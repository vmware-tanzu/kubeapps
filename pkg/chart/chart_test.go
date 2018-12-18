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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/arschles/assert"
	corev1 "k8s.io/api/core/v1"
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
	data := `{
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
	}`
	expectedDetails := Details{
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
	}
	ch := Chart{}
	details, err := ch.ParseDetails([]byte(data))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !reflect.DeepEqual(expectedDetails, *details) {
		t.Errorf("%v != %v", expectedDetails, *details)
	}
}

// Fake server for repositories and charts
type fakeHTTPClient struct {
	repoURLs  []string
	chartURLs []string
	index     *repo.IndexFile
}

func (f *fakeHTTPClient) Do(h *http.Request) (*http.Response, error) {
	for _, repoURL := range f.repoURLs {
		if h.URL.String() == fmt.Sprintf("%sindex.yaml", repoURL) {
			// Return fake chart index (not customizable per repo)
			body, err := json.Marshal(*f.index)
			if err != nil {
				fmt.Printf("Error! %v", err)
			}
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, nil
		}
	}
	for _, chartURL := range f.chartURLs {
		if h.URL.String() == chartURL {
			// Simulate download time
			time.Sleep(100 * time.Millisecond)
			// Fake chart response
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte{}))}, nil
		}
	}
	// Unexpected path
	return &http.Response{StatusCode: 404}, fmt.Errorf("Unexpected path")
}

func fakeLoadChart(in io.Reader) (*chart.Chart, error) {
	return &chart.Chart{}, nil
}

func newHTTPClient(charts []Details) fakeHTTPClient {
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
	return fakeHTTPClient{repoURLs, chartURLs, index}
}

func TestGetChart(t *testing.T) {
	target := Details{
		RepoURL:     "http://foo.com/",
		ChartName:   "test",
		ReleaseName: "foo",
		Version:     "1.0.0",
	}
	httpClient := newHTTPClient([]Details{target})
	kubeClient := fake.NewSimpleClientset()
	chUtils := Chart{
		kubeClient: kubeClient,
		load:       fakeLoadChart,
	}
	ch, err := chUtils.GetChart(&target, &httpClient)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ch == nil {
		t.Errorf("It should return a Chart")
	}
}
