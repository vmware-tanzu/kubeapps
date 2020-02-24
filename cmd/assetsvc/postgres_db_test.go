/*
Copyright (c) 2020 Bitnami

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

// Currently these tests will be skipped entirely unless the
// ENABLE_PG_INTEGRATION_TESTS env var is set.
// Run the local postgres with
// docker run --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:11.6.0-debian-9-r0
// in another terminal.
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
	_ "github.com/lib/pq"
)

func openTestManager(t *testing.T) *postgresAssetManager {
	pam, err := newPGManager(datastore.Config{
		URL:      "localhost:5432",
		Database: "testdb",
		Username: "postgres",
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = pam.Init()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return pam.(*postgresAssetManager)
}

func getInitializedManager(t *testing.T) (*postgresAssetManager, func()) {
	pam := openTestManager(t)
	cleanup := func() { pam.Close() }

	err := pam.InvalidateCache()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	return pam, cleanup
}

func countTable(t *testing.T, pam *dbutils.PostgresAssetManager, table string) int {
	var count int
	err := pam.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return count
}

func insertCharts(t *testing.T, pam *postgresAssetManager, charts []models.Chart, repo models.Repo) {
	_, err := pam.EnsureRepoExists(repo.Namespace, repo.Name)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	for _, chart := range charts {
		d, err := json.Marshal(chart)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		_, err = pam.GetDB().Exec(fmt.Sprintf(`INSERT INTO %s (repo_namespace, repo_name, chart_id, info)
		VALUES ($1, $2, $3, $4)`, dbutils.ChartTable), repo.Namespace, repo.Name, chart.ID, string(d))
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
}

func TestGetChart(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)
	const repoName = "repo-name"

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace
		existingCharts map[string][]models.Chart
		chartId        string
		namespace      string
		expectedChart  string
		expectedErr    error
	}{
		{
			name:        "it returns an error if the chart does not exist",
			chartId:     "doesnt-exist-1",
			namespace:   "doesnt-exist",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it returns the chart matching the chartid",
			existingCharts: map[string][]models.Chart{
				"namespace-1": []models.Chart{
					models.Chart{ID: "chart-1", Name: "my-chart"},
				},
			},
			chartId:       "chart-1",
			namespace:     "namespace-1",
			expectedErr:   nil,
			expectedChart: "my-chart",
		},
		// {
		// 	name: "it returns the chart matching the chartid in the specific namespace",
		// 	existingCharts: map[string][]models.Chart{
		// 		"namespace-1": []models.Chart{
		// 			models.Chart{ID: "chart-1", Name: "incorrect-chart"},
		// 		},
		// 		"namespace-2": []models.Chart{
		// 			models.Chart{ID: "chart-1", Name: "correct-chart"},
		// 		},
		// 	},
		// 	chartId:       "chart-1",
		// 	namespace:     "namespace-2",
		// 	expectedErr:   nil,
		// 	expectedChart: "correct-chart",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, charts := range tc.existingCharts {
				insertCharts(t, pam, charts, models.Repo{Name: repoName, Namespace: namespace})
			}

			chart, err := pam.getChart(tc.chartId)

			if got, want := err, tc.expectedErr; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
			if got, want := chart.Name, tc.expectedChart; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)
	const repoName = "repo-name"

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace
		existingCharts   map[string][]models.Chart
		chartId          string
		namespace        string
		requestedVersion string
		expectedVersion  string
		expectedErr      error
	}{
		{
			name:        "it returns an error if the chart does not exist",
			chartId:     "doesnt-exist-1",
			namespace:   "doesnt-exist",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it does not ?! return an error if the chart version does not exist",
			existingCharts: map[string][]models.Chart{
				"namespace-1": []models.Chart{
					models.Chart{ID: "chart-1", ChartVersions: []models.ChartVersion{
						models.ChartVersion{Version: "1.2.3"},
					}},
				},
			},
			chartId:          "chart-1",
			namespace:        "namespace-1",
			requestedVersion: "doesnt-exist",
			expectedVersion:  "1.2.3",
		},
		{
			name: "it returns the chart version matching the chartid and version",
			existingCharts: map[string][]models.Chart{
				"namespace-1": []models.Chart{
					models.Chart{ID: "chart-1", ChartVersions: []models.ChartVersion{
						models.ChartVersion{Version: "1.2.3"},
					}},
				},
			},
			chartId:          "chart-1",
			namespace:        "namespace-1",
			requestedVersion: "1.2.3",
			expectedVersion:  "1.2.3",
		},
		// {
		// 	name: "it returns the chart version matching the chartid and version",
		// 	existingCharts: map[string][]models.Chart{
		// 		"other-namespace": []models.Chart{
		// 			models.Chart{ID: "chart-1", ChartVersions: []models.ChartVersion{
		// 				models.ChartVersion{Version: "4.5.6"},
		// 			}},
		// 		},
		// 		"namespace-1": []models.Chart{
		// 			models.Chart{ID: "chart-1", ChartVersions: []models.ChartVersion{
		// 				models.ChartVersion{Version: "1.2.3"},
		// 			}},
		// 		},
		// 	},
		// 	chartId:          "chart-1",
		// 	namespace:        "namespace-1",
		// 	requestedVersion: "1.2.3",
		// 	expectedVersion:  "1.2.3",
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, charts := range tc.existingCharts {
				insertCharts(t, pam, charts, models.Repo{Name: repoName, Namespace: namespace})
			}

			chart, err := pam.getChartVersion(tc.chartId, tc.requestedVersion)

			if got, want := err, tc.expectedErr; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
			if tc.expectedErr != nil {
				return
			}
			// The function just returns the chart with only the one version?
			if got, want := len(chart.ChartVersions), 1; got != want {
				t.Fatalf("got: %d, want: %d", got, want)
			}
			if got, want := chart.ChartVersions[0].Version, tc.expectedVersion; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}
