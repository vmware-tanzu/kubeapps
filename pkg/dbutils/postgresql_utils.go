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

package dbutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
)

const (
	// ChartTable `create table charts (ID serial NOT NULL PRIMARY KEY, info jsonb NOT NULL);`
	ChartTable = "charts"
	// RepositoryTable `create table repos (ID serial NOT NULL PRIMARY KEY, name varchar unique, checksum varchar, last_update varchar);`
	RepositoryTable = "repos"
	// ChartFilesTable `create table files (ID serial NOT NULL PRIMARY KEY, chart_files_ID varchar unique, info jsonb NOT NULL);`
	ChartFilesTable = "files"
)

type postgresDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

// PostgresAssetManagerIface represents the methods of the PG asset manager
type PostgresAssetManagerIface interface {
	Init() error
	Close() error
	QueryOne(target interface{}, query string, args ...interface{}) error
	QueryAllCharts(query string, args ...interface{}) ([]*models.Chart, error)
}

// PostgresAssetManager asset manager for postgres
type PostgresAssetManager struct {
	connStr string
	DB      postgresDB
}

// NewPGManager creates an asset manager for PG
func NewPGManager(config datastore.Config) (*PostgresAssetManager, error) {
	url := strings.Split(config.URL, ":")
	if len(url) != 2 {
		return nil, fmt.Errorf("Can't parse database URL: %s", config.URL)
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		url[0], url[1], config.Username, config.Password, config.Database,
	)
	return &PostgresAssetManager{connStr, nil}, nil
}

// Init connects to PG
func (m *PostgresAssetManager) Init() error {
	db, err := sql.Open("postgres", m.connStr)
	if err != nil {
		return err
	}
	m.DB = db
	return nil
}

// Close connection
func (m *PostgresAssetManager) Close() error {
	return m.DB.Close()
}

// QueryOne element and store it in the target
func (m *PostgresAssetManager) QueryOne(target interface{}, query string, args ...interface{}) error {
	var info string
	row := m.DB.QueryRow(query, args...)
	err := row.Scan(&info)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(info), target)
}

// QueryAllCharts perform the given query and return the list of charts
func (m *PostgresAssetManager) QueryAllCharts(query string, args ...interface{}) ([]*models.Chart, error) {
	rows, err := m.DB.Query(query, args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	result := []*models.Chart{}
	for rows.Next() {
		var info string
		err := rows.Scan(&info)
		if err != nil {
			return nil, err
		}
		var chart models.Chart
		err = json.Unmarshal([]byte(info), &chart)
		if err != nil {
			return nil, err
		}
		result = append(result, &chart)
	}
	return result, nil
}
