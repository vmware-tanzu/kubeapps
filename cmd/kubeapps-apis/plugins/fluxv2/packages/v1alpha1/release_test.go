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
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

type testSpecGetInstalledPackages struct {
	repoName                  string
	repoNamespace             string
	repoIndex                 string
	chartName                 string
	chartTarGz                string
	chartSpecVersion          string // could be semver, e.g. "<=6.7.1"
	chartArtifactVersion      string // must be specific, e.g. "6.7.1"
	releaseName               string
	releaseNamespace          string
	releaseValues             map[string]interface{}
	releaseSuspend            bool
	releaseServiceAccountName string
	releaseStatus             map[string]interface{}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetInstalledPackageSummariesRequest
		existingObjs       []testSpecGetInstalledPackages
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "returns installed packages when install fails",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_failed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_failed,
				},
			},
		},
		{
			name: "returns installed packages when install is in progress",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_pending,
				},
			},
		},
		{
			name: "returns installed packages in a specific namespace",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_installed,
				},
			},
		},
		{
			name: "returns installed packages across all namespaces",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_installed,
					airflow_summary_installed,
				},
			},
		},
		{
			name: "returns limited results",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "0",
					PageSize:  1,
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_installed,
				},
				NextPageToken: "1",
			},
		},
		{
			name: "fetches results from an offset",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
				PaginationOptions: &corev1.PaginationOptions{
					PageSize:  1,
					PageToken: "1",
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					airflow_summary_installed,
				},
				NextPageToken: "2",
			},
		},
		{
			name: "fetches results from an offset (2)",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
				PaginationOptions: &corev1.PaginationOptions{
					PageSize:  1,
					PageToken: "2",
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{},
				NextPageToken:             "",
			},
		},
		{
			name: "returns installed package with semver constraint expression",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				airflow_existing_spec_semver,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					airflow_summary_semver,
				},
			},
		},
		{
			name: "returns installed package with latest '*' version",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_latest,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_latest,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			for _, existing := range tc.existingObjs {
				tarGzBytes, err := ioutil.ReadFile(existing.chartTarGz)
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
					"chart": existing.chartName,
					"sourceRef": map[string]interface{}{
						"name": existing.repoName,
						"kind": fluxHelmRepository,
					},
					"version":  existing.chartSpecVersion,
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
						"revision": existing.chartArtifactVersion,
					},
					"url": ts.URL,
				}
				chart := newChart(existing.chartName, existing.repoNamespace, chartSpec, chartStatus)
				runtimeObjs = append(runtimeObjs, chart)

				releaseSpec := map[string]interface{}{
					"chart": map[string]interface{}{
						"spec": map[string]interface{}{
							"chart":   existing.chartName,
							"version": existing.chartSpecVersion,
							"sourceRef": map[string]interface{}{
								"name":      existing.repoName,
								"kind":      fluxHelmRepository,
								"namespace": existing.repoNamespace,
							},
						},
					},
					"interval": "1m",
				}
				if len(existing.releaseValues) != 0 {
					unstructured.SetNestedMap(releaseSpec, existing.releaseValues, "values")
				}
				if existing.releaseSuspend {
					unstructured.SetNestedField(releaseSpec, existing.releaseSuspend, "suspend")
				}
				if len(existing.releaseServiceAccountName) != 0 {
					unstructured.SetNestedField(releaseSpec, existing.releaseServiceAccountName, "serviceAccountName")
				}
				release := newRelease(existing.releaseName, existing.releaseNamespace, releaseSpec, existing.releaseStatus)
				runtimeObjs = append(runtimeObjs, release)
			}

			s, mock, _, err := newServerWithChartsAndReleases(nil, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			for i, existing := range tc.existingObjs {
				if tc.request.GetPaginationOptions().GetPageSize() > 0 {
					pageOffset, err := pageOffsetFromPageToken(tc.request.GetPaginationOptions().GetPageToken())
					if err != nil {
						t.Fatalf("%+v", err)
					}
					startAt := int(tc.request.GetPaginationOptions().GetPageSize()) * pageOffset
					if i < startAt {
						continue
					} else if i >= startAt+int(tc.request.GetPaginationOptions().GetPageSize()) {
						break
					}
				}

				ts2, repo, err := newRepoWithIndex(existing.repoIndex, existing.repoName, existing.repoNamespace)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()

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

			opts := cmpopts.IgnoreUnexported(
				corev1.GetInstalledPackageSummariesResponse{},
				corev1.InstalledPackageSummary{},
				corev1.InstalledPackageReference{},
				corev1.Context{},
				corev1.VersionReference{},
				corev1.InstalledPackageStatus{},
				plugins.Plugin{})
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

type helmReleaseStub struct {
	name         string
	namespace    string
	chartVersion string
	notes        string
	status       release.Status
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetInstalledPackageDetailRequest
		existingK8sObjs    []testSpecGetInstalledPackages
		targetNamespace    string // this is where installation would actually place artifacts
		existingHelmStubs  []helmReleaseStub
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetInstalledPackageDetailResponse
	}{
		{
			name: "returns installed package detail when install fails",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "my-redis",
					Context: &corev1.Context{
						Namespace: "namespace-1",
					},
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_failed,
			},
			targetNamespace: "test",
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_failed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: redis_detail_failed,
			},
		},
		{
			name: "returns installed package detail when install is in progress",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "my-redis",
					Context: &corev1.Context{
						Namespace: "namespace-1",
					},
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending,
			},
			targetNamespace: "test",
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_pending,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: redis_detail_pending,
			},
		},
		{
			name: "returns installed package detail when install is successful",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "my-redis",
					Context: &corev1.Context{
						Namespace: "namespace-1",
					},
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			targetNamespace: "test",
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: redis_detail_completed,
			},
		},
		{
			name: "returns a 404 if the installed package is not found",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "namespace-1",
					},
					Identifier: "dontworrybehappy",
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			targetNamespace:    "test",
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "returns values and reconciliation options in package detail",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "namespace-1",
					},
					Identifier: "my-redis",
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed_with_values_and_reconciliation_options,
			},
			targetNamespace: "test",
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: redis_detail_completed_with_values_and_reconciliation_options,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs := []runtime.Object{}
			for _, existing := range tc.existingK8sObjs {
				tarGzBytes, err := ioutil.ReadFile(existing.chartTarGz)
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
					"chart": existing.chartName,
					"sourceRef": map[string]interface{}{
						"name": existing.repoName,
						"kind": fluxHelmRepository,
					},
					"version":  existing.chartSpecVersion,
					"interval": "1m",
				}
				chartStatus := map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"lastTransitionTime": "2021-08-12T03:25:38Z",
							"message":            "Fetched revision: " + existing.chartSpecVersion,
							"type":               "Ready",
							"status":             "True",
							"reason":             "ChartPullSucceeded",
						},
					},
					"artifact": map[string]interface{}{
						"revision": existing.chartArtifactVersion,
					},
					"url": ts.URL,
				}
				chart := newChart(existing.chartName, existing.repoNamespace, chartSpec, chartStatus)
				runtimeObjs = append(runtimeObjs, chart)

				releaseSpec := map[string]interface{}{
					"chart": map[string]interface{}{
						"spec": map[string]interface{}{
							"chart":   existing.chartName,
							"version": existing.chartSpecVersion,
							"sourceRef": map[string]interface{}{
								"name":      existing.repoName,
								"kind":      fluxHelmRepository,
								"namespace": existing.repoNamespace,
							},
						},
					},
					"interval":        "1m",
					"targetNamespace": tc.targetNamespace,
				}
				if len(existing.releaseValues) != 0 {
					unstructured.SetNestedMap(releaseSpec, existing.releaseValues, "values")
				}
				if existing.releaseSuspend {
					unstructured.SetNestedField(releaseSpec, existing.releaseSuspend, "suspend")
				}
				if len(existing.releaseServiceAccountName) != 0 {
					unstructured.SetNestedField(releaseSpec, existing.releaseServiceAccountName, "serviceAccountName")
				}
				release := newRelease(existing.releaseName, existing.releaseNamespace, releaseSpec, existing.releaseStatus)
				runtimeObjs = append(runtimeObjs, release)
			}

			actionConfig := newHelmActionConfig(t, tc.targetNamespace, tc.existingHelmStubs)
			s, mock, _, err := newServerWithChartsAndReleases(actionConfig, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.GetInstalledPackageDetailResponse{},
				corev1.InstalledPackageDetail{},
				corev1.InstalledPackageReference{},
				corev1.Context{},
				corev1.VersionReference{},
				corev1.InstalledPackageStatus{},
				plugins.Plugin{},
				corev1.ReconciliationOptions{},
				corev1.AvailablePackageReference{})
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

