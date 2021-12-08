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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

type testSpecGetAvailablePackageSummaries struct {
	name      string
	namespace string
	url       string
	index     string
}

func TestGetAvailablePackageSummaries(t *testing.T) {
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
			name: "it returns only the first page of results",
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
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "0",
					PageSize:  1,
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					elasticsearch_summary,
				},
				NextPageToken: "1",
			},
		},
		{
			name: "it returns only the requested page of results and includes the next page token",
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
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "1",
					PageSize:  1,
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					ghost_summary,
				},
				NextPageToken: "2",
			},
		},
		{
			name: "it returns the last page without a next page token",
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
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "2",
					PageSize:  1,
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				NextPageToken:             "",
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
			repos := []runtime.Object{}

			for _, rs := range tc.repos {
				ts2, repo, err := newRepoWithIndex(rs.index, rs.name, rs.namespace, nil, "")
				if err != nil {
					t.Fatalf("%+v", err)
				}
				defer ts2.Close()
				repos = append(repos, repo)
			}

			// the index.yaml will contain links to charts but for the purposes
			// of this test they do not matter
			s, mock, _, _, err := newServerWithRepos(t, repos, nil, nil)
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

		repoSpec := map[string]interface{}{
			"url":      "https://example.repo.com/charts",
			"interval": "1m0s",
		}

		repoStatus := map[string]interface{}{
			"artifact": map[string]interface{}{
				"checksum":       "651f952130ea96823711d08345b85e82be011dc6",
				"lastUpdateTime": "2021-07-01T05:09:45Z",
				"revision":       "651f952130ea96823711d08345b85e82be011dc6",
			},
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "True",
					"reason": "IndexationSucceed",
				},
			},
			"url": ts.URL,
		}
		repo := newRepo("testrepo", "ns2", repoSpec, repoStatus)

		s, mock, dyncli, watcher, err := newServerWithRepos(t, []runtime.Object{repo}, nil, nil)
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

		opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeUpdate.AvailablePackageSummaries, index_before_update_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// see below
		key, oldValue, err := s.redisKeyValueForRepo(repo)
		if err != nil {
			t.Fatalf("%v", err)
		}

		updateHappened = true
		// now we are going to simulate flux seeing an update of the index.yaml and modifying the
		// HelmRepository CRD which, in turn, causes k8s server to fire a MODIFY event
		unstructured.SetNestedField(repo.Object, "2", "metadata", "resourceVersion")
		unstructured.SetNestedField(repo.Object, "4e881a3c34a5430c1059d2c4f753cb9aed006803", "status", "artifact", "checksum")
		unstructured.SetNestedField(repo.Object, "4e881a3c34a5430c1059d2c4f753cb9aed006803", "status", "artifact", "revision")
		// there will be a GET to retrieve the old value from the cache followed by a SET to new value
		mock.ExpectGet(key).SetVal(string(oldValue))
		key, newValue, err := s.redisMockSetValueForRepo(mock, repo)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if repo, err = dyncli.Resource(repositoriesGvr).Namespace("ns2").Update(context.Background(), repo, metav1.UpdateOptions{}); err != nil {
			t.Fatalf("%v", err)
		}

		s.repoCache.ExpectAdd(key)
		watcher.Modify(repo)
		s.repoCache.WaitUntilDoneWith(key)

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
	})
}

