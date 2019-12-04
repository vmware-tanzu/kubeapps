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
	"errors"
	"fmt"
	"image"
	"io"
	"testing"
	"time"

	"github.com/arschles/assert"
	"github.com/globalsign/mgo/bson"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/common/datastore/mockstore"
	"github.com/stretchr/testify/mock"
)

func Test_importCharts(t *testing.T) {
	m := &mock.Mock{}
	// Ensure Upsert func is called with some arguments
	m.On("Upsert", mock.Anything)
	m.On("RemoveAll", mock.Anything)
	dbSession := mockstore.NewMockSession(m)
	index, _ := parseRepoIndex([]byte(validRepoIndexYAML))
	charts := chartsFromIndex(index, &repo{Name: "test", URL: "http://testrepo.com"})
	manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
	manager.importCharts(charts)

	m.AssertExpectations(t)
	// The Bulk Upsert method takes an array that consists of a selector followed by an interface to upsert.
	// So for x charts to upsert, there should be x*2 elements (each chart has it's own selector)
	// e.g. [selector1, chart1, selector2, chart2, ...]
	args := m.Calls[0].Arguments.Get(0).([]interface{})
	assert.Equal(t, len(args), len(charts)*2, "number of selector, chart pairs to upsert")
	for i := 0; i < len(args); i += 2 {
		c := args[i+1].(chart)
		assert.Equal(t, args[i], bson.M{"_id": "test/" + c.Name}, "selector")
	}
}

func Test_DeleteRepo(t *testing.T) {
	m := &mock.Mock{}
	m.On("RemoveAll", bson.M{
		"repo.name": "test",
	})
	m.On("RemoveAll", bson.M{
		"_id": "test",
	})
	dbSession := mockstore.NewMockSession(m)

	mongoManager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
	err := mongoManager.Delete("test")
	if err != nil {
		t.Errorf("failed to delete chart repo test: %v", err)
	}
	m.AssertExpectations(t)
}

func Test_fetchAndImportIcon(t *testing.T) {
	t.Run("no icon", func(t *testing.T) {
		m := mock.Mock{}
		dbSession := mockstore.NewMockSession(&m)
		c := chart{ID: "test/acs-engine-autoscaler"}
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.NoErr(t, manager.fetchAndImportIcon(c))
	})

	index, _ := parseRepoIndex([]byte(validRepoIndexYAML))
	charts := chartsFromIndex(index, &repo{Name: "test", URL: "http://testrepo.com"})

	t.Run("failed download", func(t *testing.T) {
		netClient = &badHTTPClient{}
		c := charts[0]
		m := mock.Mock{}
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.Err(t, fmt.Errorf("500 %s", c.Icon), manager.fetchAndImportIcon(c))
	})

	t.Run("bad icon", func(t *testing.T) {
		netClient = &badIconClient{}
		c := charts[0]
		m := mock.Mock{}
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.Err(t, image.ErrFormat, manager.fetchAndImportIcon(c))
	})

	t.Run("valid icon", func(t *testing.T) {
		netClient = &goodIconClient{}
		c := charts[0]
		m := mock.Mock{}
		dbSession := mockstore.NewMockSession(&m)
		m.On("UpdateId", c.ID, bson.M{"$set": bson.M{"raw_icon": iconBytes(), "icon_content_type": "image/png"}}).Return(nil)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.NoErr(t, manager.fetchAndImportIcon(c))
		m.AssertExpectations(t)
	})

	t.Run("valid SVG icon", func(t *testing.T) {
		netClient = &svgIconClient{}
		c := chart{
			ID:   "foo",
			Icon: "https://foo/bar/logo.svg",
			Repo: &repo{},
		}
		m := mock.Mock{}
		dbSession := mockstore.NewMockSession(&m)
		m.On("UpdateId", c.ID, bson.M{"$set": bson.M{"raw_icon": []byte("foo"), "icon_content_type": "image/svg"}}).Return(nil)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.NoErr(t, manager.fetchAndImportIcon(c))
		m.AssertExpectations(t)
	})
}

