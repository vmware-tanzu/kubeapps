// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs"
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

	// kubeappsCluster specifies the cluster on which Kubeapps is installed.
	kubeappsCluster string
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter       clientgetter.ClientGetterFunc
	actionConfigGetter clientgetter.HelmActionConfigGetterFunc

	repoCache  *cache.NamespacedResourceWatcherCache
	chartCache *cache.ChartCache

	versionsInSummary pkgutils.VersionsInSummary
	timeoutSeconds    int32
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, kubeappsCluster string, stopCh <-chan struct{}, pluginConfigPath string) (*Server, error) {
	log.Infof("+fluxv2 NewServer(kubeappsCluster: [%v], pluginConfigPath: [%s]",
		kubeappsCluster, pluginConfigPath)

	repositoriesGvr := schema.GroupVersionResource{
		Group:    sourcev1.GroupVersion.Group,
		Version:  sourcev1.GroupVersion.Version,
		Resource: fluxHelmRepositories,
	}

	if redisCli, err := common.NewRedisClientFromEnv(); err != nil {
		return nil, err
	} else if chartCache, err := cache.NewChartCache("chartCache", redisCli, stopCh); err != nil {
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
			log.Infof("+fluxv2 using custom packages config with versions: [%v]", versionsInSummary)
		} else {
			log.Infof("+fluxv2 using default config since pluginConfigPath is empty")
		}

		// register the GitOps Toolkit schema definitions
		scheme := runtime.NewScheme()
		sourcev1.AddToScheme(scheme)
		helmv2.AddToScheme(scheme)

		s := repoEventSink{
			clientGetter: clientgetter.NewBackgroundClientGetter(
				configGetter, clientgetter.Options{Scheme: scheme}),
			chartCache: chartCache,
		}
		repoCacheConfig := cache.NamespacedResourceWatcherCacheConfig{
			Gvr:          repositoriesGvr,
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
			"repoCache", repoCacheConfig, redisCli, stopCh); err != nil {
			return nil, err
		} else {
			return &Server{
				clientGetter: clientgetter.NewClientGetter(
					configGetter, clientgetter.Options{Scheme: scheme}),
				actionConfigGetter: clientgetter.NewHelmActionConfigGetter(
					configGetter, kubeappsCluster),
				repoCache:         repoCache,
				chartCache:        chartCache,
				kubeappsCluster:   kubeappsCluster,
				versionsInSummary: versionsInSummary,
				timeoutSeconds:    timeoutSecs,
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
//  accessible to the user are available to be installed in the target namespace.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageSummaries(request: [%v])", request)
	defer log.Infof("-fluxv2 GetAvailablePackageSummaries")

	// grpc compiles in getters for you which automatically return a default (empty) struct if the pointer was nil
	cluster := request.GetContext().GetCluster()
	if request != nil && cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
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
	defer log.Infof("-fluxv2 GetAvailablePackageDetail")

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

	repoN, chartName, err := pkgutils.SplitChartIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	// check specified repo exists and is in ready state
	repoName := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: repoN}

	// this verifies that the repo exists
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	}

	pkgDetail, err := s.availableChartDetail(ctx, repoName, chartName, request.GetPkgVersion())
	if err != nil {
		return nil, err
	}

	// fix up a couple of fields that don't come from the chart tarball
	repoUrl := repo.Spec.URL
	if repoUrl == "" {
		return nil, status.Errorf(codes.NotFound, "Missing required field spec.url on repository %q", repoName)
	}
	pkgDetail.RepoUrl = repoUrl
	pkgDetail.AvailablePackageRef.Context.Namespace = packageRef.Context.Namespace
	// per https://github.com/kubeapps/kubeapps/pull/3686#issue-1038093832
	pkgDetail.AvailablePackageRef.Context.Cluster = s.kubeappsCluster

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: pkgDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageVersions [%v]", request)
	defer log.Infof("-fluxv2 GetAvailablePackageVersions")

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

	repoName, chartName, err := pkgutils.SplitChartIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart [%s] in namespace [%s]", chartName, namespace)
	repo := types.NamespacedName{Namespace: namespace, Name: repoName}
	chart, err := s.getChart(ctx, repo, chartName)
	if err != nil {
		return nil, err
	} else if chart != nil {
		// found it
		return &corev1.GetAvailablePackageVersionsResponse{
			PackageAppVersions: pkgutils.PackageAppVersionsSummary(
				chart.ChartVersions,
				s.versionsInSummary),
		}, nil
	} else {
		return nil, status.Errorf(codes.Internal, "unable to retrieve versions for chart: [%s]", packageRef.Identifier)
	}
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageSummaries [%v]", request)
	pageOffset, err := paginate.PageOffsetFromInstalledRequest(request)
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
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", pkgRef.GetContext().GetCluster(), pkgRef.GetContext().GetNamespace())
	identifier := pkgRef.GetIdentifier()
	log.Infof("+fluxv2 GetInstalledPackageResourceRefs %s %s", contextMsg, identifier)

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
	log.Infof("+fluxv2 AddPackageRepository [%v]", request)
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no request provided")
	}
	if request.Context == nil || request.Context.Namespace == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Context namespace provided")
	}
	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	if request.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}

	name := types.NamespacedName{Name: request.Name, Namespace: request.Context.Namespace}

	if request.GetNamespaceScoped() {
		return nil, status.Errorf(codes.Unimplemented, "Namespaced-scoped repositories are not supported")
	} else if request.GetType() != "helm" {
		return nil, status.Errorf(codes.Unimplemented, "repository type [%s] not supported", request.GetType())
	}

	if repoRef, err := s.newRepo(ctx, name, request.GetUrl(),
		request.GetInterval(), request.GetTlsConfig(), request.GetAuth()); err != nil {
		return nil, err
	} else {
		return &corev1.AddPackageRepositoryResponse{PackageRepoRef: repoRef}, nil
	}
}

// convenience func mostly used by unit tests
func (s *Server) newBackgroundClientGetter() clientgetter.BackgroundClientGetterFunc {
	return func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return s.clientGetter(ctx, s.kubeappsCluster)
	}
}

func (s *Server) getClient(ctx context.Context, namespace string) (ctrlclient.Client, error) {
	client, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster)
	if err != nil {
		return nil, err
	}
	return ctrlclient.NewNamespacedClient(client, namespace), nil
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *plugins.Plugin {
	return common.GetPluginDetail()
}

// parsePluginConfig parses the input plugin configuration json file and return the
// configuration options.
func parsePluginConfig(pluginConfigPath string) (pkgutils.VersionsInSummary, int32, error) {
	// Note at present VersionsInSummary is the only configurable option for this plugin,
	// and if required this func can be enhanced to return fluxConfig struct

	// In the flux plugin, for example, we are interested in config for the
	// core.packages.v1alpha1 only. So the plugin defines the following struct and parses the config.
	type fluxPluginConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary pkgutils.VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`
	}
	var config fluxPluginConfig

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