func newServerWithChartsAndReleases(actionConfig *action.Configuration, chartOrRelease ...runtime.Object) (*Server, redismock.ClientMock, *watch.FakeWatcher, error) {
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

	s, mock, err := newServer(clientGetter, actionConfig)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, mock, watcher, nil
}

// newHelmActionConfig returns an action.Configuration with fake clients and memory storage.
func newHelmActionConfig(t *testing.T, namespace string, rels []helmReleaseStub) *action.Configuration {
	t.Helper()

	memDriver := driver.NewMemory()

	actionConfig := &action.Configuration{
		Releases:     storage.Init(memDriver),
		KubeClient:   &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}

	for _, r := range rels {
		config := map[string]interface{}{}
		rel := &release.Release{
			Name:      r.name,
			Namespace: r.namespace,
			Info: &release.Info{
				Status: r.status,
				Notes:  r.notes,
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version:    r.chartVersion,
					Icon:       "https://example.com/icon.png",
					AppVersion: "1.2.3",
				},
			},
			Config: config,
		}
		err := actionConfig.Releases.Create(rel)
		if err != nil {
			t.Fatal(err)
		}
	}
	// It is the namespace of the the driver which determines the results. In the prod code,
	// the actionConfigGetter sets this using StorageForSecrets(namespace, clientset).
	memDriver.SetNamespace(namespace)

	return actionConfig
}

