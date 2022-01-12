/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/disintegration/imaging"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/stretchr/testify/assert"
)

type bodyAPIListResponse struct {
	Data *apiListResponse `json:"data"`
	Meta meta             `json:"meta,omitempty"`
}

type bodyAPIResponse struct {
	Data apiResponse `json:"data"`
}

var chartsList []*models.Chart
var cc count

const (
	testChartReadme      = "# Quickstart\n\n```bash\nhelm install my-repo/my-chart\n```"
	testChartValues      = "image:\n  registry: docker.io\n  repository: my-repo/my-chart\n  tag: 0.1.0"
	testChartSchema      = `{"properties": {"type": "object"}}`
	namespace            = "namespace"
	kubeappsNamespace    = "kubeapps-namespace"
	globalReposNamespace = "kubeapps-repos-global"
	testRepoName         = "my-repo"
)

var testRepo *models.Repo = &models.Repo{Name: testRepoName, Namespace: namespace}

func iconBytes() []byte {
	var b bytes.Buffer
	img := imaging.New(1, 1, color.White)
	imaging.Encode(&b, img, imaging.PNG)
	return b.Bytes()
}

func setMockManager(t *testing.T) (sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// TODO(absoludity): Let's not use globals for storing state like this.
	origManager := manager

	manager = &utils.PostgresAssetManager{&dbutils.PostgresAssetManager{DB: db, GlobalReposNamespace: globalReposNamespace}}

	return mock, func() { db.Close(); manager = origManager }
}

func Test_extractDecodedNamespaceAndRepoAndVersionParams(t *testing.T) {
	tests := []struct {
		name             string
		namespace        string
		repo             string
		version          string
		expectedParamErr string
	}{
		{"params OK", "namespace", "repo", "version", ""},
		{"params NOK namespace", "%%1", "repo", "version", "%%1"},
		{"params NOK repo", "namespace", "%%1", "version", "%%1"},
		{"params NOK version", "namespace", "repo", "%%1", "%%1"},
		{"params NOK (returns the first one errored)", "%%1", "%%2", "%%3", "%%1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := Params{
				"namespace": tt.namespace,
				"repo":      tt.repo,
				"version":   tt.version,
			}
			namespace, repo, version, paramErr, _ := extractDecodedNamespaceAndRepoAndVersionParams(params)
			if tt.expectedParamErr != "" {
				assert.Equal(t, tt.expectedParamErr, paramErr, "expectedParamErr should be the same")
			} else {
				assert.Equal(t, tt.namespace, namespace, "namespace should be the same")
				assert.Equal(t, tt.repo, repo, "repo should be the same")
				assert.Equal(t, tt.version, version, "version should be the same")
			}
		})
	}
}
func Test_extractChartQueryFromRequest(t *testing.T) {
	tests := []struct {
		name               string
		namespace          string
		repo               string
		chartName          string
		version            string
		appVersion         string
		searchQuery        string
		repos              string
		categories         string
		expectedRepos      []string
		expectedCategories []string
	}{
		{"no params ", "", "", "", "", "", "", "", "", []string{}, []string{}},
		{"only categories", "", "", "", "", "", "", "", "cat1,cat2", []string{}, []string{"cat1", "cat2"}},
		{"only repo ", "", "repo1", "", "", "", "", "", "", []string{"repo1"}, []string{}},
		{"only repos ", "", "", "", "", "", "", "repo2,repo3", "", []string{"repo2", "repo3"}, []string{}},
		{"only repo and repos ", "", "repo1", "", "", "", "", "repo2,repo3", "", []string{"repo1", "repo2", "repo3"}, []string{}},
		{"every param", "namespace", "repo1", "chartName", "version", "appVersion", "searchQuery", "repo2,repo3", "cat1,cat2", []string{"repo1", "repo2", "repo3"}, []string{"cat1", "cat2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURI := fmt.Sprintf("/charts?name=%s&version=%s&appversion=%s&repos=%s&categories=%s&q=%s", tt.chartName, tt.version, tt.appVersion, tt.repos, tt.categories, tt.searchQuery)
			req := httptest.NewRequest("GET", requestURI, nil)

			cq := extractChartQueryFromRequest(tt.namespace, tt.repo, req)

			assert.Equal(t, tt.namespace, cq.Namespace, "namespace should be the same")
			assert.Equal(t, tt.chartName, cq.ChartName, "chartName should be the same")
			assert.Equal(t, tt.version, cq.Version, "version should be the same")
			assert.Equal(t, tt.appVersion, cq.AppVersion, "appVersion should be the same")
			assert.Equal(t, tt.searchQuery, cq.SearchQuery, "searchQuery should be the same")
			assert.Equal(t, tt.expectedRepos, cq.Repos, "repos should be the same")
			assert.Equal(t, tt.expectedCategories, cq.Categories, "categories should be the same")

		})
	}
}

