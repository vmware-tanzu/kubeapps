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
	"os"
	"strings"
	"testing"

	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type testSpecChartWithFile struct {
	name     string
	revision string
	tgzFile  string
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		testName              string
		request               *corev1.GetAvailablePackageDetailRequest
		chartCacheHit         bool
		basicAuth             bool // also known as "private" or "secure"
		tls                   bool // also known as "private" or "secure"
		expectedPackageDetail *corev1.AvailablePackageDetail
	}{
		{
			testName: "it returns details about the latest redis package in bitnami repo",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
			},
			chartCacheHit:         true,
			expectedPackageDetail: expected_detail_redis_1,
		},
		{
			testName: "it returns details about the redis package with specific version in bitnami repo",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
				PkgVersion:          "14.3.4",
			},
			chartCacheHit:         false,
			expectedPackageDetail: expected_detail_redis_2,
		},
		{
			testName: "it returns details about the latest redis package from a repo with basic auth",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
			},
			chartCacheHit:         true,
			basicAuth:             true,
			expectedPackageDetail: expected_detail_redis_1,
		},
		{
			testName: "it returns details about the latest redis package from a repo with TLS",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
			},
			chartCacheHit:         true,
			basicAuth:             false,
			tls:                   true,
			expectedPackageDetail: expected_detail_redis_1,
		},
		{
			testName: "it returns details about the latest redis package from a repo with TLS and basic auth",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
			},
			chartCacheHit:         true,
			basicAuth:             true,
			tls:                   true,
			expectedPackageDetail: expected_detail_redis_1,
		},
	}

	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	var ca, pub, priv []byte
	var err error
	if ca, err = ioutil.ReadFile("testdata/rootCA.crt"); err != nil {
		t.Fatalf("%+v", err)
	} else if pub, err = ioutil.ReadFile("testdata/crt.pem"); err != nil {
		t.Fatalf("%+v", err)
	} else if priv, err = ioutil.ReadFile("testdata/key.pem"); err != nil {
		t.Fatalf("%+v", err)
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace
			replaceUrls := make(map[string]string)
			charts := []testSpecChartWithUrl{}
			requestChartUrl := ""

			// these will be used later in a few places
			opts := &common.ClientOptions{}
			if tc.basicAuth {
				opts.Username = "foo"
				opts.Password = "bar"
			}
			if tc.tls {
				opts.CaBytes = ca
				opts.CertBytes = pub
				opts.KeyBytes = priv
			}

			for _, s := range redis_charts_spec {
				tarGzBytes, err := ioutil.ReadFile(s.tgzFile)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				// stand up an http server just for the duration of this test
				var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					w.Write(tarGzBytes)
				})
				if tc.basicAuth {
					handler = basicAuth(handler, "foo", "bar", "myrealm")
				}
				var ts *httptest.Server
				if tc.tls {
					// I cheated a bit in this test. Instead of generating my own certificates
					// and keys using openssl tool, which I found time consuming and overly complicated,
					// I just copied the ones being used by helm.sh tool for testing purposes
					// from https://github.com/helm/helm/tree/main/testdata
					// in order to save some time. Should n't affect any functionality of productionn
					// code
					ts = httptest.NewUnstartedServer(handler)
					tlsConf, err := httpclient.NewClientTLS(pub, priv, ca)
					if err != nil {
						t.Fatalf("%v", err)
					}
					ts.TLS = tlsConf
					ts.StartTLS()
				} else {
					ts = httptest.NewServer(handler)
				}
				defer ts.Close()
				replaceUrls[fmt.Sprintf("{{%s}}", s.tgzFile)] = ts.URL
				c := testSpecChartWithUrl{
					chartID:       fmt.Sprintf("%s/%s", repoName, s.name),
					chartRevision: s.revision,
					chartUrl:      ts.URL,
					opts:          opts,
					repoNamespace: repoNamespace,
				}
				if tc.request.PkgVersion == s.revision {
					requestChartUrl = ts.URL
				}
				charts = append(charts, c)
			}

			secretRef := ""
			secretObjs := []runtime.Object{}
			if tc.basicAuth && tc.tls {
				secretRef = "both-credentials"
				if secret, err := newBasicAuthTlsSecret(secretRef, repoNamespace, "foo", "bar", pub, priv, ca); err != nil {
					t.Fatalf("%+v", err)
				} else {
					secretObjs = append(secretObjs, secret)
				}
			} else if tc.basicAuth {
				secretRef = "http-credentials"
				secretObjs = append(secretObjs, newBasicAuthSecret(secretRef, repoNamespace, "foo", "bar"))
			} else if tc.tls {
				secretRef = "https-credentials"
				if secret, err := newTlsSecret(secretRef, repoNamespace, pub, priv, ca); err != nil {
					t.Fatalf("%+v", err)
				} else {
					secretObjs = append(secretObjs, secret)
				}
			}

			ts2, repo, err := newRepoWithIndex(
				"testdata/redis-two-versions.yaml", repoName, repoNamespace, replaceUrls, secretRef)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, _, _, err := newServerWithRepos(t, []runtime.Object{repo}, charts, secretObjs)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			s.redisMockExpectGetFromRepoCache(mock, nil, repo)
			chartVersion := tc.request.PkgVersion
			if chartVersion == "" {
				chartVersion = charts[0].chartRevision
				requestChartUrl = charts[0].chartUrl
			}
			chartCacheKey, err := s.chartCache.KeyFor(
				repoNamespace,
				tc.request.AvailablePackageRef.Identifier,
				chartVersion)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if !tc.chartCacheHit {
				// first a miss (there will be actually two calls to Redis GET based on current code path)
				for i := 0; i < 2; i++ {
					if err = redisMockExpectGetFromChartCache(mock, chartCacheKey, "", nil); err != nil {
						t.Fatalf("%+v", err)
					}
				}
				// followed by a set and a hit
				if err = s.redisMockSetValueForChart(mock, chartCacheKey, requestChartUrl, opts); err != nil {
					t.Fatalf("%+v", err)
				}
			}
			if err = redisMockExpectGetFromChartCache(mock, chartCacheKey, requestChartUrl, opts); err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			compareActualVsExpectedAvailablePackageDetail(t, response.AvailablePackageDetail, tc.expectedPackageDetail)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestTransientHttpFailuresAreRetriedForChartCache(t *testing.T) {
	t.Run("successfully populates chart cache when transient HTTP errors occur", func(t *testing.T) {
		repoName := "bitnami-1"
		repoNamespace := "default"
		replaceUrls := make(map[string]string)
		charts := []testSpecChartWithUrl{}

		for _, s := range redis_charts_spec {
			tarGzBytes, err := ioutil.ReadFile(s.tgzFile)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			// stand up an http server just for the duration of this test
			// this server will simulate a failure on 1st and 2nd request
			failuresAllowed := 2
			var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, nofail := r.URL.Query()["Nofail"]
				if !nofail && failuresAllowed > 0 {
					failuresAllowed--
					w.WriteHeader(503)
					w.Write([]byte(fmt.Sprintf("The server is not ready to handle the request: [%d try left before OK]", failuresAllowed)))
				} else {
					w.WriteHeader(200)
					w.Write(tarGzBytes)
				}
			})
			ts := httptest.NewServer(handler)
			defer ts.Close()
			replaceUrls[fmt.Sprintf("{{%s}}", s.tgzFile)] = ts.URL
			c := testSpecChartWithUrl{
				chartID:       fmt.Sprintf("%s/%s", repoName, s.name),
				chartRevision: s.revision,
				chartUrl:      ts.URL + "/?Nofail=true",
				repoNamespace: repoNamespace,
				numRetries:    2,
			}
			charts = append(charts, c)
		}

		ts2, repo, err := newRepoWithIndex(
			"testdata/redis-two-versions.yaml", repoName, repoNamespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		s, mock, _, _, err := newServerWithRepos(t, []runtime.Object{repo}, charts, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		packageIdentifier := repoName + "/redis"
		chartVersion := charts[0].chartRevision
		requestChartUrl := charts[0].chartUrl

		s.redisMockExpectGetFromRepoCache(mock, nil, repo)
		chartCacheKey, err := s.chartCache.KeyFor(
			repoNamespace,
			packageIdentifier,
			chartVersion)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if err = redisMockExpectGetFromChartCache(mock, chartCacheKey, requestChartUrl, nil); err != nil {
			t.Fatalf("%+v", err)
		}

		response, err := s.GetAvailablePackageDetail(
			context.Background(),
			&corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef(packageIdentifier, repoNamespace),
			})
		if err != nil {
			t.Fatalf("%+v", err)
		}

		compareActualVsExpectedAvailablePackageDetail(t,
			response.AvailablePackageDetail, expected_detail_redis_1)

		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

func TestNegativeGetAvailablePackageDetail(t *testing.T) {
	negativeTestCases := []struct {
		testName   string
		request    *corev1.GetAvailablePackageDetailRequest
		statusCode codes.Code
	}{
		{
			testName: "it fails if request is missing namespace",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "redis",
				}},
			statusCode: codes.InvalidArgument,
		},
		{
			testName: "it fails if request has invalid identifier",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("redis", "default"),
			},
			statusCode: codes.InvalidArgument,
		},
	}

	// I don't need any repos/charts to test these scenarios
	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			s, mock, _, _, err := newServerWithRepos(t, nil, nil, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			if response != nil {
				t.Fatalf("want nil, got %v", response)
			}
		})
	}
}

