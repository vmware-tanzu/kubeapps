/*
Copyright (c) 2019 Bitnami

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
	"fmt"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
)

type assetManager interface {
	Init() error
	Close() error
	getPaginatedChartList(repo string, pageNumber, pageSize int, showDuplicates bool) ([]*models.Chart, int, error)
	getChart(chartID string) (models.Chart, error)
	getChartVersion(chartID, version string) (models.Chart, error)
	getChartFiles(filesID string) (models.ChartFiles, error)
	getChartsWithFiltes(name, version, appVersion string) ([]*models.Chart, error)
	searchCharts(query, repo string) ([]*models.Chart, error)
}

func newManager(databaseType string, config datastore.Config) (assetManager, error) {
	if databaseType == "mongodb" {
		return newMongoDBManager(config), nil
	} else if databaseType == "postgresql" {
		return newPGManager(config)
	} else {
		return nil, fmt.Errorf("Unsupported database type %s", databaseType)
	}
}
