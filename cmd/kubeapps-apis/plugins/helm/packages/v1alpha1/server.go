// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	helmv1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/helm/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/helm/packages/v1alpha1/utils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/agent"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	chartutils "github.com/vmware-tanzu/kubeapps/pkg/chart"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	authorizationv1 "k8s.io/api/authorization/v1"
	corek8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type helmActionConfigGetter func(ctx context.Context, pkgContext *corev1.Context) (*action.Configuration, error)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

const (
	UserAgentPrefix = "kubeapps-apis/plugins"
)

type createRelease func(*action.Configuration, string, string, string, *chart.Chart, map[string]string, int32) (*release.Release, error)
type newRepoClient func(appRepo *appRepov1.AppRepository, secret *corek8sv1.Secret) (httpclient.Client, error)

// Server implements the helm packages v1alpha1 interface.
type Server struct {
	helmv1.UnimplementedHelmPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientgetter.ClientProviderInterface
	globalPackagingNamespace string
	globalPackagingCluster   string
	manager                  utils.AssetManager
	actionConfigGetter       helmActionConfigGetter
	chartClientFactory       chartutils.ChartClientFactoryInterface
	createReleaseFunc        createRelease
	kubeappsCluster          string // Specifies the cluster on which Kubeapps is installed.
	kubeappsNamespace        string // Namespace in which Kubeapps is installed
	pluginConfig             *common.HelmPluginConfig
	repoClientGetter         newRepoClient
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster string, globalPackagingNamespace string, pluginConfigPath string) *Server {
	var ASSET_SYNCER_DB_URL = os.Getenv("ASSET_SYNCER_DB_URL")
	var ASSET_SYNCER_DB_NAME = os.Getenv("ASSET_SYNCER_DB_NAME")
	var ASSET_SYNCER_DB_USERNAME = os.Getenv("ASSET_SYNCER_DB_USERNAME")
	var ASSET_SYNCER_DB_USERPASSWORD = os.Getenv("ASSET_SYNCER_DB_USERPASSWORD")

	// Namespace where Kubeapps is installed to be in line with the asset syncer
	kubeappsNamespace := os.Getenv("POD_NAMESPACE")
	if kubeappsNamespace == "" {
		log.Fatal("POD_NAMESPACE environment variable should be defined")
	}

	var dbConfig = dbutils.Config{URL: ASSET_SYNCER_DB_URL, Database: ASSET_SYNCER_DB_NAME, Username: ASSET_SYNCER_DB_USERNAME, Password: ASSET_SYNCER_DB_USERPASSWORD}

	log.Infof("+helm NewServer(globalPackagingCluster: [%v], globalPackagingNamespace: [%v], pluginConfigPath: [%s]",
		globalPackagingCluster, globalPackagingNamespace, pluginConfigPath)

	// If no config is provided, we default to the existing values for backwards
	// compatibility.
	pluginConfig := common.NewDefaultPluginConfig()
	if pluginConfigPath != "" {
		pluginConfig, err := common.ParsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+helm using custom config: [%v]", *pluginConfig)
	} else {
		log.Info("+helm using default config since pluginConfigPath is empty")
	}

	// TODO(agamez): currently, globalPackagingNamespace and pluginConfig.GlobalPackagingNamespace always match, but we might stop passing the config via CLI args in the future
	// and we will want to use the one from the pluginConfig
	effectiveGlobalPackagingNamespace := globalPackagingNamespace
	if pluginConfig.GlobalPackagingNamespace != "" {
		effectiveGlobalPackagingNamespace = pluginConfig.GlobalPackagingNamespace
	}

	log.Infof("+helm NewServer effective globalPackagingNamespace: [%v]", effectiveGlobalPackagingNamespace)

	manager, err := utils.NewPGManager(dbConfig, effectiveGlobalPackagingNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	err = manager.Init()
	if err != nil {
		log.Fatalf("%s", err)
	}

	// Register custom scheme
	scheme := runtime.NewScheme()
	err = appRepov1.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("%s", err)
	}

	clientProvider, err := clientgetter.NewClientProvider(configGetter, clientgetter.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("%s", err)
	}

	return &Server{
		clientGetter: clientProvider,
		actionConfigGetter: func(ctx context.Context, pkgContext *corev1.Context) (*action.Configuration, error) {
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
		kubeappsNamespace:        kubeappsNamespace,
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		chartClientFactory:       &chartutils.ChartClientFactory{},
		pluginConfig:             pluginConfig,
		createReleaseFunc:        agent.CreateRelease,
		repoClientGetter:         newRepositoryClient,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}

	// Usually only one of the clients is used for a given scenario,
	// but with this we keep backwards compatibility
	clients, err := s.clientGetter.GetClients(ctx, cluster)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get clients : %v", err))
	}

	dynamicClient, err := clients.Dynamic()
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	typedClient, err := clients.Typed()
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

// GetGlobalPackagingNamespace returns the configured global packaging namespace in in the plugin config if any,
// otherwise it uses the one passed as a cmd argument to the kubeapps-apis server for backwards compatibilty.
func (s *Server) GetGlobalPackagingNamespace() string {
	if s.pluginConfig.GlobalPackagingNamespace != "" {
		return s.pluginConfig.GlobalPackagingNamespace
	} else {
		return s.globalPackagingNamespace
	}
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.InfoS("+helm GetAvailablePackageSummaries", "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

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
		namespace = s.GetGlobalPackagingNamespace()
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
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
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

	charts, err := s.manager.GetPaginatedChartListWithFilters(cq, itemOffset, int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve charts: %v", err)
	}

	// Convert the charts response into a GetAvailablePackageSummariesResponse
	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg, err := pkgutils.AvailablePackageSummaryFromChart(chart, GetPluginDetail())
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
		nextPageToken = fmt.Sprintf("%d", itemOffset+int(pageSize))
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: responsePackages,
		NextPageToken:             nextPageToken,
		Categories:                categories,
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'helm' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	log.InfoS("+helm GetAvailablePackageDetail", "cluster", request.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.GetAvailablePackageRef().GetContext().GetNamespace())

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

	unescapedChartID, err := pkgutils.GetUnescapedPackageID(request.AvailablePackageRef.Identifier)
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
	fileID := fileIDForChart(unescapedChartID, version)
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
	log.InfoS("+helm GetAvailablePackageVersions", "cluster", request.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.GetAvailablePackageRef().GetContext().GetNamespace())

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

	unescapedChartID, err := pkgutils.GetUnescapedPackageID(request.GetAvailablePackageRef().GetIdentifier())
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart '%s' (latest version) in ns '%s'", unescapedChartID, namespace)
	chart, err := s.manager.GetChart(namespace, unescapedChartID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to retrieve chart: %v", err)
	}

	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: pkgutils.PackageAppVersionsSummary(chart.ChartVersions, s.pluginConfig.VersionsInSummary),
	}, nil
}

// AvailablePackageDetailFromChart builds an AvailablePackageDetail from a Chart
func AvailablePackageDetailFromChart(chart *models.Chart, chartFiles *models.ChartFiles) (*corev1.AvailablePackageDetail, error) {
	pkg := &corev1.AvailablePackageDetail{}

	isValid, err := pkgutils.IsValidChart(chart)
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
	if namespace == s.GetGlobalPackagingNamespace() {
		return nil
	}
	client, err := s.clientGetter.Typed(ctx, cluster)
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
		return status.Errorf(codes.PermissionDenied, "The current user has no access to the namespace %q", namespace)
	}
	return nil
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'helm' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	log.InfoS("+helm GetInstalledPackageSummaries", "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

	actionConfig, err := s.actionConfigGetter(ctx, request.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}
	cmd := action.NewList(actionConfig)
	if request.GetContext().GetNamespace() == "" {
		cmd.AllNamespaces = true
	}

	cmd.Limit = int(request.GetPaginationOptions().GetPageSize())
	cmd.Offset, err = paginate.ItemOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
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
		installedPkgSummaries[i].InstalledPackageRef.Plugin = GetPluginDetail()
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
		charts, err := s.manager.GetPaginatedChartListWithFilters(cq, 0, 0)
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
	if cmd.Limit > 0 && len(releases) == cmd.Limit {
		response.NextPageToken = fmt.Sprintf("%d", cmd.Offset+cmd.Limit)
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
	log.InfoS("+helm GetInstalledPackageDetail", "cluster", request.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.GetInstalledPackageRef().GetContext().GetNamespace())

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
	charts, err := s.manager.GetPaginatedChartListWithFilters(cq, 0, 0)
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

// CreateInstalledPackage creates an installed package.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	log.InfoS("+helm CreateInstalledPackage", "cluster", request.GetTargetContext().GetCluster(), "namespace", request.GetTargetContext().GetNamespace())

	typedClient, err := s.clientGetter.Typed(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := request.GetAvailablePackageRef().GetIdentifier()
	repoNamespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	repoName, chartName, err := pkgutils.SplitPackageIdentifier(chartID)
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
		return nil, status.Errorf(codes.PermissionDenied, "Missing permissions %v", err)
	}

	// Create an action config for the target namespace.
	actionConfig, err := s.actionConfigGetter(ctx, request.GetTargetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := s.createReleaseFunc(actionConfig, request.GetName(), request.GetTargetContext().GetNamespace(), request.GetValues(), ch, registrySecrets, s.pluginConfig.TimeoutSeconds)
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
	log.InfoS("+helm UpdateInstalledPackage", "cluster", installedRef.GetContext().GetCluster(), "namespace", installedRef.GetContext().GetNamespace())

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

	typedClient, err := s.clientGetter.Typed(ctx, s.globalPackagingCluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create kubernetes clientset: %v", err)
	}
	chartID := availablePkgRef.GetIdentifier()
	repoName, chartName, err := pkgutils.SplitPackageIdentifier(chartID)
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
		return nil, status.Errorf(codes.PermissionDenied, "Missing permissions %v", err)
	}

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := agent.UpgradeRelease(actionConfig, releaseName, request.GetValues(), ch, registrySecrets, s.pluginConfig.TimeoutSeconds)
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

// getAppRepoAndRelatedSecrets retrieves the given repo from its cluster and namespace
func (s *Server) getAppRepoAndRelatedSecrets(ctx context.Context, cluster, appRepoName, appRepoNamespace string) (*appRepov1.AppRepository, *corek8sv1.Secret, *corek8sv1.Secret, *corek8sv1.Secret, error) {

	// We currently get app repositories on the kubeapps cluster only.
	typedClient, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, nil, nil, nil, err
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
		return nil, nil, nil, nil, fmt.Errorf("unable to get app repository %s/%s: %v", appRepoNamespace, appRepoName, err)
	}

	var appRepo appRepov1.AppRepository
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(appRepoUnstructured.UnstructuredContent(), &appRepo)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to convert unstructured AppRepository for %s/%s to a structured AppRepository: %v", appRepoNamespace, appRepoName, err)
	}

	auth := appRepo.Spec.Auth
	var caCertSecret *corek8sv1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		caCertSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *corek8sv1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unable to read secret %q: %v", secretName, err)
		}
	}

	var imagesPullSecret *corek8sv1.Secret
	if len(appRepo.Spec.DockerRegistrySecrets) > 0 {
		secretName := appRepo.Spec.DockerRegistrySecrets[0]
		imagesPullSecret, err = typedClient.CoreV1().Secrets(appRepoNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("unable to read images pull secret %q: %v", secretName, err)
		}
	}

	return &appRepo, caCertSecret, authSecret, imagesPullSecret, nil
}

