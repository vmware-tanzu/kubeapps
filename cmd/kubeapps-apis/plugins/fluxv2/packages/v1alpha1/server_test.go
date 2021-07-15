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
	"sync"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8stesting "k8s.io/client-go/testing"
)

func TestNilClientGetter(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter server.KubernetesClientGetter
		statusCode   codes.Code
	}{
		{
			name:         "returns failed-precondition error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, mock, err := newServer(tc.clientGetter)
			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestBadClientGetter(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter server.KubernetesClientGetter
		statusCode   codes.Code
	}{
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, err := newServer(tc.clientGetter)
			if err != nil {
				t.Fatalf("%v", err)
			}

			_, err = s.GetAvailablePackageSummaries(
				context.Background(),
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			err = mock.ExpectationsWereMet()
			if err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name       string
		repo       *unstructured.Unstructured
		statusCode codes.Code
	}{
		{
			name: "returns without error if response status does not contain conditions",
			repo: newRepo("test", "default",
				map[string]interface{}{
					"url":      "http://example.com",
					"interval": "1m0s",
				},
				nil),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response status does not contain conditions (2)",
			repo: newRepo("test", "default",
				map[string]interface{}{
					"url":      "http://example.com",
					"interval": "1m0s",
				},
				map[string]interface{}{
					"zot": "xyz",
				}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain ready repos",
			repo: newRepo("test", "default",
				map[string]interface{}{
					"url":      "http://example.com",
					"interval": "1m0s",
				},
				map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Ready",
							"status": "False",
							"reason": "IndexationFailed",
						},
					}}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain namespace",
			repo: newRepo("test", "", nil, map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "IndexationSucceed",
					},
				}}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain spec url",
			repo: newRepo("test", "default", nil, map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "IndexationSucceed",
					},
				}}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain status url",
			repo: newRepo("test", "default", map[string]interface{}{
				"url":      "http://example.com",
				"interval": "1m0s",
			}, map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "IndexationSucceed",
					},
				}}),
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, _, err := newServerWithWatcher(tc.repo)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			if err = beforeCallGetAvailablePackageSummaries(mock, nil); err != nil {
				t.Fatalf("%v", err)
			}

			response, err := s.GetAvailablePackageSummaries(
				context.Background(),
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, error: %v, want: %+v", got, err, want)

				if got == codes.OK {
					if len(response.AvailablePackagesSummaries) != 0 {
						t.Errorf("unexpected response: %v", response)
					} else if response != nil {
						t.Errorf("unexpected response: %v", response)
					}
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

type testRepoStruct struct {
	name      string
	namespace string
	url       string
	index     string
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		testName         string
		request          *corev1.GetAvailablePackageSummariesRequest
		testRepos        []testRepoStruct
		expectedResponse *corev1.GetAvailablePackageSummariesResponse
	}{
		{
			testName: "it returns a couple of fluxv2 packages from the cluster (no request ns specified)",
			testRepos: []testRepoStruct{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "acs-engine-autoscaler",
						LatestPkgVersion: "2.1.1",
						IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
						ShortDescription: "Scales worker nodes within agent pools",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/acs-engine-autoscaler",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "wordpress",
						LatestPkgVersion: "0.7.5",
						IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
						ShortDescription: "new description!",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/wordpress",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "it returns a couple of fluxv2 packages from the cluster (when request namespace is specified)",
			testRepos: []testRepoStruct{
				{
					name:      "bitnami-1",
					namespace: "default",
					url:       "https://example.repo.com/charts",
					index:     "testdata/valid-index.yaml",
				},
			},
			request: &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{Namespace: "default"}},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "acs-engine-autoscaler",
						LatestPkgVersion: "2.1.1",
						IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
						ShortDescription: "Scales worker nodes within agent pools",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/acs-engine-autoscaler",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "wordpress",
						LatestPkgVersion: "0.7.5",
						IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
						ShortDescription: "new description!",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/wordpress",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "it returns all fluxv2 packages from the cluster (when request namespace is does not match repo namespace)",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "acs-engine-autoscaler",
						LatestPkgVersion: "2.1.1",
						IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
						ShortDescription: "Scales worker nodes within agent pools",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/acs-engine-autoscaler",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "cert-manager",
						LatestPkgVersion: "v1.4.0",
						IconUrl:          "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
						ShortDescription: "A Helm chart for cert-manager",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "jetstack-1/cert-manager",
							Context:    &corev1.Context{Namespace: "ns1"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "wordpress",
						LatestPkgVersion: "0.7.5",
						IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
						ShortDescription: "new description!",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "bitnami-1/wordpress",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on existing repo",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "cert-manager",
						LatestPkgVersion: "v1.4.0",
						IconUrl:          "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
						ShortDescription: "A Helm chart for cert-manager",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "jetstack-1/cert-manager",
							Context:    &corev1.Context{Namespace: "ns1"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on non-existing repo",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			testName: "uses a filter based on existing categories",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "elasticsearch",
						LatestPkgVersion: "15.5.0",
						IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
						ShortDescription: "A highly scalable open-source full-text search and analytics engine",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/elasticsearch",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on existing categories (2)",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "elasticsearch",
						LatestPkgVersion: "15.5.0",
						IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
						ShortDescription: "A highly scalable open-source full-text search and analytics engine",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/elasticsearch",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "ghost",
						LatestPkgVersion: "13.0.14",
						IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
						ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/ghost",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on non-existing categories",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			testName: "uses a filter based on existing appVersion",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "ghost",
						LatestPkgVersion: "13.0.14",
						IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
						ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/ghost",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on non-existing appVersion",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			testName: "uses a filter based on existing pkgVersion",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "elasticsearch",
						LatestPkgVersion: "15.5.0",
						IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
						ShortDescription: "A highly scalable open-source full-text search and analytics engine",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/elasticsearch",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on non-existing pkgVersion",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			testName: "uses a filter based on existing query text (chart name)",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "elasticsearch",
						LatestPkgVersion: "15.5.0",
						IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
						ShortDescription: "A highly scalable open-source full-text search and analytics engine",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/elasticsearch",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on existing query text (chart keywords)",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "ghost",
						LatestPkgVersion: "13.0.14",
						IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
						ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/ghost",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
			},
		},
		{
			testName: "uses a filter based on non-existing query text",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{},
			},
		},
		{
			testName: "it returns only the requested page of results and includes the next page token",
			testRepos: []testRepoStruct{
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
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "ghost",
						LatestPkgVersion: "13.0.14",
						IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
						ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Identifier: "index-with-categories-1/ghost",
							Context:    &corev1.Context{Namespace: "default"},
							Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "2",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repos := []runtime.Object{}

			for _, rs := range tc.testRepos {
				indexYAMLBytes, err := ioutil.ReadFile(rs.index)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				// stand up an http server just for the duration of this test
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintln(w, string(indexYAMLBytes))
				}))
				defer ts.Close()

				repoSpec := map[string]interface{}{
					"url":      rs.url,
					"interval": "1m0s",
				}
				repoStatus := map[string]interface{}{
					"conditions": []interface{}{
						map[string]interface{}{
							"type":   "Ready",
							"status": "True",
							"reason": "IndexationSucceed",
						},
					},
					"url": ts.URL,
				}
				repos = append(repos, newRepo(rs.name, rs.namespace, repoSpec, repoStatus))
			}

			s, mock, _, err := newServerWithWatcherAndReadyRepos(repos...)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			if err = beforeCallGetAvailablePackageSummaries(mock, tc.request.FilterOptions, repos...); err != nil {
				t.Fatalf("%v", err)
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%v", err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageSummariesResponse{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{})
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

		s, mock, watcher, err := newServerWithWatcherAndReadyRepos(repo)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = beforeCallGetAvailablePackageSummaries(mock, nil, repo); err != nil {
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

		expectedPackagesBeforeUpdate := []*corev1.AvailablePackageSummary{
			{
				DisplayName:      "alpine",
				LatestPkgVersion: "0.2.0",
				IconUrl:          "",
				ShortDescription: "Deploy a basic Alpine Linux pod",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "testrepo/alpine",
					Context:    &corev1.Context{Namespace: "ns2"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			},
			{
				DisplayName:      "nginx",
				LatestPkgVersion: "1.1.0",
				IconUrl:          "",
				ShortDescription: "Create a basic nginx HTTP server",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "testrepo/nginx",
					Context:    &corev1.Context{Namespace: "ns2"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			}}

		opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeUpdate.AvailablePackagesSummaries, expectedPackagesBeforeUpdate; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		updateHappened = true
		// now we are going to simulate flux seeing an update of the index.yaml and modifying the
		// HelmRepository CRD which, in turn, causes k8s server to fire a MODIFY event
		s.cache.eventProcessingWaitGroup.Add(1)

		key, bytes, err := redisKeyValueForRuntimeObject(repo)
		if err != nil {
			t.Fatalf("%v", err)
		}
		mock.ExpectSet(key, bytes, 0).SetVal("")

		watcher.Modify(repo)

		s.cache.eventProcessingWaitGroup.Wait()

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		mock.ExpectScan(0, "", 0).SetVal([]string{key}, 0)
		mock.ExpectGet(key).SetVal(string(bytes))

		responsePackagesAfterUpdate, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		expectedPackagesAfterUpdate := []*corev1.AvailablePackageSummary{
			{
				DisplayName:      "alpine",
				LatestPkgVersion: "0.3.0",
				IconUrl:          "",
				ShortDescription: "Deploy a basic Alpine Linux pod",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "testrepo/alpine",
					Context:    &corev1.Context{Namespace: "ns2"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			},
			{
				DisplayName:      "nginx",
				LatestPkgVersion: "1.1.0",
				IconUrl:          "",
				ShortDescription: "Create a basic nginx HTTP server",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "testrepo/nginx",
					Context:    &corev1.Context{Namespace: "ns2"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			}}

		if got, want := responsePackagesAfterUpdate.AvailablePackagesSummaries, expectedPackagesAfterUpdate; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

func TestGetAvailablePackageSummaryAfterFluxHelmRepoDelete(t *testing.T) {
	t.Run("test get available package summaries after flux helm repository CRD gets deleted", func(t *testing.T) {
		indexYaml, err := ioutil.ReadFile("testdata/valid-index.yaml")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// stand up an http server just for the duration of this test
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, string(indexYaml))
		}))
		defer ts.Close()

		repoSpec := map[string]interface{}{
			"url":      "https://example.repo.com/charts",
			"interval": "1m0s",
		}

		repoStatus := map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":   "Ready",
					"status": "True",
					"reason": "IndexationSucceed",
				},
			},
			"url": ts.URL,
		}
		repo := newRepo("bitnami-1", "default", repoSpec, repoStatus)

		s, mock, watcher, err := newServerWithWatcherAndReadyRepos(repo)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		if err = beforeCallGetAvailablePackageSummaries(mock, nil, repo); err != nil {
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

		expectedPackagesBeforeDelete := []*corev1.AvailablePackageSummary{
			{
				DisplayName:      "acs-engine-autoscaler",
				LatestPkgVersion: "2.1.1",
				IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
				ShortDescription: "Scales worker nodes within agent pools",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/acs-engine-autoscaler",
					Context:    &corev1.Context{Namespace: "default"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			},
			{
				DisplayName:      "wordpress",
				LatestPkgVersion: "0.7.5",
				IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
				ShortDescription: "new description!",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/wordpress",
					Context:    &corev1.Context{Namespace: "default"},
					Plugin:     &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"},
				},
			},
		}

		opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
		opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)
		if got, want := responseBeforeDelete.AvailablePackagesSummaries, expectedPackagesBeforeDelete; !cmp.Equal(got, want, opt1, opt2) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
		}

		// now we are going to simulate flux seeing an update of the index.yaml and modifying the
		// HelmRepository CRD which, in turn, causes k8s server to fire a MODIFY event
		s.cache.eventProcessingWaitGroup.Add(1)
		key := redisKeyForRuntimeObject(repo)
		mock.ExpectDel(key).SetVal(0)

		watcher.Delete(repo)

		s.cache.eventProcessingWaitGroup.Wait()

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}

		mock.ExpectScan(0, "", 0).SetVal([]string{}, 0)

		responseAfterDelete, err := s.GetAvailablePackageSummaries(
			context.Background(),
			&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})
		if err != nil {
			t.Fatalf("%v", err)
		}

		if len(responseAfterDelete.AvailablePackagesSummaries) != 0 {
			t.Errorf("expected empty array, got: %s", responseAfterDelete)
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
			name: "returns expected repositories in specific namespace",
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
			s, _, mock, err := newServerWithRepos(newRepos(tc.repoSpecs, tc.repoNamespace)...)
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

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		testName              string
		request               *corev1.GetAvailablePackageDetailRequest
		repoName              string
		repoNamespace         string
		chartName             string
		chartTarGz            string
		expectedPackageDetail *corev1.AvailablePackageDetail
	}{
		{
			testName:      "it returns details about the redis package in bitnami repo",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			chartTarGz: "testdata/redis-14.4.0.tgz",
			expectedPackageDetail: &corev1.AvailablePackageDetail{
				// TODO (gfichtenholt) other fields
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "bitnami-1/redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				},
				Name:            "redis",
				LongDescription: "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
			},
		},
		// TODO (gfichtenholt) specific version
		// TODO (gfichtenholt) negative test
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			tarGzBytes, err := ioutil.ReadFile(tc.chartTarGz)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				w.Write(tarGzBytes)
			}))
			defer ts.Close()

			chartSpec := map[string]interface{}{
				"chart": tc.chartName,
				"sourceRef": map[string]interface{}{
					"name": tc.repoName,
					"kind": fluxHelmRepository,
				},
				"interval": "10m",
			}
			chartStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "ChartPullSucceeded",
					},
				},
				"url": ts.URL,
			}
			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)

			s, _, mock, err := newServerWithCharts(chart)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageReference{}, corev1.Context{})
			opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "LongDescription")
			if got, want := response.AvailablePackageDetail, tc.expectedPackageDetail; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
			if !strings.Contains(response.AvailablePackageDetail.LongDescription, tc.expectedPackageDetail.LongDescription) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.LongDescription, response.AvailablePackageDetail.LongDescription)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}

	negativeTestCases := []struct {
		testName      string
		request       *corev1.GetAvailablePackageDetailRequest
		repoName      string
		repoNamespace string
		chartName     string
		statusCode    codes.Code
	}{
		{
			testName:      "it fails if request is missing namespace",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "redis",
				}},
			chartName:  "redis",
			statusCode: codes.InvalidArgument,
		},
		{
			testName:      "it fails if request has invalid identifier",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "redis",
					Context: &corev1.Context{
						Namespace: "default",
					},
				}},
			chartName:  "redis",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			chartSpec := map[string]interface{}{
				"chart": tc.chartName,
				"sourceRef": map[string]interface{}{
					"name": "does-not-matter-for-now",
					"kind": "HelmRepository",
				},
				"interval": "10m",
			}
			chartStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "ChartPullSucceeded",
					},
				},
				"url": "does-not-matter",
			}
			chart := newChart(tc.chartName, tc.repoNamespace, chartSpec, chartStatus)
			s, _, mock, err := newServerWithCharts(chart)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			_, err = s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

