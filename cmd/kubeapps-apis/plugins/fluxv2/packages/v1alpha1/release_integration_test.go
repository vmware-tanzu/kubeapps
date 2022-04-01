// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	rbacv1 "k8s.io/api/rbac/v1"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
// 3) run './kind-cluster-setup.sh deploy' once prior to these tests

type integrationTestCreatePackageSpec struct {
	testName          string
	repoUrl           string
	repoInterval      time.Duration // 0 for default (10m)
	request           *corev1.CreateInstalledPackageRequest
	expectedDetail    *corev1.InstalledPackageDetail
	expectedPodPrefix string
	// what follows are boolean flags to test various negative scenarios
	// different from expectedStatusCode due to async nature of install
	expectInstallFailure bool
	noPreCreateNs        bool
	noCleanup            bool
	expectedStatusCode   codes.Code
	expectedResourceRefs []*corev1.ResourceRef
}

func TestKindClusterCreateInstalledPackage(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []integrationTestCreatePackageSpec{
		{
			testName:             "create test (simplest case)",
			repoUrl:              podinfo_repo_url,
			request:              create_request_basic,
			expectedDetail:       expected_detail_basic,
			expectedPodPrefix:    "my-podinfo-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_basic,
		},
		{
			testName:             "create package (semver constraint)",
			repoUrl:              podinfo_repo_url,
			request:              create_request_semver_constraint,
			expectedDetail:       expected_detail_semver_constraint,
			expectedPodPrefix:    "my-podinfo-2-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_semver_constraint,
		},
		{
			testName:             "create package (reconcile options)",
			repoUrl:              podinfo_repo_url,
			request:              create_request_reconcile_options,
			expectedDetail:       expected_detail_reconcile_options,
			expectedPodPrefix:    "my-podinfo-3-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_reconcile_options,
		},
		{
			testName:             "create package (with values)",
			repoUrl:              podinfo_repo_url,
			request:              create_request_with_values,
			expectedDetail:       expected_detail_with_values,
			expectedPodPrefix:    "my-podinfo-4-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_with_values,
		},
		{
			testName:             "install fails",
			repoUrl:              podinfo_repo_url,
			request:              create_request_install_fails,
			expectedDetail:       expected_detail_install_fails,
			expectInstallFailure: true,
			expectedStatusCode:   codes.OK,
		},
		{
			testName:           "unauthorized",
			repoUrl:            podinfo_repo_url,
			request:            create_request_basic,
			expectedStatusCode: codes.Unauthenticated,
		},
		{
			testName:           "wrong cluster",
			repoUrl:            podinfo_repo_url,
			request:            create_request_wrong_cluster,
			expectedStatusCode: codes.Unimplemented,
		},
		{
			testName:           "target namespace does not exist",
			repoUrl:            podinfo_repo_url,
			request:            create_request_target_ns_doesnt_exist,
			noPreCreateNs:      true,
			expectedStatusCode: codes.NotFound,
		},
	}

	grpcContext, err := newGrpcAdminContext(t, "test-create-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcContext)
		})
	}
}

type integrationTestUpdatePackageSpec struct {
	integrationTestCreatePackageSpec
	request *corev1.UpdateInstalledPackageRequest
	// this is expected AFTER the update call completes
	expectedDetailAfterUpdate *corev1.InstalledPackageDetail
	expectedRefsAfterUpdate   []*corev1.ResourceRef
	unauthorized              bool
}

