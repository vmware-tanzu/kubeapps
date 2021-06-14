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
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"

	"github.com/kubeapps/kubeapps/pkg/helm"
	tar "github.com/kubeapps/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/
	fluxGroup              = "source.toolkit.fluxcd.io"
	fluxVersion            = "v1beta1"
	fluxHelmRepository     = "helmrepository"
	fluxHelmRepositories   = "helmrepositories"
	fluxHelmRepositoryList = "HelmRepositoryList"
	fluxHelmChart          = "HelmChart"
	fluxHelmCharts         = "helmcharts"
	fluxHelmChartList      = "HelmChartList"
)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedFluxV2PackagesServiceServer

	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter func(context.Context) (dynamic.Interface, error)
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(clientGetter func(context.Context) (dynamic.Interface, error)) *Server {
	return &Server{
		clientGetter: clientGetter,
	}
}

// getClient ensures a client getter is available and uses it to return the client.
func (s *Server) GetClient(ctx context.Context) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	client, err := s.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return client, nil
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
func (s *Server) GetPackageRepositories(ctx context.Context, request *v1alpha1.GetPackageRepositoriesRequest) (*v1alpha1.GetPackageRepositoriesResponse, error) {
	log.Infof("+GetPackageRepositories(request: [%v])", request)

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
		obj := repoUnstructured.Object
		repo := &v1alpha1.PackageRepository{}
		name, found, err := unstructured.NestedString(obj, "metadata", "name")
		if err != nil || !found {
			return nil, status.Errorf(
				codes.Internal,
				"required field metadata.name not found on HelmRepository: %v:\n%v", err, obj)
		}
		repo.Name = name

		// namespace is optional according to https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
		namespace, found, err := unstructured.NestedString(obj, "metadata", "namespace")

		// TODO(absoludity): When testing, write failing test for the case of a
		// cluster-scoped object without a namespace, then fix.
		if err == nil && found {
			repo.Namespace = namespace
		}

		url, found, err := unstructured.NestedString(obj, "spec", "url")
		if err != nil || !found {
			return nil, status.Errorf(
				codes.Internal, "required field spec.url not found on HelmRepository: %v:\n%v", err, obj)
		}
		repo.Url = url

		responseRepos = append(responseRepos, repo)
	}
	return &v1alpha1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

// GetAvailablePackageSummaries streams the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+GetAvailablePackageSummaries(request: [%v])", request)

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

	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, unstructuredRepo := range repos.Items {
		obj := unstructuredRepo.Object
		name, found, err := unstructured.NestedString(obj, "metadata", "name")
		if err != nil || !found {
			log.Errorf("required field metadata.name not found on HelmRepository: %w:\n%v", err, obj)
			// just skip over to the next one
			continue
		}

		ready, err := isRepoReady(obj)
		if err != nil || !ready {
			log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state:%v\n%v", name, err, obj)
			continue
		}

		url, found, err := unstructured.NestedString(obj, "status", "url")
		if err != nil || !found {
			log.Infof("expected field status.url not found on HelmRepository [%s]: %v:\n%v", name, err, obj)
			continue
		}

		log.Infof("Found repository: [%s], index URL: [%s]", name, url)
		repo := v1alpha1.PackageRepository{
			Name: name,
		}
		// namespace is optional according to https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
		namespace, found, err := unstructured.NestedString(obj, "metadata", "namespace")
		if err == nil && found {
			repo.Namespace = namespace
		}

		repoPackages, err := readPackagesFromRepoIndex(&repo, url)
		if err != nil {
			// just skip this repo
			log.Errorf("Failed to read packages for repository [%s] due to %v", name, err)
		} else {
			responsePackages = append(responsePackages, repoPackages...)
		}
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: responsePackages,
	}, nil
}

