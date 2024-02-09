// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	authorizationv1 "k8s.io/api/authorization/v1"
	k8stesting "k8s.io/client-go/testing"

	"github.com/bufbuild/connect-go"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

type testSpecGetAvailablePackageSummaries struct {
	name      string
	namespace string
	url       string
	index     string
}

func TestGetAvailablePackageSummariesWithoutPagination(t *testing.T) {
	testCases := []struct {
		name                 string
		request              *corev1.GetAvailablePackageSummariesRequest
		repos                []testSpecGetAvailablePackageSummaries
		expectedResponse     *corev1.GetAvailablePackageSummariesResponse
		expectedErrorCode    connect.Code
		noCrossNamespaceRefs bool
	}{
		{
			name: "it returns a couple of fluxv2 packages from the cluster (no request ns specified)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
			},
			request:          &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: valid_index_available_package_summaries_resp,
		},
		{
			name: "it returns a couple of fluxv2 packages from the cluster (when request namespace is specified)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
			},
			request:          &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "default"}},
			expectedResponse: valid_index_available_package_summaries_resp,
		},
		{
			name: "it returns a couple of fluxv2 packages from the cluster (when request cluster is specified and matches the kubeapps cluster)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			}},
			expectedResponse: valid_index_available_package_summaries_resp,
		},
		{
			name: "it returns all fluxv2 packages from the cluster (when request namespace is does not match repo namespace)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     testYaml("jetstack-index.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "non-default"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: append(valid_index_available_package_summaries, cert_manager_summary),
			},
		},
		{
			name: "uses a filter based on existing repo",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     testYaml("jetstack-index.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Repositories: []string{"jetstack-1"},
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					cert_manager_summary,
				},
			},
		},
		{
			name: "uses a filter based on non-existing repo",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     testYaml("jetstack-index.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Repositories: []string{"jetstack-2"},
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			name: "uses a filter based on existing categories",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Categories: []string{"Analytics"},
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					elasticsearch_summary,
				},
			},
		},
		{
			name: "uses a filter based on existing categories (2)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Categories: []string{"Analytics", "CMS"},
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: index_with_categories_summaries,
			},
		},
		{
			name: "uses a filter based on non-existing categories",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Categories: []string{"Foo"},
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			name: "uses a filter based on existing appVersion",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					AppVersion: "4.7.0",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					ghost_summary,
				},
			},
		},
		{
			name: "uses a filter based on non-existing appVersion",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					AppVersion: "99.99.99",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			name: "uses a filter based on existing pkgVersion",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					PkgVersion: "15.5.0",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					elasticsearch_summary,
				},
			},
		},
		{
			name: "uses a filter based on non-existing pkgVersion",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					PkgVersion: "99.99.99",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			name: "uses a filter based on existing query text (chart name)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Query: "ela",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					elasticsearch_summary,
				},
			},
		},
		{
			name: "uses a filter based on existing query text (chart keywords)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Query: "vascrip",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					ghost_summary,
				},
			},
		},
		{
			name: "uses a filter based on non-existing query text",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "index-with-categories-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("index-with-categories.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{Namespace: "blah"},
				FilterOptions: &corev1.FilterOptions{
					Query: "qwerty",
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			name: "it returns an error if a cluster other than the kubeapps cluster is specified",
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{
				Cluster: "not-kubeapps-cluster",
			}},
			expectedErrorCode: connect.CodeUnimplemented,
		},
		{
			name: "it returns expected fluxv2 packages when noCrossNamespaceRefs flag is set",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     testYaml("valid-index.yaml"),
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     testYaml("jetstack-index.yaml"),
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "ns1"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					cert_manager_summary,
				},
			},
			noCrossNamespaceRefs: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repos := []sourcev1beta2.HelmRepository{}

			for _, rs := range tc.repos {
				ts2, repo, err := newHttpRepoAndServeIndex(rs.index, rs.name, rs.namespace, nil, "")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()
				repos = append(repos, *repo)
			}

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, mock, err := newSimpleServerWithRepos(t, repos)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			if tc.noCrossNamespaceRefs {
				s.pluginConfig.NoCrossNamespaceRefs = true
				for _, r := range repos {
					if r.Namespace == tc.request.Context.Namespace {
						if err = s.redisMockExpectGetFromRepoCache(mock, nil, r); err != nil {
							t.Fatal(err)
						}
					}
				}
			} else {
				if err = s.redisMockExpectGetFromRepoCache(mock, tc.request.FilterOptions, repos...); err != nil {
					t.Fatal(err)
				}
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %v, want: %v", got, want)
			}
			// If an error code was expected, then no need to continue checking
			// the response.
			if tc.expectedErrorCode != 0 {
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			compareAvailablePackageSummaries(t, response.Msg, tc.expectedResponse)
		})
	}
}

