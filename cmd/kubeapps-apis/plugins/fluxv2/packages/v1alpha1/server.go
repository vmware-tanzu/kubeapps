// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resources"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/helm"

	helmv2beta2 "github.com/fluxcd/helm-controller/api/v2beta2"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	corev1connect "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1/v1alpha1connect"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1connect.PackagesServiceHandler = (*Server)(nil)
var _ corev1connect.RepositoriesServiceHandler = (*Server)(nil)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedFluxV2PackagesServiceServer
	v1alpha1.UnimplementedFluxV2RepositoriesServiceServer

	// kubeappsCluster specifies the cluster on which Kubeapps is installed.
	kubeappsCluster string
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	// It is meant for in-band interactions (i.e. in the context of a caller)
	// with k8s API server
	clientGetter clientgetter.ClientProviderInterface
	// for interactions with k8s API server in the context of
	// kubeapps-internal-kubeappsapis service account
	serviceAccountClientGetter clientgetter.FixedClusterClientProviderInterface

	actionConfigGetter helm.HelmActionConfigGetterFunc

	repoCache  *cache.NamespacedResourceWatcherCache
	chartCache *cache.ChartCache

	pluginConfig *common.FluxPluginConfig
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, kubeappsCluster string, stopCh <-chan struct{}, pluginConfigPath string, clientQPS float32, clientBurst int) (*Server, error) {
	log.Infof("+fluxv2 NewServer(kubeappsCluster: [%v], pluginConfigPath: [%s]",
		kubeappsCluster, pluginConfigPath)

	if redisCli, err := common.NewRedisClientFromEnv(stopCh); err != nil {
		return nil, err
	} else if chartCache, err := cache.NewChartCache("chartCache", redisCli, stopCh); err != nil {
		return nil, err
	} else {
		pluginConfig := common.NewDefaultPluginConfig()
		if pluginConfigPath != "" {
			pluginConfig, err = common.ParsePluginConfig(pluginConfigPath)
			if err != nil {
				log.Fatalf("%s", err)
			}
			log.Infof("+fluxv2 using custom config: [%v]", *pluginConfig)
		} else {
			log.Info("+fluxv2 using default config since pluginConfigPath is empty")
		}

		// register the GitOps Toolkit schema definitions
		scheme := runtime.NewScheme()
		err = sourcev1beta2.AddToScheme(scheme)
		if err != nil {
			log.Fatalf("%s", err)
		}
		err = helmv2beta2.AddToScheme(scheme)
		if err != nil {
			log.Fatalf("%s", err)
		}

		backgroundClientGetter := clientgetter.NewBackgroundClientProvider(clientgetter.Options{Scheme: scheme}, clientQPS, clientBurst)

		s := repoEventSink{
			clientGetter: backgroundClientGetter,
			chartCache:   chartCache,
		}
		repoCacheConfig := cache.NamespacedResourceWatcherCacheConfig{
			Gvr:          common.GetRepositoriesGvr(),
			ClientGetter: s.clientGetter,
			OnAddFunc:    s.onAddRepo,
			OnModifyFunc: s.onModifyRepo,
			OnGetFunc:    s.onGetRepo,
			OnDeleteFunc: s.onDeleteRepo,
			OnResyncFunc: s.onResync,
			NewObjFunc:   func() ctrlclient.Object { return &sourcev1beta2.HelmRepository{} },
			NewListFunc:  func() ctrlclient.ObjectList { return &sourcev1beta2.HelmRepositoryList{} },
			ListItemsFunc: func(ol ctrlclient.ObjectList) []ctrlclient.Object {
				if hl, ok := ol.(*sourcev1beta2.HelmRepositoryList); !ok {
					log.Errorf("Expected: *sourcev1beta2.HelmRepositoryList, got: %T", ol)
					return nil
				} else {
					ret := make([]ctrlclient.Object, len(hl.Items))
					for i, hr := range hl.Items {
						ret[i] = hr.DeepCopy()
					}
					return ret
				}
			},
		}
		if repoCache, err := cache.NewNamespacedResourceWatcherCache(
			"repoCache", repoCacheConfig, redisCli, stopCh, false); err != nil {
			return nil, err
		} else {
			clientProvider, err := clientgetter.NewClientProvider(configGetter, clientgetter.Options{Scheme: scheme})
			if err != nil {
				log.Fatalf("%s", err)
			}
			return &Server{
				clientGetter:               clientProvider,
				serviceAccountClientGetter: backgroundClientGetter,
				actionConfigGetter: helm.NewHelmActionConfigGetter(
					configGetter, kubeappsCluster),
				repoCache:       repoCache,
				chartCache:      chartCache,
				kubeappsCluster: kubeappsCluster,
				pluginConfig:    pluginConfig,
			}, nil
		}
	}
}

