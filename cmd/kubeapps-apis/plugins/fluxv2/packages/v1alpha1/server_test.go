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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
)

func TestGetAvailablePackagesStatus(t *testing.T) {
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
			name: "returns without error if response status does not contain conditions",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
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
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
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
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
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
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
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

			response, err := s.GetAvailablePackageSummaries(
				context.Background(),
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

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
		"name":       name,
		"generation": int64(1),
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
		status["observedGeneration"] = int64(1)
		obj["status"] = status
	}

	return &unstructured.Unstructured{
		Object: obj,
	}
}

// newRepos takes a map of specs keyed by object name converting them to runtime objects.
func newRepos(specs map[string]map[string]interface{}, namespace string) []runtime.Object {
	repos := []runtime.Object{}
	for name, spec := range specs {
		repo := newRepo(name, namespace, spec, nil)
		repos = append(repos, repo)
	}
	return repos
}

type testRepoStruct struct {
	name      string
	namespace string
	url       string
	index     string
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		testName         string
		request          *corev1.GetAvailablePackageSummariesRequest
		testRepos        []testRepoStruct
		expectedPackages []*corev1.AvailablePackageSummary
	}{
		{
			testName: "it returns a couple of fluxv2 packages from the cluster (no request ns specified)",
			testRepos: []testRepoStruct{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:      "acs-engine-autoscaler",
					LatestPkgVersion: "2.1.1",
					IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/acs-engine-autoscaler",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
				{
					DisplayName:      "wordpress",
					LatestPkgVersion: "0.7.5",
					IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/wordpress",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
			},
		},
		{
			testName: "it returns a couple of fluxv2 packages from the cluster (when request namespace is specified)",
			testRepos: []testRepoStruct{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "default"}},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:      "acs-engine-autoscaler",
					LatestPkgVersion: "2.1.1",
					IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/acs-engine-autoscaler",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
				{
					DisplayName:      "wordpress",
					LatestPkgVersion: "0.7.5",
					IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/wordpress",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
			},
		},
		{
			testName: "it returns all fluxv2 packages from the cluster (when request namespace is does not match repo namespace)",
			testRepos: []testRepoStruct{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     "testdata/jetstack-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "non-default"}},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					DisplayName:      "acs-engine-autoscaler",
					LatestPkgVersion: "2.1.1",
					IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/acs-engine-autoscaler",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
				{
					DisplayName:      "cert-manager",
					LatestPkgVersion: "v1.4.0",
					IconUrl:          "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "jetstack-1/cert-manager",
						Context:    &corev1.Context{Namespace: "ns1"},
					},
				},
				{
					DisplayName:      "wordpress",
					LatestPkgVersion: "0.7.5",
					IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Identifier: "bitnami-1/wordpress",
						Context:    &corev1.Context{Namespace: "default"},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repos := []runtime.Object{}
			for _, rs := range tc.testRepos {
				indexYAMLBytes, err := ioutil.ReadFile(rs.index)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				// stand up an http server just for the duration of this test
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, string(indexYAMLBytes))
				}))
				defer ts.Close()

				repoSpec := map[string]interface{}{
					"url":      rs.url,
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
				repos = append(repos, newRepo(rs.name, rs.namespace, repoSpec, repoStatus))
			}
			s := Server{
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
						},
						repos...,
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
			repoNamespace: "default",
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
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
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

func newChart(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":       name,
		"generation": int64(1),
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}

	obj := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
		"kind":       fluxHelmChart,
		"metadata":   metadata,
	}

	if spec != nil {
		obj["spec"] = spec
	}

	if status != nil {
		status["observedGeneration"] = int64(1)
		obj["status"] = status
	}

	return &unstructured.Unstructured{
		Object: obj,
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		testName              string
		request               *corev1.GetAvailablePackageDetailRequest
		repoName              string
		repoNamespace         string
		chartName             string
		chartTarGz            string
		expectedPackageDetail *corev1.AvailablePackageDetail
	}{
		{
			testName:      "it returns details about the redis package in bitnami repo",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			chartTarGz: "testdata/redis-14.4.0.tgz",
			expectedPackageDetail: &corev1.AvailablePackageDetail{
				// TODO (gfichtenholt) other fields
				LongDescription: "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
			},
		},
		// TODO (gfichtenholt) specific version
		// TODO (gfichtenholt) negative test
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			tarGzBytes, err := ioutil.ReadFile(tc.chartTarGz)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write(tarGzBytes)
			}))
			defer ts.Close()

			chartSpec := map[string]interface{}{
				"chart": tc.chartName,
				"sourceRef": map[string]interface{}{
					"name": tc.repoName,
					"kind": fluxHelmRepository,
				},
				"interval": "10m",
			}
			chartStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "ChartPullSucceeded",
					},
				},
				"url": ts.URL,
			}
			repo := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			s := Server{
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}: fluxHelmChartList,
						},
						repo,
					), nil
				},
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageReference{}, corev1.Context{})
			opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "LongDescription")
			if got, want := response.AvailablePackageDetail, tc.expectedPackageDetail; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
			if !strings.Contains(response.AvailablePackageDetail.LongDescription, tc.expectedPackageDetail.LongDescription) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.LongDescription, response.AvailablePackageDetail.LongDescription)
			}
		})
	}

	negativeTestCases := []struct {
		testName      string
		request       *corev1.GetAvailablePackageDetailRequest
		repoName      string
		repoNamespace string
		chartName     string
		statusCode    codes.Code
	}{
		{
			testName:      "it fails if request is missing namespace",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "redis",
				}},
			chartName:  "redis",
			statusCode: codes.InvalidArgument,
		},
		{
			testName:      "it fails if request has invalid identifier",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			chartSpec := map[string]interface{}{
				"chart": tc.chartName,
				"sourceRef": map[string]interface{}{
					"name": "does-not-matter-for-now",
					"kind": "HelmRepository",
				},
				"interval": "10m",
			}
			chartStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "ChartPullSucceeded",
					},
				},
				"url": "does-not-matter",
			}
			repo := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			s := Server{
				clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}: fluxHelmChartList,
						},
						repo,
					), nil
				},
			}

			_, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
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
