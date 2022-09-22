// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/tarutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func (s *Server) getChartInCluster(ctx context.Context, key types.NamespacedName) (*sourcev1.HelmChart, error) {
	client, err := s.getClient(ctx, key.Namespace)
	if err != nil {
		return nil, err
	}
	var chartObj sourcev1.HelmChart
	if err = client.Get(ctx, key, &chartObj); err != nil {
		return nil, statuserror.FromK8sError("get", "HelmChart", key.String(), err)
	}
	return &chartObj, nil
}

// TODO (gfichtenholt) this func is too long. Break it up
func (s *Server) availableChartDetail(ctx context.Context, packageRef *corev1.AvailablePackageReference, chartVersion string) (*corev1.AvailablePackageDetail, error) {
	log.Infof("+availableChartDetail(%s, %s)", packageRef, chartVersion)

	repoN, chartName, err := pkgutils.SplitPackageIdentifier(packageRef.Identifier)
	if err != nil {
		return nil, err
	}

	// check specified repo exists and is in ready state
	repoName := types.NamespacedName{Namespace: packageRef.Context.Namespace, Name: repoN}

	// this verifies that the repo exists
	repo, err := s.getRepoInCluster(ctx, repoName)
	if err != nil {
		return nil, err
	} else if !isRepoReady(*repo) {
		return nil, status.Errorf(codes.Internal, "repository [%s] is not in Ready state", repoName)
	}

	chartID := fmt.Sprintf("%s/%s", repoName.Name, chartName)
	// first, try the happy path - we have the chart version and the corresponding entry
	// happens to be in the cache
	var byteArray []byte
	if chartVersion != "" {
		if key, err := s.chartCache.KeyFor(repoName.Namespace, chartID, chartVersion); err != nil {
			return nil, err
		} else if byteArray, err = s.chartCache.Fetch(key); err != nil {
			return nil, err
		}
	}

	if byteArray == nil {
		// no specific chart version was provided or a cache miss, need to do a bit of work
		chartModel, err := s.getChartModel(ctx, repoName, chartName)
		if err != nil {
			return nil, err
		} else if chartModel == nil {
			return nil, status.Errorf(codes.NotFound, "chart [%s] not found", chartName)
		}

		if chartVersion == "" {
			chartVersion = chartModel.ChartVersions[0].Version
		}

		var key string
		if key, err = s.chartCache.KeyFor(repoName.Namespace, chartID, chartVersion); err != nil {
			return nil, err
		}

		var fn cache.DownloadChartFn
		if chartModel.Repo.Type == "oci" {
			if ociRepo, err := s.newOCIChartRepositoryAndLogin(ctx, repoName); err != nil {
				return nil, err
			} else {
				fn = downloadOCIChartFn(ociRepo)
			}
		} else {
			if opts, err := s.httpClientOptionsForRepo(ctx, repoName); err != nil {
				return nil, err
			} else {
				fn = downloadHttpChartFn(opts)
			}
		}
		if byteArray, err = s.chartCache.Get(key, chartModel, fn); err != nil {
			return nil, err
		}

		if byteArray == nil {
			return nil, status.Errorf(codes.Internal, "failed to load details for chart [%s]", chartModel.ID)
		}
	}

	chartDetail, err := tarutil.FetchChartDetailFromTarball(bytes.NewReader(byteArray), chartID)
	if err != nil {
		return nil, err
	}
	log.Infof("checkpoint 5")

	pkgDetail, err := availablePackageDetailFromChartDetail(chartID, chartDetail)
	if err != nil {
		return nil, err
	}
	log.Infof("checkpoint 6")

	// fix up a couple of fields that don't come from the chart tarball
	repoUrl := repo.Spec.URL
	if repoUrl == "" {
		return nil, status.Errorf(codes.NotFound, "Missing required field spec.url on repository %q", repoName)
	}

	pkgDetail.RepoUrl = repoUrl
	pkgDetail.AvailablePackageRef.Context.Namespace = packageRef.Context.Namespace
	// per https://github.com/vmware-tanzu/kubeapps/pull/3686#issue-1038093832
	pkgDetail.AvailablePackageRef.Context.Cluster = s.kubeappsCluster
	return pkgDetail, nil
}

func (s *Server) getChartModel(ctx context.Context, repoName types.NamespacedName, chartName string) (*models.Chart, error) {
	if s.repoCache == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server cache has not been properly initialized")
	} else if ok, err := s.hasAccessToNamespace(ctx, common.GetChartsGvr(), repoName.Namespace); err != nil {
		return nil, err
	} else if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "user has no [get] access for HelmCharts in namespace [%s]", repoName.Namespace)
	}

	key := s.repoCache.KeyForNamespacedName(repoName)
	value, err := s.repoCache.Get(key)
	if err != nil {
		return nil, err
	} else {
		typedValue, err := s.repoCacheEntryFromUntyped(key, value)
		if err != nil {
			return nil, err
		} else if typedValue == nil {
			return nil, nil
		} else {
			for _, chart := range typedValue.Charts {
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

func filterAndPaginateCharts(filters *corev1.FilterOptions, pageSize int32, itemOffset int, charts map[string][]models.Chart) ([]*corev1.AvailablePackageSummary, error) {
	// this loop is here for 3 reasons:
	// 1) to convert from []interface{} which is what the generic cache implementation
	// returns for cache hits to a typed array object.
	// 2) perform any filtering of the results as needed, pending redis support for
	// querying values stored in cache (see discussion in https://github.com/vmware-tanzu/kubeapps/issues/3032)
	// 3) if pagination was requested, only return up to one page size of results
	summaries := make([]*corev1.AvailablePackageSummary, 0)
	i := 0
	startAt := -1
	if pageSize > 0 {
		startAt = itemOffset
	}
	for _, packages := range charts {
		for p := range packages {
			chart := packages[p] // avoid implicit memory aliasing
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
		return nil, status.Errorf(codes.Internal, "No chart manifest found for chart: [%s]", chartID)
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

func downloadHttpChartFn(options *common.HttpClientOptions) func(chartID, chartUrl, chartVersion string) ([]byte, error) {
	return func(chartID, chartUrl, chartVersion string) ([]byte, error) {
		client, headers, err := common.NewHttpClientAndHeaders(options)
		if err != nil {
			return nil, err
		}

		reader, _, err := httpclient.GetStream(chartUrl, client, headers)
		if reader != nil {
			defer reader.Close()
		}
		if err != nil {
			return nil, err
		}

		chartTgz, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		return chartTgz, nil
	}
}
