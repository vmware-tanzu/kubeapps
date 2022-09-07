// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"reflect"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)
var _ corev1.RepositoriesServiceServer = (*Server)(nil)

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
	clientGetter clientgetter.ClientGetterFunc
	// for interactions with k8s API server in the context of
	// kubeapps-internal-kubeappsapis service account
	serviceAccountClientGetter clientgetter.BackgroundClientGetterFunc

	actionConfigGetter clientgetter.HelmActionConfigGetterFunc

	repoCache  *cache.NamespacedResourceWatcherCache
	chartCache *cache.ChartCache

	pluginConfig *common.FluxPluginConfig
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, kubeappsCluster string, stopCh <-chan struct{}, pluginConfigPath string) (*Server, error) {
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
		err = sourcev1.AddToScheme(scheme)
		if err != nil {
			log.Fatalf("%s", err)
		}
		err = helmv2.AddToScheme(scheme)
		if err != nil {
			log.Fatalf("%s", err)
		}

		backgroundClientGetter := clientgetter.NewBackgroundClientGetter(
			configGetter, clientgetter.Options{Scheme: scheme})

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
			NewObjFunc:   func() ctrlclient.Object { return &sourcev1.HelmRepository{} },
			NewListFunc:  func() ctrlclient.ObjectList { return &sourcev1.HelmRepositoryList{} },
			ListItemsFunc: func(ol ctrlclient.ObjectList) []ctrlclient.Object {
				if hl, ok := ol.(*sourcev1.HelmRepositoryList); !ok {
					log.Errorf("Expected: *sourcev1.HelmRepositoryList, got: %s", reflect.TypeOf(ol))
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
			return &Server{
				clientGetter: clientgetter.NewClientGetter(
					configGetter, clientgetter.Options{Scheme: scheme}),
				serviceAccountClientGetter: backgroundClientGetter,
				actionConfigGetter: clientgetter.NewHelmActionConfigGetter(
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
// state. For the fluxv2 plugin, the request context namespace (the target
// namespace) is not relevant since charts from a repository in any namespace
// accessible to the user are available to be installed in the target namespace.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageSummaries(request: [%v])", request)
	defer log.Info("-fluxv2 GetAvailablePackageSummaries")

	// grpc compiles in getters for you which automatically return a default (empty) struct
	// if the pointer was nil
	cluster := request.GetContext().GetCluster()
	if request != nil && cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	itemOffset, err := paginate.ItemOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	charts, err := s.getChartsForRepos(ctx, request.GetFilterOptions().GetRepositories())
	if err != nil {
		return nil, err
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	packageSummaries, err := filterAndPaginateCharts(
		request.GetFilterOptions(), pageSize, itemOffset, charts)
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

	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: packageSummaries,
		NextPageToken:             nextPageToken,
		// TODO (gfichtenholt) Categories?
		// Just happened to notice that helm plug-in returning this.
		// Never discussed this and the design doc appears to have a lot of back-and-forth comments
		// about this, semantics aren't very clear
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageDetail(request: [%v])", request)
	defer log.Info("-fluxv2 GetAvailablePackageDetail")

	if request == nil || request.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef provided")
	}

	packageRef := request.AvailablePackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "AvailablePackageReference is missing required 'namespace' field")
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}

	pkgDetail, err := s.availableChartDetail(ctx, request.GetAvailablePackageRef(), request.GetPkgVersion())
	if err != nil {
		return nil, err
	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: pkgDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageVersions [%v]", request)
	defer log.Info("-fluxv2 GetAvailablePackageVersions")

	if request.GetPkgVersion() != "" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.GetPkgVersion(): [%v]",
			request.GetPkgVersion())
	}

	packageRef := request.GetAvailablePackageRef()
	namespace := packageRef.GetContext().GetNamespace()
	if namespace == "" || packageRef.GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}

	cluster := packageRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}

	repoName, chartName, err := pkgutils.SplitPackageIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart [%s] in namespace [%s]", chartName, namespace)
	repo := types.NamespacedName{Namespace: namespace, Name: repoName}
	chart, err := s.getChartModel(ctx, repo, chartName)
	if err != nil {
		return nil, err
	} else if chart != nil {
		// found it
		return &corev1.GetAvailablePackageVersionsResponse{
			PackageAppVersions: pkgutils.PackageAppVersionsSummary(
				chart.ChartVersions,
				s.pluginConfig.VersionsInSummary),
		}, nil
	} else {
		return nil, status.Errorf(codes.Internal, "unable to retrieve versions for chart: [%s]", packageRef.Identifier)
	}
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageSummaries [%v]", request)
	itemOffset, err := paginate.ItemOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			cluster)
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	installedPkgSummaries, err := s.paginatedInstalledPkgSummaries(
		ctx, request.GetContext().GetNamespace(), pageSize, itemOffset)
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
	return response, nil
}

// GetInstalledPackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageDetail(ctx context.Context, request *corev1.GetInstalledPackageDetailRequest) (*corev1.GetInstalledPackageDetailResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageDetail [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	packageRef := request.InstalledPackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "InstalledPackageReference is missing required 'namespace' field")
	}

	cluster := packageRef.Context.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.InstalledPackageRef.Context.Cluster: [%v]",
			cluster)
	}

	key := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: packageRef.Identifier}
	pkgDetail, err := s.installedPackageDetail(ctx, key)
	if err != nil {
		return nil, err
	}

	return &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: pkgDetail,
	}, nil
}

// CreateInstalledPackage creates an installed package based on the request.
func (s *Server) CreateInstalledPackage(ctx context.Context, request *corev1.CreateInstalledPackageRequest) (*corev1.CreateInstalledPackageResponse, error) {
	log.Infof("+fluxv2 CreateInstalledPackage [%v]", request)

	if request == nil || request.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef provided")
	}
	packageRef := request.AvailablePackageRef
	if packageRef.GetContext().GetNamespace() == "" || packageRef.GetIdentifier() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "required context or identifier not provided")
	}
	cluster := packageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.AvailablePackageRef.Context.Cluster: [%v]",
			cluster)
	}
	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}
	if request.TargetContext == nil || request.TargetContext.Namespace == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request TargetContext namespace provided")
	}
	cluster = request.TargetContext.GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.TargetContext.Cluster: [%v]",
			request.TargetContext.Cluster)
	}

	name := types.NamespacedName{Name: request.Name, Namespace: request.TargetContext.Namespace}

	if installedRef, err := s.newRelease(
		ctx,
		request.AvailablePackageRef,
		name,
		request.PkgVersionReference,
		request.ReconciliationOptions,
		request.Values); err != nil {
		return nil, err
	} else {
		return &corev1.CreateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}, nil
	}
}

// UpdateInstalledPackage updates an installed package based on the request.
func (s *Server) UpdateInstalledPackage(ctx context.Context, request *corev1.UpdateInstalledPackageRequest) (*corev1.UpdateInstalledPackageResponse, error) {
	log.Infof("+fluxv2 UpdateInstalledPackage [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	installedPackageRef := request.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
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
		return &corev1.UpdateInstalledPackageResponse{
			InstalledPackageRef: installedRef,
		}, nil
	}
}

// DeleteInstalledPackage deletes an installed package.
func (s *Server) DeleteInstalledPackage(ctx context.Context, request *corev1.DeleteInstalledPackageRequest) (*corev1.DeleteInstalledPackageResponse, error) {
	log.Infof("+fluxv2 DeleteInstalledPackage [%v]", request)

	if request == nil || request.InstalledPackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request InstalledPackageRef provided")
	}

	installedPackageRef := request.InstalledPackageRef
	cluster := installedPackageRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.installedPackageRef.Context.Cluster: [%v]",
			cluster)
	}

	if err := s.deleteRelease(ctx, request.InstalledPackageRef); err != nil {
		return nil, err
	} else {
		return &corev1.DeleteInstalledPackageResponse{}, nil
	}
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes
// resources created by an installed package.
func (s *Server) GetInstalledPackageResourceRefs(ctx context.Context, request *corev1.GetInstalledPackageResourceRefsRequest) (*corev1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	identifier := pkgRef.GetIdentifier()
	log.InfoS("+fluxv2 GetInstalledPackageResourceRefs", "cluster", pkgRef.GetContext().GetCluster(), "namespace", pkgRef.GetContext().GetNamespace(), "id", identifier)

	key := types.NamespacedName{Namespace: pkgRef.Context.Namespace, Name: identifier}
	rel, err := s.getReleaseInCluster(ctx, key)
	if err != nil {
		return nil, err
	}
	hrName := helmReleaseName(key, rel)
	refs, err := resourcerefs.GetInstalledPackageResourceRefs(ctx, hrName, s.actionConfigGetter)
	if err != nil {
		return nil, err
	} else {
		return &corev1.GetInstalledPackageResourceRefsResponse{
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
		}, nil
	}
}

func (s *Server) AddPackageRepository(ctx context.Context, request *corev1.AddPackageRepositoryRequest) (*corev1.AddPackageRepositoryResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request provided")
	}
	if request.Context == nil || request.Context.Namespace == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Context namespace provided")
	}

	cluster := request.GetContext().GetCluster()
	namespace := request.GetContext().GetNamespace()
	repoName := request.GetName()
	log.InfoS("+fluxv2 AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", repoName)

	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	if repoRef, err := s.newRepo(ctx, request); err != nil {
		return nil, err
	} else {
		return &corev1.AddPackageRepositoryResponse{PackageRepoRef: repoRef}, nil
	}
}

