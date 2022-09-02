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
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ca, pub, priv := getCertsForTesting(t)

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace
			replaceUrls := make(map[string]string)
			charts := []testSpecChartWithUrl{}
			requestChartUrl := ""

			// these will be used later in a few places
			opts := &common.HttpClientOptions{}
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
				tarGzBytes, err := os.ReadFile(s.tgzFile)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				// stand up an http server just for the duration of this test
				var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(200)
					_, err = w.Write(tarGzBytes)
					if err != nil {
						t.Fatalf("%+v", err)
					}
				})
				if tc.basicAuth {
					handler = basicAuth(handler, "foo", "bar", "myrealm")
				}
				var ts *httptest.Server
				if tc.tls {
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
				secret := newBasicAuthTlsSecret(types.NamespacedName{
					Name:      secretRef,
					Namespace: repoNamespace}, "foo", "bar", pub, priv, ca)
				secretObjs = append(secretObjs, secret)
			} else if tc.basicAuth {
				secretRef = "http-credentials"
				secretObjs = append(secretObjs, newBasicAuthSecret(
					types.NamespacedName{
						Name:      secretRef,
						Namespace: repoNamespace}, "foo", "bar"))
			} else if tc.tls {
				secretRef = "https-credentials"
				secret := newTlsSecret(types.NamespacedName{
					Name:      secretRef,
					Namespace: repoNamespace}, pub, priv, ca)
				secretObjs = append(secretObjs, secret)
			}

			ts2, repo, err := newHttpRepoAndServeIndex(
				testYaml("redis-two-versions.yaml"), repoName, repoNamespace, replaceUrls, secretRef)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, charts, secretObjs)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}
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
					if err = redisMockExpectGetFromHttpChartCache(mock, chartCacheKey, "", nil); err != nil {
						t.Fatalf("%+v", err)
					}
				}
				// followed by a set and a hit
				if err = redisMockSetValueForHttpChart(mock, chartCacheKey, requestChartUrl, opts); err != nil {
					t.Fatalf("%+v", err)
				}
			}
			if err = redisMockExpectGetFromHttpChartCache(mock, chartCacheKey, requestChartUrl, opts); err != nil {
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
			tarGzBytes, err := os.ReadFile(s.tgzFile)
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
					_, err = w.Write([]byte(fmt.Sprintf("The server is not ready to handle the request: [%d try left before OK]", failuresAllowed)))
					if err != nil {
						t.Fatalf("%+v", err)
					}
				} else {
					w.WriteHeader(200)
					_, err := w.Write(tarGzBytes)
					if err != nil {
						t.Fatalf("%+v", err)
					}
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

		ts2, repo, err := newHttpRepoAndServeIndex(
			testYaml("redis-two-versions.yaml"), repoName, repoNamespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, charts, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		packageIdentifier := repoName + "/redis"
		chartVersion := charts[0].chartRevision
		requestChartUrl := charts[0].chartUrl

		err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		chartCacheKey, err := s.chartCache.KeyFor(
			repoNamespace,
			packageIdentifier,
			chartVersion)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if err = redisMockExpectGetFromHttpChartCache(mock, chartCacheKey, requestChartUrl, nil); err != nil {
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
			s, mock, err := newSimpleServerWithRepos(t, nil)
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
					chartID:       fmt.Sprintf("%s/%s", tc.repoName, s.name),
					chartRevision: s.revision,
					chartUrl:      ts.URL,
					repoNamespace: tc.repoNamespace,
				}
				charts = append(charts, c)
			}

			ts2, repo, err := newHttpRepoAndServeIndex(
				testYaml("redis-two-versions.yaml"), tc.repoName, tc.repoNamespace, replaceUrls, "")
			if err != nil {
				t.Fatalf("%+v", err)
			}
			defer ts2.Close()

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, charts, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			requestRepoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			requestRepoNamespace := tc.request.AvailablePackageRef.Context.Namespace

			repoExists := requestRepoName == tc.repoName && requestRepoNamespace == tc.repoNamespace
			if repoExists {
				err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo)
				if err != nil {
					t.Fatalf("%+v", err)
				}
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
						if err = redisMockExpectGetFromHttpChartCache(mock, chartCacheKey, "", nil); err != nil {
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
			s, mock, err := newSimpleServerWithRepos(t, nil)
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
			repoIndex:     testYaml("redis-many-versions.yaml"),
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
			repoIndex:     testYaml("redis-many-versions.yaml"),
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
					chartID:       fmt.Sprintf("%s/%s", repoName, s.name),
					chartRevision: s.revision,
					chartUrl:      ts.URL,
					repoNamespace: repoNamespace,
				}
				charts = append(charts, c)
			}
			ts, repo, err := newHttpRepoAndServeIndex(tc.repoIndex, tc.repoName, tc.repoNamespace, replaceUrls, "")
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
				t.Fatal(err)
			}

			if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
				t.Fatal(err)
			}

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

