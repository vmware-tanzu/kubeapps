// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/disintegration/imaging"
	"github.com/itchyny/gojq"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"helm.sh/helm/v3/pkg/chart"
	helmregistry "helm.sh/helm/v3/pkg/registry"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	additionalCAFile     = "/usr/local/share/ca-certificates/ca.crt"
	numWorkers           = 10
	chartsIndexMediaType = "application/vnd.vmware.charts.index.config.v1+json"
)

type Config struct {
	DatabaseURL              string
	DatabaseName             string
	DatabaseUser             string
	DatabasePassword         string
	Debug                    bool
	Namespace                string
	OciRepositories          []string
	TlsInsecureSkipVerify    bool
	FilterRules              string
	PassCredentials          bool
	UserAgent                string
	UserAgentComment         string
	GlobalPackagingNamespace string
	KubeappsNamespace        string
	AuthorizationHeader      string
	DockerConfigJson         string
}

type importChartFilesJob struct {
	ID           string
	Repo         *models.Repo
	ChartVersion models.ChartVersion
}

type pullChartJob struct {
	AppName string
	Tag     string
}

type pullChartResult struct {
	Chart *models.Chart
	Error error
}

type checkTagJob struct {
	AppName string
	Tag     string
}

type checkTagResult struct {
	checkTagJob
	isHelmChart bool
	Error       error
}

func parseRepoURL(repoURL string) (*url.URL, error) {
	repoURL = strings.TrimSpace(repoURL)
	return url.ParseRequestURI(repoURL)
}

type assetManager interface {
	Delete(repo models.Repo) error
	Sync(repo models.Repo, charts []models.Chart) error
	LastChecksum(repo models.Repo) string
	UpdateLastCheck(repoNamespace, repoName, checksum string, now time.Time) error
	Init() error
	Close() error
	InvalidateCache() error
	updateIcon(repo models.Repo, data []byte, contentType, ID string) error
	filesExist(repo models.Repo, chartFilesID, digest string) bool
	insertFiles(chartID string, files models.ChartFiles) error
}

func newManager(config dbutils.Config, globalPackagingNamespace string) (assetManager, error) {
	return newPGManager(config, globalPackagingNamespace)
}

func getSha256(src []byte) (string, error) {
	f := bytes.NewReader(src)
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// Repo defines the methods to retrieve information from the given repository
type Repo interface {
	Checksum() (string, error)
	Repo() *models.RepoInternal
	FilterIndex()
	Charts(fetchLatestOnly bool) ([]models.Chart, error)
	FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error)
}

// HelmRepo implements the Repo interface for chartmuseum-like repositories
type HelmRepo struct {
	content []byte
	*models.RepoInternal
	netClient httpclient.Client
	filter    *apprepov1alpha1.FilterRuleSpec
}

// Checksum returns the sha256 of the repo
func (r *HelmRepo) Checksum() (string, error) {
	return getSha256(r.content)
}

// Repo returns the repo information
func (r *HelmRepo) Repo() *models.RepoInternal {
	return r.RepoInternal
}

// FilterRepo is a no-op for a Helm repo
func (r *HelmRepo) FilterIndex() {
	// no-op
}

func compileJQ(rule *apprepov1alpha1.FilterRuleSpec) (*gojq.Code, []interface{}, error) {
	query, err := gojq.Parse(rule.JQ)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse jq query: %v", err)
	}
	varNames := []string{}
	varValues := []interface{}{}
	for name, val := range rule.Variables {
		varNames = append(varNames, name)
		varValues = append(varValues, val)
	}
	code, err := gojq.Compile(
		query,
		gojq.WithVariables(varNames),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to compile jq: %v", err)
	}
	return code, varValues, nil
}

func satisfy(chartInput map[string]interface{}, code *gojq.Code, vars []interface{}) (bool, error) {
	res, _ := code.Run(chartInput, vars...).Next()
	if err, ok := res.(error); ok {
		return false, fmt.Errorf("unable to run jq: %v", err)
	}

	satisfied, ok := res.(bool)
	if !ok {
		return false, fmt.Errorf("unable to convert jq result to boolean. Got: %v", res)
	}
	return satisfied, nil
}

