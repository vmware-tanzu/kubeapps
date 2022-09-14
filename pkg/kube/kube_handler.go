// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	apprepoclientset "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	v1alpha1typed "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/typed/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	authorizationv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	corev1typed "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	log "k8s.io/klog/v2"
)

const OCIImageManifestMediaType = "application/vnd.oci.image.manifest.v1+json"

// ClusterConfig contains required info to talk to additional clusters.
type ClusterConfig struct {
	Name                     string `json:"name"`
	APIServiceURL            string `json:"apiServiceURL"`
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`
	// When parsing config we decode the cert auth data to ensure it is valid
	// and store it since it's required when using the data.
	CertificateAuthorityDataDecoded string
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

	// Insecure should only be used in test or development environments and enables
	// TLS requests without requiring the cert authority validation.
	Insecure bool `json:"insecure"`

	// PinnipedConfig is an optional per-cluster configuration specifying
	// the pinniped namespace, authenticator type and authenticator name
	// that should be used for any credential exchange.
	PinnipedConfig PinnipedConciergeConfig `json:"pinnipedConfig,omitempty"`

	// IsKubeappsCluster is an optional per-cluster configuration specifying
	// that this cluster is the one in which Kubeapps is being installed.
	// Often this is inferred as the cluster without an explicit APIServiceURL, but
	// if every cluster defines an APIServiceURL, we can no longer infer the cluster
	// on which Kubeapps is installed.
	IsKubeappsCluster bool `json:"isKubeappsCluster,omitempty"`
}

// PinnipedConciergeConfig enables each cluster configuration to specify the
// pinniped-concierge installation to use for any credential exchange.
type PinnipedConciergeConfig struct {
	// Enabled flags whether this cluster should use
	// pinniped to exchange credentials.
	Enabled bool `json:"enabled"`
	// Enable is deprecated and will be removed in a future release.
	Enable bool `json:"enable"`
	// The Namespace, AuthenticatorType and Authenticator name to use
	// when exchanging credentials.
	Namespace         string `json:"namespace,omitempty"`
	AuthenticatorType string `json:"authenticatorType,omitempty"`
	AuthenticatorName string `json:"authenticatorName,omitempty"`
}

// ClustersConfig is an alias for a map of additional cluster configs.
type ClustersConfig struct {
	KubeappsClusterName      string
	GlobalPackagingNamespace string
	PinnipedProxyURL         string
	PinnipedProxyCACert      string
	Clusters                 map[string]ClusterConfig
}

// NewClusterConfig returns a copy of an in-cluster config with a user token (leave blank for
// when configuring a service account). and/or custom cluster host
func NewClusterConfig(inClusterConfig *rest.Config, userToken string, cluster string, clustersConfig ClustersConfig) (*rest.Config, error) {
	config := rest.CopyConfig(inClusterConfig)
	config.BearerToken = userToken
	config.BearerTokenFile = ""

	// If the cluster is empty, we assume the rest of the inClusterConfig is correct. This can be the case when
	// the cluster on which Kubeapps is installed is not one presented in the UI as a target (hence not in the
	// `clusters` configuration).
	if cluster == "" {
		return config, nil
	}

	clusterConfig, ok := clustersConfig.Clusters[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster %q has no configuration", cluster)
	}

	if userToken != "" && (clusterConfig.PinnipedConfig.Enabled || clusterConfig.PinnipedConfig.Enable) {
		// Create a config for routing requests via the pinniped-proxy for credential
		// exchange.
		config.Host = clustersConfig.PinnipedProxyURL
		// set roundtripper.
		// https://github.com/kubernetes/client-go/issues/407
		existingWrapTransport := config.WrapTransport
		config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			if existingWrapTransport != nil {
				rt = existingWrapTransport(rt)
			}
			headers := map[string][]string{}
			if clusterConfig.APIServiceURL != "" {
				headers["PINNIPED_PROXY_API_SERVER_URL"] = []string{clusterConfig.APIServiceURL}
			}
			if clusterConfig.CertificateAuthorityData != "" {
				headers["PINNIPED_PROXY_API_SERVER_CERT"] = []string{clusterConfig.CertificateAuthorityData}
			}
			return &pinnipedProxyRoundTripper{
				headers: headers,
				rt:      rt,
			}
		}

		// If pinniped-proxy is configured with TLS, we need to set the
		// CACert.
		config.CAFile = clustersConfig.PinnipedProxyCACert

		return config, nil
	}

	if cluster == clustersConfig.KubeappsClusterName {
		return config, nil
	}

	config.Host = clusterConfig.APIServiceURL
	config.TLSClientConfig = rest.TLSClientConfig{}
	config.TLSClientConfig.Insecure = clusterConfig.Insecure
	if clusterConfig.CertificateAuthorityDataDecoded != "" {
		config.TLSClientConfig.CAData = []byte(clusterConfig.CertificateAuthorityDataDecoded)
		config.CAFile = clusterConfig.CAFile
	}
	return config, nil
}

func ParseClusterConfig(configPath, caFilesPrefix string, pinnipedProxyURL, PinnipedProxyCACert string) (ClustersConfig, func(), error) {
	caFilesDir, err := os.MkdirTemp(caFilesPrefix, "")
	if err != nil {
		return ClustersConfig{}, func() {}, err
	}
	deferFn := func() {
		err = os.RemoveAll(caFilesDir)
	}

	// #nosec G304
	content, err := os.ReadFile(configPath)
	if err != nil {
		return ClustersConfig{}, deferFn, err
	}

	var clusterConfigs []ClusterConfig
	if err = json.Unmarshal(content, &clusterConfigs); err != nil {
		return ClustersConfig{}, deferFn, err
	}

	configs := ClustersConfig{Clusters: map[string]ClusterConfig{}}
	configs.PinnipedProxyURL = pinnipedProxyURL
	configs.PinnipedProxyCACert = PinnipedProxyCACert
	for _, c := range clusterConfigs {
		// Select the cluster in which Kubeapps in installed. We look for either
		// `isKubeappsCluster: true` or an empty `APIServiceURL`.
		isKubeappsClusterCandidate := c.IsKubeappsCluster || c.APIServiceURL == ""
		if isKubeappsClusterCandidate {
			if configs.KubeappsClusterName == "" {
				configs.KubeappsClusterName = c.Name
			} else {
				return ClustersConfig{}, nil, fmt.Errorf("only one cluster can be configured using either 'isKubeappsCluster: true' or without an apiServiceURL to refer to the cluster on which Kubeapps is installed, two defined: %q, %q", configs.KubeappsClusterName, c.Name)
			}
		}

		// We need to decode the base64-encoded cadata from the input.
		if c.CertificateAuthorityData != "" {
			decodedCAData, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
			if err != nil {
				return ClustersConfig{}, deferFn, err
			}
			c.CertificateAuthorityDataDecoded = string(decodedCAData)

			// We also need a CAFile field because Helm uses the genericclioptions.ConfigFlags
			// struct which does not support CAData.
			// https://github.com/kubernetes/cli-runtime/issues/8
			c.CAFile = filepath.Join(caFilesDir, c.Name)
			// #nosec G306
			// TODO(agamez): check if we can set perms to 0600 instead of 0644.
			err = os.WriteFile(c.CAFile, decodedCAData, 0644)
			if err != nil {
				return ClustersConfig{}, deferFn, err
			}
		}
		configs.Clusters[c.Name] = c
	}
	return configs, deferFn, nil
}

// combinedClientsetInterface provides both the app repository clientset and the corev1 clientset.
type combinedClientsetInterface interface {
	KubeappsV1alpha1() v1alpha1typed.KubeappsV1alpha1Interface
	CoreV1() corev1typed.CoreV1Interface
	AuthorizationV1() authorizationv1.AuthorizationV1Interface
	RestClient() rest.Interface
	MaxWorkers() int
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

func (c *combinedClientset) MaxWorkers() int {
	return int(c.restCli.GetRateLimiter().QPS())
}

type KubeOptions struct {
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

	// kubeappsSvcClientset is the clientset using the pod serviceaccount on the
	// cluster on which kubeapps is installed.
	kubeappsSvcClientset combinedClientsetInterface

	// Configuration for additional clusters which may be requested.
	clustersConfig ClustersConfig

	// clientsetForConfig is a field on the struct only so it can be switched
	// for a fake version when testing. NewAppRepositoryhandler sets it to the
	// proper function below so that production code always has the real
	// version (and since this is a private struct, external code cannot change
	// the function).
	clientsetForConfig func(*rest.Config) (combinedClientsetInterface, error)

	// Additional options from Kubeops arguments
	options KubeOptions
}

// userHandler is an extension of kubeHandler for a specific service account
type userHandler struct {
	// The namespace in which (currently) app repositories are created.
	kubeappsNamespace string

	// clientset using the pod serviceaccount for the specific cluster
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
	RefreshAppRepository(repoName string, requestNamespace string) (*v1alpha1.AppRepository, error)
	DeleteAppRepository(name, namespace string) error
	GetSecret(name, namespace string) (*corev1.Secret, error)
	GetAppRepository(repoName, repoNamespace string) (*v1alpha1.AppRepository, error)
	ValidateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*ValidationResponse, error)
}

// AuthHandler exposes Handler functionality as a user or the current serviceaccount
type AuthHandler interface {
	AsUser(token, cluster string) (handler, error)
	AsSVC(cluster string) (handler, error)
	GetOptions() KubeOptions
}

func (a *kubeHandler) getSvcClientsetForCluster(cluster string, config *rest.Config) (combinedClientsetInterface, error) {
	// Just use the service clientset if we're on the cluster on which Kubeapps
	// is installed, but otherwise create a new clientset using a configured
	// service token for a specific cluster. This is used when requesting the
	// namespaces for a cluster (to populate the selector) iff the users own
	// credential does not suffice. If a service token is not configured for the
	// cluster, the namespace selector remains unpopulated.
	var svcClientset combinedClientsetInterface
	var err error
	if cluster == "" || cluster == a.clustersConfig.KubeappsClusterName {
		svcClientset = a.kubeappsSvcClientset
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
	return svcClientset, nil
}

func (a *kubeHandler) GetOptions() KubeOptions {
	return a.options
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

	svcClientset, err := a.getSvcClientsetForCluster(cluster, config)
	if err != nil {
		log.Errorf("unable to create svc clientset: %v", err)
		return nil, err
	}

	return &userHandler{
		kubeappsNamespace: a.kubeappsNamespace,
		svcClientset:      svcClientset,
		clientset:         clientset,
	}, nil
}

func (a *kubeHandler) AsSVC(cluster string) (handler, error) {
	config, err := NewClusterConfig(&a.config, "", cluster, a.clustersConfig)
	if err != nil {
		log.Errorf("unable to create svc clientset: %v", err)
		return nil, err
	}

	svcClientset, err := a.getSvcClientsetForCluster(cluster, config)
	if err != nil {
		log.Errorf("unable to create svc clientset: %v", err)
		return nil, err
	}

	return &userHandler{
		kubeappsNamespace: a.kubeappsNamespace,
		svcClientset:      svcClientset,
		clientset:         svcClientset,
	}, nil
}

// appRepositoryRequest is used to parse the JSON request
type appRepositoryRequest struct {
	AppRepository appRepositoryRequestDetails `json:"appRepository"`
}

type appRepositoryRequestDetails struct {
	Name                  string                  `json:"name"`
	Type                  string                  `json:"type"`
	RepoURL               string                  `json:"repoURL"`
	AuthHeader            string                  `json:"authHeader"`
	CustomCA              string                  `json:"customCA"`
	AuthRegCreds          string                  `json:"authRegCreds"`
	RegistrySecrets       []string                `json:"registrySecrets"`
	SyncJobPodTemplate    corev1.PodTemplateSpec  `json:"syncJobPodTemplate"`
	ResyncRequests        int                     `json:"resyncRequests"`
	OCIRepositories       []string                `json:"ociRepositories"`
	TLSInsecureSkipVerify bool                    `json:"tlsInsecureSkipVerify"`
	FilterRule            v1alpha1.FilterRuleSpec `json:"filterRule"`
	Description           string                  `json:"description"`
	PassCredentials       bool                    `json:"passCredentials"`
}

// ErrGlobalRepositoryWithSecrets defines the error returned when an attempt is
// made to create registry secrets for a global repo.
var ErrGlobalRepositoryWithSecrets = fmt.Errorf("docker registry secrets cannot be set for app repositories available in all namespaces")

// ErrEmptyOCIRegistry defines the error returned when an attempt is
// made to create an OCI registry with no repositories
var ErrEmptyOCIRegistry = fmt.Errorf("You need to specify at least one repository for an OCI registry")

// NewHandler returns a handler configured with a service account client set and a config
// with a blank token to be copied when creating user client sets with specific tokens.
func NewHandler(kubeappsNamespace string, burst int, qps float32, clustersConfig ClustersConfig) (AuthHandler, error) {
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
	// Modify the default number of requests that the given client can do
	// This is useful to handle a large number of namespaces which are check in parallel
	// Burst is the initial number of request made in parallel
	// Then further requests are performed following the QPS rate
	config.Burst = burst
	config.QPS = qps

	options := KubeOptions{}

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
		clientsetForConfig:   clientsetForConfig,
		kubeappsSvcClientset: svcClientset,
		clustersConfig:       clustersConfig,
		options:              options,
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
		log.InfoS("unable to decode", "err", err)
		return nil, err
	}
	return &appRepoRequest, nil
}

func (a *userHandler) applyAppRepositorySecret(repoSecret *corev1.Secret, requestNamespace string) error {
	// TODO: pass request context through from user request to clientset.
	// Create the secret in the requested namespace if it's not an existing docker config secret
	_, err := a.clientset.CoreV1().Secrets(requestNamespace).Create(context.TODO(), repoSecret, metav1.CreateOptions{})
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = a.clientset.CoreV1().Secrets(requestNamespace).Update(context.TODO(), repoSecret, metav1.UpdateOptions{})
	}
	return err
}

// Deprecated: Remove when the new Package Repository API implementation is completed
// TODO(#1647): Move app repo sync to namespaces so secret copy not required.
func (a *userHandler) copyAppRepositorySecret(repoSecret *corev1.Secret, appRepo *v1alpha1.AppRepository) error {
	repoSecret.ObjectMeta.Name = KubeappsSecretNameForRepo(appRepo.ObjectMeta.Name, appRepo.ObjectMeta.Namespace)
	repoSecret.ObjectMeta.Namespace = a.kubeappsNamespace
	repoSecret.ObjectMeta.OwnerReferences = nil
	repoSecret.ObjectMeta.ResourceVersion = ""
	_, err := a.svcClientset.CoreV1().Secrets(a.kubeappsNamespace).Create(context.TODO(), repoSecret, metav1.CreateOptions{})
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		_, err = a.clientset.CoreV1().Secrets(a.kubeappsNamespace).Update(context.TODO(), repoSecret, metav1.UpdateOptions{})
	}
	return err
}

// ListAppRepositories list AppRepositories in a namespace, bypass RBAC if the requeste namespace is the global one
func (a *userHandler) ListAppRepositories(requestNamespace string) (*v1alpha1.AppRepositoryList, error) {
	if a.kubeappsNamespace == requestNamespace {
		return a.svcClientset.KubeappsV1alpha1().AppRepositories(requestNamespace).List(context.TODO(), metav1.ListOptions{})
	}
	return a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).List(context.TODO(), metav1.ListOptions{})
}

// Deprecated: Remove when the new Package Repository API implementation is completed
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

	repoSecret, err := a.secretForRequest(appRepoRequest, appRepo, requestNamespace)
	if err != nil {
		return nil, err
	}

	if repoSecret != nil {
		// If the secret is a docker config, the secret already exists so we don't need to create it
		if _, ok := repoSecret.Data[".dockerconfigjson"]; !ok {
			err = a.applyAppRepositorySecret(repoSecret, requestNamespace)
		}
		if err != nil {
			return nil, err
		}
		// If the namespace is different than the Kubeapps one, the secret needs to be copied
		if requestNamespace != a.kubeappsNamespace {
			err = a.copyAppRepositorySecret(repoSecret, appRepo)
		}
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

	repoSecret, err := a.secretForRequest(appRepoRequest, appRepo, requestNamespace)
	if err != nil {
		return nil, err
	}

	if repoSecret != nil {
		// If the secret is a docker config, the secret already exists so we don't need to create it
		if _, ok := repoSecret.Data[".dockerconfigjson"]; !ok {
			err = a.applyAppRepositorySecret(repoSecret, requestNamespace)
		}
		if err != nil {
			return nil, err
		}
		// If the namespace is different than the Kubeapps one, the secret needs to be copied
		if requestNamespace != a.kubeappsNamespace {
			err = a.copyAppRepositorySecret(repoSecret, appRepo)
		}
		if err != nil {
			return nil, err
		}
	}
	return appRepo, nil
}

// RefreshAppRepository forces a refresh in a given apprepository (by updating resyncRequests property)
func (a *userHandler) RefreshAppRepository(repoName string, requestNamespace string) (*v1alpha1.AppRepository, error) {
	// Retrieve the repo object with name=repoName
	appRepo, err := a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).Get(context.TODO(), repoName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// An update is forced if the ResyncRequests property changes,
	// so we increase it in the retrieved object
	appRepo.Spec.ResyncRequests++

	// Update existing repo with the new spec (ie ResyncRequests++)
	appRepo, err = a.clientset.KubeappsV1alpha1().AppRepositories(requestNamespace).Update(context.TODO(), appRepo, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
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

// Deprecated: Remove when the new Package Repository API implementation is completed
func (a *userHandler) getValidationCli(appRepoBody io.ReadCloser, requestNamespace, kubeappsNamespace string) (*v1alpha1.AppRepository, httpclient.Client, error) {
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

	repoSecret, err := a.secretForRequest(appRepoRequest, appRepo, requestNamespace)
	if err != nil {
		return nil, nil, err
	}

	cli, err := InitNetClient(appRepo, repoSecret, repoSecret, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create HTTP client: %w", err)
	}
	return appRepo, cli, nil
}

// repoTagsList stores the list of tags for an OCI repository.
type repoTagsList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type repoConfig struct {
	MediaType string `json:"mediaType"`
}

// repoManifest stores the mediatype for an OCI repository.
type repoManifest struct {
	Config repoConfig `json:"config"`
}

// Deprecated: Remove when the new Package Repository API implementation is completed
//
//	getOCIAppRepositoryTag  get a tag for the given repoURL & repoName
func getOCIAppRepositoryTag(cli httpclient.Client, repoURL string, repoName string) (string, error) {
	// This function is the implementation of below curl command
	// curl -XGET -H "Authorization: Basic $harborauthz"
	//		-H "Accept: application/vnd.oci.image.manifest.v1+json"
	//		-s https://demo.goharbor.io/v2/test10/podinfo/podinfo/tags/list\?n\=1

	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", err
	}

	parsedURL.Path = path.Join("v2", parsedURL.Path, repoName, "tags", "list")
	q := parsedURL.Query()
	q.Add("n", "1")
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	//This header is required for a successful request
	req.Header.Set("Accept", OCIImageManifestMediaType)

	resp, err := cli.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected status code when querying %q: %d", repoName, resp.StatusCode)
	}

	var body []byte
	var repoTagsData repoTagsList

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("io.ReadAll : unable to get: %v", err)
		return "", err
	}

	err = json.Unmarshal(body, &repoTagsData)
	if err != nil {
		err = fmt.Errorf("OCI Repo tag at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}

	if len(repoTagsData.Tags) == 0 {
		err = fmt.Errorf("OCI Repo tag at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}

	tagVersion := repoTagsData.Tags[0]
	return tagVersion, nil
}

// Deprecated: Remove when the new Package Repository API implementation is completed
//
//	getOCIAppRepositoryMediaType  get manifests config.MediaType for the given repoURL & repoName
func getOCIAppRepositoryMediaType(cli httpclient.Client, repoURL string, repoName string, tagVersion string) (string, error) {
	// This function is the implementation of below curl command
	// curl -XGET -H "Authorization: Basic $harborauthz"
	//		 -H "Accept: application/vnd.oci.image.manifest.v1+json"
	//		-s https://demo.goharbor.io/v2/test10/podinfo/podinfo/manifests/6.0.0

	parsedURL, err := url.ParseRequestURI(repoURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = path.Join("v2", parsedURL.Path, repoName, "manifests", tagVersion)

	log.InfoS("parsedURL", "URL", parsedURL.String())
	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	//This header is required for a successful request
	req.Header.Set("Accept", OCIImageManifestMediaType)

	resp, err := cli.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mediaData repoManifest

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &mediaData)
	if err != nil {
		err = fmt.Errorf("OCI Repo manifest at %q could not be parsed: %w", parsedURL.String(), err)
		return "", err
	}
	mediaType := mediaData.Config.MediaType
	return mediaType, nil
}

// ValidateOCIAppRepository validates OCI Repos only
// return true if mediaType == "application/vnd.cncf.helm.config" otherwise false
func ValidateOCIAppRepository(appRepo *v1alpha1.AppRepository, cli httpclient.Client) (bool, error) {

	repoURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/")

	// For the OCI case, we want to validate that all the given repositories are valid
	if len(appRepo.Spec.OCIRepositories) == 0 {
		return false, ErrEmptyOCIRegistry
	}
	for _, repoName := range appRepo.Spec.OCIRepositories {
		tagVersion, err := getOCIAppRepositoryTag(cli, repoURL, repoName)
		if err != nil {
			return false, err
		}

		mediaType, err := getOCIAppRepositoryMediaType(cli, repoURL, repoName, tagVersion)
		if err != nil {
			return false, err
		}

		if !strings.HasPrefix(mediaType, "application/vnd.cncf.helm.config") {
			err := fmt.Errorf("%v is not a Helm OCI Repo. mediaType starting with %q expected, found %q", repoName, "application/vnd.cncf.helm.config", mediaType)
			return false, err
		}
	}
	return true, nil
}

// HttpValidator is an interface for checking the validity of an AppRepo via Http requests.
type HttpValidator interface {
	// Validate returns a validation response.
	Validate(cli httpclient.Client) (*ValidationResponse, error)
}

// HelmNonOCIValidator is an HttpValidator for non-OCI Helm repositories.
type HelmNonOCIValidator struct {
	Req *http.Request
}

func (r HelmNonOCIValidator) Validate(cli httpclient.Client) (*ValidationResponse, error) {

	res, err := cli.Do(r.Req)
	if err != nil {
		// If the request fail, it's not an internal error
		return &ValidationResponse{Code: 400, Message: err.Error()}, nil
	}
	response := &ValidationResponse{Code: res.StatusCode, Message: "OK"}
	if response.Code != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse validation response. Got: %w", err)
		}
		response.Message = string(body)
	}

	return response, nil
}

type HelmOCIValidator struct {
	AppRepo *v1alpha1.AppRepository
}

func (r HelmOCIValidator) Validate(cli httpclient.Client) (*ValidationResponse, error) {

	var response *ValidationResponse
	response = &ValidationResponse{Code: 200, Message: "OK"}

	// If there was an error validating the OCI repository, it's not an internal error.
	isValidRepo, err := ValidateOCIAppRepository(r.AppRepo, cli)
	if err != nil || !isValidRepo {
		response = &ValidationResponse{Code: 400, Message: err.Error()}
	}
	return response, nil
}

// Deprecated: Remove when the new Package Repository API implementation is completed
// getValidator return appropriate HttpValidator interface for OCI and non-OCI Repos
func getValidator(appRepo *v1alpha1.AppRepository) (HttpValidator, error) {

	repoURL := strings.TrimSuffix(strings.TrimSpace(appRepo.Spec.URL), "/")

	if appRepo.Spec.Type == "oci" {
		// For the OCI case, we want to validate that all the given repositories are valid
		if len(appRepo.Spec.OCIRepositories) == 0 {
			return nil, ErrEmptyOCIRegistry
		}
		return HelmOCIValidator{
			AppRepo: appRepo,
		}, nil
	} else {
		parsedURL, err := url.ParseRequestURI(repoURL)
		if err != nil {
			return nil, err
		}
		parsedURL.Path = path.Join(parsedURL.Path, "index.yaml")
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			return nil, err
		}
		return HelmNonOCIValidator{
			Req: req,
		}, nil

	}
}

// Deprecated: Remove when the new Package Repository API implementation is completed
func (a *userHandler) ValidateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*ValidationResponse, error) {
	// Split body parsing to a different function for ease testing
	appRepo, cli, err := a.getValidationCli(appRepoBody, requestNamespace, a.kubeappsNamespace)
	if err != nil {
		return &ValidationResponse{
			Code:    400,
			Message: err.Error(),
		}, nil
	}
	httpValidator, err := getValidator(appRepo)
	if err != nil {
		return nil, err
	}
	response, err := httpValidator.Validate(cli)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetAppRepository returns an AppRepository resource from a namespace.
// Optionally set a token to get the AppRepository using a custom serviceaccount
func (a *userHandler) GetAppRepository(repoName, repoNamespace string) (*v1alpha1.AppRepository, error) {
	return a.clientset.KubeappsV1alpha1().AppRepositories(repoNamespace).Get(context.TODO(), repoName, metav1.GetOptions{})
}

// Deprecated: Remove when the new Package Repository API implementation is completed
// appRepositoryForRequest takes care of parsing the request data into an AppRepository.
func appRepositoryForRequest(appRepoRequest *appRepositoryRequest) *v1alpha1.AppRepository {
	appRepo := appRepoRequest.AppRepository

	var auth v1alpha1.AppRepositoryAuth
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
	if appRepo.AuthRegCreds != "" {
		auth.Header = &v1alpha1.AppRepositoryAuthHeader{
			SecretKeyRef: corev1.SecretKeySelector{
				Key: ".dockerconfigjson",
				LocalObjectReference: corev1.LocalObjectReference{
					Name: appRepo.AuthRegCreds,
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

	if appRepo.Type == "" {
		// Use helm type by default
		appRepo.Type = "helm"
	}
	return &v1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name: appRepo.Name,
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:                   appRepo.RepoURL,
			Type:                  appRepo.Type,
			Auth:                  auth,
			DockerRegistrySecrets: appRepo.RegistrySecrets,
			SyncJobPodTemplate:    appRepo.SyncJobPodTemplate,
			ResyncRequests:        appRepo.ResyncRequests,
			OCIRepositories:       appRepo.OCIRepositories,
			TLSInsecureSkipVerify: appRepo.TLSInsecureSkipVerify,
			FilterRule:            appRepo.FilterRule,
			Description:           appRepo.Description,
			PassCredentials:       appRepo.PassCredentials,
		},
	}
}

// Deprecated: Remove when the new Package Repository API implementation is completed
// secretForRequest takes care of parsing the request data into a secret for an AppRepository.
func (a *userHandler) secretForRequest(appRepoRequest *appRepositoryRequest, appRepo *v1alpha1.AppRepository, namespace string) (*corev1.Secret, error) {
	if len(appRepoRequest.AppRepository.AuthRegCreds) > 0 {
		return a.GetSecret(appRepoRequest.AppRepository.AuthRegCreds, namespace)
	}
	appRepoDetails := appRepoRequest.AppRepository
	secrets := map[string]string{}
	if appRepoDetails.AuthHeader != "" {
		secrets["authorizationHeader"] = appRepoDetails.AuthHeader
	}
	if appRepoDetails.CustomCA != "" {
		secrets["ca.crt"] = appRepoDetails.CustomCA
	}

	if len(secrets) == 0 {
		return nil, nil
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
	}, nil
}

// Deprecated: Remove when the new Package Repository API implementation is completed
func secretNameForRepo(repoName string) string {
	return fmt.Sprintf("apprepo-%s", repoName)
}

// KubeappsSecretNameForRepo returns a name suitable for recording a copy of
// a per-namespace repository secret in the kubeapps namespace.
func KubeappsSecretNameForRepo(repoName, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, secretNameForRepo(repoName))
}

// GetSecret return the a secret from a namespace using a token if given
func (a *userHandler) GetSecret(name, namespace string) (*corev1.Secret, error) {
	return a.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
