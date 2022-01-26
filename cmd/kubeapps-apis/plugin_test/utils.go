// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package plugin_test

import (
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
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
)

var defaultInstalledPackageStatus = &pkgsGRPCv1alpha1.InstalledPackageStatus{
	Ready:      true,
	Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
	UserReason: "ReconciliationSucceeded",
}

func MakeAvailablePackageSummary(name string, plugin *pluginsGRPCv1alpha1.Plugin) *pkgsGRPCv1alpha1.AvailablePackageSummary {
	return &pkgsGRPCv1alpha1.AvailablePackageSummary{
		Name:        name,
		DisplayName: name,
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
		IconUrl:          DefaultIconURL,
		Categories:       []string{DefaultCategory},
		ShortDescription: DefaultDescription,
		AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
			Context:    &pkgsGRPCv1alpha1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: DefaultId,
			Plugin:     plugin,
		},
	}
}

func MakeInstalledPackageSummary(name string, plugin *pluginsGRPCv1alpha1.Plugin) *pkgsGRPCv1alpha1.InstalledPackageSummary {
	return &pkgsGRPCv1alpha1.InstalledPackageSummary{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Namespace: DefaultNamespace,
			},
			Identifier: name,
			Plugin:     plugin,
		},
		Name:    name,
		IconUrl: DefaultIconURL,
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: DefaultPkgVersion,
		},
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
		PkgDisplayName:   name,
		ShortDescription: DefaultDescription,
		Status:           defaultInstalledPackageStatus,
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: DefaultPkgVersion,
			AppVersion: DefaultAppVersion,
		},
	}
}

func MakeAvailablePackageDetail(name string, plugin *pluginsGRPCv1alpha1.Plugin) *pkgsGRPCv1alpha1.AvailablePackageDetail {
	return &pkgsGRPCv1alpha1.AvailablePackageDetail{
		AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
			Context:    &pkgsGRPCv1alpha1.Context{Cluster: GlobalPackagingCluster, Namespace: DefaultNamespace},
			Identifier: DefaultId,
			Plugin:     plugin,
		},
		Name: name,
		Version: &pkgsGRPCv1alpha1.PackageAppVersion{
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
		Maintainers:      []*pkgsGRPCv1alpha1.Maintainer{{Name: DefaultMaintainerName, Email: DefaultMaintainerEmail}},
	}
}

func MakeInstalledPackageDetail(name string, plugin *pluginsGRPCv1alpha1.Plugin) *pkgsGRPCv1alpha1.InstalledPackageDetail {
	return &pkgsGRPCv1alpha1.InstalledPackageDetail{
		InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Namespace: DefaultReleaseNamespace,
				Cluster:   GlobalPackagingCluster,
			},
			Identifier: DefaultReleaseName,
		},
		PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
			Version: DefaultReleaseVersion,
		},
		Name: DefaultReleaseName,
		CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: DefaultReleaseVersion,
			AppVersion: DefaultAppVersion,
		},
		LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
			PkgVersion: DefaultReleaseVersion,
			AppVersion: DefaultAppVersion,
		},
		ValuesApplied:         DefaultReleaseValues,
		PostInstallationNotes: DefaultReleaseNotes,
		Status:                defaultInstalledPackageStatus,
		AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
			Context: &pkgsGRPCv1alpha1.Context{
				Namespace: DefaultReleaseNamespace,
				Cluster:   GlobalPackagingCluster,
			},
			Identifier: DefaultId,
			Plugin:     plugin,
		},
		CustomDetail: nil,
	}
}

func MakePackageAppVersion(appVersion, pkgVersion string) *pkgsGRPCv1alpha1.PackageAppVersion {
	return &pkgsGRPCv1alpha1.PackageAppVersion{
		AppVersion: appVersion,
		PkgVersion: pkgVersion,
	}
}
