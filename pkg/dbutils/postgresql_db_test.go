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

// Currently these tests will be skipped entirely if a connection to
// a local postgres database `testdb` cannot be established.
// To have the tests run, simply run
// docker run --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:11.6.0-debian-9-r0
// in another terminal.
package dbutils

import (
	"testing"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
	_ "github.com/lib/pq"
)

func openTestManager(t *testing.T) *PostgresAssetManager {
	pam, err := NewPGManager(datastore.Config{
		URL:      "localhost:5432",
		Database: "testdb",
		Username: "postgres",
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}

	err = pam.Init()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	return pam
}

func getInitializedManager(t *testing.T) (*PostgresAssetManager, func()) {
	pam := openTestManager(t)
	cleanup := func() { pam.Close() }

	err := pam.InvalidateCache()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	return pam, cleanup
}

func TestEnsureRepoExists(t *testing.T) {
	dbutilstest.SkipIfNoPostgres(t)

	testCases := []struct {
		name          string
		existingRepos []models.Repo
		newRepo       models.Repo
		expectedId    int
	}{
		{
			name: "it returns a new ID if it does not yet exist",
			existingRepos: []models.Repo{
				models.Repo{Namespace: "my-namespace", Name: "other-repo"},
				models.Repo{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedId: 3,
		},
		{
			name: "it returns the existing ID if the repo exists in the db",
			existingRepos: []models.Repo{
				models.Repo{Namespace: "my-namespace", Name: "my-name"},
				models.Repo{Namespace: "my-namespace", Name: "other-repo"},
				models.Repo{Namespace: "other-namespace", Name: "my-repo"},
			},
			newRepo:    models.Repo{Namespace: "my-namespace", Name: "my-name"},
			expectedId: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()

			for _, repo := range tc.existingRepos {
				_, err := pam.EnsureRepoExists(repo.Namespace, repo.Name)
				if err != nil {
					t.Fatalf("%+v", err)
				}
			}

			id, err := pam.EnsureRepoExists(tc.newRepo.Namespace, tc.newRepo.Name)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := id, tc.expectedId; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}