func TestKindClusterUpdateInstalledPackage(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []integrationTestUpdatePackageSpec{
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (simplest case)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_5_2_1,
				expectedDetail:       expected_detail_podinfo_5_2_1,
				expectedPodPrefix:    "my-podinfo-6-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1,
			},
			request:                   update_request_1,
			expectedDetailAfterUpdate: expected_detail_podinfo_6_0_0,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (add values)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_5_2_1_no_values,
				expectedDetail:       expected_detail_podinfo_5_2_1_no_values,
				expectedPodPrefix:    "my-podinfo-7-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_no_values,
			},
			request:                   update_request_2,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_no_values,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (change values)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_5_2_1_values_2,
				expectedDetail:       expected_detail_podinfo_5_2_1_values_2,
				expectedPodPrefix:    "my-podinfo-8-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_2,
			},
			request:                   update_request_3,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_3,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_2,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (remove values)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_5_2_1_values_4,
				expectedDetail:       expected_detail_podinfo_5_2_1_values_4,
				expectedPodPrefix:    "my-podinfo-9-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_4,
			},
			request:                   update_request_4,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_5,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_4,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (values dont change)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_5_2_1_values_6,
				expectedDetail:       expected_detail_podinfo_5_2_1_values_6,
				expectedPodPrefix:    "my-podinfo-10-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_6,
			},
			request:                   update_request_5,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_6,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_6,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update unauthorized test",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_7,
				expectedDetail:       expected_detail_podinfo_7,
				expectedPodPrefix:    "my-podinfo-11-",
				expectedResourceRefs: expected_resource_refs_podinfo_7,
			},
			request:      update_request_6,
			unauthorized: true,
		},
	}

	grpcContext, err := newGrpcAdminContext(t, "test-create-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			installedRef := createAndWaitForHelmRelease(
				t, tc.integrationTestCreatePackageSpec, fluxPluginClient, grpcContext)
			tc.request.InstalledPackageRef = installedRef

			ctx := grpcContext
			if tc.unauthorized {
				ctx = context.TODO()
			}
			_, err := fluxPluginClient.UpdateInstalledPackage(ctx, tc.request)
			if tc.unauthorized {
				if status.Code(err) != codes.Unauthenticated {
					t.Fatalf("Expected Unathenticated, got: %v", status.Code(err))
				}
				return // done, nothing more to check
			} else if err != nil {
				t.Fatalf("%+v", err)
			}

			actualRespAfterUpdate, actualRefsAfterUpdate :=
				waitUntilInstallCompletes(t, fluxPluginClient, grpcContext, installedRef, false)

			tc.expectedDetailAfterUpdate.InstalledPackageRef = installedRef
			tc.expectedDetailAfterUpdate.Name = tc.integrationTestCreatePackageSpec.request.Name
			tc.expectedDetailAfterUpdate.ReconciliationOptions = &corev1.ReconciliationOptions{
				Interval: 60,
			}
			tc.expectedDetailAfterUpdate.AvailablePackageRef = tc.integrationTestCreatePackageSpec.request.AvailablePackageRef
			tc.expectedDetailAfterUpdate.PostInstallationNotes = strings.ReplaceAll(
				tc.expectedDetailAfterUpdate.PostInstallationNotes,
				"@TARGET_NS@",
				tc.integrationTestCreatePackageSpec.request.TargetContext.Namespace)

			expectedResp := &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: tc.expectedDetailAfterUpdate,
			}

			compareActualVsExpectedGetInstalledPackageDetailResponse(t, actualRespAfterUpdate, expectedResp)

			if tc.expectedRefsAfterUpdate != nil {
				expectedRefsCopy := []*corev1.ResourceRef{}
				for _, r := range tc.expectedRefsAfterUpdate {
					newR := &corev1.ResourceRef{
						ApiVersion: r.ApiVersion,
						Kind:       r.Kind,
						Name:       strings.ReplaceAll(r.Name, "@TARGET_NS@", tc.integrationTestCreatePackageSpec.request.TargetContext.Namespace),
						Namespace:  tc.integrationTestCreatePackageSpec.request.TargetContext.Namespace,
					}
					expectedRefsCopy = append(expectedRefsCopy, newR)
				}
				opts := cmpopts.IgnoreUnexported(corev1.ResourceRef{})
				if got, want := actualRefsAfterUpdate.ResourceRefs, expectedRefsCopy; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}

func TestKindClusterAutoUpdateInstalledPackage(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	spec := integrationTestCreatePackageSpec{
		testName:             "create test (auto update)",
		repoUrl:              podinfo_repo_url,
		repoInterval:         30 * time.Second,
		request:              create_request_auto_update,
		expectedDetail:       expected_detail_auto_update,
		expectedPodPrefix:    "my-podinfo-16",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_auto_update,
	}

	grpcContext, err := newGrpcAdminContext(t, "test-auto-update", "default")
	if err != nil {
		t.Fatal(err)
	}

	// this will also make sure that response looks like expected_detail_auto_update
	installedRef := createAndWaitForHelmRelease(t, spec, fluxPluginClient, grpcContext)
	podName, err := getFluxPluginTestdataPodName()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("podName = [%s]", podName)

	if err = kubeCopyFileToPod(
		t,
		testTgz("podinfo-6.0.3.tgz"),
		*podName,
		"/usr/share/nginx/html/podinfo/podinfo-6.0.3.tgz"); err != nil {
		t.Fatal(err)
	}
	if err = kubeCopyFileToPod(
		t,
		testYaml("podinfo-index-updated.yaml"),
		*podName,
		"/usr/share/nginx/html/podinfo/index.yaml"); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err = kubeCopyFileToPod(
			t,
			testYaml("podinfo-index.yaml"),
			*podName,
			"/usr/share/nginx/html/podinfo/index.yaml"); err != nil {
			t.Logf("Error reverting to previos podinfo index: %v", err)
		}
	})
	t.Logf("Waiting 45 seconds...")
	time.Sleep(45 * time.Second)

	resp, err := fluxPluginClient.GetInstalledPackageDetail(
		grpcContext, &corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	}
	expected_detail_auto_update_2.InstalledPackageRef = installedRef
	expected_detail_auto_update_2.PostInstallationNotes = strings.ReplaceAll(
		expected_detail_auto_update_2.PostInstallationNotes,
		"@TARGET_NS@",
		spec.request.TargetContext.Namespace)
	compareActualVsExpectedGetInstalledPackageDetailResponse(
		t, resp, &corev1.GetInstalledPackageDetailResponse{
			InstalledPackageDetail: expected_detail_auto_update_2,
		})
}

