// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"net/url"
	"path"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/disintegration/imaging"
	"github.com/itchyny/gojq"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalog_client"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"helm.sh/helm/v3/pkg/chart"
	helmregistry "helm.sh/helm/v3/pkg/registry"
	log "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/slice"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"sigs.k8s.io/yaml"
)

const (
	additionalCAFile = "/usr/local/share/ca-certificates/ca.crt"
	// numWorkersFiles is the number of workers used when pulling non-OCI charts
	// to extract files (Readme, values, etc.), as well as chart icons
	// generally.
	numWorkersFiles = 10

	// numWorkersOCI is the number of workers used when pulling charts from OCI
	// registries. This may need to be adjusted depending on public
	// registry limits until we can better manage request limits for
	// the Bitnami OCI repo on dockerhub.
	numWorkersOCI            = 10
	maxOCIVersionsForOneSync = 5
	chartsIndexMediaType     = "application/vnd.vmware.charts.index.config.v1+json"
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
	OCICatalogURL            string
}

type importChartFilesJob struct {
	ID           string
	Repo         *models.AppRepository
	ChartVersion models.ChartVersion
}

type pullChartJob struct {
	AppName        string
	VersionsToSync []string
	Chart          *models.Chart
}

type pullChartResult struct {
	Chart  models.Chart
	Errors []error
}

func parseRepoURL(repoURL string) (*url.URL, error) {
	repoURL = strings.TrimSpace(repoURL)
	return url.ParseRequestURI(repoURL)
}

type assetManager interface {
	Delete(repo models.AppRepository) error
	Sync(repo models.AppRepository, chart models.Chart) error
	LastChecksum(repo models.AppRepository) string
	UpdateLastCheck(repoNamespace, repoName, checksum string, now time.Time) error
	ChartsForRepo(repo models.AppRepository) (map[string]*models.Chart, error)
	Init() error
	Close() error
	InvalidateCache() error
	RemoveMissingCharts(repo models.AppRepository, chartNames []string) error
	updateIcon(repo models.AppRepository, data []byte, contentType, ID string) error
	filesExist(repo models.AppRepository, chartFilesID, digest string) bool
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

// ChartCatalog defines the methods to retrieve information from the given repository
type ChartCatalog interface {
	Checksum(ctx context.Context) (string, error)
	AppRepository() *models.AppRepositoryInternal
	Charts(ctx context.Context, fetchLatestOnly bool, charts chan pullChartResult) ([]string, error)
	FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error)
	Filters() *apprepov1alpha1.FilterRuleSpec
}

// HelmRepo implements the ChartCatalog interface for chartmuseum-like repositories
type HelmRepo struct {
	content []byte
	*models.AppRepositoryInternal
	netClient *http.Client
	filter    *apprepov1alpha1.FilterRuleSpec
	manager   assetManager
}

// Checksum returns the sha256 of the repo
func (r *HelmRepo) Checksum(ctx context.Context) (string, error) {
	return getSha256(r.content)
}

// AppRepository returns the repo information
func (r *HelmRepo) AppRepository() *models.AppRepositoryInternal {
	return r.AppRepositoryInternal
}