func TestGetAvailablePackageSummariesWithPagination(t *testing.T) {
	// one big test case that can't really be broken down to smaller cases because
	// the tests aren't independent/idempotent: there is state that needs to be
	// kept track from one call to the next

	t.Run("test GetAvailablePackageSummaries with pagination", func(t *testing.T) {
		existingRepos := []testSpecGetAvailablePackageSummaries{
			{
				name:      "index-with-categories-1",
				namespace: "default",
				url:       "https://example.repo.com/charts",
				index:     testYaml("index-with-categories.yaml"),
			},
		}
		repos := []sourcev1beta2.HelmRepository{}
		for _, rs := range existingRepos {
			ts2, repo, err := newHttpRepoAndServeIndex(rs.index, rs.name, rs.namespace, nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()
			repos = append(repos, *repo)
		}

		// the index.yaml will contain links to charts but for the purposes
		// of this test they do not matter
		s, mock, err := newSimpleServerWithRepos(t, repos)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repos...); err != nil {
			t.Fatal(err)
		}

		request1 := &corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{Namespace: "blah"},
			PaginationOptions: &corev1.PaginationOptions{
				PageToken: "0",
				PageSize:  1,
			},
		}

		response1, err := s.GetAvailablePackageSummaries(context.Background(), connect.NewRequest(request1))
		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %v, want: %v", got, want)
		}

		opts := cmpopts.IgnoreUnexported(
			corev1.GetAvailablePackageSummariesResponse{},
			corev1.AvailablePackageSummary{},
			corev1.AvailablePackageReference{},
			corev1.Context{},
			plugins.Plugin{},
			corev1.PackageAppVersion{})
		opts2 := cmpopts.SortSlices(lessAvailablePackageFunc)

		expectedResp1 := &corev1.GetAvailablePackageSummariesResponse{
			AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
				elasticsearch_summary,
			},
			NextPageToken: "1",
		}
		expectedResp2 := &corev1.GetAvailablePackageSummariesResponse{
			AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
				ghost_summary,
			},
			NextPageToken: "1",
		}

		match := false
		var nextExpectedResp *corev1.GetAvailablePackageSummariesResponse
		if got, want := response1.Msg, expectedResp1; cmp.Equal(want, got, opts, opts2) {
			match = true
			nextExpectedResp = expectedResp2
			nextExpectedResp.NextPageToken = "2"
		} else if got, want := response1.Msg, expectedResp2; cmp.Equal(want, got, opts, opts2) {
			match = true
			nextExpectedResp = expectedResp1
			nextExpectedResp.NextPageToken = "2"
		}
		if !match {
			t.Fatalf("Expected one of:\n%s\n%s, but got:\n%s", expectedResp1, expectedResp2, response1)
		}

		request2 := &corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{Namespace: "blah"},
			PaginationOptions: &corev1.PaginationOptions{
				PageToken: "1",
				PageSize:  1,
			},
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repos...); err != nil {
			t.Fatal(err)
		}
		response2, err := s.GetAvailablePackageSummaries(context.Background(), connect.NewRequest(request2))
		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %v, want: %v", err, want)
		}
		compareAvailablePackageSummaries(t, response2.Msg, nextExpectedResp)

		request3 := &corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{Namespace: "blah"},
			PaginationOptions: &corev1.PaginationOptions{
				PageToken: "2",
				PageSize:  1,
			},
		}
		nextExpectedResp = &corev1.GetAvailablePackageSummariesResponse{
			AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
			NextPageToken:             "",
		}
		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repos...); err != nil {
			t.Fatal(err)
		}
		response3, err := s.GetAvailablePackageSummaries(context.Background(), connect.NewRequest(request3))
		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %v, want: %v", err, want)
		}
		compareAvailablePackageSummaries(t, response3.Msg, nextExpectedResp)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGetAvailablePackageSummaryAfterRepoIndexUpdate(t *testing.T) {
	t.Run("test get available package summaries after repo index is updated", func(t *testing.T) {
		indexYamlBeforeUpdateBytes, err := os.ReadFile(testYaml("index-before-update.yaml"))
		if err != nil {
			t.Fatalf("%+v", err)
		}

		indexYamlAfterUpdateBytes, err := os.ReadFile(testYaml("index-after-update.yaml"))
		if err != nil {
			t.Fatalf("%+v", err)
		}

		updateHappened := false
		// stand up an http server just for the duration of this test
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if updateHappened {
				fmt.Fprintln(w, string(indexYamlAfterUpdateBytes))
			} else {
				fmt.Fprintln(w, string(indexYamlBeforeUpdateBytes))
			}
		}))
		defer ts.Close()

		repoSpec := &sourcev1beta2.HelmRepositorySpec{
			URL:      "https://example.repo.com/charts",
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		}

		repoStatus := &sourcev1beta2.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Digest:         "651f952130ea96823711d08345b85e82be011dc6",
				LastUpdateTime: metav1.Time{Time: lastUpdateTime},
				Revision:       "651f952130ea96823711d08345b85e82be011dc6",
			},
			Conditions: []metav1.Condition{
				{
					Type:   fluxmeta.ReadyCondition,
					Status: metav1.ConditionTrue,
					Reason: fluxmeta.SucceededReason,
				},
			},
			URL: ts.URL,
		}

		repoName := types.NamespacedName{Namespace: "ns2", Name: "testrepo"}
		repo := newRepo(repoName.Name, repoName.Namespace, repoSpec, repoStatus)

		s, mock, err := newSimpleServerWithRepos(t, []sourcev1beta2.HelmRepository{repo})
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		responseBeforeUpdate, err := s.GetAvailablePackageSummaries(
			ctx,
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
		if err != nil {
			t.Fatal(err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		compareAvailablePackageSummaries(t, responseBeforeUpdate.Msg, expected_summaries_before_update)

		// see below
		key, oldValue, err := s.redisKeyValueForRepo(repo)
		if err != nil {
			t.Fatal(err)
		}

		ctrlClient, _, err := ctrlClientAndWatcher(t, s)
		if err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Get(ctx, repoName, &repo); err != nil {
			t.Fatal(err)
		} else {
			updateHappened = true
			// now we are going to simulate flux seeing an update of the index.yaml and modifying the
			// HelmRepository CRD which, in turn, causes k8s server to fire a MODIFY event
			repo.Status.Artifact.Digest = "4e881a3c34a5430c1059d2c4f753cb9aed006803"
			repo.Status.Artifact.Revision = "4e881a3c34a5430c1059d2c4f753cb9aed006803"

			// there will be a GET to retrieve the old value from the cache followed by a SET
			// to new value
			_, newValue, err := s.redisMockSetValueForRepo(mock, repo, oldValue)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			s.repoCache.ExpectAdd(key)

			if err = ctrlClient.Update(ctx, &repo); err != nil {
				// unlike dynamic.Interface.Update, client.Update will update an object in k8s
				// and an Modified event will be fired
				t.Fatal(err)
			}
			s.repoCache.WaitUntilForgotten(key)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			mock.ExpectGet(key).SetVal(string(newValue))

			responsePackagesAfterUpdate, err := s.GetAvailablePackageSummaries(
				ctx,
				connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
			if err != nil {
				t.Fatal(err)
			}
			compareAvailablePackageSummaries(t, responsePackagesAfterUpdate.Msg, expected_summaries_after_update)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		}
	})
}

func TestGetAvailablePackageSummaryAfterFluxHelmRepoDelete(t *testing.T) {
	t.Run("test get available package summaries after flux helm repository CRD gets deleted", func(t *testing.T) {
		repoName := types.NamespacedName{Namespace: "default", Name: "bitnami-1"}
		replaceUrls := make(map[string]string)
		charts := []testSpecChartWithUrl{}
		for _, s := range valid_index_charts_spec {
			tarGzBytes, err := os.ReadFile(s.tgzFile)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, err = w.Write(tarGzBytes)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}))
			defer ts.Close()
			replaceUrls[fmt.Sprintf("{{%s}}", s.tgzFile)] = ts.URL
			c := testSpecChartWithUrl{
				chartID:       fmt.Sprintf("%s/%s", repoName.Name, s.name),
				chartRevision: s.revision,
				chartUrl:      ts.URL,
				repoNamespace: repoName.Namespace,
			}
			charts = append(charts, c)
		}
		ts, repo, err := newHttpRepoAndServeIndex(
			testYaml("valid-index.yaml"), repoName.Name, repoName.Namespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts.Close()

		s, mock, err := newServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo}, charts, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatal(err)
		}

		responseBeforeDelete, err := s.GetAvailablePackageSummaries(
			context.Background(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
		if err != nil {
			t.Fatal(err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		compareAvailablePackageSummaries(t, responseBeforeDelete.Msg, valid_index_available_package_summaries_resp)

		// now we are going to simulate the user deleting a HelmRepository CR which, in turn,
		// causes k8s server to fire a DELETE event
		chartsInCache := []string{
			"acs-engine-autoscaler:2.1.1",
			"wordpress:0.7.5",
		}

		repoKey, err := redisKeyForRepoNamespacedName(repoName)
		if err != nil {
			t.Fatal(err)
		}

		if err = redisMockExpectDeleteRepoWithCharts(mock, repoName, chartsInCache); err != nil {
			t.Fatal(err)
		}

		chartCacheKeys := []string{}
		for _, c := range chartsInCache {
			chartCacheKeys = append(chartCacheKeys, fmt.Sprintf("helmcharts:%s:%s/%s", repoName.Namespace, repoName.Name, c))
		}

		s.repoCache.ExpectAdd(repoKey)
		for _, k := range chartCacheKeys {
			s.chartCache.ExpectAdd(k)
		}

		ctx := context.Background()
		if ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Delete(ctx, repo); err != nil {
			t.Fatal(err)
		}

		s.repoCache.WaitUntilForgotten(repoKey)
		for _, k := range chartCacheKeys {
			s.chartCache.WaitUntilForgotten(k)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		responseAfterDelete, err := s.GetAvailablePackageSummaries(
			context.Background(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
		if err != nil {
			t.Fatal(err)
		}

		if len(responseAfterDelete.Msg.AvailablePackageSummaries) != 0 {
			t.Errorf("expected empty array, got: %s", responseAfterDelete)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

// test that causes RetryWatcher to stop and the cache needs to resync
func TestGetAvailablePackageSummaryAfterCacheResync(t *testing.T) {
	t.Run("test that causes RetryWatcher to stop and the cache needs to resync", func(t *testing.T) {
		ts2, repo, err := newHttpRepoAndServeIndex(testYaml("valid-index.yaml"), "bitnami-1", "default", nil, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		s, mock, err := newSimpleServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo})
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatal(err)
		}

		responseBeforeResync, err := s.GetAvailablePackageSummaries(
			context.Background(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
		if err != nil {
			t.Fatal(err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		compareAvailablePackageSummaries(t, responseBeforeResync.Msg, valid_index_available_package_summaries_resp)

		resyncCh, err := s.repoCache.ExpectResync()
		if err != nil {
			t.Fatal(err)
		}

		// now lets try to simulate HTTP 410 GONE exception which should force
		// RetryWatcher to stop and force a cache resync. The ERROR event which
		// we'll send below should trigger a re-sync of the cache in the
		// background: a FLUSHDB followed by a SET
		_, watcher, err := ctrlClientAndWatcher(t, s)
		if err != nil {
			t.Fatal(err)
		}

		watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)

		// wait for the server to start the resync process. Don't care how big the work queue is
		<-resyncCh

		// set up expectations
		mock.ExpectFlushDB().SetVal("OK")
		if _, _, err := s.redisMockSetValueForRepo(mock, *repo, nil); err != nil {
			t.Fatalf("%+v", err)
		}

		// tell server its okay to proceed
		resyncCh <- 0
		s.repoCache.WaitUntilResyncComplete()

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatal(err)
		}

		responseAfterResync, err := s.GetAvailablePackageSummaries(
			context.Background(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}}))
		if err != nil {
			t.Fatal(err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		compareAvailablePackageSummaries(t, responseAfterResync.Msg, valid_index_available_package_summaries_resp)
	})
}

// test that causes RetryWatcher to stop and the cache needs to resync when there are
// lots of pending work items. this test is focused on the repo cache work queue
func TestGetAvailablePackageSummariesAfterCacheResyncQueueNotIdle(t *testing.T) {
	t.Run("test that causes RetryWatcher to stop and the repo cache needs to resync", func(t *testing.T) {
		// start with an empty server that only has an empty repo cache
		s, mock, err := newSimpleServerWithRepos(t, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		// first, I'd like to fill up the work queue with a whole bunch of work items
		repos := []*sourcev1beta2.HelmRepository{}
		mapReposCached := make(map[string][]byte)
		keysInOrder := []string{}

		const MAX_REPOS = 20
		for i := 0; i < MAX_REPOS; i++ {
			repoName := fmt.Sprintf("bitnami-%d", i)

			ts, repo, err := newHttpRepoAndServeIndex(testYaml("valid-index.yaml"), repoName, "default", nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts.Close()

			// the GETs and SETs will be done by background worker upon processing of Add event in
			// NamespaceResourceWatcherCache.onAddOrModify()
			key, byteArray, err := s.redisKeyValueForRepo(*repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			mapReposCached[key] = byteArray
			keysInOrder = append(keysInOrder, key)
			redisMockSetValueForRepo(mock, key, byteArray, nil)
			repos = append(repos, repo)
		}

		s.repoCache.ExpectAdd(keysInOrder[0])

		ctrlClient, watcher, err := ctrlClientAndWatcher(t, s)
		if err != nil {
			t.Fatal(err)
		} else {
			ctx := context.Background()
			for _, r := range repos {
				if err = ctrlClient.Create(ctx, r); err != nil {
					t.Fatal(err)
				}
			}
		}

		done := make(chan int, 1)

		go func() {
			// wait until the first of the added repos have been fully processed
			s.repoCache.WaitUntilForgotten(keysInOrder[0])

			// pretty delicate dance between the server and the client below using
			// bi-directional channels in order to make sure the right expectations
			// are set at the right time.
			resyncCh, err := s.repoCache.ExpectResync()
			if err != nil {
				t.Errorf("%v", err)
			}

			// now we will simulate a HTTP 410 Gone error in the watcher
			watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)
			// we need to wait until server can guarantee no more Redis SETs after
			// this until resync() kicks in
			len := <-resyncCh
			if len == 0 {
				t.Errorf("ERROR: Expected non-empty repo work queue!")
			} else {
				mock.ExpectFlushDB().SetVal("OK")
				// *SOME* of the repos have already been cached into redis at this point
				// via the repo cache background worker triggered by the Add event in the
				// main goroutine. Those SET calls will need to be repeated due to
				// populateWith() which will re-populate the cache from scratch based on
				// the current state in k8s (all MAX_REPOS repos).
				for i := 0; i <= (MAX_REPOS - len); i++ {
					redisMockSetValueForRepo(mock, keysInOrder[i], mapReposCached[keysInOrder[i]], nil)
				}
				// now we can signal to the server it's ok to proceed
				resyncCh <- 0
				s.repoCache.WaitUntilResyncComplete()
				// we do ClearExpect() here to avoid things like
				// "there is a remaining expectation which was not matched:
				// [get helmrepositories:default:bitnami-4]"
				// which might happened because the for loop in the main goroutine may have done a GET
				// right before resync() kicked in. We don't care about that
				mock.ClearExpect()
			}
			done <- 0
		}()

		<-done

		// in case the side go-routine had failures
		if t.Failed() {
			t.FailNow()
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		// at this point I'd like to make sure that GetAvailablePackageSummaries returns
		// packages from all repos
		for key, byteArray := range mapReposCached {
			mock.ExpectGet(key).SetVal(string(byteArray))
		}

		resp, err := s.GetAvailablePackageSummaries(context.TODO(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{}))
		if err != nil {
			t.Fatal(err)
		}

		// we need to make sure that response contains packages from all existing repositories
		// regardless whether they're in the cache or not
		expected := sets.Set[string]{}
		for i := 0; i < len(repos); i++ {
			repo := fmt.Sprintf("bitnami-%d", i)
			expected.Insert(repo)
		}
		for _, s := range resp.Msg.AvailablePackageSummaries {
			id := strings.Split(s.AvailablePackageRef.Identifier, "/")
			expected.Delete(id[0])
		}

		if expected.Len() != 0 {
			t.Fatalf("Expected to get packages from these repositories: %s, but did not get any",
				expected.UnsortedList())
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

// test that causes RetryWatcher to stop and the cache needs to resync when there are
// lots of pending work items
// this test is focused on the repo cache work queue
func TestGetAvailablePackageSummariesAfterCacheResyncQueueIdle(t *testing.T) {
	t.Run("test that causes RetryWatcher to stop and the repo cache needs to resync (idle queue)", func(t *testing.T) {
		// start with an empty server that only has an empty repo cache
		s, mock, err := newSimpleServerWithRepos(t, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		// first, I'd like to make sure there is a single item in the queue
		repoName := "bitnami-0"
		repoNamespace := "default"

		ts, repo, err := newHttpRepoAndServeIndex(testYaml("valid-index.yaml"), repoName, repoNamespace, nil, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts.Close()

		// the GETs and SETs will be done by background worker upon processing of Add event in
		// NamespaceResourceWatcherCache.onAddOrModify()
		key, byteArray, err := s.redisKeyValueForRepo(*repo)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		redisMockSetValueForRepo(mock, key, byteArray, nil)

		s.repoCache.ExpectAdd(key)

		ctrlClient, watcher, err := ctrlClientAndWatcher(t, s)
		if err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Create(context.Background(), repo); err != nil {
			t.Fatal(err)
		}

		done := make(chan int, 1)

		go func() {
			// wait until the first of the added repos have been fully processed
			s.repoCache.WaitUntilForgotten(key)

			// pretty delicate dance between the server and the client below using
			// bi-directional channels in order to make sure the right expectations
			// are set at the right time.
			resyncCh, err := s.repoCache.ExpectResync()
			if err != nil {
				t.Errorf("%v", err)
			}

			// now we will simulate a HTTP 410 Gone error in the watcher
			watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)
			// we need to wait until server can guarantee no more Redis SETs after
			// this until resync() kicks in
			len := <-resyncCh
			if len != 0 {
				t.Errorf("ERROR: Expected empty repo work queue!")
			} else {
				mock.ExpectFlushDB().SetVal("OK")
				redisMockSetValueForRepo(mock, key, byteArray, nil)
				// now we can signal to the server it's ok to proceed
				resyncCh <- 0
				s.repoCache.WaitUntilResyncComplete()
			}
			done <- 0
		}()

		<-done

		// in case the side go-routine had failures
		if t.Failed() {
			t.FailNow()
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}

		// at this point I'd like to make sure that GetAvailablePackageSummaries returns
		// packages from all repos
		mock.ExpectGet(key).SetVal(string(byteArray))

		resp, err := s.GetAvailablePackageSummaries(context.TODO(),
			connect.NewRequest(&corev1.GetAvailablePackageSummariesRequest{}))
		if err != nil {
			t.Fatal(err)
		}

		// we need to make sure that response contains packages from all existing repositories
		// regardless whether they're in the cache or not
		expected := sets.Set[string]{}
		expected.Insert(repoName)
		for _, s := range resp.Msg.AvailablePackageSummaries {
			id := strings.Split(s.AvailablePackageRef.Identifier, "/")
			expected.Delete(id[0])
		}

		if expected.Len() != 0 {
			t.Fatalf("Expected to get packages from these repositories: %s, but did not get any",
				expected.UnsortedList())
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestAddPackageRepository(t *testing.T) {
	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	ca, pub, priv := getCertsForTesting(t)

	testCases := []struct {
		name                  string
		request               *corev1.AddPackageRepositoryRequest
		expectedResponse      *corev1.AddPackageRepositoryResponse
		expectedRepo          *sourcev1beta2.HelmRepository
		errorCode             connect.Code
		existingSecret        *apiv1.Secret
		expectedCreatedSecret *apiv1.Secret
		userManagedSecrets    bool
	}{
		{
			name:      "returns error if no namespace is provided",
			request:   &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{}},
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if no name is provided",
			request:   &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Namespace: "foo"}},
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if namespaced scoped",
			request:   add_repo_req_1,
			errorCode: connect.CodeUnimplemented,
		},
		{
			name:      "returns error if wrong repository type",
			request:   add_repo_req_2,
			errorCode: connect.CodeUnimplemented,
		},
		{
			name:      "returns error if no url",
			request:   add_repo_req_3,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if insecureskipverify is set",
			request:   add_repo_req_4,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:             "simple add package repository scenario",
			request:          add_repo_req_5,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_1,
		},
		{
			name:             "package repository with tls cert authority",
			request:          add_repo_req_6(ca),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_2,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"bar",
				newTlsSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, nil, nil, ca))),
		},
		{
			name:      "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:   add_repo_req_7,
			errorCode: connect.CodeNotFound,
		},
		{
			name:             "package repository with secret key reference",
			request:          add_repo_req_7,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_3,
			existingSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "foo"}, nil, nil, ca),
			userManagedSecrets: true,
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            add_repo_req_7,
			errorCode:          connect.CodeNotFound,
			userManagedSecrets: true,
		},
		{
			name:      "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:   add_repo_req_7,
			errorCode: connect.CodeNotFound,
		},
		{
			name:             "package repository with basic auth and pass_credentials flag",
			request:          add_repo_req_8,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_4,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"bar",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, "baz", "zot"))),
		},
		{
			name:             "package repository with TLS authentication",
			request:          add_repo_req_9(pub, priv),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_2,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef("bar",
				newTlsSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, pub, priv, nil))),
		},
		{
			name:      "errors for package repository with bearer token",
			request:   add_repo_req_10,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "errors for package repository with custom auth token",
			request:   add_repo_req_11,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "package repository with docker config JSON authentication",
			request:   add_repo_req_12,
			errorCode: connect.CodeInternal,
		},
		{
			name:             "package repository with basic auth and existing secret",
			request:          add_repo_req_13,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_3,
			existingSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "foo"}, "baz", "zot"),
			userManagedSecrets: true,
		},
		{
			name:      "package repository with basic auth and existing secret (kubeapps managed secrets)",
			request:   add_repo_req_13,
			errorCode: connect.CodeNotFound,
		},
		{
			name:      "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:   add_repo_req_14,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            add_repo_req_14,
			errorCode:          connect.CodeInvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          add_repo_req_20,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_5,
		},
		{
			name:             "add basic OCI package repository",
			request:          add_repo_req_26,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_6,
		},
		{
			name:             "add OCI package repository with gcp provider",
			request:          add_repo_req_29(),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_7,
		},
		{
			name:      "returns error when mix referenced secrets and user provided secret data",
			request:   add_repo_req_30,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:             "simple repo with description",
			request:          add_repo_req_31,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_8,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secrets := []runtime.Object{}
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}
			s, mock, err := newServerWithRepos(t, nil, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			nsname := types.NamespacedName{Namespace: tc.request.Context.Namespace, Name: tc.request.Name}
			if tc.errorCode == 0 {
				key, err := redisKeyForRepoNamespacedName(nsname)
				if err != nil {
					t.Fatal(err)
				}
				mock.ExpectGet(key).RedisNil()
			}

			ctx := context.Background()
			response, err := s.AddPackageRepository(ctx, connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.errorCode == 0 {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.AddPackageRepositoryResponse{},
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
					)
					if got, want := response.Msg, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// purposefully not calling mock.ExpectationsWereMet() here because
			// AddPackageRepository will trigger an ADD event that will be processed
			// asynchronously, so it may or may not have enough time to get to the
			// point where the cache worker does a GET

			// We don't need to check anything else for non-OK codes.
			if tc.errorCode != 0 {
				return
			}

			// check expected HelmReleass CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				var actualRepo sourcev1beta2.HelmRepository
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				} else {
					if tc.userManagedSecrets {
						if tc.expectedCreatedSecret != nil {
							t.Fatalf("Error: unexpected state")
						}

						// Manually setting TypeMeta, as the fakeclient doesn't do it anymore:
						// https://github.com/kubernetes-sigs/controller-runtime/pull/2633
						actualRepo.TypeMeta = tc.expectedRepo.TypeMeta

						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}
					} else {
						// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
						opt1 := cmpopts.IgnoreFields(sourcev1beta2.HelmRepositorySpec{}, "SecretRef")

						// Manually setting TypeMeta, as the fakeclient doesn't do it anymore:
						// https://github.com/kubernetes-sigs/controller-runtime/pull/2633
						actualRepo.TypeMeta = tc.expectedRepo.TypeMeta

						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got, opt1) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
						}

						if tc.expectedCreatedSecret != nil {
							// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
							if !strings.HasPrefix(actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name) {
								t.Errorf("SecretRef [%s] was expected to start with [%s]",
									actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name)
							}
							opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
							// check expected secret has been created
							if typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster); err != nil {
								t.Fatal(err)
								// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
							} else if secret, err := typedClient.CoreV1().Secrets(nsname.Namespace).Get(ctx, actualRepo.Spec.SecretRef.Name, metav1.GetOptions{}); err != nil {
								t.Fatal(err)
							} else if got, want := secret, tc.expectedCreatedSecret; !cmp.Equal(want, got, opt2) {
								t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
							} else if !strings.HasPrefix(secret.Name, tc.expectedCreatedSecret.Name) {
								t.Errorf("Secret Name [%s] was expected to start with [%s]",
									secret.Name, tc.expectedCreatedSecret.Name)
							}
						} else if actualRepo.Spec.SecretRef != nil {
							t.Fatalf("Expected no secret, but found: [%q]", actualRepo.Spec.SecretRef.Name)
							// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
						} else if tc.expectedRepo.Spec.SecretRef != nil {
							t.Fatalf("Error: unexpected state")
						}
					}
				}
			}
		})
	}
}

