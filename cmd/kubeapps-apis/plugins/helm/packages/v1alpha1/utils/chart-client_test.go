// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	helmfake "github.com/vmware-tanzu/kubeapps/pkg/helm/fake"
	helmtest "github.com/vmware-tanzu/kubeapps/pkg/helm/test"
	corev1 "k8s.io/api/core/v1"

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
			assert.Equal(t, chartURL.String(), tt.wantedURL, "url")
		})
	}
}

func TestGetChartHttp(t *testing.T) {
	const repoName = "foo-repo"
	testCases := []struct {
		name         string
		chartVersion string
		userAgent    string
		tarballPath  string
	}{
		{
			name:         "gets the chart with tarballURL",
			chartVersion: "5.1.1-apiVersionV1",
			tarballPath:  "/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets the chart without a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "",
			tarballPath:  "/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets the chart with a user agent",
			chartVersion: "5.1.1-apiVersionV1",
			userAgent:    "kubeapps-apis/devel",
			tarballPath:  "/nginx-5.1.1-apiVersionV1.tgz",
		},
		{
			name:         "gets a v2 chart without error when v1 support not required",
			chartVersion: "5.1.1-apiVersionV2",
			tarballPath:  "/nginx-5.1.1-apiVersionV2.tgz",
		},
	}

	for _, tc := range testCases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == tc.tarballPath {
				data, err := os.ReadFile(path.Join(".", "testdata", tc.tarballPath))
				if err != nil {
					t.Fatalf("%+v", err)
				}
				w.WriteHeader(200)
				_, err = w.Write(data)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				return
			}
			w.WriteHeader(404)
		}))
		defer server.Close()
		target := ChartDetails{
			AppRepositoryResourceName: repoName,
			ChartName:                 "nginx",
			Version:                   tc.chartVersion,
			TarballURL:                server.URL + tc.tarballPath,
		}
		t.Run(tc.name, func(t *testing.T) {

			chUtils := HelmRepoClient{
				userAgent: tc.userAgent,
			}
			chUtils.netClient = server.Client()
			ch, err := chUtils.GetChart(&target, server.URL)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if ch == nil {
				t.Errorf("got: nil, want: non-nil")
			}

			if got, want := ch.Name(), "nginx"; got != want {
				t.Errorf("got: %q, want: %q", got, want)
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
		if !strings.Contains(err.Error(), "missing chart details") {
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
		ch, err := cli.GetChart(&ChartDetails{ChartName: "nginx", Version: "5.1.1", TarballURL: "oci://foo/bar/nginx:5.1.1"}, "http://foo/bar")
		if ch == nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if ch.Name() != "nginx" || ch.Metadata.Version != "5.1.1" {
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
		ch, err := cli.GetChart(&ChartDetails{ChartName: "nginx", Version: "5.1.1", TarballURL: "oci://foo/bar%2Fbar/nginx:5.1.1"}, "http://foo/bar%2Fbar")
		if ch == nil {
			t.Errorf("Unexpected error: %s", err)
		}
		if ch.Name() != "nginx" || ch.Metadata.Version != "5.1.1" {
			t.Errorf("Unexpected chart %s:%s", ch.Name(), ch.Metadata.Version)
		}
	})
}
