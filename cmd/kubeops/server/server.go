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
	"github.com/vmware-tanzu/kubeapps/cmd/kubeops/internal/httphandler"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"

	log "k8s.io/klog/v2"
)

type ServeOptions struct {
	ClustersConfigPath  string
	PinnipedProxyURL    string
	PinnipedProxyCACert string
	Burst               int
	Qps                 float32
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
		clustersConfig, cleanupCAFiles, err = kube.ParseClusterConfig(serveOpts.ClustersConfigPath, clustersCAFilesPrefix, serveOpts.PinnipedProxyURL, serveOpts.PinnipedProxyCACert)
		if err != nil {
			return fmt.Errorf("unable to parse additional clusters config: %+v", err)
		}
		defer cleanupCAFiles()
	}

	r := mux.NewRouter()

	// Healthcheck
	// TODO: add app specific health and readiness checks as per https://github.com/heptiolabs/healthcheck
	health := healthcheck.NewHandler()
	r.Handle("/live", health)
	r.Handle("/ready", health)

	err := httphandler.SetupDefaultRoutes(r.PathPrefix("/backend/v1").Subrouter(), serveOpts.Burst, serveOpts.Qps, clustersConfig)
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
		ReadHeaderTimeout: 60 * time.Second, // mitigate slowloris attacks, set to nginx's default
		Addr:              addr,
		Handler:           n,
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
