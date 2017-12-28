/*
Copyright (c) 2017 Bitnami

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

var dbSession datastore.Session

func main() {
	dbURL := flag.String("mongo-url", "localhost", "MongoDB URL (see https://godoc.org/labix.org/v2/mgo#Dial for format)")
	dbName := flag.String("mongo-database", "charts", "MongoDB database")
	dbUsername := flag.String("mongo-user", "", "MongoDB user")
	dbPassword := os.Getenv("MONGO_PASSWORD")
	flag.Parse()

	mongoConfig := datastore.Config{URL: *dbURL, Database: *dbName, Username: *dbUsername, Password: dbPassword}
	var err error
	dbSession, err = datastore.NewSession(mongoConfig)
	if err != nil {
		log.WithFields(log.Fields{"host": *dbURL}).Fatal(err)
	}

	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/charts").HandlerFunc(listCharts)
	apiv1.Methods("GET").Path("/charts/{repo}").Handler(WithParams(listRepoCharts))
	apiv1.Methods("GET").Path("/charts/{repo}/{chartName}").Handler(WithParams(getChart))
	apiv1.Methods("GET").Path("/charts/{repo}/{chartName}/versions").Handler(WithParams(listChartVersions))
	apiv1.Methods("GET").Path("/charts/{repo}/{chartName}/versions/{version}").Handler(WithParams(getChartVersion))
	apiv1.Methods("GET").Path("/assets/{repo}/{chartName}/logo-160x160-fit.png").Handler(WithParams(getChartIcon))
	apiv1.Methods("GET").Path("/assets/{repo}/{chartName}/versions/{version}/README.md").Handler(WithParams(getChartVersionReadme))

	n := negroni.Classic()
	n.UseHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.WithFields(log.Fields{"addr": addr}).Info("Started RateSvc")
	http.ListenAndServe(addr, n)
}
