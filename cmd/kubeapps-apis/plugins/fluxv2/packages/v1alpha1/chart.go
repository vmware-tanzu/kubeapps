// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strings"

	yaml "github.com/ghodss/yaml"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pkgutils "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	statuserror "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	tarutil "github.com/kubeapps/kubeapps/pkg/tarutil"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmchart "helm.sh/helm/v3/pkg/chart"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smetaunstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8stypes "k8s.io/apimachinery/pkg/types"
	k8dynamicclient "k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
)

// chart-related utilities

const (
	// see docs at https://fluxcd.io/docs/components/source/
	fluxHelmChart     = "HelmChart"
	fluxHelmCharts    = "helmcharts"
	fluxHelmChartList = "HelmChartList"
)

func (s *Server) getChartsResourceInterface(ctx context.Context, namespace string) (k8dynamicclient.ResourceInterface, error) {
	_, client, _, err := s.GetClients(ctx)
	if err != nil {
		return nil, err
	}

	chartsResource := k8sschema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmCharts,
	}

	return client.Resource(chartsResource).Namespace(namespace), nil
}

func (s *Server) listChartsInCluster(ctx context.Context, namespace string) (*k8smetaunstructuredv1.UnstructuredList, error) {
	resourceIfc, err := s.getChartsResourceInterface(ctx, namespace)
	if err != nil {
		return nil, err
	}

	if chartList, err := resourceIfc.List(ctx, k8smetav1.ListOptions{}); err != nil {
		return nil, statuserror.FromK8sError("list", "HelmCharts", "", err)
	} else {
		return chartList, nil
	}
}

func (s *Server) availableChartDetail(ctx context.Context, repoName k8stypes.NamespacedName, chartName, chartVersion string) (*pkgsGRPCv1alpha1.AvailablePackageDetail, error) {
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
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "chart [%s] not found", chartName)
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
			return nil, grpcstatus.Errorf(grpccodes.Internal, "failed to load details for chart [%s]", chartModel.ID)
		}
	}

	chartDetail, err := tarutil.FetchChartDetailFromTarball(bytes.NewReader(byteArray), chartID)
	if err != nil {
		return nil, err
	}

	return availablePackageDetailFromChartDetail(chartID, chartDetail)
}

func (s *Server) getChart(ctx context.Context, repo k8stypes.NamespacedName, chartName string) (*chartmodels.Chart, error) {
	if s.repoCache == nil {
		return nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "server cache has not been properly initialized")
	}

	key := s.repoCache.KeyForNamespacedName(repo)
	if entry, err := s.repoCache.GetForOne(key); err != nil {
		return nil, err
	} else if entry != nil {
		if typedEntry, ok := entry.(repoCacheEntryValue); !ok {
			return nil, grpcstatus.Errorf(
				grpccodes.Internal,
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

func passesFilter(chart chartmodels.Chart, filters *pkgsGRPCv1alpha1.FilterOptions) bool {
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

func filterAndPaginateCharts(filters *pkgsGRPCv1alpha1.FilterOptions, pageSize int32, pageOffset int, charts map[string][]chartmodels.Chart) ([]*pkgsGRPCv1alpha1.AvailablePackageSummary, error) {
	// this loop is here for 3 reasons:
	// 1) to convert from []interface{} which is what the generic cache implementation
	// returns for cache hits to a typed array object.
	// 2) perform any filtering of the results as needed, pending redis support for
	// querying values stored in cache (see discussion in https://github.com/kubeapps/kubeapps/issues/3032)
	// 3) if pagination was requested, only return up to one page size of results
	summaries := make([]*pkgsGRPCv1alpha1.AvailablePackageSummary, 0)
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
						return nil, grpcstatus.Errorf(
							grpccodes.Internal,
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

func availablePackageDetailFromChartDetail(chartID string, chartDetail map[string]string) (*pkgsGRPCv1alpha1.AvailablePackageDetail, error) {
	chartYaml, ok := chartDetail[chartmodels.ChartYamlKey]
	// TODO (gfichtenholt): if there is no chart yaml (is that even possible?),
	// fall back to chart info from repo index.yaml
	if !ok || chartYaml == "" {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "No chart manifest found for chart [%s]", chartID)
	}
	var chartMetadata helmchart.Metadata
	err := yaml.Unmarshal([]byte(chartYaml), &chartMetadata)
	if err != nil {
		return nil, err
	}

	maintainers := []*pkgsGRPCv1alpha1.Maintainer{}
	for _, maintainer := range chartMetadata.Maintainers {
		m := &pkgsGRPCv1alpha1.Maintainer{Name: maintainer.Name, Email: maintainer.Email}
		maintainers = append(maintainers, m)
	}

	var categories []string
	category, found := chartMetadata.Annotations["category"]
	if found && category != "" {
		categories = []string{category}
	}

	pkg := &pkgsGRPCv1alpha1.AvailablePackageDetail{
		Name: chartMetadata.Name,
		Version: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: chartMetadata.Version,
			AppVersion: chartMetadata.AppVersion,
		},
		HomeUrl:          chartMetadata.Home,
		IconUrl:          chartMetadata.Icon,
		DisplayName:      chartMetadata.Name,
		ShortDescription: chartMetadata.Description,
		Categories:       categories,
		Readme:           chartDetail[chartmodels.ReadmeKey],
		DefaultValues:    chartDetail[chartmodels.ValuesKey],
		ValuesSchema:     chartDetail[chartmodels.SchemaKey],
		SourceUrls:       chartMetadata.Sources,
		Maintainers:      maintainers,
		AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
			Identifier: chartID,
			Plugin:     GetPluginDetail(),
			Context:    &pkgsGRPCv1alpha1.Context{},
		},
	}
	// TODO: (gfichtenholt) LongDescription?

	// note, the caller will set pkg.AvailablePackageRef namespace as that information
	// is not included in the tarball
	return pkg, nil
}
