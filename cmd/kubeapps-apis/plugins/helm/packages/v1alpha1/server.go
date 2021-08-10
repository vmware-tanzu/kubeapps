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
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type clientGetter func(context.Context) (kubernetes.Interface, dynamic.Interface, error)
type helmActionConfigGetter func(ctx context.Context, namespace string) (*action.Configuration, error)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

const (
	MajorVersionsInSummary = 3
	MinorVersionsInSummary = 3
	PatchVersionsInSummary = 3
)

// Server implements the helm packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedHelmPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientGetter
	globalPackagingNamespace string
	manager                  utils.AssetManager
	actionConfigGetter       helmActionConfigGetter
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter server.KubernetesConfigGetter) *Server {
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

	return &Server{
		clientGetter: func(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
			if configGetter == nil {
				return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx)
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
		actionConfigGetter: func(ctx context.Context, namespace string) (*action.Configuration, error) {
			if configGetter == nil {
				return nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
			}

			restClientGetter := agent.NewConfigFlagsFromCluster(namespace, config)
			clientSet, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to create kubernetes client : %v", err))
			}
			// TODO(mnelson): Update to allow different helm storage options.
			storage := agent.StorageForSecrets(namespace, clientSet)
			return &action.Configuration{
				RESTClientGetter: restClientGetter,
				KubeClient:       kube.New(restClientGetter),
				Releases:         storage,
				Log:              log.Infof,
			}, nil
		},
		manager:                  manager,
		globalPackagingNamespace: kubeappsNamespace,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx)
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
	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+helm GetAvailablePackageSummaries %s", contextMsg)

	// Check the request context (namespace and cluster)
	namespace := ""
	if request.Context != nil {
		if request.Context.Cluster != "" {
			return nil, status.Errorf(codes.Unimplemented, "Not supported yet: request.Context.Cluster: [%v]", request.Context.Cluster)
		}
		namespace = request.Context.Namespace
	}
	// Check the requested namespace: if any, return "everything a user can read";
	// otherwise, first check if the user can access the requested ns
	if namespace == "" {
		// TODO(agamez): not including a namespace means that it returns everything a user can read
		return nil, status.Errorf(codes.Unimplemented, "Not supported yet: not including a namespace means that it returns everything a user can read")
	}
	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, namespace); err != nil {
		return nil, err
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
		pkg.LatestPkgVersion = chart.ChartVersions[0].Version
		pkg.LatestAppVersion = chart.ChartVersions[0].AppVersion
	}

	return pkg, nil
}

// getUnescapedChartID takes a chart id with URI-encoded characters and decode them. Ex: 'foo%2Fbar' becomes 'foo/bar'
func getUnescapedChartID(chartID string) (string, error) {
	unescapedChartID, err := url.QueryUnescape(chartID)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Unable to decode chart ID chart: %v", chartID)
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
	if request.AvailablePackageRef == nil || request.AvailablePackageRef.Context == nil {
		return nil, status.Errorf(codes.InvalidArgument, "No request AvailablePackageRef.Context provided")
	}
	contextMsg := fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.AvailablePackageRef.Context.Cluster, request.AvailablePackageRef.Context.Namespace)
	log.Infof("+helm GetAvailablePackageDetail %s", contextMsg)

	// Retrieve namespace, chartID, version from the request
	namespace := request.AvailablePackageRef.Context.Namespace
	version := request.PkgVersion

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, namespace); err != nil {
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

	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	if namespace == "" || request.GetAvailablePackageRef().GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Required context or identifier not provided")
	}
	contextMsg := fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.AvailablePackageRef.Context.Cluster, request.AvailablePackageRef.Context.Namespace)
	log.Infof("+helm GetAvailablePackageVersions %s", contextMsg)

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, namespace); err != nil {
		return nil, err
	}

	unescapedChartID, err := getUnescapedChartID(request.AvailablePackageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
	chart, err := s.manager.GetChart(namespace, unescapedChartID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve chart: %v", err)
	}
	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: packageAppVersionsSummary(chart.ChartVersions),
	}, nil
}

