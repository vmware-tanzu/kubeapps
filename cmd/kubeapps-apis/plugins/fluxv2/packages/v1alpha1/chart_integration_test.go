// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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
			if err = kubeAddHelmRepository(t, repo, "https://charts.bitnami.com/bitnami", "default", "", 0); err != nil {
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
	grpcContext, err := newGrpcAdminContext(t, "test-create-admin", "default")
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
			repo := fmt.Sprintf("bitnami-%d", totalRepos)
			// this is to make sure we allow enough time for repository to be created and come to ready state
			if err = kubeAddHelmRepository(t, repo, "https://charts.bitnami.com/bitnami", "default", "", 0); err != nil {
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
