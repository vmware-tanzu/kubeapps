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
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

const KubeappsCluster = "default"

func TestBadClientGetter(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter common.ClientGetterFunc
		statusCode   codes.Code
	}{
		{
			name:         "returns internal error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.Internal,
		},
		{
			name: "returns failed-precondition when clientGetter itself errors",
			clientGetter: func(context.Context) (dynamic.Interface, apiext.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			_, err := s.getDynamicClient(context.Background())
			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

		})
	}
}

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name       string
		repo       runtime.Object
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
			s, mock, _, _, err := newServerWithRepos(tc.repo)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			// these (negative) tests are all testing very unlikely scenarios which were kind of hard to fit into
			// redisMockBeforeCallToGetAvailablePackageSummaries() so I put special logic here instead
			if isRepoReady(tc.repo.(*unstructured.Unstructured).Object) {
				if key, err := redisKeyForRuntimeObject(tc.repo); err == nil {
					// TODO explain why 3 calls to ExpectGet. It has to do with the way
					// the caching internals work: a call such as GetAvailablePackageSummaries
					// will cause multiple fetches
					mock.ExpectGet(key).RedisNil()
					mock.ExpectGet(key).RedisNil()
					mock.ExpectGet(key).RedisNil()
					mock.ExpectDel(key).SetVal(0)
				}
			}

			response, err := s.GetAvailablePackageSummaries(
				context.Background(),
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, error: %v, want: %+v", got, err, want)
			} else if got == codes.OK {
				if response == nil || len(response.AvailablePackageSummaries) != 0 {
					t.Fatalf("unexpected response: %v", response)
				}
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

// This func does not create a kubernetes dynamic client. It is meant to work in conjunction with
// a call to fake.NewSimpleDynamicClientWithCustomListKinds. The reason for argument repos
// (unlike charts or releases) is that repos are treated special because
// a new instance of a Server object is only returned once the cache has been synced with indexed repos
func newServer(clientGetter common.ClientGetterFunc, actionConfig *action.Configuration, repos ...runtime.Object) (*Server, redismock.ClientMock, error) {
	redisCli, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(false)

	if clientGetter != nil {
		// if client getter returns an error, FLUSHDB call does not take place, because
		// newCacheWithRedisClient() raises an error before redisCli.FlushDB() call
		if _, _, err := clientGetter(context.TODO()); err == nil {
			mock.ExpectFlushDB().SetVal("OK")
		}
	}

	chartCache, err := NewChartCache(redisCli)
	if err != nil {
		return nil, mock, err
	}

	config := cache.NamespacedResourceWatcherCacheConfig{
		Gvr:          repositoriesGvr,
		ClientGetter: clientGetter,
		OnAddFunc:    chartCache.wrapOnAddFunc(onAddRepo, onGetRepo),
		OnModifyFunc: onModifyRepo,
		OnGetFunc:    onGetRepo,
		OnDeleteFunc: onDeleteRepo,
	}

	for _, r := range repos {
		if isRepoReady(r.(*unstructured.Unstructured).Object) {
			// we are willfully ignoring any errors coming from redisMockSetValueForRepo()
			// here and just skipping over to next repo
			redisMockSetValueForRepo(r, mock)
		}
	}

	repoCache, err := cache.NewNamespacedResourceWatcherCache(config, redisCli)
	if err != nil {
		return nil, mock, err
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		return nil, mock, err
	}

	s := &Server{
		clientGetter: clientGetter,
		actionConfigGetter: func(context.Context, string) (*action.Configuration, error) {
			return actionConfig, nil
		},
		repoCache:       repoCache,
		chartCache:      chartCache,
		kubeappsCluster: KubeappsCluster,
	}
	return s, mock, nil
}

// these are helpers to compare slices ignoring order
func lessAvailablePackageFunc(p1, p2 *corev1.AvailablePackageSummary) bool {
	return p1.DisplayName < p2.DisplayName
}

func lessPackageRepositoryFunc(p1, p2 *v1alpha1.PackageRepository) bool {
	return p1.Name < p2.Name && p1.Namespace < p2.Namespace
}

// misc global vars that get re-used in multiple tests
var fluxPlugin = &plugins.Plugin{Name: "fluxv2.packages", Version: "v1alpha1"}
var fluxHelmRepositoryCRD = &apiextv1.CustomResourceDefinition{
	TypeMeta: metav1.TypeMeta{
		Kind:       "CustomResourceDefinition",
		APIVersion: "apiextensions.k8s.io/v1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name: "helmrepositories.source.toolkit.fluxcd.io",
	},
	Status: apiextv1.CustomResourceDefinitionStatus{
		Conditions: []apiextv1.CustomResourceDefinitionCondition{
			{
				Type:   "Established",
				Status: "True",
			},
		},
	},
}
