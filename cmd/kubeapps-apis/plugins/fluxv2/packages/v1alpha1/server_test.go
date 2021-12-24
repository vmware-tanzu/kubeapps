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
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error) {
				return nil, nil, nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			_, _, _, err := s.GetClients(context.Background())
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
			s, mock, _, _, err := newServerWithRepos(t, []runtime.Object{tc.repo}, nil, nil)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			// these (negative) tests are all testing very unlikely scenarios which were kind of hard to fit into
			// redisMockBeforeCallToGetAvailablePackageSummaries() so I put special logic here instead
			if isRepoReady(tc.repo.(*unstructured.Unstructured).Object) {
				if key, err := redisKeyForRepo(tc.repo); err == nil {
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
type testSpecChartWithUrl struct {
	chartID       string
	chartRevision string
	chartUrl      string
	opts          *common.ClientOptions
	repoNamespace string
	// this is for a negative test TestTransientHttpFailuresAreRetriedForChartCache
	numRetries int
}

// This func does not create a kubernetes dynamic client. It is meant to work in conjunction with
// a call to fake.NewSimpleDynamicClientWithCustomListKinds. The reason for argument repos
// (unlike charts or releases) is that repos are treated special because
// a new instance of a Server object is only returned once the cache has been synced with indexed repos
func newServer(t *testing.T, clientGetter common.ClientGetterFunc, actionConfig *action.Configuration, repos []runtime.Object, charts []testSpecChartWithUrl) (*Server, redismock.ClientMock, error) {
	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	redisCli, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(false)

	if clientGetter != nil {
		// if client getter returns an error, FLUSHDB call does not take place, because
		// newCacheWithRedisClient() raises an error before redisCli.FlushDB() call
		if _, _, _, err := clientGetter(context.TODO()); err == nil {
			mock.ExpectFlushDB().SetVal("OK")
		}
	}

	cs := repoCacheCallSite{
		clientGetter: clientGetter,
		chartCache:   nil,
	}

	okRepos := sets.String{}
	for _, r := range repos {
		if isRepoReady(r.(*unstructured.Unstructured).Object) {
			// we are willfully ignoring any errors coming from redisMockSetValueForRepo()
			// here and just skipping over to next repo. This is done for test
			// TestGetAvailablePackagesStatus where we make sure that even if the flux CRD happens
			// to be invalid flux plug in can still operate
			key, _, err := cs.redisMockSetValueForRepo(mock, r)
			if err != nil {
				t.Logf("Skipping repo [%s] due to %+v", key, err)
			} else {
				okRepos.Insert(key)
			}
		}
	}

	var chartCache *cache.ChartCache
	var err error
	cachedChartKeys := sets.String{}
	cachedChartIds := sets.String{}

	if charts != nil {
		chartCache, err = cache.NewChartCache("chartCacheTest", redisCli, stopCh)
		if err != nil {
			return nil, mock, err
		}
		t.Cleanup(func() { chartCache.Shutdown() })

		// for now we only cache latest version of each chart
		for _, c := range charts {
			// very simple logic for now, relies on the order of elements in the array
			// to pick the latest version.
			if !cachedChartIds.Has(c.chartID) {
				cachedChartIds.Insert(c.chartID)
				key, err := chartCache.KeyFor(c.repoNamespace, c.chartID, c.chartRevision)
				if err != nil {
					return nil, mock, err
				}
				repoName := types.NamespacedName{
					Name:      strings.Split(c.chartID, "/")[0],
					Namespace: c.repoNamespace}

				repoKey, err := redisKeyForRepoNamespacedName(repoName)
				if err == nil && okRepos.Has(repoKey) {
					for i := 0; i < c.numRetries+1; i++ {
						mock.ExpectExists(key).SetVal(0)
					}
					err = redisMockSetValueForChart(mock, key, c.chartUrl, c.opts)
					if err != nil {
						return nil, mock, err
					}
					cachedChartKeys.Insert(key)
					chartCache.ExpectAdd(key)
				}
			}
		}
		cs = repoCacheCallSite{
			clientGetter: clientGetter,
			chartCache:   chartCache,
		}
	}

	cacheConfig := cache.NamespacedResourceWatcherCacheConfig{
		Gvr:          repositoriesGvr,
		ClientGetter: clientGetter,
		OnAddFunc:    cs.onAddRepo,
		OnModifyFunc: cs.onModifyRepo,
		OnGetFunc:    cs.onGetRepo,
		OnDeleteFunc: cs.onDeleteRepo,
		OnResyncFunc: cs.onResync,
	}

	repoCache, err := cache.NewNamespacedResourceWatcherCache("repoCacheTest", cacheConfig, redisCli, stopCh)
	if err != nil {
		return nil, mock, err
	}

	// need to wait until ChartCache has finished syncing
	for key := range cachedChartKeys {
		chartCache.WaitUntilForgotten(key)
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

// ref: https://stackoverflow.com/questions/21936332/idiomatic-way-of-requiring-http-basic-auth-in-go
func basicAuth(handler http.HandlerFunc, username, password, realm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}
		handler(w, r)
	}
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