type integrationTestDeletePackageSpec struct {
	integrationTestCreatePackageSpec
	unauthorized bool
}

func TestKindClusterDeleteInstalledPackage(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []integrationTestDeletePackageSpec{
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "delete test (simplest case)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_for_delete_1,
				expectedDetail:       expected_detail_podinfo_for_delete_1,
				expectedPodPrefix:    "my-podinfo-12-",
				expectedResourceRefs: expected_resource_refs_for_delete_1,
				noCleanup:            true,
			},
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "delete test (unauthorized)",
				repoUrl:              podinfo_repo_url,
				request:              create_request_podinfo_for_delete_2,
				expectedDetail:       expected_detail_podinfo_for_delete_2,
				expectedPodPrefix:    "my-podinfo-13-",
				expectedResourceRefs: expected_resource_refs_for_delete_2,
				noCleanup:            true,
			},
			unauthorized: true,
		},
	}

	grpcContext, err := newGrpcAdminContext(t, "test-delete-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(t, tc.integrationTestCreatePackageSpec, fluxPluginClient, grpcContext)

			ctx := grpcContext
			if tc.unauthorized {
				ctx = context.TODO()
			}
			_, err := fluxPluginClient.DeleteInstalledPackage(ctx, &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: installedRef,
			})
			if tc.unauthorized {
				if status.Code(err) != codes.Unauthenticated {
					t.Fatalf("Expected Unathenticated, got: %v", status.Code(err))
				}
				// still need to delete the release though
				if err = kubeDeleteHelmRelease(t, installedRef.Identifier, installedRef.Context.Namespace); err != nil {
					t.Logf("Failed to delete helm release due to %v", err)
				}
				return // done, nothing more to check
			} else if err != nil {
				t.Fatalf("%+v", err)
			}

			const maxWait = 25
			for i := 0; i <= maxWait; i++ {
				grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
				defer cancel()

				_, err := fluxPluginClient.GetInstalledPackageDetail(
					grpcContext, &corev1.GetInstalledPackageDetailRequest{
						InstalledPackageRef: installedRef,
					})
				if err != nil {
					if status.Code(err) == codes.NotFound {
						break // this is the only way to break out of this loop successfully
					} else {
						t.Fatalf("%+v", err)
					}
				}
				if i == maxWait {
					t.Fatalf("Timed out waiting for delete of installed package [%s], last error: [%v]", installedRef, err)
				} else {
					t.Logf("Waiting 1s for package [%s] to be deleted, attempt [%d/%d]...", installedRef, i+1, maxWait)
					time.Sleep(1 * time.Second)
				}
			}

			// confidence test
			exists, err := kubeExistsHelmRelease(t, installedRef.Identifier, installedRef.Context.Namespace)
			if err != nil {
				t.Fatalf("%+v", err)
			} else if exists {
				t.Fatalf("helmrelease [%s] still exists", installedRef)
			}

			// flux is supposed to clean up or "garbage collect" artifacts created by the release,
			// in the targetNamespace, except the namespace itself. Wait to make sure this is done
			// (https://fluxcd.io/docs/components/helm/) it clearly says: Prunes Helm releases removed
			// from cluster (garbage collection)
			for i := 0; i <= maxWait; i++ {
				if pods, err := kubeGetPodNames(t, tc.request.TargetContext.Namespace); err != nil {
					t.Fatalf("%+v", err)
				} else if len(pods) == 0 {
					break
				} else if len(pods) != 1 {
					t.Errorf("expected 1 pod, got: %s", pods)
				} else if !strings.HasPrefix(pods[0], tc.expectedPodPrefix) {
					t.Errorf("expected pod with prefix [%s] not found in namespace [%s], pods found: [%v]",
						tc.expectedPodPrefix, tc.request.TargetContext.Namespace, pods)
				} else if i == maxWait {
					t.Fatalf("Timed out waiting for garbage collection, of [%s], last error: [%v]", pods[0], err)
				} else {
					t.Logf("Waiting 2s for garbage collection of [%s], attempt [%d/%d]...", pods[0], i+1, maxWait)
					time.Sleep(2 * time.Second)
				}
			}
		})
	}
}

