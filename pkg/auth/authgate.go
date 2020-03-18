package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	"github.com/urfave/negroni"
)

// Context key type for request contexts
type contextKey int

// UserKey is the context key for the User data in the request context
const UserKey contextKey = 0

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// AuthGate implements middleware to check if the user has access to the specific namespace
// before continuing. If the path being handled by the AuthGate middleware does not include
// the 'namespace' mux var, the kubeapps namespace will be assumed.
func AuthGate(kubeappsNamespace string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		token := ExtractToken(req.Header.Get("Authorization"))
		if token == "" {
			response.NewErrorResponse(http.StatusUnauthorized, "Unauthorized").Write(w)
			return
		}
		userAuth, err := NewAuth(token)
		if err != nil {
			response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
			return
		}
		namespace := mux.Vars(req)["namespace"]
		if namespace == "" {
			namespace = kubeappsNamespace
		}
		authz, err := userAuth.ValidateForNamespace(namespace)

		if err != nil || !authz {
			msg := fmt.Sprintf("Unable to validate user for namespace %q", namespace)
			if err != nil {
				msg = fmt.Sprintf("%s: %s", msg, err.Error())
			}
			response.NewErrorResponse(http.StatusUnauthorized, msg).Write(w)
			return
		}
		ctx := context.WithValue(req.Context(), UserKey, userAuth)
		next(w, req.WithContext(ctx))
	}
}

// ExtractToken extracts the token from a correctly formatted Authorization header.
func ExtractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[len(tokenPrefix):]
	} else {
		return ""
	}
}
