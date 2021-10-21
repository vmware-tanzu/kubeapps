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
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
)

// see https://blog.golang.org/context
// this is used exclusively for unit tests to signal conditions between production
// and unit test code. The key type is unexported to prevent collisions with context
// keys defined in other packages.
type contextKey int

// waitGroupKey is the context key for the waitGroup.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const waitGroupKey contextKey = 0

func fromContext(ctx context.Context) (*sync.WaitGroup, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the sync.WaitGroup type assertion returns ok=false for nil.
	wg, ok := ctx.Value(waitGroupKey).(*sync.WaitGroup)
	return wg, ok
}

func newContext(ctx context.Context, wg *sync.WaitGroup) context.Context {
	return context.WithValue(ctx, waitGroupKey, wg)
}

//
// miscellaneous utility funcs
//
func prettyPrintObject(o runtime.Object) string {
	prettyBytes, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", o)
	}
	return string(prettyBytes)
}

func prettyPrintMap(m map[string]interface{}) string {
	prettyBytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", m)
	}
	return string(prettyBytes)
}

// pageOffsetFromPageToken converts a page token to an integer offset
// representing the page of results.
// TODO(gfichtenholt): it'd be better if we ensure that the page_token
// contains an offset to the item, not the page so we can
// aggregate paginated results. Same as helm hlug-in.
// Update this when helm plug-in does so
func pageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 1, nil
	}
	offset, err := strconv.ParseUint(pageToken, 10, 0)
	if err != nil {
		return 0, err
	}
	return int(offset), nil
}

// getUnescapedChartID takes a chart id with URI-encoded characters and decode them. Ex: 'foo%2Fbar' becomes 'foo/bar'
// also checks that the chart ID is in the expected format, namely "repoName/chartName"
func getUnescapedChartID(chartID string) (string, error) {
	unescapedChartID, err := url.QueryUnescape(chartID)
	if err != nil {
		return "", status.Errorf(codes.Internal, "Unable to decode chart ID chart: %v", chartID)
	}
	// TODO(agamez): support ID with multiple slashes, eg: aaa/bbb/ccc
	chartIDParts := strings.Split(unescapedChartID, "/")
	if len(chartIDParts) != 2 {
		return "", status.Errorf(codes.InvalidArgument, "Incorrect package ref dentifier, currently just 'foo/bar' patterns are supported: %s", chartID)
	}
	return unescapedChartID, nil
}

// Confirm the state we are observing is for the current generation
// returns true if object's status.observedGeneration == metadata.generation
// false otherwise
func checkGeneration(unstructuredObj map[string]interface{}) bool {
	observedGeneration, found, err := unstructured.NestedInt64(unstructuredObj, "status", "observedGeneration")
	if err != nil || !found {
		return false
	}
	generation, found, err := unstructured.NestedInt64(unstructuredObj, "metadata", "generation")
	if err != nil || !found {
		return false
	}
	return generation == observedGeneration
}

func namespacedName(unstructuredObj map[string]interface{}) (*types.NamespacedName, error) {
	name, found, err := unstructured.NestedString(unstructuredObj, "metadata", "name")
	if err != nil || !found {
		return nil,
			status.Errorf(codes.Internal, "required field 'metadata.name' not found on resource: %v:\n%s",
				err,
				prettyPrintMap(unstructuredObj))
	}

	namespace, found, err := unstructured.NestedString(unstructuredObj, "metadata", "namespace")
	if err != nil || !found {
		return nil,
			status.Errorf(codes.Internal, "required field 'metadata.namespace' not found on resource: %v:\n%s",
				err,
				prettyPrintMap(unstructuredObj))
	}
	return &types.NamespacedName{Name: name, Namespace: namespace}, nil
}

func newHelmActionConfigGetter(configGetter server.KubernetesConfigGetter, cluster string) helmActionConfigGetter {
	return func(ctx context.Context, namespace string) (*action.Configuration, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		config, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
		}

		restClientGetter := agent.NewConfigFlagsFromCluster(namespace, config)
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to create kubernetes client due to: %v", err)
		}
		// TODO(mnelson): Update to allow different helm storage options.
		storage := agent.StorageForSecrets(namespace, clientSet)
		return &action.Configuration{
			RESTClientGetter: restClientGetter,
			KubeClient:       kube.New(restClientGetter),
			Releases:         storage,
			Log:              log.Infof,
		}, nil
	}
}

func newClientGetter(configGetter server.KubernetesConfigGetter, cluster string) clientGetter {
	return func(ctx context.Context) (dynamic.Interface, apiext.Interface, error) {
		if configGetter == nil {
			return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		if config, err := configGetter(ctx, cluster); err != nil {
			if status.Code(err) == codes.Unauthenticated {
				// want to make sure we return same status in this case
				return nil, nil, status.Errorf(codes.Unauthenticated, "unable to get config due to: %v", err)
			} else {
				return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}
		} else {
			return clientGetterHelper(config)
		}
	}
}

// https://github.com/kubeapps/kubeapps/issues/3560
// flux plug-in runs out-of-request interactions with the Kubernetes API server.
// Although we've already ensured that if the flux plugin is selected, that the service account
// will be granted additional read privileges, we also need to ensure that the plugin can get a
// config based on the service account rather than the request context
func newBackgroundClientGetter() clientGetter {
	return func(ctx context.Context) (dynamic.Interface, apiext.Interface, error) {
		if config, err := rest.InClusterConfig(); err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get in cluster config due to: %v", err)
		} else {
			return clientGetterHelper(config)
		}
	}
}

func clientGetterHelper(config *rest.Config) (dynamic.Interface, apiext.Interface, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client due to: %v", err)
	}
	apiExtensions, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get api extensions client due to: %v", err)
	}
	return dynamicClient, apiExtensions, nil
}