// scenario:
// 1) create new namespace ns1
// 2) add podinfo repo in ns1
// 3) create new namespace ns2
// 4) create these service-accounts in default namespace:
//   a) - "...-admin", with cluster-wide access to everything
//   b) - "...-loser", without cluster-wide access or any access to any of the namespaces
//   c) - "...-helmreleases", with only permissions to 'get' HelmReleases in ns2
//   d) - "...-helmreleases-and-charts", with only permissions to 'get' HelmCharts in ns1, HelmReleases in ns2
// 5) as user 4a) install package podinfo in ns2
// 6) verify GetInstalledPackageSummaries:
//    a) as 4a) returns 1 result
//    b) as 4b) raises PermissionDenied error
//    c) as 4c) returns 1 result but without the corresponding chart details
//    d) as 4d) returns 1 result with details from corresponding chart
// 7) verify GetInstalledPackageDetail:
//    a) as 4a) returns full detail
//    b) as 4b) returns PermissionDenied error
//    c) as 4c) returns full detail
//    d) as 4d) returns full detail
// 8) verify GetInstalledPackageResourceRefs:
//    a) as 4a) returns all refs
//    b) as 4b) returns PermissionDenied error
//    c) as 4c) returns all refs
// ref https://github.com/vmware-tanzu/kubeapps/issues/4390
func TestKindClusterRBAC_ReadRelease(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespace(t, ns1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteNamespace(t, ns1); err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", ns1, err)
		}
	})

	grpcCtxAdmin, err := newGrpcAdminContext(t, "test-release-rbac-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, "test-release-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	out := kubectlCanI(
		t, "test-release-rbac-admin", "default", "get", "helmcharts", ns1)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}
	out = kubectlCanI(
		t, "test-release-rbac-loser", "default", "get", "helmcharts", ns1)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	tc := integrationTestCreatePackageSpec{
		testName: "test chart RBAC",
		repoUrl:  podinfo_repo_url,
		request: &corev1.CreateInstalledPackageRequest{
			AvailablePackageRef: availableRef("podinfo-1/podinfo", ns1),
			Name:                "my-podinfo",
			TargetContext: &corev1.Context{
				// note that Namespace is just the prefix - the actual name will
				// have a random string appended at the end, e.g. "test-ns2-h23r"
				// this will happen during the running of the test
				Namespace: "test-ns2",
				Cluster:   KubeappsCluster,
			},
		},
		expectedDetail:       expected_detail_test_release_rbac,
		expectedPodPrefix:    "my-podinfo-",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_basic,
	}

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	out = kubectlCanI(
		t, "test-release-rbac-admin", "default", "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel := context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtx,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})
	if err != nil {
		t.Fatal(err)
	} else if len(resp.InstalledPackageSummaries) != 1 {
		t.Errorf("Unexpected response: %s", common.PrettyPrint(resp))
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
	defer cancel()

	resp2, err := fluxPluginClient.GetInstalledPackageDetail(
		grpcCtx,
		&corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	} else {
		expected_detail_test_release_rbac_2.InstalledPackageDetail.InstalledPackageRef = installedRef
		expected_detail_test_release_rbac_2.InstalledPackageDetail.PostInstallationNotes = strings.ReplaceAll(
			expected_detail_test_release_rbac_2.InstalledPackageDetail.PostInstallationNotes,
			"@TARGET_NS@", ns2)
		expected_detail_test_release_rbac_2.InstalledPackageDetail.AvailablePackageRef.Context.Namespace = ns1
		compareActualVsExpectedGetInstalledPackageDetailResponse(t, resp2, expected_detail_test_release_rbac_2)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
	defer cancel()

	expectedRefsCopy := []*corev1.ResourceRef{}
	for _, r := range expected_resource_refs_basic {
		newR := &corev1.ResourceRef{
			ApiVersion: r.ApiVersion,
			Kind:       r.Kind,
			Name:       r.Name,
			Namespace:  ns2,
		}
		expectedRefsCopy = append(expectedRefsCopy, newR)
	}

	resp3, err := fluxPluginClient.GetInstalledPackageResourceRefs(
		grpcCtx,
		&corev1.GetInstalledPackageResourceRefsRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	} else {
		opts := cmpopts.IgnoreUnexported(corev1.ResourceRef{})
		if got, want := resp3.ResourceRefs, expectedRefsCopy; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
	}

	out = kubectlCanI(
		t, "test-release-rbac-loser", "default", "get", fluxHelmReleases, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtx,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.GetInstalledPackageDetail(
		grpcCtx,
		&corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.GetInstalledPackageResourceRefs(
		grpcCtx,
		&corev1.GetInstalledPackageResourceRefsRequest{
			InstalledPackageRef: installedRef,
		})
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	rules := map[string][]rbacv1.PolicyRule{
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"get", "list"},
			},
			// a little weird but currently this is required too:
			// without it when calling GetInstalledPackageDetail() you will get
			// error "Failed to get helm release due to rpc
			// error: code = NotFound desc = Unable to run Helm Get action for release
			// [test-ns2-8ynh/my-podinfo] in namespace [test-ns2-8ynh]: query: failed to query with
			// labels: secrets is forbidden: User "system:serviceaccount:default:test-release-rbac-helmreleases"
			// cannot list resource "secrets" in API group "" in the namespace "test-ns2-8ynh"
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
		},
	}

	grpcCtxReadHelmReleases, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases", "default", rules)
	if err != nil {
		t.Fatal(err)
	}

	out = kubectlCanI(
		t, "test-release-rbac-helmreleases", "default", "get", fluxHelmRepositories, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanI(
		t, "test-release-rbac-helmreleases", "default", "get", "helmcharts", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanI(
		t, "test-release-rbac-helmreleases", "default", "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleases, defaultContextTimeout)
	defer cancel()

	resp, err = fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtx,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})

	opts2 := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageSummariesResponse{},
		corev1.InstalledPackageSummary{},
		corev1.InstalledPackageReference{},
		corev1.InstalledPackageStatus{},
		plugins.Plugin{},
		corev1.VersionReference{},
		corev1.PackageAppVersion{},
		corev1.Context{})

	if err != nil {
		t.Fatal(err)
	} else {
		// should return installed package summaries without chart details
		expected_summaries_test_release_rbac_1.InstalledPackageSummaries[0].InstalledPackageRef.Context.Namespace = ns2
		if got, want := resp, expected_summaries_test_release_rbac_1; !cmp.Equal(want, got, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts2))
		}
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleases, defaultContextTimeout)
	defer cancel()

	resp2, err = fluxPluginClient.GetInstalledPackageDetail(
		grpcCtx,
		&corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	} else {
		compareActualVsExpectedGetInstalledPackageDetailResponse(t, resp2, expected_detail_test_release_rbac_2)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleases, defaultContextTimeout)
	defer cancel()

	resp3, err = fluxPluginClient.GetInstalledPackageResourceRefs(
		grpcCtx,
		&corev1.GetInstalledPackageResourceRefsRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	} else {
		opts := cmpopts.IgnoreUnexported(corev1.ResourceRef{})
		if got, want := resp3.ResourceRefs, expectedRefsCopy; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
	}

	nsToRules := map[string][]rbacv1.PolicyRule{
		ns1: {
			{
				APIGroups: []string{sourcev1.GroupVersion.Group},
				Resources: []string{"helmcharts"},
				Verbs:     []string{"get", "list"},
			},
		},
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"get", "list"},
			},
			{ // see comment above
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list"},
			},
		},
	}

	grpcCtxReadHelmReleasesAndCharts, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-and-charts", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	out = kubectlCanI(
		t, "test-release-rbac-helmreleases-and-charts", "default", "get", "helmcharts", ns1)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}
	out = kubectlCanI(
		t, "test-release-rbac-helmreleases-and-charts", "default", "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleasesAndCharts, defaultContextTimeout)
	defer cancel()

	resp, err = fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtx,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})
	if err != nil {
		t.Fatal(err)
	} else {
		// should return installed package summaries with chart details
		expected_summaries_test_release_rbac_2.InstalledPackageSummaries[0].InstalledPackageRef.Context.Namespace = ns2
		if got, want := resp, expected_summaries_test_release_rbac_2; !cmp.Equal(want, got, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts2))
		}
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleasesAndCharts, defaultContextTimeout)
	defer cancel()

	resp2, err = fluxPluginClient.GetInstalledPackageDetail(
		grpcCtx,
		&corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	} else {
		compareActualVsExpectedGetInstalledPackageDetailResponse(t, resp2, expected_detail_test_release_rbac_2)
	}
}

