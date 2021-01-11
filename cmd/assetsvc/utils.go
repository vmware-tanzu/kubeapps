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
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
)

type assetManager interface {
	Init() error
	Close() error
	getChart(namespace, chartID string) (models.Chart, error)
	getChartVersion(namespace, chartID, version string) (models.Chart, error)
	getChartFiles(namespace, filesID string) (models.ChartFiles, error)
	getPaginatedChartListWithFilters(cq chartQuery, pageNumber, pageSize int) ([]*models.Chart, int, error)
	getPaginatedChartListWithFilters(cq ChartQuery, pageNumber, pageSize int) ([]*models.Chart, int, error)
	getAllChartCategories(namespace, repo string) ([]*models.ChartCategory, error)
}

// ChartQuery is a container for passing the supported query paramters for generating the WHERE query
type ChartQuery struct {
	namespace   string
	chartName   string
	version     string
	appVersion  string
	searchQuery string
	repos       []string
	categories  []string
}

func newManager(databaseType string, config datastore.Config, kubeappsNamespace string) (assetManager, error) {
	return newPGManager(config, kubeappsNamespace)
}
