// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pkgfluxv2common "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	pkgutils "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	statuserror "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	tarutil "github.com/kubeapps/kubeapps/pkg/tarutil"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smetaunstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8dynamicclient "k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	// see docs at https://fluxcd.io/docs/components/helm/api/
	fluxHelmReleaseGroup   = "helm.toolkit.fluxcd.io"
	fluxHelmReleaseVersion = "v2beta1"
	fluxHelmRelease        = "HelmRelease"
	fluxHelmReleases       = "helmreleases"
	fluxHelmReleaseList    = "HelmReleaseList"

	defaultReconcileInterval = "1m"
)

func (s *Server) getReleasesResourceInterface(ctx context.Context, namespace string) (k8dynamicclient.ResourceInterface, error) {
	_, client, _, err := s.GetClients(ctx)
	if err != nil {
		return nil, err
	}

	releasesResource := k8sschema.GroupVersionResource{
		Group:    fluxHelmReleaseGroup,
		Version:  fluxHelmReleaseVersion,
		Resource: fluxHelmReleases,
	}

	return client.Resource(releasesResource).Namespace(namespace), nil
}

// namespace maybe "", in which case releases from all namespaces are returned
func (s *Server) listReleasesInCluster(ctx context.Context, namespace string) (*k8smetaunstructuredv1.UnstructuredList, error) {
	releasesIfc, err := s.getReleasesResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if releases, err := releasesIfc.List(ctx, k8smetav1.ListOptions{}); err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "unable to list fluxv2 helmreleases: %v", err)
	} else {
		return releases, nil
	}
}

func (s *Server) paginatedInstalledPkgSummaries(ctx context.Context, namespace string, pageSize int32, pageOffset int) ([]*pkgsGRPCv1alpha1.InstalledPackageSummary, error) {
	releasesFromCluster, err := s.listReleasesInCluster(ctx, namespace)
	if err != nil {
		return nil, err
	}

	installedPkgSummaries := []*pkgsGRPCv1alpha1.InstalledPackageSummary{}
	if len(releasesFromCluster.Items) > 0 {
		// we're going to need this later
		// TODO (gfichtenholt) for now we get all charts and later find one that helmrelease is using
		// there is probably a more efficient way to do this
		chartsFromCluster, err := s.listChartsInCluster(ctx, k8scorev1.NamespaceAll)
		if err != nil {
			return nil, err
		}

		startAt := -1
		if pageSize > 0 {
			startAt = int(pageSize) * pageOffset
		}

		for i, releaseUnstructured := range releasesFromCluster.Items {
			if startAt <= i {
				summary, err := s.installedPkgSummaryFromRelease(ctx, releaseUnstructured.Object, chartsFromCluster)
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

func (s *Server) installedPkgSummaryFromRelease(ctx context.Context, unstructuredRelease map[string]interface{}, chartsFromCluster *k8smetaunstructuredv1.UnstructuredList) (*pkgsGRPCv1alpha1.InstalledPackageSummary, error) {
	// first check if release CR is ready or is in "flux"
	if !pkgfluxv2common.CheckGeneration(unstructuredRelease) {
		return nil, nil
	}

	name, err := pkgfluxv2common.NamespacedName(unstructuredRelease)
	if err != nil {
		return nil, err
	}

	var latestPkgVersion *pkgsGRPCv1alpha1.PackageAppVersion
	var pkgVersion *pkgsGRPCv1alpha1.VersionReference

	version, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersion = &pkgsGRPCv1alpha1.VersionReference{
			Version: version,
		}
	}

	var pkgDetail *pkgsGRPCv1alpha1.AvailablePackageDetail

	// see https://fluxcd.io/docs/components/helm/helmreleases/
	helmChartRef, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "status", "helmChart")
	if err != nil || !found || helmChartRef == "" {
		log.Warningf("Missing element grpcstatus.helmChart on HelmRelease [%s]", name)
	}

	repoName, _, _ := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "name")
	repoNamespace, _, _ := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "namespace")
	chartName, _, _ := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "chart")

	if repoName != "" && helmChartRef != "" && chartName != "" {
		if parts := strings.Split(helmChartRef, "/"); len(parts) != 2 {
			return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Incorrect package ref dentifier, currently just 'foo/bar' patterns are supported: %s", helmChartRef)
		} else if ifc, err := s.getChartsResourceInterface(ctx, parts[0]); err != nil {
			log.Warningf("Failed to get HelmChart [%s] due to: %+v", helmChartRef, err)
		} else if unstructuredChart, err := ifc.Get(ctx, parts[1], k8smetav1.GetOptions{}); err != nil {
			log.Warningf("Failed to get HelmChart [%s] due to: %+v", helmChartRef, err)
		} else {
			tarUrl, found, err := k8smetaunstructuredv1.NestedString(unstructuredChart.Object, "status", "url")
			if err != nil || !found || tarUrl == "" {
				log.Warningf("Missing element grpcstatus.url on HelmRelease [%s]", name)
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
				pkgDetail, err = availablePackageDetailFromChartDetail(chartID, chartDetail)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// according to flux docs repoNamespace is optional
	if repoNamespace == "" {
		repoNamespace = name.Namespace
	}
	repo := k8stypes.NamespacedName{Namespace: repoNamespace, Name: repoName}
	chartFromCache, err := s.getChart(ctx, repo, chartName)
	if err != nil {
		return nil, err
	} else if chartFromCache != nil && len(chartFromCache.ChartVersions) > 0 {
		// charts in cache are already sorted with the latest being at position 0
		latestPkgVersion = &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: chartFromCache.ChartVersions[0].Version,
			AppVersion: chartFromCache.ChartVersions[0].AppVersion,
		}
	}

	return &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
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
		Status:              installedPackageStatusFromUnstructured(unstructuredRelease),
		LatestVersion:       latestPkgVersion,
		// TODO (gfichtenholt) LatestMatchingPkgVersion
		// Only non-empty if an available upgrade matches the specified pkg_version_reference.
		// For example, if the pkg_version_reference is ">10.3.0 < 10.4.0" and 10.3.1
		// is installed, then:
		//   * if 10.3.2 is available, latest_matching_version should be 10.3.2, but
		//   * if 10.4 is available while >10.3.1 is not, this should remain empty.
	}, nil
}

