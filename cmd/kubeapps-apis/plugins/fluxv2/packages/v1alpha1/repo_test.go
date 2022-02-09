// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	typfake "k8s.io/client-go/kubernetes/fake"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_package_summaries,
			},
		},
		{
			name: "it returns a couple of fluxv2 packages from the cluster (when request namespace is specified)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "default"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_package_summaries,
			},
		},
		{
			name: "it returns a couple of fluxv2 packages from the cluster (when request cluster is specified and matches the kubeapps cluster)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: valid_index_package_summaries,
			},
		},
		{
			name: "it returns all fluxv2 packages from the cluster (when request namespace is does not match repo namespace)",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     "testdata/jetstack-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "non-default"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: append(valid_index_package_summaries, cert_manager_summary),
			},
		},
		{
			name: "uses a filter based on existing repo",
			repos: []testSpecGetAvailablePackageSummaries{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     "testdata/jetstack-index.yaml",
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
					index:     "testdata/valid-index.yaml",
				},
				{
					name:      "jetstack-1",
					namespace: "ns1",
					url:       "https://charts.jetstack.io",
					index:     "testdata/jetstack-index.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
					index:     "testdata/index-with-categories.yaml",
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
				ts2, repo, err := newRepoWithIndex(rs.index, rs.name, rs.namespace, nil, "")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()
				repos = append(repos, *repo)
			}

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, mock, err := newServerWithRepos(t, repos, nil, nil)
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
				index:     "testdata/index-with-categories.yaml",
			},
		}
		repos := []sourcev1.HelmRepository{}
		for _, rs := range existingRepos {
			ts2, repo, err := newRepoWithIndex(rs.index, rs.name, rs.namespace, nil, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()
			repos = append(repos, *repo)
		}

		// the index.yaml will contain links to charts but for the purposes
		// of this test they do not matter
		s, mock, err := newServerWithRepos(t, repos, nil, nil)
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
		indexYamlBeforeUpdateBytes, err := ioutil.ReadFile("testdata/index-before-update.yaml")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		indexYamlAfterUpdateBytes, err := ioutil.ReadFile("testdata/index-after-update.yaml")
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

		lastUpdateTime, err := time.Parse(time.RFC3339, "2021-07-01T05:09:45Z")
		if err != nil {
			t.Fatalf("%v", err)
		}

		repoStatus := &sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Checksum:       "651f952130ea96823711d08345b85e82be011dc6",
				LastUpdateTime: metav1.Time{Time: lastUpdateTime},
				Revision:       "651f952130ea96823711d08345b85e82be011dc6",
			},
			Conditions: []metav1.Condition{
				{
					Type:   "Ready",
					Status: "True",
					Reason: sourcev1.IndexationSucceededReason,
				},
			},
			URL: ts.URL,
		}

		repoName := types.NamespacedName{Namespace: "ns2", Name: "testrepo"}
		repo := newRepo(repoName.Name, repoName.Namespace, repoSpec, repoStatus)

		s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{repo}, nil, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
			t.Fatalf("%v", err)
		}

		responseBeforeUpdate, err := s.GetAvailablePackageSummaries(
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
		if got, want := responseBeforeUpdate.AvailablePackageSummaries, index_before_update_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// see below
		key, oldValue, err := s.redisKeyValueForRepo(repo)
		if err != nil {
			t.Fatalf("%v", err)
		}

		ctx := context.Background()
		if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Get(ctx, repoName, &repo); err != nil {
			t.Fatal(err)
		} else {
			updateHappened = true
			// now we are going to simulate flux seeing an update of the index.yaml and modifying the
			// HelmRepository CRD which, in turn, causes k8s server to fire a MODIFY event
			repo.Status.Artifact.Checksum = "4e881a3c34a5430c1059d2c4f753cb9aed006803"
			repo.Status.Artifact.Revision = "4e881a3c34a5430c1059d2c4f753cb9aed006803"

			// there will be a GET to retrieve the old value from the cache followed by a SET to new value
			mock.ExpectGet(key).SetVal(string(oldValue))
			key, newValue, err := s.redisMockSetValueForRepo(mock, repo)
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
				context.Background(),
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
			tarGzBytes, err := ioutil.ReadFile(s.tgzFile)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write(tarGzBytes)
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
		ts, repo, err := newRepoWithIndex(
			"testdata/valid-index.yaml", repoName.Name, repoName.Namespace, replaceUrls, "")
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
		if got, want := responseBeforeDelete.AvailablePackageSummaries, valid_index_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
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
		ts2, repo, err := newRepoWithIndex("testdata/valid-index.yaml", "bitnami-1", "default", nil, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, nil, nil)
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
		if got, want := responseBeforeResync.AvailablePackageSummaries, valid_index_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		resyncCh, err := s.repoCache.ExpectResync()
		if err != nil {
			t.Fatalf("%v", err)
		}

		// now lets try to simulate HTTP 410 GONE exception which should force RetryWatcher to stop and force
		// a cache resync. The ERROR eventwhich we'll send below should trigger a re-sync of the cache in the
		// background: a FLUSHDB followed by a SET
		ctx := context.Background()
		var watcher *watch.RaceFreeFakeWatcher
		if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else if ww, ok := ctrlClient.(*withWatchWrapper); !ok {
			t.Fatalf("Unexpected condition: %s", reflect.TypeOf(ww))
		} else if watcher = ww.watcher; watcher == nil {
			t.Fatalf("Unexpected condition: watcher is nil")
		}

		watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)

		// wait for the server to start the resync process. Don't care how big the work queue is
		<-resyncCh

		// set up expectations
		mock.ExpectFlushDB().SetVal("OK")
		if _, _, err := s.redisMockSetValueForRepo(mock, *repo); err != nil {
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

		if got, want := responseAfterResync.AvailablePackageSummaries, valid_index_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}
	})
}