func Test_chartAttributes(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"chart enconded has no icon", models.Chart{
			Repo: testRepo, Name: "foo%2Fwordpress", ID: "my-repo/foo%2Fwordpress",
		}},
		{"chart has no icon", models.Chart{
			Repo: testRepo, Name: "wordpress", ID: "my-repo/wordpress",
		}},
		{"chart has a icon", models.Chart{
			Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", RawIcon: iconBytes(), IconContentType: "image/svg",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chartAttributes(namespace, tt.chart)
			assert.Equal(t, tt.chart.ID, c.ID)
			assert.Equal(t, tt.chart.RawIcon, c.RawIcon)
			if len(tt.chart.RawIcon) == 0 {
				assert.Equal(t, len(c.Icon), 0, "icon url should be undefined")
			} else {
				assert.Equal(t, pathPrefix+"/ns/"+namespace+"/assets/"+tt.chart.ID+"/logo", c.Icon, "the icon url should be the same")
				assert.Equal(t, tt.chart.IconContentType, c.IconContentType, "the icon content type should be the same")
			}
		})
	}
}

func Test_chartVersionAttributes(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"my-chart", models.Chart{
			Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}},
		}},
		{"foo%2Fmy-chart", models.Chart{
			Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := chartVersionAttributes(namespace, tt.chart.Repo.Name, tt.chart.Name, tt.chart.ChartVersions[0])
			assert.Equal(t, cv.Version, tt.chart.ChartVersions[0].Version, "version string should be the same")
			assert.Equal(t, cv.Readme, pathPrefix+"/ns/"+namespace+"/assets/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version+"/README.md", "README.md resource path should be the same")
			assert.Equal(t, cv.Values, pathPrefix+"/ns/"+namespace+"/assets/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version+"/values.yaml", "values.yaml resource path should be the same")
		})
	}
}

func Test_newChartResponse(t *testing.T) {
	tests := []struct {
		name      string
		chartName string
		chart     models.Chart
	}{
		{"chart has only one version", "my-chart", models.Chart{
			Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}}},
		},
		{"chart encoded has only one version", "foo%2Fmy-chart", models.Chart{
			Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}}},
		},
		{"chart has many versions", "my-chart", models.Chart{
			Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.2"}, {Version: "0.1.0"}},
		}},
		{"raw_icon is never sent down the wire", "my-chart", models.Chart{
			Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}}, RawIcon: iconBytes(), IconContentType: "image/svg",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cResponse := newChartResponse(&tt.chart)
			assert.Equal(t, cResponse.Type, "chart", "response type is chart")
			assert.Equal(t, cResponse.ID, tt.chart.ID, "chart ID should be the same")
			assert.Equal(t, cResponse.Relationships["latestChartVersion"].Data.(models.ChartVersion).Version, tt.chart.ChartVersions[0].Version, "latestChartVersion should match version at index 0")
			assert.Equal(t, cResponse.Links.(selfLink).Self, pathPrefix+"/ns/"+namespace+"/charts/"+tt.chart.ID, "self link should be the same")
			// We don't send the raw icon down the wire.
			assert.Nil(t, cResponse.Attributes.(models.Chart).RawIcon)
		})
	}
}

