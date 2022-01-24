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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	fluxplugin "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"golang.org/x/sync/semaphore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"
)

// This is an integration test: it tests the full integration of flux plugin with flux back-end
// To run these tests, enable ENABLE_FLUX_INTEGRATION_TESTS variable
// pre-requisites for these tests to run:
// 1) kind cluster with flux deployed
// 2) kubeapps apis apiserver service running with fluxv2 plug-in enabled, port forwarded to 8080, e.g.
//      kubectl -n kubeapps port-forward svc/kubeapps-internal-kubeappsapis 8080:8080
// 3) run './kind-cluster-setup.sh deploy' once prior to these tests

const (
	// This is local copy of the first few entries
	// on "https://stefanprodan.github.io/podinfo/index.yaml" as of Sept 10 2021 with the chart
	// urls modified to link to .tgz files also within the local cluster.
	// If we want other repos, we'll have add directories and tinker with ./Dockerfile and NGINX conf.
	// This relies on fluxv2plugin-testdata-svc service stood up by testdata/kind-cluster-setup.sh
	podinfo_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo"

	// same as above but requires HTTP basic authentication: user: foo, password: bar
	podinfo_basic_auth_repo_url = "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth"
)

type integrationTestCreateSpec struct {
	testName          string
	repoUrl           string
	request           *corev1.CreateInstalledPackageRequest
	expectedDetail    *corev1.InstalledPackageDetail
	expectedPodPrefix string
	// what follows are boolean flags to test various negative scenarios
	// different from expectedStatusCode due to async nature of install
	expectInstallFailure bool
	noPreCreateNs        bool
	noCleanup            bool
	expectedStatusCode   codes.Code
}

func TestKindClusterCreateInstalledPackage(t *testing.T) {
	fluxPluginClient := checkEnv(t)

	testCases := []integrationTestCreateSpec{
		{
			testName:           "create test (simplest case)",
			repoUrl:            podinfo_repo_url,
			request:            create_request_basic,
			expectedDetail:     expected_detail_basic,
			expectedPodPrefix:  "@TARGET_NS@-my-podinfo-",
			expectedStatusCode: codes.OK,
		},
		{
			testName:           "create package (semver constraint)",
			repoUrl:            podinfo_repo_url,
			request:            create_request_semver_constraint,
			expectedDetail:     expected_detail_semver_constraint,
			expectedPodPrefix:  "@TARGET_NS@-my-podinfo-2-",
			expectedStatusCode: codes.OK,
		},
		{
			testName:           "create package (reconcile options)",
			repoUrl:            podinfo_repo_url,
			request:            create_request_reconcile_options,
			expectedDetail:     expected_detail_reconcile_options,
			expectedPodPrefix:  "@TARGET_NS@-my-podinfo-3-",
			expectedStatusCode: codes.OK,
		},
		{
			testName:           "create package (with values)",
			repoUrl:            podinfo_repo_url,
			request:            create_request_with_values,
			expectedDetail:     expected_detail_with_values,
			expectedPodPrefix:  "@TARGET_NS@-my-podinfo-4-",
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.Internal,
		},
	}

	grpcContext := newGrpcAdminContext(t, "test-create-admin")

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
	unauthorized              bool
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
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "update unauthorized test",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_7,
				expectedDetail:    expected_detail_podinfo_7,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-11-",
			},
			request:      update_request_6,
			unauthorized: true,
		},
		// TODO (gfichtenholt) test automatic upgrade to new version when it becomes available
	}

	grpcContext := newGrpcAdminContext(t, "test-create-admin")

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {

			installedRef := createAndWaitForHelmRelease(
				t, tc.integrationTestCreateSpec, fluxPluginClient, grpcContext)
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

			actualRespAfterUpdate := waitUntilInstallCompletes(t, fluxPluginClient, grpcContext, installedRef, false)

			tc.expectedDetailAfterUpdate.InstalledPackageRef = installedRef
			tc.expectedDetailAfterUpdate.Name = tc.integrationTestCreateSpec.request.Name
			tc.expectedDetailAfterUpdate.ReconciliationOptions = &corev1.ReconciliationOptions{
				Interval: 60,
			}
			tc.expectedDetailAfterUpdate.AvailablePackageRef = tc.integrationTestCreateSpec.request.AvailablePackageRef
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

type integrationTestDeleteSpec struct {
	integrationTestCreateSpec
	unauthorized bool
}

func TestKindClusterDeleteInstalledPackage(t *testing.T) {
	fluxPluginClient := checkEnv(t)

	testCases := []integrationTestDeleteSpec{
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "delete test (simplest case)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_for_delete_1,
				expectedDetail:    expected_detail_podinfo_for_delete_1,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-12-",
				noCleanup:         true,
			},
		},
		{
			integrationTestCreateSpec: integrationTestCreateSpec{
				testName:          "delete test (unauthorized)",
				repoUrl:           podinfo_repo_url,
				request:           create_request_podinfo_for_delete_2,
				expectedDetail:    expected_detail_podinfo_for_delete_2,
				expectedPodPrefix: "@TARGET_NS@-my-podinfo-13-",
				noCleanup:         true,
			},
			unauthorized: true,
		},
	}

	grpcContext := newGrpcAdminContext(t, "test-delete-admin")

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			installedRef := createAndWaitForHelmRelease(t, tc.integrationTestCreateSpec, fluxPluginClient, grpcContext)

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

