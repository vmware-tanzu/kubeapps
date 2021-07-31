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

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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

type testSpecGetInstalledPackageSummaries struct {
	repoName         string
	repoNamespace    string
	repoIndex        string
	chartName        string
	chartTarGz       string
	chartRevision    string
	releaseName      string
	releaseNamespace string
	releaseStatus    map[string]interface{}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetInstalledPackageSummariesRequest
		existingObjs       []testSpecGetInstalledPackageSummaries
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "returns installed packages when install fails",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackageSummaries{
				{
					repoName:         "bitnami-1",
					repoNamespace:    "default",
					repoIndex:        "testdata/redis-many-versions.yaml",
					chartName:        "redis",
					chartTarGz:       "testdata/redis-14.4.0.tgz",
					chartRevision:    "14.4.0",
					releaseName:      "my-redis",
					releaseNamespace: "namespace-1",
					releaseStatus: map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
								"reason": "InstallFailed",
							},
						},
						"lastAttemptedRevision": "14.4.0",
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Namespace: "namespace-1",
							},
							Identifier: "my-redis",
						},
						Name:    "my-redis",
						IconUrl: "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "14.4.0",
						},
						CurrentAppVersion: "6.2.4",
						PkgDisplayName:    "redis",
						ShortDescription:  "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
						Status: &corev1.InstalledPackageStatus{
							Ready:      false,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
							UserReason: "InstallFailed",
						},
						LatestPkgVersion: "14.6.1",
					},
				},
			},
		},
		{
			name: "returns installed packages when install is in progress",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackageSummaries{
				{
					repoName:         "bitnami-1",
					repoNamespace:    "default",
					repoIndex:        "testdata/redis-many-versions.yaml",
					chartName:        "redis",
					chartTarGz:       "testdata/redis-14.4.0.tgz",
					chartRevision:    "14.4.0",
					releaseName:      "my-redis",
					releaseNamespace: "namespace-1",
					releaseStatus: map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "Unknown",
								"reason": "Progressing",
							},
						},
						"lastAttemptedRevision": "14.4.0",
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Namespace: "namespace-1",
							},
							Identifier: "my-redis",
						},
						Name:    "my-redis",
						IconUrl: "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "14.4.0",
						},
						CurrentAppVersion: "6.2.4",
						PkgDisplayName:    "redis",
						ShortDescription:  "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
						Status: &corev1.InstalledPackageStatus{
							Ready:      false,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
							UserReason: "Progressing",
						},
						LatestPkgVersion: "14.6.1",
					},
				},
			},
		},
		{
			name: "returns installed packages in a specific namespace",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackageSummaries{
				{
					repoName:         "bitnami-1",
					repoNamespace:    "default",
					repoIndex:        "testdata/redis-many-versions.yaml",
					chartName:        "redis",
					chartTarGz:       "testdata/redis-14.4.0.tgz",
					chartRevision:    "14.4.0",
					releaseName:      "my-redis",
					releaseNamespace: "namespace-1",
					releaseStatus: map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
								"reason": "ReconciliationSucceeded",
							},
						},
						"lastAppliedRevision":   "14.4.0",
						"lastAttemptedRevision": "14.4.0",
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Namespace: "namespace-1",
							},
							Identifier: "my-redis",
						},
						Name:    "my-redis",
						IconUrl: "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "14.4.0",
						},
						CurrentPkgVersion: "14.4.0",
						CurrentAppVersion: "6.2.4",
						PkgDisplayName:    "redis",
						ShortDescription:  "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "ReconciliationSucceeded",
						},
						LatestPkgVersion: "14.6.1",
					},
				},
			},
		},
		{
			name: "returns installed packages across all namespaces",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackageSummaries{
				{
					repoName:         "bitnami-1",
					repoNamespace:    "default",
					repoIndex:        "testdata/redis-many-versions.yaml",
					chartName:        "redis",
					chartTarGz:       "testdata/redis-14.4.0.tgz",
					chartRevision:    "14.4.0",
					releaseName:      "my-redis",
					releaseNamespace: "namespace-1",
					releaseStatus: map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
								"reason": "ReconciliationSucceeded",
							},
						},
						"lastAppliedRevision":   "14.4.0",
						"lastAttemptedRevision": "14.4.0",
					},
				},
				{
					repoName:         "bitnami-2",
					repoNamespace:    "default",
					repoIndex:        "testdata/airflow-many-versions.yaml",
					chartName:        "airflow",
					chartTarGz:       "testdata/airflow-6.7.1.tgz",
					chartRevision:    "6.7.1",
					releaseName:      "my-airflow",
					releaseNamespace: "namespace-2",
					releaseStatus: map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
								"reason": "ReconciliationSucceeded",
							},
						},
						"lastAppliedRevision":   "6.7.1",
						"lastAttemptedRevision": "6.7.1",
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{

						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Namespace: "namespace-1",
							},
							Identifier: "my-redis",
						},
						Name:    "my-redis",
						IconUrl: "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "14.4.0",
						},
						CurrentPkgVersion: "14.4.0",
						CurrentAppVersion: "6.2.4",
						PkgDisplayName:    "redis",
						ShortDescription:  "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "ReconciliationSucceeded",
						},
						LatestPkgVersion: "14.6.1",
					},
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Namespace: "namespace-2",
							},
							Identifier: "my-airflow",
						},
						Name:    "my-airflow",
						IconUrl: "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "6.7.1",
						},
						CurrentPkgVersion: "6.7.1",
						LatestPkgVersion:  "10.2.1",
						CurrentAppVersion: "1.10.12",
						ShortDescription:  "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
						PkgDisplayName:    "airflow",
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "ReconciliationSucceeded",
						},
					},
				},
			},
		},
		/*
			{
				name: "returns limited results",
				request: &corev1.GetInstalledPackageSummariesRequest{
					Context: &corev1.Context{Namespace: ""},
					PaginationOptions: &corev1.PaginationOptions{
						PageSize: 2,
					},
				},
				existingReleases: []releaseStub{
					{
						name:         "my-release-1",
						namespace:    "namespace-1",
						chartVersion: "1.2.3",
						status:       release.StatusDeployed,
					},
					{
						name:         "my-release-2",
						namespace:    "namespace-2",
						status:       release.StatusDeployed,
						chartVersion: "3.4.5",
					},
					{
						name:         "my-release-3",
						namespace:    "namespace-3",
						chartVersion: "4.5.6",
						status:       release.StatusDeployed,
					},
				},
				expectedStatusCode: codes.OK,
				expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
					InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
						{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Namespace: "namespace-1",
								},
								Identifier: "my-release-1",
							},
							Name:    "my-release-1",
							IconUrl: "https://example.com/icon.png",
							PkgVersionReference: &corev1.VersionReference{
								Version: "1.2.3",
							},
							CurrentPkgVersion: "1.2.3",
							LatestPkgVersion:  "1.2.3",
							CurrentAppVersion: DefaultAppVersion,
							Status: &corev1.InstalledPackageStatus{
								Ready:      true,
								Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
								UserReason: "deployed",
							},
						},
						{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Namespace: "namespace-2",
								},
								Identifier: "my-release-2",
							},
							Name:    "my-release-2",
							IconUrl: "https://example.com/icon.png",
							PkgVersionReference: &corev1.VersionReference{
								Version: "3.4.5",
							},
							CurrentPkgVersion: "3.4.5",
							LatestPkgVersion:  "3.4.5",
							CurrentAppVersion: DefaultAppVersion,
							Status: &corev1.InstalledPackageStatus{
								Ready:      true,
								Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
								UserReason: "deployed",
							},
						},
					},
					NextPageToken: "3",
				},
			},
			{
				name: "fetches results from an offset",
				request: &corev1.GetInstalledPackageSummariesRequest{
					Context: &corev1.Context{Namespace: ""},
					PaginationOptions: &corev1.PaginationOptions{
						PageSize:  2,
						PageToken: "2",
					},
				},
				existingReleases: []releaseStub{
					{
						name:         "my-release-1",
						namespace:    "namespace-1",
						chartVersion: "1.2.3",
						status:       release.StatusDeployed,
					},
					{
						name:         "my-release-2",
						namespace:    "namespace-2",
						status:       release.StatusDeployed,
						chartVersion: "3.4.5",
					},
					{
						name:         "my-release-3",
						namespace:    "namespace-3",
						chartVersion: "4.5.6",
						status:       release.StatusDeployed,
					},
				},
				expectedStatusCode: codes.OK,
				expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
					InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
						{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Namespace: "namespace-3",
								},
								Identifier: "my-release-3",
							},
							Name:    "my-release-3",
							IconUrl: "https://example.com/icon.png",
							PkgVersionReference: &corev1.VersionReference{
								Version: "4.5.6",
							},
							CurrentPkgVersion: "4.5.6",
							LatestPkgVersion:  "4.5.6",
							CurrentAppVersion: DefaultAppVersion,
							Status: &corev1.InstalledPackageStatus{
								Ready:      true,
								Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
								UserReason: "deployed",
							},
						},
					},
					NextPageToken: "",
				},
			},
			{
				name: "includes a latest package version when available",
				request: &corev1.GetInstalledPackageSummariesRequest{
					Context: &corev1.Context{Namespace: "namespace-1"},
				},
				existingReleases: []releaseStub{
					{
						name:         "my-release-1",
						namespace:    "namespace-1",
						chartVersion: "1.2.3",
						status:       release.StatusDeployed,
					},
				},
				expectedStatusCode: codes.OK,
				expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
					InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
						{
							InstalledPackageRef: &corev1.InstalledPackageReference{
								Context: &corev1.Context{
									Namespace: "namespace-1",
								},
								Identifier: "my-release-1",
							},
							Name:    "my-release-1",
							IconUrl: "https://example.com/icon.png",
							PkgVersionReference: &corev1.VersionReference{
								Version: "1.2.3",
							},
							CurrentPkgVersion: "1.2.3",
							LatestPkgVersion:  "1.2.5",
							CurrentAppVersion: DefaultAppVersion,
							Status: &corev1.InstalledPackageStatus{
								Ready:      true,
								Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
								UserReason: "deployed",
							},
						},
					},
				},
			},
		*/
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			for _, available := range tc.existingObjs {
				tarGzBytes, err := ioutil.ReadFile(available.chartTarGz)
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
					"chart": available.chartName,
					"sourceRef": map[string]interface{}{
						"name": available.repoName,
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
						"revision": available.chartRevision,
					},
					"url": ts.URL,
				}
				chart := newChart(available.chartName, available.repoNamespace, chartSpec, chartStatus)
				runtimeObjs = append(runtimeObjs, chart)

				releaseSpec := map[string]interface{}{
					"chart": map[string]interface{}{
						"spec": map[string]interface{}{
							"chart":   available.chartName,
							"version": available.chartRevision,
							"sourceRef": map[string]interface{}{
								"name":      available.repoName,
								"kind":      fluxHelmRepository,
								"namespace": available.repoNamespace,
							},
							"interval": "1m",
						},
					},
				}

				release := newRelease(available.releaseName, available.releaseNamespace, releaseSpec, available.releaseStatus)
				runtimeObjs = append(runtimeObjs, release)
			}

			s, mock, _, err := newServerWithChartsAndReleases(runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			for _, existing := range tc.existingObjs {
				indexYAMLBytes, err := ioutil.ReadFile(existing.repoIndex)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				// stand up an http server just for the duration of this test
				ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, string(indexYAMLBytes))
				}))
				defer ts2.Close()

				repoSpec := map[string]interface{}{
					"url":      "https://example.repo.com/charts",
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
					"url": ts2.URL,
				}
				repo := newRepo(existing.repoName, existing.repoNamespace, repoSpec, repoStatus)

				redisKey, bytes, err := redisKeyValueForRuntimeObject(repo)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				mock.ExpectGet(redisKey).SetVal(string(bytes))
			}

			response, err := s.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, corev1.VersionReference{}, corev1.InstalledPackageStatus{})
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

func newRelease(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":            name,
		"generation":      int64(1),
		"resourceVersion": "1",
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}

	obj := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", fluxHelmReleaseGroup, fluxHelmReleaseVersion),
		"kind":       fluxHelmRelease,
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

func newServerWithChartsAndReleases(chartOrRelease ...runtime.Object) (*Server, redismock.ClientMock, *watch.FakeWatcher, error) {
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}:                         fluxHelmChartList,
			{Group: fluxHelmReleaseGroup, Version: fluxHelmReleaseVersion, Resource: fluxHelmReleases}: fluxHelmReleaseList,
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}:                   fluxHelmRepositoryList,
		},
		chartOrRelease...)

	clientGetter := func(context.Context) (dynamic.Interface, error) {
		return dynamicClient, nil
	}

	watcher := watch.NewFake()

	// see chart_test.go for explanation
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

	dynamicClient.Fake.PrependWatchReactor(
		fluxHelmCharts,
		k8stesting.DefaultWatchReactor(watcher, nil))

	s, mock, err := newServerWithClientGetter(clientGetter)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, mock, watcher, nil
}
