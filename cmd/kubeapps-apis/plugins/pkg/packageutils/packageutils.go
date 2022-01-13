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
	"github.com/Masterminds/semver"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
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