func TestGetOciAvailablePackageVersions(t *testing.T) {
	seed_data_2, err := newFakeRemoteOciRegistryData_2()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name               string
		repoNamespace      string
		repoName           string
		repoUrl            string
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
		seedData           *fakeRemoteOciRegistryData
		charts             []testSpecChartWithUrl
	}{
		{
			name:          "it returns the package version summary for podinfo chart in oci repo",
			repoNamespace: "namespace-1",
			repoName:      "repo-1",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: availableRef("repo-1/podinfo", "namespace-1"),
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   expected_versions_podinfo_2,
			seedData:           seed_data_2,
			charts:             oci_podinfo_charts_spec_2,
			repoUrl:            "oci://localhost:54321/userX/charts",
		},
		{
			name:          "it returns error for non-existent chart",
			repoNamespace: "namespace-1",
			repoName:      "repo-1",
			repoUrl:       "oci://localhost:54321/userX/charts",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: availableRef("repo-1/zippity", "namespace-1"),
			},
			expectedStatusCode: codes.Internal,
			seedData:           seed_data_2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace

			repo, err := newOciRepo(repoName, repoNamespace, tc.repoUrl)
			if err != nil {
				t.Fatal(err)
			}

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, tc.charts, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			if err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo); err != nil {
				t.Fatal(err)
			}

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
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
		s, mock, err := newServerWithRepos(t, nil, []testSpecChartWithUrl{}, nil)
		if err != nil {
			t.Fatalf("error instantiating the server: %v", err)
		}

		// what I need is a single repo with a whole bunch of unique charts (packages)
		tarGzBytes, err := os.ReadFile(testTgz("redis-14.4.0.tgz"))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		// stand up an http server just for the duration of this test
		var handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, err = w.Write(tarGzBytes)
			if err != nil {
				t.Fatalf("%+v", err)
			}
		})
		ts := httptest.NewServer(handler)
		defer ts.Close()

		const NUM_CHARTS = 20
		// create a YAML index file that contains this many unique packages
		tmpFile, err := os.CreateTemp(os.TempDir(), "*.yaml")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer os.Remove(tmpFile.Name())

		templateYAMLBytes, err := os.ReadFile(testYaml("single-package-template.yaml"))
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
		replaceUrls["{{./testdata/charts/redis-14.4.0.tgz}}"] = ts.URL
		ts2, r, err := newHttpRepoAndServeIndex(
			tmpFile.Name(), repoName, repoNamespace, replaceUrls, "")
		if err != nil {
			t.Fatalf("%+v", err)
		}
		defer ts2.Close()

		repoKey, repoBytes, err := s.redisKeyValueForRepo(*r)
		if err != nil {
			t.Fatalf("%+v", err)
		} else {
			redisMockSetValueForRepo(mock, repoKey, repoBytes, nil)
		}

		opts := &common.HttpClientOptions{}
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
				fn := downloadHttpChartFn(opts)
				chartBytes, err = cache.ChartCacheComputeValue(chartID, ts.URL, chartVersion, fn)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}
			redisMockSetValueForChart(mock, chartCacheKey, chartBytes)
			s.chartCache.ExpectAdd(chartCacheKey)
			chartCacheKeys = append(chartCacheKeys, chartCacheKey)
		}

		s.repoCache.ExpectAdd(repoKey)

		ctrlClient, watcher, err := ctrlClientAndWatcher(t, s)
		if err != nil {
			t.Fatal(err)
		} else if err = ctrlClient.Create(context.Background(), r); err != nil {
			t.Fatal(err)
		}

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
				redisMockSetValueForRepo(mock, repoKey, repoBytes, nil)
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

		// sanity check
		if err = mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