func Test_newChartListResponse(t *testing.T) {
	tests := []struct {
		name   string
		input  []*models.Chart
		result []*models.Chart
	}{
		{"no charts", []*models.Chart{}, []*models.Chart{}},
		{"has one chart", []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
		}, []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
		}},
		{"has two charts", []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "wordpress", ID: "my-repo/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}, []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "wordpress", ID: "my-repo/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}},
		{"has two encoded charts", []*models.Chart{
			{Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "foo%2Fwordpress", ID: "my-repo/foo%2Fwordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}, []*models.Chart{
			{Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "foo%2Fwordpress", ID: "my-repo/foo%2Fwordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clResponse := newChartListResponse(tt.input)
			assert.Equal(t, len(clResponse), len(tt.result), "number of charts in response should be the same")
			for i := range tt.result {
				assert.Equal(t, "chart", clResponse[i].Type, "response type is chart")
				assert.Equal(t, tt.result[i].ID, clResponse[i].ID, "chart ID should be the same")
				assert.Equal(t, tt.result[i].ChartVersions[0].Version, clResponse[i].Relationships["latestChartVersion"].Data.(models.ChartVersion).Version, "latestChartVersion should match version at index 0")
				assert.Equal(t, pathPrefix+"/ns/"+namespace+"/charts/"+tt.result[i].ID, clResponse[i].Links.(selfLink).Self, "self link should be the same")
			}
		})
	}
}

func Test_newChartVersionResponse(t *testing.T) {
	tests := []struct {
		name         string
		chart        models.Chart
		expectedIcon string
	}{
		{
			name: "my-chart",
			chart: models.Chart{
				Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.2.3"}},
			},
		},
		{
			name: "foo%2Fmy-chart",
			chart: models.Chart{
				Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.2.3"}},
			},
		},
		{
			name: "RawIcon is never sent down the wire",
			chart: models.Chart{
				Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}}, RawIcon: iconBytes(), IconContentType: "image/svg",
			},
			expectedIcon: "/v1/ns/" + namespace + "/assets/my-repo/my-chart/logo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := range tt.chart.ChartVersions {
				cvResponse := newChartVersionResponse(&tt.chart, tt.chart.ChartVersions[i])
				assert.Equal(t, "chartVersion", cvResponse.Type, "response type is chartVersion")
				assert.Equal(t, tt.chart.Repo.Name+"/"+tt.chart.Name+"-"+tt.chart.ChartVersions[i].Version, cvResponse.ID, "response id should have chart version suffix")
				assert.Equal(t, pathPrefix+"/ns/"+namespace+"/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[i].Version, cvResponse.Links.(interface{}).(selfLink).Self, "self link should be the same")
				assert.Equal(t, tt.chart.ChartVersions[i].Version, cvResponse.Attributes.(models.ChartVersion).Version, "chart version in the response should be the same")

				// The chart should have had its icon url set and raw icon data removed.
				expectedChart := tt.chart
				expectedChart.RawIcon = nil
				expectedChart.Icon = tt.expectedIcon
				expectedChart.ChartVersions = []models.ChartVersion{}
				assert.Equal(t, cvResponse.Relationships["chart"].Data.(interface{}).(models.Chart), expectedChart, "chart in relatioship matches")
			}
		})
	}
}

func Test_newChartVersionListResponse(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"chart has no versions", models.Chart{
			Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{},
		}},
		{"chart has one version", models.Chart{
			Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}},
		}},
		{"chart has many versions", models.Chart{
			Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}, {Version: "0.0.2"}},
		}},
		{"chart encoded has many versions", models.Chart{
			Repo: testRepo, ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}, {Version: "0.0.2"}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cvListResponse := newChartVersionListResponse(&tt.chart)
			assert.Equal(t, len(cvListResponse), len(tt.chart.ChartVersions), "number of chart versions in response should be the same")
			for i := range tt.chart.ChartVersions {
				assert.Equal(t, "chartVersion", cvListResponse[i].Type, "response type is chartVersion")
				assert.Equal(t, tt.chart.ID+"-"+tt.chart.ChartVersions[i].Version, cvListResponse[i].ID, "response id should have chart version suffix")
				assert.Equal(t, pathPrefix+"/ns/"+namespace+"/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[i].Version, cvListResponse[i].Links.(interface{}).(selfLink).Self, "self link should be the same")
				assert.Equal(t, tt.chart.ChartVersions[i].Version, cvListResponse[i].Attributes.(models.ChartVersion).Version, "chart version in the response should be the same")
			}
		})
	}
}