// GetAvailablePackageDetail returns the package metadata managed by the 'fluxv2' plugin
func (s *Server) GetAvailablePackageDetail(ctx context.Context, request *corev1.GetAvailablePackageDetailRequest) (*corev1.GetAvailablePackageDetailResponse, error) {
	log.Infof("+GetAvailablePackageDetail(request: [%v])", request)

	if request == nil || request.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "No request AvailablePackageRef provided")
	}

	if request.Version != "" {
		return nil, status.Errorf(
			codes.Unimplemented,
			"Not supported yet: version: [%v]",
			request.Version)
	}

	url, err := s.pullChartTarball(ctx, request.AvailablePackageRef)
	if err != nil {
		return nil, err
	}
	log.Infof("Found chart url: [%s]", *url)

	// unzip and untar .tgz file
	// TODO: userAgent, authz and netClient w/TLS config similar to asset-syncer utils.initNetClient(),
	// see if we can re-factor code to reuse in both places
	detail, err := tar.FetchChartDetailFromTarball(request.AvailablePackageRef.Identifier, *url, "", "", &http.Client{})
	if err != nil {
		return nil, err
	}

	return &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: &corev1.AvailablePackageDetail{
			LongDescription: detail[chart.ReadmeKey],
		},
	}, nil
}

func (s *Server) pullChartTarball(ctx context.Context, packageRef *corev1.AvailablePackageReference) (*string, error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	chartsResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmCharts,
	}

	// TODO: use the namespace from the request
	resourceIfc := client.Resource(chartsResource).Namespace("default")

	// see if we the chart already exists
	// TODO:
	// https://github.com/kubeapps/kubeapps/pull/2915
	// @minelson says: You should be able to use the metav1.ListOptions{} above
	// to specify filtering using the FieldSelector.
	// More info at https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/ .
	// You may even be able to specify that the pull should be complete to be included (ie.
	// that status conditions ready is true), not sure, but that'd be nice.
	// @gfichtenholt As I remember I did try a few things here. The problem is the set of supported
	// fields in FieldSelector is very small. things like spec.chart are certainly not supported.
	// see kubernetes/client-go#713 and https://github.com/flant/shell-operator/blob/8fa3c3b8cfeb1ddb37b070b7a871561fdffe788b/// HOOKS.md#fieldselector. Nevertheless, I made a TODO comment here just as you suggested to see
	// if I can improve later
	chartList, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// TODO it'd be better if we could filter on server-side
	for _, unstructuredChart := range chartList.Items {
		chartName, found, err := unstructured.NestedString(unstructuredChart.Object, "spec", "chart")
		// TODO compare chart versions too
		if err == nil && found && chartName == packageRef.Identifier {
			done, err := isChartPullComplete(&unstructuredChart)
			if err != nil {
				return nil, err
			} else if done {
				url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
				if err != nil || !found {
					return nil, status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
				}
				log.Infof("Found existing HelmChart for: [%s]", packageRef.Identifier)
				return &url, nil
			}
			// TODO waitUntilChartPullComplete
		}
	}

	// did not find the chart, need to create
	// see https://fluxcd.io/docs/components/source/helmcharts/
	unstructuredChart := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmChart,
			"metadata": map[string]interface{}{
				"generateName": fmt.Sprintf("%s-", packageRef.Identifier),
			},
			"spec": map[string]interface{}{
				"chart": packageRef.Identifier,
				"sourceRef": map[string]interface{}{
					"name": packageRef.Identifier,
					"kind": "HelmRepository",
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

	return waitUntilChartPullComplete(watcher)
}

// TODO:
// https://github.com/kubeapps/kubeapps/pull/2915
// @minelson says Nice - great use of both channel and the watcher to handle the
// async request. Note that this effectively makes pullChartTarball a blocking call,
// which is fine in it's current use (where it's only called in a context of one chart).
// Not for now, but in the future you might instead want to consider something like
// passing a results channel (of string urls) to pullChartTarball, so it returns
// immediately and you wait on the results channel at the call-site, which would mean
// you could call it for 20 different charts and just wait for the results to come in
//  whatever order they happen to take, rather than serially.
func waitUntilChartPullComplete(watcher watch.Interface) (*string, error) {
	ch := watcher.ResultChan()
	// LISTEN TO CHANNEL
	for {
		event := <-ch
		if event.Type == watch.Modified {
			unstructuredChart, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				return nil, status.Errorf(codes.Internal, "Could not cast to unstructured.Unstructured")
			}

			done, err := isChartPullComplete(unstructuredChart)
			if err != nil {
				return nil, err
			} else if done {
				url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
				if err != nil || !found {
					return nil, status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
				}
				return &url, nil
			}
		} else {
			// TODO handle other kinds of events
			return nil, status.Errorf(codes.Internal, "got unexpected event: %v", event)
		}
	}
}

