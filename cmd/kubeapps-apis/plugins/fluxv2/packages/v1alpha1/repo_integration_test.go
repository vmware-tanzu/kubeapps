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
	if err = kubeAddHelmRepository(t, "bitnami-1", "https://charts.bitnami.com/bitnami", "default", ""); err != nil {
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
	fluxPluginClient, _ := checkEnv(t)

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

	if err := kubeAddHelmRepository(t, repoName, podinfo_basic_auth_repo_url, "default", secretName); err != nil {
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

	grpcContext := newGrpcAdminContext(t, "test-create-admin-basic-auth")

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
	ctx, cancel := context.WithTimeout(newGrpcFluxPluginContext(t, fluxPluginServiceAccount), defaultContextTimeout)
	defer cancel()
	_, err := fluxPluginClient.GetAvailablePackageDetail(
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

type integrationTestAddRepoSpec struct {
	testName                 string
	request                  *corev1.AddPackageRepositoryRequest
	existingSecret           *apiv1.Secret
	expectedResponse         *corev1.AddPackageRepositoryResponse
	expectedStatusCode       codes.Code
	expectedReconcileFailure bool
}

func TestKindClusterAddPackageRepository(t *testing.T) {
	_, fluxPluginReposClient := checkEnv(t)

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

	grpcContext := newGrpcAdminContext(t, "test-add-repo-admin")

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
