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
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arschles/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

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

	chartMeta := chart.Metadata{Name: name, Version: version}
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
			name: "parses repoUrl and auth",
			data: `{
	        	"repoUrl": "foo.com",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar",
	        	"auth": {
	        		"header": {
	        			"secretKeyRef": {
	        				"key": "bar"
	        			}
	        		}
	        	}
	        }`,
			expected: &Details{
				RepoURL:     "foo.com",
				ChartName:   "test",
				ReleaseName: "foo",
				Version:     "1.0.0",
				Values:      "foo: bar",
				Auth: Auth{
					Header: &AuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							Key: "bar",
						},
					},
				},
			},
		},
		{
			name: "parses app repo resource",
			data: `{
				"appRepositoryResourceName": "my-chart-repo",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar"
	        }`,
			expected: &Details{
				AppRepositoryResourceName: "my-chart-repo",
				ChartName:                 "test",
				ReleaseName:               "foo",
				Version:                   "1.0.0",
				Values:                    "foo: bar",
			},
		},
		{
			name: "error returned if both resource and repo url specified",
			data: `{
	        	"repoUrl": "foo.com",
				"appRepositoryResourceName": "my-chart-repo",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar"
			}`,
			err: true,
		},
		{
			name: "error returned if both resource and auth header specified",
			data: `{
				"appRepositoryResourceName": "my-chart-repo",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar",
	        	"auth": {
	        		"header": {
	        			"secretKeyRef": {
	        				"key": "bar"
	        			}
	        		}
	        	}
			}`,
			err: true,
		},
		{
			name: "error returned if both resource and auth CA specified",
			data: `{
				"appRepositoryResourceName": "my-chart-repo",
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar",
	        	"auth": {
	        		"customCA": {
	        			"secretKeyRef": {
	        				"key": "bar"
	        			}
	        		}
	        	}
			}`,
			err: true,
		},
		{
			name: "specifying neither repoUrl nor app repo resource is valid",
			data: `{
	        	"chartName": "test",
	        	"releaseName": "foo",
	        	"version": "1.0.0",
	        	"values": "foo: bar"
			}`,
			expected: &Details{
				ChartName:   "test",
				ReleaseName: "foo",
				Version:     "1.0.0",
				Values:      "foo: bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ch := Chart{}
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

// fakeLoadChart implements LoadChart interface.
func fakeLoadChart(in io.Reader) (*chart.Chart, error) {
	return &chart.Chart{}, nil
}

const pem_cert = `
-----BEGIN CERTIFICATE-----
MIIDETCCAfkCFEY03BjOJGqOuIMoBewOEDORMewfMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkRFMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMTkwODE5MDQxNzU5WhcNMTkxMDA4MDQx
NzU5WjBFMQswCQYDVQQGEwJERTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAzA+X6HcScuHxqxCc5gs68weW8i72qMjvcWvBG064SvpTuNDK
ECEGvug6f8SFJjpA+hWjlqR5+UPMdfjMKPUEg1CI8JZm6lyNiB54iY50qvhv+qQg
1STdAWNTzvqUXUMGIImzeXFnErxlq8WwwLGwPNT4eFxF8V8fzIhR8sqQKFLOqvpS
7sCQwF5QOhziGfS+zParDLFsBoXQpWyDKqxb/yBSPwqijKkuW7kF4jGfPHD0Re3+
rspXiq8+jWSwSJIPSIbya8DQqrMwFeLCAxABidPnlrwS0UUion557ylaBK6Cv0UB
MojA4SMfjm5xRdzrOcoE8EcabxqoQD5rCIBgFQIDAQABMA0GCSqGSIb3DQEBCwUA
A4IBAQCped08LTojPejkPqmp1edZa9rWWrCMviY5cvqb6t3P3erse+jVcBi9NOYz
8ewtDbR0JWYvSW6p3+/nwyDG4oVfG5TiooAZHYHmgg4x9+5h90xsnmgLhIsyopPc
Rltj86tRCl1YiuRpkWrOfRBGdYfkGEG4ihJzLHWRMCd1SmMwnmLliBctD7IeqBKw
UKt8wcroO8/sj/Xd1/LCtNZ79/FdQFa4l3HnzhOJOrlQyh4gyK05EKdg6vv3un17
l6NEPfiXd7dZvsWi9uY/PGBhu9EY/bdvuIOWDNNK262azk1A56HINpMrYBUcfti1
YrvYQHgOtHsqCB/hFHWfZp1lg2Sx
-----END CERTIFICATE-----
`

func TestInitNetClient(t *testing.T) {
	// TODO(mnelson): currently the InitNetClient swallows any error during
	// call to SystemCertPool, silently creating an empty cert pool. If that
	// path is taken on the test system, this test will fail. Find out why.
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	const authHeaderSecret = "really-secret-stuff"

	testCases := []struct {
		name             string
		details          *Details
		secretData       string
		errorExpected    bool
		numCertsExpected int
	}{
		{
			name: "default cert pool without auth",
			details: &Details{
				Auth: Auth{},
			},
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "cert added when present in auth",
			details: &Details{
				Auth: Auth{
					CustomCA: &CustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"custom-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			secretData:       pem_cert,
			numCertsExpected: len(systemCertPool.Subjects()) + 1,
		},
		{
			name: "errors if secret for custom CA cannot be found",
			details: &Details{
				Auth: Auth{
					CustomCA: &CustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"other-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			secretData:    pem_cert,
			errorExpected: true,
		},
		{
			name: "errors if custom CA cannot be found in secret",
			details: &Details{
				Auth: Auth{
					CustomCA: &CustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"custom-secret-name"},
							"some-other-secret-key",
							nil,
						},
					},
				},
			},
			secretData:    pem_cert,
			errorExpected: true,
		},
		{
			name: "errors if custom CA cannot be parsed",
			details: &Details{
				Auth: Auth{
					CustomCA: &CustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"custom-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			secretData:    "not a valid cert",
			errorExpected: true,
		},
		{
			name: "authorization header added when present in auth",
			details: &Details{
				Auth: Auth{
					Header: &AuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"custom-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			secretData:       authHeaderSecret,
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "errors if auth secret cannot be found",
			details: &Details{
				Auth: Auth{
					Header: &AuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{"other-secret-name"},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			secretData:    authHeaderSecret,
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		// Create the client with the test case data.
		kubeClient := fake.NewSimpleClientset(&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "custom-secret-name",
			},
			Data: map[string][]byte{
				"custom-secret-key": []byte(tc.secretData),
			},
		})
		chUtils := Chart{
			kubeClient: kubeClient,
			load:       fakeLoadChart,
		}

		t.Run(tc.name, func(t *testing.T) {
			httpClient, err := chUtils.InitNetClient(tc.details)

			if err != nil {
				if tc.errorExpected {
					return
				}
				t.Fatalf("%+v", err)
			}

			clientWithDefaultHeaders, ok := httpClient.(*clientWithDefaultHeaders)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			client, ok := clientWithDefaultHeaders.client.(*http.Client)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			transport, ok := client.Transport.(*http.Transport)
			certPool := transport.TLSClientConfig.RootCAs

			if got, want := len(certPool.Subjects()), tc.numCertsExpected; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// If the Auth header was set, the default Authorization header should be set
			// from the secret.
			if tc.details.Auth.Header != nil {
				_, ok := clientWithDefaultHeaders.defaultHeaders["Authorization"]
				if !ok {
					t.Fatalf("expected Authorization header but found none")
				}
				if got, want := clientWithDefaultHeaders.defaultHeaders.Get("Authorization"), authHeaderSecret; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			}
		})
	}
}

// Fake server for repositories and charts
type fakeHTTPClient struct {
	repoURLs  []string
	chartURLs []string
	index     *repo.IndexFile
	userAgent string
	// TODO(absoludity): perhaps switch to use httptest instead of our own fake?
	requests []*http.Request
}

func (f *fakeHTTPClient) Do(h *http.Request) (*http.Response, error) {
	// Record the request for later test assertions.
	f.requests = append(f.requests, h)
	if f.userAgent != "" && h.Header.Get("User-Agent") != f.userAgent {
		return nil, fmt.Errorf("Wrong user agent: %s", h.Header.Get("User-Agent"))
	}
	for _, repoURL := range f.repoURLs {
		if h.URL.String() == fmt.Sprintf("%sindex.yaml", repoURL) {
			// Return fake chart index (not customizable per repo)
			body, err := json.Marshal(*f.index)
			if err != nil {
				return nil, err
			}
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, nil
		}
	}
	for _, chartURL := range f.chartURLs {
		if h.URL.String() == chartURL {
			// Fake chart response
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte{}))}, nil
		}
	}
	// Unexpected path
	return &http.Response{StatusCode: 404}, fmt.Errorf("Unexpected path")
}

func newHTTPClient(charts []Details, userAgent string) HTTPClient {
	var repoURLs []string
	var chartURLs []string
	entries := map[string]repo.ChartVersions{}
	// Populate Chart registry with content of the given helmReleases
	for _, ch := range charts {
		repoURLs = append(repoURLs, ch.RepoURL)
		chartMeta := chart.Metadata{Name: ch.ChartName, Version: ch.Version}
		chartURL := fmt.Sprintf("%s%s-%s.tgz", ch.RepoURL, ch.ChartName, ch.Version)
		chartURLs = append(chartURLs, chartURL)
		chartVersion := repo.ChartVersion{Metadata: &chartMeta, URLs: []string{chartURL}}
		chartVersions := []*repo.ChartVersion{&chartVersion}
		entries[ch.ChartName] = chartVersions
	}
	index := &repo.IndexFile{APIVersion: "v1", Generated: time.Now(), Entries: entries}
	return &clientWithDefaultHeaders{
		client: &fakeHTTPClient{
			repoURLs:  repoURLs,
			chartURLs: chartURLs,
			index:     index,
			userAgent: userAgent,
		},
		defaultHeaders: http.Header{"User-Agent": []string{userAgent}},
	}
}

// getFakeClientRequests returns the requests which were issued to the fake test client.
func getFakeClientRequests(t *testing.T, c HTTPClient) []*http.Request {
	clientWithDefaultUA, ok := c.(*clientWithDefaultHeaders)
	if !ok {
		t.Fatalf("client was not a clientWithDefaultUA")
	}
	fakeClient, ok := clientWithDefaultUA.client.(*fakeHTTPClient)
	if !ok {
		t.Fatalf("client was not a fakeHTTPClient")
	}
	return fakeClient.requests
}

func TestGetChart(t *testing.T) {
	target := Details{
		RepoURL:     "http://foo.com/",
		ChartName:   "test",
		ReleaseName: "foo",
		Version:     "1.0.0",
	}
	testCases := []struct {
		name      string
		userAgent string
	}{
		{
			name:      "GetChart without user agent",
			userAgent: "",
		},
		{
			name:      "GetChart with user agent",
			userAgent: "tiller-proxy/devel",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := newHTTPClient([]Details{target}, tc.userAgent)
			kubeClient := fake.NewSimpleClientset()
			chUtils := Chart{
				kubeClient: kubeClient,
				load:       fakeLoadChart,
				userAgent:  tc.userAgent,
			}
			ch, err := chUtils.GetChart(&target, httpClient)

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			// Currently tests return an empty chart object.
			if got, want := ch, &(chart.Chart{}); !cmp.Equal(got, want) {
				t.Errorf("got: %v, want: %v", got, want)
			}

			requests := getFakeClientRequests(t, httpClient)
			// We expect one request for the index and one for the chart.
			if got, want := len(requests), 2; got != want {
				t.Fatalf("got: %d, want %d", got, want)
			}

			for i, url := range []string{
				target.RepoURL + "index.yaml",
				fmt.Sprintf("%s%s-%s.tgz", target.RepoURL, target.ChartName, target.Version),
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
			client := &clientWithDefaultHeaders{
				client:         &fakeHTTPClient{},
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
