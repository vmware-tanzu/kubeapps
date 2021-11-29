/*
Copyright © 2021 VMware
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
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/kubeapps/common/datastore"
	appRepov1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	helmv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/agent"
	chartutils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	authorizationv1 "k8s.io/api/authorization/v1"
	corek8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)
type helmActionConfigGetter func(ctx context.Context, pkgContext *corev1.Context) (*action.Configuration, error)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

const (
	MajorVersionsInSummary       = 3
	MinorVersionsInSummary       = 3
	PatchVersionsInSummary       = 3
	UserAgentPrefix              = "kubeapps-apis/plugins"
	DefaultTimeoutSeconds  int32 = 300
)

// Wapper struct to include three version constants
type VersionsInSummary struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

type createRelease func(*action.Configuration, string, string, string, *chart.Chart, map[string]string, int32) (*release.Release, error)

// Server implements the helm packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedHelmPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientGetter
	globalPackagingNamespace string
	globalPackagingCluster   string
	manager                  utils.AssetManager
	actionConfigGetter       helmActionConfigGetter
	chartClientFactory       chartutils.ChartClientFactoryInterface
	versionsInSummary        VersionsInSummary
	timeoutSeconds           int32
	createReleaseFunc        createRelease
}

// parsePluginConfig parses the input plugin configuration json file and return the configuration options.
func parsePluginConfig(pluginConfigPath string) (VersionsInSummary, int32, error) {

	// Note at present VersionsInSummary is the only configurable option for this plugin,
	// and if required this func can be enhaned to return helmConfig struct

	// In the helm plugin, for example, we are interested in config for the
	// core.packages.v1alpha1 only. So the plugin defines the following struct and parses the config.
	type helmConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`
	}
	var config helmConfig

	pluginConfig, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return VersionsInSummary{}, 0, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return VersionsInSummary{}, 0, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	// return configured value
	return config.Core.Packages.V1alpha1.VersionsInSummary, config.Core.Packages.V1alpha1.TimeoutSeconds, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster string, pluginConfigPath string) *Server {
	var kubeappsNamespace = os.Getenv("POD_NAMESPACE")
	var ASSET_SYNCER_DB_URL = os.Getenv("ASSET_SYNCER_DB_URL")
	var ASSET_SYNCER_DB_NAME = os.Getenv("ASSET_SYNCER_DB_NAME")
	var ASSET_SYNCER_DB_USERNAME = os.Getenv("ASSET_SYNCER_DB_USERNAME")
	var ASSET_SYNCER_DB_USERPASSWORD = os.Getenv("ASSET_SYNCER_DB_USERPASSWORD")

	var dbConfig = datastore.Config{URL: ASSET_SYNCER_DB_URL, Database: ASSET_SYNCER_DB_NAME, Username: ASSET_SYNCER_DB_USERNAME, Password: ASSET_SYNCER_DB_USERPASSWORD}

	manager, err := utils.NewPGManager(dbConfig, kubeappsNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}

	// If no config is provided, we default to the existing values for backwards
	// compatibility.
	versionsInSummary := VersionsInSummary{
		Major: MajorVersionsInSummary,
		Minor: MinorVersionsInSummary,
		Patch: PatchVersionsInSummary,
	}
	timeoutSeconds := DefaultTimeoutSeconds
	if pluginConfigPath != "" {
		versionsInSummary, timeoutSeconds, err = parsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+helm using custom packages config with %v and timeout %d\n", versionsInSummary, timeoutSeconds)
	} else {
		log.Infof("+helm using default config since pluginConfigPath is empty")
	}

	return &Server{
		clientGetter: func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
			if configGetter == nil {
				return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
			}
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get dynamic client : %v", err))
			}
			typedClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get typed client : %v", err))
			}
			return typedClient, dynamicClient, nil
		},
		actionConfigGetter: func(ctx context.Context, pkgContext *corev1.Context) (*action.Configuration, error) {
			if configGetter == nil {
				return nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			cluster := pkgContext.GetCluster()
			// Don't force clients to send a cluster unless we are sure all use-cases
			// of kubeapps-api are multicluster.
			if cluster == "" {
				cluster = globalPackagingCluster
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
			}

			restClientGetter := agent.NewConfigFlagsFromCluster(pkgContext.GetNamespace(), config)
			clientSet, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to create kubernetes client : %v", err))
			}
			// TODO(mnelson): Update to allow different helm storage options.
			storage := agent.StorageForSecrets(pkgContext.GetNamespace(), clientSet)
			return &action.Configuration{
				RESTClientGetter: restClientGetter,
				KubeClient:       kube.New(restClientGetter),
				Releases:         storage,
				Log:              log.Infof,
			}, nil
		},
		manager:                  manager,
		globalPackagingNamespace: kubeappsNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		chartClientFactory:       &chartutils.ChartClientFactory{},
		versionsInSummary:        versionsInSummary,
		timeoutSeconds:           timeoutSeconds,
		createReleaseFunc:        agent.CreateRelease,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}

// GetManager ensures a manager is available and returns it.
func (s *Server) GetManager() (utils.AssetManager, error) {
	if s.manager == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with manager")
	}
	manager := s.manager
	return manager, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageSummaries %s", contextMsg)

	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()

	// Check the requested namespace: if any, return "everything a user can read";
	// otherwise, first check if the user can access the requested ns
	if namespace == "" {
		// TODO(agamez): not including a namespace means that it returns everything a user can read
		return nil, status.Errorf(codes.Unimplemented, "Not supported yet: not including a namespace means that it returns everything a user can read")
	}
	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, cluster, namespace); err != nil {
		return nil, err
	}

	// If the request is for available packages on another cluster, we only
	// return the global packages (ie. kubeapps namespace)
	if cluster != "" && cluster != s.globalPackagingCluster {
		namespace = s.globalPackagingNamespace
	}

	// Create the initial chart query with the namespace
	cq := utils.ChartQuery{
		Namespace: namespace,
	}

	// Add any other filter if a FilterOptions is passed
	if request.FilterOptions != nil {
		cq.Categories = request.FilterOptions.Categories
		cq.SearchQuery = request.FilterOptions.Query
		cq.Repos = request.FilterOptions.Repositories
		cq.Version = request.FilterOptions.PkgVersion
		cq.AppVersion = request.FilterOptions.AppVersion
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to intepret page token %q: %v", request.GetPaginationOptions().GetPageToken(), err)
	}

	// This plugin will include, as part of the GetAvailablePackageSummariesResponse,
	// a "Categories" field containing only the distinct category names considering just the namespace
	chartCategories, err := s.manager.GetAllChartCategories(utils.ChartQuery{Namespace: namespace})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to fetch chart categories: %v", err)
	}

	var categories []string
	for _, cat := range chartCategories {
		categories = append(categories, cat.Name)
	}

	// The current assetsvc manager works on a page number (ie. 1 for the first page),
	// rather than an offset.
	pageNumber := pageOffset + 1
	charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, pageNumber, int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve charts: %v", err)
	}

	// Convert the charts response into a GetAvailablePackageSummariesResponse
	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg, err := AvailablePackageSummaryFromChart(chart)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to parse chart to an AvailablePackageSummary: %v", err)
		}
		// We currently support app repositories on the kubeapps cluster only.
		pkg.AvailablePackageRef.Context.Cluster = s.globalPackagingCluster
		responsePackages = append(responsePackages, pkg)
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(responsePackages) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: responsePackages,
		NextPageToken:             nextPageToken,
		Categories:                categories,
	}, nil
}

// pageOffsetFromPageToken converts a page token to an integer offset
// representing the page of results.
// TODO(mnelson): When aggregating results from different plugins, we'll
// need to update the actual query in GetPaginatedChartListWithFilters to
// use a row offset rather than a page offset (as not all rows may be consumed
// for a specific plugin when combining).
func pageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.ParseUint(pageToken, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(offset), nil
}

// AvailablePackageSummaryFromChart builds an AvailablePackageSummary from a Chart
func AvailablePackageSummaryFromChart(chart *models.Chart) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	isValid, err := isValidChart(chart)
	if !isValid || err != nil {
		return nil, status.Errorf(codes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.Name = chart.Name
	// Helm's Chart.yaml (and hence our model) does not include a separate
	// display name, so the chart name is also used here.
	pkg.DisplayName = chart.Name
	pkg.IconUrl = chart.Icon
	pkg.ShortDescription = chart.Description
	pkg.Categories = []string{chart.Category}

	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     GetPluginDetail(),
	}
	pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}

	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.LatestVersion = &corev1.PackageAppVersion{
			PkgVersion: chart.ChartVersions[0].Version,
			AppVersion: chart.ChartVersions[0].AppVersion,
		}
	}

	return pkg, nil
}

// getUnescapedChartID takes a chart id with URI-encoded characters and decode them. Ex: 'foo%2Fbar' becomes 'foo/bar'
func getUnescapedChartID(chartID string) (string, error) {
	unescapedChartID, err := url.QueryUnescape(chartID)
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "Unable to decode chart ID chart: %v", chartID)
	}
	// TODO(agamez): support ID with multiple slashes, eg: aaa/bbb/ccc
	chartIDParts := strings.Split(unescapedChartID, "/")
	if len(chartIDParts) != 2 {
		return "", status.Errorf(codes.InvalidArgument, "Incorrect request.AvailablePackageRef.Identifier, currently just 'foo/bar' patters are supported: %s", chartID)
	}
	return unescapedChartID, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageDetail %s", contextMsg)

	if request.GetAvailablePackageRef().GetContext() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "No request AvailablePackageRef.Context provided")
	}

	// Retrieve namespace, cluster, version from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	version := request.PkgVersion

	// Currently we support available packages on the kubeapps cluster only.
	if cluster != "" && cluster != s.globalPackagingCluster {
		return nil, status.Errorf(codes.InvalidArgument, "Requests for available packages on clusters other than %q not supported. Requested cluster was: %q", s.globalPackagingCluster, cluster)
	}

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, cluster, namespace); err != nil {
		return nil, err
	}

	unescapedChartID, err := getUnescapedChartID(request.AvailablePackageRef.Identifier)
	if err != nil {
		return nil, err
	}

	// Since the version is optional, in case of an empty one, fall back to get all versions and get the first one
	// Note that the chart version array should be ordered
	var chart models.Chart
	if version == "" {
		log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
		chart, err = s.manager.GetChart(namespace, unescapedChartID)
	} else {
		log.Infof("Requesting chart '%s' (version %s) in ns '%s'", unescapedChartID, version, namespace)
		chart, err = s.manager.GetChartVersion(namespace, unescapedChartID, version)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve chart: %v", err)
	}

	if len(chart.ChartVersions) == 0 {
		return nil, status.Errorf(codes.Internal, "Chart returned without any versions: %+v", chart)
	}
	if version == "" {
		version = chart.ChartVersions[0].Version
	}
	fileID := fileIDForChart(unescapedChartID, chart.ChartVersions[0].Version)
	chartFiles, err := s.manager.GetChartFiles(namespace, fileID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve chart files: %v", err)
	}

	availablePackageDetail, err := AvailablePackageDetailFromChart(&chart, &chartFiles)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to parse chart to an availablePackageDetail: %v", err)
	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}, nil
}

// fileIDForChart returns a file ID given a chart id and version.
func fileIDForChart(id, version string) string {
	return fmt.Sprintf("%s-%s", id, version)
}

// GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageVersions %s", contextMsg)

	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	if namespace == "" || request.GetAvailablePackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Required context or identifier not provided")
	}
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	// Currently we support available packages on the kubeapps cluster only.
	if cluster != "" && cluster != s.globalPackagingCluster {
		return nil, status.Errorf(codes.InvalidArgument, "Requests for versions of available packages on clusters other than %q not supported. Requested cluster was %q.", s.globalPackagingCluster, cluster)
	}

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, cluster, namespace); err != nil {
		return nil, err
	}

	unescapedChartID, err := getUnescapedChartID(request.GetAvailablePackageRef().GetIdentifier())
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
	chart, err := s.manager.GetChart(namespace, unescapedChartID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve chart: %v", err)
	}

	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: packageAppVersionsSummary(chart.ChartVersions, s.versionsInSummary),
	}, nil
}

// packageAppVersionsSummary converts the model chart versions into the required version summary.
func packageAppVersionsSummary(versions []models.ChartVersion, versionInSummary VersionsInSummary) []*corev1.PackageAppVersion {
	pav := []*corev1.PackageAppVersion{}

	// Use a version map to be able to count how many major, minor and patch versions
	// we have included.
	version_map := map[int64]map[int64][]int64{}
	for _, v := range versions {
		version, err := semver.NewVersion(v.Version)
		if err != nil {
			continue
		}

		if _, ok := version_map[version.Major()]; !ok {
			// Don't add a new major version if we already have enough
			if len(version_map) >= versionInSummary.Major {
				continue
			}
		} else {
			// If we don't yet have this minor version
			if _, ok := version_map[version.Major()][version.Minor()]; !ok {
				// Don't add a new minor version if we already have enough for this major version
				if len(version_map[version.Major()]) >= versionInSummary.Minor {
					continue
				}
			} else {
				if len(version_map[version.Major()][version.Minor()]) >= versionInSummary.Patch {
					continue
				}
			}
		}

		// Include the version and update the version map.
		pav = append(pav, &corev1.PackageAppVersion{
			PkgVersion: v.Version,
			AppVersion: v.AppVersion,
		})

		if _, ok := version_map[version.Major()]; !ok {
			version_map[version.Major()] = map[int64][]int64{}
		}
		version_map[version.Major()][version.Minor()] = append(version_map[version.Major()][version.Minor()], version.Patch())
	}

	return pav
}

// AvailablePackageDetailFromChart builds an AvailablePackageDetail from a Chart
func AvailablePackageDetailFromChart(chart *models.Chart, chartFiles *models.ChartFiles) (*corev1.AvailablePackageDetail, error) {
	pkg := &corev1.AvailablePackageDetail{}

	isValid, err := isValidChart(chart)
	if !isValid || err != nil {
		return nil, status.Errorf(codes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.DisplayName = chart.Name
	pkg.HomeUrl = chart.Home
	pkg.IconUrl = chart.Icon
	pkg.Name = chart.Name
	pkg.ShortDescription = chart.Description
	pkg.Categories = []string{chart.Category}
	pkg.SourceUrls = chart.Sources

	pkg.Maintainers = []*corev1.Maintainer{}
	for _, maintainer := range chart.Maintainers {
		m := &corev1.Maintainer{Name: maintainer.Name, Email: maintainer.Email}
		pkg.Maintainers = append(pkg.Maintainers, m)
	}

	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     GetPluginDetail(),
	}
	if chart.Repo != nil {
		pkg.RepoUrl = chart.Repo.URL
		pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}
	}

	// We assume that chart.ChartVersions[0] will always contain either: the latest version or the specified version
	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.Version = &corev1.PackageAppVersion{
			PkgVersion: chart.ChartVersions[0].Version,
			AppVersion: chart.ChartVersions[0].AppVersion,
		}
		pkg.Readme = chartFiles.Readme
		pkg.DefaultValues = chartFiles.Values
		pkg.ValuesSchema = chartFiles.Schema
	}
	return pkg, nil
}

// hasAccessToNamespace returns an error if the client does not have read access to a given namespace
func (s *Server) hasAccessToNamespace(ctx context.Context, cluster, namespace string) error {
	// If checking the global namespace, allow access always
	if namespace == s.globalPackagingNamespace {
		return nil
	}
	client, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return err
	}

	res, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Group:     "",
				Resource:  "secrets",
				Verb:      "get",
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return status.Errorf(codes.Internal, "Unable to check if the user has access to the namespace: %s", err)
	}
	if !res.Status.Allowed {
		// If the user has not access, return a unauthenticated response, otherwise, continue
		return status.Errorf(codes.Unauthenticated, "The current user has no access to the namespace %q", namespace)
	}
	return nil
}

// isValidChart returns true if the chart model passed defines a value
// for each required field described at the Helm website:
// https://helm.sh/docs/topics/charts/#the-chartyaml-file
// together with required fields for our model.
func isValidChart(chart *models.Chart) (bool, error) {
	if chart.Name == "" {
		return false, status.Errorf(codes.Internal, "required field .Name not found on helm chart: %v", chart)
	}
	if chart.ID == "" {
		return false, status.Errorf(codes.Internal, "required field .ID not found on helm chart: %v", chart)
	}
	if chart.Repo == nil {
		return false, status.Errorf(codes.Internal, "required field .Repo not found on helm chart: %v", chart)
	}
	if chart.ChartVersions == nil || len(chart.ChartVersions) == 0 {
		return false, status.Errorf(codes.Internal, "required field .chart.ChartVersions[0] not found on helm chart: %v", chart)
	} else {
		for _, chartVersion := range chart.ChartVersions {
			if chartVersion.Version == "" {
				return false, status.Errorf(codes.Internal, "required field .ChartVersions[i].Version not found on helm chart: %v", chart)
			}
		}
	}
	for _, maintainer := range chart.Maintainers {
		if maintainer.Name == "" {
			return false, status.Errorf(codes.Internal, "required field .Maintainers[i].Name not found on helm chart: %v", chart)
		}
	}
	return true, nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+helm GetInstalledPackageSummaries %s", contextMsg)

	actionConfig, err := s.actionConfigGetter(ctx, request.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}
	cmd := action.NewList(actionConfig)
	if request.GetContext().GetNamespace() == "" {
		cmd.AllNamespaces = true
	}

	cmd.Limit = int(request.GetPaginationOptions().GetPageSize())
	cmd.Offset, err = pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to intepret page token %q: %v", request.GetPaginationOptions().GetPageToken(), err)
	}

	// TODO(mnelson): Check whether we need to support ListAll (status == "all" in existing helm support)

	releases, err := cmd.Run()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to run Helm List action: %v", err)
	}

	installedPkgSummaries := make([]*corev1.InstalledPackageSummary, len(releases))
	cluster := request.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	for i, r := range releases {
		installedPkgSummaries[i] = installedPkgSummaryFromRelease(r)
		installedPkgSummaries[i].InstalledPackageRef.Context.Cluster = cluster
	}

	// Fill in the latest package version for each.
	// TODO(mnelson): Update to do this with a single query rather than iterating and
	// querying per release.
	for i, rel := range releases {
		// Helm does not store a back-reference to the chart used to create a
		// release (https://github.com/helm/helm/issues/6464), so for each
		// release, we look up a chart with than name and version available in
		// the release namespace, and if one is found, pull out the latest chart
		// version.
		cq := utils.ChartQuery{
			Namespace:  rel.Namespace,
			ChartName:  rel.Chart.Metadata.Name,
			Version:    rel.Chart.Metadata.Version,
			AppVersion: rel.Chart.Metadata.AppVersion,
		}
		charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, 1, 0)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Error while fetching related charts: %v", err)
		}
		// TODO(agamez): deal with multiple matches, perhaps returning []AvailablePackageRef ?
		// Example: global + namespaced repo including an overlapping subset of packages.
		if len(charts) > 0 && len(charts[0].ChartVersions) > 0 {
			installedPkgSummaries[i].LatestVersion = &corev1.PackageAppVersion{
				PkgVersion: charts[0].ChartVersions[0].Version,
				AppVersion: charts[0].ChartVersions[0].AppVersion,
			}
		}
		installedPkgSummaries[i].Status = &corev1.InstalledPackageStatus{
			Ready:      rel.Info.Status == release.StatusDeployed,
			Reason:     statusReasonForHelmStatus(rel.Info.Status),
			UserReason: rel.Info.Status.String(),
		}
		installedPkgSummaries[i].CurrentVersion = &corev1.PackageAppVersion{
			PkgVersion: rel.Chart.Metadata.Version,
			AppVersion: rel.Chart.Metadata.AppVersion,
		}
	}

	response := &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummaries,
	}
	if len(releases) == cmd.Limit {
		response.NextPageToken = fmt.Sprintf("%d", cmd.Limit+1)
	}
	return response, nil
}

func statusReasonForHelmStatus(s release.Status) corev1.InstalledPackageStatus_StatusReason {
	switch s {
	case release.StatusDeployed:
		return corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED
	case release.StatusFailed:
		return corev1.InstalledPackageStatus_STATUS_REASON_FAILED
	case release.StatusPendingInstall, release.StatusPendingRollback, release.StatusPendingUpgrade, release.StatusUninstalling:
		return corev1.InstalledPackageStatus_STATUS_REASON_PENDING
	}
	// Both StatusUninstalled and StatusSuperseded will be unknown/unspecified.
	return corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED
}

func installedPkgSummaryFromRelease(r *release.Release) *corev1.InstalledPackageSummary {
	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: r.Namespace,
			},
			Identifier: r.Name,
		},
		Name: r.Name,
		PkgVersionReference: &corev1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
		},
		IconUrl:          r.Chart.Metadata.Icon,
		PkgDisplayName:   r.Chart.Name(),
		ShortDescription: r.Chart.Metadata.Description,
	}
}

// GetInstalledPackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *corev1.GetInstalledPackageDetailRequest) (*corev1.GetInstalledPackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetInstalledPackageRef().GetContext().GetCluster(), request.GetInstalledPackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetInstalledPackageDetail %s", contextMsg)

	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	identifier := request.GetInstalledPackageRef().GetIdentifier()

	actionConfig, err := s.actionConfigGetter(ctx, request.GetInstalledPackageRef().GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	// First grab the release.
	getcmd := action.NewGet(actionConfig)
	release, err := getcmd.Run(identifier)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", identifier, namespace, err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to run Helm get action: %v", err)
	}
	installedPkgDetail, err := installedPkgDetailFromRelease(release, request.GetInstalledPackageRef())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create installed package detail from release: %v", err)
	}

	// Grab the released values.
	valuescmd := action.NewGetValues(actionConfig)
	values, err := valuescmd.Run(identifier)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to get Helm release values: %v", err)
	}
	valuesMarshalled, err := json.Marshal(values)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to marshal Helm release values: %v", err)
	}
	installedPkgDetail.ValuesApplied = string(valuesMarshalled)

	// Check for a chart matching the installed package.
	cq := utils.ChartQuery{
		Namespace:  release.Namespace,
		ChartName:  release.Chart.Metadata.Name,
		Version:    release.Chart.Metadata.Version,
		AppVersion: release.Chart.Metadata.AppVersion,
	}
	charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, 1, 0)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Error while fetching related chart: %v", err)
	}
	// TODO(agamez): deal with multiple matches, perhaps returning []AvailablePackageRef ?
	// Example: global + namespaced repo including an overlapping subset of packages.
	if len(charts) > 0 {
		installedPkgDetail.AvailablePackageRef = &corev1.AvailablePackageReference{
			Identifier: charts[0].ID,
			Plugin:     GetPluginDetail(),
		}
		if charts[0].Repo != nil {
			installedPkgDetail.AvailablePackageRef.Context = &corev1.Context{
				Namespace: charts[0].Repo.Namespace,
				Cluster:   s.globalPackagingCluster,
			}
		}
		if len(charts[0].ChartVersions) > 0 {
			cv := charts[0].ChartVersions[0]
			installedPkgDetail.LatestVersion = &corev1.PackageAppVersion{
				PkgVersion: cv.Version,
				AppVersion: cv.AppVersion,
			}
		}
	}

	return &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPkgDetail,
	}, nil
}

func installedPkgDetailFromRelease(r *release.Release, ref *corev1.InstalledPackageReference) (*corev1.InstalledPackageDetail, error) {
	customDetailHelm, err := anypb.New(&helmv1.InstalledPackageDetailCustomDataHelm{
		ReleaseRevision: int32(r.Version),
	})
	if err != nil {
		return nil, err
	}
	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: ref,
		Name:                r.Name,
		PkgVersionReference: &corev1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
		},
		PostInstallationNotes: r.Info.Notes,
		Status: &corev1.InstalledPackageStatus{
			Ready:      r.Info.Status == release.StatusDeployed,
			Reason:     statusReasonForHelmStatus(r.Info.Status),
			UserReason: r.Info.Status.String(),
		},
		CustomDetail: customDetailHelm,
	}, nil
}

func splitChartIdentifier(chartID string) (repoName, chartName string, err error) {
	// getUnescapedChartID also ensures that there are two parts (ie. repo/chart-name only)
	unescapedChartID, err := getUnescapedChartID(chartID)
	if err != nil {
		return "", "", err
	}
	chartIDParts := strings.Split(unescapedChartID, "/")
	return chartIDParts[0], chartIDParts[1], nil
}

// CreateInstalledPackage creates an installed package.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetTargetContext().GetCluster(), request.GetTargetContext().GetNamespace())
	log.Infof("+helm CreateInstalledPackage %s", contextMsg)

	typedClient, _, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := request.GetAvailablePackageRef().GetIdentifier()
	repoNamespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	repoName, chartName, err := splitChartIdentifier(chartID)
	if err != nil {
		return nil, err
	}
	chartDetails := &chartutils.Details{
		AppRepositoryResourceName:      repoName,
		AppRepositoryResourceNamespace: repoNamespace,
		ChartName:                      chartName,
		Version:                        request.GetPkgVersionReference().GetVersion(),
	}
	ch, registrySecrets, err := s.fetchChartWithRegistrySecrets(ctx, chartDetails, typedClient)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Missing permissions %v", err)
	}

	// Create an action config for the target namespace.
	actionConfig, err := s.actionConfigGetter(ctx, request.GetTargetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := s.createReleaseFunc(actionConfig, request.GetName(), request.GetTargetContext().GetNamespace(), request.GetValues(), ch, registrySecrets, s.timeoutSeconds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create helm release %q in the namespace %q: %v", request.GetName(), request.GetTargetContext().GetNamespace(), err)
	}

	cluster := request.GetTargetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	return &corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: release.Namespace,
			},
			Identifier: release.Name,
			Plugin:     GetPluginDetail(),
		},
	}, nil
}

// UpdateInstalledPackage updates an installed package.
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *corev1.UpdateInstalledPackageRequest) (*corev1.UpdateInstalledPackageResponse, error) {
	installedRef := request.GetInstalledPackageRef()
	releaseName := installedRef.GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", installedRef.GetContext().GetCluster(), installedRef.GetContext().GetNamespace())
	log.Infof("+helm UpdateInstalledPackage %s", contextMsg)

	// Determine the chart used for this installed package.
	// We may want to include the AvailablePackageRef in the request, given
	// that it can be ambiguous, but we dont yet have a UI that allows the
	// user to select which chart to use, so until then, we're fetching the
	// available package ref via the detail (bit of a short-cut, could query
	// for the ref directly).
	detailResponse, err := s.GetInstalledPackageDetail(ctx, &corev1.GetInstalledPackageDetailRequest{
		InstalledPackageRef: installedRef,
	})
	if err != nil {
		return nil, err
	}

	availablePkgRef := detailResponse.GetInstalledPackageDetail().GetAvailablePackageRef()
	if availablePkgRef == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Unable to find the available package used to deploy %q in the namespace %q.", releaseName, installedRef.GetContext().GetNamespace())
	}

	typedClient, _, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := availablePkgRef.GetIdentifier()
	repoName, chartName, err := splitChartIdentifier(chartID)
	if err != nil {
		return nil, err
	}
	chartDetails := &chartutils.Details{
		AppRepositoryResourceName:      repoName,
		AppRepositoryResourceNamespace: availablePkgRef.GetContext().GetNamespace(),
		ChartName:                      chartName,
		Version:                        request.GetPkgVersionReference().GetVersion(),
	}
	ch, registrySecrets, err := s.fetchChartWithRegistrySecrets(ctx, chartDetails, typedClient)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Missing permissions %v", err)
	}

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := agent.UpgradeRelease(actionConfig, releaseName, request.GetValues(), ch, registrySecrets, s.timeoutSeconds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to upgrade helm release %q in the namespace %q: %v", releaseName, installedRef.GetContext().GetNamespace(), err)
	}

	cluster := installedRef.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	return &corev1.UpdateInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: release.Namespace,
			},
			Identifier: release.Name,
			Plugin:     GetPluginDetail(),
		},
	}, nil
}

// GetAppRepoAndRelatedSecrets retrieves the given repo from its namespace
// Depending on the repo namespace and the
func (s *Server) getAppRepoAndRelatedSecrets(ctx context.Context, appRepoName, appRepoNamespace string) (*appRepov1.AppRepository, *corek8sv1.Secret, *corek8sv1.Secret, error) {

	// We currently get app repositories on the kubeapps cluster only.
	typedClient, dynClient, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, nil, nil, err
	}

	// Using the dynamic client to get the AppRepository rather than updating the interface
	// returned by the client getter, then converting the result back to a typed AppRepository
	// since our existing code that we're re-using expects one.
	gvr := schema.GroupVersionResource{
		Group:    "kubeapps.com",
		Version:  "v1alpha1",
		Resource: "apprepositories",
	}
	appRepoUnstructured, err := dynClient.Resource(gvr).Namespace(appRepoNamespace).Get(ctx, appRepoName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to get app repository %s/%s: %v", appRepoNamespace, appRepoName, err)
	}

	var appRepo appRepov1.AppRepository
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(appRepoUnstructured.UnstructuredContent(), &appRepo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to convert unstructured AppRepository for %s/%s to a structured AppRepository: %v", appRepoNamespace, appRepoName, err)
	}

	auth := appRepo.Spec.Auth
	var caCertSecret *corek8sv1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		caCertSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *corek8sv1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", secretName, err)
		}
	}

	return &appRepo, caCertSecret, authSecret, nil
}

// fetchChartWithRegistrySecrets returns the chart and related registry secrets.
//
// Mainly to DRY up similar code in the create and update methods.
func (s *Server) fetchChartWithRegistrySecrets(ctx context.Context, chartDetails *chartutils.Details, client kubernetes.Interface) (*chart.Chart, map[string]string, error) {
	// Most of the existing code that we want to reuse is based on having a typed AppRepository.
	appRepo, caCertSecret, authSecret, err := s.getAppRepoAndRelatedSecrets(ctx, chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Unable to fetch app repo %q from namespace %q: %v", chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace, err)
	}

	userAgentString := fmt.Sprintf("%s/%s/%s/%s", UserAgentPrefix, pluginDetail.Name, pluginDetail.Version, version)

	chartID := fmt.Sprintf("%s/%s", appRepo.Name, chartDetails.ChartName)
	log.Infof("fetching chart %q with user-agent %q", chartID, userAgentString)

	// Look up the cachedChart cached in our DB to populate the tarball URL
	cachedChart, err := s.manager.GetChartVersion(chartDetails.AppRepositoryResourceNamespace, chartID, chartDetails.Version)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Unable to fetch the chart %s (version %s) from the namespace %q: %v", chartID, chartDetails.Version, chartDetails.AppRepositoryResourceNamespace, err)
	}
	var tarballURL string
	if cachedChart.ChartVersions != nil && len(cachedChart.ChartVersions) == 1 && cachedChart.ChartVersions[0].URLs != nil {
		// The tarball URL will always be the first URL in the repo.chartVersions:
		// https://helm.sh/docs/topics/chart_repository/#the-index-file
		// https://github.com/helm/helm/blob/v3.7.1/cmd/helm/search/search_test.go#L63
		tarballURL = chartTarballURL(cachedChart.Repo, cachedChart.ChartVersions[0])
		log.Infof("using chart tarball url %q", tarballURL)
	}

	// Grab the chart itself
	ch, err := handlerutil.GetChart(
		&chartutils.Details{
			AppRepositoryResourceName:      appRepo.Name,
			AppRepositoryResourceNamespace: appRepo.Namespace,
			ChartName:                      chartDetails.ChartName,
			Version:                        chartDetails.Version,
			TarballURL:                     tarballURL,
		},
		appRepo,
		caCertSecret, authSecret,
		s.chartClientFactory.New(appRepo.Spec.Type, userAgentString),
	)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Unable to fetch the chart %s from the namespace %q: %v", chartDetails.ChartName, appRepo.Namespace, err)
	}

	registrySecrets, err := chartutils.RegistrySecretsPerDomain(ctx, appRepo.Spec.DockerRegistrySecrets, appRepo.Namespace, client)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Unable to fetch registry secrets from the namespace %q: %v", appRepo.Namespace, err)
	}

	return ch, registrySecrets, nil
}

func chartTarballURL(r *models.Repo, cv models.ChartVersion) string {
	source := strings.TrimSpace(cv.URLs[0])
	parsedUrl, err := url.ParseRequestURI(source)
	if err != nil || parsedUrl.Scheme == "" {
		// If the chart URL is not absolute, join with repo URL. It's fine if the
		// URL we build here is invalid as we can catch this error when actually
		// making the request
		u, _ := url.Parse(r.URL)
		u.Path = path.Join(u.Path, source)
		return u.String()
	}
	return source
}

// DeleteInstalledPackage deletes an installed package.
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *corev1.DeleteInstalledPackageRequest) (*corev1.DeleteInstalledPackageResponse, error) {
	installedRef := request.GetInstalledPackageRef()
	releaseName := installedRef.GetIdentifier()
	namespace := installedRef.GetContext().GetNamespace()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", installedRef.GetContext().GetCluster(), namespace)
	log.Infof("+helm DeleteInstalledPackage %s", contextMsg)

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	keepHistory := false
	err = agent.DeleteRelease(actionConfig, releaseName, keepHistory, s.timeoutSeconds)
	if err != nil {
		log.Errorf("error: %+v", err)
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", releaseName, namespace, err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to delete helm release %q in the namespace %q: %v", releaseName, namespace, err)
	}

	return &corev1.DeleteInstalledPackageResponse{}, nil
}

// RollbackInstalledPackage updates an installed package.
func (s *Server) RollbackInstalledPackage(ctx context.Context, request *helmv1.RollbackInstalledPackageRequest) (*helmv1.RollbackInstalledPackageResponse, error) {
	installedRef := request.GetInstalledPackageRef()
	releaseName := installedRef.GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", installedRef.GetContext().GetCluster(), installedRef.GetContext().GetNamespace())
	log.Infof("+helm RollbackInstalledPackage %s", contextMsg)

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := agent.RollbackRelease(actionConfig, releaseName, int(request.GetReleaseRevision()), s.timeoutSeconds)
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", releaseName, installedRef.GetContext().GetNamespace(), err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to rollback helm release %q in the namespace %q: %v", releaseName, installedRef.GetContext().GetNamespace(), err)
	}

	cluster := installedRef.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	return &helmv1.RollbackInstalledPackageResponse{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: release.Namespace,
			},
			Identifier: release.Name,
			Plugin:     GetPluginDetail(),
		},
	}, nil
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes
// resources created by an installed package.
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *corev1.GetInstalledPackageResourceRefsRequest) (*corev1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", pkgRef.GetContext().GetCluster(), pkgRef.GetContext().GetNamespace())
	identifier := pkgRef.GetIdentifier()
	log.Infof("+helm GetResources %s %s", contextMsg, identifier)

	namespace := pkgRef.GetContext().GetNamespace()

	actionConfig, err := s.actionConfigGetter(ctx, request.GetInstalledPackageRef().GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	// Grab the released manifest from the release.
	getcmd := action.NewGet(actionConfig)
	release, err := getcmd.Run(identifier)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", identifier, namespace, err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to run Helm get action: %v", err)
	}

	refs, err := resourceRefsFromManifest(release.Manifest)
	if err != nil {
		return nil, err
	}

	return &corev1.GetInstalledPackageResourceRefsResponse{
		Context:      pkgRef.GetContext(),
		ResourceRefs: refs,
	}, nil
}

type YAMLMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type YAMLResource struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   YAMLMetadata   `json:"metadata"`
	Items      []YAMLResource `json:"items"`
}

// resourceRefsFromManifest returns the resource refs for a given yaml manifest.
func resourceRefsFromManifest(m string) ([]*corev1.ResourceRef, error) {
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(m))
	refs := []*corev1.ResourceRef{}
	doc := YAMLResource{}
	for {
		err := decoder.Decode(&doc)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, status.Errorf(codes.Internal, "Unable to decode yaml manifest: %v", err)
		}
		if doc.Kind == "" {
			continue
		}
		if doc.Kind == "List" || doc.Kind == "RoleList" || doc.Kind == "ClusterRoleList" {
			for _, i := range doc.Items {
				refs = append(refs, &corev1.ResourceRef{
					ApiVersion: i.APIVersion,
					Kind:       i.Kind,
					Name:       i.Metadata.Name,
					Namespace:  i.Metadata.Namespace,
				})
			}
			continue
		}
		refs = append(refs, &corev1.ResourceRef{
			ApiVersion: doc.APIVersion,
			Kind:       doc.Kind,
			Name:       doc.Metadata.Name,
			Namespace:  doc.Metadata.Namespace,
		})
	}

	return refs, nil
}
