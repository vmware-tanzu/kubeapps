// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
)

const AllNamespaces = "_all"

// TODO(mnelson): standardise error API for package.
var ErrChartVersionNotFound = errors.New("chart version not found")

// TODO(agamez): temporary flag, use the fallback behavior just when necessary, not globally
var enableFallbackQueryMode = true

type PostgresAssetManager struct {
	dbutils.PostgresAssetManagerIface
}

func NewPGManager(config dbutils.Config, globalPackagingNamespace string) (AssetManager, error) {
	m, err := dbutils.NewPGManager(config, globalPackagingNamespace)
	if err != nil {
		return nil, err
	}
	return &PostgresAssetManager{m}, nil
}

func (m *PostgresAssetManager) GetAllChartCategories(cq ChartQuery) ([]*models.ChartCategory, error) {
	whereQuery, whereQueryParams, err := m.GenerateWhereClause(cq)
	if err != nil {
		return nil, err
	}
	dbQuery := fmt.Sprintf("SELECT (info ->> 'category') AS name, COUNT( (info ->> 'category')) AS count FROM %s %s GROUP BY (info ->> 'category') ORDER BY (info ->> 'category') ASC", dbutils.ChartTable, whereQuery)

	chartsCategories, err := m.QueryAllChartCategories(dbQuery, whereQueryParams...)
	if err != nil {
		return nil, err
	}
	return chartsCategories, nil
}

func (m *PostgresAssetManager) GetPaginatedChartList(whereQuery string, whereQueryParams []interface{}, startItemNumber, pageSize int) ([]*models.Chart, error) {
	paginationClause := ""
	if pageSize > 0 {
		paginationClause = fmt.Sprintf("LIMIT %d", pageSize)
	}
	if startItemNumber > 0 {
		paginationClause = fmt.Sprintf("%s OFFSET %d", paginationClause, startItemNumber)
	}

	dbQuery := fmt.Sprintf("SELECT info FROM %s %s ORDER BY (info->>'name') ASC %s", dbutils.ChartTable, whereQuery, paginationClause)
	charts, err := m.QueryAllCharts(dbQuery, whereQueryParams...)
	if err != nil {
		return nil, err
	}

	return charts, nil
}

func (m *PostgresAssetManager) GetChart(namespace, chartID string) (models.Chart, error) {
	return m.GetChartWithFallback(namespace, chartID, enableFallbackQueryMode)
}

func (m *PostgresAssetManager) GetChartWithFallback(namespace, chartID string, withFallback bool) (models.Chart, error) {
	var chart models.ChartIconString

	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		splittedID := strings.Split(chartID, "/")
		if withFallback && len(splittedID) == 2 {
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

func (m *PostgresAssetManager) GetChartVersion(namespace, chartID, version string) (models.Chart, error) {
	return m.GetChartVersionWithFallback(namespace, chartID, version, enableFallbackQueryMode)
}

func (m *PostgresAssetManager) GetChartVersionWithFallback(namespace, chartID, version string, withFallback bool) (models.Chart, error) {

	var chart models.Chart
	err := m.QueryOne(&chart, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_id = $2", dbutils.ChartTable), namespace, chartID)
	if err != nil {
		splittedID := strings.Split(chartID, "/")
		if withFallback && len(splittedID) == 2 {
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

func (m *PostgresAssetManager) GetChartFiles(namespace, filesID string) (models.ChartFiles, error) {
	return m.GetChartFilesWithFallback(namespace, filesID, enableFallbackQueryMode)
}

func (m *PostgresAssetManager) GetChartFilesWithFallback(namespace, filesID string, withFallback bool) (models.ChartFiles, error) {
	var chartFiles models.ChartFiles
	err := m.QueryOne(&chartFiles, fmt.Sprintf("SELECT info FROM %s WHERE repo_namespace = $1 AND chart_files_id = $2", dbutils.ChartFilesTable), namespace, filesID)
	if err != nil {
		splittedID := strings.Split(filesID, "/")
		if withFallback && len(splittedID) == 2 {
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

func (m *PostgresAssetManager) GetPaginatedChartListWithFilters(cq ChartQuery, startItemNumber, pageSize int) ([]*models.Chart, error) {
	whereQuery, whereQueryParams, err := m.GenerateWhereClause(cq)
	if err != nil {
		return nil, err
	}
	charts, err := m.GetPaginatedChartList(whereQuery, whereQueryParams, startItemNumber, pageSize)
	if err != nil {
		return nil, err
	}
	return charts, nil
}

func (m *PostgresAssetManager) GenerateWhereClause(cq ChartQuery) (string, []interface{}, error) {
	whereClauses := []string{}
	whereQueryParams := []interface{}{}
	whereQuery := ""

	if cq.Namespace != AllNamespaces {
		whereQueryParams = append(whereQueryParams, cq.Namespace, m.GetGlobalPackagingNamespace())
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(repo_namespace = $%d OR repo_namespace = $%d)", len(whereQueryParams)-1, len(whereQueryParams),
		))
	}
	if cq.ChartName != "" {
		whereQueryParams = append(whereQueryParams, cq.ChartName)
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(info->>'name' = $%d)", len(whereQueryParams),
		))
	}
	if cq.Version != "" && cq.AppVersion != "" {
		if !containsOnlyAllowedChars(cq.Version) {
			return "", nil, errors.New("invalid version")
		}
		if !containsOnlyAllowedChars(cq.AppVersion) {
			return "", nil, errors.New("invalid app version")
		}
		parametrizedJsonbLiteral := fmt.Sprintf(`[{"version":"%s","app_version":"%s"}]`, cq.Version, cq.AppVersion)
		whereQueryParams = append(whereQueryParams, parametrizedJsonbLiteral)
		whereClauses = append(whereClauses, fmt.Sprintf("(info->'chartVersions' @> $%d::jsonb)", len(whereQueryParams)))
	}

	if cq.Repos != nil && len(cq.Repos) > 0 {
		repoClauses := []string{}
		for _, repo := range cq.Repos {
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
	if cq.Categories != nil && len(cq.Categories) > 0 {
		categoryClauses := []string{}
		for _, category := range cq.Categories {
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
	if cq.SearchQuery != "" {
		whereQueryParams = append(whereQueryParams, "%"+cq.SearchQuery+"%")
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

	return whereQuery, whereQueryParams, nil
}

// See https://semver.org/#backusnaur-form-grammar-for-valid-semver-versions
const allowed string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+-."

// Using the same "semver" validation logic for parsing version
// see https://github.com/Masterminds/semver/blob/v3.1.1/version.go
func containsOnlyAllowedChars(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool {
		return !strings.ContainsRune(allowed, r)
	}) == -1
}
