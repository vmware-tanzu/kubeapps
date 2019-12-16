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

	"github.com/kubeapps/common/datastore"
)

// MongodbAssetManager struct containing mongodb info
type MongodbAssetManager struct {
	mongoConfig datastore.Config
	DBSession   datastore.Session
}

// NewMongoDBManager creates an asset manager for MongoDB
func NewMongoDBManager(config datastore.Config) *MongodbAssetManager {
	return &MongodbAssetManager{config, nil}
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
