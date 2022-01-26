// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httphandler

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"

	mux "github.com/gorilla/mux"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	authutils "github.com/kubeapps/kubeapps/pkg/auth"
	kubeutils "github.com/kubeapps/kubeapps/pkg/kube"
	log "github.com/sirupsen/logrus"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// namespacesResponse is used to marshal the JSON response
type namespacesResponse struct {
	Namespaces []k8scorev1.Namespace `json:"namespaces"`
}

// appRepositoryResponse is used to marshal the JSON response
type appRepositoryResponse struct {
	AppRepository apprepov1alpha1.AppRepository `json:"appRepository"`
	Secret        k8scorev1.Secret              `json:"secret"`
}

// appRepositoryListResponse is used to marshal the JSON response
type appRepositoryListResponse struct {
	AppRepositoryList apprepov1alpha1.AppRepositoryList `json:"appRepository"`
}

type allowedResponse struct {
	Allowed bool `json:"allowed"`
}

// JSONError returns an error code and a JSON response
func JSONError(w http.ResponseWriter, err interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(err)
}

func returnK8sError(err error, action string, resource string, w http.ResponseWriter) {
	if statusErr, ok := err.(*k8serrors.StatusError); ok {
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

// getHeaderNamespaces returns a list of namespaces from the header request
// The name and the value of the header field is specified by 2 variables:
// - headerName is a name of the expected header field, e.g. X-Consumer-Groups
// - headerPattern is a regular expression and it matches only single regex group, e.g. ^namespace:([\w-]+)$
func getHeaderNamespaces(req *http.Request, headerName, headerPattern string) ([]k8scorev1.Namespace, error) {
	var namespaces = []k8scorev1.Namespace{}
	if headerName == "" || headerPattern == "" {
		return []k8scorev1.Namespace{}, nil
	}
	r, err := regexp.Compile(headerPattern)
	if err != nil {
		log.Errorf("unable to compile regular expression: %v", err)
		return namespaces, err
	}
	headerNamespacesOrigin := strings.Split(req.Header.Get(headerName), ",")
	for _, n := range headerNamespacesOrigin {
		rns := r.FindStringSubmatch(strings.TrimSpace(n))
		if rns == nil || len(rns) < 2 {
			continue
		}
		ns := k8scorev1.Namespace{
			ObjectMeta: k8smetav1.ObjectMeta{Name: rns[1]},
			Status:     k8scorev1.NamespaceStatus{Phase: k8scorev1.NamespaceActive},
		}
		namespaces = append(namespaces, ns)
	}
	return namespaces, nil
}

// ListAppRepositories list app repositories
func ListAppRepositories(handler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "list", "AppRepositories", w)
			return
		}

		appRepos, err := clientset.ListAppRepositories(requestNamespace)
		if err != nil {
			returnK8sError(err, "list", "AppRepositories", w)
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
func CreateAppRepository(handler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "create", "AppRepository", w)
			return
		}

		appRepo, err := clientset.CreateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, "create", "AppRepository", w)
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
func UpdateAppRepository(handler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "update", "AppRepository", w)
			return
		}

		appRepo, err := clientset.UpdateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, "update", "AppRepository", w)
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

// RefreshAppRepository forces a refresh in a given apprepository (by updating resyncRequests property)
func RefreshAppRepository(handler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		repoName := mux.Vars(req)["name"]
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "refresh", "AppRepository", w)
			return
		}

		appRepo, err := clientset.RefreshAppRepository(repoName, requestNamespace)
		if err != nil {
			returnK8sError(err, "refresh", "AppRepository", w)
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
func ValidateAppRepository(handler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := handler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "validate", "AppRepository", w)
			return
		}

		res, err := clientset.ValidateAppRepository(req.Body, requestNamespace)
		if err != nil {
			returnK8sError(err, "validate", "AppRepository", w)
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
func DeleteAppRepository(kubeHandler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		repoName := mux.Vars(req)["name"]
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := kubeHandler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "delete", "AppRepository", w)
			return
		}

		err = clientset.DeleteAppRepository(repoName, requestNamespace)
		if err != nil {
			returnK8sError(err, "delete", "AppRepository", w)
		}
	}
}

