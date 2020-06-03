/*
Copyright (c) 2019 Bitnami

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

package dbutils

import (
	"fmt"

	"github.com/globalsign/mgo"
	"github.com/kubeapps/common/datastore"
)

const (
	ChartCollection      = "charts"
	RepositoryCollection = "repos"
	ChartFilesCollection = "files"
)

// MongodbAssetManager struct containing mongodb info
type MongodbAssetManager struct {
	mongoConfig       datastore.Config
	DBSession         datastore.Session
	KubeappsNamespace string
}

// NewMongoDBManager creates an asset manager for MongoDB
func NewMongoDBManager(config datastore.Config, kubeappsNamespace string) *MongodbAssetManager {
	return &MongodbAssetManager{config, nil, kubeappsNamespace}
}

// Init creates dbsession
func (m *MongodbAssetManager) Init() error {
	dbSession, err := datastore.NewSession(m.mongoConfig)
	if err != nil {
		return fmt.Errorf("Can't connect to mongoDB: %v", err)
	}
	m.DBSession = dbSession
	return nil
}

// Close (no-op)
func (m *MongodbAssetManager) Close() error {
	return nil
}

// InitCollections ensure indexes of the different collections
func (m *MongodbAssetManager) InitCollections() error {
	db, closer := m.DBSession.DB()
	defer closer()

	err := db.C(ChartCollection).EnsureIndex(mgo.Index{
		Key:        []string{"chart_id", "repo.namespace", "repo.name"},
		Unique:     true,
		DropDups:   true,
		Background: false,
	})
	if err != nil {
		return err
	}
	return db.C(ChartFilesCollection).EnsureIndex(mgo.Index{
		Key:        []string{"file_id", "repo.namespace", "repo.name"},
		Background: false,
	})
}

// InvalidateCache drops the different collections and initialize them again
func (m *MongodbAssetManager) InvalidateCache() error {
	db, closer := m.DBSession.DB()
	defer closer()

	err := db.C(ChartCollection).DropCollection()
	// We ignore "ns not found" which relates to an operation on a non-existent collection.
	if err != nil && err.Error() != "ns not found" {
		return err
	}

	err = m.InitCollections()
	if err != nil {
		return err
	}

	return m.DBSession.Fsync(false)
}
