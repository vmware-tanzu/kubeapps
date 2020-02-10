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
	"github.com/heptiolabs/healthcheck"
	"github.com/kubeapps/kubeapps/cmd/kubeops/internal/handler"
	"github.com/kubeapps/kubeapps/pkg/agent"
	appRepoHandler "github.com/kubeapps/kubeapps/pkg/apprepo"
	"github.com/kubeapps/kubeapps/pkg/auth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/urfave/negroni"
	"k8s.io/helm/pkg/helm/environment"
)

var (
	settings         environment.EnvSettings
	assetsvcURL      string
	helmDriverArg    string
	userAgentComment string
	listLimit        int
	timeout          int64
)

func init() {
	settings.AddFlags(pflag.CommandLine)
	pflag.StringVar(&assetsvcURL, "assetsvc-url", "https://kubeapps-internal-assetsvc:8080", "URL to the internal assetsvc")
	pflag.StringVar(&helmDriverArg, "helm-driver", "", "which Helm driver type to use")
	pflag.IntVar(&listLimit, "list-max", 256, "maximum number of releases to fetch")
	pflag.StringVar(&userAgentComment, "user-agent-comment", "", "UserAgent comment used during outbound requests")
	// Default timeout from https://github.com/helm/helm/blob/b0b0accdfc84e154b3d48ec334cd5b4f9b345667/cmd/helm/install.go#L216
	pflag.Int64Var(&timeout, "timeout", 300, "Timeout to perform release operations (install, upgrade, rollback, delete)")
}

func main() {
	pflag.Parse()
	settings.Init(pflag.CommandLine)

	options := handler.Options{
		ListLimit: listLimit,
		Timeout:   timeout,
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
	addRoute("GET", "/releases", handler.ListAllReleases)
	addRoute("GET", "/namespaces/{namespace}/releases", handler.ListReleases)
	addRoute("POST", "/namespaces/{namespace}/releases", handler.CreateRelease)
	addRoute("GET", "/namespaces/{namespace}/releases/{releaseName}", handler.GetRelease)
	addRoute("PUT", "/namespaces/{namespace}/releases/{releaseName}", handler.OperateRelease)
	addRoute("DELETE", "/namespaces/{namespace}/releases/{releaseName}", handler.DeleteRelease)

	// Backend routes unrelated to kubeops functionality.
	appreposHandler, err := appRepoHandler.NewAppRepositoriesHandler(os.Getenv("POD_NAMESPACE"))
	if err != nil {
		log.Fatalf("Unable to create app repositories handler: %+v", err)
	}
	backendAPIv1 := r.PathPrefix("/backend/v1").Subrouter()
	backendAPIv1.Methods("POST").Path("/apprepositories").Handler(negroni.New(
		negroni.WrapFunc(appreposHandler.Create),
	))
	backendAPIv1.Methods("GET").Path("/namespaces").Handler(negroni.New(
		negroni.WrapFunc(appreposHandler.GetNamespaces),
	))

	// assetsvc reverse proxy
	authGate := auth.AuthGate()
	parsedAssetsvcURL, err := url.Parse(assetsvcURL)
	if err != nil {
		log.Fatalf("Unable to parse the assetsvc URL: %v", err)
	}
	assetsvcProxy := httputil.NewSingleHostReverseProxy(parsedAssetsvcURL)
	assetsvcPrefix := "/assetsvc"
	assetsvcRouter := r.PathPrefix(assetsvcPrefix).Subrouter()
	// Logos don't require authentication so bypass that step
	assetsvcRouter.Methods("GET").Path("/v1/assets/{repo}/{id}/logo").Handler(negroni.New(
		negroni.Wrap(http.StripPrefix(assetsvcPrefix, assetsvcProxy)),
	))
	assetsvcRouter.Methods("GET").Handler(negroni.New(
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