func (r *HelmRepo) Filters() *apprepov1alpha1.FilterRuleSpec {
	return r.filter
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
func unescapeChartData(chart models.Chart) models.Chart {
	chart.Name = unescapeOrDefaultValue(chart.Name)
	chart.ID = unescapeOrDefaultValue(chart.ID)
	return chart
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

func filterMatches(chart models.Chart, filterRule *apprepov1alpha1.FilterRuleSpec) (bool, error) {
	if filterRule == nil || filterRule.JQ == "" {
		// No filter
		return true, nil
	}
	jqCode, vars, err := compileJQ(filterRule)
	if err != nil {
		return false, err
	}
	// Convert the chart to a map[interface]{}
	chartBytes, err := json.Marshal(chart)
	if err != nil {
		return false, fmt.Errorf("unable to parse chart: %v", err)
	}
	chartInput := map[string]interface{}{}
	err = json.Unmarshal(chartBytes, &chartInput)
	if err != nil {
		return false, fmt.Errorf("unable to parse chart: %v", err)
	}

	satisfied, err := satisfy(chartInput, jqCode, vars)
	if err != nil {
		return false, err
	}
	return satisfied, nil
}

// Charts retrieve the list of charts exposed in the repo
func (r *HelmRepo) Charts(ctx context.Context, fetchLatestOnly bool, chartResults chan pullChartResult) ([]string, error) {
	repo := &models.AppRepository{
		Namespace: r.Namespace,
		Name:      r.Name,
		URL:       r.URL,
		Type:      r.Type,
	}
	// ChartsFromIndex currently gets all charts quick quickly (as it is just
	// parsing the index).
	charts, err := helm.ChartsFromIndex(r.content, repo, fetchLatestOnly)
	if err != nil {
		return nil, err
	}
	if len(charts) == 0 {
		close(chartResults)
		return nil, nil
	}

	unescapedCharts := []models.Chart{}
	newChartNames := []string{}
	for _, c := range charts {
		unescapedChart := unescapeChartData(c)
		newChartNames = append(newChartNames, unescapedChart.Name)
		unescapedCharts = append(unescapedCharts, unescapedChart)
	}

	go func() {
		for _, chart := range unescapedCharts {
			chartResults <- pullChartResult{
				Chart: chart,
			}
		}
		close(chartResults)
	}()

	syncedChartsForRepo, err := r.manager.ChartsForRepo(*repo)
	if err != nil {
		return nil, err
	}
	chartsForDeletion := []string{}
	for syncedChartName := range syncedChartsForRepo {
		if !slice.ContainsString(newChartNames, syncedChartName, func(s string) string { return s }) {
			chartsForDeletion = append(chartsForDeletion, syncedChartName)
		}
	}
	return chartsForDeletion, nil
}

// FetchFiles retrieves the important files of a chart and version from the repo
func (r *HelmRepo) FetchFiles(cv models.ChartVersion, userAgent string, passCredentials bool) (map[string]string, error) {
	authorizationHeader := ""
	chartTarballURL := chartTarballURL(r.AppRepositoryInternal, cv)

	if passCredentials || len(r.AuthorizationHeader) > 0 && isURLDomainEqual(chartTarballURL, r.URL) {
		authorizationHeader = r.AuthorizationHeader
	}

	// If URL points to an OCI chart, we transform its URL to its tgz blob URL
	if strings.HasPrefix(chartTarballURL, "oci://") {
		return FetchChartDetailFromOciUrl(chartTarballURL, userAgent, authorizationHeader, r.netClient)
	} else {
		return tarutil.FetchChartDetailFromTarballUrl(chartTarballURL, userAgent, authorizationHeader, r.netClient)
	}
}

// Fetches helm chart details from an OCI url
func FetchChartDetailFromOciUrl(chartTarballURL string, userAgent string, authz string, netClient *http.Client) (map[string]string, error) {
	headers := http.Header{}
	if len(userAgent) > 0 {
		headers.Add("User-Agent", userAgent)
	}
	if len(authz) > 0 {
		headers.Add("Authorization", authz)
	}

	puller := &helm.OCIPuller{Resolver: docker.NewResolver(docker.ResolverOptions{Headers: headers, Client: netClient})}

	ref := strings.TrimPrefix(strings.TrimSpace(chartTarballURL), "oci://")
	chartBuffer, _, err := puller.PullOCIChart(ref)
	if err != nil {
		return nil, err
	}

	return tarutil.FetchChartDetailFromTarball(chartBuffer)
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
	*models.AppRepositoryInternal
	puller  helm.ChartPuller
	ociCli  ociAPI
	filter  *apprepov1alpha1.FilterRuleSpec
	manager assetManager
}

func doReq(url string, cli *http.Client, headers map[string]string, userAgent string) ([]byte, error) {
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
	CatalogAvailable(ctx context.Context, userAgent string) (bool, error)
	Catalog(ctx context.Context, userAgent string) ([]string, error)
}

// OciAPIClient enables basic interactions with an OCI registry
// using (for now) a combination of the gRPC OCI catalog service
// and the http distribution spec API.
type OciAPIClient struct {
	RegistryNamespaceUrl *url.URL
	// The HttpClient is used for all http requests to the OCI Distribution
	// spec API.
	HttpClient *http.Client
	// The GrpcClient is used when querying our OCI Catalog service, which
	// aims to work around some of the shortfalls of the OCI Distribution spec
	// API
	GrpcClient ocicatalog.OCICatalogServiceClient
	AuthorizationHeader string
}

func (o *OciAPIClient) getOrasRepoClient(appName string, userAgent string) (*remote.Repository, error) {
	url := *o.RegistryNamespaceUrl
	repoName := path.Join(url.Path, appName)
	repoRef := path.Join(url.Host, repoName)
	orasRepoClient, err := remote.NewRepository(repoRef)
	if err != nil {
		return nil, fmt.Errorf("unable to create ORAS client for %q: %w", repoRef, err)
	}
	if url.Scheme == "http" {
		orasRepoClient.PlainHTTP = true
	}

	// Set the http client using our own which adds headers for auth.
	header := auth.DefaultClient.Header.Clone()
	if userAgent != "" {
		header.Set("User-Agent", userAgent)
	}
	if o.AuthorizationHeader != "" {
		header.Set("Authorization", o.AuthorizationHeader)
	}
	orasRepoClient.Client = &auth.Client{
		Client: o.HttpClient,
		Cache:  auth.DefaultCache,
		Header: header,
	}
	return orasRepoClient, nil
}

// TagList retrieves the list of tags for an asset
func (o *OciAPIClient) TagList(appName string, userAgent string) (*TagList, error) {
	orasRepoClient, err := o.getOrasRepoClient(appName, userAgent)
	if err != nil {
		return nil, err
	}

	tags := []string{}

	err = orasRepoClient.Tags(context.TODO(), "", func(ts []string) error {
		tags = append(tags, ts...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &TagList{
		Name: orasRepoClient.Reference.Repository,
		Tags: tags,
	}, nil
}

func (o *OciAPIClient) IsHelmChart(appName, tag, userAgent string) (bool, error) {
	orasRepoClient, err := o.getOrasRepoClient(appName, userAgent)
	if err != nil {
		return false, err
	}
	ctx := context.TODO()
	descriptor, err := orasRepoClient.Resolve(ctx, tag)
	if err != nil {
		log.Errorf("got error: %+v", err)
		return false, err
	}
	rc, err := orasRepoClient.Fetch(ctx, descriptor)
	if err != nil {
		return false, err
	}
	defer rc.Close()

	manifestData, err := content.ReadAll(rc, descriptor)
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
func (o *OciAPIClient) CatalogAvailable(ctx context.Context, userAgent string) (bool, error) {
	manifest, err := o.catalogManifest(userAgent)
	if err == nil {
		return manifest.Config.MediaType == chartsIndexMediaType, nil
	}
	log.V(4).Infof("Unable to get VAC-published catalog manifest: %+v", err)
	if o.GrpcClient == nil {
		// This is not currently an error as the oci-catalog service is
		// still optional.
		log.Errorf("VAC index not available and OCI-Catalog client is nil. Unable to determine catalog")
		return false, nil
	}
	log.V(4).Infof("Attempting catalog retrieval via oci-catalog service.")

	repos_stream, err := o.GrpcClient.ListRepositoriesForRegistry(ctx, &ocicatalog.ListRepositoriesForRegistryRequest{
		Registry:     o.RegistryNamespaceUrl.Host,
		Namespace:    o.RegistryNamespaceUrl.Path,
		ContentTypes: []string{ocicatalog_client.CONTENT_TYPE_HELM},
	})
	if err != nil {
		log.Errorf("Error querying OCI Catalog for repos: %+v", err)
		return false, fmt.Errorf("error querying OCI catalog for repos: %+v", err)
	}

	// It's enough to receive a single repo to be valid.
	_, err = repos_stream.Recv()
	if err != nil {
		if err == io.EOF {
			log.Errorf("OCI catalog returned zero repositories for %q", o.RegistryNamespaceUrl.String())
			return false, nil
		}
		log.Errorf("Error receiving OCI Repositories: %+v", err)
		return false, fmt.Errorf("error receiving OCI Repositories: %+v", err)
	}
	return true, nil
}

func (o *OciAPIClient) catalogManifest(userAgent string) (*OCIManifest, error) {
	indexURL := *o.RegistryNamespaceUrl
	indexURL.Path = path.Join("v2", indexURL.Path, "charts-index", "manifests", "latest")
	log.V(4).Infof("Getting tag %s", indexURL.String())
	headers := map[string]string{
		"Accept": "application/vnd.oci.image.manifest.v1+json",
	}
	manifestData, err := doReq(indexURL.String(), o.HttpClient, headers, userAgent)
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

func (o *OciAPIClient) getVACReposForManifest(manifest *OCIManifest, userAgent string) ([]string, error) {
	if len(manifest.Layers) != 1 || manifest.Layers[0].MediaType != "application/vnd.vmware.charts.index.layer.v1+json" {
		log.Errorf("Unexpected layer in index manifest: %v", manifest)
		return nil, fmt.Errorf("unexpected layer in index manifest")
	}

	blobDigest := manifest.Layers[0].Digest

	blobURL := *o.RegistryNamespaceUrl
	blobURL.Path = path.Join("v2", blobURL.Path, "charts-index", "blobs", blobDigest)
	log.V(4).Infof("Getting blob %s", blobURL.String())
	headers := map[string]string{
		"Accept": "application/vnd.vmware.charts.index.layer.v1+json",
	}
	blobData, err := doReq(blobURL.String(), o.HttpClient, headers, userAgent)
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

// Catalog returns the list of repositories in the (namespaced) registry
// when discoverable.
func (o *OciAPIClient) Catalog(ctx context.Context, userAgent string) ([]string, error) {
	manifest, err := o.catalogManifest(userAgent)
	if err == nil {
		return o.getVACReposForManifest(manifest, userAgent)
	}
	if o.GrpcClient != nil {
		log.V(4).Infof("Unable to find VAC index: %+v. Attempting OCI-Catalog", err)
		repos_stream, err := o.GrpcClient.ListRepositoriesForRegistry(ctx, &ocicatalog.ListRepositoriesForRegistryRequest{
			Registry:     o.RegistryNamespaceUrl.Host,
			Namespace:    o.RegistryNamespaceUrl.Path,
			ContentTypes: []string{ocicatalog_client.CONTENT_TYPE_HELM},
		})
		if err != nil {
			return nil, fmt.Errorf("error querying OCI catalog for repos: %+v", err)
		}

		repos := []string{}
		for {
			repo, err := repos_stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.V(4).Infof("Received repos from oci-catalog service: %+v", repos)
					return repos, nil
				}
				return nil, fmt.Errorf("error receiving OCI Repositories: %+v", err)
			}
			repos = append(repos, repo.Name)
		}
	} else {
		log.V(4).Infof("Unable to find VAC index: %+v and oci-catalog service not configured", err)
		return nil, err
	}
}

// Checksum returns a random sha256 so that the OCI sync will always be
// performed.  This is because we check each app individually whether it has any
// versions that need syncing.
func (r *OCIRegistry) Checksum(ctx context.Context) (string, error) {
	data := make([]byte, 10)
	if _, err := rand.Read(data); err == nil {
		return fmt.Sprintf("%x", sha256.Sum256(data)), nil
	} else {
		return "", err
	}
}

func (r *OCIRegistry) Filters() *apprepov1alpha1.FilterRuleSpec {
	return r.filter
}

// AppRepository returns the repo information
func (r *OCIRegistry) AppRepository() *models.AppRepositoryInternal {
	return r.AppRepositoryInternal
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
		Repo:          &models.AppRepository{Namespace: r.Namespace, Name: r.Name, URL: r.URL, Type: r.Type},
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
		// Note that j.Chart will only be non-nil if this chart was previously synced.
		chart := j.Chart
		errors := []error{}
		log.V(4).Infof("Pulling chart, name=%s, tags=%s", j.AppName, j.VersionsToSync)
		for _, tag := range j.VersionsToSync {
			c, err := pullAndExtract(repoURL, j.AppName, tag, r.puller, r)
			if err != nil {
				errors = append(errors, err)
				continue
			}
			// The model is *weird*, but the first (latest) chart is used as the
			// main chart, and has a single chart version of itself, so
			// subsequent charts are just used for the extra chart versions.
			if chart == nil {
				chart = c
			} else {
				chart.ChartVersions = append(chart.ChartVersions, c.ChartVersions...)
			}
		}

		// Re-sort the ChartVersions
		orderedChartVersions(chart.ChartVersions)

		resultChan <- pullChartResult{*chart, errors}
	}
}

// orderedChartVersions orders the chart versions in descending semver
func orderedChartVersions(chartVersions []models.ChartVersion) {
	slices.SortFunc(chartVersions, func(a, b models.ChartVersion) int {
		va, err := semver.NewVersion(a.Version)
		if err != nil {
			return +1
		}
		vb, err := semver.NewVersion(b.Version)
		if err != nil {
			return -1
		}
		return vb.Compare(va)
	})
}

// orderVersions orders the slice of versions using reverse semver
func orderVersions(versions []string) ([]string, error) {
	vs := make([]*semver.Version, len(versions))
	for i, r := range versions {
		v, err := semver.NewVersion(r)
		if err != nil {
			log.Errorf("Error parsing version: %s", err)
		}
		vs[i] = v
	}
	sort.Sort(sort.Reverse(semver.Collection(vs)))
	orderedVersions := make([]string, len(versions))
	for i, v := range vs {
		orderedVersions[i] = v.String()
	}
	return orderedVersions, nil
}

// Charts retrieve the list of actual charts needing syncing in the repo.
func (r *OCIRegistry) Charts(ctx context.Context, fetchLatestOnly bool, chartResults chan pullChartResult) ([]string, error) {
	repoURL, err := parseRepoURL(r.AppRepositoryInternal.URL)
	if err != nil {
		return nil, err
	}
	if len(r.repositories) == 0 {
		repos, err := r.ociCli.Catalog(ctx, "")
		if err != nil {
			return nil, err
		}
		r.repositories = repos
	}
	chartJobs := make(chan pullChartJob, numWorkersOCI)
	workerChartResults := make(chan pullChartResult, numWorkersOCI)
	var wg sync.WaitGroup
	// Process n apps at a time
	for i := 0; i < numWorkersOCI; i++ {
		wg.Add(1)
		go func() {
			chartImportWorker(repoURL, r, chartJobs, workerChartResults)
			wg.Done()
		}()
	}
	// When we know all workers have sent their data in chartChan, close it.
	go func() {
		wg.Wait()
		close(workerChartResults)
	}()

	// Get the current versions that we're aware of from the DB
	repo := models.AppRepository{Namespace: r.Namespace, Name: r.Name, URL: r.URL, Type: r.Type}
	syncedChartsForRepo, err := r.manager.ChartsForRepo(repo)
	if err != nil {
		return nil, err
	}

	log.V(4).Infof("Starting %d workers for importing OCI charts", numWorkersOCI)
	go func() {
		for _, appName := range r.repositories {
			// Get the list of tags for the app
			tagList, err := r.ociCli.TagList(appName, GetUserAgent("", ""))
			if err != nil {
				log.V(3).ErrorS(err, "unable to list tags")
				log.Errorf("unable to list tags: %+v", err)
				close(chartJobs)
				return
			}
			tags, err := orderVersions(tagList.Tags)
			if err != nil {
				log.V(3).ErrorS(err, "Error parsing version")
				close(chartJobs)
				return
			}
			// Find the tags present in DB, in order verify the difference.
			syncedChart := syncedChartsForRepo[appName]
			syncedVersions := []string{}
			if syncedChart != nil {
				for _, cv := range syncedChart.ChartVersions {
					syncedVersions = append(syncedVersions, cv.Version)
				}
			}
			// We want to sync only those versions that we don't already have synced
			versionsToSync := []string{}
			for _, tag := range tags {
				if !slice.ContainsString(syncedVersions, tag, func(s string) string { return s }) {
					versionsToSync = append(versionsToSync, tag)
				}
			}

			if len(versionsToSync) == 0 {
				log.V(4).Infof("No versions requiring sync for %q", appName)
				continue
			}

			if fetchLatestOnly {
				// TODO(minelson): There's a small but non-zero chance that the
				// latest tag is for non-chart data. Worst case here is that the app
				// won't appear in the UI until the non-shallow sync syncs its chart tags.
				chartJobs <- pullChartJob{
					AppName:        appName,
					VersionsToSync: []string{versionsToSync[0]},
					Chart:          syncedChart,
				}
				log.V(4).Infof("Queued only the first tag for %q for shallow sync : %q", appName, versionsToSync[0])
			} else {
				limitedVersionsToSync := versionsToSync
				if len(limitedVersionsToSync) > maxOCIVersionsForOneSync {
					limitedVersionsToSync = limitedVersionsToSync[:maxOCIVersionsForOneSync]
					log.V(4).Infof("Queued only the next %d versions of %q during this sync: %v", maxOCIVersionsForOneSync, appName, versionsToSync[:maxOCIVersionsForOneSync])
				} else {
					log.V(4).Infof("Queued all remaining  versions for %q: %v", appName, versionsToSync)
				}
				chartJobs <- pullChartJob{
					AppName:        appName,
					VersionsToSync: limitedVersionsToSync,
					Chart:          syncedChart,
				}
			}
		}
		close(chartJobs)
	}()

	go func() {
		// Start receiving charts from the multiple workers and pass them down
		// to the caller.
		for res := range workerChartResults {
			chartResults <- res
		}
		close(chartResults)
	}()

	chartsForDeletion := []string{}
	for syncedChartName := range syncedChartsForRepo {
		if !slice.ContainsString(r.repositories, syncedChartName, func(s string) string { return s }) {
			chartsForDeletion = append(chartsForDeletion, syncedChartName)
		}
	}

	return chartsForDeletion, nil
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

func getHelmRepo(namespace, name, repoURL, authorizationHeader string, filter *apprepov1alpha1.FilterRuleSpec, netClient *http.Client, userAgent string, manager assetManager) (ChartCatalog, error) {
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
		AppRepositoryInternal: &models.AppRepositoryInternal{
			Namespace:           namespace,
			Name:                name,
			URL:                 url.String(),
			AuthorizationHeader: authorizationHeader,
		},
		netClient: netClient,
		filter:    filter,
		manager:   manager,
	}, nil
}

func getOCIRepo(namespace, name, repoURL, authorizationHeader string, filter *apprepov1alpha1.FilterRuleSpec, ociRepos []string, netClient *http.Client, grpcClient *ocicatalog.OCICatalogServiceClient, manager assetManager) (ChartCatalog, error) {
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
		repositories:          ociRepos,
		AppRepositoryInternal: &models.AppRepositoryInternal{Namespace: namespace, Name: name, URL: url.String(), AuthorizationHeader: authorizationHeader},
		puller:                &helm.OCIPuller{Resolver: ociResolver},
		ociCli:                &OciAPIClient{RegistryNamespaceUrl: url, HttpClient: netClient, GrpcClient: *grpcClient, AuthorizationHeader: authorizationHeader},
		filter:                filter,
		manager:               manager,
	}, nil
}

func fetchRepoIndex(url, authHeader string, cli *http.Client, userAgent string) ([]byte, error) {
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

func chartTarballURL(r *models.AppRepositoryInternal, cv models.ChartVersion) string {
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
	netClient *http.Client
}

func (f *fileImporter) fetchFiles(inputCharts chan models.Chart, repo ChartCatalog, userAgent string, passCredentials bool, done chan bool) {
	iconJobs := make(chan models.Chart, numWorkersFiles)
	chartFilesJobs := make(chan importChartFilesJob, numWorkersFiles)
	var wg sync.WaitGroup

	log.V(4).Infof("Starting %d file importer workers", numWorkersFiles)
	for i := 0; i < numWorkersFiles; i++ {
		wg.Add(1)
		go f.importWorker(&wg, iconJobs, chartFilesJobs, repo, userAgent, passCredentials)
	}

	// Enqueue jobs to process chart icons and record the charts for further
	// processing.
	charts := []models.Chart{}
	for c := range inputCharts {
		iconJobs <- c
		charts = append(charts, c)
	}
	log.V(4).Infof("Finished queueing icon jobs")
	// Close the iconJobs channel to signal the worker pools to move on to the
	// chart files jobs
	close(iconJobs)

	// Iterate through the list of charts and enqueue the latest chart version to
	// be processed. Append the rest of the chart versions to a list to be
	// enqueued later
	var toEnqueue []importChartFilesJob
	log.V(4).Infof("Enqueuing chart file imports for first versions")
	for _, c := range charts {
		chartFilesJobs <- importChartFilesJob{c.ID, c.Repo, c.ChartVersions[0]}
		for _, cv := range c.ChartVersions[1:] {
			toEnqueue = append(toEnqueue, importChartFilesJob{c.ID, c.Repo, cv})
		}
	}

	// Enqueue all the remaining chart versions
	log.V(4).Infof("Enqueuing chart file imports for remaining versions")
	for _, cfj := range toEnqueue {
		chartFilesJobs <- cfj
	}
	// Close the chartFilesJobs channel to signal the worker pools that there are
	// no more jobs to process
	close(chartFilesJobs)

	// Wait for the worker pools to finish processing
	log.V(4).Infof("Waiting for file import workers to complete.")
	wg.Wait()

	log.V(4).Infof("File importing complete")
	done <- true
}

func (f *fileImporter) importWorker(wg *sync.WaitGroup, icons <-chan models.Chart, chartFiles <-chan importChartFilesJob, repo ChartCatalog, userAgent string, passCredentials bool) {
	defer wg.Done()
	for c := range icons {
		log.V(4).Infof("Importing icon, name=%s", c.Name)
		if err := f.fetchAndImportIcon(c, repo.AppRepository(), userAgent, passCredentials); err != nil {
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

func (f *fileImporter) fetchAndImportIcon(c models.Chart, r *models.AppRepositoryInternal, userAgent string, passCredentials bool) error {
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

	return f.manager.updateIcon(models.AppRepository{Namespace: r.Namespace, Name: r.Name}, b, contentType, c.ID)
}

func (f *fileImporter) fetchAndImportFiles(chartID string, repo ChartCatalog, cv models.ChartVersion, userAgent string, passCredentials bool) error {
	r := repo.AppRepository()
	chartFilesID := fmt.Sprintf("%s-%s", chartID, cv.Version)

	// Check if we already have indexed files for this chart version and digest
	if f.manager.filesExist(models.AppRepository{Namespace: r.Namespace, Name: r.Name}, chartFilesID, cv.Digest) {
		log.V(4).Infof("Skipping existing files, id: %s, version: %s", chartID, cv.Version)
		return nil
	}
	log.V(4).Infof("Fetching files, id=%s, version=%s", chartID, cv.Version)

	files, err := repo.FetchFiles(cv, userAgent, passCredentials)
	if err != nil {
		return err
	}

	chartFiles := models.ChartFiles{ID: chartFilesID, Repo: &models.AppRepository{Name: r.Name, Namespace: r.Namespace, URL: r.URL}, Digest: cv.Digest}
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
