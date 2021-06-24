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
	"math"
	"strings"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	_ "github.com/lib/pq"
)

// TODO(mnelson): standardise error API for package.
var ErrChartVersionNotFound = errors.New("chart version not found")

// TODO(agamez): temporary flag, use the fallback behavior just when necessary, not globally
var enableFallbackQueryMode = true

type postgresAssetManager struct {
	dbutils.PostgresAssetManagerIface
}

func NewPGManager(config datastore.Config, kubeappsNamespace string) (AssetManager, error) {
	m, err := dbutils.NewPGManager(config, kubeappsNamespace)
	if err != nil {
		return nil, err
	}
	return &postgresAssetManager{m}, nil
}

func (m *postgresAssetManager) GetAllChartCategories(cq ChartQuery) ([]*models.ChartCategory, error) {
	whereQuery, whereQueryParams := m.GenerateWhereClause(cq)
	dbQuery := fmt.Sprintf("SELECT (info ->> 'category') AS name, COUNT( (info ->> 'category')) AS count FROM %s %s GROUP BY (info ->> 'category') ORDER BY (info ->> 'category') ASC", dbutils.ChartTable, whereQuery)

	chartsCategories, err := m.QueryAllChartCategories(dbQuery, whereQueryParams...)
	if err != nil {
		return nil, err
	}
	return chartsCategories, nil
}

func (m *postgresAssetManager) GetPaginatedChartList(whereQuery string, whereQueryParams []interface{}, pageNumber, pageSize int) ([]*models.Chart, int, error) {
	// Default (pageNumber,pageSize) = (1, 0) as in the handler.go
	if pageNumber <= 0 {
		pageNumber = 1
	}

	paginationClause := ""
	if pageSize > 0 {
		offset := (pageNumber - 1) * pageSize
		paginationClause = fmt.Sprintf("LIMIT %d OFFSET %d", pageSize, offset)
	}

	dbQuery := fmt.Sprintf("SELECT info FROM %s %s ORDER BY (info->>'name') ASC %s", dbutils.ChartTable, whereQuery, paginationClause)
	charts, err := m.QueryAllCharts(dbQuery, whereQueryParams...)
	if err != nil {
		return nil, 0, err
	}

	numPages := 1
	if pageSize > 0 {
		dbCountQuery := fmt.Sprintf("SELECT count(info) FROM %s %s", dbutils.ChartTable, whereQuery)
		count, err := m.QueryCount(dbCountQuery, whereQueryParams...)
		if err != nil {
			return nil, 0, err
		}
		numPages = int(math.Ceil(float64(count) / float64(pageSize)))
	}
	return charts, numPages, nil
}

func (m *postgresAssetManager) GetChart(namespace, chartID string) (models.Chart, error) {
	return m.GetChartWithFallback(namespace, chartID, enableFallbackQueryMode)
}

