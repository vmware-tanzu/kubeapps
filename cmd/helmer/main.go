package main

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/helmer/internal/handler"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/helm/pkg/helm/environment"
)

var (
	settings         environment.EnvSettings
	chartsvcURL      string
	userAgentComment string
	listLimit        int
	timeout          int64
)

func init() {
	settings.AddFlags(pflag.CommandLine) // necessary???
	pflag.StringVar(&chartsvcURL, "chartsvc-url", "https://kubeapps-internal-chartsvc:8080", "URL to the internal chartsvc")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests") // necessary???
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	pflag.Int64Var(&timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
}

func main() {
	pflag.Parse()
	settings.Init(pflag.CommandLine)

	options := agent.Options{
		ListLimit: listLimit,
		Timeout:   timeout,
	}

	r := mux.NewRouter()
	withAgentContext := handler.WithAgentContext(options)

	// Routes
	// Auth not necessary here with Helm 3 because it's done by Kubernetes.
	apiv1 := r.PathPrefix("/v1").Subrouter()
	apiv1.Methods("GET").Path("/releases").Handler(negroni.New(
		negroni.Wrap(withAgentContext(handler.ListAllReleases)),
	))
	apiv1.Methods("GET").Path("/namespaces/{namespace}/releases").Handler(negroni.New(
		negroni.Wrap(withAgentContext(handler.ListReleases)),
	))

	// Chartsvc reverse proxy
	authGate := auth.AuthGate()
	parsedChartsvcURL, err := url.Parse(chartsvcURL)
	if err != nil {
		log.Fatalf("Unable to parse the chartsvc URL: %v", err)
	}
	chartsvcProxy := httputil.NewSingleHostReverseProxy(parsedChartsvcURL)
	chartsvcPrefix := "/chartsvc"
	chartsvcRouter := r.PathPrefix(chartsvcPrefix).Subrouter()
	// Logos don't require authentication so bypass that step
	chartsvcRouter.Methods("GET").Path("/v1/assets/{repo}/{id}/logo").Handler(negroni.New(
		negroni.Wrap(http.StripPrefix(chartsvcPrefix, chartsvcProxy)),
	))
	chartsvcRouter.Methods("GET").Handler(negroni.New(
		authGate,
		negroni.Wrap(http.StripPrefix(chartsvcPrefix, chartsvcProxy)),
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
