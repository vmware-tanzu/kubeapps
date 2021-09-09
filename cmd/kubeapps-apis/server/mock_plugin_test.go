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
package server

import (
	"context"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
)

const (
	globalPackagingCluster  = "default"
	defaultNs               = "ns"
	defaultId               = "repo-1/chart-id"
	defaultPkgVersion       = "1.0.0"
	defaultAppVersion       = "1.2.6"
	defaultChartDescription = "default chart description"
	defaultChartIconURL     = "https://example.com/chart.svg"
	defaultChartCategory    = "cat1"
	releaseNamespace        = "my-namespace-1"
	releaseName             = "my-release-1"
	releaseVersion          = "1.2.3"
	releaseValues           = "{\"value\":\"new\"}"
	releaseNotes            = "some notes"
)

type TestPackagingPlugin struct {
	packages.UnimplementedPackagesServiceServer
	plugin *plugins.Plugin
}

func NewTestPackagingPlugin(plugin *plugins.Plugin) *TestPackagingPlugin {
	return &TestPackagingPlugin{
		plugin: plugin,
	}
}

func makeTestPackagingPlugin(pluginName string) *pkgsPluginWithServer {
	pluginDetails := &plugins.Plugin{
		Name:    pluginName,
		Version: "v1alpha1",
	}
	return &pkgsPluginWithServer{
		plugin: pluginDetails,
		server: &TestPackagingPlugin{
			plugin: pluginDetails,
		},
	}
}

func makeAvailablePackageSummary(name string, plugin *plugins.Plugin) *corev1.AvailablePackageSummary {
	return &corev1.AvailablePackageSummary{
		Name:        name,
		DisplayName: name,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		IconUrl:          defaultChartIconURL,
		Categories:       []string{defaultChartCategory},
		ShortDescription: defaultChartDescription,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: defaultNs},
			Identifier: defaultId,
			Plugin:     plugin,
		},
	}
}

func makeInstalledPackageSummary(name string, plugin *plugins.Plugin) *corev1.InstalledPackageSummary {
	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: defaultNs,
			},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:    name,
		IconUrl: defaultChartIconURL,
		PkgVersionReference: &corev1.VersionReference{
			Version: defaultPkgVersion,
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		PkgDisplayName:   name,
		ShortDescription: "short description",
		Status: &corev1.InstalledPackageStatus{
			Ready:      true,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
			UserReason: "ReconciliationSucceeded",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
	}
}

func makeAvailablePackageDetail(name string, plugin *plugins.Plugin) *corev1.AvailablePackageDetail {
	return &corev1.AvailablePackageDetail{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: defaultNs},
			Identifier: defaultId,
			Plugin:     plugin,
		},
		Name: name,
		Version: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		RepoUrl:          "https://example.com/repo",
		IconUrl:          defaultChartIconURL,
		HomeUrl:          "https://example.com/home",
		DisplayName:      name,
		Categories:       []string{"cat1"},
		ShortDescription: "short description",
		Readme:           "readme",
		DefaultValues:    "values",
		ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
		SourceUrls:       []string{"https://example.com/source", "https://example.com/home"},
		Maintainers: []*corev1.Maintainer{
			{
				Name:  "me",
				Email: "me@example.com",
			},
		},
	}
}

func makeInstalledPackageDetail(name string, plugin *plugins.Plugin) *corev1.InstalledPackageDetail {
	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: releaseNamespace,
				Cluster:   globalPackagingCluster,
			},
			Identifier: releaseName,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: releaseVersion,
		},
		Name: releaseName,
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: releaseVersion,
			AppVersion: defaultAppVersion,
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: releaseVersion,
			AppVersion: defaultAppVersion,
		},
		ValuesApplied:         releaseValues,
		PostInstallationNotes: releaseNotes,
		Status: &corev1.InstalledPackageStatus{
			Ready:      true,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
			UserReason: "deployed",
		},
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Namespace: releaseNamespace,
				Cluster:   globalPackagingCluster,
			},
			Identifier: "myrepo/" + name,
			Plugin:     plugin,
		},
		CustomDetail: nil,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPlugin) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
			makeAvailablePackageSummary("pkg-1", s.plugin),
			makeAvailablePackageSummary("pkg-2", s.plugin),
		},
		Categories:    []string{"cat1"},
		NextPageToken: "1",
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPlugin) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: makeAvailablePackageDetail("pkg-1", s.plugin),
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPlugin) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
			makeInstalledPackageSummary("pkg-1", s.plugin),
			makeInstalledPackageSummary("pkg-2", s.plugin),
		},
		NextPageToken: "1",
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPlugin) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: makeInstalledPackageDetail("pkg-1", s.plugin),
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPlugin) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	return &packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{
				PkgVersion: "3.0.0/" + s.plugin.Name,
				AppVersion: defaultAppVersion,
			},
			{
				PkgVersion: "2.0.0/" + s.plugin.Name,
				AppVersion: defaultAppVersion,
			},
			{
				PkgVersion: "1.0.0/" + s.plugin.Name,
				AppVersion: defaultAppVersion,
			},
		},
	}, nil
}
