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
	"strings"

	"github.com/ghodss/yaml"
	appRepov1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	helm3chart "helm.sh/helm/v3/pkg/chart"
	helm3loader "helm.sh/helm/v3/pkg/chart/loader"
	corev1 "k8s.io/api/core/v1"
	helm2loader "k8s.io/helm/pkg/chartutil"
	helm2chart "k8s.io/helm/pkg/proto/hapi/chart"
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

// ChartMultiVersion includes both Helm2Chart and Helm3Chart
type ChartMultiVersion struct {
	Helm2Chart *helm2chart.Chart
	Helm3Chart *helm3chart.Chart
}

// LoadHelm2Chart should return a helm2 Chart struct from an IOReader
type LoadHelm2Chart func(in io.Reader) (*helm2chart.Chart, error)

// LoadHelm3Chart returns a helm3 Chart struct from an IOReader
type LoadHelm3Chart func(in io.Reader) (*helm3chart.Chart, error)

// Resolver for exposed funcs
type Resolver interface {
	ParseDetails(data []byte) (*Details, error)
	GetChart(details *Details, netClient kube.HTTPClient, requireV1Support bool) (*ChartMultiVersion, error)
	InitNetClient(details *Details, userAuthToken string) (kube.HTTPClient, error)
	RegistrySecretsPerDomain() map[string]string
}

// Client struct contains the clients required to retrieve charts info
type Client struct {
	appRepoHandler           kube.AuthHandler
	userAgent                string
	kubeappsCluster          string
	kubeappsNamespace        string
	appRepo                  *appRepov1.AppRepository
	registrySecretsPerDomain map[string]string
}

// NewChartClient returns a new ChartClient
func NewChartClient(appRepoHandler kube.AuthHandler, kubeappsCluster, kubeappsNamespace, userAgent string) *Client {
	return &Client{
		appRepoHandler:    appRepoHandler,
		userAgent:         userAgent,
		kubeappsCluster:   kubeappsCluster,
		kubeappsNamespace: kubeappsNamespace,
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
func fetchChart(netClient *kube.HTTPClient, chartURL string, requireV1Support bool) (*ChartMultiVersion, error) {
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
	// We only return an error when loading using the helm2loader (ie. chart v1)
	// if we require v1 support, otherwise we continue to load using the
	// helm3 v2 loader.
	helm2Chart, err := helm2loader.LoadArchive(bytes.NewReader(data))
	if err != nil && requireV1Support {
		return nil, err
	}
	helm3Chart, err := helm3loader.LoadArchive(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &ChartMultiVersion{Helm2Chart: helm2Chart, Helm3Chart: helm3Chart}, nil
}

// ParseDetails return Chart details
func (c *Client) ParseDetails(data []byte) (*Details, error) {
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

func (c *Client) parseDetailsForHTTPClient(details *Details, userAuthToken string) (*appRepov1.AppRepository, *corev1.Secret, *corev1.Secret, error) {
	// We grab the specified app repository (for later access to the repo URL, as well as any specified
	// auth).
	client, err := c.appRepoHandler.AsUser(userAuthToken, c.kubeappsCluster)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to create clientset: %v", err)
	}
	if details.AppRepositoryResourceNamespace == c.kubeappsNamespace {
		// If we're parsing a global repository (from the kubeappsNamespace), use a service client.
		client = c.appRepoHandler.AsSVC()
	}
	appRepo, err := client.GetAppRepository(details.AppRepositoryResourceName, details.AppRepositoryResourceNamespace)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to get app repository %q: %v", details.AppRepositoryResourceName, err)
	}
	c.appRepo = appRepo
	auth := appRepo.Spec.Auth

	var caCertSecret *corev1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		caCertSecret, err = client.GetSecret(secretName, details.AppRepositoryResourceNamespace)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *corev1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = client.GetSecret(secretName, details.AppRepositoryResourceNamespace)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return appRepo, caCertSecret, authSecret, nil
}

// InitNetClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func (c *Client) InitNetClient(details *Details, userAuthToken string) (kube.HTTPClient, error) {
	appRepo, caCertSecret, authSecret, err := c.parseDetailsForHTTPClient(details, userAuthToken)
	if err != nil {
		return nil, err
	}

	c.registrySecretsPerDomain, err = getRegistrySecretsPerDomain(c.appRepo.Spec.DockerRegistrySecrets, c.kubeappsCluster, details.AppRepositoryResourceNamespace, userAuthToken, c.appRepoHandler)
	if err != nil {
		return nil, err
	}

	return kube.InitNetClient(appRepo, caCertSecret, authSecret, http.Header{"User-Agent": []string{c.userAgent}})
}

// GetChart retrieves and loads a Chart from a registry in both
// v2 and v3 formats.
func (c *Client) GetChart(details *Details, netClient kube.HTTPClient, requireV1Support bool) (*ChartMultiVersion, error) {
	indexURL := strings.TrimSuffix(strings.TrimSpace(c.appRepo.Spec.URL), "/") + "/index.yaml"

	repoIndex, err := fetchRepoIndex(&netClient, indexURL)
	if err != nil {
		return nil, err
	}

	chartURL, err := findChartInRepoIndex(repoIndex, indexURL, details.ChartName, details.Version)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading %s ...", chartURL)
	chart, err := fetchChart(&netClient, chartURL, requireV1Support)
	if err != nil {
		return nil, err
	}

	return chart, nil
}

// RegistrySecretsPerDomain checks the app repo and available secrets
// to return the secret names per registry domain.
//
// These are actually calculated during InitNetClient when we already have a
// k8s client with the user token.
func (c *Client) RegistrySecretsPerDomain() map[string]string {
	return c.registrySecretsPerDomain
}

func getRegistrySecretsPerDomain(appRepoSecrets []string, cluster, namespace, token string, authHandler kube.AuthHandler) (map[string]string, error) {
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

		dockerConfigJSON := credentialprovider.DockerConfigJson{}
		if err := json.Unmarshal(dockerConfigJSONBytes, &dockerConfigJSON); err != nil {
			return nil, err
		}

		for key, _ := range dockerConfigJSON.Auths {
			secretsPerDomain[key] = secretName
		}

	}
	return secretsPerDomain, nil
}