func TestGetPackageRepositoryDetail(t *testing.T) {
	ca, pub, priv := getCertsForTesting(t)
	testCases := []struct {
		name               string
		request            *corev1.GetPackageRepositoryDetailRequest
		repoIndex          string
		repoName           string
		repoNamespace      string
		repoSecret         *apiv1.Secret
		pending            bool
		failed             bool
		expectedErrorCode  connect.Code
		expectedResponse   *corev1.GetPackageRepositoryDetailResponse
		userManagedSecrets bool
	}{
		{
			name:             "get package repository detail simplest case",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_1,
		},
		{
			name:              "fails with NotFound when wrong identifier",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           get_repo_detail_req_2,
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name:              "fails with NotFound when wrong namespace",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           get_repo_detail_req_3,
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name:              "it returns an invalid arg error status if no context is provided",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           get_repo_detail_req_4,
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:              "it returns an error status if cluster is not the global/kubeapps one",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           get_repo_detail_req_5,
			expectedErrorCode: connect.CodeUnimplemented,
		},
		{
			name:          "it returns package repository detail with TLS cert aurthority",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, nil, nil, ca),
			request:            get_repo_detail_req_1,
			expectedResponse:   get_repo_detail_resp_6,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with TLS cert authority (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo1",
				newTlsSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, nil, nil, ca))),
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_6a,
		},
		{
			name:             "get package repository with pending status",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_7,
			pending:          true,
		},
		{
			name:             "get package repository with failed status",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_8,
			failed:           true,
		},
		{
			name:          "it returns package repository detail with TLS cert authentication",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, pub, priv, nil),
			request:            get_repo_detail_req_1,
			expectedResponse:   get_repo_detail_resp_9,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with TLS cert authentication (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newTlsSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, pub, priv, nil))),
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_9a,
		},
		{
			name:          "it returns package repository detail with basic authentication",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            get_repo_detail_req_1,
			expectedResponse:   get_repo_detail_resp_10,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with basic authentication (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, "foo", "bar"))),
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_10a,
		},
		{
			name:             "get package repository detail description",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secretRef := ""
			secrets := []runtime.Object{}
			if tc.repoSecret != nil {
				secretRef = tc.repoSecret.Name
				secrets = append(secrets, tc.repoSecret)
			}
			var repo *sourcev1beta2.HelmRepository
			if !tc.pending && !tc.failed {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, secretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else if tc.pending {
				repoSpec := &sourcev1beta2.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							Type:               fluxmeta.ReadyCondition,
							Status:             metav1.ConditionUnknown,
							Reason:             fluxmeta.ProgressingReason,
							Message:            "reconciliation in progress",
						},
					},
				}
				repo1 := newRepo(tc.repoName, tc.repoNamespace, repoSpec, repoStatus)
				repo = &repo1
			} else { // failed
				repoSpec := &sourcev1beta2.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							Type:               fluxmeta.ReadyCondition,
							Status:             metav1.ConditionFalse,
							Reason:             fluxmeta.FailedReason,
							Message:            "failed to fetch https://invalid.example.com/index.yaml : 404 Not Found",
						},
					},
				}
				repo1 := newRepo(tc.repoName, tc.repoNamespace, repoSpec, repoStatus)
				repo = &repo1
			}

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, _, err := newServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			ctx := context.Background()
			actualResp, err := s.GetPackageRepositoryDetail(ctx, connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedErrorCode == 0 {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					comparePackageRepositoryDetail(t, actualResp.Msg, tc.expectedResponse)
				}
			}
		})
	}
}

