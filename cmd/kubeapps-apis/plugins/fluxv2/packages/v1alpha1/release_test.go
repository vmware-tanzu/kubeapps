// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	releaseValues             *v1.JSON
	releaseSuspend            bool
	releaseServiceAccountName string
	releaseStatus             helmv2.HelmReleaseStatus
	// only used to test edge cases now, most tests should not set this
	targetNamespace string
}

func TestGetInstalledPackageSummariesWithoutPagination(t *testing.T) {
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
				Context: &corev1.Context{Namespace: "test"},
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
				Context: &corev1.Context{Namespace: "test"},
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
			name: "returns installed packages when install is in progress (2)",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "test"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_pending_2,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_pending_2,
				},
			},
		},
		{
			name: "returns installed packages in a specific namespace",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "test"},
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
		{
			// see https://github.com/kubeapps/kubeapps/issues/4189 for discussion
			// this is testing a configuration where a customer has manually set a
			// .targetNamespace field of Flux HelmRelease CR
			name: "returns installed packages when HelmRelease targetNamespace is set",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "test"},
			},
			existingObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_target_ns_is_set,
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					redis_summary_installed,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			charts, releases, cleanup := newChartsAndReleases(t, tc.existingObjs)
			s, mock, err := newServerWithChartsAndReleases(t, nil, charts, releases)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer cleanup()

			for _, existing := range tc.existingObjs {
				ts2, repo, err := newRepoWithIndex(
					existing.repoIndex, existing.repoName, existing.repoNamespace, nil, "")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()

				redisKey, bytes, err := s.redisKeyValueForRepo(*repo)
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
				corev1.PackageAppVersion{},
				plugins.Plugin{})
			opts2 := cmpopts.SortSlices(lessInstalledPackageSummaryFunc)
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts, opts2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetInstalledPackageSummariesWithPagination(t *testing.T) {
	// one big test case that can't really be broken down to smaller cases because
	// the tests aren't independent/idempotent: there is state that needs to be
	// kept track from one call to the next
	t.Run("tests GetInstalledPackageSummaries() pagination", func(t *testing.T) {
		existingObjs := []testSpecGetInstalledPackages{
			redis_existing_spec_completed,
			airflow_existing_spec_completed,
		}
		charts, releases, cleanup := newChartsAndReleases(t, existingObjs)
		s, mock, err := newServerWithChartsAndReleases(t, nil, charts, releases)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer cleanup()

		for _, existing := range existingObjs {
			ts2, repo, err := newRepoWithIndex(
				existing.repoIndex, existing.repoName, existing.repoNamespace, nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			redisKey, bytes, err := s.redisKeyValueForRepo(*repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			mock.ExpectGet(redisKey).SetVal(string(bytes))
		}

		request1 := &corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{Namespace: ""},
			PaginationOptions: &corev1.PaginationOptions{
				PageToken: "0",
				PageSize:  1,
			},
		}
		response1, err := s.GetInstalledPackageSummaries(context.Background(), request1)

		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
		}

		t.Logf("got response: [%s]", response1.InstalledPackageSummaries[0].Name)

		opts := cmpopts.IgnoreUnexported(
			corev1.GetInstalledPackageSummariesResponse{},
			corev1.InstalledPackageSummary{},
			corev1.InstalledPackageReference{},
			corev1.Context{},
			corev1.VersionReference{},
			corev1.InstalledPackageStatus{},
			corev1.PackageAppVersion{},
			plugins.Plugin{})
		opts2 := cmpopts.SortSlices(lessInstalledPackageSummaryFunc)

		expectedResp1 := &corev1.GetInstalledPackageSummariesResponse{
			InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
				redis_summary_installed,
			},
			NextPageToken: "1",
		}
		expectedResp2 := &corev1.GetInstalledPackageSummariesResponse{
			InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
				airflow_summary_installed,
			},
			NextPageToken: "1",
		}

		match := false
		var nextExpectedResp *corev1.GetInstalledPackageSummariesResponse
		if got, want := response1, expectedResp1; cmp.Equal(want, got, opts, opts2) {
			match = true
			nextExpectedResp = expectedResp2
			nextExpectedResp.NextPageToken = "2"
		} else if got, want := response1, expectedResp2; cmp.Equal(want, got, opts, opts2) {
			match = true
			nextExpectedResp = expectedResp1
			nextExpectedResp.NextPageToken = "2"
		}
		if !match {
			t.Fatalf("Expected one of:\n%s\n%s, but got:\n%s", expectedResp1, expectedResp2, response1)
		}

		request2 := &corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{Namespace: ""},
			PaginationOptions: &corev1.PaginationOptions{
				PageSize:  1,
				PageToken: "1",
			},
		}

		response2, err := s.GetInstalledPackageSummaries(context.Background(), request2)

		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
		}
		if got, want := response2, nextExpectedResp; !cmp.Equal(want, got, opts, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
		}

		request3 := &corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{Namespace: ""},
			PaginationOptions: &corev1.PaginationOptions{
				PageSize:  1,
				PageToken: "2",
			},
		}

		nextExpectedResp = &corev1.GetInstalledPackageSummariesResponse{
			InstalledPackageSummaries: []*corev1.InstalledPackageSummary{},
			NextPageToken:             "",
		}

		response3, err := s.GetInstalledPackageSummaries(context.Background(), request3)

		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
		}
		if got, want := response3, nextExpectedResp; !cmp.Equal(want, got, opts, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

type helmReleaseStub struct {
	name         string
	namespace    string
	chartVersion string
	notes        string
	manifest     string
	status       release.Status
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetInstalledPackageDetailRequest
		existingK8sObjs    testSpecGetInstalledPackages
		existingHelmStub   helmReleaseStub
		expectedStatusCode codes.Code
		expectedDetail     *corev1.InstalledPackageDetail
	}{
		{
			name: "returns installed package detail when install fails",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs:    redis_existing_spec_failed,
			existingHelmStub:   redis_existing_stub_failed,
			expectedStatusCode: codes.OK,
			expectedDetail:     redis_detail_failed,
		},
		{
			name: "returns installed package detail when install is in progress",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs:    redis_existing_spec_pending,
			existingHelmStub:   redis_existing_stub_pending,
			expectedStatusCode: codes.OK,
			expectedDetail:     redis_detail_pending,
		},
		{
			name: "returns installed package detail when install is successful",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs:    redis_existing_spec_completed,
			existingHelmStub:   redis_existing_stub_completed,
			expectedStatusCode: codes.OK,
			expectedDetail:     redis_detail_completed,
		},
		{
			name: "returns a 404 if the installed package is not found",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: installedRef("dontworrybehappy", "namespace-1"),
			},
			existingK8sObjs:    redis_existing_spec_completed,
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "returns values and reconciliation options in package detail",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs:    redis_existing_spec_completed_with_values_and_reconciliation_options,
			existingHelmStub:   redis_existing_stub_completed,
			expectedStatusCode: codes.OK,
			expectedDetail:     redis_detail_completed_with_values_and_reconciliation_options,
		},
		{
			// see https://github.com/kubeapps/kubeapps/issues/4189 for discussion
			// this is testing a configuration where a customer has manually set a
			// .targetNamespace field of Flux HelmRelease CR
			name: "returns installed package detail when targetNamespace is set",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs:    redis_existing_spec_target_ns_is_set,
			existingHelmStub:   redis_existing_stub_target_ns_is_set,
			expectedStatusCode: codes.OK,
			expectedDetail:     redis_detail_completed,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			charts, repos, cleanup := newChartsAndReleases(t, []testSpecGetInstalledPackages{tc.existingK8sObjs})
			defer cleanup()
			helmReleaseNamespace := tc.existingK8sObjs.targetNamespace
			if helmReleaseNamespace == "" {
				// this would be most cases now
				helmReleaseNamespace = tc.existingK8sObjs.releaseNamespace
			}
			actionConfig := newHelmActionConfig(
				t, helmReleaseNamespace, []helmReleaseStub{tc.existingHelmStub})
			s, mock, err := newServerWithChartsAndReleases(t, actionConfig, charts, repos)
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

			expectedResp := &corev1.GetInstalledPackageDetailResponse{
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
		request            *corev1.CreateInstalledPackageRequest
		existingObjs       testSpecCreateInstalledPackage
		expectedStatusCode codes.Code
		expectedResponse   *corev1.CreateInstalledPackageResponse
		expectedRelease    *helmv2.HelmRelease
	}{
		{
			name: "create package (simple)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_basic,
		},
		{
			name: "create package (semver constraint)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "> 5",
				},
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_semver_constraint,
		},
		{
			name: "create package (reconcile options)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				ReconciliationOptions: &corev1.ReconciliationOptions{
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
			expectedStatusCode: codes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_reconcile_options,
		},
		{
			name: "create package (values JSON override)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_values,
		},
		{
			name: "create package (values YAML override)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
				Name:                "my-podinfo",
				TargetContext: &corev1.Context{
					Namespace: "test",
				},
				Values: "# Default values for podinfo.\n---\nui:\n  message: what we do in the shadows",
			},
			existingObjs: testSpecCreateInstalledPackage{
				repoName:      "podinfo",
				repoNamespace: "namespace-1",
				repoIndex:     "testdata/podinfo-index.yaml",
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   create_installed_package_resp_my_podinfo,
			expectedRelease:    flux_helm_release_values,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, repo, err := newRepoWithIndex(
				tc.existingObjs.repoIndex, tc.existingObjs.repoName, tc.existingObjs.repoNamespace, nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts.Close()

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			redisKey, bytes, err := s.redisKeyValueForRepo(*repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			mock.ExpectGet(redisKey).SetVal(string(bytes))

			response, err := s.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.CreateInstalledPackageResponse{},
				corev1.InstalledPackageReference{},
				plugins.Plugin{},
				corev1.Context{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmReleass CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(context.Background(), s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				key := types.NamespacedName{Namespace: tc.request.TargetContext.Namespace, Name: tc.request.Name}
				var actualRel helmv2.HelmRelease
				if err = ctrlClient.Get(context.Background(), key, &actualRel); err != nil {
					t.Fatal(err)
				} else {
					// Values are JSON string and need to be compared as such
					opts = cmpopts.IgnoreFields(helmv2.HelmReleaseSpec{}, "Values")

					if got, want := &actualRel, tc.expectedRelease; !cmp.Equal(want, got, opts) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
					}
					compareJSON(t, tc.expectedRelease.Spec.Values, actualRel.Spec.Values)
				}
			}
		})
	}
}

func TestUpdateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.UpdateInstalledPackageRequest
		existingK8sObjs    *testSpecGetInstalledPackages
		expectedStatusCode codes.Code
		expectedResponse   *corev1.UpdateInstalledPackageResponse
		expectedRelease    *helmv2.HelmRelease
	}{
		{
			name: "update package (simple)",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
				PkgVersionReference: &corev1.VersionReference{
					Version: ">14.4.0",
				},
			},
			existingK8sObjs:    &redis_existing_spec_completed,
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: my_redis_ref,
			},
			expectedRelease: flux_helm_release_updated_1,
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: installedRef("not-a-valid-identifier", "default"),
			},
			expectedStatusCode: codes.NotFound,
		},
		{
			// see https://github.com/kubeapps/kubeapps/issues/4189 for discussion
			// this is testing a configuration where a customer has manually set a
			// .targetNamespace field of Flux HelmRelease CR
			name: "updates a package when targetNamespace is set",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
				PkgVersionReference: &corev1.VersionReference{
					Version: ">14.4.0",
				},
			},
			existingK8sObjs:    &redis_existing_spec_target_ns_is_set,
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: my_redis_ref,
			},
			expectedRelease: flux_helm_release_updated_target_ns_is_set,
		},
		{
			name: "update package (values JSON override)",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
				Values:              "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
			},
			existingK8sObjs:    &redis_existing_spec_completed,
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: my_redis_ref,
			},
			expectedRelease: flux_helm_release_updated_2,
		},
		{
			name: "update package (values YAML override)",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
				Values:              "# Default values.\n---\nui:\n  message: what we do in the shadows",
			},
			existingK8sObjs:    &redis_existing_spec_completed,
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: my_redis_ref,
			},
			expectedRelease: flux_helm_release_updated_2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			existingObjs := []testSpecGetInstalledPackages(nil)
			if tc.existingK8sObjs != nil {
				existingObjs = []testSpecGetInstalledPackages{*tc.existingK8sObjs}
			}
			charts, repos, cleanup := newChartsAndReleases(t, existingObjs)
			defer cleanup()
			s, mock, err := newServerWithChartsAndReleases(t, nil, charts, repos)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.UpdateInstalledPackageResponse{},
				corev1.InstalledPackageReference{},
				plugins.Plugin{},
				corev1.Context{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmReleass CRD has been updated
			key := types.NamespacedName{
				Namespace: tc.expectedResponse.InstalledPackageRef.Context.Namespace,
				Name:      tc.expectedResponse.InstalledPackageRef.Identifier,
			}
			ctx := context.Background()
			var actualRel helmv2.HelmRelease
			if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else if err = ctrlClient.Get(ctx, key, &actualRel); err != nil {
				t.Fatal(err)
			}

			// Values are JSON string and need to be compared as such
			opts = cmpopts.IgnoreFields(helmv2.HelmReleaseSpec{}, "Values")

			if got, want := &actualRel, tc.expectedRelease; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
			compareJSON(t, tc.expectedRelease.Spec.Values, actualRel.Spec.Values)
		})
	}
}

func TestDeleteInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.DeleteInstalledPackageRequest
		existingK8sObjs    []testSpecGetInstalledPackages
		expectedStatusCode codes.Code
		expectedResponse   *corev1.DeleteInstalledPackageResponse
	}{
		{
			name: "delete package",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: my_redis_ref,
			},
			existingK8sObjs: []testSpecGetInstalledPackages{
				redis_existing_spec_completed,
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   &corev1.DeleteInstalledPackageResponse{},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "default",
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedStatusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			charts, repos, cleanup := newChartsAndReleases(t, tc.existingK8sObjs)
			defer cleanup()
			s, mock, err := newServerWithChartsAndReleases(t, nil, charts, repos)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.DeleteInstalledPackageResponse{})

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// check expected HelmRelease CRD has been deleted
			key := types.NamespacedName{
				Namespace: tc.request.InstalledPackageRef.Context.Namespace,
				Name:      tc.request.InstalledPackageRef.Identifier,
			}
			ctx := context.Background()
			var actualRel helmv2.HelmRelease
			if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else if err = ctrlClient.Get(ctx, key, &actualRel); !errors.IsNotFound(err) {
				t.Errorf("mismatch expected, NotFound, got %+v", err)
			}
		})
	}
}

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	// sanity check
	if len(resourcerefstest.TestCases2) < 12 {
		t.Fatalf("Expected array [resourcerefstest.TestCases2] size of at least 12")
	}

	type testCase struct {
		baseTestCase       resourcerefstest.TestCase
		request            *corev1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *corev1.GetInstalledPackageResourceRefsResponse
		expectedStatusCode codes.Code
		targetNamespaceSet bool
	}

	// Using the redis_existing_stub_completed data with
	// different manifests for each test.
	var (
		flux_obj_namespace = redis_existing_spec_completed.releaseNamespace
		flux_obj_name      = redis_existing_spec_completed.releaseName
	)

	// newTestCase is a function to take an existing test-case
	// (a so-called baseTestCase in pkg/resourcerefs module, which contains a LOT of useful data)
	// and "enrich" it with some new fields to create a different kind of test case
	// that tests server.GetInstalledPackageResourceRefs() func
	newTestCase := func(tc int, response bool, code codes.Code, targetNamespaceSet bool) testCase {
		newCase := testCase{
			baseTestCase: resourcerefstest.TestCases2[tc],
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: flux_obj_namespace,
					},
					Identifier: flux_obj_name,
				},
			},
		}
		if response {
			newCase.expectedResponse = &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: flux_obj_namespace,
				},
				ResourceRefs: resourcerefstest.TestCases2[tc].ExpectedResourceRefs,
			}
		}
		newCase.expectedStatusCode = code
		newCase.targetNamespaceSet = targetNamespaceSet
		return newCase
	}

	testCases := []testCase{
		newTestCase(0, true, codes.OK, false),
		newTestCase(1, true, codes.OK, false),
		newTestCase(2, true, codes.OK, false),
		newTestCase(3, true, codes.OK, false),
		newTestCase(4, false, codes.NotFound, false),
		newTestCase(5, false, codes.Internal, false),
		// See https://github.com/kubeapps/kubeapps/issues/632
		newTestCase(6, true, codes.OK, false),
		newTestCase(7, true, codes.OK, false),
		newTestCase(8, true, codes.OK, false),
		// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
		newTestCase(9, true, codes.OK, false),
		newTestCase(10, true, codes.OK, false),
		// see https://github.com/kubeapps/kubeapps/issues/4189 for discussion
		// this is testing a configuration where a customer has manually set a
		// .targetNamespace field of Flux HelmRelease CR
		newTestCase(11, true, codes.OK, true),
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageResourceRefsResponse{},
		corev1.ResourceRef{},
		corev1.Context{},
	)

	toHelmReleaseStubs := func(in []resourcerefstest.TestReleaseStub) []helmReleaseStub {
		out := []helmReleaseStub{}
		for _, r := range in {
			s := helmReleaseStub{
				name:      r.Name,
				namespace: r.Namespace,
				manifest:  r.Manifest,
			}
			out = append(out, s)
		}
		return out
	}

	for _, tc := range testCases {
		t.Run(tc.baseTestCase.Name, func(t *testing.T) {
			var spec testSpecGetInstalledPackages
			var helmReleaseNamespace string
			if !tc.targetNamespaceSet {
				spec = redis_existing_spec_completed
				helmReleaseNamespace = flux_obj_namespace
			} else {
				spec = redis_existing_spec_target_ns_is_set
				helmReleaseNamespace = "test2"
			}
			charts, releases, cleanup := newChartsAndReleases(t, []testSpecGetInstalledPackages{spec})
			defer cleanup()
			actionConfig := newHelmActionConfig(
				t,
				helmReleaseNamespace,
				toHelmReleaseStubs(tc.baseTestCase.ExistingReleases))
			server, mock, err := newServerWithChartsAndReleases(t, actionConfig, charts, releases)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
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

func newChartsAndReleases(t *testing.T, existingK8sObjs []testSpecGetInstalledPackages) (charts []sourcev1.HelmChart, releases []helmv2.HelmRelease, cleanup func()) {
	httpServers := []*httptest.Server{}
	cleanup = func() {
		for _, ts := range httpServers {
			ts.Close()
		}
	}
	charts = []sourcev1.HelmChart{}
	releases = []helmv2.HelmRelease{}

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

		chartSpec := &sourcev1.HelmChartSpec{
			Chart: existing.chartName,
			SourceRef: sourcev1.LocalHelmChartSourceReference{
				Name: existing.repoName,
				Kind: sourcev1.HelmRepositoryKind,
			},
			Version:  existing.chartSpecVersion,
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		}

		lastTransitionTime, err := time.Parse(time.RFC3339, "2021-08-12T03:25:38Z")
		if err != nil {
			t.Fatalf("%v", err)
		}

		chartStatus := &sourcev1.HelmChartStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Message:            "Fetched revision: " + existing.chartSpecVersion,
					Type:               "Ready",
					Status:             "True",
					Reason:             sourcev1.ChartPullSucceededReason,
				},
			},
			Artifact: &sourcev1.Artifact{
				Revision: existing.chartArtifactVersion,
			},
			URL: ts.URL,
		}
		chart := newChart(existing.chartName, existing.repoNamespace, chartSpec, chartStatus)
		charts = append(charts, chart)

		releaseSpec := &helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   existing.chartName,
					Version: existing.chartSpecVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Name:      existing.repoName,
						Kind:      sourcev1.HelmRepositoryKind,
						Namespace: existing.repoNamespace,
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		}
		if existing.targetNamespace != "" {
			// now this is only used when Flux CRs are not created by kubeapps
			releaseSpec.TargetNamespace = existing.targetNamespace
		}
		if existing.releaseValues != nil {
			releaseSpec.Values = existing.releaseValues
		}
		if existing.releaseSuspend {
			releaseSpec.Suspend = existing.releaseSuspend
		}
		if len(existing.releaseServiceAccountName) != 0 {
			releaseSpec.ServiceAccountName = existing.releaseServiceAccountName
		}

		release := newRelease(existing.releaseName, existing.releaseNamespace, releaseSpec, &existing.releaseStatus)
		releases = append(releases, release)
	}
	return charts, releases, cleanup
}