func TestGetAvailablePackageSummaryAfterFluxHelmRepoDelete(t *testing.T) {
	t.Run("test get available package summaries after flux helm repository CRD gets deleted", func(t *testing.T) {
		repoName := "bitnami-1"
		repoNamespace := "default"
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
				chartID:       fmt.Sprintf("%s/%s", repoName, s.name),
				chartRevision: s.revision,
				chartUrl:      ts.URL,
				repoNamespace: repoNamespace,
			}
			charts = append(charts, c)
		}
		ts, repo, err := newRepoWithIndex("testdata/valid-index.yaml", repoName, repoNamespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts.Close()

		s, mock, dyncli, watcher, err := newServerWithRepos(t, []runtime.Object{repo}, charts, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// we make sure that all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
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

		opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeDelete.AvailablePackageSummaries, valid_index_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// now we are going to simulate the user deleting a HelmRepository CR which, in turn,
		// causes k8s server to fire a DELETE event
		key, err := redisKeyForRepo(repo)
		if err != nil {
			t.Fatalf("%v", err)
		} else {
			mock.ExpectDel(key).SetVal(0)
		}
		redisMockExpectDeleteFromChartCache(mock)
		if err = dyncli.Resource(repositoriesGvr).Namespace("default").Delete(context.Background(), "bitnami-1", metav1.DeleteOptions{}); err != nil {
			t.Fatalf("%v", err)
		}
		// TODO (gfichtenholt)
		// everything hardcoded for one test for now :-)
		// will clean it up when I am done with more important stuff
		s.repoCache.ExpectAdd(key)
		chartCacheKeys := []string{
			"helmcharts:default:bitnami-1/acs-engine-autoscaler:2.1.1",
			"helmcharts:default:bitnami-1/wordpress:0.7.5",
		}
		for _, k := range chartCacheKeys {
			s.chartCache.ExpectAdd(k)
		}
		watcher.Delete(repo)
		s.repoCache.WaitUntilDoneWith(key)
		for _, k := range chartCacheKeys {
			s.chartCache.WaitUntilDoneWith(k)
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

		s, mock, _, watcher, err := newServerWithRepos(t, []runtime.Object{repo}, nil, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
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

		opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeResync.AvailablePackageSummaries, valid_index_package_summaries; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// now lets try to simulate HTTP 410 GONE exception which should force RetryWatcher to stop and force
		// a cache resync. The ERROR eventwhich we'll send below should trigger a re-sync of the cache in the
		// background: a FLUSHDB followed by a SET
		mock.ExpectFlushDB().SetVal("OK")
		if _, _, err := s.redisMockSetValueForRepo(mock, repo); err != nil {
			t.Fatalf("%+v", err)
		}

		s.repoCache.ExpectResync()
		watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)
		s.repoCache.WaitUntilResyncComplete()

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
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

func TestGetPackageRepositories(t *testing.T) {
	testCases := []struct {
		name                        string
		request                     *v1alpha1.GetPackageRepositoriesRequest
		repoNamespace               string
		repoSpecs                   map[string]map[string]interface{}
		expectedPackageRepositories []*v1alpha1.PackageRepository
		statusCode                  codes.Code
	}{
		{
			name:          "returns an internal error status if item in response cannot be converted to v1alpha1.PackageRepository",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "default",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"foo": "bar",
				},
			},
			statusCode: codes.Internal,
		},
		{
			name:          "returns expected repositories",
			request:       &v1alpha1.GetPackageRepositoriesRequest{Context: &corev1.Context{}},
			repoNamespace: "default",
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
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
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
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
			repoSpecs: map[string]map[string]interface{}{
				"repo-1": {
					"url": "https://charts.bitnami.com/bitnami",
				},
				"repo-2": {
					"url": "https://charts.helm.sh/stable",
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
			s, mock, _, _, err := newServerWithRepos(t, newRepos(tc.repoSpecs, tc.repoNamespace), nil, nil)
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

func newServerWithRepos(t *testing.T, repos []runtime.Object, charts []testSpecChartWithUrl, secrets []runtime.Object) (*Server, redismock.ClientMock, *fake.FakeDynamicClient, *watch.FakeWatcher, error) {
	typedClient := typfake.NewSimpleClientset(secrets...)
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			repositoriesGvr: fluxHelmRepositoryList,
		},
		repos...)

	// here we are essentially adding on to how List() works for HelmRepository objects
	// this is done so that the the item list returned by List() command with fake client contains
	// a "resourceVersion" field in its metadata, which happens in a real k8s environment and
	// is critical
	reactor := dynamicClient.Fake.ReactionChain[0]
	dynamicClient.Fake.PrependReactor("list", fluxHelmRepositories,
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			handled, ret, err := reactor.React(action)
			if err == nil {
				ulist, ok := ret.(*unstructured.UnstructuredList)
				if ok && ulist != nil {
					ulist.SetResourceVersion("1")
				}
			}
			return handled, ret, err
		})

	watcher := watch.NewFake()

	dynamicClient.Fake.PrependWatchReactor(
		fluxHelmRepositories,
		k8stesting.DefaultWatchReactor(watcher, nil))

	apiextIfc := apiextfake.NewSimpleClientset(fluxHelmRepositoryCRD)

	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error) {
		return typedClient, dynamicClient, apiextIfc, nil
	}

	s, mock, err := newServer(t, clientGetter, nil, repos, charts)
	return s, mock, dynamicClient, watcher, err
}

func newRepo(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":            name,
		"generation":      int64(1),
		"resourceVersion": "1",
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}
	obj := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
		"kind":       fluxHelmRepository,
		"metadata":   metadata,
	}

	if spec != nil {
		obj["spec"] = spec
	}

	if status != nil {
		status["observedGeneration"] = int64(1)
		obj["status"] = status
	}

	return &unstructured.Unstructured{
		Object: obj,
	}
}

// newRepos takes a map of specs keyed by object name converting them to unstructured objects.
func newRepos(specs map[string]map[string]interface{}, namespace string) []runtime.Object {
	repos := []runtime.Object{}
	for name, spec := range specs {
		repo := newRepo(name, namespace, spec, nil)
		repos = append(repos, repo)
	}
	return repos
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret
func newBasicAuthSecret(name, namespace, user, password string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: k8scorev1.SecretTypeBasicAuth,
		StringData: map[string]string{
			"username": user,
			"password": password,
		},
	}
}

// ref: https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
// ref: https://medium.com/avmconsulting-blog/how-to-secure-applications-on-kubernetes-ssl-tls-certificates-8f7f5751d788
// the shell commands I used to generate the relevant files:
// 1. $ openssl genrsa -out rootCA.key 4096
// 2. $ openssl req -x509 -new -key rootCA.key -days 3650 -out rootCA.crt
//    You are about to be asked to enter information that will be incorporated
//    into your certificate request.
//    What you are about to enter is what is called a Distinguished Name or a DN.
//    There are quite a few fields but you can leave some blank
//    For some fields there will be a default value,
//    If you enter '.', the field will be left blank.
//    -----
//    Country Name (2 letter code) []:US
//    State or Province Name (full name) []:California
//    Locality Name (eg, city) []:Palo Alto
//    Organization Name (eg, company) []:VMware
//    Organizational Unit Name (eg, section) []:MAPBU
//    Common Name (eg, fully qualified host name) []:kubeapps
//    Email Address []:
// 3. $ openssl genrsa -out testTLS.key 2048
// 4. $ openssl req -new -key testTLS.key -out testTLS.csr
//    You are about to be asked to enter information that will be incorporated
//    into your certificate request.
//    What you are about to enter is what is called a Distinguished Name or a DN.
//    There are quite a few fields but you can leave some blank
//    For some fields there will be a default value,
//    If you enter '.', the field will be left blank.
//    -----
//    Country Name (2 letter code) []:US
//    State or Province Name (full name) []:California
//    Locality Name (eg, city) []:Palo Alto
//    Organization Name (eg, company) []:VMware
//    Organizational Unit Name (eg, section) []:MAPBU
//    Common Name (eg, fully qualified host name) []:test
//    Email Address []:
//
//    Please enter the following 'extra' attributes
//    to be sent with your certificate request
//    A challenge password []:
// 5. $ openssl x509 -req -in testTLS.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial -days 365 -out testTLS.crt
// 6. $ openssl x509 -in testTLS.crt -out testTLS.pem
// 7. $ openssl x509 -in rootCA.crt -out rootCA.pem
// To test
// 8. $ kubectl create secret tls my-secret --cert=testTLS.pem --key=testTLS.key
func newTlsSecret(name, namespace string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: k8scorev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": nil,
			"tls.key": nil,
		},
	}
}

