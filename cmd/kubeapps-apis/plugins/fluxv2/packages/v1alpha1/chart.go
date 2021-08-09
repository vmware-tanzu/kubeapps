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
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	httpclient "github.com/kubeapps/kubeapps/pkg/http-client"
	tar "github.com/kubeapps/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/helm/pkg/proto/hapi/chart"
	log "k8s.io/klog/v2"
)

// chart-related utilities

const (
	// see docs at https://fluxcd.io/docs/components/source/
	fluxHelmChart     = "HelmChart"
	fluxHelmCharts    = "helmcharts"
	fluxHelmChartList = "HelmChartList"

	MajorVersionsInSummary = 3
	MinorVersionsInSummary = 3
	PatchVersionsInSummary = 3
)

func (s *Server) listChartsInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	resourceIfc, err := s.getChartsResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}

	// see if we the chart already exists
	// TODO (gfichtenholt):
	// see https://github.com/kubeapps/kubeapps/pull/2915
	// for context. It'd be better if we could filter on server-side. The problem is the set of supported
	// fields in FieldSelector is very small. things like "spec.chart" or "status.artifact.revision" are
	// certainly not supported.
	// see
	//  - kubernetes/client-go#713 and
	//  - https://github.com/flant/shell-operator/blob/8fa3c3b8cfeb1ddb37b070b7a871561fdffe788b///HOOKS.md#fieldselector and
	//  - https://github.com/kubernetes/kubernetes/issues/53459
	chartList, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return chartList, nil
}

// returns the url from which chart .tgz can be downloaded
// here chartVersion string, if specified at all, should be specific, like "14.4.0",
// not an expression like ">14 <15"
func (s *Server) getChartTarball(ctx context.Context, repoName string, chartName string, namespace string, chartVersion string) (url string, err error, cleanUp func()) {
	// see if we the chart already exists
	// TODO (gfichtenholt):
	// see https://github.com/kubeapps/kubeapps/pull/2915
	// for context. It'd be better if we could filter on server-side. The problem is the set of supported
	// fields in FieldSelector is very small. things like "spec.chart" or "status.artifact.revision" are
	// certainly not supported.
	// see
	//  - kubernetes/client-go#713 and
	//  - https://github.com/flant/shell-operator/blob/8fa3c3b8cfeb1ddb37b070b7a871561fdffe788b///HOOKS.md#fieldselector and
	//  - https://github.com/kubernetes/kubernetes/issues/53459
	chartList, err := s.listChartsInCluster(ctx, namespace)
	if err != nil {
		return "", err, nil
	}

	url, err = findUrlForChartInList(chartList, repoName, chartName, chartVersion)
	if err != nil {
		return "", err, nil
	} else if url != "" {
		return url, nil, nil
	}

	// did not find the chart, need to create
	// see https://fluxcd.io/docs/components/source/helmcharts/
	// notes:
	// 1. HelmChart object needs to be co-located in the same namespace as the HelmRepository it is referencing.
	// 2. As of the time of this writing, flux impersonates a "super" user when doing this
	// (see fluxv2 plug-in specific notes at the end of design doc). However, they are backing away from
	// this model toward this proposal
	// https://github.com/fluxcd/flux2/blob/1c5a25313561771d585c4192d7f330b45753cd99/docs/proposals/secure-impersonation.md
	// So we may not necessarily want to follow what flux does today
	unstructuredChart := newFluxHelmChart(chartName, repoName, chartVersion)

	resourceIfc, err := s.getChartsResourceInterface(ctx, namespace)
	if err != nil {
		return "", err, nil
	}

	newChart, err := resourceIfc.Create(ctx, &unstructuredChart, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("Error creating chart: %v\n%v", err, unstructuredChart)
		return "", err, nil
	}

	log.Infof("Created chart: [%v]", prettyPrintMap(newChart.Object))

	// Delete the created helm chart regardless of success or failure. At the end of
	// GetAvailablePackageDetail(), we've already collected the information we need,
	// so why leave a flux chart chart object hanging around?
	// Over time, they could accumulate to a very large number...
	cleanUp = func() {
		if err = resourceIfc.Delete(ctx, newChart.GetName(), metav1.DeleteOptions{}); err != nil {
			log.Errorf("Failed to delete flux helm chart [%v]", prettyPrintMap(newChart.Object))
		}
	}

	watcher, err := resourceIfc.Watch(ctx, metav1.ListOptions{
		ResourceVersion: newChart.GetResourceVersion(),
	})
	if err != nil {
		log.Errorf("Error creating watch: %v\n%v", err, unstructuredChart)
		return "", err, cleanUp
	}

	// wait til wait until flux reconciles and we have chart url available
	// TODO (gfichtenholt) note that, unlike with ResourceWatcherCache, the
	// wait time window is very short here so I am not employing the RetryWatcher
	// technique here for now
	url, err = waitUntilChartPullComplete(ctx, watcher)
	watcher.Stop()
	// only the caller should call cleanUp() when it's done with the url,
	// if we call it here, the caller will end up with a dangling link
	return url, err, cleanUp
}

