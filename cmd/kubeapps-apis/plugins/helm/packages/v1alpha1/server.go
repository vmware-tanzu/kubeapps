// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	assetmanager "github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pkghelmv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	paginate "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	pkgutils "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	resourcerefs "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	helmagent "github.com/kubeapps/kubeapps/pkg/agent"
	chartutils "github.com/kubeapps/kubeapps/pkg/chart"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	dbutils "github.com/kubeapps/kubeapps/pkg/dbutils"
	handlerutil "github.com/kubeapps/kubeapps/pkg/handlerutil"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8sauthorizationv1 "k8s.io/api/authorization/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8stypedclient "k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type helmActionConfigGetter func(ctx context.Context, pkgContext *pkgsGRPCv1alpha1.Context) (*helmaction.Configuration, error)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ pkgsGRPCv1alpha1.PackagesServiceServer = (*Server)(nil)

const (
	UserAgentPrefix             = "kubeapps-apis/plugins"
	DefaultTimeoutSeconds int32 = 300
)

type createRelease func(*helmaction.Configuration, string, string, string, *helmchart.Chart, map[string]string, int32) (*helmrelease.Release, error)

// Server implements the helm packages v1alpha1 interface.
type Server struct {
	pkghelmv1alpha1.UnimplementedHelmPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientgetter.ClientGetterFunc
	globalPackagingNamespace string
	globalPackagingCluster   string
	manager                  assetmanager.AssetManager
	actionConfigGetter       helmActionConfigGetter
	chartClientFactory       chartutils.ChartClientFactoryInterface
	versionsInSummary        pkgutils.VersionsInSummary
	timeoutSeconds           int32
	createReleaseFunc        createRelease
}

