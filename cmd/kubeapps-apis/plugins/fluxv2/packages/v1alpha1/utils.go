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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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
		return "", status.Errorf(codes.InvalidArgument, "Incorrect request.AvailablePackageRef.Identifier, currently just 'foo/bar' patters are supported: %s", chartID)
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

// returns 3 things:
// - complete whether the operation was completed
// - success (only applicable when complete == true) whether the operation was successful or failed
// - reason, if present
func checkStatusReady(unstructuredObj map[string]interface{}) (complete bool, success bool, reason string) {
	conditions, found, err := unstructured.NestedSlice(unstructuredObj, "status", "conditions")
	if err != nil || !found {
		return false, false, ""
	}

	for _, conditionUnstructured := range conditions {
		if conditionAsMap, ok := conditionUnstructured.(map[string]interface{}); ok {
			if typeString, ok := conditionAsMap["type"]; ok && typeString == "Ready" {
				if reasonString, ok := conditionAsMap["reason"]; ok {
					reason = fmt.Sprintf("%v", reasonString)
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

func nameAndNamespace(unstructuredObj map[string]interface{}) (name, namespace string, err error) {
	name, found, err := unstructured.NestedString(unstructuredObj, "metadata", "name")
	if err != nil || !found {
		return "", "",
			status.Errorf(codes.Internal, "required field metadata.name not found on resource: %v:\n%s",
				err,
				prettyPrintMap(unstructuredObj))
	}

	namespace, found, err = unstructured.NestedString(unstructuredObj, "metadata", "namespace")
	if err != nil || !found {
		return "", "",
			status.Errorf(codes.Internal, "required field metadata.namespace not found on resource: %v:\n%s",
				err,
				prettyPrintMap(unstructuredObj))
	}
	return name, namespace, nil
}
