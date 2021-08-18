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
	"sync"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

// global variable to keep track of flux helm charts created for the purpose of GetAvailablePackageDetail
var createdChartCount int

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		testName              string
		request               *corev1.GetAvailablePackageDetailRequest
		repoName              string
		repoNamespace         string
		chartName             string
		chartTarGz            string
		chartRevision         string
		chartExists           bool
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
			chartName:     "redis",
			chartTarGz:    "testdata/redis-14.4.0.tgz",
			chartRevision: "14.4.0",
			chartExists:   true,
			expectedPackageDetail: &corev1.AvailablePackageDetail{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context:    &corev1.Context{Namespace: "default"},
					Plugin:     fluxPlugin,
				},
				Name:             "redis",
				PkgVersion:       "14.4.0",
				AppVersion:       "6.2.4",
				RepoUrl:          "https://example.repo.com/charts",
				HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
				IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
				DisplayName:      "redis",
				Categories:       []string{"Database"},
				ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
				Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
				DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
				ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
				SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
				Maintainers: []*corev1.Maintainer{
					{
						Name:  "Bitnami",
						Email: "containers@bitnami.com",
					},
					{
						Name:  "desaintmartin",
						Email: "cedric@desaintmartin.fr",
					},
				},
			},
		},
		{
			testName:      "it returns details about the redis package with specific version in bitnami repo",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context:    &corev1.Context{Namespace: "default"},
				},
				PkgVersion: "14.3.4",
			},
			chartName:     "redis",
			chartTarGz:    "testdata/redis-14.3.4.tgz",
			chartRevision: "14.4.0",
			chartExists:   false,
			expectedPackageDetail: &corev1.AvailablePackageDetail{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context:    &corev1.Context{Namespace: "default"},
					Plugin:     fluxPlugin,
				},
				Name:             "redis",
				PkgVersion:       "14.3.4",
				AppVersion:       "6.2.4",
				RepoUrl:          "https://example.repo.com/charts",
				IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
				HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
				DisplayName:      "redis",
				Categories:       []string{"Database"},
				ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
				Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
				DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
				ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
				SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
				Maintainers: []*corev1.Maintainer{
					{
						Name:  "Bitnami",
						Email: "containers@bitnami.com",
					},
					{
						Name:  "desaintmartin",
						Email: "cedric@desaintmartin.fr",
					},
				},
			},
		},
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

			ts2, repo, err := newRepoWithIndex("testdata/valid-index.yaml", tc.repoName, tc.repoNamespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			charts := []runtime.Object{}
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
				"artifact": map[string]interface{}{
					"revision": tc.chartRevision,
				},
				"url": ts.URL,
			}
			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			if tc.chartExists {
				charts = append(charts, chart)
			}

			s, mock, watcher, err := newServerWithRepoAndCharts(repo, charts...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			wg := sync.WaitGroup{}
			wg.Add(1)

			if !tc.chartExists {
				go func() {
					wg.Wait()
					watcher.Modify(chart)
				}()
			}

			chartCountBefore := createdChartCount

			key, bytes, err := redisKeyValueForRuntimeObject(repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			mock.ExpectGet(key).SetVal(string(bytes))

			response, err := s.GetAvailablePackageDetail(newContext(context.Background(), &wg), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageReference{}, corev1.Context{}, corev1.Maintainer{}, plugins.Plugin{})
			// these few fields a bit special in that they are all very long strings,
			// so we'll do a 'Contains' check for these instead of 'Equals'
			opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "Readme", "DefaultValues", "ValuesSchema")
			if got, want := response.AvailablePackageDetail, tc.expectedPackageDetail; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
			if !strings.Contains(response.AvailablePackageDetail.Readme, tc.expectedPackageDetail.Readme) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.Readme, response.AvailablePackageDetail.Readme)
			}
			if !strings.Contains(response.AvailablePackageDetail.DefaultValues, tc.expectedPackageDetail.DefaultValues) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.DefaultValues, response.AvailablePackageDetail.DefaultValues)
			}
			if !strings.Contains(response.AvailablePackageDetail.ValuesSchema, tc.expectedPackageDetail.ValuesSchema) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.ValuesSchema, response.AvailablePackageDetail.ValuesSchema)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			// make sure we've cleaned up after ourselves
			chartCountAfter := createdChartCount
			if chartCountBefore != chartCountAfter {
				t.Fatalf("flux helmchart count (-want +got): -[%d], [%d]", chartCountBefore, chartCountAfter)
			}
		})
	}
}

func TestNegativeGetAvailablePackageDetail(t *testing.T) {
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
			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)

			ts2, repo, err := newRepoWithIndex("testdata/valid-index.yaml", tc.repoName, tc.repoNamespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, _, err := newServerWithRepoAndCharts(repo, chart)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			if response != nil {
				t.Fatalf("want nil, got %v", response)
			}
		})
	}
}