func TestGetOciPackageRepositoryDetail(t *testing.T) {
	seed_data_1, err := newFakeRemoteOciRegistryData_1()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name              string
		request           *corev1.GetPackageRepositoryDetailRequest
		repoName          string
		repoNamespace     string
		repoUrl           string
		expectedErrorCode connect.Code
		expectedResponse  *corev1.GetPackageRepositoryDetailResponse
		seedData          *fakeRemoteOciRegistryData
	}{
		{
			name:             "get package repository detail for OCI repository",
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			repoUrl:          "oci://localhost:54321/userX/charts",
			request:          get_repo_detail_req_1,
			expectedResponse: get_repo_detail_resp_19,
			seedData:         seed_data_1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)

			repo, err := newOciRepo(tc.repoName, tc.repoNamespace, tc.repoUrl)
			if err != nil {
				t.Fatal(err)
			}

			s, mock, err := newServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo}, nil, nil)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			ctx := context.Background()
			actualResp, err := s.GetPackageRepositoryDetail(ctx, connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedErrorCode == 0 {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					comparePackageRepositoryDetail(t, actualResp.Msg, tc.expectedResponse)
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGetPackageRepositorySummaries(t *testing.T) {
	// some prep
	indexYAMLBytes, err := os.ReadFile(testYaml("valid-index.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(indexYAMLBytes))
	}))
	defer ts.Close()
	get_summaries_repo_1.Status.URL = ts.URL
	get_summaries_repo_2.Status.URL = ts.URL

	testCases := []struct {
		name              string
		request           *corev1.GetPackageRepositorySummariesRequest
		existingRepos     []sourcev1beta2.HelmRepository
		expectedErrorCode connect.Code
		expectedResponse  *corev1.GetPackageRepositorySummariesResponse
	}{
		{
			name: "returns package summaries when namespace not specified",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{},
			},
			existingRepos: []sourcev1beta2.HelmRepository{
				get_summaries_repo_1,
				get_summaries_repo_2,
				get_summaries_repo_3,
				get_summaries_repo_4,
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					get_summaries_summary_1,
					get_summaries_summary_2,
					get_summaries_summary_3,
					get_summaries_summary_4,
				},
			},
		},
		{
			name: "returns package summaries when namespace is specified",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Namespace: "foo"},
			},
			existingRepos: []sourcev1beta2.HelmRepository{
				get_summaries_repo_1,
				get_summaries_repo_2,
				get_summaries_repo_3,
				get_summaries_repo_4,
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					get_summaries_summary_1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, err := newSimpleServerWithRepos(t, tc.existingRepos)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetPackageRepositorySummaries(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedErrorCode != 0 {
				return
			}

			comparePackageRepositorySummaries(t, response.Msg, tc.expectedResponse)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdatePackageRepository(t *testing.T) {
	ca, pub, priv := getCertsForTesting(t)
	testCases := []struct {
		name                  string
		request               *corev1.UpdatePackageRepositoryRequest
		repoIndex             string
		repoName              string
		repoNamespace         string
		oldRepoSecret         *apiv1.Secret
		newRepoSecret         *apiv1.Secret
		expectedCreatedSecret *apiv1.Secret
		pending               bool
		expectedErrorCode     connect.Code
		expectedResponse      *corev1.UpdatePackageRepositoryResponse
		expectedDetail        *corev1.GetPackageRepositoryDetailResponse
		userManagedSecrets    bool
	}{
		{
			name:             "update repository url",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_1,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_1,
		},
		{
			name:             "update repository poll interval",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_2,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_2,
		},
		{
			name:             "update repository pass credentials flag",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_3,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_3,
		},
		{
			name:          "update repository set TLS cert authority",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			newRepoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, nil, nil, ca),
			request:            update_repo_req_4,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_4,
			userManagedSecrets: true,
		},
		{
			name:          "update repository unset TLS cert authority",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, nil, nil, ca),
			request:            update_repo_req_5,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_5,
			userManagedSecrets: true,
		},
		{
			name:          "update repository set basic auth",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			newRepoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            update_repo_req_6,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_6,
			userManagedSecrets: true,
		},
		{
			name:          "update repository unset basic auth",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            update_repo_req_7,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_7,
			userManagedSecrets: true,
		},
		{
			name:             "update repository set TLS cert/key (kubeapps-managed secrets)",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_8(pub, priv),
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_8,
		},
		{
			name:          "update repository unset TLS cert/key (kubeapps-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newTlsSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, pub, priv, nil))),
			request:          update_repo_req_9,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_9,
		},
		{
			name:          "update repository change from TLS cert/key to basic auth (kubeapps-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newTlsSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, pub, priv, nil))),
			request:          update_repo_req_10,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_10,
		},
		{
			name:              "updates to pending repo is not allowed",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           update_repo_req_1,
			expectedErrorCode: connect.CodeInternal,
			pending:           true,
		},
		{
			name:          "updates url for repo preserve secret in kubeapps managed env",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "foo", "bar"))),
			request:          update_repo_req_16,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_15,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "foo", "bar"))),
		},
		{
			name:              "returns error when mix referenced secrets and user provided secret data",
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			request:           update_repo_req_19,
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:          "update repository change Auth management mode (user-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            update_repo_req_20,
			expectedErrorCode:  connect.CodeInvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:          "update repository change Auth management mode (kubeapps-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "secret-1",
					Namespace: "namespace-1"}, "foo", "bar"))),
			request:           update_repo_req_21,
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name:          "issue5747 - update auth password: username was incorrectly overridden to redacted string",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "foo", "bar"))),
			request:          update_repo_req_22,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_18,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "foo", "doe"))),
		},
		{
			name:          "issue5747 - update basic auth but not tls ca: basic auth updates are ignored",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthTlsSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "foo", "bar", nil, nil, ca))),
			request:          update_repo_req_23,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_19,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newBasicAuthTlsSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, "john", "doe", nil, nil, ca))),
		},
		{
			name:             "issue5747 - starts with no auth/tls, adding tls is being ignored",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_24(ca),
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_20,
			expectedCreatedSecret: setSecretManagedByKubeapps(setSecretOwnerRef(
				"repo-1",
				newTlsSecret(types.NamespacedName{
					Name:      "repo-1",
					Namespace: "namespace-1"}, nil, nil, ca))),
		},
		{
			name:             "update description",
			repoIndex:        testYaml("valid-index.yaml"),
			repoName:         "repo-1",
			repoNamespace:    "namespace-1",
			request:          update_repo_req_25,
			expectedResponse: update_repo_resp_1,
			expectedDetail:   update_repo_detail_21,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldSecretRef := ""
			secrets := []runtime.Object{}
			if tc.oldRepoSecret != nil {
				oldSecretRef = tc.oldRepoSecret.Name
				secrets = append(secrets, tc.oldRepoSecret)
			}
			if tc.newRepoSecret != nil {
				secrets = append(secrets, tc.newRepoSecret)
			}
			var repo *sourcev1beta2.HelmRepository
			if !tc.pending {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, oldSecretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else {
				repoSpec := &sourcev1beta2.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							Type:               fluxmeta.ReadyCondition,
							Status:             metav1.ConditionUnknown,
							Reason:             fluxmeta.ProgressingReason,
							Message:            "reconciliation in progress",
						},
					},
				}
				repo1 := newRepo(tc.repoName, tc.repoNamespace, repoSpec, repoStatus)
				repo = &repo1
			}
			// update to the repo in a failed state will be tested in integration test

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, _, err := newServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			ctx := context.Background()
			actualResp, err := s.UpdatePackageRepository(ctx, connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedErrorCode == 0 {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
						corev1.UpdatePackageRepositoryResponse{},
					)
					if got, want := actualResp.Msg, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			} else {
				// We don't need to check anything else for non-OK codes.
				return
			}

			actualDetail, err := s.GetPackageRepositoryDetail(ctx, connect.NewRequest(&corev1.GetPackageRepositoryDetailRequest{
				PackageRepoRef: actualResp.Msg.PackageRepoRef,
			}))
			if err != nil {
				t.Fatalf("got: %+v, want: %+v, err: %+v", connect.CodeOf(err), 0, err)
			}

			if actualDetail == nil {
				t.Fatalf("got: nil, want: detail")
			} else {
				comparePackageRepositoryDetail(t, actualDetail.Msg, tc.expectedDetail)
			}

			// ensures the secret has been created/updated correctly
			if !tc.userManagedSecrets && (tc.oldRepoSecret != nil || tc.expectedCreatedSecret != nil) {
				typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster)
				if err != nil {
					t.Fatal(err)
				}
				ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster)
				if err != nil {
					t.Fatal(err)
				}

				// check the secret has been deleted
				if tc.oldRepoSecret != nil && tc.expectedCreatedSecret == nil {
					if _, err = typedClient.CoreV1().Secrets(tc.repoNamespace).Get(ctx, tc.oldRepoSecret.Name, metav1.GetOptions{}); err == nil {
						t.Fatalf("Expected secret [%q] to have been deleted", tc.oldRepoSecret.Name)
					}
				}

				// check the created/updated secret
				if tc.expectedCreatedSecret != nil {
					var actualRepo sourcev1beta2.HelmRepository
					if err = ctrlClient.Get(ctx, types.NamespacedName{Namespace: tc.repoNamespace, Name: tc.repoName}, &actualRepo); err != nil {
						t.Fatal(err)
					}
					// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
					if actualRepo.Spec.SecretRef == nil {
						t.Fatalf("Expected repo to have a secret ref, none found")
					}

					opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
					// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
					if secret, err := typedClient.CoreV1().Secrets(tc.repoNamespace).Get(ctx, actualRepo.Spec.SecretRef.Name, metav1.GetOptions{}); err != nil {
						t.Fatal(err)
					} else if got, want := secret, tc.expectedCreatedSecret; !cmp.Equal(want, got, opt2) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
					} else if !strings.HasPrefix(secret.Name, tc.expectedCreatedSecret.Name) {
						t.Errorf("Secret Name [%s] was expected to start with [%s]", secret.Name, tc.expectedCreatedSecret.Name)
					}

				}
			}
		})
	}
}

