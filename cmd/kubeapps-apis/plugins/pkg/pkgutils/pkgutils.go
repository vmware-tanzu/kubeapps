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
/*
 Utility functions that apply to "packages", e.g. helm charts or carvel packages
*/
package pkgutils

import (
	"net/url"
	"strings"

	"github.com/Masterminds/semver"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Contains miscellaneous package-utilities used by multiple plug-ins
const (
	MajorVersionsInSummary = 3
	MinorVersionsInSummary = 3
	PatchVersionsInSummary = 3
)

// Wrapper struct to include three version constants
type VersionsInSummary struct {
	Major int `json:"major"`
	Minor int `json:"minor"`
	Patch int `json:"patch"`
}

var (
	defaultVersionsInSummary = VersionsInSummary{
		Major: MajorVersionsInSummary,
		Minor: MinorVersionsInSummary,
		Patch: PatchVersionsInSummary,
	}
)

func GetDefaultVersionsInSummary() VersionsInSummary {
	return defaultVersionsInSummary
}

// packageAppVersionsSummary converts the model chart versions into the required version summary.
func PackageAppVersionsSummary(versions []models.ChartVersion, versionInSummary VersionsInSummary) []*corev1.PackageAppVersion {
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
			if len(version_map) >= versionInSummary.Major {
				continue
			}
		} else {
			// If we don't yet have this minor version
			if _, ok := version_map[version.Major()][version.Minor()]; !ok {
				// Don't add a new minor version if we already have enough for this major version
				if len(version_map[version.Major()]) >= versionInSummary.Minor {
					continue
				}
			} else {
				if len(version_map[version.Major()][version.Minor()]) >= versionInSummary.Patch {
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

// isValidChart returns true if the chart model passed defines a value
// for each required field described at the Helm website:
// https://helm.sh/docs/topics/charts/#the-chartyaml-file
// together with required fields for our model.
func IsValidChart(chart *models.Chart) (bool, error) {
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

// AvailablePackageSummaryFromChart builds an AvailablePackageSummary from a Chart
func AvailablePackageSummaryFromChart(chart *models.Chart, plugin *plugins.Plugin) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	isValid, err := IsValidChart(chart)
	if !isValid || err != nil {
		return nil, status.Errorf(codes.Internal, "invalid chart: %s", err.Error())
	}

	pkg.Name = chart.Name
	// Helm's Chart.yaml (and hence our model) does not include a separate
	// display name, so the chart name is also used here.
	pkg.DisplayName = chart.Name
	pkg.IconUrl = chart.Icon
	pkg.ShortDescription = chart.Description
	// TODO (gfichtenholt) I think when chart.Category is "" (i.e. missing)
	// we should return nil rather than []string{""}. For now keeping the logic
	// as is, so as not to break existing unit tests
	pkg.Categories = []string{chart.Category}

	pkg.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: chart.ID,
		Plugin:     plugin,
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

// TODO @gfichtenholt: I really wanted to put helm plugin's implementation of AvailablePackageDetailFromChart()
// here, and use it in flux plugin as well. But I found out a couple of flaws in the implementation and decided
// against it. Namely:
// 1. This assumption:
//    // We assume that chart.ChartVersions[0] will always contain either: the latest version or
//    // the specified version
//  This is a hack and forces the caller to prepate the input argument in a certain way such that
//  chart.ChartVersions[0] will be the desired version. I other words, the abstraction between the caller and
//  the call site is completely broken here. Yuck.
// 2. IMHO, it would have been better to get most of the detail info, including the current version out of
//   parsed chart YAML file, which this implementation chooses to ignore.
// I did consider using flux's implementation of AvailablePackageDetailFromChart but did not feel comfortable
// chaning helm plugin to use it before talking to @minelson
// Update Michael replied he is okay with my proposal:
// https://github.com/kubeapps/kubeapps/pull/4094#discussion_r790349962.
// Will come back to this

// GetUnescapedChartID takes a chart id with URI-encoded characters and decode them. Ex: 'foo%2Fbar' becomes 'foo/bar'
// also checks that the chart ID is in the expected format, namely "repoName/chartName"
func GetUnescapedChartID(chartID string) (string, error) {
	unescapedChartID, err := url.QueryUnescape(chartID)
	if err != nil {
		return "", status.Errorf(codes.InvalidArgument, "Unable to decode chart ID chart: %v", chartID)
	}
	// TODO(agamez): support ID with multiple slashes, eg: aaa/bbb/ccc
	chartIDParts := strings.Split(unescapedChartID, "/")
	if len(chartIDParts) != 2 {
		return "", status.Errorf(codes.InvalidArgument, "Incorrect package ref dentifier, currently just 'foo/bar' patterns are supported: %s", chartID)
	}
	return unescapedChartID, nil
}

func SplitChartIdentifier(chartID string) (repoName, chartName string, err error) {
	// getUnescapedChartID also ensures that there are two parts (ie. repo/chart-name only)
	unescapedChartID, err := GetUnescapedChartID(chartID)
	if err != nil {
		return "", "", err
	}
	chartIDParts := strings.Split(unescapedChartID, "/")
	return chartIDParts[0], chartIDParts[1], nil
}
