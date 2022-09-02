// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package pkgutils

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
)

const (
	DefaultAppVersion       = "1.2.6"
	DefaultChartDescription = "default chart description"
	DefaultChartIconURL     = "https://example.com/chart.svg"
	DefaultChartHomeURL     = "https://helm.sh/helm"
	DefaultChartCategory    = "cat1"
)

func TestPackageAppVersionsSummary(t *testing.T) {
	testCases := []struct {
		name                      string
		chart_versions            []models.ChartVersion
		version_summary           []*corev1.PackageAppVersion
		input_versions_in_summary VersionsInSummary
	}{
		{
			name: "it includes the latest three major versions only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "7.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "5.5.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "7.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three minor versions for each major version only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.6", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three patch versions for each minor version only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
		{
			name: "it includes the latest three patch versions of the latest three minor versions of the latest three major versions only",
			chart_versions: []models.ChartVersion{
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
			version_summary: []*corev1.PackageAppVersion{
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
			chart_versions: []models.ChartVersion{
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
			version_summary: []*corev1.PackageAppVersion{
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
			chart_versions: []models.ChartVersion{
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
			version_summary: []*corev1.PackageAppVersion{
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
		{
			name: "it includes latest versions ordered (descending) by package version",
			chart_versions: []models.ChartVersion{
				{Version: "6.1.3", AppVersion: DefaultAppVersion},
				{Version: "6.0.0", AppVersion: DefaultAppVersion},
				{Version: "6.0.3", AppVersion: DefaultAppVersion},
				{Version: "6.1.6", AppVersion: DefaultAppVersion},
				{Version: "5.2.1", AppVersion: DefaultAppVersion},
				{Version: "6.1.4", AppVersion: DefaultAppVersion},
				{Version: "6.1.5", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.PackageAppVersion{
				{PkgVersion: "6.1.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.1.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.1.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.0.3", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.0.0", AppVersion: DefaultAppVersion},
				{PkgVersion: "5.2.1", AppVersion: DefaultAppVersion},
			},
			input_versions_in_summary: GetDefaultVersionsInSummary(),
		},
	}

	opts := cmpopts.IgnoreUnexported(corev1.PackageAppVersion{})

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
		in       *models.Chart
		expected bool
	}{
		{
			name: "it returns true if the chart name, ID, repo and versions are specified",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: true,
		},
		{
			name: "it returns false if the chart name is missing",
			in: &models.Chart{
				ID: "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart ID is missing",
			in: &models.Chart{
				Name: "foo",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart repo is missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the ChartVersions are missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
			},
			expected: false,
		},
		{
			name: "it returns false if a ChartVersions.Version is missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []models.ChartVersion{
					{Version: "3.0.0"},
					{AppVersion: DefaultAppVersion},
				},
			},
			expected: false,
		},
		{
			name: "it returns true if the minimum (+maintainer) chart is correct",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []chart.Maintainer{{Name: "me"}},
			},
			expected: true,
		},
		{
			name: "it returns false if a Maintainer.Name is missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []chart.Maintainer{{Name: "me"}, {Email: "you"}},
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
	invalidChart := &models.Chart{Name: "foo"}

	testCases := []struct {
		name       string
		in         *models.Chart
		expected   *corev1.AvailablePackageSummary
		statusCode codes.Code
	}{
		{
			name: "it returns a complete AvailablePackageSummary for a complete chart",
			in: &models.Chart{
				Name:        "foo",
				ID:          "foo/bar",
				Category:    DefaultChartCategory,
				Description: "best chart",
				Icon:        "foo.bar/icon.svg",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				Maintainers: []chart.Maintainer{{Name: "me", Email: "me@me.me"}},
				ChartVersions: []models.ChartVersion{
					{Version: "3.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "2.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "1.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
				},
			},
			expected: &corev1.AvailablePackageSummary{
				Name:        "foo",
				DisplayName: "foo",
				LatestVersion: &corev1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				IconUrl:          "foo.bar/icon.svg",
				ShortDescription: "best chart",
				Categories:       []string{DefaultChartCategory},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it returns a valid AvailablePackageSummary if the minimal chart is correct",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version:    "3.0.0",
						AppVersion: DefaultAppVersion,
					},
				},
			},
			expected: &corev1.AvailablePackageSummary{
				Name:        "foo",
				DisplayName: "foo",
				LatestVersion: &corev1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Categories: []string{""},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns internal error if empty chart",
			in:         &models.Chart{},
			statusCode: codes.Internal,
		},
		{
			name:       "it returns internal error if chart is invalid",
			in:         invalidChart,
			statusCode: codes.Internal,
		},
	}

	pluginDetail := plugins.Plugin{
		Name:    "helm.packages",
		Version: "v1alpha1",
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			availablePackageSummary, err := AvailablePackageSummaryFromChart(tc.in, &pluginDetail)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
				if got, want := availablePackageSummary, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetUnescapedPackageID(t *testing.T) {
	testCases := []struct {
		name       string
		in         string
		out        string
		statusCode codes.Code
	}{
		{
			name:       "it returns a chartID for a valid input",
			in:         "foo/bar",
			out:        "foo/bar",
			statusCode: codes.OK,
		},
		{
			name:       "it returns a chartID for a valid input (2)",
			in:         "foo%2Fbar",
			out:        "foo/bar",
			statusCode: codes.OK,
		},
		{
			name:       "allows chart with multiple slashes",
			in:         "foo/bar/zot",
			out:        "foo/bar%2Fzot",
			statusCode: codes.OK,
		},
		{
			name:       "it fails for an invalid chartID",
			in:         "foo%ZZbar",
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "it fails for an invalid chartID (2)",
			in:         "foo",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOut, err := GetUnescapedPackageID(tc.in)
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := actualOut, tc.out; !cmp.Equal(got, want) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
			}
		})
	}
}

func TestSplitPackageIdentifier(t *testing.T) {
	testCases := []struct {
		name       string
		in         string
		repoName   string
		chartName  string
		statusCode codes.Code
	}{
		{
			name:       "it returns a repoName and chartName for a valid input",
			in:         "foo/bar",
			repoName:   "foo",
			chartName:  "bar",
			statusCode: codes.OK,
		},
		{
			name:       "it allows chart with multiple slashes",
			in:         "foo/bar/zot",
			repoName:   "foo",
			chartName:  "bar%2Fzot",
			statusCode: codes.OK,
		},
		{
			name:       "it fails for invalid input",
			in:         "foo",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, chartName, err := SplitPackageIdentifier(tc.in)
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
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

func TestDefaultValuesFromSchema(t *testing.T) {
	tests := []struct {
		name           string
		isCommentedOut bool
		schema         []byte
		expected       string
		expectedErr    error
	}{
		{"schema with defaults", false, []byte(`properties:
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`valueWithDefault: 80
`, nil},
		{"schema with without defaults (integer)", false, []byte(`properties:
  missingDefaultInteger:
    description: Missing default
    type: integer`,
		),
			`missingDefaultInteger: 0
`, nil},
		{"schema with without defaults (number)", false, []byte(`properties:
  missingDefaultNumber:
    description: Missing default
    type: number`,
		),
			`missingDefaultNumber: 0
`, nil},
		{"schema with without defaults (string)", false, []byte(`properties:
  missingDefaultString:
    description: Missing default
    type: string`,
		),
			`missingDefaultString: ""
`, nil},
		{"schema with without defaults (boolean)", false, []byte(`properties:
  missingDefaultBoolean:
    description: Missing default
    type: boolean`,
		),
			`missingDefaultBoolean: false
`, nil},
		{"schema with without defaults (array)", false, []byte(`properties:
  missingDefaultArray:
    description: Missing default
    type: array`,
		),
			`missingDefaultArray: []
`, nil},
		{"schema with without defaults (object)", false, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object`,
		),
			`missingDefaultObject: {}
`, nil},
		{"schema (mixed) with isCommentedOut=false", false, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`missingDefaultObject: {}
valueWithDefault: 80
`, nil},
		{"schema (mixed) with isCommentedOut=true", true, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`# missingDefaultObject: {}
# valueWithDefault: 80
`, nil},
		{"good schema (w/ additionalProperties: true, as per jsonschema draft 4)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties: true
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
		{"good schema (w/ additionalProperties: <schema>)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties:
      type: string
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
		{"bad schema (w/ additionalProperties: string)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties: string
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := DefaultValuesFromSchema(tt.schema, tt.isCommentedOut)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("unexpected error = %v", err)
			}
			if tt.expectedErr != nil {
				if want, got := tt.expectedErr.Error(), err.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
				}
			} else {
				if want, got := tt.expected, values; !cmp.Equal(want, got) {
					t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
				}
			}
		})
	}
}

// From https://github.com/kubernetes/apiextensions-apiserver/blob/release-1.21/pkg/apiserver/schema/defaulting/algorithm_test.go
// With a new test case ("object without default values")
func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		schema   *structuralschema.Structural
		expected string
	}{
		{"empty", "null", nil, "null"},
		{"scalar", "4", &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Default: structuralschema.JSON{Object: "foo"},
			},
		}, "4"},
		{"scalar array", "[1,2]", &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{Object: "foo"},
				},
			},
		}, "[1,2]"},
		{"object array", `[{"a":1},{"b":1},{"c":1}]`, &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "A"},
						},
					},
					"b": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "B"},
						},
					},
					"c": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "C"},
						},
					},
				},
			},
		}, `[{"a":1,"b":"B","c":"C"},{"a":"A","b":1,"c":"C"},{"a":"A","b":"B","c":1}]`},
		// New test case checking our tweaks
		{"object without default values", `{}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Type: "string",
					},
				},
				"b": {
					Generic: structuralschema.Generic{
						Type: "boolean",
					},
				},
				"c": {
					Generic: structuralschema.Generic{
						Type: "array",
					},
				},
				"d": {
					Generic: structuralschema.Generic{
						Type: "number",
					},
				},
				"e": {
					Generic: structuralschema.Generic{
						Type: "integer",
					},
				},
				"f": {
					Generic: structuralschema.Generic{
						Type: "object",
					},
				},
			},
		}, `{
  "a": "",
  "b": false,
  "c": [],
  "d": 0,
  "e": 0,
  "f": {}
}
`},
		{"object array object", `{"array":[{"a":1},{"b":2}],"object":{"a":1},"additionalProperties":{"x":{"a":1},"y":{"b":2}}}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"array": {
					Items: &structuralschema.Structural{
						Properties: map[string]structuralschema.Structural{
							"a": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{Object: "A"},
								},
							},
							"b": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{Object: "B"},
								},
							},
						},
					},
				},
				"object": {
					Properties: map[string]structuralschema.Structural{
						"a": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "N"},
							},
						},
						"b": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{Object: "O"},
							},
						},
					},
				},
				"additionalProperties": {
					Generic: structuralschema.Generic{
						AdditionalProperties: &structuralschema.StructuralOrBool{
							Structural: &structuralschema.Structural{
								Properties: map[string]structuralschema.Structural{
									"a": {
										Generic: structuralschema.Generic{
											Default: structuralschema.JSON{Object: "alpha"},
										},
									},
									"b": {
										Generic: structuralschema.Generic{
											Default: structuralschema.JSON{Object: "beta"},
										},
									},
								},
							},
						},
					},
				},
				"foo": {
					Generic: structuralschema.Generic{
						Default: structuralschema.JSON{Object: "bar"},
					},
				},
			},
		}, `{"array":[{"a":1,"b":"B"},{"a":"A","b":2}],"object":{"a":1,"b":"O"},"additionalProperties":{"x":{"a":1,"b":"beta"},"y":{"a":"alpha","b":2}},"foo":"bar"}`},
		{"empty and null", `[{},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":null},{"a":[]},{"a":{}}]`, &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}, `[{"a":"A"},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":"A"},{"a":[]},{"a":{}}]`},
		{"null in nullable list", `[null]`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Nullable: true,
			},
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}, `[null]`},
		{"null in non-nullable list", `[null]`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Nullable: false,
			},
			Items: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{Object: "A"},
				},
			},
		}, `["A"]`},
		{"null in nullable object", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{},
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Nullable: true,
						Default:  structuralschema.JSON{Object: "A"},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object", `{"a": null}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Nullable: false,
						Default:  structuralschema.JSON{Object: "A"},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null in nullable object with additionalProperties", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Structural: &structuralschema.Structural{
						Generic: structuralschema.Generic{
							Nullable: true,
							Default:  structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object with additionalProperties", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Structural: &structuralschema.Structural{
						Generic: structuralschema.Generic{
							Nullable: false,
							Default:  structuralschema.JSON{Object: "A"},
						},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null unknown field", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Bool: true,
				},
			},
		}, `{"a": null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in interface{}
			if err := json.Unmarshal([]byte(tt.json), &in); err != nil {
				t.Fatal(err)
			}

			var expected interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatal(err)
			}

			defaultValues(in, tt.schema)
			if !reflect.DeepEqual(in, expected) {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				err := enc.Encode(in)
				if err != nil {
					t.Fatalf("unexpected result mashalling error: %v", err)
				}
				if tt.expected != buf.String() {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.expected, buf.String()))
				}
			}
		})
	}
}