// misc global vars that get re-used in multiple tests scenarios
var redis_summary_installed = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
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
}

var redis_summary_failed = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
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
}

var redis_summary_pending = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
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
}

var airflow_summary_installed = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-2",
		},
		Identifier: "my-airflow",
		Plugin:     fluxPlugin,
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
}

var redis_summary_latest = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
	},
	Name:    "my-redis",
	IconUrl: "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
	PkgVersionReference: &corev1.VersionReference{
		Version: "*",
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
}

var airflow_summary_semver = &corev1.InstalledPackageSummary{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-2",
		},
		Identifier: "my-airflow",
		Plugin:     fluxPlugin,
	},
	Name:    "my-airflow",
	IconUrl: "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
	PkgVersionReference: &corev1.VersionReference{
		Version: "<=6.7.1",
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
}

var redis_existing_spec_completed = testSpecGetInstalledPackages{
	repoName:             "bitnami-1",
	repoNamespace:        "default",
	repoIndex:            "testdata/redis-many-versions.yaml",
	chartName:            "redis",
	chartTarGz:           "testdata/redis-14.4.0.tgz",
	chartSpecVersion:     "14.4.0",
	chartArtifactVersion: "14.4.0",
	releaseName:          "my-redis",
	releaseNamespace:     "namespace-1",
	releaseStatus: map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"lastTransitionTime": "2021-08-11T08:46:03Z",
				"type":               "Ready",
				"status":             "True",
				"reason":             "ReconciliationSucceeded",
				"message":            "Release reconciliation succeeded",
			},
			map[string]interface{}{
				"lastTransitionTime": "2021-08-11T08:46:03Z",
				"type":               "Released",
				"status":             "True",
				"reason":             "InstallSucceeded",
				"message":            "Helm install succeeded",
			},
		},
		"lastAppliedRevision":   "14.4.0",
		"lastAttemptedRevision": "14.4.0",
	},
}

var redis_existing_stub_completed = helmReleaseStub{
	name:         "test-my-redis",
	namespace:    "test",
	chartVersion: "14.4.0",
	notes:        "some notes",
	status:       release.StatusDeployed,
}

var redis_existing_spec_completed_with_values_and_reconciliation_options = testSpecGetInstalledPackages{
	repoName:                  "bitnami-1",
	repoNamespace:             "default",
	repoIndex:                 "testdata/redis-many-versions.yaml",
	chartName:                 "redis",
	chartTarGz:                "testdata/redis-14.4.0.tgz",
	chartSpecVersion:          "14.4.0",
	chartArtifactVersion:      "14.4.0",
	releaseName:               "my-redis",
	releaseNamespace:          "namespace-1",
	releaseSuspend:            true,
	releaseServiceAccountName: "foo",
	releaseValues: map[string]interface{}{
		"replica": []interface{}{
			map[string]interface{}{
				"replicaCount":  "1",
				"configuration": "xyz",
			},
		},
	},
	releaseStatus: map[string]interface{}{
		"conditions": []interface{}{
			map[string]interface{}{
				"lastTransitionTime": "2021-08-11T08:46:03Z",
				"type":               "Ready",
				"status":             "True",
				"reason":             "ReconciliationSucceeded",
				"message":            "Release reconciliation succeeded",
			},
			map[string]interface{}{
				"lastTransitionTime": "2021-08-11T08:46:03Z",
				"type":               "Released",
				"status":             "True",
				"reason":             "InstallSucceeded",
				"message":            "Helm install succeeded",
			},
		},
		"lastAppliedRevision":   "14.4.0",
		"lastAttemptedRevision": "14.4.0",
	},
}

var redis_existing_spec_failed = testSpecGetInstalledPackages{
	repoName:             "bitnami-1",
	repoNamespace:        "default",
	repoIndex:            "testdata/redis-many-versions.yaml",
	chartName:            "redis",
	chartTarGz:           "testdata/redis-14.4.0.tgz",
	chartSpecVersion:     "14.4.0",
	chartArtifactVersion: "14.4.0",
	releaseName:          "my-redis",
	releaseNamespace:     "namespace-1",
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
}