//
// utilities
//
func newRepo(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":       name,
		"generation": int64(1),
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

// newRepos takes a map of specs keyed by object name converting them to runtime objects.
func newRepos(specs map[string]map[string]interface{}, namespace string) []runtime.Object {
	repos := []runtime.Object{}
	for name, spec := range specs {
		repo := newRepo(name, namespace, spec, nil)
		repos = append(repos, repo)
	}
	return repos
}

func newChart(name string, namespace string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{
		"name":       name,
		"generation": int64(1),
	}
	if namespace != "" {
		metadata["namespace"] = namespace
	}

	obj := map[string]interface{}{
		"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
		"kind":       fluxHelmChart,
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

func newServer(clientGetter server.KubernetesClientGetter) (*Server, redismock.ClientMock, error) {
	redisCli, mock := redismock.NewClientMock()
	if clientGetter != nil {
		mock.ExpectPing().SetVal("PONG")
	}
	repositoriesGvr := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}
	config := cacheConfig{
		gvr:          repositoriesGvr,
		clientGetter: clientGetter,
		onAdd:        onAddOrModifyRepo,
		onModify:     onAddOrModifyRepo,
		onGet:        onGetRepo,
		onDelete:     onDeleteRepo,
	}
	cache, err := newCacheWithRedisClient(config, redisCli)
	if err != nil {
		return nil, mock, err
	}
	s := &Server{
		clientGetter: clientGetter,
		cache:        cache,
	}
	return s, mock, nil
}

func newServerWithRepos(repos ...runtime.Object) (*Server, *fake.FakeDynamicClient, redismock.ClientMock, error) {
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
		},
		repos...)

	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return nil, dynamicClient, nil
	}

	s, mock, err := newServer(clientGetter)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, dynamicClient, mock, nil
}