// fetchChartWithRegistrySecrets returns the chart and related registry secrets.
//
// Mainly to DRY up similar code in the create and update methods.
func (s *Server) fetchChartWithRegistrySecrets(ctx context.Context, chartDetails *chartutils.Details, client kubernetes.Interface) (*chart.Chart, map[string]string, error) {
	// Most of the existing code that we want to reuse is based on having a typed AppRepository.
	appRepo, caCertSecret, authSecret, _, err := s.getAppRepoAndRelatedSecrets(ctx, s.globalPackagingCluster, chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace)
	if err != nil {
		return nil, nil, status.Errorf(codes.Internal, "Unable to fetch app repo %q from namespace %q: %v", chartDetails.AppRepositoryResourceName, chartDetails.AppRepositoryResourceNamespace, err)
	}

	userAgentString := fmt.Sprintf("%s/%s/%s/%s", UserAgentPrefix, pluginDetail.Name, pluginDetail.Version, version)

	chartID := fmt.Sprintf("%s/%s", appRepo.Name, chartDetails.ChartName)
	log.InfoS("fetching chart with user-agent", "chartID", chartID, "userAgentString", userAgentString)

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
		log.InfoS("using chart tarball", "url", tarballURL)
	}

	// Grab the chart itself
	ch, err := utils.GetChart(
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
	log.InfoS("+helm DeleteInstalledPackage", "cluster", installedRef.GetContext().GetCluster(), "namespace", namespace)

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	keepHistory := false
	err = agent.DeleteRelease(actionConfig, releaseName, keepHistory, s.pluginConfig.TimeoutSeconds)
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
	log.InfoS("+helm RollbackInstalledPackage", "cluster", installedRef.GetContext().GetCluster(), "namespace", installedRef.GetContext().GetNamespace())

	// Create an action config for the installed pkg context.
	actionConfig, err := s.actionConfigGetter(ctx, installedRef.GetContext())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	release, err := agent.RollbackRelease(actionConfig, releaseName, int(request.GetReleaseRevision()), s.pluginConfig.TimeoutSeconds)
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
	identifier := pkgRef.GetIdentifier()
	log.InfoS("+helm GetInstalledPackageResourceRefs", "cluster", pkgRef.GetContext().GetCluster(), "namespace", pkgRef.GetContext().GetNamespace(), "id", identifier)

	fn := func(ctx context.Context, namespace string) (*action.Configuration, error) {
		actionGetter, err := s.actionConfigGetter(ctx, pkgRef.GetContext())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
		}
		return actionGetter, nil
	}

	refs, err := resourcerefs.GetInstalledPackageResourceRefs(
		ctx, types.NamespacedName{Name: identifier, Namespace: pkgRef.Context.Namespace}, fn)
	if err != nil {
		return nil, err
	} else {
		return &corev1.GetInstalledPackageResourceRefsResponse{
			Context:      pkgRef.GetContext(),
			ResourceRefs: refs,
		}, nil
	}
}

