// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	log "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/credentialprovider"
)

func Sync(serveOpts Config, version string, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("Need exactly three arguments: [REPO NAME] [REPO URL] [REPO TYPE] (got %v)", len(args))
	}

	dbConfig := dbutils.Config{URL: serveOpts.DatabaseURL, Database: serveOpts.DatabaseName, Username: serveOpts.DatabaseUser, Password: serveOpts.DatabasePassword}
	globalPackagingNamespace := serveOpts.GlobalPackagingNamespace
	manager, err := newManager(dbConfig, globalPackagingNamespace)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	err = manager.Init()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	defer manager.Close()

	netClient, err := httpclient.NewWithCertFile(additionalCAFile, serveOpts.TlsInsecureSkipVerify)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	authorizationHeader := serveOpts.AuthorizationHeader
	// The auth header may be a dockerconfig that we need to parse
	if serveOpts.DockerConfigJson != "" {
		dockerConfig := &credentialprovider.DockerConfigJSON{}
		err = json.Unmarshal([]byte(serveOpts.DockerConfigJson), dockerConfig)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
		authorizationHeader, err = kube.GetAuthHeaderFromDockerConfig(dockerConfig)
		if err != nil {
			return fmt.Errorf("Error: %v", err)
		}
	}

	filters, err := parseFilters(serveOpts.FilterRules)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	var repoIface Repo
	if args[2] == "helm" {
		repoIface, err = getHelmRepo(serveOpts.Namespace, args[0], args[1], authorizationHeader, filters, netClient, serveOpts.UserAgent)
	} else {
		repoIface, err = getOCIRepo(serveOpts.Namespace, args[0], args[1], authorizationHeader, filters, serveOpts.OciRepositories, netClient)
	}
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	repo := repoIface.Repo()
	checksum, err := repoIface.Checksum()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	// Check if the repo has been already processed
	lastChecksum := manager.LastChecksum(models.Repo{Namespace: repo.Namespace, Name: repo.Name})
	log.Infof("Last checksum: %v", lastChecksum)
	if lastChecksum == checksum {
		log.Infof("Skipping repository since there are no updatesrepo.URL= %v", repo.URL)
		return nil
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
			return fmt.Errorf("Error: %v", err)
		}
		if len(charts) == 0 {
			log.Infof("No charts in repository to be synced, repo.URL= %v", repo.URL)
			return nil
		}
		if err = manager.Sync(models.Repo{Name: repo.Name, Namespace: repo.Namespace}, charts); err != nil {
			return fmt.Errorf("Can't add chart repository to database: %v", err)
		}

		// Fetch and store chart icons
		fImporter := fileImporter{manager, netClient}
		fImporter.fetchFiles(charts, repoIface, serveOpts.UserAgent, serveOpts.PassCredentials)
		log.Infof("Repository synced, shallow=%v", fetchLatestOnly)

	}

	// Update cache in the database
	if err = manager.UpdateLastCheck(repo.Namespace, repo.Name, checksum, time.Now()); err != nil {
		return fmt.Errorf("Error: %v", err)
	}

	log.Infof("Stored repository update in cache, repo.URL= %v", repo.URL)
	log.Infof("Successfully added the package repository %s to database", args[0])
	return nil
}
