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
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
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

// the goal of this fn is to answer whether or not to stop waiting for chart reconciliation
// which is different from answering whether the chart was pulled successfully
// TODO (gfichtenholt): As above, hopefully this fn isn't required if we can only list charts that we know are ready.
func isChartPullComplete(unstructuredChart *unstructured.Unstructured) (bool, error) {
	// see docs at https://fluxcd.io/docs/components/source/helmcharts/
	// Confirm the state we are observing is for the current generation
	observedGeneration, found, err := unstructured.NestedInt64(unstructuredChart.Object, "status", "observedGeneration")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}
	generation, found, err := unstructured.NestedInt64(unstructuredChart.Object, "metadata", "generation")
	if err != nil {
		return false, err
	} else if !found {
		return false, nil
	}
	if generation != observedGeneration {
		return false, nil
	}

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
				if statusString, ok := conditionAsMap["status"]; ok {
					if statusString == "True" {
						if reasonString, ok := conditionAsMap["reason"]; !ok || reasonString != "ChartPullSucceeded" {
							// should not happen
							log.Infof("unexpected status of HelmChart: %v", *unstructuredChart)
						}
						return true, nil
					} else if statusString == "False" {
						var msg string
						if msg, ok = conditionAsMap["message"].(string); !ok {
							msg = fmt.Sprintf("No message available in condition: %v", conditionAsMap)
						}
						// chart pull is done and it's a failure
						return true, status.Errorf(codes.Internal, msg)
					}
				}
			}
		}
	}
	return false, nil
}

// TODO (gfichtenholt):
// see https://github.com/kubeapps/kubeapps/pull/2915 for context
// In the future you might instead want to consider something like
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