func Test_listCharts(t *testing.T) {
	tests := []struct {
		name   string
		charts []*models.Chart
		meta   meta
	}{
		{"no charts", []*models.Chart{}, meta{TotalPages: 1}},
		{"one chart", []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
		}, meta{TotalPages: 1}},
		{"two charts", []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "dokuwiki", ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}, meta{TotalPages: 1}},
		{"two charts, one encoded", []*models.Chart{
			{Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "dokuwiki", ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}, meta{TotalPages: 1}},
		{"four charts", []*models.Chart{
			{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, Name: "dokuwiki", ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, Name: "drupal", ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, Name: "wordpress", ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})
			rowCount := sqlmock.NewRows([]string{"count"}).AddRow(len(tt.charts))

			for _, chart := range tt.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			mock.ExpectQuery("SELECT info FROM").
				WithArgs(namespace, globalReposNamespace).
				WillReturnRows(rows)

			mock.ExpectQuery("^SELECT count(.+) FROM").
				WillReturnRows(rowCount)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts", nil)
			listChartsWithFilters(w, req, Params{"namespace": namespace})

			assert.Equal(t, http.StatusOK, w.Code)

			var b bodyAPIListResponse
			json.NewDecoder(w.Body).Decode(&b)
			if b.Data == nil {
				t.Fatal("In '" + tt.name + "': " + "chart list shouldn't be null")
			}
			data := *b.Data
			assert.Len(t, data, len(tt.charts))
			for i, resp := range data {
				assert.Equal(t, resp.ID, tt.charts[i].ID, "In '"+tt.name+"': "+"chart id in the response should be the same")
				assert.Equal(t, resp.Type, "chart", "In '"+tt.name+"': "+"response type is chart")
				assert.Equal(t, resp.Links.(map[string]interface{})["self"], pathPrefix+"/ns/"+namespace+"/charts/"+tt.charts[i].ID, "In '"+tt.name+"': "+"self link should be the same")
				assert.Equal(t, resp.Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], tt.charts[i].ChartVersions[0].Version, "In '"+tt.name+"': "+"latestChartVersion should match version at index 0")
			}
			assert.Equal(t, b.Meta, tt.meta, "In '"+tt.name+"': "+"response meta should be the same")
		})
	}
}

func Test_listRepoCharts(t *testing.T) {
	tests := []struct {
		name   string
		repo   string
		query  string
		charts []*models.Chart
		meta   meta
	}{
		{"repo has no charts", "my-repo", "", []*models.Chart{}, meta{TotalPages: 1}},
		{"repo has one chart", "my-repo", "", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts", "my-repo", "", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "my-repo/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}, meta{TotalPages: 1}},
		{"repo has many encoded charts with pagination", "my-repo", "?page=1&size=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/foo%2Fdokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/foo%2Fdrupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 2}},
		{"repo has many charts with pagination (2 pages)", "my-repo", "?page=2&size=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 2}},
		{"repo has many charts with pagination (non existing page)", "my-repo", "?page=3&size=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 2}},
		{"repo has many charts with pagination (out of range size)", "my-repo", "?page=1&size=100", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts with pagination (w/ page, w size)", "my-repo", "?page=2size=3", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts with pagination (w/ page, w zero size)", "my-repo", "?page=2size=0", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts with pagination (w/ wrong page, w/ size)", "my-repo", "?page=-2size=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts with pagination (w/ page, w/o size)", "my-repo", "?page=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
		{"repo has many charts with pagination (w/o page, w/ size)", "my-repo", "?size=2", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 2}},
		{"repo has many charts with pagination (w/o page, w/o size)", "my-repo", "", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}}},
			{Repo: testRepo, ID: "stable/drupal", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "12345"}}},
			{Repo: testRepo, ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "123456"}}},
		}, meta{TotalPages: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})
			rowCount := sqlmock.NewRows([]string{"count"}).AddRow(len(tt.charts))

			for _, chart := range tt.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			expectedParams := []driver.Value{namespace, globalReposNamespace}
			if tt.repo != "" {
				expectedParams = append(expectedParams, tt.repo)
			}
			mock.ExpectQuery("SELECT info FROM charts").
				WithArgs(expectedParams...).
				WillReturnRows(rows)

			mock.ExpectQuery("^SELECT count(.+) FROM").
				WillReturnRows(rowCount)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.repo+tt.query, nil)

			page, size := getPageAndSizeParams(req)

			params := Params{
				"repo":      "my-repo",
				"namespace": namespace,
				"page":      strconv.Itoa(page),
				"size":      strconv.Itoa(size),
			}

			listChartsWithFilters(w, req, params)

			assert.Equal(t, http.StatusOK, w.Code)

			var b bodyAPIListResponse
			json.NewDecoder(w.Body).Decode(&b)
			data := *b.Data
			assert.Len(t, data, len(tt.charts))
			for i, resp := range data {
				assert.Equal(t, tt.charts[i].ID, resp.ID, "In '"+tt.name+"': "+"chart id in the response should be the same")
				assert.Equal(t, "chart", resp.Type, "In '"+tt.name+"': "+"response type is chart")
				assert.Equal(t, tt.charts[i].ChartVersions[0].Version, resp.Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], "In '"+tt.name+"': "+"latestChartVersion should match version at index 0")
			}
			assert.Equal(t, tt.meta, b.Meta, "In '"+tt.name+"': "+"response meta should be the same")
		})
	}
}

