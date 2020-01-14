package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	appRepo "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/chart/helm3to2"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	"github.com/urfave/negroni"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	authHeader     = "Authorization"
	namespaceParam = "namespace"
	nameParam      = "releaseName"
)

const isV1SupportRequired = false

// This type represents the fact that a regular handler cannot actually be created until we have access to the request,
// because a valid action config (and hence handler config) cannot be created until then.
// If the handler config were a "this" argument instead of an explicit argument, it would be easy to create a handler with a "zero" config.
// This approach practically eliminates that risk; it is much easier to use WithHandlerConfig to create a handler guaranteed to use a valid handler config.
type dependentHandler func(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params)

// Options represents options that can be created without a bearer token, i.e. once at application startup.
type Options struct {
	ListLimit int
	Timeout   int64
	UserAgent string
}

// Config represents data needed by each handler to be able to create Helm 3 actions.
// It cannot be created without a bearer token, so a new one must be created upon each HTTP request.
type Config struct {
	ActionConfig *action.Configuration
	Options      Options
	ChartClient  chartUtils.Resolver
}

// NewInClusterConfig returns an internal cluster config replacing the token.
func NewInClusterConfig(token string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	return config, nil
}

// WithHandlerConfig takes a dependentHandler and creates a regular (WithParams) handler that,
// for every request, will create a handler config for itself.
// Written in a curried fashion for convenient usage; see cmd/kubeops/main.go.
func WithHandlerConfig(storageForDriver agent.StorageForDriver, options Options) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
			namespace := params[namespaceParam]
			token := auth.ExtractToken(req.Header.Get(authHeader))
			restConfig, err := NewInClusterConfig(token)
			if err != nil {
				// TODO log details rather than return potentially sensitive details in error.
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			kubeClient, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				// TODO log details rather than return potentially sensitive details in error.
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			appRepoClient, err := appRepo.NewForConfig(restConfig)
			if err != nil {
				// TODO log details rather than return potentially sensitive details in error.
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			actionConfig, err := agent.NewActionConfig(storageForDriver, restConfig, kubeClient, namespace)
			if err != nil {
				// TODO log details rather than return potentially sensitive details in error.
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			cfg := Config{
				Options:      options,
				ActionConfig: actionConfig,
				ChartClient:  chartUtils.NewChartClient(kubeClient, appRepoClient, options.UserAgent),
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

// ListReleases list existing releases.
func ListReleases(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	apps, err := agent.ListReleases(cfg.ActionConfig, params[namespaceParam], cfg.Options.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
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
	chartDetails, chartMulti, err := handlerutil.ParseAndGetChart(req, cfg.ChartClient, isV1SupportRequired)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	ch := chartMulti.Helm3Chart
	releaseName := chartDetails.ReleaseName
	namespace := params[namespaceParam]
	valuesString := chartDetails.Values
	release, err := agent.CreateRelease(cfg.ActionConfig, releaseName, namespace, valuesString, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
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
	chartDetails, chartMulti, err := handlerutil.ParseAndGetChart(req, cfg.ChartClient, isV1SupportRequired)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	ch := chartMulti.Helm3Chart
	rel, err := agent.UpgradeRelease(cfg.ActionConfig, releaseName, chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	compatRelease, err := helm3to2.Convert(*rel)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(compatRelease).Write(w)
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
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	rel, err := agent.RollbackRelease(cfg.ActionConfig, releaseName, int(revisionInt))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	compatRelease, err := newDashboardCompatibleRelease(*rel)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(compatRelease).Write(w)
}

// GetRelease returns a release.
func GetRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	// Namespace is already known by the RESTClientGetter.
	releaseName := params[nameParam]
	release, err := agent.GetRelease(cfg.ActionConfig, releaseName)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	compatRelease, err := helm3to2.Convert(*release)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(compatRelease).Write(w)
}

// DeleteRelease deletes a release.
func DeleteRelease(cfg Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	releaseName := params[nameParam]
	purge := handlerutil.QueryParamIsTruthy("purge", req)
	// Helm 3 has --purge by default; --keep-history in Helm 3 corresponds to omitting --purge in Helm 2.
	// https://stackoverflow.com/a/59210923/2135002
	keepHistory := !purge
	err := agent.DeleteRelease(cfg.ActionConfig, releaseName, keepHistory)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	w.Header().Set("Status-Code", "200")
	w.Write([]byte("OK"))
}
