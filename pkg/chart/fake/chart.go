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

package fake

import (
	"encoding/json"
	"net/http"

	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	chart3 "helm.sh/helm/v3/pkg/chart"
	chart2 "k8s.io/helm/pkg/proto/hapi/chart"
	"sigs.k8s.io/yaml"
)

type FakeChart struct{}

func (f *FakeChart) ParseDetails(data []byte) (*chartUtils.Details, error) {
	details := &chartUtils.Details{}
	err := json.Unmarshal(data, details)
	return details, err
}

func (f *FakeChart) GetChart(details *chartUtils.Details, netClient chartUtils.HTTPClient, requireV1Support bool) (*chartUtils.ChartMultiVersion, error) {
	vals, err := getValues([]byte(details.Values))
	if err != nil {
		return nil, err
	}
	return &chartUtils.ChartMultiVersion{
		Helm2Chart: &chart2.Chart{
			Metadata: &chart2.Metadata{
				Name: details.ChartName,
			},
			Values: &chart2.Config{
				Raw: details.Values,
			},
		},
		Helm3Chart: &chart3.Chart{
			Metadata: &chart3.Metadata{
				Name: details.ChartName,
			},
			Values: vals,
		},
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

func (f *FakeChart) InitNetClient(details *chartUtils.Details) (chartUtils.HTTPClient, error) {
	return &http.Client{}, nil
}
