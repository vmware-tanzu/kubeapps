// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	dbutils "github.com/kubeapps/kubeapps/pkg/dbutils"
	log "k8s.io/klog/v2"
)

func Delete(serveOpts Config, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("exactly one argument: [REPO NAME] (got %v)", len(args))
	}

	dbConfig := dbutils.Config{URL: serveOpts.DatabaseURL, Database: serveOpts.DatabaseName, Username: serveOpts.DatabaseUser, Password: serveOpts.DatabasePassword}
	manager, err := newManager(dbConfig, serveOpts.GlobalReposNamespace)
	if err != nil {
		return fmt.Errorf("Error file creating a mananger: %v", err)
	}
	err = manager.Init()
	if err != nil {
		return fmt.Errorf("Error file initializing the mananger: %v", err)
	}
	defer manager.Close()

	repo := chartmodels.Repo{Name: args[0], Namespace: serveOpts.Namespace}
	if err = manager.Delete(repo); err != nil {
		return fmt.Errorf("Can't delete chart repository %s from database: %v", args[0], err)
	}

	log.Infof("Successfully deleted the chart repository %s from database", args[0])
	return nil
}