func newServerWithWatcher(repos ...runtime.Object) (*Server, redismock.ClientMock, *watch.FakeWatcher, error) {
	s, dynamicClient, mock, err := newServerWithRepos(repos...)
	if err != nil {
		return s, mock, nil, err
	}

	// this is so we can emulate actual k8s server firing events
	// see https://github.com/kubernetes/kubernetes/issues/54075 for explanation
	watcher := watch.NewFake()

	dynamicClient.Fake.PrependWatchReactor(
		"*",
		k8stesting.DefaultWatchReactor(watcher, nil))

	mock.MatchExpectationsInOrder(false)

	return s, mock, watcher, nil
}

func newServerWithWatcherAndReadyRepos(repos ...runtime.Object) (*Server, redismock.ClientMock, *watch.FakeWatcher, error) {
	s, mock, watcher, err := newServerWithWatcher(repos...)
	if err != nil {
		return s, mock, watcher, err
	}

	s.cache.eventProcessingWaitGroup = &sync.WaitGroup{}

	for _, r := range repos {
		s.cache.eventProcessingWaitGroup.Add(1)
		key, bytes, err := redisKeyValueForRuntimeObject(r)
		if err != nil {
			return s, mock, watcher, err
		}
		mock.ExpectSet(key, bytes, 0).SetVal("")

		// fire an ADD event for this repo as k8s server would do
		watcher.Add(r)
	}

	// sanity check
	if err = s.cache.checkInit(); err != nil {
		return s, mock, watcher, err
	}

	// here we wait until all repos have been indexed on the server-side
	s.cache.eventProcessingWaitGroup.Wait()

	if err = mock.ExpectationsWereMet(); err != nil {
		return s, mock, watcher, err
	}
	return s, mock, watcher, nil
}

