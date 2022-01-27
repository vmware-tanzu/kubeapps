// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	chartutils "github.com/kubeapps/kubeapps/pkg/chart"
	helmchart "helm.sh/helm/v3/pkg/chart"
	k8scorev1 "k8s.io/api/core/v1"
	k8syaml "sigs.k8s.io/yaml"
)

// ChartClient implements Resolver inteface
type ChartClient struct{}

// GetChart fake
func (f *ChartClient) GetChart(details *chartutils.Details, repoURL string) (*helmchart.Chart, error) {
	vals, err := getValues([]byte(details.Values))
	if err != nil {
		return nil, err
	}
	return &helmchart.Chart{
		Metadata: &helmchart.Metadata{
			Name:    details.ChartName,
			Version: details.Version,
		},
		Values: vals,
	}, nil
}

func getValues(raw []byte) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := k8syaml.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// Init fake
func (f *ChartClient) Init(appRepo *apprepov1alpha1.AppRepository, caCertSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret) error {
	return nil
}

// ChartClientFactory is a fake implementation of the ChartClientFactory interface.
type ChartClientFactory struct{}

// New returns a fake ChartClient
func (c *ChartClientFactory) New(repoType, userAgent string) chartutils.ChartClient {
	return &ChartClient{}
}
