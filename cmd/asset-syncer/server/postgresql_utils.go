// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	log "k8s.io/klog/v2"
)

var ErrMultipleRows = fmt.Errorf("more than one row returned in query result")

type postgresAssetManager struct {
	*dbutils.PostgresAssetManager
}

func newPGManager(config dbutils.Config, globalPackagingNamespace string) (assetManager, error) {
	m, err := dbutils.NewPGManager(config, globalPackagingNamespace)
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
func (m *postgresAssetManager) Sync(repo models.AppRepository, chart models.Chart) error {
	_, err := m.EnsureRepoExists(repo.Namespace, repo.Name)
	if err != nil {
		return err
	}

	d, err := json.Marshal(chart)
	if err != nil {
		return err
	}
	_, err = m.DB.Exec(fmt.Sprintf(`INSERT INTO %s (repo_namespace, repo_name, chart_id, info)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chart_id, repo_namespace, repo_name)
		DO UPDATE SET info = $4
		`, dbutils.ChartTable), repo.Namespace, repo.Name, chart.ID, string(d))
	return err
}

func (m *postgresAssetManager) LastChecksum(repo models.AppRepository) string {
	var lastChecksum string
	row := m.DB.QueryRow(fmt.Sprintf("SELECT checksum FROM %s WHERE name = $1 AND namespace = $2", dbutils.RepositoryTable), repo.Name, repo.Namespace)
	if row != nil {
		err := row.Scan(&lastChecksum)
		if err != nil {
			return ""
		}
	}
	return lastChecksum
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

func (m *postgresAssetManager) RemoveMissingCharts(repo models.AppRepository, chartNames []string) error {
	if len(chartNames) == 0 {
		log.V(4).Infof("No synced charts missing from repository. Nothing to remove.")
		return nil
	}
	var quotedChartNames []string
	for _, chartName := range chartNames {
		quotedChartNames = append(quotedChartNames, fmt.Sprintf("'%s'", chartName))
	}
	chartNamesString := strings.Join(quotedChartNames, ", ")
	log.V(4).Infof("Removing the following charts that are no longer present in the repo: %s", chartNamesString)
	rows, err := m.DB.Query(fmt.Sprintf("DELETE FROM %s WHERE info->>'name' IN (%s) AND repo_name = $1 AND repo_namespace = $2", dbutils.ChartTable, chartNamesString), repo.Name, repo.Namespace)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) Delete(repo models.AppRepository) error {
	rows, err := m.DB.Query(fmt.Sprintf("DELETE FROM %s WHERE name = $1 AND namespace = $2", dbutils.RepositoryTable), repo.Name, repo.Namespace)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

func (m *postgresAssetManager) updateIcon(repo models.AppRepository, data []byte, contentType, ID string) error {
	rows, err := m.DB.Query(fmt.Sprintf(
		`UPDATE charts SET info = info || '{"raw_icon": "%s", "icon_content_type": "%s"}' WHERE chart_id = $1 AND repo_namespace = $2 AND repo_name = $3 RETURNING ID`,
		base64.StdEncoding.EncodeToString(data), contentType,
	), ID, repo.Namespace, repo.Name)
	if rows != nil {
		defer rows.Close()
		var id int
		if !rows.Next() {
			return sql.ErrNoRows
		}
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		if rows.Next() {
			return fmt.Errorf("more than one icon updated for chart id %q: %w", ID, ErrMultipleRows)
		}
	}
	return err
}

func (m *postgresAssetManager) filesExist(repo models.AppRepository, chartFilesID, digest string) bool {
	var exists bool
	err := m.DB.QueryRow(
		fmt.Sprintf(`
	SELECT EXISTS(
		SELECT 1 FROM %s
		WHERE chart_files_id = $1 AND
			repo_name = $2 AND
			repo_namespace = $3 AND
			info ->> 'Digest' = $4
		)`, dbutils.ChartFilesTable),
		chartFilesID, repo.Name, repo.Namespace, digest).Scan(&exists)
	return err == nil && exists
}

func (m *postgresAssetManager) insertFiles(chartId string, files models.ChartFiles) error {
	if files.Repo == nil {
		return fmt.Errorf("unable to insert file without repo: %q", files.ID)
	}
	query := fmt.Sprintf(`INSERT INTO %s (chart_id, repo_name, repo_namespace, chart_files_ID, info)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (repo_namespace, chart_files_ID)
	DO UPDATE SET info = $5
	`, dbutils.ChartFilesTable)
	rows, err := m.DB.Query(query, chartId, files.Repo.Name, files.Repo.Namespace, files.ID, files)
	if rows != nil {
		defer rows.Close()
	}
	return err
}

// ChartsForRepo returns a map of charts with all previously synced charts in
// the repo.
func (m *postgresAssetManager) ChartsForRepo(repo models.AppRepository) (map[string]*models.Chart, error) {
	dbQuery := fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND repo_name = $2 ORDER BY (info->>'name')", dbutils.ChartTable)

	charts, err := m.QueryAllCharts(dbQuery, repo.Namespace, repo.Name)
	if err != nil {
		return nil, err
	}

	chartsByName := map[string]*models.Chart{}
	for _, chart := range charts {
		chartsByName[chart.Name] = chart
	}
	return chartsByName, nil
}
