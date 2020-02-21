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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	_ "github.com/lib/pq"
)

type postgresAssetManager struct {
	*dbutils.PostgresAssetManager
}

func newPGManager(config datastore.Config) (assetManager, error) {
	m, err := dbutils.NewPGManager(config)
	if err != nil {
		return nil, err
	}
	return &postgresAssetManager{m}, nil
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
func (m *postgresAssetManager) Sync(repo models.RepoInternal, charts []models.Chart) error {
	m.initTables()

	// Ensure the repo exists so FK constraints will be met.
	_, err := m.ensureRepoExists(repo.Namespace, repo.Name)
	if err != nil {
		return err
	}

	err = m.importCharts(charts, models.Repo{Namespace: repo.Namespace, Name: repo.Name})
	if err != nil {
		return err
	}

	// Remove charts no longer existing in index
	return m.removeMissingCharts(charts)
}

func (m *postgresAssetManager) initTables() error {
	// Repository table should have a namespace column, and chart table should reference repositories.
	_, err := m.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	ID serial NOT NULL PRIMARY KEY,
	namespace varchar NOT NULL,
	name varchar NOT NULL,
	checksum varchar,
	last_update varchar,
	UNIQUE(namespace, name)
)`, dbutils.RepositoryTable))
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
)`, dbutils.ChartTable, dbutils.RepositoryTable))
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	ID serial NOT NULL PRIMARY KEY,
	repo_name varchar NOT NULL,
	repo_namespace varchar NOT NULL,
	chart_files_ID varchar NOT NULL,
	info jsonb NOT NULL,
	UNIQUE(repo_namespace, chart_files_ID),
	FOREIGN KEY (repo_name, repo_namespace) REFERENCES %s (name, namespace) ON DELETE CASCADE
)`, dbutils.ChartFilesTable, dbutils.RepositoryTable))
	if err != nil {
		return err
	}
	return nil
}

func (m *postgresAssetManager) RepoAlreadyProcessed(repoName, repoChecksum string) bool {
	var lastChecksum string
	row := m.DB.QueryRow(fmt.Sprintf("SELECT checksum FROM %s WHERE name = $1", dbutils.RepositoryTable), repoName)
	if row != nil {
		err := row.Scan(&lastChecksum)
		return err == nil && lastChecksum == repoChecksum
	}
	return false
}

// ensureRepoExists upserts to get the primary key of a repo.
func (m *postgresAssetManager) ensureRepoExists(repoNamespace, repoName string) (int, error) {
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
`, dbutils.RepositoryTable, dbutils.RepositoryTable, dbutils.RepositoryTable)

	var id int
	err := m.DB.QueryRow(query, repoNamespace, repoName).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *postgresAssetManager) UpdateLastCheck(repoNamespace, repoName, checksum string, now time.Time) error {
	query := fmt.Sprintf(`INSERT INTO %s (namespace, name, checksum, last_update)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (namespace, name)
	DO UPDATE SET last_update = $4, checksum = $3
	`, dbutils.RepositoryTable)
	rows, err := m.DB.Query(query, repoNamespace, repoName, checksum, now.String())
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) importCharts(charts []models.Chart, repo models.Repo) error {
	for _, chart := range charts {
		d, err := json.Marshal(chart)
		if err != nil {
			return err
		}
		_, err = m.DB.Exec(fmt.Sprintf(`INSERT INTO %s (repo_namespace, repo_name, chart_id, info)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chart_id, repo_namespace, repo_name)
		DO UPDATE SET info = $4
		`, dbutils.ChartTable), repo.Namespace, repo.Name, chart.ID, string(d))
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *postgresAssetManager) removeMissingCharts(charts []models.Chart) error {
	var chartIDs []string
	for _, chart := range charts {
		chartIDs = append(chartIDs, fmt.Sprintf("'%s'", chart.ID))
	}
	chartIDsString := strings.Join(chartIDs, ", ")
	rows, err := m.DB.Query(fmt.Sprintf("DELETE FROM %s WHERE info ->> 'ID' NOT IN (%s) AND info -> 'repo' ->> 'name' = $1", dbutils.ChartTable, chartIDsString), charts[0].Repo.Name)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) Delete(repo models.Repo) error {
	rows, err := m.DB.Query(fmt.Sprintf("DELETE FROM %s WHERE name = $1 AND namespace = $2", dbutils.RepositoryTable), repo.Name, repo.Namespace)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) updateIcon(data []byte, contentType, ID string) error {
	rows, err := m.DB.Query(fmt.Sprintf(
		`UPDATE charts SET info = info || '{"raw_icon": "%s", "icon_content_type": "%s"}'  WHERE info ->> 'ID' = '%s'`,
		base64.StdEncoding.EncodeToString(data), contentType, ID,
	))
	if rows != nil {
		rows.Close()
	}
	return err
}

func (m *postgresAssetManager) filesExist(chartFilesID, digest string) bool {
	rows, err := m.DB.Query(
		fmt.Sprintf("SELECT * FROM %s WHERE chart_files_id = $1 AND info ->> 'Digest' = $2", dbutils.ChartFilesTable),
		chartFilesID,
		digest,
	)
	hasEntries := false
	if rows != nil {
		defer rows.Close()
		hasEntries = rows.Next()
	}
	return err == nil && hasEntries
}

func (m *postgresAssetManager) insertFiles(chartFilesID string, files models.ChartFiles) error {
	if files.Repo == nil {
		return fmt.Errorf("unable to insert file without repo: %q", files.ID)
	}
	query := fmt.Sprintf(`INSERT INTO %s (repo_name, repo_namespace, chart_files_ID, info)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (repo_namespace, chart_files_ID)
	DO UPDATE SET info = $4
	`, dbutils.ChartFilesTable)
	rows, err := m.DB.Query(query, files.Repo.Name, files.Repo.Namespace, chartFilesID, files)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

// InvalidateCache for postgresql deletes and re-writes the schema
func (m *postgresAssetManager) InvalidateCache() error {
	tables := strings.Join([]string{dbutils.RepositoryTable, dbutils.ChartTable, dbutils.ChartFilesTable}, ",")
	_, err := m.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tables))
	if err != nil {
		return err
	}

	return m.initTables()
}