func (s *Server) getChartsResourceInterface(ctx context.Context, namespace string) (dynamic.ResourceInterface, error) {
	client, err := s.getDynamicClient(ctx)
	if err != nil {
		return nil, err
	}

	chartsResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmCharts,
	}

	return client.Resource(chartsResource).Namespace(namespace), nil
}

func (s *Server) fetchChartFromCache(repoNamespace, repoName, chartName string) (*models.Chart, error) {
	if s.cache == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	}

	charts, err := s.cache.fetchForOne(s.cache.keyForNamespaceAndName(repoNamespace, repoName))
	if err != nil {
		return nil, err
	}

	if charts != nil {
		if typedCharts, ok := charts.([]models.Chart); !ok {
			return nil, status.Errorf(
				codes.Internal,
				"unexpected value fetched from cache: %v", charts)
		} else {
			for _, chart := range typedCharts {
				if chart.Name == chartName {
					return &chart, nil // found it
				}
			}
		}
	}
	return nil, nil
}

// the main goal of this func is to answer whether or not to stop waiting for chart reconciliation
// which is different from answering whether the chart was pulled successfully
// TODO (gfichtenholt): As above, hopefully this func isn't required if we can only list charts that we know are ready.
func isChartPullComplete(unstructuredChart map[string]interface{}) (complete, success bool, reason string) {
	// see docs at https://fluxcd.io/docs/components/source/helmcharts/
	// Confirm the state we are observing is for the current generation
	if !checkGeneration(unstructuredChart) {
		return false, false, ""
	} else {
		return checkStatusReady(unstructuredChart)
	}
}