// Make sure charts are treated without escaped data
func unescapeChartsData(charts []models.Chart) []models.Chart {
	result := []models.Chart{}
	for _, chart := range charts {
		chart.Name = unescapeOrDefaultValue(chart.Name)
		chart.ID = unescapeOrDefaultValue(chart.ID)
		result = append(result, chart)
	}
	return result
}

// Unescape string or return value itself if error
func unescapeOrDefaultValue(value string) string {
	// Ensure any escaped `/` (%2F) in a chart name will remain escaped.
	// Kubeapps splits the chart ID, such as "repo-name/harbor-project%2Fchart-name", on the slash.
	// See PR comment at
	// https://github.com/vmware-tanzu/kubeapps/pull/3863#pullrequestreview-819141298
	// and instance of the issue cropping up via Harbor at
	// https://github.com/vmware-tanzu/kubeapps/issues/5897
	value = strings.Replace(value, "%2F", "%252F", -1)
	unescapedValue, err := url.PathUnescape(value)
	if err != nil {
		return value
	} else {
		return unescapedValue
	}
}

func filterCharts(charts []models.Chart, filterRule *apprepov1alpha1.FilterRuleSpec) ([]models.Chart, error) {
	if filterRule == nil || filterRule.JQ == "" {
		// No filter
		return charts, nil
	}
	jqCode, vars, err := compileJQ(filterRule)
	if err != nil {
		return nil, err
	}
	result := []models.Chart{}
	for _, chart := range charts {
		// Convert the chart to a map[interface]{}
		chartBytes, err := json.Marshal(chart)
		if err != nil {
			return nil, fmt.Errorf("unable to parse chart: %v", err)
		}
		chartInput := map[string]interface{}{}
		err = json.Unmarshal(chartBytes, &chartInput)
		if err != nil {
			return nil, fmt.Errorf("unable to parse chart: %v", err)
		}

		satisfied, err := satisfy(chartInput, jqCode, vars)
		if err != nil {
			return nil, err
		}
		if satisfied {
			// All rules have been checked and matched
			result = append(result, chart)
		}
	}
	return result, nil
}

// Charts retrieve the list of charts exposed in the repo
func (r *HelmRepo) Charts(fetchLatestOnly bool) ([]models.Chart, error) {
	repo := &models.Repo{
		Namespace: r.Namespace,
		Name:      r.Name,
		URL:       r.URL,
		Type:      r.Type,
	}
	charts, err := helm.ChartsFromIndex(r.content, repo, fetchLatestOnly)
	if err != nil {
		return []models.Chart{}, err
	}
	if len(charts) == 0 {
		return []models.Chart{}, nil
	}

	return filterCharts(unescapeChartsData(charts), r.filter)
}

// FetchFiles retrieves the important files of a chart and version from the repo
func (r *HelmRepo) FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error) {
	authorizationHeader := ""
	chartTarballURL := chartTarballURL(r.RepoInternal, cv)

	if passCredentials || len(r.AuthorizationHeader) > 0 && isURLDomainEqual(chartTarballURL, r.URL) {
		authorizationHeader = r.AuthorizationHeader
	}

	return tarutil.FetchChartDetailFromTarballUrl(
		chartTarballURL,
		userAgent,
		authorizationHeader,
		r.netClient)
}

// TagList represents a list of tags as specified at
// https://github.com/opencontainers/distribution-spec/blob/main/spec.md#content-discovery
type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// OCIRegistry implements the Repo interface for OCI repositories
type OCIRegistry struct {
	repositories []string
	*models.RepoInternal
	tags   map[string]TagList
	puller helm.ChartPuller
	ociCli ociAPI
	filter *apprepov1alpha1.FilterRuleSpec
}

func doReq(url string, cli httpclient.Client, headers map[string]string, userAgent string) ([]byte, error) {
	headers["User-Agent"] = userAgent
	return httpclient.Get(url, cli, headers)
}

// OCILayer represents a single OCI layer
type OCILayer struct {
	MediaType string `json:"mediaType"`
	Digest    string `json:"digest"`
	Size      int    `json:"size"`
}

