// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	paginate "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	resourcerefstest "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmchartutil "helm.sh/helm/v3/pkg/chartutil"
	helmkubefake "helm.sh/helm/v3/pkg/kube/fake"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8sapiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8sapiextensionsclientfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smetaunstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8swatch "k8s.io/apimachinery/pkg/watch"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8dynamicclientfake "k8s.io/client-go/dynamic/fake"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type testSpecGetInstalledPackages struct {
	repoName                  string
	repoNamespace             string
	repoIndex                 string
	chartName                 string
	chartTarGz                string
	chartSpecVersion          string // could be semver constraint, e.g. "<=6.7.1"
	chartArtifactVersion      string // must be specific, e.g. "6.7.1"
	releaseName               string
	releaseNamespace          string
	releaseValues             map[string]interface{}
	releaseSuspend            bool
	releaseServiceAccountName string
	releaseStatus             map[string]interface{}
	targetNamespace           string
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest
		existingObjs       []testSpecGetInstalledPackages
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "returns installed packages when install fails",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_failed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_failed,
				},
			},
		},
		{
			name: "returns installed packages when install is in progress",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_pending,
				},
			},
		},
		{
			name: "returns installed packages when install is in progress (2)",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending_2,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_pending_2,
				},
			},
		},
		{
			name: "returns installed packages in a specific namespace",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_installed,
				},
			},
		},
		{
			name: "returns installed packages across all namespaces",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_installed,
					airflow_summary_installed,
				},
			},
		},
		{
			name: "returns limited results",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageToken: "0",
					PageSize:  1,
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_installed,
				},
				NextPageToken: "1",
			},
		},
		{
			name: "fetches results from an offset",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageSize:  1,
					PageToken: "1",
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					airflow_summary_installed,
				},
				NextPageToken: "2",
			},
		},
		{
			name: "fetches results from an offset (2)",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageSize:  1,
					PageToken: "2",
				},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
				airflow_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{},
				NextPageToken:             "",
			},
		},
		{
			name: "returns installed package with semver constraint expression",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				airflow_existing_spec_semver,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					airflow_summary_semver,
				},
			},
		},
		{
			name: "returns installed package with latest '*' version",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_latest,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					redis_summary_latest,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs, cleanup := newRuntimeObjects(t, tc.existingObjs)
			s, mock, _, err := newServerWithChartsAndReleases(t, nil, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer cleanup()

			for i, existing := range tc.existingObjs {
				if tc.request.GetPaginationOptions().GetPageSize() > 0 {
					pageOffset, err := paginate.PageOffsetFromInstalledRequest(tc.request)
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

				ts2, repo, err := newRepoWithIndex(existing.repoIndex, existing.repoName, existing.repoNamespace, nil, "")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()

				redisKey, bytes, err := s.redisKeyValueForRepo(repo)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				mock.ExpectGet(redisKey).SetVal(string(bytes))
			}

			response, err := s.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{},
				pkgsGRPCv1alpha1.InstalledPackageSummary{},
				pkgsGRPCv1alpha1.InstalledPackageReference{},
				pkgsGRPCv1alpha1.Context{},
				pkgsGRPCv1alpha1.VersionReference{},
				pkgsGRPCv1alpha1.InstalledPackageStatus{},
				pkgsGRPCv1alpha1.PackageAppVersion{},
				pluginsGRPCv1alpha1.Plugin{})
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
	manifest     string
	status       helmrelease.Status
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest
		existingK8sObjs    []testSpecGetInstalledPackages
		targetNamespace    string // this is where installation would actually place artifacts
		existingHelmStubs  []helmReleaseStub
		expectedStatusCode grpccodes.Code
		expectedDetail     *pkgsGRPCv1alpha1.InstalledPackageDetail
	}{
		{
			name: "returns installed package detail when install fails",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_failed,
			},
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_failed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedDetail:     redis_detail_failed,
		},
		{
			name: "returns installed package detail when install is in progress",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending,
			},
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_pending,
			},
			expectedStatusCode: grpccodes.OK,
			expectedDetail:     redis_detail_pending,
		},
		{
			name: "returns installed package detail when install is successful",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedDetail:     redis_detail_completed,
		},
		{
			name: "returns a 404 if the installed package is not found",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: installedRef("dontworrybehappy", "namespace-1"),
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.NotFound,
		},
		{
			name: "returns values and reconciliation options in package detail",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed_with_values_and_reconciliation_options,
			},
			existingHelmStubs: []helmReleaseStub{
				redis_existing_stub_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedDetail:     redis_detail_completed_with_values_and_reconciliation_options,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs, cleanup := newRuntimeObjects(t, tc.existingK8sObjs)
			defer cleanup()
			actionConfig := newHelmActionConfig(t, tc.targetNamespace, tc.existingHelmStubs)
			s, mock, _, err := newServerWithChartsAndReleases(t, actionConfig, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			expectedResp := &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: tc.expectedDetail,
			}

			compareActualVsExpectedGetInstalledPackageDetailResponse(t, response, expectedResp)

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

type testSpecCreateInstalledPackage struct {
	repoName      string
	repoNamespace string
	repoIndex     string
}

func TestCreateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.CreateInstalledPackageRequest
		existingObjs       testSpecCreateInstalledPackage
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.CreateInstalledPackageResponse
		expectedRelease    map[string]interface{}
	}{
		{
			name: "create package (simple)",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "test",
				},
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_basic,
		},
		{
			name: "create package (semver constraint)",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "test",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "> 5",
				},
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_semver_constraint,
		},
		{
			name: "create package (reconcile options)",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "test",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					Interval:           60,
					Suspend:            false,
					ServiceAccountName: "foo",
				},
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_reconcile_options,
		},
		{
			name: "create package (values override)",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "test",
				},
				Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_values,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs := []k8sruntime.Object{}

			ts, repo, err := newRepoWithIndex(
				tc.existingObjs.repoIndex, tc.existingObjs.repoName, tc.existingObjs.repoNamespace, nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts.Close()

			runtimeObjs = append(runtimeObjs, repo)
			s, mock, _, _, err := newServerWithRepos(t, runtimeObjs, nil, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			redisKey, bytes, err := s.redisKeyValueForRepo(repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			mock.ExpectGet(redisKey).SetVal(string(bytes))

			response, err := s.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				pkgsGRPCv1alpha1.CreateInstalledPackageResponse{},
				pkgsGRPCv1alpha1.InstalledPackageReference{},
				pluginsGRPCv1alpha1.Plugin{},
				pkgsGRPCv1alpha1.Context{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmReleass CRD has been created
			_, dynamicClient, _, err = s.clientGetter(context.Background())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			releaseObj, err := dynamicClient.Resource(releasesGvr).Namespace(tc.request.TargetContext.Namespace).Get(
				context.Background(),
				tc.request.Name,
				k8smetav1.GetOptions{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := releaseObj.Object, tc.expectedRelease; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestUpdateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest
		existingK8sObjs    []testSpecGetInstalledPackages
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.UpdateInstalledPackageResponse
		expectedRelease    map[string]interface{}
	}{
		{
			name: "update package (simple)",
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: ">14.4.0",
				},
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
				InstalledPackageRef: my_redis_ref,
			},
			expectedRelease: flux_helm_release_updated_1,
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: installedRef("not-a-valid-identifier", "default"),
			},
			expectedStatusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs, cleanup := newRuntimeObjects(t, tc.existingK8sObjs)
			defer cleanup()
			s, mock, _, err := newServerWithChartsAndReleases(t, nil, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{},
				pkgsGRPCv1alpha1.InstalledPackageReference{},
				pluginsGRPCv1alpha1.Plugin{},
				pkgsGRPCv1alpha1.Context{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmReleass CRD has been updated
			_, dynamicClient, _, err = s.clientGetter(context.Background())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			releaseObj, err := dynamicClient.Resource(releasesGvr).
				Namespace(tc.expectedResponse.InstalledPackageRef.Context.Namespace).Get(
				context.Background(),
				tc.expectedResponse.InstalledPackageRef.Identifier,
				k8smetav1.GetOptions{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := releaseObj.Object, tc.expectedRelease; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestDeleteInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest
		existingK8sObjs    []testSpecGetInstalledPackages
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.DeleteInstalledPackageResponse
	}{
		{
			name: "delete package",
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedStatusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runtimeObjs, cleanup := newRuntimeObjects(t, tc.existingK8sObjs)
			defer cleanup()
			s, mock, _, err := newServerWithChartsAndReleases(t, nil, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmReleass CRD has been updated
			_, dynamicClient, _, err = s.clientGetter(context.Background())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			_, err = dynamicClient.Resource(releasesGvr).
				Namespace(tc.request.InstalledPackageRef.Context.Namespace).Get(
				context.Background(),
				tc.request.InstalledPackageRef.Identifier,
				k8smetav1.GetOptions{})
			if !k8serrors.IsNotFound(err) {
				t.Errorf("mismatch expected, NotFound, got %+v", err)
			}
		})
	}
}

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	// sanity check
	if len(resourcerefstest.TestCases2) < 11 {
		t.Fatalf("Expected array [resourcerefstest.TestCases2] size of at least 11")
		return
	}

	type testCase struct {
		baseTestCase       resourcerefstest.TestCase
		request            *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse
		expectedStatusCode grpccodes.Code
	}

	// newTestCase is a function to take an existing test-case
	// (a so-called baseTestCase in pkg/resourcerefs module, which contains a LOT of useful data)
	// and "enrich" it with some new fields to create a different kind of test case
	// that tests server.GetInstalledPackageResourceRefs() func
	newTestCase := func(tc int, response bool, code grpccodes.Code) testCase {
		// Using the redis_existing_stub_completed data with
		// different manifests for each test.
		var (
			releaseNamespace = redis_existing_stub_completed.namespace
			releaseName      = redis_existing_stub_completed.name
		)

		newCase := testCase{
			baseTestCase: resourcerefstest.TestCases2[tc],
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: releaseNamespace,
					},
					Identifier: releaseName,
				},
			},
		}
		if response {
			newCase.expectedResponse = &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: releaseNamespace,
				},
				ResourceRefs: resourcerefstest.TestCases2[tc].ExpectedResourceRefs,
			}
		}
		newCase.expectedStatusCode = code
		return newCase
	}

	testCases := []testCase{
		newTestCase(0, true, grpccodes.OK),
		newTestCase(1, true, grpccodes.OK),
		newTestCase(2, true, grpccodes.OK),
		newTestCase(3, true, grpccodes.OK),
		newTestCase(4, false, grpccodes.NotFound),
		newTestCase(5, false, grpccodes.Internal),
		// See https://github.com/kubeapps/kubeapps/issues/632
		newTestCase(6, true, grpccodes.OK),
		newTestCase(7, true, grpccodes.OK),
		newTestCase(8, true, grpccodes.OK),
		// See https://k8stypedclient.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
		newTestCase(9, true, grpccodes.OK),
		newTestCase(10, true, grpccodes.OK),
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{},
		pkgsGRPCv1alpha1.ResourceRef{},
		pkgsGRPCv1alpha1.Context{},
	)

	toHelmReleaseStubs := func(in []resourcerefstest.TestReleaseStub) []helmReleaseStub {
		out := []helmReleaseStub{}
		for _, r := range in {
			out = append(out, helmReleaseStub{name: r.Name, namespace: r.Namespace, manifest: r.Manifest})
		}
		return out
	}

	for _, tc := range testCases {
		t.Run(tc.baseTestCase.Name, func(t *testing.T) {
			runtimeObjs, cleanup := newRuntimeObjects(t, []testSpecGetInstalledPackages{redis_existing_spec_completed})
			defer cleanup()
			actionConfig := newHelmActionConfig(
				t,
				tc.request.InstalledPackageRef.GetContext().GetNamespace(),
				toHelmReleaseStubs(tc.baseTestCase.ExistingReleases))
			server, mock, _, err := newServerWithChartsAndReleases(t, actionConfig, runtimeObjs...)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func newRuntimeObjects(t *testing.T, existingK8sObjs []testSpecGetInstalledPackages) (runtimeObjs []k8sruntime.Object, cleanup func()) {
	httpServers := []*httptest.Server{}
	cleanup = func() {
		for _, ts := range httpServers {
			ts.Close()
		}
	}

	for _, existing := range existingK8sObjs {
		tarGzBytes, err := ioutil.ReadFile(existing.chartTarGz)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// stand up an http server just for the duration of this test
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write(tarGzBytes)
		}))
		httpServers = append(httpServers, ts)

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
			"interval": "1m",
			"install": map[string]interface{}{
				"createNamespace": true,
			},
		}
		if len(existing.targetNamespace) != 0 {
			k8smetaunstructuredv1.SetNestedField(releaseSpec, existing.targetNamespace, "targetNamespace")
		}
		if len(existing.releaseValues) != 0 {
			k8smetaunstructuredv1.SetNestedMap(releaseSpec, existing.releaseValues, "values")
		}
		if existing.releaseSuspend {
			k8smetaunstructuredv1.SetNestedField(releaseSpec, existing.releaseSuspend, "suspend")
		}
		if len(existing.releaseServiceAccountName) != 0 {
			k8smetaunstructuredv1.SetNestedField(releaseSpec, existing.releaseServiceAccountName, "serviceAccountName")
		}
		release := newRelease(existing.releaseName, existing.releaseNamespace, releaseSpec, existing.releaseStatus)
		runtimeObjs = append(runtimeObjs, release)
	}
	return runtimeObjs, cleanup
}

