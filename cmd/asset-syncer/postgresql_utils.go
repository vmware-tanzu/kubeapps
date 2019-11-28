/*
Copyright (c) 2018 Bitnami

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
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	chartTable      = "charts"
	repositoryTable = "repos"
	chartFilesTable = "files"
)

type postgresDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type postgresAssetManager struct {
	db postgresDB
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
func (m *postgresAssetManager) Sync(repoName, repoURL string, authorizationHeader string) error {
	// TODO(andresmgot): Implement this :)
	return nil
}

func (m *postgresAssetManager) Delete(repoName string) error {
	tables := []string{chartTable, repositoryTable, chartFilesTable}
	for _, table := range tables {
		_, err := m.db.Query(fmt.Sprintf("DELETE FROM %s WHERE info -> 'repo' ->> 'name' = '%s'", table, repoName))
		if err != nil {
			return err
		}
	}
	return nil
}
