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
)

const (
	chartCollection      = "charts"
	repositoryCollection = "repos"
	chartFilesCollection = "files"
)

type mongodbAssetManager struct {
	mongoConfig datastore.Config
	dbSession   datastore.Session
}

func newMongoDBManager(config datastore.Config) assetManager {
	return &mongodbAssetManager{config, nil}
}

func (m *mongodbAssetManager) Init() error {
	dbSession, err := datastore.NewSession(m.mongoConfig)
	if err != nil {
		return fmt.Errorf("Can't connect to mongoDB: %v", err)
	}
	m.dbSession = dbSession
	return nil
}

func (m *mongodbAssetManager) Close() error {
	return nil
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
func (m *mongodbAssetManager) Sync(charts []chart) error {
	return m.importCharts(charts)
}

func (m *mongodbAssetManager) RepoAlreadyProcessed(repoName string, checksum string) bool {
	db, closer := m.dbSession.DB()
	defer closer()
	lastCheck := &repoCheck{}
	err := db.C(repositoryCollection).Find(bson.M{"_id": repoName}).One(lastCheck)
	return err == nil && checksum == lastCheck.Checksum
}

func (m *mongodbAssetManager) UpdateLastCheck(repoName string, checksum string, now time.Time) error {
	db, closer := m.dbSession.DB()
	defer closer()
	_, err := db.C(repositoryCollection).UpsertId(repoName, bson.M{"$set": bson.M{"last_update": now, "checksum": checksum}})
	return err
}

func (m *mongodbAssetManager) Delete(repoName string) error {
	db, closer := m.dbSession.DB()
	defer closer()
	_, err := db.C(chartCollection).RemoveAll(bson.M{
		"repo.name": repoName,
	})
	if err != nil {
		return err
	}

	_, err = db.C(chartFilesCollection).RemoveAll(bson.M{
		"repo.name": repoName,
	})
	if err != nil {
		return err
	}

	_, err = db.C(repositoryCollection).RemoveAll(bson.M{
		"_id": repoName,
	})
	return err
}

func (m *mongodbAssetManager) importCharts(charts []chart) error {
	var pairs []interface{}
	var chartIDs []string
	for _, c := range charts {
		chartIDs = append(chartIDs, c.ID)
		// charts to upsert - pair of selector, chart
		pairs = append(pairs, bson.M{"_id": c.ID}, c)
	}

	db, closer := m.dbSession.DB()
	defer closer()
	bulk := db.C(chartCollection).Bulk()

	// Upsert pairs of selectors, charts
	bulk.Upsert(pairs...)

	// Remove charts no longer existing in index
	bulk.RemoveAll(bson.M{
		"_id": bson.M{
			"$nin": chartIDs,
		},
		"repo.name": charts[0].Repo.Name,
	})

	_, err := bulk.Run()
	return err
}

func (m *mongodbAssetManager) updateIcon(data []byte, contentType, ID string) error {
	db, closer := m.dbSession.DB()
	defer closer()
	return db.C(chartCollection).UpdateId(ID, bson.M{"$set": bson.M{"raw_icon": data, "icon_content_type": contentType}})
}

func (m *mongodbAssetManager) filesExist(chartFilesID, digest string) bool {
	db, closer := m.dbSession.DB()
	defer closer()
	err := db.C(chartFilesCollection).Find(bson.M{"_id": chartFilesID, "digest": digest}).One(&chartFiles{})
	return err == nil
}

func (m *mongodbAssetManager) insertFiles(chartFilesID string, files chartFiles) error {
	db, closer := m.dbSession.DB()
	defer closer()
	_, err := db.C(chartFilesCollection).UpsertId(chartFilesID, files)
	return err
}