func (s *Server) installedPackageDetail(ctx context.Context, name k8stypes.NamespacedName) (*pkgsGRPCv1alpha1.InstalledPackageDetail, error) {
	releasesIfc, err := s.getReleasesResourceInterface(ctx, name.Namespace)
	if err != nil {
		return nil, err
	}
	unstructuredRelease, err := releasesIfc.Get(ctx, name.Name, k8smetav1.GetOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "HelmRelease", name.String(), err)
	}

	log.V(4).Infof("installedPackageDetail:\n[%s]", pkgfluxv2common.PrettyPrintMap(unstructuredRelease.Object))

	obj := unstructuredRelease.Object
	var pkgVersionRef *pkgsGRPCv1alpha1.VersionReference
	version, found, err := k8smetaunstructuredv1.NestedString(obj, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersionRef = &pkgsGRPCv1alpha1.VersionReference{
			Version: version,
		}
	}

	valuesApplied := ""
	valuesMap, found, err := k8smetaunstructuredv1.NestedMap(obj, "spec", "values")
	if found && err == nil && len(valuesMap) != 0 {
		bytes, err := json.Marshal(valuesMap)
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to marshal Helm release values due to: %v", err)
		}
		valuesApplied = string(bytes)
	}
	// TODO (gfichtenholt) what about ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`?
	// ValuesReference maybe a config map or a secret

	// this will only be present if install/upgrade succeeded
	pkgVersion, found, err := k8smetaunstructuredv1.NestedString(obj, "status", "lastAppliedRevision")
	if !found || err != nil || pkgVersion == "" {
		// this is the back-up option: will be there if the reconciliation is in progress or has failed
		pkgVersion, _, _ = k8smetaunstructuredv1.NestedString(obj, "status", "lastAttemptedRevision")
	}

	availablePackageRef, err := installedPackageAvailablePackageRefFromUnstructured(obj)
	if err != nil {
		return nil, err
	}
	// per https://github.com/kubeapps/kubeapps/pull/3686#issue-1038093832
	availablePackageRef.Context.Cluster = s.kubeappsCluster

	appVersion, postInstallNotes := "", ""
	release, err := s.helmReleaseFromUnstructured(ctx, name, obj)
	// err maybe NotFound if this object has just been created and flux hasn't had time
	// to invoke helm layer yet
	if err == nil && release != nil {
		// a couple of fields currrently only available via helm API
		if release.Chart != nil {
			appVersion = release.Chart.AppVersion()
		}
		if release.Info != nil {
			postInstallNotes = release.Info.Notes
		}
	} else if err != nil && !k8serrors.IsNotFound(err) {
		log.Warningf("Failed to get helm release due to %v", err)
	}

	return &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Namespace: name.Namespace,
				Cluster:   s.kubeappsCluster,
			},
			Identifier: name.Name,
			Plugin:     GetPluginDetail(),
		},
		Name:                name.Name,
		PkgVersionReference: pkgVersionRef,
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: pkgVersion,
			AppVersion: appVersion,
		},
		ValuesApplied:         valuesApplied,
		ReconciliationOptions: installedPackageReconciliationOptionsFromUnstructured(obj),
		AvailablePackageRef:   availablePackageRef,
		PostInstallationNotes: postInstallNotes,
		Status:                installedPackageStatusFromUnstructured(obj),
	}, nil
}

