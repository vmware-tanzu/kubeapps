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
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"regexp"
	"time"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
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

func (s *Server) getRepoResourceInterface(ctx context.Context, namespace string) (dynamic.ResourceInterface, error) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return nil, err
	}

	repositoriesResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	return client.Resource(repositoriesResource).Namespace(namespace), nil
}

// namespace maybe apiv1.NamespaceAll, in which case repositories from all namespaces are returned
func (s *Server) listReposInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	resourceIfc, err := s.getRepoResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if repos, err := resourceIfc.List(ctx, metav1.ListOptions{}); err != nil {
		if errors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "%q", err)
		} else if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, status.Errorf(codes.Unauthenticated, "unable to list repositories in namespace [%s] due to %v", namespace, err)
		} else {
			return nil, status.Errorf(codes.Internal, "unable to list repositories in namespace [%s] due to %v", namespace, err)
		}
	} else {
		return repos, nil
	}
}

func (s *Server) getRepoInCluster(ctx context.Context, name types.NamespacedName) (*unstructured.Unstructured, error) {
	resourceIfc, err := s.getRepoResourceInterface(ctx, name.Namespace)
	if err != nil {
		return nil, err
	}

	result, err := resourceIfc.Get(ctx, name.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "%q", err)
		} else if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, status.Errorf(codes.Unauthenticated, "unable to get repository [%s] due to %v", name, err)
		} else {
			return nil, status.Errorf(codes.Internal, "unable to get repository [%s] due to %v", name, err)
		}
	} else if result == nil {
		return nil, status.Errorf(codes.NotFound, "unable to find repository [%s]", name)
	}
	return result, nil
}

// regexp expressions are used for matching actual names against expected patters
func (s *Server) filterReadyReposByName(repoList *unstructured.UnstructuredList, match []string) ([]string, error) {
	if s.repoCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	}

	resultKeys := make([]string, 0)
	for _, repo := range repoList.Items {
		// first check if repo is in ready state
		if !isRepoReady(repo.Object) {
			// just skip it
			continue
		}
		name, err := namespacedName(repo.Object)
		if err != nil {
			// just skip it
			continue
		}
		// see if name matches the filter
		matched := false
		if len(match) > 0 {
			for _, m := range match {
				if matched, err = regexp.MatchString(m, name.Name); matched && err == nil {
					break
				}
			}
		} else {
			matched = true
		}
		if matched {
			resultKeys = append(resultKeys, s.repoCache.keyForNamespacedName(*name))
		}
	}
	return resultKeys, nil
}

func (s *Server) getChartsForRepos(ctx context.Context, match []string) (map[string][]models.Chart, error) {
	// 1. with flux an available package may be from a repo in any namespace
	// 2. can't rely on cache as a real source of truth for key names
	//    because redis may evict cache entries due to memory pressure to make room for new ones
	repoList, err := s.listReposInCluster(ctx, apiv1.NamespaceAll)
	if err != nil {
		return nil, err
	}

	repoNames, err := s.filterReadyReposByName(repoList, match)
	if err != nil {
		return nil, err
	}

	chartsUntyped, err := s.repoCache.getForMultiple(repoNames, repoList)
	if err != nil {
		return nil, err
	}

	chartsTyped := make(map[string][]models.Chart)
	for key, value := range chartsUntyped {
		if value == nil {
			chartsTyped[key] = nil
		} else {
			typedValue, ok := value.(repoCacheEntry)
			if !ok {
				return nil, status.Errorf(
					codes.Internal,
					"unexpected value fetched from cache: type: [%s], value: [%v]",
					reflect.TypeOf(value), value)
			}
			chartsTyped[key] = typedValue.Charts
		}
	}
	return chartsTyped, nil
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

	completed, success, _ := isHelmRepositoryReady(unstructuredRepo)
	return completed && success
}

// it is assumed the caller has already checked that this repo is ready
// At present, there is only one caller of indexOneRepo() and this check is already done by it
func indexOneRepo(unstructuredRepo map[string]interface{}) (charts []models.Chart, err error) {
	startTime := time.Now()

	repo, err := packageRepositoryFromUnstructured(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
	indexUrl, found, err := unstructured.NestedString(unstructuredRepo, "status", "url")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal,
			"expected field status.url not found on HelmRepository\n[%s], error %v",
			repo.Name, err)
	}

	log.Infof("+indexOneRepo: [%s], index URL: [%s]", repo.Name, indexUrl)

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
	charts, err = helm.ChartsFromIndex(bytes, modelRepo, false)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	msg := fmt.Sprintf("-indexOneRepo: [%s], indexed [%d] packages in [%d] ms", repo.Name, len(charts), duration.Milliseconds())
	if len(charts) > 0 {
		log.Info(msg)
	} else {
		// this is kind of a red flag - an index with 0 charts, most likely contents of index.yaml is
		// messed up and didn't parse but the helm library didn't raise an error
		log.Warning(msg)
	}
	return charts, nil
}