// OCIManifest representation
type OCIManifest struct {
	Schema int        `json:"schema"`
	Config OCILayer   `json:"config"`
	Layers []OCILayer `json:"layers"`
}

// VACCatalog representation
type VACCatalog struct {
	APIVersion string         `json:"apiVersion"`
	Entries    map[string]any `json:"entries"`
}

type ociAPI interface {
	TagList(appName, userAgent string) (*TagList, error)
	IsHelmChart(appName, tag, userAgent string) (bool, error)
	CatalogAvailable(userAgent string) bool
	Catalog(userAgent string) ([]string, error)
}

// OciAPIClient enables basic interactions with an OCI registry.
//
// The AuthHeader is optional - if empty the http client's default Authorization
// header will be used.
type OciAPIClient struct {
	AuthHeader string
	Url        *url.URL
	NetClient  httpclient.Client
}

// TagList retrieves the list of tags for an asset
func (o *OciAPIClient) TagList(appName string, userAgent string) (*TagList, error) {
	url := *o.Url
	url.Path = path.Join("v2", url.Path, appName, "tags", "list")
	headers := map[string]string{}
	if o.AuthHeader != "" {
		headers["Authorization"] = o.AuthHeader
	}
	data, err := doReq(url.String(), o.NetClient, headers, userAgent)
	if err != nil {
		return nil, err
	}

	var appTags TagList
	err = json.Unmarshal(data, &appTags)
	if err != nil {
		return nil, err
	}
	return &appTags, nil
}

func (o *OciAPIClient) IsHelmChart(appName, tag, userAgent string) (bool, error) {
	repoURL := *o.Url
	repoURL.Path = path.Join("v2", repoURL.Path, appName, "manifests", tag)
	log.V(4).Infof("Getting tag %s", repoURL.String())
	headers := map[string]string{
		"Accept": "application/vnd.oci.image.manifest.v1+json",
	}
	if o.AuthHeader != "" {
		headers["Authorization"] = o.AuthHeader
	}
	manifestData, err := doReq(repoURL.String(), o.NetClient, headers, userAgent)
	if err != nil {
		return false, err
	}
	var manifest OCIManifest
	err = json.Unmarshal(manifestData, &manifest)
	if err != nil {
		return false, err
	}
	return manifest.Config.MediaType == helmregistry.ConfigMediaType, nil
}

// CatalogAvailable returns whether Kubeapps can return a catalog for
// this OCI repository.
//
// Currently this checks only for a VMware Application Catalog index as
// documented at:
// https://docs.vmware.com/en/VMware-Application-Catalog/services/main/GUID-using-consume-metadata.html#method-2-obtain-metadata-from-the-oci-registry-10
// although examples have "chart-index" rather than "index" as the artifact
// name.
// In the future, this should check the oci-catalog service for possible
// catalogs.
func (o *OciAPIClient) CatalogAvailable(userAgent string) bool {
	manifest, err := o.catalogManifest(userAgent)
	if err != nil {
		log.Errorf("Unable to get catalog manifest: %+v", err)
		return false
	}
	return manifest.Config.MediaType == chartsIndexMediaType
}

func (o *OciAPIClient) catalogManifest(userAgent string) (*OCIManifest, error) {
	indexURL := *o.Url
	indexURL.Path = path.Join("v2", indexURL.Path, "charts-index", "manifests", "latest")
	log.V(4).Infof("Getting tag %s", indexURL.String())
	headers := map[string]string{
		"Accept": "application/vnd.oci.image.manifest.v1+json",
	}
	if o.AuthHeader != "" {
		headers["Authorization"] = o.AuthHeader
	}
	manifestData, err := doReq(indexURL.String(), o.NetClient, headers, userAgent)
	if err != nil {
		return nil, err
	}
	var manifest OCIManifest
	err = json.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, err
	}
	return &manifest, nil
}