func (s *Server) helmReleaseFromUnstructured(ctx context.Context, name k8stypes.NamespacedName, unstructuredRelease map[string]interface{}) (*helmrelease.Release, error) {
	// post installation notes can only be retrieved via helm APIs, flux doesn't do it
	// see discussion in https://cloud-native.slack.com/archives/CLAJ40HV3/p1629244025187100
	if s.actionConfigGetter == nil {
		return nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "Server is not configured with actionConfigGetter")
	}

	helmReleaseName, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "ReleaseName")
	// according to docs ReleaseName is optional and defaults to a composition of
	// '[TargetNamespace-]Name'.
	if !found || err != nil || helmReleaseName == "" {
		targetNamespace, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "targetNamespace")
		// according to docs targetNamespace is optional and defaults to the namespace of the HelmRelease
		if !found || err != nil || targetNamespace == "" {
			targetNamespace = name.Namespace
		}
		helmReleaseName = fmt.Sprintf("%s-%s", targetNamespace, name.Name)
	}

	actionConfig, err := s.actionConfigGetter(ctx, name.Namespace)
	if err != nil || actionConfig == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config in namespace [%s] due to: %v", name.Namespace, err)
	}
	cmd := helmaction.NewGet(actionConfig)
	release, err := cmd.Run(helmReleaseName)
	if err != nil {
		if err == helmstoragedriver.ErrReleaseNotFound {
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to find Helm release [%s] in namespace [%s]", helmReleaseName, name.Namespace)
		}
		return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to run Helm Get action for release [%s] in namespace [%s]: %v", helmReleaseName, name.Namespace, err)
	}
	return release, nil
}

