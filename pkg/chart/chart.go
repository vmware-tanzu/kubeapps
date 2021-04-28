/*
Copyright (c) 2018 Bitnami

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

package chart

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/ghodss/yaml"
	appRepov1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/helm"
	"github.com/kubeapps/kubeapps/pkg/kube"
	helm3chart "helm.sh/helm/v3/pkg/chart"
	helm3loader "helm.sh/helm/v3/pkg/chart/loader"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/helm/pkg/repo"
	"k8s.io/kubernetes/pkg/credentialprovider"
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
	// ReleaseName is the Name of the release given to Tiller.
	ReleaseName string `json:"releaseName"`
	// Version is the chart version.
	Version string `json:"version"`
	// Values is a string containing (unparsed) YAML values.
	Values string `json:"values,omitempty"`
}

// LoadHelmChart returns a helm3 Chart struct from an IOReader
type LoadHelmChart func(in io.Reader) (*helm3chart.Chart, error)

// Resolver for exposed funcs
type Resolver interface {
	InitClient(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error
	GetChart(details *Details, repoURL string) (*helm3chart.Chart, error)
}

// Client struct contains the clients required to retrieve charts info
type Client struct {
	userAgent string
	netClient kube.HTTPClient
}

// NewChartClient returns a new ChartClient
func NewChartClient(userAgent string) Resolver {
	return &Client{
		userAgent: userAgent,
	}
}

// OCIClient struct contains the clients required to retrieve charts info from an OCI registry
type OCIClient struct {
	userAgent string
	puller    helm.ChartPuller
}

// NewOCIClient returns a new OCIClient
func NewOCIClient(userAgent string) Resolver {
	return &OCIClient{
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

func readResponseBody(res *http.Response) ([]byte, error) {
	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("chart download request failed")
	}

	body, err := ioutil.ReadAll(res.Body)
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
// is an expensive operation. See https://github.com/kubeapps/kubeapps/issues/1052
func getIndexFromCache(repoURL string, data []byte) (*repo.IndexFile, string) {
	sha := checksum(data)
	if repoIndexes[repoURL] == nil || repoIndexes[repoURL].checksum != sha {
		// The repository is not in the cache or the content changed
		return nil, sha
	}
	return repoIndexes[repoURL].index, sha
}

func storeIndexInCache(repoURL string, index *repo.IndexFile, sha string) {
	repoIndexes[repoURL] = &repoIndex{sha, index}
}

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
func fetchRepoIndex(netClient *kube.HTTPClient, repoURL string) (*repo.IndexFile, error) {
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

func resolveChartURL(index, chart string) (string, error) {
	indexURL, err := url.Parse(strings.TrimSpace(index))
	if err != nil {
		return "", err
	}
	chartURL, err := indexURL.Parse(strings.TrimSpace(chart))
	if err != nil {
		return "", err
	}
	return chartURL.String(), nil
}

// findChartInRepoIndex returns the URL of a chart given a Helm repository and its name and version
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
func fetchChart(netClient *kube.HTTPClient, chartURL string) (*helm3chart.Chart, error) {
	req, err := getReq(chartURL)
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
	return helm3loader.LoadArchive(bytes.NewReader(data))
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

// InitClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func (c *Client) InitClient(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	var err error
	c.netClient, err = kube.InitNetClient(appRepo, caCertSecret, authSecret, http.Header{"User-Agent": []string{c.userAgent}})
	return err
}

// GetChart retrieves and loads a Chart from a registry in both
// v2 and v3 formats.
func (c *Client) GetChart(details *Details, repoURL string) (*helm3chart.Chart, error) {
	if c.netClient == nil {
		return nil, fmt.Errorf("unable to retrieve chart, InitClient should be called first")
	}
	var chart *helm3chart.Chart
	indexURL := strings.TrimSuffix(strings.TrimSpace(repoURL), "/") + "/index.yaml"
	repoIndex, err := fetchRepoIndex(&c.netClient, indexURL)
	if err != nil {
		return nil, err
	}

	chartURL, err := findChartInRepoIndex(repoIndex, indexURL, details.ChartName, details.Version)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading %s ...", chartURL)
	chart, err = fetchChart(&c.netClient, chartURL)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// RegistrySecretsPerDomain checks the app repo and available secrets
// to return the secret names per registry domain.
func RegistrySecretsPerDomain(appRepoSecrets []string, cluster, namespace, token string, authHandler kube.AuthHandler) (map[string]string, error) {
	secretsPerDomain := map[string]string{}
	client, err := authHandler.AsUser(token, cluster)
	if err != nil {
		return nil, err
	}

	for _, secretName := range appRepoSecrets {
		secret, err := client.GetSecret(secretName, namespace)
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

// InitClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
// TODO(andresmgot): Using a custom CA cert is not supported by ORAS (neither helm), only using the insecure flag
func (c *OCIClient) InitClient(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
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
		auth, err = kube.GetData(appRepo.Spec.Auth.Header.SecretKeyRef.Key, authSecret)
		if err != nil {
			return err
		}
		headers.Set("Authorization", string(auth))
	}

	c.puller = &helm.OCIPuller{Resolver: docker.NewResolver(docker.ResolverOptions{Headers: headers, Client: netClient})}
	return err
}

// GetChart retrieves and loads a Chart from a OCI registry
func (c *OCIClient) GetChart(details *Details, repoURL string) (*helm3chart.Chart, error) {
	if c.puller == nil {
		return nil, fmt.Errorf("unable to retrieve chart, InitClient should be called first")
	}
	url, err := url.ParseRequestURI(strings.TrimSpace(repoURL))
	if err != nil {
		return nil, err
	}

	ref := path.Join(url.Host, url.Path, fmt.Sprintf("%s:%s", details.ChartName, details.Version))
	chartBuffer, _, err := c.puller.PullOCIChart(ref)
	if err != nil {
		return nil, err
	}

	return helm3loader.LoadArchive(chartBuffer)
}