// Catalog returns the list of repositories in the (namespaced) registry
// when discoverable.
func (o *OciAPIClient) Catalog(userAgent string) ([]string, error) {
	// TODO(minelson): all Kubeapps interactions with OCI registries should
	// be updated to use the oras go lib.
	manifest, err := o.catalogManifest(userAgent)
	if err != nil {
		return nil, err
	}

	if len(manifest.Layers) != 1 || manifest.Layers[0].MediaType != "application/vnd.vmware.charts.index.layer.v1+json" {
		log.Errorf("Unexpected layer in index manifest: %v", manifest)
		return nil, fmt.Errorf("unexpected layer in index manifest")
	}

	blobDigest := manifest.Layers[0].Digest

	blobURL := *o.Url
	blobURL.Path = path.Join("v2", blobURL.Path, "charts-index", "blobs", blobDigest)
	log.V(4).Infof("Getting blob %s", blobURL.String())
	headers := map[string]string{
		"Accept": "application/vnd.vmware.charts.index.layer.v1+json",
	}
	if o.AuthHeader != "" {
		headers["Authorization"] = o.AuthHeader
	}
	blobData, err := doReq(blobURL.String(), o.NetClient, headers, userAgent)
	if err != nil {
		return nil, err
	}
	var vacCatalog VACCatalog
	err = json.Unmarshal(blobData, &vacCatalog)
	if err != nil {
		return nil, err
	}

	repos := make([]string, 0, len(vacCatalog.Entries))
	for r := range vacCatalog.Entries {
		repos = append(repos, r)
	}
	sort.Strings(repos)
	return repos, nil
}

func tagCheckerWorker(o ociAPI, tagJobs <-chan checkTagJob, resultChan chan checkTagResult) {
	for j := range tagJobs {
		isHelmChart, err := o.IsHelmChart(j.AppName, j.Tag, GetUserAgent("", ""))
		resultChan <- checkTagResult{j, isHelmChart, err}
	}
}

// Checksum returns the sha256 of the repo by concatenating tags for
// all repositories within the registry and returning the sha256.
// Caveat: Mutated image tags won't be detected as new
func (r *OCIRegistry) Checksum() (string, error) {
	r.tags = map[string]TagList{}

	if len(r.repositories) == 0 {
		repos, err := r.ociCli.Catalog("")
		if err != nil {
			return "", err
		}
		r.repositories = repos
	}

	for _, appName := range r.repositories {
		tags, err := r.ociCli.TagList(appName, GetUserAgent("", ""))
		if err != nil {
			return "", err
		}
		r.tags[appName] = *tags
	}

	content, err := json.Marshal(r.tags)
	if err != nil {
		return "", err
	}

	return getSha256(content)
}

// Repo returns the repo information
func (r *OCIRegistry) Repo() *models.RepoInternal {
	return r.RepoInternal
}

func pullAndExtract(repoURL *url.URL, appName, tag string, puller helm.ChartPuller, r *OCIRegistry) (*models.Chart, error) {
	ref := path.Join(repoURL.Host, repoURL.Path, fmt.Sprintf("%s:%s", appName, tag))

	chartBuffer, digest, err := puller.PullOCIChart(ref)
	if err != nil {
		return nil, err
	}

	files, err := tarutil.FetchChartDetailFromTarball(chartBuffer)
	if err != nil {
		return nil, err
	}
	chartMetadata := chart.Metadata{}
	err = yaml.Unmarshal([]byte(files[models.ChartYamlKey]), &chartMetadata)
	if err != nil {
		return nil, err
	}

	// Format Data
	chartVersion := models.ChartVersion{
		Version:                 chartMetadata.Version,
		AppVersion:              chartMetadata.AppVersion,
		Digest:                  digest,
		URLs:                    chartMetadata.Sources,
		Readme:                  files[models.ReadmeKey],
		DefaultValues:           files[models.DefaultValuesKey],
		AdditionalDefaultValues: additional_default_values_from_files(files),
		Schema:                  files[models.SchemaKey],
	}

	maintainers := []chart.Maintainer{}
	for _, m := range chartMetadata.Maintainers {
		maintainers = append(maintainers, chart.Maintainer{
			Name:  m.Name,
			Email: m.Email,
			URL:   m.URL,
		})
	}

	// Encode repository names to store them in the database.
	encodedAppNameForID := url.PathEscape(appName)

	return &models.Chart{
		ID:            path.Join(r.Name, encodedAppNameForID),
		Name:          chartMetadata.Name,
		Repo:          &models.Repo{Namespace: r.Namespace, Name: r.Name, URL: r.URL, Type: r.Type},
		Description:   chartMetadata.Description,
		Home:          chartMetadata.Home,
		Keywords:      chartMetadata.Keywords,
		Maintainers:   maintainers,
		Sources:       chartMetadata.Sources,
		Icon:          chartMetadata.Icon,
		Category:      chartMetadata.Annotations["category"],
		ChartVersions: []models.ChartVersion{chartVersion},
	}, nil
}