func (s *Server) newRelease(ctx context.Context, packageRef *pkgsGRPCv1alpha1.AvailablePackageReference, targetName k8stypes.NamespacedName, versionRef *pkgsGRPCv1alpha1.VersionReference, reconcile *pkgsGRPCv1alpha1.ReconciliationOptions, valuesString string) (*pkgsGRPCv1alpha1.InstalledPackageReference, error) {
	repoName, chartName, err := pkgutils.SplitChartIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	repo := k8stypes.NamespacedName{Namespace: packageRef.Context.Namespace, Name: repoName}
	chart, err := s.getChart(ctx, repo, chartName)
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	if valuesString != "" {
		values = make(map[string]interface{})
		err = k8syaml.Unmarshal([]byte(valuesString), &values)
		if err != nil {
			return nil, err
		}
	}

	fluxHelmRelease, err := s.newFluxHelmRelease(chart, targetName, versionRef, reconcile, values)
	if err != nil {
		return nil, err
	}

	// per https://github.com/kubeapps/kubeapps/pull/3640#issuecomment-949315105
	// the helm release CR to also be created in the target namespace (where the helm release itself is currently created)
	resourceIfc, err := s.getReleasesResourceInterface(ctx, targetName.Namespace)
	if err != nil {
		return nil, err
	}
	newRelease, err := resourceIfc.Create(ctx, fluxHelmRelease, k8smetav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsForbidden(err) || k8serrors.IsUnauthorized(err) {
			// TODO (gfichtenholt) I think in some cases we should be returning grpccodes.PermissionDenied instead,
			// but that has to be done consistently across all plug-in operations, not just here
			return nil, grpcstatus.Errorf(grpccodes.Unauthenticated, "Unable to create release due to %v", err)
		} else {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create release due to %v", err)
		}
	}

	name, err := pkgfluxv2common.NamespacedName(newRelease.Object)
	if err != nil {
		return nil, err
	}
	return &pkgsGRPCv1alpha1.InstalledPackageReference{
		Context: &pkgsGRPCv1alpha1.Context{
			Namespace: name.Namespace,
			Cluster:   s.kubeappsCluster,
		},
		Identifier: name.Name,
		Plugin:     GetPluginDetail(),
	}, nil
}

func (s *Server) updateRelease(ctx context.Context, packageRef *pkgsGRPCv1alpha1.InstalledPackageReference, versionRef *pkgsGRPCv1alpha1.VersionReference, reconcile *pkgsGRPCv1alpha1.ReconciliationOptions, valuesString string) (*pkgsGRPCv1alpha1.InstalledPackageReference, error) {
	ifc, err := s.getReleasesResourceInterface(ctx, packageRef.Context.Namespace)
	if err != nil {
		return nil, err
	}

	unstructuredRel, err := ifc.Get(ctx, packageRef.Identifier, k8smetav1.GetOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "HelmRelease", packageRef.Identifier, err)
	}

	if versionRef.GetVersion() != "" {
		if err = k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, versionRef.GetVersion(), "spec", "chart", "spec", "version"); err != nil {
			return nil, err
		}
	} else {
		k8smetaunstructuredv1.RemoveNestedField(unstructuredRel.Object, "spec", "chart", "spec", "version")
	}

	if valuesString != "" {
		values := make(map[string]interface{})
		if err = k8syaml.Unmarshal([]byte(valuesString), &values); err != nil {
			return nil, err
		} else if err = k8smetaunstructuredv1.SetNestedMap(unstructuredRel.Object, values, "spec", "values"); err != nil {
			return nil, err
		}
	} else {
		k8smetaunstructuredv1.RemoveNestedField(unstructuredRel.Object, "spec", "values")
	}

	setInterval, setServiceAccount := false, false
	if reconcile != nil {
		if reconcile.Interval > 0 {
			reconcileInterval := (time.Duration(reconcile.Interval) * time.Second).String()
			if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcileInterval, "spec", "interval"); err != nil {
				return nil, err
			}
			setInterval = true
		}
		if reconcile.ServiceAccountName != "" {
			if err = k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcile.ServiceAccountName, "spec", "serviceAccountName"); err != nil {
				setServiceAccount = true
			}
		}
		if err = k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcile.Suspend, "spec", "suspend"); err != nil {
			return nil, err
		}
	}

	if !setInterval {
		// interval is a required field
		if err = k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, defaultReconcileInterval, "spec", "interval"); err != nil {
			return nil, err
		}
	}
	if !setServiceAccount {
		k8smetaunstructuredv1.RemoveNestedField(unstructuredRel.Object, "spec", "serviceAccountName")
	}

	// get rid of the status field, since now there will be a new reconciliation process and the current status no
	// longer applies. metadata and spec I want to keep, as they may have had added labels and/or annotations and/or
	// even other changes made by the user.
	k8smetaunstructuredv1.RemoveNestedField(unstructuredRel.Object, "status")

	// replace the object in k8s with a new desired state
	unstructuredRel, err = ifc.Update(ctx, unstructuredRel, k8smetav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsForbidden(err) || k8serrors.IsUnauthorized(err) {
			return nil, grpcstatus.Errorf(grpccodes.Unauthenticated, "Unable to update release due to %v", err)
		} else {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to update release due to %v", err)
		}
	}

	log.V(4).Infof("Updated release: %s", pkgfluxv2common.PrettyPrintMap(unstructuredRel.Object))

	return &pkgsGRPCv1alpha1.InstalledPackageReference{
		Context: &pkgsGRPCv1alpha1.Context{
			Namespace: packageRef.Context.Namespace,
			Cluster:   s.kubeappsCluster,
		},
		Identifier: packageRef.Identifier,
		Plugin:     GetPluginDetail(),
	}, nil
}

