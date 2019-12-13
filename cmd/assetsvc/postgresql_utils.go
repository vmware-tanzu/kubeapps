/*
Copyright (c) 2018 Bitnami

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

package main

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/models"
)

type postgresDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Begin() (*sql.Tx, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Close() error
}

type postgresAssetManager struct {
	connStr string
	db      postgresDB
}

func newPGManager(config datastore.Config) (assetManager, error) {
	url := strings.Split(config.URL, ":")
	if len(url) != 2 {
		return nil, fmt.Errorf("Can't parse database URL: %s", config.URL)
	}
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		url[0], url[1], config.Username, config.Password, config.Database,
	)
	return &postgresAssetManager{connStr, nil}, nil
}

func (m *postgresAssetManager) Init() error {
	db, err := sql.Open("postgres", m.connStr)
	if err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *postgresAssetManager) Close() error {
	return m.db.Close()
}

func (m *postgresAssetManager) getPaginatedChartList(repo string, pageNumber, pageSize int, showDuplicates bool) ([]*models.Chart, int, error) {
	// TODO: Implement this!
	return nil, 0, nil
}

func (m *postgresAssetManager) getChart(chartID string) (models.Chart, error) {
	// TODO: Implement this!
	return models.Chart{}, nil
}

func (m *postgresAssetManager) getChartVersion(chartID, version string) (models.Chart, error) {
	// TODO: Implement this!
	return models.Chart{}, nil
}

func (m *postgresAssetManager) getChartFiles(filesID string) (models.ChartFiles, error) {
	// TODO: Implement this!
	return models.ChartFiles{}, nil
}

func (m *postgresAssetManager) getChartsWithFiltes(name, version, appVersion string) ([]*models.Chart, error) {
	// TODO: Implement this!
	return nil, nil
}

func (m *postgresAssetManager) searchCharts(query, repo string) ([]*models.Chart, error) {
	// TODO: Implement this!
	return nil, nil
}
