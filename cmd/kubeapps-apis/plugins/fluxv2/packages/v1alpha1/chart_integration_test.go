// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
// 3) run './integ-test-env.sh deploy' once prior to these tests

// this integration test is meant to test a scenario when the redis cache is confiured with maxmemory
// too small to be able to fit all the repos needed to satisfy the request for GetAvailablePackageSummaries
// and redis cache eviction kicks in. Also, the kubeapps-apis pod should have a large memory limit (1Gb) set
// To set up such environment one can use  "-f ./site/content/docs/latest/reference/manifests/kubeapps-local-dev-redis-tiny-values.yaml"
// option when installing kubeapps via "helm upgrade"
// It is worth noting that exactly how many copies of bitnami repo can be held in the cache at any given time varies
// This is because the size of the index.yaml we get from bitnami does fluctuate quite a bit over time:
// [kubeapps]$ ls -l bitnami_index.yaml
// -rw-r--r--@ 1 gfichtenholt  staff  8432962 Jun 20 02:35 bitnami_index.yaml
// [kubeapps]$ ls -l bitnami_index.yaml
// -rw-rw-rw-@ 1 gfichtenholt  staff  10394218 Nov  7 19:41 bitnami_index.yaml
// Also now we are caching helmcharts themselves for each repo so that will affect how many will fit too
func TestKindClusterGetAvailablePackageSummariesForLargeReposAndTinyRedis(t *testing.T) {
	fluxPlugin, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	redisCli, err := newRedisClientForIntegrationTest(t)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// assume 30Mb redis cache for now. See comment above
	if err = redisCheckTinyMaxMemory(t, redisCli, "31457280"); err != nil {
		t.Fatalf("%v", err)
	}

	// ref https://redis.io/topics/notifications
	if err = redisCli.ConfigSet(redisCli.Context(), "notify-keyspace-events", "EA").Err(); err != nil {
		t.Fatalf("%+v", err)
	}

	if err = usesBitnamiCatalog(t); err != nil {
		t.Fatalf("Failed to get number of charts in bitnami catalog due to: %v", err)
	}

	const MAX_REPOS_NEVER = 100
	var totalRepos = 0
	// ref https://stackoverflow.com/questions/32840687/timeout-for-waitgroup-wait
	evictedRepos := sets.String{}

	// do this part in a func so we can defer subscribe.Close
	func() {
		// ref https://medium.com/nerd-for-tech/redis-getting-notified-when-a-key-is-expired-or-changed-ca3e1f1c7f0a
		subscribe := redisCli.PSubscribe(redisCli.Context(), "__keyevent@0__:*")
		defer subscribe.Close()

		sem := semaphore.NewWeighted(MAX_REPOS_NEVER)
		if err := sem.Acquire(context.Background(), MAX_REPOS_NEVER); err != nil {
			t.Fatalf("%v", err)
		}

		go redisReceiveNotificationsLoop(t, subscribe.Channel(), sem, &evictedRepos)

		// now load some large repos (bitnami)
		// I didn't want to store a large (>10MB) copy of bitnami repo in our git,
		// so for now let it fetch directly from bitnami website
		// we'll keep adding repos one at a time, until we get an event from redis
		// about the first evicted repo entry
		for ; totalRepos < MAX_REPOS_NEVER && evictedRepos.Len() == 0; totalRepos++ {
			repo := types.NamespacedName{
				Name:      fmt.Sprintf("bitnami-%d", totalRepos),
				Namespace: "default",
			}
			// this is to make sure we allow enough time for repository to be created and come to ready state
			if err = kubeAddHelmRepositoryAndCleanup(t, repo, "", in_cluster_bitnami_url, "", 0); err != nil {
				t.Fatalf("%v", err)
			}
			// wait until this repo have been indexed and cached up to 10 minutes
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			defer cancel()
			if err := sem.Acquire(ctx, 1); err != nil {
				t.Fatalf("Timed out waiting for Redis event: %v", err)
			}
		}
		t.Logf("Done with first part of the test, total repos: [%d], evicted repos: [%d]",
			totalRepos, len(evictedRepos))
	}()

	if evictedRepos.Len() == 0 {
		t.Fatalf("Failing because redis did not evict any entries")
	}

	if keys, err := redisCli.Keys(redisCli.Context(), "helmrepositories:*").Result(); err != nil {
		t.Fatalf("%v", err)
	} else {
		// the cache should only big enough to be able to hold at most (totalRepos-1) of the keys
		// one (or more) entries may have been evicted
		if len(keys) > totalRepos-1 {
			t.Fatalf("Expected at most [%d] keys in cache but got: %s", totalRepos-1, keys)
		}
	}

	// one particular code path I'd like to test:
	// make sure that GetAvailablePackageVersions() works w.r.t. a cache entry that's been evicted
	name := types.NamespacedName{
		Name:      "test-create-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcContext, err := newGrpcAdminContext(t, name)
	if err != nil {
		t.Fatal(err)
	}

	// copy the evicted list because before ForEach loop below will modify it in a goroutine
	evictedCopy := sets.StringKeySet(evictedRepos)

	// do this part in a func so we can defer subscribe.Close
	func() {
		subscribe := redisCli.PSubscribe(redisCli.Context(), "__keyevent@0__:*")
		defer subscribe.Close()

		go redisReceiveNotificationsLoop(t, subscribe.Channel(), nil, &evictedRepos)

		for _, k := range evictedCopy.List() {
			name := strings.Split(k, ":")[2]
			t.Logf("Checking apache version in repo [%s]...", name)
			grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()
			resp, err := fluxPlugin.GetAvailablePackageVersions(
				grpcContext, &corev1.GetAvailablePackageVersionsRequest{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context: &corev1.Context{
							Namespace: "default",
						},
						Identifier: name + "/apache",
					},
				})
			if err != nil {
				t.Fatalf("%v", err)
			} else if len(resp.PackageAppVersions) < 5 {
				t.Fatalf("Expected at least 5 versions for apache chart, got: %s", resp)
			}
		}

		t.Logf("Done with second part of the test")
	}()

	// do this part in a func so we can defer subscribe.Close
	func() {
		subscribe := redisCli.PSubscribe(redisCli.Context(), "__keyevent@0__:*")
		defer subscribe.Close()

		// above loop should cause a few more entries to be evicted, but just to be sure let's
		// load a few more copies of bitnami repo into the cache. The goal of this for loop is
		// to force redis to evict more repo(s)
		sem := semaphore.NewWeighted(MAX_REPOS_NEVER)
		if err := sem.Acquire(context.Background(), MAX_REPOS_NEVER); err != nil {
			t.Fatalf("%v", err)
		}
		go redisReceiveNotificationsLoop(t, subscribe.Channel(), sem, &evictedRepos)

		for ; totalRepos < MAX_REPOS_NEVER && evictedRepos.Len() == evictedCopy.Len(); totalRepos++ {
			repo := types.NamespacedName{
				Name:      fmt.Sprintf("bitnami-%d", totalRepos),
				Namespace: "default",
			}
			// this is to make sure we allow enough time for repository to be created and come to ready state
			if err = kubeAddHelmRepositoryAndCleanup(t, repo, "", in_cluster_bitnami_url, "", 0); err != nil {
				t.Fatalf("%v", err)
			}
			// wait until this repo have been indexed and cached up to 10 minutes
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			defer cancel()
			if err := sem.Acquire(ctx, 1); err != nil {
				t.Fatalf("Timed out waiting for Redis event: %v", err)
			}
		}

		t.Logf("Done with third part of the test")
	}()

	if keys, err := redisCli.Keys(redisCli.Context(), "helmrepositories:*").Result(); err != nil {
		t.Fatalf("%v", err)
	} else {
		// the cache should only big enough to be able to hold at most (totalRepos-1) of the keys
		// one (or more) entries MUST have been evicted
		if len(keys) > totalRepos-1 {
			t.Fatalf("Expected at most %d keys in cache but got [%s]", totalRepos-1, keys)
		}
	}

	// not related to low maxmemory but as long as we are here might as well check that
	// there is a Unauthenticated failure when there are no credenitals in the request
	_, err = fluxPlugin.GetAvailablePackageSummaries(context.TODO(), &corev1.GetAvailablePackageSummariesRequest{})
	if err == nil || status.Code(err) != codes.Unauthenticated {
		t.Fatalf("Expected Unauthenticated, got %v", err)
	}

	grpcContext, cancel := context.WithTimeout(grpcContext, 90*time.Second)
	defer cancel()
	resp2, err := fluxPlugin.GetAvailablePackageSummaries(grpcContext, &corev1.GetAvailablePackageSummariesRequest{})
	if err != nil {
		t.Fatalf("%v", err)
	}

	// we need to make sure that response contains packages from all existing repositories
	// regardless whether they're in the cache or not
	expected := sets.String{}
	for i := 0; i < totalRepos; i++ {
		repo := fmt.Sprintf("bitnami-%d", i)
		expected.Insert(repo)
	}
	for _, s := range resp2.AvailablePackageSummaries {
		id := strings.Split(s.AvailablePackageRef.Identifier, "/")
		expected.Delete(id[0])
	}

	if expected.Len() != 0 {
		t.Fatalf("Expected to get packages from these repositories: %s, but did not get any",
			expected.List())
	}
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
// ref https://github.com/vmware-tanzu/kubeapps/issues/4390
func TestKindClusterRepoAndChartRBAC(t *testing.T) {
	fluxPluginClient, _, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	names := []types.NamespacedName{
		{Name: "podinfo-1", Namespace: "test-" + randSeq(4)},
		{Name: "podinfo-2", Namespace: "test-" + randSeq(4)},
	}

	for _, n := range names {
		if err := kubeCreateNamespaceAndCleanup(t, n.Namespace); err != nil {
			t.Fatal(err)
		} else if err = kubeAddHelmRepositoryAndCleanup(t, n, "", podinfo_repo_url, "", 0); err != nil {
			t.Fatal(err)
		}
		// wait until this repo reaches 'Ready'
		if err = kubeWaitUntilHelmRepositoryIsReady(t, n); err != nil {
			t.Fatal(err)
		}
	}

	svcAcctName := types.NamespacedName{
		Name:      "test-repo-rbac-admin-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxAdmin, err := newGrpcAdminContext(t, svcAcctName)
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range names {
		out := kubectlCanI(t, svcAcctName, "get", fluxHelmRepositories, n.Namespace)
		if out != "yes" {
			t.Errorf("Expected [yes], got [%s]", out)
		}
	}

	svcAcctName2 := types.NamespacedName{
		Name:      "test-repo-rbac-loser-" + randSeq(4),
		Namespace: "default",
	}
	grpcCtxLoser, err := newGrpcContextForServiceAccountWithoutAccessToAnyNamespace(t, svcAcctName2)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		out := kubectlCanI(
			t, svcAcctName2, "get", fluxHelmRepositories, n.Namespace)
		if out != "no" {
			t.Errorf("Expected [no], got [%s]", out)
		}
	}

	rules := map[string][]rbacv1.PolicyRule{
		names[1].Namespace: {
			{
				APIGroups: []string{sourcev1.GroupVersion.Group},
				Resources: []string{fluxHelmRepositories},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{sourcev1.GroupVersion.Group},
				Resources: []string{"helmcharts"},
				Verbs:     []string{"get", "list"},
			},
		},
	}

	svcAcctName3 := types.NamespacedName{
		Name:      "test-repo-rbac-limited-" + randSeq(4),
		Namespace: "default",
	}

	grpcCtxLimited, err := newGrpcContextForServiceAccountWithRules(t, svcAcctName3, rules)
	if err != nil {
		t.Fatal(err)
	}
	for i, n := range names {
		out := kubectlCanI(t, svcAcctName3, "get", fluxHelmRepositories, n.Namespace)
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
		resp, err = fluxPluginClient.GetAvailablePackageSummaries(
			grpcCtx,
			&corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: n.Namespace,
				},
			})
		if err != nil {
			t.Fatal(err)
		} else if len(resp.AvailablePackageSummaries) != 1 {
			t.Errorf("Unexpected response: %s", common.PrettyPrint(resp))
		}

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

type testCaseKindClusterAvailablePackageEndpointsForOCISpec struct {
	testName        string
	registryUrl     string
	secret          *apiv1.Secret
	unauthenticated bool
	unauthorized    bool
}

func testCaseKindClusterAvailablePackageEndpointsForGitHub(t *testing.T) []testCaseKindClusterAvailablePackageEndpointsForOCISpec {
	ghUser := os.Getenv("GITHUB_USER")
	ghToken := os.Getenv("GITHUB_TOKEN")
	if ghUser == "" || ghToken == "" {
		t.Fatalf("Environment variables GITHUB_USER and GITHUB_TOKEN need to be set to run this test")
	}
	return []testCaseKindClusterAvailablePackageEndpointsForOCISpec{
		{
			testName:    "Testing [" + github_stefanprodan_podinfo_oci_registry_url + "] with basic auth secret",
			registryUrl: github_stefanprodan_podinfo_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				ghUser,
				ghToken,
			),
		},
		{
			testName:    "Testing [" + github_stefanprodan_podinfo_oci_registry_url + "] with dockerconfigjson secret",
			registryUrl: github_stefanprodan_podinfo_oci_registry_url,
			secret: newDockerConfigJsonSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				"ghcr.io", ghUser, ghToken,
			),
		},
	}
}

