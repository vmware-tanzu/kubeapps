/*
Copyright (c) 2020 Bitnami

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

// Currently these tests will be skipped entirely unless the
// ENABLE_MONGO_INTEGRATION_TESTS env var is set.
// Run the local mongodb with
// docker run --publish 27017:27017 -e MONGODB_ROOT_PASSWORD=testpassword -e ALLOW_EMPTY_PASSWORD=yes bitnami/mongodb:4.2.3-debian-10-r31
// in another terminal.
package main

import (
	"testing"

	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest/mgtest"
)

func getInitializedMongoManager(t *testing.T) (*mongodbAssetManager, func()) {
	manager, cleanup := mgtest.GetInitializedManager(t)
	return &mongodbAssetManager{manager}, cleanup
}

func TestMongoImportCharts(t *testing.T) {
	mgtest.SkipIfNoDB(t)

	repo := models.Repo{
		Name:      "repo-name",
		Namespace: "repo-namespace",
	}

	testCases := []struct {
		name   string
		charts []models.Chart
	}{
		{
			name: "it inserts the charts",
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, cleanup := getInitializedMongoManager(t)
			defer cleanup()

			err := manager.importCharts(tc.charts)
			if err != nil {
				t.Errorf("%+v", err)
			}
		})
	}
}
