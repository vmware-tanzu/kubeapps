// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

// scenario:
// 1) create two namespaces
// 2) create two helm repositories, each containing a single package: one in each of the namespaces from (1)
// 3) create 3 service-accounts in default namespace:
//   a) - "...-admin", with cluster-wide access
//   b) - "...-loser", without cluster-wide access or any access to any of the namespaces
//   c) - "...-limited", without cluster-wide access, but with read access to one namespace
// 4) execute GetAvailablePackageSummaries():
//   a) with 3a) => should return 2 packages
//   b) with 3b) => should return 0 packages
//   c) with 3c) => should return 1 package
// 5) execute GetAvailablePackageDetail():
//   a) with 3a) => should work 2 times
//   b) with 3b) => should fail 2 times with PermissionDenied error
//   c) with 3c) => should fail once and work once
// 6) execute GetAvailablePackageVersions():
//   a) with 3a) => should work 2 times
//   b) with 3b) => should fail 2 times with PermissionDenied error
//   c) with 3c) => should fail once and work once

func TestKindClusterRepoRBAC(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	names := []types.NamespacedName{
		{Name: "podinfo-1", Namespace: "test-" + randSeq(4)},
		{Name: "podinfo-2", Namespace: "test-" + randSeq(4)},
	}

	for _, n := range names {
		nm, ns := n.Name, n.Namespace
		if err := kubeCreateNamespace(t, ns); err != nil {
			t.Fatal(err)
		} else if err = kubeAddHelmRepository(t, nm, podinfo_repo_url, ns, "", 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err = kubeDeleteHelmRepository(t, nm, ns); err != nil {
				t.Logf("Failed to delete helm source due to [%v]", err)
			} else if err := kubeDeleteNamespace(t, ns); err != nil {
				t.Logf("Failed to delete namespace [%s] due to [%v]", ns, err)
			}
		})
		// wait until this repo reaches 'Ready'
		if err = kubeWaitUntilHelmRepositoryIsReady(t, nm, ns); err != nil {
			t.Fatal(err)
		}
	}

	grpcCtxAdmin, err := newGrpcAdminContext(t, "test-repo-rbac-admin", "default")
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range names {
		out := kubectlCanIgetHelmRepositoriesInNamespace(t, "test-repo-rbac-admin", "default", n.Namespace)
		if out != "yes" {
			t.Errorf("Expected [yes], got [%s]", out)
		}
	}

	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, "test-repo-rbac-loser", "default")
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		out := kubectlCanIgetHelmRepositoriesInNamespace(t, "test-repo-rbac-loser", "default", n.Namespace)
		if out != "no" {
			t.Errorf("Expected [no], got [%s]", out)
		}
	}

	grpcCtxLimited, err := newGrpcContextForServiceAccountWithAccessToNamespace(t, "test-repo-rbac-limited", "default", names[1].Namespace)
	if err != nil {
		t.Fatal(err)
	}
	for i, n := range names {
		out := kubectlCanIgetHelmRepositoriesInNamespace(t, "test-repo-rbac-limited", "default", n.Namespace)
		if i == 0 {
			if out != "no" {
				t.Errorf("Expected [no], got [%s]", out)
			}
		} else {
			if out != "yes" {
				t.Errorf("Expected [yes], got [%s]", out)
			}
		}
	}

	grpcCtx, cancel := context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
	defer cancel()
	resp, err := fluxPluginClient.GetAvailablePackageSummaries(
		grpcCtx,
		&corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{},
		})
	if err != nil {
		t.Fatal(err)
	} else if len(resp.AvailablePackageSummaries) != 2 {
		t.Errorf("Expected 2 packages, got %s", common.PrettyPrint(resp))
	}

	for _, n := range names {
		grpcCtx, cancel = context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
		defer cancel()
		resp2, err := fluxPluginClient.GetAvailablePackageDetail(
			grpcCtx,
			&corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if err != nil {
			t.Fatal(err)
		} else if resp2.AvailablePackageDetail.SourceUrls[0] != "https://github.com/stefanprodan/podinfo" {
			t.Errorf("Unexpected response: %s", common.PrettyPrint(resp2))
		}

		grpcCtx, cancel = context.WithTimeout(grpcCtxAdmin, defaultContextTimeout)
		defer cancel()
		resp3, err := fluxPluginClient.GetAvailablePackageVersions(
			grpcCtx,
			&corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if err != nil {
			t.Fatal(err)
		} else if len(resp3.PackageAppVersions) != 2 {
			t.Errorf("Unexpected response: %s", common.PrettyPrint(resp3))
		}
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
	defer cancel()
	resp, err = fluxPluginClient.GetAvailablePackageSummaries(
		grpcCtx,
		&corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{},
		})
	if err != nil {
		t.Fatal(err)
	} else if len(resp.AvailablePackageSummaries) != 0 {
		t.Errorf("Expected 0 packages, got %s", common.PrettyPrint(resp))
	}

	for _, n := range names {
		grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
		defer cancel()
		_, err := fluxPluginClient.GetAvailablePackageDetail(
			grpcCtx,
			&corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if status.Code(err) != codes.PermissionDenied {
			t.Fatalf("Expected PermissionDenied error, got %v", err)
		}

		grpcCtx, cancel = context.WithTimeout(grpcCtxLoser, defaultContextTimeout)
		defer cancel()
		_, err = fluxPluginClient.GetAvailablePackageVersions(
			grpcCtx,
			&corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if status.Code(err) != codes.PermissionDenied {
			t.Fatalf("Expected PermissionDenied error, got %v", err)
		}
	}

	grpcCtx, cancel = context.WithTimeout(grpcCtxLimited, defaultContextTimeout)
	defer cancel()
	resp, err = fluxPluginClient.GetAvailablePackageSummaries(
		grpcCtx,
		&corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{},
		})
	if err != nil {
		t.Fatal(err)
	} else if len(resp.AvailablePackageSummaries) != 1 {
		t.Errorf("Unexpected response: %s", common.PrettyPrint(resp))
	}

	for i, n := range names {
		grpcCtx, cancel = context.WithTimeout(grpcCtxLimited, defaultContextTimeout)
		defer cancel()
		resp2, err := fluxPluginClient.GetAvailablePackageDetail(
			grpcCtx,
			&corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if i == 0 {
			if status.Code(err) != codes.PermissionDenied {
				t.Fatalf("Expected PermissionDenied error, got %v", err)
			}
		} else if resp2.AvailablePackageDetail.SourceUrls[0] != "https://github.com/stefanprodan/podinfo" {
			t.Errorf("Unexpected response: %s", common.PrettyPrint(resp2))
		}

		grpcCtx, cancel = context.WithTimeout(grpcCtxLimited, defaultContextTimeout)
		defer cancel()
		resp3, err := fluxPluginClient.GetAvailablePackageVersions(
			grpcCtx,
			&corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: n.Namespace},
					Identifier: n.Name + "/podinfo",
				},
			})
		if i == 0 {
			if status.Code(err) != codes.PermissionDenied {
				t.Fatalf("Expected PermissionDenied error, got %v", err)
			}
		} else if len(resp3.PackageAppVersions) != 2 {
			t.Errorf("UnexpectedResponse: %s", common.PrettyPrint(resp3))
		}
	}
}