func newServerWithCharts(charts ...runtime.Object) (*Server, *fake.FakeDynamicClient, redismock.ClientMock, error) {
	dynamicClient := fake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmCharts}: fluxHelmChartList,
		},
		charts...)

	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return nil, dynamicClient, nil
	}

	s, mock, err := newServer(clientGetter)
	if err != nil {
		return nil, nil, nil, err
	}
	return s, dynamicClient, mock, nil
}

func beforeCallGetAvailablePackageSummaries(mock redismock.ClientMock, filterOptions *corev1.FilterOptions, repos ...runtime.Object) error {
	mapVals := make(map[string][]byte)
	keys := []string{}
	for _, r := range repos {
		key, bytes, err := redisKeyValueForRuntimeObject(r)
		if err != nil {
			return err
		}
		keys = append(keys, key)
		mapVals[key] = bytes
	}
	if filterOptions == nil || len(filterOptions.GetRepositories()) == 0 {
		mock.ExpectScan(0, "", 0).SetVal(keys, 0)
		for _, k := range keys {
			mock.ExpectGet(k).SetVal(string(mapVals[k]))
		}
	} else {
		for _, r := range filterOptions.GetRepositories() {
			keys := []string{}
			for k, _ := range mapVals {
				if strings.HasSuffix(k, ":"+r) {
					keys = append(keys, k)
				}
			}
			mock.ExpectScan(0, "helmrepositories:*:"+r, 0).SetVal(keys, 0)
			for _, k := range keys {
				mock.ExpectGet(k).SetVal(string(mapVals[k]))
			}
		}
	}
	return nil
}

func redisKeyValueForRuntimeObject(r runtime.Object) (string, []byte, error) {
	key := redisKeyForRuntimeObject(r)
	bytes, _, err := onAddOrModifyRepo(key, r.(*unstructured.Unstructured).Object)
	if err != nil {
		return "", nil, err
	}
	return key, bytes.([]byte), nil
}

func redisKeyForRuntimeObject(r runtime.Object) string {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return fmt.Sprintf("%s:%s:%s",
		fluxHelmRepositories,
		r.(*unstructured.Unstructured).GetNamespace(),
		r.(*unstructured.Unstructured).GetName())

}

// these are helpers to compare slices ignoring order
func lessAvailablePackageFunc(p1, p2 *corev1.AvailablePackageSummary) bool {
	return p1.DisplayName < p2.DisplayName
}

func lessPackageRepositoryFunc(p1, p2 *v1alpha1.PackageRepository) bool {
	return p1.Name < p2.Name && p1.Namespace < p2.Namespace
}
