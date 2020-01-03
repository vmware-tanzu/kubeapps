package handler

import (
	"net/http"

	"github.com/kubeapps/common/response"
	appRepo "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	authHeader     = "Authorization"
	namespaceParam = "namespace"
	nameParam      = "releaseName"
)

// This type represents the fact that a regular handler cannot actually be created until we have access to the request,
// because a valid action config (and hence agent config) cannot be created until then.
// If the agent config were a "this" argument instead of an explicit argument, it would be easy to create a handler with a "zero" config.
// This approach practically eliminates that risk; it is much easier to use WithAgentConfig to create a handler guaranteed to use a valid agent config.
type dependentHandler func(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params)

func NewInClusterConfig(token string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	config.BearerToken = token
	config.BearerTokenFile = ""
	return config, nil
}

// WithAgentConfig takes a dependentHandler and creates a regular (WithParams) handler that,
// for every request, will create an agent config for itself.
// Written in a curried fashion for convenient usage; see cmd/kubeops/main.go.
func WithAgentConfig(storageForDriver agent.StorageForDriver, options agent.Options) func(f dependentHandler) handlerutil.WithParams {
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
			cfg := agent.Config{
				AgentOptions: options,
				ActionConfig: actionConfig,
				ChartClient:  chartUtils.NewChartClient(kubeClient, appRepoClient, options.UserAgent),
			}
			f(cfg, w, req, params)
		}
	}
}

func ListReleases(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	apps, err := agent.ListReleases(cfg.ActionConfig, params[namespaceParam], cfg.AgentOptions.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func ListAllReleases(cfg agent.Config, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(cfg, w, req, make(map[string]string))
}

func CreateRelease(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	requireV1Support := false
	chartDetails, chartMulti, err := handlerutil.ParseAndGetChart(req, cfg.ChartClient, requireV1Support)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	ch := chartMulti.Helm3Chart
	releaseName := chartDetails.ReleaseName
	namespace := params[namespaceParam]
	valuesString := chartDetails.Values
	release, err := agent.CreateRelease(cfg, releaseName, namespace, valuesString, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(release).Write(w)
}

func GetRelease(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	// Namespace is already known by the RESTClientGetter.
	releaseName := params[nameParam]
	release, err := agent.GetRelease(cfg.ActionConfig, releaseName)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(newDashboardCompatibleRelease(*release)).Write(w)
}

func DeleteRelease(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
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
