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

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"image/color"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/disintegration/imaging"
	"github.com/kubeapps/common/datastore/mockstore"
	"github.com/kubeapps/kubeapps/cmd/chartsvc/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type bodyAPIListResponse struct {
	Data apiListResponse `json:"data"`
}

type bodyAPIResponse struct {
	Data apiResponse `json:"data"`
}

var chartsList []*models.Chart

const testChartReadme = "# Quickstart\n\n```bash\nhelm install my-repo/my-chart\n```"
const testChartValues = "image:\n  registry: docker.io\n  repository: my-repo/my-chart\n  tag: 0.1.0"

func iconBytes() []byte {
	var b bytes.Buffer
	img := imaging.New(1, 1, color.White)
	imaging.Encode(&b, img, imaging.PNG)
	return b.Bytes()
}

func Test_chartAttributes(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"chart has no icon", models.Chart{
			ID: "stable/wordpress",
		}},
		{"chart has a icon", models.Chart{
			ID: "repo/mychart", RawIcon: iconBytes(),
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chartAttributes(tt.chart)
			assert.Equal(t, tt.chart.ID, c.ID)
			assert.Equal(t, tt.chart.RawIcon, c.RawIcon)
			if len(tt.chart.RawIcon) == 0 {
				assert.Equal(t, len(c.Icon), 0, "icon url should be undefined")
			} else {
				assert.Equal(t, c.Icon, pathPrefix+"/assets/"+tt.chart.ID+"/logo-160x160-fit.png", "the icon url should be the same")
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
			ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := chartVersionAttributes(tt.chart.ID, tt.chart.ChartVersions[0])
			assert.Equal(t, cv.Version, tt.chart.ChartVersions[0].Version, "version string should be the same")
			assert.Equal(t, cv.Readme, pathPrefix+"/assets/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version+"/README.md", "README.md resource path should be the same")
			assert.Equal(t, cv.Values, pathPrefix+"/assets/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version+"/values.yaml", "values.yaml resource path should be the same")
		})
	}
}

func Test_newChartResponse(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"chart has only one version", models.Chart{
			ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}}},
		},
		{"chart has many versions", models.Chart{
			ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.2"}, {Version: "0.1.0"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cResponse := newChartResponse(&tt.chart)
			assert.Equal(t, cResponse.Type, "chart", "response type is chart")
			assert.Equal(t, cResponse.ID, tt.chart.ID, "chart ID should be the same")
			assert.Equal(t, cResponse.Links.(selfLink).Self, pathPrefix+"/charts/"+tt.chart.ID, "self link should be the same")
			assert.Equal(t, cResponse.Attributes.(models.Chart).ChartVersions, tt.chart.ChartVersions, "chart version in the response should be the same")
		})
	}
}

func Test_newChartListResponse(t *testing.T) {
	tests := []struct {
		name   string
		charts []*models.Chart
	}{
		{"no charts", []*models.Chart{}},
		{"has one chart", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}}},
		}},
		{"has two charts", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}}},
			{ID: "stable/wordpress", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}, {Version: "1.2.2"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clResponse := newChartListResponse(tt.charts)
			assert.Equal(t, len(clResponse), len(tt.charts), "number of charts in response should be the same")
			for i := range tt.charts {
				assert.Equal(t, clResponse[i].Type, "chart", "response type is chart")
				assert.Equal(t, clResponse[i].ID, tt.charts[i].ID, "chart ID should be the same")
				assert.Equal(t, clResponse[i].Links.(selfLink).Self, pathPrefix+"/charts/"+tt.charts[i].ID, "self link should be the same")
				assert.Equal(t, clResponse[i].Attributes.(models.Chart).ChartVersions, tt.charts[i].ChartVersions, "chart version in the response should be the same")
			}
		})
	}
}