func compareActualVsExpectedGetInstalledPackageDetailResponse(t *testing.T, actualResp *pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse, expectedResp *pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse) {
	opts := cmpopts.IgnoreUnexported(
		pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{},
		pkgsGRPCv1alpha1.InstalledPackageDetail{},
		pkgsGRPCv1alpha1.InstalledPackageReference{},
		pkgsGRPCv1alpha1.Context{},
		pkgsGRPCv1alpha1.VersionReference{},
		pkgsGRPCv1alpha1.InstalledPackageStatus{},
		pkgsGRPCv1alpha1.PackageAppVersion{},
		pluginsGRPCv1alpha1.Plugin{},
		pkgsGRPCv1alpha1.ReconciliationOptions{},
		pkgsGRPCv1alpha1.AvailablePackageReference{})
	// see comment in release_integration_test.go. Intermittently we get an inconsistent error message from flux
	opts2 := cmpopts.IgnoreFields(pkgsGRPCv1alpha1.InstalledPackageStatus{}, "UserReason")
	if got, want := actualResp, expectedResp; !cmp.Equal(want, got, opts, opts2) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
	}
	if !strings.Contains(actualResp.InstalledPackageDetail.Status.UserReason, expectedResp.InstalledPackageDetail.Status.UserReason) {
		t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", expectedResp.InstalledPackageDetail.Status.UserReason, actualResp.InstalledPackageDetail.Status.UserReason)
	}
}

