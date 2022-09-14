// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"database/sql/driver"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
)

func getMockManager(t *testing.T) (*PostgresAssetManager, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	pgManager := &PostgresAssetManager{&dbutils.PostgresAssetManager{DB: db, GlobalPackagingNamespace: "kubeapps"}}

	return pgManager, mock, func() { db.Close() }
}

func Test_PGGetChart(t *testing.T) {
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

	chart, err := pgManager.GetChart("namespace", "foo")
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

func Test_PGGetChartVersion(t *testing.T) {
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

	chart, err := pgManager.GetChartVersion("namespace", "foo", "1.0.0")
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

func Test_GetChartFiles(t *testing.T) {
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

	files, err := pgManager.GetChartFiles("namespace", "foo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	if !cmp.Equal(files, expectedFiles) {
		t.Errorf("Unexpected result %v", cmp.Diff(files, expectedFiles))
	}
}

func Test_GetChartFiles_withSlashes(t *testing.T) {
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

	files, err := pgManager.GetChartFiles("namespace", "fo%2Fo")
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	if !cmp.Equal(files, expectedFiles) {
		t.Errorf("Unexpected result %v", cmp.Diff(files, expectedFiles))
	}
}

func Test_GetChartsWithFilters(t *testing.T) {
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

	version := "1.0.0"
	appVersion := "1.0.1"
	parametrizedJsonbLiteral := fmt.Sprintf(`[{"version":"%s","app_version":"%s"}]`, version, appVersion)

	mock.ExpectQuery("SELECT info FROM charts WHERE *").
		WithArgs("namespace", "kubeapps", "foo", parametrizedJsonbLiteral).
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(dbChartJSON))

	charts, err := pgManager.GetPaginatedChartListWithFilters(ChartQuery{Namespace: "namespace", ChartName: "foo", Version: version, AppVersion: appVersion}, 1, 0)
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedCharts := []*models.Chart{{
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

func Test_GetChartsWithFilters_withSlashes(t *testing.T) {
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

	version := "1.0.0"
	appVersion := "1.0.1"
	parametrizedJsonbLiteral := fmt.Sprintf(`[{"version":"%s","app_version":"%s"}]`, version, appVersion)

	mock.ExpectQuery("SELECT info FROM charts WHERE *").
		WithArgs("namespace", "kubeapps", "fo%2Fo", parametrizedJsonbLiteral).
		WillReturnRows(sqlmock.NewRows([]string{"info"}).AddRow(dbChartJSON))

	charts, err := pgManager.GetPaginatedChartListWithFilters(ChartQuery{Namespace: "namespace", ChartName: "fo%2Fo", Version: version, AppVersion: appVersion}, 1, 0)
	if err != nil {
		t.Errorf("Found error %v", err)
	}
	expectedCharts := []*models.Chart{{
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

func Test_GetAllChartCategories(t *testing.T) {

	tests := []struct {
		name                    string
		namespace               string
		repo                    string
		expectedChartCategories []*models.ChartCategory
	}{
		{
			name:      "without repo",
			namespace: "other-namespace",
			repo:      "",
			expectedChartCategories: []*models.ChartCategory{
				{Name: "cat1", Count: 1},
				{Name: "cat2", Count: 2},
				{Name: "cat3", Count: 3},
			},
		},
		{
			name:      "with repo",
			namespace: "other-namespace",
			repo:      "bitnami",
			expectedChartCategories: []*models.ChartCategory{
				{Name: "cat1", Count: 1},
				{Name: "cat2", Count: 2},
				{Name: "cat3", Count: 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgManager, mock, cleanup := getMockManager(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"name", "count"})
			for _, chartCategories := range tt.expectedChartCategories {
				rows.AddRow(chartCategories.Name, chartCategories.Count)
			}

			expectedParams := []driver.Value{"other-namespace", "kubeapps"}
			if tt.repo != "" {
				expectedParams = append(expectedParams, tt.repo)
			}
			mock.ExpectQuery("SELECT (info ->> 'category')*").
				WithArgs(expectedParams...).
				WillReturnRows(rows)

			chartCategories, err := pgManager.GetAllChartCategories(ChartQuery{Namespace: tt.namespace, Repos: []string{tt.repo}})
			if err != nil {
				t.Fatalf("Found error %v", err)
			}
			if !cmp.Equal(chartCategories, tt.expectedChartCategories) {
				t.Errorf("Unexpected result %v", cmp.Diff(chartCategories, tt.expectedChartCategories))
			}
		})
	}
}
func Test_GetPaginatedChartList(t *testing.T) {
	availableCharts := []*models.Chart{
		{ID: "bar", ChartVersions: []models.ChartVersion{{Digest: "456"}}},
		{ID: "copyFoo", ChartVersions: []models.ChartVersion{{Digest: "123"}}},
		{ID: "foo", ChartVersions: []models.ChartVersion{{Digest: "123"}}},
		{ID: "fo%2Fo", ChartVersions: []models.ChartVersion{{Digest: "321"}}},
	}
	tests := []struct {
		name            string
		namespace       string
		repo            string
		startItemNumber int
		pageSize        int
		expectedCharts  []*models.Chart
	}{
		{
			name:            "one page with duplicates with repo",
			namespace:       "other-namespace",
			repo:            "bitnami",
			startItemNumber: 0,
			pageSize:        100,
			expectedCharts:  availableCharts,
		},
		{
			name:            "one page with duplicates",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 0,
			pageSize:        100,
			expectedCharts:  availableCharts,
		},
		{
			name:            "repo has many charts with pagination (2 pages)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 2,
			pageSize:        2,
			expectedCharts:  []*models.Chart{availableCharts[2], availableCharts[3]},
		},
		{
			name:            "repo has many charts with pagination (non existing page)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 5,
			pageSize:        2,
			expectedCharts:  []*models.Chart{},
		},
		{
			name:            "repo has many charts with pagination (out of range size)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 0,
			pageSize:        100,
			expectedCharts:  availableCharts,
		},
		{
			name:           "repo has many charts with pagination (w/ startItem, w size)",
			namespace:      "other-namespace",
			repo:           "",
			pageSize:       3,
			expectedCharts: []*models.Chart{availableCharts[0], availableCharts[1], availableCharts[2]},
		},
		{
			name:            "repo has many charts with pagination (w/ startItem, w zero size)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 2,
			pageSize:        0,
			expectedCharts:  availableCharts,
		},
		{
			name:            "repo has many charts with pagination (w/ wrong page, w/ size)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: -2,
			pageSize:        2,
			expectedCharts:  []*models.Chart{availableCharts[0], availableCharts[1]},
		},
		{
			name:            "repo has many charts with pagination (w/ page, w/o size)",
			namespace:       "other-namespace",
			repo:            "",
			startItemNumber: 2,
			expectedCharts:  availableCharts,
		},
		{
			name:           "repo has many charts with pagination (w/o page, w/ size)",
			namespace:      "other-namespace",
			repo:           "",
			pageSize:       2,
			expectedCharts: []*models.Chart{availableCharts[0], availableCharts[1]},
		},
		{
			name:           "repo has many charts with pagination (w/o page, w/o size)",
			namespace:      "other-namespace",
			repo:           "",
			expectedCharts: availableCharts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgManager, mock, cleanup := getMockManager(t)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})
			rowCount := sqlmock.NewRows([]string{"count"}).AddRow(len(availableCharts))

			for _, chart := range tt.expectedCharts {
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

			charts, err := pgManager.GetPaginatedChartListWithFilters(ChartQuery{Namespace: tt.namespace, Repos: []string{tt.repo}}, tt.startItemNumber, tt.pageSize)
			if err != nil {
				t.Fatalf("Found error %v", err)
			}
			if tt.pageSize > 0 {
				if len(charts) > tt.pageSize {
					t.Errorf("Unexpected number of charts, got %d expecting %d", len(charts), tt.pageSize)
				}
			}
			if !cmp.Equal(charts, tt.expectedCharts) {
				t.Errorf("Unexpected result %v", cmp.Diff(tt.expectedCharts, charts))
			}
		})
	}
}

func Test_GenerateWhereClause(t *testing.T) {
	tests := []struct {
		name           string
		namespace      string
		chartName      string
		version        string
		appVersion     string
		repos          []string
		categories     []string
		query          string
		expectedClause string
		expectedParams []interface{}
	}{
		{
			name:           "returns where clause - no params",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2)",
			expectedParams: []interface{}{string(""), string("kubeapps")},
		},
		{
			name:           "returns where clause - single param - namespace",
			namespace:      "my-ns",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2)",
			expectedParams: []interface{}{string("my-ns"), string("kubeapps")},
		},
		{
			name:           "returns where clause - single param - name",
			namespace:      "",
			chartName:      "my-chart",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2) AND (info->>'name' = $3)",
			expectedParams: []interface{}{string(""), string("kubeapps"), string("my-chart")},
		},
		{
			name:           "returns where clause - single param - version",
			namespace:      "",
			chartName:      "",
			version:        "1.0.0",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2)", //needs both version and appVersion
			expectedParams: []interface{}{string(""), string("kubeapps")},
		},
		{
			name:           "returns where clause - single param - appVersion",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "0.1.0",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2)", //needs both version and appVersion
			expectedParams: []interface{}{string(""), string("kubeapps")},
		},
		{
			name:           "returns where clause - single param - version AND appVersion",
			namespace:      "",
			chartName:      "",
			version:        "1.0.0",
			appVersion:     "0.1.0",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND (info->'chartVersions' @> $3::jsonb)`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string(`[{"version":"1.0.0","app_version":"0.1.0"}]`)},
		},
		{
			name:           "returns where clause - single param - version AND appVersion malformed with quotes",
			namespace:      "",
			chartName:      "",
			version:        "'\"1.0.0'",
			appVersion:     "'\"0.1.0'",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: ``,
			expectedParams: nil,
		},
		{
			name:           "returns where clause - no params",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "",
			expectedClause: "WHERE (repo_namespace = $1 OR repo_namespace = $2)",
			expectedParams: []interface{}{string(""), string("kubeapps")},
		},
		{
			name:           "returns where clause - single param - single repo",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{"my-repo1"},
			categories:     []string{""},
			query:          "",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((repo_name = $3))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("my-repo1")},
		},
		{
			name:           "returns where clause - single param - multiple repos",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{"my-repo1", "my-repo2"},
			categories:     []string{""},
			query:          "",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((repo_name = $3) OR (repo_name = $4))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("my-repo1"), string("my-repo2")},
		},
		{
			name:           "returns where clause - single param - single category",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{"my-category1"},
			query:          "",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND (info->>'category' = $3)`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("my-category1")},
		},
		{
			name:           "returns where clause - single param - multiple categories",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{"my-category1", "my-category2"},
			query:          "",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND (info->>'category' = $3 OR info->>'category' = $4)`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("my-category1"), string("my-category2")},
		},
		{
			name:           "returns where clause - single param - query (one word)",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "chart",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((info ->> 'name' ILIKE $3) OR (info ->> 'description' ILIKE $3) OR (info -> 'repo' ->> 'name' ILIKE $3) OR (info ->> 'keywords' ILIKE $3) OR (info ->> 'sources' ILIKE $3) OR (info -> 'maintainers' ->> 'name' ILIKE $3))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("%chart%")},
		},
		{
			name:           "returns where clause - single param - query (two words)",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "my chart",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((info ->> 'name' ILIKE $3) OR (info ->> 'description' ILIKE $3) OR (info -> 'repo' ->> 'name' ILIKE $3) OR (info ->> 'keywords' ILIKE $3) OR (info ->> 'sources' ILIKE $3) OR (info -> 'maintainers' ->> 'name' ILIKE $3))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("%my chart%")},
		},
		{
			name:           "returns where clause - single param - query (with slash)",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "my/chart",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((info ->> 'name' ILIKE $3) OR (info ->> 'description' ILIKE $3) OR (info -> 'repo' ->> 'name' ILIKE $3) OR (info ->> 'keywords' ILIKE $3) OR (info ->> 'sources' ILIKE $3) OR (info -> 'maintainers' ->> 'name' ILIKE $3))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("%my/chart%")},
		},
		{
			name:           "returns where clause - single param - query (encoded)",
			namespace:      "",
			chartName:      "",
			version:        "",
			appVersion:     "",
			repos:          []string{""},
			categories:     []string{""},
			query:          "my%2Fchart",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND ((info ->> 'name' ILIKE $3) OR (info ->> 'description' ILIKE $3) OR (info -> 'repo' ->> 'name' ILIKE $3) OR (info ->> 'keywords' ILIKE $3) OR (info ->> 'sources' ILIKE $3) OR (info -> 'maintainers' ->> 'name' ILIKE $3))`,
			expectedParams: []interface{}{string(""), string("kubeapps"), string("%my%2Fchart%")},
		},
		{
			name:           "returns where clause - every param",
			namespace:      "my-ns",
			chartName:      "my-chart",
			version:        "1.0.0",
			appVersion:     "0.1.0",
			repos:          []string{"my-repo1", "my-repo2"},
			categories:     []string{"my-category1", "my-category2"},
			query:          "best chart",
			expectedClause: `WHERE (repo_namespace = $1 OR repo_namespace = $2) AND (info->>'name' = $3) AND (info->'chartVersions' @> $4::jsonb) AND ((repo_name = $5) OR (repo_name = $6)) AND (info->>'category' = $7 OR info->>'category' = $8) AND ((info ->> 'name' ILIKE $9) OR (info ->> 'description' ILIKE $9) OR (info -> 'repo' ->> 'name' ILIKE $9) OR (info ->> 'keywords' ILIKE $9) OR (info ->> 'sources' ILIKE $9) OR (info -> 'maintainers' ->> 'name' ILIKE $9))`,
			expectedParams: []interface{}{string("my-ns"), string("kubeapps"), string("my-chart"), string(`[{"version":"1.0.0","app_version":"0.1.0"}]`), string("my-repo1"), string("my-repo2"), string("my-category1"), string("my-category2"), string("%best chart%")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pgManager, _, cleanup := getMockManager(t)
			defer cleanup()

			cq := ChartQuery{
				Namespace:   tt.namespace,
				ChartName:   tt.chartName,
				Version:     tt.version,
				AppVersion:  tt.appVersion,
				SearchQuery: tt.query,
				Repos:       tt.repos,
				Categories:  tt.categories,
			}
			whereQuery, whereQueryParams, _ := pgManager.GenerateWhereClause(cq)

			if tt.expectedClause != whereQuery {
				t.Errorf("Expecting query:\n'%s'\nreceived query:\n'%s'\nin '%s'", tt.expectedClause, whereQuery, tt.name)
			}

			if !cmp.Equal(tt.expectedParams, whereQueryParams) {
				t.Errorf("Param mismatch in '%s': %s", tt.name, cmp.Diff(tt.expectedParams, whereQueryParams))
			}
		})
	}
}