// packageAppVersionsSummary converts the model chart versions into the required version summary.
func packageAppVersionsSummary(versions []models.ChartVersion) []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion {
	pav := []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{}

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
			if len(version_map) >= MajorVersionsInSummary {
				continue
			}
		} else {
			// If we don't yet have this minor version
			if _, ok := version_map[version.Major()][version.Minor()]; !ok {
				// Don't add a new minor version if we already have enough for this major version
				if len(version_map[version.Major()]) >= MinorVersionsInSummary {
					continue
				}
			} else {
				if len(version_map[version.Major()][version.Minor()]) >= PatchVersionsInSummary {
					continue
				}
			}
		}

		// Include the version and update the version map.
		pav = append(pav, &corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
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
	pkg.IconUrl = chart.Icon
	pkg.Name = chart.Name
	pkg.ShortDescription = chart.Description
	pkg.Categories = []string{chart.Category}

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
		pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}
	}

	// We assume that chart.ChartVersions[0] will always contain either: the latest version or the specified version
	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.PkgVersion = chart.ChartVersions[0].Version
		pkg.AppVersion = chart.ChartVersions[0].AppVersion
		pkg.Readme = chartFiles.Readme
		pkg.DefaultValues = chartFiles.Values
		pkg.ValuesSchema = chartFiles.Schema
	}
	return pkg, nil
}

// hasAccessToNamespace returns an error if the client does not have read access to a given namespace
func (s *Server) hasAccessToNamespace(ctx context.Context, namespace string) error {
	// If checking the global namespace, allow access always
	if namespace == s.globalPackagingNamespace {
		return nil
	}
	client, _, err := s.GetClients(ctx)
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
	namespace := request.GetContext().GetNamespace()
	actionConfig, err := s.actionConfigGetter(ctx, namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}
	cmd := action.NewList(actionConfig)
	if namespace == "" {
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
	for i, r := range releases {
		installedPkgSummaries[i] = installedPkgSummaryFromRelease(r)
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
		if len(charts) == 1 && len(charts[0].ChartVersions) > 0 {
			installedPkgSummaries[i].LatestPkgVersion = charts[0].ChartVersions[0].Version
		}
		installedPkgSummaries[i].Status = &corev1.InstalledPackageStatus{
			Ready:      rel.Info.Status == release.StatusDeployed,
			Reason:     statusReasonForHelmStatus(rel.Info.Status),
			UserReason: rel.Info.Status.String(),
		}
		installedPkgSummaries[i].CurrentAppVersion = rel.Chart.Metadata.AppVersion
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
		CurrentPkgVersion: r.Chart.Metadata.Version,
		IconUrl:           r.Chart.Metadata.Icon,
		PkgDisplayName:    r.Chart.Name(),
		ShortDescription:  r.Chart.Metadata.Description,
	}
}

// GetInstalledPackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *corev1.GetInstalledPackageDetailRequest) (*corev1.GetInstalledPackageDetailResponse, error) {
	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	identifier := request.GetInstalledPackageRef().GetIdentifier()
	actionConfig, err := s.actionConfigGetter(ctx, namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	// First grab the release.
	getcmd := action.NewGet(actionConfig)
	release, err := getcmd.Run(identifier)
	if err != nil {
		log.Errorf("Error is %+v", err)
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", identifier, namespace, err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to run Helm get action: %v", err)
	}
	installedPkgDetail := installedPkgDetailFromRelease(release, request.GetInstalledPackageRef())

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
	if len(charts) == 1 {
		log.Errorf("Got chart: %+v", charts[0])
		installedPkgDetail.AvailablePackageRef = &corev1.AvailablePackageReference{
			Identifier: charts[0].ID,
			Plugin:     GetPluginDetail(),
		}
		if charts[0].Repo != nil {
			installedPkgDetail.AvailablePackageRef.Context = &corev1.Context{Namespace: charts[0].Repo.Namespace}
		}
	}

	return &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPkgDetail,
	}, nil
}

func installedPkgDetailFromRelease(r *release.Release, ref *corev1.InstalledPackageReference) *corev1.InstalledPackageDetail {
	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: ref,
		Name:                r.Name,
		PkgVersionReference: &corev1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentPkgVersion:     r.Chart.Metadata.Version,
		PostInstallationNotes: r.Info.Notes,
		Status: &corev1.InstalledPackageStatus{
			Ready:      r.Info.Status == release.StatusDeployed,
			Reason:     statusReasonForHelmStatus(r.Info.Status),
			UserReason: r.Info.Status.String(),
		},
	}
}