var redis_existing_stub_failed = helmReleaseStub{
	name:         "test-my-redis",
	namespace:    "test",
	chartVersion: "14.4.0",
	notes:        "some notes",
	status:       release.StatusFailed,
}

var airflow_existing_spec_completed = testSpecGetInstalledPackages{
	repoName:             "bitnami-2",
	repoNamespace:        "default",
	repoIndex:            "testdata/airflow-many-versions.yaml",
	chartName:            "airflow",
	chartTarGz:           "testdata/airflow-6.7.1.tgz",
	chartSpecVersion:     "6.7.1",
	chartArtifactVersion: "6.7.1",
	releaseName:          "my-airflow",
	releaseNamespace:     "namespace-2",
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
}

var airflow_existing_spec_semver = testSpecGetInstalledPackages{
	repoName:             "bitnami-2",
	repoNamespace:        "default",
	repoIndex:            "testdata/airflow-many-versions.yaml",
	chartName:            "airflow",
	chartTarGz:           "testdata/airflow-6.7.1.tgz",
	chartSpecVersion:     "<=6.7.1",
	chartArtifactVersion: "6.7.1",
	releaseName:          "my-airflow",
	releaseNamespace:     "namespace-2",
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
}

var redis_existing_spec_pending = testSpecGetInstalledPackages{
	repoName:             "bitnami-1",
	repoNamespace:        "default",
	repoIndex:            "testdata/redis-many-versions.yaml",
	chartName:            "redis",
	chartTarGz:           "testdata/redis-14.4.0.tgz",
	chartSpecVersion:     "14.4.0",
	chartArtifactVersion: "14.4.0",
	releaseName:          "my-redis",
	releaseNamespace:     "namespace-1",
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
}

var redis_existing_stub_pending = helmReleaseStub{
	name:         "test-my-redis",
	namespace:    "test",
	chartVersion: "14.4.0",
	notes:        "some notes",
	status:       release.StatusPendingInstall,
}

var redis_existing_spec_latest = testSpecGetInstalledPackages{
	repoName:             "bitnami-1",
	repoNamespace:        "default",
	repoIndex:            "testdata/redis-many-versions.yaml",
	chartName:            "redis",
	chartTarGz:           "testdata/redis-14.4.0.tgz",
	chartSpecVersion:     "*",
	chartArtifactVersion: "14.4.0",
	releaseName:          "my-redis",
	releaseNamespace:     "namespace-1",
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
}

var redis_detail_failed = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
	},
	Name: "my-redis",
	PkgVersionReference: &corev1.VersionReference{
		Version: "14.4.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      false,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
		UserReason: "InstallFailed",
	},
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "bitnami-1/redis",
		Context:    &corev1.Context{Namespace: "default"},
		Plugin:     fluxPlugin,
	},
	PostInstallationNotes: "some notes",
}

var redis_detail_pending = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
	},
	Name: "my-redis",
	PkgVersionReference: &corev1.VersionReference{
		Version: "14.4.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      false,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
		UserReason: "Progressing",
	},
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "bitnami-1/redis",
		Context:    &corev1.Context{Namespace: "default"},
		Plugin:     fluxPlugin,
	},
	PostInstallationNotes: "some notes",
}

var redis_detail_completed = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
	},
	Name:              "my-redis",
	CurrentPkgVersion: "14.4.0",
	PkgVersionReference: &corev1.VersionReference{
		Version: "14.4.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval: 60,
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded",
	},
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "bitnami-1/redis",
		Context:    &corev1.Context{Namespace: "default"},
		Plugin:     fluxPlugin,
	},
	PostInstallationNotes: "some notes",
}

var redis_detail_completed_with_values_and_reconciliation_options = &corev1.InstalledPackageDetail{
	InstalledPackageRef: &corev1.InstalledPackageReference{
		Context: &corev1.Context{
			Namespace: "namespace-1",
		},
		Identifier: "my-redis",
		Plugin:     fluxPlugin,
	},
	Name:              "my-redis",
	CurrentPkgVersion: "14.4.0",
	PkgVersionReference: &corev1.VersionReference{
		Version: "14.4.0",
	},
	ReconciliationOptions: &corev1.ReconciliationOptions{
		Interval:           60,
		Suspend:            true,
		ServiceAccountName: "foo",
	},
	Status: &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded",
	},
	ValuesApplied: "{\"replica\":[{\"configuration\":\"xyz\",\"replicaCount\":\"1\"}]}",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "bitnami-1/redis",
		Context:    &corev1.Context{Namespace: "default"},
		Plugin:     fluxPlugin,
	},
	PostInstallationNotes: "some notes",
}
