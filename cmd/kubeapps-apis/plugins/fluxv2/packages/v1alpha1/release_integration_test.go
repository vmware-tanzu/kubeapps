// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"math"
	"os"
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
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	fluxHelmReleases = "helmreleases"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
// 3) run './integ-test-env.sh deploy' once prior to these tests

type integrationTestCreatePackageSpec struct {
	testName          string
	repoType          string
	repoUrl           string
	repoInterval      time.Duration // 0 for default (10m)
	repoSecret        *apiv1.Secret
	request           *corev1.CreateInstalledPackageRequest
	expectedDetail    *corev1.InstalledPackageDetail
	expectedPodPrefix string
	// what follows are boolean flags to test various negative scenarios
	// different from expectedStatusCode due to async nature of install
	expectInstallFailure bool
	dontCreateNs         bool
	noCleanup            bool
	expectedStatusCode   codes.Code
	expectedResourceRefs []*corev1.ResourceRef
}

func TestKindClusterCreateInstalledPackage(t *testing.T) {
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ghUser := os.Getenv("GITHUB_USER")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghUser == "" || ghToken == "" {
		t.Fatalf("Environment variables GITHUB_USER and GITHUB_TOKEN need to be set to run this test")
	}

	gcpUser := "oauth2accesstoken"
	gcpPasswd, err := gcloudPrintAccessToken(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []integrationTestCreatePackageSpec{
		{
			testName:             "create test (simplest case)",
			repoUrl:              podinfo_repo_url,
			request:              create_installed_package_request_basic,
			expectedDetail:       expected_detail_installed_package_basic,
			expectedPodPrefix:    "my-podinfo-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_basic,
		},
		{
			testName:             "create package (semver constraint)",
			repoUrl:              podinfo_repo_url,
			request:              create_installed_package_request_semver_constraint,
			expectedDetail:       expected_detail_installed_package_semver_constraint,
			expectedPodPrefix:    "my-podinfo-2-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_semver_constraint,
		},
		{
			testName:             "create package (reconcile options)",
			repoUrl:              podinfo_repo_url,
			request:              create_installed_package_request_reconcile_options,
			expectedDetail:       expected_detail_installed_package_reconcile_options,
			expectedPodPrefix:    "my-podinfo-3-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_reconcile_options,
		},
		{
			testName:             "create package (with values)",
			repoUrl:              podinfo_repo_url,
			request:              create_installed_package_request_with_values,
			expectedDetail:       expected_detail_installed_package_with_values,
			expectedPodPrefix:    "my-podinfo-4-",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_with_values,
		},
		{
			testName:             "install fails",
			repoUrl:              podinfo_repo_url,
			request:              create_installed_package_request_install_fails,
			expectedDetail:       expected_detail_installed_package_install_fails,
			expectInstallFailure: true,
			expectedStatusCode:   codes.OK,
		},
		{
			testName:           "unauthorized",
			repoUrl:            podinfo_repo_url,
			request:            create_installed_package_request_basic,
			expectedStatusCode: codes.Unauthenticated,
		},
		{
			testName:           "wrong cluster",
			repoUrl:            podinfo_repo_url,
			request:            create_installed_package_request_wrong_cluster,
			expectedStatusCode: codes.Unimplemented,
		},
		{
			testName:           "target namespace does not exist",
			repoUrl:            podinfo_repo_url,
			request:            create_installed_package_request_target_ns_doesnt_exist,
			dontCreateNs:       true,
			expectedStatusCode: codes.NotFound,
		},
		{
			testName: "create OCI helm release from [" + github_gfichtenholt_podinfo_oci_registry_url + "]",
			repoType: "oci",
			repoUrl:  github_gfichtenholt_podinfo_oci_registry_url,
			repoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				ghUser,
				ghToken,
			),
			request:              create_installed_package_request_oci,
			expectedDetail:       expected_detail_installed_package_oci,
			expectedPodPrefix:    "my-podinfo-17",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_oci,
		},
		{
			testName: "create OCI helm release from [" + gcp_stefanprodan_podinfo_oci_registry_url + "]",
			repoType: "oci",
			repoUrl:  gcp_stefanprodan_podinfo_oci_registry_url,
			repoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				gcpUser,
				string(gcpPasswd),
			),
			request:              create_installed_package_request_oci_2,
			expectedDetail:       expected_detail_installed_package_oci_2,
			expectedPodPrefix:    "my-podinfo-19",
			expectedStatusCode:   codes.OK,
			expectedResourceRefs: expected_resource_refs_oci_2,
		},
	}

	name := types.NamespacedName{
		Name:      "test-create-admin" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			createAndWaitForHelmRelease(t, tc, fluxPluginPackagesClient, fluxPluginReposClient, grpcContext)
		})
	}
}

