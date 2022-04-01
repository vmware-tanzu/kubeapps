// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package dbutils

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
)

func Test_NewPGManager(t *testing.T) {
	config := Config{URL: "10.11.12.13:5432", Database: "assets", Username: "postgres", Password: "123"}
	m, err := NewPGManager(config, "kubeapps")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedConnStr := "host=10.11.12.13 port=5432 user=postgres password=123 dbname=assets sslmode=disable"
	if m.connStr != expectedConnStr {
		t.Errorf("Expected %s got %s", expectedConnStr, m.connStr)
	}
}

func Test_Close(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	manager := PostgresAssetManager{
		connStr: "localhost",
		DB:      db,
	}
	mock.ExpectClose()
	err = manager.Close()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func Test_QueryOne(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	manager := PostgresAssetManager{
		connStr: "localhost",
		DB:      db,
	}
	query := "SELECT * from charts"
	rows := sqlmock.NewRows([]string{"info"}).AddRow(`{"ID": "foo"}`)
	mock.ExpectQuery("^SELECT (.+)$").WillReturnRows(rows)
	target := models.Chart{}
	err = manager.QueryOne(&target, query)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	expectedChart := models.Chart{ID: "foo"}
	if !cmp.Equal(target, expectedChart) {
		t.Errorf("Unexpected result %v", cmp.Diff(target, expectedChart))
	}
}

func Test_QueryAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	manager := PostgresAssetManager{
		connStr: "localhost",
		DB:      db,
	}
	query := "SELECT * from charts"
	rows := sqlmock.NewRows([]string{"info"}).
		AddRow(`{"ID": "foo"}`).
		AddRow(`{"ID": "bar"}`)
	mock.ExpectQuery("^SELECT (.+)$").WillReturnRows(rows)
	charts, err := manager.QueryAllCharts(query)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	expectedCharts := []*models.Chart{{ID: "foo"}, {ID: "bar"}}
	if !cmp.Equal(charts, expectedCharts) {
		t.Errorf("Unexpected result %v", cmp.Diff(charts, expectedCharts))
	}

}

func Test_QueryAllChartCategories(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	manager := PostgresAssetManager{
		connStr: "localhost",
		DB:      db,
	}
	query := "SELECT (info ->> 'category') AS name, COUNT( (info ->> 'category')) AS count FROM charts WHERE (repo_namespace = 'kubeapps' OR repo_namespace = 'default') GROUP BY (info ->> 'category') ORDER BY (info ->> 'category') ASC"
	rows := sqlmock.NewRows([]string{"name", "count"}).
		AddRow("cat1", 1).
		AddRow("cat2", 2).
		AddRow("cat3", 3)
	mock.ExpectQuery("SELECT (info ->> 'category')*").WillReturnRows(rows)
	chartCategories, err := manager.QueryAllChartCategories(query)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	expectedChartCategories := []*models.ChartCategory{
		{Name: "cat1", Count: 1},
		{Name: "cat2", Count: 2},
		{Name: "cat3", Count: 3},
	}
	if !cmp.Equal(chartCategories, expectedChartCategories) {
		t.Errorf("Unexpected result %v", cmp.Diff(chartCategories, expectedChartCategories))
	}
}
