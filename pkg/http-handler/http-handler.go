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

	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/apprepo"
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

func returnK8sError(err error, w http.ResponseWriter) {
	if statusErr, ok := err.(*k8sErrors.StatusError); ok {
		status := statusErr.ErrStatus
		log.Infof("unable to create app repo: %v", status.Reason)
		http.Error(w, status.Message, int(status.Code))
	} else {
		log.Errorf("unable to create app repo: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CreateAppRepository creates App Repository
func CreateAppRepository(appRepo apprepo.Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace := mux.Vars(req)["namespace"]
		appRepo, err := appRepo.CreateAppRepository(req, requestNamespace)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// DeleteAppRepository deletes an App Repository
func DeleteAppRepository(appRepo apprepo.Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		repoNamespace := mux.Vars(req)["namespace"]
		repoName := mux.Vars(req)["name"]

		err := appRepo.DeleteAppRepository(req, repoName, repoNamespace)

		if err != nil {
			returnK8sError(err, w)
		}
	}
}

// GetNamespaces return the list of namespaces
func GetNamespaces(appRepo apprepo.Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		namespaces, err := appRepo.GetNamespaces(req)
		if err != nil {
			returnK8sError(err, w)
		}
		response := namespacesResponse{
			Namespaces: namespaces,
		}
		responseBody, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// SetupDefaultRoutes enables call-sites to use the backend api's default routes with minimal setup.
func SetupDefaultRoutes(r *mux.Router) error {
	backendHandler, err := apprepo.NewAppRepositoriesHandler(os.Getenv("POD_NAMESPACE"))
	if err != nil {
		return err
	}
	r.Methods("GET").Path("/namespaces").Handler(http.HandlerFunc(GetNamespaces(backendHandler)))
	r.Methods("POST").Path("/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(CreateAppRepository(backendHandler)))
	r.Methods("DELETE").Path("/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(DeleteAppRepository(backendHandler)))
	return nil
}
