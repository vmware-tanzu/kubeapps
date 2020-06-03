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
	"fmt"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
)

var ErrRepoMismatch = fmt.Errorf("chart repository did not match import repository")

type mongodbAssetManager struct {
	*dbutils.MongodbAssetManager
}

func newMongoDBManager(config datastore.Config, kubeappsNamespace string) assetManager {
	m := dbutils.NewMongoDBManager(config, kubeappsNamespace)
	return &mongodbAssetManager{m}
}

// Syncing is performed in the following steps:
// 1. Update database to match chart metadata from index
// 2. Concurrently process icons for charts (concurrently)
// 3. Concurrently process the README and values.yaml for the latest chart version of each chart
// 4. Concurrently process READMEs and values.yaml for historic chart versions
//
// These steps are processed in this way to ensure relevant chart data is
// imported into the database as fast as possible. E.g. we want all icons for
// charts before fetching readmes for each chart and version pair.
func (m *mongodbAssetManager) Sync(repo models.Repo, charts []models.Chart) error {
	err := m.InitCollections()
	if err != nil {
		return err
	}

	return m.importCharts(charts, repo)
}

func (m *mongodbAssetManager) RepoAlreadyProcessed(repo models.Repo, checksum string) bool {
	db, closer := m.DBSession.DB()
	defer closer()
	lastCheck := &models.RepoCheck{}
	err := db.C(dbutils.RepositoryCollection).Find(bson.M{"name": repo.Name, "namespace": repo.Namespace}).One(lastCheck)
	return err == nil && checksum == lastCheck.Checksum
}

func (m *mongodbAssetManager) UpdateLastCheck(repoNamespace, repoName, checksum string, now time.Time) error {
	db, closer := m.DBSession.DB()
	defer closer()
	_, err := db.C(dbutils.RepositoryCollection).Upsert(bson.M{"name": repoName, "namespace": repoNamespace}, bson.M{"$set": bson.M{"last_update": now, "checksum": checksum}})
	return err
}

func (m *mongodbAssetManager) Delete(repo models.Repo) error {
	db, closer := m.DBSession.DB()
	defer closer()
	_, err := db.C(dbutils.ChartCollection).RemoveAll(bson.M{
		"repo.name":      repo.Name,
		"repo.namespace": repo.Namespace,
	})
	if err != nil {
		return err
	}

	_, err = db.C(dbutils.ChartFilesCollection).RemoveAll(bson.M{
		"repo.name":      repo.Name,
		"repo.namespace": repo.Namespace,
	})
	if err != nil {
		return err
	}

	_, err = db.C(dbutils.RepositoryCollection).RemoveAll(bson.M{
		"name":      repo.Name,
		"namespace": repo.Namespace,
	})
	return err
}

func (m *mongodbAssetManager) importCharts(charts []models.Chart, repo models.Repo) error {
	var pairs []interface{}
	var chartIDs []string
	for _, c := range charts {
		if c.Repo == nil || c.Repo.Namespace != repo.Namespace || c.Repo.Name != repo.Name {
			return fmt.Errorf("%w: chart repo: %+v, import repo: %+v", ErrRepoMismatch, c.Repo, repo)
		}
		chartIDs = append(chartIDs, c.ID)
		// charts to upsert - pair of selector, chart
		// Mongodb generates the unique _id, we rely on the compound unique index on chart_id and repo.
		pairs = append(pairs, bson.M{"chart_id": c.ID, "repo.name": repo.Name, "repo.namespace": repo.Namespace}, bson.M{"$set": c})
	}

	db, closer := m.DBSession.DB()
	defer closer()
	bulk := db.C(dbutils.ChartCollection).Bulk()

	// Upsert pairs of selectors, charts
	bulk.Upsert(pairs...)

	// Remove charts no longer existing in index
	bulk.RemoveAll(bson.M{
		"chart_id": bson.M{
			"$nin": chartIDs,
		},
		"repo.name":      repo.Name,
		"repo.namespace": repo.Namespace,
	})

	_, err := bulk.Run()
	return err
}

func (m *mongodbAssetManager) updateIcon(repo models.Repo, data []byte, contentType, ID string) error {
	db, closer := m.DBSession.DB()
	defer closer()
	_, err := db.C(dbutils.ChartCollection).Upsert(bson.M{"chart_id": ID, "repo.name": repo.Name, "repo.namespace": repo.Namespace}, bson.M{"$set": bson.M{"raw_icon": data, "icon_content_type": contentType}})
	return err
}

func (m *mongodbAssetManager) filesExist(repo models.Repo, chartFilesID, digest string) bool {
	db, closer := m.DBSession.DB()
	defer closer()
	err := db.C(dbutils.ChartFilesCollection).Find(bson.M{"file_id": chartFilesID, "repo.name": repo.Name, "repo.namespace": repo.Namespace, "digest": digest}).One(&models.ChartFiles{})
	return err == nil
}

func (m *mongodbAssetManager) insertFiles(chartId string, files models.ChartFiles) error {
	db, closer := m.DBSession.DB()
	defer closer()

	_, err := db.C(dbutils.ChartFilesCollection).Upsert(bson.M{"file_id": files.ID, "repo.name": files.Repo.Name, "repo.namespace": files.Repo.Namespace}, files)
	return err
}