// parsePluginConfig parses the input plugin configuration json file and return the configuration options.
func parsePluginConfig(pluginConfigPath string) (pkgutils.VersionsInSummary, int32, error) {

	// Note at present VersionsInSummary is the only configurable option for this plugin,
	// and if required this func can be enhaned to return helmConfig struct

	// In the helm plugin, for example, we are interested in config for the
	// apiscore.packages.v1alpha1 only. So the plugin defines the following struct and parses the config.
	type helmConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary pkgutils.VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`
	}
	var config helmConfig

	pluginConfig, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return pkgutils.VersionsInSummary{}, 0, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return pkgutils.VersionsInSummary{}, 0, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	// return configured value
	return config.Core.Packages.V1alpha1.VersionsInSummary, config.Core.Packages.V1alpha1.TimeoutSeconds, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter apiscore.KubernetesConfigGetter, globalPackagingCluster string, globalReposNamespace string, pluginConfigPath string) *Server {
	var ASSET_SYNCER_DB_URL = os.Getenv("ASSET_SYNCER_DB_URL")
	var ASSET_SYNCER_DB_NAME = os.Getenv("ASSET_SYNCER_DB_NAME")
	var ASSET_SYNCER_DB_USERNAME = os.Getenv("ASSET_SYNCER_DB_USERNAME")
	var ASSET_SYNCER_DB_USERPASSWORD = os.Getenv("ASSET_SYNCER_DB_USERPASSWORD")

	var dbConfig = dbutils.Config{URL: ASSET_SYNCER_DB_URL, Database: ASSET_SYNCER_DB_NAME, Username: ASSET_SYNCER_DB_USERNAME, Password: ASSET_SYNCER_DB_USERPASSWORD}

	manager, err := assetmanager.NewPGManager(dbConfig, globalReposNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}

	// If no config is provided, we default to the existing values for backwards
	// compatibility.
	versionsInSummary := pkgutils.GetDefaultVersionsInSummary()
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
		clientGetter: clientgetter.NewClientGetter(configGetter),
		actionConfigGetter: func(ctx context.Context, pkgContext *pkgsGRPCv1alpha1.Context) (*helmaction.Configuration, error) {
			cluster := pkgContext.GetCluster()
			// Don't force clients to send a cluster unless we are sure all use-cases
			// of kubeapps-api are multicluster.
			if cluster == "" {
				cluster = globalPackagingCluster
			}
			fn := clientgetter.NewHelmActionConfigGetter(configGetter, cluster)
			return fn(ctx, pkgContext.GetNamespace())
		},
		manager:                  manager,
		globalPackagingNamespace: globalReposNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		chartClientFactory:       &chartutils.ChartClientFactory{},
		versionsInSummary:        versionsInSummary,
		timeoutSeconds:           timeoutSeconds,
		createReleaseFunc:        helmagent.CreateRelease,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context, cluster string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}

// GetManager ensures a manager is available and returns it.
func (s *Server) GetManager() (assetmanager.AssetManager, error) {
	if s.manager == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "server not configured with manager")
	}
	manager := s.manager
	return manager, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageSummaries %s", contextMsg)

	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()

	// Check the requested namespace: if any, return "everything a user can read";
	// otherwise, first check if the user can access the requested ns
	if namespace == "" {
		// TODO(agamez): not including a namespace means that it returns everything a user can read
		return nil, grpcstatus.Errorf(grpccodes.Unimplemented, "Not supported yet: not including a namespace means that it returns everything a user can read")
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
	cq := assetmanager.ChartQuery{
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
	pageOffset, err := paginate.PageOffsetFromAvailableRequest(request)
	if err != nil {
		return nil, err
	}

	// This plugin will include, as part of the GetAvailablePackageSummariesResponse,
	// a "Categories" field containing only the distinct category names considering just the namespace
	chartCategories, err := s.manager.GetAllChartCategories(assetmanager.ChartQuery{Namespace: namespace})
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to fetch chart categories: %v", err)
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
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to retrieve charts: %v", err)
	}

	// Convert the charts response into a GetAvailablePackageSummariesResponse
	responsePackages := []*pkgsGRPCv1alpha1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg, err := pkgutils.AvailablePackageSummaryFromChart(chart, GetPluginDetail())
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to parse chart to an AvailablePackageSummary: %v", err)
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
	return &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: responsePackages,
		NextPageToken:             nextPageToken,
		Categories:                categories,
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageDetail %s", contextMsg)

	if request.GetAvailablePackageRef().GetContext() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "No request AvailablePackageRef.Context provided")
	}

	// Retrieve namespace, cluster, version from the request
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	version := request.PkgVersion

	// Currently we support available packages on the kubeapps cluster only.
	if cluster != "" && cluster != s.globalPackagingCluster {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Requests for available packages on clusters other than %q not supported. Requested cluster was: %q", s.globalPackagingCluster, cluster)
	}

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, cluster, namespace); err != nil {
		return nil, err
	}

	unescapedChartID, err := pkgutils.GetUnescapedChartID(request.AvailablePackageRef.Identifier)
	if err != nil {
		return nil, err
	}

	// Since the version is optional, in case of an empty one, fall back to get all versions and get the first one
	// Note that the chart version array should be ordered
	var chart chartmodels.Chart
	if version == "" {
		log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
		chart, err = s.manager.GetChart(namespace, unescapedChartID)
	} else {
		log.Infof("Requesting chart '%s' (version %s) in ns '%s'", unescapedChartID, version, namespace)
		chart, err = s.manager.GetChartVersion(namespace, unescapedChartID, version)
	}
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to retrieve chart: %v", err)
	}

	if len(chart.ChartVersions) == 0 {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Chart returned without any versions: %+v", chart)
	}
	if version == "" {
		version = chart.ChartVersions[0].Version
	}
	fileID := fileIDForChart(unescapedChartID, chart.ChartVersions[0].Version)
	chartFiles, err := s.manager.GetChartFiles(namespace, fileID)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to retrieve chart files: %v", err)
	}

	availablePackageDetail, err := AvailablePackageDetailFromChart(&chart, &chartFiles)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to parse chart to an availablePackageDetail: %v", err)
	}

	return &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: availablePackageDetail,
	}, nil
}

// fileIDForChart returns a file ID given a chart id and version.
func fileIDForChart(id, version string) string {
	return fmt.Sprintf("%s-%s", id, version)
}

// GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetAvailablePackageVersions %s", contextMsg)

	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	if namespace == "" || request.GetAvailablePackageRef().GetIdentifier() == "" {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Required context or identifier not provided")
	}
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	// Currently we support available packages on the kubeapps cluster only.
	if cluster != "" && cluster != s.globalPackagingCluster {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Requests for versions of available packages on clusters other than %q not supported. Requested cluster was %q.", s.globalPackagingCluster, cluster)
	}

	// After requesting a specific namespace, we have to ensure the user can actually access to it
	if err := s.hasAccessToNamespace(ctx, cluster, namespace); err != nil {
		return nil, err
	}

	unescapedChartID, err := pkgutils.GetUnescapedChartID(request.GetAvailablePackageRef().GetIdentifier())
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
	chart, err := s.manager.GetChart(namespace, unescapedChartID)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to retrieve chart: %v", err)
	}

	return &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: pkgutils.PackageAppVersionsSummary(chart.ChartVersions, s.versionsInSummary),
	}, nil
}

// AvailablePackageDetailFromChart builds an AvailablePackageDetail from a Chart
func AvailablePackageDetailFromChart(chart *chartmodels.Chart, chartFiles *chartmodels.ChartFiles) (*pkgsGRPCv1alpha1.AvailablePackageDetail, error) {
	pkg := &pkgsGRPCv1alpha1.AvailablePackageDetail{}

	isValid, err := pkgutils.IsValidChart(chart)
	if !isValid || err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.DisplayName = chart.Name
	pkg.HomeUrl = chart.Home
	pkg.IconUrl = chart.Icon
	pkg.Name = chart.Name
	pkg.ShortDescription = chart.Description
	pkg.Categories = []string{chart.Category}
	pkg.SourceUrls = chart.Sources

	pkg.Maintainers = []*pkgsGRPCv1alpha1.Maintainer{}
	for _, maintainer := range chart.Maintainers {
		m := &pkgsGRPCv1alpha1.Maintainer{Name: maintainer.Name, Email: maintainer.Email}
		pkg.Maintainers = append(pkg.Maintainers, m)
	}

	pkg.AvailablePackageRef = &pkgsGRPCv1alpha1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     GetPluginDetail(),
	}
	if chart.Repo != nil {
		pkg.RepoUrl = chart.Repo.URL
		pkg.AvailablePackageRef.Context = &pkgsGRPCv1alpha1.Context{Namespace: chart.Repo.Namespace}
	}

	// We assume that helmchart.ChartVersions[0] will always contain either: the latest version or the specified version
	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.Version = &pkgsGRPCv1alpha1.PackageAppVersion{
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

	res, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(context.TODO(), &k8sauthorizationv1.SelfSubjectAccessReview{
		Spec: k8sauthorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &k8sauthorizationv1.ResourceAttributes{
				Group:     "",
				Resource:  "secrets",
				Verb:      "get",
				Namespace: namespace,
			},
		},
	}, k8smetav1.CreateOptions{})
	if err != nil {
		return grpcstatus.Errorf(grpccodes.Internal, "Unable to check if the user has access to the namespace: %s", err)
	}
	if !res.Status.Allowed {
		// If the user has not access, return a unauthenticated response, otherwise, continue
		return grpcstatus.Errorf(grpccodes.Unauthenticated, "The current user has no access to the namespace %q", namespace)
	}
	return nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+helm GetInstalledPackageSummaries %s", contextMsg)

	actionConfig, err := s.actionConfigGetter(ctx, request.GetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}
	cmd := helmaction.NewList(actionConfig)
	if request.GetContext().GetNamespace() == "" {
		cmd.AllNamespaces = true
	}

	cmd.Limit = int(request.GetPaginationOptions().GetPageSize())
	cmd.Offset, err = paginate.PageOffsetFromInstalledRequest(request)
	if err != nil {
		return nil, err
	}

	// TODO(mnelson): Check whether we need to support ListAll (status == "all" in existing helm support)

	releases, err := cmd.Run()
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to run Helm List action: %v", err)
	}

	installedPkgSummaries := make([]*pkgsGRPCv1alpha1.InstalledPackageSummary, len(releases))
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
	// querying per helmrelease.
	for i, rel := range releases {
		// Helm does not store a back-reference to the chart used to create a
		// release (https://github.com/helm/helm/issues/6464), so for each
		// release, we look up a chart with than name and version available in
		// the release namespace, and if one is found, pull out the latest chart
		// version.
		cq := assetmanager.ChartQuery{
			Namespace:  rel.Namespace,
			ChartName:  rel.Chart.Metadata.Name,
			Version:    rel.Chart.Metadata.Version,
			AppVersion: rel.Chart.Metadata.AppVersion,
		}
		charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, 1, 0)
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Error while fetching related charts: %v", err)
		}
		// TODO(agamez): deal with multiple matches, perhaps returning []AvailablePackageRef ?
		// Example: global + namespaced repo including an overlapping subset of packages.
		if len(charts) > 0 && len(charts[0].ChartVersions) > 0 {
			installedPkgSummaries[i].LatestVersion = &pkgsGRPCv1alpha1.PackageAppVersion{
				PkgVersion: charts[0].ChartVersions[0].Version,
				AppVersion: charts[0].ChartVersions[0].AppVersion,
			}
		}
		installedPkgSummaries[i].Status = &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      rel.Info.Status == helmrelease.StatusDeployed,
			Reason:     statusReasonForHelmStatus(rel.Info.Status),
			UserReason: rel.Info.Status.String(),
		}
		installedPkgSummaries[i].CurrentVersion = &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: rel.Chart.Metadata.Version,
			AppVersion: rel.Chart.Metadata.AppVersion,
		}
	}

	response := &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummaries,
	}
	if len(releases) == cmd.Limit {
		response.NextPageToken = fmt.Sprintf("%d", cmd.Limit+1)
	}
	return response, nil
}

func statusReasonForHelmStatus(s helmrelease.Status) pkgsGRPCv1alpha1.InstalledPackageStatus_StatusReason {
	switch s {
	case helmrelease.StatusDeployed:
		return pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED
	case helmrelease.StatusFailed:
		return pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED
	case helmrelease.StatusPendingInstall, helmrelease.StatusPendingRollback, helmrelease.StatusPendingUpgrade, helmrelease.StatusUninstalling:
		return pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING
	}
	// Both StatusUninstalled and StatusSuperseded will be unknown/unspecified.
	return pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED
}

func installedPkgSummaryFromRelease(r *helmrelease.Release) *pkgsGRPCv1alpha1.InstalledPackageSummary {
	return &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Namespace: r.Namespace,
			},
			Identifier: r.Name,
		},
		Name: r.Name,
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
		},
		IconUrl:          r.Chart.Metadata.Icon,
		PkgDisplayName:   r.Chart.Name(),
		ShortDescription: r.Chart.Metadata.Description,
	}
}

// GetInstalledPackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetInstalledPackageRef().GetContext().GetCluster(), request.GetInstalledPackageRef().GetContext().GetNamespace())
	log.Infof("+helm GetInstalledPackageDetail %s", contextMsg)

	namespace := request.GetInstalledPackageRef().GetContext().GetNamespace()
	identifier := request.GetInstalledPackageRef().GetIdentifier()

	actionConfig, err := s.actionConfigGetter(ctx, request.GetInstalledPackageRef().GetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	// First grab the helmrelease.
	getcmd := helmaction.NewGet(actionConfig)
	release, err := getcmd.Run(identifier)
	if err != nil {
		if err == helmstoragedriver.ErrReleaseNotFound {
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", identifier, namespace, err)
		}
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to run Helm get action: %v", err)
	}
	installedPkgDetail, err := installedPkgDetailFromRelease(release, request.GetInstalledPackageRef())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create installed package detail from release: %v", err)
	}

	// Grab the released values.
	valuescmd := helmaction.NewGetValues(actionConfig)
	values, err := valuescmd.Run(identifier)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get Helm release values: %v", err)
	}
	valuesMarshalled, err := json.Marshal(values)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to marshal Helm release values: %v", err)
	}
	installedPkgDetail.ValuesApplied = string(valuesMarshalled)

	// Check for a chart matching the installed package.
	cq := assetmanager.ChartQuery{
		Namespace:  release.Namespace,
		ChartName:  release.Chart.Metadata.Name,
		Version:    release.Chart.Metadata.Version,
		AppVersion: release.Chart.Metadata.AppVersion,
	}
	charts, _, err := s.manager.GetPaginatedChartListWithFilters(cq, 1, 0)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Error while fetching related chart: %v", err)
	}
	// TODO(agamez): deal with multiple matches, perhaps returning []AvailablePackageRef ?
	// Example: global + namespaced repo including an overlapping subset of packages.
	if len(charts) > 0 {
		installedPkgDetail.AvailablePackageRef = &pkgsGRPCv1alpha1.AvailablePackageReference{
			Identifier: charts[0].ID,
			Plugin:     GetPluginDetail(),
		}
		if charts[0].Repo != nil {
			installedPkgDetail.AvailablePackageRef.Context = &pkgsGRPCv1alpha1.Context{
				Namespace: charts[0].Repo.Namespace,
				Cluster:   s.globalPackagingCluster,
			}
		}
		if len(charts[0].ChartVersions) > 0 {
			cv := charts[0].ChartVersions[0]
			installedPkgDetail.LatestVersion = &pkgsGRPCv1alpha1.PackageAppVersion{
				PkgVersion: cv.Version,
				AppVersion: cv.AppVersion,
			}
		}
	}

	return &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: installedPkgDetail,
	}, nil
}

func installedPkgDetailFromRelease(r *helmrelease.Release, ref *pkgsGRPCv1alpha1.InstalledPackageReference) (*pkgsGRPCv1alpha1.InstalledPackageDetail, error) {
	customDetailHelm, err := anypb.New(&pkghelmv1alpha1.InstalledPackageDetailCustomDataHelm{
		ReleaseRevision: int32(r.Version),
	})
	if err != nil {
		return nil, err
	}
	return &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: ref,
		Name:                r.Name,
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: r.Chart.Metadata.Version,
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: r.Chart.Metadata.Version,
			AppVersion: r.Chart.Metadata.AppVersion,
		},
		PostInstallationNotes: r.Info.Notes,
		Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
			Ready:      r.Info.Status == helmrelease.StatusDeployed,
			Reason:     statusReasonForHelmStatus(r.Info.Status),
			UserReason: r.Info.Status.String(),
		},
		CustomDetail: customDetailHelm,
	}, nil
}

// CreateInstalledPackage creates an installed package.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.CreateInstalledPackageRequest) (*pkgsGRPCv1alpha1.CreateInstalledPackageResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetTargetContext().GetCluster(), request.GetTargetContext().GetNamespace())
	log.Infof("+helm CreateInstalledPackage %s", contextMsg)

	typedClient, _, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := request.GetAvailablePackageRef().GetIdentifier()
	repoNamespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	repoName, chartName, err := pkgutils.SplitChartIdentifier(chartID)
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
		return nil, grpcstatus.Errorf(grpccodes.Unauthenticated, "Missing permissions %v", err)
	}

	// Create an action config for the target namespace.
	actionConfig, err := s.actionConfigGetter(ctx, request.GetTargetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := s.createReleaseFunc(actionConfig, request.GetName(), request.GetTargetContext().GetNamespace(), request.GetValues(), ch, registrySecrets, s.timeoutSeconds)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create helm release %q in the namespace %q: %v", request.GetName(), request.GetTargetContext().GetNamespace(), err)
	}

	cluster := request.GetTargetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	return &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Cluster:   cluster,
				Namespace: release.Namespace,
			},
			Identifier: release.Name,
			Plugin:     GetPluginDetail(),
		},
	}, nil
}

// UpdateInstalledPackage updates an installed package.
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest) (*pkgsGRPCv1alpha1.UpdateInstalledPackageResponse, error) {
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
	detailResponse, err := s.GetInstalledPackageDetail(ctx, &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
		InstalledPackageRef: installedRef,
	})
	if err != nil {
		return nil, err
	}

	availablePkgRef := detailResponse.GetInstalledPackageDetail().GetAvailablePackageRef()
	if availablePkgRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "Unable to find the available package used to deploy %q in the namespace %q.", releaseName, installedRef.GetContext().GetNamespace())
	}

	typedClient, _, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := availablePkgRef.GetIdentifier()
	repoName, chartName, err := pkgutils.SplitChartIdentifier(chartID)
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
		return nil, grpcstatus.Errorf(grpccodes.Unauthenticated, "Missing permissions %v", err)
	}

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := helmagent.UpgradeRelease(actionConfig, releaseName, request.GetValues(), ch, registrySecrets, s.timeoutSeconds)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to upgrade helm release %q in the namespace %q: %v", releaseName, installedRef.GetContext().GetNamespace(), err)
	}

	cluster := installedRef.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	return &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
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
func (s *Server) getAppRepoAndRelatedSecrets(ctx context.Context, appRepoName, appRepoNamespace string) (*apprepov1alpha1.AppRepository, *k8scorev1.Secret, *k8scorev1.Secret, error) {

	// We currently get app repositories on the kubeapps cluster only.
	typedClient, dynClient, err := s.GetClients(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, nil, nil, err
	}

	// Using the dynamic client to get the AppRepository rather than updating the interface
	// returned by the client getter, then converting the result back to a typed AppRepository
	// since our existing code that we're re-using expects one.
	gvr := k8sschema.GroupVersionResource{
		Group:    "kubeapps.com",
		Version:  "v1alpha1",
		Resource: "apprepositories",
	}
	appRepoUnstructured, err := dynClient.Resource(gvr).Namespace(appRepoNamespace).Get(ctx, appRepoName, k8smetav1.GetOptions{})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unable to get app repository %s/%s: %v", appRepoNamespace, appRepoName, err)
	}

	var appRepo apprepov1alpha1.AppRepository
	err = k8sruntime.DefaultUnstructuredConverter.FromUnstructured(appRepoUnstructured.UnstructuredContent(), &appRepo)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to convert unstructured AppRepository for %s/%s to a structured AppRepository: %v", appRepoNamespace, appRepoName, err)
	}

	auth := appRepo.Spec.Auth
	var caCertSecret *k8scorev1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		caCertSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, k8smetav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *k8scorev1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, k8smetav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", secretName, err)
		}
	}

	return &appRepo, caCertSecret, authSecret, nil
}

// fetchChartWithRegistrySecrets returns the chart and related registry secrets.
//
// Mainly to DRY up similar code in the create and update methods.
func (s *Server) fetchChartWithRegistrySecrets(ctx context.Context, chartDetails *chartutils.Details, client k8stypedclient.Interface) (*helmchart.Chart, map[string]string, error) {
	// Most of the existing code that we want to reuse is based on having a typed AppRepository.
	appRepo, caCertSecret, authSecret, err := s.getAppRepoAndRelatedSecrets(ctx, chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace)
	if err != nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to fetch app repo %q from namespace %q: %v", chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace, err)
	}

	userAgentString := fmt.Sprintf("%s/%s/%s/%s", UserAgentPrefix, pluginDetail.Name, pluginDetail.Version, version)

	chartID := fmt.Sprintf("%s/%s", appRepo.Name, chartDetails.ChartName)
	log.Infof("fetching chart %q with user-agent %q", chartID, userAgentString)

	// Look up the cachedChart cached in our DB to populate the tarball URL
	cachedChart, err := s.manager.GetChartVersion(chartDetails.AppRepositoryResourceNamespace, chartID, chartDetails.Version)
	if err != nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to fetch the chart %s (version %s) from the namespace %q: %v", chartID, chartDetails.Version, chartDetails.AppRepositoryResourceNamespace, err)
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
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to fetch the chart %s from the namespace %q: %v", chartDetails.ChartName, appRepo.Namespace, err)
	}

	registrySecrets, err := chartutils.RegistrySecretsPerDomain(ctx, appRepo.Spec.DockerRegistrySecrets, appRepo.Namespace, client)
	if err != nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to fetch registry secrets from the namespace %q: %v", appRepo.Namespace, err)
	}

	return ch, registrySecrets, nil
}

func chartTarballURL(r *chartmodels.Repo, cv chartmodels.ChartVersion) string {
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
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest) (*pkgsGRPCv1alpha1.DeleteInstalledPackageResponse, error) {
	installedRef := request.GetInstalledPackageRef()
	releaseName := installedRef.GetIdentifier()
	namespace := installedRef.GetContext().GetNamespace()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", installedRef.GetContext().GetCluster(), namespace)
	log.Infof("+helm DeleteInstalledPackage %s", contextMsg)

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	keepHistory := false
	err = helmagent.DeleteRelease(actionConfig, releaseName, keepHistory, s.timeoutSeconds)
	if err != nil {
		log.Errorf("error: %+v", err)
		if errors.Is(err, helmstoragedriver.ErrReleaseNotFound) {
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", releaseName, namespace, err)
		}
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to delete helm release %q in the namespace %q: %v", releaseName, namespace, err)
	}

	return &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{}, nil
}

// RollbackInstalledPackage updates an installed package.
func (s *Server) RollbackInstalledPackage(ctx context.Context, request *pkghelmv1alpha1.RollbackInstalledPackageRequest) (*pkghelmv1alpha1.RollbackInstalledPackageResponse, error) {
	installedRef := request.GetInstalledPackageRef()
	releaseName := installedRef.GetIdentifier()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", installedRef.GetContext().GetCluster(), installedRef.GetContext().GetNamespace())
	log.Infof("+helm RollbackInstalledPackage %s", contextMsg)

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := helmagent.RollbackRelease(actionConfig, releaseName, int(request.GetReleaseRevision()), s.timeoutSeconds)
	if err != nil {
		if errors.Is(err, helmstoragedriver.ErrReleaseNotFound) {
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", releaseName, installedRef.GetContext().GetNamespace(), err)
		}
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to rollback helm release %q in the namespace %q: %v", releaseName, installedRef.GetContext().GetNamespace(), err)
	}

	cluster := installedRef.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	return &pkghelmv1alpha1.RollbackInstalledPackageResponse{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
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
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", pkgRef.GetContext().GetCluster(), pkgRef.GetContext().GetNamespace())
	identifier := pkgRef.GetIdentifier()
	log.Infof("+helm GetInstalledPackageResourceRefs %s %s", contextMsg, identifier)

	fn := func(ctx context.Context, namespace string) (*helmaction.Configuration, error) {
		actionGetter, err := s.actionConfigGetter(ctx, pkgRef.GetContext())
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
		}
		return actionGetter, nil
	}
	return resourcerefs.GetInstalledPackageResourceRefs(ctx, request, fn)
}
