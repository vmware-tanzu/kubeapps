// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/disintegration/imaging"
	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/kubeapps/cmd/assetsvc/pkg/utils"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
)

type bodyAPIListResponse struct {
	Data *apiListResponse `json:"data"`
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
