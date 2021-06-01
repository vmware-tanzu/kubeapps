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
	"io/ioutil"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
	helmrepo "k8s.io/helm/pkg/repo"
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
)

// these should be constants but alas go does not allow const structs
var (
	repositoriesResource = schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories}

	chartsResource = schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmCharts,
	}
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
func NewServer() *Server {
	return &Server{
		clientGetter: clientForRequestContext,
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
func (s *Server) GetPackageRepositories(ctx context.Context, request *corev1.GetPackageRepositoriesRequest) (*corev1.GetPackageRepositoriesResponse, error) {
	log.Infof("+GetPackageRepositories(namespace=[%s], cluster=[%s])", request.Namespace, request.Cluster)
	if request.Cluster != "" {
		return nil, status.Errorf(codes.Unimplemented, "Not supported yet")
	}

	repos, err := s.getHelmRepos(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}

	responseRepos := []*corev1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		obj := repoUnstructured.Object
		repo := &corev1.PackageRepository{}
		name, found, err := unstructured.NestedString(obj, "metadata", "name")
		if err != nil || !found {
			return nil, status.Errorf(codes.Internal, "required field metadata.name not found on HelmRepository: %v:\n%v", err, obj)
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
	return &corev1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

// GetAvailablePackages streams the available packages based on the request.
func (s *Server) GetAvailablePackages(ctx context.Context, request *corev1.GetAvailablePackagesRequest) (*corev1.GetAvailablePackagesResponse, error) {
	log.Infof("+GetAvailablePackages(namespace=[%s], cluster=[%s])", request.Namespace, request.Cluster)
	if request.Cluster != "" {
		return nil, status.Errorf(codes.Unimplemented, "Not supported yet")
	}

	repos, err := s.getHelmRepos(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}

	responsePackages := []*corev1.AvailablePackage{}
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
		repoRef := corev1.AvailablePackage_PackageRepositoryReference{
			Name: name,
		}
		// namespace is optional according to https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
		namespace, found, err := unstructured.NestedString(obj, "metadata", "namespace")
		if err == nil && found {
			repoRef.Namespace = namespace
		}

		repoPackages, err := readPackagesFromRepoIndex(&repoRef, url)
		if err != nil {
			// just skip this repo
			log.Errorf("Failed to read packages for repository [%s] due to %v", name, err)
		} else {
			responsePackages = append(responsePackages, repoPackages...)
		}
	}
	return &corev1.GetAvailablePackagesResponse{
		Packages: responsePackages,
	}, nil
}

// GetPackageMeta streams the package metadata based on the request.
func (s *Server) GetPackageMeta(ctx context.Context, request *corev1.GetPackageMetaRequest) (*corev1.GetPackageMetaResponse, error) {
	log.Infof("+GetPackageMeta()")

	unstructuredChart := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmChart,
			"metadata": map[string]interface{}{
				"generateName": "redis-",
			},
			"spec": map[string]interface{}{
				"chart":   "redis",
				"version": "10.5.x",
				"sourceRef": map[string]interface{}{
					"name": "bitnami",
					"kind": "HelmRepository",
				},
				"interval": "10m",
			},
		},
	}

	url, err := s.createAndPullChart(ctx, &unstructuredChart)
	if err != nil {
		return nil, err
	}
	log.Infof("Found chart url: [%s]", *url)

	return nil, status.Errorf(codes.Unimplemented, "not implemented yet")
}

func (s *Server) createAndPullChart(ctx context.Context, unstructuredChart *unstructured.Unstructured) (*string, error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	resourceIfc := client.Resource(chartsResource).Namespace("default")

	newChart, err := resourceIfc.Create(ctx, unstructuredChart, metav1.CreateOptions{})
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

func waitUntilChartPullComplete(watcher watch.Interface) (*string, error) {
	ch := watcher.ResultChan()
	// LISTEN TO CHANNEL
	for {
		event := <-ch
		// check if ready=True
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

// clientForRequestContext returns a k8s client for use during interactions with the cluster.
// This will be updated to use the user credential from the request context but for now
// simply returns th in-cluster config (which is linked to a service-account with demo RBAC).
func clientForRequestContext(ctx context.Context) (dynamic.Interface, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get client config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	return client, nil
}

func (s *Server) getHelmRepos(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
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
		// like we do in GetAvailablePackages()?
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

//
// TODO the semantics of this really should be do we need to keep polling or not
//
func isChartPullComplete(unstructuredChart *unstructured.Unstructured) (bool, error) {
	// see docs at https://fluxcd.io/docs/components/source/helmcharts/
	conditions, found, err := unstructured.NestedSlice(unstructuredChart.Object, "status", "conditions")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}

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

func readPackagesFromRepoIndex(repoRef *corev1.AvailablePackage_PackageRepositoryReference, indexURL string) ([]*corev1.AvailablePackage, error) {
	index, err := getHelmIndexFileFromURL(indexURL)
	if err != nil {
		return nil, err
	}

	responsePackages := []*corev1.AvailablePackage{}
	for _, entry := range index.Entries {
		// note that 'entry' itself is an array of chart versions
		// after index.SortEntires() call, it looks like there is only one entry per package,
		// and entry[0] should be the most recent chart version, e.g. Name: "mariadb" Version: "9.3.12"
		// while the rest of the elements in the entry array keep track of previous chart versions, e.g.
		// "mariadb" version "9.3.11", "9.3.10", etc. For entry "mariadb", bitnami catalog has
		// almost 200 chart versions going all the way back many years to version "2.1.4".
		// So for now, let's just keep track of the latest, not to overwhelm the caller with
		// all these outdated versions
		if entry[0].GetDeprecated() {
			log.Infof("skipping deprecated chart: [%s]", entry[0].Name)
			continue
		}

		pkg := &corev1.AvailablePackage{
			Name:       entry[0].Name,
			Version:    entry[0].Version,
			IconUrl:    entry[0].Icon,
			Repository: repoRef,
		}
		responsePackages = append(responsePackages, pkg)
	}
	return responsePackages, nil
}

func getHelmIndexFileFromURL(indexURL string) (*helmrepo.IndexFile, error) {
	log.Infof("+getHelmIndexFileFromURL(%s) 1", indexURL)
	// Get the response bytes from the url
	response, err := http.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, status.Errorf(codes.FailedPrecondition, "received non OK response code: [%d]", response.StatusCode)
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var index helmrepo.IndexFile
	err = yaml.Unmarshal(contents, &index)
	if err != nil {
		return nil, err
	}
	index.SortEntries()
	log.Infof("-getHelmIndexFileFromURL(%s)", indexURL)
	return &index, nil
}
