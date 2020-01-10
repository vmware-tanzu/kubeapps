/*
Copyright (c) Bitnami

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
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/stretchr/testify/mock"
)

type fakePGManager struct {
	*mock.Mock
}

func (f *fakePGManager) Init() error {
	return nil
}

func (f *fakePGManager) Close() error {
	return nil
}

func (f *fakePGManager) QueryOne(target interface{}, query string, args ...interface{}) error {
	f.Called(target, query, args)
	return nil
}

var chartsResponse []*models.Chart

func (f *fakePGManager) QueryAllCharts(query string, args ...interface{}) ([]*models.Chart, error) {
	f.Called(query, args)
	return chartsResponse, nil
}

func Test_NewPGManager(t *testing.T) {
	config := datastore.Config{URL: "10.11.12.13:5432"}
	_, err := newPGManager(config)
	if err != nil {
		t.Errorf("Found error %v", err)
	}
}

func Test_PGgetChart(t *testing.T) {
	m := &mock.Mock{}
	fpg := &fakePGManager{m}
	pg := postgresAssetManager{fpg}

	icon := []byte("test")
	iconB64 := base64.StdEncoding.EncodeToString(icon)
	dbChart := models.ChartIconString{
		Chart:   models.Chart{ID: "foo"},
		RawIcon: iconB64,
	}
	m.On("QueryOne", &models.ChartIconString{}, "SELECT info FROM charts WHERE info ->> 'ID' = $1", []interface{}{"foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.ChartIconString) = dbChart
	})

	chart, err := pg.getChart("foo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedChart := models.Chart{
		ID:      "foo",
		RawIcon: icon,
	}
	if !cmp.Equal(chart, expectedChart) {
		t.Errorf("Unexpected result %v", cmp.Diff(chart, expectedChart))
	}
}

func Test_PGgetChartVersion(t *testing.T) {
	m := &mock.Mock{}
	fpg := &fakePGManager{m}
	pg := postgresAssetManager{fpg}

	dbChart := models.Chart{
		ID: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0"},
			{Version: "2.0.0"},
		},
	}
	m.On("QueryOne", &models.Chart{}, "SELECT info FROM charts WHERE info ->> 'ID' = $1", []interface{}{"foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.Chart) = dbChart
	})

	chart, err := pg.getChartVersion("foo", "1.0.0")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedChart := models.Chart{
		ID: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0"},
		},
	}
	if !cmp.Equal(chart, expectedChart) {
		t.Errorf("Unexpected result %v", cmp.Diff(chart, expectedChart))
	}
}

func Test_getChartFiles(t *testing.T) {
	m := &mock.Mock{}
	fpg := &fakePGManager{m}
	pg := postgresAssetManager{fpg}

	expectedFiles := models.ChartFiles{ID: "foo"}
	m.On("QueryOne", &models.ChartFiles{}, "SELECT info FROM files WHERE chart_files_id = $1", []interface{}{"foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.ChartFiles) = expectedFiles
	})

	files, err := pg.getChartFiles("foo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	if !cmp.Equal(files, expectedFiles) {
		t.Errorf("Unexpected result %v", cmp.Diff(files, expectedFiles))
	}
}

func Test_getChartWithFilters(t *testing.T) {
	m := &mock.Mock{}
	fpg := &fakePGManager{m}
	pg := postgresAssetManager{fpg}

	dbChart := models.Chart{
		Name: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "2.0.0", AppVersion: "2.0.2"},
			{Version: "1.0.0", AppVersion: "1.0.1"},
		},
	}
	chartsResponse = []*models.Chart{&dbChart}
	m.On("QueryAllCharts", "SELECT info FROM charts WHERE info ->> 'name' = $1", []interface{}{"foo"})

	charts, err := pg.getChartsWithFilters("foo", "1.0.0", "1.0.1")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedCharts := []*models.Chart{&models.Chart{
		Name: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "2.0.0", AppVersion: "2.0.2"},
		},
	}}
	if !cmp.Equal(charts, expectedCharts) {
		t.Errorf("Unexpected result %v", cmp.Diff(charts, expectedCharts))
	}
}

func Test_getPaginatedChartList(t *testing.T) {
	availableCharts := []*models.Chart{
		{ID: "foo", ChartVersions: []models.ChartVersion{{Digest: "123"}}},
		{ID: "bar", ChartVersions: []models.ChartVersion{{Digest: "456"}}},
		{ID: "copyFoo", ChartVersions: []models.ChartVersion{{Digest: "123"}}},
	}
	tests := []struct {
		name               string
		repo               string
		pageNumber         int
		pageSize           int
		showDuplicates     bool
		expectedCharts     []*models.Chart
		expectedTotalPages int
	}{
		{"one page with duplicates with repo", "bitnami", 1, 100, true, availableCharts, 1},
		{"one page withuot duplicates", "", 1, 100, false, []*models.Chart{availableCharts[0], availableCharts[1]}, 1},
		// TODO(andresmgot): several pages
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mock.Mock{}
			fpg := &fakePGManager{m}
			pg := postgresAssetManager{fpg}

			chartsResponse = availableCharts
			expectedQuery := ""
			if tt.repo != "" {
				expectedQuery = fmt.Sprintf("WHERE info -> 'repo' ->> 'name' = '%s'", tt.repo)
			}
			expectedQuery = fmt.Sprintf("SELECT info FROM %s %s ORDER BY info ->> 'name' ASC", dbutils.ChartTable, expectedQuery)
			m.On("QueryAllCharts", expectedQuery, []interface{}(nil))
			charts, totalPages, err := pg.getPaginatedChartList(tt.repo, tt.pageNumber, tt.pageSize, tt.showDuplicates)
			if err != nil {
				t.Errorf("Found error %v", err)
			}
			if totalPages != tt.expectedTotalPages {
				t.Errorf("Unexpected number of pages, got %d expecting %d", totalPages, tt.expectedTotalPages)
			}
			if !cmp.Equal(charts, tt.expectedCharts) {
				t.Errorf("Unexpected result %v", cmp.Diff(charts, tt.expectedCharts))
			}
		})
	}
}

func Test_searchCharts(t *testing.T) {
	tests := []struct {
		name          string
		query         string
		repo          string
		expectedQuery string
	}{
		{"without repo", "foo", "", "SELECT info FROM charts WHERE  (info ->> 'name' ~ $1) OR (info ->> 'description' ~ $1) OR (info -> 'repo' ->> 'name' ~ $1) OR (info ->> 'keywords' ~ $1)OR (info ->> 'sources' ~ $1)OR (info ->> 'maintainers' ~ $1)"},
		{"with repo", "foo", "bar", "SELECT info FROM charts WHERE info -> 'repo' ->> 'name' = 'bar' AND (info ->> 'name' ~ $1) OR (info ->> 'description' ~ $1) OR (info -> 'repo' ->> 'name' ~ $1) OR (info ->> 'keywords' ~ $1)OR (info ->> 'sources' ~ $1)OR (info ->> 'maintainers' ~ $1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mock.Mock{}
			fpg := &fakePGManager{m}
			pg := postgresAssetManager{fpg}

			m.On("QueryAllCharts", tt.expectedQuery, []interface{}{tt.query})
			pg.searchCharts(tt.query, tt.repo)
			m.AssertExpectations(t)
		})
	}
}