// ref https://github.com/vmware-tanzu/kubeapps/issues/4381
// [fluxv2] non-FQDN chart url fails on chart view #4381
func TestChartWithRelativeURL(t *testing.T) {
	repoName := "testRepo"
	repoNamespace := "default"

	tarGzBytes, err := os.ReadFile(testTgz("airflow-1.0.0.tgz"))
	if err != nil {
		t.Fatal(err)
	}

	indexYAMLBytes, err := os.ReadFile(testYaml("chart-with-relative-url.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/index.yaml" {
			fmt.Fprintln(w, string(indexYAMLBytes))
		} else if r.RequestURI == "/charts/airflow-1.0.0.tgz" {
			w.WriteHeader(200)
			_, err = w.Write(tarGzBytes)
			if err != nil {
				t.Fatalf("%+v", err)
			}
		} else {
			w.WriteHeader(404)
		}
	}))

	repoSpec := &sourcev1.HelmRepositorySpec{
		URL:      ts.URL,
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
		URL: ts.URL + "/index.yaml",
	}
	repo := newRepo(repoName, repoNamespace, repoSpec, repoStatus)
	defer ts.Close()

	s, mock, err := newServerWithRepos(t,
		[]sourcev1.HelmRepository{repo},
		[]testSpecChartWithUrl{
			{
				chartID:       fmt.Sprintf("%s/airflow", repoName),
				chartRevision: "1.0.0",
				chartUrl:      ts.URL + "/charts/airflow-1.0.0.tgz",
				repoNamespace: repoNamespace,
			},
		}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err = s.redisMockExpectGetFromRepoCache(mock, nil, repo); err != nil {
		t.Fatal(err)
	}

	response, err := s.GetAvailablePackageVersions(
		context.Background(), &corev1.GetAvailablePackageVersionsRequest{
			AvailablePackageRef: availableRef(repoName+"/airflow", repoNamespace),
		})
	if err != nil {
		t.Fatal(err)
	}
	opts := cmpopts.IgnoreUnexported(
		corev1.GetAvailablePackageVersionsResponse{},
		corev1.PackageAppVersion{})
	if got, want := response, expected_versions_airflow; !cmp.Equal(want, got, opts) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetOciAvailablePackageDetail(t *testing.T) {
	seed_data_1, err := newFakeRemoteOciRegistryData_1()
	if err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		testName              string
		request               *corev1.GetAvailablePackageDetailRequest
		chartCacheHit         bool
		expectedPackageDetail *corev1.AvailablePackageDetail
		seedData              *fakeRemoteOciRegistryData
		charts                []testSpecChartWithUrl
		repoUrl               string
	}{
		{
			testName: "it returns details about the latest podinfo package in oci repo",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: availableRef("repo-1/podinfo", "namespace-1"),
			},
			chartCacheHit:         true,
			expectedPackageDetail: expected_detail_podinfo_1,
			seedData:              seed_data_1,
			charts:                oci_podinfo_charts_spec,
			repoUrl:               "oci://localhost:54321/userX/charts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			initOciFakeClientBuilder(t, *tc.seedData)
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace

			repo, err := newOciRepo(repoName, repoNamespace, tc.repoUrl)
			if err != nil {
				t.Fatal(err)
			}

			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{*repo}, tc.charts, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			err = s.redisMockExpectGetFromRepoCache(mock, nil, *repo)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			chartVersion := tc.request.PkgVersion
			if chartVersion == "" {
				chartVersion = tc.charts[0].chartRevision
			}
			chartCacheKey, err := s.chartCache.KeyFor(
				repoNamespace,
				tc.request.AvailablePackageRef.Identifier,
				chartVersion)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			namespacedRepoName := types.NamespacedName{
				Name:      repoName,
				Namespace: repoNamespace}
			ociChartRepo, err := s.newOCIChartRepositoryAndLogin(context.Background(), namespacedRepoName)
			if err != nil {
				t.Fatal(err)
			}
			if !tc.chartCacheHit {
				// first a miss (there will be actually two calls to Redis GET based on current code path)
				for i := 0; i < 2; i++ {
					if err = redisMockExpectGetFromOciChartCache(mock, chartCacheKey, "", nil); err != nil {
						t.Fatal(err)
					}
				}
				// followed by a set and a hit
				err = redisMockSetValueForOciChart(mock, chartCacheKey, tc.charts[0].chartUrl, ociChartRepo)
				if err != nil {
					t.Fatal(err)
				}
			}
			if err = redisMockExpectGetFromOciChartCache(mock, chartCacheKey, tc.charts[0].chartUrl, ociChartRepo); err != nil {
				t.Fatal(err)
			}

			response, err := s.GetAvailablePackageDetail(context.Background(), tc.request)
			if err != nil {
				t.Fatal(err)
			}

			compareActualVsExpectedAvailablePackageDetail(t, response.AvailablePackageDetail, tc.expectedPackageDetail)

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func newChart(name, namespace string, spec *sourcev1.HelmChartSpec, status *sourcev1.HelmChartStatus) sourcev1.HelmChart {
	helmChart := sourcev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Generation: int64(1),
		},
	}
	if namespace != "" {
		helmChart.ObjectMeta.Namespace = namespace
	}

	if spec != nil {
		helmChart.Spec = *spec.DeepCopy()
	}

	if status != nil {
		helmChart.Status = *status.DeepCopy()
		helmChart.Status.ObservedGeneration = int64(1)
	}

	return helmChart
}