func chartImportWorker(repoURL *url.URL, r *OCIRegistry, chartJobs <-chan pullChartJob, resultChan chan pullChartResult) {
	for j := range chartJobs {
		log.V(4).Infof("Pulling chart, name=%s, tag=%s", j.AppName, j.Tag)
		chart, err := pullAndExtract(repoURL, j.AppName, j.Tag, r.puller, r)
		resultChan <- pullChartResult{chart, err}
	}
}

// FilterIndex remove non chart tags
func (r *OCIRegistry) FilterIndex() {
	unfilteredTags := r.tags
	r.tags = map[string]TagList{}
	checktagJobs := make(chan checkTagJob, numWorkers)
	tagcheckRes := make(chan checkTagResult, numWorkers)
	var wg sync.WaitGroup

	// Process 10 tags at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			tagCheckerWorker(r.ociCli, checktagJobs, tagcheckRes)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(tagcheckRes)
	}()

	go func() {
		for _, appName := range r.repositories {
			for _, tag := range unfilteredTags[appName].Tags {
				checktagJobs <- checkTagJob{AppName: appName, Tag: tag}
			}
		}
		close(checktagJobs)
	}()

	// Start receiving tags
	for res := range tagcheckRes {
		if res.Error == nil {
			if res.isHelmChart {
				r.tags[res.AppName] = TagList{
					Name: unfilteredTags[res.AppName].Name,
					Tags: append(r.tags[res.AppName].Tags, res.Tag),
				}
			}
		} else {
			log.Errorf("Failed to pull chart. Got %v", res.Error)
		}
	}

	// Order tags by semver
	for _, appName := range r.repositories {
		vs := make([]*semver.Version, len(r.tags[appName].Tags))
		for i, r := range r.tags[appName].Tags {
			v, err := semver.NewVersion(r)
			if err != nil {
				log.Errorf("Error parsing version: %s", err)
			}
			vs[i] = v
		}
		sort.Sort(sort.Reverse(semver.Collection(vs)))
		orderedTags := []string{}
		for _, v := range vs {
			orderedTags = append(orderedTags, v.String())
		}
		r.tags[appName] = TagList{
			Name: r.tags[appName].Name,
			Tags: orderedTags,
		}
	}
}

// Charts retrieve the list of charts exposed in the repo
func (r *OCIRegistry) Charts(fetchLatestOnly bool) ([]models.Chart, error) {
	result := map[string]*models.Chart{}
	repoURL, err := parseRepoURL(r.RepoInternal.URL)
	if err != nil {
		return nil, err
	}

	if len(r.repositories) == 0 {
		repos, err := r.ociCli.Catalog("")
		if err != nil {
			return nil, err
		}
		r.repositories = repos
	}

	chartJobs := make(chan pullChartJob, numWorkers)
	chartResults := make(chan pullChartResult, numWorkers)
	var wg sync.WaitGroup
	// Process 10 charts at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			chartImportWorker(repoURL, r, chartJobs, chartResults)
			wg.Done()
		}()
	}
	// When we know all workers have sent their data in chartChan, close it.
	go func() {
		wg.Wait()
		close(chartResults)
	}()

	log.V(4).Infof("Starting %d workers", numWorkers)
	go func() {
		for _, appName := range r.repositories {
			if fetchLatestOnly {
				chartJobs <- pullChartJob{AppName: appName, Tag: r.tags[appName].Tags[0]}
			} else {
				for _, tag := range r.tags[appName].Tags {
					chartJobs <- pullChartJob{AppName: appName, Tag: tag}
				}
			}
		}
		close(chartJobs)
	}()

	// Start receiving charts
	for res := range chartResults {
		if res.Error == nil {
			ch := res.Chart
			log.V(4).Infof("Received chart %s from channel", ch.ID)
			if r, ok := result[ch.ID]; ok {
				// Chart already exists, append version
				r.ChartVersions = append(r.ChartVersions, ch.ChartVersions...)
			} else {
				result[ch.ID] = ch
			}
		} else {
			log.Errorf("Failed to pull chart. Got %v", res.Error)
		}
	}

	charts := []models.Chart{}
	for _, c := range result {
		charts = append(charts, *c)
	}
	return filterCharts(charts, r.filter)
}

