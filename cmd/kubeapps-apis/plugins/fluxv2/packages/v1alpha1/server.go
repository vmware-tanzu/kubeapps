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
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedFluxV2PackagesServiceServer

	// kubeappsCluster specifies the cluster on which Kubeapps is installed.
	kubeappsCluster string
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter       common.ClientGetterFunc
	actionConfigGetter common.HelmActionConfigGetterFunc

	repoCache  *cache.NamespacedResourceWatcherCache
	chartCache *ChartCache
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, kubeappsCluster string) (*Server, error) {
	log.Infof("+fluxv2 NewServer(kubeappsCluster: [%v])", kubeappsCluster)
	repositoriesGvr := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	if redisCli, err := common.NewRedisClientFromEnv(); err != nil {
		return nil, err
	} else if chartCache, err := NewChartCache(redisCli); err != nil {
		return nil, err
	} else {
		repoCacheConfig := cache.NamespacedResourceWatcherCacheConfig{
			Gvr:          repositoriesGvr,
			ClientGetter: common.NewBackgroundClientGetter(),
			OnAddFunc:    chartCache.wrapOnAddFunc(onAddRepo, onGetRepo),
			OnModifyFunc: onModifyRepo,
			OnGetFunc:    onGetRepo,
			OnDeleteFunc: onDeleteRepo,
		}
		if repoCache, err := cache.NewNamespacedResourceWatcherCache(repoCacheConfig, redisCli); err != nil {
			return nil, err
		} else {
			return &Server{
				clientGetter:       common.NewClientGetter(configGetter, kubeappsCluster),
				actionConfigGetter: common.NewHelmActionConfigGetter(configGetter, kubeappsCluster),
				repoCache:          repoCache,
				chartCache:         chartCache,
				kubeappsCluster:    kubeappsCluster,
			}, nil
		}
	}
}

// getDynamicClient returns a dynamic k8s client.
func (s *Server) getDynamicClient(ctx context.Context) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	dynamicClient, _, err := s.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}
	return dynamicClient, nil
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

// GetPackageRepositories returns the package repositories based on the request.
// note that this func currently returns ALL repositories, not just those in 'ready' (reconciled) state
func (s *Server) GetPackageRepositories(ctx context.Context, request *v1alpha1.GetPackageRepositoriesRequest) (*v1alpha1.GetPackageRepositoriesResponse, error) {
	log.Infof("+fluxv2 GetPackageRepositories(request: [%v])", request)

	if request == nil || request.Context == nil {
		return nil, status.Errorf(codes.InvalidArgument, "no context provided")
	}

	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	repos, err := s.listReposInCluster(ctx, request.Context.Namespace)
	if err != nil {
		return nil, err
	}

	responseRepos := []*v1alpha1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo, err := packageRepositoryFromUnstructured(repoUnstructured.Object)
		if err != nil {
			return nil, err
		}
		responseRepos = append(responseRepos, repo)
	}
	return &v1alpha1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

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

	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := common.PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	}

	charts, err := s.getChartsForRepos(ctx, request.GetFilterOptions().GetRepositories())
	if err != nil {
		return nil, err
	}

	packageSummaries, err := filterAndPaginateCharts(request.GetFilterOptions(), pageSize, pageOffset, charts)
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

	unescapedChartID, err := common.GetUnescapedChartID(packageRef.Identifier)
	if err != nil {
		return nil, err
	}
	packageIdParts := strings.Split(unescapedChartID, "/")

	// check specified repo exists and is in ready state
	repo := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: packageIdParts[0]}

	// this verifies that the repo exists
	repoUnstructured, err := s.getRepoInCluster(ctx, repo)
	if err != nil {
		return nil, err
	}

	tarUrl, cleanUp, err := s.getChartTarballUrl(ctx, repoUnstructured, packageIdParts[1], request.PkgVersion)
	if cleanUp != nil {
		defer cleanUp()
	}
	if err != nil {
		return nil, err
	}
	log.V(4).Infof("Found chart url: [%s] for chart [%s]", tarUrl, packageRef.Identifier)

	pkgDetail, err := availablePackageDetailFromTarball(packageRef.Identifier, tarUrl)
	if err != nil {
		return nil, err
	}

	// fix up a couple of fields that don't come from the chart tarball
	repoUrl, found, err := unstructured.NestedString(repoUnstructured.Object, "spec", "url")
	if err != nil || !found {
		return nil, status.Errorf(codes.NotFound, "Missing required field spec.url on repository %q", repo)
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

	unescapedChartID, err := common.GetUnescapedChartID(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	log.Infof("Requesting chart [%s] (latest version) in ns [%s]", unescapedChartID, namespace)
	packageIdParts := strings.Split(unescapedChartID, "/")
	repo := types.NamespacedName{Namespace: namespace, Name: packageIdParts[0]}
	chart, err := s.getChart(ctx, repo, packageIdParts[1])
	if err != nil {
		return nil, err
	} else if chart != nil {
		// found it
		return &corev1.GetAvailablePackageVersionsResponse{
			PackageAppVersions: packageAppVersionsSummary(chart.ChartVersions),
		}, nil
	} else {
		return nil, status.Errorf(codes.Internal, "unable to retrieve versions for chart: [%s]", packageRef.Identifier)
	}
}

// GetInstalledPackageSummaries returns the installed packages managed by the 'fluxv2' plugin
func (s *Server) GetInstalledPackageSummaries(ctx context.Context, request *corev1.GetInstalledPackageSummariesRequest) (*corev1.GetInstalledPackageSummariesResponse, error) {
	log.Infof("+fluxv2 GetInstalledPackageSummaries [%v]", request)
	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := common.PageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	}

	cluster := request.GetContext().GetCluster()
	if cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			cluster)
	}

	installedPkgSummaries, err := s.paginatedInstalledPkgSummaries(ctx, request.GetContext().GetNamespace(), pageSize, pageOffset)
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

	name := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: packageRef.Identifier}
	pkgDetail, err := s.installedPackageDetail(ctx, name)
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
	log.Infof("+fluxv2 GetResourceRefs %s %s", contextMsg, identifier)

	namespace := pkgRef.GetContext().GetNamespace()

	actionConfig, err := s.actionConfigGetter(ctx, request.GetInstalledPackageRef().GetContext().GetNamespace())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	// Grab the released manifest from the release.
	// TODO(minelson): We're currently getting the resource refs for a package
	// install by checking the helm manifest, as we do for the helm plugin. With
	// certain assumptions about the RBAC of the Kubeapps user, we may be able
	// to instead query for labelled resources. See the discussion following for
	// more details:
	// https://github.com/kubeapps/kubeapps/pull/3811#issuecomment-977689570
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
