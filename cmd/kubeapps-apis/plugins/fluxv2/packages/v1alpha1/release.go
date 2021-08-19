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
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
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

	latestPkgVersion := ""
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
			latestPkgVersion = chartFromCache.ChartVersions[0].Version
		}
	}

	// this will only be present if install/upgrade succeeded
	lastAppliedRevision, _, _ := unstructured.NestedString(unstructuredRelease, "status", "lastAppliedRevision")

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
		CurrentPkgVersion:   lastAppliedRevision,
		CurrentAppVersion:   pkgDetail.GetAppVersion(),
		IconUrl:             pkgDetail.GetIconUrl(),
		PkgDisplayName:      pkgDetail.GetDisplayName(),
		ShortDescription:    pkgDetail.GetShortDescription(),
		Status:              installedPackageStatusFromUnstructured(unstructuredRelease),
		LatestPkgVersion:    latestPkgVersion,
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

	var pkgVersion *corev1.VersionReference
	version, found, err := unstructured.NestedString(unstructuredRelease.Object, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersion = &corev1.VersionReference{
			Version: version,
		}
	}

	valuesApplied := ""
	valuesMap, found, err := unstructured.NestedMap(unstructuredRelease.Object, "spec", "values")
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
	lastAppliedRevision, _, _ := unstructured.NestedString(unstructuredRelease.Object, "status", "lastAppliedRevision")

	availablePackageRef, err := installedPackageAvailablePackageRefFromUnstructured(unstructuredRelease.Object)
	if err != nil {
		return nil, err
	}

	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: name.Namespace,
			},
			Identifier: name.Name,
			Plugin:     GetPluginDetail(),
		},
		Name:                  name.Name,
		PkgVersionReference:   pkgVersion,
		CurrentPkgVersion:     lastAppliedRevision,
		ValuesApplied:         valuesApplied,
		ReconciliationOptions: installedPackageReconciliationOptionsFromUnstructured(unstructuredRelease.Object),
		AvailablePackageRef:   availablePackageRef,
		// TODO (gfichtenholt) PostInstallationNotes
		// looks like I may have to resort to getting the notes via helm APIs, flux doesn't do it
		// see discussion in https://cloud-native.slack.com/archives/CLAJ40HV3/p1629244025187100
		// bottom line: "helm get notes" is a read-only activity that you can replicate in your
		// own code by copying the implementation, or calling to the library function if there is one
		// https://github.com/helm/helm/blob/main/cmd/helm/get_notes.go.
		// Also look at https://gist.github.com/DzeryCZ/c4adf39d4a1a99ae6e594a183628eaee
		Status: installedPackageStatusFromUnstructured(unstructuredRelease.Object),
	}, nil
}

func installedPackageStatusFromUnstructured(unstructuredRelease map[string]interface{}) *corev1.InstalledPackageStatus {
	complete, success, reason := checkStatusReady(unstructuredRelease)
	status := &corev1.InstalledPackageStatus{
		Ready:      complete && success,
		UserReason: reason,
	}
	if complete && success {
		status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED
	} else if complete && !success {
		status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_FAILED
	} else {
		status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_PENDING
	}
	return status
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