func TestDeletePackageRepository(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.DeletePackageRepositoryRequest
		repoIndex          string
		repoName           string
		repoNamespace      string
		oldRepoSecret      *apiv1.Secret
		newRepoSecret      *apiv1.Secret
		pending            bool
		expectedErrorCode  connect.Code
		userManagedSecrets bool
	}{
		{
			name:          "delete repository",
			request:       delete_repo_req_1,
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
		},
		{
			name:              "returns not found if package repo doesn't exist",
			request:           delete_repo_req_2,
			repoIndex:         testYaml("valid-index.yaml"),
			repoName:          "repo-1",
			repoNamespace:     "namespace-1",
			expectedErrorCode: connect.CodeNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldSecretRef := ""
			secrets := []runtime.Object{}
			if tc.oldRepoSecret != nil {
				oldSecretRef = tc.oldRepoSecret.Name
				secrets = append(secrets, tc.oldRepoSecret)
			}
			if tc.newRepoSecret != nil {
				secrets = append(secrets, tc.newRepoSecret)
			}
			var repo *sourcev1beta2.HelmRepository
			if !tc.pending {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, oldSecretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else {
				repoSpec := &sourcev1beta2.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1beta2.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							LastTransitionTime: metav1.Time{Time: lastTransitionTime},
							Type:               fluxmeta.ReadyCondition,
							Status:             metav1.ConditionUnknown,
							Reason:             fluxmeta.ProgressingReason,
							Message:            "reconciliation in progress",
						},
					},
				}
				repo1 := newRepo(tc.repoName, tc.repoNamespace, repoSpec, repoStatus)
				repo = &repo1
			}
			// update to the repo in a failed state will be tested in integration test

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, _, err := newServerWithRepos(t, []sourcev1beta2.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			ctx := context.Background()
			ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster)
			if err != nil {
				t.Fatal(err)
			}
			nsname := types.NamespacedName{
				Namespace: tc.request.PackageRepoRef.Context.Namespace,
				Name:      tc.request.PackageRepoRef.Identifier,
			}
			var actualRepo sourcev1beta2.HelmRepository
			if tc.expectedErrorCode == 0 {
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				}
			}

			_, err = s.DeletePackageRepository(ctx, connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedErrorCode == 0 {
				// check the repository CRD is gone from the cluster
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err == nil {
					t.Fatalf("Expected repository [%s] to have been deleted but still exists", nsname)
				}
			}
		})
	}
}

