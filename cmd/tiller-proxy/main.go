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
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/kubeapps/cmd/tiller-proxy/internal/handler"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	backendHandlers "github.com/kubeapps/kubeapps/pkg/http-handler"
	"github.com/kubeapps/kubeapps/pkg/kube"
	tillerProxy "github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/tlsutil"
)

var (
	settings   environment.EnvSettings
	proxy      *tillerProxy.Proxy
	kubeClient kubernetes.Interface
	listLimit  int
	timeout    int64

	tlsCaCertFile string // path to TLS CA certificate file
	tlsCertFile   string // path to TLS certificate file
	tlsKeyFile    string // path to TLS key file
	tlsVerify     bool   // enable TLS and verify remote certificates
	tlsEnable     bool   // enable TLS

	tlsCaCertDefault = fmt.Sprintf("%s/ca.crt", os.Getenv("HELM_HOME"))
	tlsCertDefault   = fmt.Sprintf("%s/tls.crt", os.Getenv("HELM_HOME"))
	tlsKeyDefault    = fmt.Sprintf("%s/tls.key", os.Getenv("HELM_HOME"))

	assetsvcURL string
)

func init() {
	settings.AddFlags(pflag.CommandLine)
	// TLS Flags
	pflag.StringVar(&tlsCaCertFile, "tls-ca-cert", tlsCaCertDefault, "path to TLS CA certificate file")
	pflag.StringVar(&tlsCertFile, "tls-cert", tlsCertDefault, "path to TLS certificate file")
	pflag.StringVar(&tlsKeyFile, "tls-key", tlsKeyDefault, "path to TLS key file")
	pflag.BoolVar(&tlsVerify, "tls-verify", false, "enable TLS for request and verify remote")
	pflag.BoolVar(&tlsEnable, "tls", false, "enable TLS for request")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	pflag.Int64Var(&timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
	pflag.StringVar(&assetsvcURL, "assetsvc-url", "http://kubeapps-internal-assetsvc:8080", "URL to the internal assetsvc")
	// Probably need to tell tiller-proxy what the kubeappsCluster is named.
	// If so, set the same on kubeops
}

func main() {
	pflag.Parse()

	// set defaults from environment
	settings.Init(pflag.CommandLine)

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Unable to get cluster config: %v", err)
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

	proxy = tillerProxy.NewProxy(kubeClient, helmClient, timeout)
	kubeappsNamespace := os.Getenv("POD_NAMESPACE")
	if kubeappsNamespace == "" {
		log.Fatalf("POD_NAMESPACE should be defined")
	}

	kubeHandler, err := kube.NewHandler(kubeappsNamespace, kube.ClustersConfig{})
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	chartClient := chartUtils.NewChartClient(kubeHandler, "default", kubeappsNamespace, userAgent())

	r := mux.NewRouter()

	// Healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// HTTP Handler
	h := handler.TillerProxy{
		CheckerForRequest: auth.AuthCheckerForRequest,
		ListLimit:         listLimit,
		ChartClient:       chartClient,
		ProxyClient:       proxy,
	}

	// Routes
	// Deprecate non-cluster-aware URIs.
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").Handler(handlerutil.WithoutParams(h.ListAllReleases))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(handlerutil.WithParams(h.ListReleases))
	apiv1.Methods("POST").Path("/namespaces/{namespace}/releases").Handler(handlerutil.WithParams(h.CreateRelease))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.GetRelease))
	apiv1.Methods("PUT").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.OperateRelease))
	apiv1.Methods("DELETE").Path("/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.DeleteRelease))
	apiv1.Methods("GET").Path("/clusters/{cluster}/releases").Handler(handlerutil.WithoutParams(h.ListAllReleases))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/releases").Handler(handlerutil.WithParams(h.ListReleases))
	apiv1.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/releases").Handler(handlerutil.WithParams(h.CreateRelease))
	apiv1.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.GetRelease))
	apiv1.Methods("PUT").Path("/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.OperateRelease))
	apiv1.Methods("DELETE").Path("/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}").Handler(handlerutil.WithParams(h.DeleteRelease))

	// Backend routes unrelated to tiller-proxy functionality.
	err = backendHandlers.SetupDefaultRoutes(r.PathPrefix("/backend/v1").Subrouter(), kube.ClustersConfig{})
	if err != nil {
		log.Fatalf("Unable to setup backend routes: %+v", err)
	}

	// assetsvc reverse proxy
	// TODO(mnelson) remove this reverse proxy once the haproxy frontend
	// proxies requests directly to the assetsvc. Move the authz to the
	// assetsvc itself.
	parsedAssetsvcURL, err := url.Parse(assetsvcURL)
	if err != nil {
		log.Fatalf("Unable to parse the assetsvc URL: %v", err)
	}
	assetsvcProxy := httputil.NewSingleHostReverseProxy(parsedAssetsvcURL)
	assetsvcPrefix := "/assetsvc"
	assetsvcRouter := r.PathPrefix(assetsvcPrefix).Subrouter()
	// Logos don't require authentication so bypass that step
	assetsvcRouter.Methods("GET").Path("/v1/ns/{ns}/assets/{repo}/{id}/logo").Handler(http.StripPrefix(assetsvcPrefix, assetsvcProxy))
	authGate := auth.AuthGate(kubeappsNamespace)
	assetsvcRouter.PathPrefix("/v1/ns/{namespace}/").Handler(negroni.New(
		authGate,
		negroni.Wrap(http.StripPrefix(assetsvcPrefix, assetsvcProxy)),
	))

	n := negroni.Classic()
	n.UseHandler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	srv := &http.Server{
		Addr:    addr,
		Handler: n,
	}

	go func() {
		log.WithFields(log.Fields{"addr": addr}).Info("Started Tiller Proxy")
		err = srv.ListenAndServe()
		if err != nil {
			log.Info(err)
		}
	}()

	// Catch SIGINT and SIGTERM
	// Set up channel on which to send signal notifications.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	log.Debug("Set system to get notified on signals")
	s := <-c
	log.Infof("Received signal: %v. Waiting for existing requests to finish", s)
	// Set a timeout value high enough to let k8s terminationGracePeriodSeconds to act
	// accordingly and send a SIGKILL if needed
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	log.Info("All requests have been served. Exiting")
	os.Exit(0)
}