// ===== general note on error handling ========
// using fmt.Errorf vs status.Errorf in functions exposed as grpc:
//
// grpc itself will transform any error into a grpc status code (which is
// then translated into an http status via grpc gateway), so we'll need to
// be using status.Errorf(...) here, rather than fmt.Errorf(...), the former
// allowing you to specify a status code with the error which can be used
// for grpc and translated or http. Without doing this, the grpc status will
// be codes.Unknown which is translated to a 500. you might have a helper
// function that returns an error, then your actual handler function handles
// that error by returning a status.Errorf with the appropriate code

// GetAvailablePackageSummaries returns the available packages based on the request.
// Note that currently packages are returned only from repos that are in a 'Ready'
// state. For the fluxv2 plugin:
//   - if flux helm-controller flag "-no-cross-namespace-refs=true" is
//     enabled only the request target namespace is relevant
//     ref https://github.com/vmware-tanzu/kubeapps/issues/5541
//   - otherwise the request context namespace (the target
//     namespace) is not relevant since charts from a repository in any namespace
//     accessible to the user are available to be installed in the target namespace.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageSummariesRequest]) (*connect.Response[corev1.GetAvailablePackageSummariesResponse], error) {
	log.Infof("+fluxv2 GetAvailablePackageSummaries(request: [%v])", request)
	defer log.Info("-fluxv2 GetAvailablePackageSummaries")

	// grpc compiles in getters for you which automatically return a default (empty) struct
	// if the pointer was nil
	if request == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The request was nil"))
	}
	cluster := request.Msg.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.Context.Cluster: [%v]", request.Msg.Context.Cluster))
	}

	itemOffset, err := paginate.ItemOffsetFromPageToken(request.Msg.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	ns := metav1.NamespaceAll
	if s.pluginConfig.NoCrossNamespaceRefs {
		ns = request.Msg.Context.Namespace
	}

	charts, err := s.getChartsForRepos(ctx, request.Header(), ns, request.Msg.GetFilterOptions().GetRepositories())
	if err != nil {
		return nil, err
	}

	pageSize := request.Msg.GetPaginationOptions().GetPageSize()
	packageSummaries, err := filterAndPaginateCharts(
		request.Msg.GetFilterOptions(), pageSize, itemOffset, charts)
	if err != nil {
		return nil, err
	}

	// per https://github.com/vmware-tanzu/kubeapps/pull/3686#issue-1038093832
	for _, summary := range packageSummaries {
		summary.AvailablePackageRef.Context.Cluster = s.kubeappsCluster
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(packageSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", itemOffset+int(pageSize))
	}

	return connect.NewResponse(&corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: packageSummaries,
		NextPageToken:             nextPageToken,
		// TODO (gfichtenholt) Categories?
		// Just happened to notice that helm plug-in returning this.
		// Never discussed this and the design doc appears to have a lot of back-and-forth comments
		// about this, semantics aren't very clear
	}), nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageDetailRequest]) (*connect.Response[corev1.GetAvailablePackageDetailResponse], error) {
	log.Infof("+fluxv2 GetAvailablePackageDetail(request: [%v])", request)
	defer log.Info("-fluxv2 GetAvailablePackageDetail")

	if request == nil || request.Msg.AvailablePackageRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef provided"))
	}

	packageRef := request.Msg.AvailablePackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("AvailablePackageReference is missing required 'namespace' field"))
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]", cluster))
	}

	pkgDetail, err := s.availableChartDetail(ctx, request.Header(), request.Msg.GetAvailablePackageRef(), request.Msg.GetPkgVersion())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: pkgDetail,
	}), nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageVersionsRequest]) (*connect.Response[corev1.GetAvailablePackageVersionsResponse], error) {
	log.Infof("+fluxv2 GetAvailablePackageVersions [%v]", request)
	defer log.Info("-fluxv2 GetAvailablePackageVersions")

	if request.Msg.GetPkgVersion() != "" {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.GetPkgVersion(): [%v]", request.Msg.GetPkgVersion()))
	}

	packageRef := request.Msg.GetAvailablePackageRef()
	namespace := packageRef.GetContext().GetNamespace()
	if namespace == "" || packageRef.GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]", cluster))
	}

	repoName, chartName, err := pkgutils.SplitPackageIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart [%s] in namespace [%s]", chartName, namespace)
	repo := types.NamespacedName{Namespace: namespace, Name: repoName}
	chart, err := s.getChartModel(ctx, request.Header(), repo, chartName)
	if err != nil {
		return nil, err
	} else if chart != nil {
		// found it
		return connect.NewResponse(&corev1.GetAvailablePackageVersionsResponse{
			PackageAppVersions: pkgutils.PackageAppVersionsSummary(
				chart.ChartVersions,
				s.pluginConfig.VersionsInSummary),
		}), nil
	} else {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to retrieve versions for chart: [%s]", packageRef.Identifier))
	}
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageSummariesRequest]) (*connect.Response[corev1.GetInstalledPackageSummariesResponse], error) {
	log.Infof("+fluxv2 GetInstalledPackageSummaries [%v]", request)
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.Msg.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	cluster := request.Msg.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.Context.Cluster: [%v]", cluster))
	}

	pageSize := request.Msg.GetPaginationOptions().GetPageSize()
	installedPkgSummaries, err := s.paginatedInstalledPkgSummaries(
		ctx, request.Header(), request.Msg.GetContext().GetNamespace(), pageSize, itemOffset)
	if err != nil {
		return nil, err
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(installedPkgSummaries) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", itemOffset+int(pageSize))
	}

	response := &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: installedPkgSummaries,
		NextPageToken:             nextPageToken,
	}
	return connect.NewResponse(response), nil
}

// GetInstalledPackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageDetailRequest]) (*connect.Response[corev1.GetInstalledPackageDetailResponse], error) {
	log.Infof("+fluxv2 GetInstalledPackageDetail [%v]", request)

	if request == nil || request.Msg.InstalledPackageRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request InstalledPackageRef provided"))
	}

	packageRef := request.Msg.InstalledPackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("InstalledPackageReference is missing required 'namespace' field"))
	}

	cluster := packageRef.Context.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.InstalledPackageRef.Context.Cluster: [%v]", cluster))
	}

	key := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: packageRef.Identifier}
	pkgDetail, err := s.installedPackageDetail(ctx, request.Header(), key)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: pkgDetail,
	}), nil
}

// CreateInstalledPackage creates an installed package based on the request.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *connect.Request[corev1.CreateInstalledPackageRequest]) (*connect.Response[corev1.CreateInstalledPackageResponse], error) {
	log.Infof("+fluxv2 CreateInstalledPackage [%v]", request)

	if request == nil || request.Msg.AvailablePackageRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef provided"))
	}
	packageRef := request.Msg.AvailablePackageRef
	if packageRef.GetContext().GetNamespace() == "" || packageRef.GetIdentifier() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Required context or identifier not provided"))
	}
	cluster := packageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]", cluster))
	}
	if request.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Name provided"))
	}
	if request.Msg.TargetContext == nil || request.Msg.TargetContext.Namespace == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request TargetContext namespace provided"))
	}
	cluster = request.Msg.TargetContext.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.TargetContext.Cluster: [%v]", request.Msg.TargetContext.Cluster))
	}

	name := types.NamespacedName{Name: request.Msg.Name, Namespace: request.Msg.TargetContext.Namespace}

	if installedRef, err := s.newRelease(
		ctx,
		request.Header(),
		request.Msg.AvailablePackageRef,
		name,
		request.Msg.PkgVersionReference,
		request.Msg.ReconciliationOptions,
		request.Msg.Values); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.CreateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}), nil
	}
}

