// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	pkgfluxv2v1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	pkgfluxv2cache "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	pkgfluxv2common "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	paginate "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	pkgutils "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	resourcerefs "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8sapiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8smetaunstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8stypedclient "k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ pkgsGRPCv1alpha1.PackagesServiceServer = (*Server)(nil)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	pkgfluxv2v1alpha1.UnimplementedFluxV2PackagesServiceServer

	// kubeappsCluster specifies the cluster on which Kubeapps is installed.
	kubeappsCluster string
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter       clientgetter.ClientGetterWithApiExtFunc
	actionConfigGetter clientgetter.HelmActionConfigGetterFunc

	repoCache  *pkgfluxv2cache.NamespacedResourceWatcherCache
	chartCache *pkgfluxv2cache.ChartCache

	versionsInSummary pkgutils.VersionsInSummary
	timeoutSeconds    int32
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter apiscore.KubernetesConfigGetter, kubeappsCluster string, stopCh <-chan struct{}, pluginConfigPath string) (*Server, error) {
	log.Infof("+fluxv2 NewServer(kubeappsCluster: [%v], pluginConfigPath: [%s]",
		kubeappsCluster, pluginConfigPath)

	repositoriesGvr := k8sschema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	if redisCli, err := pkgfluxv2common.NewRedisClientFromEnv(); err != nil {
		return nil, err
	} else if chartCache, err := pkgfluxv2cache.NewChartCache("chartCache", redisCli, stopCh); err != nil {
		return nil, err
	} else {
		// If no config is provided, we default to the existing values for backwards
		// compatibility.
		versionsInSummary := pkgutils.GetDefaultVersionsInSummary()
		timeoutSecs := int32(-1)
		if pluginConfigPath != "" {
			versionsInSummary, timeoutSecs, err = parsePluginConfig(pluginConfigPath)
			if err != nil {
				log.Fatalf("%s", err)
			}
			log.Infof("+fluxv2 using custom packages config with %v\n", versionsInSummary)
		} else {
			log.Infof("+fluxv2 using default config since pluginConfigPath is empty")
		}

		s := repoEventSink{
			clientGetter: clientgetter.NewBackgroundClientGetter(),
			chartCache:   chartCache,
		}
		repoCacheConfig := pkgfluxv2cache.NamespacedResourceWatcherCacheConfig{
			Gvr:          repositoriesGvr,
			ClientGetter: s.clientGetter,
			OnAddFunc:    s.onAddRepo,
			OnModifyFunc: s.onModifyRepo,
			OnGetFunc:    s.onGetRepo,
			OnDeleteFunc: s.onDeleteRepo,
			OnResyncFunc: s.onResync,
		}
		if repoCache, err := pkgfluxv2cache.NewNamespacedResourceWatcherCache(
			"repoCache", repoCacheConfig, redisCli, stopCh); err != nil {
			return nil, err
		} else {
			return &Server{
				clientGetter:       clientgetter.NewClientGetterWithApiExt(configGetter, kubeappsCluster),
				actionConfigGetter: clientgetter.NewHelmActionConfigGetter(configGetter, kubeappsCluster),
				repoCache:          repoCache,
				chartCache:         chartCache,
				kubeappsCluster:    kubeappsCluster,
				versionsInSummary:  versionsInSummary,
				timeoutSeconds:     timeoutSecs,
			}, nil
		}
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, nil, grpcstatus.Errorf(grpccodes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, apiExtClient, err := s.clientGetter(ctx)
	if err != nil {
		if grpcstatus.Code(err) == grpccodes.Unknown {
			return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get client due to: %v", err)
		} else {
			// this could be grpccodes.Unauthorized which we want to pass through
			return nil, nil, nil, err
		}
	}
	return typedClient, dynamicClient, apiExtClient, nil
}

// ===== general note on error handling ========
// using fmt.Errorf vs grpcstatus.Errorf in functions exposed as grpc:
//
// grpc itself will transform any error into a grpc status code (which is
// then translated into an http status via grpc gateway), so we'll need to
// be using grpcstatus.Errorf(...) here, rather than fmt.Errorf(...), the former
// allowing you to specify a status code with the error which can be used
// for grpc and translated or http. Without doing this, the grpc status will
// be grpccodes.Unknown which is translated to a 500. you might have a helper
// function that returns an error, then your actual handler function handles
// that error by returning a grpcstatus.Errorf with the appropriate code

// GetPackageRepositories returns the package repositories based on the request.
// note that this func currently returns ALL repositories, not just those in 'ready' (reconciled) state
func (s *Server) GetPackageRepositories(ctx context.Context, request *pkgfluxv2v1alpha1.GetPackageRepositoriesRequest) (*pkgfluxv2v1alpha1.GetPackageRepositoriesResponse, error) {
	log.Infof("+fluxv2 GetPackageRepositories(request: [%v])", request)

	if request == nil || request.Context == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no context provided")
	}

	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	repos, err := s.listReposInNamespace(ctx, request.Context.Namespace)
	if err != nil {
		return nil, err
	}

	responseRepos := []*pkgfluxv2v1alpha1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo, err := packageRepositoryFromUnstructured(repoUnstructured.Object)
		if err != nil {
			return nil, err
		}
		responseRepos = append(responseRepos, repo)
	}
	return &pkgfluxv2v1alpha1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
// Note that currently packages are returned only from repos that are in a 'Ready'
// state. For the fluxv2 plugin, the request context namespace (the target
// namespace) is not relevant since charts from a repository in any namespace
//  accessible to the user are available to be installed in the target namespace.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageSummaries(request: [%v])", request)
	defer log.Infof("-fluxv2 GetAvailablePackageSummaries")

	// grpc compiles in getters for you which automatically return a default (empty) struct if the pointer was nil
	cluster := request.GetContext().GetCluster()
	if request != nil && cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	pageOffset, err := paginate.PageOffsetFromAvailableRequest(request)
	if err != nil {
		return nil, err
	}

	charts, err := s.getChartsForRepos(ctx, request.GetFilterOptions().GetRepositories())
	if err != nil {
		return nil, err
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	packageSummaries, err := filterAndPaginateCharts(
		request.GetFilterOptions(), pageSize, pageOffset, charts)
	if err != nil {
		return nil, err
	}

	// per https://github.com/kubeapps/kubeapps/pull/3686#issue-1038093832
	for _, summary := range packageSummaries {
		summary.AvailablePackageRef.Context.Cluster = s.kubeappsCluster
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(packageSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}

	return &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: packageSummaries,
		NextPageToken:             nextPageToken,
		// TODO (gfichtenholt) Categories?
		// Just happened to notice that helm plug-in returning this.
		// Never discussed this and the design doc appears to have a lot of back-and-forth comments
		// about this, semantics aren't very clear
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageDetail(request: [%v])", request)
	defer log.Infof("-fluxv2 GetAvailablePackageDetail")

	if request == nil || request.AvailablePackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request AvailablePackageRef provided")
	}

	packageRef := request.AvailablePackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "AvailablePackageReference is missing required 'namespace' field")
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}

	repoName, chartName, err := pkgutils.SplitChartIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	// check specified repo exists and is in ready state
	repo := k8stypes.NamespacedName{Namespace: packageRef.Context.Namespace, Name: repoName}

	// this verifies that the repo exists
	repoUnstructured, err := s.getRepoInCluster(ctx, repo)
	if err != nil {
		return nil, err
	}

	pkgDetail, err := s.availableChartDetail(ctx, repo, chartName, request.GetPkgVersion())
	if err != nil {
		return nil, err
	}

	// fix up a couple of fields that don't come from the chart tarball
	repoUrl, found, err := k8smetaunstructuredv1.NestedString(repoUnstructured.Object, "spec", "url")
	if err != nil || !found {
		return nil, grpcstatus.Errorf(grpccodes.NotFound, "Missing required field spec.url on repository %q", repo)
	}
	pkgDetail.RepoUrl = repoUrl
	pkgDetail.AvailablePackageRef.Context.Namespace = packageRef.Context.Namespace
	// per https://github.com/kubeapps/kubeapps/pull/3686#issue-1038093832
	pkgDetail.AvailablePackageRef.Context.Cluster = s.kubeappsCluster

	return &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: pkgDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageVersions [%v]", request)
	defer log.Infof("-fluxv2 GetAvailablePackageVersions")

	if request.GetPkgVersion() != "" {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.GetPkgVersion(): [%v]",
			request.GetPkgVersion())
	}

	packageRef := request.GetAvailablePackageRef()
	namespace := packageRef.GetContext().GetNamespace()
	if namespace == "" || packageRef.GetIdentifier() == "" {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "required context or identifier not provided")
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}

	repoName, chartName, err := pkgutils.SplitChartIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart [%s] in namespace [%s]", chartName, namespace)
	repo := k8stypes.NamespacedName{Namespace: namespace, Name: repoName}
	chart, err := s.getChart(ctx, repo, chartName)
	if err != nil {
		return nil, err
	} else if chart != nil {
		// found it
		return &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
			PackageAppVersions: pkgutils.PackageAppVersionsSummary(
				chart.ChartVersions,
				s.versionsInSummary),
		}, nil
	} else {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "unable to retrieve versions for chart: [%s]", packageRef.Identifier)
	}
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageSummaries [%v]", request)
	pageOffset, err := paginate.PageOffsetFromInstalledRequest(request)
	if err != nil {
		return nil, err
	}

	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			cluster)
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	installedPkgSummaries, err := s.paginatedInstalledPkgSummaries(
		ctx, request.GetContext().GetNamespace(), pageSize, pageOffset)
	if err != nil {
		return nil, err
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(installedPkgSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}

	response := &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummaries,
		NextPageToken:             nextPageToken,
	}
	return response, nil
}

// GetInstalledPackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageDetail [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	packageRef := request.InstalledPackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "InstalledPackageReference is missing required 'namespace' field")
	}

	cluster := packageRef.Context.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.InstalledPackageRef.Context.Cluster: [%v]",
			cluster)
	}

	name := k8stypes.NamespacedName{Namespace: packageRef.Context.Namespace, Name: packageRef.Identifier}
	pkgDetail, err := s.installedPackageDetail(ctx, name)
	if err != nil {
		return nil, err
	}

	return &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: pkgDetail,
	}, nil
}

// CreateInstalledPackage creates an installed package based on the request.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.CreateInstalledPackageRequest) (*pkgsGRPCv1alpha1.CreateInstalledPackageResponse, error) {
	log.Infof("+fluxv2 CreateInstalledPackage [%v]", request)

	if request == nil || request.AvailablePackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request AvailablePackageRef provided")
	}
	packageRef := request.AvailablePackageRef
	if packageRef.GetContext().GetNamespace() == "" || packageRef.GetIdentifier() == "" {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "required context or identifier not provided")
	}
	cluster := packageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}
	if request.Name == "" {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request Name provided")
	}
	if request.TargetContext == nil || request.TargetContext.Namespace == "" {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request TargetContext namespace provided")
	}
	cluster = request.TargetContext.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.TargetContext.Cluster: [%v]",
			request.TargetContext.Cluster)
	}

	name := k8stypes.NamespacedName{Name: request.Name, Namespace: request.TargetContext.Namespace}

	if installedRef, err := s.newRelease(
		ctx,
		request.AvailablePackageRef,
		name,
		request.PkgVersionReference,
		request.ReconciliationOptions,
		request.Values); err != nil {
		return nil, err
	} else {
		return &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}, nil
	}
}

