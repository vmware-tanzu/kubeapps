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
	"errors"
	"fmt"
	"strings"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	_ "github.com/lib/pq"
)

// TODO(mnelson): standardise error API for package.
var ErrChartVersionNotFound = errors.New("chart version not found")

type postgresAssetManager struct {
	dbutils.PostgresAssetManagerIface
}

func newPGManager(config datastore.Config, kubeappsNamespace string) (assetManager, error) {
	m, err := dbutils.NewPGManager(config, kubeappsNamespace)
	if err != nil {
		return nil, err
	}
	return &postgresAssetManager{m}, nil
}

func exists(current []string, str string) bool {
	for _, s := range current {
		if s == str {
			return true
		}
	}
	return false
}

func (m *postgresAssetManager) getPaginatedChartList(namespace, repo string, pageNumber, pageSize int, showDuplicates bool) ([]*models.Chart, int, error) {
	clauses := []string{}
	queryParams := []interface{}{}
	if namespace != dbutils.AllNamespaces {
		queryParams = append(queryParams, namespace, m.GetKubeappsNamespace())
		clauses = append(clauses, "(repo_namespace = $1 OR repo_namespace = $2)")
	}
	if repo != "" {
		queryParams = append(queryParams, repo)
		clauses = append(clauses, fmt.Sprintf("repo_name = $%d", len(queryParams)))
	}
	repoQuery := ""
	if len(clauses) > 0 {
		repoQuery = strings.Join(clauses, " AND ")
		repoQuery = "WHERE " + repoQuery
	}
	dbQuery := fmt.Sprintf("SELECT info FROM %s %s ORDER BY info ->> 'name' ASC", dbutils.ChartTable, repoQuery)
	charts, err := m.QueryAllCharts(dbQuery, queryParams...)
	if err != nil {
		return nil, 0, nil
	}
	if !showDuplicates {
		// Group by unique digest for the latest version (remove duplicates)
		uniqueCharts := []*models.Chart{}
		digests := []string{}
		for _, c := range charts {
			if len(c.ChartVersions) == 0 {
				return nil, 0, fmt.Errorf("chart %q missing chart versions", c.ID)
			}
			if !exists(digests, c.ChartVersions[0].Digest) {
				digests = append(digests, c.ChartVersions[0].Digest)
				uniqueCharts = append(uniqueCharts, c)
			}
		}
		// TODO(andresmgot): Implement pagination but currently Kubeapps don't support it
		return uniqueCharts, 1, nil
	}
	return charts, 1, nil
}

func (m *postgresAssetManager) getChart(namespace, chartID string) (models.Chart, error) {
	var chart models.ChartIconString
	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		return models.Chart{}, err
	}

	// TODO(andresmgot): Store raw_icon as a byte array
	icon, err := base64.StdEncoding.DecodeString(chart.RawIcon)
	if err != nil {
		return models.Chart{}, err
	}
	return models.Chart{
		ID:              chart.ID,
		Name:            chart.Name,
		Repo:            chart.Repo,
		Description:     chart.Description,
		Home:            chart.Home,
		Keywords:        chart.Keywords,
		Maintainers:     chart.Maintainers,
		Sources:         chart.Sources,
		Icon:            chart.Icon,
		RawIcon:         icon,
		IconContentType: chart.IconContentType,
		Category:        chart.Category,
		ChartVersions:   chart.ChartVersions,
	}, nil
}

func (m *postgresAssetManager) getChartVersion(namespace, chartID, version string) (models.Chart, error) {
	var chart models.Chart
	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		return models.Chart{}, err
	}
	found := false
	for _, c := range chart.ChartVersions {
		if c.Version == version {
			chart.ChartVersions = []models.ChartVersion{c}
			found = true
			break
		}
	}
	if !found {
		return models.Chart{}, ErrChartVersionNotFound
	}
	return chart, nil
}

func (m *postgresAssetManager) getChartFiles(namespace, filesID string) (models.ChartFiles, error) {
	var chartFiles models.ChartFiles
	err := m.QueryOne(&chartFiles, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_files_id = $2", dbutils.ChartFilesTable), namespace, filesID)
	if err != nil {
		return models.ChartFiles{}, err
	}
	return chartFiles, nil
}

func containsVersionAndAppVersion(chartVersions []models.ChartVersion, version, appVersion string) (models.ChartVersion, bool) {
	for _, ch := range chartVersions {
		if ch.Version == version && ch.AppVersion == appVersion {
			return ch, true
		}
	}
	return models.ChartVersion{}, false
}

func (m *postgresAssetManager) getChartsWithFilters(namespace, name, version, appVersion string) ([]*models.Chart, error) {
	clauses := []string{"info ->> 'name' = $1"}
	queryParams := []interface{}{name, namespace}
	if namespace != dbutils.AllNamespaces {
		queryParams = append(queryParams, m.GetKubeappsNamespace())
		clauses = append(clauses, "(repo_namespace = $2 OR repo_namespace = $3)")
	} else {
		clauses = append(clauses, "repo_namespace = $2")
	}
	dbQuery := fmt.Sprintf("SELECT info FROM %s WHERE %s ORDER BY info ->> 'ID' ASC", dbutils.ChartTable, strings.Join(clauses, " AND "))
	charts, err := m.QueryAllCharts(dbQuery, queryParams...)
	if err != nil {
		return nil, err
	}
	result := []*models.Chart{}
	for _, c := range charts {
		if _, found := containsVersionAndAppVersion(c.ChartVersions, version, appVersion); found {
			result = append(result, c)
		}
	}
	return result, nil
}
