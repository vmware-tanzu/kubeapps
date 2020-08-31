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

func (f *fakePGManager) InvalidateCache() error {
	return nil
}

func (f *fakePGManager) InitTables() error {
	return nil
}

func (f *fakePGManager) EnsureRepoExists(namespace, name string) (int, error) {
	return 0, nil
}

func (f *fakePGManager) GetDB() dbutils.PostgresDB {
	return nil
}

func (f *fakePGManager) GetKubeappsNamespace() string {
	return "kubeapps"
}

func Test_NewPGManager(t *testing.T) {
	config := datastore.Config{URL: "10.11.12.13:5432"}
	_, err := newPGManager(config, "kubeapps")
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
	m.On("QueryOne", &models.ChartIconString{}, "SELECT info FROM charts WHERE repo_namespace = $1 AND chart_id = $2", []interface{}{"namespace", "foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.ChartIconString) = dbChart
	})

	chart, err := pg.getChart("namespace", "foo")
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
	m.On("QueryOne", &models.Chart{}, "SELECT info FROM charts WHERE repo_namespace = $1 AND chart_id = $2", []interface{}{"namespace", "foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.Chart) = dbChart
	})

	chart, err := pg.getChartVersion("namespace", "foo", "1.0.0")
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
	m.On("QueryOne", &models.ChartFiles{}, "SELECT info FROM files WHERE repo_namespace = $1 AND chart_files_id = $2", []interface{}{"namespace", "foo"}).Run(func(args mock.Arguments) {
		*args.Get(0).(*models.ChartFiles) = expectedFiles
	})

	files, err := pg.getChartFiles("namespace", "foo")
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
	m.On("QueryAllCharts", "SELECT info FROM charts WHERE info ->> 'name' = $1 AND (repo_namespace = $2 OR repo_namespace = $3) ORDER BY info ->> 'ID' ASC", []interface{}{"foo", "namespace", "kubeapps"})

	charts, err := pg.getChartsWithFilters("namespace", "foo", "1.0.0", "1.0.1")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedCharts := []*models.Chart{&models.Chart{
		Name: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "2.0.0", AppVersion: "2.0.2"},
			{Version: "1.0.0", AppVersion: "1.0.1"},
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
		namespace          string
		repo               string
		pageNumber         int
		pageSize           int
		showDuplicates     bool
		expectedCharts     []*models.Chart
		expectedTotalPages int
	}{
		{
			name:               "one page with duplicates with repo",
			namespace:          "other-namespace",
			repo:               "bitnami",
			pageNumber:         1,
			pageSize:           100,
			showDuplicates:     true,
			expectedCharts:     availableCharts,
			expectedTotalPages: 1,
		},
		{
			name:               "one page withuot duplicates",
			namespace:          "other-namespace",
			repo:               "",
			pageNumber:         1,
			pageSize:           100,
			showDuplicates:     false,
			expectedCharts:     []*models.Chart{availableCharts[0], availableCharts[1]},
			expectedTotalPages: 1,
		},
		// TODO(andresmgot): several pages
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mock.Mock{}
			fpg := &fakePGManager{m}
			pg := postgresAssetManager{fpg}

			chartsResponse = availableCharts
			expectedQuery := "WHERE (repo_namespace = $1 OR repo_namespace = $2)"
			expectedParams := []interface{}{"other-namespace", "kubeapps"}
			if tt.repo != "" {
				expectedQuery = expectedQuery + " AND repo_name = $3"
				expectedParams = append(expectedParams, "bitnami")
			}
			expectedQuery = fmt.Sprintf("SELECT info FROM %s %s ORDER BY info ->> 'name' ASC", dbutils.ChartTable, expectedQuery)
			m.On("QueryAllCharts", expectedQuery, expectedParams)
			charts, totalPages, err := pg.getPaginatedChartList(tt.namespace, tt.repo, tt.pageNumber, tt.pageSize, tt.showDuplicates)
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