func TestKindClusterUpdateInstalledPackage(t *testing.T) {
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		integrationTestCreatePackageSpec
		request *corev1.UpdateInstalledPackageRequest
		// this is expected AFTER the update call completes
		expectedDetailAfterUpdate *corev1.InstalledPackageDetail
		expectedRefsAfterUpdate   []*corev1.ResourceRef
		unauthorized              bool
	}{
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (simplest case)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_5_2_1,
				expectedDetail:       expected_detail_installed_package_podinfo_5_2_1,
				expectedPodPrefix:    "my-podinfo-6-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1,
			},
			request:                   update_request_1,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_6_0_0,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (add values)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_5_2_1_no_values,
				expectedDetail:       expected_detail_installed_package_podinfo_5_2_1_no_values,
				expectedPodPrefix:    "my-podinfo-7-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_no_values,
			},
			request:                   update_request_2,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_5_2_1_values,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_no_values,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (change values)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_5_2_1_values_2,
				expectedDetail:       expected_detail_installed_package_podinfo_5_2_1_values_2,
				expectedPodPrefix:    "my-podinfo-8-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_2,
			},
			request:                   update_request_3,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_5_2_1_values_3,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_2,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (remove values)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_5_2_1_values_4,
				expectedDetail:       expected_detail_installed_package_podinfo_5_2_1_values_4,
				expectedPodPrefix:    "my-podinfo-9-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_4,
			},
			request:                   update_request_4,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_5_2_1_values_5,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_4,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update test (values dont change)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_5_2_1_values_6,
				expectedDetail:       expected_detail_installed_package_podinfo_5_2_1_values_6,
				expectedPodPrefix:    "my-podinfo-10-",
				expectedResourceRefs: expected_resource_refs_podinfo_5_2_1_values_6,
			},
			request:                   update_request_5,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_5_2_1_values_6,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_5_2_1_values_6,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update unauthorized test",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_7,
				expectedDetail:       expected_detail_installed_package_podinfo_7,
				expectedPodPrefix:    "my-podinfo-11-",
				expectedResourceRefs: expected_resource_refs_podinfo_7,
			},
			request:      update_request_6,
			unauthorized: true,
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "update installed package in failed state is allowed",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_8,
				expectedDetail:       expected_detail_installed_package_podinfo_8,
				expectInstallFailure: true,
			},
			request:                   update_request_7,
			expectedDetailAfterUpdate: expected_detail_installed_package_podinfo_9,
			expectedRefsAfterUpdate:   expected_resource_refs_podinfo_9,
		},
		// TODO (gfichtenholt) update OCI helmrelease
	}

	name := types.NamespacedName{
		Name:      "test-update-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(
				t, tc.integrationTestCreatePackageSpec, fluxPluginPackagesClient, fluxPluginReposClient, grpcContext)
			tc.request.InstalledPackageRef = installedRef

			ctx := grpcContext
			if tc.unauthorized {
				ctx = context.TODO()
			}

			// Every once in a while (very infrequently, like 1 out of 25 tries)
			// I get rpc error: code = Internal desc = unable to update the HelmRelease
			// 'test-12-i7a4/my-podinfo-12' due to 'Operation cannot be fulfilled on
			// helmreleases.helm.toolkit.fluxcd.io "my-podinfo-12": the object has been
			// modified; please apply your changes to the latest version and try again'
			// ... so this is the reason for the loop loop with retries
			var i, maxRetries = 0, 5
			for ; i < maxRetries; i++ {
				_, err := fluxPluginPackagesClient.UpdateInstalledPackage(ctx, tc.request)
				if tc.unauthorized {
					if status.Code(err) != codes.Unauthenticated {
						t.Fatalf("Expected Unathenticated, got: %v", status.Code(err))
					}
					return // done, nothing more to check
				} else if err != nil {
					if strings.Contains(err.Error(), " the object has been modified; please apply your changes to the latest version and try again") {
						waitTime := int64(math.Pow(2, float64(i)))
						t.Logf("Retrying update in [%d] sec due to %s...", waitTime, err.Error())
						time.Sleep(time.Duration(waitTime) * time.Second)
					} else {
						t.Fatalf("%+v", err)
					}
				} else {
					break
				}
			}
			if i == maxRetries {
				t.Fatalf("Update retries exhaused for package [%s], last error: [%v]", installedRef, err)
			}

			actualRespAfterUpdate, actualRefsAfterUpdate :=
				waitUntilInstallCompletes(t, fluxPluginPackagesClient, grpcContext, installedRef, false)

			tc.expectedDetailAfterUpdate.InstalledPackageRef = installedRef
			tc.expectedDetailAfterUpdate.Name = tc.integrationTestCreatePackageSpec.request.Name
			tc.expectedDetailAfterUpdate.ReconciliationOptions = &corev1.ReconciliationOptions{
				Interval: "1m",
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

// The goal of this integration test is to ensure that when the contents of remote HTTP helm repo is changed,
// that fact is recorded locally and processed properly (repo/chart cache is updated with latest, etc.)
func TestKindClusterAutoUpdateInstalledPackageFromHttpRepo(t *testing.T) {
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	spec := integrationTestCreatePackageSpec{
		testName:             "create test (HTTP repo auto update)",
		repoUrl:              podinfo_repo_url,
		repoInterval:         30 * time.Second,
		request:              create_installed_package_request_auto_update,
		expectedDetail:       expected_detail_installed_package_auto_update,
		expectedPodPrefix:    "my-podinfo-16",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_auto_update,
	}

	name := types.NamespacedName{
		Name:      "test-auto-update-" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	// this will also make sure that response looks like expected_detail_installed_package_auto_update
	installedRef := createAndWaitForHelmRelease(t, spec, fluxPluginPackagesClient, fluxPluginReposClient, grpcContext)
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
			t.Logf("Error reverting to previous podinfo index: %v", err)
		}
	})
	t.Logf("Waiting 45 seconds...")
	time.Sleep(45 * time.Second)

	resp, err := fluxPluginPackagesClient.GetInstalledPackageDetail(
		grpcContext, &corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	}
	expected_detail_installed_package_auto_update_2.InstalledPackageRef = installedRef
	expected_detail_installed_package_auto_update_2.PostInstallationNotes = strings.ReplaceAll(
		expected_detail_installed_package_auto_update_2.PostInstallationNotes,
		"@TARGET_NS@",
		spec.request.TargetContext.Namespace)
	compareActualVsExpectedGetInstalledPackageDetailResponse(
		t, resp, &corev1.GetInstalledPackageDetailResponse{
			InstalledPackageDetail: expected_detail_installed_package_auto_update_2,
		})
}