func redisMockSetValueForHttpChart(mock redismock.ClientMock, key, url string, opts *common.HttpClientOptions) error {
	_, chartID, version, err := fromRedisKeyForChart(key)
	if err != nil {
		return err
	}
	fn := downloadHttpChartFn(opts)
	byteArray, err := cache.ChartCacheComputeValue(chartID, url, version, fn)
	if err != nil {
		return fmt.Errorf("chartCacheComputeValue failed due to: %+v", err)
	}
	redisMockSetValueForChart(mock, key, byteArray)
	return nil
}

func redisMockSetValueForOciChart(mock redismock.ClientMock, key, url string, ociRepo *OCIChartRepository) error {
	_, chartID, version, err := fromRedisKeyForChart(key)
	if err != nil {
		return err
	}
	fn := downloadOCIChartFn(ociRepo)
	byteArray, err := cache.ChartCacheComputeValue(chartID, url, version, fn)
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
func redisMockExpectGetFromHttpChartCache(mock redismock.ClientMock, key, url string, opts *common.HttpClientOptions) error {
	if url != "" {
		_, chartID, version, err := fromRedisKeyForChart(key)
		if err != nil {
			return err
		}
		fn := downloadHttpChartFn(opts)
		bytes, err := cache.ChartCacheComputeValue(chartID, url, version, fn)
		if err != nil {
			return err
		}
		mock.ExpectGet(key).SetVal(string(bytes))
	} else {
		mock.ExpectGet(key).RedisNil()
	}
	return nil
}

// does a series of mock.ExpectGet(...)
func redisMockExpectGetFromOciChartCache(mock redismock.ClientMock, key, url string, ociRepo *OCIChartRepository) error {
	if url != "" {
		_, chartID, version, err := fromRedisKeyForChart(key)
		if err != nil {
			return err
		}
		fn := downloadOCIChartFn(ociRepo)
		bytes, err := cache.ChartCacheComputeValue(chartID, url, version, fn)
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