// GetAppRepository gets an App Repository with a related secret if present.
func GetAppRepository(kubeHandler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		requestNamespace, requestCluster := getNamespaceAndCluster(req)
		repoName := mux.Vars(req)["name"]
		token := authutils.ExtractToken(req.Header.Get("Authorization"))

		clientset, err := kubeHandler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "get", "AppRepository", w)
			return
		}

		appRepo, err := clientset.GetAppRepository(repoName, requestNamespace)
		if err != nil {
			returnK8sError(err, "get", "AppRepository", w)
			return
		}

		response := appRepositoryResponse{
			AppRepository: *appRepo,
		}

		auth := &appRepo.Spec.Auth
		if auth != nil {
			var secretSelector *k8scorev1.SecretKeySelector
			if auth.CustomCA != nil {
				secretSelector = &auth.CustomCA.SecretKeyRef
			} else if auth.Header != nil {
				secretSelector = &auth.Header.SecretKeyRef
			}
			if secretSelector != nil {
				secret, err := clientset.GetSecret(secretSelector.Name, requestNamespace)
				if err != nil {
					returnK8sError(err, "get", "Secret", w)
					return
				}
				response.Secret = *secret
			}
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			JSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(responseBody)
	}
}

// GetNamespaces return the list of namespaces
func GetNamespaces(kubeHandler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token := authutils.ExtractToken(req.Header.Get("Authorization"))
		_, requestCluster := getNamespaceAndCluster(req)

		options := kubeHandler.GetOptions()

		clientset, err := kubeHandler.AsUser(token, requestCluster)
		if err != nil {
			returnK8sError(err, "get", "Namespaces", w)
			return
		}

		headerNamespaces, err := getHeaderNamespaces(req, options.NamespaceHeaderName, options.NamespaceHeaderPattern)
		if err != nil {
			returnK8sError(err, "get", "Namespaces", w)
		}

		namespaces, err := clientset.GetNamespaces(headerNamespaces)
		if err != nil {
			returnK8sError(err, "get", "Namespaces", w)
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
func GetOperatorLogo(kubeHandler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
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
		w.Write(logo)
	}
}

// CanI returns a boolean if the user can do the given action
func CanI(kubeHandler kubeutils.AuthHandler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		token := authutils.ExtractToken(req.Header.Get("Authorization"))
		_, requestCluster := getNamespaceAndCluster(req)

		clientset, err := kubeHandler.AsUser(token, requestCluster)

		if err != nil {
			returnK8sError(err, "get", "CanI", w)
			return
		}

		defer req.Body.Close()
		attributes, err := kubeutils.ParseSelfSubjectAccessRequest(req.Body)
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
		w.Write(responseBody)
	}
}

// SetupDefaultRoutes enables call-sites to use the backend api's default routes with minimal setup.
func SetupDefaultRoutes(r *mux.Router, namespaceHeaderName, namespaceHeaderPattern string, burst int, qps float32, clustersConfig kubeutils.ClustersConfig) error {
	backendHandler, err := kubeutils.NewHandler(os.Getenv("POD_NAMESPACE"), namespaceHeaderName, namespaceHeaderPattern, burst, qps, clustersConfig)
	if err != nil {
		return err
	}
	//TODO(agamez): move these endpoints to a separate plugin when possible
	r.Methods("POST").Path("/clusters/{cluster}/can-i").Handler(http.HandlerFunc(CanI(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces").Handler(http.HandlerFunc(GetNamespaces(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/apprepositories").Handler(http.HandlerFunc(ListAppRepositories(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(ListAppRepositories(backendHandler)))
	r.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories").Handler(http.HandlerFunc(CreateAppRepository(backendHandler)))
	r.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/validate").Handler(http.HandlerFunc(ValidateAppRepository(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(GetAppRepository(backendHandler)))
	r.Methods("PUT").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(UpdateAppRepository(backendHandler)))
	r.Methods("POST").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}/refresh").Handler(http.HandlerFunc(RefreshAppRepository(backendHandler)))
	r.Methods("DELETE").Path("/clusters/{cluster}/namespaces/{namespace}/apprepositories/{name}").Handler(http.HandlerFunc(DeleteAppRepository(backendHandler)))
	r.Methods("GET").Path("/clusters/{cluster}/namespaces/{namespace}/operator/{name}/logo").Handler(http.HandlerFunc(GetOperatorLogo(backendHandler)))
	return nil
}
