// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	chartUtils "github.com/vmware-tanzu/kubeapps/pkg/chart"
	"helm.sh/helm/v3/pkg/chart"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// ChartClient implements Resolver inteface
type ChartClient struct{}

// GetChart fake
func (f *ChartClient) GetChart(details *chartUtils.Details, repoURL string) (*chart.Chart, error) {
	vals, err := getValues([]byte(details.Values))
	if err != nil {
		return nil, err
	}
	return &chart.Chart{
		Metadata: &chart.Metadata{
			Name:    details.ChartName,
			Version: details.Version,
		},
		Values: vals,
	}, nil
}

func getValues(raw []byte) (map[string]interface{}, error) {
	values := make(map[string]interface{})
	err := yaml.Unmarshal(raw, &values)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// Init fake
func (f *ChartClient) Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	return nil
}

// ChartClientFactory is a fake implementation of the ChartClientFactory interface.
type ChartClientFactory struct{}

// New returns a fake ChartClient
func (c *ChartClientFactory) New(repoType, userAgent string) chartUtils.ChartClient {
	return &ChartClient{}
}
