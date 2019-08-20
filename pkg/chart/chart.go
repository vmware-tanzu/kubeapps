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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

const (
	defaultNamespace      = metav1.NamespaceSystem
	defaultRepoURL        = "https://kubernetes-charts.storage.googleapis.com"
	defaultTimeoutSeconds = 180
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
	// RepoURL is the URL of the repository. Defaults to stable repo.
	RepoURL string `json:"repoUrl,omitempty"`
	// AppRepositoryResourceName specifies an app repository resource to use
	// for the request.
	// TODO(absoludity): Intended to supercede RepoURL and Auth below. Remove
	// RepoURL and Auth once #1110 complete.
	AppRepositoryResourceName string `json:"appRepositoryResourceName,omitempty"`
	// ChartName is the name of the chart within the repo.
	ChartName string `json:"chartName"`
	// ReleaseName is the Name of the release given to Tiller.
	ReleaseName string `json:"releaseName"`
	// Version is the chart version.
	Version string `json:"version"`
	// Auth is the authentication.
	Auth Auth `json:"auth,omitempty"`
	// Values is a string containing (unparsed) YAML values.
	Values string `json:"values,omitempty"`
}

// Auth contains the information to authenticate against a private registry
type Auth struct {
	// Header is header based Authorization
	Header *AuthHeader `json:"header,omitempty"`
	// CustomCA is an additional CA
	CustomCA *CustomCA `json:"customCA,omitempty"`
}

// AuthHeader contains the secret information for authenticate
type CustomCA struct {
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// AuthHeader contains the secret information for authenticate
type AuthHeader struct {
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// HTTPClient Interface to perform HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// LoadChart should return a Chart struct from an IOReader
type LoadChart func(in io.Reader) (*chart.Chart, error)

// Resolver for exposed funcs
type Resolver interface {
	ParseDetails(data []byte) (*Details, error)
	GetChart(details *Details, netClient HTTPClient) (*chart.Chart, error)
	InitNetClient(details *Details) (HTTPClient, error)
}

// Chart struct contains the clients required to retrieve charts info
type Chart struct {
	kubeClient kubernetes.Interface
	load       LoadChart
	userAgent  string
}

// NewChart returns a new Chart
func NewChart(kubeClient kubernetes.Interface, load LoadChart, userAgent string) *Chart {
	return &Chart{
		kubeClient,
		load,
		userAgent,
	}
}

func getReq(rawURL, authHeader string) (*http.Request, error) {
	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(authHeader) > 0 {
		req.Header.Set("Authorization", authHeader)
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
func fetchRepoIndex(netClient *HTTPClient, repoURL string, authHeader string) (*repo.IndexFile, error) {
	req, err := getReq(repoURL, authHeader)
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

// fetchChart returns the Chart content given an URL and the auth header if needed
func fetchChart(netClient *HTTPClient, chartURL, authHeader string, load LoadChart) (*chart.Chart, error) {
	req, err := getReq(chartURL, authHeader)
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
	return load(bytes.NewReader(data))
}

// ParseDetails return Chart details
func (c *Chart) ParseDetails(data []byte) (*Details, error) {
	details := &Details{}
	err := json.Unmarshal(data, details)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request body: %v", err)
	}

	if (details.RepoURL != "" || details.Auth.Header != nil || details.Auth.CustomCA != nil) && details.AppRepositoryResourceName != "" {
		return nil, fmt.Errorf("repoUrl or auth specified together with appRepositoryResourceName")
	}
	return details, nil
}

// clientWithDefaultHeaders implements chart.HTTPClient interface
// and includes an override of the Do method which injects our default
// headers - User-Agent and Authorization (when present)
type clientWithDefaultHeaders struct {
	client         HTTPClient
	defaultHeaders http.Header
}

// Do HTTP request
func (c *clientWithDefaultHeaders) Do(req *http.Request) (*http.Response, error) {
	for k, v := range c.defaultHeaders {
		// Only add the default header if it's not already set in the request.
		if _, ok := req.Header[k]; !ok {
			req.Header[k] = v
		}
	}
	return c.client.Do(req)
}

// InitNetClient returns an HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func (c *Chart) InitNetClient(details *Details) (HTTPClient, error) {
	// Get the SystemCertPool, continue with an empty pool on error
	caCertPool, _ := x509.SystemCertPool()
	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
	}

	// If additionalCA is set, load it
	customCA := details.Auth.CustomCA
	if customCA != nil {
		namespace := os.Getenv("POD_NAMESPACE")
		caCertSecret, err := c.kubeClient.CoreV1().Secrets(namespace).Get(customCA.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("unable to read secret %q: %v", customCA.SecretKeyRef.Name, err)
		}

		// Append our cert to the system pool
		customData, ok := caCertSecret.Data[customCA.SecretKeyRef.Key]
		if !ok {
			return nil, fmt.Errorf("secret %q did not contain key %q", customCA.SecretKeyRef.Name, customCA.SecretKeyRef.Key)
		}
		if ok := caCertPool.AppendCertsFromPEM(customData); !ok {
			return nil, fmt.Errorf("Failed to append %s to RootCAs", customCA.SecretKeyRef.Name)
		}
	}

	// Return Transport for testing purposes
	return &clientWithDefaultHeaders{
		client: &http.Client{
			Timeout: time.Second * defaultTimeoutSeconds,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{
					RootCAs: caCertPool,
				},
			},
		},
		defaultHeaders: http.Header{"User-Agent": []string{c.userAgent}},
	}, nil
}

// GetChart retrieves and loads a Chart from a registry
func (c *Chart) GetChart(details *Details, netClient HTTPClient) (*chart.Chart, error) {
	repoURL := details.RepoURL
	if repoURL == "" {
		// FIXME: Make configurable
		repoURL = defaultRepoURL
	}
	repoURL = strings.TrimSuffix(strings.TrimSpace(repoURL), "/") + "/index.yaml"

	authHeader := ""
	if details.Auth.Header != nil {
		namespace := os.Getenv("POD_NAMESPACE")
		if namespace == "" {
			namespace = defaultNamespace
		}

		secret, err := c.kubeClient.Core().Secrets(namespace).Get(details.Auth.Header.SecretKeyRef.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		authHeader = string(secret.Data[details.Auth.Header.SecretKeyRef.Key])
	}

	log.Printf("Downloading repo %s index...", repoURL)
	repoIndex, err := fetchRepoIndex(&netClient, repoURL, authHeader)
	if err != nil {
		return nil, err
	}

	chartURL, err := findChartInRepoIndex(repoIndex, repoURL, details.ChartName, details.Version)
	if err != nil {
		return nil, err
	}

	log.Printf("Downloading %s ...", chartURL)
	chartRequested, err := fetchChart(&netClient, chartURL, authHeader, c.load)
	if err != nil {
		return nil, err
	}
	return chartRequested, nil
}
