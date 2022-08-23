// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

var (
	// default reconcile interval is 1 min
	defaultReconcileInterval = metav1.Duration{Duration: 1 * time.Minute}
)

// namespace maybe "", in which case releases from all namespaces are returned
func (s *Server) listReleasesInCluster(ctx context.Context, namespace string) ([]helmv2.HelmRelease, error) {
	client, err := s.getClient(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// TODO (gfichtenholt)
	// 1) we need to make sure that .List() always returns results in the
	// same order, otherwise pagination is broken (duplicates maybe returned and some entries
	// missing).
	// 2) there is a "consistent snapshot" problem, where the client doesn't want to
	// see any results created/updated/deleted after the first request is issued
	// To fix this, we must make use of resourceVersion := relList.GetResourceVersion()
	var relList helmv2.HelmReleaseList
	if err = client.List(ctx, &relList); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmRelease", namespace+"/*", err)
	} else {
		return relList.Items, nil
	}
}

func (s *Server) getReleaseInCluster(ctx context.Context, key types.NamespacedName) (*helmv2.HelmRelease, error) {
	client, err := s.getClient(ctx, key.Namespace)
	if err != nil {
		return nil, err
	}

	var rel helmv2.HelmRelease
	if err = client.Get(ctx, key, &rel); err != nil {
		return nil, statuserror.FromK8sError("get", "HelmRelease", key.String(), err)
	}
	return &rel, nil
}

func (s *Server) paginatedInstalledPkgSummaries(ctx context.Context, namespace string, pageSize int32, itemOffset int) ([]*corev1.InstalledPackageSummary, error) {
	releasesFromCluster, err := s.listReleasesInCluster(ctx, namespace)
	if err != nil {
		return nil, err
	}

	installedPkgSummaries := []*corev1.InstalledPackageSummary{}
	if len(releasesFromCluster) > 0 {
		startAt := -1
		if pageSize > 0 {
			startAt = itemOffset
		}

		for i, r := range releasesFromCluster {
			if startAt <= i {
				summary, err := s.installedPkgSummaryFromRelease(ctx, r)
				if err != nil {
					return nil, err
				} else if summary == nil {
					// not ready yet
					continue
				}
				installedPkgSummaries = append(installedPkgSummaries, summary)
				if pageSize > 0 && len(installedPkgSummaries) == int(pageSize) {
					break
				}
			}
		}
	}
	return installedPkgSummaries, nil
}

func (s *Server) installedPkgSummaryFromRelease(ctx context.Context, rel helmv2.HelmRelease) (*corev1.InstalledPackageSummary, error) {
	// first check if release CR is ready or is in "flux"
	if !checkReleaseGeneration(rel) {
		return nil, nil
	}

	name, err := common.NamespacedName(&rel)
	if err != nil {
		return nil, err
	}

	var latestPkgVersion *corev1.PackageAppVersion
	var pkgVersion *corev1.VersionReference

	version := rel.Spec.Chart.Spec.Version
	if version != "" {
		pkgVersion = &corev1.VersionReference{
			Version: version,
		}
	}

	var pkgDetail *corev1.AvailablePackageDetail

	// see https://fluxcd.io/docs/components/helm/helmreleases/
	helmChartRef := rel.Status.HelmChart
	if helmChartRef == "" {
		log.Warningf("Missing element status.helmChart on HelmRelease [%s]", name)
	}

	repoName := rel.Spec.Chart.Spec.SourceRef.Name
	repoNamespace := rel.Spec.Chart.Spec.SourceRef.Namespace
	chartName := rel.Spec.Chart.Spec.Chart

	if repoName != "" && helmChartRef != "" && chartName != "" {
		parts := strings.Split(helmChartRef, "/")
		if len(parts) != 2 {
			return nil, status.Errorf(codes.InvalidArgument, "Incorrect package ref dentifier, currently just 'foo/bar' patterns are supported: %s", helmChartRef)
		} else {
			chartKey := types.NamespacedName{Name: parts[1], Namespace: parts[0]}
			// not important to use the chart cache here, since the tar URL will be from a local cluster
			if chart, err := s.getChartInCluster(ctx, chartKey); err != nil {
				log.Warningf("Failed to get HelmChart [%s] due to: %+v", helmChartRef, err)
			} else {
				tarUrl := chart.Status.URL
				if tarUrl == "" {
					log.Warningf("Missing element status.url on HelmRelease [%s]", name)
				} else {
					chartID := fmt.Sprintf("%s/%s", repoName, chartName)
					// fetch, unzip and untar .tgz file
					// no need to provide authz, userAgent or any of the TLS details, as we are pulling .tgz file from
					// local cluster, not remote repo.
					// E.g. http://source-controller.flux-system.svc.cluster.local./helmchart/default/redis-j6wtx/redis-latest.tgz
					// Flux does the hard work of pulling the bits from remote repo
					// based on secretRef associated with HelmRepository, if applicable
					chartDetail, err := tarutil.FetchChartDetailFromTarballUrl(chartID, tarUrl, "", "", httpclient.New())
					if err != nil {
						return nil, err
					}
					if pkgDetail, err = availablePackageDetailFromChartDetail(chartID, chartDetail); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// according to flux docs repoNamespace is optional
	if repoNamespace == "" {
		repoNamespace = name.Namespace
	}
	repo := types.NamespacedName{Namespace: repoNamespace, Name: repoName}
	chartFromCache, err := s.getChartModel(ctx, repo, chartName)
	if err != nil {
		log.Warningf("%v", err)
	} else if chartFromCache != nil && len(chartFromCache.ChartVersions) > 0 {
		// charts in cache are already sorted with the latest being at position 0
		latestPkgVersion = &corev1.PackageAppVersion{
			PkgVersion: chartFromCache.ChartVersions[0].Version,
			AppVersion: chartFromCache.ChartVersions[0].AppVersion,
		}
	}

	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: name.Namespace,
				Cluster:   s.kubeappsCluster,
			},
			Identifier: name.Name,
			Plugin:     GetPluginDetail(),
		},
		Name:                name.Name,
		PkgVersionReference: pkgVersion,
		CurrentVersion:      pkgDetail.GetVersion(),
		IconUrl:             pkgDetail.GetIconUrl(),
		PkgDisplayName:      pkgDetail.GetDisplayName(),
		ShortDescription:    pkgDetail.GetShortDescription(),
		Status:              installedPackageStatus(rel),
		LatestVersion:       latestPkgVersion,
		// TODO (gfichtenholt) LatestMatchingPkgVersion
		// Only non-empty if an available upgrade matches the specified pkg_version_reference.
		// For example, if the pkg_version_reference is ">10.3.0 < 10.4.0" and 10.3.1
		// is installed, then:
		//   * if 10.3.2 is available, latest_matching_version should be 10.3.2, but
		//   * if 10.4 is available while >10.3.1 is not, this should remain empty.
	}, nil
}

func (s *Server) installedPackageDetail(ctx context.Context, key types.NamespacedName) (*corev1.InstalledPackageDetail, error) {
	rel, err := s.getReleaseInCluster(ctx, key)
	if err != nil {
		return nil, err
	}

	log.V(4).Infof("installedPackageDetail:\n[%s]", common.PrettyPrint(rel))

	var pkgVersionRef *corev1.VersionReference
	version := rel.Spec.Chart.Spec.Version
	if version != "" {
		pkgVersionRef = &corev1.VersionReference{
			Version: version,
		}
	}

	var valuesApplied = ""
	valuesJson := rel.Spec.Values
	if valuesJson != nil {
		valuesApplied = string(valuesJson.Raw)
	}

	// TODO (gfichtenholt) what about ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`?
	// ValuesReference maybe a config map or a secret

	// this will only be present if install/upgrade succeeded
	pkgVersion := rel.Status.LastAppliedRevision
	if pkgVersion == "" {
		// this is the back-up option: will be there if the reconciliation is in progress or has failed
		pkgVersion = rel.Status.LastAttemptedRevision
	}

	availablePackageRef, err := installedPackageAvailablePackageRef(rel)
	if err != nil {
		return nil, err
	}
	// per https://github.com/vmware-tanzu/kubeapps/pull/3686#issue-1038093832
	availablePackageRef.Context.Cluster = s.kubeappsCluster

	appVersion, postInstallNotes := "", ""
	rel2, err := s.getReleaseViaHelmApi(ctx, key, rel)
	// err maybe NotFound if this object has just been created and flux hasn't had time
	// to invoke helm layer yet
	if err == nil && rel != nil {
		// a couple of fields currrently only available via helm API
		if rel2.Chart != nil {
			appVersion = rel2.Chart.AppVersion()
		}
		if rel2.Info != nil {
			postInstallNotes = rel2.Info.Notes
		}
	} else if err != nil && !errors.IsNotFound(err) {
		log.Warningf("Failed to get helm release due to %v", err)
	}

	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: key.Namespace,
				Cluster:   s.kubeappsCluster,
			},
			Identifier: key.Name,
			Plugin:     GetPluginDetail(),
		},
		Name:                key.Name,
		PkgVersionReference: pkgVersionRef,
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: pkgVersion,
			AppVersion: appVersion,
		},
		ValuesApplied:         valuesApplied,
		ReconciliationOptions: installedPackageReconciliationOptions(rel),
		AvailablePackageRef:   availablePackageRef,
		PostInstallationNotes: postInstallNotes,
		Status:                installedPackageStatus(*rel),
	}, nil
}

func (s *Server) getReleaseViaHelmApi(ctx context.Context, key types.NamespacedName, rel *helmv2.HelmRelease) (*release.Release, error) {
	// post installation notes can only be retrieved via helm APIs, flux doesn't do it
	// see discussion in https://cloud-native.slack.com/archives/CLAJ40HV3/p1629244025187100
	if s.actionConfigGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Server is not configured with actionConfigGetter")
	}

	helmRel := helmReleaseName(key, rel)
	actionConfig, err := s.actionConfigGetter(ctx, helmRel.Namespace)
	if err != nil || actionConfig == nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config in namespace [%s] due to: %v", key.Namespace, err)
	}
	cmd := action.NewGet(actionConfig)
	release, err := cmd.Run(helmRel.Name)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release [%s] in namespace [%s]", helmRel, key.Namespace)
		}
		return nil, status.Errorf(codes.NotFound, "Unable to run Helm Get action for release [%s] in namespace [%s]: %v", helmRel, key.Namespace, err)
	}
	return release, nil
}

func (s *Server) newRelease(ctx context.Context, packageRef *corev1.AvailablePackageReference, targetName types.NamespacedName, versionRef *corev1.VersionReference, reconcile *corev1.ReconciliationOptions, valuesString string) (*corev1.InstalledPackageReference, error) {
	repoName, chartName, err := pkgutils.SplitPackageIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	repo := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: repoName}
	chart, err := s.getChartModel(ctx, repo, chartName)
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	if valuesString != "" {
		// maybe JSON or YAML
		values = make(map[string]interface{})
		if err = yaml.Unmarshal([]byte(valuesString), &values); err != nil {
			return nil, err
		}
	}

	// Calculate the version constraints
	versionExpr := versionRef.GetVersion()
	if versionExpr != "" {
		versionExpr, err = pkgutils.VersionConstraintWithUpgradePolicy(
			versionRef.GetVersion(), s.pluginConfig.DefaultUpgradePolicy)
		if err != nil {
			return nil, err
		}
	}

	fluxRelease, err := s.newFluxHelmRelease(chart, targetName, versionExpr, reconcile, values)
	if err != nil {
		return nil, err
	}

	// per https://github.com/vmware-tanzu/kubeapps/pull/3640#issuecomment-949315105
	// the helm release CR to also be created in the target namespace (where the helm
	// release itself is currently created)
	client, err := s.getClient(ctx, targetName.Namespace)
	if err != nil {
		return nil, err
	}

	if err = client.Create(ctx, fluxRelease); err != nil {
		return nil, statuserror.FromK8sError("create", "HelmRelease", targetName.String(), err)
	}

	return &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: targetName.Namespace,
			Cluster:   s.kubeappsCluster,
		},
		Identifier: targetName.Name,
		Plugin:     GetPluginDetail(),
	}, nil
}

func (s *Server) updateRelease(ctx context.Context, packageRef *corev1.InstalledPackageReference, versionRef *corev1.VersionReference, reconcile *corev1.ReconciliationOptions, valuesString string) (*corev1.InstalledPackageReference, error) {
	key := types.NamespacedName{Name: packageRef.Identifier, Namespace: packageRef.Context.Namespace}

	rel, err := s.getReleaseInCluster(ctx, key)
	if err != nil {
		return nil, err
	}

	// TODO (gfichtenholt): there is an intermittent issue
	// rpc error: code = Internal desc = unable to update the HelmRelease
	// 'test-12-i7a4/my-podinfo-12' due to 'Operation cannot be fulfilled on
	// helmreleases.helm.toolkit.fluxcd.io "my-podinfo-12": the object has been
	// modified; please apply your changes to the latest version and try again'
	// the problem is
	//  1) we get a CR then
	//  2) do some modifications of this CR
	//  3) call Update() with this CR.
	// Every once in a while there the CR gets updated (by flux) between (1) and (3)
	// and we get this error.
	// I think one way to fix it would be to implement a fixed number of retries in
	// flux plugin with exponential back-off.
	// Another solution might be to push the decision all the way to the end user to retry
	// the Update operation

	// As Michael and I agreed 4/12/2022, initially we'll disallow updates to pending releases
	// to simplify the initial case, though we may implement support later. Updates to
	// non-pending releases  (i.e. success or failed status) are allowed
	_, reason, _ := isHelmReleaseReady(*rel)
	if reason == corev1.InstalledPackageStatus_STATUS_REASON_PENDING {
		return nil, status.Errorf(codes.Internal, "updates to helm releases pending reconciliation are not supported")
	}

	versionExpr := versionRef.GetVersion()
	if versionExpr != "" {
		versionExpr, err = pkgutils.VersionConstraintWithUpgradePolicy(
			versionRef.GetVersion(), s.pluginConfig.DefaultUpgradePolicy)
		if err != nil {
			return nil, err
		}
		rel.Spec.Chart.Spec.Version = versionExpr
	} else {
		rel.Spec.Chart.Spec.Version = ""
	}

	if valuesString != "" {
		// could be JSON or YAML
		var values map[string]interface{}
		if err = yaml.Unmarshal([]byte(valuesString), &values); err != nil {
			return nil, err
		}
		byteArray, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		rel.Spec.Values = &v1.JSON{Raw: byteArray}
	} else {
		rel.Spec.Values = nil
	}

	setInterval, setServiceAccount := false, false
	if reconcile != nil {
		if reconcile.Interval != "" {
			reconcileInterval, err := pkgutils.ToDuration(reconcile.Interval)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "the reconciliation interval is invalid: %v", err)
			}
			rel.Spec.Interval = *reconcileInterval
			setInterval = true
		}
		if reconcile.ServiceAccountName != "" {
			rel.Spec.ServiceAccountName = reconcile.ServiceAccountName
			setServiceAccount = true
		}

		rel.Spec.Suspend = reconcile.Suspend
	}

	if !setInterval {
		// interval is a required field
		rel.Spec.Interval = defaultReconcileInterval
	}
	if !setServiceAccount {
		rel.Spec.ServiceAccountName = ""
	}

	// get rid of the status field, since now there will be a new reconciliation
	// process and the current status no longer applies. metadata and spec I want
	// to keep, as they may have had added labels and/or annotations and/or
	// even other changes made by the user.
	rel.Status = helmv2.HelmReleaseStatus{}

	client, err := s.getClient(ctx, packageRef.Context.Namespace)
	if err != nil {
		return nil, err
	}

	if err = client.Update(ctx, rel); err != nil {
		return nil, statuserror.FromK8sError("update", "HelmRelease", key.String(), err)
	}

	log.V(4).Infof("Updated release: %s", common.PrettyPrint(rel))

	return &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: packageRef.Context.Namespace,
			Cluster:   s.kubeappsCluster,
		},
		Identifier: packageRef.Identifier,
		Plugin:     GetPluginDetail(),
	}, nil
}

func (s *Server) deleteRelease(ctx context.Context, packageRef *corev1.InstalledPackageReference) error {
	client, err := s.getClient(ctx, packageRef.Context.Namespace)
	if err != nil {
		return err
	}

	log.V(4).Infof("Deleting release: [%s]", packageRef.Identifier)

	rel := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      packageRef.Identifier,
			Namespace: packageRef.Context.Namespace,
		},
	}

	if err = client.Delete(ctx, rel); err != nil {
		return statuserror.FromK8sError("delete", "HelmRelease", packageRef.Identifier, err)
	}
	return nil
}

// Potentially, there are 3 different namespaces that can be specified here
// 1. spec.chart.spec.sourceRef.namespace, where HelmRepository CRD object referenced exists
// 2. metadata.namespace, where this HelmRelease CRD will exist, same as (3) below
//    per https://github.com/vmware-tanzu/kubeapps/pull/3640#issuecomment-949315105
// 3. spec.targetNamespace, where flux will install any artifacts from the release
func (s *Server) newFluxHelmRelease(chart *models.Chart, targetName types.NamespacedName, versionExpr string, reconcile *corev1.ReconciliationOptions, values map[string]interface{}) (*helmv2.HelmRelease, error) {
	fluxRelease := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName.Name,
			Namespace: targetName.Namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: chart.Name,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      chart.Repo.Name,
						Kind:      sourcev1.HelmRepositoryKind,
						Namespace: chart.Repo.Namespace,
					},
				},
			},
		},
	}
	if versionExpr != "" {
		fluxRelease.Spec.Chart.Spec.Version = versionExpr
	}

	reconcileInterval := defaultReconcileInterval // unless explicitly specified
	if reconcile != nil {
		if reconcile.Interval != "" {
			if duration, err := pkgutils.ToDuration(reconcile.Interval); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "the reconciliation interval is invalid: %v", err)
			} else {
				reconcileInterval = *duration
			}
		}
		fluxRelease.Spec.Suspend = reconcile.Suspend
		if reconcile.ServiceAccountName != "" {
			fluxRelease.Spec.ServiceAccountName = reconcile.ServiceAccountName
		}
	}
	if values != nil {
		byteArray, err := json.Marshal(values)
		if err != nil {
			return nil, err
		}
		fluxRelease.Spec.Values = &v1.JSON{Raw: byteArray}
	}

	// required fields, without which flux controller will fail to create the CRD
	fluxRelease.Spec.Interval = reconcileInterval

	// ref https://fluxcd.io/docs/components/helm/helmreleases/
	// TODO (gfichtenholt) in theory, flux allows a timeout per release.
	// So far we just use one configured per server/installation, if specified,
	// same as helm plug-in.
	// Otherwise the default timeout is used.
	if s.pluginConfig.TimeoutSeconds > 0 {
		timeoutInterval := metav1.Duration{Duration: time.Duration(s.pluginConfig.TimeoutSeconds) * time.Second}
		fluxRelease.Spec.Timeout = &timeoutInterval
	}
	return fluxRelease, nil
}

// returns 3 things:
// - ready:  whether the HelmRelease object is in a ready state
// - reason: one of SUCCESS/FAILURE/PENDING/UNSPECIFIED,
// - userReason: textual description of why the object is in current state, if present
// docs:
// 1. https://fluxcd.io/docs/components/helm/helmreleases/#examples
// 2. discussion on private slack channel. Summary:
//    - "ready" field: - it's not indicating that the resource has completed (i.e. whether the task
//      completed with install or failure), but rather just whether the resource is ready or not.
//      So it can be false because of either a final state (failure) or a pending state
//      (reconciliation in progress or whatever). That means the ready flag will only be set to true
//       when install completes with success
//    - "reason" field: failure only when flux returns "InstallFailed" reason
//       otherwise pending or unspecified when there are no status conditions to go by
//
func isHelmReleaseReady(rel helmv2.HelmRelease) (ready bool, status corev1.InstalledPackageStatus_StatusReason, userReason string) {
	if !checkReleaseGeneration(rel) {
		return false, corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED, ""
	}

	isInstallFailed := false
	readyCond := meta.FindStatusCondition(rel.GetConditions(), fluxmeta.ReadyCondition)

	if readyCond != nil {
		if readyCond.Reason != "" {
			// this could be something like
			// "reason": "InstallFailed"
			// i.e. not super-useful
			userReason = readyCond.Reason
			if userReason == helmv2.InstallFailedReason ||
				userReason == helmv2.UpgradeFailedReason {
				isInstallFailed = true
			}
		}
		if readyCond.Message != "" {
			// whereas this could be something like:
			// "message": 'Helm install failed: unable to build kubernetes objects from
			// release manifest: error validating "": error validating data:
			// ValidationError(Deployment.spec.replicas): invalid type for
			// io.k8s.api.apps.v1.DeploymentSpec.replicas: got "string", expected "integer"'
			// i.e. a little more useful, so we'll just return them both
			userReason += ": " + readyCond.Message
		}
		if readyCond.Status == metav1.ConditionTrue {
			return true, corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED, userReason
		} else if isInstallFailed {
			return false, corev1.InstalledPackageStatus_STATUS_REASON_FAILED, userReason
		}
	}

	// catch all: unless we know something else, install is pending
	return false, corev1.InstalledPackageStatus_STATUS_REASON_PENDING, userReason
}

func installedPackageStatus(rel helmv2.HelmRelease) *corev1.InstalledPackageStatus {
	ready, reason, userReason := isHelmReleaseReady(rel)
	return &corev1.InstalledPackageStatus{
		Ready:      ready,
		Reason:     reason,
		UserReason: userReason,
	}
}

func installedPackageReconciliationOptions(rel *helmv2.HelmRelease) *corev1.ReconciliationOptions {
	reconciliationOptions := &corev1.ReconciliationOptions{}
	reconciliationOptions.Interval = pkgutils.FromDuration(&rel.Spec.Interval)
	reconciliationOptions.Suspend = rel.Spec.Suspend
	reconciliationOptions.ServiceAccountName = rel.Spec.ServiceAccountName
	return reconciliationOptions
}

func installedPackageAvailablePackageRef(rel *helmv2.HelmRelease) (*corev1.AvailablePackageReference, error) {
	repoName := rel.Spec.Chart.Spec.SourceRef.Name
	if repoName == "" {
		return nil, status.Errorf(codes.Internal, "missing required field spec.chart.spec.sourceRef.name")
	}
	chartName := rel.Spec.Chart.Spec.Chart
	if chartName == "" {
		return nil, status.Errorf(codes.Internal, "missing required field spec.chart.spec.chart")
	}
	repoNamespace := rel.Spec.Chart.Spec.SourceRef.Namespace
	// CrossNamespaceObjectReference namespace is optional, so
	if repoNamespace == "" {
		name, err := common.NamespacedName(rel)
		if err != nil {
			return nil, err
		}
		repoNamespace = name.Namespace
	}
	return &corev1.AvailablePackageReference{
		Identifier: fmt.Sprintf("%s/%s", repoName, chartName),
		Plugin:     GetPluginDetail(),
		Context:    &corev1.Context{Namespace: repoNamespace},
	}, nil
}

// ref https://fluxcd.io/docs/components/helm/helmreleases/
func helmReleaseName(key types.NamespacedName, rel *helmv2.HelmRelease) types.NamespacedName {
	helmReleaseName := rel.Spec.ReleaseName
	// according to docs ReleaseName is optional and defaults to a composition of
	// '[TargetNamespace-]Name'.
	if helmReleaseName == "" {
		// according to docs targetNamespace is optional and defaults to the namespace
		// of the HelmRelease
		if rel.Spec.TargetNamespace == "" {
			helmReleaseName = key.Name
		} else {
			helmReleaseName = fmt.Sprintf("%s-%s", rel.Spec.TargetNamespace, key.Name)
		}
	}

	helmReleaseNamespace := rel.Spec.TargetNamespace
	if helmReleaseNamespace == "" {
		helmReleaseNamespace = key.Namespace
	}
	return types.NamespacedName{Name: helmReleaseName, Namespace: helmReleaseNamespace}
}

func checkReleaseGeneration(rel helmv2.HelmRelease) bool {
	generation := rel.GetGeneration()
	observedGeneration := rel.Status.ObservedGeneration
	return generation > 0 && generation == observedGeneration
}
