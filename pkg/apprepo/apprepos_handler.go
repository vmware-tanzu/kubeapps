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

package apprepo

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	apprepoclientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	v1alpha1typed "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/typed/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/auth"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// coreClientsetInterface provides just the core client
type coreClientsetInterface interface {
	CoreV1() corev1typed.CoreV1Interface
}

// combinedClientsetInterface provides both the app repository clientset and the corev1 clientset.
type combinedClientsetInterface interface {
	KubeappsV1alpha1() v1alpha1typed.KubeappsV1alpha1Interface
	CoreV1() corev1typed.CoreV1Interface
	AuthorizationV1() authorizationv1.AuthorizationV1Interface
}

// Need to use a type alias to embed the two Clientset's without a name clash.
type appRepoClientsetAlias = apprepoclientset.Clientset
type combinedClientset struct {
	*appRepoClientsetAlias
	*kubernetes.Clientset
}

// appRepositories handles http requests for operating on app repositories
// in Kubeapps, without exposing implementation details to 3rd party integrations.
type appRepositoriesHandler struct {
	// The config set internally here cannot be used on its own as a valid
	// token is required. Call-sites use ConfigForToken to obtain a valid
	// config with a specific token.
	config rest.Config

	// The namespace in which (currently) app repositories are created.
	kubeappsNamespace string

	// The Kubernetes client using the pod serviceaccount
	svcKubeClient coreClientsetInterface

	// clientsetForConfig is a field on the struct only so it can be switched
	// for a fake version when testing. NewAppRepositoryHandler sets it to the
	// proper function below so that production code always has the real
	// version (and since this is a private struct, external code cannot change
	// the function).
	clientsetForConfig func(*rest.Config) (combinedClientsetInterface, error)
}

// appRepositoryRequest is used to parse the JSON request
type appRepositoryRequest struct {
	AppRepository appRepositoryRequestDetails `json:"appRepository"`
}

type appRepositoryRequestDetails struct {
	Name               string                 `json:"name"`
	Namespace          string                 `json:"namespace"`
	RepoURL            string                 `json:"repoURL"`
	AuthHeader         string                 `json:"authHeader"`
	CustomCA           string                 `json:"customCA"`
	SyncJobPodTemplate corev1.PodTemplateSpec `json:"syncJobPodTemplate"`
	ResyncRequests     uint                   `json:"resyncRequests"`
}

// appRepositoryResponse is used to marshal the JSON response
type appRepositoryResponse struct {
	AppRepository v1alpha1.AppRepository `json:"appRepository"`
}

// NewAppRepositoriesHandler returns an AppRepositories and Kubernetes handler configured with
// the in-cluster config but overriding the token with an empty string, so that
// ConfigForToken must be called to obtain a valid config.
func NewAppRepositoriesHandler(kubeappsNamespace string) (*appRepositoriesHandler, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			AuthInfo: clientcmdapi.AuthInfo{
				// These three override their respective file or string
				// data.
				ClientCertificateData: []byte{},
				ClientKeyData:         []byte{},
				// A non empty value is required to override, it seems.
				TokenFile: " ",
			},
		},
	)
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	svcRestConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	svcKubeClient, err := kubernetes.NewForConfig(svcRestConfig)
	if err != nil {
		return nil, err
	}

	return &appRepositoriesHandler{
		config:            *config,
		kubeappsNamespace: kubeappsNamespace,
		// See comment in the struct defn above.
		clientsetForConfig: clientsetForConfig,
		svcKubeClient:      svcKubeClient,
	}, nil
}

