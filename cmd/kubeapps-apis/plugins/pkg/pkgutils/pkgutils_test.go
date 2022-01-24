/*
Copyright © 2022 VMware
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
package pkgutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
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

func TestGetUnescapedChartID(t *testing.T) {
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
			name:       "it fails for an invalid chartID",
			in:         "foo%ZZbar",
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "it fails for an invalid chartID (2)",
			in:         "foo/bar/zot",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualOut, err := GetUnescapedChartID(tc.in)
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

func TestSplitChartIdentifier(t *testing.T) {
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
			name:       "it fails for invalid input",
			in:         "foo/bar/zot",
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			repoName, chartName, err := SplitChartIdentifier(tc.in)
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