type integrationTestAddRepoSpec struct {
	testName                 string
	request                  *corev1.AddPackageRepositoryRequest
	existingSecret           *apiv1.Secret
	expectedResponse         *corev1.AddPackageRepositoryResponse
	expectedStatusCode       codes.Code
	expectedReconcileFailure bool
}

func TestKindClusterAddPackageRepository(t *testing.T) {
	_, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	ca, pub, priv := getCertsForTesting(t)

	testCases := []integrationTestAddRepoSpec{
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

// global vars
var (
	available_package_summaries_podinfo_basic_auth = &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
			{
				Name:                "podinfo",
				AvailablePackageRef: availableRef("podinfo-basic-auth/podinfo", "default"),
				LatestVersion:       &corev1.PackageAppVersion{PkgVersion: "6.0.0", AppVersion: "6.0.0"},
				DisplayName:         "podinfo",
				ShortDescription:    "Podinfo Helm chart for Kubernetes",
				Categories:          []string{""},
			},
		},
	}

	expected_detail_podinfo_basic_auth = &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: &corev1.AvailablePackageDetail{
			AvailablePackageRef: availableRef("podinfo-basic-auth/podinfo", "default"),
			Name:                "podinfo",
			Version:             &corev1.PackageAppVersion{PkgVersion: "6.0.0", AppVersion: "6.0.0"},
			RepoUrl:             "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth",
			HomeUrl:             "https://github.com/stefanprodan/podinfo",
			DisplayName:         "podinfo",
			ShortDescription:    "Podinfo Helm chart for Kubernetes",
			SourceUrls:          []string{"https://github.com/stefanprodan/podinfo"},
			Maintainers: []*corev1.Maintainer{
				{Name: "stefanprodan", Email: "stefanprodan@users.noreply.github.com"},
			},
			Readme:        "Podinfo is used by CNCF projects like [Flux](https://github.com/fluxcd/flux2)",
			DefaultValues: "Default values for podinfo.\n\nreplicaCount: 1\n",
		},
	}
)
