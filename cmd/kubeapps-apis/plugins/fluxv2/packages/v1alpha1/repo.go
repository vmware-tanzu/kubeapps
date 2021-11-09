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
	if s.cache == nil {
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
			resultKeys = append(resultKeys, s.cache.keyForNamespacedName(*name))
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

	chartsTyped := make(map[string][]models.Chart)

	// at any given moment, the redis cache may only have a subset of the entire set of existing keys.
	// Some key may have been evicted due to memory pressure and LRU eviction policy.
	// ref: https://redis.io/topics/lru-cache
	// so, first, let's fetch the entries that are still cached before redis evicts those
	chartsUntyped, err := s.cache.fetchForMultiple(repoNames)
	if err != nil {
		return nil, err
	}

	// now, re-compute and fetch the ones that are are left over from the previous operation,
	// TODO: (gfichtenholt) I think this for loop can be done for all entries processed in
	//   parallel rather than one-at-a-time. The computation part certainly can be parallelized,
	//   and cache PUTs can also be done in parallel, potentially forcing redis to evict recently
	//   computed entires to make room for new ones along the way
	// TODO: (gfichtenholt) a bit of an inconcistency here. Some keys/values may have been
	//   added to the cache by a background go-routine running in the context of
	//   "kubeapps-internal-kubeappsapis" service account, whereas keys/values added below are
	//   going to be fetched from k8s on behalf of the caller which is a different service account
	//   with different RBAC settings
	for key, value := range chartsUntyped {
		if value == nil {
			// this cache miss may be due to one of these reasons:
			// 1) key truly does not exist in k8s (there is no repo with the given name in the "Ready" state)
			// 2) key exists and the "Ready" repo currently being indexed but has not yet completed
			// 3) key exists in k8s but the corresponding cache entry has been evicted by redis due to
			//    LRU maxmemory policies or entry TTL expiry (doesn't apply currently, cuz we use TTL=0
			//    for all entries)
			// In the 3rd case we want to re-compute the key and add it to the cache, which may potentially
			// cause other entries to be evicted in order to make room for the ones being added

			// TODO (gfichtenholt) handle (2) - there is a (small) time window during which two
			// threads will be doing the same work. No inconsistent results will occur, but still.
			if name, err := s.cache.fromKey(key); err != nil {
				return nil, err
			} else {
				for _, repo := range repoList.Items {
					if repoName, err := namespacedName(repo.Object); err != nil {
						return nil, err
					} else if *repoName == *name {
						if err = s.cache.onAddOrModify(true, repo.Object); err == nil {
							if value, err = s.cache.fetchForOne(key); err != nil {
								return nil, err
							}
						}
						break
					}
				}
			}
		}
		if value != nil {
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

// onAddRepo essentially tells the cache what to store for a given key
func onAddRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	// first check the repo is ready
	if isRepoReady(unstructuredRepo) {
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		checksum, found, err := unstructured.NestedString(unstructuredRepo, "status", "artifact", "checksum")
		if err != nil || !found {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact.checksum not found on HelmRepository\n[%s], error %v",
				prettyPrintMap(unstructuredRepo), err)
		}

		charts, err := indexOneRepo(unstructuredRepo)
		if err != nil {
			return nil, false, err
		}

		cacheEntry := repoCacheEntry{
			Checksum: checksum,
			Charts:   charts,
		}

		// launch go-routine that will download tarball and extract relevant details for each chart
		// and cache them for fast lookup
		// TODO (gfichtenholt) go cacheLatestChartDetails(charts)

		// use gob encoding instead of json, it peforms much better
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err = enc.Encode(cacheEntry); err != nil {
			return nil, false, err
		}
		return buf.Bytes(), true, nil
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

// onModifyRepo essentially tells the cache what to store for a given key
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
			newCharts, err := indexOneRepo(unstructuredRepo)
			if err != nil {
				return nil, false, err
			}

			cacheEntry = repoCacheEntry{
				Checksum: newChecksum,
				Charts:   newCharts,
			}

			// launch go-routine that will download tarball and extract relevant details for each chart
			// and cache them for fast lookup
			// TODO (gfichtenholt) go cacheLatestChartDetails(charts)

			// use gob encoding instead of json, it peforms much better
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			if err = enc.Encode(cacheEntry); err != nil {
				return nil, false, err
			}
			return buf.Bytes(), true, nil
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

func onDeleteRepo(key string, unstructuredRepo map[string]interface{}) (bool, error) {
	return true, nil
}
