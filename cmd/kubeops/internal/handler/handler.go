package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	"github.com/kubeapps/kubeapps/pkg/chart"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"github.com/kubeapps/kubeapps/pkg/response"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	authHeader     = "Authorization"
	clusterParam   = "cluster"
	namespaceParam = "namespace"
	nameParam      = "releaseName"
	authUserError  = "Unexpected error while configuring authentication"
)

// This type represents the fact that a regular handler cannot actually be created until we have access to the request,
// because a valid action config (and hence handler config) cannot be created until then.
// If the handler config were a "this" argument instead of an explicit argument, it would be easy to create a handler with a "zero" config.
// This approach practically eliminates that risk; it is much easier to use WithHandlerConfig to create a handler guaranteed to use a valid handler config.
type dependentHandler func(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params)

// Options represents options that can be created without a bearer token, i.e. once at application startup.
type Options struct {
	ListLimit              int
	Timeout                int64
	UserAgent              string
	KubeappsNamespace      string
	ClustersConfig         kube.ClustersConfig
	Burst                  int
	QPS                    float32
	NamespaceHeaderName    string
	NamespaceHeaderPattern string
}

// Config represents data needed by each handler to be able to create Helm 3 actions.
// It cannot be created without a bearer token, so a new one must be created upon each HTTP request.
type Config struct {
	ActionConfig       *action.Configuration
	Options            Options
	KubeHandler        kube.AuthHandler
	ChartClientFactory chartUtils.ChartClientFactoryInterface
	Cluster            string
	Token              string
	userClientSet      kubernetes.Interface
}

// WithHandlerConfig takes a dependentHandler and creates a regular (WithParams) handler that,
// for every request, will create a handler config for itself.
// Written in a curried fashion for convenient usage; see cmd/kubeops/main.go.
func WithHandlerConfig(storageForDriver agent.StorageForDriver, options Options) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
			// Don't assume the cluster name was in the url for backwards compatibility
			// for now.
			cluster, ok := params[clusterParam]
			if !ok {
				cluster = options.ClustersConfig.KubeappsClusterName
			}
			namespace := params[namespaceParam]
			token := auth.ExtractToken(req.Header.Get(authHeader))

			inClusterConfig, err := rest.InClusterConfig()
			if err != nil {
				log.Errorf("Failed to create in-cluster config: %v", err)
				response.NewErrorResponse(http.StatusInternalServerError, authUserError).Write(w)
				return
			}

			restConfig, err := kube.NewClusterConfig(inClusterConfig, token, cluster, options.ClustersConfig)
			if err != nil {
				log.Errorf("Failed to create in-cluster config with user token: %v", err)
				response.NewErrorResponse(http.StatusInternalServerError, authUserError).Write(w)
				return
			}
			userKubeClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				log.Errorf("Failed to create kube client with user config: %v", err)
				response.NewErrorResponse(http.StatusInternalServerError, authUserError).Write(w)
				return
			}
			actionConfig, err := agent.NewActionConfig(storageForDriver, restConfig, userKubeClient, namespace)
			if err != nil {
				log.Errorf("Failed to create action config with user client: %v", err)
				response.NewErrorResponse(http.StatusInternalServerError, authUserError).Write(w)
				return
			}

			kubeHandler, err := kube.NewHandler(options.KubeappsNamespace, options.NamespaceHeaderName, options.NamespaceHeaderPattern, options.Burst, options.QPS, options.ClustersConfig)
			if err != nil {
				log.Errorf("Failed to create handler: %v", err)
				response.NewErrorResponse(http.StatusInternalServerError, authUserError).Write(w)
				return
			}

			cfg := Config{
				Options:            options,
				ActionConfig:       actionConfig,
				KubeHandler:        kubeHandler,
				Cluster:            cluster,
				Token:              token,
				ChartClientFactory: &chartUtils.ChartClientFactory{},
				userClientSet:      userKubeClient,
			}
			f(cfg, w, req, params)
		}
	}
}

// AddRouteWith makes it easier to define routes in main.go and avoids code repetition.
func AddRouteWith(
	r *mux.Router,
	withHandlerConfig func(dependentHandler) handlerutil.WithParams,
) func(verb, path string, handler dependentHandler) {
	return func(verb, path string, handler dependentHandler) {
		r.Methods(verb).Path(path).Handler(negroni.New(negroni.Wrap(withHandlerConfig(handler))))
	}
}

func returnForbiddenActions(forbiddenActions []auth.Action, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(forbiddenActions)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewErrorResponse(http.StatusForbidden, string(body)).Write(w)
}

func returnErrMessage(err error, w http.ResponseWriter) {
	code := handlerutil.ErrorCode(err)
	errMessage := err.Error()
	if code == http.StatusForbidden {
		forbiddenActions := auth.ParseForbiddenActions(errMessage)
		if len(forbiddenActions) > 0 {
			returnForbiddenActions(forbiddenActions, w)
		} else {
			// Unable to parse forbidden actions, return the raw message
			response.NewErrorResponse(code, errMessage).Write(w)
		}
	} else {
		response.NewErrorResponse(code, errMessage).Write(w)
	}
}

