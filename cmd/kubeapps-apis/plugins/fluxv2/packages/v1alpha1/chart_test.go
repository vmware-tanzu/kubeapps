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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
		/*
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
		*/
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			repoName := strings.Split(tc.request.AvailablePackageRef.Identifier, "/")[0]
			repoNamespace := tc.request.AvailablePackageRef.Context.Namespace
			replaceUrls := make(map[string]string)
			charts := []testSpecChartWithUrl{}
			requestChartUrl := ""
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
					ts = httptest.NewTLSServer(handler)
				} else {
					ts = httptest.NewServer(handler)
				}
				defer ts.Close()
				replaceUrls[fmt.Sprintf("{{%s}}", s.tgzFile)] = ts.URL
				opts := []getter.Option{getter.WithURL(ts.URL)}
				if tc.basicAuth {
					opts = append(opts, getter.WithBasicAuth("foo", "bar"))
				}
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
			// TODO (gfichtenholt) in theory we could have both TLS AND basic auth
			if tc.basicAuth {
				secretRef = "http-credentials"
				secretObjs = append(secretObjs, newBasicAuthSecret(secretRef, repoNamespace, "foo", "bar"))
			} else if tc.tls {
				secretRef = "https-credentials"
				secretObjs = append(secretObjs, newTlsSecret(secretRef, repoNamespace))
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
			version := tc.request.PkgVersion
			if version == "" {
				version = charts[0].chartRevision
				requestChartUrl = charts[0].chartUrl
			}
			chartCacheKey, err := s.chartCache.keyFor(
				repoNamespace,
				tc.request.AvailablePackageRef.Identifier,
				version)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			opts := []getter.Option{getter.WithURL(requestChartUrl)}
			if tc.basicAuth {
				opts = append(opts, getter.WithBasicAuth("foo", "bar"))
			}
			if !tc.chartCacheHit {
				// first a miss
				if err = redisMockExpectGetFromChartCache(mock, chartCacheKey, "", nil); err != nil {
					t.Fatalf("%+v", err)
				}
				// followed by a set and a hit
				if err = redisMockSetValueForChart(mock, chartCacheKey, requestChartUrl, opts); err != nil {
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

			opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageReference{}, corev1.Context{}, corev1.Maintainer{}, plugins.Plugin{}, corev1.PackageAppVersion{})
			// these few fields a bit special in that they are all very long strings,
			// so we'll do a 'Contains' check for these instead of 'Equals'
			opt2 := cmpopts.IgnoreFields(corev1.AvailablePackageDetail{}, "Readme", "DefaultValues", "ValuesSchema")
			if got, want := response.AvailablePackageDetail, tc.expectedPackageDetail; !cmp.Equal(got, want, opt1, opt2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1, opt2))
			}
			if !strings.Contains(response.AvailablePackageDetail.Readme, tc.expectedPackageDetail.Readme) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.Readme, response.AvailablePackageDetail.Readme)
			}
			if !strings.Contains(response.AvailablePackageDetail.DefaultValues, tc.expectedPackageDetail.DefaultValues) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.DefaultValues, response.AvailablePackageDetail.DefaultValues)
			}
			if !strings.Contains(response.AvailablePackageDetail.ValuesSchema, tc.expectedPackageDetail.ValuesSchema) {
				t.Errorf("substring mismatch (-want: %s\n+got: %s):\n", tc.expectedPackageDetail.ValuesSchema, response.AvailablePackageDetail.ValuesSchema)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
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
					chartCacheKey, err := s.chartCache.keyFor(
						requestRepoNamespace,
						tc.request.AvailablePackageRef.Identifier,
						tc.request.PkgVersion)
					if err != nil {
						t.Fatalf("%+v", err)
					}
					redisMockExpectGetFromChartCache(mock, chartCacheKey, "", nil)
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

func redisMockSetValueForChart(mock redismock.ClientMock, key, url string, opts []getter.Option) error {
	_, chartID, version, err := fromRedisKeyForChart(key)
	if err != nil {
		return err
	}
	byteArray, err := chartCacheComputeValue(chartID, url, version, opts)
	if err != nil {
		return fmt.Errorf("chartCacheComputeValue failed due to: %+v", err)
	}
	mock.ExpectSet(key, byteArray, 0).SetVal("OK")
	mock.ExpectInfo("memory").SetVal("used_memory_rss_human:NA\r\nmaxmemory_human:NA")
	return nil
}

// does a series of mock.ExpectGet(...)
func redisMockExpectGetFromChartCache(mock redismock.ClientMock, key, url string, opts []getter.Option) error {
	if url != "" {
		_, chartID, version, err := fromRedisKeyForChart(key)
		if err != nil {
			return err
		}
		bytes, err := chartCacheComputeValue(chartID, url, version, opts)
		if err != nil {
			return err
		}
		mock.ExpectGet(key).SetVal(string(bytes))
	} else {
		mock.ExpectGet(key).RedisNil()
	}
	return nil
}

func redisMockExpectDeleteFromChartCache(mock redismock.ClientMock) error {
	// TODO (gfichtenholt)
	// everything hardcoded for one test for now :-)
	// will clean it up when I am done with more important stuff
	keys := []string{
		"helmcharts:default:bitnami-1/acs-engine-autoscaler:2.1.1",
		"helmcharts:default:bitnami-1/wordpress:0.7.5",
	}
	mock.ExpectScan(0, "helmcharts:default:bitnami-1/*:*", 0).SetVal(keys, 0)
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
