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
	"fmt"
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
	helmChartUtil "k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/tlsutil"

	"github.com/kubeapps/kubeapps/cmd/tiller-proxy/internal/handler"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	tillerProxy "github.com/kubeapps/kubeapps/pkg/proxy"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	settings    environment.EnvSettings
	proxy       *tillerProxy.Proxy
	kubeClient  kubernetes.Interface
	netClient   *http.Client
	disableAuth bool
	listLimit   int

	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsVerify     bool   // enable TLS and verify remote certificates
	tlsEnable     bool   // enable TLS

	tlsCaCertDefault = fmt.Sprintf("%s/ca.crt", os.Getenv("HELM_HOME"))
	tlsCertDefault   = fmt.Sprintf("%s/tls.crt", os.Getenv("HELM_HOME"))
	tlsKeyDefault    = fmt.Sprintf("%s/tls.key", os.Getenv("HELM_HOME"))
)

func init() {
	settings.AddFlags(pflag.CommandLine)
	// TLS Flags
	pflag.StringVar(&tlsCaCertFile, "tls-ca-cert", tlsCaCertDefault, "path to TLS CA certificate file")
	pflag.StringVar(&tlsCertFile, "tls-cert", tlsCertDefault, "path to TLS certificate file")
	pflag.StringVar(&tlsKeyFile, "tls-key", tlsKeyDefault, "path to TLS key file")
	pflag.BoolVar(&tlsVerify, "tls-verify", false, "enable TLS for request and verify remote")
	pflag.BoolVar(&tlsEnable, "tls", false, "enable TLS for request")
	pflag.BoolVar(&disableAuth, "disable-auth", false, "Disable authorization check")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")

	netClient = &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
	}
}

func main() {
	pflag.Parse()

	// set defaults from environment
	settings.Init(pflag.CommandLine)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to get cluter config: %v", err)
	}

	kubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Unable to create a kubernetes client: %v", err)
	}

	log.Printf("Using tiller host: %s", settings.TillerHost)
	helmOptions := []helm.Option{helm.Host(settings.TillerHost)}
	if tlsVerify || tlsEnable {
		if tlsCaCertFile == "" {
			tlsCaCertFile = settings.Home.TLSCaCert()
		}
		if tlsCertFile == "" {
			tlsCertFile = settings.Home.TLSCert()
		}
		if tlsKeyFile == "" {
			tlsKeyFile = settings.Home.TLSKey()
		}
		log.Printf("Using Key=%q, Cert=%q, CA=%q", tlsKeyFile, tlsCertFile, tlsCaCertFile)
		tlsopts := tlsutil.Options{KeyFile: tlsKeyFile, CertFile: tlsCertFile, InsecureSkipVerify: true}
		if tlsVerify {
			tlsopts.CaCertFile = tlsCaCertFile
			tlsopts.InsecureSkipVerify = false
		}
		tlscfg, err := tlsutil.ClientConfig(tlsopts)
		if err != nil {
			log.Fatal(err)
		}
		helmOptions = append(helmOptions, helm.WithTLS(tlscfg))
	}
	helmClient := helm.NewClient(helmOptions...)
	err = helmClient.PingTiller()
	if err != nil {
		log.Fatalf("Unable to connect to Tiller: %v", err)
	}

	proxy = tillerProxy.NewProxy(kubeClient, helmClient)
	chartutils := chartUtils.NewChart(kubeClient, netClient, helmChartUtil.LoadArchive)

	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	authGate := handler.AuthGate()

	// HTTP Handler
	h := handler.TillerProxy{
		DisableAuth: disableAuth,
		ListLimit:   listLimit,
		ChartClient: chartutils,
		ProxyClient: proxy,
	}
	// Routes
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithoutParams(h.ListAllReleases)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.ListReleases)),
	))
	apiv1.Methods("POST").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.CreateRelease)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.GetRelease)),
	))
	apiv1.Methods("PUT").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.UpgradeRelease)),
	))
	apiv1.Methods("DELETE").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(negroni.New(
		authGate,
		negroni.Wrap(handler.WithParams(h.DeleteRelease)),
	))

	n := negroni.Classic()
	n.UseHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.WithFields(log.Fields{"addr": addr}).Info("Started Tiller Proxy")
	err = http.ListenAndServe(addr, n)
	if err != nil {
		log.Fatalf("Unable to start the server: %v", err)
	}
}