func TestKindClusterRBAC_CreateRelease(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespace(t, ns1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteNamespace(t, ns1); err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", ns1, err)
		}
	})

	err = kubeAddHelmRepository(t, "podinfo", podinfo_repo_url, ns1, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		err = kubeDeleteHelmRepository(t, "podinfo", ns1)
		if err != nil {
			t.Logf("Failed to delete helm source due to [%v]", err)
		}
	})

	err = kubeWaitUntilHelmRepositoryIsReady(t, "podinfo", ns1)
	if err != nil {
		t.Fatal(err)
	}

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, "test-release-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	ns2 := "test-ns2-" + randSeq(4)
	if err := kubeCreateNamespace(t, ns2); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteNamespace(t, ns2); err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", ns2, err)
		}
	})

	out := kubectlCanI(t, "test-release-rbac-loser", "default", "get", "helmcharts", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	out = kubectlCanI(t, "test-release-rbac-loser", "default", "get", "helmreleases", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	out = kubectlCanI(t, "test-release-rbac-loser", "default", "create", "helmreleases", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	ctx, cancel := context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	req := &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", ns1),
		Name:                "podinfo",
		TargetContext: &corev1.Context{
			Namespace: ns2,
			Cluster:   KubeappsCluster,
		},
	}
	_, err = fluxPluginClient.CreateInstalledPackage(ctx, req)
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	nsToRules := map[string][]rbacv1.PolicyRule{
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"create"},
			},
		},
	}

	grpcCtx2, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-2", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx2, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.CreateInstalledPackage(ctx, req)
	// perhaps not super intuitive BUT
	// this should fail due to not having 'get' access for HelmCharts in ns1
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	nsToRules = map[string][]rbacv1.PolicyRule{
		ns1: {
			{
				APIGroups: []string{sourcev1.GroupVersion.Group},
				Resources: []string{"helmcharts"},
				Verbs:     []string{"get"},
			},
		},
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"create"},
			},
		},
	}

	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-3", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginClient.CreateInstalledPackage(ctx, req)
	if err != nil {
		t.Fatal(err)
	} else {
		// not necessary to delete release because the whole namespace ns2 will be deleted
		expectedRef := installedRef("podinfo", ns2)
		opts := cmpopts.IgnoreUnexported(
			corev1.InstalledPackageDetail{},
			corev1.InstalledPackageReference{},
			plugins.Plugin{},
			corev1.Context{})
		if got, want := resp.InstalledPackageRef, expectedRef; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
	}
}

