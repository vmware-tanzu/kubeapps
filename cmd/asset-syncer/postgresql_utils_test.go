package main

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
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

	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
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
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On("QueryRow", "SELECT checksum FROM repos WHERE name = $1", []interface{}{"foo"})
	pgManager.RepoAlreadyProcessed("foo", "123")
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
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
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
	charts := []models.Chart{{ID: "foo", Repo: &models.Repo{Name: "repo"}}, {ID: "bar"}}
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On("Query", "DELETE FROM charts WHERE info ->> 'ID' NOT IN ('foo', 'bar') AND info -> 'repo' ->> 'name' = $1", []interface{}{"repo"})
	pgManager.removeMissingCharts(charts)
	m.AssertExpectations(t)
}

func Test_PGupdateIcon(t *testing.T) {
	data := []byte("foo")
	contentType := "image/png"
	id := "stable/wordpress"
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On(
		"Query",
		`UPDATE charts SET info = info || '{"raw_icon": "Zm9v", "icon_content_type": "image/png"}'  WHERE info ->> 'ID' = 'stable/wordpress'`,
		[]interface{}(nil),
	)
	err := pgManager.updateIcon(data, contentType, id)
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
	rows := sqlmock.NewRows([]string{"info"}).AddRow(`{"ID": "foo"}`)
	mock.ExpectQuery(`^SELECT \* FROM files WHERE chart_files_id = \$1 AND info ->> 'Digest' = \$2$`).WillReturnRows(rows)
	id := "stable/wordpress"
	digest := "foo"
	man := &dbutils.PostgresAssetManager{DB: db}
	pgManager := &postgresAssetManager{man}
	exists := pgManager.filesExist(id, digest)
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
		id        = "stable/wordpress"
		repoName  = "my-repo"
	)
	files := models.ChartFiles{ID: id, Readme: "foo", Values: "bar", Repo: &models.Repo{Namespace: namespace, Name: repoName}}
	m := &mockDB{&mock.Mock{}}
	man, _ := dbutils.NewPGManager(datastore.Config{URL: "localhost:4123"})
	man.DB = m
	pgManager := &postgresAssetManager{man}
	m.On(
		"Query",
		`INSERT INTO files (repo_name, repo_namespace, chart_files_ID, info)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (repo_namespace, chart_files_ID)
	DO UPDATE SET info = $4
	`,
		[]interface{}{repoName, namespace, id, files},
	)
	err := pgManager.insertFiles(id, files)
	if err != nil {
		t.Errorf("Failed to insert files: %+v", err)
	}
	m.AssertExpectations(t)
}
