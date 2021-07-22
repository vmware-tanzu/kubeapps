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

// repo-related utilities

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	chart "github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/helm"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	log "k8s.io/klog/v2"
)

func isRepoReady(obj map[string]interface{}) (bool, error) {
	// see docs at https://fluxcd.io/docs/components/source/helmrepositories/
	// Confirm the state we are observing is for the current generation
	observedGeneration, found, err := unstructured.NestedInt64(obj, "status", "observedGeneration")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}
	generation, found, err := unstructured.NestedInt64(obj, "metadata", "generation")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}
	if generation != observedGeneration {
		return false, nil
	}

	conditions, found, err := unstructured.NestedSlice(obj, "status", "conditions")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}

	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				if statusString, ok := conditionAsMap["status"]; ok {
					if statusString == "True" {
						// note that the current doc on https://fluxcd.io/docs/components/source/helmrepositories/
						// incorrectly states the example status reason as "IndexationSucceeded".
						// The actual string is "IndexationSucceed"
						if reasonString, ok := conditionAsMap["reason"]; !ok || reasonString != "IndexationSucceed" {
							// should not happen
							log.Infof("Unexpected status of HelmRepository: %v", obj)
						}
						return true, nil
					} else if statusString == "False" {
						var msg string
						if msg, ok = conditionAsMap["message"].(string); !ok {
							msg = fmt.Sprintf("No message available in condition: %v", conditionAsMap)
						}
						return false, status.Errorf(codes.Internal, msg)
					}
				}
			}
		}
	}
	return false, nil
}

func indexOneRepo(unstructuredRepo map[string]interface{}) ([]chart.Chart, error) {
	startTime := time.Now()

	repo, err := newPackageRepository(unstructuredRepo)
	if err != nil {
		return nil, err
	}

	// TODO: (gfichtenholt) the caller already checks this before invoking, see if I can remove this
	ready, err := isRepoReady(unstructuredRepo)
	if err != nil || !ready {
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

	modelRepo := &chart.Repo{
		Namespace: repo.Namespace,
		Name:      repo.Name,
		URL:       repo.Url,
		Type:      "helm",
	}

	// this is potentially a very expensive operation for large repos like 'bitnami'
	charts, err := helm.ChartsFromIndex(bytes, modelRepo, true)
	if err != nil {
		return nil, err
	}

	duration := time.Since(startTime)
	log.Infof("indexOneRepo: indexed [%d] packages in repository [%s] in [%d] ms", len(charts), repo.Name, duration.Milliseconds())
	return charts, nil
}

func newPackageRepository(unstructuredRepo map[string]interface{}) (*v1alpha1.PackageRepository, error) {
	name, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"required field metadata.name not found on HelmRepository:\n%s, error: %v", prettyPrintMap(unstructuredRepo), err)
	}
	namespace, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(
			codes.Internal,
			"field metadata.namespace not found on HelmRepository:\n%s, error: %v", prettyPrintMap(unstructuredRepo), err)
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

// implements plug-in specific cache-related functionality
// onAddOrModifyRepo essentially tells the cache what to store for a given key
func onAddOrModifyRepo(key string, unstructuredRepo map[string]interface{}) (interface{}, bool, error) {
	ready, err := isRepoReady(unstructuredRepo)
	if err != nil {
		return nil, false, err
	}

	if ready {
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

	var charts []chart.Chart
	err := json.Unmarshal(bytes, &charts)
	if err != nil {
		return nil, err
	}
	return charts, nil
}

func onDeleteRepo(key string, unstructuredRepo map[string]interface{}) (bool, error) {
	return true, nil
}
