package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/kubeapps/cmd/kubeops/internal/handler"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	backendHandlers "github.com/kubeapps/kubeapps/pkg/http-handler"
	"github.com/kubeapps/kubeapps/pkg/kube"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/helm/pkg/helm/environment"
)

const clustersCAFilesPrefix = "/etc/additional-clusters-cafiles"

var (
	additionalClustersConfigPath string
	clustersConfigPath           string
	assetsvcURL                  string
	helmDriverArg                string
	listLimit                    int
	settings                     environment.EnvSettings
	timeout                      int64
	userAgentComment             string
)

func init() {
	settings.AddFlags(pflag.CommandLine)
	pflag.StringVar(&assetsvcURL, "assetsvc-url", "https://kubeapps-internal-assetsvc:8080", "URL to the internal assetsvc")
	pflag.StringVar(&helmDriverArg, "helm-driver", "", "which Helm driver type to use")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	pflag.Int64Var(&timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
	pflag.StringVar(&clustersConfigPath, "clusters-config-path", "", "Configuration for clusters")
	pflag.StringVar(&additionalClustersConfigPath, "additional-clusters-config-path", "", "Configuration for clusters")
}

func main() {
	pflag.Parse()
	settings.Init(pflag.CommandLine)

	kubeappsNamespace := os.Getenv("POD_NAMESPACE")
	if kubeappsNamespace == "" {
		log.Fatal("POD_NAMESPACE should be defined")
	}

	var clustersConfig kube.ClustersConfig
	// TODO(absoludity): remove support for --additional-clusters-config-path once we're +2 releases away.
	if clustersConfigPath == "" && additionalClustersConfigPath != "" {
		clustersConfigPath = additionalClustersConfigPath
	}
	if clustersConfigPath != "" {
		var err error
		var cleanupCAFiles func()
		clustersConfig, cleanupCAFiles, err = parseClusterConfig(clustersConfigPath, clustersCAFilesPrefix)
		if err != nil {
			log.Fatalf("unable to parse additional clusters config: %+v", err)
		}
		defer cleanupCAFiles()
	}

	options := handler.Options{
		ListLimit:         listLimit,
		Timeout:           timeout,
		KubeappsNamespace: kubeappsNamespace,
		ClustersConfig:    clustersConfig,
	}

	storageForDriver := agent.StorageForSecrets
	if helmDriverArg != "" {
		var err error
		storageForDriver, err = agent.ParseDriverType(helmDriverArg)
		if err != nil {
			panic(err)
		}
	}
	withHandlerConfig := handler.WithHandlerConfig(storageForDriver, options)
	r := mux.NewRouter()

	// Healthcheck
	// TODO: add app specific health and readiness checks as per https://github.com/heptiolabs/healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	// Routes
	// Auth not necessary here with Helm 3 because it's done by Kubernetes.
	addRoute := handler.AddRouteWith(r.PathPrefix("/v1").Subrouter(), withHandlerConfig)
	// Deprecate non-cluster-aware URIs.
	addRoute("GET", "/releases", handler.ListAllReleases)
	addRoute("GET", "/namespaces/{namespace}/releases", handler.ListReleases)
	addRoute("POST", "/namespaces/{namespace}/releases", handler.CreateRelease)
	addRoute("GET", "/namespaces/{namespace}/releases/{releaseName}", handler.GetRelease)
	addRoute("PUT", "/namespaces/{namespace}/releases/{releaseName}", handler.OperateRelease)
	addRoute("DELETE", "/namespaces/{namespace}/releases/{releaseName}", handler.DeleteRelease)
	addRoute("GET", "/clusters/{cluster}/releases", handler.ListAllReleases)
	addRoute("GET", "/clusters/{cluster}/namespaces/{namespace}/releases", handler.ListReleases)
	addRoute("POST", "/clusters/{cluster}/namespaces/{namespace}/releases", handler.CreateRelease)
	addRoute("GET", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.GetRelease)
	addRoute("PUT", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.OperateRelease)
	addRoute("DELETE", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.DeleteRelease)

	// Backend routes unrelated to kubeops functionality.
	err := backendHandlers.SetupDefaultRoutes(r.PathPrefix("/backend/v1").Subrouter(), clustersConfig)
	if err != nil {
		log.Fatalf("Unable to setup backend routes: %+v", err)
	}

	// assetsvc reverse proxy
	// TODO(mnelson) remove this reverse proxy once the haproxy frontend
	// proxies requests directly to the assetsvc. Move the authz to the
	// assetsvc itself.
	authGate := auth.AuthGate(kubeappsNamespace)
	parsedAssetsvcURL, err := url.Parse(assetsvcURL)
	if err != nil {
		log.Fatalf("Unable to parse the assetsvc URL: %v", err)
	}
	assetsvcProxy := httputil.NewSingleHostReverseProxy(parsedAssetsvcURL)
	assetsvcPrefix := "/assetsvc"
	assetsvcRouter := r.PathPrefix(assetsvcPrefix).Subrouter()
	// Logos don't require authentication so bypass that step
	assetsvcRouter.Methods("GET").Path("/v1/ns/{namespace}/assets/{repo}/{id}/logo").Handler(negroni.New(
		negroni.Wrap(http.StripPrefix(assetsvcPrefix, assetsvcProxy)),
	))
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
		log.WithFields(log.Fields{"addr": addr}).Info("Started Kubeops")
		err := srv.ListenAndServe()
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

func parseClusterConfig(configPath, caFilesPrefix string) (kube.ClustersConfig, func(), error) {
	caFilesDir, err := ioutil.TempDir(caFilesPrefix, "")
	if err != nil {
		return kube.ClustersConfig{}, func() {}, err
	}
	deferFn := func() { os.RemoveAll(caFilesDir) }
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return kube.ClustersConfig{}, deferFn, err
	}

	var clusterConfigs []kube.ClusterConfig
	if err = json.Unmarshal(content, &clusterConfigs); err != nil {
		return kube.ClustersConfig{}, deferFn, err
	}

	configs := kube.ClustersConfig{KubeappsClusterName: "default", Clusters: map[string]kube.ClusterConfig{}}
	for _, c := range clusterConfigs {
		// We need to decode the base64-encoded cadata from the input.
		if c.CertificateAuthorityData != "" {
			decodedCAData, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
			if err != nil {
				return kube.ClustersConfig{}, deferFn, err
			}
			c.CertificateAuthorityData = string(decodedCAData)

			// We also need a CAFile field because Helm uses the genericclioptions.ConfigFlags
			// struct which does not support CAData.
			// https://github.com/kubernetes/cli-runtime/issues/8
			c.CAFile = filepath.Join(caFilesDir, c.Name)
			err = ioutil.WriteFile(c.CAFile, decodedCAData, 0644)
			if err != nil {
				return kube.ClustersConfig{}, deferFn, err
			}
		}
		configs.Clusters[c.Name] = c
	}
	return configs, deferFn, nil
}