// ListReleases list existing releases.
func ListReleases(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	apps, err := agent.ListReleases(cfg.ActionConfig, params[namespaceParam], cfg.Options.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

// ListAllReleases list all the releases available.
func ListAllReleases(cfg Config, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(cfg, w, req, make(map[string]string))
}

// CreateRelease creates a release.
func CreateRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	chartDetails, err := handlerutil.ParseRequest(req)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	// TODO: currently app repositories are only supported on the cluster on which Kubeapps is installed. #1982
	appRepo, caCertSecret, authSecret, err := chart.GetAppRepoAndRelatedSecrets(chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace, cfg.KubeHandler, cfg.Token, cfg.Options.ClustersConfig.KubeappsClusterName, cfg.Options.ClustersConfig.GlobalReposNamespace, cfg.Options.ClustersConfig.KubeappsClusterName)
	if err != nil {
		returnErrMessage(fmt.Errorf("unable to get app repository %q: %v", chartDetails.AppRepositoryResourceName, err), w)
		return
	}
	ch, err := handlerutil.GetChart(
		chartDetails,
		appRepo,
		caCertSecret, authSecret,
		cfg.ChartClientFactory.New(appRepo.Spec.Type, cfg.Options.UserAgent),
	)
	if err != nil {
		returnErrMessage(err, w)
		return
	}

	releaseName := chartDetails.ReleaseName
	namespace := params[namespaceParam]
	valuesString := chartDetails.Values
	registrySecrets, err := chartUtils.RegistrySecretsPerDomain(req.Context(), appRepo.Spec.DockerRegistrySecrets, appRepo.Namespace, cfg.userClientSet)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	release, err := agent.CreateRelease(cfg.ActionConfig, releaseName, namespace, valuesString, ch, registrySecrets, 0)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewDataResponse(release).Write(w)
}

// OperateRelease decides which method to call depending on the "action" query param.
func OperateRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	switch req.FormValue("action") {
	case "upgrade":
		upgradeRelease(cfg, w, req, params)
	case "rollback":
		rollbackRelease(cfg, w, req, params)
	// TODO: Add "test" case here.
	default:
		// By default, for maintaining compatibility, we call upgrade.
		upgradeRelease(cfg, w, req, params)
	}
}

func upgradeRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	releaseName := params[nameParam]
	chartDetails, err := handlerutil.ParseRequest(req)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	// TODO: currently app repositories are only supported on the cluster on which Kubeapps is installed. #1982
	appRepo, caCertSecret, authSecret, err := chart.GetAppRepoAndRelatedSecrets(chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace, cfg.KubeHandler, cfg.Token, cfg.Options.ClustersConfig.KubeappsClusterName, cfg.Options.KubeappsNamespace, cfg.Options.ClustersConfig.KubeappsClusterName)
	if err != nil {
		returnErrMessage(fmt.Errorf("unable to get app repository %q: %v", chartDetails.AppRepositoryResourceName, err), w)
		return
	}
	ch, err := handlerutil.GetChart(
		chartDetails,
		appRepo,
		caCertSecret, authSecret,
		cfg.ChartClientFactory.New(appRepo.Spec.Type, cfg.Options.UserAgent),
	)
	registrySecrets, err := chartUtils.RegistrySecretsPerDomain(req.Context(), appRepo.Spec.DockerRegistrySecrets, appRepo.Namespace, cfg.userClientSet)
	if err != nil {
		returnErrMessage(err, w)
		return
	}

	rel, err := agent.UpgradeRelease(cfg.ActionConfig, releaseName, chartDetails.Values, ch, registrySecrets, 0)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewDataResponse(*rel).Write(w)
}

func rollbackRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	releaseName := params[nameParam]
	revision := req.FormValue("revision")
	if revision == "" {
		response.NewErrorResponse(http.StatusUnprocessableEntity, "Missing revision to rollback in request").Write(w)
		return
	}
	revisionInt, err := strconv.ParseInt(revision, 10, 32)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	rel, err := agent.RollbackRelease(cfg.ActionConfig, releaseName, int(revisionInt), 0)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewDataResponse(*rel).Write(w)
}

// GetRelease returns a release.
func GetRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	// Namespace is already known by the RESTClientGetter.
	releaseName := params[nameParam]
	release, err := agent.GetRelease(cfg.ActionConfig, releaseName)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	response.NewDataResponse(*release).Write(w)
}

// DeleteRelease deletes a release.
func DeleteRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	releaseName := params[nameParam]
	purge := handlerutil.QueryParamIsTruthy("purge", req)
	// Helm 3 has --purge by default; --keep-history in Helm 3 corresponds to omitting --purge in Helm 2.
	// https://stackoverflow.com/a/59210923/2135002
	keepHistory := !purge
	err := agent.DeleteRelease(cfg.ActionConfig, releaseName, keepHistory, 0)
	if err != nil {
		returnErrMessage(err, w)
		return
	}
	w.Header().Set("Status-Code", "200")
	w.Write([]byte("OK"))
}