func testCaseKindClusterAvailablePackageEndpointsForHarbor(t *testing.T) []testCaseKindClusterAvailablePackageEndpointsForOCISpec {
	if err := setupHarborStefanProdanClone(t); err != nil {
		t.Fatal(err)
	}

	harborRobotName, harborRobotSecret, err := setupHarborRobotAccount(t)
	if err != nil {
		t.Fatal(err)
	}

	// per agamez request (slack thread 8/31/22)
	harborCorpVMwareHost := os.Getenv("HARBOR_VMWARE_CORP_HOST")
	if harborCorpVMwareHost == "" {
		t.Fatal("Environment variable [HARBOR_VMWARE_CORP_HOST] needs to be set to run this test")
	}
	harborCorpVMwareRepoUrl := "oci://" + harborCorpVMwareHost + "/kubeapps_flux_integration"
	harborCorpVMwareRepoRobotUser := os.Getenv("HARBOR_VMWARE_CORP_ROBOT_USER")
	if harborCorpVMwareRepoRobotUser == "" {
		t.Fatal("Environment variable [HARBOR_VMWARE_CORP_ROBOT_USER] needs to be set to run this test")
	}
	harborCorpVMwareRepoRobotSecret := os.Getenv("HARBOR_VMWARE_CORP_ROBOT_SECRET")
	if harborCorpVMwareRepoRobotSecret == "" {
		t.Fatal("Environment variable [HARBOR_VMWARE_CORP_ROBOT_SECRET] needs to be set to run this test")
	}

	return []testCaseKindClusterAvailablePackageEndpointsForOCISpec{
		{
			testName:    "Testing [" + harbor_stefanprodan_podinfo_oci_registry_url + "] with basic auth secret (admin)",
			registryUrl: harbor_stefanprodan_podinfo_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				harbor_admin_user,
				harbor_admin_pwd,
			),
		},
		{
			testName:    "Testing [" + harbor_stefanprodan_podinfo_oci_registry_url + "] with basic auth secret (robot)",
			registryUrl: harbor_stefanprodan_podinfo_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				harborRobotName,
				harborRobotSecret,
			),
		},
		// harbor private repo (admin)
		{
			testName:    "Testing [" + harbor_stefanprodan_podinfo_private_oci_registry_url + "] with basic auth secret (admin)",
			registryUrl: harbor_stefanprodan_podinfo_private_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				harbor_admin_user,
				harbor_admin_pwd,
			),
		},
		// harbor private repo (robot)
		{
			testName:    "Testing [" + harbor_stefanprodan_podinfo_private_oci_registry_url + "] with basic auth secret (robot)",
			registryUrl: harbor_stefanprodan_podinfo_private_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				harborRobotName,
				harborRobotSecret,
			),
		},
		// harbor private repo (negative test for no secret)
		{
			testName:        "Testing [" + harbor_stefanprodan_podinfo_private_oci_registry_url + "] without secret",
			registryUrl:     harbor_stefanprodan_podinfo_private_oci_registry_url,
			unauthenticated: true,
		},
		// harbor private repo (negative test bad username/secret)
		{
			testName:    "Testing [" + harbor_stefanprodan_podinfo_private_oci_registry_url + "] bad username/secret",
			registryUrl: harbor_stefanprodan_podinfo_private_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				"kaka",
				"kaka"),
			unauthorized: true,
		},
		{
			testName:    "Testing [" + harborCorpVMwareRepoUrl + "]",
			registryUrl: harborCorpVMwareRepoUrl,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				harborCorpVMwareRepoRobotUser,
				harborCorpVMwareRepoRobotSecret),
		},
	}
}