func TestGetOciAvailablePackageSummariesWithoutPagination(t *testing.T) {
	seed_data_1, err := newFakeRemoteOciRegistryData_1()
	if err != nil {
		t.Fatal(err)
	}

	seed_data_3, err := newFakeRemoteOciRegistryData_3()
	if err != nil {
		t.Fatal(err)
	}

	type testSpecGetOciAvailablePackageSummaries struct {
		repoName      string
		repoNamespace string
		repoUrl       string
	}

	testCases := []struct {
		name              string
		request           *corev1.GetAvailablePackageSummariesRequest
		repos             []testSpecGetOciAvailablePackageSummaries
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
		expectedErrorCode connect.Code
		seedData          *fakeRemoteOciRegistryData
	}{
		{
			name: "returns a single available package",
			repos: []testSpecGetOciAvailablePackageSummaries{
				{
					repoName:      "repo-1",
					repoNamespace: "namespace-1",
					repoUrl:       "oci://localhost:54321/userX/charts",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: oci_repo_available_package_summaries,
			},
			seedData: seed_data_1,
		},
		{
			name: "returns available packages from multiple repos",
			repos: []testSpecGetOciAvailablePackageSummaries{
				{
					repoName:      "repo-1",
					repoNamespace: "namespace-1",
					repoUrl:       "oci://localhost:54321/userX/charts",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: oci_repo_available_package_summaries_2,
			},
			seedData: seed_data_3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)

			repos := []sourcev1beta2.HelmRepository{}

			for _, rs := range tc.repos {
				repo, err := newOciRepo(rs.repoName, rs.repoNamespace, rs.repoUrl)
				if err != nil {
					t.Fatal(err)
				}
				repos = append(repos, *repo)
			}

			s, mock, err := newSimpleServerWithRepos(t, repos)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			if err = s.redisMockExpectGetFromRepoCache(mock, tc.request.FilterOptions, repos...); err != nil {
				t.Fatal(err)
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), connect.NewRequest(tc.request))
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %v, want: %v, err: %v", got, want, err)
			}
			// If an error code was expected, then no need to continue checking
			// the response.
			if tc.expectedErrorCode != 0 {
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
			compareAvailablePackageSummaries(t, response.Msg, tc.expectedResponse)
		})
	}
}

