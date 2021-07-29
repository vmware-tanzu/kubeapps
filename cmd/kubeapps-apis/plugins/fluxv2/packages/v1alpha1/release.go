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

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func installedPkgSummaryFromRelease(unstructuredRelease map[string]interface{}) (*corev1.InstalledPackageSummary, error) {
	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: r.Namespace,
			},
			Identifier: r.Name,
		},
		Name: r.Name,
		PkgVersionReference: &corev1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentPkgVersion: r.Chart.Metadata.Version,
		IconUrl:           r.Chart.Metadata.Icon,
		PkgDisplayName:    r.Chart.Name(),
		ShortDescription:  r.Chart.Metadata.Description,
	}, nil
}
