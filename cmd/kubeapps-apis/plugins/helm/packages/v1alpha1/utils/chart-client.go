// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/containerd/containerd/remotes/docker"
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	corev1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"

	k8scorev1 "k8s.io/api/core/v1"
)

// ChartDetails contains the information to retrieve a Chart
type ChartDetails struct {
	// AppRepositoryResourceName specifies an app repository resource to use
	// for the request.
	AppRepositoryResourceName string `json:"appRepositoryResourceName,omitempty"`
	// AppRepositoryResourceNamespace specifies the namespace for the app repository
	AppRepositoryResourceNamespace string `json:"appRepositoryResourceNamespace,omitempty"`
	// resource for the request.
	// ChartName is the name of the chart within the repo.
	ChartName string `json:"chartName"`
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
	GetChart(details *ChartDetails, repoURL string) (*chart.Chart, error)
}

// HelmRepoClient struct contains the clients required to retrieve charts info
type HelmRepoClient struct {
	userAgent string
	netClient *http.Client
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

// Init initialises the HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
func (c *HelmRepoClient) Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	var err error
	c.netClient, err = helm.InitNetClient(appRepo, caCertSecret, authSecret, http.Header{"User-Agent": []string{c.userAgent}})
	return err
}

// GetChart loads a Chart from a given tarball, if the tarball URL is not passed,
// it will try to retrieve the chart by parsing the whole repo index
func (c *HelmRepoClient) GetChart(details *ChartDetails, repoURL string) (*chart.Chart, error) {
	if c.netClient == nil {
		return nil, fmt.Errorf("unable to retrieve chart, Init should be called first")
	}

	if details.TarballURL == "" {
		return nil, fmt.Errorf("calling GetChart '%s - %s' without any tarball url", details.ChartName, details.Version)
	}

	chartURL, err := resolveChartURL(repoURL, details.TarballURL)
	if err != nil {
		return nil, err
	}

	log.Infof("Downloading %s ...", chartURL)
	chart, err := fetchChart(c.netClient, chartURL.String())
	if err != nil {
		return nil, err
	}

	return chart, nil
}

func resolveChartURL(indexURL, chartURL string) (*url.URL, error) {
	parsedIndexURL, err := url.Parse(strings.TrimSpace(indexURL))
	if err != nil {
		return nil, err
	}
	parsedChartURL, err := parsedIndexURL.Parse(strings.TrimSpace(chartURL))
	if err != nil {
		return nil, err
	}
	return parsedChartURL, nil
}

// fetchChart returns the Chart content given an URL
func fetchChart(netClient *http.Client, chartURL string) (*chart.Chart, error) {
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

// Init initialises the HTTP client based on the chart details loading a
// custom CA if provided (as a secret)
// TODO(andresmgot): Using a custom CA cert is not supported by ORAS (neither helm), only using the insecure flag
func (c *OCIRepoClient) Init(appRepo *appRepov1.AppRepository, caCertSecret *corev1.Secret, authSecret *corev1.Secret) error {
	var err error
	headers := http.Header{
		"User-Agent": []string{c.userAgent},
	}
	netClient, err := helm.InitHTTPClient(appRepo, caCertSecret)
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
func (c *OCIRepoClient) GetChart(details *ChartDetails, repoURL string) (*chart.Chart, error) {
	if c.puller == nil {
		return nil, fmt.Errorf("unable to retrieve chart, Init should be called first")
	}
	if details == nil || details.TarballURL == "" {
		return nil, fmt.Errorf("unable to retrieve chart, missing chart details")
	}
	chartURL, err := resolveChartURL(repoURL, details.TarballURL)
	if err != nil {
		return nil, err
	}

	ref := path.Join(chartURL.Host, chartURL.Path)
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
func (c *ChartClientFactory) New(tarballUrl string, userAgent string) ChartClient {
	var client ChartClient
	if strings.HasPrefix(tarballUrl, "oci://") {
		client = NewOCIClient(userAgent)
	} else {
		client = NewChartClient(userAgent)
	}
	return client
}

// GetChart retrieves a chart
func GetChart(chartDetails *ChartDetails, appRepo *appRepov1.AppRepository, caCertSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret, chartClient ChartClient) (*chart.Chart, error) {
	err := chartClient.Init(appRepo, caCertSecret, authSecret)
	if err != nil {
		return nil, err
	}
	ch, err := chartClient.GetChart(chartDetails, appRepo.Spec.URL)
	if err != nil {
		return nil, err
	}
	return ch, nil
}
