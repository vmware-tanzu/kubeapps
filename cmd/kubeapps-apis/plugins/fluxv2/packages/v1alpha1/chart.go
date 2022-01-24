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
	"fmt"
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
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

	if chartList, err := resourceIfc.List(ctx, metav1.ListOptions{}); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmCharts", "", err)
	} else {
		return chartList, nil
	}
}

func (s *Server) availableChartDetail(ctx context.Context, repoName types.NamespacedName, chartName, chartVersion string) (*corev1.AvailablePackageDetail, error) {
	log.Infof("+availableChartDetail(%s, %s, %s)", repoName, chartName, chartVersion)

	chartID := fmt.Sprintf("%s/%s", repoName.Name, chartName)
	// first, try the happy path - we have the chart version and the corresponding entry
	// happens to be in the cache
	var byteArray []byte
	if chartVersion != "" {
		if key, err := s.chartCache.KeyFor(repoName.Namespace, chartID, chartVersion); err != nil {
			return nil, err
		} else if byteArray, err = s.chartCache.FetchForOne(key); err != nil {
			return nil, err
		}
	}

	if byteArray == nil {
		// no specific chart version was provided or a cache miss, need to do a bit of work
		chartModel, err := s.getChart(ctx, repoName, chartName)
		if err != nil {
			return nil, err
		} else if chartModel == nil {
			return nil, status.Errorf(codes.NotFound, "chart [%s] not found", chartName)
		}

		if chartVersion == "" {
			chartVersion = chartModel.ChartVersions[0].Version
		}

		if key, err := s.chartCache.KeyFor(repoName.Namespace, chartID, chartVersion); err != nil {
			return nil, err
		} else if opts, err := s.clientOptionsForRepo(ctx, repoName); err != nil {
			return nil, err
		} else if byteArray, err = s.chartCache.GetForOne(key, chartModel, opts); err != nil {
			return nil, err
		} else if byteArray == nil {
			return nil, status.Errorf(codes.Internal, "failed to load details for chart [%s]", chartModel.ID)
		}
	}

	chartDetail, err := tarutil.FetchChartDetailFromTarball(bytes.NewReader(byteArray), chartID)
	if err != nil {
		return nil, err
	}

	return availablePackageDetailFromChartDetail(chartID, chartDetail)
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
					pkg, err := pkgutils.AvailablePackageSummaryFromChart(&chart, GetPluginDetail())
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
			Plugin:     GetPluginDetail(),
			Context:    &corev1.Context{},
		},
	}
	// TODO: (gfichtenholt) LongDescription?

	// note, the caller will set pkg.AvailablePackageRef namespace as that information
	// is not included in the tarball
	return pkg, nil
}