// FetchFiles do nothing for the OCI case since they have been already fetched in the Charts() method
func (r *OCIRegistry) FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error) {
	return map[string]string{
		models.DefaultValuesKey: cv.DefaultValues,
		models.ReadmeKey:        cv.Readme,
		models.SchemaKey:        cv.Schema,
	}, nil
}

func parseFilters(filters string) (*apprepov1alpha1.FilterRuleSpec, error) {
	filterSpec := &apprepov1alpha1.FilterRuleSpec{}
	if len(filters) > 0 {
		err := json.Unmarshal([]byte(filters), filterSpec)
		if err != nil {
			return nil, err
		}
	}
	return filterSpec, nil
}

func getHelmRepo(namespace, name, repoURL, authorizationHeader string, filter *apprepov1alpha1.FilterRuleSpec, netClient httpclient.Client, userAgent string) (Repo, error) {
	url, err := parseRepoURL(repoURL)
	if err != nil {
		log.Errorf("Failed to parse URL, url=%s: %v", repoURL, err)
		return nil, err
	}

	repoBytes, err := fetchRepoIndex(url.String(), authorizationHeader, netClient, userAgent)
	if err != nil {
		return nil, err
	}

	return &HelmRepo{
		content: repoBytes,
		RepoInternal: &models.RepoInternal{
			Namespace:           namespace,
			Name:                name,
			URL:                 url.String(),
			AuthorizationHeader: authorizationHeader,
		},
		netClient: netClient,
		filter:    filter,
	}, nil
}

func getOCIRepo(namespace, name, repoURL, authorizationHeader string, filter *apprepov1alpha1.FilterRuleSpec, ociRepos []string, netClient *http.Client) (Repo, error) {
	url, err := parseRepoURL(repoURL)
	if err != nil {
		log.Errorf("Failed to parse URL, url=%s: %v", repoURL, err)
		return nil, err
	}

	// If the AppRepo has the URL specified as `oci://` then replace it with
	// https for talking with the API. If people are using non-https OCI
	// registries (?!) then they can specify the URL with http.
	if url.Scheme == "oci" {
		url.Scheme = "https"
	}
	headers := http.Header{}
	if authorizationHeader != "" {
		headers["Authorization"] = []string{authorizationHeader}
	}
	ociResolver := docker.NewResolver(docker.ResolverOptions{Headers: headers, Client: netClient})

	return &OCIRegistry{
		repositories: ociRepos,
		RepoInternal: &models.RepoInternal{Namespace: namespace, Name: name, URL: url.String(), AuthorizationHeader: authorizationHeader},
		puller:       &helm.OCIPuller{Resolver: ociResolver},
		ociCli:       &OciAPIClient{AuthHeader: authorizationHeader, Url: url, NetClient: netClient},
		filter:       filter,
	}, nil
}

func fetchRepoIndex(url, authHeader string, cli httpclient.Client, userAgent string) ([]byte, error) {
	indexURL, err := parseRepoURL(url)
	if err != nil {
		log.Errorf("Failed to parse URL, url=%s: %v", url, err)
		return nil, err
	}
	indexURL.Path = path.Join(indexURL.Path, "index.yaml")
	headers := map[string]string{}
	if authHeader != "" {
		headers["Authorization"] = authHeader
	}
	return doReq(indexURL.String(), cli, headers, userAgent)
}

