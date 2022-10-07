// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	helmfake "github.com/vmware-tanzu/kubeapps/pkg/helm/fake"
	helmtest "github.com/vmware-tanzu/kubeapps/pkg/helm/test"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/repo"
	"io"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
			assert.NoError(t, err)
			assert.Equal(t, chartURL, tt.wantedURL, "url")
		})
	}
}

func TestGetChartHttp(t *testing.T) {
	const repoName = "foo-repo"
	testCases := []struct {
		name          string
		chartVersion  string
		userAgent     string
		tarballURL    string
		errorExpected bool
	}{
		{
			name:         "gets the chart with tarballURL",
			chartVersion: "5.1.1-apiVersionV1",
			tarballURL:   "http://example.com/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets the chart without a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "",
			tarballURL:   "http://example.com/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets the chart with a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "kubeapps-apis/devel",
			tarballURL:   "http://example.com/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets a v2 chart without error when v1 support not required",
			chartVersion: "5.1.1-apiVersionV2",
			tarballURL:   "http://example.com/nginx-5.1.1-apiVersionV2.tgz",
		},
	}

	const repoURL = "http://example.com/"
	for _, tc := range testCases {
		target := ChartDetails{
			AppRepositoryResourceName: repoName,
			ChartName:                 "nginx",
			Version:                   tc.chartVersion,
			TarballURL:                tc.tarballURL,
		}
		t.Run(tc.name, func(t *testing.T) {
			httpClient := newHTTPClient(repoURL, []ChartDetails{target}, tc.userAgent)
			chUtils := HelmRepoClient{
				userAgent: tc.userAgent,
			}
			chUtils.netClient = httpClient
			ch, err := chUtils.GetChart(&target, repoURL)

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
			if ch == nil {
				t.Errorf("got: nil, want: non-nil")
			} else if got, want := ch.Name(), "nginx"; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}

			requests := getFakeClientRequests(t, httpClient)
			expectedLen := 1
			if tc.tarballURL == "" {
				// We expect one request for the index and one for the chart
				expectedLen = 2
			}

			if got, want := len(requests), expectedLen; got != want {
				t.Fatalf("got: %d, want %d", got, want)
			}
			for i, url := range []string{
				repoURL + "index.yaml",
				fmt.Sprintf("%s%s-%s.tgz", repoURL, target.ChartName, target.Version),
			} {
				// Skip the index.yaml request if a tarballURL is passed
				if tc.tarballURL != "" {
					continue
				}
				if got, want := requests[i].URL.String(), url; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
				if got, want := requests[i].Header.Get("User-Agent"), tc.userAgent; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			}

		})
	}

	t.Run("it should fail if the netClient is not instantiated", func(t *testing.T) {
		cli := NewChartClient("")
		_, err := cli.GetChart(nil, "")
		assert.Error(t, fmt.Errorf("unable to retrieve chart, Init should be called first"), err)
	})
}