// UpdateInstalledPackage updates an installed package based on the request.
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest) (*pkgsGRPCv1alpha1.UpdateInstalledPackageResponse, error) {
	log.Infof("+fluxv2 UpdateInstalledPackage [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	installedPackageRef := request.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.installedPackageRef.Context.Cluster: [%v]",
			cluster)
	}

	if installedRef, err := s.updateRelease(
		ctx,
		installedPackageRef,
		request.PkgVersionReference,
		request.ReconciliationOptions,
		request.Values); err != nil {
		return nil, err
	} else {
		return &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}, nil
	}
}

// DeleteInstalledPackage deletes an installed package.
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest) (*pkgsGRPCv1alpha1.DeleteInstalledPackageResponse, error) {
	log.Infof("+fluxv2 DeleteInstalledPackage [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	installedPackageRef := request.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, grpcstatus.Errorf(
			grpccodes.Unimplemented,
			"not supported yet: request.installedPackageRef.Context.Cluster: [%v]",
			cluster)
	}

	if err := s.deleteRelease(ctx, request.InstalledPackageRef); err != nil {
		return nil, err
	} else {
		return &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{}, nil
	}
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes
// resources created by an installed package.
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", pkgRef.GetContext().GetCluster(), pkgRef.GetContext().GetNamespace())
	identifier := pkgRef.GetIdentifier()
	log.Infof("+fluxv2 GetInstalledPackageResourceRefs %s %s", contextMsg, identifier)

	return resourcerefs.GetInstalledPackageResourceRefs(ctx, request, s.actionConfigGetter)
}

// GetPluginDetail returns a apiscore.plugins.Plugin describing itself.
func GetPluginDetail() *pluginsGRPCv1alpha1.Plugin {
	return pkgfluxv2common.GetPluginDetail()
}

// parsePluginConfig parses the input plugin configuration json file and return the
// configuration options.
func parsePluginConfig(pluginConfigPath string) (pkgutils.VersionsInSummary, int32, error) {
	// Note at present VersionsInSummary is the only configurable option for this plugin,
	// and if required this func can be enhanced to return fluxConfig struct

	// In the flux plugin, for example, we are interested in config for the
	// apiscore.packages.v1alpha1 only. So the plugin defines the following struct and parses the config.
	type fluxConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary pkgutils.VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`
	}
	var config fluxConfig

	pluginConfig, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return pkgutils.VersionsInSummary{}, 0, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return pkgutils.VersionsInSummary{}, 0, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	// return configured value
	return config.Core.Packages.V1alpha1.VersionsInSummary,
		config.Core.Packages.V1alpha1.TimeoutSeconds, nil
}
