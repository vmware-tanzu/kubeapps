// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Currently these tests will be skipped entirely unless the
// ENABLE_PG_INTEGRATION_TESTS env var is set.
// Run the local postgres with
// docker run --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:14.5.0-debian-11-r6
// in another terminal.
package server

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils/dbutilstest/pgtest"
)

func getInitializedManager(t *testing.T) (*postgresAssetManager, func()) {
	pam, cleanup := pgtest.GetInitializedManager(t)
	return &postgresAssetManager{pam}, cleanup
}

func TestEnsureRepoExists(t *testing.T) {
	pgtest.SkipIfNoDB(t)

	testCases := []struct {
		name          string
		existingRepos []models.Repo
		newRepo       models.Repo
		expectedID    int
	}{
		{
			name: "it returns a new ID if it does not yet exist",
			existingRepos: []models.Repo{
				{Namespace: "my-namespace", Name: "other-repo"},
				{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedID: 3,
		},
		{
			name: "it returns the existing ID if the repo exists in the db",
			existingRepos: []models.Repo{
				{Namespace: "my-namespace", Name: "my-name"},
				{Namespace: "my-namespace", Name: "other-repo"},
				{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedID: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()

			for _, repo := range tc.existingRepos {
				_, err := pam.EnsureRepoExists(repo.Namespace, repo.Name)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}

			id, err := pam.EnsureRepoExists(tc.newRepo.Namespace, tc.newRepo.Name)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := id, tc.expectedID; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestImportCharts(t *testing.T) {
	pgtest.SkipIfNoDB(t)

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
				{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
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
	pgtest.SkipIfNoDB(t)
	const (
		namespace = "my-namespace"
		chartID   = "my-chart-id"
		filesID   = chartID + "-1.0"
	)

	testCases := []struct {
		name          string
		existingFiles []models.ChartFiles
		chartFiles    models.ChartFiles
		fileCount     int
	}{
		{
			name:       "it inserts new chart files",
			chartFiles: models.ChartFiles{ID: filesID, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  1,
		},
		{
			name: "it updates existing chart files",
			existingFiles: []models.ChartFiles{
				{ID: filesID, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			},
			chartFiles: models.ChartFiles{ID: filesID, Readme: "A New Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  1,
		},
		{
			name: "it imports the same repo name and chart version in different namespaces",
			existingFiles: []models.ChartFiles{
				{ID: filesID, Readme: "A different Readme", Repo: &models.Repo{Namespace: "another-namespace"}},
			},
			chartFiles: models.ChartFiles{ID: filesID, Readme: "A Readme", Repo: &models.Repo{Namespace: namespace}},
			fileCount:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			ensureFilesExist(t, pam, chartID, tc.existingFiles)

			pgtest.EnsureChartsExist(t, pam, []models.Chart{{ID: chartID}}, *tc.chartFiles.Repo)

			err := pam.insertFiles(chartID, tc.chartFiles)
			if err != nil {
				t.Errorf("%+v", err)
			}

			if got, want := pgtest.CountRows(t, pam.DB, dbutils.ChartFilesTable), tc.fileCount; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func ensureFilesExist(t *testing.T, pam *postgresAssetManager, chartID string, files []models.ChartFiles) {
	for _, f := range files {
		pgtest.EnsureChartsExist(t, pam, []models.Chart{{ID: chartID}}, *f.Repo)
		err := pam.insertFiles(chartID, f)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
}

func TestDelete(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const (
		repoNamespace = "my-namespace"
		repoName      = "my-repo"
		chartID       = repoName + "/my-chart"
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
				{ID: chartID + "-1.8", Readme: "A Readme", Repo: &repoToDelete},
			},
			repo:           repoToDelete,
			expectedRepos:  0,
			expectedCharts: 0,
			expectedFiles:  0,
		},
		{
			name: "it deletes the repo, chart and files from that namespace only",
			existingFiles: []models.ChartFiles{
				{ID: chartID + "-1.8", Readme: "A Readme", Repo: &repoToDelete},
				{ID: chartID + "-1.8", Readme: "A Readme", Repo: &otherRepo},
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
			ensureFilesExist(t, pam, chartID, tc.existingFiles)

			err := pam.Delete(repoToDelete)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := pgtest.CountRows(t, pam.DB, dbutils.RepositoryTable), tc.expectedRepos; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := pgtest.CountRows(t, pam.DB, dbutils.ChartTable), tc.expectedCharts; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := pgtest.CountRows(t, pam.DB, dbutils.ChartFilesTable), tc.expectedFiles; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestRemoveMissingCharts(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const (
		repoName = "my-repo"
	)
	repo := models.Repo{Name: repoName, Namespace: "my-namespace"}
	repoOtherNameSameNamespace := models.Repo{Name: "other-repo", Namespace: "my-namespace"}
	repoSameNameOtherNamespace := models.Repo{Name: repoName, Namespace: "other-namespace"}

	testCases := []struct {
		name string
		// existingFiles maps a chartID to a slice of files for different
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
					{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					{ID: "other-chart-1", Readme: "A Readme", Repo: &repo},
					{ID: "other-chart-2", Readme: "A Readme", Repo: &repo},
					{ID: "other-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				{ID: "my-chart"},
			},
			expectedCharts: 1,
			expectedFiles:  1,
		},
		{
			name: "it leaves two charts while removing one",
			existingFiles: map[string][]models.ChartFiles{
				"my-chart": {
					{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					{ID: "other-chart-1", Readme: "A Readme", Repo: &repo},
					{ID: "other-chart-2", Readme: "A Readme", Repo: &repo},
					{ID: "other-chart-3", Readme: "A Readme", Repo: &repo},
				},
				"third-chart": {
					{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				{ID: "my-chart"},
				{ID: "other-chart"},
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
					{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					{ID: "other-chart-1", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
					{ID: "other-chart-2", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
					{ID: "other-chart-3", Readme: "A Readme", Repo: &repoOtherNameSameNamespace},
				},
				"third-chart": {
					{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				{ID: "my-chart"},
				{ID: "third-chart"},
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
					{ID: "my-chart-1", Readme: "A Readme", Repo: &repo},
				},
				"other-chart": {
					{ID: "other-chart-1", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
					{ID: "other-chart-2", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
					{ID: "other-chart-3", Readme: "A Readme", Repo: &repoSameNameOtherNamespace},
				},
				"third-chart": {
					{ID: "third-chart-1", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-2", Readme: "A Readme", Repo: &repo},
					{ID: "third-chart-3", Readme: "A Readme", Repo: &repo},
				},
			},
			remainingCharts: []models.Chart{
				{ID: "my-chart"},
				{ID: "third-chart"},
			},
			expectedCharts: 3,
			expectedFiles:  7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for chartID, files := range tc.existingFiles {
				ensureFilesExist(t, pam, chartID, files)
			}

			err := pam.removeMissingCharts(repo, tc.remainingCharts)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := pgtest.CountRows(t, pam.DB, dbutils.ChartTable), tc.expectedCharts; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := pgtest.CountRows(t, pam.DB, dbutils.ChartFilesTable), tc.expectedFiles; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestRepoAlreadyProcessed(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const (
		repoNamespace = "my-namespace"
		repoName      = "my-repo"
		checksum      = "deadbeef"
	)
	repo := models.Repo{Namespace: repoNamespace, Name: repoName}

	pam, cleanup := getInitializedManager(t)
	defer cleanup()

	// not processed when it doesn't exist
	got := pam.LastChecksum(repo)
	if got != "" {
		t.Errorf("got: %s, want: %s", got, "")
	}

	// not processed when repo exists but has not been processed
	_, err := pam.EnsureRepoExists(repoNamespace, repoName)
	if err != nil {
		t.Errorf("%+v", err)
	}
	got = pam.LastChecksum(repo)
	if got != "" {
		t.Errorf("got: %s, want: %s", got, "")
	}

	err = pam.UpdateLastCheck(repoNamespace, repoName, checksum, time.Now())
	if err != nil {
		t.Errorf("%+v", err)
	}
	// processed when checksums match
	_, err = pam.EnsureRepoExists(repoNamespace, repoName)
	if err != nil {
		t.Errorf("%+v", err)
	}
	got = pam.LastChecksum(repo)
	if got != checksum {
		t.Errorf("got: %s, want: %s", got, checksum)
	}

	// it does not match the same repo in a different namespace
	got = pam.LastChecksum(models.Repo{Namespace: "other-namespace", Name: repo.Name})
	if got != "" {
		t.Errorf("got: %s, want: %s", got, "")
	}
}

func TestFilesExist(t *testing.T) {
	pgtest.SkipIfNoDB(t)

	const (
		namespace = "my-namespace"
		repoName  = "my-repo"
		chartID   = repoName + "/chart-name"
		filesID   = chartID + "-1.0"
		digest    = "some-digest"
	)
	repo := models.Repo{Namespace: namespace, Name: repoName}
	pam, cleanup := getInitializedManager(t)
	defer cleanup()

	// false when it does not exist
	if got, want := pam.filesExist(repo, filesID, digest), false; got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}

	// false when it exists with a different digest
	ensureFilesExist(t, pam, chartID, []models.ChartFiles{{ID: filesID, Repo: &repo, Digest: "other-digest"}})
	if got, want := pam.filesExist(repo, filesID, digest), false; got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}

	// true when it exists in the repo with the correct digest
	ensureFilesExist(t, pam, chartID, []models.ChartFiles{{ID: filesID, Repo: &repo, Digest: digest}})
	if got, want := pam.filesExist(repo, filesID, digest), true; got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}

	// false when it exists in another repo
	if got, want := pam.filesExist(models.Repo{Namespace: repo.Namespace, Name: "other-name"}, filesID, digest), false; got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}
	if got, want := pam.filesExist(models.Repo{Namespace: "other-namespace", Name: repo.Name}, filesID, digest), false; got != want {
		t.Errorf("got: %t, want: %t", got, want)
	}
}

func TestUpdateIcon(t *testing.T) {
	pgtest.SkipIfNoDB(t)

	const (
		iconContentType = "icon-content-type"
		repoNamespace   = "repo-namespace"
		repoName        = "repo-name"
		chartID         = repoName + "/chart-id"
	)
	iconData := []byte("icon-data")
	repo := models.Repo{Namespace: repoNamespace, Name: repoName}

	testCases := []struct {
		name           string
		existingCharts map[string][]string
		expectedErr    error
	}{
		{
			name:        "it errors if the chart does not exist",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it updates the chart if it exists in the repo",
			existingCharts: map[string][]string{
				repoNamespace: {chartID},
			},
		},
		{
			name: "it updates only the chart in the specific namespace",
			existingCharts: map[string][]string{
				repoNamespace:     {chartID},
				"other-namespace": {chartID},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, chartIDs := range tc.existingCharts {
				charts := []models.Chart{}
				for _, chartID := range chartIDs {
					charts = append(charts, models.Chart{ID: chartID})
				}
				pgtest.EnsureChartsExist(t, pam, charts, models.Repo{Namespace: namespace, Name: repoName})
			}

			err := pam.updateIcon(repo, iconData, iconContentType, chartID)

			if got, want := err, tc.expectedErr; !errors.Is(got, want) {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
		})
	}
}