func Test_newChartVersionResponse(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"my-chart", models.Chart{
			ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.2.3"}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := range tt.chart.ChartVersions {
				cvResponse := newChartVersionResponse(&tt.chart, tt.chart.ChartVersions[i])
				assert.Equal(t, cvResponse.Type, "chartVersion", "response type is chartVersion")
				assert.Equal(t, cvResponse.ID, tt.chart.ID+"-"+tt.chart.ChartVersions[i].Version, "reponse id should have chart version suffix")
				assert.Equal(t, cvResponse.Links.(interface{}).(selfLink).Self, pathPrefix+"/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[i].Version, "self link should be the same")
				assert.Equal(t, cvResponse.Attributes.(models.ChartVersion).Version, tt.chart.ChartVersions[i].Version, "chart version in the response should be the same")
			}
		})
	}
}

func Test_newChartVersionListResponse(t *testing.T) {
	tests := []struct {
		name  string
		chart models.Chart
	}{
		{"my-chart", models.Chart{
			ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}, {Version: "0.0.2"}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cvListResponse := newChartVersionListResponse(&tt.chart)
			assert.Equal(t, len(cvListResponse), len(tt.chart.ChartVersions), "number of chart versions in response should be the same")
			for i := range tt.chart.ChartVersions {
				assert.Equal(t, cvListResponse[i].Type, "chartVersion", "response type is chartVersion")
				assert.Equal(t, cvListResponse[i].ID, tt.chart.ID+"-"+tt.chart.ChartVersions[i].Version, "reponse id should have chart version suffix")
				assert.Equal(t, cvListResponse[i].Links.(interface{}).(selfLink).Self, pathPrefix+"/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[i].Version, "self link should be the same")
				assert.Equal(t, cvListResponse[i].Attributes.(models.ChartVersion).Version, tt.chart.ChartVersions[i].Version, "chart version in the response should be the same")
			}
		})
	}
}

func Test_listCharts(t *testing.T) {
	tests := []struct {
		name   string
		charts []*models.Chart
	}{
		{"no charts", []*models.Chart{}},
		{"one chart", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}}},
		}},
		{"two charts", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}}},
			{ID: "stable/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}, {Version: "1.2.2"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			m.On("All", &chartsList).Run(func(args mock.Arguments) {
				*args.Get(0).(*[]*models.Chart) = tt.charts
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts", nil)
			listCharts(w, req)

			m.AssertExpectations(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var b bodyAPIListResponse
			json.NewDecoder(w.Body).Decode(&b)
			assert.Len(t, b.Data, len(tt.charts))
			for i := range b.Data {
				assert.Equal(t, b.Data[i].ID, tt.charts[i].ID, "chart id in the response should be the same")
				assert.Equal(t, b.Data[i].Type, "chart", "response type is chart")
				assert.Equal(t, b.Data[i].Links.(map[string]interface{})["self"], pathPrefix+"/charts/"+tt.charts[i].ID, "self link should be the same")
				assert.Equal(t, b.Data[i].Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], tt.charts[i].ChartVersions[0].Version, "version should match latest chart version")
			}
		})
	}
}