func (s *Server) getHelmRepos(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	repositoriesResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	var resource dynamic.NamespaceableResourceInterface = client.Resource(repositoriesResource)
	var resourceIfc dynamic.ResourceInterface = resource
	if namespace != "" {
		resourceIfc = resource.Namespace(namespace)
	}
	repos, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list fluxv2 helmrepositories: %v", err)
	} else {
		// TODO: should we filter out those repos that don't have .status.condition.Ready == True?
		// like we do in GetAvailablePackageSummaries()?
		// i.e. should GetAvailableRepos() call semantics be such that only "Ready" repos are returned
		// ongoing slack discussion https://vmware.slack.com/archives/C4HEXCX3N/p1621846518123800
		return repos, nil
	}
}

func isRepoReady(obj map[string]interface{}) (bool, error) {
	// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
	conditions, found, err := unstructured.NestedSlice(obj, "status", "conditions")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}

	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				if statusString, ok := conditionAsMap["status"]; ok && statusString == "True" {
					// note that the current doc on https://fluxcd.io/docs/components/source/helmrepositories/
					// incorrectly states the example status reason as "IndexationSucceeded".
					// The actual string is "IndexationSucceed"
					if reasonString, ok := conditionAsMap["reason"]; ok && reasonString == "IndexationSucceed" {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

// TODO: As above, hopefully this fn isn't required if we can only list charts that we know are ready.
func isChartPullComplete(unstructuredChart *unstructured.Unstructured) (bool, error) {
	// see docs at https://fluxcd.io/docs/components/source/helmcharts/
	conditions, found, err := unstructured.NestedSlice(unstructuredChart.Object, "status", "conditions")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}

	// check if ready=True
	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				if statusString, ok := conditionAsMap["status"]; ok && statusString == "True" {
					if reasonString, ok := conditionAsMap["reason"]; ok && reasonString == "ChartPullSucceeded" {
						return true, nil
					}
				}
				// TODO handle the case when chart pull fails
			}
		}
	}
	return false, nil
}

func readPackagesFromRepoIndex(repo *v1alpha1.PackageRepository, indexURL string) ([]*corev1.AvailablePackageSummary, error) {
	// todo set up httpClient properly with userAgent and TLS config
	// similar to what is done in asset syncer
	bytes, err := httpclient.Get(indexURL, &http.Client{}, map[string]string{})
	if err != nil {
		return nil, err
	}

	modelRepo := &models.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Url,
		Type:      "helm",
	}
	charts, err := helm.ChartsFromIndex(bytes, modelRepo, true)
	if err != nil {
		return nil, err
	}

	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, chart := range charts {
		pkg := &corev1.AvailablePackageSummary{
			DisplayName:   chart.Name,
			LatestVersion: chart.ChartVersions[0].Version,
			IconUrl:       chart.Icon,
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Context:    &corev1.Context{Namespace: repo.Namespace},
				Identifier: chart.ID,
			},
		}
		responsePackages = append(responsePackages, pkg)
	}
	return responsePackages, nil
}
