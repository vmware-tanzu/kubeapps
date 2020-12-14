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
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
)

func getMockManager(t *testing.T) (*postgresAssetManager, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	pgManager := &postgresAssetManager{&dbutils.PostgresAssetManager{DB: db, KubeappsNamespace: "kubeapps"}}

	return pgManager, mock, func() { db.Close() }
}

func Test_PGgetChart(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	icon := []byte("test")
	iconB64 := base64.StdEncoding.EncodeToString(icon)
	dbChart := models.ChartIconString{
		Chart:   models.Chart{ID: "foo"},
		RawIcon: iconB64,
	}
	dbChartJSON, err := json.Marshal(dbChart)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	mock.ExpectQuery("SELECT info FROM charts*").
		WithArgs("namespace", "foo").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(string(dbChartJSON)))

	chart, err := pgManager.getChart("namespace", "foo")
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
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	dbChart := models.Chart{
		ID: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0"},
			{Version: "2.0.0"},
		},
	}
	dbChartJSON, err := json.Marshal(dbChart)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	mock.ExpectQuery("SELECT info FROM charts*").
		WithArgs("namespace", "foo").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(string(dbChartJSON)))

	chart, err := pgManager.getChartVersion("namespace", "foo", "1.0.0")
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
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	expectedFiles := models.ChartFiles{ID: "foo"}
	filesJSON, err := json.Marshal(expectedFiles)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	mock.ExpectQuery("SELECT info FROM files*").
		WithArgs("namespace", "foo").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(string(filesJSON)))

	files, err := pgManager.getChartFiles("namespace", "foo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	if !cmp.Equal(files, expectedFiles) {
		t.Errorf("Unexpected result %v", cmp.Diff(files, expectedFiles))
	}
}

func Test_getChartFiles_withSlashes(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	expectedFiles := models.ChartFiles{ID: "fo%2Fo"}
	filesJSON, err := json.Marshal(expectedFiles)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	mock.ExpectQuery("SELECT info FROM files*").
		WithArgs("namespace", "fo%2Fo").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(string(filesJSON)))

	files, err := pgManager.getChartFiles("namespace", "fo%2Fo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	if !cmp.Equal(files, expectedFiles) {
		t.Errorf("Unexpected result %v", cmp.Diff(files, expectedFiles))
	}
}

func Test_getChartsWithFilters(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	dbChart := models.Chart{
		Name: "foo",
		ChartVersions: []models.ChartVersion{
			{Version: "2.0.0", AppVersion: "2.0.2"},
			{Version: "1.0.0", AppVersion: "1.0.1"},
		},
	}
	dbChartJSON, err := json.Marshal(dbChart)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	mock.ExpectQuery("SELECT info FROM charts WHERE info*").
		WithArgs("foo", "namespace", "kubeapps").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(dbChartJSON))

	charts, err := pgManager.getChartsWithFilters("namespace", "foo", "1.0.0", "1.0.1")
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

func Test_getChartsWithFilters_withSlashes(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	dbChart := models.Chart{
		Name: "fo%2Fo",
		ChartVersions: []models.ChartVersion{
			{Version: "2.0.0", AppVersion: "2.0.2"},
			{Version: "1.0.0", AppVersion: "1.0.1"},
		},
	}
	dbChartJSON, err := json.Marshal(dbChart)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	mock.ExpectQuery("SELECT info FROM charts WHERE info*").
		WithArgs("fo%2Fo", "namespace", "kubeapps").
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(dbChartJSON))

	charts, err := pgManager.getChartsWithFilters("namespace", "fo%2Fo", "1.0.0", "1.0.1")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedCharts := []*models.Chart{&models.Chart{
		Name: "fo%2Fo",
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
		{ID: "fo%2Fo", ChartVersions: []models.ChartVersion{{Digest: "321"}}},
		{ID: "bar", ChartVersions: []models.ChartVersion{{Digest: "456"}}},
		{ID: "copyFoo", ChartVersions: []models.ChartVersion{{Digest: "123"}}},
	}
	tests := []struct {
		name               string
		namespace          string
		repo               string
		pageNumber         int
		pageSize           int
		expectedCharts     []*models.Chart
		expectedTotalPages int
	}{
		{
			name:               "one page with duplicates with repo",
			namespace:          "other-namespace",
			repo:               "bitnami",
			pageNumber:         1,
			pageSize:           100,
			expectedCharts:     availableCharts,
			expectedTotalPages: 1,
		},
		{
			name:               "one page with duplicates",
			namespace:          "other-namespace",
			repo:               "",
			pageNumber:         1,
			pageSize:           100,
			expectedCharts:     availableCharts,
			expectedTotalPages: 1,
		},
		// TODO(andresmgot): several pages
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgManager, mock, cleanup := getMockManager(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})
			rowCount := sqlmock.NewRows([]string{"count"}).AddRow(len(availableCharts))

			for _, chart := range availableCharts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			expectedParams := []driver.Value{"other-namespace", "kubeapps"}
			if tt.repo != "" {
				expectedParams = append(expectedParams, "bitnami")
			}
			mock.ExpectQuery("SELECT info FROM *").
				WithArgs(expectedParams...).
				WillReturnRows(rows)

			mock.ExpectQuery("^SELECT count(.+) FROM").
				WillReturnRows(rowCount)

			charts, totalPages, err := pgManager.getPaginatedChartList(tt.namespace, tt.repo, tt.pageNumber, tt.pageSize)
			if err != nil {
				t.Fatalf("Found error %v", err)
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