func chartTarballURL(r *models.RepoInternal, cv models.ChartVersion) string {
	source := cv.URLs[0]
	if _, err := parseRepoURL(source); err != nil {
		// If the chart URL is not absolute, join with repo URL. It's fine if the
		// URL we build here is invalid as we can catch this error when actually
		// making the request
		u, _ := url.Parse(r.URL)
		u.Path = path.Join(u.Path, source)
		return u.String()
	}
	return source
}

type fileImporter struct {
	manager   assetManager
	netClient httpclient.Client
}

func (f *fileImporter) fetchFiles(charts []models.Chart, repo Repo, userAgent string, passCredentials bool) {
	iconJobs := make(chan models.Chart, numWorkers)
	chartFilesJobs := make(chan importChartFilesJob, numWorkers)
	var wg sync.WaitGroup

	log.V(4).Infof("Starting %d workers", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go f.importWorker(&wg, iconJobs, chartFilesJobs, repo, userAgent, passCredentials)
	}

	// Enqueue jobs to process chart icons
	for _, c := range charts {
		iconJobs <- c
	}
	// Close the iconJobs channel to signal the worker pools to move on to the
	// chart files jobs
	close(iconJobs)

	// Iterate through the list of charts and enqueue the latest chart version to
	// be processed. Append the rest of the chart versions to a list to be
	// enqueued later
	var toEnqueue []importChartFilesJob
	for _, c := range charts {
		// TODO: Should we use the chart id, chart name with prefix, or helm chart name here? The database actually stores the chart ID so we could
		// pass that in instead? Why don't we?
		chartFilesJobs <- importChartFilesJob{c.ID, c.Repo, c.ChartVersions[0]}
		for _, cv := range c.ChartVersions[1:] {
			toEnqueue = append(toEnqueue, importChartFilesJob{c.ID, c.Repo, cv})
		}
	}

	// Enqueue all the remaining chart versions
	for _, cfj := range toEnqueue {
		chartFilesJobs <- cfj
	}
	// Close the chartFilesJobs channel to signal the worker pools that there are
	// no more jobs to process
	close(chartFilesJobs)

	// Wait for the worker pools to finish processing
	wg.Wait()
}

func (f *fileImporter) importWorker(wg *sync.WaitGroup, icons <-chan models.Chart, chartFiles <-chan importChartFilesJob, repo Repo, userAgent string, passCredentials bool) {
	defer wg.Done()
	for c := range icons {
		log.V(4).Infof("Importing icon, name=%s", c.Name)
		if err := f.fetchAndImportIcon(c, repo.Repo(), userAgent, passCredentials); err != nil {
			log.Errorf("Failed to import icon, name=%s: %v", c.Name, err)
		}
	}
	for j := range chartFiles {
		log.V(4).Infof("Importing readme and values, ID=%s, version=%s", j.ID, j.ChartVersion.Version)
		if err := f.fetchAndImportFiles(j.ID, repo, j.ChartVersion, userAgent, passCredentials); err != nil {
			log.Errorf("Failed to import files, ID=%s, version=%s: %v", j.ID, j.ChartVersion.Version, err)
		}
	}
}

func (f *fileImporter) fetchAndImportIcon(c models.Chart, r *models.RepoInternal, userAgent string, passCredentials bool) error {
	if c.Icon == "" {
		log.Infof("Icon not found, name=%s", c.Name)
		return nil
	}

	reqHeaders := make(map[string]string)
	reqHeaders["User-Agent"] = userAgent
	if passCredentials || len(r.AuthorizationHeader) > 0 && isURLDomainEqual(c.Icon, r.URL) {
		reqHeaders["Authorization"] = r.AuthorizationHeader
	}

	reader, contentType, err := httpclient.GetStream(c.Icon, f.netClient, reqHeaders)
	if reader != nil {
		defer reader.Close()
	}
	if err != nil {
		return err
	}

	var img image.Image
	// if the icon is in any other format try to convert it to PNG
	if strings.Contains(contentType, "image/svg") {
		// if the icon is an SVG, it requires special processing
		icon, err := oksvg.ReadIconStream(reader)
		if err != nil {
			log.Errorf("Failed to decode icon, name=%s: %v", c.Name, err)
			return err
		}
		w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
		icon.SetTarget(0, 0, float64(w), float64(h))
		rgba := image.NewNRGBA(image.Rect(0, 0, w, h))
		icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)
		img = rgba
	} else {
		img, err = imaging.Decode(reader)
		if err != nil {
			log.Errorf("Failed to decode icon, name=%s: %v", c.Name, err)
			return err
		}
	}

	// TODO: make this configurable?
	resizedImg := imaging.Fit(img, 160, 160, imaging.Lanczos)
	var buf bytes.Buffer
	err = imaging.Encode(&buf, resizedImg, imaging.PNG)
	if err != nil {
		log.Errorf("Failed to encode icon, name=%s: %v", c.Name, err)
		return err
	}
	b := buf.Bytes()
	contentType = "image/png"

	return f.manager.updateIcon(models.Repo{Namespace: r.Namespace, Name: r.Name}, b, contentType, c.ID)
}