// test that causes RetryWatcher to stop and the cache needs to resync when there are
// lots of pending work items
// this test is focused on the repo cache work queue
func TestGetAvailablePackageSummariesAfterCacheResyncQueueNotIdle(t *testing.T) {
	t.Run("test that causes RetryWatcher to stop and the repo cache needs to resync", func(t *testing.T) {
		// start with an empty server that only has an empty repo cache
		s, mock, err := newServerWithRepos(t, nil, nil, nil)
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

			ts, repo, err := newRepoWithIndex("testdata/valid-index.yaml", repoName, "default", nil, "")
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
			mock.ExpectGet(key).RedisNil()
			redisMockSetValueForRepo(mock, key, byteArray)
			repos = append(repos, repo)
		}

		s.repoCache.ExpectAdd(keysInOrder[0])

		var watcher *watch.RaceFreeFakeWatcher
		ctx := context.Background()
		if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else {
			for _, r := range repos {
				if err = ctrlClient.Create(ctx, r); err != nil {
					t.Fatal(err)
				}
			}
			if ww, ok := ctrlClient.(*withWatchWrapper); !ok {
				t.Fatalf("Unexpected condition: %s", reflect.TypeOf(ww))
			} else if watcher = ww.watcher; watcher == nil {
				t.Fatalf("Unexpected condition watcher is nil")
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
					redisMockSetValueForRepo(mock, keysInOrder[i], mapReposCached[keysInOrder[i]])
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
		s, mock, err := newServerWithRepos(t, nil, nil, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		// first, I'd like to make sure there is a single item in the queue
		repoName := "bitnami-0"
		repoNamespace := "default"

		ts, repo, err := newRepoWithIndex("testdata/valid-index.yaml", repoName, repoNamespace, nil, "")
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
		mock.ExpectGet(key).RedisNil()
		redisMockSetValueForRepo(mock, key, byteArray)

		s.repoCache.ExpectAdd(key)

		var watcher *watch.RaceFreeFakeWatcher
		ctx := context.Background()
		if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Create(ctx, repo); err != nil {
			t.Fatal(err)
		} else if ww, ok := ctrlClient.(*withWatchWrapper); !ok {
			t.Fatalf("Unexpected condition: %s", reflect.TypeOf(ww))
		} else if watcher = ww.watcher; watcher == nil {
			t.Fatalf("Unexpected condition: watcher is nil")
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
				redisMockSetValueForRepo(mock, key, byteArray)
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

func TestGetPackageRepositories(t *testing.T) {
	testCases := []struct {
		name                        string
		request                     *v1alpha1.GetPackageRepositoriesRequest
		repoNamespace               string
		repoSpecs                   map[string]sourcev1.HelmRepositorySpec
		expectedPackageRepositories []*v1alpha1.PackageRepository
		statusCode                  codes.Code
	}{
		{
			name:          "returns an internal error status if item in response cannot be converted to v1alpha1.PackageRepository",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "default",
			repoSpecs: map[string]sourcev1.HelmRepositorySpec{
				"repo-1": {},
			},
			statusCode: codes.Internal,
		},
		{
			name:          "returns expected repositories",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "default",
			repoSpecs: map[string]sourcev1.HelmRepositorySpec{
				"repo-1": {
					URL: "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					URL: "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{
				{
					Name:      "repo-1",
					Namespace: "default",
					Url:       "https://charts.bitnami.com/bitnami",
				},
				{
					Name:      "repo-2",
					Namespace: "default",
					Url:       "https://charts.helm.sh/stable",
				},
			},
		},
		{
			name: "returns expected repositories in specific namespace",
			request: &v1alpha1.GetPackageRepositoriesRequest{
				Context: &corev1.Context{
					Namespace: "default",
				},
			},
			repoNamespace: "non-default",
			repoSpecs: map[string]sourcev1.HelmRepositorySpec{
				"repo-1": {
					URL: "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					URL: "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{},
		},
		{
			name: "returns expected repositories in specific namespace (2)",
			request: &v1alpha1.GetPackageRepositoriesRequest{
				Context: &corev1.Context{
					Namespace: "default",
				},
			},
			repoNamespace: "default",
			repoSpecs: map[string]sourcev1.HelmRepositorySpec{
				"repo-1": {
					URL: "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					URL: "https://charts.helm.sh/stable",
				},
			},
			expectedPackageRepositories: []*v1alpha1.PackageRepository{
				{
					Name:      "repo-1",
					Namespace: "default",
					Url:       "https://charts.bitnami.com/bitnami",
				},
				{
					Name:      "repo-2",
					Namespace: "default",
					Url:       "https://charts.helm.sh/stable",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, err := newServerWithRepos(t, newRepos(tc.repoSpecs, tc.repoNamespace), nil, nil)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			response, err := s.GetPackageRepositories(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(v1alpha1.PackageRepository{}, corev1.Context{})
					opt2 := cmpopts.SortSlices(lessPackageRepositoryFunc)
					if got, want := response.Repositories, tc.expectedPackageRepositories; !cmp.Equal(got, want, opt1, opt2) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
					}
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func newServerWithRepos(t *testing.T, repos []sourcev1.HelmRepository, charts []testSpecChartWithUrl, secrets []runtime.Object) (*Server, redismock.ClientMock, error) {
	typedClient := typfake.NewSimpleClientset(secrets...)
	apiextIfc := apiextfake.NewSimpleClientset(fluxHelmRepositoryCRD)

	// register the GitOps Toolkit schema definitions
	scheme := runtime.NewScheme()
	_ = sourcev1.AddToScheme(scheme)
	_ = helmv2.AddToScheme(scheme)

	rm := apimeta.NewDefaultRESTMapper([]schema.GroupVersion{sourcev1.GroupVersion, helmv2.GroupVersion})
	rm.Add(schema.GroupVersionKind{
		Group:   sourcev1.GroupVersion.Group,
		Version: sourcev1.GroupVersion.Version,
		Kind:    sourcev1.HelmRepositoryKind},
		apimeta.RESTScopeNamespace)
	rm.Add(schema.GroupVersionKind{
		Group:   helmv2.GroupVersion.Group,
		Version: helmv2.GroupVersion.Version,
		Kind:    helmv2.HelmReleaseKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	if len(repos) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(&sourcev1.HelmRepositoryList{Items: repos})
	}
	ctrlClient := &withWatchWrapper{delegate: ctrlClientBuilder.Build()}

	clientGetter := func(context.Context, string) (clientgetter.ClientInterfaces, error) {
		return clientgetter.
			NewBuilder().
			WithTyped(typedClient).
			WithApiExt(apiextIfc).
			WithControllerRuntime(ctrlClient).
			Build(), nil
	}

	return newServer(t, clientGetter, nil, repos, charts)
}

func newRepo(name string, namespace string, spec *sourcev1.HelmRepositorySpec, status *sourcev1.HelmRepositoryStatus) sourcev1.HelmRepository {
	helmRepository := sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Generation: int64(1),
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
		helmRepository.Status.ObservedGeneration = int64(1)
	}

	return helmRepository
}

// newRepos takes a map of specs keyed by object name converting them to typed flux HelmRepository objects.
func newRepos(specs map[string]sourcev1.HelmRepositorySpec, namespace string) []sourcev1.HelmRepository {
	repos := []sourcev1.HelmRepository{}
	for name, spec := range specs {
		repo := newRepo(name, namespace, &spec, nil)
		repos = append(repos, repo)
	}
	return repos
}

// these functiosn should affect only unit test, not production code
// does a series of mock.ExpectGet(...)
func (s *Server) redisMockExpectGetFromRepoCache(mock redismock.ClientMock, filterOptions *corev1.FilterOptions, repos ...sourcev1.HelmRepository) error {
	mapVals := make(map[string][]byte)
	for _, r := range repos {
		key, bytes, err := s.redisKeyValueForRepo(r)
		if err != nil {
			return err
		}
		mapVals[key] = bytes
	}
	if filterOptions == nil || len(filterOptions.GetRepositories()) == 0 {
		for k, v := range mapVals {
			mock.ExpectGet(k).SetVal(string(v))
		}
	} else {
		for _, r := range filterOptions.GetRepositories() {
			for k, v := range mapVals {
				if strings.HasSuffix(k, ":"+r) {
					mock.ExpectGet(k).SetVal(string(v))
				}
			}
		}
	}
	return nil
}

func (s *Server) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1.HelmRepository) (key string, bytes []byte, err error) {
	backgroundClientGetter := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return s.clientGetter(ctx, s.kubeappsCluster)
	}
	sink := repoEventSink{
		clientGetter: backgroundClientGetter,
		chartCache:   nil,
	}
	return sink.redisMockSetValueForRepo(mock, repo)
}

func (sink *repoEventSink) redisMockSetValueForRepo(mock redismock.ClientMock, repo sourcev1.HelmRepository) (key string, byteArray []byte, err error) {
	if key, err = redisKeyForRepo(repo); err != nil {
		return key, nil, err
	}
	if key, byteArray, err = sink.redisKeyValueForRepo(repo); err != nil {
		mock.ExpectDel(key).SetVal(0)
		return key, nil, err
	} else {
		redisMockSetValueForRepo(mock, key, byteArray)
		return key, byteArray, nil
	}
}

func redisMockSetValueForRepo(mock redismock.ClientMock, key string, byteArray []byte) {
	mock.ExpectSet(key, byteArray, 0).SetVal("OK")
	mock.ExpectInfo("memory").SetVal("used_memory_rss_human:NA\r\nmaxmemory_human:NA")
}

func (s *Server) redisKeyValueForRepo(r sourcev1.HelmRepository) (key string, byteArray []byte, err error) {
	sink := repoEventSink{
		clientGetter: s.newBackgroundClientGetter(),
		chartCache:   nil,
	}
	return sink.redisKeyValueForRepo(r)
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

func newRepoWithIndex(repoIndex, repoName, repoNamespace string, replaceUrls map[string]string, secretRef string) (*httptest.Server, *sourcev1.HelmRepository, error) {
	indexYAMLBytes, err := ioutil.ReadFile(repoIndex)
	if err != nil {
		return nil, nil, err
	}

	for k, v := range replaceUrls {
		indexYAMLBytes = []byte(strings.ReplaceAll(string(indexYAMLBytes), k, v))
	}

	// stand up a plain text http server to server the contents of index.yaml just for the
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
		repoSpec.SecretRef = &meta.LocalObjectReference{Name: secretRef}
	}

	lastUpdateTime, err := time.Parse(time.RFC3339, "2021-07-01T05:09:45Z")
	if err != nil {
		return nil, nil, err
	}

	repoStatus := &sourcev1.HelmRepositoryStatus{
		Artifact: &sourcev1.Artifact{
			Checksum:       "651f952130ea96823711d08345b85e82be011dc6",
			LastUpdateTime: metav1.Time{Time: lastUpdateTime},
			Revision:       "651f952130ea96823711d08345b85e82be011dc6",
		},
		Conditions: []metav1.Condition{
			{
				Type:   "Ready",
				Status: "True",
				Reason: sourcev1.IndexationSucceededReason,
			},
		},
		URL: ts.URL,
	}
	repo := newRepo(repoName, repoNamespace, repoSpec, repoStatus)
	return ts, &repo, nil
}

// misc global vars that get re-used in multiple tests scenarios
var repositoriesGvr = schema.GroupVersionResource{
	Group:    sourcev1.GroupVersion.Group,
	Version:  sourcev1.GroupVersion.Version,
	Resource: fluxHelmRepositories,
}

var valid_index_charts_spec = []testSpecChartWithFile{
	{
		name:     "acs-engine-autoscaler",
		tgzFile:  "testdata/acs-engine-autoscaler-2.1.1.tgz",
		revision: "2.1.1",
	},
	{
		name:     "wordpress",
		tgzFile:  "testdata/wordpress-0.7.5.tgz",
		revision: "0.7.5",
	},
	{
		name:     "wordpress",
		tgzFile:  "testdata/wordpress-0.7.4.tgz",
		revision: "0.7.4",
	},
}

var valid_index_package_summaries = []*corev1.AvailablePackageSummary{
	{
		Name:        "acs-engine-autoscaler",
		DisplayName: "acs-engine-autoscaler",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "2.1.1",
			AppVersion: "2.1.1",
		},
		IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
		ShortDescription: "Scales worker nodes within agent pools",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "bitnami-1/acs-engine-autoscaler",
			Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	},
	{
		Name:        "wordpress",
		DisplayName: "wordpress",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "0.7.5",
			AppVersion: "4.9.1",
		},
		IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
		ShortDescription: "new description!",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "bitnami-1/wordpress",
			Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	},
}

var cert_manager_summary = &corev1.AvailablePackageSummary{
	Name:        "cert-manager",
	DisplayName: "cert-manager",
	LatestVersion: &corev1.PackageAppVersion{
		PkgVersion: "v1.4.0",
		AppVersion: "v1.4.0",
	},
	IconUrl:          "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
	ShortDescription: "A Helm chart for cert-manager",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "jetstack-1/cert-manager",
		Context:    &corev1.Context{Namespace: "ns1", Cluster: KubeappsCluster},
		Plugin:     fluxPlugin,
	},
	Categories: []string{""},
}

var elasticsearch_summary = &corev1.AvailablePackageSummary{
	Name:        "elasticsearch",
	DisplayName: "elasticsearch",
	LatestVersion: &corev1.PackageAppVersion{
		PkgVersion: "15.5.0",
		AppVersion: "7.13.2",
	},
	IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
	ShortDescription: "A highly scalable open-source full-text search and analytics engine",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "index-with-categories-1/elasticsearch",
		Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
		Plugin:     fluxPlugin,
	},
	Categories: []string{"Analytics"},
}

var ghost_summary = &corev1.AvailablePackageSummary{
	Name:        "ghost",
	DisplayName: "ghost",
	LatestVersion: &corev1.PackageAppVersion{
		PkgVersion: "13.0.14",
		AppVersion: "4.7.0",
	},
	IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
	ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Identifier: "index-with-categories-1/ghost",
		Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
		Plugin:     fluxPlugin,
	},
	Categories: []string{"CMS"},
}

var index_with_categories_summaries = []*corev1.AvailablePackageSummary{
	elasticsearch_summary,
	ghost_summary,
}

var index_before_update_summaries = []*corev1.AvailablePackageSummary{
	{
		Name:        "alpine",
		DisplayName: "alpine",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "0.2.0",
		},
		IconUrl:          "",
		ShortDescription: "Deploy a basic Alpine Linux pod",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "testrepo/alpine",
			Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	},
	{
		Name:        "nginx",
		DisplayName: "nginx",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "1.1.0",
		},
		IconUrl:          "",
		ShortDescription: "Create a basic nginx HTTP server",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "testrepo/nginx",
			Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	},
}

var index_after_update_summaries = []*corev1.AvailablePackageSummary{
	{
		Name:        "alpine",
		DisplayName: "alpine",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "0.3.0",
		},
		IconUrl:          "",
		ShortDescription: "Deploy a basic Alpine Linux pod",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "testrepo/alpine",
			Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	},
	{
		Name:        "nginx",
		DisplayName: "nginx",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "1.1.0",
		},
		IconUrl:          "",
		ShortDescription: "Create a basic nginx HTTP server",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "testrepo/nginx",
			Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	}}
