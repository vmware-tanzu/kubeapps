/*
Copyright (c) 2018 Bitnami

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
package chart

import (
	"bytes"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arschles/assert"
	appRepov1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	chartv2 "k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

const testChartArchive = "./testdata/nginx-apiVersion-v1-5.1.1.tgz"

func Test_resolveChartURL(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		chartURL  string
		wantedURL string
	}{
		{
			"absolute url",
			"http://www.google.com",
			"http://charts.example.com/repo/wordpress-0.1.0.tgz",
			"http://charts.example.com/repo/wordpress-0.1.0.tgz",
		},
		{
			"relative, repo url",
			"http://charts.example.com/repo/",
			"wordpress-0.1.0.tgz",
			"http://charts.example.com/repo/wordpress-0.1.0.tgz",
		},
		{
			"relative, repo index url",
			"http://charts.example.com/repo/index.yaml",
			"wordpress-0.1.0.tgz",
			"http://charts.example.com/repo/wordpress-0.1.0.tgz",
		},
		{
			"relative, repo url - no trailing slash",
			"http://charts.example.com/repo",
			"wordpress-0.1.0.tgz",
			"http://charts.example.com/wordpress-0.1.0.tgz",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chartURL, err := resolveChartURL(tt.baseURL, tt.chartURL)
			assert.NoErr(t, err)
			assert.Equal(t, chartURL, tt.wantedURL, "url")
		})
	}
}

func TestFindChartInRepoIndex(t *testing.T) {
	name := "foo"
	version := "v1.0.0"
	chartURL := "wordpress-0.1.0.tgz"
	repoURL := "http://charts.example.com/repo/"
	expectedURL := fmt.Sprintf("%s%s", repoURL, chartURL)

	chartMeta := chartv2.Metadata{Name: name, Version: version}
	chartVersion := repo.ChartVersion{URLs: []string{chartURL}}
	chartVersion.Metadata = &chartMeta
	chartVersions := []*repo.ChartVersion{&chartVersion}
	entries := map[string]repo.ChartVersions{}
	entries[name] = chartVersions
	index := &repo.IndexFile{APIVersion: "v1", Generated: time.Now(), Entries: entries}

	res, err := findChartInRepoIndex(index, repoURL, name, version)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if res != expectedURL {
		t.Errorf("Expecting %s to be resolved as %s", res, expectedURL)
	}
}

func TestParseDetails(t *testing.T) {
	testCases := []struct {
		name     string
		data     string
		expected *Details
		err      bool
	}{
		{
			name: "parses request including app repo resource",
			data: `{
				"appRepositoryResourceName": "my-chart-repo",
				"appRepositoryResourceNamespace": "my-repo-namespace",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar"
	        }`,
			expected: &Details{
				AppRepositoryResourceName:      "my-chart-repo",
				AppRepositoryResourceNamespace: "my-repo-namespace",
				ChartName:                      "test",
				ReleaseName:                    "foo",
				Version:                        "1.0.0",
				Values:                         "foo: bar",
			},
		},
		{
			name: "errors if appRepositoryResourceName is not present",
			data: `{
				"appRepositoryResourceNamespace": "my-repo-namespace",
				"chartName": "test",
				"releaseName": "foo",
				"version": "1.0.0",
				"values": "foo: bar"
			}`,
			err: true,
		},
		{
			name: "errors if appRepositoryResourceName is empty",
			data: `{
				"appRepositoryResourceName": "",
				"appRepositoryResourceNamespace": "my-repo-namespace",
				"chartName": "test",
				"releaseName": "foo",
				"version": "1.0.0",
				"values": "foo: bar"
			}`,
			err: true,
		},
		{
			name: "errors if appRepositoryResourceNamespace is empty",
			data: `{
				"appRepositoryResourceName": "my-repo",
				"appRepositoryResourceNamespace": "",
				"chartName": "test",
				"releaseName": "foo",
				"version": "1.0.0",
				"values": "foo: bar"
			}`,
			err: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := Client{}
			details, err := ch.ParseDetails([]byte(tc.data))

			if tc.err {
				if err == nil {
					t.Fatalf("expected error")
				} else {
					return
				}
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !cmp.Equal(tc.expected, details) {
				t.Errorf(cmp.Diff(tc.expected, details))
			}
		})
	}
}

// fakeLoadChartV2 implements LoadChartV2 interface.
func fakeLoadChartV2(in io.Reader) (*chartv2.Chart, error) {
	return &chartv2.Chart{}, nil
}

func TestParseDetailsForHTTPClient(t *testing.T) {
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	const (
		authHeaderSecretName = "auth-header-secret-name"
		authHeaderSecretData = "really-secret-stuff"
		customCASecretName   = "custom-ca-secret-name"
		customCASecretData   = "some-cert-data"
		appRepoName          = "custom-repo"
		appRepoNamespace     = "my-namespace"
	)

	testCases := []struct {
		name             string
		details          *Details
		appRepoSpec      appRepov1.AppRepositorySpec
		errorExpected    bool
		numCertsExpected int
	}{
		{
			name: "default cert pool without auth",
			details: &Details{
				AppRepositoryResourceName:      appRepoName,
				AppRepositoryResourceNamespace: appRepoNamespace,
			},
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "custom CA added when passed an AppRepository CRD",
			details: &Details{
				AppRepositoryResourceName:      appRepoName,
				AppRepositoryResourceNamespace: appRepoNamespace,
			},
			appRepoSpec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					CustomCA: &appRepov1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{customCASecretName},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			numCertsExpected: len(systemCertPool.Subjects()) + 1,
		},
		{
			name: "errors if secret for custom CA secret cannot be found",
			details: &Details{
				AppRepositoryResourceName:      appRepoName,
				AppRepositoryResourceNamespace: appRepoNamespace,
			},
			appRepoSpec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					CustomCA: &appRepov1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"other-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			errorExpected: true,
		},
		{
			name: "authorization header added when passed an AppRepository CRD",
			details: &Details{
				AppRepositoryResourceName:      appRepoName,
				AppRepositoryResourceNamespace: appRepoNamespace,
			},
			appRepoSpec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					Header: &appRepov1.AppRepositoryAuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{authHeaderSecretName},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "errors if auth secret cannot be found",
			details: &Details{
				AppRepositoryResourceName:      appRepoName,
				AppRepositoryResourceNamespace: appRepoNamespace,
			},
			appRepoSpec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					CustomCA: &appRepov1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"other-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		// The fake k8s client will contain secret for the CA and header respectively.
		secrets := []*corev1.Secret{&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      customCASecretName,
				Namespace: appRepoNamespace,
			},
			Data: map[string][]byte{
				"custom-secret-key": []byte(customCASecretName),
			},
		}, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      authHeaderSecretName,
				Namespace: appRepoNamespace,
			},
			Data: map[string][]byte{
				"custom-secret-key": []byte(authHeaderSecretData),
			},
		}}

		apprepos := []*appRepov1.AppRepository{&appRepov1.AppRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      tc.details.AppRepositoryResourceName,
				Namespace: appRepoNamespace,
			},
			Spec: tc.appRepoSpec,
		}}

		chUtils := Client{
			appRepoHandler:    &kube.FakeHandler{Secrets: secrets, AppRepos: apprepos},
			kubeappsNamespace: metav1.NamespaceSystem,
		}

		t.Run(tc.name, func(t *testing.T) {
			appRepo, caCertSecret, authSecret, err := chUtils.parseDetailsForHTTPClient(tc.details, "dummy-user-token")

			if err != nil {
				if tc.errorExpected {
					return
				}
				t.Fatalf("%+v", err)
			} else {
				if tc.errorExpected {
					t.Fatalf("got: nil, want: error")
				}
			}

			// If the Auth header was set, secrets should be returned
			if tc.appRepoSpec.Auth.Header != nil && authSecret == nil {
				t.Errorf("Expecting auth secret")
			}
			if tc.appRepoSpec.Auth.CustomCA != nil && caCertSecret == nil {
				t.Errorf("Expecting auth secret")
			}
			// The client holds a reference to the appRepo.
			if got, want := appRepo, apprepos[0]; !cmp.Equal(got, want) {
				t.Errorf(cmp.Diff(got, want))
			}
		})
	}
}

// Fake server for repositories and charts
type fakeHTTPClient struct {
	repoURL   string
	chartURLs []string
	index     *repo.IndexFile
	userAgent string
	// TODO(absoludity): perhaps switch to use httptest instead of our own fake?
	requests       []*http.Request
	defaultHeaders http.Header
}

// Do for this fake client will return a chart if it exists in the
// index *and* the corresponding chart exists in the testdata directory.
func (f *fakeHTTPClient) Do(h *http.Request) (*http.Response, error) {
	// Record the request for later test assertions.
	for k, v := range f.defaultHeaders {
		// Only add the default header if it's not already set in the request.
		if _, ok := h.Header[k]; !ok {
			h.Header[k] = v
		}
	}
	f.requests = append(f.requests, h)
	if f.userAgent != "" && h.Header.Get("User-Agent") != f.userAgent {
		return nil, fmt.Errorf("Wrong user agent: %s", h.Header.Get("User-Agent"))
	}
	if h.URL.String() == fmt.Sprintf("%sindex.yaml", f.repoURL) {
		// Return fake chart index
		body, err := json.Marshal(*f.index)
		if err != nil {
			return nil, err
		}
		return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, nil
	}
	for _, chartURL := range f.chartURLs {
		if h.URL.String() == chartURL {
			// Fake chart response
			testChartPath := path.Join(".", "testdata", h.URL.Path)
			f, err := os.Open(testChartPath)
			if err != nil {
				return &http.Response{StatusCode: 404}, fmt.Errorf("unable to open test chart archive: %q", testChartPath)
			}
			return &http.Response{StatusCode: 200, Body: f}, nil
		}
	}
	// Unexpected path
	return &http.Response{StatusCode: 404}, fmt.Errorf("Unexpected path %q for chartURLs %+v", h.URL.String(), f.chartURLs)
}

func newHTTPClient(repoURL string, charts []Details, userAgent string) kube.HTTPClient {
	var chartURLs []string
	entries := map[string]repo.ChartVersions{}
	// Populate Chart registry with content of the given helmReleases
	for _, ch := range charts {
		chartMeta := chartv2.Metadata{Name: ch.ChartName, Version: ch.Version}
		chartURL := fmt.Sprintf("%s%s-%s.tgz", repoURL, ch.ChartName, ch.Version)
		chartURLs = append(chartURLs, chartURL)
		chartVersion := repo.ChartVersion{Metadata: &chartMeta, URLs: []string{chartURL}}
		chartVersions := []*repo.ChartVersion{&chartVersion}
		entries[ch.ChartName] = chartVersions
	}
	index := &repo.IndexFile{APIVersion: "v1", Generated: time.Now(), Entries: entries}
	return &fakeHTTPClient{
		repoURL:        repoURL,
		chartURLs:      chartURLs,
		index:          index,
		userAgent:      userAgent,
		defaultHeaders: http.Header{"User-Agent": []string{userAgent}},
	}
}

// getFakeClientRequests returns the requests which were issued to the fake test client.
func getFakeClientRequests(t *testing.T, c kube.HTTPClient) []*http.Request {
	fakeClient, ok := c.(*fakeHTTPClient)
	if !ok {
		t.Fatalf("client was not a fakeHTTPClient")
	}
	return fakeClient.requests
}

func TestGetChart(t *testing.T) {
	const repoName = "foo-repo"
	testCases := []struct {
		name             string
		chartVersion     string
		userAgent        string
		requireV1Support bool
		errorExpected    bool
	}{
		{
			name:         "gets the chart without a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "",
		},
		{
			name:         "gets the chart with a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "tiller-proxy/devel",
		},
		{
			name:             "gets a v2 chart without error when v1 support not required",
			chartVersion:     "5.1.1-apiVersionV2",
			requireV1Support: false,
		},
		{
			name:             "returns an error for a v2 chart if v1 support required",
			chartVersion:     "5.1.1-apiVersionV2",
			requireV1Support: true,
			errorExpected:    true,
		},
	}

	const repoURL = "http://example.com/"
	for _, tc := range testCases {
		target := Details{
			AppRepositoryResourceName: repoName,
			ChartName:                 "nginx",
			ReleaseName:               "foo",
			Version:                   tc.chartVersion,
		}
		t.Run(tc.name, func(t *testing.T) {
			httpClient := newHTTPClient(repoURL, []Details{target}, tc.userAgent)
			chUtils := Client{
				userAgent: tc.userAgent,
				appRepo: &appRepov1.AppRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      repoName,
						Namespace: metav1.NamespaceSystem,
					},
					Spec: appRepov1.AppRepositorySpec{
						URL: repoURL,
					},
				},
			}
			ch, err := chUtils.GetChart(&target, httpClient, tc.requireV1Support)

			if err != nil {
				if tc.errorExpected {
					if got, want := err.Error(), "apiVersion 'v2' is not valid. The value must be \"v1\""; got != want {
						t.Fatalf("got: %q, want: %q", got, want)
					} else {
						// Continue to the next test.
						return
					}
				}
				t.Fatalf("Unexpected error: %v", err)
			}
			// Currently tests return an nginx chart from ./testdata
			// We need to ensure it got loaded in both version formats.
			if got, want := ch.Helm2Chart.GetMetadata().GetName(), "nginx"; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
			if ch.Helm3Chart == nil {
				t.Errorf("got: nil, want: non-nil")
			} else if got, want := ch.Helm3Chart.Name(), "nginx"; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}

			requests := getFakeClientRequests(t, httpClient)
			// We expect one request for the index and one for the chart.
			if got, want := len(requests), 2; got != want {
				t.Fatalf("got: %d, want %d", got, want)
			}

			for i, url := range []string{
				chUtils.appRepo.Spec.URL + "index.yaml",
				fmt.Sprintf("%s%s-%s.tgz", chUtils.appRepo.Spec.URL, target.ChartName, target.Version),
			} {
				if got, want := requests[i].URL.String(), url; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
				if got, want := requests[i].Header.Get("User-Agent"), tc.userAgent; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			}

		})
	}
}

func TestGetIndexFromCache(t *testing.T) {
	repoURL := "https://test.com"
	data := []byte("foo")
	index, sha := getIndexFromCache(repoURL, data)
	if index != nil {
		t.Error("Index should be empty since it's not in the cache yet")
	}
	fakeIndex := &repo.IndexFile{}
	storeIndexInCache(repoURL, fakeIndex, sha)
	index, _ = getIndexFromCache(repoURL, data)
	if index != fakeIndex {
		t.Error("It should return the stored index")
	}
}

func TestClientWithDefaultHeaders(t *testing.T) {
	testCases := []struct {
		name            string
		requestHeaders  http.Header
		defaultHeaders  http.Header
		expectedHeaders http.Header
	}{
		{
			name:            "no headers added when none set",
			defaultHeaders:  http.Header{},
			expectedHeaders: http.Header{},
		},
		{
			name:            "existing headers in the request remain present",
			requestHeaders:  http.Header{"Some-Other": []string{"value"}},
			defaultHeaders:  http.Header{},
			expectedHeaders: http.Header{"Some-Other": []string{"value"}},
		},
		{
			name: "headers are set when present",
			defaultHeaders: http.Header{
				"User-Agent":    []string{"foo/devel"},
				"Authorization": []string{"some-token"},
			},
			expectedHeaders: http.Header{
				"User-Agent":    []string{"foo/devel"},
				"Authorization": []string{"some-token"},
			},
		},
		{
			name: "headers can have multiple values",
			defaultHeaders: http.Header{
				"Authorization": []string{"some-token", "some-other-token"},
			},
			expectedHeaders: http.Header{
				"Authorization": []string{"some-token", "some-other-token"},
			},
		},
		{
			name: "default headers do not overwrite request headers",
			requestHeaders: http.Header{
				"Authorization":        []string{"request-auth-token"},
				"Other-Request-Header": []string{"other-request-header"},
			},
			defaultHeaders: http.Header{
				"Authorization":        []string{"default-auth-token"},
				"Other-Default-Header": []string{"other-default-header"},
			},
			expectedHeaders: http.Header{
				"Authorization":        []string{"request-auth-token"},
				"Other-Request-Header": []string{"other-request-header"},
				"Other-Default-Header": []string{"other-default-header"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeHTTPClient{
				defaultHeaders: tc.defaultHeaders,
			}

			request, err := http.NewRequest("GET", "http://example.com/foo", nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			for k, v := range tc.requestHeaders {
				request.Header[k] = v
			}
			client.Do(request)

			requestsWithHeaders := getFakeClientRequests(t, client)
			if got, want := len(requestsWithHeaders), 1; got != want {
				t.Fatalf("got: %d, want: %d", got, want)
			}

			requestWithHeader := requestsWithHeaders[0]

			if got, want := requestWithHeader.Header, tc.expectedHeaders; !cmp.Equal(got, want) {
				t.Errorf(cmp.Diff(want, got))
			}
		})
	}
}

func TestGetRegistrySecretsPerDomain(t *testing.T) {
	const (
		userAuthToken = "ignored"
		namespace     = "user-namespace"
		// Secret created with
		// k create secret docker-registry test-secret --dry-run --docker-email=a@b.com --docker-password='password' --docker-username='username' --docker-server='https://index.docker.io/v1/' -o yaml
		indexDockerIOCred   = `{"auths":{"https://index.docker.io/v1/":{"username":"username","password":"password","email":"a@b.com","auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`
		otherExampleComCred = `{"auths":{"other.example.com":{"username":"username","password":"password","email":"a@b.com","auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`
	)

	testCases := []struct {
		name             string
		secretNames      []string
		existingSecrets  []*corev1.Secret
		secretsPerDomain map[string]string
		expectError      bool
	}{
		{
			name:             "it returns an empty map if there are no secret names",
			secretNames:      nil,
			secretsPerDomain: map[string]string{},
		},
		{
			name:        "it returns an error if a secret does not exist",
			secretNames: []string{"should-exist"},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret is not a dockerConfigJSON type",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte("whatevs"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret data does not have .dockerconfigjson key",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"custom-secret-key": []byte("whatevs"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret .dockerconfigjson value is not json decodable",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte("not json"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns the registry secrets per domain",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(indexDockerIOCred),
					},
				},
			},
			secretsPerDomain: map[string]string{
				"https://index.docker.io/v1/": "bitnami-repo",
			},
		},
		{
			name:        "it includes secrets for multiple servers",
			secretNames: []string{"bitnami-repo1", "bitnami-repo2"},
			existingSecrets: []*corev1.Secret{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo1",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(indexDockerIOCred),
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo2",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(otherExampleComCred),
					},
				},
			},
			secretsPerDomain: map[string]string{
				"https://index.docker.io/v1/": "bitnami-repo1",
				"other.example.com":           "bitnami-repo2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &kube.FakeHandler{Secrets: tc.existingSecrets}

			secretsPerDomain, err := getRegistrySecretsPerDomain(tc.secretNames, "default", namespace, "token", client)
			if got, want := err != nil, tc.expectError; !cmp.Equal(got, want) {
				t.Fatalf("got: %t, want: %t, err was: %+v", got, want, err)
			}
			if err != nil {
				return
			}

			if got, want := secretsPerDomain, tc.secretsPerDomain; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
