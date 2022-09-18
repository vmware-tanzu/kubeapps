// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package dbutils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
)

const (
	// ChartTable table containing Charts info
	ChartTable = "charts"
	// RepositoryTable table containing repositories sync info
	RepositoryTable = "repos"
	// ChartFilesTable table containing files related to other charts
	ChartFilesTable = "files"
	// EnvvarPostgresTests enables tests that run against a local postgres
	EnvvarPostgresTests = "ENABLE_PG_INTEGRATION_TESTS"
)

type PostgresDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// PostgresAssetManagerIface represents the methods of the PG asset manager
// The interface is used by the tests to implement a fake PostgresAssetManagerIface
type PostgresAssetManagerIface interface {
	AssetManager
	QueryCount(query string, args ...interface{}) (int, error)
	QueryOne(target interface{}, query string, args ...interface{}) error
	QueryAllCharts(query string, args ...interface{}) ([]*models.Chart, error)
	QueryAllChartCategories(query string, args ...interface{}) ([]*models.ChartCategory, error)
	InitTables() error
	InvalidateCache() error
	EnsureRepoExists(repoNamespace, repoName string) (int, error)
	GetDB() PostgresDB
	GetGlobalPackagingNamespace() string
}

// PostgresAssetManager asset manager for postgres
type PostgresAssetManager struct {
	connStr                  string
	DB                       PostgresDB
	GlobalPackagingNamespace string
}

// NewPGManager creates an asset manager for PG
func NewPGManager(config Config, globalPackagingNamespace string) (*PostgresAssetManager, error) {
	url := strings.Split(config.URL, ":")
	if len(url) != 2 {
		return nil, fmt.Errorf("Can't parse database URL: %s", config.URL)
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		url[0], url[1], config.Username, config.Password, config.Database,
	)
	return &PostgresAssetManager{connStr, nil, globalPackagingNamespace}, nil
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

// QueryAllChartCategories performs the query and return the array of all the chart categories
func (m *PostgresAssetManager) QueryAllChartCategories(query string, args ...interface{}) ([]*models.ChartCategory, error) {
	rows, err := m.DB.Query(query, args...)
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return nil, err
	}
	result := []*models.ChartCategory{}
	for rows.Next() {
		var name string
		var count int
		err := rows.Scan(&name, &count)
		if err != nil {
			return nil, err
		}
		chartCategory := models.ChartCategory{Name: name, Count: count}
		result = append(result, &chartCategory)
	}
	return result, nil
}

// QueryCount count the returned results from a given query
func (m *PostgresAssetManager) QueryCount(query string, args ...interface{}) (int, error) {
	var count int
	row := m.DB.QueryRow(query, args...)
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// InitTables creates the required tables for the postgresql backend for assets.
func (m *PostgresAssetManager) InitTables() error {
	// Repository table should have a namespace column, and chart table should reference repositories.
	_, err := m.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	ID serial NOT NULL PRIMARY KEY,
	namespace varchar NOT NULL,
	name varchar NOT NULL,
	checksum varchar,
	last_update varchar,
	UNIQUE(namespace, name)
)`, RepositoryTable))
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	ID serial NOT NULL PRIMARY KEY,
	repo_name varchar NOT NULL,
	repo_namespace varchar NOT NULL,
	chart_id varchar,
	info jsonb NOT NULL,
	UNIQUE(repo_name, repo_namespace, chart_id),
	FOREIGN KEY (repo_name, repo_namespace) REFERENCES %s (name, namespace) ON DELETE CASCADE
)`, ChartTable, RepositoryTable))
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	ID serial NOT NULL PRIMARY KEY,
	chart_id varchar NOT NULL,
	repo_name varchar NOT NULL,
	repo_namespace varchar NOT NULL,
	chart_files_ID varchar NOT NULL,
	info jsonb NOT NULL,
	UNIQUE(repo_namespace, chart_files_ID),
	FOREIGN KEY (repo_name, repo_namespace) REFERENCES %s (name, namespace) ON DELETE CASCADE,
	FOREIGN KEY (repo_name, repo_namespace, chart_id) REFERENCES %s (repo_name, repo_namespace, chart_id) ON DELETE CASCADE
)`, ChartFilesTable, RepositoryTable, ChartTable))
	if err != nil {
		return err
	}
	return nil
}

// InvalidateCache for postgresql deletes and re-writes the schema
func (m *PostgresAssetManager) InvalidateCache() error {
	tables := strings.Join([]string{RepositoryTable, ChartTable, ChartFilesTable}, ",")
	_, err := m.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tables))
	if err != nil {
		return err
	}

	return m.InitTables()
}

// EnsureRepoExists upserts to get the primary key of a repo.
func (m *PostgresAssetManager) EnsureRepoExists(repoNamespace, repoName string) (int, error) {
	// The only query I could find for inserting a new repo or selecting the existing one
	// to find the ID in a single query.
	query := fmt.Sprintf(`
WITH new_repo AS (
	INSERT INTO %s (namespace, name)
	SELECT CAST($1 AS VARCHAR), CAST($2 AS VARCHAR) WHERE NOT EXISTS (
		SELECT * FROM %s WHERE namespace=$1 AND name=$2)
	RETURNING ID
)
SELECT ID FROM new_repo
UNION
SELECT ID FROM %s WHERE namespace=$1 AND name=$2
`, RepositoryTable, RepositoryTable, RepositoryTable)

	var id int
	err := m.DB.QueryRow(query, repoNamespace, repoName).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *PostgresAssetManager) GetDB() PostgresDB {
	return m.DB
}

func (m *PostgresAssetManager) GetGlobalPackagingNamespace() string {
	return m.GlobalPackagingNamespace
}
