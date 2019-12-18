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

package handler

import (
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	"github.com/kubeapps/kubeapps/pkg/auth"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// appRepositories handles http requests for operating on app repositories
// in Kubeapps, without exposing implementation details to 3rd party integrations.
type appRepositoriesHandler struct {
	// The config set internally here cannot be used on its own as a valid
	// token is required. Call-sites use ConfigForToken to obtain a valid
	// config with a specific token.
	config rest.Config

	// The namespace in which (currently) app repositories are created.
	kubeappsNamespace string

	// clientsetForConfig is a field on the struct only so it can be switched
	// for a fake version when testing. NewAppRepositoryHandler sets it to the
	// proper function below so that production code always has the real
	// version (and since this is a private struct, external code cannot change
	// the function).
	clientsetForConfig func(*rest.Config) (clientset.Interface, error)
}

// appRepositoryRequest is used to parse the JSON request
type appRepositoryRequest struct {
	AppRepository appRepositoryRequestDetails `json:"appRepository"`
}

type appRepositoryRequestDetails struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	// TODO(mnelson): Add credential support for private repositories
	// so this can be used for the UI request also.
}

// NewAppRepositoriesHandler returns an AppRepositories handler configured with
// the in-cluster config but overriding the token with an empty string, so that
// ConfigForToken must be called to obtain a valid config.
func NewAppRepositoriesHandler(kubeappsNamespace string) (*appRepositoriesHandler, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			AuthInfo: clientcmdapi.AuthInfo{
				Token:     "",
				TokenFile: "",
			},
		},
	)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	return &appRepositoriesHandler{
		config:            *config,
		kubeappsNamespace: kubeappsNamespace,
		// See comment in the struct defn above.
		clientsetForConfig: clientsetForConfig,
	}, nil
}

// clientsetForConfig returns a clientset using the provided config.
func clientsetForConfig(config *rest.Config) (clientset.Interface, error) {
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

// ConfigForToken returns a new config for a given auth token.
func (a *appRepositoriesHandler) ConfigForToken(token string) *rest.Config {
	configCopy := a.config
	configCopy.BearerToken = token
	return &configCopy
}

// Create creates an AppRepository resource based on the request data
func (a *appRepositoriesHandler) Create(w http.ResponseWriter, req *http.Request) {
	if a.kubeappsNamespace == "" {
		log.Errorf("attempt to use app repositories handler without kubeappsNamespace configured")
		http.Error(w, "kubeappsNamespace must be configured to enable app repository handler", http.StatusUnauthorized)
	}

	token := auth.ExtractToken(req.Header.Get("Authorization"))
	clientset, err := a.clientsetForConfig(a.ConfigForToken(token))
	if err != nil {
		log.Errorf("unable to create clientset: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	appRepo, err := appRepositoryForRequestData(req.Body)
	if err != nil {
		log.Infof("unable to decode: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO(mnelson): validate both required data and request for index
	// https://github.com/kubeapps/kubeapps/issues/1330

	_, err = clientset.KubeappsV1alpha1().AppRepositories(a.kubeappsNamespace).Create(appRepo)

	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok {
			status := statusErr.ErrStatus
			log.Infof("unable to create app repo: %v", status.Reason)
			http.Error(w, status.Message, int(status.Code))
		} else {
			log.Errorf("unable to create app repo: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}

// appRepositoryForRequestData takes care of parsing the request data into an AppRepository.
func appRepositoryForRequestData(body io.ReadCloser) (*v1alpha1.AppRepository, error) {
	var appRepoRequest appRepositoryRequest
	err := json.NewDecoder(body).Decode(&appRepoRequest)
	if err != nil {
		return nil, err
	}
	return &v1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name: appRepoRequest.AppRepository.Name,
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL: appRepoRequest.AppRepository.URL,
		},
	}, nil
}