func (s *Server) AddPackageRepository(ctx context.Context, request *corev1.AddPackageRepositoryRequest) (*corev1.AddPackageRepositoryResponse, error) {
	repoName := request.GetName()
	repoUrl := request.GetUrl()
	log.Infof("+helm AddPackageRepository '%s' pointing to '%s'", repoName, repoUrl)

	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request provided")
	}
	if request.Context == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request Context provided")
	}
	cluster := request.GetContext().GetCluster()
	if cluster == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no cluster specified: request.Context.Cluster: [%v]", request.Context.Cluster)
	}

	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no package repository Name provided")
	}
	namespace := request.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.GetGlobalPackagingNamespace()
	}
	if request.GetNamespaceScoped() != (namespace != s.GetGlobalPackagingNamespace()) {
		return nil, status.Errorf(codes.InvalidArgument, "Namespace Scope is inconsistent with the provided Namespace")
	}
	name := types.NamespacedName{
		Name:      request.Name,
		Namespace: namespace,
	}

	// Get Helm-specific values
	var customDetail *helmv1.HelmPackageRepositoryCustomDetail
	if request.CustomDetail != nil {
		customDetail = &helmv1.HelmPackageRepositoryCustomDetail{}
		if err := request.CustomDetail.UnmarshalTo(customDetail); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "customDetail could not be parsed: [%v]", request.CustomDetail)
		}
		log.Infof("+helm customDetail [%v]", customDetail)
	}

	helmRepo := &HelmRepository{
		cluster:      cluster,
		name:         name,
		url:          request.GetUrl(),
		repoType:     request.GetType(),
		description:  request.GetDescription(),
		interval:     request.GetInterval(),
		tlsConfig:    request.GetTlsConfig(),
		auth:         request.GetAuth(),
		customDetail: customDetail,
	}
	if repoRef, err := s.newRepo(ctx, helmRepo); err != nil {
		return nil, err
	} else {
		return &corev1.AddPackageRepositoryResponse{PackageRepoRef: repoRef}, nil
	}
}

