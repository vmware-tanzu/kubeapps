package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"github.com/kubeapps/kubeapps/pkg/response"
	"github.com/urfave/negroni"
)

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// CheckerForRequest defines a function type so we can also inject a fake for tests
// rather than setting a context value.
type CheckerForRequest func(clustersConfig kube.ClustersConfig, req *http.Request) (Checker, error)

func AuthCheckerForRequest(clustersConfig kube.ClustersConfig, req *http.Request) (Checker, error) {
	token := ExtractToken(req.Header.Get("Authorization"))
	if token == "" {
		return nil, fmt.Errorf("Authorization token missing")
	}
	clusterName := mux.Vars(req)["cluster"]
	return NewAuth(token, clusterName, clustersConfig)
}

// AuthGate implements middleware to check if the user has access to charts from
// the specific namespace before continuing.
//   * If the path being handled by the
//     AuthGate middleware does not include the 'namespace' mux var, or the value
//     is _all, then the check is for cluster-wide access.
//   * If the namespace is the global chart namespace (ie. kubeappsNamespace) then
//     we allow read access regardless.
func AuthGate(clustersConfig kube.ClustersConfig, kubeappsNamespace string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		userAuth, err := AuthCheckerForRequest(clustersConfig, req)
		if err != nil {
			response.NewErrorResponse(http.StatusUnauthorized, err.Error()).Write(w)
			return
		}
		namespace := mux.Vars(req)["namespace"]
		if namespace == dbutils.AllNamespaces {
			namespace = ""
		}

		// The auth-gate is used only for access to the asset-svc and the functionality should be
		// moved to the assetsvc itself if and when the assetsvc is updated to be cluster aware.
		authz := false
		// TODO(absoludity): Update to allow access to assets from the global kubeapps namespace
		// on the kubeapps cluster only. See #2037.
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