func testCaseKindClusterAvailablePackageEndpointsForGcp(t *testing.T) []testCaseKindClusterAvailablePackageEndpointsForOCISpec {
	// ref: https://cloud.google.com/artifact-registry/docs/helm/authentication#token
	gcpUser := "oauth2accesstoken"
	gcpPasswd, err := gcloudPrintAccessToken(t)
	if err != nil {
		t.Fatal(err)
	}
	// ref https://cloud.google.com/artifact-registry/docs/helm/authentication#json-key
	gcpKeyFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if gcpKeyFile == "" {
		t.Fatalf("Environment variable [GOOGLE_APPLICATION_CREDENTIALS] needs to be set to run this test")
	}

	gcpUser2 := "_json_key"
	gcpServer2 := "us-west1-docker.pkg.dev"
	gcpPasswd2, err := os.ReadFile(gcpKeyFile)
	if err != nil {
		t.Fatal(err)
	}

	return []testCaseKindClusterAvailablePackageEndpointsForOCISpec{
		{
			testName:    "Testing [" + gcp_stefanprodan_podinfo_oci_registry_url + "] with service access token",
			registryUrl: gcp_stefanprodan_podinfo_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				gcpUser,
				string(gcpPasswd),
			),
		},
		{
			testName:    "Testing [" + gcp_stefanprodan_podinfo_oci_registry_url + "] with JSON key",
			registryUrl: gcp_stefanprodan_podinfo_oci_registry_url,
			secret: newDockerConfigJsonSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				gcpServer2,
				gcpUser2,
				string(gcpPasswd2),
			),
		},
		// negative test for no secret
		{
			testName:        "Testing [" + gcp_stefanprodan_podinfo_oci_registry_url + "] without a secret",
			registryUrl:     gcp_stefanprodan_podinfo_oci_registry_url,
			unauthenticated: true,
		},
		// negative test for bad username/secret
		{
			testName:    "Testing [" + gcp_stefanprodan_podinfo_oci_registry_url + "] bad username/secret",
			registryUrl: gcp_stefanprodan_podinfo_oci_registry_url,
			secret: newDockerConfigJsonSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				gcpServer2,
				"kaka",
				"kaka",
			),
			unauthorized: true,
		},
	}
}