func Test_getChart(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		chart    models.Chart
		wantCode int
	}{
		{
			"chart does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			nil,
			models.Chart{Repo: testRepo, Name: "foo%2Fmy-chart", ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{Repo: testRepo, Name: "my-chart", ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			chartID := tt.chart.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM charts").
				WithArgs(namespace, chartID)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				chartJSON, err := json.Marshal(tt.chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(chartJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID, nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"namespace": namespace,
				"repo":      parts[0],
				"chartName": parts[1],
			}

			getChart(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIResponse
				json.NewDecoder(w.Body).Decode(&b)
				assert.Equal(t, tt.chart.ID, b.Data.ID, "In '"+tt.name+"': "+"chart id in the response should be the same")
				assert.Equal(t, "chart", b.Data.Type, "In '"+tt.name+"': "+"response type is chart")
				assert.Equal(t, pathPrefix+"/ns/"+namespace+"/charts/"+tt.chart.ID, b.Data.Links.(map[string]interface{})["self"], "In '"+tt.name+"': "+"self link should be the same")
				assert.Equal(t, tt.chart.ChartVersions[0].Version, b.Data.Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], "In '"+tt.name+"': "+"latestChartVersion should match version at index 0")
			}
		})
	}
}

func Test_listChartVersions(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		chart    models.Chart
		wantCode int
	}{
		{
			"chart does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			chartID := tt.chart.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM charts").
				WithArgs(namespace, chartID)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				chartJSON, err := json.Marshal(tt.chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(chartJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID+"/versions", nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"namespace": namespace,
				"repo":      parts[0],
				"chartName": parts[1],
			}

			listChartVersions(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIListResponse
				json.NewDecoder(w.Body).Decode(&b)
				data := *b.Data
				for i, resp := range data {
					assert.Equal(t, resp.ID, tt.chart.ID+"-"+tt.chart.ChartVersions[i].Version, "In '"+tt.name+"': "+"chart id in the response should be the same")
					assert.Equal(t, resp.Type, "chartVersion", "In '"+tt.name+"': "+"response type is chartVersion")
					assert.Equal(t, resp.Attributes.(map[string]interface{})["version"], tt.chart.ChartVersions[i].Version, "In '"+tt.name+"': "+"chart version should match")
				}
			}
		})
	}
}