func newRelease(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *k8smetaunstructuredv1.Unstructured {
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

	return &k8smetaunstructuredv1.Unstructured{
		Object: obj,
	}
}

func newServerWithChartsAndReleases(t *testing.T, actionConfig *helmaction.Configuration, chartOrRelease ...k8sruntime.Object) (*Server, redismock.ClientMock, *k8swatch.FakeWatcher, error) {
	typedClient := k8stypedclientfake.NewSimpleClientset()
	dynamicClient := k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
		k8sruntime.NewScheme(),
		map[k8sschema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}:                         fluxHelmChartList,
			{Group: fluxHelmReleaseGroup, Version: fluxHelmReleaseVersion, Resource: fluxHelmReleases}: fluxHelmReleaseList,
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}:                   fluxHelmRepositoryList,
		},
		chartOrRelease...)

	apiextIfc := k8sapiextensionsclientfake.NewSimpleClientset(fluxHelmRepositoryCRD)

	clientGetter := func(context.Context) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error) {
		return typedClient, dynamicClient, apiextIfc, nil
	}

	watcher := k8swatch.NewFake()

	// see chart_test.go for explanation
	reactor := dynamicClient.Fake.ReactionChain[0]
	dynamicClient.Fake.PrependReactor("list", fluxHelmRepositories,
		func(action k8stesting.Action) (bool, k8sruntime.Object, error) {
			handled, ret, err := reactor.React(action)
			ulist, ok := ret.(*k8smetaunstructuredv1.UnstructuredList)
			if ok && ulist != nil {
				ulist.SetResourceVersion("1")
			}
			return handled, ret, err
		})

	dynamicClient.Fake.PrependWatchReactor(
		fluxHelmCharts,
		k8stesting.DefaultWatchReactor(watcher, nil))

	s, mock, err := newServer(t, clientGetter, actionConfig, nil, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, mock, watcher, nil
}

