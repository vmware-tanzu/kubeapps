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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetClient(t *testing.T) {

	testCases := []struct {
		name         string
		clientGetter server.KubernetesClientGetter
		statusCode   codes.Code
	}{
		{
			name:         "returns internal error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.Internal,
		},
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
		{
			name: "returns client without error when configured correctly",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
				), nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			typedClient, dynamicClient, err := s.GetClients(context.Background())

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the clients should not be nil.
			if tc.statusCode == codes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}
			}
		})
	}

}

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter server.KubernetesClientGetter
		statusCode   codes.Code
	}{
		{
			name: "returns an internal error status if response does not contain packageRef.refName",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec("1.2.3", map[string]interface{}{
						"packageRef": map[string]interface{}{},
					}, t),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns an internal error status if response does not contain version",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(nil, map[string]interface{}{
						"packageRef": map[string]interface{}{
							"refName": "someName",
						},
					}, t),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns OK status if items contain required fields",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec("1.2.3", map[string]interface{}{
						"packageRef": map[string]interface{}{
							"refName": "someName",
						},
					}, t),
				), nil
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			_, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}
		})
	}

}

func packageFromSpec(version interface{}, spec map[string]interface{}, t *testing.T) *unstructured.Unstructured {
	pkgRef, ok := spec["packageRef"].(map[string]interface{})
	if !ok {
		t.Fatalf("unable to convert %+v to a map[string]interface{}", spec["packageRef"])
	}
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", packagingGroup, packageVersion),
			"kind":       packageResource,
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s.%s", pkgRef["refName"], version),
			},
			"spec": spec,
			"status": map[string]interface{}{
				"version": version,
			},
		},
	}
}

func packagesFromSpecs(specs []map[string]interface{}, t *testing.T) []runtime.Object {
	pkgs := []runtime.Object{}
	for _, s := range specs {
		pkgSpec, ok := s["spec"].(map[string]interface{})
		if !ok {
			t.Fatalf("unable to convert %+v to a map[string]interface{}", s["spec"])
		}
		pkgs = append(pkgs, packageFromSpec(s["version"], pkgSpec, t))
	}
	return pkgs
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name             string
		packageSpecs     []map[string]interface{}
		expectedPackages []*corev1.AvailablePackageSummary
	}{
		{
			name: "it returns carvel packages from the cluster",
			packageSpecs: []map[string]interface{}{
				{
					"spec": map[string]interface{}{
						"packageRef": map[string]interface{}{
							"refName": "tetris.foo.example.com",
						},
					},
					"version": "1.2.3",
				},
				{
					"spec": map[string]interface{}{
						"packageRef": map[string]interface{}{
							"refName": "another.foo.example.com",
						},
					},
					"version": "1.2.5",
				},
			},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:      "another.foo.example.com",
					LatestPkgVersion: "1.2.5",
				},
				{
					DisplayName:      "tetris.foo.example.com",
					LatestPkgVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkgs := packagesFromSpecs(tc.packageSpecs, t)
			s := Server{
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
						},
						pkgs...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{}, corev1.Context{})
			if got, want := response.AvailablePackagesSummaries, tc.expectedPackages; !cmp.Equal(got, want, opt1) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
			}
		})
	}

}

func repositoryFromSpec(name string, spec map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", packagingGroup, installPackageVersion),
			"kind":       repositoryResource,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": globalPackagingNamespace,
			},
			"spec": spec,
		},
	}
}

// spec is just a testing object spec, defined only to avoid writing
// a map of named specs as map[string]map[string]interface{}.
type spec map[string]interface{}

// repositoryFromSpecs takes a map of specs keyed by object name converting them to runtime objects.
func repositoriesFromSpecs(specs map[string]spec) []runtime.Object {
	repos := []runtime.Object{}
	for name, s := range specs {
		repos = append(repos, repositoryFromSpec(name, s))
	}
	return repos
}