// UpdateInstalledPackage updates an installed package based on the request.
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *connect.Request[corev1.UpdateInstalledPackageRequest]) (*connect.Response[corev1.UpdateInstalledPackageResponse], error) {
	log.Infof("+fluxv2 UpdateInstalledPackage [%v]", request)

	if request == nil || request.Msg.InstalledPackageRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request InstalledPackageRef provided"))
	}

	installedPackageRef := request.Msg.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.installedPackageRef.Context.Cluster: [%v]", cluster))
	}

	if installedRef, err := s.updateRelease(
		ctx,
		request.Header(),
		installedPackageRef,
		request.Msg.PkgVersionReference,
		request.Msg.ReconciliationOptions,
		request.Msg.Values); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.UpdateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}), nil
	}
}

// DeleteInstalledPackage deletes an installed package.
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *connect.Request[corev1.DeleteInstalledPackageRequest]) (*connect.Response[corev1.DeleteInstalledPackageResponse], error) {
	log.Infof("+fluxv2 DeleteInstalledPackage [%v]", request)

	if request == nil || request.Msg.InstalledPackageRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request InstalledPackageRef provided"))
	}

	installedPackageRef := request.Msg.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.installedPackageRef.Context.Cluster: [%v]", cluster))
	}

	if err := s.deleteRelease(ctx, request.Header(), request.Msg.InstalledPackageRef); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.DeleteInstalledPackageResponse{}), nil
	}
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes
// resources created by an installed package.
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *connect.Request[corev1.GetInstalledPackageResourceRefsRequest]) (*connect.Response[corev1.GetInstalledPackageResourceRefsResponse], error) {
	pkgRef := request.Msg.GetInstalledPackageRef()
	identifier := pkgRef.GetIdentifier()
	log.InfoS("+fluxv2 GetInstalledPackageResourceRefs", "cluster", pkgRef.GetContext().GetCluster(), "namespace", pkgRef.GetContext().GetNamespace(), "id", identifier)

	key := types.NamespacedName{Namespace: pkgRef.Context.Namespace, Name: identifier}
	rel, err := s.getReleaseInCluster(ctx, request.Header(), key)
	if err != nil {
		return nil, err
	}
	hrName := helmReleaseName(key, rel)
	refs, err := resourcerefs.GetInstalledPackageResourceRefs(request.Header(), hrName, s.actionConfigGetter)
	if err != nil {
		return nil, err
	} else {
		return connect.NewResponse(
			&corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster: s.kubeappsCluster,
					// TODO (gfichtenholt) it is not specifically called out in the spec why there is a
					// need for a Context in the response and MORE imporantly what the value of Namespace
					// field should be. In particular, there is use case when Flux Helm Release in
					// installed in ns1 but specifies targetNamespace as test2. Should we:
					//  (a) return ns1 (the namespace where CRs are installed) OR
					//  (b) return ns2 (the namespace where flux installs the resources specified by the
					//    release).
					// For now lets use (a)
					Namespace: key.Namespace,
				},
				ResourceRefs: refs,
			}), nil
	}
}