func Test_getChartVersion(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		chart    models.Chart
		wantCode int
	}{
		{
			"chart does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{Repo: testRepo, ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/foo%2Fmy-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			chartID := tt.chart.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM charts").
				WithArgs(namespace, chartID)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				chartJSON, err := json.Marshal(tt.chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(chartJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version, nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"namespace": namespace,
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   tt.chart.ChartVersions[0].Version,
			}

			getChartVersion(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIResponse
				json.NewDecoder(w.Body).Decode(&b)
				assert.Equal(t, b.Data.ID, tt.chart.ID+"-"+tt.chart.ChartVersions[0].Version, "In '"+tt.name+"': "+"chart id in the response should be the same")
				assert.Equal(t, b.Data.Type, "chartVersion", "In '"+tt.name+"': "+"response type is chartVersion")
				assert.Equal(t, b.Data.Attributes.(map[string]interface{})["version"], tt.chart.ChartVersions[0].Version, "In '"+tt.name+"': "+"chart version should match")
			}
		})
	}
}

func Test_getChartIcon(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		chart    models.Chart
		wantCode int
	}{
		{
			"chart does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			errors.New("return an error when checking if chart exists"),
			models.Chart{ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart has icon",
			nil,
			models.Chart{ID: "my-repo/my-chart", RawIcon: iconBytes(), IconContentType: "image/png"},
			http.StatusOK,
		},
		{
			"chart encoded has icon",
			nil,
			models.Chart{ID: "my-repo/foo%2Fmy-chart", RawIcon: iconBytes(), IconContentType: "image/png"},
			http.StatusOK,
		},
		{
			"chart does not have a icon",
			nil,
			models.Chart{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded does not have a icon",
			nil,
			models.Chart{ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart has icon with custom type",
			nil,
			models.Chart{ID: "my-repo/my-chart", RawIcon: iconBytes(), IconContentType: "image/svg"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			chartID := tt.chart.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM charts").
				WithArgs(namespace, chartID)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				chartJSON, err := json.Marshal(tt.chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(chartJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.chart.ID+"/logo", nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"namespace": namespace,
			}

			getChartIcon(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, tt.chart.RawIcon, w.Body.Bytes(), "raw icon data should match")
				assert.Equal(t, tt.chart.IconContentType, w.Header().Get("Content-Type"), "icon content type should match")
			}
		})
	}
}

func Test_getChartVersionReadme(t *testing.T) {
	chartName := "my-chart"
	chartEncodedName := "foo%2Fmy-chart"
	tests := []struct {
		name     string
		version  string
		err      error
		files    models.ChartFiles
		wantCode int
	}{
		{
			"chart does not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/" + chartName},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/" + chartEncodedName},
			http.StatusNotFound,
		},
		{
			"chart exists",
			"1.2.3",
			nil,
			models.ChartFiles{ID: "my-repo/" + chartName, Readme: testChartReadme},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			"1.2.3",
			nil,
			models.ChartFiles{ID: "my-repo/" + chartEncodedName, Readme: testChartReadme},
			http.StatusOK,
		},
		{
			"chart does not have a readme",
			"1.1.1",
			nil,
			models.ChartFiles{ID: "my-repo/" + chartName},
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			filesID := tt.files.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM files").
				WithArgs(namespace, filesID+"-0.1.0")

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				filesJSON, err := json.Marshal(tt.files)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(filesJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.files.ID+"/versions/"+tt.version+"/README.md", nil)
			parts := strings.Split(tt.files.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   "0.1.0",
				"namespace": namespace,
			}

			getChartVersionReadme(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, tt.files.Readme, string(w.Body.Bytes()), "content of the readme should match")
			}
		})
	}
}

func Test_getChartVersionValues(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		err      error
		files    models.ChartFiles
		wantCode int
	}{
		{
			"chart does not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded does not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			"3.2.1",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart", Values: testChartValues},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			"3.2.1",
			nil,
			models.ChartFiles{ID: "my-repo/foo%2Fmy-chart", Values: testChartValues},
			http.StatusOK,
		},
		{
			"chart does not have values.yaml",
			"2.2.2",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			filesID := tt.files.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM files").
				WithArgs(namespace, filesID+"-"+tt.version)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				filesJSON, err := json.Marshal(tt.files)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(filesJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.files.ID+"/versions/"+tt.version+"/values.yaml", nil)
			parts := strings.Split(tt.files.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   tt.version,
				"namespace": namespace,
			}

			getChartVersionValues(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, tt.files.Values, string(w.Body.Bytes()), "content of values.yaml should match")
			}
		})
	}
}

