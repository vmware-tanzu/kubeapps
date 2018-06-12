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
	"fmt"
	"testing"
	"time"

	"github.com/arschles/assert"
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

	res, err := FindChartInRepoIndex(index, repoURL, name, version)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if res != expectedURL {
		t.Errorf("Expecting %s to be resolved as %s", res, expectedURL)
	}
}
