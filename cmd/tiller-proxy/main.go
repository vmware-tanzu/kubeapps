/*
Copyright (c) 2018 Bitnami

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
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"

	tillerProxy "github.com/kubeapps/kubeapps/pkg/utils/proxy"
)

var (
	settings environment.EnvSettings
	proxy    *tillerProxy.Proxy
)

const (
	defaultTimeoutSeconds = 180
)

func init() {
	settings.AddFlags(pflag.CommandLine)
}

func main() {
	pflag.Parse()

	// set defaults from environment
	settings.Init(pflag.CommandLine)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to get cluter config: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create a kubernetes client: %v", err)
	}

	log.Printf("Using tiller host: %s", settings.TillerHost)
	helmClient := helm.NewClient(helm.Host(settings.TillerHost))
	err = helmClient.PingTiller()
	if err != nil {
		log.Fatalf("Unable to connect to Tiller: %v", err)
	}

	netClient := &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
	}

	proxy = tillerProxy.NewProxy(kubeClient, helmClient, netClient, chartutil.LoadArchive)

	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").HandlerFunc(listAllReleases)
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(WithParams(listReleases))
	apiv1.Methods("POST").Path("/namespaces/{namespace}/releases").Handler(WithParams(deployRelease))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(WithParams(getRelease))
	apiv1.Methods("PUT").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(WithParams(deployRelease))
	apiv1.Methods("DELETE").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(WithParams(deleteRelease))

	n := negroni.Classic()
	n.UseHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.WithFields(log.Fields{"addr": addr}).Info("Started Tiller Proxy")
	http.ListenAndServe(addr, n)
}