func TestGetPackageRepositories(t *testing.T) {
	testCases := []struct {
		name                        string
		request                     *v1alpha1.GetPackageRepositoriesRequest
		repoSpecs                   map[string]spec
		expectedPackageRepositories []*v1alpha1.PackageRepository
		statusCode                  codes.Code
	}{
		{
			name:    "returns an internal error status if item in response cannot be converted to v1alpha1.PackageRepository",
			request: &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoSpecs: map[string]spec{
				"repo-1": {
					"fetch": "unexpected",
				},
			},
			statusCode: codes.Internal,
		},
		{
			name:    "returns expected repositories",
			request: &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoSpecs: map[string]spec{
				"repo-1": {
					"fetch": map[string]interface{}{
						"imgpkgBundle": map[string]interface{}{
							"image": "projects.registry.example.com/repo-1/main@sha256:abcd",
						},
					},
				},
				"repo-2": {
					"fetch": map[string]interface{}{
						"imgpkgBundle": map[string]interface{}{
							"image": "projects.registry.example.com/repo-2/main@sha256:abcd",
						},
					},
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{
				{
					Name:      "repo-1",
					Url:       "projects.registry.example.com/repo-1/main@sha256:abcd",
					Namespace: globalPackagingNamespace,
				},
				{
					Name:      "repo-2",
					Url:       "projects.registry.example.com/repo-2/main@sha256:abcd",
					Namespace: globalPackagingNamespace,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: packagingGroup, Version: installPackageVersion, Resource: repositoriesResource}: "PackageRepositoryList",
						},
						repositoriesFromSpecs(tc.repoSpecs)...,
					), nil
				},
			}

			response, err := s.GetPackageRepositories(context.Background(), &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(v1alpha1.PackageRepository{}, corev1.Context{})
					if got, want := response.Repositories, tc.expectedPackageRepositories; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}
		})
	}
}

func TestPackageRepositoryFromUnstructured(t *testing.T) {
	validSpec := map[string]interface{}{
		"fetch": map[string]interface{}{
			"imgpkgBundle": map[string]interface{}{
				"image": "projects.registry.example.com/repo-1/main@sha256:abcd",
			},
		},
	}
	invalidSpec := map[string]interface{}{
		"fetch": "unexpected",
	}
	testCases := []struct {
		name       string
		in         *unstructured.Unstructured
		expected   *v1alpha1.PackageRepository
		statusCode codes.Code
	}{
		{
			name:       "returns internal error if empty name",
			in:         repositoryFromSpec("", validSpec),
			statusCode: codes.Internal,
		},
		{
			name:       "returns internal error if spec is invalid",
			in:         repositoryFromSpec("valid-name", invalidSpec),
			statusCode: codes.Internal,
		},
		{
			name: "returns a repo for an imgpkgBundle type",
			in:   repositoryFromSpec("valid-name", validSpec),
			expected: &v1alpha1.PackageRepository{
				Name:      "valid-name",
				Url:       "projects.registry.example.com/repo-1/main@sha256:abcd",
				Namespace: globalPackagingNamespace,
			},
		},
		{
			name: "returns a repo for an image type",
			in: repositoryFromSpec("valid-name", map[string]interface{}{
				"fetch": map[string]interface{}{
					"image": map[string]interface{}{
						"url": "host.com/username/image:v0.1.0",
					},
				},
			}),
			expected: &v1alpha1.PackageRepository{
				Name:      "valid-name",
				Url:       "host.com/username/image:v0.1.0",
				Namespace: globalPackagingNamespace,
			},
		},
		{
			name: "returns a repo for an http type",
			in: repositoryFromSpec("valid-name", map[string]interface{}{
				"fetch": map[string]interface{}{
					"http": map[string]interface{}{
						"url": "https://host.com/archive.tgz",
					},
				},
			}),
			expected: &v1alpha1.PackageRepository{
				Name:      "valid-name",
				Url:       "https://host.com/archive.tgz",
				Namespace: globalPackagingNamespace,
			},
		},
		{
			name: "returns a repo for a git type",
			in: repositoryFromSpec("valid-name", map[string]interface{}{
				"fetch": map[string]interface{}{
					"git": map[string]interface{}{
						"url": "https://github.com/k14s/k8s-simple-app-example",
					},
				},
			}),
			expected: &v1alpha1.PackageRepository{
				Name:      "valid-name",
				Url:       "https://github.com/k14s/k8s-simple-app-example",
				Namespace: globalPackagingNamespace,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := packageRepositoryFromUnstructured(tc.in)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(v1alpha1.PackageRepository{}, corev1.Context{})
				if got, want := repo, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}
