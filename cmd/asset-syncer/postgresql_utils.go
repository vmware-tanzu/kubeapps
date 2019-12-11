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
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kubeapps/common/datastore"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	// create table charts (ID serial NOT NULL PRIMARY KEY, info jsonb NOT NULL);
	chartTable = "charts"
	// create table repos (ID serial NOT NULL PRIMARY KEY, name varchar unique, checksum varchar, last_update varchar);
	repositoryTable = "repos"
	// create table files (ID serial NOT NULL PRIMARY KEY, chart_files_ID varchar unique, info jsonb NOT NULL);
	chartFilesTable = "files"
)

type postgresDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

type postgresAssetManager struct {
	connStr string
	db      postgresDB
}

func newPGManager(config datastore.Config) (assetManager, error) {
	url := strings.Split(config.URL, ":")
	if len(url) != 2 {
		return nil, fmt.Errorf("Can't parse database URL: %s", config.URL)
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		url[0], url[1], config.Username, config.Password, config.Database,
	)
	return &postgresAssetManager{connStr, nil}, nil
}

func (m *postgresAssetManager) Init() error {
	db, err := sql.Open("postgres", m.connStr)
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *postgresAssetManager) Close() error {
	return m.db.Close()
}

// Syncing is performed in the following steps:
// 1. Update database to match chart metadata from index
// 2. Concurrently process icons for charts (concurrently)
// 3. Concurrently process the README and values.yaml for the latest chart version of each chart
// 4. Concurrently process READMEs and values.yaml for historic chart versions
//
// These steps are processed in this way to ensure relevant chart data is
// imported into the database as fast as possible. E.g. we want all icons for
// charts before fetching readmes for each chart and version pair.
func (m *postgresAssetManager) Sync(charts []chart) error {
	err := m.importCharts(charts)
	if err != nil {
		return err
	}

	// Remove charts no longer existing in index
	return m.removeMissingCharts(charts)
}

func (m *postgresAssetManager) RepoAlreadyProcessed(repoName, repoChecksum string) bool {
	var lastChecksum string
	row := m.db.QueryRow(fmt.Sprintf("SELECT checksum FROM %s WHERE name = $1", repositoryTable), repoName)
	if row != nil {
		err := row.Scan(&lastChecksum)
		return err == nil && lastChecksum == repoChecksum
	}
	return false
}

func (m *postgresAssetManager) UpdateLastCheck(repoName, checksum string, now time.Time) error {
	query := fmt.Sprintf(`INSERT INTO %s (name, checksum, last_update)
	VALUES ($1, $2, $3)
	ON CONFLICT (name) 
	DO UPDATE SET last_update = $3, checksum = $2
	`, repositoryTable)
	rows, err := m.db.Query(query, repoName, checksum, now.String())
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) importCharts(charts []chart) error {
	txn, err := m.db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := txn.Prepare(pq.CopyIn(chartTable, "info"))
	if err != nil {
		return err
	}

	for _, chart := range charts {
		d, err := json.Marshal(chart)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(string(d))
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (m *postgresAssetManager) removeMissingCharts(charts []chart) error {
	var chartIDs []string
	for _, chart := range charts {
		chartIDs = append(chartIDs, fmt.Sprintf("'%s'", chart.ID))
	}
	chartIDsString := strings.Join(chartIDs, ", ")
	rows, err := m.db.Query(fmt.Sprintf("DELETE FROM %s WHERE info ->> 'ID' NOT IN (%s)", chartTable, chartIDsString))
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) Delete(repoName string) error {
	tables := []string{chartTable, chartFilesTable}
	for _, table := range tables {
		rows, err := m.db.Query(fmt.Sprintf("DELETE FROM %s WHERE info -> 'repo' ->> 'name' = $1", table), repoName)
		if rows != nil {
			defer rows.Close()
		}
		if err != nil {
			return err
		}
	}
	rows, err := m.db.Query(fmt.Sprintf("DELETE FROM %s WHERE name = $1", repositoryTable), repoName)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) updateIcon(data []byte, contentType, ID string) error {
	rows, err := m.db.Query(fmt.Sprintf(
		`UPDATE charts SET info = info || '{"raw_icon": "%s", "icon_content_type": "%s"}'  WHERE info ->> 'ID' = '%s'`,
		base64.StdEncoding.EncodeToString(data), contentType, ID,
	))
	if rows != nil {
		rows.Close()
	}
	return err
}

func (m *postgresAssetManager) filesExist(chartFilesID, digest string) bool {
	rows, err := m.db.Query("SELECT * FROM files WHERE info -> 'ID' = $1 AND info -> 'digest' = $2", chartFilesID, digest)
	if rows != nil {
		defer rows.Close()
	}
	return err == nil
}

func (m *postgresAssetManager) insertFiles(chartFilesID string, files chartFiles) error {
	query := fmt.Sprintf(`INSERT INTO %s (chart_files_ID, info)
	VALUES ($1, $2)
	ON CONFLICT (chart_files_ID) 
	DO UPDATE SET info = $2
	`, chartFilesTable)
	rows, err := m.db.Query(query, chartFilesID, files)
	if rows != nil {
		defer rows.Close()
	}
	return err
}
