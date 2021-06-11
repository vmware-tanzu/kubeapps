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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func TestGetAvailablePackagesStatus(t *testing.T) {
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
			name: "returns without error if response status does not contain conditions",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", "", nil, nil),
				), nil
			},
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response status does not contain conditions",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", "", map[string]interface{}{
						"foo": "bar",
					}, map[string]interface{}{
						"zot": "xyz",
					}),
				), nil
			},
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain ready repos",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", "", map[string]interface{}{}, map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
								"reason": "IndexationFailed",
							},
						}}),
				), nil
			},
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain status url",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", "", nil, map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
								"reason": "IndexationSucceed",
							},
						}}),
				), nil
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)

				if got == codes.OK {
					if len(response.AvailablePackagesSummaries) != 0 {
						t.Errorf("unexpected response: %v", response)
					} else if response != nil {
						t.Errorf("unexpected response: %v", response)
					}
				}
			}
		})
	}
}

func newRepo(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name": name,
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}

	obj := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
		"kind":       fluxHelmRepository,
		"metadata":   metadata,
	}

	if spec != nil {
		obj["spec"] = spec
	}

	if status != nil {
		obj["status"] = status
	}

	return &unstructured.Unstructured{
		Object: obj,
	}
}

// repositoryFromSpecs takes a map of specs keyed by object name converting them to runtime objects.
func newRepos(specs map[string]map[string]interface{}, namespace string) []runtime.Object {
	repos := []runtime.Object{}
	for name, spec := range specs {
		repo := newRepo(name, namespace, spec, nil)
		repos = append(repos, repo)
	}
	return repos
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		testName         string
		request          *corev1.GetAvailablePackageSummariesRequest
		repoName         string
		repoNamespace    string
		repoUrl          string
		repoIndex        string
		expectedPackages []*corev1.AvailablePackageSummary
	}{
		{
			testName:      "it returns a couple of fluxv2 packages from the cluster",
			repoName:      "bitnami-1",
			repoNamespace: "",
			request:       &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			repoUrl:       "https://example.repo.com/charts",
			repoIndex:     "testdata/valid-index.yaml",
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:   "acs-engine-autoscaler",
					LatestVersion: "2.1.1",
					IconUrl:       "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1",
					},
				},
				{
					DisplayName:   "wordpress",
					LatestVersion: "0.7.5",
					IconUrl:       "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1",
					},
				},
			},
		},
		{
			testName:      "it returns all of fluxv2 packages from the cluster when request does not specify ns",
			repoName:      "bitnami-2",
			repoNamespace: "non-default",
			request:       &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			repoUrl:       "https://example.repo.com/charts",
			repoIndex:     "testdata/valid-index.yaml",
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:   "acs-engine-autoscaler",
					LatestVersion: "2.1.1",
					IconUrl:       "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-2",
						Context: &corev1.Context{
							Namespace: "non-default",
						},
					},
				},
				{
					DisplayName:   "wordpress",
					LatestVersion: "0.7.5",
					IconUrl:       "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-2",
						Context: &corev1.Context{
							Namespace: "non-default",
						},
					},
				},
			},
		},
		{
			testName:      "it returns all of fluxv2 packages from the cluster that match request ns",
			repoName:      "bitnami-3",
			repoNamespace: "non-default",
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "non-default",
				},
			},
			repoUrl:   "https://example.repo.com/charts",
			repoIndex: "testdata/valid-index.yaml",
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:   "acs-engine-autoscaler",
					LatestVersion: "2.1.1",
					IconUrl:       "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-3",
						Context: &corev1.Context{
							Namespace: "non-default",
						},
					},
				},
				{
					DisplayName:   "wordpress",
					LatestVersion: "0.7.5",
					IconUrl:       "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-3",
						Context: &corev1.Context{
							Namespace: "non-default",
						},
					},
				},
			},
		},
		{
			testName:      "it returns none of fluxv2 packages from the cluster that don't match request ns",
			repoName:      "bitnami-4",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "non-default",
				},
			},
			repoUrl:          "https://example.repo.com/charts",
			repoIndex:        "testdata/valid-index.yaml",
			expectedPackages: []*corev1.AvailablePackageSummary{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			indexYAMLBytes, err := ioutil.ReadFile(tc.repoIndex)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, string(indexYAMLBytes))
			}))
			defer ts.Close()

			repoSpec := map[string]interface{}{
				"url":      tc.repoUrl,
				"interval": "1m0s",
			}
			repoStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "IndexationSucceed",
					},
				},
				"url": ts.URL,
			}
			repo := newRepo(tc.repoName, tc.repoNamespace, repoSpec, repoStatus)
			s := Server{
				clientGetter: func(context.Context) (dynamic.Interface, error) {
					return fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
						},
						repo,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{})
			opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
			if got, want := response.AvailablePackagesSummaries, tc.expectedPackages; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
		})
	}
}

func TestGetPackageRepositories(t *testing.T) {
	testCases := []struct {
		name                        string
		request                     *v1alpha1.GetPackageRepositoriesRequest
		repoNamespace               string
		repoSpecs                   map[string]map[string]interface{}
		expectedPackageRepositories []*v1alpha1.PackageRepository
		statusCode                  codes.Code
	}{
		{
			name:          "returns an internal error status if item in response cannot be converted to v1alpha1.PackageRepository",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"foo": "bar",
				},
			},
			statusCode: codes.Internal,
		},
		{
			name:          "returns expected repositories",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{
				{
					Name: "repo-1",
					Url:  "https://charts.bitnami.com/bitnami",
				},
				{
					Name: "repo-2",
					Url:  "https://charts.helm.sh/stable",
				},
			},
		},
		{
			name: "returns expected repositories in specific namespace",
			request: &v1alpha1.GetPackageRepositoriesRequest{
				Context: &corev1.Context{
					Namespace: "default",
				},
			},
			repoNamespace: "non-default",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{},
		},
		{
			name: "returns expected repositories in specific namespace",
			request: &v1alpha1.GetPackageRepositoriesRequest{
				Context: &corev1.Context{
					Namespace: "default",
				},
			},
			repoNamespace: "default",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{
				{
					Name:      "repo-1",
					Namespace: "default",
					Url:       "https://charts.bitnami.com/bitnami",
				},
				{
					Name:      "repo-2",
					Namespace: "default",
					Url:       "https://charts.helm.sh/stable",
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
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
						},
						newRepos(tc.repoSpecs, tc.repoNamespace)...,
					), nil
				},
			}

			response, err := s.GetPackageRepositories(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(v1alpha1.PackageRepository{}, corev1.Context{})
					opt2 := cmpopts.SortSlices(lessPackageRepositoryFunc)
					if got, want := response.Repositories, tc.expectedPackageRepositories; !cmp.Equal(got, want, opt1, opt2) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
					}
				}
			}
		})
	}
}

// these are helpers to compare slices ignoring order
func lessAvailablePackageFunc(p1, p2 *corev1.AvailablePackageSummary) bool {
	return p1.DisplayName < p2.DisplayName
}

func lessPackageRepositoryFunc(p1, p2 *v1alpha1.PackageRepository) bool {
	return p1.Name < p2.Name && p1.Namespace < p2.Namespace
}
