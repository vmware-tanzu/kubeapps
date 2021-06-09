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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func TestGetClient(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter func(context.Context) (dynamic.Interface, error)
		statusCode   codes.Code
	}{
		{
			name:         "returns internal error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.Internal,
		},
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
		{
			name: "returns client without error when configured correctly",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
				), nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			client, err := s.GetClient(context.Background())

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCode == codes.OK {
				if _, ok := client.(dynamic.Interface); !ok {
					t.Errorf("got: %T, want: dynamic.Interface", client)
				}
			}
		})
	}

}

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter func(context.Context) (dynamic.Interface, error)
		statusCode   codes.Code
	}{
		{
			name: "returns an internal error status if response does not contain publicName",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"version": "1.2.3",
					}),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns an internal error status if response does not contain version",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"publicName": "someName",
					}),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns OK status if items contain required fields",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"publicName": "someName",
						"version":    "1.2.3",
					}),
				), nil
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			_, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}
		})
	}

}

func packageFromSpec(spec map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", packageGroup, packageVersion),
			"kind":       "Package",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s.%s", spec["publicName"], spec["version"]),
			},
			"spec": spec,
		},
	}
}

func packagesFromSpecs(specs []map[string]interface{}) []runtime.Object {
	pkgs := []runtime.Object{}
	for _, s := range specs {
		pkgs = append(pkgs, packageFromSpec(s))
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
					"publicName": "tetris.foo.example.com",
					"version":    "1.2.3",
				},
				{
					"publicName": "another.foo.example.com",
					"version":    "1.2.5",
				},
			},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:   "another.foo.example.com",
					LatestVersion: "1.2.5",
				},
				{
					DisplayName:   "tetris.foo.example.com",
					LatestVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkgs := packagesFromSpecs(tc.packageSpecs)
			s := Server{
				clientGetter: func(context.Context) (dynamic.Interface, error) {
					return fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
						},
						pkgs...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response.AvailablePackagesSummaries, tc.expectedPackages; !cmp.Equal(got, want, cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{})) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{})))
			}
		})
	}

}

func repositoryFromSpec(name string, spec map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", installPackageGroup, installPackageVersion),
			"kind":       "PackageRepository",
			"metadata": map[string]interface{}{
				"name": name,
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
		request                     *corev1.GetPackageRepositoriesRequest
		repoSpecs                   map[string]spec
		expectedPackageRepositories []*corev1.PackageRepository
		statusCode                  codes.Code
	}{
		{
			name:    "returns an internal error status if item in response cannot be converted to corev1.PackageRepository",
			request: &corev1.GetPackageRepositoriesRequest{},
			repoSpecs: map[string]spec{
				"repo-1": {
					"fetch": "unexpected",
				},
			},
			statusCode: codes.Internal,
		},
		{
			name:    "returns expected repositories",
			request: &corev1.GetPackageRepositoriesRequest{},
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
			expectedPackageRepositories: []*corev1.PackageRepository{
				{
					Name: "repo-1",
					Url:  "projects.registry.example.com/repo-1/main@sha256:abcd",
				},
				{
					Name: "repo-2",
					Url:  "projects.registry.example.com/repo-2/main@sha256:abcd",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{
				clientGetter: func(context.Context) (dynamic.Interface, error) {
					return fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: installPackageGroup, Version: installPackageVersion, Resource: repositoriesResource}: "PackageRepositoryList",
						},
						repositoriesFromSpecs(tc.repoSpecs)...,
					), nil
				},
			}

			response, err := s.GetPackageRepositories(context.Background(), &corev1.GetPackageRepositoriesRequest{})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					if got, want := response.Repositories, tc.expectedPackageRepositories; !cmp.Equal(got, want, cmpopts.IgnoreUnexported(corev1.PackageRepository{})) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpopts.IgnoreUnexported(corev1.PackageRepository{})))
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
		expected   *corev1.PackageRepository
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
			expected: &corev1.PackageRepository{
				Name: "valid-name",
				Url:  "projects.registry.example.com/repo-1/main@sha256:abcd",
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
			expected: &corev1.PackageRepository{
				Name: "valid-name",
				Url:  "host.com/username/image:v0.1.0",
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
			expected: &corev1.PackageRepository{
				Name: "valid-name",
				Url:  "https://host.com/archive.tgz",
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
			expected: &corev1.PackageRepository{
				Name: "valid-name",
				Url:  "https://github.com/k14s/k8s-simple-app-example",
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
				if got, want := repo, tc.expected; !cmp.Equal(got, want, cmpopts.IgnoreUnexported(corev1.PackageRepository{})) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpopts.IgnoreUnexported(corev1.PackageRepository{})))
				}
			}
		})
	}
}
