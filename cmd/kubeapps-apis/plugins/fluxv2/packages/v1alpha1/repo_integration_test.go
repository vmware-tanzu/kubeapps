// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
// 3) run './kind-cluster-setup.sh deploy' once prior to these tests

// this test is testing a scenario when a repo that takes a long time to index is added
// and while the indexing is in progress this repo is deleted by another request.
// The goal is to make sure that the events are processed by the cache fully in the order
// they were received and the cache does not end up in inconsistent state
func TestKindClusterAddThenDeleteRepo(t *testing.T) {
	checkEnv(t)

	redisCli, err := newRedisClientForIntegrationTest(t)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// now load some large repos (bitnami)
	// I didn't want to store a large (10MB) copy of bitnami repo in our git,
	// so for now let it fetch from bitnami website
	if err = kubeAddHelmRepository(t, "bitnami-1", "https://charts.bitnami.com/bitnami", "default", "", 0); err != nil {
		t.Fatalf("%v", err)
	}
	// wait until this repo reaches 'Ready' state so that long indexation process kicks in
	if err = kubeWaitUntilHelmRepositoryIsReady(t, "bitnami-1", "default"); err != nil {
		t.Fatalf("%v", err)
	}

	if err = kubeDeleteHelmRepository(t, "bitnami-1", "default"); err != nil {
		t.Fatalf("%v", err)
	}

	t.Logf("Waiting up to 30 seconds...")
	time.Sleep(30 * time.Second)

	if keys, err := redisCli.Keys(redisCli.Context(), "*").Result(); err != nil {
		t.Fatalf("%v", err)
	} else {
		if len(keys) != 0 {
			t.Fatalf("Failing due to unexpected state of the cache. Current keys: %s", keys)
		}
	}
}

func TestKindClusterRepoWithBasicAuth(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	secretName := "podinfo-basic-auth-secret"
	repoName := "podinfo-basic-auth"

	if err := kubeCreateSecret(t, newBasicAuthSecret(secretName, "default", "foo", "bar")); err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		err := kubeDeleteSecret(t, "default", secretName)
		if err != nil {
			t.Logf("Failed to delete helm repository due to [%v]", err)
		}
	})

	if err := kubeAddHelmRepository(t, repoName, podinfo_basic_auth_repo_url, "default", secretName, 0); err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		err := kubeDeleteHelmRepository(t, repoName, "default")
		if err != nil {
			t.Logf("Failed to delete helm repository due to [%v]", err)
		}
	})

	// wait until this repo reaches 'Ready'
	if err := kubeWaitUntilHelmRepositoryIsReady(t, repoName, "default"); err != nil {
		t.Fatalf("%v", err)
	}

	grpcContext, err := newGrpcAdminContext(t, "test-create-admin-basic-auth", "default")
	if err != nil {
		t.Fatal(err)
	}

	const maxWait = 25
	for i := 0; i <= maxWait; i++ {
		grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
		defer cancel()
		resp, err := fluxPluginClient.GetAvailablePackageSummaries(
			grpcContext,
			&corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "default",
				},
			})
		if err == nil {
			opt1 := cmpopts.IgnoreUnexported(
				corev1.GetAvailablePackageSummariesResponse{},
				corev1.AvailablePackageSummary{},
				corev1.AvailablePackageReference{},
				corev1.Context{},
				plugins.Plugin{},
				corev1.PackageAppVersion{})
			opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
			if got, want := resp, available_package_summaries_podinfo_basic_auth; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
			break
		} else if i == maxWait {
			t.Fatalf("Timed out waiting for available package summaries, last response: %v, last error: [%v]", resp, err)
		} else {
			t.Logf("Waiting 2s for repository [%s] to be indexed, attempt [%d/%d]...", repoName, i+1, maxWait)
			time.Sleep(2 * time.Second)
		}
	}

	availablePackageRef := availableRef(repoName+"/podinfo", "default")

	// first try the negative case, no auth - should fail due to not being able to
	// read secrets in all namespaces
	fluxPluginServiceAccount := "test-repo-with-basic-auth"
	grpcCtx, err := newGrpcFluxPluginContext(t, fluxPluginServiceAccount, "default")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(grpcCtx, defaultContextTimeout)
	defer cancel()
	_, err = fluxPluginClient.GetAvailablePackageDetail(
		ctx,
		&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
	if err == nil {
		t.Fatalf("Expected error, did not get one")
	} else if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("GetAvailablePackageDetailRequest expected: PermissionDenied, got: %v", err)
	}

	// this should succeed as it is done in the context of cluster admin
	grpcContext, cancel = context.WithTimeout(grpcContext, defaultContextTimeout)
	defer cancel()
	resp, err := fluxPluginClient.GetAvailablePackageDetail(
		grpcContext,
		&corev1.GetAvailablePackageDetailRequest{AvailablePackageRef: availablePackageRef})
	if err != nil {
		t.Fatalf("%v", err)
	}

	compareActualVsExpectedAvailablePackageDetail(t, resp.AvailablePackageDetail, expected_detail_podinfo_basic_auth.AvailablePackageDetail)
}

