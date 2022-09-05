// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package chart

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/containerd/containerd/remotes/docker"
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/credentialprovider"
	"sigs.k8s.io/yaml"
)

const (
	dockerConfigJSONType = "kubernetes.io/dockerconfigjson"
	dockerConfigJSONKey  = ".dockerconfigjson"
)

type repoIndex struct {
	checksum string
	index    *repo.IndexFile
}

var repoIndexes map[string]*repoIndex

func init() {
	repoIndexes = map[string]*repoIndex{}
}

// Details contains the information to retrieve a Chart
type Details struct {
	// AppRepositoryResourceName specifies an app repository resource to use
	// for the request.
	AppRepositoryResourceName string `json:"appRepositoryResourceName,omitempty"`
	// AppRepositoryResourceNamespace specifies the namespace for the app repository
	AppRepositoryResourceNamespace string `json:"appRepositoryResourceNamespace,omitempty"`
	// resource for the request.
	// ChartName is the name of the chart within the repo.
	ChartName string `json:"chartName"`
	// ReleaseName is the Name of the release given to Helm.
	ReleaseName string `json:"releaseName"`
	// Version is the chart version.
	Version string `json:"version"`
	// Values is a string containing (unparsed) YAML values.
	Values string `json:"values,omitempty"`
	// TarballURL is the URL to the tarball file
	TarballURL string `json:"tarballURL,omitempty"`
}

// LoadHelmChart returns a helm3 Chart struct from an IOReader
type LoadHelmChart func(in io.Reader) (*chart.Chart, error)

// ChartClient for exposed funcs
type ChartClient interface {
	Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error
	GetChart(details *Details, repoURL string) (*chart.Chart, error)
}

// HelmRepoClient struct contains the clients required to retrieve charts info
type HelmRepoClient struct {
	userAgent string
	netClient httpclient.Client
}

// NewChartClient returns a new ChartClient
func NewChartClient(userAgent string) ChartClient {
	return &HelmRepoClient{
		userAgent: userAgent,
	}
}

// OCIRepoClient struct contains the clients required to retrieve charts info from an OCI registry
type OCIRepoClient struct {
	userAgent string
	puller    helm.ChartPuller
}

// NewOCIClient returns a new OCIClient
func NewOCIClient(userAgent string) ChartClient {
	return &OCIRepoClient{
		userAgent: userAgent,
	}
}

func getReq(rawURL string) (*http.Request, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// TODO(agamez): remove this method once it is no longer used in kubeops
func readResponseBody(res *http.Response) ([]byte, error) {
	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chart download request failed")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func checksum(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return string(hasher.Sum(nil))
}

// Cache the result of parsing the repo index since parsing this YAML
// is an expensive operation. See https://github.com/vmware-tanzu/kubeapps/issues/1052
// TODO(agamez): remove this method once it is no longer used in kubeops
func getIndexFromCache(repoURL string, data []byte) (*repo.IndexFile, string) {
	sha := checksum(data)
	if repoIndexes[repoURL] == nil || repoIndexes[repoURL].checksum != sha {
		// The repository is not in the cache or the content changed
		return nil, sha
	}
	return repoIndexes[repoURL].index, sha
}

// TODO(agamez): remove this method once it is no longer used in kubeops
func storeIndexInCache(repoURL string, index *repo.IndexFile, sha string) {
	repoIndexes[repoURL] = &repoIndex{sha, index}
}

// TODO(agamez): remove this method once it is no longer used in kubeops
func parseIndex(data []byte) (*repo.IndexFile, error) {
	index := &repo.IndexFile{}
	err := yaml.Unmarshal(data, index)
	if err != nil {
		return index, err
	}
	index.SortEntries()
	return index, nil
}

// fetchRepoIndex returns a Helm repository
// TODO(agamez): remove this method once it is no longer used in kubeops
func fetchRepoIndex(netClient *httpclient.Client, repoURL string) (*repo.IndexFile, error) {
	req, err := getReq(repoURL)
	if err != nil {
		return nil, err
	}

	res, err := (*netClient).Do(req)
	if err != nil {
		return nil, err
	}
	data, err := readResponseBody(res)
	if err != nil {
		return nil, err
	}

	index, sha := getIndexFromCache(repoURL, data)
	if index == nil {
		// index not found in the cache, parse it
		index, err = parseIndex(data)
		if err != nil {
			return nil, err
		}
		storeIndexInCache(repoURL, index, sha)
	}
	return index, nil
}

func resolveChartURL(indexURL, chartURL string) (string, error) {
	parsedIndexURL, err := url.Parse(strings.TrimSpace(indexURL))
	if err != nil {
		return "", err
	}
	parsedChartURL, err := parsedIndexURL.Parse(strings.TrimSpace(chartURL))
	if err != nil {
		return "", err
	}
	return parsedChartURL.String(), nil
}

// findChartInRepoIndex returns the URL of a chart given a Helm repository and its name and version
// TODO(agamez): remove this method once it is no longer used in kubeops
func findChartInRepoIndex(repoIndex *repo.IndexFile, repoURL, chartName, chartVersion string) (string, error) {
	errMsg := fmt.Sprintf("chart %q", chartName)
	if chartVersion != "" {
		errMsg = fmt.Sprintf("%s version %q", errMsg, chartVersion)
	}
	cv, err := repoIndex.Get(chartName, chartVersion)
	if err != nil {
		return "", fmt.Errorf("%s not found in repository", errMsg)
	}
	if len(cv.URLs) == 0 {
		return "", fmt.Errorf("%s has no downloadable URLs", errMsg)
	}
	return resolveChartURL(repoURL, cv.URLs[0])
}

// fetchChart returns the Chart content given an URL
func fetchChart(netClient *httpclient.Client, chartURL string) (*chart.Chart, error) {
	req, err := getReq(chartURL)
	if err != nil {
		return nil, err
	}

	res, err := (*netClient).Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chart download request failed")
	}

	return loader.LoadArchive(res.Body)
}

