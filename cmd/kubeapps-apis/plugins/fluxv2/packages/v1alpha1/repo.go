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
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	log "k8s.io/klog/v2"
)

const (
	// see docs at https://fluxcd.io/docs/components/source/ and
	// https://fluxcd.io/docs/components/helm/api/
	fluxGroup              = "source.toolkit.fluxcd.io"
	fluxVersion            = "v1beta1"
	fluxHelmRepository     = "HelmRepository"
	fluxHelmRepositories   = "helmrepositories"
	fluxHelmRepositoryList = "HelmRepositoryList"
)

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

	if repos, err := client.Resource(repositoriesResource).Namespace(namespace).List(ctx, metav1.ListOptions{}); err != nil {
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
		thisNamespace, thisName, err := s.cache.fromKey(key)
		if err != nil {
			return false, err
		}
		if thisNamespace == namespace && thisName == repoName {
			return true, nil
		}
	}
	return false, nil
}

//
// repo-related utilities
//

func isRepoReady(unstructuredRepo map[string]interface{}) bool {
	// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
	// Confirm the state we are observing is for the current generation
	if !checkGeneration(unstructuredRepo) {
		return false
	}

	completed, success, _ := checkStatusReady(unstructuredRepo)
	return completed && success
}

func indexOneRepo(unstructuredRepo map[string]interface{}) ([]models.Chart, error) {
	startTime := time.Now()

	repo, err := newPackageRepository(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	// this is just a future-proofing sanity check.
	// At present, there is only one caller of indexOneRepo() and this check is already done by it,
	// so this should never really happen
	ready := isRepoReady(unstructuredRepo)
	if !ready {
		return nil, status.Errorf(codes.Internal,
			"cannot index repository [%s] because it is not in 'Ready' state. error: %v",
			repo.Name,
			err)
	}

	indexUrl, found, err := unstructured.NestedString(unstructuredRepo, "status", "url")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal,
			"expected field status.url not found on HelmRepository\n[%s], error %v",
			repo.Name, err)
	}

	log.Infof("indexOneRepo: [%s], index URL: [%s]", repo.Name, indexUrl)

	// no need to provide authz, userAgent or any of the TLS details, as we are reading index.yaml file from
	// local cluster, not some remote repo.
	// e.g. http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index.yaml
	// Flux does the hard work of pulling the index file from remote repo
	// into local cluster based on secretRef associated with HelmRepository, if applicable
	bytes, err := httpclient.Get(indexUrl, httpclient.New(), map[string]string{})
	if err != nil {
		return nil, err
	}

	modelRepo := &models.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Url,
		Type:      "helm",
	}

	// this is potentially a very expensive operation for large repos like 'bitnami'
	// shallow = true  => 8-9 sec
	// shallow = false => 12-13 sec, so deep copy adds 50% to cost, but we need it to
	// for GetAvailablePackageVersions()
	charts, err := helm.ChartsFromIndex(bytes, modelRepo, false)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	msg := fmt.Sprintf("indexOneRepo: indexed [%d] packages in repository [%s] in [%d] ms", len(charts), repo.Name, duration.Milliseconds())
	if len(charts) > 0 {
		log.Info(msg)
	} else {
		// this is kind of a red flag - an index with 0 charts, most likely contents of index.yaml is
		// messed up and didn't parse but the helm library didn't raise an error
		log.Warning(msg)
	}
	return charts, nil
}

func newPackageRepository(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, namespace, err := nameAndNamespace(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	url, found, err := unstructured.NestedString(unstructuredRepo, "spec", "url")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field spec.url not found on HelmRepository:\n%s, error: %v", prettyPrintMap(unstructuredRepo), err)
	}
	return &v1alpha1.PackageRepository{
		Name:      name,
		Namespace: namespace,
		Url:       url,
	}, nil
}

//
// implements plug-in specific cache-related functionality
//

// onAddOrModifyRepo essentially tells the cache what to store for a given key
func onAddOrModifyRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	if isRepoReady(unstructuredRepo) {
		charts, err := indexOneRepo(unstructuredRepo)
		if err != nil {
			return nil, false, err
		}

		jsonBytes, err := json.Marshal(charts)
		if err != nil {
			return nil, false, err
		}

		return jsonBytes, true, nil
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func onGetRepo(key string, value interface{}) (interface{}, error) {
	bytes, ok := value.([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unexpected value found in cache for key [%s]: %v", key, value)
	}

	var charts []models.Chart
	err := json.Unmarshal(bytes, &charts)
	if err != nil {
		return nil, err
	}
	return charts, nil
}

func onDeleteRepo(key string, unstructuredRepo map[string]interface{}) (bool, error) {
	return true, nil
}
