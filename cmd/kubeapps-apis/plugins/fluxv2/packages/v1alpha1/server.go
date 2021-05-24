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
	"k8s.io/client-go/dynamic"

	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"k8s.io/client-go/rest"
	helmrepo "k8s.io/helm/pkg/repo"
	log "k8s.io/klog/v2"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/
	fluxGroup            = "source.toolkit.fluxcd.io"
	fluxVersion          = "v1beta1"
	fluxHelmRepoResource = "helmrepositories"
)

// Server implements the fluxv2 packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *corev1.GetPackageRepositoriesRequest) (*corev1.GetPackageRepositoriesResponse, error) {

	repos, err := getHelmRepos(ctx)
	if err != nil {
		return nil, err
	}

	responseRepos := []*corev1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo := &corev1.PackageRepository{}
		name, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "name")
		if err != nil || !found {
			return nil, fmt.Errorf("required field metadata.name not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Name = name

		namespace, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "namespace")
		// TODO(absoludity): When testing, write failing test for the case of a
		// cluster-scoped object without a namespace, then fix.
		if err != nil || !found {
			return nil, fmt.Errorf("required field metadata.namespace not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Namespace = namespace

		url, found, err := unstructured.NestedString(repoUnstructured.Object, "spec", "url")
		if err != nil || !found {
			return nil, fmt.Errorf("required field spec.url not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
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
	log.Infof("+GetAvailablePackages(namespace=[%s])", request.Namespace)
	repos, err := getHelmRepos(ctx)
	if err != nil {
		return nil, err
	}

	responsePackages := []*corev1.AvailablePackage{}
	for _, repoUnstructured := range repos.Items {
		name, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "name")
		if err != nil || !found {
			log.Errorf("required field metadata.name not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
			// just skip over to the next one
			continue
		}

		// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
		conditions, found, err := unstructured.NestedSlice(repoUnstructured.Object, "status", "conditions")
		if err != nil || !found {
			log.Infof("Skipping packages for repository [%s] because it has not reached 'Ready' state:%w\n%v", name, err, repoUnstructured.Object)
			continue
		}

		ready := false
		for _, conditionUnstructured := range conditions {
			if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
				if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
					if statusString, ok := conditionAsMap["status"]; ok && statusString == "True" {
						if reasonString, ok := conditionAsMap["reason"]; ok && reasonString == "IndexationSucceed" {
							ready = true
							break
						}
					}
				}
			}
		}

		if !ready {
			log.Infof("Skipping packages for repository [%s] because it is not in Ready state:n%v", name, repoUnstructured.Object)
			continue
		}

		url, found, err := unstructured.NestedString(repoUnstructured.Object, "status", "url")
		if err != nil || !found {
			log.Infof("expected field status.url not found on HelmRepository: %w:\n%v", err, repoUnstructured.Object)
			continue
		}

		log.Infof("Found repository: [%s], index URL: [%s]", name, url)
		repoRef := corev1.AvailablePackage_PackageRepositoryReference{
			Name: name,
		}
		// namespace is optional according to https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
		namespace, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "namespace")
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

func getHelmRepos(ctx context.Context) (*unstructured.UnstructuredList, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create incluster config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	repositoryResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepoResource}

	// Currently checks globally. Update to handle namespaced requests (?)
	repos, err := client.Resource(repositoryResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list fluxv2 helmrepositories: %w", err)
	} else {
		// TODO: should we filter out those repos that don't have .status.condition.Ready == True?
		// like we do in GetAvailablePackages()?
		// i.e. should GetAvailableRepos() call semantics be such that only "Ready" repos are returned
		// ongoing slack discussion https://vmware.slack.com/archives/C4HEXCX3N/p1621846518123800
		return repos, nil
	}
}

func readPackagesFromRepoIndex(repoRef *corev1.AvailablePackage_PackageRepositoryReference, indexURL string) ([]*corev1.AvailablePackage, error) {
	// Get the response bytes from the url
	response, err := http.Get(indexURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non OK response code: [%d]", response.StatusCode)
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

	responsePackages := []*corev1.AvailablePackage{}
	for _, entry := range index.Entries {
		// after SortEntires call, entry[0] should be the latest chart, e.g. mariadb 9.3.12
		// while entry[1] might be mariadb 9.3.11, etc. For mariadb, bitnami catalog has almost
		// 200 entries going all the way back to version 2.1.4. So for now let's just keep the latest,
		// not to overwhelm the caller with all these old versions

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