func Test_getChartVersionSchema(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		err      error
		files    models.ChartFiles
		wantCode int
	}{
		{
			"chart does not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart encoded not exist",
			"0.1.0",
			errors.New("return an error when checking if chart exists"),
			models.ChartFiles{ID: "my-repo/foo%2Fmy-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			"3.2.1",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart", Schema: testChartSchema},
			http.StatusOK,
		},
		{
			"chart encoded exists",
			"3.2.1",
			nil,
			models.ChartFiles{ID: "my-repo/foo%2Fmy-chart", Schema: testChartSchema},
			http.StatusOK,
		},
		{
			"chart does not have values.yaml",
			"2.2.2",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			filesID := tt.files.ID

			mockQuery := mock.ExpectQuery("SELECT info FROM files").
				WithArgs(namespace, filesID+"-"+tt.version)

			if tt.err != nil {
				mockQuery.WillReturnError(tt.err)
			} else {
				filesJSON, err := json.Marshal(tt.files)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mockQuery.WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(filesJSON))
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.files.ID+"/versions/"+tt.version+"/values.schema.json", nil)
			parts := strings.Split(tt.files.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   tt.version,
				"namespace": namespace,
			}

			getChartVersionSchema(w, req, params)

			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, tt.files.Schema, string(w.Body.Bytes()), "content of values.schema.json should match")
			}
		})
	}
}

