/*
Copyright (c) 2017 The Helm Authors

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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// tests the GET /live endpoint
func Test_GetLive(t *testing.T) {
	var m mock.Mock
	manager = getMockManager(&m)

	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/live")
	assert.NoError(t, err, "should not return an error")
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK, "http status code should match")
}

// tests the GET /ready endpoint
func Test_GetReady(t *testing.T) {
	var m mock.Mock
	manager = getMockManager(&m)

	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

	res, err := http.Get(ts.URL + "/ready")
	assert.NoError(t, err, "should not return an error")
	defer res.Body.Close()
	assert.Equal(t, res.StatusCode, http.StatusOK, "http status code should match")
}

// tests the GET /{apiVersion}/ns/{namespace}/charts endpoint
func Test_GetCharts(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

	tests := []struct {
		name   string
		charts []*models.Chart
	}{
		{"no charts", []*models.Chart{}},
		{"one chart", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1"}}}},
		},
		{"two charts", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "my-repo/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			manager = getMockManager(&m)
			m.On("All", &chartsList).Run(func(args mock.Arguments) {
				*args.Get(0).(*[]*models.Chart) = tt.charts
			})

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/charts")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, http.StatusOK, "http status code should match")

			var b bodyAPIListResponse
			json.NewDecoder(res.Body).Decode(&b)
			assert.Len(t, *b.Data, len(tt.charts))
		})
	}
}

// tests the GET /{apiVersion}/ns/{namespace}/charts/{repo} endpoint
func Test_GetChartsInRepo(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

	tests := []struct {
		name   string
		repo   string
		charts []*models.Chart
	}{
		{"repo has no charts", "my-repo", []*models.Chart{}},
		{"repo has one chart", "my-repo", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
		}},
		{"repo has many charts", "my-repo", []*models.Chart{
			{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.0.1", Digest: "123"}}},
			{Repo: testRepo, ID: "my-repo/dokuwiki", ChartVersions: []models.ChartVersion{{Version: "1.2.3", Digest: "1234"}, {Version: "1.2.2", Digest: "12345"}}},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			manager = getMockManager(&m)
			m.On("All", &chartsList).Run(func(args mock.Arguments) {
				*args.Get(0).(*[]*models.Chart) = tt.charts
			})

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/charts/" + tt.repo)
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, http.StatusOK, "http status code should match")

			var b bodyAPIListResponse
			json.NewDecoder(res.Body).Decode(&b)
			assert.Len(t, *b.Data, len(tt.charts))
		})
	}
}

// tests the GET /{apiVersion}/ns/charts/{repo}/{chartName} endpoint
func Test_GetChartInRepo(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
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
			var m mock.Mock
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/charts/" + tt.chart.ID)
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/charts/{repo}/{chartName}/versions endpoint
func Test_ListChartVersions(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
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
			var m mock.Mock
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/charts/" + tt.chart.ID + "/versions")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/charts/{repo}/{chartName}/versions/{:version} endpoint
func Test_GetChartVersion(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			"chart exists",
			nil,
			models.Chart{Repo: testRepo, ID: "my-repo/my-chart", ChartVersions: []models.ChartVersion{{Version: "0.1.0"}}},
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
			var m mock.Mock
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/charts/" + tt.chart.ID + "/versions/" + tt.chart.ChartVersions[0].Version)
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/assets/{repo}/{chartName}/logo-160x160-fit.png endpoint
func Test_GetChartIcon(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.Chart{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.Chart) = tt.chart
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/assets/" + tt.chart.ID + "/logo")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/assets/{repo}/{chartName}/versions/{version}/README.md endpoint
func Test_GetChartReadme(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.ChartFiles{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.ChartFiles) = tt.files
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/assets/" + tt.files.ID + "/versions/" + tt.version + "/README.md")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, tt.wantCode, res.StatusCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/assets/{repo}/{chartName}/versions/{version}/values.yaml endpoint
func Test_GetChartValues(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.ChartFiles{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.ChartFiles) = tt.files
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/assets/" + tt.files.ID + "/versions/" + tt.version + "/values.yaml")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}

// tests the GET /{apiVersion}/ns/assets/{repo}/{chartName}/versions/{version}/values/schema.json endpoint
func Test_GetChartSchema(t *testing.T) {
	ts := httptest.NewServer(setupRoutes())
	defer ts.Close()

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
			models.ChartFiles{ID: "my-repo/my-chart", Schema: testChartSchema},
			http.StatusOK,
		},
		{
			"chart does not have values.schema.json",
			"2.2.2",
			nil,
			models.ChartFiles{ID: "my-repo/my-chart"},
			http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m mock.Mock
			manager = getMockManager(&m)
			if tt.err != nil {
				m.On("One", mock.Anything).Return(tt.err)
			} else {
				m.On("One", &models.ChartFiles{}).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(0).(*models.ChartFiles) = tt.files
				})
			}

			res, err := http.Get(ts.URL + pathPrefix + "/ns/kubeapps/assets/" + tt.files.ID + "/versions/" + tt.version + "/values.schema.json")
			assert.NoError(t, err)
			defer res.Body.Close()

			m.AssertExpectations(t)
			assert.Equal(t, res.StatusCode, tt.wantCode, "http status code should match")
		})
	}
}
