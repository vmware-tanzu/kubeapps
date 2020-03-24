package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/urfave/negroni"
)

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// CheckerForRequest defines a function type so we can also inject a fake for tests
// rather than setting a context value.
type CheckerForRequest func(req *http.Request) (Checker, error)

func AuthCheckerForRequest(req *http.Request) (Checker, error) {
	token := ExtractToken(req.Header.Get("Authorization"))
	if token == "" {
		return nil, fmt.Errorf("Authorization token missing")
	}
	return NewAuth(token)
}

// AuthGate implements middleware to check if the user has access to read from
// the specific namespace before continuing.
//   * If the path being handled by the
//     AuthGate middleware does not include the 'namespace' mux var, or the value
//     is _all, then the check is for cluster-wide access.
//   * If the namespace is the global chart namespace (ie. kubeappsNamespace) then
//     we allow read access regardless.
func AuthGate(kubeappsNamespace string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		userAuth, err := AuthCheckerForRequest(req)
		if err != nil {
			response.NewErrorResponse(http.StatusUnauthorized, err.Error()).Write(w)
			return
		}
		namespace := mux.Vars(req)["namespace"]
		if namespace == dbutils.AllNamespaces {
			namespace = ""
		}

		// If the request is for the global public charts (ie. kubeappsNamespace)
		// we do not check authz.
		authz := false
		if namespace == kubeappsNamespace {
			authz = true
		} else {
			authz, err = userAuth.ValidateForNamespace(namespace)
		}

		if err != nil || !authz {
			msg := fmt.Sprintf("Unable to validate user for namespace %q", namespace)
			if err != nil {
				msg = fmt.Sprintf("%s: %s", msg, err.Error())
			}
			response.NewErrorResponse(http.StatusForbidden, msg).Write(w)
			return
		}
		next(w, req)
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