func packageRepositoryFromUnstructured(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, err := namespacedName(unstructuredRepo)
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
		Name:      name.Name,
		Namespace: name.Namespace,
		Url:       url,
	}, nil
}

// returns 3 things:
// - complete whether the operation was completed
// - success (only applicable when complete == true) whether the operation was successful or failed
// - reason, if present
// docs:
// 1. https://fluxcd.io/docs/components/source/helmrepositories/#status-examples
func isHelmRepositoryReady(unstructuredObj map[string]interface{}) (complete bool, success bool, reason string) {
	if !checkGeneration(unstructuredObj) {
		return false, false, ""
	}

	conditions, found, err := unstructured.NestedSlice(unstructuredObj, "status", "conditions")
	if err != nil || !found {
		return false, false, ""
	}

	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				// this could be something like
				// "reason": "ChartPullFailed"
				// i.e. not super-useful
				if reasonString, ok := conditionAsMap["reason"]; ok {
					reason = fmt.Sprintf("%v", reasonString)
				}
				// whereas this could be something like:
				// "message": 'invalid chart URL format'
				// i.e. a little more useful, so we'll just return them both
				if messageString, ok := conditionAsMap["message"]; ok {
					reason += fmt.Sprintf(": %v", messageString)
				}
				if statusString, ok := conditionAsMap["status"]; ok {
					if statusString == "True" {
						return true, true, reason
					} else if statusString == "False" {
						return true, false, reason
					}
					// statusString == "Unknown" falls in here
				}
				break
			}
		}
	}
	return false, false, reason
}

//
// implements plug-in specific cache-related functionality
//

// this is what we store in the cache for each cached repo
// all struct fields are capitalized so they're exported by gob encoding
type repoCacheEntry struct {
	Checksum string
	Charts   []models.Chart
}

// onAddRepo essentially tells the cache whether to and what to store for a given key
func onAddRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	// first, check the repo is ready
	if isRepoReady(unstructuredRepo) {
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		checksum, found, err := unstructured.NestedString(unstructuredRepo, "status", "artifact", "checksum")
		if err != nil || !found {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact.checksum not found on HelmRepository\n[%s], error %v",
				prettyPrintMap(unstructuredRepo), err)
		}
		return indexAndEncode(checksum, unstructuredRepo)
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

// onModifyRepo essentially tells the cache whether to and what to store for a given key
func onModifyRepo(key string, unstructuredRepo map[string]interface{}, oldValue interface{}) (interface{}, bool, error) {
	// first check the repo is ready
	if isRepoReady(unstructuredRepo) {
		// We should to compare checksums on what's stored in the cache
		// vs the modified object to see if the contents has really changed before embarking on
		// expensive operation indexOneRepo() below.
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		newChecksum, found, err := unstructured.NestedString(unstructuredRepo, "status", "artifact", "checksum")
		if err != nil || !found {
			return nil, false, status.Errorf(
				codes.Internal,
				"expected field status.artifact.checksum not found on HelmRepository\n[%s], error %v",
				prettyPrintMap(unstructuredRepo), err)
		}

		cacheEntryUntyped, err := onGetRepo(key, oldValue)
		if err != nil {
			return nil, false, err
		}

		cacheEntry, ok := cacheEntryUntyped.(repoCacheEntry)
		if !ok {
			return nil, false, status.Errorf(
				codes.Internal,
				"unexpected value found in cache for key [%s]: %v",
				key, cacheEntryUntyped)
		}

		if cacheEntry.Checksum != newChecksum {
			return indexAndEncode(newChecksum, unstructuredRepo)
		} else {
			// skip because the content did not change
			return nil, false, nil
		}
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.V(4).Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func onGetRepo(key string, value interface{}) (interface{}, error) {
	b, ok := value.([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unexpected value found in cache for key [%s]: %v", key, value)
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	var entry repoCacheEntry
	if err := dec.Decode(&entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func onDeleteRepo(key string) (bool, error) {
	return true, nil
}

func indexAndEncode(checksum string, unstructuredRepo map[string]interface{}) ([]byte, bool, error) {
	charts, err := indexOneRepo(unstructuredRepo)
	if err != nil {
		return nil, false, err
	}

	cacheEntry := repoCacheEntry{
		Checksum: checksum,
		Charts:   charts,
	}

	// use gob encoding instead of json, it peforms much better
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(cacheEntry); err != nil {
		return nil, false, err
	}
	return buf.Bytes(), true, nil
}