func Test_listRepoCharts(t *testing.T) {
	tests := []struct {
		name   string
		repo   string
		charts []*models.Chart
	}{
		{"repo has no charts", "my-repo", []*models.Chart{}},
		{"repo has one chart", "my-repo", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
		}},
		{"repo has many charts", "my-repo", []*models.Chart{
			{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			{ID: "my-repo/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3"}, {Version: "1.2.2"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			m.On("All", &chartsList).Run(func(args mock.Arguments) {
				*args.Get(0).(*[]*models.Chart) = tt.charts
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.repo, nil)
			params := Params{
				"repo": "my-repo",
			}

			listRepoCharts(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, http.StatusOK, w.Code)

			var b bodyAPIListResponse
			json.NewDecoder(w.Body).Decode(&b)
			assert.Len(t, b.Data, len(tt.charts))
			for i := range b.Data {
				assert.Equal(t, b.Data[i].ID, tt.charts[i].ID, "chart id in the response should be the same")
				assert.Equal(t, b.Data[i].Type, "chart", "response type is chart")
				assert.Equal(t, b.Data[i].Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], tt.charts[i].ChartVersions[0].Version, "version should match latest chart version")
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
			models.Chart{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID, nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
			}

			getChart(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIResponse
				json.NewDecoder(w.Body).Decode(&b)
				assert.Equal(t, b.Data.ID, tt.chart.ID, "chart id in the response should be the same")
				assert.Equal(t, b.Data.Type, "chart", "response type is chart")
				assert.Equal(t, b.Data.Links.(map[string]interface{})["self"], pathPrefix+"/charts/"+tt.chart.ID, "self link should be the same")
				assert.Equal(t, b.Data.Relationships["latestChartVersion"].Data.(map[string]interface{})["version"], tt.chart.ChartVersions[0].Version, "version should match latest chart version")
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
			models.Chart{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID+"/versions", nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
			}

			listChartVersions(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIListResponse
				json.NewDecoder(w.Body).Decode(&b)
				for i := range b.Data {
					assert.Equal(t, b.Data[i].ID, tt.chart.ID+"-"+tt.chart.ChartVersions[i].Version, "chart id in the response should be the same")
					assert.Equal(t, b.Data[i].Type, "chartVersion", "response type is chartVersion")
					assert.Equal(t, b.Data[i].Attributes.(map[string]interface{})["version"], tt.chart.ChartVersions[i].Version, "chart version should match")
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
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusNotFound,
		},
		{
			"chart exists",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
			http.StatusOK,
		},
		{
			"chart has multiple versions",
			nil,
			models.Chart{ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}, {Version: "0.0.1"}}},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/charts/"+tt.chart.ID+"/versions/"+tt.chart.ChartVersions[0].Version, nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   tt.chart.ChartVersions[0].Version,
			}

			getChartVersion(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var b bodyAPIResponse
				json.NewDecoder(w.Body).Decode(&b)
				assert.Equal(t, b.Data.ID, tt.chart.ID+"-"+tt.chart.ChartVersions[0].Version, "chart id in the response should be the same")
				assert.Equal(t, b.Data.Type, "chartVersion", "response type is chartVersion")
				assert.Equal(t, b.Data.Attributes.(map[string]interface{})["version"], tt.chart.ChartVersions[0].Version, "chart version should match")
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
			"chart has icon",
			nil,
			models.Chart{ID: "my-repo/my-chart", RawIcon: iconBytes()},
			http.StatusOK,
		},
		{
			"chart does not have a icon",
			nil,
			models.Chart{ID: "my-repo/my-chart"},
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.chart.ID+"/logo-160x160-fit.png", nil)
			parts := strings.Split(tt.chart.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
			}

			getChartIcon(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, w.Body.Bytes(), tt.chart.RawIcon, "raw icon data should match")
			}
		})
	}
}

func Test_getChartVersionReadme(t *testing.T) {
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
			"chart exists",
			"1.2.3",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart", Readme: testChartReadme},
			http.StatusOK,
		},
		{
			"chart does not have a readme",
			"1.1.1",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.ChartFiles{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.ChartFiles) = tt.files
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.files.ID+"/versions/"+tt.version+"/README.md", nil)
			parts := strings.Split(tt.files.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   "0.1.0",
			}

			getChartVersionReadme(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, string(w.Body.Bytes()), tt.files.Readme, "content of the readme should match")
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
			"chart exists",
			"3.2.1",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart", Values: testChartValues},
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
			var m mock.Mock
			dbSession = mockstore.NewMockSession(&m)

			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.ChartFiles{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.ChartFiles) = tt.files
				})
			}

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/assets/"+tt.files.ID+"/versions/"+tt.version+"/values.yaml", nil)
			parts := strings.Split(tt.files.ID, "/")
			params := Params{
				"repo":      parts[0],
				"chartName": parts[1],
				"version":   "0.1.0",
			}

			getChartVersionValues(w, req, params)

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, w.Code, "http status code should match")
			if tt.wantCode == http.StatusOK {
				assert.Equal(t, string(w.Body.Bytes()), tt.files.Values, "content of values.yaml should match")
			}
		})
	}
}
