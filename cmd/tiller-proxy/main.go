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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/kubeapps/cmd/tiller-proxy/internal/handler"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	tillerProxy "github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	helmChartUtil "k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/tlsutil"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	settings    environment.EnvSettings
	proxy       *tillerProxy.Proxy
	kubeClient  kubernetes.Interface
	netClient   *clientWithDefaultUserAgent
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

	registryCaCerts []string
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
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	pflag.StringSliceVar(&registryCaCerts, "helm-registry-ca-cert", []string{}, "Secret holding additional CA Certificates for chart registries")
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

	caCertPool := x509.NewCertPool()
	systemCACert, err := ioutil.ReadFile("/etc/ssl/certs/ca-certificates.crt")
	if err != nil {
		log.Fatalf("Unable to read system CA certs: %v", err)
	}
	caCertPool.AppendCertsFromPEM(systemCACert)

	if len(registryCaCerts) > 0 {
		namespace := os.Getenv("POD_NAMESPACE")
		for _, registryCaCert := range registryCaCerts {
			caCertSecret, err := kubeClient.CoreV1().Secrets(namespace).Get(registryCaCert, metav1.GetOptions{})
			if err != nil {
				log.Fatalf("Unable to read the given CA cert: %v", err)
			}
			for _, caCert := range caCertSecret.Data {
				ok := caCertPool.AppendCertsFromPEM(caCert)
				log.Printf("Adding CA cert from %s: %v", registryCaCert, ok)
			}
		}
	}
	netClient = &clientWithDefaultUserAgent{
		Client: http.Client{
			Timeout: time.Second * defaultTimeoutSeconds,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
		},
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