func (s *Server) AddPackageRepository(ctx context.Context, request *connect.Request[corev1.AddPackageRepositoryRequest]) (*connect.Response[corev1.AddPackageRepositoryResponse], error) {
	if request == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request provided"))
	}
	if request.Msg.Context == nil || request.Msg.Context.Namespace == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Context namespace provided"))
	}

	cluster := request.Msg.GetContext().GetCluster()
	namespace := request.Msg.GetContext().GetNamespace()
	repoName := request.Msg.GetName()
	log.InfoS("+fluxv2 AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", repoName)

	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.Context.Cluster: [%v]", request.Msg.Context.Cluster))
	}

	if repoRef, err := s.newRepo(ctx, request); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.AddPackageRepositoryResponse{PackageRepoRef: repoRef.Msg}), nil
	}
}

func (s *Server) GetPackageRepositoryDetail(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryDetailRequest]) (*connect.Response[corev1.GetPackageRepositoryDetailResponse], error) {
	log.Infof("+fluxv2 GetPackageRepositoryDetail [%v]", request)
	defer log.Info("-fluxv2 GetPackageRepositoryDetail")
	if request == nil || request.Msg.PackageRepoRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request AvailablePackageRef provided"))
	}

	repoRef := request.Msg.PackageRepoRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if repoRef.Context == nil || len(repoRef.Context.Namespace) == 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("PackageRepositoryReference is missing required namespace"))
	}

	cluster := repoRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.PackageRepoRef.Context.Cluster: [%v]", cluster))
	}

	repoDetail, err := s.repoDetail(ctx, request.Header(), repoRef)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&corev1.GetPackageRepositoryDetailResponse{
		Detail: repoDetail,
	}), nil
}

// GetPackageRepositorySummaries returns the package repositories managed by the 'fluxv2' plugin
func (s *Server) GetPackageRepositorySummaries(ctx context.Context, request *connect.Request[corev1.GetPackageRepositorySummariesRequest]) (*connect.Response[corev1.GetPackageRepositorySummariesResponse], error) {
	log.Infof("+fluxv2 GetPackageRepositorySummaries [%v]", request)
	cluster := request.Msg.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.Context.Cluster: [%v]", cluster))
	}

	if summaries, err := s.repoSummaries(ctx, request.Header(), request.Msg.GetContext().GetNamespace()); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.GetPackageRepositorySummariesResponse{
			PackageRepositorySummaries: summaries,
		}), nil
	}
}

// UpdatePackageRepository updates a package repository based on the request.
func (s *Server) UpdatePackageRepository(ctx context.Context, request *connect.Request[corev1.UpdatePackageRepositoryRequest]) (*connect.Response[corev1.UpdatePackageRepositoryResponse], error) {
	log.Infof("+fluxv2 UpdatePackageRepository [%v]", request)
	if request == nil || request.Msg.PackageRepoRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request PackageRepoRef provided"))
	}

	repoRef := request.Msg.PackageRepoRef
	cluster := repoRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.packageRepoRef.Context.Cluster: [%v]", cluster))
	}

	if responseRef, err := s.updateRepo(ctx, repoRef, request); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.UpdatePackageRepositoryResponse{
			PackageRepoRef: responseRef,
		}), nil
	}
}

// DeletePackageRepository deletes a package repository based on the request.
func (s *Server) DeletePackageRepository(ctx context.Context, request *connect.Request[corev1.DeletePackageRepositoryRequest]) (*connect.Response[corev1.DeletePackageRepositoryResponse], error) {
	log.Infof("+fluxv2 DeletePackageRepository [%v]", request)
	if request == nil || request.Msg.PackageRepoRef == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request PackageRepoRef provided"))
	}

	repoRef := request.Msg.PackageRepoRef
	cluster := repoRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Not supported yet: request.packageRepoRef.Context.Cluster: [%v]", cluster))
	}

	if err := s.deleteRepo(ctx, request.Header(), repoRef); err != nil {
		return nil, err
	} else {
		return connect.NewResponse(&corev1.DeletePackageRepositoryResponse{}), nil
	}
}

