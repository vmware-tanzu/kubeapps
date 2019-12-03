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
	"os"
	"time"

	"database/sql"

	"github.com/kubeapps/common/datastore"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync [REPO NAME] [REPO URL]",
	Short: "add a new chart repository, and resync its charts periodically",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			logrus.Info("Need exactly two arguments: [REPO NAME] [REPO URL]")
			cmd.Help()
			return
		}

		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			logrus.Fatal(err)
		}
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		database, err := cmd.Flags().GetString("database-type")
		var manager assetManager
		if database == "mongodb" {
			mongoURL, err := cmd.Flags().GetString("mongo-url")
			if err != nil {
				logrus.Fatal(err)
			}
			mongoDB, err := cmd.Flags().GetString("mongo-database")
			if err != nil {
				logrus.Fatal(err)
			}
			mongoUser, err := cmd.Flags().GetString("mongo-user")
			if err != nil {
				logrus.Fatal(err)
			}
			mongoPW := os.Getenv("MONGO_PASSWORD")
			mongoConfig := datastore.Config{URL: mongoURL, Database: mongoDB, Username: mongoUser, Password: mongoPW}
			dbSession, err := datastore.NewSession(mongoConfig)
			if err != nil {
				logrus.Fatalf("Can't connect to mongoDB: %v", err)
			}
			manager = &mongodbAssetManager{dbSession}
		} else if database == "postgresql" {
			pgHost, err := cmd.Flags().GetString("pg-host")
			if err != nil {
				logrus.Fatal(err)
			}
			pgPort, err := cmd.Flags().GetString("pg-port")
			if err != nil {
				logrus.Fatal(err)
			}
			pgDB, err := cmd.Flags().GetString("pg-database")
			if err != nil {
				logrus.Fatal(err)
			}
			pgUser, err := cmd.Flags().GetString("pg-user")
			if err != nil {
				logrus.Fatal(err)
			}
			pgPW := os.Getenv("POSTGRESQL_PASSWORD")

			connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgHost, pgPort, pgUser, pgPW, pgDB)
			// TODO(andresmgot): Open the DB connection only when needed.
			// We are opening the connection now to be able to test the Delete method
			// but ideally this method should be mocked.
			db, err := sql.Open("postgres", connStr)
			if err != nil {
				logrus.Fatal(err)
			}
			defer db.Close()
			manager = &postgresAssetManager{db}
		} else {
			logrus.Fatalf("Unsupported database type %s", database)
		}

		authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
		r, err := getRepo(args[0], args[1], authorizationHeader)
		if err != nil {
			logrus.Fatal(err)
		}

		// Check if the repo has been already processed
		if manager.RepoAlreadyProcessed(r.Name, r.Checksum) {
			logrus.WithFields(logrus.Fields{"url": r.URL}).Info("Skipping repository since there are no updates")
			return
		}

		index, err := parseRepoIndex(r.Content)
		if err != nil {
			logrus.Fatal(err)
		}

		charts := chartsFromIndex(index, r)
		if len(charts) == 0 {
			logrus.Fatal("no charts in repository index")
		}

		if err = manager.Sync(charts); err != nil {
			logrus.Fatalf("Can't add chart repository to database: %v", err)
		}

		// Update cache in the database
		if err = manager.UpdateLastCheck(r.Name, r.Checksum, time.Now()); err != nil {
			logrus.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{"url": r.URL}).Info("Stored repository update in cache")

		logrus.Infof("Successfully added the chart repository %s to database", args[0])
	},
}