func TestKindClusterAddPackageRepository(t *testing.T) {
	_, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	ca, pub, priv := getCertsForTesting(t)

	testCases := []struct {
		testName                 string
		request                  *corev1.AddPackageRepositoryRequest
		existingSecret           *apiv1.Secret
		expectedResponse         *corev1.AddPackageRepositoryResponse
		expectedStatusCode       codes.Code
		expectedReconcileFailure bool
	}{
		{
			testName: "add repo test (simplest case)",
			request: &corev1.AddPackageRepositoryRequest{
				Name:    "my-podinfo",
				Context: &corev1.Context{Namespace: "default"},
				Type:    "helm",
				Url:     podinfo_repo_url,
			},
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   KubeappsCluster,
					},
					Identifier: "my-podinfo",
					Plugin:     fluxPlugin,
				},
			},
			expectedStatusCode: codes.OK,
		},
		{
			testName: "package repository with basic auth",
			request: &corev1.AddPackageRepositoryRequest{
				Name:    "my-podinfo-2",
				Context: &corev1.Context{Namespace: "default"},
				Type:    "helm",
				Url:     podinfo_basic_auth_repo_url,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: "foo",
							Password: "bar",
						},
					},
				},
			},
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   KubeappsCluster,
					},
					Identifier: "my-podinfo-2",
					Plugin:     fluxPlugin,
				},
			},
			expectedStatusCode: codes.OK,
		},
		{
			testName: "package repository with wrong basic auth fails",
			request: &corev1.AddPackageRepositoryRequest{
				Name:    "my-podinfo-3",
				Context: &corev1.Context{Namespace: "default"},
				Type:    "helm",
				Url:     podinfo_basic_auth_repo_url,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: "foo",
							Password: "bar-2",
						},
					},
				},
			},
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   KubeappsCluster,
					},
					Identifier: "my-podinfo-3",
					Plugin:     fluxPlugin,
				},
			},
			expectedStatusCode:       codes.OK,
			expectedReconcileFailure: true,
		},
		{
			testName: "package repository with basic auth and existing secret",
			request: &corev1.AddPackageRepositoryRequest{
				Name:    "my-podinfo-4",
				Context: &corev1.Context{Namespace: "default"},
				Type:    "helm",
				Url:     podinfo_basic_auth_repo_url,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
						SecretRef: &corev1.SecretKeyReference{
							Name: "secret-1",
						},
					},
				},
			},
			existingSecret: newBasicAuthSecret("secret-1", "default", "foo", "bar"),
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   KubeappsCluster,
					},
					Identifier: "my-podinfo-4",
					Plugin:     fluxPlugin,
				},
			},
			expectedStatusCode: codes.OK,
		},
		{
			testName: "package repository with TLS",
			request: &corev1.AddPackageRepositoryRequest{
				Name:    "my-podinfo-4",
				Context: &corev1.Context{Namespace: "default"},
				Type:    "helm",
				Url:     podinfo_tls_repo_url,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
						SecretRef: &corev1.SecretKeyReference{
							Name: "secret-2",
						},
					},
				},
			},
			existingSecret: newTlsSecret("secret-2", "default", pub, priv, ca),
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   KubeappsCluster,
					},
					Identifier: "my-podinfo-4",
					Plugin:     fluxPlugin,
				},
			},
			expectedStatusCode: codes.OK,
		},
	}

	grpcContext, err := newGrpcAdminContext(t, "test-add-repo-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()
			if tc.existingSecret != nil {
				if err := kubeCreateSecret(t, tc.existingSecret); err != nil {
					t.Fatalf("%v", err)
				}
				t.Cleanup(func() {
					err := kubeDeleteSecret(t, tc.existingSecret.Namespace, tc.existingSecret.Name)
					if err != nil {
						t.Logf("Failed to delete secret due to [%v]", err)
					}
				})
			}
			resp, err := fluxPluginReposClient.AddPackageRepository(ctx, tc.request)
			if tc.expectedStatusCode != codes.OK {
				if status.Code(err) != tc.expectedStatusCode {
					t.Fatalf("Expected %v, got: %v", tc.expectedStatusCode, err)
				}
				return // done, nothing more to check
			} else if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				err := kubeDeleteHelmRepository(t, tc.request.Name, tc.request.Context.Namespace)
				if err != nil {
					t.Logf("Failed to delete helm source due to [%v]", err)
				}
			})
			opt1 := cmpopts.IgnoreUnexported(
				corev1.AddPackageRepositoryResponse{},
				corev1.Context{},
				corev1.PackageRepositoryReference{},
				plugins.Plugin{},
			)
			if got, want := resp, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
			}

			// TODO wait for reconcile. To do it properly, we need "R" in CRUD to be
			// designed and implemented
			err = kubeWaitUntilHelmRepositoryIsReady(t, tc.request.Name, tc.request.Context.Namespace)
			if err != nil && !tc.expectedReconcileFailure {
				t.Fatal(err)
			} else if err == nil && tc.expectedReconcileFailure {
				t.Fatalf("Expected error got nil")
			}
		})
	}
}