func (s *Server) GetPackageRepositoryPermissions(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryPermissionsRequest]) (*connect.Response[corev1.GetPackageRepositoryPermissionsResponse], error) {
	log.Infof("+fluxv2 GetPackageRepositoryPermissions [%v]", request)

	cluster := request.Msg.GetContext().GetCluster()
	namespace := request.Msg.GetContext().GetNamespace()
	if cluster == "" && namespace != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Cluster must be specified when namespace is present: %s", namespace))
	}
	typedClient, err := s.clientGetter.Typed(request.Header(), cluster)
	if err != nil {
		return nil, err
	}

	resource := schema.GroupResource{
		Group:    sourcev1beta2.GroupVersion.Group,
		Resource: fluxHelmRepositories,
	}

	permissions := &corev1.PackageRepositoriesPermissions{
		Plugin: GetPluginDetail(),
	}

	// Flux does not really have a notion of global repositories

	// Namespace permissions
	if namespace != "" {
		permissions.Namespace, err = resources.GetPermissionsOnResource(ctx, typedClient, resource, request.Msg.GetContext().GetNamespace())
		if err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&corev1.GetPackageRepositoryPermissionsResponse{
		Permissions: []*corev1.PackageRepositoriesPermissions{permissions},
	}), nil
}

// makes the server look like a repo event sink. Facilitates code reuse between
// use cases when something happens in background as a result of a watch event,
// aka an "out-of-band" interaction and use cases when the user wants something
// done explicitly, aka "in-band" interaction
func (s *Server) newRepoEventSink() repoEventSink {

	cg := &clientgetter.FixedClusterClientProvider{ClientsFunc: func(ctx context.Context) (*clientgetter.ClientGetter, error) {
		// Empty headers used here since this getter is for a service account
		// only.
		// TODO: (minelson) We need to pass the headers of the request down to
		// here, updating the ClientsFunc signature.
		return s.clientGetter.GetClients(http.Header{}, s.kubeappsCluster)
	}}

	// notice a bit of inconsistency here, we are using the context
	// of the incoming request to read the secret
	// as opposed to s.repoCache.clientGetter (which uses the context of
	//	User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis")
	// which is what is used when the repo is being processed/indexed.
	// I don't think it's necessarily a bad thing if the incoming user's RBAC
	// settings are more permissive than that of the default RBAC for
	// kubeapps-internal-kubeappsapis account. If we don't like that behavior,
	// I can easily switch to BackgroundClientGetter here
	return repoEventSink{
		clientGetter: cg,
		chartCache:   s.chartCache,
	}
}

func (s *Server) getClient(headers http.Header, namespace string) (ctrlclient.Client, error) {
	client, err := s.clientGetter.ControllerRuntime(headers, s.kubeappsCluster)
	if err != nil {
		return nil, err
	}
	return ctrlclient.NewNamespacedClient(client, namespace), nil
}

// hasAccessToNamespace returns an error if the client does not have read access to a given namespace
func (s *Server) hasAccessToNamespace(ctx context.Context, headers http.Header, gvr schema.GroupVersionResource, namespace string) (bool, error) {
	typedCli, err := s.clientGetter.Typed(headers, s.kubeappsCluster)
	if err != nil {
		return false, err
	}

	res, err := typedCli.AuthorizationV1().SelfSubjectAccessReviews().Create(
		ctx,
		&authorizationv1.SelfSubjectAccessReview{
			Spec: authorizationv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authorizationv1.ResourceAttributes{
					Group:     gvr.Group,
					Version:   gvr.Version,
					Resource:  gvr.Resource,
					Verb:      "get",
					Namespace: namespace,
				},
			},
		}, metav1.CreateOptions{})
	if err != nil {
		return false, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to check if the user has access to the namespace: %w", err))
	}
	return res.Status.Allowed, nil
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *plugins.Plugin {
	return common.GetPluginDetail()
}

func (s *Server) GetAvailablePackageMetadatas(ctx context.Context, request *connect.Request[corev1.GetAvailablePackageMetadatasRequest]) (*connect.Response[corev1.GetAvailablePackageMetadatasResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Unimplemented"))
}
