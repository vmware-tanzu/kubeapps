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
	"time"

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
				IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
				DisplayName:      "redis",
				Categories:       []string{"Database"},
				ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
				Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
				DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
				ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
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
				IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
				DisplayName:      "redis",
				Categories:       []string{"Database"},
				ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
				Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
				DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
				ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
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

			s, watcher, mock, err := newServerWithCharts(charts...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if !tc.chartExists {
				go func() {
					// TODO (gfichtenholt) yes, I don't like time.Sleep() anymore than the next guy
					// see if there is a better way to force waitUntilChartPushIsComplete to stop
					// on the server-side. I think the better way would be to do something similar to
					// what I did with ResourceWatcherCache - add a sync.WaitGroup into production code
					// somewhere in GetAvailablePackageDetail() that would only be used from within this
					// unit test
					time.Sleep(500 * time.Millisecond)
					watcher.Modify(chart)
				}()
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
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
			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			s, _, mock, err := newServerWithCharts(chart)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			_, err = s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func newChart(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
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

func newServerWithCharts(charts ...runtime.Object) (*Server, *watch.FakeWatcher, redismock.ClientMock, error) {
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}:       fluxHelmChartList,
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
		},
		charts...)

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
			uobj, ok := ret.(*unstructured.Unstructured)
			if ok && uobj != nil {
				uobj.SetResourceVersion("1")
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

	s, mock, err := newServerWithClientGetter(clientGetter)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, watcher, mock, nil
}
