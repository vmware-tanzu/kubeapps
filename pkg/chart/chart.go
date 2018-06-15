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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

const (
	defaultNamespace = metav1.NamespaceSystem
	defaultRepoURL   = "https://kubernetes-charts.storage.googleapis.com"
)

// Details contains the information to retrieve a Chart
type Details struct {
	// RepoURL is the URL of the repository. Defaults to stable repo.
	RepoURL string `json:"repoUrl,omitempty"`
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

	return parseIndex(data)
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

// LoadChart should return a Chart struct from an IOReader
type LoadChart func(in io.Reader) (*chart.Chart, error)

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
func ParseDetails(data []byte) (*Details, error) {
	details := &Details{}
	err := json.Unmarshal(data, details)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse request body: %v", err)
	}
	return details, nil
}

// GetChart retrieves and loads a Chart from a registry
func GetChart(details *Details, kubeClient kubernetes.Interface, netClient HTTPClient, load LoadChart) (*chart.Chart, error) {
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

		secret, err := kubeClient.Core().Secrets(namespace).Get(details.Auth.Header.SecretKeyRef.Name, metav1.GetOptions{})
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
	chartRequested, err := fetchChart(&netClient, chartURL, authHeader, load)
	if err != nil {
		return nil, err
	}
	return chartRequested, nil
}