// clientsetForConfig returns a clientset using the provided config.
func clientsetForConfig(config *rest.Config) (combinedClientsetInterface, error) {
	arclientset, err := apprepoclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	coreclientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &combinedClientset{arclientset, coreclientset}, nil
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
		return
	}

	token := auth.ExtractToken(req.Header.Get("Authorization"))
	clientset, err := a.clientsetForConfig(a.ConfigForToken(token))
	if err != nil {
		log.Errorf("unable to create clientset: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var appRepoRequest appRepositoryRequest
	err = json.NewDecoder(req.Body).Decode(&appRepoRequest)
	if err != nil {
		log.Infof("unable to decode: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Remove once dashboard always sends namespace.
	if appRepoRequest.AppRepository.Namespace == "" {
		appRepoRequest.AppRepository.Namespace = a.kubeappsNamespace
	}
	appRepo := appRepositoryForRequest(appRepoRequest)

	// TODO(mnelson): validate both required data and request for index
	// https://github.com/kubeapps/kubeapps/issues/1330

	appRepo, err = clientset.KubeappsV1alpha1().AppRepositories(appRepo.ObjectMeta.Namespace).Create(appRepo)

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

	repoSecret := secretForRequest(appRepoRequest, appRepo)
	if repoSecret != nil {
		_, err = clientset.CoreV1().Secrets(repoSecret.ObjectMeta.Namespace).Create(repoSecret)
		if err != nil {
			if statusErr, ok := err.(*errors.StatusError); ok {
				status := statusErr.ErrStatus
				log.Infof("unable to create app repo secret: %v", status.Reason)
				http.Error(w, status.Message, int(status.Code))
			} else {
				log.Errorf("unable to create app repo secret: %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// If the namespace isn't kubeapps (ie. this is a per-namespace
		// AppRepository), save a copy of the repository secret in kubeapps
		// namespace using the service account clientset. This enables the
		// existing assetsync service to be able to sync private
		// AppRepositories in other namespaces. It is not ideal and is a
		// temporary work-around until the asset-sync is updated to run
		// cronjobs in other namespaces with the assetsvc receiving the data.
		// See the relevant section of the design doc for details:
		// https://docs.google.com/document/d/1YEeKC6nPLoq4oaxs9v8_UsmxrRfWxB6KCyqrh2-Q8x0/edit?ts=5e2adf87#heading=h.kilvd2vii0w
		if repoSecret.ObjectMeta.Namespace != a.kubeappsNamespace {
			repoSecret.ObjectMeta.Namespace = a.kubeappsNamespace
			repoSecret.ObjectMeta.Name = kubeappsSecretNameForRepo(appRepo.ObjectMeta.Name, appRepo.ObjectMeta.Namespace)
			_, err = a.svcKubeClient.CoreV1().Secrets(repoSecret.ObjectMeta.Namespace).Create(repoSecret)
			if err != nil {
				if statusErr, ok := err.(*errors.StatusError); ok {
					status := statusErr.ErrStatus
					log.Infof("unable to create app repo secret in Kubeapps' namespace: %v", status.Reason)
					http.Error(w, status.Message, int(status.Code))
				} else {
					log.Errorf("unable to create app repo secret in Kubeapps' namespace: %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}
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

// appRepositoryForRequest takes care of parsing the request data into an AppRepository.
func appRepositoryForRequest(appRepoRequest appRepositoryRequest) *v1alpha1.AppRepository {
	appRepo := appRepoRequest.AppRepository

	var auth v1alpha1.AppRepositoryAuth
	if appRepo.AuthHeader != "" || appRepo.CustomCA != "" {
		secretName := secretNameForRepo(appRepo.Name)
		if appRepo.AuthHeader != "" {
			auth.Header = &v1alpha1.AppRepositoryAuthHeader{
				SecretKeyRef: corev1.SecretKeySelector{
					Key: "authorizationHeader",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
				},
			}
		}
		if appRepo.CustomCA != "" {
			auth.CustomCA = &v1alpha1.AppRepositoryCustomCA{
				SecretKeyRef: corev1.SecretKeySelector{
					Key: "ca.crt",
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
				},
			}
		}
	}

	return &v1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appRepo.Name,
			Namespace: appRepo.Namespace,
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:                appRepo.RepoURL,
			Type:               "helm",
			Auth:               auth,
			SyncJobPodTemplate: appRepo.SyncJobPodTemplate,
			ResyncRequests:     appRepo.ResyncRequests,
		},
	}
}

// secretForRequest takes care of parsing the request data into a secret for an AppRepository.
func secretForRequest(appRepoRequest appRepositoryRequest, appRepo *v1alpha1.AppRepository) *corev1.Secret {
	appRepoDetails := appRepoRequest.AppRepository
	secrets := map[string]string{}
	if appRepoDetails.AuthHeader != "" {
		secrets["authorizationHeader"] = appRepoDetails.AuthHeader
	}
	if appRepoDetails.CustomCA != "" {
		secrets["ca.crt"] = appRepoDetails.CustomCA
	}

	if len(secrets) == 0 {
		return nil
	}
	blockOwnerDeletion := true
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretNameForRepo(appRepo.Name),
			Namespace: appRepoDetails.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion:         "kubeapps.com/v1alpha1",
					Kind:               "AppRepository",
					Name:               appRepo.ObjectMeta.Name,
					UID:                appRepo.ObjectMeta.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				},
			},
		},
		StringData: secrets,
	}
}

func secretNameForRepo(repoName string) string {
	return fmt.Sprintf("apprepo-%s-secrets", repoName)
}

// kubeappsSecretNameForRepo returns a name suitable for recording a copy of
// a per-namespace repository secret in the kubeapps namespace.
func kubeappsSecretNameForRepo(repoName, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, secretNameForRepo(repoName))
}

func filterAllowedNamespaces(userClientset combinedClientsetInterface, namespaces *corev1.NamespaceList) ([]corev1.Namespace, error) {
	allowedNamespaces := []corev1.Namespace{}
	for _, namespace := range namespaces.Items {
		res, err := userClientset.AuthorizationV1().SelfSubjectAccessReviews().Create(&authorizationapi.SelfSubjectAccessReview{
			Spec: authorizationapi.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationapi.ResourceAttributes{
					Group:     "",
					Resource:  "secrets",
					Verb:      "get",
					Namespace: namespace.Name,
				},
			},
		})
		if err != nil {
			return nil, err
		}
		if res.Status.Allowed {
			allowedNamespaces = append(allowedNamespaces, namespace)
		}
	}
	return allowedNamespaces, nil
}

// GetNamespaces return the list of namespaces that the user has permission to access
// TODO(andresmgot): I am adding this method in this package for simplicity
// (since it already allows to impersonate the user)
// We should refactor this code to make it more generic (not apprepository-specific)
func (a *appRepositoriesHandler) GetNamespaces(w http.ResponseWriter, req *http.Request) {
	token := auth.ExtractToken(req.Header.Get("Authorization"))
	userClientset, err := a.clientsetForConfig(a.ConfigForToken(token))
	if err != nil {
		log.Errorf("unable to create clientset: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Try to list namespaces with the user token, for backward compatibility
	namespaces, err := userClientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		if errors.IsForbidden(err) {
			// The user doesn't have permissions to list namespaces, use the current serviceaccount
			namespaces, err = a.svcKubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
		}
		if err != nil && !errors.IsForbidden(err) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	allowedNamespaces, err := filterAllowedNamespaces(userClientset, namespaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(allowedNamespaces)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(responseBody)
}
