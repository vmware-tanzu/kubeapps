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
	tables := []string{chartTable, chartFilesTable}
	for _, table := range tables {
		q := fmt.Sprintf("DELETE FROM %s WHERE info -> 'repo' ->> 'name' = $1", table)
		// Since we are not specifying any argument, Query is called with []interface{}(nil)
		m.On("Query", q, []interface{}{repoName})
	}
	m.On("Query", "DELETE FROM repos WHERE name = $1", []interface{}{repoName})

	pgManager := &postgresAssetManager{"", m}
	err := pgManager.Delete(repoName)
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
	m.AssertExpectations(t)
}

func Test_PGRepoAlreadyPropcessed(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	pgManager := &postgresAssetManager{"", m}
	m.On("QueryRow", "SELECT checksum FROM repos WHERE name = $1", []interface{}{"foo"})
	pgManager.RepoAlreadyProcessed("foo", "123")
	m.AssertExpectations(t)
}

func Test_PGUpdateLastCheck(t *testing.T) {
	m := &mockDB{&mock.Mock{}}
	repoName := "foo"
	checksum := "bar"
	now := time.Now()
	pgManager := &postgresAssetManager{"", m}
	expectedQuery := `INSERT INTO repos (name, checksum, last_update)
	VALUES ($1, $2, $3)
	ON CONFLICT (name) 
	DO UPDATE SET last_update = $3, checksum = $2
	`
	m.On("Query", expectedQuery, []interface{}{repoName, checksum, now.String()})
	pgManager.UpdateLastCheck(repoName, checksum, now)
	m.AssertExpectations(t)
}

func Test_PGremoveMissingCharts(t *testing.T) {
	charts := []chart{{ID: "foo"}, {ID: "bar"}}
	m := &mockDB{&mock.Mock{}}
	pgManager := &postgresAssetManager{"", m}
	m.On("Query", "DELETE FROM charts WHERE info ->> 'ID' NOT IN ('foo', 'bar')", []interface{}(nil))
	pgManager.removeMissingCharts(charts)
	m.AssertExpectations(t)
}
