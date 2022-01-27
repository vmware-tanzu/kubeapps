// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Currently these tests will be skipped entirely unless the
// ENABLE_PG_INTEGRATION_TESTS env var is set.
// Run the local postgres with
// docker run --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:11.14.0-debian-10-r28
// in another terminal.
package utils

import (
	"database/sql"
	"sort"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	dbutilstest "github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest"
	pgtest "github.com/kubeapps/kubeapps/pkg/dbutils/dbutilstest/pgtest"
	_ "github.com/lib/pq"
)

func getInitializedManager(t *testing.T) (*PostgresAssetManager, func()) {
	pam, cleanup := pgtest.GetInitializedManager(t)
	return &PostgresAssetManager{pam}, cleanup
}

func TestGetChart(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const repoName = "repo-name"

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace
		existingCharts map[string][]chartmodels.Chart
		chartId        string
		namespace      string
		expectedChart  string
		expectedErr    error
	}{
		{
			name:        "it returns an error if the chart does not exist",
			chartId:     "doesnt-exist-1",
			namespace:   "doesnt-exist",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it returns an error if the chart does not exist in that repo",
			existingCharts: map[string][]chartmodels.Chart{
				"namespace-1": []chartmodels.Chart{
					chartmodels.Chart{ID: "chart-1", Name: "my-chart"},
				},
			},
			chartId:     "chart-1",
			namespace:   "other-namespace",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it returns the chart matching the chartid",
			existingCharts: map[string][]chartmodels.Chart{
				"namespace-1": []chartmodels.Chart{
					chartmodels.Chart{ID: "chart-1", Name: "my-chart"},
				},
			},
			chartId:       "chart-1",
			namespace:     "namespace-1",
			expectedErr:   nil,
			expectedChart: "my-chart",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, charts := range tc.existingCharts {
				pgtest.EnsureChartsExist(t, pam, charts, chartmodels.Repo{Name: repoName, Namespace: namespace})
			}

			chart, err := pam.GetChart(tc.namespace, tc.chartId)

			if got, want := err, tc.expectedErr; got != want {
				t.Fatalf("In '"+tc.name+"': "+"got: %+v, want: %+v", got, want)
			}
			if got, want := chart.Name, tc.expectedChart; got != want {
				t.Errorf("In '"+tc.name+"': "+"got: %q, want: %q", got, want)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const repoName = "repo-name"

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace
		existingCharts   map[string][]chartmodels.Chart
		chartId          string
		namespace        string
		requestedVersion string
		expectedVersion  string
		expectedErr      error
	}{
		{
			name:        "it returns an error if the chart does not exist",
			chartId:     "doesnt-exist-1",
			namespace:   "doesnt-exist",
			expectedErr: sql.ErrNoRows,
		},
		{
			name: "it returns an error if the chart version does not exist",
			existingCharts: map[string][]chartmodels.Chart{
				"namespace-1": []chartmodels.Chart{
					chartmodels.Chart{ID: "chart-1", ChartVersions: []chartmodels.ChartVersion{
						chartmodels.ChartVersion{Version: "1.2.3"},
					}},
				},
			},
			chartId:          "chart-1",
			namespace:        "namespace-1",
			requestedVersion: "doesnt-exist",
			expectedErr:      ErrChartVersionNotFound,
		},
		{
			name: "it returns an error if the chart version does not exist in that namespace",
			existingCharts: map[string][]chartmodels.Chart{
				"namespace-1": []chartmodels.Chart{
					chartmodels.Chart{ID: "chart-1", ChartVersions: []chartmodels.ChartVersion{
						chartmodels.ChartVersion{Version: "1.2.3"},
					}},
				},
			},
			chartId:          "chart-1",
			namespace:        "other-namespace",
			requestedVersion: "1.2.3",
			expectedErr:      sql.ErrNoRows,
		},
		{
			name: "it returns the chart version matching the chartid and version",
			existingCharts: map[string][]chartmodels.Chart{
				"namespace-1": []chartmodels.Chart{
					chartmodels.Chart{ID: "chart-1", ChartVersions: []chartmodels.ChartVersion{
						chartmodels.ChartVersion{Version: "1.2.3"},
						chartmodels.ChartVersion{Version: "4.5.6"},
					}},
				},
			},
			chartId:          "chart-1",
			namespace:        "namespace-1",
			requestedVersion: "1.2.3",
			expectedVersion:  "1.2.3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, charts := range tc.existingCharts {
				pgtest.EnsureChartsExist(t, pam, charts, chartmodels.Repo{Name: repoName, Namespace: namespace})
			}

			chart, err := pam.GetChartVersion(tc.namespace, tc.chartId, tc.requestedVersion)

			if got, want := err, tc.expectedErr; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
			if tc.expectedErr != nil {
				return
			}
			// The function just returns the chart with only the one version.
			if got, want := len(chart.ChartVersions), 1; got != want {
				t.Fatalf("got: %d, want: %d", got, want)
			}
			if got, want := chart.ChartVersions[0].Version, tc.expectedVersion; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}

func TestGetPaginatedChartList(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const (
		repoName      = "repo-name"
		namespaceName = "namespace-name"
	)

	chartVersions := []chartmodels.ChartVersion{
		chartmodels.ChartVersion{
			Digest: "abc-123",
		},
	}

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace and repo
		existingCharts   map[string]map[string][]chartmodels.Chart
		namespace        string
		repo             string
		expectedCharts   []*chartmodels.Chart
		expectedErr      error
		expectedNumPages int
	}{
		{
			name:             "it returns an empty list if the repo or namespace do not exist",
			repo:             "repo-doesnt-exist",
			namespace:        "doesnt-exist",
			expectedCharts:   []*chartmodels.Chart{},
			expectedNumPages: 0,
		},
		{
			name:      "it returns charts from a specific repo in a specific namespace",
			repo:      repoName,
			namespace: namespaceName,
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
					},
				},
			},
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
			},
			expectedNumPages: 1,
		},
		{
			name: "it returns charts from multiple repos in a specific namespace",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
					},
				},
			},
			repo:      "",
			namespace: namespaceName,
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
				&chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
			},
			expectedNumPages: 1,
		},
		{
			name: "it includes charts from global repositories and the specific namespace",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
					},
				},
				dbutilstest.KubeappsTestNamespace: map[string][]chartmodels.Chart{
					"global-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "global-repo/global-chart", Name: "global-chart"},
					},
				},
			},
			repo:      "",
			namespace: namespaceName,
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
				&chartmodels.Chart{ID: "global-repo/global-chart", Name: "global-chart"},
				&chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
			},
			expectedNumPages: 1,
		},
		{
			name: "it returns charts from multiple repos across all namespaces",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
					},
				},
			},
			repo:      "",
			namespace: "_all",
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
				&chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
				&chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
			},
			expectedNumPages: 1,
		},
		{
			name: "it returns charts from a single repo across all namespaces",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
					},
				},
			},
			repo:      repoName,
			namespace: "_all",
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1"},
				&chartmodels.Chart{ID: repoName + "/chart-in-other-namespace", Name: "chart-in-other-namespace"},
			},
			expectedNumPages: 1,
		},
		{
			name: "it does not remove duplicates",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: map[string][]chartmodels.Chart{
					repoName: []chartmodels.Chart{
						chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1", ChartVersions: chartVersions},
					},
					"other-repo": []chartmodels.Chart{
						chartmodels.Chart{ID: "other-repo/same-chart-different-repo", Name: "same-chart-different-repo", ChartVersions: chartVersions},
					},
				},
			},
			repo:      "",
			namespace: namespaceName,
			expectedCharts: []*chartmodels.Chart{
				&chartmodels.Chart{ID: repoName + "/chart-1", Name: "chart-1", ChartVersions: chartVersions},
				&chartmodels.Chart{ID: "other-repo/same-chart-different-repo", Name: "same-chart-different-repo", ChartVersions: chartVersions},
			},
			expectedNumPages: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, chartsPerRepo := range tc.existingCharts {
				for repo, charts := range chartsPerRepo {
					pgtest.EnsureChartsExist(t, pam, charts, chartmodels.Repo{Name: repo, Namespace: namespace})
				}
			}

			charts, numPages, err := pam.GetPaginatedChartListWithFilters(ChartQuery{Namespace: tc.namespace, Repos: []string{tc.repo}}, 1, 10)

			if got, want := err, tc.expectedErr; got != want {
				t.Fatalf("In '"+tc.name+"': "+"got err: %+v, want: %+v", got, want)
			}
			if got, want := numPages, tc.expectedNumPages; got != want {
				t.Fatalf("In '"+tc.name+"': "+"got numPages: %+v, want: %+v", got, want)
			}
			if got, want := charts, tc.expectedCharts; !cmp.Equal(want, got) {
				t.Errorf("In '"+tc.name+"': "+"mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

// ByID implements sort.Interface for []chartmodels.Chart based on
// the ID field.
type byID []*chartmodels.Chart

func (a byID) Len() int           { return len(a) }
func (a byID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byID) Less(i, j int) bool { return a[i].ID < a[j].ID }

func TestGetChartsWithFilters(t *testing.T) {
	pgtest.SkipIfNoDB(t)
	const (
		repoName1     = "repo-name-1"
		repoName2     = "repo-name-2"
		namespaceName = "namespace-name"
	)

	chartVersion := chartmodels.ChartVersion{
		Digest:     "abc-123",
		Version:    "1.0chart",
		AppVersion: "2.0app",
	}

	chartVersions := []chartmodels.ChartVersion{chartVersion}

	chartWithVersionRepo1 := chartmodels.Chart{ID: repoName1 + "/chart-1", Name: "chart-1", ChartVersions: chartVersions}
	chartWithVersionRepo2 := chartmodels.Chart{ID: repoName2 + "/chart-1", Name: "chart-1", ChartVersions: chartVersions}

	testCases := []struct {
		name string
		// existingCharts is a map of charts per namespace and repo
		existingCharts map[string]map[string][]chartmodels.Chart
		namespace      string
		chartName      string
		chartVersion   string
		appVersion     string
		expectedCharts []*chartmodels.Chart
		expectedErr    error
	}{
		{
			name: "returns charts in the specific namespace",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: {
					repoName1: {chartWithVersionRepo1},
					"other-repo": []chartmodels.Chart{
						{ID: "other-repo/other-chart", Name: "other-chart"},
					},
				},
				"other-namespace": {
					repoName1: {chartWithVersionRepo1},
				},
			},
			namespace:    namespaceName,
			chartName:    chartWithVersionRepo1.Name,
			chartVersion: chartWithVersionRepo1.ChartVersions[0].Version,
			appVersion:   chartWithVersionRepo1.ChartVersions[0].AppVersion,
			expectedCharts: []*chartmodels.Chart{
				&chartWithVersionRepo1,
			},
		},
		{
			name: "returns charts from different repos in the specific namespace",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: {
					repoName1:    {chartWithVersionRepo1},
					"other-repo": {chartWithVersionRepo2},
				},
			},
			namespace:    namespaceName,
			chartName:    chartWithVersionRepo1.Name,
			chartVersion: chartWithVersionRepo1.ChartVersions[0].Version,
			appVersion:   chartWithVersionRepo1.ChartVersions[0].AppVersion,
			expectedCharts: []*chartmodels.Chart{
				&chartWithVersionRepo1,
				&chartWithVersionRepo2,
			},
		},
		{
			name: "includes charts from global repositories",
			existingCharts: map[string]map[string][]chartmodels.Chart{
				namespaceName: {
					repoName1: {chartWithVersionRepo1},
				},
				dbutilstest.KubeappsTestNamespace: {
					"other-repo": {chartWithVersionRepo2},
				},
			},
			namespace:    namespaceName,
			chartName:    chartWithVersionRepo1.Name,
			chartVersion: chartWithVersionRepo1.ChartVersions[0].Version,
			appVersion:   chartWithVersionRepo1.ChartVersions[0].AppVersion,
			expectedCharts: []*chartmodels.Chart{
				&chartWithVersionRepo1,
				&chartWithVersionRepo2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pam, cleanup := getInitializedManager(t)
			defer cleanup()
			for namespace, chartsPerRepo := range tc.existingCharts {
				for repo, charts := range chartsPerRepo {
					pgtest.EnsureChartsExist(t, pam, charts, chartmodels.Repo{Name: repo, Namespace: namespace})
				}
			}

			charts, _, err := pam.GetPaginatedChartListWithFilters(ChartQuery{Namespace: tc.namespace, ChartName: tc.chartName, Version: tc.chartVersion, AppVersion: tc.appVersion}, 1, 0)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			sort.Sort(byID(charts))
			sort.Sort(byID(tc.expectedCharts))
			if got, want := charts, tc.expectedCharts; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})

	}
}