func TestKindClusterRBAC_UpdateRelease(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespace(t, ns1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteNamespace(t, ns1); err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", ns1, err)
		}
	})

	tc := integrationTestCreatePackageSpec{
		testName: "test chart RBAC",
		repoUrl:  podinfo_repo_url,
		request: &corev1.CreateInstalledPackageRequest{
			AvailablePackageRef: availableRef("podinfo-1/podinfo", ns1),
			Name:                "my-podinfo",
			TargetContext: &corev1.Context{
				// note that Namespace is just the prefix - the actual name will
				// have a random string appended at the end, e.g. "test-ns2-h23r"
				// this will happen during the running of the test
				Namespace: "test-ns2",
				Cluster:   KubeappsCluster,
			},
		},
		expectedDetail:       expected_detail_test_release_rbac_3,
		expectedPodPrefix:    "my-podinfo-",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_basic,
	}

	grpcCtxAdmin, err := newGrpcAdminContext(t, "test-release-rbac-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, "test-release-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	req := &corev1.UpdateInstalledPackageRequest{
		InstalledPackageRef: installedRef,
		Values:              "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}
	_, err = fluxPluginClient.UpdateInstalledPackage(ctx, req)
	// should fail due to rpc error: code = PermissionDenied desc = Forbidden to get the
	// HelmRelease 'test-ns2-b8jg/my-podinfo' due to 'helmreleases.helm.toolkit.fluxcd.io
	// "my-podinfo" is forbidden: User "system:serviceaccount:default:test-release-rbac-loser"
	// cannot get resource "helmreleases" in API group "helm.toolkit.fluxcd.io" in the namespace
	// "test-ns2-b8jg"'
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	nsToRules := map[string][]rbacv1.PolicyRule{
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"get"},
			},
		},
	}

	grpcCtx2, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-2", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx2, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.UpdateInstalledPackage(ctx, req)
	// should fail due to rpc error: code = PermissionDenied desc = Forbidden to update the
	// HelmRelease 'test-ns2-w8xd/my-podinfo' due to 'helmreleases.helm.toolkit.fluxcd.io
	// "my-podinfo" is forbidden: User "system:serviceaccount:default:test-release-rbac-helmreleases-2"
	// cannot update resource "helmreleases" in API group "helm.toolkit.fluxcd.io" in the
	// namespace "test-ns2-w8xd"'
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	nsToRules = map[string][]rbacv1.PolicyRule{
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"get", "update"},
			},
		},
	}

	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-3", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginClient.UpdateInstalledPackage(ctx, req)
	if err != nil {
		t.Fatal(err)
	} else {
		// not necessary to delete release because the whole namespace ns2 will be deleted
		opts := cmpopts.IgnoreUnexported(
			corev1.InstalledPackageDetail{},
			corev1.InstalledPackageReference{},
			plugins.Plugin{},
			corev1.Context{})
		if got, want := resp.InstalledPackageRef, installedRef; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
	}
}