// newHelmActionConfig returns an helmaction.Configuration with fake clients and memory helmstorage.
func newHelmActionConfig(t *testing.T, namespace string, rels []helmReleaseStub) *helmaction.Configuration {
	t.Helper()

	memDriver := helmstoragedriver.NewMemory()

	actionConfig := &helmaction.Configuration{
		Releases:     helmstorage.Init(memDriver),
		KubeClient:   &helmkubefake.FailingKubeClient{PrintingKubeClient: helmkubefake.PrintingKubeClient{Out: ioutil.Discard}},
		Capabilities: helmchartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}

	for _, r := range rels {
		config := map[string]interface{}{}
		rel := &helmrelease.Release{
			Name:      r.name,
			Namespace: r.namespace,
			Info: &helmrelease.Info{
				Status: r.status,
				Notes:  r.notes,
			},
			Chart: &helmchart.Chart{
				Metadata: &helmchart.Metadata{
					Version:    r.chartVersion,
					Icon:       "https://example.com/icon.png",
					AppVersion: "1.2.3",
				},
			},
			Config:   config,
			Manifest: r.manifest,
		}
		err := actionConfig.Releases.Create(rel)
		if err != nil {
			t.Fatal(err)
		}
	}
	// It is the namespace of the driver which determines the results. In the prod code,
	// the actionConfigGetter sets this using StorageForSecrets(namespace, clientset).
	memDriver.SetNamespace(namespace)

	return actionConfig
}

