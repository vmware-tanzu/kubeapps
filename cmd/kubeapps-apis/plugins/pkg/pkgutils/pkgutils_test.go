// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package pkgutils

import (
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmchart "helm.sh/helm/v3/pkg/chart"
)

const (
	DefaultAppVersion       = "1.2.6"
	DefaultChartDescription = "default chart description"
	DefaultChartIconURL     = "https://example.com/helmchart.svg"
	DefaultChartHomeURL     = "https://helm.sh/helm"
	DefaultChartCategory    = "cat1"
)

func TestPackageAppVersionsSummary(t *testing.T) {
	testCases := []struct {
		name                      string
		chart_versions            []chartmodels.ChartVersion
		version_summary           []*pkgsGRPCv1alpha1.PackageAppVersion
		input_versions_in_summary VersionsInSummary
	}{
		{
			name: "it includes the latest three major versions only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "7.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "5.5.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "7.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three minor versions for each major version only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.6", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three patch versions for each minor version only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three patch versions of the latest three minor versions of the latest three major versions only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.5", AppVersion: DefaultAppVersion},
				{Version: "8.4.4", AppVersion: DefaultAppVersion},
				{Version: "8.4.3", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.5", AppVersion: DefaultAppVersion},
				{Version: "8.3.4", AppVersion: DefaultAppVersion},
				{Version: "8.3.3", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.5", AppVersion: DefaultAppVersion},
				{Version: "8.2.4", AppVersion: DefaultAppVersion},
				{Version: "8.2.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.5", AppVersion: DefaultAppVersion},
				{Version: "6.5.4", AppVersion: DefaultAppVersion},
				{Version: "6.5.3", AppVersion: DefaultAppVersion},
				{Version: "6.4.6", AppVersion: DefaultAppVersion},
				{Version: "6.4.5", AppVersion: DefaultAppVersion},
				{Version: "6.4.4", AppVersion: DefaultAppVersion},
				{Version: "6.4.3", AppVersion: DefaultAppVersion},
				{Version: "6.3.6", AppVersion: DefaultAppVersion},
				{Version: "6.3.5", AppVersion: DefaultAppVersion},
				{Version: "6.3.4", AppVersion: DefaultAppVersion},
				{Version: "6.3.3", AppVersion: DefaultAppVersion},
				{Version: "6.2.6", AppVersion: DefaultAppVersion},
				{Version: "6.2.5", AppVersion: DefaultAppVersion},
				{Version: "6.2.4", AppVersion: DefaultAppVersion},
				{Version: "6.2.3", AppVersion: DefaultAppVersion},
				{Version: "4.5.6", AppVersion: DefaultAppVersion},
				{Version: "4.5.5", AppVersion: DefaultAppVersion},
				{Version: "4.5.4", AppVersion: DefaultAppVersion},
				{Version: "4.5.3", AppVersion: DefaultAppVersion},
				{Version: "4.4.6", AppVersion: DefaultAppVersion},
				{Version: "4.4.5", AppVersion: DefaultAppVersion},
				{Version: "4.4.4", AppVersion: DefaultAppVersion},
				{Version: "4.4.3", AppVersion: DefaultAppVersion},
				{Version: "4.3.6", AppVersion: DefaultAppVersion},
				{Version: "4.3.5", AppVersion: DefaultAppVersion},
				{Version: "4.3.4", AppVersion: DefaultAppVersion},
				{Version: "4.3.3", AppVersion: DefaultAppVersion},
				{Version: "4.2.6", AppVersion: DefaultAppVersion},
				{Version: "4.2.5", AppVersion: DefaultAppVersion},
				{Version: "4.2.4", AppVersion: DefaultAppVersion},
				{Version: "4.2.3", AppVersion: DefaultAppVersion},
				{Version: "2.5.6", AppVersion: DefaultAppVersion},
				{Version: "2.5.5", AppVersion: DefaultAppVersion},
				{Version: "2.5.4", AppVersion: DefaultAppVersion},
				{Version: "2.5.3", AppVersion: DefaultAppVersion},
				{Version: "2.4.6", AppVersion: DefaultAppVersion},
				{Version: "2.4.5", AppVersion: DefaultAppVersion},
				{Version: "2.4.4", AppVersion: DefaultAppVersion},
				{Version: "2.4.3", AppVersion: DefaultAppVersion},
				{Version: "2.3.6", AppVersion: DefaultAppVersion},
				{Version: "2.3.5", AppVersion: DefaultAppVersion},
				{Version: "2.3.4", AppVersion: DefaultAppVersion},
				{Version: "2.3.3", AppVersion: DefaultAppVersion},
				{Version: "2.2.6", AppVersion: DefaultAppVersion},
				{Version: "2.2.5", AppVersion: DefaultAppVersion},
				{Version: "2.2.4", AppVersion: DefaultAppVersion},
				{Version: "2.2.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.4", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest four patch versions of the latest one minor versions of the latest two major versions only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.5", AppVersion: DefaultAppVersion},
				{Version: "8.4.4", AppVersion: DefaultAppVersion},
				{Version: "8.4.3", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.5", AppVersion: DefaultAppVersion},
				{Version: "8.3.4", AppVersion: DefaultAppVersion},
				{Version: "8.3.3", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.5", AppVersion: DefaultAppVersion},
				{Version: "8.2.4", AppVersion: DefaultAppVersion},
				{Version: "8.2.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.5", AppVersion: DefaultAppVersion},
				{Version: "6.5.4", AppVersion: DefaultAppVersion},
				{Version: "6.5.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.2", AppVersion: DefaultAppVersion},
				{Version: "6.5.1", AppVersion: DefaultAppVersion},
				{Version: "6.4.6", AppVersion: DefaultAppVersion},
				{Version: "6.4.5", AppVersion: DefaultAppVersion},
				{Version: "6.4.4", AppVersion: DefaultAppVersion},
				{Version: "6.4.3", AppVersion: DefaultAppVersion},
				{Version: "6.3.6", AppVersion: DefaultAppVersion},
				{Version: "6.3.5", AppVersion: DefaultAppVersion},
				{Version: "6.3.4", AppVersion: DefaultAppVersion},
				{Version: "6.3.3", AppVersion: DefaultAppVersion},
				{Version: "6.2.6", AppVersion: DefaultAppVersion},
				{Version: "6.2.5", AppVersion: DefaultAppVersion},
				{Version: "6.2.4", AppVersion: DefaultAppVersion},
				{Version: "6.2.3", AppVersion: DefaultAppVersion},
				{Version: "4.5.6", AppVersion: DefaultAppVersion},
				{Version: "4.5.5", AppVersion: DefaultAppVersion},
				{Version: "4.5.4", AppVersion: DefaultAppVersion},
				{Version: "4.5.3", AppVersion: DefaultAppVersion},
				{Version: "4.4.6", AppVersion: DefaultAppVersion},
				{Version: "4.4.5", AppVersion: DefaultAppVersion},
				{Version: "4.4.4", AppVersion: DefaultAppVersion},
				{Version: "4.4.3", AppVersion: DefaultAppVersion},
				{Version: "4.3.6", AppVersion: DefaultAppVersion},
				{Version: "4.3.5", AppVersion: DefaultAppVersion},
				{Version: "4.3.4", AppVersion: DefaultAppVersion},
				{Version: "4.3.3", AppVersion: DefaultAppVersion},
				{Version: "4.2.6", AppVersion: DefaultAppVersion},
				{Version: "4.2.5", AppVersion: DefaultAppVersion},
				{Version: "4.2.4", AppVersion: DefaultAppVersion},
				{Version: "4.2.3", AppVersion: DefaultAppVersion},
				{Version: "2.5.6", AppVersion: DefaultAppVersion},
				{Version: "2.5.5", AppVersion: DefaultAppVersion},
				{Version: "2.5.4", AppVersion: DefaultAppVersion},
				{Version: "2.5.3", AppVersion: DefaultAppVersion},
				{Version: "2.4.6", AppVersion: DefaultAppVersion},
				{Version: "2.4.5", AppVersion: DefaultAppVersion},
				{Version: "2.4.4", AppVersion: DefaultAppVersion},
				{Version: "2.4.3", AppVersion: DefaultAppVersion},
				{Version: "2.3.6", AppVersion: DefaultAppVersion},
				{Version: "2.3.5", AppVersion: DefaultAppVersion},
				{Version: "2.3.4", AppVersion: DefaultAppVersion},
				{Version: "2.3.3", AppVersion: DefaultAppVersion},
				{Version: "2.2.6", AppVersion: DefaultAppVersion},
				{Version: "2.2.5", AppVersion: DefaultAppVersion},
				{Version: "2.2.4", AppVersion: DefaultAppVersion},
				{Version: "2.2.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.3", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.3", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: VersionsInSummary{
				Major: 2,
				Minor: 1,
				Patch: 4},
		},
		{
			name: "it includes the latest zero patch versions of the latest zero minor versions of the latest six major versions only",
			chart_versions: []chartmodels.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.5", AppVersion: DefaultAppVersion},
				{Version: "8.4.4", AppVersion: DefaultAppVersion},
				{Version: "8.4.3", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.5", AppVersion: DefaultAppVersion},
				{Version: "8.3.4", AppVersion: DefaultAppVersion},
				{Version: "8.3.3", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.5", AppVersion: DefaultAppVersion},
				{Version: "8.2.4", AppVersion: DefaultAppVersion},
				{Version: "8.2.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.5", AppVersion: DefaultAppVersion},
				{Version: "6.5.4", AppVersion: DefaultAppVersion},
				{Version: "6.5.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.2", AppVersion: DefaultAppVersion},
				{Version: "6.5.1", AppVersion: DefaultAppVersion},
				{Version: "6.4.6", AppVersion: DefaultAppVersion},
				{Version: "6.4.5", AppVersion: DefaultAppVersion},
				{Version: "6.4.4", AppVersion: DefaultAppVersion},
				{Version: "6.4.3", AppVersion: DefaultAppVersion},
				{Version: "6.3.6", AppVersion: DefaultAppVersion},
				{Version: "6.3.5", AppVersion: DefaultAppVersion},
				{Version: "6.3.4", AppVersion: DefaultAppVersion},
				{Version: "6.3.3", AppVersion: DefaultAppVersion},
				{Version: "6.2.6", AppVersion: DefaultAppVersion},
				{Version: "6.2.5", AppVersion: DefaultAppVersion},
				{Version: "6.2.4", AppVersion: DefaultAppVersion},
				{Version: "6.2.3", AppVersion: DefaultAppVersion},
				{Version: "4.5.6", AppVersion: DefaultAppVersion},
				{Version: "4.5.5", AppVersion: DefaultAppVersion},
				{Version: "4.5.4", AppVersion: DefaultAppVersion},
				{Version: "4.5.3", AppVersion: DefaultAppVersion},
				{Version: "4.4.6", AppVersion: DefaultAppVersion},
				{Version: "4.4.5", AppVersion: DefaultAppVersion},
				{Version: "4.4.4", AppVersion: DefaultAppVersion},
				{Version: "4.4.3", AppVersion: DefaultAppVersion},
				{Version: "4.3.6", AppVersion: DefaultAppVersion},
				{Version: "4.3.5", AppVersion: DefaultAppVersion},
				{Version: "4.3.4", AppVersion: DefaultAppVersion},
				{Version: "4.3.3", AppVersion: DefaultAppVersion},
				{Version: "4.2.6", AppVersion: DefaultAppVersion},
				{Version: "4.2.5", AppVersion: DefaultAppVersion},
				{Version: "4.2.4", AppVersion: DefaultAppVersion},
				{Version: "4.2.3", AppVersion: DefaultAppVersion},
				{Version: "3.4.6", AppVersion: DefaultAppVersion},
				{Version: "3.4.5", AppVersion: DefaultAppVersion},
				{Version: "3.4.4", AppVersion: DefaultAppVersion},
				{Version: "2.4.3", AppVersion: DefaultAppVersion},
				{Version: "2.3.6", AppVersion: DefaultAppVersion},
				{Version: "2.3.5", AppVersion: DefaultAppVersion},
				{Version: "2.3.4", AppVersion: DefaultAppVersion},
				{Version: "2.3.3", AppVersion: DefaultAppVersion},
				{Version: "1.2.6", AppVersion: DefaultAppVersion},
				{Version: "1.2.5", AppVersion: DefaultAppVersion},
				{Version: "1.2.4", AppVersion: DefaultAppVersion},
				{Version: "1.2.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*pkgsGRPCv1alpha1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "3.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "2.4.3", AppVersion: DefaultAppVersion},
				{PkgVersion: "1.2.6", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: VersionsInSummary{Major: 6,
				Minor: 0,
				Patch: 0},
		},
	}

	opts := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.PackageAppVersion{})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := PackageAppVersionsSummary(tc.chart_versions, tc.input_versions_in_summary), tc.version_summary; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestIsValidChart(t *testing.T) {
	testCases := []struct {
		name     string
		in       *chartmodels.Chart
		expected bool
	}{
		{
			name: "it returns true if the chart name, ID, repo and versions are specified",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: true,
		},
		{
			name: "it returns false if the chart name is missing",
			in: &chartmodels.Chart{
				ID: "foo/bar",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart ID is missing",
			in: &chartmodels.Chart{
				Name: "foo",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart repo is missing",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the ChartVersions are missing",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
			},
			expected: false,
		},
		{
			name: "it returns false if a ChartVersions.Version is missing",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []chartmodels.ChartVersion{
					{Version: "3.0.0"},
					{AppVersion: DefaultAppVersion},
				},
			},
			expected: false,
		},
		{
			name: "it returns true if the minimum (+maintainer) chart is correct",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []helmchart.Maintainer{{Name: "me"}},
			},
			expected: true,
		},
		{
			name: "it returns false if a Maintainer.Name is missing",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []helmchart.Maintainer{{Name: "me"}, {Email: "you"}},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IsValidChart(tc.in)
			if got, want := res, tc.expected; got != want {
				t.Fatalf("got: %+v, want: %+v, res: %+v (%+v)", got, want, res, err)
			}
		})
	}
}

func TestAvailablePackageSummaryFromChart(t *testing.T) {
	invalidChart := &chartmodels.Chart{Name: "foo"}

	testCases := []struct {
		name       string
		in         *chartmodels.Chart
		expected   *pkgsGRPCv1alpha1.AvailablePackageSummary
		statusCode grpccodes.Code
	}{
		{
			name: "it returns a complete AvailablePackageSummary for a complete chart",
			in: &chartmodels.Chart{
				Name:        "foo",
				ID:          "foo/bar",
				Category:    DefaultChartCategory,
				Description: "best chart",
				Icon:        "foo.bar/icon.svg",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				Maintainers: []helmchart.Maintainer{{Name: "me", Email: "me@me.me"}},
				ChartVersions: []chartmodels.ChartVersion{
					{Version: "3.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "2.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "1.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
				},
			},
			expected: &pkgsGRPCv1alpha1.AvailablePackageSummary{
				Name:        "foo",
				DisplayName: "foo",
				LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				IconUrl:          "foo.bar/icon.svg",
				ShortDescription: "best chart",
				Categories:       []string{DefaultChartCategory},
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it returns a valid AvailablePackageSummary if the minimal chart is correct",
			in: &chartmodels.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &chartmodels.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []chartmodels.ChartVersion{
					{
						Version:    "3.0.0",
						AppVersion: DefaultAppVersion,
					},
				},
			},
			expected: &pkgsGRPCv1alpha1.AvailablePackageSummary{
				Name:        "foo",
				DisplayName: "foo",
				LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Categories: []string{""},
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns internal error if empty chart",
			in:         &chartmodels.Chart{},
			statusCode: grpccodes.Internal,
		},
		{
			name:       "it returns internal error if chart is invalid",
			in:         invalidChart,
			statusCode: grpccodes.Internal,
		},
	}

	pluginDetail := pluginsGRPCv1alpha1.Plugin{
		Name:    "helm.packages",
		Version: "v1alpha1",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			availablePackageSummary, err := AvailablePackageSummaryFromChart(tc.in, &pluginDetail)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				opt1 := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.AvailablePackageDetail{}, pkgsGRPCv1alpha1.AvailablePackageSummary{}, pkgsGRPCv1alpha1.AvailablePackageReference{}, pkgsGRPCv1alpha1.Context{}, pluginsGRPCv1alpha1.Plugin{}, pkgsGRPCv1alpha1.Maintainer{}, pkgsGRPCv1alpha1.PackageAppVersion{})
				if got, want := availablePackageSummary, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetUnescapedChartID(t *testing.T) {
	testCases := []struct {
		name       string
		in         string
		out        string
		statusCode grpccodes.Code
	}{
		{
			name:       "it returns a chartID for a valid input",
			in:         "foo/bar",
			out:        "foo/bar",
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns a chartID for a valid input (2)",
			in:         "foo%2Fbar",
			out:        "foo/bar",
			statusCode: grpccodes.OK,
		},
		{
			name:       "it fails for an invalid chartID",
			in:         "foo%ZZbar",
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name:       "it fails for an invalid chartID (2)",
			in:         "foo/bar/zot",
			statusCode: grpccodes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOut, err := GetUnescapedChartID(tc.in)
			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				if got, want := actualOut, tc.out; !cmp.Equal(got, want) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
			}
		})
	}
}

func TestSplitChartIdentifier(t *testing.T) {
	testCases := []struct {
		name       string
		in         string
		repoName   string
		chartName  string
		statusCode grpccodes.Code
	}{
		{
			name:       "it returns a repoName and chartName for a valid input",
			in:         "foo/bar",
			repoName:   "foo",
			chartName:  "bar",
			statusCode: grpccodes.OK,
		},
		{
			name:       "it fails for invalid input",
			in:         "foo/bar/zot",
			statusCode: grpccodes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, chartName, err := SplitChartIdentifier(tc.in)
			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				if got, want := repoName, tc.repoName; !cmp.Equal(got, want) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
				if got, want := chartName, tc.chartName; !cmp.Equal(got, want) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
			}
		})
	}
}
