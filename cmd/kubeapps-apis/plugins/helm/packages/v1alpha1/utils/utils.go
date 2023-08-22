// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	dockerConfigJSONType = "kubernetes.io/dockerconfigjson"
	dockerConfigJSONKey  = ".dockerconfigjson"
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

// RegistrySecretsPerDomain checks the app repo and available secrets
// to return the secret names per registry domain.
func RegistrySecretsPerDomain(ctx context.Context, appRepoSecrets []string, namespace string, client kubernetes.Interface) (map[string]string, error) {
	secretsPerDomain := map[string]string{}

	for _, secretName := range appRepoSecrets {
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if secret.Type != dockerConfigJSONType {
			return nil, fmt.Errorf("the AppRepository secret must be of type %q. Secret %q had type %q", dockerConfigJSONType, secretName, secret.Type)
		}

		dockerConfigJSONBytes, ok := secret.Data[dockerConfigJSONKey]
		if !ok {
			return nil, fmt.Errorf("the AppRepository secret must have a data map with a key %q. Secret %q did not", dockerConfigJSONKey, secretName)
		}

		dockerConfigJSON := kube.DockerConfigJSON{}
		if err := json.Unmarshal(dockerConfigJSONBytes, &dockerConfigJSON); err != nil {
			return nil, err
		}

		for key := range dockerConfigJSON.Auths {
			secretsPerDomain[key] = secretName
		}

	}
	return secretsPerDomain, nil
}