func (s *Server) getClient(ctx context.Context, cluster string, namespace string) (ctrlclient.Client, error) {
	client, err := s.clientGetter.ControllerRuntime(ctx, cluster)
	if err != nil {
		return nil, err
	}
	return ctrlclient.NewNamespacedClient(client, namespace), nil
}

func (s *Server) GetPackageRepositoryDetail(ctx context.Context, request *corev1.GetPackageRepositoryDetailRequest) (*corev1.GetPackageRepositoryDetailResponse, error) {
	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request PackageRepoRef provided")
	}
	repoRef := request.GetPackageRepoRef()
	if repoRef.GetContext() == nil || repoRef.GetContext().GetCluster() == "" || repoRef.GetContext().GetNamespace() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no valid context provided")
	}
	log.Infof("+helm GetPackageRepositoryDetail '%s' in context [%v]", repoRef.Identifier, repoRef.Context)

	cluster := repoRef.GetContext().GetCluster()
	namespace := repoRef.GetContext().GetNamespace()
	name := repoRef.GetIdentifier()

	log.InfoS("+helm GetPackageRepositoryDetail", "cluster", cluster, "namespace", namespace, "name", name)

	// Retrieve repository
	appRepo, caCertSecret, authSecret, imagesPullSecret, err := s.getAppRepoAndRelatedSecrets(ctx, cluster, name, namespace)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to retrieve AppRepository '%s/%s' due to [%v]", namespace, name, err)
	}

	// Map to target struct
	repositoryDetail, err := s.mapToPackageRepositoryDetail(appRepo, cluster, namespace, caCertSecret, authSecret, imagesPullSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to convert the AppRepository: %v", err)
	}

	// response
	response := &corev1.GetPackageRepositoryDetailResponse{
		Detail: repositoryDetail,
	}

	return response, nil
}