func compareActualVsExpectedGetInstalledPackageDetailResponse(t *testing.T, actualResp *corev1.GetInstalledPackageDetailResponse, expectedResp *corev1.GetInstalledPackageDetailResponse) {
	opts := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageDetailResponse{},
		corev1.InstalledPackageDetail{},
		corev1.InstalledPackageReference{},
		corev1.Context{},
		corev1.VersionReference{},
		corev1.InstalledPackageStatus{},
		corev1.PackageAppVersion{},
		plugins.Plugin{},
		corev1.ReconciliationOptions{},
		corev1.AvailablePackageReference{})
	// see comment in release_integration_test.go. Intermittently we get an inconsistent error message from flux
	opts2 := cmpopts.IgnoreFields(corev1.InstalledPackageStatus{}, "UserReason")
	// Values Applied are JSON string and need to be compared as such
	opts3 := cmpopts.IgnoreFields(corev1.InstalledPackageDetail{}, "ValuesApplied")
	if got, want := actualResp, expectedResp; !cmp.Equal(want, got, opts, opts2, opts3) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2, opts3))
	}
	if !strings.Contains(actualResp.InstalledPackageDetail.Status.UserReason, expectedResp.InstalledPackageDetail.Status.UserReason) {
		t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", expectedResp.InstalledPackageDetail.Status.UserReason, actualResp.InstalledPackageDetail.Status.UserReason)
	}
	compareJSONStrings(t, expectedResp.InstalledPackageDetail.ValuesApplied, actualResp.InstalledPackageDetail.ValuesApplied)
}

