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
	"sync"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func TestBadClientGetter(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter clientGetter
		statusCode   codes.Code
	}{
		{
			name:         "returns failed-precondition error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.FailedPrecondition,
		},
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, mock, err := newServer(tc.clientGetter, nil)
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
			s, mock, _, err := newServerWithRepos(tc.repo)
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
					if len(response.AvailablePackageSummaries) != 0 {
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

//
// utilities
//

// This func does not create a kubernetes dynamic client. It is meant to work in conjunction with
// a call to fake.NewSimpleDynamicClientWithCustomListKinds. The reason for argument repos
// (unlike charts or releases) is that repos are treated special because
// a new instance of a Server object is only returned once the cache has been synced with indexed repos
func newServer(clientGetter clientGetter, actionConfig *action.Configuration, repos ...runtime.Object) (*Server, redismock.ClientMock, error) {
	redisCli, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(false)

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

	eventProcessingWaitGroup := &sync.WaitGroup{}
	for _, r := range repos {
		eventProcessingWaitGroup.Add(1)
		if isRepoReady(r.(*unstructured.Unstructured).Object) {
			key, bytes, err := redisKeyValueForRuntimeObject(r)
			if err != nil {
				continue
			}
			mock.ExpectSet(key, bytes, 0).SetVal("")
		}
	}

	cache, err := newCacheWithRedisClient(config, redisCli, eventProcessingWaitGroup)
	if err != nil {
		return nil, mock, err
	}

	eventProcessingWaitGroup.Wait()

	if err := mock.ExpectationsWereMet(); err != nil {
		return nil, mock, err
	}

	s := &Server{
		clientGetter: clientGetter,
		actionConfigGetter: func(context.Context, string) (*action.Configuration, error) {
			return actionConfig, nil
		},
		cache: cache,
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