// ParseDetails return Chart details
func ParseDetails(data []byte) (*Details, error) {
	details := &Details{}
	err := json.Unmarshal(data, details)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request body: %v", err)
	}

	if details.AppRepositoryResourceName == "" {
		return nil, fmt.Errorf("an AppRepositoryResourceName is required")
	}

	if details.AppRepositoryResourceNamespace == "" {
		return nil, fmt.Errorf("an AppRepositoryResourceNamespace is required")
	}

	return details, nil
}

// GetAppRepoAndRelatedSecrets retrieves the given repo from its namespace
// Depending on the repo namespace and the
func GetAppRepoAndRelatedSecrets(appRepoName, appRepoNamespace string, handler kube.AuthHandler, userAuthToken, cluster, kubeappsNamespace string, kubeappsCluster string) (*appRepov1.AppRepository, *corev1.Secret, *corev1.Secret, error) {
	client, err := handler.AsUser(userAuthToken, cluster)
	// If the UI clusters configuration did not include the cluster on which Kubeapps is installed then
	// we won't know the kubeappsNamespace but it will be empty. In this scenario Kubeapps only supports
	// global app repositories (#1982).
	isKubeappsCluster := cluster == kubeappsCluster
	nonUIKubeappsCluster := cluster == ""
	isGlobalAppRepository := kubeappsNamespace == appRepoNamespace
	if isKubeappsCluster && (nonUIKubeappsCluster || isGlobalAppRepository) {
		// If we're parsing a global repository then use a service client.
		// AppRepositories are only allowed in the default cluster for the moment
		client, err = handler.AsSVC(cluster)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create clientset: %v", err)
	}
	appRepo, err := client.GetAppRepository(appRepoName, appRepoNamespace)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to get app repository %q: %v", appRepoName, err)
	}

	auth := appRepo.Spec.Auth
	var caCertSecret *corev1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		caCertSecret, err = client.GetSecret(secretName, appRepo.Namespace)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *corev1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = client.GetSecret(secretName, appRepo.Namespace)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return appRepo, caCertSecret, authSecret, nil
}

// Init initialises the HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func (c *HelmRepoClient) Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	var err error
	c.netClient, err = kube.InitNetClient(appRepo, caCertSecret, authSecret, http.Header{"User-Agent": []string{c.userAgent}})
	return err
}