func TestKindClusterAvailablePackageEndpointsForOCI(t *testing.T) {
	fluxPluginClient, fluxPluginReposClient, err := checkEnv(t)
	if err != nil {
		t.Fatal(err)
	}

	// this is written this way so its relatively easy to comment out and run just a subset
	// of the test cases when debugging failures
	testCases := []testCaseKindClusterAvailablePackageEndpointsForOCISpec{}
	testCases = append(testCases, testCaseKindClusterAvailablePackageEndpointsForGitHub(t)...)
	testCases = append(testCases, testCaseKindClusterAvailablePackageEndpointsForHarbor(t)...)
	testCases = append(testCases, testCaseKindClusterAvailablePackageEndpointsForGcp(t)...)

	// TODO (gfichtenholt) harbor plainHTTP (not HTTPS) repo with robot account
	//   this may or may not work see https://github.com/fluxcd/source-controller/issues/807
	// TODO (gfichtenholt) TLS secret with CA
	// TODO (gfichtenholt) TLS secret with CA, pub, priv assuming Flux supports it

	/*
		{
			// this gets set up in ./testdata/integ-test-env.sh
			// currently fails with AuthenticationFailed: failed to log into registry
			//  'oci://registry-app-svc.default.svc.cluster.local:5000/helm-charts':
			// Get "https://registry-app-svc.default.svc.cluster.local:5000/v2/":
			// http: server gave HTTP response to HTTPS client
			// the error comes from flux source-controller
			// opened a new issue per souleb's request:
			// https://github.com/fluxcd/source-controller/issues/805

			testName:    "Testing [" + in_cluster_oci_registry_url + "]",
			registryUrl: in_cluster_oci_registry_url,
			secret: newBasicAuthSecret(types.NamespacedName{
				Name:      "oci-repo-secret-" + randSeq(4),
				Namespace: "default"},
				"foo",
				"bar",
			),
		},
	*/

	testKindClusterAvailablePackageEndpointsForOCIHelper(t, testCases, fluxPluginClient, fluxPluginReposClient)
}