func TestKindClusterRBAC_DeleteRelease(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespace(t, ns1); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := kubeDeleteNamespace(t, ns1); err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", ns1, err)
		}
	})

	tc := integrationTestCreatePackageSpec{
		testName: "test chart RBAC",
		repoUrl:  podinfo_repo_url,
		request: &corev1.CreateInstalledPackageRequest{
			AvailablePackageRef: availableRef("podinfo-1/podinfo", ns1),
			Name:                "my-podinfo",
			TargetContext: &corev1.Context{
				// note that Namespace is just the prefix - the actual name will
				// have a random string appended at the end, e.g. "test-ns2-h23r"
				// this will happen during the running of the test
				Namespace: "test-ns2",
				Cluster:   KubeappsCluster,
			},
		},
		expectedDetail:       expected_detail_test_release_rbac_4,
		expectedPodPrefix:    "my-podinfo-",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_basic,
	}

	grpcCtxAdmin, err := newGrpcAdminContext(t, "test-release-rbac-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, "test-release-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	req := &corev1.DeleteInstalledPackageRequest{
		InstalledPackageRef: installedRef,
	}

	_, err = fluxPluginClient.DeleteInstalledPackage(ctx, req)
	// should fail due to rpc error: code = PermissionDenied desc = Forbidden to delete the
	// HelmRelease 'my-podinfo' due to 'helmreleases.helm.toolkit.fluxcd.io "my-podinfo" is
	// forbidden: User "system:serviceaccount:default:test-release-rbac-loser" cannot delete
	// resource "helmreleases" in API group "helm.toolkit.fluxcd.io" in the namespace "test-ns2-g4yp"'
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	nsToRules := map[string][]rbacv1.PolicyRule{
		ns2: {
			{
				APIGroups: []string{helmv2.GroupVersion.Group},
				Resources: []string{fluxHelmReleases},
				Verbs:     []string{"delete"},
			},
		},
	}

	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, "test-release-rbac-helmreleases-3", "default", nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginClient.DeleteInstalledPackage(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
}

func createAndWaitForHelmRelease(t *testing.T, tc integrationTestCreatePackageSpec, fluxPluginClient fluxplugin.FluxV2PackagesServiceClient, grpcContext context.Context) *corev1.InstalledPackageReference {
	availablePackageRef := tc.request.AvailablePackageRef
	idParts := strings.Split(availablePackageRef.Identifier, "/")
	err := kubeAddHelmRepository(t, idParts[0], tc.repoUrl, availablePackageRef.Context.Namespace, "", tc.repoInterval)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	t.Cleanup(func() {
		err = kubeDeleteHelmRepository(t, idParts[0], availablePackageRef.Context.Namespace)
		if err != nil {
			t.Logf("Failed to delete helm source due to [%v]", err)
		}
	})

	// need to wait until repo is indexed by flux plugin
	const maxWait = 25
	for i := 0; i <= maxWait; i++ {
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		resp, err := fluxPluginClient.GetAvailablePackageDetail(
			grpcContext,
			&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
		if err == nil {
			break
		} else if i == maxWait {
			t.Fatalf("Timed out waiting for available package [%s], last response: %v, last error: [%v]", availablePackageRef, resp, err)
		} else {
			t.Logf("Waiting 1s for repository [%s] to be indexed, attempt [%d/%d]...", idParts[0], i+1, maxWait)
			time.Sleep(1 * time.Second)
		}
	}

	// generate a unique target namespace for each test to avoid situations when tests are
	// run multiple times in a row and they fail due to the fact that the specified namespace
	// in 'Terminating' state
	if tc.request.TargetContext.Namespace != "" {
		tc.request.TargetContext.Namespace += "-" + randSeq(4)

		if !tc.noPreCreateNs {
			// per https://github.com/vmware-tanzu/kubeapps/pull/3640#issuecomment-950383123
			if err := kubeCreateNamespace(t, tc.request.TargetContext.Namespace); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err = kubeDeleteNamespace(t, tc.request.TargetContext.Namespace); err != nil {
					t.Logf("Failed to delete namespace [%s] due to [%v]", tc.request.TargetContext.Namespace, err)
				}
			})
		}
	}

	if tc.request.ReconciliationOptions != nil && tc.request.ReconciliationOptions.ServiceAccountName != "" {
		_, err = kubeCreateAdminServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, tc.request.TargetContext.Namespace)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		// it appears that if service account is deleted before the helmrelease object that uses it,
		// when you try to delete the helmrelease, the "delete" operation gets stuck and the only
		// way to get it "unstuck" is to edit the CRD and remove the finalizer.
		// So we'll cleanup the service account only after the corresponding helmrelease has been deleted
		t.Cleanup(func() {
			if !tc.expectInstallFailure {
				for i := 0; i < 20; i++ {
					exists, _ := kubeExistsHelmRelease(t, tc.expectedDetail.InstalledPackageRef.Identifier, tc.expectedDetail.InstalledPackageRef.Context.Namespace)
					if exists {
						time.Sleep(300 * time.Millisecond)
					} else {
						break
					}
				}
			}
			err := kubeDeleteServiceAccountWithClusterRoleBinding(t, tc.request.ReconciliationOptions.ServiceAccountName, tc.request.TargetContext.Namespace)
			if err != nil {
				t.Logf("Failed to delete service account due to [%v]", err)
			}
		})
	}

	ctx, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
	defer cancel()

	if tc.expectedStatusCode == codes.Unauthenticated {
		ctx = context.TODO()
	}
	resp, err := fluxPluginClient.CreateInstalledPackage(ctx, tc.request)
	if tc.expectedStatusCode != codes.OK {
		if status.Code(err) != tc.expectedStatusCode {
			t.Fatalf("Expected %v, got: %v", tc.expectedStatusCode, err)
		}
		return nil // done, nothing more to check
	} else if err != nil {
		t.Fatalf("CreateInstalledPackage failed due to: %+v", err)
	}

	if tc.expectedDetail != nil {
		// set some of the expected fields here to values we already know to expect,
		// the rest should be specified explicitly
		tc.expectedDetail.InstalledPackageRef = installedRef(tc.request.Name, tc.request.TargetContext.Namespace)
		tc.expectedDetail.AvailablePackageRef = tc.request.AvailablePackageRef
		tc.expectedDetail.Name = tc.request.Name
		if tc.request.ReconciliationOptions == nil {
			tc.expectedDetail.ReconciliationOptions = &corev1.ReconciliationOptions{
				Interval: 60,
			}
		}
	}

	installedPackageRef := resp.InstalledPackageRef
	opts := cmpopts.IgnoreUnexported(
		corev1.InstalledPackageDetail{},
		corev1.InstalledPackageReference{},
		plugins.Plugin{},
		corev1.Context{})
	if got, want := installedPackageRef, tc.expectedDetail.InstalledPackageRef; !cmp.Equal(want, got, opts) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
	}

	if !tc.noCleanup {
		t.Cleanup(func() {
			err = kubeDeleteHelmRelease(t, installedPackageRef.Identifier, installedPackageRef.Context.Namespace)
			if err != nil {
				t.Logf("Failed to delete helm release due to [%v]", err)
			}
		})
	}

	actualDetailResp, actualRefsResp := waitUntilInstallCompletes(t, fluxPluginClient, grpcContext, installedPackageRef, tc.expectInstallFailure)

	tc.expectedDetail.PostInstallationNotes = strings.ReplaceAll(
		tc.expectedDetail.PostInstallationNotes, "@TARGET_NS@", tc.request.TargetContext.Namespace)

	expectedResp := &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: tc.expectedDetail,
	}

	compareActualVsExpectedGetInstalledPackageDetailResponse(t, actualDetailResp, expectedResp)

	if !tc.expectInstallFailure {
		// check artifacts in target namespace:
		expectedPodPrefix := strings.ReplaceAll(
			tc.expectedPodPrefix, "@TARGET_NS@", tc.request.TargetContext.Namespace)
		pods, err := kubeGetPodNames(t, tc.request.TargetContext.Namespace)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if len(pods) != 1 {
			t.Errorf("expected 1 pod, got: %s", pods)
		} else if !strings.HasPrefix(pods[0], expectedPodPrefix) {
			t.Errorf("expected pod with prefix [%s] not found in namespace [%s], pods found: [%v]",
				expectedPodPrefix, tc.request.TargetContext.Namespace, pods)
		}
	}

	if tc.expectedResourceRefs != nil {
		expectedRefsCopy := []*corev1.ResourceRef{}
		for _, r := range tc.expectedResourceRefs {
			newR := &corev1.ResourceRef{
				ApiVersion: r.ApiVersion,
				Kind:       r.Kind,
				Name:       r.Name,
				Namespace:  tc.request.TargetContext.Namespace,
			}
			expectedRefsCopy = append(expectedRefsCopy, newR)
		}
		opts := cmpopts.IgnoreUnexported(corev1.ResourceRef{})
		if got, want := actualRefsResp.ResourceRefs, expectedRefsCopy; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
	}
	return installedPackageRef
}

