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

package kube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	apprepoclientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	v1alpha1typed "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/typed/apprepository/v1alpha1"
	log "github.com/sirupsen/logrus"
	authorizationapi "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ClusterConfig contains required info to talk to additional clusters.
type ClusterConfig struct {
	Name                     string `json:"name"`
	APIServiceURL            string `json:"apiServiceURL"`
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`
	// The genericclioptions.ConfigFlags struct includes only a CAFile field, not
	// a CAData field.
	// https://github.com/kubernetes/cli-runtime/issues/8
	// Embedding genericclioptions.ConfigFlags in a struct which includes the actual rest.Config
	// and returning that for ToRESTConfig() isn't enough, so we each configured cert out and
	// include a CAFile field in the config.
	CAFile string
	// ServiceToken can be configured so that the Kubeapps application itself
	// has access to get all namespaces on additional clusters, for example. It
	// should *not* be for reading secrets or similar, but limited to the
	// required functionality.
	ServiceToken string

	Insecure bool `json:"insecure"`
}

// ClustersConfig is an alias for a map of additional cluster configs.
type ClustersConfig struct {
	KubeappsClusterName string
	Clusters            map[string]ClusterConfig
}

// NewClusterConfig returns a copy of an in-cluster config with a custom token and/or custom cluster host
func NewClusterConfig(inClusterConfig *rest.Config, token string, cluster string, clustersConfig ClustersConfig) (*rest.Config, error) {
	config := rest.CopyConfig(inClusterConfig)
	config.BearerToken = token
	config.BearerTokenFile = ""

	if cluster == clustersConfig.KubeappsClusterName {
		return config, nil
	}

	additionalCluster, ok := clustersConfig.Clusters[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster %q has no configuration", cluster)
	}

	config.Host = additionalCluster.APIServiceURL
	config.TLSClientConfig = rest.TLSClientConfig{}
	config.TLSClientConfig.Insecure = additionalCluster.Insecure
	if additionalCluster.CertificateAuthorityData != "" {
		config.TLSClientConfig.CAData = []byte(additionalCluster.CertificateAuthorityData)
		config.CAFile = additionalCluster.CAFile
	}
	return config, nil
}

// combinedClientsetInterface provides both the app repository clientset and the corev1 clientset.
type combinedClientsetInterface interface {
	KubeappsV1alpha1() v1alpha1typed.KubeappsV1alpha1Interface
	CoreV1() corev1typed.CoreV1Interface
	AuthorizationV1() authorizationv1.AuthorizationV1Interface
	RestClient() rest.Interface
}

// Need to use a type alias to embed the two Clientset's without a name clash.
type kubeClientsetAlias = apprepoclientset.Clientset
type combinedClientset struct {
	*kubeClientsetAlias
	*kubernetes.Clientset
	restCli rest.Interface
}

func (c *combinedClientset) RestClient() rest.Interface {
	return c.restCli
}

// kubeHandler handles http requests for operating on app repositories and k8s resources
// in Kubeapps, without exposing implementation details to 3rd party integrations.
type kubeHandler struct {
	// The config set internally here cannot be used on its own as a valid
	// token is required. Call-sites use NewClusterConfig to obtain a valid
	// config with a specific token.
	config rest.Config

	// The namespace in which (currently) app repositories are created on the default cluster.
	kubeappsNamespace string

	// clientset using the pod serviceaccount on the default cluster.
	svcClientset combinedClientsetInterface

	// Configuration for additional clusters which may be requested.
	clustersConfig ClustersConfig

	// clientsetForConfig is a field on the struct only so it can be switched
	// for a fake version when testing. NewAppRepositoryhandler sets it to the
	// proper function below so that production code always has the real
	// version (and since this is a private struct, external code cannot change
	// the function).
	clientsetForConfig func(*rest.Config) (combinedClientsetInterface, error)
}

// userHandler is an extension of kubeHandler for a specific service account
type userHandler struct {
	// The namespace in which (currently) app repositories are created.
	kubeappsNamespace string

	// clientset using the pod serviceaccount for the default cluster
	svcClientset combinedClientsetInterface

	// clientset for a specific user token on a specific cluster.
	clientset combinedClientsetInterface
}

// ValidationResponse represents the response after validating a repo
type ValidationResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// This interface is explicitly private so that it cannot be used in function
// args, so that call-sites cannot accidentally pass a service handler in place
// of a user handler.
// TODO(mnelson): We could instead just create a UserHandler interface which embeds
// this one and adds one method, to force call-sites to explicitly use a UserHandler
// or ServiceHandler.
type handler interface {
	ListAppRepositories(requestNamespace string) (*v1alpha1.AppRepositoryList, error)
	CreateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*v1alpha1.AppRepository, error)
	UpdateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*v1alpha1.AppRepository, error)
	DeleteAppRepository(name, namespace string) error
	GetNamespaces() ([]corev1.Namespace, error)
	GetSecret(name, namespace string) (*corev1.Secret, error)
	GetAppRepository(repoName, repoNamespace string) (*v1alpha1.AppRepository, error)
	ValidateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*ValidationResponse, error)
	GetOperatorLogo(namespace, name string) ([]byte, error)
}

// AuthHandler exposes Handler functionality as a user or the current serviceaccount
type AuthHandler interface {
	AsUser(token, cluster string) (handler, error)
	AsSVC() handler
}

func (a *kubeHandler) AsUser(token, cluster string) (handler, error) {
	config, err := NewClusterConfig(&a.config, token, cluster, a.clustersConfig)
	if err != nil {
		log.Errorf("unable to create config: %v", err)
		return nil, err
	}
	clientset, err := a.clientsetForConfig(config)
	if err != nil {
		log.Errorf("unable to create clientset: %v", err)
		return nil, err
	}

	// Just use the service clientset if we're on the default cluster, but otherwise
	// create a new clientset using a configured service token for a specific cluster.
	// This is used when requesting the namespaces for a cluster (to populate the selector)
	// iff the users own credential does not suffice. If a service token is not configured
	// for the cluster, the namespace selector remains unpopulated.
	var svcClientset combinedClientsetInterface
	if cluster == a.clustersConfig.KubeappsClusterName {
		svcClientset = a.svcClientset
	} else {
		additionalCluster, ok := a.clustersConfig.Clusters[cluster]
		if !ok {
			return nil, fmt.Errorf("cluster %q has no configuration", cluster)
		}
		svcConfig := *config
		svcConfig.BearerToken = additionalCluster.ServiceToken

		svcClientset, err = a.clientsetForConfig(config)
		if err != nil {
			log.Errorf("unable to create clientset: %v", err)
			return nil, err
		}
	}

	return &userHandler{
		kubeappsNamespace: a.kubeappsNamespace,
		svcClientset:      svcClientset,
		clientset:         clientset,
	}, nil
}

func (a *kubeHandler) AsSVC() handler {
	return &userHandler{
		kubeappsNamespace: a.kubeappsNamespace,
		svcClientset:      a.svcClientset,
		clientset:         a.svcClientset,
	}
}

// appRepositoryRequest is used to parse the JSON request
type appRepositoryRequest struct {
	AppRepository appRepositoryRequestDetails `json:"appRepository"`
}

type appRepositoryRequestDetails struct {
	Name               string                 `json:"name"`
	RepoURL            string                 `json:"repoURL"`
	AuthHeader         string                 `json:"authHeader"`
	CustomCA           string                 `json:"customCA"`
	RegistrySecrets    []string               `json:"registrySecrets"`
	SyncJobPodTemplate corev1.PodTemplateSpec `json:"syncJobPodTemplate"`
	ResyncRequests     uint                   `json:"resyncRequests"`
}

// ErrGlobalRepositoryWithSecrets defines the error returned when an attempt is
// made to create registry secrets for a global repo.
var ErrGlobalRepositoryWithSecrets = fmt.Errorf("docker registry secrets cannot be set for app repositories available in all namespaces")

// NewHandler returns a handler configured with a service account client set and a config
// with a blank token to be copied when creating user client sets with specific tokens.
func NewHandler(kubeappsNamespace string, clustersConfig ClustersConfig) (AuthHandler, error) {
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
	svcClientset, err := clientsetForConfig(svcRestConfig)
	if err != nil {
		return nil, err
	}

	return &kubeHandler{
		config:            *config,
		kubeappsNamespace: kubeappsNamespace,
		// See comment in the struct defn above.
		clientsetForConfig: clientsetForConfig,
		svcClientset:       svcClientset,
		clustersConfig:     clustersConfig,
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
	return &combinedClientset{arclientset, coreclientset, coreclientset.RESTClient()}, nil
}

func parseRepoRequest(appRepoBody io.ReadCloser) (*appRepositoryRequest, error) {
	var appRepoRequest appRepositoryRequest
	err := json.NewDecoder(appRepoBody).Decode(&appRepoRequest)
	if err != nil {
		log.Infof("unable to decode: %v", err)
		return nil, err
	}
	return &appRepoRequest, nil
}

func (a *userHandler) applyAppRepositorySecret(repoSecret *corev1.Secret, requestNamespace string, appRepo *v1alpha1.AppRepository) error {
	// TODO: pass request context through from user request to clientset.
	_, err := a.clientset.CoreV1().Secrets(requestNamespace).Create(context.TODO(), repoSecret, metav1.CreateOptions{})
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = a.clientset.CoreV1().Secrets(requestNamespace).Update(context.TODO(), repoSecret, metav1.UpdateOptions{})
	}
	if err != nil {
		return err
	}

	// TODO(#1647): Move app repo sync to namespaces so secret copy not required.
	if requestNamespace != a.kubeappsNamespace {
		repoSecret.ObjectMeta.Name = KubeappsSecretNameForRepo(appRepo.ObjectMeta.Name, appRepo.ObjectMeta.Namespace)
		repoSecret.ObjectMeta.OwnerReferences = nil
		_, err = a.svcClientset.CoreV1().Secrets(a.kubeappsNamespace).Create(context.TODO(), repoSecret, metav1.CreateOptions{})
		if err != nil && k8sErrors.IsAlreadyExists(err) {
			_, err = a.clientset.CoreV1().Secrets(a.kubeappsNamespace).Update(context.TODO(), repoSecret, metav1.UpdateOptions{})
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// ListAppRepositories list AppRepositories in a namespace, bypass RBAC if the requeste namespace is the global one
func (a *userHandler) ListAppRepositories(requestNamespace string) (*v1alpha1.AppRepositoryList, error) {
	if a.kubeappsNamespace == requestNamespace {
		return a.svcClientset.KubeappsV1alpha1().AppRepositories(requestNamespace).List(context.TODO(), metav1.ListOptions{})
	}
	return a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).List(context.TODO(), metav1.ListOptions{})
}

// CreateAppRepository creates an AppRepository resource based on the request data
func (a *userHandler) CreateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*v1alpha1.AppRepository, error) {
	if a.kubeappsNamespace == "" {
		log.Errorf("attempt to use app repositories handler without kubeappsNamespace configured")
		return nil, fmt.Errorf("kubeappsNamespace must be configured to enable app repository handler")
	}

	appRepoRequest, err := parseRepoRequest(appRepoBody)
	if err != nil {
		return nil, err
	}

	appRepo := appRepositoryForRequest(appRepoRequest)
	if err != nil {
		return nil, err
	}

	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && requestNamespace == a.kubeappsNamespace {
		return nil, ErrGlobalRepositoryWithSecrets
	}

	appRepo, err = a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).Create(context.TODO(), appRepo, metav1.CreateOptions{})

	if err != nil {
		return nil, err
	}

	repoSecret := secretForRequest(appRepoRequest, appRepo)
	if repoSecret != nil {
		a.applyAppRepositorySecret(repoSecret, requestNamespace, appRepo)
		if err != nil {
			return nil, err
		}
	}
	return appRepo, nil
}

// UpdateAppRepository updates an AppRepository resource based on the request data
func (a *userHandler) UpdateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*v1alpha1.AppRepository, error) {
	if a.kubeappsNamespace == "" {
		log.Errorf("attempt to use app repositories handler without kubeappsNamespace configured")
		return nil, fmt.Errorf("kubeappsNamespace must be configured to enable app repository handler")
	}

	appRepoRequest, err := parseRepoRequest(appRepoBody)
	if err != nil {
		return nil, err
	}

	appRepo := appRepositoryForRequest(appRepoRequest)
	if err != nil {
		return nil, err
	}

	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && requestNamespace == a.kubeappsNamespace {
		return nil, ErrGlobalRepositoryWithSecrets
	}

	existingAppRepo, err := a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).Get(context.TODO(), appRepo.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Update existing repo with the new spec
	existingAppRepo.Spec = appRepo.Spec
	appRepo, err = a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).Update(context.TODO(), existingAppRepo, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	repoSecret := secretForRequest(appRepoRequest, appRepo)
	if repoSecret != nil {
		a.applyAppRepositorySecret(repoSecret, requestNamespace, appRepo)
		if err != nil {
			return nil, err
		}
	}
	return appRepo, nil
}

// DeleteAppRepository deletes an AppRepository resource from a namespace.
func (a *userHandler) DeleteAppRepository(repoName, repoNamespace string) error {
	appRepo, err := a.clientset.KubeappsV1alpha1().AppRepositories(repoNamespace).Get(context.TODO(), repoName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	hasCredentials := appRepo.Spec.Auth.Header != nil || appRepo.Spec.Auth.CustomCA != nil
	err = a.clientset.KubeappsV1alpha1().AppRepositories(repoNamespace).Delete(context.TODO(), repoName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	// If the app repo was in a namespace other than the kubeapps one, we also delete the copy of
	// the repository credentials kept in the kubeapps namespace (the repo credentials in the actual
	// namespace should be deleted when the owning app repo is deleted).
	if hasCredentials && repoNamespace != a.kubeappsNamespace {
		err = a.clientset.CoreV1().Secrets(a.kubeappsNamespace).Delete(context.TODO(), KubeappsSecretNameForRepo(repoName, repoNamespace), metav1.DeleteOptions{})
	}
	return err
}

func getValidationCliAndReq(appRepoBody io.ReadCloser, requestNamespace, kubeappsNamespace string) (HTTPClient, *http.Request, error) {
	appRepoRequest, err := parseRepoRequest(appRepoBody)
	if err != nil {
		return nil, nil, err
	}

	appRepo := appRepositoryForRequest(appRepoRequest)
	if err != nil {
		return nil, nil, err
	}
	if len(appRepo.Spec.DockerRegistrySecrets) > 0 && requestNamespace == kubeappsNamespace {
		// TODO(mnelson): we may also want to validate that any docker registry secrets listed
		// already exist in the namespace.
		return nil, nil, ErrGlobalRepositoryWithSecrets
	}

	repoSecret := secretForRequest(appRepoRequest, appRepo)
	cli, err := InitNetClient(appRepo, repoSecret, repoSecret, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create HTTP client: %w", err)
	}
	indexURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/") + "/index.yaml"
	req, err := http.NewRequest("GET", indexURL, nil)
	if err != nil {
		return nil, nil, err
	}
	return cli, req, nil
}

func doValidationRequest(cli HTTPClient, req *http.Request) (*ValidationResponse, error) {
	res, err := cli.Do(req)
	if err != nil {
		// If the request fail, it's not an internal error
		return &ValidationResponse{Code: 400, Message: err.Error()}, nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse validation response. Got: %v", err)
	}
	return &ValidationResponse{Code: res.StatusCode, Message: string(body)}, nil
}

func (a *userHandler) ValidateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*ValidationResponse, error) {
	// Split body parsing to a different function for ease testing
	cli, req, err := getValidationCliAndReq(appRepoBody, requestNamespace, a.kubeappsNamespace)
	if err != nil {
		return nil, err
	}
	return doValidationRequest(cli, req)
}

// GetAppRepository returns an AppRepository resource from a namespace.
// Optionally set a token to get the AppRepository using a custom serviceaccount
func (a *userHandler) GetAppRepository(repoName, repoNamespace string) (*v1alpha1.AppRepository, error) {
	return a.clientset.KubeappsV1alpha1().AppRepositories(repoNamespace).Get(context.TODO(), repoName, metav1.GetOptions{})
}

// appRepositoryForRequest takes care of parsing the request data into an AppRepository.
func appRepositoryForRequest(appRepoRequest *appRepositoryRequest) *v1alpha1.AppRepository {
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
			Name: appRepo.Name,
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:                   appRepo.RepoURL,
			Type:                  "helm",
			Auth:                  auth,
			DockerRegistrySecrets: appRepo.RegistrySecrets,
			SyncJobPodTemplate:    appRepo.SyncJobPodTemplate,
			ResyncRequests:        appRepo.ResyncRequests,
		},
	}
}

// secretForRequest takes care of parsing the request data into a secret for an AppRepository.
func secretForRequest(appRepoRequest *appRepositoryRequest, appRepo *v1alpha1.AppRepository) *corev1.Secret {
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
			Name: secretNameForRepo(appRepo.Name),
			OwnerReferences: []metav1.OwnerReference{
				{
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
	return fmt.Sprintf("apprepo-%s", repoName)
}

// KubeappsSecretNameForRepo returns a name suitable for recording a copy of
// a per-namespace repository secret in the kubeapps namespace.
func KubeappsSecretNameForRepo(repoName, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, secretNameForRepo(repoName))
}

func filterAllowedNamespaces(userClientset combinedClientsetInterface, namespaces *corev1.NamespaceList) (*corev1.NamespaceList, error) {
	allowedNamespaces := []corev1.Namespace{}
	for _, namespace := range namespaces.Items {
		res, err := userClientset.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &authorizationapi.SelfSubjectAccessReview{
			Spec: authorizationapi.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationapi.ResourceAttributes{
					Group:     "",
					Resource:  "secrets",
					Verb:      "get",
					Namespace: namespace.Name,
				},
			},
		}, metav1.CreateOptions{})
		if err != nil {
			return nil, err
		}
		if res.Status.Allowed {
			allowedNamespaces = append(allowedNamespaces, namespace)
		}
	}
	namespaces.Items = allowedNamespaces
	return namespaces, nil
}

// GetNamespaces return the list of namespaces that the user has permission to access
func (a *userHandler) GetNamespaces() ([]corev1.Namespace, error) {
	// Try to list namespaces with the user token, for backward compatibility
	namespaces, err := a.clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		// TODO: #1763 Check if we have a service token for getting namespaces on other clusters.
		if k8sErrors.IsForbidden(err) {
			// The user doesn't have permissions to list namespaces, use the current serviceaccount
			namespaces, err = a.svcClientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
			if err != nil && k8sErrors.IsForbidden(err) {
				// If the configured svcclient doesn't have permission, just return an empty list.
				return []corev1.Namespace{}, nil
			}

			// Only if we obtained the namespaces from the svc client do we filter it using
			// the user clientset.
			namespaces, err = filterAllowedNamespaces(a.clientset, namespaces)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return namespaces.Items, nil
}

// GetSecret return the a secret from a namespace using a token if given
func (a *userHandler) GetSecret(name, namespace string) (*corev1.Secret, error) {
	return a.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetNamespaces return the list of namespaces that the user has permission to access
func (a *userHandler) GetOperatorLogo(namespace, name string) ([]byte, error) {
	return a.clientset.RestClient().Get().AbsPath(fmt.Sprintf("/apis/packages.operators.coreos.com/v1/namespaces/%s/packagemanifests/%s/icon", namespace, name)).Do(context.TODO()).Raw()
}
