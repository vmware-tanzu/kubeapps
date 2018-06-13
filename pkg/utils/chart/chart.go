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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

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

// FetchRepoIndex returns a Helm repository
func FetchRepoIndex(netClient *HTTPClient, repoURL string, authHeader string) (*repo.IndexFile, error) {
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

// FindChartInRepoIndex returns the URL of a chart given a Helm repository and its name and version
func FindChartInRepoIndex(repoIndex *repo.IndexFile, repoURL, chartName, chartVersion string) (string, error) {
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

// FetchChart returns the Chart content given an URL and the auth header if needed
func FetchChart(netClient *HTTPClient, chartURL, authHeader string, load LoadChart) (*chart.Chart, error) {
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
