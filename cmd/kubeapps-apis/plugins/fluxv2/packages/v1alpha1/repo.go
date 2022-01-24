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
	"strings"
	"time"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
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
	_, client, _, err := s.GetClients(ctx)
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
func (s *Server) listReposInNamespace(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	resourceIfc, err := s.getRepoResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if repos, err := resourceIfc.List(ctx, metav1.ListOptions{}); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmRepository", namespace+"/*", err)
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
		return nil, statuserror.FromK8sError("get", "HelmRepository", name.String(), err)
	} else if result == nil {
		return nil, status.Errorf(codes.NotFound, "unable to find HelmRepository [%s]", name)
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
		name, err := common.NamespacedName(repo.Object)
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
			resultKeys = append(resultKeys, s.repoCache.KeyForNamespacedName(*name))
		}
	}
	return resultKeys, nil
}

func (s *Server) getChartsForRepos(ctx context.Context, match []string) (map[string][]models.Chart, error) {
	// 1. with flux an available package may be from a repo in any namespace
	// 2. can't rely on cache as a real source of truth for key names
	//    because redis may evict cache entries due to memory pressure to make room for new ones
	repoList, err := s.listReposInNamespace(ctx, apiv1.NamespaceAll)
	if err != nil {
		return nil, err
	}

	repoNames, err := s.filterReadyReposByName(repoList, match)
	if err != nil {
		return nil, err
	}

	chartsUntyped, err := s.repoCache.GetForMultiple(repoNames)
	if err != nil {
		return nil, err
	}

	chartsTyped := make(map[string][]models.Chart)
	for key, value := range chartsUntyped {
		if value == nil {
			chartsTyped[key] = nil
		} else {
			typedValue, ok := value.(repoCacheEntryValue)
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

func (s *Server) clientOptionsForRepo(ctx context.Context, repo types.NamespacedName) (*common.ClientOptions, error) {
	unstructuredRepo, err := s.getRepoInCluster(ctx, repo)
	if err != nil {
		return nil, err
	}
	// notice a bit of inconsistency here, we are using s.clientGetter
	// (i.e. the context of the incoming request) to read the secret
	// as opposed to s.repoCache.clientGetter (which uses the context of
	//	User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis")
	// which is what is used when the repo is being processed/indexed.
	// I don't think it's necessarily a bad thing if the incoming user's RBAC
	// settings are more permissive than that of the default RBAC for
	// kubeapps-internal-kubeappsapis account. If we don't like that behavior,
	// I can easily switch to using common.NewBackgroundClientGetter here
	callSite := repoEventSink{
		clientGetter: s.clientGetter,
		chartCache:   s.chartCache,
	}
	return callSite.clientOptionsForRepo(ctx, unstructuredRepo.Object)
}

//
// implements plug-in specific cache-related functionality
//
type repoEventSink struct {
	clientGetter clientgetter.ClientGetterWithApiExtFunc
	chartCache   *cache.ChartCache // chartCache maybe nil only in unit tests
}

// this is what we store in the cache for each cached repo
// all struct fields are capitalized so they're exported by gob encoding
type repoCacheEntryValue struct {
	Checksum string
	Charts   []models.Chart
}

// onAddRepo essentially tells the cache whether to and what to store for a given key
func (s *repoEventSink) onAddRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	log.V(4).Info("+onAddRepo()")
	defer log.V(4).Info("-onAddRepo()")

	// TODO (gfichtenholt) use
	// runtime.DefaultUnstructuredConverter.FromUnstructured to convert to flux typed API
	// https://fluxcd.io/docs/components/source/api/#source.toolkit.fluxcd.io/v1beta1.HelmRepository

	// first, check the repo is ready
	if isRepoReady(unstructuredRepo) {
		// ref https://fluxcd.io/docs/components/source/helmrepositories/#status
		checksum, found, err := unstructured.NestedString(unstructuredRepo, "status", "artifact", "checksum")
		if err != nil || !found {
			return nil, false, status.Errorf(codes.Internal,
				"expected field status.artifact.checksum not found on HelmRepository\n[%s], error %v",
				common.PrettyPrintMap(unstructuredRepo), err)
		}
		return s.indexAndEncode(checksum, unstructuredRepo)
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", key)
		return nil, false, nil
	}
}

func (s *repoEventSink) indexAndEncode(checksum string, unstructuredRepo map[string]interface{}) ([]byte, bool, error) {
	charts, err := s.indexOneRepo(unstructuredRepo)
	if err != nil {
		return nil, false, err
	}

	cacheEntryValue := repoCacheEntryValue{
		Checksum: checksum,
		Charts:   charts,
	}

	// use gob encoding instead of json, it peforms much better
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(cacheEntryValue); err != nil {
		return nil, false, err
	}

	if s.chartCache != nil {
		if opts, err := s.clientOptionsForRepo(context.Background(), unstructuredRepo); err != nil {
			// ref: https://github.com/kubeapps/kubeapps/pull/3899#issuecomment-990446931
			// I don't want this func to fail onAdd/onModify() if we can't read
			// the corresponding secret due to something like default RBAC settings:
			// "secrets "podinfo-basic-auth-secret" is forbidden:
			// User "system:serviceaccount:kubeapps:kubeapps-internal-kubeappsapis" cannot get
			// resource "secrets" in API group "" in the namespace "default"
			// So we still finish the indexing of the repo but skip the charts
			log.Errorf("Failed to read secret for repo due to: %+v", err)
		} else if err = s.chartCache.SyncCharts(charts, opts); err != nil {
			return nil, false, err
		}
	}
	return buf.Bytes(), true, nil
}

// it is assumed the caller has already checked that this repo is ready
// At present, there is only one caller of indexOneRepo() and this check is already done by it
func (s *repoEventSink) indexOneRepo(unstructuredRepo map[string]interface{}) ([]models.Chart, error) {
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

	// In production, there should be no need to provide authz, userAgent or any of the TLS details,
	// as we are reading index.yaml file from local cluster, not some remote repo.
	// e.g. http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index.yaml
	// Flux does the hard work of pulling the index file from remote repo
	// into local cluster based on secretRef associated with HelmRepository, if applicable
	// This is only true of index.yaml, not the individual chart URLs within it

	// if a transient error occurs the item will be re-queued and retried after a back-off period
	byteArray, err := httpclient.Get(indexUrl, httpclient.New(), nil)
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
	charts, err := helm.ChartsFromIndex(byteArray, modelRepo, false)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	msg := fmt.Sprintf("-indexOneRepo: [%s], indexed [%d] packages in [%d] ms", repo.Name, len(charts), duration.Milliseconds())
	if len(charts) > 0 {
		log.Info(msg)
	} else {
		// this is kind of a red flag - an index with 0 charts, most likely contents of index.yaml is
		// messed up and didn't parse successfully but the helm library didn't raise an error
		log.Warning(msg)
	}
	return charts, nil
}

// onModifyRepo essentially tells the cache whether to and what to store for a given key
func (s *repoEventSink) onModifyRepo(key string, unstructuredRepo map[string]interface{}, oldValue interface{}) (interface{}, bool, error) {
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
				common.PrettyPrintMap(unstructuredRepo), err)
		}

		cacheEntryUntyped, err := s.onGetRepo(key, oldValue)
		if err != nil {
			return nil, false, err
		}

		cacheEntry, ok := cacheEntryUntyped.(repoCacheEntryValue)
		if !ok {
			return nil, false, status.Errorf(
				codes.Internal,
				"unexpected value found in cache for key [%s]: %v",
				key, cacheEntryUntyped)
		}

		if cacheEntry.Checksum != newChecksum {
			return s.indexAndEncode(newChecksum, unstructuredRepo)
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

func (s *repoEventSink) onGetRepo(key string, value interface{}) (interface{}, error) {
	b, ok := value.([]byte)
	if !ok {
		return nil, status.Errorf(codes.Internal, "unexpected value found in cache for key [%s]: %v", key, value)
	}

	dec := gob.NewDecoder(bytes.NewReader(b))
	var entryValue repoCacheEntryValue
	if err := dec.Decode(&entryValue); err != nil {
		return nil, err
	}
	return entryValue, nil
}

func (s *repoEventSink) onDeleteRepo(key string) (bool, error) {
	if s.chartCache != nil {
		if name, err := s.fromKey(key); err != nil {
			return false, err
		} else if err := s.chartCache.DeleteChartsForRepo(name); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (s *repoEventSink) onResync() error {
	if s.chartCache != nil {
		return s.chartCache.OnResync()
	} else {
		return nil
	}
}

// TODO (gfichtenholt) low priority: don't really like the fact that these 4 lines of code
// basically repeat same logic as NamespacedResourceWatcherCache.fromKey() but can't
// quite come up with with a more elegant alternative right now
func (c *repoEventSink) fromKey(key string) (*types.NamespacedName, error) {
	parts := strings.Split(key, cache.KeySegmentsSeparator)
	if len(parts) != 3 || parts[0] != fluxHelmRepositories || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return &types.NamespacedName{Namespace: parts[1], Name: parts[2]}, nil
}

// unstructuredRepo is passed as map[string]interface{}
// this is only until https://github.com/kubeapps/kubeapps/issues/3496
// "Investigate and propose package repositories API with similar core interface to packages API"
// gets implemented. After that, the auth should be part of packageRepositoryFromUnstructured()
// The reason I do this here is to set up auth that may be needed to fetch chart tarballs by
// ChartCache
func (s *repoEventSink) clientOptionsForRepo(ctx context.Context, unstructuredRepo map[string]interface{}) (*common.ClientOptions, error) {
	secretName, found, err := unstructured.NestedString(unstructuredRepo, "spec", "secretRef", "name")
	if !found || err != nil || secretName == "" {
		return nil, nil
	}
	if s == nil || s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "unexpected state in clientGetterHolder instance")
	}
	typedClient, _, _, err := s.clientGetter(ctx)
	if err != nil {
		return nil, err
	}
	repoName, err := common.NamespacedName(unstructuredRepo)
	if err != nil {
		return nil, err
	}
	secret, err := typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, status.Errorf(codes.Unauthenticated, "unable to get secret due to %v", err)
		} else {
			return nil, status.Errorf(codes.Internal, "unable to get secret due to %v", err)
		}
	}
	return common.ClientOptionsFromSecret(*secret)
}

//
// repo-related utilities
//

func isRepoReady(unstructuredRepo map[string]interface{}) bool {
	// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
	// Confirm the state we are observing is for the current generation
	if !common.CheckGeneration(unstructuredRepo) {
		return false
	}

	completed, success, _ := isHelmRepositoryReady(unstructuredRepo)
	return completed && success
}

func packageRepositoryFromUnstructured(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, err := common.NamespacedName(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	url, found, err := unstructured.NestedString(unstructuredRepo, "spec", "url")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field spec.url not found on HelmRepository:\n%s, error: %v", common.PrettyPrintMap(unstructuredRepo), err)
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
	if !common.CheckGeneration(unstructuredObj) {
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
