// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
)

const (
	GlobalPackagingCluster  = "default"
	DefaultAppVersion       = "1.2.6"
	DefaultCategory         = "cat-1"
	DefaultIconURL          = "https://example.com/package.svg"
	DefaultHomeURL          = "https://example.com/home"
	DefaultRepoURL          = "https://example.com/repo"
	DefaultDescription      = "description"
	DefaultId               = "repo-1/package-id"
	DefaultNamespace        = "my-namespace-1"
	DefaultPkgVersion       = "1.0.0"
	DefaultPkgUpdateVersion = "2.0.0"
	DefaultReleaseName      = "my-release-1"
	DefaultReleaseNamespace = "my-release-namespace-1"
	DefaultReleaseNotes     = "some notes"
	DefaultReleaseValues    = "{\"value\":\"new\"}"
	DefaultReleaseVersion   = "1.2.3"
	DefaultValuesSchema     = "\"$schema\": \"http://json-schema.org/schema#\""
	DefaultReadme           = "#readme"
	DefaultValues           = "key: value"
	DefaultMaintainerName   = "me"
	DefaultMaintainerEmail  = "me@example.com"
	DefaultRepoInterval     = "1m"
)

var defaultInstalledPackageStatus = &corev1.InstalledPackageStatus{
	Ready:      true,
	Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
	UserReason: "ReconciliationSucceeded",
}

var defaultRepoStatus = &corev1.PackageRepositoryStatus{
	Ready:      true,
	Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
	UserReason: "IndexationSucceed",
}

func MakeAvailablePackageSummary(name string, plugin *plugins.Plugin) *corev1.AvailablePackageSummary {
	return &corev1.AvailablePackageSummary{
		Name:        name,
		DisplayName: name,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
		IconUrl:          DefaultIconURL,
		Categories:       []string{DefaultCategory},
		ShortDescription: DefaultDescription,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: DefaultId,
			Plugin:     plugin,
		},
	}
}

func MakeInstalledPackageSummary(name string, plugin *plugins.Plugin) *corev1.InstalledPackageSummary {
	return &corev1.InstalledPackageSummary{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: DefaultNamespace,
			},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:    name,
		IconUrl: DefaultIconURL,
		PkgVersionReference: &corev1.VersionReference{
			Version: DefaultPkgVersion,
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
		PkgDisplayName:   name,
		ShortDescription: DefaultDescription,
		Status:           defaultInstalledPackageStatus,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
	}
}

func MakeAvailablePackageDetail(name string, plugin *plugins.Plugin) *corev1.AvailablePackageDetail {
	return &corev1.AvailablePackageDetail{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: DefaultId,
			Plugin:     plugin,
		},
		Name: name,
		Version: &corev1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
		RepoUrl:          DefaultRepoURL,
		IconUrl:          DefaultIconURL,
		HomeUrl:          DefaultHomeURL,
		DisplayName:      name,
		Categories:       []string{DefaultCategory},
		ShortDescription: DefaultDescription,
		Readme:           DefaultReadme,
		DefaultValues:    DefaultValues,
		ValuesSchema:     DefaultValuesSchema,
		SourceUrls:       []string{DefaultHomeURL},
		Maintainers:      []*corev1.Maintainer{{Name: DefaultMaintainerName, Email: DefaultMaintainerEmail}},
	}
}

func MakeInstalledPackageDetail(name string, plugin *plugins.Plugin) *corev1.InstalledPackageDetail {
	return &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: DefaultReleaseNamespace,
				Cluster:   GlobalPackagingCluster,
			},
			Identifier: DefaultReleaseName,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: DefaultReleaseVersion,
		},
		Name: DefaultReleaseName,
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: DefaultReleaseVersion,
			AppVersion: DefaultAppVersion,
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: DefaultReleaseVersion,
			AppVersion: DefaultAppVersion,
		},
		ValuesApplied:         DefaultReleaseValues,
		PostInstallationNotes: DefaultReleaseNotes,
		Status:                defaultInstalledPackageStatus,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Namespace: DefaultReleaseNamespace,
				Cluster:   GlobalPackagingCluster,
			},
			Identifier: DefaultId,
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

func MakePackageRepositoryDetail(name string, plugin *plugins.Plugin) *corev1.PackageRepositoryDetail {
	return &corev1.PackageRepositoryDetail{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context:    &corev1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:            name,
		Description:     DefaultDescription,
		NamespaceScoped: false,
		Type:            "helm",
		Url:             DefaultRepoURL,
		Interval:        DefaultRepoInterval,
		TlsConfig:       nil,
		Auth:            nil,
		CustomDetail:    nil,
		Status:          defaultRepoStatus,
	}
}

func MakePackageRepositorySummary(name string, plugin *plugins.Plugin) *corev1.PackageRepositorySummary {
	return &corev1.PackageRepositorySummary{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context:    &corev1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:            name,
		Description:     DefaultDescription,
		NamespaceScoped: false,
		Type:            "helm",
		Url:             DefaultRepoURL,
		Status:          defaultRepoStatus,
		RequiresAuth:    false,
	}
}
