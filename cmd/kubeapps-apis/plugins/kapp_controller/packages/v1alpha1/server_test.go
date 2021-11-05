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
	pluginv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
)

var ignoreUnexported = cmpopts.IgnoreUnexported(
	corev1.AvailablePackageReference{},
	corev1.AvailablePackageSummary{},
	corev1.Context{},
	corev1.PackageAppVersion{},
	pluginv1.Plugin{},
)

func TestGetClient(t *testing.T) {

	testCases := []struct {
		name         string
		clientGetter clientGetter
		statusCode   codes.Code
	}{
		{
			name:         "returns internal error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.Internal,
		},
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context, string) (dynamic.Interface, error) {
				return nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
		{
			name: "returns client without error when configured correctly",
			clientGetter: func(context.Context, string) (dynamic.Interface, error) {
				return dynfake.NewSimpleDynamicClient(runtime.NewScheme()), nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			dynamicClient, err := s.getDynamicClient(context.Background(), "default")

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the clients should not be nil.
			if tc.statusCode == codes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}
			}
		})
	}

}

func packagesFromSpecs(t *testing.T, namespace string, specs map[string]interface{}) []runtime.Object {
	pkgs := []runtime.Object{}
	for name, s := range specs {
		pkgs = append(pkgs, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/%s", packageGroup, packageVersion),
				"kind":       packageResource,
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": namespace,
				},
				"spec": s,
			},
		})
	}
	return pkgs
}

func metadatasFromSpecs(t *testing.T, namespace string, specs map[string]interface{}) []runtime.Object {
	metadatas := []runtime.Object{}
	for name, s := range specs {
		metadatas = append(metadatas, &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": fmt.Sprintf("%s/%s", packageGroup, packageVersion),
				"kind":       packageMetadataResource,
				"metadata": map[string]interface{}{
					"name":      name,
					"namespace": namespace,
				},
				"spec": s,
			},
		})
	}
	return metadatas

}

func TestGetAvailablePackageSummaries(t *testing.T) {
	defaultContext := &corev1.Context{Cluster: "default", Namespace: "default"}
	testCases := []struct {
		name               string
		existingObjects    []runtime.Object
		expectedPackages   []*corev1.AvailablePackageSummary
		expectedStatusCode codes.Code
	}{
		{
			name: "it returns an internal error status if a package meta does not contain spec.displayName",
			existingObjects: metadatasFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com": map[string]interface{}{
					"longDescription": "Some long text",
				},
			}),
			expectedStatusCode: codes.Internal,
		},
		{
			name: "it returns an internal error status if a package does not contain version",
			existingObjects: metadatasFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com": map[string]interface{}{
					"displayName": "Classic Tetris",
				},
			}),
			expectedStatusCode: codes.Internal,
		},
		{
			name: "it returns carvel package summaries with basic info from the cluster",
			existingObjects: append(metadatasFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com": map[string]interface{}{
					"displayName": "Classic Tetris",
				},
				"another.foo.example.com": map[string]interface{}{
					"displayName": "Some Other Game",
				},
			}), packagesFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com.1.2.3": map[string]interface{}{
					"refName": "tetris.foo.example.com",
					"version": "1.2.3",
				},
				"another.foo.example.com.1.2.5": map[string]interface{}{
					"refName": "another.foo.example.com",
					"version": "1.2.5",
				},
			})...),
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "another.foo.example.com",
					},
					Name:          "another.foo.example.com",
					DisplayName:   "Some Other Game",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.5"},
				},
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:          "tetris.foo.example.com",
					DisplayName:   "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
				},
			},
		},
		{
			name: "it returns carvel package summaries with complete metadata",
			existingObjects: append(metadatasFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com": map[string]interface{}{
					"displayName":      "Classic Tetris",
					"iconSVGBase64":    "Tm90IHJlYWxseSBTVkcK",
					"shortDescription": "A great game for arcade gamers",
					"categories":       []interface{}{"logging", "daemon-set"},
				},
			}), packagesFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com.1.2.3": map[string]interface{}{
					"refName": "tetris.foo.example.com",
					"version": "1.2.3",
				},
			})...),
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:             "tetris.foo.example.com",
					DisplayName:      "Classic Tetris",
					LatestVersion:    &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					Categories:       []string{"logging", "daemon-set"},
				},
			},
		},
		{
			name: "it returns the latest semver version in the latest version field",
			existingObjects: append(metadatasFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com": map[string]interface{}{
					"displayName": "Classic Tetris",
				},
			}), packagesFromSpecs(t, defaultContext.Namespace, map[string]interface{}{
				"tetris.foo.example.com.1.2.3": map[string]interface{}{
					"refName": "tetris.foo.example.com",
					"version": "1.2.3",
				},
				"tetris.foo.example.com.1.2.7": map[string]interface{}{
					"refName": "tetris.foo.example.com",
					"version": "1.2.7",
				},
				"tetris.foo.example.com.1.2.4": map[string]interface{}{
					"refName": "tetris.foo.example.com",
					"version": "1.2.4",
				},
			})...),
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:          "tetris.foo.example.com",
					DisplayName:   "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.7"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{
				clientGetter: func(context.Context, string) (dynamic.Interface, error) {
					return dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: packageGroup, Version: packageVersion, Resource: packageResources}:         "PackageList",
							{Group: packageGroup, Version: packageVersion, Resource: packageMetadataResources}: "PackageMetadataList",
						},
						tc.existingObjects...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: defaultContext})

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			if got, want := response.AvailablePackageSummaries, tc.expectedPackages; !cmp.Equal(got, want, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}

}

func repositoryFromSpec(name string, spec map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", repositoryGroup, repositoryVersion),
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
				clientGetter: func(context.Context, string) (dynamic.Interface, error) {
					return dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: repositoryGroup, Version: repositoryVersion, Resource: repositoriesResource}: "PackageRepositoryList",
						},
						repositoriesFromSpecs(tc.repoSpecs)...,
					), nil
				},
			}

			response, err := s.GetPackageRepositories(context.Background(), &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v, response: %+v", got, want, err, response)
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
