package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	*mock.Mock
}

func (d *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	d.Called(query, args)
	return nil, nil
}

func (d *mockDB) Begin() (*sql.Tx, error) {
	d.Called()
	return nil, nil
}

func (d *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	d.Called(query, args)
	return nil
}

func (d *mockDB) Close() error {
	return nil
}

func (d *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func Test_DeletePGRepo(t *testing.T) {
	repo := models.Repo{Name: "testrepo", Namespace: "testnamespace"}
	m := &mockDB{&mock.Mock{}}
	m.On("Query", "DELETE FROM repos WHERE name = $1 AND namespace = $2", []interface{}{repo.Name, repo.Namespace})

	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	err := pgManager.Delete(repo)
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
	m.AssertExpectations(t)
}

func Test_PGRepoAlreadyPropcessed(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On("QueryRow", "SELECT checksum FROM repos WHERE name = $1 AND namespace = $2", []interface{}{"foo", "repo-namespace"})
	pgManager.RepoAlreadyProcessed(models.Repo{Namespace: "repo-namespace", Name: "foo"}, "123")
	m.AssertExpectations(t)
}

func Test_PGUpdateLastCheck(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	const (
		repoNamespace = "repoNamespace"
		repoName      = "foo"
		checksum      = "bar"
	)
	now := time.Now()
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	expectedQuery := `INSERT INTO repos (namespace, name, checksum, last_update)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (namespace, name)
	DO UPDATE SET last_update = $4, checksum = $3
	`
	m.On("Query", expectedQuery, []interface{}{repoNamespace, repoName, checksum, now.String()})
	pgManager.UpdateLastCheck(repoNamespace, repoName, checksum, now)
	m.AssertExpectations(t)
}

func Test_PGremoveMissingCharts(t *testing.T) {
	repo := models.Repo{Name: "repo"}
	charts := []models.Chart{{ID: "foo", Repo: &repo}, {ID: "bar"}}
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On("Query", "DELETE FROM charts WHERE chart_id NOT IN ('foo', 'bar') AND repo_name = $1 AND repo_namespace = $2", []interface{}{repo.Name, repo.Namespace})
	pgManager.removeMissingCharts(repo, charts)
	m.AssertExpectations(t)
}

func Test_PGupdateIcon(t *testing.T) {
	data := []byte("foo")
	contentType := "image/png"
	id := "stable/wordpress"
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On(
		"Query",
		`UPDATE charts SET info = info || '{"raw_icon": "Zm9v", "icon_content_type": "image/png"}' WHERE chart_id = $1 AND repo_namespace = $2 AND repo_name = $3 RETURNING ID`,
		[]interface{}{"stable/wordpress", "repo-namespace", "repo-name"},
	)
	err := pgManager.updateIcon(models.Repo{Namespace: "repo-namespace", Name: "repo-name"}, data, contentType, id)
	if err != nil {
		t.Errorf("Failed to update icon")
	}
	m.AssertExpectations(t)
}

func Test_PGfilesExist(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	rows := sqlmock.NewRows([]string{"info"}).AddRow(`true`)
	mock.ExpectQuery(`^SELECT EXISTS\(
	SELECT 1 FROM files
	WHERE chart_files_id = \$1 AND
		repo_name = \$2 AND
		repo_namespace = \$3 AND
		info ->> 'Digest' = \$4
	\)$`).WillReturnRows(rows)
	id := "stable/wordpress"
	digest := "foo"
	man := &dbutils.PostgresAssetManager{DB: db}
	pgManager := &postgresAssetManager{man}
	exists := pgManager.filesExist(models.Repo{Namespace: "namespace", Name: "repo-name"}, id, digest)
	if exists != true {
		t.Errorf("Failed to check if file exists")
	}
	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Errorf("err %v", err)
	}
}

func Test_PGinsertFiles(t *testing.T) {
	const (
		namespace = "my-namespace"
		repoName  = "my-repo"
		chartId   = repoName + "/wordpress"
		filesId   = chartId + "-2.1.3"
	)
	files := models.ChartFiles{ID: filesId, Readme: "foo", Values: "bar", Repo: &models.Repo{Namespace: namespace, Name: repoName}}
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"}, dbutilstest.KubeappsTestNamespace)
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On(
		"Query",
		`INSERT INTO files (chart_id, repo_name, repo_namespace, chart_files_ID, info)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (repo_namespace, chart_files_ID)
	DO UPDATE SET info = $5
	`,
		[]interface{}{chartId, repoName, namespace, filesId, files},
	)
	err := pgManager.insertFiles(chartId, files)
	if err != nil {
		t.Errorf("Failed to insert files: %+v", err)
	}
	m.AssertExpectations(t)
}
