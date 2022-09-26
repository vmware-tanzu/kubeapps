// Copyright 2021-2022 the Kubeapps contributors.
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

	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	redismock "github.com/go-redis/redismock/v8"
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
		name              string
		request           *corev1.GetAvailablePackageSummariesRequest
		repos             []testSpecGetAvailablePackageSummaries
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
		expectedErrorCode codes.Code
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
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_available_package_summaries,
			},
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
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "default"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_available_package_summaries,
			},
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
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_available_package_summaries,
			},
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
			expectedErrorCode: codes.Unimplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repos := []sourcev1.HelmRepository{}

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

			if err = s.redisMockExpectGetFromRepoCache(mock, tc.request.FilterOptions, repos...); err != nil {
				t.Fatalf("%v", err)
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), tc.request)
			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %v, want: %v", got, want)
			}
			// If an error code was expected, then no need to continue checking
			// the response.
			if tc.expectedErrorCode != codes.OK {
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageSummariesResponse{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
			opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
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
		repos := []sourcev1.HelmRepository{}
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
			t.Fatalf("%v", err)
		}

		request1 := &corev1.GetAvailablePackageSummariesRequest{
			Context: &corev1.Context{Namespace: "blah"},
			PaginationOptions: &corev1.PaginationOptions{
				PageToken: "0",
				PageSize:  1,
			},
		}

		response1, err := s.GetAvailablePackageSummaries(context.Background(), request1)
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
		if got, want := response1, expectedResp1; cmp.Equal(want, got, opts, opts2) {
			match = true
			nextExpectedResp = expectedResp2
			nextExpectedResp.NextPageToken = "2"
		} else if got, want := response1, expectedResp2; cmp.Equal(want, got, opts, opts2) {
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
			t.Fatalf("%v", err)
		}
		response2, err := s.GetAvailablePackageSummaries(context.Background(), request2)
		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %v, want: %v", err, want)
		}
		if got, want := response2, nextExpectedResp; !cmp.Equal(want, got, opts, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
		}

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
			t.Fatalf("%v", err)
		}
		response3, err := s.GetAvailablePackageSummaries(context.Background(), request3)
		if got, want := status.Code(err), codes.OK; got != want {
			t.Fatalf("got: %v, want: %v", err, want)
		}
		if got, want := response3, nextExpectedResp; !cmp.Equal(want, got, opts, opts2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
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

		repoSpec := &sourcev1.HelmRepositorySpec{
			URL:      "https://example.repo.com/charts",
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		}

		repoStatus := &sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Checksum:       "651f952130ea96823711d08345b85e82be011dc6",
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

		s, mock, err := newSimpleServerWithRepos(t, []sourcev1.HelmRepository{repo})
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
			t.Fatalf("%v", err)
		}

		ctx := context.Background()
		responseBeforeUpdate, err := s.GetAvailablePackageSummaries(
			ctx,
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		opt1 := cmpopts.IgnoreUnexported(
			corev1.AvailablePackageDetail{},
			corev1.AvailablePackageSummary{},
			corev1.AvailablePackageReference{},
			corev1.Context{},
			plugins.Plugin{},
			corev1.Maintainer{},
			corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeUpdate.AvailablePackageSummaries, index_before_update_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// see below
		key, oldValue, err := s.redisKeyValueForRepo(repo)
		if err != nil {
			t.Fatalf("%v", err)
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
			repo.Status.Artifact.Checksum = "4e881a3c34a5430c1059d2c4f753cb9aed006803"
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
				t.Fatalf("%v", err)
			}

			mock.ExpectGet(key).SetVal(string(newValue))

			responsePackagesAfterUpdate, err := s.GetAvailablePackageSummaries(
				ctx,
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
			if err != nil {
				t.Fatalf("%v", err)
			}

			if got, want := responsePackagesAfterUpdate.AvailablePackageSummaries, index_after_update_summaries; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
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

		s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, charts, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatalf("%v", err)
		}

		responseBeforeDelete, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		opt1 := cmpopts.IgnoreUnexported(
			corev1.AvailablePackageDetail{},
			corev1.AvailablePackageSummary{},
			corev1.AvailablePackageReference{},
			corev1.Context{},
			plugins.Plugin{},
			corev1.Maintainer{},
			corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeDelete.AvailablePackageSummaries, valid_index_available_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// now we are going to simulate the user deleting a HelmRepository CR which, in turn,
		// causes k8s server to fire a DELETE event
		chartsInCache := []string{
			"acs-engine-autoscaler:2.1.1",
			"wordpress:0.7.5",
		}

		repoKey, err := redisKeyForRepoNamespacedName(repoName)
		if err != nil {
			t.Fatalf("%v", err)
		}

		if err = redisMockExpectDeleteRepoWithCharts(mock, repoName, chartsInCache); err != nil {
			t.Fatalf("%v", err)
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
		if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Delete(ctx, repo); err != nil {
			t.Fatal(err)
		}

		s.repoCache.WaitUntilForgotten(repoKey)
		for _, k := range chartCacheKeys {
			s.chartCache.WaitUntilForgotten(k)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		responseAfterDelete, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if len(responseAfterDelete.AvailablePackageSummaries) != 0 {
			t.Errorf("expected empty array, got: %s", responseAfterDelete)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
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

		s, mock, err := newSimpleServerWithRepos(t, []sourcev1.HelmRepository{*repo})
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatalf("%v", err)
		}

		responseBeforeResync, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		opt1 := cmpopts.IgnoreUnexported(
			corev1.AvailablePackageDetail{},
			corev1.AvailablePackageSummary{},
			corev1.AvailablePackageReference{},
			corev1.Context{}, plugins.Plugin{},
			corev1.Maintainer{},
			corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeResync.AvailablePackageSummaries, valid_index_available_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		resyncCh, err := s.repoCache.ExpectResync()
		if err != nil {
			t.Fatalf("%v", err)
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
			t.Fatalf("%v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
			t.Fatalf("%v", err)
		}

		responseAfterResync, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		if got, want := responseAfterResync.AvailablePackageSummaries, valid_index_available_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}
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
		repos := []*sourcev1.HelmRepository{}
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
				// via the repo cache backround worker triggered by the Add event in the
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
			t.Fatalf("%v", err)
		}

		// at this point I'd like to make sure that GetAvailablePackageSummaries returns
		// packages from all repos
		for key, byteArray := range mapReposCached {
			mock.ExpectGet(key).SetVal(string(byteArray))
		}

		resp, err := s.GetAvailablePackageSummaries(context.TODO(),
			&corev1.GetAvailablePackageSummariesRequest{})
		if err != nil {
			t.Fatalf("%v", err)
		}

		// we need to make sure that response contains packages from all existing repositories
		// regardless whether they're in the cache or not
		expected := sets.String{}
		for i := 0; i < len(repos); i++ {
			repo := fmt.Sprintf("bitnami-%d", i)
			expected.Insert(repo)
		}
		for _, s := range resp.AvailablePackageSummaries {
			id := strings.Split(s.AvailablePackageRef.Identifier, "/")
			expected.Delete(id[0])
		}

		if expected.Len() != 0 {
			t.Fatalf("Expected to get packages from these repositories: %s, but did not get any",
				expected.List())
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
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
			t.Fatalf("%v", err)
		}

		// at this point I'd like to make sure that GetAvailablePackageSummaries returns
		// packages from all repos
		mock.ExpectGet(key).SetVal(string(byteArray))

		resp, err := s.GetAvailablePackageSummaries(context.TODO(),
			&corev1.GetAvailablePackageSummariesRequest{})
		if err != nil {
			t.Fatalf("%v", err)
		}

		// we need to make sure that response contains packages from all existing repositories
		// regardless whether they're in the cache or not
		expected := sets.String{}
		expected.Insert(repoName)
		for _, s := range resp.AvailablePackageSummaries {
			id := strings.Split(s.AvailablePackageRef.Identifier, "/")
			expected.Delete(id[0])
		}

		if expected.Len() != 0 {
			t.Fatalf("Expected to get packages from these repositories: %s, but did not get any",
				expected.List())
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
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
		expectedRepo          *sourcev1.HelmRepository
		statusCode            codes.Code
		existingSecret        *apiv1.Secret
		expectedCreatedSecret *apiv1.Secret
		userManagedSecrets    bool
	}{
		{
			name:       "returns error if no namespace is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if no name is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Namespace: "foo"}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if namespaced scoped",
			request:    add_repo_req_1,
			statusCode: codes.Unimplemented,
		},
		{
			name:       "returns error if wrong repository type",
			request:    add_repo_req_2,
			statusCode: codes.Unimplemented,
		},
		{
			name:       "returns error if no url",
			request:    add_repo_req_3,
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if insecureskipverify is set",
			request:    add_repo_req_4,
			statusCode: codes.InvalidArgument,
		},
		{
			name:             "simple add package repository scenario",
			request:          add_repo_req_5,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_1,
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with tls cert authority",
			request:          add_repo_req_6(ca),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_2,
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newTlsSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, nil, nil, ca)),
			statusCode: codes.OK,
		},
		{
			name:       "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:    add_repo_req_7,
			statusCode: codes.InvalidArgument,
		},
		{
			name:             "package repository with secret key reference",
			request:          add_repo_req_7,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_3,
			statusCode:       codes.OK,
			existingSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "foo"}, nil, nil, ca),
			userManagedSecrets: true,
		},
		{
			name:               "failes when package repository links to non-existing secret",
			request:            add_repo_req_7,
			statusCode:         codes.NotFound,
			userManagedSecrets: true,
		},
		{
			name:       "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:    add_repo_req_7,
			statusCode: codes.InvalidArgument,
		},
		{
			name:             "package repository with basic auth and pass_credentials flag",
			request:          add_repo_req_8,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_4,
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newBasicAuthSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, "baz", "zot")),
			statusCode: codes.OK,
		},
		{
			name:             "package repository with TLS authentication",
			request:          add_repo_req_9(pub, priv),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_2,
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newTlsSecret(types.NamespacedName{
					Name:      "bar-",
					Namespace: "foo"}, pub, priv, nil)),
			statusCode: codes.OK,
		},
		{
			name:       "errors for package repository with bearer token",
			request:    add_repo_req_10,
			statusCode: codes.Internal,
		},
		{
			name:       "errors for package repository with custom auth token",
			request:    add_repo_req_11,
			statusCode: codes.Internal,
		},
		{
			name:       "package repository with docker config JSON authentication",
			request:    add_repo_req_12,
			statusCode: codes.Internal,
		},
		{
			name:             "package repository with basic auth and existing secret",
			request:          add_repo_req_13,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_3,
			existingSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "foo"}, "baz", "zot"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
		},
		{
			name:       "package repository with basic auth and existing secret (kubeapps managed secrets)",
			request:    add_repo_req_13,
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:    add_repo_req_14,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            add_repo_req_14,
			statusCode:         codes.InvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          add_repo_req_20,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_5,
			statusCode:       codes.OK,
		},
		{
			name:             "add basic OCI package repository",
			request:          add_repo_req_26,
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_6,
			statusCode:       codes.OK,
		},
		{
			name:             "add OCI package repository with gcp provider",
			request:          add_repo_req_29(),
			expectedResponse: add_repo_expected_resp,
			expectedRepo:     &add_repo_7,
			statusCode:       codes.OK,
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
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets

			nsname := types.NamespacedName{Namespace: tc.request.Context.Namespace, Name: tc.request.Name}
			if tc.statusCode == codes.OK {
				key, err := redisKeyForRepoNamespacedName(nsname)
				if err != nil {
					t.Fatal(err)
				}
				mock.ExpectGet(key).RedisNil()
			}

			ctx := context.Background()
			response, err := s.AddPackageRepository(ctx, tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.AddPackageRepositoryResponse{},
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
					)
					if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// purposefully not calling mock.ExpectationsWereMet() here because
			// AddPackageRepository will trigger an ADD event that will be processed
			// asynchronously, so it may or may not have enough time to get to the
			// point where the cache worker does a GET

			// We don't need to check anything else for non-OK codes.
			if tc.statusCode != codes.OK {
				return
			}

			// check expected HelmReleass CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				var actualRepo sourcev1.HelmRepository
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				} else {
					if tc.userManagedSecrets {
						if tc.expectedCreatedSecret != nil {
							t.Fatalf("Error: unexpected state")
						}
						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}
					} else {
						opt1 := cmpopts.IgnoreFields(sourcev1.HelmRepositorySpec{}, "SecretRef")

						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got, opt1) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
						}

						if tc.expectedCreatedSecret != nil {
							if !strings.HasPrefix(actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name) {
								t.Errorf("SecretRef [%s] was expected to start with [%s]",
									actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name)
							}
							opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
							// check expected secret has been created
							if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
								t.Fatal(err)
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
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetPackageRepositoryDetailResponse
		userManagedSecrets bool
	}{
		{
			name:               "get package repository detail simplest case",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_1,
		},
		{
			name:               "fails with NotFound when wrong identifier",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_2,
			expectedStatusCode: codes.NotFound,
		},
		{
			name:               "fails with NotFound when wrong namespace",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_3,
			expectedStatusCode: codes.NotFound,
		},
		{
			name:               "it returns an invalid arg error status if no context is provided",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_4,
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name:               "it returns an error status if cluster is not the global/kubeapps one",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_5,
			expectedStatusCode: codes.Unimplemented,
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
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_6,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with TLS cert authority (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, nil, nil, ca),
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_6a,
		},
		{
			name:               "get package repository with pending status",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_7,
			pending:            true,
		},
		{
			name:               "get package repository with failed status",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_8,
			failed:             true,
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
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_9,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with TLS cert authentication (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, pub, priv, nil),
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_9a,
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
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_10,
			userManagedSecrets: true,
		},
		{
			name:          "it returns package repository detail with basic authentication (kubeapps managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			repoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_10a,
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
			var repo *sourcev1.HelmRepository
			if !tc.pending && !tc.failed {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, secretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else if tc.pending {
				repoSpec := &sourcev1.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1.HelmRepositoryStatus{
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
				repoSpec := &sourcev1.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1.HelmRepositoryStatus{
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
			s, _, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets

			ctx := context.Background()
			actualResp, err := s.GetPackageRepositoryDetail(ctx, tc.request)
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedStatusCode == codes.OK {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
						corev1.GetPackageRepositoryDetailResponse{},
						corev1.PackageRepositoryDetail{},
						corev1.PackageRepositoryStatus{},
						corev1.PackageRepositoryAuth{},
						corev1.PackageRepositoryTlsConfig{},
						corev1.SecretKeyReference{},
						corev1.TlsCertKey{},
						corev1.UsernamePassword{},
					)
					if got, want := actualResp, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
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
		name               string
		request            *corev1.GetPackageRepositoryDetailRequest
		repoName           string
		repoNamespace      string
		repoUrl            string
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetPackageRepositoryDetailResponse
		seedData           *fakeRemoteOciRegistryData
	}{
		{
			name:               "get package repository detail for OCI repository",
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			repoUrl:            "oci://localhost:54321/userX/charts",
			request:            get_repo_detail_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   get_repo_detail_resp_19,
			seedData:           seed_data_1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)

			repo, err := newOciRepo(tc.repoName, tc.repoNamespace, tc.repoUrl)
			if err != nil {
				t.Fatal(err)
			}

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, nil)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			ctx := context.Background()
			actualResp, err := s.GetPackageRepositoryDetail(ctx, tc.request)
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedStatusCode == codes.OK {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
						corev1.GetPackageRepositoryDetailResponse{},
						corev1.PackageRepositoryDetail{},
						corev1.PackageRepositoryStatus{},
						corev1.PackageRepositoryAuth{},
						corev1.PackageRepositoryTlsConfig{},
						corev1.SecretKeyReference{},
						corev1.TlsCertKey{},
						corev1.UsernamePassword{},
					)
					if got, want := actualResp, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// FWIW GetPackageRepositoryDetail curently does not use the redis cache
			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
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
		name               string
		request            *corev1.GetPackageRepositorySummariesRequest
		existingRepos      []sourcev1.HelmRepository
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetPackageRepositorySummariesResponse
	}{
		{
			name: "returns package summaries when namespace not specified",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{},
			},
			existingRepos: []sourcev1.HelmRepository{
				get_summaries_repo_1,
				get_summaries_repo_2,
				get_summaries_repo_3,
				get_summaries_repo_4,
			},
			expectedStatusCode: codes.OK,
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
			existingRepos: []sourcev1.HelmRepository{
				get_summaries_repo_1,
				get_summaries_repo_2,
				get_summaries_repo_3,
				get_summaries_repo_4,
			},
			expectedStatusCode: codes.OK,
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

			response, err := s.GetPackageRepositorySummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositorySummariesResponse{},
				corev1.PackageRepositorySummary{},
				corev1.PackageRepositoryReference{},
				corev1.PackageRepositoryStatus{},
			)
			opts2 := cmpopts.SortSlices(lessPackageRepositorySummaryFunc)
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts, opts2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdatePackageRepository(t *testing.T) {
	ca, pub, priv := getCertsForTesting(t)
	testCases := []struct {
		name               string
		request            *corev1.UpdatePackageRepositoryRequest
		repoIndex          string
		repoName           string
		repoNamespace      string
		oldRepoSecret      *apiv1.Secret
		newRepoSecret      *apiv1.Secret
		pending            bool
		expectedStatusCode codes.Code
		expectedResponse   *corev1.UpdatePackageRepositoryResponse
		expectedDetail     *corev1.GetPackageRepositoryDetailResponse
		userManagedSecrets bool
	}{
		{
			name:               "update repository url",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            update_repo_req_1,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_1,
		},
		{
			name:               "update repository poll interval",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            update_repo_req_2,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_2,
		},
		{
			name:               "update repository pass credentials flag",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            update_repo_req_3,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_3,
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_7,
			userManagedSecrets: true,
		},
		{
			name:               "update repository set TLS cert/key (kubeapps-managed secrets)",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            update_repo_req_8(pub, priv),
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_8,
		},
		{
			name:          "update repository unset TLS cert/key (kubeapps-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, pub, priv, nil),
			request:            update_repo_req_9,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_9,
		},
		{
			name:          "update repository change from TLS cert/key to basic auth (kubeapps-managed secrets)",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newTlsSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, pub, priv, nil),
			request:            update_repo_req_10,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_10,
		},
		{
			name:               "updates to pending repo is not allowed",
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			request:            update_repo_req_1,
			expectedStatusCode: codes.Internal,
			pending:            true,
		},
		{
			name:          "updates url for repo preserve secret in kubeapps managed env",
			repoIndex:     testYaml("valid-index.yaml"),
			repoName:      "repo-1",
			repoNamespace: "namespace-1",
			oldRepoSecret: newBasicAuthSecret(types.NamespacedName{
				Name:      "secret-1",
				Namespace: "namespace-1"}, "foo", "bar"),
			request:            update_repo_req_16,
			expectedStatusCode: codes.OK,
			expectedResponse:   update_repo_resp_1,
			expectedDetail:     update_repo_detail_15,
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
			var repo *sourcev1.HelmRepository
			if !tc.pending {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, oldSecretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else {
				repoSpec := &sourcev1.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1.HelmRepositoryStatus{
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
			s, _, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets

			ctx := context.Background()
			actualResp, err := s.UpdatePackageRepository(ctx, tc.request)
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedStatusCode == codes.OK {
				if actualResp == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
						corev1.UpdatePackageRepositoryResponse{},
					)
					if got, want := actualResp, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			} else {
				// We don't need to check anything else for non-OK codes.
				return
			}

			actualDetail, err := s.GetPackageRepositoryDetail(ctx, &corev1.GetPackageRepositoryDetailRequest{
				PackageRepoRef: actualResp.PackageRepoRef,
			})
			if got, want := status.Code(err), codes.OK; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if actualDetail == nil {
				t.Fatalf("got: nil, want: detail")
			} else {
				opt1 := cmpopts.IgnoreUnexported(
					corev1.Context{},
					corev1.PackageRepositoryReference{},
					plugins.Plugin{},
					corev1.GetPackageRepositoryDetailResponse{},
					corev1.PackageRepositoryDetail{},
					corev1.PackageRepositoryStatus{},
					corev1.PackageRepositoryAuth{},
					corev1.PackageRepositoryTlsConfig{},
					corev1.SecretKeyReference{},
					corev1.TlsCertKey{},
					corev1.UsernamePassword{},
				)
				if got, want := actualDetail, tc.expectedDetail; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}

			if !tc.userManagedSecrets && tc.oldRepoSecret != nil && actualDetail.Detail.Auth.Type == corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
				// check the secret's been deleted
				if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
					t.Fatal(err)
				} else if _, err = typedClient.CoreV1().Secrets(tc.repoNamespace).Get(ctx, tc.oldRepoSecret.Name, metav1.GetOptions{}); err == nil {
					t.Fatalf("Expected secret [%q] to have been deleted", tc.oldRepoSecret.Name)
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
		expectedStatusCode codes.Code
		userManagedSecrets bool
	}{
		{
			name:               "delete repository",
			request:            delete_repo_req_1,
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			expectedStatusCode: codes.OK,
		},
		{
			name:               "returns not found if package repo doesn't exist",
			request:            delete_repo_req_2,
			repoIndex:          testYaml("valid-index.yaml"),
			repoName:           "repo-1",
			repoNamespace:      "namespace-1",
			expectedStatusCode: codes.NotFound,
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
			var repo *sourcev1.HelmRepository
			if !tc.pending {
				var ts *httptest.Server
				var err error
				ts, repo, err = newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, nil, oldSecretRef)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts.Close()
			} else {
				repoSpec := &sourcev1.HelmRepositorySpec{
					URL:      "https://example.repo.com/charts",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				}
				repoStatus := &sourcev1.HelmRepositoryStatus{
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
			s, _, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, secrets)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets

			ctx := context.Background()
			ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster)
			if err != nil {
				t.Fatal(err)
			}
			nsname := types.NamespacedName{
				Namespace: tc.request.PackageRepoRef.Context.Namespace,
				Name:      tc.request.PackageRepoRef.Identifier,
			}
			var actualRepo sourcev1.HelmRepository
			if tc.expectedStatusCode == codes.OK {
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				}
			}

			_, err = s.DeletePackageRepository(ctx, tc.request)
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.expectedStatusCode == codes.OK {
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
		expectedErrorCode codes.Code
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
			expectedErrorCode: codes.OK,
			seedData:          seed_data_1,
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
			expectedErrorCode: codes.OK,
			seedData:          seed_data_3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)

			repos := []sourcev1.HelmRepository{}

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
				t.Fatalf("%v", err)
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), tc.request)
			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %v, want: %v, err: %v", got, want, err)
			}
			// If an error code was expected, then no need to continue checking
			// the response.
			if tc.expectedErrorCode != codes.OK {
				return
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageSummariesResponse{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
			opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
		})
	}
}

func newRepo(name string, namespace string, spec *sourcev1.HelmRepositorySpec, status *sourcev1.HelmRepositoryStatus) sourcev1.HelmRepository {
	helmRepository := sourcev1.HelmRepository{
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

// these functiosn should affect only unit test, not production code
// does a series of mock.ExpectGet(...)
func (s *Server) redisMockExpectGetFromRepoCache(mock redismock.ClientMock, filterOptions *corev1.FilterOptions, repos ...sourcev1.HelmRepository) error {
	mapVals := make(map[string][]byte)
	ociRepoKeys := sets.String{}
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
				// see comment in chart.go about caching helm OCI chart repos
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

func (s *Server) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1.HelmRepository, oldValue []byte) (key string, bytes []byte, err error) {
	backgroundClientGetter := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return s.clientGetter(ctx, s.kubeappsCluster)
	}
	sink := repoEventSink{
		clientGetter: backgroundClientGetter,
		chartCache:   nil,
	}
	return sink.redisMockSetValueForRepo(mock, repo, oldValue)
}

func (sink *repoEventSink) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1.HelmRepository, oldValue []byte) (key string, newValue []byte, err error) {
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

func (s *Server) redisKeyValueForRepo(r sourcev1.HelmRepository) (key string, byteArray []byte, err error) {
	cg := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return s.clientGetter(ctx, s.kubeappsCluster)
	}
	sinkNoChartCache := repoEventSink{clientGetter: cg}
	return sinkNoChartCache.redisKeyValueForRepo(r)
}

func (sink *repoEventSink) redisKeyValueForRepo(r sourcev1.HelmRepository) (key string, byteArray []byte, err error) {
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

func redisKeyForRepo(r sourcev1.HelmRepository) (string, error) {
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

func newHttpRepoAndServeIndex(repoIndex, repoName, repoNamespace string, replaceUrls map[string]string, secretRef string) (*httptest.Server, *sourcev1.HelmRepository, error) {
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

	repoSpec := &sourcev1.HelmRepositorySpec{
		URL:      "https://example.repo.com/charts",
		Interval: metav1.Duration{Duration: 1 * time.Minute},
	}

	if secretRef != "" {
		repoSpec.SecretRef = &fluxmeta.LocalObjectReference{Name: secretRef}
	}

	revision := "651f952130ea96823711d08345b85e82be011dc6"
	sz := int64(31989)

	repoStatus := &sourcev1.HelmRepositoryStatus{
		Artifact: &sourcev1.Artifact{
			Path:           fmt.Sprintf("helmrepository/%s/%s/index-%s.yaml", repoNamespace, repoName, revision),
			Checksum:       revision,
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

func newOciRepo(repoName, repoNamespace, repoUrl string) (*sourcev1.HelmRepository, error) {
	timeout := metav1.Duration{Duration: 60 * time.Second}
	repoSpec := &sourcev1.HelmRepositorySpec{
		URL:      repoUrl,
		Interval: metav1.Duration{Duration: 1 * time.Minute},
		Timeout:  &timeout,
		Type:     "oci",
	}

	repoStatus := &sourcev1.HelmRepositoryStatus{
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
