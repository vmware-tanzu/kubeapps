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
	"encoding/json"
	"os"
	"time"

	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/kubernetes/pkg/credentialprovider"
)

var syncCmd = &cobra.Command{
	Use:   "sync [REPO NAME] [REPO URL] [REPO TYPE]",
	Short: "add a new chart repository, and resync its charts periodically",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			logrus.Info("Need exactly two arguments: [REPO NAME] [REPO URL] [REPO TYPE]")
			cmd.Help()
			return
		}

		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		dbConfig := datastore.Config{URL: databaseURL, Database: databaseName, Username: databaseUser, Password: databasePassword}
		kubeappsNamespace := os.Getenv("POD_NAMESPACE")
		manager, err := newManager(dbConfig, kubeappsNamespace)
		if err != nil {
			logrus.Fatal(err)
		}
		err = manager.Init()
		if err != nil {
			logrus.Fatal(err)
		}
		defer manager.Close()

		netClient, err := httpclient.NewWithCertFile(additionalCAFile, tlsInsecureSkipVerify)
		if err != nil {
			logrus.Fatal(err)
		}

		authorizationHeader := os.Getenv("AUTHORIZATION_HEADER")
		// The auth header may be a dockerconfig that we need to parse
		if os.Getenv("DOCKER_CONFIG_JSON") != "" {
			dockerConfig := &credentialprovider.DockerConfigJSON{}
			err = json.Unmarshal([]byte(os.Getenv("DOCKER_CONFIG_JSON")), dockerConfig)
			if err != nil {
				logrus.Fatal(err)
			}
			authorizationHeader, err = kube.GetAuthHeaderFromDockerConfig(dockerConfig)
		}

		filters, err := parseFilters(filterRules)
		if err != nil {
			logrus.Fatal(err)
		}

		var repoIface Repo
		if args[2] == "helm" {
			repoIface, err = getHelmRepo(namespace, args[0], args[1], authorizationHeader, filters, netClient)
		} else {
			repoIface, err = getOCIRepo(namespace, args[0], args[1], authorizationHeader, filters, ociRepositories, netClient)
		}
		if err != nil {
			logrus.Fatal(err)
		}
		repo := repoIface.Repo()
		checksum, err := repoIface.Checksum()
		if err != nil {
			logrus.Fatal(err)
		}

		// Check if the repo has been already processed
		lastChecksum := manager.LastChecksum(models.Repo{Namespace: repo.Namespace, Name: repo.Name})
		logrus.Infof("Last checksum: %v", lastChecksum)
		if lastChecksum == checksum {
			logrus.WithFields(logrus.Fields{"url": repo.URL}).Info("Skipping repository since there are no updates")
			return
		}

		// First filter the list of charts (still without applying custom filters)
		repoIface.FilterIndex()

		fetchLatestOnlySlice := []bool{false}
		if lastChecksum == "" {
			// If the repo has never been processed, run first a shallow sync to give early feedback
			// then sync all the repositories
			fetchLatestOnlySlice = []bool{true, false}
		}

		for _, fetchLatestOnly := range fetchLatestOnlySlice {
			charts, err := repoIface.Charts(fetchLatestOnly)
			if err != nil {
				logrus.Fatal(err)
			}
			if err = manager.Sync(models.Repo{Name: repo.Name, Namespace: repo.Namespace}, charts); err != nil {
				logrus.Fatalf("Can't add chart repository to database: %v", err)
			}

			// Fetch and store chart icons
			fImporter := fileImporter{manager, netClient}
			fImporter.fetchFiles(charts, repoIface)
			logrus.WithFields(logrus.Fields{"shallow": fetchLatestOnly}).Info("Repository synced")
		}

		// Update cache in the database
		if err = manager.UpdateLastCheck(repo.Namespace, repo.Name, checksum, time.Now()); err != nil {
			logrus.Fatal(err)
		}
		logrus.WithFields(logrus.Fields{"url": repo.URL}).Info("Stored repository update in cache")

		logrus.Infof("Successfully added the chart repository %s to database", args[0])
	},
}
