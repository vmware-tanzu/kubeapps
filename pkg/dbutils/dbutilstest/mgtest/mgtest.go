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

package mgtest

import (
	"testing"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
)

const (
	// EnvvarMongoTests enables tests that run against a local mongo db
	EnvvarMongoTests = "ENABLE_MONGO_INTEGRATION_TESTS"
)

func SkipIfNoDB(t *testing.T) {
	if !dbutilstest.IsEnvVarTrue(t, EnvvarMongoTests) {
		t.Skipf("skipping mongodb tests as %q not set to be true", EnvvarMongoTests)
	}
}

func OpenTestManager(t *testing.T) *dbutils.MongodbAssetManager {
	manager := dbutils.NewMongoDBManager(datastore.Config{
		URL:      "localhost:27017",
		Username: "root",
		Password: "testpassword",
	}, dbutilstest.KubeappsTestNamespace)

	err := manager.Init()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return manager
}

// GetInitializedManager returns an initialized mongodb manager ready for testing.
func GetInitializedManager(t *testing.T) (*dbutils.MongodbAssetManager, func()) {
	manager := OpenTestManager(t)
	cleanup := func() { manager.Close() }

	err := manager.InvalidateCache()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	return manager, cleanup
}
