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
	"os"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
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

		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		dbConfig := datastore.Config{URL: databaseURL, Database: databaseName, Username: databaseUser, Password: databasePassword}
		kubeappsNamespace := os.Getenv("POD_NAMESPACE")
		manager, err := newManager(databaseType, dbConfig, kubeappsNamespace)
		if err != nil {
			logrus.Fatal(err)
		}
		err = manager.Init()
		if err != nil {
			logrus.Fatal(err)
		}
		defer manager.Close()

		repo := models.Repo{Name: args[0], Namespace: namespace}
		if err = manager.Delete(repo); err != nil {
			logrus.Fatalf("Can't delete chart repository %s from database: %v", args[0], err)
		}

		logrus.Infof("Successfully deleted the chart repository %s from database", args[0])
	},
}