func (s *Server) deleteRelease(ctx context.Context, packageRef *pkgsGRPCv1alpha1.InstalledPackageReference) error {
	ifc, err := s.getReleasesResourceInterface(ctx, packageRef.Context.Namespace)
	if err != nil {
		return err
	}

	log.V(4).Infof("Deleted release: [%s]", packageRef.Identifier)

	if err = ifc.Delete(ctx, packageRef.Identifier, k8smetav1.DeleteOptions{}); err != nil {
		return statuserror.FromK8sError("delete", "HelmRelease", packageRef.Identifier, err)
	}
	return nil
}

// Potentially, there are 3 different namespaces that can be specified here
// 1. spec.chart.spec.sourceRef.namespace, where HelmRepository CRD object referenced exists
// 2. metadata.namespace, where this HelmRelease CRD will exist, same as (3) below
//    per https://github.com/kubeapps/kubeapps/pull/3640#issuecomment-949315105
// 3. spec.targetNamespace, where flux will install any artifacts from the release
func (s *Server) newFluxHelmRelease(chart *chartmodels.Chart, targetName k8stypes.NamespacedName, versionRef *pkgsGRPCv1alpha1.VersionReference, reconcile *pkgsGRPCv1alpha1.ReconciliationOptions, values map[string]interface{}) (*k8smetaunstructuredv1.Unstructured, error) {
	unstructuredRel := k8smetaunstructuredv1.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxHelmReleaseGroup, fluxHelmReleaseVersion),
			"kind":       fluxHelmRelease,
			"metadata": map[string]interface{}{
				"name":      targetName.Name,
				"namespace": targetName.Namespace,
			},
			"spec": map[string]interface{}{
				"chart": map[string]interface{}{
					"spec": map[string]interface{}{
						"chart": chart.Name,
						"sourceRef": map[string]interface{}{
							"name":      chart.Repo.Name,
							"kind":      fluxHelmRepository,
							"namespace": chart.Repo.Namespace,
						},
					},
				},
				"targetNamespace": targetName.Namespace,
			},
		},
	}
	if versionRef.GetVersion() != "" {
		if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, versionRef.GetVersion(), "spec", "chart", "spec", "version"); err != nil {
			return nil, err
		}
	}
	reconcileInterval := defaultReconcileInterval // unless explicitly specified
	if reconcile != nil {
		if reconcile.Interval > 0 {
			reconcileInterval = (time.Duration(reconcile.Interval) * time.Second).String()
		}
		if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcile.Suspend, "spec", "suspend"); err != nil {
			return nil, err
		}
		if reconcile.ServiceAccountName != "" {
			if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcile.ServiceAccountName, "spec", "serviceAccountName"); err != nil {
				return nil, err
			}
		}
	}
	if values != nil {
		if err := k8smetaunstructuredv1.SetNestedMap(unstructuredRel.Object, values, "spec", "values"); err != nil {
			return nil, err
		}
	}

	// required fields, without which flux controller will fail to create the CRD
	if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, reconcileInterval, "spec", "interval"); err != nil {
		return nil, err
	}

	// ref https://fluxcd.io/docs/components/helm/helmreleases/
	// TODO (gfichtenholt) in theory, flux allows a timeout per helmrelease. So far we just use one
	// configured per server/installation, if specified, same as helm plug-in.
	// Otherwise the default timeout is used.
	if s.timeoutSeconds > 0 {
		timeoutInterval := (time.Duration(s.timeoutSeconds) * time.Second).String()
		if err := k8smetaunstructuredv1.SetNestedField(unstructuredRel.Object, timeoutInterval, "spec", "timeout"); err != nil {
			return nil, err
		}
	}
	return &unstructuredRel, nil
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
func isHelmReleaseReady(unstructuredObj map[string]interface{}) (ready bool, status pkgsGRPCv1alpha1.InstalledPackageStatus_StatusReason, userReason string) {
	if !pkgfluxv2common.CheckGeneration(unstructuredObj) {
		return false, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED, ""
	}

	conditions, found, err := k8smetaunstructuredv1.NestedSlice(unstructuredObj, "status", "conditions")
	if err != nil || !found {
		return false, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED, ""
	}

	isInstallFailed := false

	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				// this could be something like
				// "reason": "InstallFailed"
				// i.e. not super-useful
				if reasonString, ok := conditionAsMap["reason"]; ok {
					userReason = fmt.Sprintf("%v", reasonString)
					if reasonString == "InstallFailed" {
						isInstallFailed = true
					}
				}
				// whereas this could be something like:
				// "message": 'Helm install failed: unable to build kubernetes objects from
				// release manifest: error validating "": error validating data:
				// ValidationError(Deployment.spec.replicas): invalid type for
				// io.k8s.api.apps.v1.DeploymentSpec.replicas: got "string", expected "integer"'
				// i.e. a little more useful, so we'll just return them both
				if messageString, ok := conditionAsMap["message"]; ok {
					userReason += fmt.Sprintf(": %v", messageString)
				}
				if statusString, ok := conditionAsMap["status"]; ok {
					if statusString == "True" {
						return true, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED, userReason
					} else if isInstallFailed {
						return false, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED, userReason
					} else {
						return false, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING, userReason
					}
				}
			}
		}
	}
	// catch all: unless we know something else, install is pending
	return false, pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING, userReason
}