// GetChart loads a Chart from a given tarball, if the tarball URL is not passed,
// it will try to retrieve the chart by parsing the whole repo index
func (c *HelmRepoClient) GetChart(details *Details, repoURL string) (*chart.Chart, error) {
	if c.netClient == nil {
		return nil, fmt.Errorf("unable to retrieve chart, Init should be called first")
	}
	var chartURL string
	var err error
	if details.TarballURL != "" {
		chartURL, err = resolveChartURL(repoURL, details.TarballURL)
		if err != nil {
			return nil, err
		}
	} else {
		// TODO(agamez): remove this branch as it is really expensive and it is solely used in a few places in kubeops
		log.Info("calling GetChart without any tarball url, please note this action is memory-expensive")
		indexURL := strings.TrimSuffix(strings.TrimSpace(repoURL), "/") + "/index.yaml"
		repoIndex, err := fetchRepoIndex(&c.netClient, indexURL)
		if err != nil {
			return nil, err
		}
		chartURL, err = findChartInRepoIndex(repoIndex, indexURL, details.ChartName, details.Version)
		if err != nil {
			return nil, err
		}
	}
	log.Infof("Downloading %s ...", chartURL)
	chart, err := fetchChart(&c.netClient, chartURL)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// RegistrySecretsPerDomain checks the app repo and available secrets
// to return the secret names per registry domain.
func RegistrySecretsPerDomain(ctx context.Context, appRepoSecrets []string, namespace string, client kubernetes.Interface) (map[string]string, error) {
	secretsPerDomain := map[string]string{}

	for _, secretName := range appRepoSecrets {
		secret, err := client.CoreV1().Secrets(namespace).Get(ctx, secretName, v1.GetOptions{})
		if err != nil {
			return nil, err
		}

		if secret.Type != dockerConfigJSONType {
			return nil, fmt.Errorf("AppRepository secret must be of type %q. Secret %q had type %q", dockerConfigJSONType, secretName, secret.Type)
		}

		dockerConfigJSONBytes, ok := secret.Data[dockerConfigJSONKey]
		if !ok {
			return nil, fmt.Errorf("AppRepository secret must have a data map with a key %q. Secret %q did not", dockerConfigJSONKey, secretName)
		}

		dockerConfigJSON := credentialprovider.DockerConfigJSON{}
		if err := json.Unmarshal(dockerConfigJSONBytes, &dockerConfigJSON); err != nil {
			return nil, err
		}

		for key := range dockerConfigJSON.Auths {
			secretsPerDomain[key] = secretName
		}

	}
	return secretsPerDomain, nil
}

// Init initialises the HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
// TODO(andresmgot): Using a custom CA cert is not supported by ORAS (neither helm), only using the insecure flag
func (c *OCIRepoClient) Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	var err error
	headers := http.Header{
		"User-Agent": []string{c.userAgent},
	}
	netClient, err := kube.InitHTTPClient(appRepo, caCertSecret)
	if err != nil {
		return err
	}
	if authSecret != nil && appRepo.Spec.Auth.Header != nil {
		var auth string
		auth, err = kube.GetDataFromSecret(appRepo.Spec.Auth.Header.SecretKeyRef.Key, authSecret)
		if err != nil {
			return err
		}
		headers.Set("Authorization", string(auth))
	}

	c.puller = &helm.OCIPuller{Resolver: docker.NewResolver(docker.ResolverOptions{Headers: headers, Client: netClient})}
	return err
}

// GetChart retrieves and loads a Chart from a OCI registry
func (c *OCIRepoClient) GetChart(details *Details, repoURL string) (*chart.Chart, error) {
	if c.puller == nil {
		return nil, fmt.Errorf("unable to retrieve chart, Init should be called first")
	}
	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(repoURL))
	if err != nil {
		return nil, err
	}
	unescapedChartName, err := url.QueryUnescape(details.ChartName)
	if err != nil {
		return nil, err
	}

	ref := path.Join(parsedURL.Host, parsedURL.Path, fmt.Sprintf("%s:%s", unescapedChartName, details.Version))
	chartBuffer, _, err := c.puller.PullOCIChart(ref)
	if err != nil {
		return nil, err
	}

	return loader.LoadArchive(chartBuffer)
}

// ChartClientFactoryInterface defines how a ChartClientFactory implementation
// can return a chart client.
//
// This can be implemented with a fake for tests.
type ChartClientFactoryInterface interface {
	New(repoType, userAgent string) ChartClient
}

// ChartClientFactory provides a real implementation of the ChartClientFactory interface
// returning either an OCI repository client or a traditional helm repository chart client.
type ChartClientFactory struct{}

// New for ClientResolver
func (c *ChartClientFactory) New(repoType, userAgent string) ChartClient {
	var client ChartClient
	switch repoType {
	case "oci":
		client = NewOCIClient(userAgent)
	default:
		client = NewChartClient(userAgent)
	}
	return client
}
