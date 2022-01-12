/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"fmt"

	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
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

	repo := models.Repo{Name: args[0], Namespace: serveOpts.Namespace}
	if err = manager.Delete(repo); err != nil {
		return fmt.Errorf("Can't delete chart repository %s from database: %v", args[0], err)
	}

	log.Infof("Successfully deleted the chart repository %s from database", args[0])
	return nil
}
