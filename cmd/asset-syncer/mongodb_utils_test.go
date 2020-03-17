/*
Copyright (c) 2018 The Helm Authors

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
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/globalsign/mgo/bson"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/common/datastore/mockstore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
	"github.com/stretchr/testify/mock"
)

func getMockManager(m *mock.Mock) *mongodbAssetManager {
	dbSession := mockstore.NewMockSession(m)
	man := dbutils.NewMongoDBManager(datastore.Config{}, dbutilstest.KubeappsTestNamespace)
	man.DBSession = dbSession
	return &mongodbAssetManager{man}
}

func Test_importCharts(t *testing.T) {
	m := &mock.Mock{}
	// Ensure Upsert func is called with some arguments
	m.On("Upsert", mock.Anything)
	m.On("RemoveAll", mock.Anything)
	index, _ := parseRepoIndex([]byte(validRepoIndexYAML))
	repo := models.Repo{
		Name:      "repo-name",
		Namespace: "repo-namespace",
		URL:       "http://testrepo.example.com",
	}
	charts := chartsFromIndex(index, &repo)
	manager := getMockManager(m)
	manager.importCharts(charts, repo)

	m.AssertExpectations(t)
	// The Bulk Upsert method takes an array that consists of a selector followed by an interface to upsert.
	// So for x charts to upsert, there should be x*2 elements (each chart has it's own selector)
	// e.g. [selector1, chart1, selector2, chart2, ...]
	args := m.Calls[0].Arguments.Get(0).([]interface{})
	assert.Equal(t, len(args), len(charts)*2, "number of selector, chart pairs to upsert")
	for i := 0; i < len(args); i += 2 {
		m := args[i+1].(bson.M)
		c := m["$set"].(models.Chart)
		assert.Equal(t, args[i], bson.M{"chart_id": "repo-name/" + c.Name, "repo.name": "repo-name", "repo.namespace": "repo-namespace"}, "selector")
	}
}

func Test_DeleteRepo(t *testing.T) {
	m := &mock.Mock{}
	repo := models.Repo{Name: "repo-name", Namespace: "repo-namespace"}
	m.On("RemoveAll", bson.M{
		"repo.name":      repo.Name,
		"repo.namespace": repo.Namespace,
	})
	m.On("RemoveAll", bson.M{
		"name":      repo.Name,
		"namespace": repo.Namespace,
	})
	manager := getMockManager(m)
	err := manager.Delete(repo)
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
	m.AssertExpectations(t)
}

func Test_emptyChartRepo(t *testing.T) {
	r := &models.Repo{Name: "testRepo", URL: "https://my.examplerepo.com"}
	i, err := parseRepoIndex(emptyRepoIndexYAMLBytes)
	assert.NoErr(t, err)
	charts := chartsFromIndex(i, r)
	assert.Equal(t, len(charts), 0, "charts")
}

func Test_repoAlreadyProcessed(t *testing.T) {
	tests := []struct {
		name            string
		checksum        string
		mockedLastCheck models.RepoCheck
		processed       bool
	}{
		{"not processed yet", "bar", models.RepoCheck{}, false},
		{"already processed", "bar", models.RepoCheck{Checksum: "bar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mock.Mock{}
			repo := &models.RepoCheck{}
			m.On("One", repo).Run(func(args mock.Arguments) {
				*args.Get(0).(*models.RepoCheck) = tt.mockedLastCheck
			}).Return(nil)
			manager := getMockManager(m)
			res := manager.RepoAlreadyProcessed(models.Repo{Namespace: "repo-namespace", Name: "repo-name"}, tt.checksum)
			if res != tt.processed {
				t.Errorf("Expected alreadyProcessed to be %v got %v", tt.processed, res)
			}
		})
	}
}

func Test_updateLastCheck(t *testing.T) {
	m := &mock.Mock{}
	const (
		repoNamespace = "repoNamespace"
		repoName      = "foo"
		checksum      = "bar"
	)
	now := time.Now()
	m.On("Upsert", bson.M{"name": repoName, "namespace": repoNamespace}, bson.M{"$set": bson.M{"last_update": now, "checksum": checksum}}).Return(nil)
	manager := getMockManager(m)
	err := manager.UpdateLastCheck(repoNamespace, repoName, checksum, now)
	m.AssertExpectations(t)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if len(m.Calls) != 1 {
		t.Errorf("Expected one call got %d", len(m.Calls))
	}
}