func TestKindClusterGetPackageRepositoryDetail(t *testing.T) {
	_, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		testName           string
		request            *corev1.GetPackageRepositoryDetailRequest
		repoName           string
		repoUrl            string
		unauthorized       bool
		expectedResponse   *corev1.GetPackageRepositoryDetailResponse
		expectedStatusCode codes.Code
		existingSecret     *apiv1.Secret
	}{
		{
			testName:           "gets detail for podinfo package repository",
			request:            get_repo_detail_req_6,
			repoName:           "my-podinfo",
			repoUrl:            podinfo_repo_url,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_11,
		},
		{
			testName:           "gets detail for bitnami package repository",
			request:            get_repo_detail_req_7,
			repoName:           "my-bitnami",
			repoUrl:            "https://charts.bitnami.com/bitnami",
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_12,
		},
		{
			testName:           "get detail fails for podinfo basic auth package repository without creds",
			request:            get_repo_detail_req_9,
			repoName:           "my-podinfo-2",
			repoUrl:            podinfo_basic_auth_repo_url,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_13,
		},
		{
			testName:           "get detail succeeds for podinfo basic auth package repository with creds",
			request:            get_repo_detail_req_10,
			repoName:           "my-podinfo-3",
			repoUrl:            podinfo_basic_auth_repo_url,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_14,
			existingSecret:     newBasicAuthSecret("secret-1", "TBD", "foo", "bar"),
		},
		{
			testName:           "get detail returns NotFound error for wrong repo",
			request:            get_repo_detail_req_8,
			repoName:           "my-podinfo",
			repoUrl:            podinfo_repo_url,
			expectedStatusCode: codes.NotFound,
		},
		{
			testName:           "get detail returns PermissionDenied error",
			request:            get_repo_detail_req_11,
			repoName:           "my-podinfo-11",
			repoUrl:            podinfo_repo_url,
			expectedStatusCode: codes.PermissionDenied,
			unauthorized:       true,
		},
	}

	grpcAdmin, err := newGrpcAdminContext(t, "test-get-repo-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	grpcLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, "test-get-repo-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repoNamespace := "test-" + randSeq(4)

			if err := kubeCreateNamespace(t, repoNamespace); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := kubeDeleteNamespace(t, repoNamespace); err != nil {
					t.Logf("Failed to delete namespace [%s] due to [%v]", repoNamespace, err)
				}
			})

			secretName := ""
			if tc.existingSecret != nil {
				tc.existingSecret.Namespace = repoNamespace
				if err := kubeCreateSecret(t, tc.existingSecret); err != nil {
					t.Fatalf("%v", err)
				}
				secretName = tc.existingSecret.Name
				t.Cleanup(func() {
					err := kubeDeleteSecret(t, tc.existingSecret.Namespace, tc.existingSecret.Name)
					if err != nil {
						t.Logf("Failed to delete secret due to [%v]", err)
					}
				})
			}

			if err = kubeAddHelmRepository(t, tc.repoName, tc.repoUrl, repoNamespace, secretName, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err = kubeDeleteHelmRepository(t, tc.repoName, repoNamespace); err != nil {
					t.Logf("Failed to delete helm source due to [%v]", err)
				}
			})

			tc.request.PackageRepoRef.Context.Namespace = repoNamespace
			if tc.expectedResponse != nil {
				tc.expectedResponse.Detail.PackageRepoRef.Context.Namespace = repoNamespace
			}

			var grpcCtx context.Context
			var cancel context.CancelFunc
			if tc.unauthorized {
				grpcCtx, cancel = context.WithTimeout(grpcLoser, defaultContextTimeout)
			} else {
				grpcCtx, cancel = context.WithTimeout(grpcAdmin, defaultContextTimeout)
			}
			defer cancel()

			var resp *corev1.GetPackageRepositoryDetailResponse
			for {
				resp, err = fluxPluginReposClient.GetPackageRepositoryDetail(grpcCtx, tc.request)
				if got, want := status.Code(err), tc.expectedStatusCode; got != want {
					t.Fatalf("got: %v, want: %v", err, want)
				}
				if tc.expectedStatusCode != codes.OK {
					// we are done
					return
				}
				if resp.Detail.Status.Reason != corev1.PackageRepositoryStatus_STATUS_REASON_PENDING {
					break
				} else {
					t.Logf("Waiting 2s for repository reconciliation to complete...")
					time.Sleep(2 * time.Second)
				}
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				corev1.PackageRepositoryReference{},
				plugins.Plugin{},
				corev1.GetPackageRepositoryDetailResponse{},
				corev1.PackageRepositoryDetail{},
				corev1.PackageRepositoryStatus{},
				corev1.PackageRepositoryAuth{},
				corev1.PackageRepositoryTlsConfig{},
				corev1.SecretKeyReference{},
			)

			opts2 := cmpopts.IgnoreFields(corev1.PackageRepositoryStatus{}, "UserReason")

			if got, want := resp, tc.expectedResponse; !cmp.Equal(want, got, opts, opts2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts, opts2))
			}

			if !strings.HasPrefix(resp.GetDetail().Status.UserReason, tc.expectedResponse.Detail.Status.UserReason) {
				t.Errorf("unexpected response (status.UserReason): (-want +got):\n- %s\n+ %s",
					tc.expectedResponse.Detail.Status.UserReason,
					resp.GetDetail().Status.UserReason)
			}
		})
	}
}

