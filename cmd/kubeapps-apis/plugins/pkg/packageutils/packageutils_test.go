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
package packageutils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"helm.sh/helm/v3/pkg/chart"
)

const (
	DefaultAppVersion = "1.2.6"
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
