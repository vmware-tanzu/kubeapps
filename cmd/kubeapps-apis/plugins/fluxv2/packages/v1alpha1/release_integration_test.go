// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
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
//   c) - "...-helmreleases", with only permissions to read HelmReleases in ns2
// 5) as user 4a) install package podinfo in ns2
// 6) verify GetInstalledPackageSummaries:
//    a) as 4a) returns 1 result
//    b) as 4b) raises PermissionDenied error
//    c) as 4c) returns 1 result but without the corresponding chart details
// ref https://github.com/kubeapps/kubeapps/issues/4390
func TestKindClusterReleaseRBAC(t *testing.T) {
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

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, "test-release-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	out := kubectlCanIGetThisInNamespace(t, "test-release-rbac-admin", "default", "helmcharts", ns1)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}
	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-loser", "default", "helmcharts", ns1)
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

	createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-admin", "default", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	resp, err := fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtxAdmin,
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

	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-loser", "default", fluxHelmReleases, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	_, err = fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtxLoser,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{helmv2.GroupVersion.Group},
			Resources: []string{fluxHelmReleases},
			Verbs:     []string{"get", "list"},
		},
	}

	grpcCtxReadHelmReleases, err := newGrpcContextForServiceAccountWithAccessToNamespace(t, "test-release-rbac-helmreleases", "default", ns2, rules)
	if err != nil {
		t.Fatal(err)
	}

	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-helmreleases", "default", fluxHelmRepositories, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-helmreleases", "default", "helmcharts", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanIGetThisInNamespace(t, "test-release-rbac-helmreleases", "default", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	resp2, err := fluxPluginClient.GetInstalledPackageSummaries(
		grpcCtxReadHelmReleases,
		&corev1.GetInstalledPackageSummariesRequest{
			Context: &corev1.Context{
				Namespace: ns2,
			},
		})
	if err != nil {
		t.Fatal(err)
	} else {
		// should return installed package summaries without chart details
		expected_summaries_test_release_rbac_1.InstalledPackageSummaries[0].InstalledPackageRef.Context.Namespace = ns2
		opts := cmpopts.IgnoreUnexported(
			corev1.GetInstalledPackageSummariesResponse{},
			corev1.InstalledPackageSummary{},
			corev1.InstalledPackageReference{},
			corev1.InstalledPackageStatus{},
			plugins.Plugin{},
			corev1.VersionReference{},
			corev1.Context{})
		if got, want := resp2, expected_summaries_test_release_rbac_1; !cmp.Equal(want, got, opts) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
		}
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
			// per https://github.com/kubeapps/kubeapps/pull/3640#issuecomment-950383123
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
		t.Fatalf("%+v", err)
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

// global vars
// why define these here? see https://github.com/kubeapps/kubeapps/pull/3736#discussion_r745246398
var (
	create_request_basic = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-1/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			// note that Namespace is just the prefix - the actual name will
			// have a random string appended at the end, e.g. "test-1-h23r"
			// this will happen during the running of the test
			Namespace: "test-1",
			Cluster:   KubeappsCluster,
		},
	}

	// specify just the fields that cannot be easily computed based on the request
	expected_detail_basic = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
	}

	expected_resource_refs_basic = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo",
		},
	}

	create_request_semver_constraint = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-2/podinfo", "default"),
		Name:                "my-podinfo-2",
		TargetContext: &corev1.Context{
			Namespace: "test-2",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
	}

	expected_detail_semver_constraint = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-2 8080:9898\n",
	}

	expected_resource_refs_semver_constraint = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-2",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-2",
		},
	}

	create_request_reconcile_options = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-3/podinfo", "default"),
		Name:                "my-podinfo-3",
		TargetContext: &corev1.Context{
			Namespace: "test-3",
			Cluster:   KubeappsCluster,
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            false,
			ServiceAccountName: "foo",
		},
	}

	expected_detail_reconcile_options = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            false,
			ServiceAccountName: "foo",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-3 8080:9898\n",
	}

	expected_resource_refs_reconcile_options = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-3",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-3",
		},
	}

	create_request_with_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-4/podinfo", "default"),
		Name:                "my-podinfo-4",
		TargetContext: &corev1.Context{
			Namespace: "test-4",
			Cluster:   KubeappsCluster,
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	expected_detail_with_values = &corev1.InstalledPackageDetail{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-4 8080:9898\n",
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_resource_refs_with_values = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-4",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-4",
		},
	}

	create_request_install_fails = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-5/podinfo", "default"),
		Name:                "my-podinfo-5",
		TargetContext: &corev1.Context{
			Namespace: "test-5",
			Cluster:   KubeappsCluster,
		},
		Values: "{\"replicaCount\": \"what we do in the shadows\" }",
	}

	expected_detail_install_fails = &corev1.InstalledPackageDetail{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:  false,
			Reason: corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			// most of the time it fails with
			//   "InstallFailed: install retries exhausted",
			// but every once in a while you get
			//   "InstallFailed: Helm install failed: unable to build kubernetes objects from release manifest: error
			//    validating "": error validating data: ValidationError(Deployment.spec.replicas): invalid type for
			//    io.k8s.api.apps.v1.DeploymentSpec.replicas: got "string""
			// so we'll just test the prefix
			UserReason: "InstallFailed: ",
		},
		ValuesApplied: "{\"replicaCount\":\"what we do in the shadows\"}",
	}

	create_request_podinfo_5_2_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
		Name:                "my-podinfo-6",
		TargetContext: &corev1.Context{
			Namespace: "test-6",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-6 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-6",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-6",
		},
	}

	expected_detail_podinfo_6_0_0 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-6 8080:9898\n",
	}

	create_request_podinfo_5_2_1_no_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
		Name:                "my-podinfo-7",
		TargetContext: &corev1.Context{
			Namespace: "test-7",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1_no_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-7 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_no_values = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-7",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-7",
		},
	}

	expected_detail_podinfo_5_2_1_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-7 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
		Name:                "my-podinfo-8",
		TargetContext: &corev1.Context{
			Namespace: "test-8",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-8 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_2 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-8",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-8",
		},
	}

	expected_detail_podinfo_5_2_1_values_3 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"Le Bureau des Légendes\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-8 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_4 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
		Name:                "my-podinfo-9",
		TargetContext: &corev1.Context{
			Namespace: "test-9",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_4 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-9 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_4 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-9",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-9",
		},
	}

	expected_detail_podinfo_5_2_1_values_5 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-9 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_6 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-10/podinfo", "default"),
		Name:                "my-podinfo-10",
		TargetContext: &corev1.Context{
			Namespace: "test-10",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_6 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-10 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_6 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-10",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-10",
		},
	}

	create_request_podinfo_7 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-11/podinfo", "default"),
		Name:                "my-podinfo-11",
		TargetContext: &corev1.Context{
			Namespace: "test-11",
			Cluster:   KubeappsCluster,
		},
	}

	expected_detail_podinfo_7 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-11 8080:9898\n",
	}

	expected_resource_refs_podinfo_7 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-11",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-11",
		},
	}

	update_request_1 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
	}

	update_request_2 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	update_request_3 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"Le Bureau des Légendes\" } }",
	}

	update_request_4 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "",
	}

	update_request_5 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	update_request_6 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	create_request_podinfo_for_delete_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-12/podinfo", "default"),
		Name:                "my-podinfo-12",
		TargetContext: &corev1.Context{
			Namespace: "test-12",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_for_delete_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-12 8080:9898\n",
	}

	expected_resource_refs_for_delete_1 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-12",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-12",
		},
	}

	create_request_podinfo_for_delete_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-13/podinfo", "default"),
		Name:                "my-podinfo-13",
		TargetContext: &corev1.Context{
			Namespace: "test-13",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_for_delete_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-13 8080:9898\n",
	}

	expected_resource_refs_for_delete_2 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-13",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-13",
		},
	}

	create_request_wrong_cluster = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-14/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-14",
			Cluster:   "this is not the cluster you're looking for",
		},
	}

	create_request_target_ns_doesnt_exist = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-15/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-15",
			Cluster:   KubeappsCluster,
		},
	}

	create_request_auto_update = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-16/podinfo", "default"),
		Name:                "my-podinfo-16",
		TargetContext: &corev1.Context{
			Namespace: "test-16",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
	}

	expected_detail_auto_update = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-16 8080:9898\n",
	}

	expected_detail_auto_update_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.3",
			AppVersion: "6.0.3",
		},
		Name:   "my-podinfo-16",
		Status: statusInstalled,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			},
			Identifier: "podinfo-16/podinfo",
			Plugin:     fluxPlugin,
		},
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-16 8080:9898\n",
	}

	expected_resource_refs_auto_update = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-16",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-16",
		},
	}

	expected_detail_test_release_rbac = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
	}

	expected_summaries_test_release_rbac_1 = &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
			{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   KubeappsCluster,
						Namespace: "@TARGET_NS@",
					},
					Identifier: "my-podinfo",
					Plugin:     fluxPlugin,
				},
				Name: "my-podinfo",
				PkgVersionReference: &corev1.VersionReference{
					Version: "*",
				},
				Status: &corev1.InstalledPackageStatus{
					Ready:      true,
					Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
				},
				// notice that the details from the corresponding chart, like LatestPkgVersion, are missing
			},
		},
	}
)
