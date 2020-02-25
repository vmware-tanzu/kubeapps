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

func countTable(t *testing.T, pam *postgresAssetManager, table string) int {
	var count int
	err := pam.DB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return count
}

func TestImportCharts(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)

	repo := models.Repo{
		Name:      "repo-name",
		Namespace: "repo-namespace",
	}

	testCases := []struct {
		name   string
		charts []models.Chart
	}{
		{
			name: "it inserts the charts",
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			_, err := pam.EnsureRepoExists(repo.Namespace, repo.Name)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			err = pam.importCharts(tc.charts, repo)
			if err != nil {
				t.Errorf("%+v", err)
			}
		})
	}
}

func TestInsertFiles(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)
	const (
		namespace = "my-namespace"
		chartId   = "my-chart-id"
		filesId   = chartId + "-1.0"
	)

	testCases := []struct {
		name          string
		existingFiles []models.ChartFiles
		chartFiles    models.ChartFiles
		fileCount     int
	}{
		{
			name:       "it inserts new chart files",
			chartFiles: models.ChartFiles{ID: filesId, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  1,
		},
		{
			name: "it updates existing chart files",
			existingFiles: []models.ChartFiles{
				models.ChartFiles{ID: filesId, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			},
			chartFiles: models.ChartFiles{ID: filesId, Readme: "A New Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  1,
		},
		{
			name: "it imports the same repo name and chart version in different namespaces",
			existingFiles: []models.ChartFiles{
				models.ChartFiles{ID: filesId, Readme: "A different Readme", Repo: &models.Repo{Namespace: "another-namespace"}},
			},
			chartFiles: models.ChartFiles{ID: filesId, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			ensureFilesExist(t, pam, chartId, tc.existingFiles)

			ensureChartExists(t, pam, models.Chart{ID: chartId, Repo: tc.chartFiles.Repo})

			err := pam.insertFiles(chartId, tc.chartFiles)
			if err != nil {
				t.Errorf("%+v", err)
			}

			if got, want := countTable(t, pam, dbutils.ChartFilesTable), tc.fileCount; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func ensureChartExists(t *testing.T, pam *postgresAssetManager, chart models.Chart) {
	_, err := pam.EnsureRepoExists(chart.Repo.Namespace, chart.Repo.Name)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	err = pam.importCharts([]models.Chart{chart}, *chart.Repo)
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func ensureFilesExist(t *testing.T, pam *postgresAssetManager, chartId string, files []models.ChartFiles) {
	for _, f := range files {
		ensureChartExists(t, pam, models.Chart{ID: chartId, Repo: f.Repo})
		err := pam.insertFiles(chartId, f)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
}

func TestDelete(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)
	const (
		repoNamespace = "my-namespace"
		repoName      = "my-repo"
		chartId       = repoName + "/my-chart"
	)
	repoToDelete := models.Repo{Namespace: repoNamespace, Name: repoName}
	otherRepo := models.Repo{Namespace: "other-namespace", Name: repoName}

	testCases := []struct {
		name           string
		existingFiles  []models.ChartFiles
		repo           models.Repo
		expectedRepos  int
		expectedCharts int
		expectedFiles  int
	}{
		{
			name: "it deletes the repo, chart and files",
			existingFiles: []models.ChartFiles{
				models.ChartFiles{ID: chartId + "-1.8", Readme: "A Readme", Repo: &repoToDelete},
			},
			repo:           repoToDelete,
			expectedRepos:  0,
			expectedCharts: 0,
			expectedFiles:  0,
		},
		{
			name: "it deletes the repo, chart and files from that namespace only",
			existingFiles: []models.ChartFiles{
				models.ChartFiles{ID: chartId + "-1.8", Readme: "A Readme", Repo: &repoToDelete},
				models.ChartFiles{ID: chartId + "-1.8", Readme: "A Readme", Repo: &otherRepo},
			},
			repo:           repoToDelete,
			expectedRepos:  1,
			expectedCharts: 1,
			expectedFiles:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			ensureFilesExist(t, pam, chartId, tc.existingFiles)

			err := pam.Delete(repoToDelete)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := countTable(t, pam, dbutils.RepositoryTable), tc.expectedRepos; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := countTable(t, pam, dbutils.ChartTable), tc.expectedCharts; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := countTable(t, pam, dbutils.ChartFilesTable), tc.expectedFiles; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestRemoveMissingCharts(t *testing.T) {
	const (
		repoName = "my-repo"
	)
	repo := models.Repo{Name: repoName, Namespace: "my-namespace"}
	repoOtherNameSameNamespace := models.Repo{Name: "other-repo", Namespace: "my-namespace"}
	repoSameNameOtherNamespace := models.Repo{Name: repoName, Namespace: "other-namespace"}

	testCases := []struct {
		name string
		// existingFiles maps a chartId to a slice of files for different
		// versions of that chart.
		existingFiles   map[string][]models.ChartFiles
		remainingCharts []models.Chart
		expectedCharts  int
		expectedFiles   int
	}{
		{
			name: "it removes missing charts and files",
			existingFiles: map[string][]models.ChartFiles{
				"my-chart": {
					models.ChartFiles{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					models.ChartFiles{ID: "other-chart-1", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "other-chart-2", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "other-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				models.Chart{ID: "my-chart"},
			},
			expectedCharts: 1,
			expectedFiles:  1,
		},
		{
			name: "it leaves two charts while removing one",
			existingFiles: map[string][]models.ChartFiles{
				"my-chart": {
					models.ChartFiles{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					models.ChartFiles{ID: "other-chart-1", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "other-chart-2", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "other-chart-3", Readme: "A Readme", Repo: &repo},
				},
				"third-chart": {
					models.ChartFiles{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				models.Chart{ID: "my-chart"},
				models.Chart{ID: "other-chart"},
			},
			expectedCharts: 2,
			expectedFiles:  4,
		},
		{
			name: "it leaves the same chart in a different repo of same namespace",
			// None of the charts or files are removed because my-chart and third-chart
			// are the only charts in the specific repo.
			existingFiles: map[string][]models.ChartFiles{
				"my-chart": {
					models.ChartFiles{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					models.ChartFiles{ID: "other-chart-1", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
					models.ChartFiles{ID: "other-chart-2", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
					models.ChartFiles{ID: "other-chart-3", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
				},
				"third-chart": {
					models.ChartFiles{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				models.Chart{ID: "my-chart"},
				models.Chart{ID: "third-chart"},
			},
			expectedCharts: 3,
			expectedFiles:  7,
		},
		{
			name: "it leaves the same chart in a repo in a different",
			// None of the charts or files are removed because my-chart and third-chart
			// are the only charts in the specific repo.
			existingFiles: map[string][]models.ChartFiles{
				"my-chart": {
					models.ChartFiles{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					models.ChartFiles{ID: "other-chart-1", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
					models.ChartFiles{ID: "other-chart-2", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
					models.ChartFiles{ID: "other-chart-3", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
				},
				"third-chart": {
					models.ChartFiles{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					models.ChartFiles{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				models.Chart{ID: "my-chart"},
				models.Chart{ID: "third-chart"},
			},
			expectedCharts: 3,
			expectedFiles:  7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for chartId, files := range tc.existingFiles {
				ensureFilesExist(t, pam, chartId, files)
			}

			err := pam.removeMissingCharts(repo, tc.remainingCharts)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := countTable(t, pam, dbutils.ChartTable), tc.expectedCharts; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := countTable(t, pam, dbutils.ChartFilesTable), tc.expectedFiles; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}

}
