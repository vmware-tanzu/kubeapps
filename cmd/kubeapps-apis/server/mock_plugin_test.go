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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	globalPackagingCluster  = "default"
	defaultAppVersion       = "1.2.6"
	defaultCategory         = "cat-1"
	defaultIconURL          = "https://example.com/package.svg"
	defaultHomeURL          = "https://example.com/home"
	defaultRepoURL          = "https://example.com/repo"
	defaultDescription      = "description"
	defaultId               = "repo-1/package-id"
	defaultNamespace        = "my-namespace-1"
	defaultPkgVersion       = "1.0.0"
	defaultPkgUpdateVersion = "2.0.0"
	defaultReleaseName      = "my-release-1"
	defaultReleaseNamespace = "my-release-namespace-1"
	defaultReleaseNotes     = "some notes"
	defaultReleaseValues    = "{\"value\":\"new\"}"
	defaultReleaseVersion   = "1.2.3"
	defaultValuesSchema     = "\"$schema\": \"http://json-schema.org/schema#\""
	defaultReadme           = "#readme"
	defaultValues           = "key: value"
	defaultMaintainerName   = "me"
	defaultMaintainerEmail  = "me@example.com"
)

var defaultInstalledPackageStatus = &corev1.InstalledPackageStatus{
	Ready:      true,
	Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
	UserReason: "ReconciliationSucceeded",
}

type TestPackagingPlugin struct {
	packages.UnimplementedPackagesServiceServer
	plugin                    *plugins.Plugin
	availablePackageSummaries []*corev1.AvailablePackageSummary
	availablePackageDetail    *corev1.AvailablePackageDetail
	installedPackageSummaries []*corev1.InstalledPackageSummary
	installedPackageDetail    *corev1.InstalledPackageDetail
	packageAppVersions        []*corev1.PackageAppVersion
	categories                []string
	nextPageToken             string
	status                    codes.Code
}

func NewTestPackagingPlugin(plugin *plugins.Plugin) *TestPackagingPlugin {
	return &TestPackagingPlugin{
		plugin: plugin,
	}
}

func MakeTestPackagingPlugin(
	plugin *plugins.Plugin,
	availablePackageSummaries []*corev1.AvailablePackageSummary,
	availablePackageDetail *corev1.AvailablePackageDetail,
	installedPackageSummaries []*corev1.InstalledPackageSummary,
	installedPackageDetail *corev1.InstalledPackageDetail,
	packageAppVersions []*corev1.PackageAppVersion,
	nextPageToken string,
	categories []string,
	status codes.Code,
) *PkgsPluginWithServer {
	return &PkgsPluginWithServer{
		plugin: plugin,
		server: &TestPackagingPlugin{
			plugin:                    plugin,
			availablePackageSummaries: availablePackageSummaries,
			availablePackageDetail:    availablePackageDetail,
			installedPackageSummaries: installedPackageSummaries,
			installedPackageDetail:    installedPackageDetail,
			packageAppVersions:        packageAppVersions,
			categories:                categories,
			nextPageToken:             nextPageToken,
			status:                    status,
		},
	}
}

func MakeAvailablePackageSummary(name string, plugin *plugins.Plugin) *corev1.AvailablePackageSummary {
	return &corev1.AvailablePackageSummary{
		Name:        name,
		DisplayName: name,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		IconUrl:          defaultIconURL,
		Categories:       []string{defaultCategory},
		ShortDescription: defaultDescription,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: defaultNamespace},
			Identifier: defaultId,
			Plugin:     plugin,
		},
	}
}

func MakeInstalledPackageSummary(name string, plugin *plugins.Plugin) *corev1.InstalledPackageSummary {
	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: defaultNamespace,
			},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:    name,
		IconUrl: defaultIconURL,
		PkgVersionReference: &corev1.VersionReference{
			Version: defaultPkgVersion,
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		PkgDisplayName:   name,
		ShortDescription: defaultDescription,
		Status:           defaultInstalledPackageStatus,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
	}
}

func MakeAvailablePackageDetail(name string, plugin *plugins.Plugin) *corev1.AvailablePackageDetail {
	return &corev1.AvailablePackageDetail{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: defaultNamespace},
			Identifier: defaultId,
			Plugin:     plugin,
		},
		Name: name,
		Version: &corev1.PackageAppVersion{
			PkgVersion: defaultPkgVersion,
			AppVersion: defaultAppVersion,
		},
		RepoUrl:          defaultRepoURL,
		IconUrl:          defaultIconURL,
		HomeUrl:          defaultHomeURL,
		DisplayName:      name,
		Categories:       []string{defaultCategory},
		ShortDescription: defaultDescription,
		Readme:           defaultReadme,
		DefaultValues:    defaultValues,
		ValuesSchema:     defaultValuesSchema,
		SourceUrls:       []string{defaultHomeURL},
		Maintainers:      []*corev1.Maintainer{{Name: defaultMaintainerName, Email: defaultMaintainerEmail}},
	}
}

func MakeInstalledPackageDetail(name string, plugin *plugins.Plugin) *corev1.InstalledPackageDetail {
	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: defaultReleaseNamespace,
				Cluster:   globalPackagingCluster,
			},
			Identifier: defaultReleaseName,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: defaultReleaseVersion,
		},
		Name: defaultReleaseName,
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultReleaseVersion,
			AppVersion: defaultAppVersion,
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: defaultReleaseVersion,
			AppVersion: defaultAppVersion,
		},
		ValuesApplied:         defaultReleaseValues,
		PostInstallationNotes: defaultReleaseNotes,
		Status:                defaultInstalledPackageStatus,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Namespace: defaultReleaseNamespace,
				Cluster:   globalPackagingCluster,
			},
			Identifier: defaultId,
			Plugin:     plugin,
		},
		CustomDetail: nil,
	}
}

func MakePackageAppVersion(appVersion, pkgVersion string) *corev1.PackageAppVersion {
	return &corev1.PackageAppVersion{
		AppVersion: appVersion,
		PkgVersion: pkgVersion,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s TestPackagingPlugin) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: s.availablePackageSummaries,
		Categories:                s.categories,
		NextPageToken:             s.nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s TestPackagingPlugin) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: s.availablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s TestPackagingPlugin) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: s.installedPackageSummaries,
		NextPageToken:             s.nextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s TestPackagingPlugin) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: s.installedPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s TestPackagingPlugin) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	if s.status != codes.OK {
		return nil, status.Errorf(s.status, "Non-OK response")
	}
	return &packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: s.packageAppVersions,
	}, nil
}
