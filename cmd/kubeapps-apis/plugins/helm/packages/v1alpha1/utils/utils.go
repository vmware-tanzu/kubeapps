// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	chartUtils "github.com/vmware-tanzu/kubeapps/pkg/chart"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"helm.sh/helm/v3/pkg/chart"
	k8scorev1 "k8s.io/api/core/v1"
)

type AssetManager interface {
	Init() error
	Close() error
	GetChart(namespace, chartID string) (models.Chart, error)
	GetChartVersion(namespace, chartID, version string) (models.Chart, error)
	GetChartFiles(namespace, filesID string) (models.ChartFiles, error)
	GetPaginatedChartListWithFilters(cq ChartQuery, startItemNumber, pageSize int) ([]*models.Chart, error)
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

func NewManager(databaseType string, config dbutils.Config, globalPackagingNamespace string) (AssetManager, error) {
	return NewPGManager(config, globalPackagingNamespace)
}

// GetChart retrieves a chart
func GetChart(chartDetails *chartUtils.Details, appRepo *appRepov1.AppRepository, caCertSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret, chartClient chartUtils.ChartClient) (*chart.Chart, error) {
	err := chartClient.Init(appRepo, caCertSecret, authSecret)
	if err != nil {
		return nil, err
	}
	ch, err := chartClient.GetChart(chartDetails, appRepo.Spec.URL)
	if err != nil {
		return nil, err
	}
	return ch, nil
}
