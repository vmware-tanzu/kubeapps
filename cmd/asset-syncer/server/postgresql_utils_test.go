// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
)

func getMockManager(t *testing.T) (*postgresAssetManager, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	pgManager := &postgresAssetManager{&dbutils.PostgresAssetManager{DB: db}}

	return pgManager, mock, func() { db.Close() }
}

func Test_DeletePGRepo(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	repo := models.Repo{Name: "testrepo", Namespace: "testnamespace"}
	mock.ExpectQuery(`DELETE FROM repos WHERE name = \$1 AND namespace = \$2`).
		WithArgs(repo.Name, repo.Namespace).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	err := pgManager.Delete(repo)
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
}

func Test_PGRepoLastChecksum(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	mock.ExpectQuery("SELECT checksum FROM repos *").
		WithArgs("foo", "repo-namespace").
		WillReturnRows(sqlmock.NewRows([]string{"checksum"}).AddRow("123"))

	got := pgManager.LastChecksum(models.Repo{Namespace: "repo-namespace", Name: "foo"})
	if got != "123" {
		t.Errorf("got: %s, want: %s", got, "123")
	}
}

func Test_PGUpdateLastCheck(t *testing.T) {
	const (
		repoNamespace = "repoNamespace"
		repoName      = "foo"
		checksum      = "bar"
	)

	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	now := time.Now()
	mock.ExpectQuery("INSERT INTO repos *").
		WithArgs(repoNamespace, repoName, checksum, now.String()).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

	err := pgManager.UpdateLastCheck(repoNamespace, repoName, checksum, now)
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func Test_PGremoveMissingCharts(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	repo := models.Repo{Name: "repo"}
	charts := []models.Chart{{ID: "foo", Repo: &repo}, {ID: "bar"}}
	mock.ExpectQuery(`^DELETE FROM charts WHERE chart_id NOT IN \('foo', 'bar'\) AND repo_name = \$1 AND repo_namespace = \$2`).
		WithArgs(repo.Name, repo.Namespace).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1).AddRow(2))

	err := pgManager.removeMissingCharts(repo, charts)
	if err != nil {
		t.Errorf("%+v", err)
	}
}

func Test_PGupdateIcon(t *testing.T) {
	data := []byte("foo")
	contentType := "image/png"
	id := "stable/wordpress"

	t.Run("one icon only", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		mock.ExpectQuery(`^UPDATE charts SET info = info *`).
			WithArgs("stable/wordpress", "repo-namespace", "repo-name").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

		err := pgManager.updateIcon(models.Repo{Namespace: "repo-namespace", Name: "repo-name"}, data, contentType, id)

		if err != nil {
			t.Errorf("%+v", err)
		}
	})

	t.Run("no rows returned", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		mock.ExpectQuery(`^UPDATE charts SET info = info *`).
			WithArgs("stable/wordpress", "repo-namespace", "repo-name").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}))

		err := pgManager.updateIcon(models.Repo{Namespace: "repo-namespace", Name: "repo-name"}, data, contentType, id)

		if got, want := err, sql.ErrNoRows; got != want {
			t.Errorf("got: %+v, want: %+v", got, want)
		}
	})

	t.Run("more than one icon", func(t *testing.T) {
		pgManager, mock, cleanup := getMockManager(t)
		defer cleanup()
		mock.ExpectQuery(`^UPDATE charts SET info = info *`).
			WithArgs("stable/wordpress", "repo-namespace", "repo-name").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1).AddRow(2))

		err := pgManager.updateIcon(models.Repo{Namespace: "repo-namespace", Name: "repo-name"}, data, contentType, id)

		if got, want := errors.Is(err, ErrMultipleRows), true; got != want {
			t.Errorf("got: %t, want: %t", got, want)
		}
	})
}

func Test_PGfilesExist(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()

	const (
		id     = "stable/wordpress"
		digest = "foo"
	)
	repo := models.Repo{Namespace: "namespace", Name: "repo-name"}

	rows := sqlmock.NewRows([]string{"info"}).AddRow(`true`)
	mock.ExpectQuery(`^SELECT EXISTS\(
		SELECT 1 FROM files
		WHERE chart_files_id = \$1 AND
			repo_name = \$2 AND
			repo_namespace = \$3 AND
			info ->> 'Digest' = \$4
		\)$`).
		WithArgs(id, repo.Name, repo.Namespace, digest).
		WillReturnRows(rows)

	exists := pgManager.filesExist(repo, id, digest)
	if exists != true {
		t.Errorf("Failed to check if file exists")
	}
}

func Test_PGinsertFiles(t *testing.T) {
	pgManager, mock, cleanup := getMockManager(t)
	defer cleanup()
	const (
		namespace        = "my-namespace"
		repoName         = "my-repo"
		chartID   string = repoName + "/wordpress"
		filesID   string = chartID + "-2.1.3"
	)
	files := models.ChartFiles{ID: filesID, Readme: "foo", Values: "bar", Repo: &models.Repo{Namespace: namespace, Name: repoName}}
	mock.ExpectQuery(`INSERT INTO files \(chart_id, repo_name, repo_namespace, chart_files_ID, info\)*`).
		WithArgs(chartID, repoName, namespace, filesID, files).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow("3"))

	err := pgManager.insertFiles(chartID, files)
	if err != nil {
		t.Errorf("%+v", err)
	}
}
