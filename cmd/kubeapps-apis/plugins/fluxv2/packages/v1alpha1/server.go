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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"k8s.io/client-go/rest"
)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *corev1.GetPackageRepositoriesRequest) (*corev1.GetPackageRepositoriesResponse, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create incluster config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	repositoryResource := schema.GroupVersionResource{Group: "source.toolkit.fluxcd.io", Version: "v1beta1", Resource: "helmrepositories"}

	// Currently checks globally. Update to handle namespaced requests (?)
	repos, err := client.Resource(repositoryResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list helm-fluxv2 repositories: %w", err)
	}

	responseRepos := []*corev1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo := &corev1.PackageRepository{}
		name, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "name")
		if err != nil || !found {
			return nil, fmt.Errorf("required field metadata.name not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Name = name

		namespace, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "namespace")
		// TODO(absoludity): When testing, write failing test for the case of a
		// cluster-scoped object without a namespace, then fix.
		if err != nil || !found {
			return nil, fmt.Errorf("required field metadata.namespace not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Namespace = namespace

		url, found, err := unstructured.NestedString(repoUnstructured.Object, "spec", "url")
		if err != nil || !found {
			return nil, fmt.Errorf("required field spec.url not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Url = url

		responseRepos = append(responseRepos, repo)
	}
	return &corev1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}
