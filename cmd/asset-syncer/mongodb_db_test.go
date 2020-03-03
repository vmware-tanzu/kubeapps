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
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
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
	repoSameNameOtherNamespace := models.Repo{
		Name:      "repo-name",
		Namespace: "other-namespace",
	}

	testCases := []struct {
		name           string
		existingCharts []models.Chart
		charts         []models.Chart
		expectedCharts []models.Chart
		expectedError  error
	}{
		{
			name: "it inserts the charts",
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
			expectedCharts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
		{
			name: "it errors if asked to insert a chart in a different namespace",
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
				models.Chart{Name: "my-chart1", Repo: &repoSameNameOtherNamespace, ID: "foo/bar:123"},
			},
			expectedError: ErrRepoMismatch,
		},
		{
			name: "it updates existing charts in the chart namespace",
			existingCharts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "Old description"},
			},
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
			expectedCharts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
		{
			name: "it removes charts that are not included in the import",
			existingCharts: []models.Chart{
				models.Chart{Name: "my-chart-old", Repo: &repo, ID: "foo/old:123"},
			},
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
			expectedCharts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
		{
			name: "it does not remove charts from other namespaces",
			existingCharts: []models.Chart{
				models.Chart{Name: "my-chart-old", Repo: &repoSameNameOtherNamespace, ID: "foo/other:123"},
			},
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
			expectedCharts: []models.Chart{
				models.Chart{Name: "my-chart-old", Repo: &repoSameNameOtherNamespace, ID: "foo/other:123"},
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
		{
			name: "it does not remove charts from other namespaces even if they have the same repo name",
			existingCharts: []models.Chart{
				models.Chart{Name: "my-chart-old", Repo: &repoSameNameOtherNamespace, ID: "foo/bar:123"},
			},
			charts: []models.Chart{
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
			expectedCharts: []models.Chart{
				models.Chart{Name: "my-chart-old", Repo: &repoSameNameOtherNamespace, ID: "foo/bar:123"},
				models.Chart{Name: "my-chart1", Repo: &repo, ID: "foo/bar:123", Description: "New description"},
				models.Chart{Name: "my-chart2", Repo: &repo, ID: "foo/bar:456"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager, cleanup := getInitializedMongoManager(t)
			defer cleanup()
			if len(tc.existingCharts) > 0 {
				err := manager.importCharts(tc.existingCharts, *tc.existingCharts[0].Repo)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}

			err := manager.importCharts(tc.charts, repo)
			if tc.expectedError != nil {
				if got, want := err, tc.expectedError; !errors.Is(got, want) {
					t.Fatalf("got: %+v, want: %+v", got, want)
				}
			} else if err != nil {
				t.Fatalf("%+v", err)
			}

			opts := cmpopts.EquateEmpty()
			if got, want := getAllCharts(t, manager), tc.expectedCharts; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func getAllCharts(t *testing.T, manager *mongodbAssetManager) []models.Chart {
	var result []models.Chart
	db, closer := manager.DBSession.DB()
	defer closer()

	coll := db.C(dbutils.ChartCollection)
	err := coll.Find(nil).Sort("repo.name", "repo.namespace", "id").All(&result)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return result
}