func (s *Server) GetPackageRepositorySummaries(ctx context.Context, request *corev1.GetPackageRepositorySummariesRequest) (*corev1.GetPackageRepositorySummariesResponse, error) {
	log.Infof("+helm GetPackageRepositorySummaries [%v]", request)
	if request.GetContext() == nil || request.GetContext().GetCluster() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no valid context provided")
	}

	if summaries, err := s.repoSummaries(ctx, request.GetContext().GetCluster(), request.GetContext().GetNamespace()); err != nil {
		return nil, err
	} else {
		return &corev1.GetPackageRepositorySummariesResponse{
			PackageRepositorySummaries: summaries,
		}, nil
	}
}

func (s *Server) UpdatePackageRepository(ctx context.Context, request *corev1.UpdatePackageRepositoryRequest) (*corev1.UpdatePackageRepositoryResponse, error) {
	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request PackageRepoRef provided")
	}
	repoRef := request.GetPackageRepoRef()
	if repoRef.GetContext() == nil || repoRef.GetContext().GetCluster() == "" || repoRef.GetContext().GetNamespace() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no valid context provided")
	}
	log.Infof("+helm UpdatePackageRepository '%s' in context [%v]", repoRef.Identifier, repoRef.Context)

	// Get Helm-specific values
	var customDetail *helmv1.HelmPackageRepositoryCustomDetail
	if request.CustomDetail != nil {
		customDetail = &helmv1.HelmPackageRepositoryCustomDetail{}
		if err := request.CustomDetail.UnmarshalTo(customDetail); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "customDetail could not be parsed: [%v]", request.CustomDetail)
		}
		log.V(4).Infof("+helm upgrade repo %s customDetail [%v]", repoRef.Identifier, customDetail)
	}

	helmRepo := &HelmRepository{
		cluster: repoRef.GetContext().GetCluster(),
		name: types.NamespacedName{
			Name:      repoRef.Identifier,
			Namespace: repoRef.GetContext().GetNamespace(),
		},
		url:          request.GetUrl(),
		description:  request.GetDescription(),
		interval:     request.GetInterval(),
		tlsConfig:    request.GetTlsConfig(),
		auth:         request.GetAuth(),
		customDetail: customDetail,
	}
	if responseRef, err := s.updateRepo(ctx, helmRepo); err != nil {
		return nil, err
	} else {
		return &corev1.UpdatePackageRepositoryResponse{
			PackageRepoRef: responseRef,
		}, nil
	}
}

func (s *Server) DeletePackageRepository(ctx context.Context, request *corev1.DeletePackageRepositoryRequest) (*corev1.DeletePackageRepositoryResponse, error) {
	log.Infof("+helm DeletePackageRepository [%v]", request)

	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request PackageRepoRef provided")
	}
	repoRef := request.GetPackageRepoRef()
	if repoRef.GetContext() == nil || repoRef.GetContext().GetCluster() == "" || repoRef.GetContext().GetNamespace() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no valid context provided")
	}
	cluster := repoRef.GetContext().GetCluster()

	if err := s.deleteRepo(ctx, cluster, repoRef); err != nil {
		return nil, err
	} else {
		return &corev1.DeletePackageRepositoryResponse{}, nil
	}
}

// SetUserManagedSecrets This endpoint exists only for integration unit tests
func (s *Server) SetUserManagedSecrets(ctx context.Context, request *helmv1.SetUserManagedSecretsRequest) (*helmv1.SetUserManagedSecretsResponse, error) {
	log.Infof("+helm SetUserManagedSecrets [%t]", request.Value)
	oldVal := s.pluginConfig.UserManagedSecrets
	s.pluginConfig.UserManagedSecrets = request.Value
	return &helmv1.SetUserManagedSecretsResponse{
		Value: oldVal,
	}, nil
}