// this integration test is meant to test a scenario when the redis cache is confiured with maxmemory
// too small to be able to fit all the repos needed to satisfy the request for GetAvailablePackageSummaries
// and redis cache eviction kicks in. Also, the kubeapps-apis pod should have a large memory limit (1Gb) set
// To set up such environment one can use  "-f ./docs/user/manifests/kubeapps-local-dev-redis-tiny-values.yaml"
// option when installing kubeapps via "helm upgrade"
// It is worth noting that exactly how many copies of bitnami repo can be held in the cache at any given time varies
// This is because the size of the index.yaml we get from bitnami does fluctuate quite a bit over time:
// [kubeapps]$ ls -l bitnami_index.yaml
// -rw-r--r--@ 1 gfichtenholt  staff  8432962 Jun 20 02:35 bitnami_index.yaml
// [kubeapps]$ ls -l bitnami_index.yaml
// -rw-rw-rw-@ 1 gfichtenholt  staff  10394218 Nov  7 19:41 bitnami_index.yaml
// Also now we are caching helmcharts themselves for each repo so that will affect how many will fit too
func TestKindClusterGetAvailablePackageSummariesForLargeReposAndTinyRedis(t *testing.T) {
	fluxPlugin := checkEnv(t)

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
	t.Cleanup(func() {
		t.Logf("Resetting notify-keyspace-events")
		if err = redisCli.ConfigSet(redisCli.Context(), "notify-keyspace-events", "").Err(); err != nil {
			t.Logf("%v", err)
		}
	})

	if err = initNumberOfChartsInBitnamiCatalog(t); err != nil {
		t.Errorf("Failed to get number of charts in bitnami catalog due to: %v", err)
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
			repo := fmt.Sprintf("bitnami-%d", totalRepos)
			// this is to make sure we allow enough time for repository to be created and come to ready state
			if err = kubeCreateHelmRepository(t, repo, "https://charts.bitnami.com/bitnami", "default", ""); err != nil {
				t.Fatalf("%v", err)
			}
			t.Cleanup(func() {
				if err = kubeDeleteHelmRepository(t, repo, "default"); err != nil {
					t.Logf("%v", err)
				}
			})
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
	grpcContext := newGrpcAdminContext(t, "test-create-admin")

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
			repo := fmt.Sprintf("bitnami-%d", totalRepos)
			// this is to make sure we allow enough time for repository to be created and come to ready state
			if err = kubeCreateHelmRepository(t, repo, "https://charts.bitnami.com/bitnami", "default", ""); err != nil {
				t.Fatalf("%v", err)
			}
			t.Cleanup(func() {
				if err = kubeDeleteHelmRepository(t, repo, "default"); err != nil {
					t.Logf("%v", err)
				}
			})
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

	grpcContext, cancel := context.WithTimeout(grpcContext, 60*time.Second)
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

// this test is testing a scenario when a repo that takes a long time to index is added
// and while the indexing is in progress this repo is deleted by another request.
// The goal is to make sure that the events are processed by the cache fully in the order
// they were received and the cache does not end up in inconsistent state
func TestKindClusterAddThenDeleteRepo(t *testing.T) {
	_ = checkEnv(t)

	redisCli, err := newRedisClientForIntegrationTest(t)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// now load some large repos (bitnami)
	// I didn't want to store a large (10MB) copy of bitnami repo in our git,
	// so for now let it fetch from bitnami website
	if err = kubeCreateHelmRepository(t, "bitnami-1", "https://charts.bitnami.com/bitnami", "default", ""); err != nil {
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
	fluxPluginClient := checkEnv(t)

	secretName := "podinfo-basic-auth-secret"
	repoName := "podinfo-basic-auth"

	if err := kubeCreateBasicAuthSecret(t, "default", secretName, "foo", "bar"); err != nil {
		t.Fatalf("%v", err)
	}
	t.Cleanup(func() {
		err := kubeDeleteSecret(t, "default", secretName)
		if err != nil {
			t.Logf("Failed to delete helm repository due to [%v]", err)
		}
	})

	if err := kubeCreateHelmRepository(t, repoName, podinfo_basic_auth_repo_url, "default", secretName); err != nil {
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
	} else if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("GetAvailablePackageDetailRequest expected Unauthenticated got %v", err)
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

func createAndWaitForHelmRelease(t *testing.T, tc integrationTestCreateSpec, fluxPluginClient fluxplugin.FluxV2PackagesServiceClient, grpcContext context.Context) *corev1.InstalledPackageReference {
	availablePackageRef := tc.request.AvailablePackageRef
	idParts := strings.Split(availablePackageRef.Identifier, "/")
	err := kubeCreateHelmRepository(t, idParts[0], tc.repoUrl, availablePackageRef.Context.Namespace, "")
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
			kubeCreateNamespace(t, tc.request.TargetContext.Namespace)
			t.Cleanup(func() {
				err = kubeDeleteNamespace(t, tc.request.TargetContext.Namespace)
				if err != nil {
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
			err := kubeDeleteServiceAccount(t, tc.request.ReconciliationOptions.ServiceAccountName, tc.request.TargetContext.Namespace)
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
		t.Logf("Waiting 2s until install completes due to: [%s], userReason: [%s], attempt: [%d/%d]...",
			resp2.InstalledPackageDetail.Status.Reason, resp2.InstalledPackageDetail.Status.UserReason, i+1, maxWait)
		time.Sleep(2 * time.Second)
	}

	if actualResp == nil {
		t.Fatalf("Timed out waiting for task to complete")
	}
	return actualResp
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-2 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-3 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-4 8080:9898\n",
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-6 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-6 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-7 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-7 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-8 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-8 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-9 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-9 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-10 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-11 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-12 8080:9898\n",
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
			"kubectl -n @TARGET_NS@ port-forward deploy/@TARGET_NS@-my-podinfo-13 8080:9898\n",
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
