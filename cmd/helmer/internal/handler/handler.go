package handler

import (
	"net/http"
	"strings"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
)

const (
	authHeader     = "Authorization"
	namespaceParam = "namespace"
	tokenPrefix    = "Bearer "
)

// This type represents the fact that a regular handler cannot actually be created until we have access to the request,
// because a valid context cannot be created until then.
// If the context were a "this" argument instead of an explicit argument, it would be easy to create a handler with a "zero" context.
// This approach practically eliminates that risk; it is much easier to use WithAgentContext to create a handler guaranteed to use a valid context.
type dependentHandler func(ctx agent.Context, w http.ResponseWriter, req *http.Request, params handlerutil.Params)

// WithAgentContext takes a dependentHandler and creates a regular (WithParams) handler that,
// for every request, will create a context for itself.
// Written in a curried fashion for convenient usage; see cmd/helmer/main.go.
func WithAgentContext(options agent.Options) func(f dependentHandler) handlerutil.WithParams {
	return func(f dependentHandler) handlerutil.WithParams {
		return func(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
			namespace := params[namespaceParam]
			token := extractToken(req.Header.Get(authHeader))
			ctx := agent.Context{
				AgentOptions: options,
				ActionConfig: agent.NewConfig(token, namespace),
			}
			f(ctx, w, req, params)
		}
	}
}

func ListReleases(ctx agent.Context, w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	apps, err := agent.ListReleases(ctx, params[namespaceParam], req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func ListAllReleases(ctx agent.Context, w http.ResponseWriter, req *http.Request, _ handlerutil.Params) {
	ListReleases(ctx, w, req, map[string]string{namespaceParam: ""})
}

// A best effort at extracting the actual token from the Authorization header.
// We assume that the token is either preceded by tokenPrefix or not preceded by anything at all.
func extractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[len(tokenPrefix):]
	} else {
		return headerValue
	}
}
