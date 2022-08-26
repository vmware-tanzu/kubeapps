// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httphandler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	log "k8s.io/klog/v2"
)

type allowedResponse struct {
	Allowed bool `json:"allowed"`
}

// tokenPrefix is the string preceding the token in the Authorization header.
const tokenPrefix = "Bearer "

// ExtractToken extracts the token from a correctly formatted Authorization header.
func extractToken(headerValue string) string {
	if strings.HasPrefix(headerValue, tokenPrefix) {
		return headerValue[len(tokenPrefix):]
	} else {
		return ""
	}
}

// JSONError returns an error code and a JSON response
func JSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	e := json.NewEncoder(w).Encode(err)
	if e != nil {
		return
	}
}

func returnK8sError(err error, action string, resource string, w http.ResponseWriter) {
	if statusErr, ok := err.(*k8sErrors.StatusError); ok {
		status := statusErr.ErrStatus
		log.Infof("unable to %s %s: %v", action, resource, status.Reason)
		JSONError(w, statusErr.ErrStatus, int(status.Code))
	} else {
		log.Errorf("unable to %s %s: %v", action, resource, err)
		JSONError(w, err.Error(), http.StatusInternalServerError)
	}
}

func getNamespaceAndCluster(req *http.Request) (string, string) {
	requestNamespace := mux.Vars(req)["namespace"]
	requestCluster := mux.Vars(req)["cluster"]
	return requestNamespace, requestCluster
}

func GetOperatorLogo(kubeHandler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		name := mux.Vars(req)["name"]
		ns, requestCluster := getNamespaceAndCluster(req)
		clientset, err := kubeHandler.AsSVC(requestCluster)
		if err != nil {
			returnK8sError(err, "get", "OperatorLogo", w)
			return
		}

		logo, err := clientset.GetOperatorLogo(ns, name)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctype := http.DetectContentType(logo)
		if strings.Contains(ctype, "text/") {
			// DetectContentType is unable to return svg icons since they are in fact text
			ctype = "image/svg+xml"
		}
		w.Header().Set("Content-Type", ctype)
		_, err = w.Write(logo)
		if err != nil {
			return
		}
	}
}

// CanI returns a boolean if the user can do the given action
func CanI(kubeHandler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token := extractToken(req.Header.Get("Authorization"))
		_, requestCluster := getNamespaceAndCluster(req)

		clientset, err := kubeHandler.AsUser(token, requestCluster)

		if err != nil {
			returnK8sError(err, "get", "CanI", w)
			return
		}

		defer req.Body.Close()
		attributes, err := kube.ParseSelfSubjectAccessRequest(req.Body)
		if err != nil {
			returnK8sError(err, "get", "CanI", w)
			return
		}
		allowed, err := clientset.CanI(attributes)
		if err != nil {
			returnK8sError(err, "get", "CanI", w)
			return
		}

		response := allowedResponse{
			Allowed: allowed,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(responseBody)
		if err != nil {
			return
		}
	}
}

// SetupDefaultRoutes enables call-sites to use the backend api's default routes with minimal setup.
func SetupDefaultRoutes(r *mux.Router, burst int, qps float32, clustersConfig kube.ClustersConfig) error {
	backendHandler, err := kube.NewHandler(os.Getenv("POD_NAMESPACE"), burst, qps, clustersConfig)
	if err != nil {
		return err
	}
	//TODO(agamez): move these endpoints to a separate plugin when possible
	r.Methods("POST").Path("/clusters/{cluster}/can-i").Handler(http.HandlerFunc(CanI(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/operator/{name}/logo").Handler(http.HandlerFunc(GetOperatorLogo(backendHandler)))
	return nil
}
