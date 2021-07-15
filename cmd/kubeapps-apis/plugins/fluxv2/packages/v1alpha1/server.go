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
	"k8s.io/client-go/kubernetes"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"

	tar "github.com/kubeapps/kubeapps/pkg/tarutil"
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

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedFluxV2PackagesServiceServer

	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter server.KubernetesClientGetter

	cache *ResourceWatcherCache
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(clientGetter server.KubernetesClientGetter) (*Server, error) {
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

// getClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}
	return typedClient, dynamicClient, nil
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
		return nil, status.Errorf(codes.InvalidArgument, "No context provided")
	}

	if request.Context.Cluster != "" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"Not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	repos, err := s.getHelmRepos(ctx, request.Context.Namespace)
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

	packageSummaries, err := filterAndPaginateCharts(request.GetFilterOptions(), int(pageSize), pageOffset, cachedCharts)
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
		AvailablePackagesSummaries: packageSummaries,
		NextPageToken:              nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	log.Infof("+fluxv2 GetAvailablePackageDetail(request: [%v])", request)

	if request == nil || request.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "No request AvailablePackageRef provided")
	}

	packageRef := request.AvailablePackageRef
	// flux CRDs require a namespace, cluster-wide resources are not supported
	if packageRef.Context == nil || len(packageRef.Context.Namespace) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "AvailablePackageReference is missing required 'namespace' field")
	}
	packageIdParts := strings.Split(packageRef.Identifier, "/")
	if len(packageIdParts) != 2 {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid package ref identifier: [%s]", packageRef.Identifier)
	}

	if request.PkgVersion != "" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"Not supported yet: version: [%v]",
			request.PkgVersion)
	}

	// TODO (gfichtenholt) check if the repo has been indexed, stored in the cache and requested
	// package is part of it. Otherwise, there is a time window when this scenario can happen:
	// - GetAvailablePackageSummaries may return {} while a ready repo is being indexed BUT
	// - GetAvailablePackageDetail may return package detail
	url, err := s.pullChartTarball(ctx, packageIdParts[0], packageIdParts[1], packageRef.Context.Namespace)
	if err != nil {
		return nil, err
	}
	log.Infof("Found chart url: [%s]", *url)

	// unzip and untar .tgz file
	// no need to provide authz, userAgent or any of the TLS details, as we are pulling .tgz file from
	// local cluster, not remote repo.
	// E.g. http://source-controller.flux-system.svc.cluster.local./helmchart/default/redis-j6wtx/redis-latest.tgz
	// Flux does the hard work of pulling the bits from remote repo
	// based on secretRef associated with HelmRepository, if applicable
	detail, err := tar.FetchChartDetailFromTarball(packageRef.Identifier, *url, "", "", httpclient.New())
	if err != nil {
		return nil, err
	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: &corev1.AvailablePackageDetail{
			AvailablePackageRef: packageRef, // copy just for now
			Name:                packageIdParts[1],
			LongDescription:     detail[chart.ReadmeKey],
		},
	}, nil
}

// returns the url from which chart .tgz can be downloaded
func (s *Server) pullChartTarball(ctx context.Context, repoName string, chartName string, namespace string) (*string, error) {
	_, client, err := s.GetClients(ctx)
	if err != nil {
		return nil, err
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
	// fields in FieldSelector is very small. things like "spec.chart" are certainly not supported.
	// see
	//  - kubernetes/client-go#713 and
	//  - https://github.com/flant/shell-operator/blob/8fa3c3b8cfeb1ddb37b070b7a871561fdffe788b///HOOKS.md#fieldselector and
	//  - https://github.com/kubernetes/kubernetes/issues/53459
	chartList, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, unstructuredChart := range chartList.Items {
		thisChartName, found, err := unstructured.NestedString(unstructuredChart.Object, "spec", "chart")
		thisRepoName, found2, err2 := unstructured.NestedString(unstructuredChart.Object, "spec", "sourceRef", "name")

		// TODO (gfichtenholt) compare chart versions too
		if err == nil && err2 == nil && found && found2 && repoName == thisRepoName && chartName == thisChartName {
			done, err := isChartPullComplete(&unstructuredChart)
			if err != nil {
				return nil, err
			} else if done {
				url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
				if err != nil || !found {
					return nil, status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
				}
				log.Infof("Found existing HelmChart for: [%s/%s]", repoName, chartName)
				return &url, nil
			}
			// TODO (gfichtenholt) waitUntilChartPullComplete?
		}
	}

	// did not find the chart, need to create
	// see https://fluxcd.io/docs/components/source/helmcharts/
	// TODO (gfichtenholt)
	// 1. HelmChart object needs to be co-located in the same namespace as the HelmRepository it is referencing.
	// 2. flux impersonates a "super" user when doing this (see fluxv2 plug-in specific notes at the end of
	//	design doc). We should probably be doing simething similar to avoid RBAC-related problems
	unstructuredChart := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmChart,
			"metadata": map[string]interface{}{
				"generateName": fmt.Sprintf("%s-", chartName),
			},
			"spec": map[string]interface{}{
				"chart": chartName,
				"sourceRef": map[string]interface{}{
					"name": repoName,
					"kind": fluxHelmRepository,
				},
				"interval": "10m",
			},
		},
	}

	newChart, err := resourceIfc.Create(ctx, &unstructuredChart, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("error creating chart: %v\n%v", err, unstructuredChart)
		return nil, err
	}

	log.Infof("created chart: [%v]", newChart)

	// wait until flux reconciles
	watcher, err := resourceIfc.Watch(ctx, metav1.ListOptions{
		ResourceVersion: newChart.GetResourceVersion(),
	})
	if err != nil {
		log.Errorf("error creating watch: %v\n%v", err, unstructuredChart)
		return nil, err
	}

	// wait til we have chart url available
	return waitUntilChartPullComplete(watcher)
}

// namespace maybe "", in which case repositories from all namespaces are returned
func (s *Server) getHelmRepos(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	_, client, err := s.GetClients(ctx)
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