func testKindClusterAvailablePackageEndpointsForOCIHelper(
	t *testing.T,
	testCases []testCaseKindClusterAvailablePackageEndpointsForOCISpec,
	fluxPluginClient fluxplugin.FluxV2PackagesServiceClient,
	fluxPluginReposClient fluxplugin.FluxV2RepositoriesServiceClient) {

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repoName := types.NamespacedName{
				Name:      "my-podinfo-" + randSeq(4),
				Namespace: "default",
			}

			secretName := ""
			if tc.secret != nil {
				secretName = tc.secret.Name

				if err := kubeCreateSecretAndCleanup(t, tc.secret); err != nil {
					t.Fatal(err)
				}
			}

			setUserManagedSecretsAndCleanup(t, fluxPluginReposClient, true)

			if err := kubeAddHelmRepositoryAndCleanup(
				t, repoName, "oci", tc.registryUrl, secretName, 0); err != nil {
				t.Fatal(err)
			}
			// wait until this repo reaches 'Ready'
			err := kubeWaitUntilHelmRepositoryIsReady(t, repoName)
			if !tc.unauthorized {
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					if strings.Contains(err.Error(), "AuthenticationFailed: failed to login to registry") {
						return // nothing more to check
					} else {
						t.Fatal(err)
					}
				} else {
					t.Fatal("expected error, got nil")
				}
			}

			adminName := types.NamespacedName{
				Name:      "test-admin-" + randSeq(4),
				Namespace: "default",
			}
			grpcContext, err := newGrpcAdminContext(t, adminName)
			if err != nil {
				t.Fatal(err)
			}

			grpcContext, cancel := context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()

			resp, err := fluxPluginClient.GetAvailablePackageSummaries(
				grpcContext,
				&corev1.GetAvailablePackageSummariesRequest{})
			if err != nil {
				t.Fatal(err)
			}

			opt1 := cmpopts.IgnoreUnexported(
				corev1.GetAvailablePackageSummariesResponse{},
				corev1.AvailablePackageSummary{},
				corev1.AvailablePackageReference{},
				corev1.Context{},
				plugins.Plugin{},
				corev1.PackageAppVersion{})
			opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
			if !tc.unauthenticated {
				if got, want := resp, expected_oci_stefanprodan_podinfo_available_summaries(repoName.Name); !cmp.Equal(got, want, opt1, opt2) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
				}
			} else {
				if got, want := resp, no_available_summaries(repoName.Name); !cmp.Equal(got, want, opt1, opt2) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
				}
				return // nothing more to check
			}

			grpcContext, cancel = context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()
			resp2, err := fluxPluginClient.GetAvailablePackageVersions(
				grpcContext, &corev1.GetAvailablePackageVersionsRequest{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context: &corev1.Context{
							Namespace: "default",
						},
						Identifier: repoName.Name + "/podinfo",
					},
				})
			if err != nil {
				t.Fatal(err)
			}
			opts := cmpopts.IgnoreUnexported(
				corev1.GetAvailablePackageVersionsResponse{},
				corev1.PackageAppVersion{})
			if got, want := resp2, expected_versions_stefanprodan_podinfo; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			grpcContext, cancel = context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()
			resp3, err := fluxPluginClient.GetAvailablePackageDetail(
				grpcContext,
				&corev1.GetAvailablePackageDetailRequest{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context: &corev1.Context{
							Namespace: "default",
						},
						Identifier: repoName.Name + "/podinfo",
					},
				})
			if err != nil {
				t.Fatal(err)
			}

			compareActualVsExpectedAvailablePackageDetail(
				t,
				resp3.AvailablePackageDetail,
				expected_detail_oci_stefanprodan_podinfo(repoName.Name, tc.registryUrl).AvailablePackageDetail)

			// try a few older versions
			grpcContext, cancel = context.WithTimeout(grpcContext, defaultContextTimeout)
			defer cancel()
			resp4, err := fluxPluginClient.GetAvailablePackageDetail(
				grpcContext,
				&corev1.GetAvailablePackageDetailRequest{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context: &corev1.Context{
							Namespace: "default",
						},
						Identifier: repoName.Name + "/podinfo",
					},
					PkgVersion: "6.1.6",
				})
			if err != nil {
				t.Fatal(err)
			}

			compareActualVsExpectedAvailablePackageDetail(
				t,
				resp4.AvailablePackageDetail,
				expected_detail_oci_stefanprodan_podinfo_2(repoName.Name, tc.registryUrl).AvailablePackageDetail)
		})
	}
}