func TestKindClusterGetPackageRepositorySummaries(t *testing.T) {
	_, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	type repoSpec struct {
		name string
		ns   string
		url  string
	}

	ns1 := "ns1-" + randSeq(4)
	ns2 := "ns2-" + randSeq(4)
	ns3 := "ns3-" + randSeq(4)

	for _, namespace := range []string{ns1, ns2, ns3} {
		ns := namespace
		if err := kubeCreateNamespace(t, ns); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := kubeDeleteNamespace(t, ns); err != nil {
				t.Logf("Failed to delete namespace [%s] due to [%v]", ns, err)
			}
		})
	}

	testCases := []struct {
		testName           string
		request            *corev1.GetPackageRepositorySummariesRequest
		existingRepos      []repoSpec
		unauthorized       bool
		expectedResponse   *corev1.GetPackageRepositorySummariesResponse
		expectedStatusCode codes.Code
	}{
		{
			testName: "admin gets summaries (all namespaces)",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{},
			},
			existingRepos: []repoSpec{
				{name: "podinfo-1", ns: ns1, url: podinfo_repo_url},
				{name: "podinfo-2", ns: ns2, url: podinfo_repo_url},
				{name: "podinfo-3", ns: ns3, url: podinfo_repo_url},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					get_summaries_summary_5("podinfo-1", ns1),
					get_summaries_summary_5("podinfo-2", ns2),
					get_summaries_summary_5("podinfo-3", ns3),
				},
			},
		},
		{
			testName: "admin gets summaries (one namespace)",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Namespace: ns2},
			},
			existingRepos: []repoSpec{
				{name: "podinfo-4", ns: ns1, url: podinfo_repo_url},
				{name: "podinfo-5", ns: ns2, url: podinfo_repo_url},
				{name: "podinfo-6", ns: ns3, url: podinfo_repo_url},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					get_summaries_summary_5("podinfo-5", ns2),
				},
			},
		},
		{
			testName: "loser gets summaries (one namespace => permission denined)",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Namespace: ns2},
			},
			existingRepos: []repoSpec{
				{name: "podinfo-7", ns: ns1, url: podinfo_repo_url},
				{name: "podinfo-8", ns: ns2, url: podinfo_repo_url},
				{name: "podinfo-9", ns: ns3, url: podinfo_repo_url},
			},
			expectedStatusCode: codes.PermissionDenied,
			unauthorized:       true,
		},
		{
			testName: "loser gets summaries (all namespaces => empty list)",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{},
			},
			existingRepos: []repoSpec{
				{name: "podinfo-10", ns: ns1, url: podinfo_repo_url},
				{name: "podinfo-11", ns: ns2, url: podinfo_repo_url},
				{name: "podinfo-12", ns: ns3, url: podinfo_repo_url},
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   &corev1.GetPackageRepositorySummariesResponse{
				// PackageRepositorySummaries: is empty so is omitted
			},
			unauthorized: true,
		},
	}

	grpcAdmin, err := newGrpcAdminContext(t, "test-get-repo-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	grpcLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, "test-get-repo-loser", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			for _, repo := range tc.existingRepos {
				name, namespace := repo.name, repo.ns
				if err = kubeAddHelmRepository(t, name, repo.url, namespace, "", 0); err != nil {
					t.Fatal(err)
				}
				// want to wait until all repos reach Ready state
				kubeWaitUntilHelmRepositoryIsReady(t, name, namespace)
				t.Cleanup(func() {
					if err = kubeDeleteHelmRepository(t, name, namespace); err != nil {
						t.Logf("Failed to delete helm source due to [%v]", err)
					}
				})
			}

			var grpcCtx context.Context
			var cancel context.CancelFunc
			if tc.unauthorized {
				grpcCtx, cancel = context.WithTimeout(grpcLoser, defaultContextTimeout)
			} else {
				grpcCtx, cancel = context.WithTimeout(grpcAdmin, defaultContextTimeout)
			}
			defer cancel()

			resp, err := fluxPluginReposClient.GetPackageRepositorySummaries(grpcCtx, tc.request)
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %v, want: %v", err, want)
			}
			if tc.expectedStatusCode != codes.OK {
				// we are done
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				corev1.PackageRepositoryReference{},
				plugins.Plugin{},
				corev1.PackageRepositoryStatus{},
				corev1.GetPackageRepositorySummariesResponse{},
				corev1.PackageRepositorySummary{},
			)

			if got, want := resp, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts))
			}
		})
	}
}
