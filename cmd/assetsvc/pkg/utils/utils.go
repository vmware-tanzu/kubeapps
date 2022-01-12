/*
Copyright 2021 VMware. All Rights Reserved.

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

package utils

import (
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
)

type AssetManager interface {
	Init() error
	Close() error
	GetChart(namespace, chartID string) (models.Chart, error)
	GetChartVersion(namespace, chartID, version string) (models.Chart, error)
	GetChartFiles(namespace, filesID string) (models.ChartFiles, error)
	GetPaginatedChartListWithFilters(cq ChartQuery, pageNumber, pageSize int) ([]*models.Chart, int, error)
	GetAllChartCategories(cq ChartQuery) ([]*models.ChartCategory, error)
}

// ChartQuery is a container for passing the supported query parameters for generating the WHERE query
type ChartQuery struct {
	Namespace   string
	ChartName   string
	Version     string
	AppVersion  string
	SearchQuery string
	Repos       []string
	Categories  []string
}

func NewManager(databaseType string, config dbutils.Config, globalReposNamespace string) (AssetManager, error) {
	return NewPGManager(config, globalReposNamespace)
}
