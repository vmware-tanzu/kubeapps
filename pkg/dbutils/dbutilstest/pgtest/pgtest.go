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

package pgtest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
)

const (
	// EnvvarPostgresTests enables tests that run against a local postgres db
	EnvvarPostgresTests = "ENABLE_PG_INTEGRATION_TESTS"
)

func SkipIfNoDB(t *testing.T) {
	if !dbutilstest.IsEnvVarTrue(t, EnvvarPostgresTests) {
		t.Skipf("skipping postgres tests as %q not set to be true", EnvvarPostgresTests)
	}
}

func openTestManager(t *testing.T) *dbutils.PostgresAssetManager {
	pam, err := dbutils.NewPGManager(dbutils.Config{
		URL:      "localhost:5432",
		Database: "testdb",
		Username: "postgres",
	}, dbutilstest.KubeappsTestNamespace)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = pam.Init()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return pam
}

// GetInitializedPGManager returns an initialized postgres manager ready for testing.
func GetInitializedManager(t *testing.T) (*dbutils.PostgresAssetManager, func()) {
	pam := openTestManager(t)
	cleanup := func() { pam.Close() }

	err := pam.InvalidateCache()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	return pam, cleanup
}

func CountRows(t *testing.T, db dbutils.PostgresDB, table string) int {
	var count int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return count
}

func EnsureChartsExist(t *testing.T, pam dbutils.PostgresAssetManagerIface, charts []models.Chart, repo models.Repo) {
	_, err := pam.EnsureRepoExists(repo.Namespace, repo.Name)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	for _, chart := range charts {
		d, err := json.Marshal(chart)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		_, err = pam.GetDB().Exec(fmt.Sprintf(`INSERT INTO %s (repo_namespace, repo_name, chart_id, info)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (chart_id, repo_namespace, repo_name)
		DO UPDATE SET info = $4
		`, dbutils.ChartTable), repo.Namespace, repo.Name, chart.ID, string(d))
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
}
