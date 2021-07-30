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
	"fmt"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// see docs at https://fluxcd.io/docs/components/helm/api/
	fluxHelmReleaseGroup   = "helm.toolkit.fluxcd.io"
	fluxHelmReleaseVersion = "v2beta1"
	fluxHelmRelease        = "HelmRelease"
	fluxHelmReleases       = "helmreleases"
	fluxHelmReleaseList    = "HelmReleaseList"
)

// namespace maybe "", in which case repositories from all namespaces are returned
func (s *Server) listReleasesInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return nil, err
	}

	releasesResource := schema.GroupVersionResource{
		Group:    fluxHelmReleaseGroup,
		Version:  fluxHelmReleaseVersion,
		Resource: fluxHelmReleases,
	}

	if releases, err := client.Resource(releasesResource).Namespace(namespace).List(ctx, metav1.ListOptions{}); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list fluxv2 helmreleases: %v", err)
	} else {
		return releases, nil
	}
}

func installedPkgSummaryFromRelease(unstructuredRelease map[string]interface{}, chartList *unstructured.UnstructuredList) (*corev1.InstalledPackageSummary, error) {
	// first check if release CR is ready or is in "flux"
	observedGeneration, found, err := unstructured.NestedInt64(unstructuredRelease, "status", "observedGeneration")
	if err != nil || !found {
		return nil, nil // not ready
	}
	generation, found, err := unstructured.NestedInt64(unstructuredRelease, "metadata", "generation")
	if err != nil || !found {
		return nil, nil
	}
	if generation != observedGeneration {
		return nil, nil
	}

	// see https://fluxcd.io/docs/components/helm/helmreleases/
	name, found, err := unstructured.NestedString(unstructuredRelease, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field metadata.name not found on HelmRelease:\n%s, error: %v", prettyPrintMap(unstructuredRelease), err)
	}

	namespace, found, err := unstructured.NestedString(unstructuredRelease, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"field metadata.namespace not found on HelmRelease:\n%s, error: %v", prettyPrintMap(unstructuredRelease), err)
	}

	var pkgVersion *corev1.VersionReference
	version, found, err := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "version")
	if found && err == nil && version != "" {
		pkgVersion = &corev1.VersionReference{
			Version: version,
		}
	}

	// this will only be present if install/upgrade succeeded
	lastAppliedRevision, _, _ := unstructured.NestedString(unstructuredRelease, "status", "lastAppliedRevision")

	repoName, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "name")
	repoNamespace, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "sourceRef", "namespace")
	chartName, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "chart")
	chartVersion, _, _ := unstructured.NestedString(unstructuredRelease, "spec", "chart", "spec", "version")

	var pkgDetail *corev1.AvailablePackageDetail
	if repoName != "" && repoNamespace != "" && chartName != "" && chartVersion != "" {
		url, _ := findUrlForChartInList(chartList, repoName, chartName, chartVersion)
		if url != "" {
			chartID := fmt.Sprintf("%s/%s", repoName, chartName)
			pkgDetail, _ = availablePackageDetailFromTarball(chartID, url)
		}
	}

	var status *corev1.InstalledPackageStatus
	if conditions, found, err := unstructured.NestedSlice(unstructuredRelease, "status", "conditions"); found && err == nil {
		for _, conditionUnstructured := range conditions {
			if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
				if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
					status = &corev1.InstalledPackageStatus{
						Ready: false,
					}
					if statusString, ok := conditionAsMap["status"]; ok {
						if statusString == "True" {
							status.Ready = true
							status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED
						} else if statusString == "False" {
							status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_FAILED
						} else {
							status.Reason = corev1.InstalledPackageStatus_STATUS_REASON_PENDING
						}
						if reasonString, ok := conditionAsMap["reason"].(string); ok {
							status.UserReason = reasonString
						}
					}
				}
			}
		}
	}

	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: namespace,
			},
			Identifier: name,
		},
		Name:                name,
		PkgVersionReference: pkgVersion,
		CurrentPkgVersion:   lastAppliedRevision,
		CurrentAppVersion:   pkgDetail.GetAppVersion(),
		IconUrl:             pkgDetail.GetIconUrl(),
		PkgDisplayName:      pkgDetail.GetDisplayName(),
		ShortDescription:    pkgDetail.GetShortDescription(),
		Status:              status,
		// LatestMatchingPkgVersion
		// Only non-empty if an available upgrade matches the specified pkg_version_reference.
		// For example, if the pkg_version_reference is ">10.3.0 < 10.4.0" and 10.3.1
		// is installed, then:
		//   * if 10.3.2 is available, latest_matching_version should be 10.3.2, but
		//   * if 10.4 is available while >10.3.1 is not, this should remain empty.

		// LatestPkgVersion
		// The latest version available for this package, regardless of the pkg_version_reference.
	}, nil
}