func (s *Server) GetPackageRepositoryDetail(ctx context.Context, request *corev1.GetPackageRepositoryDetailRequest) (*corev1.GetPackageRepositoryDetailResponse, error) {
	log.Infof("+fluxv2 GetPackageRepositoryDetail [%v]", request)
	defer log.Info("-fluxv2 GetPackageRepositoryDetail")
	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request AvailablePackageRef provided")
	}

	repoRef := request.PackageRepoRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if repoRef.Context == nil || len(repoRef.Context.Namespace) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "PackageRepositoryReference is missing required namespace")
	}

	cluster := repoRef.Context.Cluster
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.PackageRepoRef.Context.Cluster: [%v]",
			cluster)
	}

	repoDetail, err := s.repoDetail(ctx, repoRef)
	if err != nil {
		return nil, err
	}

	return &corev1.GetPackageRepositoryDetailResponse{
		Detail: repoDetail,
	}, nil
}

// GetPackageRepositorySummaries returns the package repositories managed by the 'fluxv2' plugin
func (s *Server) GetPackageRepositorySummaries(ctx context.Context, request *corev1.GetPackageRepositorySummariesRequest) (*corev1.GetPackageRepositorySummariesResponse, error) {
	log.Infof("+fluxv2 GetPackageRepositorySummaries [%v]", request)
	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			cluster)
	}

	if summaries, err := s.repoSummaries(ctx, request.GetContext().GetNamespace()); err != nil {
		return nil, err
	} else {
		return &corev1.GetPackageRepositorySummariesResponse{
			PackageRepositorySummaries: summaries,
		}, nil
	}
}

// UpdatePackageRepository updates a package repository based on the request.
func (s *Server) UpdatePackageRepository(ctx context.Context, request *corev1.UpdatePackageRepositoryRequest) (*corev1.UpdatePackageRepositoryResponse, error) {
	log.Infof("+fluxv2 UpdatePackageRepository [%v]", request)
	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request PackageRepoRef provided")
	}

	repoRef := request.PackageRepoRef
	cluster := repoRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.packageRepoRef.Context.Cluster: [%v]",
			cluster)
	}

	if responseRef, err := s.updateRepo(ctx, repoRef, request.Url, request.Interval, request.TlsConfig, request.Auth); err != nil {
		return nil, err
	} else {
		return &corev1.UpdatePackageRepositoryResponse{
			PackageRepoRef: responseRef,
		}, nil
	}
}

// DeletePackageRepository deletes a package repository based on the request.
func (s *Server) DeletePackageRepository(ctx context.Context, request *corev1.DeletePackageRepositoryRequest) (*corev1.DeletePackageRepositoryResponse, error) {
	log.Infof("+fluxv2 DeletePackageRepository [%v]", request)
	if request == nil || request.PackageRepoRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request PackageRepoRef provided")
	}

	repoRef := request.PackageRepoRef
	cluster := repoRef.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.packageRepoRef.Context.Cluster: [%v]",
			cluster)
	}

	if err := s.deleteRepo(ctx, repoRef); err != nil {
		return nil, err
	} else {
		return &corev1.DeletePackageRepositoryResponse{}, nil
	}
}

// This endpoint exists only for integration unit tests
func (s *Server) SetUserManagedSecrets(ctx context.Context, request *v1alpha1.SetUserManagedSecretsRequest) (*v1alpha1.SetUserManagedSecretsResponse, error) {
	log.Infof("+fluxv2 SetUserManagedSecrets [%t]", request.Value)
	oldVal := s.pluginConfig.UserManagedSecrets
	s.pluginConfig.UserManagedSecrets = request.Value
	return &v1alpha1.SetUserManagedSecretsResponse{
		Value: oldVal,
	}, nil
}

// makes the server look like a repo event sink. Facilitates code reuse between
// use cases when something happens in background as a result of a watch event,
// aka an "out-of-band" interaction and use cases when the user wants something
// done explicitly, aka "in-band" interaction
func (s *Server) newRepoEventSink() repoEventSink {
	cg := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return s.clientGetter(ctx, s.kubeappsCluster)
	}

	// notice a bit of inconsistency here, we are using s.clientGetter
	// (i.e. the context of the incoming request) to read the secret
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

func (s *Server) getClient(ctx context.Context, namespace string) (ctrlclient.Client, error) {
	client, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster)
	if err != nil {
		return nil, err
	}
	return ctrlclient.NewNamespacedClient(client, namespace), nil
}

// hasAccessToNamespace returns an error if the client does not have read access to a given namespace
func (s *Server) hasAccessToNamespace(ctx context.Context, gvr schema.GroupVersionResource, namespace string) (bool, error) {
	typedCli, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
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
		return false, status.Errorf(codes.Internal, "Unable to check if the user has access to the namespace: %s", err)
	}
	return res.Status.Allowed, nil
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *plugins.Plugin {
	return common.GetPluginDetail()
}