func installedRef(id, namespace string) *pkgsGRPCv1alpha1.InstalledPackageReference {
	return &pkgsGRPCv1alpha1.InstalledPackageReference{
		Context: &pkgsGRPCv1alpha1.Context{
			Namespace: namespace,
			Cluster:   KubeappsCluster,
		},
		Identifier: id,
		Plugin:     fluxPlugin,
	}
}

// misc global vars that get re-used in multiple tests scenarios
var (
	releasesGvr = k8sschema.GroupVersionResource{
		Group:    fluxHelmReleaseGroup,
		Version:  fluxHelmReleaseVersion,
		Resource: fluxHelmReleases,
	}

	statusInstalled = &pkgsGRPCv1alpha1.InstalledPackageStatus{
		Ready:      true,
		Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	}

	my_redis_ref = installedRef("my-redis", "namespace-1")

	redis_summary_installed = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           statusInstalled,
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_failed = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      false,
			Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      false,
			Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending_2 = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      false,
			Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "ArtifactFailed: HelmChart 'default/kubeapps-my-redis' is not ready",
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_installed = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "6.7.1",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           statusInstalled,
	}

	redis_summary_latest = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           statusInstalled,
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_semver = &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "<=6.7.1",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           statusInstalled,
	}

	redis_existing_spec_completed = testSpecGetInstalledPackages{
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
			"helmChart":             "default/redis",
			"lastAppliedRevision":   "14.4.0",
			"lastAttemptedRevision": "14.4.0",
		},
		targetNamespace: "test",
	}

	redis_existing_stub_completed = helmReleaseStub{
		name:         "test-my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       helmrelease.StatusDeployed,
	}

	redis_existing_spec_completed_with_values_and_reconciliation_options = testSpecGetInstalledPackages{
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
		targetNamespace: "test",
	}

	redis_existing_spec_failed = testSpecGetInstalledPackages{
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
					"lastTransitionTime": "2021-09-06T10:24:34Z",
					"type":               "Ready",
					"status":             "False",
					"message":            "install retries exhausted",
					"reason":             "InstallFailed",
				},
				map[string]interface{}{
					"lastTransitionTime": "2021-09-06T10:24:34Z",
					"type":               "Released",
					"status":             "False",
					"message":            "Helm install failed: unable to build kubernetes objects from release manifest: error validating \"\": error validating data: ValidationError(Deployment.spec.replicas): invalid type for io.k8s.api.apps.v1.DeploymentSpec.replicas: got \"string\", expected \"integer\"",
					"reason":             "InstallFailed",
				},
			},
			"helmChart":             "default/redis",
			"failures":              "14",
			"installFailures":       "1",
			"lastAttemptedRevision": "14.4.0",
		},
		targetNamespace: "test",
	}

	redis_existing_stub_failed = helmReleaseStub{
		name:         "test-my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       helmrelease.StatusFailed,
	}

	airflow_existing_spec_completed = testSpecGetInstalledPackages{
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
			"helmChart":             "default/airflow",
			"lastAppliedRevision":   "6.7.1",
			"lastAttemptedRevision": "6.7.1",
		},
	}

	airflow_existing_spec_semver = testSpecGetInstalledPackages{
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
			"helmChart":             "default/airflow",
			"lastAppliedRevision":   "6.7.1",
			"lastAttemptedRevision": "6.7.1",
		},
	}

	redis_existing_spec_pending = testSpecGetInstalledPackages{
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
					"status":             "Unknown",
					"reason":             "Progressing",
					"message":            "reconciliation in progress",
				},
			},
			"helmChart":             "default/redis",
			"lastAttemptedRevision": "14.4.0",
		},
		targetNamespace: "test",
	}

	redis_existing_spec_pending_2 = testSpecGetInstalledPackages{
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
					"lastTransitionTime": "2021-09-06T05:26:52Z",
					"message":            "HelmChart 'default/kubeapps-my-redis' is not ready",
					"reason":             "ArtifactFailed",
					"status":             "False",
					"type":               "Ready",
				},
			},
			"failures":              "2",
			"helmChart":             "default/redis",
			"lastAttemptedRevision": "14.4.0",
		},
	}

	redis_existing_stub_pending = helmReleaseStub{
		name:         "test-my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       helmrelease.StatusPendingInstall,
	}

	redis_existing_spec_latest = testSpecGetInstalledPackages{
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
			"helmChart":             "default/redis",
			"lastAppliedRevision":   "14.4.0",
			"lastAttemptedRevision": "14.4.0",
		},
	}

	redis_detail_failed = &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
			Interval: 60,
		},
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      false,
			Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_pending = &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
			Interval: 60,
		},
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      false,
			Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed = &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
			Interval: 60,
		},
		Status:                statusInstalled,
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed_with_values_and_reconciliation_options = &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
			Interval:           60,
			Suspend:            true,
			ServiceAccountName: "foo",
		},
		Status:                statusInstalled,
		ValuesApplied:         "{\"replica\":[{\"configuration\":\"xyz\",\"replicaCount\":\"1\"}]}",
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	flux_helm_release_basic = map[string]interface{}{
		"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
		"kind":       "HelmRelease",
		"metadata": map[string]interface{}{
			"name":      "my-podinfo",
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"chart": "podinfo",
					"sourceRef": map[string]interface{}{
						"kind":      "HelmRepository",
						"name":      "podinfo",
						"namespace": "namespace-1",
					},
				},
			},
			"interval":        "1m",
			"targetNamespace": "test",
		},
	}

	flux_helm_release_semver_constraint = map[string]interface{}{
		"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
		"kind":       "HelmRelease",
		"metadata": map[string]interface{}{
			"name":      "my-podinfo",
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"chart": "podinfo",
					"sourceRef": map[string]interface{}{
						"kind":      "HelmRepository",
						"name":      "podinfo",
						"namespace": "namespace-1",
					},
					"version": "> 5",
				},
			},
			"interval":        "1m",
			"targetNamespace": "test",
		},
	}

	flux_helm_release_reconcile_options = map[string]interface{}{
		"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
		"kind":       "HelmRelease",
		"metadata": map[string]interface{}{
			"name":      "my-podinfo",
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"chart": "podinfo",
					"sourceRef": map[string]interface{}{
						"kind":      "HelmRepository",
						"name":      "podinfo",
						"namespace": "namespace-1",
					},
				},
			},
			"interval":           "1m0s",
			"serviceAccountName": "foo",
			"suspend":            false,
			"targetNamespace":    "test",
		},
	}

	flux_helm_release_values = map[string]interface{}{
		"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
		"kind":       "HelmRelease",
		"metadata": map[string]interface{}{
			"name":      "my-podinfo",
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"chart": "podinfo",
					"sourceRef": map[string]interface{}{
						"kind":      "HelmRepository",
						"name":      "podinfo",
						"namespace": "namespace-1",
					},
				},
			},
			"interval":        "1m",
			"targetNamespace": "test",
			"values": map[string]interface{}{
				"ui": map[string]interface{}{"message": "what we do in the shadows"},
			},
		},
	}

	create_installed_package_resp_my_podinfo = &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
		InstalledPackageRef: installedRef("my-podinfo", "test"),
	}

	flux_helm_release_updated_1 = map[string]interface{}{
		"apiVersion": "helm.toolkit.fluxcd.io/v2beta1",
		"kind":       "HelmRelease",
		"metadata": map[string]interface{}{
			"name":            "my-redis",
			"namespace":       "namespace-1",
			"generation":      int64(1),
			"resourceVersion": "1",
		},
		"spec": map[string]interface{}{
			"chart": map[string]interface{}{
				"spec": map[string]interface{}{
					"chart": "redis",
					"sourceRef": map[string]interface{}{
						"kind":      "HelmRepository",
						"name":      "bitnami-1",
						"namespace": "default",
					},
					"version": ">14.4.0",
				},
			},
			"install": map[string]interface{}{
				"createNamespace": true,
			},
			"interval":        "1m",
			"targetNamespace": "test",
		},
	}
)