func waitUntilInstallCompletes(
	t *testing.T,
	fluxPluginClient fluxplugin.FluxV2PackagesServiceClient,
	grpcContext context.Context,
	installedPackageRef *corev1.InstalledPackageReference,
	expectInstallFailure bool) (
	actualDetailResp *corev1.GetInstalledPackageDetailResponse,
	actualRefsResp *corev1.GetInstalledPackageResourceRefsResponse) {
	const maxWait = 30
	for i := 0; i <= maxWait; i++ {
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		resp2, err := fluxPluginClient.GetInstalledPackageDetail(
			grpcContext,
			&corev1.GetInstalledPackageDetailRequest{InstalledPackageRef: installedPackageRef})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if !expectInstallFailure {
			if resp2.InstalledPackageDetail.Status.Ready == true &&
				resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED {
				actualDetailResp = resp2
				break
			}
		} else {
			if resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_FAILED ||
				resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED {
				actualDetailResp = resp2
				break
			}
		}
		t.Logf("Waiting 2s until install completes due to: [%s], userReason: [%s], attempt: [%d/%d]...",
			resp2.InstalledPackageDetail.Status.Reason, resp2.InstalledPackageDetail.Status.UserReason, i+1, maxWait)
		time.Sleep(2 * time.Second)
	}

	if actualDetailResp == nil {
		t.Fatalf("Timed out waiting for task to complete")
	} else if actualDetailResp.InstalledPackageDetail.Status.Ready {
		t.Logf("Getting installed package resource refs for [%s]...", installedPackageRef.Identifier)
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		var err error
		actualRefsResp, err = fluxPluginClient.GetInstalledPackageResourceRefs(
			grpcContext,
			&corev1.GetInstalledPackageResourceRefsRequest{InstalledPackageRef: installedPackageRef})
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
	return actualDetailResp, actualRefsResp
}
