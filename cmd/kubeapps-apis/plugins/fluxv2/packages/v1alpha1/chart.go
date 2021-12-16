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
	"reflect"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
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

func (s *Server) getChartsResourceInterface(ctx context.Context, namespace string) (dynamic.ResourceInterface, error) {
	_, client, _, err := s.GetClients(ctx)
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

func (s *Server) listChartsInCluster(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	resourceIfc, err := s.getChartsResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}

	chartList, err := resourceIfc.List(ctx, metav1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "%q", err)
		} else if errors.IsForbidden(err) || errors.IsUnauthorized(err) {
			return nil, status.Errorf(codes.Unauthenticated, "unable to list charts due to %v", err)
		} else {
			return nil, status.Errorf(codes.Internal, "unable to list charts due to %v", err)
		}
	}
	return chartList, nil
}

func (s *Server) availableChartDetail(ctx context.Context, repoName types.NamespacedName, chartName, chartVersion string) (*corev1.AvailablePackageDetail, error) {
	log.V(4).Infof("+availableChartDetail(%s, %s, %s)", repoName, chartName, chartVersion)
	chartModel, err := s.getChart(ctx, repoName, chartName)
	if err != nil {
		return nil, err
	} else if chartModel == nil {
		return nil, status.Errorf(codes.NotFound, "Chart [%s] not found", chartName)
	}

	opts, err := s.clientOptionsForRepo(ctx, repoName)
	if err != nil {
		return nil, err
	}

	if chartVersion == "" {
		chartVersion = chartModel.ChartVersions[0].Version
	}
	key, err := s.chartCache.KeyFor(repoName.Namespace, chartModel.ID, chartVersion)
	if err != nil {
		return nil, err
	}

	byteArray, err := s.chartCache.GetForOne(key, chartModel, opts)
	if err != nil {
		return nil, err
	} else if byteArray == nil {
		return nil, status.Errorf(codes.Internal, "Failed to load details for chart [%s]", chartModel.ID)
	}

	chartDetail, err := tarutil.FetchChartDetailFromTarball(bytes.NewReader(byteArray), chartModel.ID)
	if err != nil {
		return nil, err
	}

	return availablePackageDetailFromChartDetail(chartModel.ID, chartDetail)
}

func (s *Server) getChart(ctx context.Context, repo types.NamespacedName, chartName string) (*models.Chart, error) {
	if s.repoCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	}

	key := s.repoCache.KeyForNamespacedName(repo)
	if entry, err := s.repoCache.GetForOne(key); err != nil {
		return nil, err
	} else if entry != nil {
		if typedEntry, ok := entry.(repoCacheEntryValue); !ok {
			return nil, status.Errorf(
				codes.Internal,
				"unexpected value fetched from cache: type: [%s], value: [%v]", reflect.TypeOf(entry), entry)
		} else {
			for _, chart := range typedEntry.Charts {
				if chart.Name == chartName {
					return &chart, nil // found it
				}
			}
		}
	}
	return nil, nil
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
		Plugin:     common.GetPluginDetail(),
	}
	pkg.AvailablePackageRef.Context = &corev1.Context{Namespace: chart.Repo.Namespace}

	if chart.ChartVersions != nil || len(chart.ChartVersions) != 0 {
		pkg.LatestVersion = &corev1.PackageAppVersion{
			PkgVersion: chart.ChartVersions[0].Version,
			AppVersion: chart.ChartVersions[0].AppVersion,
		}
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

func filterAndPaginateCharts(filters *corev1.FilterOptions, pageSize int32, pageOffset int, charts map[string][]models.Chart) ([]*corev1.AvailablePackageSummary, error) {
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
	for _, packages := range charts {
		for _, chart := range packages {
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
	return summaries, nil
}

func availablePackageDetailFromChartDetail(chartID string, chartDetail map[string]string) (*corev1.AvailablePackageDetail, error) {
	chartYaml, ok := chartDetail[models.ChartYamlKey]
	// TODO (gfichtenholt): if there is no chart yaml (is that even possible?),
	// fall back to chart info from repo index.yaml
	if !ok || chartYaml == "" {
		return nil, status.Errorf(codes.Internal, "No chart manifest found for chart [%s]", chartID)
	}
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

	var categories []string
	category, found := chartMetadata.Annotations["category"]
	if found && category != "" {
		categories = []string{category}
	}

	pkg := &corev1.AvailablePackageDetail{
		Name: chartMetadata.Name,
		Version: &corev1.PackageAppVersion{
			PkgVersion: chartMetadata.Version,
			AppVersion: chartMetadata.AppVersion,
		},
		HomeUrl:          chartMetadata.Home,
		IconUrl:          chartMetadata.Icon,
		DisplayName:      chartMetadata.Name,
		ShortDescription: chartMetadata.Description,
		Categories:       categories,
		Readme:           chartDetail[models.ReadmeKey],
		DefaultValues:    chartDetail[models.ValuesKey],
		ValuesSchema:     chartDetail[models.SchemaKey],
		SourceUrls:       chartMetadata.Sources,
		Maintainers:      maintainers,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: chartID,
			Plugin:     common.GetPluginDetail(),
			Context:    &corev1.Context{},
		},
	}
	// TODO: (gfichtenholt) LongDescription?

	// note, the caller will set pkg.AvailablePackageRef namespace as that information
	// is not included in the tarball
	return pkg, nil
}

// packageAppVersionsSummary converts the model chart versions into the required version summary.
func packageAppVersionsSummary(versions []models.ChartVersion) []*corev1.PackageAppVersion {
	pav := []*corev1.PackageAppVersion{}

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
		pav = append(pav, &corev1.PackageAppVersion{
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