// TODO (gfichtenholt):
// see https://github.com/kubeapps/kubeapps/pull/2915 for context
// In the future you might instead want to consider something like
// passing a results channel (of string urls) to getChartTarball, so it returns
// immediately and you wait on the results channel at the call-site, which would mean
// you could call it for 20 different charts and just wait for the results to come in
// whatever order they happen to take, rather than serially.
func waitUntilChartPullComplete(ctx context.Context, watcher watch.Interface) (string, error) {
	ch := watcher.ResultChan()
	log.Infof("Waiting until chart pull is complete...")

	// unit test-related trigger that allows another concurrently running goroutine to
	// mock sending a watch Modify event to the channel at this point
	wg, ok := fromContext(ctx)
	if ok && wg != nil {
		wg.Done()
	}

	for {
		event, ok := <-ch
		if !ok {
			// let the user retry
			return "", status.Errorf(codes.Internal, "operation failed because a channel was closed")
		}
		if event.Type == watch.Modified {
			unstructuredChart, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				return "", status.Errorf(codes.Internal, "could not cast to unstructured.Unstructured")
			}

			done, success, reason := isChartPullComplete(unstructuredChart.Object)
			if done {
				if success {
					url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url")
					if err != nil || !found {
						return "", status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
					}
					return url, nil
				} else {
					return "", status.Errorf(codes.Internal, "Failed to pull chart due to %s", reason)
				}
			}
		} else {
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
	for _, maintainer := range chart.Maintainers {
		if maintainer.Name == "" {
			return false, status.Errorf(codes.Internal, "required field .Maintainers[i].Name not found on helm chart: %v", chart)
		}
	}
	return true, nil
}

// note that chartVersion here could be a semver constraint expression, e.g. something like "<= 6.7.1",
// as opposed to a simple literal expression, like "1.2.3"
// see https://github.com/Masterminds/semver/blob/master/README.md#checking-version-constraints
func findUrlForChartInList(chartList *unstructured.UnstructuredList, repoName, chartName, chartVersion string) (string, error) {
	var semVerConstraints *semver.Constraints
	if chartVersion != "" {
		var err error
		if semVerConstraints, err = semver.NewConstraint(chartVersion); err != nil {
			return "", err
		}
	}
	for _, unstructuredChart := range chartList.Items {
		thisChartName, found, err := unstructured.NestedString(unstructuredChart.Object, "spec", "chart")
		thisRepoName, found2, err2 := unstructured.NestedString(unstructuredChart.Object, "spec", "sourceRef", "name")

		if err == nil && err2 == nil && found && found2 && repoName == thisRepoName && chartName == thisChartName {
			if done, success, reason := isChartPullComplete(unstructuredChart.Object); done {
				if success {
					if url, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "url"); err != nil || !found {
						return "", status.Errorf(codes.Internal, "expected field status.url not found on HelmChart: %v:\n%v", err, unstructuredChart)
					} else {
						if semVerConstraints != nil {
							// refer to https://github.com/fluxcd/source-controller/blob/main/api/v1beta1/helmchart_types.go &
							// https://github.com/fluxcd/source-controller/blob/40a47670aadebc0f4e3a623be47725106bac2d55/api/v1beta1/artifact_types.go#L27
							artifactVerString, found, err := unstructured.NestedString(unstructuredChart.Object, "status", "artifact", "revision")
							if err != nil || !found {
								return "", status.Errorf(codes.Internal, "expected field status.artifact.revision not found on HelmChart: %v:\n%v", err, unstructuredChart)
							} else if artifactVerString != "" && semVerConstraints != nil {
								if artifactVer, err := semver.NewVersion(artifactVerString); err != nil {
									return "", err
								} else if !semVerConstraints.Check(artifactVer) {
									continue
								}
							}
						}
						log.Infof("Found existing HelmChart for: [%s/%s]", repoName, chartName)
						return url, nil
					}
				} else {
					return "", status.Errorf(codes.Internal, "Chart pull failed due to %s", reason)
				}
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

func filterAndPaginateCharts(filters *corev1.FilterOptions, pageSize int32, pageOffset int, cachedCharts map[string]interface{}) ([]*corev1.AvailablePackageSummary, error) {
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

func availablePackageDetailFromTarball(chartID, url string) (*corev1.AvailablePackageDetail, error) {
	// fetch, unzip and untar .tgz file
	// no need to provide authz, userAgent or any of the TLS details, as we are pulling .tgz file from
	// local cluster, not remote repo.
	// E.g. http://source-controller.flux-system.svc.cluster.local./helmchart/default/redis-j6wtx/redis-latest.tgz
	// Flux does the hard work of pulling the bits from remote repo
	// based on secretRef associated with HelmRepository, if applicable
	chartDetail, err := tar.FetchChartDetailFromTarball(chartID, url, "", "", httpclient.New())
	if err != nil {
		return nil, err
	}

	chartYaml := chartDetail[models.ChartYamlKey]
	// TODO (gfichtenholt): if there is no chart yaml (is that even possible?), fall back to chart info from
	// repo index.yaml
	var chartMetadata chart.Metadata
	err = yaml.Unmarshal([]byte(chartYaml), &chartMetadata)
	if err != nil {
		return nil, err
	}

	maintainers := []*corev1.Maintainer{}
	for _, maintainer := range chartMetadata.Maintainers {
		m := &corev1.Maintainer{Name: maintainer.Name, Email: maintainer.Email}
		maintainers = append(maintainers, m)
	}

	var categories []string
	category, found := chartMetadata.Annotations["category"]
	if found && category != "" {
		categories = []string{category}
	}

	pkg := &corev1.AvailablePackageDetail{
		Name:             chartMetadata.Name,
		PkgVersion:       chartMetadata.Version,
		AppVersion:       chartMetadata.AppVersion,
		IconUrl:          chartMetadata.Icon,
		DisplayName:      chartMetadata.Name,
		ShortDescription: chartMetadata.Description,
		Categories:       categories,
		Readme:           chartDetail[models.ReadmeKey],
		DefaultValues:    chartDetail[models.ValuesKey],
		ValuesSchema:     chartDetail[models.SchemaKey],
		Maintainers:      maintainers,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: chartID,
			Plugin:     GetPluginDetail(),
			Context:    &corev1.Context{},
		},
	}
	// TODO: (gfichtenholt) LongDescription?

	// note, the caller will set pkg.AvailablePackageRef namespace as that information
	// is not included in the tarball
	return pkg, nil
}

// packageAppVersionsSummary converts the model chart versions into the required version summary.
func packageAppVersionsSummary(versions []models.ChartVersion) []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion {
	pav := []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{}

	// Use a version map to be able to count how many major, minor and patch versions
	// we have included.
	version_map := map[int64]map[int64][]int64{}
	for _, v := range versions {
		version, err := semver.NewVersion(v.Version)
		if err != nil {
			continue
		}

		if _, ok := version_map[version.Major()]; !ok {
			// Don't add a new major version if we already have enough
			if len(version_map) >= MajorVersionsInSummary {
				continue
			}
		} else {
			// If we don't yet have this minor version
			if _, ok := version_map[version.Major()][version.Minor()]; !ok {
				// Don't add a new minor version if we already have enough for this major version
				if len(version_map[version.Major()]) >= MinorVersionsInSummary {
					continue
				}
			} else {
				if len(version_map[version.Major()][version.Minor()]) >= PatchVersionsInSummary {
					continue
				}
			}
		}

		// Include the version and update the version map.
		pav = append(pav, &corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
			PkgVersion: v.Version,
			AppVersion: v.AppVersion,
		})

		if _, ok := version_map[version.Major()]; !ok {
			version_map[version.Major()] = map[int64][]int64{}
		}
		version_map[version.Major()][version.Minor()] = append(version_map[version.Major()][version.Minor()], version.Patch())
	}
	return pav
}
