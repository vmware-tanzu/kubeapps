// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"

	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	log "k8s.io/klog/v2"
)

func InvalidateCache(serveOpts Config, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("This command does not take any arguments (got %v)", len(args))
	}

	dbConfig := dbutils.Config{URL: serveOpts.DatabaseURL, Database: serveOpts.DatabaseName, Username: serveOpts.DatabaseUser, Password: serveOpts.DatabasePassword}
	manager, err := newManager(dbConfig, serveOpts.GlobalPackagingNamespace)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	err = manager.Init()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	defer manager.Close()

	err = manager.InvalidateCache()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	log.Infof("Successfully invalidated cache")
	return nil
}