func newRepo(name string, namespace string, spec *sourcev1beta2.HelmRepositorySpec, status *sourcev1beta2.HelmRepositoryStatus) sourcev1beta2.HelmRepository {
	helmRepository := sourcev1beta2.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Generation: 1,
		},
	}
	if namespace != "" {
		helmRepository.ObjectMeta.Namespace = namespace
	}

	// Just FYI, if spec is not specified, the repo object will have a default one, e.g.
	// {
	//	"metadata": {
	//	 ...
	//	},
	//	"spec": {
	//	  "url": "",
	//	  "interval": "0s"
	//	},
	//	"status": {
	//  ...
	if spec != nil {
		helmRepository.Spec = *spec.DeepCopy()
	}

	if status != nil {
		helmRepository.Status = *status.DeepCopy()
		if status.ObservedGeneration == 0 {
			helmRepository.Status.ObservedGeneration = 1
		}
	}

	return helmRepository
}

// these functions should affect only unit test, not production code
// does a series of mock.ExpectGet(...)
func (s *Server) redisMockExpectGetFromRepoCache(mock redismock.ClientMock, filterOptions *corev1.FilterOptions, repos ...sourcev1beta2.HelmRepository) error {
	mapVals := make(map[string][]byte)
	ociRepoKeys := sets.Set[string]{}
	for _, r := range repos {
		key, bytes, err := s.redisKeyValueForRepo(r)
		if err != nil {
			return err
		}
		mapVals[key] = bytes
		if r.Spec.Type == "oci" {
			ociRepoKeys.Insert(key)
		}
	}
	if filterOptions == nil || len(filterOptions.GetRepositories()) == 0 {
		for k, v := range mapVals {
			maxTries := 1
			if ociRepoKeys.Has(k) {
				// see comment in repo.go repoCacheEntryFromUntyped() func about caching helm OCI chart repos
				maxTries = 3
			}
			for i := 0; i < maxTries; i++ {
				mock.ExpectGet(k).SetVal(string(v))
			}
		}
	} else {
		for _, r := range filterOptions.GetRepositories() {
			for k, v := range mapVals {
				if strings.HasSuffix(k, ":"+r) {
					maxTries := 1
					if ociRepoKeys.Has(k) {
						// see comment in chart.go about caching helm OCI chart repos
						maxTries = 3
					}
					for i := 0; i < maxTries; i++ {
						mock.ExpectGet(k).SetVal(string(v))
					}
				}
			}
		}
	}
	return nil
}

func (s *Server) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1beta2.HelmRepository, oldValue []byte) (key string, bytes []byte, err error) {
	bg := &clientgetter.FixedClusterClientProvider{ClientsFunc: func(ctx context.Context) (*clientgetter.ClientGetter, error) {
		return s.clientGetter.GetClients(http.Header{}, s.kubeappsCluster)
	}}
	sinkNoCache := repoEventSink{clientGetter: bg}
	return sinkNoCache.redisMockSetValueForRepo(mock, repo, oldValue)
}

func (sink *repoEventSink) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1beta2.HelmRepository, oldValue []byte) (key string, newValue []byte, err error) {
	if key, newValue, err = sink.redisKeyValueForRepo(repo); err != nil {
		if oldValue == nil {
			mock.ExpectGet(key).RedisNil()
		} else {
			mock.ExpectGet(key).SetVal(string(oldValue))
		}
		mock.ExpectDel(key).SetVal(0)
		return key, nil, err
	} else {
		redisMockSetValueForRepo(mock, key, newValue, oldValue)
		return key, newValue, nil
	}
}

