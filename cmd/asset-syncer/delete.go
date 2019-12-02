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

	"database/sql"

	"github.com/kubeapps/common/datastore"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [REPO NAME]",
	Short: "delete a chart repository",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logrus.Info("Need exactly one argument: [REPO NAME]")
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

		if err = manager.Delete(args[0]); err != nil {
			logrus.Fatalf("Can't delete chart repository %s from database: %v", args[0], err)
		}

		logrus.Infof("Successfully deleted the chart repository %s from database", args[0])
	},
}