func TestOCIClient(t *testing.T) {
	t.Run("InitClient - Creates puller with User-Agent header", func(t *testing.T) {
		cli := NewOCIClient("foo")
		err := cli.Init(&appRepov1.AppRepository{}, &corev1.Secret{}, &corev1.Secret{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		helmtest.CheckHeader(t, cli.(*OCIRepoClient).puller, "User-Agent", "foo")
	})

	t.Run("InitClient - Creates puller with Authorization", func(t *testing.T) {
		cli := NewOCIClient("")
		appRepo := &appRepov1.AppRepository{
			Spec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					Header: &appRepov1.AppRepositoryAuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{},
							Key:                  "custom-secret-key",
						},
					},
				},
			},
		}
		authSecret := &corev1.Secret{
			Data: map[string][]byte{
				"custom-secret-key": []byte("Basic Auth"),
			},
		}
		err := cli.Init(appRepo, &corev1.Secret{}, authSecret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		helmtest.CheckHeader(t, cli.(*OCIRepoClient).puller, "Authorization", "Basic Auth")
	})

	t.Run("InitClient - Creates puller with Docker Creds Authorization", func(t *testing.T) {
		cli := NewOCIClient("")
		appRepo := &appRepov1.AppRepository{
			Spec: appRepov1.AppRepositorySpec{
				Auth: appRepov1.AppRepositoryAuth{
					Header: &appRepov1.AppRepositoryAuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{},
							Key:                  ".dockerconfigjson",
						},
					},
				},
			},
		}
		authSecret := &corev1.Secret{
			Data: map[string][]byte{
				".dockerconfigjson": []byte(`{"auths":{"foo":{"username":"foo","password":"bar"}}}`),
			},
		}
		err := cli.Init(appRepo, &corev1.Secret{}, authSecret)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		// Authorization: Basic base64('foo:bar')
		helmtest.CheckHeader(t, cli.(*OCIRepoClient).puller, "Authorization", "Basic Zm9vOmJhcg==")
	})

	t.Run("GetChart - Fails if the puller has not been instantiated", func(t *testing.T) {
		cli := NewOCIClient("foo")
		_, err := cli.GetChart(nil, "")
		assert.Error(t, fmt.Errorf("unable to retrieve chart, Init should be called first"), err)
	})

	t.Run("GetChart - Fails if the URL is not valid", func(t *testing.T) {
		cli := NewOCIClient("foo")
		cli.(*OCIRepoClient).puller = &helmfake.OCIPuller{}
		_, err := cli.GetChart(nil, "foo")
		if !strings.Contains(err.Error(), "invalid URI for request") {
			t.Errorf("Unexpected error %v", err)
		}
	})

	t.Run("GetChart - Returns a chart", func(t *testing.T) {
		cli := NewOCIClient("foo")
		data, err := os.ReadFile("./testdata/nginx-5.1.1-apiVersionV2.tgz")
		assert.NoError(t, err)
		cli.(*OCIRepoClient).puller = &helmfake.OCIPuller{
			ExpectedName: "foo/bar/nginx:5.1.1",
			Content:      map[string]*bytes.Buffer{"5.1.1": bytes.NewBuffer(data)},
		}
		ch, err := cli.GetChart(&ChartDetails{ChartName: "nginx", Version: "5.1.1"}, "http://foo/bar")
		if ch == nil {
			t.Errorf("Unexpected error: %s", err)
		} else if ch.Name() != "nginx" || ch.Metadata.Version != "5.1.1" {
			t.Errorf("Unexpected chart %s:%s", ch.Name(), ch.Metadata.Version)
		}
	})

	t.Run("GetChart - Returns a chart with multiple slashes", func(t *testing.T) {
		cli := NewOCIClient("foo")
		data, err := os.ReadFile("./testdata/nginx-5.1.1-apiVersionV2.tgz")
		assert.NoError(t, err)
		cli.(*OCIRepoClient).puller = &helmfake.OCIPuller{
			ExpectedName: "foo/bar/bar/nginx:5.1.1",
			Content:      map[string]*bytes.Buffer{"5.1.1": bytes.NewBuffer(data)},
		}
		ch, err := cli.GetChart(&ChartDetails{ChartName: "nginx", Version: "5.1.1"}, "http://foo/bar%2Fbar")
		if ch == nil {
			t.Errorf("Unexpected error: %s", err)
		} else if ch.Name() != "nginx" || ch.Metadata.Version != "5.1.1" {
			t.Errorf("Unexpected chart %s:%s", ch.Name(), ch.Metadata.Version)
		}
	})
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
			_, err = client.Do(request)
			if err != nil && !strings.Contains(err.Error(), "Unexpected path") {
				t.Fatalf("%+v", err)
			}
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
		return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader(body))}, nil
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

// getFakeClientRequests returns the requests which were issued to the fake test client.
func getFakeClientRequests(t *testing.T, c httpclient.Client) []*http.Request {
	fakeClient, ok := c.(*fakeHTTPClient)
	if !ok {
		t.Fatalf("client was not a fakeHTTPClient")
	}
	return fakeClient.requests
}

func newHTTPClient(repoURL string, charts []ChartDetails, userAgent string) httpclient.Client {
	var chartURLs []string
	entries := map[string]repo.ChartVersions{}
	// Populate Chart registry with content of the given helmReleases
	for _, ch := range charts {
		chartMeta := chart.Metadata{Name: ch.ChartName, Version: ch.Version}
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
