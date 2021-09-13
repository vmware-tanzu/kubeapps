/*
Copyright © 2021 VMware
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	// see docs at https://fluxcd.io/docs/components/helm/api/
	fluxHelmReleaseGroup   = "helm.toolkit.fluxcd.io"
	fluxHelmReleaseVersion = "v2beta1"
	fluxHelmRelease        = "HelmRelease"
	fluxHelmReleases       = "helmreleases"
	fluxHelmReleaseList    = "HelmReleaseList"
)

func (s *Server) getReleasesResourceInterface(ctx context.Context, namespace string) (dynamic.ResourceInterface, error) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return nil, err
	}

	releasesResource := schema.GroupVersionResource{
		Group:    fluxHelmReleaseGroup,
		Version:  fluxHelmReleaseVersion,
		Resource: fluxHelmReleases,
	}

	return client.Resource(releasesResource).Namespace(namespace), nil
}

// namespace maybe "", in which case releases from all namespaces are returned
func (s *Server) listReleasesInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	releasesIfc, err := s.getReleasesResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if releases, err := releasesIfc.List(ctx, metav1.ListOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list fluxv2 helmreleases: %v", err)
	} else {
		return releases, nil
	}
}

func (s *Server) getReleaseInCluster(ctx context.Context, name types.NamespacedName) (*unstructured.Unstructured, error) {
	releasesIfc, err := s.getReleasesResourceInterface(ctx, name.Namespace)
	if err != nil {
		return nil, err
	}
	return releasesIfc.Get(ctx, name.Name, metav1.GetOptions{})
}

func (s *Server) paginatedInstalledPkgSummaries(ctx context.Context, namespace string, pageSize int32, pageOffset int) ([]*corev1.InstalledPackageSummary, error) {
	releasesFromCluster, err := s.listReleasesInCluster(ctx, namespace)
	if err != nil {
		return nil, err
	}

	installedPkgSummaries := []*corev1.InstalledPackageSummary{}
	if len(releasesFromCluster.Items) > 0 {
		// we're going to need this later
		// TODO (gfichtenholt) for now we get all charts and later find one that helmrelease is using
		// there is probably a more efficient way to do this
		chartsFromCluster, err := s.listChartsInCluster(ctx, apiv1.NamespaceAll)
		if err != nil {
			return nil, err
		}

		startAt := -1
		if pageSize > 0 {
			startAt = int(pageSize) * pageOffset
		}

		for i, releaseUnstructured := range releasesFromCluster.Items {
			if startAt <= i {
				summary, err := s.installedPkgSummaryFromRelease(releaseUnstructured.Object, chartsFromCluster)
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

func (s *Server) installedPkgSummaryFromRelease(unstructuredRelease map[string]interface{}, chartsFromCluster *unstructured.UnstructuredList) (*corev1.InstalledPackageSummary, error) {
	// first check if release CR is ready or is in "flux"
	if !checkGeneration(unstructuredRelease) {
		return nil, nil
	}

	name, err := namespacedName(unstructuredRelease)
	if err != nil {
		return nil, err
	}

	var pkgVersion *corev1.VersionReference
	version, found, err := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersion = &corev1.VersionReference{
			Version: version,
		}
	}

	// see https://fluxcd.io/docs/components/helm/helmreleases/
	repoName, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "name")
	repoNamespace, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "namespace")
	chartName, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "chart")
	chartVersion, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "version")

	var latestPkgVersion *corev1.PackageAppVersion
	var pkgDetail *corev1.AvailablePackageDetail
	// according to docs chartVersion is optional (defaults to '*' i.e. latest when omitted)
	if repoName != "" && chartName != "" && chartVersion != "" {
		// chartName here refers to a chart template, e.g. "nginx", rather than a specific chart instance
		// e.g. "default-my-nginx". The spec somewhat vaguely states "The name of the chart as made available
		// by the HelmRepository (without any aliases), for example: podinfo". So, we can't exactly do a "get"
		// on the name, but have to iterate the complete list of available charts for a match
		tarUrl, err := findUrlForChartInList(chartsFromCluster, repoName, chartName, chartVersion)
		if err != nil {
			return nil, err
		} else if tarUrl == "" {
			return nil, status.Errorf(codes.Internal, "Failed to find find tar file url for chart [%s], version: [%s]", chartName, chartVersion)
		}
		chartID := fmt.Sprintf("%s/%s", repoName, chartName)
		if pkgDetail, err = availablePackageDetailFromTarball(chartID, tarUrl); err != nil {
			return nil, err
		}

		// according to docs repoNamespace is optional
		if repoNamespace == "" {
			repoNamespace = name.Namespace
		}
		repo := types.NamespacedName{Namespace: repoNamespace, Name: repoName}
		chartFromCache, err := s.fetchChartFromCache(repo, chartName)
		if err != nil {
			return nil, err
		} else if chartFromCache != nil && len(chartFromCache.ChartVersions) > 0 {
			// charts in cache are already sorted with the latest being at position 0
			latestPkgVersion = &corev1.PackageAppVersion{
				PkgVersion: chartFromCache.ChartVersions[0].Version,
				AppVersion: chartFromCache.ChartVersions[0].AppVersion,
			}
		}
	}

	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: name.Namespace,
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

func (s *Server) installedPackageDetail(ctx context.Context, name types.NamespacedName) (*corev1.InstalledPackageDetail, error) {
	unstructuredRelease, err := s.getReleaseInCluster(ctx, name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q due to: %v", name, err)
	}

	log.V(4).Infof("installedPackageDetail:\n[%s]", prettyPrintMap(unstructuredRelease.Object))

	obj := unstructuredRelease.Object
	var pkgVersionRef *corev1.VersionReference
	version, found, err := unstructured.NestedString(obj, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersionRef = &corev1.VersionReference{
			Version: version,
		}
	}

	valuesApplied := ""
	valuesMap, found, err := unstructured.NestedMap(obj, "spec", "values")
	if found && err == nil && len(valuesMap) != 0 {
		bytes, err := json.Marshal(valuesMap)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to marshal Helm release values due to: %v", err)
		}
		valuesApplied = string(bytes)
	}
	// TODO (gfichtenholt) what about ValuesFrom []ValuesReference `json:"valuesFrom,omitempty"`?
	// ValuesReference maybe a config map or a secret

	// this will only be present if install/upgrade succeeded
	pkgVersion, found, err := unstructured.NestedString(obj, "status", "lastAppliedRevision")
	if !found || err != nil || pkgVersion == "" {
		// this is the back-up option: will be there if the reconciliation is in progress or has failed
		pkgVersion, _, _ = unstructured.NestedString(obj, "status", "lastAttemptedRevision")
	}

	availablePackageRef, err := installedPackageAvailablePackageRefFromUnstructured(obj)
	if err != nil {
		return nil, err
	}

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
	} else if err != nil && !errors.IsNotFound(err) {
		log.Warningf("Failed to get helm release due to %v", err)
	}

	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: name.Namespace,
			},
			Identifier: name.Name,
			Plugin:     GetPluginDetail(),
		},
		Name:                name.Name,
		PkgVersionReference: pkgVersionRef,
		CurrentVersion: &corev1.PackageAppVersion{
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

func (s *Server) helmReleaseFromUnstructured(ctx context.Context, name types.NamespacedName, unstructuredRelease map[string]interface{}) (*release.Release, error) {
	// post installation notes can only be retrieved via helm APIs, flux doesn't do it
	// see discussion in https://cloud-native.slack.com/archives/CLAJ40HV3/p1629244025187100
	if s.actionConfigGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Server is not configured with actionConfigGetter")
	}

	helmReleaseName, found, err := unstructured.NestedString(unstructuredRelease, "spec", "ReleaseName")
	// according to docs ReleaseName is optional and defaults to a composition of
	// '[TargetNamespace-]Name'.
	if !found || err != nil || helmReleaseName == "" {
		targetNamespace, found, err := unstructured.NestedString(unstructuredRelease, "spec", "targetNamespace")
		// according to docs targetNamespace is optional and defaults to the namespace of the HelmRelease
		if !found || err != nil || targetNamespace == "" {
			targetNamespace = name.Namespace
		}
		helmReleaseName = fmt.Sprintf("%s-%s", targetNamespace, name.Name)
	}

	actionConfig, err := s.actionConfigGetter(ctx, name.Namespace)
	if err != nil || actionConfig == nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config in namespace [%s] due to: %v", name.Namespace, err)
	}
	cmd := action.NewGet(actionConfig)
	release, err := cmd.Run(helmReleaseName)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release [%s] in namespace [%s]", helmReleaseName, name.Namespace)
		}
		return nil, status.Errorf(codes.NotFound, "Unable to run Helm Get action for release [%s] in namespace [%s]: %v", helmReleaseName, name.Namespace, err)
	}
	return release, nil
}

func (s *Server) newRelease(ctx context.Context, packageRef *corev1.AvailablePackageReference, targetName types.NamespacedName, versionRef *corev1.VersionReference, reconcile *corev1.ReconciliationOptions, valuesString string) (*corev1.InstalledPackageReference, error) {
	// HACK: just for now assume HelmRelease CRD will live in the kubeapps namespace
	kubeappsNamespace := os.Getenv("POD_NAMESPACE")
	resourceIfc, err := s.getReleasesResourceInterface(ctx, kubeappsNamespace)
	if err != nil {
		return nil, err
	}

	availablePackageNamespace := packageRef.GetContext().GetNamespace()
	if availablePackageNamespace == "" || packageRef.GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}

	unescapedChartID, err := getUnescapedChartID(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	packageIdParts := strings.Split(unescapedChartID, "/")
	repo := types.NamespacedName{Namespace: availablePackageNamespace, Name: packageIdParts[0]}
	chart, err := s.fetchChartFromCache(repo, packageIdParts[1])
	if err != nil {
		return nil, err
	}

	var values map[string]interface{}
	if valuesString != "" {
		values = make(map[string]interface{})
		err = yaml.Unmarshal([]byte(valuesString), &values)
		if err != nil {
			return nil, err
		}
	}

	fluxHelmRelease := newFluxHelmRelease(chart, kubeappsNamespace, targetName, versionRef, reconcile, values)
	newRelease, err := resourceIfc.Create(ctx, fluxHelmRelease, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	name, err := namespacedName(newRelease.Object)
	if err != nil {
		return nil, err
	}
	return &corev1.InstalledPackageReference{
		Context:    &corev1.Context{Namespace: name.Namespace},
		Identifier: name.Name,
		Plugin:     GetPluginDetail(),
	}, nil
}

// returns 3 things:
// - ready:  whether the HelmRelease object is in a ready state
// - reason: one of SUCCESS/FAILURE/PENDING/UNKNOWN, if present
// - userReason: textual description of why the object is in current state, if present
// docs:
// 1. https://fluxcd.io/docs/components/helm/helmreleases/#examples
// 2. discussion on https://vmware.slack.com/archives/C4HEXCX3N/p1630907107078800.
func isHelmReleaseReady(unstructuredObj map[string]interface{}) (ready bool, status corev1.InstalledPackageStatus_StatusReason, userReason string) {
	if !checkGeneration(unstructuredObj) {
		return false, corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED, ""
	}

	conditions, found, err := unstructured.NestedSlice(unstructuredObj, "status", "conditions")
	if err != nil || !found {
		return false, corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED, ""
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
						return true, corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED, userReason
					} else if isInstallFailed {
						return false, corev1.InstalledPackageStatus_STATUS_REASON_FAILED, userReason
					} else {
						return false, corev1.InstalledPackageStatus_STATUS_REASON_PENDING, userReason
					}
				}
			}
		}
	}
	// catch all: unless we know something else, install is pending
	return false, corev1.InstalledPackageStatus_STATUS_REASON_PENDING, userReason
}

func installedPackageStatusFromUnstructured(unstructuredRelease map[string]interface{}) *corev1.InstalledPackageStatus {
	ready, reason, userReason := isHelmReleaseReady(unstructuredRelease)
	return &corev1.InstalledPackageStatus{
		Ready:      ready,
		Reason:     reason,
		UserReason: userReason,
	}
}

func installedPackageReconciliationOptionsFromUnstructured(unstructuredRelease map[string]interface{}) *corev1.ReconciliationOptions {
	reconciliationOptions := &corev1.ReconciliationOptions{}
	if intervalString, found, err := unstructured.NestedString(unstructuredRelease, "spec", "interval"); found && err == nil {
		if duration, err := time.ParseDuration(intervalString); err == nil {
			reconciliationOptions.Interval = int32(duration.Seconds())
		}
	}
	if suspend, found, err := unstructured.NestedBool(unstructuredRelease, "spec", "suspend"); found && err == nil {
		reconciliationOptions.Suspend = suspend
	}
	if serviceAccountName, found, err := unstructured.NestedString(unstructuredRelease, "spec", "serviceAccountName"); found && err == nil {
		reconciliationOptions.ServiceAccountName = serviceAccountName
	}
	return reconciliationOptions
}

func installedPackageAvailablePackageRefFromUnstructured(unstructuredRelease map[string]interface{}) (*corev1.AvailablePackageReference, error) {
	repoName, found, err := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "name")
	if !found || err != nil {
		return nil, status.Errorf(codes.Internal, "missing required field spec.chart.spec.sourceRef.name")
	}
	chartName, found, err := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "chart")
	if !found || err != nil {
		return nil, status.Errorf(codes.Internal, "missing required field spec.chart.spec.chart")
	}
	repoNamespace, found, err := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "namespace")
	// CrossNamespaceObjectReference namespace is optional, so
	if !found || err != nil || repoNamespace == "" {
		name, err := namespacedName(unstructuredRelease)
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

// Potentially, there are 3 different namespaces that can be specified here
// 1. spec.chart.spec.sourceRef.namespace, where HelmRepository CRD object referenced exists
// 2. metadata.namespace, where this HelmRelease CRD will exist
// 3. spec.targetNamespace, where flux will install any artifacts from the release
func newFluxHelmRelease(chart *models.Chart, releaseNamespace string, targetName types.NamespacedName, versionRef *corev1.VersionReference, reconcile *corev1.ReconciliationOptions, values map[string]interface{}) *unstructured.Unstructured {
	unstructuredRel := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxHelmReleaseGroup, fluxHelmReleaseVersion),
			"kind":       fluxHelmRelease,
			"metadata": map[string]interface{}{
				"name":      targetName.Name,
				"namespace": releaseNamespace,
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
				"install": map[string]interface{}{
					"createNamespace": true,
				},
				"targetNamespace": targetName.Namespace,
			},
		},
	}
	if versionRef != nil && versionRef.Version != "" {
		unstructured.SetNestedField(unstructuredRel.Object, versionRef.Version, "spec", "chart", "spec", "version")
	}
	reconcileInterval := "1m" // unless explictly specified
	if reconcile != nil {
		if reconcile.Interval > 0 {
			reconcileInterval = (time.Duration(reconcile.Interval) * time.Second).String()
		}
		unstructured.SetNestedField(unstructuredRel.Object, reconcile.Suspend, "spec", "suspend")
		if reconcile.ServiceAccountName != "" {
			unstructured.SetNestedField(unstructuredRel.Object, reconcile.ServiceAccountName, "spec", "serviceAccountName")
		}
	}
	if values != nil {
		unstructured.SetNestedMap(unstructuredRel.Object, values, "spec", "values")
	}

	// required fields, without which flux controller will fail to create the CRD
	unstructured.SetNestedField(unstructuredRel.Object, reconcileInterval, "spec", "interval")
	return &unstructuredRel
}
