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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/
	fluxGroup              = "source.toolkit.fluxcd.io"
	fluxVersion            = "v1beta1"
	fluxHelmRepository     = "HelmRepository"
	fluxHelmRepositories   = "helmrepositories"
	fluxHelmRepositoryList = "HelmRepositoryList"
	fluxHelmChart          = "HelmChart"
	fluxHelmCharts         = "helmcharts"
	fluxHelmChartList      = "HelmChartList"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

type clientGetter func(context.Context) (dynamic.Interface, error)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedFluxV2PackagesServiceServer

	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientGetter

	cache *ResourceWatcherCache
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter server.KubernetesConfigGetter) (*Server, error) {
	clientGetter := func(ctx context.Context) (dynamic.Interface, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		config, err := configGetter(ctx)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get dynamic client : %v", err))
		}
		return dynamicClient, nil
	}

	repositoriesGvr := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}
	config := cacheConfig{
		gvr:          repositoriesGvr,
		clientGetter: clientGetter,
		onAdd:        onAddOrModifyRepo,
		onModify:     onAddOrModifyRepo,
		onGet:        onGetRepo,
		onDelete:     onDeleteRepo,
	}
	cache, err := newCache(config)
	if err != nil {
		return nil, err
	}
	return &Server{
		clientGetter: clientGetter,
		cache:        cache,
	}, nil
}

// getDynamicClient returns a dynamic k8s client.
func (s *Server) getDynamicClient(ctx context.Context) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	dynamicClient, err := s.clientGetter(ctx)
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

	if request.Context.Cluster != "" {
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
		repo, err := newPackageRepository(repoUnstructured.Object)
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

	// grpc compiles in getters for you which automatically return a default (empty) struct if the pointer was nil
	if request != nil && request.GetContext().GetCluster() != "" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	pageSize := request.GetPaginationOptions().GetPageSize()
	pageOffset, err := pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"unable to intepret page token %q: %v",
			request.GetPaginationOptions().GetPageToken(), err)
	}

	if s.cache == nil {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"server cache has not been properly initialized")
	}

	repos, err := s.cache.listKeys(request.GetFilterOptions().GetRepositories())
	if err != nil {
		return nil, err
	}

	cachedCharts, err := s.cache.fetchForMultiple(repos)
	if err != nil {
		return nil, err
	}

	packageSummaries, err := filterAndPaginateCharts(request.GetFilterOptions(), pageSize, pageOffset, cachedCharts)
	if err != nil {
		return nil, err
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

	unescapedChartID, err := getUnescapedChartID(packageRef.Identifier)
	if err != nil {
		return nil, err
	}
	packageIdParts := strings.Split(unescapedChartID, "/")

	// check if the repo has been indexed, stored in the cache and requested
	// package is part of it. Otherwise, there is a time window when this scenario can happen:
	// - GetAvailablePackageSummaries() may return {} while a ready repo is being indexed
	//   and said index is cached BUT
	// - GetAvailablePackageDetail() may return full package detail for one of the packages
	// in the repo
	ok, err := s.repoExistsInCache(packageRef.Context.Namespace, packageIdParts[0])
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"no fully indexed repository [%s] in namespace [%s] has been found",
			packageIdParts[0],
			packageRef.Context.Namespace)
	}

	url, err, cleanUp := s.getChartTarball(ctx, packageIdParts[0], packageIdParts[1], packageRef.Context.Namespace, request.PkgVersion)
	if cleanUp != nil {
		defer cleanUp()
	}
	if err != nil {
		return nil, err
	}
	log.Infof("Found chart url: [%s] for chart [%s]", url, packageRef.Identifier)

	pkgDetail, err := availablePackageDetailFromTarball(packageRef.Identifier, url)
	if err != nil {
		return nil, err
	}

	// fix up namespace as it is not coming from chart tarball itself
	pkgDetail.AvailablePackageRef.Context.Namespace = packageRef.Context.Namespace

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: pkgDetail,
	}, nil
}