func TestNonExistingRepoOrInvalidPkgVersionGetAvailablePackageDetail(t *testing.T) {
	negativeTestCases := []struct {
		testName        string
		request         *corev1.GetAvailablePackageDetailRequest
		repoName        string
		repoExists      bool
		expectCacheMiss bool
		repoNamespace   string
		chartName       string
		statusCode      codes.Code
	}{
		{
			testName:      "it fails if request has invalid version",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			repoExists:    true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
				PkgVersion: "99.99.0",
			},
			chartName:  "redis",
			statusCode: codes.Internal,
		},
		{
			testName:      "it fails if repo does not exist",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			repoExists:    false,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			statusCode: codes.NotFound,
		},
		{
			testName:        "it fails if repo does not exist in specified namespace",
			repoName:        "bitnami-1",
			repoNamespace:   "non-default",
			repoExists:      true,
			expectCacheMiss: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			statusCode: codes.NotFound,
		},
		{
			testName:      "it fails if request has invalid chart",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			repoExists:    true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis-123",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
			},
			chartName:  "redis",
			statusCode: codes.Internal,
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			chartSpec := map[string]interface{}{
				"chart": tc.chartName,
				"sourceRef": map[string]interface{}{
					"name": "bitnami-1",
					"kind": "HelmRepository",
				},
				"interval": "10m",
			}
			chartStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":    "Ready",
						"status":  "False",
						"reason":  "ChartPullFailed",
						"message": "message: no chart name/version found",
					},
				},
			}

			ts2, repo, err := newRepoWithIndex("testdata/valid-index.yaml", tc.repoName, tc.repoNamespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			s, mock, watcher, err := newServerWithRepoAndCharts(repo, chart)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if !tc.expectCacheMiss {
				key, bytes, err := redisKeyValueForRuntimeObject(repo)
				if tc.repoExists {
					if err != nil {
						t.Fatalf("%+v", err)
					}
					mock.ExpectGet(key).SetVal(string(bytes))
				} else {
					mock.ExpectGet(key).RedisNil()
				}
			}

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				wg.Wait()
				watcher.Modify(chart)
			}()

			response, err := s.GetAvailablePackageDetail(newContext(context.Background(), &wg), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			if response != nil {
				t.Fatalf("want nil, got %v", response)
			}
		})
	}
}

func TestNegativeGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name:               "it returns invalid argument if called without a package reference",
			request:            nil,
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without namespace",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without an identifier",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, _, err := newServerWithRepos()
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		repoIndex          string
		repoNamespace      string
		repoName           string
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name:          "it returns the package version summary for redis chart in bitnami repo",
			repoIndex:     "testdata/redis-many-versions.yaml",
			repoNamespace: "kubeapps",
			repoName:      "bitnami",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
					Identifier: "bitnami/redis",
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
					{PkgVersion: "14.6.1", AppVersion: "6.2.4"},
					{PkgVersion: "14.6.0", AppVersion: "6.2.4"},
					{PkgVersion: "14.5.0", AppVersion: "6.2.4"},
					{PkgVersion: "14.4.0", AppVersion: "6.2.4"},
					{PkgVersion: "13.0.1", AppVersion: "6.2.1"},
					{PkgVersion: "13.0.0", AppVersion: "6.2.1"},
					{PkgVersion: "12.10.1", AppVersion: "6.0.12"},
					{PkgVersion: "12.10.0", AppVersion: "6.0.12"},
					{PkgVersion: "12.9.2", AppVersion: "6.0.12"},
					{PkgVersion: "12.9.1", AppVersion: "6.0.12"},
					{PkgVersion: "12.9.0", AppVersion: "6.0.12"},
					{PkgVersion: "12.8.3", AppVersion: "6.0.12"},
					{PkgVersion: "12.8.2", AppVersion: "6.0.12"},
					{PkgVersion: "12.8.1", AppVersion: "6.0.12"},
				},
			},
		},
		{
			name:          "it returns error for non-existent chart",
			repoIndex:     "testdata/redis-many-versions.yaml",
			repoNamespace: "kubeapps",
			repoName:      "bitnami",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
					Identifier: "bitnami/redis-123",
				},
			},
			expectedStatusCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts2, repo, err := newRepoWithIndex(tc.repoIndex, tc.repoName, tc.repoNamespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, _, err := newServerWithRepos(repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			redisKey, bytes, err := redisKeyValueForRuntimeObject(repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			mock.ExpectGet(redisKey).SetVal(string(bytes))

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func newChart(name, namespace string, spec, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":            name,
		"generation":      int64(1),
		"resourceVersion": "1",
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

func newServerWithRepoAndCharts(repo runtime.Object, charts ...runtime.Object) (*Server, redismock.ClientMock, *watch.FakeWatcher, error) {
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}:       fluxHelmChartList,
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
		},
		append([]runtime.Object{repo}, charts...)...)

	// here we are essentially adding on to how List() works for HelmRepository objects
	// this is done so that the the item list returned by List() command with fake client contains
	// a "resourceVersion" field in its metadata, which happens in a real k8s environment and
	// is critical
	reactor := dynamicClient.Fake.ReactionChain[0]
	dynamicClient.Fake.PrependReactor("list", fluxHelmRepositories,
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			handled, ret, err := reactor.React(action)
			ulist, ok := ret.(*unstructured.UnstructuredList)
			if ok && ulist != nil {
				ulist.SetResourceVersion("1")
			}
			return handled, ret, err
		})
	dynamicClient.Fake.PrependReactor("create", fluxHelmCharts,
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			handled, ret, err := reactor.React(action)
			if err == nil {
				uobj, ok := ret.(*unstructured.Unstructured)
				if ok && uobj != nil {
					createdChartCount++
					uobj.SetResourceVersion("1")
				}
			}
			return handled, ret, err
		})
	dynamicClient.Fake.PrependReactor("delete", fluxHelmCharts,
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			handled, ret, err := reactor.React(action)
			if err == nil {
				createdChartCount--
			}
			return handled, ret, err
		})

	clientGetter := func(context.Context) (dynamic.Interface, error) {
		return dynamicClient, nil
	}

	watcher := watch.NewFake()

	dynamicClient.Fake.PrependWatchReactor(
		fluxHelmCharts,
		k8stesting.DefaultWatchReactor(watcher, nil))

	s, mock, err := newServerWithClientGetter(clientGetter, repo)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, mock, watcher, nil
}