func newRelease(name string, namespace string, spec *helmv2.HelmReleaseSpec, status *helmv2.HelmReleaseStatus) helmv2.HelmRelease {
	helmRelease := helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Generation: int64(1),
		},
	}
	if namespace != "" {
		helmRelease.ObjectMeta.Namespace = namespace
	}

	if spec != nil {
		helmRelease.Spec = *spec.DeepCopy()
	}

	if status != nil {
		helmRelease.Status = *status.DeepCopy()
		helmRelease.Status.ObservedGeneration = int64(1)
	}
	return helmRelease
}

func newServerWithChartsAndReleases(t *testing.T, actionConfig *action.Configuration, charts []sourcev1.HelmChart, releases []helmv2.HelmRelease) (*Server, redismock.ClientMock, error) {
	apiextIfc := apiextfake.NewSimpleClientset(fluxHelmRepositoryCRD)

	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	rm := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{sourcev1.GroupVersion, helmv2.GroupVersion})
	rm.Add(schema.GroupVersionKind{
		Group:   sourcev1.GroupVersion.Group,
		Version: sourcev1.GroupVersion.Version,
		Kind:    sourcev1.HelmRepositoryKind},
		apimeta.RESTScopeNamespace)
	rm.Add(schema.GroupVersionKind{
		Group:   sourcev1.GroupVersion.Group,
		Version: sourcev1.GroupVersion.Version,
		Kind:    sourcev1.HelmChartKind},
		apimeta.RESTScopeNamespace)
	rm.Add(schema.GroupVersionKind{
		Group:   helmv2.GroupVersion.Group,
		Version: helmv2.GroupVersion.Version,
		Kind:    helmv2.HelmReleaseKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	if len(charts) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(&sourcev1.HelmChartList{Items: charts})
	}
	if len(releases) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(&helmv2.HelmReleaseList{Items: releases})
	}
	ctrlClient := &withWatchWrapper{delegate: ctrlClientBuilder.Build()}

	clientGetter := func(context.Context, string) (clientgetter.ClientInterfaces, error) {
		return clientgetter.
			NewBuilder().
			WithApiExt(apiextIfc).
			WithControllerRuntime(ctrlClient).
			Build(), nil
	}

	return newServer(t, clientGetter, actionConfig, nil, nil)
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