// these functiosn should affect only unit test, not production code
// does a series of mock.ExpectGet(...)
func (s *Server) redisMockExpectGetFromRepoCache(mock redismock.ClientMock, filterOptions *corev1.FilterOptions, repos ...runtime.Object) error {
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

func (s *Server) redisMockSetValueForRepo(mock redismock.ClientMock, repo runtime.Object) (key string, bytes []byte, err error) {
	cs := repoCacheCallSite{
		clientGetter: s.clientGetter,
		chartCache:   nil,
	}
	return cs.redisMockSetValueForRepo(mock, repo)
}

func (cs *repoCacheCallSite) redisMockSetValueForRepo(mock redismock.ClientMock, repo runtime.Object) (key string, bytes []byte, err error) {
	if key, err = redisKeyForRepo(repo); err != nil {
		return key, nil, err
	}
	if key, bytes, err = cs.redisKeyValueForRepo(repo); err != nil {
		mock.ExpectDel(key).SetVal(0)
		return key, nil, err
	} else {
		mock.ExpectSet(key, bytes, 0).SetVal("OK")
		mock.ExpectInfo("memory").SetVal("used_memory_rss_human:NA\r\nmaxmemory_human:NA")
		return key, bytes, nil
	}
}

func (s *Server) redisKeyValueForRepo(r runtime.Object) (key string, bytes []byte, err error) {
	cs := repoCacheCallSite{
		clientGetter: s.clientGetter,
		chartCache:   nil,
	}
	return cs.redisKeyValueForRepo(r)
}

func (cs *repoCacheCallSite) redisKeyValueForRepo(r runtime.Object) (key string, bytes []byte, err error) {
	if key, err = redisKeyForRepo(r); err != nil {
		return key, nil, err
	} else {
		// we are not really adding anything to the cache here, rather just calling a
		// onAddRepo to compute the value that *WOULD* be stored in the cache
		var byteArray interface{}
		var add bool
		byteArray, add, err = cs.onAddRepo(key, r.(*unstructured.Unstructured).Object)
		if err != nil {
			return key, nil, err
		} else if !add {
			return key, nil, fmt.Errorf("onAddRepo returned false for setVal")
		}
		return key, byteArray.([]byte), nil
	}
}

func redisKeyForRepo(r runtime.Object) (string, error) {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return redisKeyForRepoNamespacedName(types.NamespacedName{
		Namespace: r.(*unstructured.Unstructured).GetNamespace(),
		Name:      r.(*unstructured.Unstructured).GetName()})
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

func newRepoWithIndex(repoIndex, repoName, repoNamespace string, replaceUrls map[string]string, secretRef string) (*httptest.Server, *unstructured.Unstructured, error) {
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

	repoSpec := map[string]interface{}{
		"url":      "https://example.repo.com/charts",
		"interval": "1m0s",
	}

	if secretRef != "" {
		repoSpec["secretRef"] = map[string]interface{}{
			"name": secretRef,
		}
	}

	repoStatus := map[string]interface{}{
		"artifact": map[string]interface{}{
			"checksum":       "651f952130ea96823711d08345b85e82be011dc6",
			"lastUpdateTime": "2021-07-01T05:09:45Z",
			"revision":       "651f952130ea96823711d08345b85e82be011dc6",
		},
		"conditions": []interface{}{
			map[string]interface{}{
				"type":   "Ready",
				"status": "True",
				"reason": "IndexationSucceed",
			},
		},
		"url": ts.URL,
	}
	return ts, newRepo(repoName, repoNamespace, repoSpec, repoStatus), nil
}

// misc global vars that get re-used in multiple tests scenarios
var repositoriesGvr = schema.GroupVersionResource{
	Group:    fluxGroup,
	Version:  fluxVersion,
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
	},
	{
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
	},
}

var cert_manager_summary = &corev1.AvailablePackageSummary{
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
}

var elasticsearch_summary = &corev1.AvailablePackageSummary{
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
}

var ghost_summary = &corev1.AvailablePackageSummary{
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
}

var index_with_categories_summaries = []*corev1.AvailablePackageSummary{
	elasticsearch_summary,
	ghost_summary,
}

var index_before_update_summaries = []*corev1.AvailablePackageSummary{
	{
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
	},
	{
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
	},
}

var index_after_update_summaries = []*corev1.AvailablePackageSummary{
	{
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
	},
	{
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
	}}
