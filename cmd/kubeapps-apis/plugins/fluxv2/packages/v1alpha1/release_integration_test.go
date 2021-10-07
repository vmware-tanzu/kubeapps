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
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
//    Didn't want to spend cycles writing port-forwarding code programmatically like https://github.com/anthhub/forwarder
//    at this point.
// 3) run './kind-cluster-setup.sh deploy' once prior to these tests

// TODO (gfichtenholt) currently core server's has broken logic inside plugins.go
// createConfigGetterWithParams(...). Refer to my comment in there. I had to make suggested changes
// locally to make these tests pass

const (
	// the only repo these tests use so far. This is local copy of the first few entries
	// on "https://stefanprodan.github.io/podinfo/index.yaml" as of Sept 10 2021 with the chart
	// urls modified to link to .tgz files also within the local cluster.
	// If we want other repos, we'll have add directories and tinker with ./Dockerfile and NGINX conf.
	// This relies on fluxv2plugin-testdata-svc service stood up by testdata/kind-cluster-setup.sh
	podinfo_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80"
)

type integrationTestCreateSpec struct {
	testName             string
	repoUrl              string
	request              *corev1.CreateInstalledPackageRequest
	expectedDetail       *corev1.InstalledPackageDetail
	expectedPodPrefix    string
	expectInstallFailure bool
	noCleanup            bool
}

func TestKindClusterCreateInstalledPackage(t *testing.T) {
	fluxPluginClient := checkEnv(t)

	testCases := []integrationTestCreateSpec{
		{
			testName:          "create test (simplest case)",
			repoUrl:           podinfo_repo_url,
			request:           create_request_basic,
			expectedDetail:    expected_detail_basic,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-",
		},
		{
			testName:          "create package (semver constraint)",
			repoUrl:           podinfo_repo_url,
			request:           create_request_semver_constraint,
			expectedDetail:    expected_detail_semver_constraint,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-2-",
		},
		{
			testName:          "create package (reconcile options)",
			repoUrl:           podinfo_repo_url,
			request:           create_request_reconcile_options,
			expectedDetail:    expected_detail_reconcile_options,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-3-",
		},
		{
			testName:          "create package (with values)",
			repoUrl:           podinfo_repo_url,
			request:           create_request_with_values,
			expectedDetail:    expected_detail_with_values,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-4-",
		},
		{
			testName:             "install fails",
			repoUrl:              podinfo_repo_url,
			request:              create_request_install_fails,
			expectedDetail:       expected_detail_install_fails,
			expectInstallFailure: true,
		},
		// TODO (gfichtenholt): add a negative test for unauthenticated user
	}

	grpcContext := newGrpcContext(t, "test-create-admin")

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcContext)
		})
	}
}

type integrationTestUpdateSpec struct {
	integrationTestCreateSpec
	request *corev1.UpdateInstalledPackageRequest
	// this is expected AFTER the update call completes
	expectedDetailAfterUpdate *corev1.InstalledPackageDetail
}

func TestKindClusterUpdateInstalledPackage(t *testing.T) {
	fluxPluginClient := checkEnv(t)

	testCases := []integrationTestUpdateSpec{
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update test (simplest case)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_5_2_1,
				expectedDetail:    expected_detail_podinfo_5_2_1,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-6-",
			},
			request:                   update_request_1,
			expectedDetailAfterUpdate: expected_detail_podinfo_6_0_0,
		},
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update test (add values)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_5_2_1_no_values,
				expectedDetail:    expected_detail_podinfo_5_2_1_no_values,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-7-",
			},
			request:                   update_request_2,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values,
		},
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update test (change values)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_5_2_1_values_2,
				expectedDetail:    expected_detail_podinfo_5_2_1_values_2,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-8-",
			},
			request:                   update_request_3,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_3,
		},
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update test (remove values)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_5_2_1_values_4,
				expectedDetail:    expected_detail_podinfo_5_2_1_values_4,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-9-",
			},
			request:                   update_request_4,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_5,
		},
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update test (values dont change)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_5_2_1_values_6,
				expectedDetail:    expected_detail_podinfo_5_2_1_values_6,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-10-",
			},
			request:                   update_request_5,
			expectedDetailAfterUpdate: expected_detail_podinfo_5_2_1_values_6,
		},
		// TODO (gfichtenholt): add a negative test for unauthenticated user
	}

	grpcContext := newGrpcContext(t, "test-update-admin")

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(t, tc.integrationTestCreateSpec, fluxPluginClient, grpcContext)
			tc.request.InstalledPackageRef = installedRef

			_, err := fluxPluginClient.UpdateInstalledPackage(grpcContext, tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			actualRespAfterUpdate := waitUntilInstallCompletes(t, fluxPluginClient, grpcContext, installedRef, false)

			tc.expectedDetailAfterUpdate.PostInstallationNotes = strings.ReplaceAll(
				tc.expectedDetailAfterUpdate.PostInstallationNotes,
				"@TARGET_NS@",
				tc.integrationTestCreateSpec.request.TargetContext.Namespace)

			expectedResp := &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: tc.expectedDetailAfterUpdate,
			}

			compareActualVsExpectedGetInstalledPackageDetailResponse(t, actualRespAfterUpdate, expectedResp)
		})
	}
}