// The goal of this integration test is to ensure that when the contents of remote OCI helm repo is changed,
// that fact is recorded locally and processed properly (repo/chart cache is updated with latest, etc.)
func TestKindClusterAutoUpdateInstalledPackageFromOciRepo(t *testing.T) {
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	// see TestKindClusterAvailablePackageEndpointsForOCI for explanation
	ghUser := os.Getenv("GITHUB_USER")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghUser == "" || ghToken == "" {
		t.Fatalf("Environment variables GITHUB_USER and GITHUB_TOKEN need to be set to run this test")
	}

	spec := integrationTestCreatePackageSpec{
		testName:     "create test (OCI repo auto update)",
		repoUrl:      github_gfichtenholt_podinfo_oci_registry_url,
		repoType:     "oci",
		repoInterval: 30 * time.Second,
		repoSecret: newBasicAuthSecret(types.NamespacedName{
			Name:      "oci-repo-secret-" + randSeq(4),
			Namespace: "default"},
			ghUser,
			ghToken,
		),
		request:              create_installed_package_request_auto_update_oci,
		expectedDetail:       expected_detail_installed_package_auto_update_oci,
		expectedPodPrefix:    "my-podinfo-18",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_auto_update_oci,
	}

	name := types.NamespacedName{
		Name:      "test-auto-update-" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	installedRef := createAndWaitForHelmRelease(t, spec, fluxPluginPackagesClient, fluxPluginReposClient, grpcContext)

	podName, err := getFluxPluginTestdataPodName()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("podName = [%s]", podName)

	if err = helmPushChartToMyGithubRegistry(t, "6.1.6"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Delete remote chart version at the end of the test so there are no side-effects.
		if err = deleteChartFromMyGithubRegistry(t, "6.1.6"); err != nil {
			t.Fatal(err)
		}
	})

	t.Logf("Waiting 45 seconds...")
	time.Sleep(45 * time.Second)

	resp, err := fluxPluginPackagesClient.GetInstalledPackageDetail(
		grpcContext, &corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if err != nil {
		t.Fatal(err)
	}

	expected_detail_installed_package_auto_update_oci_2.InstalledPackageRef = installedRef
	expected_detail_installed_package_auto_update_oci_2.PostInstallationNotes = strings.ReplaceAll(
		expected_detail_installed_package_auto_update_oci_2.PostInstallationNotes,
		"@TARGET_NS@",
		spec.request.TargetContext.Namespace)
	compareActualVsExpectedGetInstalledPackageDetailResponse(
		t, resp, &corev1.GetInstalledPackageDetailResponse{
			InstalledPackageDetail: expected_detail_installed_package_auto_update_oci_2,
		})
}

func TestKindClusterDeleteInstalledPackage(t *testing.T) {
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		integrationTestCreatePackageSpec
		unauthorized bool
	}{
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "delete test (simplest case)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_for_delete_1,
				expectedDetail:       expected_detail_installed_package_podinfo_for_delete_1,
				expectedPodPrefix:    "my-podinfo-12-",
				expectedResourceRefs: expected_resource_refs_for_delete_1,
				noCleanup:            true,
			},
		},
		{
			integrationTestCreatePackageSpec: integrationTestCreatePackageSpec{
				testName:             "delete test (unauthorized)",
				repoUrl:              podinfo_repo_url,
				request:              create_installed_package_request_podinfo_for_delete_2,
				expectedDetail:       expected_detail_installed_package_podinfo_for_delete_2,
				expectedPodPrefix:    "my-podinfo-13-",
				expectedResourceRefs: expected_resource_refs_for_delete_2,
				noCleanup:            true,
			},
			unauthorized: true,
		},
		// TODO (gifchtenholt) delete OCI helmrelease
	}

	name := types.NamespacedName{
		Name:      "test-delete-admin" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(t, tc.integrationTestCreatePackageSpec, fluxPluginPackagesClient, fluxPluginReposClient, grpcContext)

			ctx := grpcContext
			if tc.unauthorized {
				ctx = context.TODO()
			}
			_, err := fluxPluginPackagesClient.DeleteInstalledPackage(ctx, &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: installedRef,
			})
			if tc.unauthorized {
				if status.Code(err) != codes.Unauthenticated {
					t.Fatalf("Expected Unathenticated, got: %v", status.Code(err))
				}
				// still need to delete the release though
				name := types.NamespacedName{
					Name:      installedRef.Identifier,
					Namespace: installedRef.Context.Namespace,
				}
				if err = kubeDeleteHelmRelease(t, name); err != nil {
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

				_, err := fluxPluginPackagesClient.GetInstalledPackageDetail(
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
			name := types.NamespacedName{
				Name:      installedRef.Identifier,
				Namespace: installedRef.Context.Namespace,
			}
			exists, err := kubeExistsHelmRelease(t, name)
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
					t.Fatalf("Timed out waiting for garbage collection of pod [%s]", pods[0])
				} else {
					t.Logf("Waiting 3s for garbage collection of pod [%s], attempt [%d/%d]...", pods[0], i+1, maxWait)
					time.Sleep(3 * time.Second)
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
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespaceAndCleanup(t, ns1); err != nil {
		t.Fatal(err)
	}

	adminAcctName := types.NamespacedName{
		Name:      "test-release-rbac-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxAdmin, err := newGrpcAdminContext(t, adminAcctName)
	if err != nil {
		t.Fatal(err)
	}

	loserAcctName := types.NamespacedName{
		Name:      "test-release-rbac-loser-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, loserAcctName)
	if err != nil {
		t.Fatal(err)
	}

	out := kubectlCanI(t, adminAcctName, "get", "helmcharts", ns1)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}
	out = kubectlCanI(t, loserAcctName, "get", "helmcharts", ns1)
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

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginPackagesClient, fluxPluginReposClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	out = kubectlCanI(t, adminAcctName, "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel := context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginPackagesClient.GetInstalledPackageSummaries(
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

	resp2, err := fluxPluginPackagesClient.GetInstalledPackageDetail(
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

	resp3, err := fluxPluginPackagesClient.GetInstalledPackageResourceRefs(
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

	out = kubectlCanI(t, loserAcctName, "get", fluxHelmReleases, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginPackagesClient.GetInstalledPackageSummaries(
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

	_, err = fluxPluginPackagesClient.GetInstalledPackageDetail(
		grpcCtx,
		&corev1.GetInstalledPackageDetailRequest{
			InstalledPackageRef: installedRef,
		})
	if status.Code(err) != codes.PermissionDenied {
		t.Errorf("Expected PermissionDenied, got %v", err)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginPackagesClient.GetInstalledPackageResourceRefs(
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

	svcAcctName := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxReadHelmReleases, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName, rules)
	if err != nil {
		t.Fatal(err)
	}

	out = kubectlCanI(t, svcAcctName, "get", fluxHelmRepositories, ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanI(t, svcAcctName, "get", "helmcharts", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}
	out = kubectlCanI(t, svcAcctName, "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleases, defaultContextTimeout)
	defer cancel()

	resp, err = fluxPluginPackagesClient.GetInstalledPackageSummaries(
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

	resp2, err = fluxPluginPackagesClient.GetInstalledPackageDetail(
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

	resp3, err = fluxPluginPackagesClient.GetInstalledPackageResourceRefs(
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

	svcAcctName2 := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-and-charts-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxReadHelmReleasesAndCharts, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName2, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	out = kubectlCanI(t, svcAcctName2, "get", "helmcharts", ns1)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}
	out = kubectlCanI(t, svcAcctName2, "get", fluxHelmReleases, ns2)
	if out != "yes" {
		t.Errorf("Expected [yes], got [%s]", out)
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxReadHelmReleasesAndCharts, defaultContextTimeout)
	defer cancel()

	resp, err = fluxPluginPackagesClient.GetInstalledPackageSummaries(
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

	resp2, err = fluxPluginPackagesClient.GetInstalledPackageDetail(
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
	fluxPluginPackagesClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespaceAndCleanup(t, ns1); err != nil {
		t.Fatal(err)
	}
	name := types.NamespacedName{
		Name:      "podinfo",
		Namespace: ns1,
	}

	err = kubeAddHelmRepositoryAndCleanup(t, name, "", podinfo_repo_url, "", 0)
	if err != nil {
		t.Fatal(err)
	}

	err = kubeWaitUntilHelmRepositoryIsReady(t, name)
	if err != nil {
		t.Fatal(err)
	}

	loserAcctName := types.NamespacedName{
		Name:      "test-release-rbac-loser-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, loserAcctName)
	if err != nil {
		t.Fatal(err)
	}

	ns2 := "test-ns2-" + randSeq(4)
	if err := kubeCreateNamespaceAndCleanup(t, ns2); err != nil {
		t.Fatal(err)
	}
	out := kubectlCanI(t, loserAcctName, "get", "helmcharts", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	out = kubectlCanI(t, loserAcctName, "get", "helmreleases", ns2)
	if out != "no" {
		t.Errorf("Expected [no], got [%s]", out)
	}

	out = kubectlCanI(t, loserAcctName, "create", "helmreleases", ns2)
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
	_, err = fluxPluginPackagesClient.CreateInstalledPackage(ctx, req)
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

	svcAcctName2 := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-2-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtx2, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName2, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx2, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginPackagesClient.CreateInstalledPackage(ctx, req)
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

	svcAcctName3 := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-3-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName3, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginPackagesClient.CreateInstalledPackage(ctx, req)
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
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespaceAndCleanup(t, ns1); err != nil {
		t.Fatal(err)
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
		expectedDetail:       expected_detail_test_release_rbac_3,
		expectedPodPrefix:    "my-podinfo-",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_basic,
	}

	adminAcctName := types.NamespacedName{
		Name:      "test-release-rbac-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxAdmin, err := newGrpcAdminContext(t, adminAcctName)
	if err != nil {
		t.Fatal(err)
	}

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginPackagesClient, fluxPluginReposClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	loserAcctName := types.NamespacedName{
		Name:      "test-release-rbac-loser-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, loserAcctName)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	req := &corev1.UpdateInstalledPackageRequest{
		InstalledPackageRef: installedRef,
		Values:              "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}
	_, err = fluxPluginPackagesClient.UpdateInstalledPackage(ctx, req)
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

	svcAcctName2 := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-2-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtx2, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName2, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx2, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginPackagesClient.UpdateInstalledPackage(ctx, req)
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

	svcAcctName3 := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-3-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName3, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	resp, err := fluxPluginPackagesClient.UpdateInstalledPackage(ctx, req)
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
	fluxPluginPackagesClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	ns1 := "test-ns1-" + randSeq(4)
	if err := kubeCreateNamespaceAndCleanup(t, ns1); err != nil {
		t.Fatal(err)
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
		expectedDetail:       expected_detail_test_release_rbac_4,
		expectedPodPrefix:    "my-podinfo-",
		expectedStatusCode:   codes.OK,
		expectedResourceRefs: expected_resource_refs_basic,
	}

	adminAcctName := types.NamespacedName{
		Name:      "test-release-rbac-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxAdmin, err := newGrpcAdminContext(t, adminAcctName)
	if err != nil {
		t.Fatal(err)
	}

	installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginPackagesClient, fluxPluginReposClient, grpcCtxAdmin)

	ns2 := tc.request.TargetContext.Namespace

	loserAcctName := types.NamespacedName{
		Name:      "test-release-rbac-loser-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(
		t, loserAcctName)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()

	req := &corev1.DeleteInstalledPackageRequest{
		InstalledPackageRef: installedRef,
	}

	_, err = fluxPluginPackagesClient.DeleteInstalledPackage(ctx, req)
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

	svcAcctName := types.NamespacedName{
		Name:      "test-release-rbac-helmreleases-3-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtx3, err := newGrpcContextForServiceAccountWithRules(
		t, svcAcctName, nsToRules)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel = context.WithTimeout(grpcCtx3, defaultContextTimeout)
	defer cancel()

	_, err = fluxPluginPackagesClient.DeleteInstalledPackage(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
}

func createAndWaitForHelmRelease(
	t *testing.T,
	tc integrationTestCreatePackageSpec,
	fluxPluginPackagesClient fluxplugin.FluxV2PackagesServiceClient,
	fluxPluginReposClient fluxplugin.FluxV2RepositoriesServiceClient,
	grpcContext context.Context) *corev1.InstalledPackageReference {

	availablePackageRef := tc.request.AvailablePackageRef
	idParts := strings.Split(availablePackageRef.Identifier, "/")
	name := types.NamespacedName{
		Name:      idParts[0],
		Namespace: availablePackageRef.Context.Namespace,
	}

	secretName := ""
	if tc.repoSecret != nil {
		secretName = tc.repoSecret.Name

		if err := kubeCreateSecretAndCleanup(t, tc.repoSecret); err != nil {
			t.Fatal(err)
		}
	}

	setUserManagedSecretsAndCleanup(t, fluxPluginReposClient, true)

	err := kubeAddHelmRepositoryAndCleanup(t, name, tc.repoType, tc.repoUrl, secretName, tc.repoInterval)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// need to wait until repo is indexed by flux plugin
	const maxWait = 25
	for i := 0; i <= maxWait; i++ {
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()

		resp, err := fluxPluginPackagesClient.GetAvailablePackageDetail(
			grpcContext,
			&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
		if err == nil {
			break
		} else if i == maxWait {
			if repo, err2 := kubeGetHelmRepository(t, name); err2 == nil && repo != nil {
				t.Fatalf("Timed out waiting for available package [%s], last response: %v, last error: [%v],\nhelm repository:%s",
					availablePackageRef, resp, err, common.PrettyPrint(repo))
			} else {
				t.Fatalf("Timed out waiting for available package [%s], last response: %v, last error: [%v]",
					availablePackageRef, resp, err)
			}
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

		if !tc.dontCreateNs {
			// per https://github.com/vmware-tanzu/kubeapps/pull/3640#issuecomment-950383123
			if err := kubeCreateNamespaceAndCleanup(t, tc.request.TargetContext.Namespace); err != nil {
				t.Fatal(err)
			}
		}
	}

	if tc.request.ReconciliationOptions != nil && tc.request.ReconciliationOptions.ServiceAccountName != "" {
		svcAcctName := types.NamespacedName{
			Name:      tc.request.ReconciliationOptions.ServiceAccountName,
			Namespace: tc.request.TargetContext.Namespace,
		}
		_, err = kubeCreateAdminServiceAccount(t, svcAcctName)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		// it appears that if service account is deleted before the helmrelease object that uses it,
		// when you try to delete the helmrelease, the "delete" operation gets stuck and the only
		// way to get it "unstuck" is to edit the CRD and remove the finalizer.
		// So we'll cleanup the service account only after the corresponding helmrelease has been deleted
		t.Cleanup(func() {
			if !tc.expectInstallFailure {
				name := types.NamespacedName{
					Name:      tc.expectedDetail.InstalledPackageRef.Identifier,
					Namespace: tc.expectedDetail.InstalledPackageRef.Context.Namespace,
				}
				for i := 0; i < 20; i++ {
					exists, _ := kubeExistsHelmRelease(t, name)
					if exists {
						t.Logf("Waiting for garbage cleanup for [%s]...", name)
						time.Sleep(300 * time.Millisecond)
					} else {
						break
					}
				}
			}

			name := types.NamespacedName{
				Name:      tc.request.ReconciliationOptions.ServiceAccountName,
				Namespace: tc.request.TargetContext.Namespace,
			}
			err := kubeDeleteServiceAccountWithClusterRoleBinding(t, name)
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
	resp, err := fluxPluginPackagesClient.CreateInstalledPackage(ctx, tc.request)
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
				Interval: "1m",
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
			name := types.NamespacedName{
				Name:      installedPackageRef.Identifier,
				Namespace: installedPackageRef.Context.Namespace,
			}
			err = kubeDeleteHelmRelease(t, name)
			if err != nil {
				t.Logf("Failed to delete helm release due to [%v]", err)
			}
		})
	}

	actualDetailResp, actualRefsResp := waitUntilInstallCompletes(t, fluxPluginPackagesClient, grpcContext, installedPackageRef, tc.expectInstallFailure)

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
	fluxPluginPackagesClient fluxplugin.FluxV2PackagesServiceClient,
	grpcContext context.Context,
	installedPackageRef *corev1.InstalledPackageReference,
	expectInstallFailure bool) (
	actualDetailResp *corev1.GetInstalledPackageDetailResponse,
	actualRefsResp *corev1.GetInstalledPackageResourceRefsResponse) {
	var i, maxRetries = 0, 30
	for ; i < maxRetries; i++ {
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		resp2, err := fluxPluginPackagesClient.GetInstalledPackageDetail(
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
			resp2.InstalledPackageDetail.Status.Reason, resp2.InstalledPackageDetail.Status.UserReason, i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if i == maxRetries {
		t.Fatalf("Timed out waiting for task to complete")
	} else if actualDetailResp == nil {
		t.Fatalf("Unexpected state: actual detail response is nil")
	} else if actualDetailResp.InstalledPackageDetail.Status.Ready {
		t.Logf("Install succeeded. Now getting installed package resource refs for [%s]...", installedPackageRef.Identifier)
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		var err error
		actualRefsResp, err = fluxPluginPackagesClient.GetInstalledPackageResourceRefs(
			grpcContext,
			&corev1.GetInstalledPackageResourceRefsRequest{InstalledPackageRef: installedPackageRef})
		if err != nil {
			t.Fatalf("%+v", err)
		}
	} else {
		t.Logf("Install of [%s/%s] completed with [%s], userReason: [%s]",
			installedPackageRef.Context.Namespace,
			installedPackageRef.Identifier,
			actualDetailResp.InstalledPackageDetail.Status.Reason,
			actualDetailResp.InstalledPackageDetail.Status.UserReason,
		)
	}
	return actualDetailResp, actualRefsResp
}