func installedPackageStatusFromUnstructured(unstructuredRelease map[string]interface{}) *pkgsGRPCv1alpha1.InstalledPackageStatus {
	ready, reason, userReason := isHelmReleaseReady(unstructuredRelease)
	return &pkgsGRPCv1alpha1.InstalledPackageStatus{
		Ready:      ready,
		Reason:     reason,
		UserReason: userReason,
	}
}

func installedPackageReconciliationOptionsFromUnstructured(unstructuredRelease map[string]interface{}) *pkgsGRPCv1alpha1.ReconciliationOptions {
	reconciliationOptions := &pkgsGRPCv1alpha1.ReconciliationOptions{}
	if intervalString, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "interval"); found && err == nil {
		if duration, err := time.ParseDuration(intervalString); err == nil {
			reconciliationOptions.Interval = int32(duration.Seconds())
		}
	}
	if suspend, found, err := k8smetaunstructuredv1.NestedBool(unstructuredRelease, "spec", "suspend"); found && err == nil {
		reconciliationOptions.Suspend = suspend
	}
	if serviceAccountName, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "serviceAccountName"); found && err == nil {
		reconciliationOptions.ServiceAccountName = serviceAccountName
	}
	return reconciliationOptions
}

func installedPackageAvailablePackageRefFromUnstructured(unstructuredRelease map[string]interface{}) (*pkgsGRPCv1alpha1.AvailablePackageReference, error) {
	repoName, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "name")
	if !found || err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "missing required field spec.chart.spec.sourceRef.name")
	}
	chartName, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "chart")
	if !found || err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "missing required field spec.chart.spec.chart")
	}
	repoNamespace, found, err := k8smetaunstructuredv1.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "namespace")
	// CrossNamespaceObjectReference namespace is optional, so
	if !found || err != nil || repoNamespace == "" {
		name, err := pkgfluxv2common.NamespacedName(unstructuredRelease)
		if err != nil {
			return nil, err
		}
		repoNamespace = name.Namespace
	}
	return &pkgsGRPCv1alpha1.AvailablePackageReference{
		Identifier: fmt.Sprintf("%s/%s", repoName, chartName),
		Plugin:     GetPluginDetail(),
		Context:    &pkgsGRPCv1alpha1.Context{Namespace: repoNamespace},
	}, nil
}