// misc global vars that get re-used in multiple tests scenarios
var (
	releasesGvr = schema.GroupVersionResource{
		Group:    helmv2.GroupVersion.Group,
		Version:  helmv2.GroupVersion.Version,
		Resource: fluxHelmReleases,
	}

	statusInstalled = &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	}

	my_redis_ref = installedRef("my-redis", "test")

	redis_summary_installed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           statusInstalled,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_failed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending_2 = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "ArtifactFailed: HelmChart 'default/kubeapps-my-redis' is not ready",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_installed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.7.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           statusInstalled,
	}

	redis_summary_latest = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           statusInstalled,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_semver = &corev1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "<=6.7.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           statusInstalled,
	}

	lastTransitionTime, _ = time.Parse(time.RFC3339, "2021-08-11T08:46:03Z")

	redis_existing_spec_completed = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            "testdata/redis-many-versions.yaml",
		chartName:            "redis",
		chartTarGz:           "testdata/redis-14.4.0.tgz",
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_completed = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusDeployed,
	}

	redis_existing_spec_completed_with_values_and_reconciliation_options_values_bytes, _ = json.Marshal(
		map[string]interface{}{
			"replica": map[string]interface{}{
				"replicaCount":  "1",
				"configuration": "xyz",
			}})

	redis_existing_spec_completed_with_values_and_reconciliation_options = testSpecGetInstalledPackages{
		repoName:                  "bitnami-1",
		repoNamespace:             "default",
		repoIndex:                 "testdata/redis-many-versions.yaml",
		chartName:                 "redis",
		chartTarGz:                "testdata/redis-14.4.0.tgz",
		chartSpecVersion:          "14.4.0",
		chartArtifactVersion:      "14.4.0",
		releaseName:               "my-redis",
		releaseNamespace:          "test",
		releaseSuspend:            true,
		releaseServiceAccountName: "foo",
		releaseValues:             &v1.JSON{Raw: redis_existing_spec_completed_with_values_and_reconciliation_options_values_bytes},
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
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
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "False",
					Reason:             helmv2.InstallFailedReason,
					Message:            "install retries exhausted",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "False",
					Reason:             helmv2.InstallFailedReason,
					Message:            "Helm install failed: unable to build kubernetes objects from release manifest: error validating \"\": error validating data: ValidationError(Deployment.spec.replicas): invalid type for io.k8s.api.apps.v1.DeploymentSpec.replicas: got \"string\", expected \"integer\"",
				},
			},
			HelmChart:             "default/redis",
			Failures:              14,
			InstallFailures:       1,
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_failed = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusFailed,
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
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/airflow",
			LastAppliedRevision:   "6.7.1",
			LastAttemptedRevision: "6.7.1",
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
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/airflow",
			LastAppliedRevision:   "6.7.1",
			LastAttemptedRevision: "6.7.1",
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
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "Unknown",
					Reason:             "Progressing",
					Message:            "reconciliation in progress",
				},
			},
			HelmChart:             "default/redis",
			LastAttemptedRevision: "14.4.0",
		},
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
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "False",
					Reason:             helmv2.ArtifactFailedReason,
					Message:            "HelmChart 'default/kubeapps-my-redis' is not ready",
				},
			},
			HelmChart:             "default/redis",
			Failures:              2,
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_pending = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusPendingInstall,
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
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_detail_failed = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_pending = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &corev1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status:                statusInstalled,
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed_with_values_and_reconciliation_options = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &corev1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            true,
			ServiceAccountName: "foo",
		},
		Status:                statusInstalled,
		ValuesApplied:         "{\"replica\": { \"replicaCount\":  \"1\", \"configuration\": \"xyz\"    }}",
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	flux_helm_release_basic = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_semver_constraint = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
					Version: "> 5",
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_reconcile_options = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
				},
			},
			Interval:           metav1.Duration{Duration: 1 * time.Minute},
			ServiceAccountName: "foo",
			Suspend:            false,
		},
	}

	flux_helm_release_values_values_bytes, _ = json.Marshal(
		map[string]interface{}{
			"ui": map[string]interface{}{
				"message": "what we do in the shadows",
			}})

	flux_helm_release_values = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
			Values:   &v1.JSON{Raw: flux_helm_release_values_values_bytes},
		},
	}

	create_installed_package_resp_my_podinfo = &corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: installedRef("my-podinfo", "test"),
	}

	flux_helm_release_updated_1 = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
					Version: ">14.4.0",
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_updated_2 = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
			Values:   &v1.JSON{Raw: flux_helm_release_values_values_bytes},
		},
	}

	redis_existing_spec_target_ns_is_set = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            "testdata/redis-many-versions.yaml",
		chartName:            "redis",
		chartTarGz:           "testdata/redis-14.4.0.tgz",
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
		targetNamespace: "test2",
	}

	redis_existing_stub_target_ns_is_set = helmReleaseStub{
		name:         "test2-my-redis",
		namespace:    "test2",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusDeployed,
	}

	flux_helm_release_updated_target_ns_is_set = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
					Version: ">14.4.0",
				},
			},
			Interval:        metav1.Duration{Duration: 1 * time.Minute},
			TargetNamespace: "test2",
		},
	}
)