func redisMockSetValueForRepo(mock redismock.ClientMock, key string, newValue, oldValue []byte) {
	if oldValue == nil {
		mock.ExpectGet(key).RedisNil()
	} else {
		mock.ExpectGet(key).SetVal(string(oldValue))
	}
	mock.ExpectSet(key, newValue, 0).SetVal("OK")
	mock.ExpectInfo("memory").SetVal("used_memory_rss_human:NA\r\nmaxmemory_human:NA")
}

func (s *Server) redisKeyValueForRepo(r sourcev1beta2.HelmRepository) (key string, byteArray []byte, err error) {
	cg := &clientgetter.FixedClusterClientProvider{ClientsFunc: func(ctx context.Context) (*clientgetter.ClientGetter, error) {
		return s.clientGetter.GetClients(http.Header{}, s.kubeappsCluster)
	}}
	sinkNoChartCache := repoEventSink{clientGetter: cg}
	return sinkNoChartCache.redisKeyValueForRepo(r)
}

func (sink *repoEventSink) redisKeyValueForRepo(r sourcev1beta2.HelmRepository) (key string, byteArray []byte, err error) {
	if key, err = redisKeyForRepo(r); err != nil {
		return key, nil, err
	} else {
		// we are not really adding anything to the cache here, rather just calling a
		// onAddRepo to compute the value that *WOULD* be stored in the cache
		var byteArray interface{}
		var add bool
		byteArray, add, err = sink.onAddRepo(key, &r)
		if err != nil {
			return key, nil, err
		} else if !add {
			return key, nil, fmt.Errorf("onAddRepo returned false for setVal")
		}
		return key, byteArray.([]byte), nil
	}
}

func redisKeyForRepo(r sourcev1beta2.HelmRepository) (string, error) {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return redisKeyForRepoNamespacedName(types.NamespacedName{
		Namespace: r.GetNamespace(),
		Name:      r.GetName()})
}

func redisKeyForRepoNamespacedName(name types.NamespacedName) (string, error) {
	if name.Name == "" || name.Namespace == "" {
		return "", fmt.Errorf("invalid key: [%s]", name)
	}
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return fmt.Sprintf("%s:%s:%s", fluxHelmRepositories, name.Namespace, name.Name), nil
}

func newHttpRepoAndServeIndex(repoIndex, repoName, repoNamespace string, replaceUrls map[string]string, secretRef string) (*httptest.Server, *sourcev1beta2.HelmRepository, error) {
	indexYAMLBytes, err := os.ReadFile(repoIndex)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range replaceUrls {
		indexYAMLBytes = []byte(strings.ReplaceAll(string(indexYAMLBytes), k, v))
	}

	// stand up a plain text http server to serve the contents of index.yaml just for the
	// duration of this test. We are never standing up a TLS server (or any kind of secured
	// server for that matter) for repo index.yaml file because this scenario should never
	// happen in production. See comments in repo.go for explanation
	// This is only true for repo index.yaml, not for the chart URLs within it.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(indexYAMLBytes))
	}))

	repoSpec := &sourcev1beta2.HelmRepositorySpec{
		URL:      "https://example.repo.com/charts",
		Interval: metav1.Duration{Duration: 1 * time.Minute},
	}

	if secretRef != "" {
		// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
		repoSpec.SecretRef = &fluxmeta.LocalObjectReference{Name: secretRef}
	}

	revision := "651f952130ea96823711d08345b85e82be011dc6"
	sz := int64(31989)

	repoStatus := &sourcev1beta2.HelmRepositoryStatus{
		Artifact: &sourcev1.Artifact{
			Path:           fmt.Sprintf("helmrepository/%s/%s/index-%s.yaml", repoNamespace, repoName, revision),
			Digest:         revision,
			LastUpdateTime: metav1.Time{Time: lastUpdateTime},
			Revision:       revision,
			Size:           &sz,
			URL:            fmt.Sprintf("http://source-controller.flux-system.svc.cluster.local./helmrepository/%s/%s/index-%s.yaml", repoNamespace, repoName, revision),
		},
		Conditions: []metav1.Condition{
			{
				Type:               fluxmeta.ReadyCondition,
				Status:             metav1.ConditionTrue,
				Reason:             fluxmeta.SucceededReason,
				Message:            fmt.Sprintf("stored artifact for revision '%s'", revision),
				LastTransitionTime: metav1.Time{Time: lastTransitionTime},
				ObservedGeneration: 1,
			},
		},
		URL: ts.URL,
	}
	repo := newRepo(repoName, repoNamespace, repoSpec, repoStatus)
	return ts, &repo, nil
}

func newOciRepo(repoName, repoNamespace, repoUrl string) (*sourcev1beta2.HelmRepository, error) {
	timeout := metav1.Duration{Duration: 60 * time.Second}
	repoSpec := &sourcev1beta2.HelmRepositorySpec{
		URL:      repoUrl,
		Interval: metav1.Duration{Duration: 1 * time.Minute},
		Timeout:  &timeout,
		Type:     "oci",
	}

	repoStatus := &sourcev1beta2.HelmRepositoryStatus{
		Conditions: []metav1.Condition{
			{
				Type:               fluxmeta.ReadyCondition,
				Status:             metav1.ConditionTrue,
				Reason:             fluxmeta.SucceededReason,
				Message:            "Helm repository is ready",
				LastTransitionTime: metav1.Time{Time: lastTransitionTime},
				ObservedGeneration: 1,
			},
		},
	}
	repo := newRepo(repoName, repoNamespace, repoSpec, repoStatus)
	return &repo, nil
}

func TestGetPackageRepositoryPermissions(t *testing.T) {

	testCases := []struct {
		name              string
		request           *corev1.GetPackageRepositoryPermissionsRequest
		expectedErrorCode connect.Code
		expectedResponse  *corev1.GetPackageRepositoryPermissionsResponse
		reactors          []*ClientReaction
	}{
		{
			name: "returns permissions for global package repositories",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster},
			},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						createAction := action.(k8stesting.CreateActionImpl)
						accessReview := createAction.Object.(*authorizationv1.SelfSubjectAccessReview)
						if accessReview.Spec.ResourceAttributes.Namespace != "" {
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
						switch accessReview.Spec.ResourceAttributes.Verb {
						case "list", "delete":
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						default:
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: GetPluginDetail(),
						// Flux does not have the concept of "global"
						Global:    nil,
						Namespace: nil,
					},
				},
			},
		},
		{
			name:    "returns local permissions when no cluster specified",
			request: &corev1.GetPackageRepositoryPermissionsRequest{},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin:    GetPluginDetail(),
						Global:    nil,
						Namespace: nil,
					},
				},
			},
		},
		{
			name: "fails when namespace is specified but not the cluster",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Namespace: "my-ns"},
			},
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name: "returns permissions for namespaced package repositories",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster, Namespace: "my-ns"},
			},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
						createAction := action.(k8stesting.CreateActionImpl)
						accessReview := createAction.Object.(*authorizationv1.SelfSubjectAccessReview)
						if accessReview.Spec.ResourceAttributes.Namespace == "" {
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						}
						switch accessReview.Spec.ResourceAttributes.Verb {
						case "list", "delete":
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						default:
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: GetPluginDetail(),
						Namespace: map[string]bool{
							"create": false,
							"delete": true,
							"get":    false,
							"list":   true,
							"update": false,
							"watch":  false,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newServerWithReactors(t, tc.reactors)

			response, err := s.GetPackageRepositoryPermissions(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedErrorCode != 0 {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositoryPermissionsResponse{},
				corev1.PackageRepositoriesPermissions{},
			)
			if got, want := response.Msg, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}
