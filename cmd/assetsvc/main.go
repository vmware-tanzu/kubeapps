/*
Copyright (c) 2017 The Helm Authors

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
	"flag"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/common/datastore"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

const pathPrefix = "/v1"

var manager assetManager

func setupRoutes() http.Handler {
	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// Routes
	apiv1 := r.PathPrefix(pathPrefix).Subrouter()
	// TODO: mnelson: Seems we could use path per endpoint handling empty params? Check.
	apiv1.Methods("GET").Path("/ns/{namespace}/charts").Queries("name", "{chartName}", "version", "{version}", "appversion", "{appversion}").Handler(WithParams(listChartsWithFilters))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts").Queries("name", "{chartName}", "version", "{version}", "appversion", "{appversion}", "showDuplicates", "{showDuplicates}").Handler(WithParams(listChartsWithFilters))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts").Handler(WithParams(listCharts))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts").Queries("showDuplicates", "{showDuplicates}").Handler(WithParams(listCharts))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts/{repo}").Handler(WithParams(listCharts))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts/{repo}/{chartName}").Handler(WithParams(getChart))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts/{repo}/{chartName}/versions").Handler(WithParams(listChartVersions))
	apiv1.Methods("GET").Path("/ns/{namespace}/charts/{repo}/{chartName}/versions/{version}").Handler(WithParams(getChartVersion))
	apiv1.Methods("GET").Path("/ns/{namespace}/assets/{repo}/{chartName}/logo").Handler(WithParams(getChartIcon))
	apiv1.Methods("GET").Path("/ns/{namespace}/assets/{repo}/{chartName}/versions/{version}/README.md").Handler(WithParams(getChartVersionReadme))
	apiv1.Methods("GET").Path("/ns/{namespace}/assets/{repo}/{chartName}/versions/{version}/values.yaml").Handler(WithParams(getChartVersionValues))
	apiv1.Methods("GET").Path("/ns/{namespace}/assets/{repo}/{chartName}/versions/{version}/values.schema.json").Handler(WithParams(getChartVersionSchema))

	n := negroni.Classic()
	n.UseHandler(r)
	return n
}

func main() {
	dbURL := flag.String("database-url", "localhost", "Database URL")
	dbName := flag.String("database-name", "charts", "Database database")
	dbUsername := flag.String("database-user", "", "Database user")
	dbType := flag.String("database-type", "mongodb", "Database type")
	dbPassword := os.Getenv("DB_PASSWORD")
	flag.Parse()

	dbConfig := datastore.Config{URL: *dbURL, Database: *dbName, Username: *dbUsername, Password: dbPassword}

	kubeappsNamespace := os.Getenv("POD_NAMESPACE")

	var err error
	manager, err = newManager(*dbType, dbConfig, kubeappsNamespace)
	if err != nil {
		log.Fatal(err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	n := setupRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.WithFields(log.Fields{"addr": addr}).Info("Started assetsvc")
	http.ListenAndServe(addr, n)
}