func Test_findLatestChart(t *testing.T) {
	chart1 := &models.Chart{
		Name:        "foo",
		ID:          "foo",
		Category:    "cat1",
		Description: "best chart",
		Repo: &models.Repo{
			Name:      "bar",
			Namespace: kubeappsNamespace,
		},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0", AppVersion: "0.1.0"},
			{Version: "1.0.0", AppVersion: "whatever"},
			{Version: "whatever", AppVersion: "1.0.0"},
		},
	}
	chart2 := &models.Chart{
		Name:        "no-name",
		ID:          "no-id",
		Category:    "no-category",
		Description: "no-description",
		Repo: &models.Repo{
			Name:      "no-repo",
			Namespace: kubeappsNamespace,
		},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0", AppVersion: "whatever"},
			{Version: "whatever", AppVersion: "1.0.0"},
		},
	}
	chartDup1 := &models.Chart{
		Name: "foo",
		ID:   "stable/foo",
		Repo: &models.Repo{Name: "bar"},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0",
				AppVersion: "0.1.0",
				Digest:     "123",
			},
		},
	}
	chartDup2 := &models.Chart{
		Name: "foo",
		ID:   "bitnami/foo",
		Repo: &models.Repo{Name: "bar"},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0",
				AppVersion: "0.1.0",
				Digest:     "123",
			},
		},
	}
	tests := []struct {
		name           string
		charts         []*models.Chart
		chartName      string
		version        string
		appVersion     string
		repos          string
		categories     string
		query          string
		expectedCharts []*models.Chart
	}{
		{
			name: "returns no charts (by name)",
			charts: []*models.Chart{
				chart1,
			},
			chartName:      "nothing",
			version:        "",
			appVersion:     "",
			repos:          "",
			categories:     "",
			query:          "",
			expectedCharts: []*models.Chart{},
		},
		{
			name: "returns mocked chart (by name)",
			charts: []*models.Chart{
				chart1,
				chart2,
			},
			chartName:  "foo",
			version:    "",
			appVersion: "",
			repos:      "",
			categories: "",
			query:      "",
			expectedCharts: []*models.Chart{
				chart1,
			},
		},
		{
			name: "returns mocked chart (by version and appversion)",
			charts: []*models.Chart{
				chart1,
				chart2,
			},
			chartName:  "",
			version:    "1.0.0",
			appVersion: "0.1.0",
			repos:      "",
			categories: "",
			query:      "",
			expectedCharts: []*models.Chart{
				chart1,
			},
		},
		{
			name: "returns mocked chart (by repos)",
			charts: []*models.Chart{
				chart1,
				chart2,
			},
			chartName:  "",
			version:    "",
			appVersion: "",
			repos:      "bar",
			categories: "",
			query:      "",
			expectedCharts: []*models.Chart{
				chart1,
			},
		},
		{
			name: "returns mocked chart (by category)",
			charts: []*models.Chart{
				chart1,
				chart2,
			},
			chartName:  "",
			version:    "",
			appVersion: "",
			repos:      "",
			categories: "cat1",
			query:      "",
			expectedCharts: []*models.Chart{
				chart1,
			},
		},
		{
			name: "returns mocked chart (by query)",
			charts: []*models.Chart{
				chart1,
				chart2,
			},
			chartName:  "",
			version:    "",
			appVersion: "",
			repos:      "",
			categories: "",
			query:      "best",
			expectedCharts: []*models.Chart{
				chart1,
			},
		},
		{
			name: "includes duplicated charts",
			charts: []*models.Chart{
				chartDup1,
				chartDup2,
			},
			chartName:  "",
			version:    "",
			appVersion: "",
			repos:      "",
			categories: "",
			query:      "",
			expectedCharts: []*models.Chart{
				chartDup1,
				chartDup2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, cleanup := setMockManager(t)
			defer cleanup()
			mockQuery := mock.ExpectQuery("SELECT info FROM charts WHERE *")
			rows := sqlmock.NewRows([]string{"info"})
			for _, chart := range tt.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				queryMatch, _ := regexp.MatchString(tt.query, string(chartJSON))
				shouldAddRow := tt.chartName == chart.Name || (tt.version == chart.ChartVersions[0].Version && tt.appVersion == chart.ChartVersions[0].AppVersion) || tt.repos == chart.Repo.Name || tt.categories == chart.Category || (tt.query != "" && queryMatch)
				if shouldAddRow {
					rows.AddRow(chartJSON)
				}
			}
			if tt.chartName != "" {
				mockQuery.WithArgs(kubeappsNamespace, globalReposNamespace, tt.chartName)
			}
			if tt.version != "" && tt.appVersion != "" {
				parametrizedJsonbLiteral := fmt.Sprintf(`[{"version":"%s","app_version":"%s"}]`, tt.version, tt.appVersion)
				mockQuery.WithArgs(kubeappsNamespace, globalReposNamespace, parametrizedJsonbLiteral)
			}
			if tt.repos != "" {
				mockQuery.WithArgs(kubeappsNamespace, globalReposNamespace, tt.repos)
			}
			if tt.categories != "" {
				mockQuery.WithArgs(kubeappsNamespace, globalReposNamespace, tt.categories)
			}
			if tt.query != "" {
				mockQuery.WithArgs(kubeappsNamespace, globalReposNamespace, "%"+tt.query+"%")
			}

			mockQuery.WillReturnRows(rows)

			w := httptest.NewRecorder()
			requestURI := fmt.Sprintf("/charts?name=%s&version=%s&appversion=%s&repos=%s&categories=%s&q=%s", tt.chartName, tt.version, tt.appVersion, tt.repos, tt.categories, tt.query)
			req := httptest.NewRequest("GET", requestURI, nil)
			params := Params{
				"namespace":  kubeappsNamespace,
				"name":       tt.chartName,
				"version":    tt.version,
				"appversion": tt.appVersion,
				"repos":      tt.repos,
				"categories": tt.categories,
				"q":          tt.query,
			}

			listChartsWithFilters(w, req, params)

			var b bodyAPIListResponse
			json.NewDecoder(w.Body).Decode(&b)
			if b.Data == nil {
				t.Fatal("chart list shouldn't be null")
			}
			data := *b.Data

			if len(tt.expectedCharts) != len(data) {
				t.Errorf("Expecting %v charts, received %v in '%s'", len(tt.expectedCharts), len(data), tt.name)
			}

			for i := range data {
				assert.Equal(t, tt.charts[i].ID, data[i].ID, "In '"+tt.name+"': "+"chart id in the response should be the same")
				assert.Equal(t, tt.charts[i].ChartVersions[0].Version, data[i].Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], "In '"+tt.name+"': "+"latestChartVersion should match version at index 0")
			}

		})
	}
}
