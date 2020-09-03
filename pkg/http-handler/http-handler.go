/*
Copyright (c) 2019 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httphandler

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/auth"
	"github.com/kubeapps/kubeapps/pkg/kube"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// namespacesResponse is used to marshal the JSON response
type namespacesResponse struct {
	Namespaces []corev1.Namespace `json:"namespaces"`
}

// appRepositoryResponse is used to marshal the JSON response
type appRepositoryResponse struct {
	AppRepository v1alpha1.AppRepository `json:"appRepository"`
}

// appRepositoryListResponse is used to marshal the JSON response
type appRepositoryListResponse struct {
	AppRepositoryList v1alpha1.AppRepositoryList `json:"appRepository"`
}

// JSONError returns an error code and a JSON response
func JSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func returnK8sError(err error, w http.ResponseWriter) {
	if statusErr, ok := err.(*k8sErrors.StatusError); ok {
		status := statusErr.ErrStatus
		log.Infof("unable to create app repo: %v", status.Reason)
		JSONError(w, statusErr.ErrStatus, int(status.Code))
	} else {
		log.Errorf("unable to create app repo: %v", err)
		JSONError(w, err.Error(), http.StatusInternalServerError)
	}
}

func getNamespaceAndCluster(req *http.Request) (string, string) {
	requestNamespace := mux.Vars(req)["namespace"]
	requestCluster := mux.Vars(req)["cluster"]
	return requestNamespace, requestCluster
}

// ListAppRepositories list app repositories
func ListAppRepositories(handler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := auth.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		appRepos, err := clientset.ListAppRepositories(requestNamespace)
		if err != nil {
			returnK8sError(err, w)
			return
		}
		response := appRepositoryListResponse{
			AppRepositoryList: *appRepos,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// CreateAppRepository creates App Repository
func CreateAppRepository(handler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := auth.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		appRepo, err := clientset.CreateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, w)
			return
		}
		w.WriteHeader(http.StatusCreated)
		response := appRepositoryResponse{
			AppRepository: *appRepo,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// UpdateAppRepository updates an App Repository
func UpdateAppRepository(handler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := auth.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		appRepo, err := clientset.UpdateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
		response := appRepositoryResponse{
			AppRepository: *appRepo,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// ValidateAppRepository returns a 200 if the connection to the AppRepo can be established
func ValidateAppRepository(handler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := auth.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		res, err := clientset.ValidateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, w)
			return
		}
		responseBody, err := json.Marshal(res)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// DeleteAppRepository deletes an App Repository
func DeleteAppRepository(kubeHandler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		repoName := mux.Vars(req)["name"]
		token := auth.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := kubeHandler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		err = clientset.DeleteAppRepository(repoName, requestNamespace)
		if err != nil {
			returnK8sError(err, w)
		}
	}
}

// GetNamespaces return the list of namespaces
func GetNamespaces(kubeHandler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token := auth.ExtractToken(req.Header.Get("Authorization"))
		_, requestCluster := getNamespaceAndCluster(req)

		clientset, err := kubeHandler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, w)
			return
		}

		namespaces, err := clientset.GetNamespaces()
		if err != nil {
			returnK8sError(err, w)
		}
		response := namespacesResponse{
			Namespaces: namespaces,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// GetOperatorLogo return the list of namespaces
func GetOperatorLogo(kubeHandler kube.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		ns := mux.Vars(req)["namespace"]
		name := mux.Vars(req)["name"]
		logo, err := kubeHandler.AsSVC().GetOperatorLogo(ns, name)
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
		w.Write(logo)
	}
}

// SetupDefaultRoutes enables call-sites to use the backend api's default routes with minimal setup.
func SetupDefaultRoutes(r *mux.Router, clustersConfig kube.ClustersConfig) error {
	backendHandler, err := kube.NewHandler(os.Getenv("POD_NAMESPACE"), clustersConfig)
	if err != nil {
		return err
	}
	// Deprecate non-cluster-aware URIs.
	r.Methods("GET").Path("/namespaces").Handler(http.HandlerFunc(GetNamespaces(backendHandler)))
	r.Methods("GET").Path("/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(ListAppRepositories(backendHandler)))
	r.Methods("POST").Path("/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(CreateAppRepository(backendHandler)))
	r.Methods("POST").Path("/namespaces/{namespace}/apprepositories/validate").Handler(http.HandlerFunc(ValidateAppRepository(backendHandler)))
	r.Methods("PUT").Path("/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(UpdateAppRepository(backendHandler)))
	r.Methods("DELETE").Path("/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(DeleteAppRepository(backendHandler)))
	r.Methods("GET").Path("/namespaces/{namespace}/operator/{name}/logo").Handler(http.HandlerFunc(GetOperatorLogo(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces").Handler(http.HandlerFunc(GetNamespaces(backendHandler)))
	r.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(CreateAppRepository(backendHandler)))
	r.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/validate").Handler(http.HandlerFunc(ValidateAppRepository(backendHandler)))
	r.Methods("PUT").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(UpdateAppRepository(backendHandler)))
	r.Methods("DELETE").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(DeleteAppRepository(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/operator/{name}/logo").Handler(http.HandlerFunc(GetOperatorLogo(backendHandler)))
	return nil
}
