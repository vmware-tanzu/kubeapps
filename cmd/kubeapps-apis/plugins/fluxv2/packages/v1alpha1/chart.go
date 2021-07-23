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
	"strings"

	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/helm/pkg/proto/hapi/chart"
	log "k8s.io/klog/v2"
)

// chart-related utilities

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
// passing a results channel (of string urls) to getChartTarball, so it returns
// immediately and you wait on the results channel at the call-site, which would mean
// you could call it for 20 different charts and just wait for the results to come in
// whatever order they happen to take, rather than serially.
func waitUntilChartPullComplete(watcher watch.Interface) (string, error) {
	ch := watcher.ResultChan()
	for {
		event := <-ch
		if event.Type == watch.Modified {
			unstructuredChart, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				return "", status.Errorf(codes.Internal, "could not cast to unstructured.Unstructured")
			}

			done, err := isChartPullComplete(unstructuredChart)
			if err != nil {
				return "", err
			} else if done {
				url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
				if err != nil || !found {
					return "", status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
				}
				return url, nil
			}
		} else {
			// TODO handle other kinds of events
			return "", status.Errorf(codes.Internal, "got unexpected event: %v", event)
		}
	}
}

// isValidChart returns true if the chart model passed defines a value
// for each required field described at the Helm website:
// https://helm.sh/docs/topics/charts/#the-chartyaml-file
// together with required fields for our model.
func isValidChart(chart *models.Chart) (bool, error) {
	if chart.Name == "" {
		return false, status.Errorf(codes.Internal, "required field .Name not found on helm chart: %v", chart)
	}
	if chart.ID == "" {
		return false, status.Errorf(codes.Internal, "required field .ID not found on helm chart: %v", chart)
	}
	if chart.Repo == nil {
		return false, status.Errorf(codes.Internal, "required field .Repo not found on helm chart: %v", chart)
	}
	if chart.ChartVersions == nil || len(chart.ChartVersions) == 0 {
		return false, status.Errorf(codes.Internal, "required field .chart.ChartVersions[0] not found on helm chart: %v", chart)
	} else {
		for _, chartVersion := range chart.ChartVersions {
			if chartVersion.Version == "" {
				return false, status.Errorf(codes.Internal, "required field .ChartVersions[i].Version not found on helm chart: %v", chart)
			}
		}
	}
	if chart.Maintainers != nil || len(chart.ChartVersions) != 0 {
		for _, maintainer := range chart.Maintainers {
			if maintainer.Name == "" {
				return false, status.Errorf(codes.Internal, "required field .Maintainers[i].Name not found on helm chart: %v", chart)
			}
		}
	}
	return true, nil
}

func findUrlForChartInList(chartList *unstructured.UnstructuredList, repoName, chartName, version string) (string, error) {
	for _, unstructuredChart := range chartList.Items {
		thisChartName, found, err := unstructured.NestedString(unstructuredChart.Object, "spec", "chart")
		thisRepoName, found2, err2 := unstructured.NestedString(unstructuredChart.Object, "spec", "sourceRef", "name")

		if err == nil && err2 == nil && found && found2 && repoName == thisRepoName && chartName == thisChartName {
			done, err := isChartPullComplete(&unstructuredChart)
			if err != nil {
				return "", err
			} else if done {
				url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
				if err != nil || !found {
					return "", status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
				}
				if version != "" {
					// refer to https://github.com/fluxcd/source-controller/blob/main/api/v1beta1/helmchart_types.go &
					// https://github.com/fluxcd/source-controller/blob/40a47670aadebc0f4e3a623be47725106bac2d55/api/v1beta1/artifact_types.go#L27
					chartVersion, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "artifact", "revision")
					if err != nil || !found {
						return "", status.Errorf(codes.Internal, "expected field status.artifact.revision not found on HelmChart: %v:\n%v", err, unstructuredChart)
					} else if chartVersion != version {
						continue
					}
				}
				log.Infof("Found existing HelmChart for: [%s/%s]", repoName, chartName)
				return url, nil
			}
			// TODO (gfichtenholt) waitUntilChartPullComplete?
		}
	}
	return "", nil
}

// availablePackageSummaryFromChart builds an AvailablePackageSummary from a Chart
func availablePackageSummaryFromChart(chart *models.Chart) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	isValid, err := isValidChart(chart)
	if !isValid || err != nil {
		return nil, status.Errorf(codes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.DisplayName = chart.Name
	pkg.IconUrl = chart.Icon
	pkg.ShortDescription = chart.Description

	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     GetPluginDetail(),
	}
	pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}

	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.LatestPkgVersion = chart.ChartVersions[0].Version
	}

	return pkg, nil
}