func (m *postgresAssetManager) GetChartWithFallback(namespace, chartID string, withFallback bool) (models.Chart, error) {
	var chart models.ChartIconString

	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		splittedID := strings.Split(chartID, "/")
		if withFallback == true && len(splittedID) == 2 {
			// fallback query when a chart_id is not being retrieved
			// it may occur when upgrading a mirrored chart (eg, jfrog/bitnami/wordpress)
			// and helms only gives 'jfrog/wordpress' but we want to retrieve 'jfrog/bitnami/wordpress'
			// this query search 'jfrog <whatever> wordpress'. If multiple results are found, returns just the first one
			alikeChartID := splittedID[0] + "%" + splittedID[1]
			err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id ILIKE $2", dbutils.ChartTable), namespace, alikeChartID)
			if err != nil {
				return models.Chart{}, err
			}
		} else {
			return models.Chart{}, err
		}
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

func (m *postgresAssetManager) GetChartVersion(namespace, chartID, version string) (models.Chart, error) {
	return m.GetChartVersionWithFallback(namespace, chartID, version, enableFallbackQueryMode)
}

func (m *postgresAssetManager) GetChartVersionWithFallback(namespace, chartID, version string, withFallback bool) (models.Chart, error) {

	var chart models.Chart
	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		splittedID := strings.Split(chartID, "/")
		if withFallback == true && len(splittedID) == 2 {
			// fallback query when a chart_id is not being retrieved
			// it may occur when upgrading a mirrored chart (eg, jfrog/bitnami/wordpress)
			// and helms only gives 'jfrog/wordpress' but we want to retrieve 'jfrog/bitnami/wordpress'
			// this query search 'jfrog <whatever> wordpress'. If multiple results are found, returns just the first one
			alikeChartID := splittedID[0] + "%" + splittedID[1]
			err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id ILIKE $2", dbutils.ChartTable), namespace, alikeChartID)
			if err != nil {
				return models.Chart{}, err
			}
		} else {
			return models.Chart{}, err
		}
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

func (m *postgresAssetManager) GetChartFiles(namespace, filesID string) (models.ChartFiles, error) {
	return m.GetChartFilesWithFallback(namespace, filesID, enableFallbackQueryMode)
}

func (m *postgresAssetManager) GetChartFilesWithFallback(namespace, filesID string, withFallback bool) (models.ChartFiles, error) {
	var chartFiles models.ChartFiles
	err := m.QueryOne(&chartFiles, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_files_id = $2", dbutils.ChartFilesTable), namespace, filesID)
	if err != nil {
		splittedID := strings.Split(filesID, "/")
		if withFallback == true && len(splittedID) == 2 {
			// fallback query when a chart_files_id is not being retrieved
			// it may occur when upgrading a mirrored chart (eg, jfrog/bitnami/wordpress)
			// and helms only gives 'jfrog/wordpress' but we want to retrieve 'jfrog/bitnami/wordpress'
			// this query search 'jfrog <whatever> wordpress'. If multiple results are found, returns just the first one
			alikeFilesID := splittedID[0] + "%" + splittedID[1]
			err := m.QueryOne(&chartFiles, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_files_id ILIKE $2", dbutils.ChartFilesTable), namespace, alikeFilesID)
			if err != nil {
				return models.ChartFiles{}, err
			}
		} else {
			return models.ChartFiles{}, err
		}
	}
	return chartFiles, nil
}

func (m *postgresAssetManager) GetPaginatedChartListWithFilters(cq ChartQuery, pageNumber, pageSize int) ([]*models.Chart, int, error) {
	whereQuery, whereQueryParams := m.GenerateWhereClause(cq)
	charts, numPages, err := m.GetPaginatedChartList(whereQuery, whereQueryParams, pageNumber, pageSize)
	if err != nil {
		return []*models.Chart{}, 0, err
	}
	return charts, numPages, nil
}

func (m *postgresAssetManager) GenerateWhereClause(cq ChartQuery) (string, []interface{}) {
	whereClauses := []string{}
	whereQueryParams := []interface{}{}
	whereQuery := ""

	if cq.namespace != dbutils.AllNamespaces {
		whereQueryParams = append(whereQueryParams, cq.namespace, m.GetKubeappsNamespace())
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(repo_namespace = $%d OR repo_namespace = $%d)", len(whereQueryParams)-1, len(whereQueryParams),
		))
	}
	if cq.chartName != "" {
		whereQueryParams = append(whereQueryParams, cq.chartName)
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(info->>'name' = $%d)", len(whereQueryParams),
		))
	}
	if cq.version != "" && cq.appVersion != "" {
		parametrizedJsonbLiteral := fmt.Sprintf(`[{"version":"%s","app_version":"%s"}]`, cq.version, cq.appVersion)
		whereQueryParams = append(whereQueryParams, parametrizedJsonbLiteral)
		whereClauses = append(whereClauses, fmt.Sprintf("(info->'chartVersions' @> $%d::jsonb)", len(whereQueryParams)))
	}

	if cq.repos != nil && len(cq.repos) > 0 {
		repoClauses := []string{}
		for _, repo := range cq.repos {
			if repo != "" {
				whereQueryParams = append(whereQueryParams, repo)
				repoClauses = append(repoClauses, fmt.Sprintf("(repo_name = $%d)", len(whereQueryParams)))
			}
		}
		if len(repoClauses) > 0 {
			repoQuery := "(" + strings.Join(repoClauses, " OR ") + ")"
			whereClauses = append(whereClauses, repoQuery)
		}
	}
	if cq.categories != nil && len(cq.categories) > 0 {
		categoryClauses := []string{}
		for _, category := range cq.categories {
			if category != "" {
				whereQueryParams = append(whereQueryParams, category)
				categoryClauses = append(categoryClauses, fmt.Sprintf("info->>'category' = $%d", len(whereQueryParams)))
			}
		}
		if len(categoryClauses) > 0 {
			categoryQuery := "(" + strings.Join(categoryClauses, " OR ") + ")"
			whereClauses = append(whereClauses, categoryQuery)
		}
	}
	if cq.searchQuery != "" {
		whereQueryParams = append(whereQueryParams, "%"+cq.searchQuery+"%")
		searchClause := fmt.Sprintf("((info ->> 'name' ILIKE $%d) OR ", len(whereQueryParams)) +
			fmt.Sprintf("(info ->> 'description' ILIKE $%d) OR ", len(whereQueryParams)) +
			fmt.Sprintf("(info -> 'repo' ->> 'name' ILIKE $%d) OR ", len(whereQueryParams)) +
			fmt.Sprintf("(info ->> 'keywords' ILIKE $%d) OR ", len(whereQueryParams)) +
			fmt.Sprintf("(info ->> 'sources' ILIKE $%d) OR ", len(whereQueryParams)) +
			fmt.Sprintf("(info -> 'maintainers' ->> 'name' ILIKE $%d))", len(whereQueryParams))
		whereClauses = append(whereClauses, searchClause)
	}
	if len(whereClauses) > 0 {
		whereQuery = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	return whereQuery, whereQueryParams
}
