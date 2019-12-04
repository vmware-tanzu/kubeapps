package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

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

func Test_DeletePGRepo(t *testing.T) {
	repoName := "test"
	m := &mockDB{&mock.Mock{}}
	tables := []string{chartTable, repositoryTable, chartFilesTable}
	for _, table := range tables {
		q := fmt.Sprintf("DELETE FROM %s WHERE info -> 'repo' ->> 'name' = '%s'", table, repoName)
		// Since we are not specifying any argument, Query is called with []interface{}(nil)
		m.On("Query", q, []interface{}(nil))
	}

	pgManager := &postgresAssetManager{"", m}
	err := pgManager.Delete(repoName)
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
	m.AssertExpectations(t)
}

func Test_PGRepoAlreadyPropcessed(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	pgManager := &postgresAssetManager{m}
	m.On("QueryRow", "SELECT checksum FROM repos WHERE name = 'foo'", []interface{}(nil))
	pgManager.RepoAlreadyProcessed("foo", "123")
	m.AssertExpectations(t)
}

func Test_PGUpdateLastCheck(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	repoName := "foo"
	checksum := "bar"
	now := time.Now()
	pgManager := &postgresAssetManager{m}
	expectedQuery := fmt.Sprintf(`INSERT INTO %s (name, checksum, last_update)
	VALUES ('%s', '%s', '%s')
	ON CONFLICT (name) 
	DO UPDATE SET last_update = '%s', checksum = '%s'
	`, repositoryTable, repoName, checksum, now.String(), now.String(), checksum)
	m.On("Query", expectedQuery, []interface{}(nil))
	pgManager.UpdateLastCheck(repoName, checksum, now)
	m.AssertExpectations(t)
}

func Test_PGremoveMissingCharts(t *testing.T) {
	charts := []chart{{ID: "foo"}, {ID: "bar"}}
	m := &mockDB{&mock.Mock{}}
	pgManager := &postgresAssetManager{m}
	m.On("Query", "DELETE FROM charts WHERE info ->> 'ID' NOT IN ('foo', 'bar')", []interface{}(nil))
	pgManager.removeMissingCharts(charts)
	m.AssertExpectations(t)
}
