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

// Currently these tests will be skipped entirely if a connection to
// a local postgres database `testdb` cannot be established.
// To have the tests run, simply run
// docker run --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:11.6.0-debian-9-r0
// in another terminal.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	_ "github.com/lib/pq"
)

const envvarPostgresTests = "ENABLE_PG_INTEGRATION_TESTS"

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

func skipIfNoPostgres(t *testing.T) {
	enableEnvVar := os.Getenv(envvarPostgresTests)
	runTests := false
	if enableEnvVar != "" {
		var err error
		runTests, err = strconv.ParseBool(enableEnvVar)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
	if !runTests {
		t.Skipf("skipping postgres tests as %q not set", envvarPostgresTests)
	}
}

func getInitializedManager(t *testing.T) (*postgresAssetManager, func()) {
	pam := openTestManager(t)
	cleanup := func() { pam.Close() }

	tables := strings.Join([]string{dbutils.RepositoryTable, dbutils.ChartTable, dbutils.ChartFilesTable}, ",")
	_, err := pam.DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tables))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = pam.initTables()
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

func TestGetRepoId(t *testing.T) {
	skipIfNoPostgres(t)

	testCases := []struct {
		name          string
		existingRepos []models.Repo
		newRepo       models.Repo
		expectedId    int
	}{
		{
			name: "it returns a new ID if it does not yet exist",
			existingRepos: []models.Repo{
				models.Repo{Namespace: "my-namespace", Name: "other-repo"},
				models.Repo{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedId: 3,
		},
		{
			name: "it returns the existing ID if the repo exists in the db",
			existingRepos: []models.Repo{
				models.Repo{Namespace: "my-namespace", Name: "my-name"},
				models.Repo{Namespace: "my-namespace", Name: "other-repo"},
				models.Repo{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedId: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()

			for _, repo := range tc.existingRepos {
				_, err := pam.ensureRepoExists(repo.Namespace, repo.Name)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}

			id, err := pam.ensureRepoExists(tc.newRepo.Namespace, tc.newRepo.Name)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := id, tc.expectedId; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestImportCharts(t *testing.T) {
	skipIfNoPostgres(t)

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
			repoId, err := pam.ensureRepoExists(repo.Namespace, repo.Name)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			err = pam.importCharts(tc.charts, repoId)
			if err != nil {
				t.Errorf("%+v", err)
			}
		})
	}
}

func TestInsertFiles(t *testing.T) {
	skipIfNoPostgres(t)
	const namespace = "my-namespace"

	testCases := []struct {
		name          string
		existingFiles []models.ChartFiles
		chartFiles    models.ChartFiles
		filesInserted int
	}{
		{
			name:          "it inserts new chart files",
			chartFiles:    models.ChartFiles{ID: "repo/chart-1.8", Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			filesInserted: 1,
		},
		{
			name: "it imports the same repo name and chart version in different namespaces",
			existingFiles: []models.ChartFiles{
				models.ChartFiles{ID: "repo/chart-1.8", Readme: "A different Readme", Repo: &models.Repo{Namespace: "another-namespace"}},
			},
			chartFiles:    models.ChartFiles{ID: "repo/chart-1.8", Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			filesInserted: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for _, cf := range tc.existingFiles {
				err := pam.insertFiles("some-id", cf)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}

			err := pam.insertFiles("some-id", tc.chartFiles)
			if err != nil {
				t.Errorf("%+v", err)
			}

			if got, want := countTable(t, pam, dbutils.ChartFilesTable), tc.filesInserted; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}