func Test_fetchAndImportFiles(t *testing.T) {
	index, _ := parseRepoIndex([]byte(validRepoIndexYAML))
	charts := chartsFromIndex(index, &repo{Name: "test", URL: "http://testrepo.com", AuthorizationHeader: "Bearer ThisSecretAccessTokenAuthenticatesTheClient1s"})
	cv := charts[0].ChartVersions[0]

	t.Run("http error", func(t *testing.T) {
		m := mock.Mock{}
		m.On("One", mock.Anything).Return(errors.New("return an error when checking if readme already exists to force fetching"))
		dbSession := mockstore.NewMockSession(&m)
		netClient = &badHTTPClient{}
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		assert.Err(t, io.EOF, manager.fetchAndImportFiles(charts[0].Name, charts[0].Repo, cv))
	})

	t.Run("file not found", func(t *testing.T) {
		netClient = &goodTarballClient{c: charts[0], skipValues: true, skipReadme: true, skipSchema: true}
		m := mock.Mock{}
		m.On("One", mock.Anything).Return(errors.New("return an error when checking if files already exists to force fetching"))
		chartFilesID := fmt.Sprintf("%s/%s-%s", charts[0].Repo.Name, charts[0].Name, cv.Version)
		m.On("UpsertId", chartFilesID, chartFiles{chartFilesID, "", "", "", charts[0].Repo, cv.Digest})
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		err := manager.fetchAndImportFiles(charts[0].Name, charts[0].Repo, cv)
		assert.NoErr(t, err)
		m.AssertExpectations(t)
	})

	t.Run("authenticated request", func(t *testing.T) {
		netClient = &authenticatedTarballClient{c: charts[0]}
		m := mock.Mock{}
		m.On("One", mock.Anything).Return(errors.New("return an error when checking if files already exists to force fetching"))
		chartFilesID := fmt.Sprintf("%s/%s-%s", charts[0].Repo.Name, charts[0].Name, cv.Version)
		m.On("UpsertId", chartFilesID, chartFiles{chartFilesID, testChartReadme, testChartValues, testChartSchema, charts[0].Repo, cv.Digest})
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		err := manager.fetchAndImportFiles(charts[0].Name, charts[0].Repo, cv)
		assert.NoErr(t, err)
		m.AssertExpectations(t)
	})

	t.Run("valid tarball", func(t *testing.T) {
		netClient = &goodTarballClient{c: charts[0]}
		m := mock.Mock{}
		m.On("One", mock.Anything).Return(errors.New("return an error when checking if files already exists to force fetching"))
		chartFilesID := fmt.Sprintf("%s/%s-%s", charts[0].Repo.Name, charts[0].Name, cv.Version)
		m.On("UpsertId", chartFilesID, chartFiles{chartFilesID, testChartReadme, testChartValues, testChartSchema, charts[0].Repo, cv.Digest})
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		err := manager.fetchAndImportFiles(charts[0].Name, charts[0].Repo, cv)
		assert.NoErr(t, err)
		m.AssertExpectations(t)
	})

	t.Run("file exists", func(t *testing.T) {
		m := mock.Mock{}
		// don't return an error when checking if files already exists
		m.On("One", mock.Anything).Return(nil)
		dbSession := mockstore.NewMockSession(&m)
		manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
		err := manager.fetchAndImportFiles(charts[0].Name, charts[0].Repo, cv)
		assert.NoErr(t, err)
		m.AssertNotCalled(t, "UpsertId", mock.Anything, mock.Anything)
	})
}

func Test_emptyChartRepo(t *testing.T) {
	r := &repo{Name: "testRepo", URL: "https://my.examplerepo.com", Checksum: "123"}
	i, err := parseRepoIndex(emptyRepoIndexYAMLBytes)
	assert.NoErr(t, err)
	charts := chartsFromIndex(i, r)
	assert.Equal(t, len(charts), 0, "charts")
}

func Test_repoAlreadyProcessed(t *testing.T) {
	tests := []struct {
		name            string
		checksum        string
		mockedLastCheck repoCheck
		processed       bool
	}{
		{"not processed yet", "bar", repoCheck{}, false},
		{"already processed", "bar", repoCheck{Checksum: "bar"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mock.Mock{}
			repo := &repoCheck{}
			m.On("One", repo).Run(func(args mock.Arguments) {
				*args.Get(0).(*repoCheck) = tt.mockedLastCheck
			}).Return(nil)
			dbSession := mockstore.NewMockSession(&m)
			manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
			res := manager.RepoAlreadyProcessed("", tt.checksum)
			if res != tt.processed {
				t.Errorf("Expected alreadyProcessed to be %v got %v", tt.processed, res)
			}
		})
	}
}

func Test_updateLastCheck(t *testing.T) {
	m := mock.Mock{}
	repoName := "foo"
	checksum := "bar"
	now := time.Now()
	m.On("UpsertId", repoName, bson.M{"$set": bson.M{"last_update": now, "checksum": checksum}}).Return(nil)
	dbSession := mockstore.NewMockSession(&m)
	manager := mongodbAssetManager{mongoConfig: datastore.Config{}, dbSession: dbSession}
	err := manager.UpdateLastCheck(repoName, checksum, now)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if len(m.Calls) != 1 {
		t.Errorf("Expected one call got %d", len(m.Calls))
	}
}