func passesFilter(chart models.Chart, filters *corev1.FilterOptions) bool {
	if filters == nil {
		return true
	}
	ok := true
	if categories := filters.GetCategories(); len(categories) > 0 {
		ok = false
		for _, cat := range categories {
			if cat == chart.Category {
				ok = true
				break
			}
		}
	}
	if ok {
		if appVersion := filters.GetAppVersion(); len(appVersion) > 0 {
			ok = appVersion == chart.ChartVersions[0].AppVersion
		}
	}
	if ok {
		if pkgVersion := filters.GetPkgVersion(); len(pkgVersion) > 0 {
			ok = pkgVersion == chart.ChartVersions[0].Version
		}
	}
	if ok {
		if query := filters.GetQuery(); len(query) > 0 {
			if strings.Contains(chart.Name, query) {
				return true
			}
			if strings.Contains(chart.Description, query) {
				return true
			}
			for _, keyword := range chart.Keywords {
				if strings.Contains(keyword, query) {
					return true
				}
			}
			for _, source := range chart.Sources {
				if strings.Contains(source, query) {
					return true
				}
			}
			for _, maintainer := range chart.Maintainers {
				if strings.Contains(maintainer.Name, query) {
					return true
				}
			}
			// could not find a match for the query text
			ok = false
		}
	}
	return ok
}

func filterAndPaginateCharts(filters *corev1.FilterOptions, pageSize, pageOffset int, cachedCharts map[string]interface{}) ([]*corev1.AvailablePackageSummary, error) {
	// this loop is here for 3 reasons:
	// 1) to convert from []interface{} which is what the generic cache implementation
	// returns for cache hits to a typed array object.
	// 2) perform any filtering of the results as needed, pending redis support for
	// querying values stored in cache (see discussion in https://github.com/kubeapps/kubeapps/issues/3032)
	// 3) if pagination was requested, only return up to one page size of results
	summaries := make([]*corev1.AvailablePackageSummary, 0)
	i := 0
	startAt := -1
	if pageSize > 0 {
		startAt = int(pageSize) * pageOffset
	}
	for _, packages := range cachedCharts {
		if packages == nil {
			continue
		}
		typedCharts, ok := packages.([]models.Chart)
		if !ok {
			return nil, status.Errorf(
				codes.Internal,
				"Unexpected value fetched from cache: %v", packages)
		} else {
			for _, chart := range typedCharts {
				if passesFilter(chart, filters) {
					i++
					if startAt < i {
						pkg, err := availablePackageSummaryFromChart(&chart)
						if err != nil {
							return nil, status.Errorf(
								codes.Internal,
								"Unable to parse chart to an AvailablePackageSummary: %v",
								err)
						}
						summaries = append(summaries, pkg)
						if pageSize > 0 && len(summaries) == int(pageSize) {
							return summaries, nil
						}
					}
				}
			}
		}
	}
	return summaries, nil
}

func newFluxHelmChart(chartName, repoName, version string) unstructured.Unstructured {
	unstructuredChart := unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmChart,
			"metadata": map[string]interface{}{
				"generateName": fmt.Sprintf("%s-", chartName),
			},
			"spec": map[string]interface{}{
				"chart": chartName,
				"sourceRef": map[string]interface{}{
					"name": repoName,
					"kind": fluxHelmRepository,
				},
				"interval": "10m",
			},
		},
	}
	if version != "" {
		unstructured.SetNestedField(unstructuredChart.Object, version, "spec", "version")
	}
	return unstructuredChart
}

func availablePackageDetailFromTarball(detail map[string]string) (*corev1.AvailablePackageDetail, error) {
	chartYaml := detail[models.ChartYamlKey]
	// TODO (gfichtenholt): if there is no chart yaml (is that even possible?), fall back to chart info from
	// repo index.yaml
	var chartMetadata chart.Metadata
	err := yaml.Unmarshal([]byte(chartYaml), &chartMetadata)
	if err != nil {
		return nil, err
	}

	maintainers := []*corev1.Maintainer{}
	for _, maintainer := range chartMetadata.Maintainers {
		m := &corev1.Maintainer{Name: maintainer.Name, Email: maintainer.Email}
		maintainers = append(maintainers, m)
	}

	return &corev1.AvailablePackageDetail{
		Name:             chartMetadata.Name,
		PkgVersion:       chartMetadata.Version,
		AppVersion:       chartMetadata.AppVersion,
		IconUrl:          chartMetadata.Icon,
		DisplayName:      chartMetadata.Name,
		ShortDescription: chartMetadata.Description,
		Readme:           detail[models.ReadmeKey],
		DefaultValues:    detail[models.ValuesKey],
		ValuesSchema:     detail[models.SchemaKey],
		Maintainers:      maintainers,
		// LongDescription ?
	}, nil
}
