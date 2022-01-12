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
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/urfave/negroni"
	log "k8s.io/klog/v2"
)

type ServeOptions struct {
	Manager              utils.AssetManager
	DbURL                string
	DbName               string
	DbUsername           string
	DbPassword           string
	KubeappsNamespace    string
	GlobalReposNamespace string
}

// TODO(absoludity): Let's not use globals for storing state like this.
var manager utils.AssetManager

const pathPrefix = "/v1"

func setupRoutes() http.Handler {
	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// Routes
	apiv1 := r.PathPrefix(pathPrefix).Subrouter()
	// TODO: mnelson: Seems we could use path per endpoint handling empty params? Check.
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts").Handler(WithParams(listChartsWithFilters)) // accepts: name, version, appversion, repos, categories, q, page, size
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/categories").Handler(WithParams(getChartCategories))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/{repo}").Handler(WithParams(listChartsWithFilters)) // accepts: name, version, appversion, repos, categories, q, page, size
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/{repo}/categories").Handler(WithParams(getChartCategories))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/{repo}/{chartName}").Handler(WithParams(getChart))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/{repo}/{chartName}/versions").Handler(WithParams(listChartVersions))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/charts/{repo}/{chartName}/versions/{version}").Handler(WithParams(getChartVersion))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/assets/{repo}/{chartName}/versions/{version}/README.md").Handler(WithParams(getChartVersionReadme))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/assets/{repo}/{chartName}/versions/{version}/values.yaml").Handler(WithParams(getChartVersionValues))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/assets/{repo}/{chartName}/versions/{version}/values.schema.json").Handler(WithParams(getChartVersionSchema))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/assets/{repo}/{chartName}/logo").Handler(WithParams(getChartIcon))

	// Leave icon on the non-cluster aware as it is used from a link in the db data :/
	apiv1.Methods("GET").Path("/ns/{namespace}/assets/{repo}/{chartName}/logo").Handler(WithParams(getChartIcon))

	n := negroni.Classic()
	n.UseHandler(r)
	return n
}

func Serve(serveOpts ServeOptions) error {
	dbConfig := dbutils.Config{URL: *&serveOpts.DbURL, Database: *&serveOpts.DbName, Username: *&serveOpts.DbUsername, Password: serveOpts.DbPassword}

	var err error
	manager, err = utils.NewManager("postgresql", dbConfig, serveOpts.GlobalReposNamespace)
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	err = manager.Init()
	if err != nil {
		return fmt.Errorf("Error: %v", err)
	}
	defer manager.Close()

	n := setupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Infof("Started assetsvc, addr=%s", addr)
	http.ListenAndServe(addr, n)
	return nil
}