func TestKindClusterDeleteInstalledPackage(t *testing.T) {
	fluxPluginClient := checkEnv(t)

	testCases := []integrationTestCreateSpec{
		{
			testName:          "delete test (simplest case)",
			repoUrl:           podinfo_repo_url,
			request:           create_request_podinfo_for_delete_1,
			expectedDetail:    expected_detail_podinfo_for_delete_1,
			expectedPodPrefix: "@TARGET_NS@-my-podinfo-11-",
			noCleanup:         true,
		},
		// TODO (gfichtenholt): add a negative test for unauthenticated user
	}

	grpcContext := newGrpcContext(t, "test-delete-admin")

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(t, tc, fluxPluginClient, grpcContext)

			_, err := fluxPluginClient.DeleteInstalledPackage(grpcContext, &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: installedRef,
			})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			const maxWait = 25
			for i := 0; i <= maxWait; i++ {
				_, err := fluxPluginClient.GetInstalledPackageDetail(grpcContext, &corev1.GetInstalledPackageDetailRequest{
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

			// sanity check
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
			expectedPodPrefix := strings.ReplaceAll(
				tc.expectedPodPrefix, "@TARGET_NS@", tc.request.TargetContext.Namespace)
			for i := 0; i <= maxWait; i++ {
				if pods, err := kubeGetPodNames(t, tc.request.TargetContext.Namespace); err != nil {
					t.Fatalf("%+v", err)
				} else if len(pods) == 0 {
					break
				} else if len(pods) != 1 {
					t.Errorf("expected 1 pod, got: %s", pods)
				} else if !strings.HasPrefix(pods[0], expectedPodPrefix) {
					t.Errorf("expected pod with prefix [%s] not found in namespace [%s], pods found: [%v]",
						expectedPodPrefix, tc.request.TargetContext.Namespace, pods)
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

func createAndWaitForHelmRelease(t *testing.T, tc integrationTestCreateSpec, fluxPluginClient fluxplugin.FluxV2PackagesServiceClient, grpcContext context.Context) *corev1.InstalledPackageReference {
	availablePackageRef := tc.request.AvailablePackageRef
	idParts := strings.Split(availablePackageRef.Identifier, "/")
	err := kubeCreateHelmRepository(t, idParts[0], tc.repoUrl, availablePackageRef.Context.Namespace)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	t.Cleanup(func() {
		err = kubeDeleteHelmRepository(t, idParts[0], availablePackageRef.Context.Namespace)
		if err != nil {
			t.Logf("Failed to delete helm source due to [%v]", err)
		}
	})

	// need to wait until repo is index by flux plugin
	const maxWait = 25
	for i := 0; i <= maxWait; i++ {
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

	if tc.request.ReconciliationOptions != nil && tc.request.ReconciliationOptions.ServiceAccountName != "" {
		_, err = kubeCreateAdminServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if !tc.noCleanup {
			t.Cleanup(func() {
				err = kubeDeleteServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, "kubeapps")
				if err != nil {
					t.Logf("Failed to delete service account due to [%v]", err)
				}
			})
		}
	}

	// generate a unique target namespace for each test to avoid situations when tests are
	// run multiple times in a row and they fail due to the fact that the specified namespace
	// in in 'Terminating' state
	if tc.request.TargetContext.Namespace != "" {
		tc.request.TargetContext.Namespace += "-" + randSeq(4)
	}

	resp, err := fluxPluginClient.CreateInstalledPackage(grpcContext, tc.request)
	if err != nil {
		t.Fatalf("%+v", err)
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
			err = kubeForceDeleteHelmRelease(t, installedPackageRef.Identifier, installedPackageRef.Context.Namespace)
			if err != nil {
				t.Logf("Failed to delete helm release due to [%v]", err)
			}
		})
	}

	t.Cleanup(func() {
		err = kubeDeleteNamespace(t, tc.request.TargetContext.Namespace)
		if err != nil {
			t.Logf("Failed to delete namespace [%s] due to [%v]", tc.request.TargetContext.Namespace, err)
		}
	})

	actualResp := waitUntilInstallCompletes(t, fluxPluginClient, grpcContext, installedPackageRef, tc.expectInstallFailure)

	tc.expectedDetail.PostInstallationNotes = strings.ReplaceAll(
		tc.expectedDetail.PostInstallationNotes, "@TARGET_NS@", tc.request.TargetContext.Namespace)

	expectedResp := &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: tc.expectedDetail,
	}

	compareActualVsExpectedGetInstalledPackageDetailResponse(t, actualResp, expectedResp)

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
	return installedPackageRef
}

func waitUntilInstallCompletes(t *testing.T, fluxPluginClient fluxplugin.FluxV2PackagesServiceClient, grpcContext context.Context, installedPackageRef *corev1.InstalledPackageReference, expectInstallFailure bool) (actualResp *corev1.GetInstalledPackageDetailResponse) {
	const maxWait = 25
	for i := 0; i <= maxWait; i++ {
		resp2, err := fluxPluginClient.GetInstalledPackageDetail(
			grpcContext,
			&corev1.GetInstalledPackageDetailRequest{InstalledPackageRef: installedPackageRef})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if !expectInstallFailure {
			if resp2.InstalledPackageDetail.Status.Ready == true &&
				resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED {
				actualResp = resp2
				break
			}
		} else {
			if resp2.InstalledPackageDetail.Status.Ready == false &&
				resp2.InstalledPackageDetail.Status.Reason == corev1.InstalledPackageStatus_STATUS_REASON_FAILED {
				actualResp = resp2
				break
			}
		}
		t.Logf("Waiting 1s due to: [%s], userReason: [%s], attempt: [%d/%d]...",
			resp2.InstalledPackageDetail.Status.Reason, resp2.InstalledPackageDetail.Status.UserReason, i+1, maxWait)
		time.Sleep(1 * time.Second)
	}

	if actualResp == nil {
		t.Fatalf("Timed out waiting for task to complete")
	}
	return actualResp
}

// global vars
var (
	create_request_basic = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-1/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-1",
		},
	}

	expected_detail_basic = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Name: "my-podinfo",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-1/podinfo", "default"),
	}

	create_request_semver_constraint = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-2/podinfo", "default"),
		Name:                "my-podinfo-2",
		TargetContext: &corev1.Context{
			Namespace: "test-2",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
	}

	expected_detail_semver_constraint = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-2", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
		Name: "my-podinfo-2",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-2 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-2/podinfo", "default"),
	}

	create_request_reconcile_options = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-3/podinfo", "default"),
		Name:                "my-podinfo-3",
		TargetContext: &corev1.Context{
			Namespace: "test-3",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            false,
			ServiceAccountName: "foo",
		},
	}

	expected_detail_reconcile_options = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-3", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Name: "my-podinfo-3",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-3 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-3/podinfo", "default"),
	}

	create_request_with_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-4/podinfo", "default"),
		Name:                "my-podinfo-4",
		TargetContext: &corev1.Context{
			Namespace: "test-4",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	expected_detail_with_values = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-4", "kubeapps"),
		Name:                "my-podinfo-4",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-4 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-4/podinfo", "default"),
		ValuesApplied:       "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	create_request_install_fails = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-5/podinfo", "default"),
		Name:                "my-podinfo-5",
		TargetContext: &corev1.Context{
			Namespace: "test-5",
		},
		Values: "{\"replicaCount\": \"what we do in the shadows\" }",
	}

	expected_detail_install_fails = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-5", "kubeapps"),
		Name:                "my-podinfo-5",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
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
		AvailablePackageRef: availableRef("podinfo-5/podinfo", "default"),
		ValuesApplied:       "{\"replicaCount\":\"what we do in the shadows\"}",
	}

	create_request_podinfo_5_2_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
		Name:                "my-podinfo-6",
		TargetContext: &corev1.Context{
			Namespace: "test-6",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-6", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-6",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-6 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
	}

	expected_detail_podinfo_6_0_0 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-6", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		Name: "my-podinfo-6",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-6 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
	}

	create_request_podinfo_5_2_1_no_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
		Name:                "my-podinfo-7",
		TargetContext: &corev1.Context{
			Namespace: "test-7",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1_no_values = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-7", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-7",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-7 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
	}

	expected_detail_podinfo_5_2_1_values = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-7", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-7",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-7 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
	}

	create_request_podinfo_5_2_1_values_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
		Name:                "my-podinfo-8",
		TargetContext: &corev1.Context{
			Namespace: "test-8",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_2 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-8", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-8",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-8 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
	}

	expected_detail_podinfo_5_2_1_values_3 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-8", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-8",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		ValuesApplied: "{\"ui\":{\"message\":\"Le Bureau des Légendes\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-8 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
	}

	create_request_podinfo_5_2_1_values_4 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
		Name:                "my-podinfo-9",
		TargetContext: &corev1.Context{
			Namespace: "test-9",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_4 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-9", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-9",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-9 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
	}

	expected_detail_podinfo_5_2_1_values_5 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-9", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-9",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-9 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
	}

	create_request_podinfo_5_2_1_values_6 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-10/podinfo", "default"),
		Name:                "my-podinfo-10",
		TargetContext: &corev1.Context{
			Namespace: "test-10",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_6 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-10", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-10",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-10 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-10/podinfo", "default"),
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

	create_request_podinfo_for_delete_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-11/podinfo", "default"),
		Name:                "my-podinfo-11",
		TargetContext: &corev1.Context{
			Namespace: "test-11",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_for_delete_1 = &corev1.InstalledPackageDetail{
		InstalledPackageRef: installedRef("my-podinfo-11", "kubeapps"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Name: "my-podinfo-11",
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 60,
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-11 8080:9898\n",
		AvailablePackageRef: availableRef("podinfo-11/podinfo", "default"),
	}
)