// TODO: (gfichtenholt) some, if not all, these scenarios probably ought to return
// codes.NotFound instead of codes.Internal. The spec does not specify yet.
func TestNonExistingRepoOrInvalidPkgVersionGetAvailablePackageDetail(t *testing.T) {
	negativeTestCases := []struct {
		testName      string
		request       *corev1.GetAvailablePackageDetailRequest
		repoName      string
		repoNamespace string
		statusCode    codes.Code
	}{
		{
			testName:      "it fails if request has invalid package version",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
				PkgVersion:          "99.99.0",
			},
			statusCode: codes.Internal,
		},
		{
			testName:      "it fails if repo does not exist",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-2/redis", "default"),
			},
			statusCode: codes.NotFound,
		},
		{
			testName:      "it fails if repo does not exist in specified namespace",
			repoName:      "bitnami-1",
			repoNamespace: "non-default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
			},
			statusCode: codes.NotFound,
		},
		{
			testName:      "it fails if request has invalid chart",
			repoName:      "bitnami-1",
			repoNamespace: "default",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("bitnami-1/redis-123", "default"),
			},
			statusCode: codes.NotFound,
		},
	}

	for _, tc := range negativeTestCases {
		t.Run(tc.testName, func(t *testing.T) {
			replaceUrls := make(map[string]string)
			charts := []testSpecChartWithUrl{}
			for _, s := range redis_charts_spec {
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
					chartID:       fmt.Sprintf("%s/%s", tc.repoName, s.name),
					chartRevision: s.revision,
					chartUrl:      ts.URL,
					repoNamespace: tc.repoNamespace,
				}
				charts = append(charts, c)
			}

			ts2, repo, err := newRepoWithIndex(
				"testdata/redis-two-versions.yaml", tc.repoName, tc.repoNamespace, replaceUrls, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, _, _, err := newServerWithRepos(t, []runtime.Object{repo}, charts, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			requestRepoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			requestRepoNamespace := tc.request.AvailablePackageRef.Context.Namespace

			repoExists := requestRepoName == tc.repoName && requestRepoNamespace == tc.repoNamespace
			if repoExists {
				s.redisMockExpectGetFromRepoCache(mock, nil, repo)
				requestChartName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[1]
				chartExists := requestChartName == "redis"
				if chartExists {
					chartCacheKey, err := s.chartCache.KeyFor(
						requestRepoNamespace,
						tc.request.AvailablePackageRef.Identifier,
						tc.request.PkgVersion)
					if err != nil {
						t.Fatalf("%+v", err)
					}
					// on a cache miss (there will be actually two calls to Redis GET based on current code path)
					for i := 0; i < 2; i++ {
						if err = redisMockExpectGetFromChartCache(mock, chartCacheKey, "", nil); err != nil {
							t.Fatalf("%+v", err)
						}
					}
				}
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err == nil {
				t.Fatalf("got nil, want error")
			}
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}

			if response != nil {
				t.Fatalf("want nil, got %v", response)
			}
		})
	}
}

func TestNegativeGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name:               "it returns invalid argument if called without a package reference",
			request:            nil,
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without namespace",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without an identifier",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, _, _, err := newServerWithRepos(t, nil, nil, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		repoIndex          string
		repoNamespace      string
		repoName           string
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name:          "it returns the package version summary for redis chart in bitnami repo",
			repoIndex:     "testdata/redis-many-versions.yaml",
			repoNamespace: "kubeapps",
			repoName:      "bitnami",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: availableRef("bitnami/redis", "kubeapps"),
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   expected_versions_redis,
		},
		{
			name:          "it returns error for non-existent chart",
			repoIndex:     "testdata/redis-many-versions.yaml",
			repoNamespace: "kubeapps",
			repoName:      "bitnami",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: availableRef("bitnami/redis-123", "kubeapps"),
			},
			expectedStatusCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace
			replaceUrls := make(map[string]string)
			charts := []testSpecChartWithUrl{}
			for _, s := range redis_charts_spec {
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
			ts, repo, err := newRepoWithIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, replaceUrls, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts.Close()

			s, mock, _, _, err := newServerWithRepos(t, []runtime.Object{repo}, charts, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			key, bytes, err := s.redisKeyValueForRepo(repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			mock.ExpectGet(key).SetVal(string(bytes))

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.GetAvailablePackageVersionsResponse{},
				corev1.PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

// test that causes RetryWatcher to stop and the chart cache needs to resync
// this test is focused on the chart cache work queue
func TestChartCacheResyncNotIdle(t *testing.T) {
	t.Run("test that causes RetryWatcher to stop and the chart cache needs to resync", func(t *testing.T) {
		// start with an empty server that only has an empty repo cache
		// passing in []testSpecChartWithUrl{} instead of nil will add support for chart cache
		s, mock, dyncli, watcher, err := newServerWithRepos(t, nil, []testSpecChartWithUrl{}, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		// what I need is a single repo with a whole bunch of unique charts (packages)
		tarGzBytes, err := ioutil.ReadFile("./testdata/redis-14.4.0.tgz")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		// stand up an http server just for the duration of this test
		var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write(tarGzBytes)
		})
		ts := httptest.NewServer(handler)
		defer ts.Close()

		const NUM_CHARTS = 20
		// create a YAML index file that contains this many unique packages
		tmpFile, err := ioutil.TempFile(os.TempDir(), "*.yaml")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer os.Remove(tmpFile.Name())

		templateYAMLBytes, err := ioutil.ReadFile("testdata/single-package-template.yaml")
		if err != nil {
			t.Fatalf("%+v", err)
		}

		text := []byte("apiVersion: v1\nentries:\n")
		if _, err = tmpFile.Write(text); err != nil {
			t.Fatalf("%+v", err)
		}
		for i := 0; i < NUM_CHARTS; i++ {
			s := strings.ReplaceAll(string(templateYAMLBytes), "{{NUM}}", fmt.Sprintf("%d", i))
			if _, err = tmpFile.Write([]byte(s)); err != nil {
				t.Fatalf("%+v", err)
			}
		}

		repoName := "multitude-of-charts"
		repoNamespace := "default"
		replaceUrls := make(map[string]string)
		replaceUrls["{{testdata/redis-14.4.0.tgz}}"] = ts.URL
		ts2, r, err := newRepoWithIndex(
			tmpFile.Name(), repoName, repoNamespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		if _, err = dyncli.Resource(repositoriesGvr).Namespace(repoNamespace).
			Create(context.Background(), r, metav1.CreateOptions{}); err != nil {
			t.Fatalf("%v", err)
		}

		repoKey, repoBytes, err := s.redisKeyValueForRepo(r)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		mock.ExpectGet(repoKey).RedisNil()
		redisMockSetValueForRepo(mock, repoKey, repoBytes)

		opts := &common.ClientOptions{}

		chartCacheKeys := []string{}
		var chartBytes []byte
		for i := 0; i < NUM_CHARTS; i++ {
			chartID := fmt.Sprintf("%s/redis-%d", repoName, i)
			chartVersion := "14.4.0"
			chartCacheKey, err := s.chartCache.KeyFor(repoNamespace, chartID, chartVersion)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if i == 0 {
				chartBytes, err = cache.ChartCacheComputeValue(chartID, ts.URL, chartVersion, opts)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}
			redisMockSetValueForChart(mock, chartCacheKey, chartBytes)
			s.chartCache.ExpectAdd(chartCacheKey)
			chartCacheKeys = append(chartCacheKeys, chartCacheKey)
		}
		s.repoCache.ExpectAdd(repoKey)

		watcher.Add(r)

		done := make(chan int, 1)

		go func() {
			// wait until the first of the added repos have been fully processed and
			// just one of the charts has been sync'ed
			s.repoCache.WaitUntilForgotten(repoKey)
			s.chartCache.WaitUntilForgotten(chartCacheKeys[0])

			// pretty delicate dance between the server and the client below using
			// bi-directional channels in order to make sure the right expectations
			// are set at the right time.
			repoResyncCh, err := s.repoCache.ExpectResync()
			if err != nil {
				t.Errorf("%v", err)
			}

			chartResyncCh, err := s.chartCache.ExpectResync()
			if err != nil {
				t.Errorf("%v", err)
			}

			// now we will simulate a HTTP 410 Gone error in the watcher
			watcher.Error(&errors.NewGone("test HTTP 410 Gone").ErrStatus)
			// we need to wait until server can guarantee no more Redis SETs after
			// this until resync() kicks in
			len := <-repoResyncCh
			if len != 0 {
				t.Errorf("ERROR: Expected empty repo work queue!")
			} else {
				mock.ExpectFlushDB().SetVal("OK")
				redisMockSetValueForRepo(mock, repoKey, repoBytes)
				// now we can signal to the server it's ok to proceed
				repoResyncCh <- 0

				len = <-chartResyncCh
				if len == 0 {
					t.Errorf("ERROR: Expected non-empty chart work queue!")
				} else {
					for i := 0; i < NUM_CHARTS; i++ {
						redisMockSetValueForChart(mock, chartCacheKeys[i], chartBytes)
						s.chartCache.ExpectAdd(chartCacheKeys[i])
					}
					// now we can signal to the server it's ok to proceed
					chartResyncCh <- 0
					s.repoCache.WaitUntilResyncComplete()
					s.chartCache.WaitUntilResyncComplete()
					for i := 0; i < NUM_CHARTS; i++ {
						s.chartCache.WaitUntilForgotten(chartCacheKeys[i])
					}
					// we do ClearExpect() here to avoid things like
					// "there is a remaining expectation which was not matched:
					// [exists helmcharts:default:multitude-of-charts/redis-2:14.4.0]"
					// which might happened because the for loop in the main goroutine may have done a GET
					// right before resync() kicked in. We don't care about that
					mock.ClearExpect()
				}
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
	})
}

func newChart(name, namespace string, spec, status map[string]interface{}) *unstructured.Unstructured {
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

func availableRef(id, namespace string) *corev1.AvailablePackageReference {
	return &corev1.AvailablePackageReference{
		Identifier: id,
		Context: &corev1.Context{
			Namespace: namespace,
			Cluster:   KubeappsCluster,
		},
		Plugin: fluxPlugin,
	}
}

func (s *Server) redisMockSetValueForChart(mock redismock.ClientMock, key, url string, opts *common.ClientOptions) error {
	cs := repoEventSink{
		clientGetter: s.clientGetter,
		chartCache:   s.chartCache,
	}
	return cs.redisMockSetValueForChart(mock, key, url, opts)
}

func (cs *repoEventSink) redisMockSetValueForChart(mock redismock.ClientMock, key, url string, opts *common.ClientOptions) error {
	_, chartID, version, err := fromRedisKeyForChart(key)
	if err != nil {
		return err
	}
	byteArray, err := cache.ChartCacheComputeValue(chartID, url, version, opts)
	if err != nil {
		return fmt.Errorf("chartCacheComputeValue failed due to: %+v", err)
	}
	redisMockSetValueForChart(mock, key, byteArray)
	return nil
}

func redisMockSetValueForChart(mock redismock.ClientMock, key string, byteArray []byte) {
	mock.ExpectExists(key).SetVal(0)
	mock.ExpectSet(key, byteArray, 0).SetVal("OK")
	mock.ExpectInfo("memory").SetVal("used_memory_rss_human:NA\r\nmaxmemory_human:NA")
}

// does a series of mock.ExpectGet(...)
func redisMockExpectGetFromChartCache(mock redismock.ClientMock, key, url string, opts *common.ClientOptions) error {
	if url != "" {
		_, chartID, version, err := fromRedisKeyForChart(key)
		if err != nil {
			return err
		}
		bytes, err := cache.ChartCacheComputeValue(chartID, url, version, opts)
		if err != nil {
			return err
		}
		mock.ExpectGet(key).SetVal(string(bytes))
	} else {
		mock.ExpectGet(key).RedisNil()
	}
	return nil
}

func redisMockExpectDeleteRepoWithCharts(mock redismock.ClientMock, name types.NamespacedName, charts []string) error {
	key, err := redisKeyForRepoNamespacedName(name)
	if err != nil {
		return err
	} else {
		mock.ExpectDel(key).SetVal(0)
	}

	// each element of charts[] array looks like "chartName:chartVersion"
	keys := []string{}
	for _, c := range charts {
		keys = append(keys, fmt.Sprintf("helmcharts:%s:%s/%s", name.Namespace, name.Name, c))
	}

	scanWildcard := fmt.Sprintf("helmcharts:%s:%s/*:*", name.Namespace, name.Name)
	mock.ExpectScan(0, scanWildcard, 0).SetVal(keys, 0)
	for _, k := range keys {
		mock.ExpectDel(k).SetVal(0)
	}
	return nil
}

func fromRedisKeyForChart(key string) (namespace, chartID, chartVersion string, err error) {
	parts := strings.Split(key, ":")
	if len(parts) != 4 || parts[0] != "helmcharts" || len(parts[1]) == 0 || len(parts[2]) == 0 || len(parts[3]) == 0 {
		return "", "", "", status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return parts[1], parts[2], parts[3], nil
}

func compareActualVsExpectedAvailablePackageDetail(t *testing.T, actual *corev1.AvailablePackageDetail, expected *corev1.AvailablePackageDetail) {
	opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageReference{}, corev1.Context{}, corev1.Maintainer{}, plugins.Plugin{}, corev1.PackageAppVersion{})
	// these few fields a bit special in that they are all very long strings,
	// so we'll do a 'Contains' check for these instead of 'Equals'
	opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "Readme", "DefaultValues", "ValuesSchema")
	if got, want := actual, expected; !cmp.Equal(got, want, opt1, opt2) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
	}
	if !strings.Contains(actual.Readme, expected.Readme) {
		t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", expected.Readme, actual.Readme)
	}
	if !strings.Contains(actual.DefaultValues, expected.DefaultValues) {
		t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", expected.DefaultValues, actual.DefaultValues)
	}
	if !strings.Contains(actual.ValuesSchema, expected.ValuesSchema) {
		t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", expected.ValuesSchema, actual.ValuesSchema)
	}
}

// global vars

var redis_charts_spec = []testSpecChartWithFile{
	{
		name:     "redis",
		tgzFile:  "testdata/redis-14.4.0.tgz",
		revision: "14.4.0",
	},
	{
		name:     "redis",
		tgzFile:  "testdata/redis-14.3.4.tgz",
		revision: "14.3.4",
	},
}

var expected_detail_redis_1 = &corev1.AvailablePackageDetail{
	AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
	Name:                "redis",
	Version: &corev1.PackageAppVersion{
		PkgVersion: "14.4.0",
		AppVersion: "6.2.4",
	},
	RepoUrl:          "https://example.repo.com/charts",
	HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
	IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
	DisplayName:      "redis",
	Categories:       []string{"Database"},
	ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
	Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
	DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
	ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
	SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
	Maintainers: []*corev1.Maintainer{
		{
			Name:  "Bitnami",
			Email: "containers@bitnami.com",
		},
		{
			Name:  "desaintmartin",
			Email: "cedric@desaintmartin.fr",
		},
	},
}

var expected_detail_redis_2 = &corev1.AvailablePackageDetail{
	AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
	Name:                "redis",
	Version: &corev1.PackageAppVersion{
		PkgVersion: "14.3.4",
		AppVersion: "6.2.4",
	},
	RepoUrl:          "https://example.repo.com/charts",
	IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
	HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
	DisplayName:      "redis",
	Categories:       []string{"Database"},
	ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
	Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
	DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
	ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
	SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
	Maintainers: []*corev1.Maintainer{
		{
			Name:  "Bitnami",
			Email: "containers@bitnami.com",
		},
		{
			Name:  "desaintmartin",
			Email: "cedric@desaintmartin.fr",
		},
	},
}

var expected_versions_redis = &corev1.GetAvailablePackageVersionsResponse{
	PackageAppVersions: []*corev1.PackageAppVersion{
		{PkgVersion: "14.4.0", AppVersion: "6.2.4"},
		{PkgVersion: "14.3.4", AppVersion: "6.2.4"},
		{PkgVersion: "14.3.3", AppVersion: "6.2.4"},
		{PkgVersion: "14.3.2", AppVersion: "6.2.3"},
		{PkgVersion: "14.2.1", AppVersion: "6.2.3"},
		{PkgVersion: "14.2.0", AppVersion: "6.2.3"},
		{PkgVersion: "13.0.1", AppVersion: "6.2.1"},
		{PkgVersion: "13.0.0", AppVersion: "6.2.1"},
		{PkgVersion: "12.10.1", AppVersion: "6.0.12"},
		{PkgVersion: "12.10.0", AppVersion: "6.0.12"},
		{PkgVersion: "12.9.2", AppVersion: "6.0.12"},
		{PkgVersion: "12.9.1", AppVersion: "6.0.12"},
		{PkgVersion: "12.9.0", AppVersion: "6.0.12"},
		{PkgVersion: "12.8.3", AppVersion: "6.0.12"},
		{PkgVersion: "12.8.2", AppVersion: "6.0.12"},
		{PkgVersion: "12.8.1", AppVersion: "6.0.12"},
	},
}
