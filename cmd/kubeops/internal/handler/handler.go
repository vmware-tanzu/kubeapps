package handler

import (
	"net/http"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
)

const (
	authHeader     = "Authorization"
	namespaceParam = "namespace"
)

// This type represents the fact that a regular handler cannot actually be created until we have access to the request,
// because a valid action config (and hence agent config) cannot be created until then.
// If the agent config were a "this" argument instead of an explicit argument, it would be easy to create a handler with a "zero" config.
// This approach practically eliminates that risk; it is much easier to use WithAgentConfig to create a handler guaranteed to use a valid agent config.
type dependentHandler func(cfg agent.Config, w http.ResponseWriter, req *http.Request, params handlerutil.Params)

// WithAgentConfig takes a dependentHandler and creates a regular (WithParams) handler that,
// for every request, will create an agent config for itself.
// Written in a curried fashion for convenient usage; see cmd/kubeops/main.go.
func WithAgentConfig(driverType agent.DriverType, options agent.Options) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
			namespace := params[namespaceParam]
			token := auth.ExtractToken(req.Header.Get(authHeader))
			cfg := agent.Config{
				AgentOptions: options,
				ActionConfig: agent.NewActionConfig(driverType, token, namespace),
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