func (f *fileImporter) fetchAndImportFiles(chartID string, repo Repo, cv models.ChartVersion, userAgent string, passCredentials bool) error {
	r := repo.Repo()
	chartFilesID := fmt.Sprintf("%s-%s", chartID, cv.Version)

	// Check if we already have indexed files for this chart version and digest
	if f.manager.filesExist(models.Repo{Namespace: r.Namespace, Name: r.Name}, chartFilesID, cv.Digest) {
		log.V(4).Infof("Skipping existing files, id: %s, version: %s", chartID, cv.Version)
		return nil
	}
	log.V(4).Infof("Fetching files, id=%s, version=%s", chartID, cv.Version)

	files, err := repo.FetchFiles(cv, userAgent, passCredentials)
	if err != nil {
		return err
	}

	chartFiles := models.ChartFiles{ID: chartFilesID, Repo: &models.Repo{Name: r.Name, Namespace: r.Namespace, URL: r.URL}, Digest: cv.Digest}
	if v, ok := files[models.ReadmeKey]; ok {
		chartFiles.Readme = v
	} else {
		log.Infof("The README.md file has not been found, id=%s, version=%s", chartID, cv.Version)
	}
	if v, ok := files[models.DefaultValuesKey]; ok {
		chartFiles.DefaultValues = v
	} else {
		log.Infof("The values.yaml file has not been found, id=%s, version=%s", chartID, cv.Version)
	}
	if v, ok := files[models.SchemaKey]; ok {
		chartFiles.Schema = v
	} else {
		log.Infof("The values.schema.json file has not been found, id=%s, version=%s", chartID, cv.Version)
	}
	chartFiles.AdditionalDefaultValues = additional_default_values_from_files(files)

	// inserts the chart files if not already indexed, or updates the existing
	// entry if digest has changed
	return f.manager.insertFiles(chartID, chartFiles)
}

// Check if two URL strings are in the same domain.
// Return true if so, and false otherwise or when an error occurs
func isURLDomainEqual(url1Str, url2Str string) bool {
	url1, err := url.ParseRequestURI(url1Str)
	if err != nil {
		return false
	}
	url2, err := url.ParseRequestURI(url2Str)
	if err != nil {
		return false
	}

	return url1.Scheme == url2.Scheme && url1.Host == url2.Host
}

// Returns the user agent to be used during calls to the chart repositories
// Examples:
// asset-syncer/devel
// asset-syncer/1.0
// asset-syncer/1.0 (foo v1.0-beta4)
// More info here https://github.com/vmware-tanzu/kubeapps/issues/767#issuecomment-436835938
func GetUserAgent(version, userAgentComment string) string {
	if version == "" && userAgentComment == "" {
		return "asset-syncer/devel"
	}
	ua := "asset-syncer/" + version
	if userAgentComment != "" {
		ua = fmt.Sprintf("%s (%s)", ua, userAgentComment)
	}
	return ua
}
func additional_default_values_from_files(files map[string]string) map[string]string {
	additional_filenames := []string{}
	for f := range files {
		if strings.HasPrefix(f, models.DefaultValuesKey+"-") {
			additional_filenames = append(additional_filenames, f)
		}
	}

	additional_defaults := map[string]string{}
	for _, f := range additional_filenames {
		additional_defaults[f] = files[f]
	}
	return additional_defaults
}