// returns the url from which chart .tgz can be downloaded
// here chartVersion string, if specified at all, should be specific, like "14.4.0",
// not an expression like ">14 <15"
func (s *Server) getChartTarball(ctx context.Context, repoName string, chartName string, namespace string, chartVersion string) (url string, err error, cleanUp func()) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return "", err, nil
	}

	chartsResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmCharts,
	}

	resourceIfc := client.Resource(chartsResource).Namespace(namespace)

	// see if we the chart already exists
	// TODO (gfichtenholt):
	// see https://github.com/kubeapps/kubeapps/pull/2915
	// for context. It'd be better if we could filter on server-side. The problem is the set of supported
	// fields in FieldSelector is very small. things like "spec.chart" or "status.artifact.revision" are
	// certainly not supported.
	// see
	//  - kubernetes/client-go#713 and
	//  - https://github.com/flant/shell-operator/blob/8fa3c3b8cfeb1ddb37b070b7a871561fdffe788b///HOOKS.md#fieldselector and
	//  - https://github.com/kubernetes/kubernetes/issues/53459
	chartList, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err, nil
	}

	url, err = findUrlForChartInList(chartList, repoName, chartName, chartVersion)
	if err != nil {
		return "", err, nil
	} else if url != "" {
		return url, nil, nil
	}

	// did not find the chart, need to create
	// see https://fluxcd.io/docs/components/source/helmcharts/
	// TODO (gfichtenholt)
	// 1. HelmChart object needs to be co-located in the same namespace as the HelmRepository it is referencing.
	// 2. flux impersonates a "super" user when doing this (see fluxv2 plug-in specific notes at the end of
	//	design doc). We should probably be doing simething similar to avoid RBAC-related problems
	unstructuredChart := newFluxHelmChart(chartName, repoName, chartVersion)

	newChart, err := resourceIfc.Create(ctx, &unstructuredChart, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("Error creating chart: %v\n%v", err, unstructuredChart)
		return "", err, nil
	}

	log.Infof("Created chart: [%v]", prettyPrintMap(newChart.Object))

	// Delete the created helm chart regardless of success or failure. At the end of
	// GetAvailablePackageDetail(), we've already collected the information we need,
	// so why leave a flux chart chart object hanging around?
	// Over time, they could accumulate to a very large number...
	cleanUp = func() {
		err = resourceIfc.Delete(ctx, newChart.GetName(), metav1.DeleteOptions{})
		if err != nil {
			log.Errorf("Failed to delete flux helm chart [%v]", prettyPrintMap(newChart.Object))
		}
	}

	watcher, err := resourceIfc.Watch(ctx, metav1.ListOptions{
		ResourceVersion: newChart.GetResourceVersion(),
	})
	if err != nil {
		log.Errorf("Error creating watch: %v\n%v", err, unstructuredChart)
		return "", err, cleanUp
	}

	// wait til wait until flux reconciles and we have chart url available
	// TODO (gfichtenholt) note that, unlike with ResourceWatcherCache, the
	// wait time window is very short here so I am not employing the RetryWatcher
	// technique here for now
	url, err = waitUntilChartPullComplete(ctx, watcher)
	watcher.Stop()
	// only the caller should call cleanUp() when it's done with the url,
	// if we call it here, the caller will end up with a dangling link
	return url, err, cleanUp
}

// namespace maybe "", in which case repositories from all namespaces are returned
func (s *Server) listReposInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return nil, err
	}

	repositoriesResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	repos, err := client.Resource(repositoriesResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list fluxv2 helmrepositories: %v", err)
	} else {
		// TODO (gfichtenholt): should we filter out those repos that don't have .status.condition.Ready == True?
		// like we do in GetAvailablePackageSummaries()?
		// i.e. should GetAvailableRepos() call semantics be such that only "Ready" repos are returned
		// ongoing slack discussion https://vmware.slack.com/archives/C4HEXCX3N/p1621846518123800
		return repos, nil
	}
}

func (s *Server) repoExistsInCache(namespace, repoName string) (bool, error) {
	if s.cache == nil {
		return false, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	}

	repos, err := s.cache.listKeys([]string{repoName})
	if err != nil {
		return false, err
	}

	for _, key := range repos {
		thisNamespace, thisName, err := s.cache.fromRedisKey(key)
		if err != nil {
			return false, err
		}
		if thisNamespace == namespace && thisName == repoName {
			return true, nil
		}
	}
	return false, nil
}
