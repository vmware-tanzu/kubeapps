// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalog_client"
	log "k8s.io/klog/v2"
)

func Sync(serveOpts Config, version string, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("need exactly three arguments: [REPO NAME] [REPO URL] [REPO TYPE] (got %v)", len(args))
	}

	ctx := context.Background()
	dbConfig := dbutils.Config{URL: serveOpts.DatabaseURL, Database: serveOpts.DatabaseName, Username: serveOpts.DatabaseUser, Password: serveOpts.DatabasePassword}
	globalPackagingNamespace := serveOpts.GlobalPackagingNamespace
	manager, err := newManager(dbConfig, globalPackagingNamespace)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	err = manager.Init()
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	defer manager.Close()

	netClient, err := httpclient.NewWithCertFile(additionalCAFile, serveOpts.TlsInsecureSkipVerify)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	authorizationHeader := serveOpts.AuthorizationHeader
	// The auth header may be a dockerconfig that we need to parse
	if serveOpts.DockerConfigJson != "" {
		dockerConfig := &kube.DockerConfigJSON{}
		err = json.Unmarshal([]byte(serveOpts.DockerConfigJson), dockerConfig)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		authorizationHeader, err = kube.GetAuthHeaderFromDockerConfig(dockerConfig)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
	}

	filters, err := parseFilters(serveOpts.FilterRules)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	var repoIface ChartCatalog
	if args[2] == "helm" {
		repoIface, err = getHelmRepo(serveOpts.Namespace, args[0], args[1], authorizationHeader, filters, netClient, serveOpts.UserAgent, manager)
	} else {
		var grpcClient ocicatalog.OCICatalogServiceClient
		if serveOpts.OCICatalogURL != "" {
			var closer func()
			grpcClient, closer, err = ocicatalog_client.NewClient(serveOpts.OCICatalogURL)
			if err != nil {
				return fmt.Errorf("unable to create oci catalog client: %w", err)
			}
			defer closer()
		}
		repoIface, err = getOCIRepo(serveOpts.Namespace, args[0], args[1], authorizationHeader, filters, serveOpts.OciRepositories, netClient, &grpcClient, manager)
	}
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}
	repo := repoIface.AppRepository()
	checksum, err := repoIface.Checksum(ctx)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	// Check if the repo has been already processed
	lastChecksum := manager.LastChecksum(models.AppRepository{Namespace: repo.Namespace, Name: repo.Name})
	log.V(3).Infof("Current checksum: %q. Previous checksum: %q", checksum, lastChecksum)
	if lastChecksum == checksum {
		log.V(3).Infof("Skipping repository since checksum has not changed. repo.URL= %q", repo.URL)
		return nil
	}

	fetchLatestOnlySlice := []bool{false}
	if lastChecksum == "" {
		// If the repo has never been processed, run first a shallow sync to give early feedback
		// then sync all the repositories
		fetchLatestOnlySlice = []bool{true, false}
	}

	for _, fetchLatestOnly := range fetchLatestOnlySlice {
		// Create the file importer to handle the icon and chart file imports.
		fImporter := fileImporter{manager, netClient}
		fileImporterJobs := make(chan models.Chart)
		fileImportsDone := make(chan bool)
		go fImporter.fetchFiles(fileImporterJobs, repoIface, serveOpts.UserAgent, serveOpts.PassCredentials, fileImportsDone)

		// We want to receive results per app so that we can sync that app
		// immediately and have a more responsive UX experience. A channel
		// gives us a way to pull results as they're generated (like an
		// iterator).
		chartResults := make(chan pullChartResult, 2)
		chartsToDelete, err := repoIface.Charts(ctx, fetchLatestOnly, chartResults)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}

		// Also need to collect the apps to be deleted, rather than simply
		// deleting everything that's not in the set of charts being synced
		// (as it does today).
		for chart := range chartResults {
			if len(chart.Errors) != 0 {
				chartName := chart.Chart.Name

				log.Infof("There were errors syncing chart %q: %v", chartName, chart.Errors)
				if len(chart.Chart.ChartVersions) == 0 {
					continue
				}
			}

			matches, err := filterMatches(chart.Chart, repoIface.Filters())
			if err != nil {
				return fmt.Errorf("error while applying filter: %w", err)
			}
			if !matches {
				continue
			}

			if err = manager.Sync(models.AppRepository{Name: repo.Name, Namespace: repo.Namespace}, chart.Chart); err != nil {
				return fmt.Errorf("can't add chart repository to database: %v", err)
			}

			// Fetch and store chart icons
			fileImporterJobs <- chart.Chart

			log.V(4).Infof("Chart %q synced, shallow=%v", chart.Chart.Name, fetchLatestOnly)
		}

		close(fileImporterJobs)

		err = manager.RemoveMissingCharts(models.AppRepository{
			Namespace: repo.Namespace,
			Name:      repo.Name,
		}, chartsToDelete)
		if err != nil {
			return fmt.Errorf("error while removing missing charts: %w", err)
		}

		// Wait for file imports to complete.
		log.V(4).Infof("Chart data syncing complete. Waiting for file imports to complete.")
		<-fileImportsDone

		log.V(4).Infof("Repository synced, shallow=%v", fetchLatestOnly)
	}

	// Update cache in the database
	if err = manager.UpdateLastCheck(repo.Namespace, repo.Name, checksum, time.Now()); err != nil {
		return fmt.Errorf("error: %v", err)
	}

	log.Infof("Stored repository update in cache, repo.URL= %v", repo.URL)
	log.Infof("Successfully added the package repository %s to database", args[0])
	return nil
}
