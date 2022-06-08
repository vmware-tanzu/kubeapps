// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	negroni "github.com/urfave/negroni/v2"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeops/internal/handler"
	"github.com/vmware-tanzu/kubeapps/pkg/agent"
	backendHandlers "github.com/vmware-tanzu/kubeapps/pkg/http-handler"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"

	log "k8s.io/klog/v2"
)

type ServeOptions struct {
	HelmDriverArg          string
	ListLimit              int
	UserAgentComment       string
	Timeout                int64
	ClustersConfigPath     string
	PinnipedProxyURL       string
	Burst                  int
	Qps                    float32
	NamespaceHeaderName    string
	NamespaceHeaderPattern string
	UserAgent              string
	GlobalReposNamespace   string
}

const clustersCAFilesPrefix = "/etc/additional-clusters-cafiles"

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(serveOpts ServeOptions) error {

	kubeappsNamespace := os.Getenv("POD_NAMESPACE")
	if kubeappsNamespace == "" {
		return fmt.Errorf("POD_NAMESPACE should be defined")
	}

	// If there is no clusters config, we default to the previous behaviour of a "default" cluster.
	clustersConfig := kube.ClustersConfig{KubeappsClusterName: "default"}
	if serveOpts.ClustersConfigPath != "" {
		var err error
		var cleanupCAFiles func()
		clustersConfig, cleanupCAFiles, err = kube.ParseClusterConfig(serveOpts.ClustersConfigPath, clustersCAFilesPrefix, serveOpts.PinnipedProxyURL)
		if err != nil {
			return fmt.Errorf("unable to parse additional clusters config: %+v", err)
		}
		defer cleanupCAFiles()
	}

	options := handler.Options{
		ListLimit:              serveOpts.ListLimit,
		Timeout:                serveOpts.Timeout,
		KubeappsNamespace:      kubeappsNamespace,
		ClustersConfig:         clustersConfig,
		Burst:                  serveOpts.Burst,
		QPS:                    serveOpts.Qps,
		NamespaceHeaderName:    serveOpts.NamespaceHeaderName,
		NamespaceHeaderPattern: serveOpts.NamespaceHeaderPattern,
		UserAgent:              serveOpts.UserAgent,
	}

	storageForDriver := agent.StorageForSecrets
	if serveOpts.HelmDriverArg != "" {
		var err error
		storageForDriver, err = agent.ParseDriverType(serveOpts.HelmDriverArg)
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
	addRoute("GET", "/clusters/{cluster}/releases", handler.ListAllReleases)
	addRoute("GET", "/clusters/{cluster}/namespaces/{namespace}/releases", handler.ListReleases)
	addRoute("POST", "/clusters/{cluster}/namespaces/{namespace}/releases", handler.CreateRelease)
	addRoute("GET", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.GetRelease)
	addRoute("PUT", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.OperateRelease)
	addRoute("DELETE", "/clusters/{cluster}/namespaces/{namespace}/releases/{releaseName}", handler.DeleteRelease)

	// Backend routes unrelated to kubeops functionality.
	err := backendHandlers.SetupDefaultRoutes(r.PathPrefix("/backend/v1").Subrouter(), serveOpts.NamespaceHeaderName, serveOpts.NamespaceHeaderPattern, serveOpts.Burst, serveOpts.Qps, clustersConfig)
	if err != nil {
		return fmt.Errorf("Unable to setup backend routes: %+v", err)
	}

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
		log.Infof("Started Kubeops, addr=%v", addr)
		err := srv.ListenAndServe()
		if err != nil {
			log.Info(err)
		}
	}()

	// Catch SIGINT and SIGTERM
	// Set up channel on which to send signal notifications.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	log.Infof("Set system to get notified on signals")
	s := <-c
	log.Infof("Received signal: %v. Waiting for existing requests to finish", s)
	// Set a timeout value high enough to let k8s terminationGracePeriodSeconds to act
	// accordingly and send a SIGKILL if needed
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	err = srv.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("Error while shutting down: %v", err)
	}
	log.Info("All requests have been served. Exiting")
	os.Exit(0)

	return nil
}
