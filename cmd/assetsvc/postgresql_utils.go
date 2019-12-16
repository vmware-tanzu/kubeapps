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
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
)

type postgresAssetManager struct {
	*dbutils.PostgresAssetManager
}

func newPGManager(config datastore.Config) (assetManager, error) {
	m, err := dbutils.NewPGManager(config)
	if err != nil {
		return nil, err
	}
	return &postgresAssetManager{m}, nil
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
