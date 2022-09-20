// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"os"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"google.golang.org/protobuf/types/known/anypb"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// global vars
// why define these here? see https://github.com/vmware-tanzu/kubeapps/pull/3736#discussion_r745246398
// plus I am putting them in a separate file, since they take up so much space they distract from
// overall test logic
var (
	create_installed_package_request_basic = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-1/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext:       targetContext("test-1"),
	}

	// specify just the fields that cannot be easily computed based on the request
	expected_detail_installed_package_basic = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo"),
	}

	expected_resource_refs_basic = podinfo_installed_refs("my-podinfo")

	create_installed_package_request_semver_constraint = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-2/podinfo", "default"),
		Name:                "my-podinfo-2",
		TargetContext:       targetContext("test-2"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
	}

	expected_detail_installed_package_semver_constraint = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-2"),
	}

	expected_resource_refs_semver_constraint = podinfo_installed_refs("my-podinfo-2")

	create_installed_package_request_reconcile_options = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-3/podinfo", "default"),
		Name:                "my-podinfo-3",
		TargetContext:       targetContext("test-3"),
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           "1m",
			Suspend:            false,
			ServiceAccountName: "foo",
		},
	}

	expected_detail_installed_package_reconcile_options = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: pkgAppVersion("6.0.0"),
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           "1m",
			Suspend:            false,
			ServiceAccountName: "foo",
		},
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-3"),
	}

	expected_resource_refs_reconcile_options = podinfo_installed_refs("my-podinfo-3")

	create_installed_package_request_with_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-4/podinfo", "default"),
		Name:                "my-podinfo-4",
		TargetContext:       targetContext("test-4"),
		Values:              "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	expected_detail_installed_package_with_values = &corev1.InstalledPackageDetail{
		CurrentVersion: pkgAppVersion("6.0.0"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-4"),
		ValuesApplied:         "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_resource_refs_with_values = podinfo_installed_refs("my-podinfo-4")

	create_installed_package_request_install_fails = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-5/podinfo", "default"),
		Name:                "my-podinfo-5",
		TargetContext:       targetContext("test-5"),
		Values:              "{\"replicaCount\": \"what we do in the shadows\" }",
	}

	expected_detail_installed_package_install_fails = &corev1.InstalledPackageDetail{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:  false,
			Reason: corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			// most of the time it fails with
			//   "InstallFailed: install retries exhausted",
			// but every once in a while you get
			//   "InstallFailed: Helm install failed: unable to build kubernetes objects from release manifest: error
			//    validating "": error validating data: ValidationError(Deployment.spec.replicas): invalid type for
			//    io.k8s.api.apps.v1.DeploymentSpec.replicas: got "string""
			// so we'll just test the prefix
			UserReason: "InstallFailed: ",
		},
		ValuesApplied: "{\"replicaCount\":\"what we do in the shadows\"}",
	}

	create_installed_package_request_podinfo_5_2_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
		Name:                "my-podinfo-6",
		TargetContext:       targetContext("test-1"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_installed_package_podinfo_5_2_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-6"),
	}

	expected_resource_refs_podinfo_5_2_1 = podinfo_installed_refs("my-podinfo-6")

	expected_detail_installed_package_podinfo_6_0_0 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-6"),
	}

	create_installed_package_request_podinfo_5_2_1_no_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
		Name:                "my-podinfo-7",
		TargetContext:       targetContext("test-7"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_installed_package_podinfo_5_2_1_no_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-7"),
	}

	expected_resource_refs_podinfo_5_2_1_no_values = podinfo_installed_refs("my-podinfo-7")

	expected_detail_installed_package_podinfo_5_2_1_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		ValuesApplied:         "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-7"),
	}

	create_installed_package_request_podinfo_5_2_1_values_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
		Name:                "my-podinfo-8",
		TargetContext:       targetContext("test-8"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_installed_package_podinfo_5_2_1_values_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		ValuesApplied:         "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-8"),
	}

	expected_resource_refs_podinfo_5_2_1_values_2 = podinfo_installed_refs("my-podinfo-8")

	expected_detail_installed_package_podinfo_5_2_1_values_3 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		ValuesApplied:         "{\"ui\":{\"message\":\"Le Bureau des Légendes\"}}",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-8"),
	}

	create_installed_package_request_podinfo_5_2_1_values_4 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
		Name:                "my-podinfo-9",
		TargetContext:       targetContext("test-9"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_installed_package_podinfo_5_2_1_values_4 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		ValuesApplied:         "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-9"),
	}

	expected_resource_refs_podinfo_5_2_1_values_4 = podinfo_installed_refs("my-podinfo-9")

	expected_detail_installed_package_podinfo_5_2_1_values_5 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-9"),
	}

	create_installed_package_request_podinfo_5_2_1_values_6 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-10/podinfo", "default"),
		Name:                "my-podinfo-10",
		TargetContext:       targetContext("test-10"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_installed_package_podinfo_5_2_1_values_6 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		ValuesApplied:         "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-10"),
	}

	expected_resource_refs_podinfo_5_2_1_values_6 = podinfo_installed_refs("my-podinfo-10")

	create_installed_package_request_podinfo_7 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-11/podinfo", "default"),
		Name:                "my-podinfo-11",
		TargetContext:       targetContext("test-11"),
	}

	create_installed_package_request_podinfo_8 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-12/podinfo", "default"),
		Name:                "my-podinfo-12",
		TargetContext:       targetContext("test-12"),
		Values:              "{\"replicaCount\": \"what we do in the shadows\" }",
	}

	expected_detail_installed_package_podinfo_7 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-11"),
	}

	expected_detail_installed_package_podinfo_8 = &corev1.InstalledPackageDetail{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:  false,
			Reason: corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			// most of the time it fails with
			//   "InstallFailed: install retries exhausted",
			// but every once in a while you get
			//   "InstallFailed: Helm install failed: unable to build kubernetes objects from release manifest: error
			//    validating "": error validating data: ValidationError(Deployment.spec.replicas): invalid type for
			//    io.k8s.api.apps.v1.DeploymentSpec.replicas: got "string""
			// so we'll just test the prefix
			UserReason: "InstallFailed: ",
		},
		ValuesApplied: "{\"replicaCount\":\"what we do in the shadows\"}",
	}

	expected_detail_installed_package_podinfo_9 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		ValuesApplied:         "{\n  \"replicaCount\": 1\n}\n",
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-12"),
	}

	expected_resource_refs_podinfo_7 = podinfo_installed_refs("my-podinfo-11")

	expected_resource_refs_podinfo_9 = podinfo_installed_refs("my-podinfo-12")

	expected_detail_installed_package_oci = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.1.5"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-17"),
	}

	expected_resource_refs_oci = podinfo_installed_refs("my-podinfo-17")

	expected_detail_installed_package_oci_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.1.8"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-19"),
	}

	expected_resource_refs_oci_2 = podinfo_installed_refs("my-podinfo-19")

	update_request_1 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
	}

	update_request_2 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	update_request_3 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"Le Bureau des Légendes\" } }",
	}

	update_request_4 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "",
	}

	update_request_5 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	update_request_6 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	update_request_7 = &corev1.UpdateInstalledPackageRequest{
		// InstalledPackageRef will be filled in by the code below after a call to create(...) completes
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		Values: "{\"replicaCount\": 1 }",
	}

	create_installed_package_request_podinfo_for_delete_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-12/podinfo", "default"),
		Name:                "my-podinfo-12",
		TargetContext:       targetContext("test-12"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_installed_package_podinfo_for_delete_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-12"),
	}

	expected_resource_refs_for_delete_1 = podinfo_installed_refs("my-podinfo-12")

	create_installed_package_request_podinfo_for_delete_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-13/podinfo", "default"),
		Name:                "my-podinfo-13",
		TargetContext:       targetContext("test-13"),
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_installed_package_podinfo_for_delete_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion:        pkgAppVersion("5.2.1"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo-13"),
	}

	expected_resource_refs_for_delete_2 = podinfo_installed_refs("my-podinfo-13")

	create_installed_package_request_wrong_cluster = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-14/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-14",
			Cluster:   "this is not the cluster you're looking for",
		},
	}

	create_installed_package_request_target_ns_doesnt_exist = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-15/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext:       targetContext("test-15"),
	}

	create_installed_package_request_auto_update = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-16/podinfo", "default"),
		Name:                "my-podinfo-16",
		TargetContext:       targetContext("test-16"),
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
	}

	expected_detail_installed_package_auto_update = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: pkgAppVersion("6.0.0"),
		Status:         status_installed,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
		PostInstallationNotes: podinfo_notes("my-podinfo-16"),
	}

	expected_detail_installed_package_auto_update_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: pkgAppVersion("6.0.3"),
		Name:           "my-podinfo-16",
		Status:         status_installed,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			},
			Identifier: "podinfo-16/podinfo",
			Plugin:     fluxPlugin,
		},
		PostInstallationNotes: podinfo_notes("my-podinfo-16"),
	}

	expected_resource_refs_auto_update = podinfo_installed_refs("my-podinfo-16")

	create_installed_package_request_oci = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-17/podinfo", "default"),
		Name:                "my-podinfo-17",
		TargetContext:       targetContext("test-17"),
	}

	create_installed_package_request_auto_update_oci = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-18/podinfo", "default"),
		Name:                "my-podinfo-18",
		TargetContext:       targetContext("test-18"),
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
	}

	expected_detail_installed_package_auto_update_oci = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: pkgAppVersion("6.1.5"),
		Status:         status_installed,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
		PostInstallationNotes: podinfo_notes("my-podinfo-18"),
	}

	expected_resource_refs_auto_update_oci = podinfo_installed_refs("my-podinfo-18")

	expected_detail_installed_package_auto_update_oci_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: pkgAppVersion("6.1.6"),
		Name:           "my-podinfo-18",
		Status:         status_installed,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "30s",
		},
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			},
			Identifier: "podinfo-18/podinfo",
			Plugin:     fluxPlugin,
		},
		PostInstallationNotes: podinfo_notes("my-podinfo-18"),
	}

	create_installed_package_request_oci_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-19/podinfo", "default"),
		Name:                "my-podinfo-19",
		TargetContext:       targetContext("test-19"),
	}

	expected_detail_test_release_rbac = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo"),
	}

	expected_summaries_test_release_rbac_1 = &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
			{
				InstalledPackageRef: installedRef("my-podinfo", "@TARGET_NS@"),
				Name:                "my-podinfo",
				PkgVersionReference: &corev1.VersionReference{
					Version: "*",
				},
				Status: &corev1.InstalledPackageStatus{
					Ready:      true,
					Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
				},
				// notice that the details from the corresponding chart, like LatestVersion, are missing
			},
		},
	}

	expected_summaries_test_release_rbac_2 = &corev1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
			{
				InstalledPackageRef: installedRef("my-podinfo", "@TARGET_NS@"),
				Name:                "my-podinfo",
				PkgVersionReference: &corev1.VersionReference{
					Version: "*",
				},
				Status: &corev1.InstalledPackageStatus{
					Ready:      true,
					Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
				},
				// notice that the details from the corresponding chart, like LatestVersion, are present
				CurrentVersion:   pkgAppVersion("6.0.0"),
				PkgDisplayName:   "podinfo",
				ShortDescription: "Podinfo Helm chart for Kubernetes",
				LatestVersion:    pkgAppVersion("6.0.0"),
			},
		},
	}

	expected_detail_test_release_rbac_2 = &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: &corev1.InstalledPackageDetail{
			InstalledPackageRef: installedRef("my-podinfo", "@TARGET_NS@"),
			PkgVersionReference: &corev1.VersionReference{Version: "*"},
			Name:                "my-podinfo",
			CurrentVersion:      pkgAppVersion("6.0.0"),
			ReconciliationOptions: &corev1.ReconciliationOptions{
				Interval: "1m",
			},
			Status:                status_installed,
			PostInstallationNotes: podinfo_notes("my-podinfo"),
			AvailablePackageRef:   availableRef("podinfo-1/podinfo", "@SOURCE_NS@"),
		},
	}

	expected_detail_test_release_rbac_3 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo"),
	}

	expected_detail_test_release_rbac_4 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion:        pkgAppVersion("6.0.0"),
		Status:                status_installed,
		PostInstallationNotes: podinfo_notes("my-podinfo"),
	}

	available_package_summaries_podinfo_basic_auth = func(name string) *corev1.GetAvailablePackageSummariesResponse {
		return &corev1.GetAvailablePackageSummariesResponse{
			AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
				{
					Name:                "podinfo",
					AvailablePackageRef: availableRef(name+"/podinfo", "default"),
					LatestVersion:       pkgAppVersion("6.0.0"),
					DisplayName:         "podinfo",
					ShortDescription:    "Podinfo Helm chart for Kubernetes",
					Categories:          []string{""},
				},
			},
		}
	}

	expected_detail_podinfo_basic_auth = func(name string) *corev1.GetAvailablePackageDetailResponse {
		return &corev1.GetAvailablePackageDetailResponse{
			AvailablePackageDetail: &corev1.AvailablePackageDetail{
				AvailablePackageRef: availableRef(name+"/podinfo", "default"),
				Name:                "podinfo",
				Version:             pkgAppVersion("6.0.0"),
				RepoUrl:             "http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth",
				HomeUrl:             "https://github.com/stefanprodan/podinfo",
				DisplayName:         "podinfo",
				ShortDescription:    "Podinfo Helm chart for Kubernetes",
				SourceUrls:          []string{"https://github.com/stefanprodan/podinfo"},
				Maintainers: []*corev1.Maintainer{
					{Name: "stefanprodan", Email: "stefanprodan@users.noreply.github.com"},
				},
				Readme:        "Podinfo is used by CNCF projects like [Flux](https://github.com/fluxcd/flux2)",
				DefaultValues: "Default values for podinfo.\n\nreplicaCount: 1\n",
			},
		}
	}

	valid_index_charts_spec = []testSpecChartWithFile{
		{
			name:     "acs-engine-autoscaler",
			tgzFile:  testTgz("acs-engine-autoscaler-2.1.1.tgz"),
			revision: "2.1.1",
		},
		{
			name:     "wordpress",
			tgzFile:  testTgz("wordpress-0.7.5.tgz"),
			revision: "0.7.5",
		},
		{
			name:     "wordpress",
			tgzFile:  testTgz("wordpress-0.7.4.tgz"),
			revision: "0.7.4",
		},
	}

	valid_index_available_package_summaries = []*corev1.AvailablePackageSummary{
		{
			Name:             "acs-engine-autoscaler",
			DisplayName:      "acs-engine-autoscaler",
			LatestVersion:    pkgAppVersion("2.1.1"),
			IconUrl:          "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
			ShortDescription: "Scales worker nodes within agent pools",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "bitnami-1/acs-engine-autoscaler",
				Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		},
		{
			Name:        "wordpress",
			DisplayName: "wordpress",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "0.7.5",
				AppVersion: "4.9.1",
			},
			IconUrl:          "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
			ShortDescription: "new description!",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "bitnami-1/wordpress",
				Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		},
	}

	cert_manager_summary = &corev1.AvailablePackageSummary{
		Name:             "cert-manager",
		DisplayName:      "cert-manager",
		LatestVersion:    pkgAppVersion("v1.4.0"),
		IconUrl:          "https://raw.githubusercontent.com/jetstack/cert-manager/master/logo/logo.png",
		ShortDescription: "A Helm chart for cert-manager",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "jetstack-1/cert-manager",
			Context:    &corev1.Context{Namespace: "ns1", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{""},
	}

	elasticsearch_summary = &corev1.AvailablePackageSummary{
		Name:        "elasticsearch",
		DisplayName: "elasticsearch",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "15.5.0",
			AppVersion: "7.13.2",
		},
		IconUrl:          "https://bitnami.com/assets/stacks/elasticsearch/img/elasticsearch-stack-220x234.png",
		ShortDescription: "A highly scalable open-source full-text search and analytics engine",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "index-with-categories-1/elasticsearch",
			Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{"Analytics"},
	}

	ghost_summary = &corev1.AvailablePackageSummary{
		Name:        "ghost",
		DisplayName: "ghost",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "13.0.14",
			AppVersion: "4.7.0",
		},
		IconUrl:          "https://bitnami.com/assets/stacks/ghost/img/ghost-stack-220x234.png",
		ShortDescription: "A simple, powerful publishing platform that allows you to share your stories with the world",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Identifier: "index-with-categories-1/ghost",
			Context:    &corev1.Context{Namespace: "default", Cluster: KubeappsCluster},
			Plugin:     fluxPlugin,
		},
		Categories: []string{"CMS"},
	}

	index_with_categories_summaries = []*corev1.AvailablePackageSummary{
		elasticsearch_summary,
		ghost_summary,
	}

	index_before_update_summaries = []*corev1.AvailablePackageSummary{
		{
			Name:        "alpine",
			DisplayName: "alpine",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "0.2.0",
			},
			IconUrl:          "",
			ShortDescription: "Deploy a basic Alpine Linux pod",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "testrepo/alpine",
				Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		},
		{
			Name:        "nginx",
			DisplayName: "nginx",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "1.1.0",
			},
			IconUrl:          "",
			ShortDescription: "Create a basic nginx HTTP server",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "testrepo/nginx",
				Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		},
	}

	index_after_update_summaries = []*corev1.AvailablePackageSummary{
		{
			Name:        "alpine",
			DisplayName: "alpine",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "0.3.0",
			},
			IconUrl:          "",
			ShortDescription: "Deploy a basic Alpine Linux pod",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "testrepo/alpine",
				Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		},
		{
			Name:        "nginx",
			DisplayName: "nginx",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "1.1.0",
			},
			IconUrl:          "",
			ShortDescription: "Create a basic nginx HTTP server",
			AvailablePackageRef: &corev1.AvailablePackageReference{
				Identifier: "testrepo/nginx",
				Context:    &corev1.Context{Namespace: "ns2", Cluster: KubeappsCluster},
				Plugin:     fluxPlugin,
			},
			Categories: []string{""},
		}}

	add_repo_1 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      "http://example.com",
			Interval: metav1.Duration{Duration: 10 * time.Minute},
		},
	}

	add_repo_2 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:       "http://example.com",
			Interval:  metav1.Duration{Duration: 10 * time.Minute},
			SecretRef: &fluxmeta.LocalObjectReference{Name: "bar-"},
		},
	}

	add_repo_3 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:       "http://example.com",
			Interval:  metav1.Duration{Duration: 10 * time.Minute},
			SecretRef: &fluxmeta.LocalObjectReference{Name: "secret-1"},
		},
	}

	add_repo_4 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:             "http://example.com",
			Interval:        metav1.Duration{Duration: 10 * time.Minute},
			SecretRef:       &fluxmeta.LocalObjectReference{Name: "bar-"},
			PassCredentials: true,
		},
	}

	add_repo_5 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:             "http://example.com",
			Interval:        metav1.Duration{Duration: 10 * time.Minute},
			PassCredentials: true,
		},
	}

	add_repo_6 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      github_stefanprodan_podinfo_oci_registry_url,
			Interval: metav1.Duration{Duration: 10 * time.Minute},
			Type:     "oci",
		},
	}

	add_repo_7 = sourcev1.HelmRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       sourcev1.HelmRepositoryKind,
			APIVersion: sourcev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL:      github_stefanprodan_podinfo_oci_registry_url,
			Interval: metav1.Duration{Duration: 10 * time.Minute},
			Type:     "oci",
			Provider: "gcp",
		},
	}

	add_repo_req_1 = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		NamespaceScoped: true,
	}

	add_repo_req_2 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "foobar",
	}

	add_repo_req_3 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
	}

	add_repo_req_4 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			InsecureSkipVerify: true,
		},
	}

	add_repo_req_5 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
	}

	add_repo_req_6 = func(ca []byte) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "bar",
			Context: &corev1.Context{Namespace: "foo"},
			Type:    "helm",
			Url:     "http://example.com",
			TlsConfig: &corev1.PackageRepositoryTlsConfig{
				PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
					CertAuthority: string(ca),
				},
			},
		}
	}

	add_repo_req_7 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-1",
				},
			},
		},
	}

	add_repo_req_8 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: "baz",
					Password: "zot",
				},
			},
			PassCredentials: true,
		},
	}

	add_repo_req_9 = func(pub, priv []byte) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "bar",
			Context: &corev1.Context{Namespace: "foo"},
			Type:    "helm",
			Url:     "http://example.com",
			Auth:    tls_auth(pub, priv),
		}
	}

	add_repo_req_10 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
				Header: "foobarzot",
			},
		},
	}

	add_repo_req_11 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
				Header: "foobarzot",
			},
		},
	}

	add_repo_req_12 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
				DockerCreds: &corev1.DockerCredentials{
					Server:   "your.private.registry.example.com",
					Username: "janedoe",
					Password: "xxxxxxxx",
					Email:    "jdoe@example.com",
				},
			},
		},
	}

	add_repo_req_13 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth:    secret_1_auth,
	}

	add_repo_req_14 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth:    secret_1_auth,
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-2",
				},
			},
		},
	}

	add_repo_req_15 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_repo_url,
	}

	add_repo_req_16 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-2",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_basic_auth_repo_url,
		Auth:    foo_bar_auth,
	}

	add_repo_req_17 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-3",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_basic_auth_repo_url,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: "foo",
					Password: "bar-2",
				},
			},
		},
	}

	add_repo_req_18 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-4",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_basic_auth_repo_url,
		Auth:    secret_1_auth,
	}

	add_repo_req_19 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-4",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_tls_repo_url,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-2",
				},
			},
		},
	}

	add_repo_req_20 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth: &corev1.PackageRepositoryAuth{
			PassCredentials: true,
		},
	}

	add_repo_req_21 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-5",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "oci",
		Url:     github_stefanprodan_podinfo_oci_registry_url,
	}

	add_repo_req_22 = func(user, password string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "my-podinfo-6",
			Context: &corev1.Context{Namespace: "default"},
			Type:    "oci",
			Url:     github_stefanprodan_podinfo_oci_registry_url,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: user,
						Password: password,
					},
				},
			},
		}
	}

	add_repo_req_23 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-7",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "oci",
		Url:     github_stefanprodan_podinfo_oci_registry_url,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-3",
				},
			},
		},
	}

	add_repo_req_24 = func(server, user, password string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "my-podinfo-8",
			Context: &corev1.Context{Namespace: "default"},
			Type:    "oci",
			Url:     github_stefanprodan_podinfo_oci_registry_url,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Server:   server,
						Username: user,
						Password: password,
					},
				},
			},
		}
	}

	add_repo_req_25 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-9",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "oci",
		Url:     github_stefanprodan_podinfo_oci_registry_url,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-4",
				},
			},
		},
	}

	add_repo_req_26 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "oci",
		Url:     github_stefanprodan_podinfo_oci_registry_url,
	}

	add_repo_req_27 = func(server, user, password string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "my-podinfo-10",
			Context: &corev1.Context{Namespace: "default"},
			Type:    "oci",
			Url:     harbor_stefanprodan_podinfo_oci_registry_url,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Server:   server,
						Username: user,
						Password: password,
					},
				},
			},
		}
	}

	add_repo_req_28 = func(server, user, password string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:    "my-podinfo-10",
			Context: &corev1.Context{Namespace: "default"},
			Type:    "oci",
			Url:     gcp_stefanprodan_podinfo_oci_registry_url,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Server:   server,
						Username: user,
						Password: password,
					},
				},
			},
		}
	}

	add_repo_req_29 = func() *corev1.AddPackageRepositoryRequest {
		customDetail, _ := anypb.New(&v1alpha1.FluxPackageRepositoryCustomDetail{
			Provider: "gcp",
		})

		return &corev1.AddPackageRepositoryRequest{
			Name:         "bar",
			Context:      &corev1.Context{Namespace: "foo"},
			Type:         "oci",
			Url:          github_stefanprodan_podinfo_oci_registry_url,
			CustomDetail: customDetail,
		}
	}

	add_repo_expected_resp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("bar", "foo"),
	}

	add_repo_expected_resp_2 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo", "default"),
	}

	add_repo_expected_resp_3 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-2", "default"),
	}

	add_repo_expected_resp_4 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-3", "default"),
	}

	add_repo_expected_resp_5 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-4", "default"),
	}

	add_repo_expected_resp_6 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-5", "default"),
	}

	add_repo_expected_resp_7 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-6", "default"),
	}

	add_repo_expected_resp_8 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-7", "default"),
	}

	add_repo_expected_resp_9 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-8", "default"),
	}

	add_repo_expected_resp_10 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-9", "default"),
	}

	add_repo_expected_resp_11 = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("my-podinfo-10", "default"),
	}

	status_installed = &corev1.InstalledPackageStatus{
		Ready:      true,
		Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
		UserReason: "ReconciliationSucceeded: Release reconciliation succeeded",
	}

	my_redis_ref = installedRef("my-redis", "test")

	redis_summary_installed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           status_installed,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_failed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	redis_summary_pending_2 = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "ArtifactFailed: HelmChart 'default/kubeapps-my-redis' is not ready",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_installed = &corev1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.7.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           status_installed,
	}

	redis_summary_latest = &corev1.InstalledPackageSummary{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		IconUrl:             "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		PkgDisplayName:   "redis",
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Status:           status_installed,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
	}

	airflow_summary_semver = &corev1.InstalledPackageSummary{
		InstalledPackageRef: installedRef("my-airflow", "namespace-2"),
		Name:                "my-airflow",
		IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
		PkgVersionReference: &corev1.VersionReference{
			Version: "<=6.7.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.7.1",
			AppVersion: "1.10.12",
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "10.2.1",
			AppVersion: "2.1.0",
		},
		ShortDescription: "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		PkgDisplayName:   "airflow",
		Status:           status_installed,
	}

	lastTransitionTime, _ = time.Parse(time.RFC3339, "2021-08-11T08:46:03Z")
	lastUpdateTime, _     = time.Parse(time.RFC3339, "2021-07-01T05:09:45Z")

	redis_existing_spec_completed = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            testYaml("redis-many-versions.yaml"),
		chartName:            "redis",
		chartTarGz:           testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_completed = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusDeployed,
	}

	redis_existing_spec_completed_with_values_and_reconciliation_options_values_bytes, _ = json.Marshal(
		map[string]interface{}{
			"replica": map[string]interface{}{
				"replicaCount":  "1",
				"configuration": "xyz",
			}})

	redis_existing_spec_completed_with_values_and_reconciliation_options = testSpecGetInstalledPackages{
		repoName:                  "bitnami-1",
		repoNamespace:             "default",
		repoIndex:                 testYaml("redis-many-versions.yaml"),
		chartName:                 "redis",
		chartTarGz:                testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:          "14.4.0",
		chartArtifactVersion:      "14.4.0",
		releaseName:               "my-redis",
		releaseNamespace:          "test",
		releaseSuspend:            true,
		releaseServiceAccountName: "foo",
		releaseValues:             &v1.JSON{Raw: redis_existing_spec_completed_with_values_and_reconciliation_options_values_bytes},
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_spec_failed = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            testYaml("redis-many-versions.yaml"),
		chartName:            "redis",
		chartTarGz:           testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             helmv2.InstallFailedReason,
					Message:            "install retries exhausted",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionFalse,
					Reason:             helmv2.InstallFailedReason,
					Message:            "Helm install failed: unable to build kubernetes objects from release manifest: error validating \"\": error validating data: ValidationError(Deployment.spec.replicas): invalid type for io.k8s.api.apps.v1.DeploymentSpec.replicas: got \"string\", expected \"integer\"",
				},
			},
			HelmChart:             "default/redis",
			Failures:              14,
			InstallFailures:       1,
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_failed = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusFailed,
	}

	airflow_existing_spec_completed = testSpecGetInstalledPackages{
		repoName:             "bitnami-2",
		repoNamespace:        "default",
		repoIndex:            testYaml("airflow-many-versions.yaml"),
		chartName:            "airflow",
		chartTarGz:           testTgz("airflow-6.7.1.tgz"),
		chartSpecVersion:     "6.7.1",
		chartArtifactVersion: "6.7.1",
		releaseName:          "my-airflow",
		releaseNamespace:     "namespace-2",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/airflow",
			LastAppliedRevision:   "6.7.1",
			LastAttemptedRevision: "6.7.1",
		},
	}

	airflow_existing_spec_semver = testSpecGetInstalledPackages{
		repoName:             "bitnami-2",
		repoNamespace:        "default",
		repoIndex:            testYaml("airflow-many-versions.yaml"),
		chartName:            "airflow",
		chartTarGz:           testTgz("airflow-6.7.1.tgz"),
		chartSpecVersion:     "<=6.7.1",
		chartArtifactVersion: "6.7.1",
		releaseName:          "my-airflow",
		releaseNamespace:     "namespace-2",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/airflow",
			LastAppliedRevision:   "6.7.1",
			LastAttemptedRevision: "6.7.1",
		},
	}

	redis_existing_spec_pending = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            testYaml("redis-many-versions.yaml"),
		chartName:            "redis",
		chartTarGz:           testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             "Unknown",
					Reason:             "Progressing",
					Message:            "reconciliation in progress",
				},
			},
			HelmChart:             "default/redis",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_spec_pending_2 = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            testYaml("redis-many-versions.yaml"),
		chartName:            "redis",
		chartTarGz:           testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             helmv2.ArtifactFailedReason,
					Message:            "HelmChart 'default/kubeapps-my-redis' is not ready",
				},
			},
			HelmChart:             "default/redis",
			Failures:              2,
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_existing_stub_pending = helmReleaseStub{
		name:         "my-redis",
		namespace:    "test",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusPendingInstall,
	}

	redis_existing_spec_latest = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            testYaml("redis-many-versions.yaml"),
		chartName:            "redis",
		chartTarGz:           testTgz("redis-14.4.0.tgz"),
		chartSpecVersion:     "*",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
	}

	redis_detail_failed = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "1m",
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_FAILED,
			UserReason: "InstallFailed: install retries exhausted",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_pending = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "1.2.3",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "1m",
		},
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "Progressing: reconciliation in progress",
		},
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &corev1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: "1m",
		},
		Status:                status_installed,
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	redis_detail_completed_with_values_and_reconciliation_options = &corev1.InstalledPackageDetail{
		InstalledPackageRef: my_redis_ref,
		Name:                "my-redis",
		CurrentVersion: &corev1.PackageAppVersion{
			AppVersion: "1.2.3",
			PkgVersion: "14.4.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "14.4.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           "1m",
			Suspend:            true,
			ServiceAccountName: "foo",
		},
		Status:                status_installed,
		ValuesApplied:         "{\"replica\": { \"replicaCount\":  \"1\", \"configuration\": \"xyz\"    }}",
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	flux_helm_release_basic = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart: "podinfo",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
	})

	flux_helm_release_semver_constraint = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart: "podinfo",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
		Version: "> 5",
	})

	flux_helm_release_reconcile_options = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
				},
			},
			Interval:           metav1.Duration{Duration: 1 * time.Minute},
			ServiceAccountName: "foo",
			Suspend:            false,
		},
	}

	flux_helm_release_values_values_bytes, _ = json.Marshal(
		map[string]interface{}{
			"ui": map[string]interface{}{
				"message": "what we do in the shadows",
			}})

	flux_helm_release_values = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-podinfo",
			Namespace:       "test",
			ResourceVersion: "1",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "podinfo",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "podinfo",
						Namespace: "namespace-1",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
			Values:   &v1.JSON{Raw: flux_helm_release_values_values_bytes},
		},
	}

	create_installed_package_resp_my_podinfo = &corev1.CreateInstalledPackageResponse{
		InstalledPackageRef: installedRef("my-podinfo", "test"),
	}

	flux_helm_release_updated_1 = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
					Version: ">14.4.0",
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_updated_2 = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
			Values:   &v1.JSON{Raw: flux_helm_release_values_values_bytes},
		},
	}

	redis_existing_spec_target_ns_is_set = testSpecGetInstalledPackages{
		repoName:             "bitnami-1",
		repoNamespace:        "default",
		repoIndex:            "testdata/charts/redis-many-versions.yaml",
		chartName:            "redis",
		chartTarGz:           "testdata/charts/redis-14.4.0.tgz",
		chartSpecVersion:     "14.4.0",
		chartArtifactVersion: "14.4.0",
		releaseName:          "my-redis",
		releaseNamespace:     "test",
		releaseStatus: helmv2.HelmReleaseStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionTrue,
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             metav1.ConditionTrue,
					Reason:             helmv2.InstallSucceededReason,
					Message:            "Helm install succeeded",
				},
			},
			HelmChart:             "default/redis",
			LastAppliedRevision:   "14.4.0",
			LastAttemptedRevision: "14.4.0",
		},
		targetNamespace: "test2",
	}

	redis_existing_stub_target_ns_is_set = helmReleaseStub{
		name:         "test2-my-redis",
		namespace:    "test2",
		chartVersion: "14.4.0",
		notes:        "some notes",
		status:       release.StatusDeployed,
	}

	flux_helm_release_updated_target_ns_is_set = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart: "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
					Version: ">14.4.0",
				},
			},
			Interval:        metav1.Duration{Duration: 1 * time.Minute},
			TargetNamespace: "test2",
		},
	}

	redis_charts_spec = []testSpecChartWithFile{
		{
			name:     "redis",
			tgzFile:  testTgz("redis-14.4.0.tgz"),
			revision: "14.4.0",
		},
		{
			name:     "redis",
			tgzFile:  testTgz("redis-14.3.4.tgz"),
			revision: "14.3.4",
		},
	}

	expected_detail_redis_1 = &corev1.AvailablePackageDetail{
		AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
		Name:                "redis",
		Version: &corev1.PackageAppVersion{
			PkgVersion: "14.4.0",
			AppVersion: "6.2.4",
		},
		RepoUrl:          "https://example.repo.com/charts",
		HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
		IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		DisplayName:      "redis",
		Categories:       []string{"Database"},
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
		DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
		ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
		SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
		Maintainers: []*corev1.Maintainer{
			{
				Name:  "Bitnami",
				Email: "containers@bitnami.com",
			},
			{
				Name:  "desaintmartin",
				Email: "cedric@desaintmartin.fr",
			},
		},
	}

	expected_detail_redis_2 = &corev1.AvailablePackageDetail{
		AvailablePackageRef: availableRef("bitnami-1/redis", "default"),
		Name:                "redis",
		Version: &corev1.PackageAppVersion{
			PkgVersion: "14.3.4",
			AppVersion: "6.2.4",
		},
		RepoUrl:          "https://example.repo.com/charts",
		IconUrl:          "https://bitnami.com/assets/stacks/redis/img/redis-stack-220x234.png",
		HomeUrl:          "https://github.com/bitnami/charts/tree/master/bitnami/redis",
		DisplayName:      "redis",
		Categories:       []string{"Database"},
		ShortDescription: "Open source, advanced key-value store. It is often referred to as a data structure server since keys can contain strings, hashes, lists, sets and sorted sets.",
		Readme:           "Redis<sup>TM</sup> Chart packaged by Bitnami\n\n[Redis<sup>TM</sup>](http://redis.io/) is an advanced key-value cache",
		DefaultValues:    "## @param global.imageRegistry Global Docker image registry",
		ValuesSchema:     "\"$schema\": \"http://json-schema.org/schema#\"",
		SourceUrls:       []string{"https://github.com/bitnami/bitnami-docker-redis", "http://redis.io/"},
		Maintainers: []*corev1.Maintainer{
			{
				Name:  "Bitnami",
				Email: "containers@bitnami.com",
			},
			{
				Name:  "desaintmartin",
				Email: "cedric@desaintmartin.fr",
			},
		},
	}

	expected_versions_redis = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "14.4.0", AppVersion: "6.2.4"},
			{PkgVersion: "14.3.4", AppVersion: "6.2.4"},
			{PkgVersion: "14.3.3", AppVersion: "6.2.4"},
			{PkgVersion: "14.3.2", AppVersion: "6.2.3"},
			{PkgVersion: "14.2.1", AppVersion: "6.2.3"},
			{PkgVersion: "14.2.0", AppVersion: "6.2.3"},
			{PkgVersion: "13.0.1", AppVersion: "6.2.1"},
			{PkgVersion: "13.0.0", AppVersion: "6.2.1"},
			{PkgVersion: "12.10.1", AppVersion: "6.0.12"},
			{PkgVersion: "12.10.0", AppVersion: "6.0.12"},
			{PkgVersion: "12.9.2", AppVersion: "6.0.12"},
			{PkgVersion: "12.9.1", AppVersion: "6.0.12"},
			{PkgVersion: "12.9.0", AppVersion: "6.0.12"},
			{PkgVersion: "12.8.3", AppVersion: "6.0.12"},
			{PkgVersion: "12.8.2", AppVersion: "6.0.12"},
			{PkgVersion: "12.8.1", AppVersion: "6.0.12"},
		},
	}

	expected_versions_airflow = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "1.0.0", AppVersion: "2.1.4"},
		},
	}

	expected_versions_stefanprodan_podinfo = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "6.1.8"},
			{PkgVersion: "6.1.7"},
			{PkgVersion: "6.1.6"},
		},
	}

	expected_versions_gfichtenholt_podinfo = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "6.1.5"},
		},
	}

	expected_versions_podinfo_2 = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "6.1.5"},
			{PkgVersion: "6.0.3"},
			{PkgVersion: "6.0.0"},
		},
	}

	expected_versions_gfichtenholt_podinfo_3 = &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: []*corev1.PackageAppVersion{
			{PkgVersion: "6.1.6"},
			{PkgVersion: "6.1.5"},
		},
	}

	create_package_simple_req = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
	}

	create_package_semver_constraint_req = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
	}

	create_package_reconcile_options_req = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           "1m",
			Suspend:            false,
			ServiceAccountName: "foo",
		},
	}

	create_package_values_json_override = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
		Values:              "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	create_package_values_yaml_override = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
		Values:              "# Default values for podinfo.\n---\nui:\n  message: what we do in the shadows",
	}

	create_package_for_test_of_upgrade_policy = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo/podinfo", "namespace-1"),
		Name:                "my-podinfo",
		TargetContext:       &corev1.Context{Namespace: "test"},
		PkgVersionReference: &corev1.VersionReference{
			Version: "5.2.1",
		},
	}

	flux_helm_release_upgrade_policy_none = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart:   "podinfo",
		Version: "5.2.1",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
	})

	flux_helm_release_upgrade_policy_major = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart:   "podinfo",
		Version: ">=5.2.1",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
	})

	flux_helm_release_upgrade_policy_minor = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart:   "podinfo",
		Version: ">=5.2.1 <6.0.0",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
	})

	flux_helm_release_upgrade_policy_patch = newFluxHelmRelease(helmv2.HelmChartTemplateSpec{
		Chart:   "podinfo",
		Version: ">=5.2.1 <5.3.0",
		SourceRef: helmv2.CrossNamespaceObjectReference{
			Kind:      sourcev1.HelmRepositoryKind,
			Name:      "podinfo",
			Namespace: "namespace-1",
		},
	})

	flux_helm_release_updated_upgrade_major = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Version: ">=14.4.0",
					Chart:   "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_updated_upgrade_minor = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Version: ">=14.4.0 <15.0.0",
					Chart:   "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	flux_helm_release_updated_upgrade_patch = &helmv2.HelmRelease{
		TypeMeta: metav1.TypeMeta{
			Kind:       helmv2.HelmReleaseKind,
			APIVersion: helmv2.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "my-redis",
			Namespace:       "test",
			Generation:      int64(1),
			ResourceVersion: "1000",
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Version: ">=14.4.0 <14.5.0",
					Chart:   "redis",
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "bitnami-1",
						Namespace: "default",
					},
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

	get_repo_detail_req_1 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
	}

	get_repo_detail_package_resp_ref = repoRef("repo-1", "namespace-1")

	get_repo_detail_resp_1 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status:          podinfo_repo_status_2,
		},
	}

	get_repo_detail_req_2 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("repo-kaka", "namespace-1"),
	}

	get_repo_detail_req_3 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-kaka"),
	}

	get_repo_detail_req_4 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Identifier: "repo-1",
		},
	}

	get_repo_detail_req_5 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: "namespace-1",
				Cluster:   "this-is-not-the-cluster-youre-looking-for",
			},
			Identifier: "repo-1",
		},
	}

	get_repo_detail_resp_6 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			TlsConfig: &corev1.PackageRepositoryTlsConfig{
				InsecureSkipVerify: false,
				PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: "secret-1",
						Key:  "caFile",
					},
				},
			},
			Status: podinfo_repo_status_2,
		},
	}

	get_repo_detail_resp_6a = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			TlsConfig: &corev1.PackageRepositoryTlsConfig{
				InsecureSkipVerify: false,
				PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
					CertAuthority: redactedString,
				},
			},
			Status: podinfo_repo_status_2,
		},
	}

	get_repo_detail_resp_7 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      false,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_PENDING,
				UserReason: "Progressing: reconciliation in progress",
			},
		},
	}

	get_repo_detail_resp_8 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      false,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_FAILED,
				UserReason: "Failed: failed to fetch https://invalid.example.com/index.yaml : 404 Not Found",
			},
		},
	}

	get_repo_detail_resp_9 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth: &corev1.PackageRepositoryAuth{
				PassCredentials: false,
				Type:            corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: "secret-1",
					},
				},
			},
			Status: podinfo_repo_status_2,
		},
	}

	get_repo_detail_resp_9a = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            tls_auth_redacted,
			Status:          podinfo_repo_status_2,
		},
	}

	get_repo_detail_req_6 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("my-podinfo", "TBD"),
	}

	get_repo_detail_resp_10 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            secret_1_auth,
			Status:          podinfo_repo_status_2,
		},
	}

	get_repo_detail_resp_10a = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "1m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_2,
		},
	}

	get_repo_detail_resp_11 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo"),
			Name:            "my-podinfo",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_repo_url,
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status:          podinfo_repo_status_3,
		},
	}

	get_repo_detail_req_7 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("my-bitnami", "TBD"),
	}

	get_repo_detail_resp_12 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-bitnami"),
			Name:            "my-bitnami",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://charts.bitnami.com/bitnami",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      true,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
				UserReason: "Succeeded: stored artifact for revision '",
			},
		},
	}

	get_repo_detail_resp_13 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-2"),
			Name:            "my-podinfo-2",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      false,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_FAILED,
				UserReason: "Failed: failed to fetch Helm repository index: failed to cache index to temporary file: failed to fetch http://fluxv2plugin-testdata-svc.default.svc.cluster.local:80/podinfo-basic-auth/index.yaml : 401 Unauthorize",
			},
		},
	}

	get_repo_detail_resp_14 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-3"),
			Name:            "my-podinfo-3",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            secret_1_auth,
			Status:          podinfo_repo_status_1,
		},
	}

	get_repo_detail_resp_14a = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-3"),
			Name:            "my-podinfo-3",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_1,
		},
	}

	get_repo_detail_req_8 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("my-kaka", "TBD"),
	}

	get_repo_detail_req_9 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-2", "TBD"),
	}

	get_repo_detail_req_10 = &corev1.GetPackageRepositoryDetailRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-3", "TBD"),
	}

	get_repo_detail_req_11 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-4", "TBD"),
	}

	get_repo_detail_req_12 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-12", "TBD"),
	}

	get_repo_detail_resp_15 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-12"),
			Name:            "my-podinfo-12",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      false,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_FAILED,
				UserReason: "Failed: failed to fetch Helm repository index: failed to cache index to temporary file: object required",
			},
		},
	}

	get_repo_detail_req_13 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-13", "TBD"),
	}

	get_repo_detail_req_14 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-14", "TBD"),
	}

	get_repo_detail_req_15 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-15", "TBD"),
	}

	get_repo_detail_req_16 = &corev1.GetPackageRepositoryDetailRequest{
		// namespace will be set when test scenario is run
		PackageRepoRef: repoRefInReq("my-podinfo-16", "TBD"),
	}

	get_repo_detail_resp_16 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-13"),
			Name:            "my-podinfo-13",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      true,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
				UserReason: "Succeeded: Helm repository is ready",
			},
		},
	}

	get_repo_detail_resp_17 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-14"),
			Name:            "my-podinfo-14",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Interval:        "10m",
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: redactedString,
						Password: redactedString,
					},
				},
			},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      true,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
				UserReason: "Succeeded: Helm repository is ready",
			},
		},
	}

	get_repo_detail_resp_18 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-15"),
			Name:            "my-podinfo-15",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Interval:        "10m",
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Username: redactedString,
						Password: redactedString,
						Server:   redactedString,
					},
				},
			},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      true,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
				UserReason: "Succeeded: Helm repository is ready",
			},
		},
	}

	get_repo_detail_resp_19 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             "oci://localhost:54321/userX/charts",
			Interval:        "1m",
			Auth:            &corev1.PackageRepositoryAuth{},
			Status:          podinfo_repo_status_4,
		},
	}

	get_repo_detail_resp_20 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-16"),
			Name:            "my-podinfo-16",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             harbor_stefanprodan_podinfo_oci_registry_url,
			Interval:        "10m",
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Username: redactedString,
						Password: redactedString,
						Server:   redactedString,
					},
				},
			},
			Status: &corev1.PackageRepositoryStatus{
				Ready:      true,
				Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
				UserReason: "Succeeded: Helm repository is ready",
			},
		},
	}

	get_summaries_repo_1 = newRepo("bar", "foo",
		&sourcev1.HelmRepositorySpec{
			URL:       "http://example.com",
			Interval:  metav1.Duration{Duration: 10 * time.Minute},
			SecretRef: &fluxmeta.LocalObjectReference{},
		},
		&sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Checksum:       "651f952130ea96823711d08345b85e82be011dc6",
				LastUpdateTime: metav1.Time{Time: lastUpdateTime},
				Path:           "helmrepository/default/bitnami/index-651f952130ea96823711d08345b85e82be011dc6.yaml",
				Revision:       "651f952130ea96823711d08345b85e82be011dc6",
				URL:            "http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index-651f952130ea96823711d08345b85e82be011dc6.yaml",
			},
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Message:            "stored artifact for revision '651f952130ea96823711d08345b85e82be011dc6'",
					Reason:             fluxmeta.SucceededReason,
					Status:             metav1.ConditionTrue,
					Type:               fluxmeta.ReadyCondition,
				},
			},
			URL: "TBD",
		})

	get_summaries_repo_2 = newRepo("zot", "xyz",
		&sourcev1.HelmRepositorySpec{
			URL:      "http://example.com",
			Interval: metav1.Duration{Duration: 10 * time.Minute},
		},
		&sourcev1.HelmRepositoryStatus{
			Artifact: &sourcev1.Artifact{
				Checksum:       "651f952130ea96823711d08345b85e82be011dc6",
				LastUpdateTime: metav1.Time{Time: lastUpdateTime},
				Path:           "helmrepository/default/bitnami/index-651f952130ea96823711d08345b85e82be011dc6.yaml",
				Revision:       "651f952130ea96823711d08345b85e82be011dc6",
				URL:            "http://source-controller.flux-system.svc.cluster.local./helmrepository/default/bitnami/index-651f952130ea96823711d08345b85e82be011dc6.yaml",
			},
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Message:            "stored artifact for revision '651f952130ea96823711d08345b85e82be011dc6'",
					Reason:             fluxmeta.SucceededReason,
					Status:             metav1.ConditionTrue,
					Type:               fluxmeta.ReadyCondition,
				},
			},
			URL: "TBD",
		})

	get_summaries_repo_3 = newRepo("pending", "xyz",
		&sourcev1.HelmRepositorySpec{
			URL:      "http://example.com",
			Interval: metav1.Duration{Duration: 10 * time.Minute},
		},
		&sourcev1.HelmRepositoryStatus{ObservedGeneration: -1},
	)

	get_summaries_repo_4 = newRepo("failed", "xyz",
		&sourcev1.HelmRepositorySpec{
			URL:      "http://example.com",
			Interval: metav1.Duration{Duration: 10 * time.Minute},
		},
		&sourcev1.HelmRepositoryStatus{
			Conditions: []metav1.Condition{
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               fluxmeta.ReadyCondition,
					Status:             metav1.ConditionFalse,
					Reason:             fluxmeta.FailedReason,
					Message:            "failed to fetch https://invalid.example.com/index.yaml : 404 Not Found",
				},
			},
		})

	get_summaries_summary_1 = &corev1.PackageRepositorySummary{
		PackageRepoRef:  repoRef("bar", "foo"),
		Name:            "bar",
		Description:     "",
		NamespaceScoped: false,
		Type:            "helm",
		Url:             "http://example.com",
		Status:          podinfo_repo_status_2,
		RequiresAuth:    true,
	}

	get_summaries_summary_2 = &corev1.PackageRepositorySummary{
		PackageRepoRef:  repoRef("zot", "xyz"),
		Name:            "zot",
		Description:     "",
		NamespaceScoped: false,
		Type:            "helm",
		Url:             "http://example.com",
		Status:          podinfo_repo_status_2,
		RequiresAuth:    false,
	}

	get_summaries_summary_3 = &corev1.PackageRepositorySummary{
		PackageRepoRef:  repoRef("pending", "xyz"),
		Name:            "pending",
		Description:     "",
		NamespaceScoped: false,
		Type:            "helm",
		Url:             "http://example.com",
		Status:          repo_status_pending,
		RequiresAuth:    false,
	}

	get_summaries_summary_4 = &corev1.PackageRepositorySummary{
		PackageRepoRef:  repoRef("failed", "xyz"),
		Name:            "failed",
		Description:     "",
		NamespaceScoped: false,
		Type:            "helm",
		Url:             "http://example.com",
		Status: &corev1.PackageRepositoryStatus{
			Ready:      false,
			Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_FAILED,
			UserReason: "Failed: failed to fetch https://invalid.example.com/index.yaml : 404 Not Found",
		},
		RequiresAuth: false,
	}

	get_summaries_summary_5 = func(name types.NamespacedName) *corev1.PackageRepositorySummary {
		return &corev1.PackageRepositorySummary{
			PackageRepoRef:  repoRef(name.Name, name.Namespace),
			Name:            name.Name,
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_repo_url,
			Status:          podinfo_repo_status_3,
			RequiresAuth:    false,
		}
	}

	get_summaries_summary_6 = func(name types.NamespacedName) *corev1.PackageRepositorySummary {
		return &corev1.PackageRepositorySummary{
			PackageRepoRef:  repoRef(name.Name, name.Namespace),
			Name:            name.Name,
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Status:          podinfo_repo_status_4,
			RequiresAuth:    false,
		}
	}

	get_summaries_summary_7 = func(name types.NamespacedName) *corev1.PackageRepositorySummary {
		return &corev1.PackageRepositorySummary{
			PackageRepoRef:  repoRef(name.Name, name.Namespace),
			Name:            name.Name,
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             harbor_stefanprodan_podinfo_oci_registry_url,
			Status:          podinfo_repo_status_4,
			RequiresAuth:    false,
		}
	}

	update_repo_req_1 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "http://newurl.com",
	}

	update_repo_req_2 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
		Interval:       "5m45s",
	}

	update_repo_req_3 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
		Auth:           &corev1.PackageRepositoryAuth{PassCredentials: true},
	}

	update_repo_req_4 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-1",
					Key:  "caFile",
				},
			},
		},
	}

	update_repo_req_5 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
	}

	update_repo_req_6 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
		Auth:           secret_1_auth,
	}

	update_repo_req_7 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
	}

	update_repo_req_8 = func(pub, priv []byte) *corev1.UpdatePackageRepositoryRequest {
		return &corev1.UpdatePackageRepositoryRequest{
			PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
			Url:            "https://example.repo.com/charts",
			Auth:           tls_auth(pub, priv),
		}
	}

	update_repo_req_9 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
	}

	update_repo_req_10 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "https://example.repo.com/charts",
		Auth:           foo_bar_auth,
	}

	update_repo_req_11 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo", "TBD"),
		Url:            podinfo_basic_auth_repo_url,
		Auth:           foo_bar_auth,
	}

	update_repo_req_12 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-2", "TBD"),
		Url:            podinfo_basic_auth_repo_url,
		Auth:           foo_bar_auth,
	}

	update_repo_req_13 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-3", "TBD"),
		Url:            podinfo_repo_url,
	}

	update_repo_req_14 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-4", "TBD"),
		Url:            podinfo_basic_auth_repo_url,
		Auth:           secret_1_auth,
	}

	update_repo_req_15 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-5", "TBD"),
		Url:            podinfo_basic_auth_repo_url,
		Auth:           foo_bar_auth,
	}

	update_repo_req_16 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
		Url:            "http://newurl.com",
		Auth:           foo_bar_auth_redacted,
	}

	update_repo_req_17 = &corev1.UpdatePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-6", "TBD"),
		Url:            podinfo_basic_auth_repo_url,
		Auth:           foo_bar_auth_redacted,
	}

	update_repo_req_18 = func(ghUser, ghPasswd string) *corev1.UpdatePackageRepositoryRequest {
		return &corev1.UpdatePackageRepositoryRequest{
			PackageRepoRef: repoRefInReq("my-podinfo-7", "TBD"),
			Url:            github_stefanprodan_podinfo_oci_registry_url,
			Auth:           github_auth(ghUser, ghPasswd),
			Interval:       "4m44s",
		}
	}

	update_repo_resp_1 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRef("repo-1", "namespace-1"),
	}

	update_repo_resp_2 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo"),
	}

	update_repo_resp_3 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo-2"),
	}

	update_repo_resp_4 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo-4"),
	}

	update_repo_resp_5 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo-5"),
	}

	update_repo_resp_6 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo-6"),
	}

	update_repo_resp_7 = &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: repoRefWithId("my-podinfo-7"),
	}

	update_repo_detail_1 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "http://newurl.com",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_2 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "5m45s",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_3 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: true},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_4 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			TlsConfig: &corev1.PackageRepositoryTlsConfig{
				PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: "secret-1",
						Key:  "caFile",
					},
				},
			},
			Status: repo_status_pending,
		},
	}

	update_repo_detail_5 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{PassCredentials: false},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_6 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            secret_1_auth,
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_7 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_8 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            tls_auth_redacted,
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_9 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            &corev1.PackageRepositoryAuth{},
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_10 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "https://example.repo.com/charts",
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_11 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo"),
			Name:            "my-podinfo",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_1,
		},
	}

	update_repo_detail_12 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-2"),
			Name:            "my-podinfo-2",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_1,
		},
	}

	update_repo_detail_13 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-4"),
			Name:            "my-podinfo-4",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            secret_1_auth,
			Status:          podinfo_repo_status_1,
		},
	}

	update_repo_detail_14 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-5"),
			Name:            "my-podinfo-5",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_1,
		},
	}

	update_repo_detail_15 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  get_repo_detail_package_resp_ref,
			Name:            "repo-1",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             "http://newurl.com",
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          repo_status_pending,
		},
	}

	update_repo_detail_16 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-6"),
			Name:            "my-podinfo-6",
			Description:     "",
			NamespaceScoped: false,
			Type:            "helm",
			Url:             podinfo_basic_auth_repo_url,
			Interval:        "10m",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_1,
		},
	}

	update_repo_detail_17 = &corev1.GetPackageRepositoryDetailResponse{
		Detail: &corev1.PackageRepositoryDetail{
			PackageRepoRef:  repoRefWithId("my-podinfo-7"),
			Name:            "my-podinfo-7",
			Description:     "",
			NamespaceScoped: false,
			Type:            "oci",
			Url:             github_stefanprodan_podinfo_oci_registry_url,
			Interval:        "4m44s",
			Auth:            foo_bar_auth_redacted,
			Status:          podinfo_repo_status_4,
		},
	}

	foo_bar_auth = &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
			UsernamePassword: &corev1.UsernamePassword{
				Username: "foo",
				Password: "bar",
			},
		},
	}

	foo_bar_auth_redacted = &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
			UsernamePassword: &corev1.UsernamePassword{
				Username: redactedString,
				Password: redactedString,
			},
		},
	}

	github_auth = func(ghUser, ghToken string) *corev1.PackageRepositoryAuth {
		return &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: ghUser,
					Password: ghToken,
				},
			},
		}
	}

	tls_auth = func(pub, priv []byte) *corev1.PackageRepositoryAuth {
		return &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_TlsCertKey{
				TlsCertKey: &corev1.TlsCertKey{
					Cert: string(pub),
					Key:  string(priv),
				},
			},
		}
	}

	tls_auth_redacted = tls_auth([]byte(redactedString), []byte(redactedString))

	secret_1_auth = &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
			SecretRef: &corev1.SecretKeyReference{
				Name: "secret-1",
			},
		},
	}

	podinfo_repo_status_1 = &corev1.PackageRepositoryStatus{
		Ready:      true,
		Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
		UserReason: "Succeeded: stored artifact for revision '9d3ac1eb708dfaebae14d7c88fd46afce8b1e0f7aace790d91758575dc8ce518'",
	}

	podinfo_repo_status_2 = &corev1.PackageRepositoryStatus{
		Ready:      true,
		Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
		UserReason: "Succeeded: stored artifact for revision '651f952130ea96823711d08345b85e82be011dc6'",
	}

	podinfo_repo_status_3 = &corev1.PackageRepositoryStatus{
		Ready:      true,
		Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
		UserReason: "Succeeded: stored artifact for revision '2867920fb8f56575f4bc95ed878ee2a0c8ae79cdd2bca210a72aa3ff04defa1b'",
	}

	podinfo_repo_status_4 = &corev1.PackageRepositoryStatus{
		Ready:      true,
		Reason:     corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS,
		UserReason: "Succeeded: Helm repository is ready",
	}

	repo_status_pending = &corev1.PackageRepositoryStatus{
		Reason: corev1.PackageRepositoryStatus_STATUS_REASON_PENDING,
	}

	podinfo_installed_refs = func(name string) []*corev1.ResourceRef {
		return []*corev1.ResourceRef{
			{
				ApiVersion: "v1",
				Kind:       "Service",
				Name:       name,
			},
			{
				ApiVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       name,
			},
		}
	}

	podinfo_notes = func(name string) string {
		return "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/" + name + " 8080:9898\n"
	}

	// similar to repoRef but w/out Cluster or Plugin fields
	repoRefInReq = func(id, namespace string) *corev1.PackageRepositoryReference {
		return &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: namespace,
			},
			Identifier: id,
		}
	}

	newFluxHelmRelease = func(chartSpec helmv2.HelmChartTemplateSpec) *helmv2.HelmRelease {
		return &helmv2.HelmRelease{
			TypeMeta: metav1.TypeMeta{
				Kind:       helmv2.HelmReleaseKind,
				APIVersion: helmv2.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "my-podinfo",
				Namespace:       "test",
				ResourceVersion: "1",
			},
			Spec: helmv2.HelmReleaseSpec{
				Chart: helmv2.HelmChartTemplate{
					Spec: chartSpec,
				},
				Interval: metav1.Duration{Duration: 1 * time.Minute},
			},
		}
	}

	targetContext = func(namespace string) *corev1.Context {
		return &corev1.Context{
			// note that Namespace is just the prefix - the actual name will
			// have a random string appended at the end, e.g. "test-1-h23r"
			// this will happen during the running of the test
			Namespace: namespace,
			Cluster:   KubeappsCluster,
		}
	}

	pkgAppVersion = func(version string) *corev1.PackageAppVersion {
		return &corev1.PackageAppVersion{
			PkgVersion: version,
			AppVersion: version,
		}
	}

	delete_repo_req_1 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-1", "namespace-1"),
	}

	delete_repo_req_2 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("repo-kaka", "namespace-kaka"),
	}

	delete_repo_req_3 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo", "TBD"),
	}

	delete_repo_req_4 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-2", "TBD"),
	}

	delete_repo_req_5 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-3", "TBD"),
	}

	delete_repo_req_6 = &corev1.DeletePackageRepositoryRequest{
		PackageRepoRef: repoRefInReq("my-podinfo-4", "TBD"),
	}

	expected_oci_stefanprodan_podinfo_available_summaries = func(name string) *corev1.GetAvailablePackageSummariesResponse {
		return &corev1.GetAvailablePackageSummariesResponse{
			AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
				{
					Name:                "podinfo",
					AvailablePackageRef: availableRef(name+"/podinfo", "default"),
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "6.1.8",
					},
					DisplayName:      "podinfo",
					ShortDescription: "Podinfo Helm chart for Kubernetes",
					Categories:       []string{""},
				},
			},
		}
	}

	no_available_summaries = func(name string) *corev1.GetAvailablePackageSummariesResponse {
		return &corev1.GetAvailablePackageSummariesResponse{}
	}

	expected_detail_oci_stefanprodan_podinfo = func(name, url string) *corev1.GetAvailablePackageDetailResponse {
		return &corev1.GetAvailablePackageDetailResponse{
			AvailablePackageDetail: &corev1.AvailablePackageDetail{
				AvailablePackageRef: availableRef(name+"/podinfo", "default"),
				Name:                "podinfo",
				Version:             pkgAppVersion("6.1.8"),
				RepoUrl:             url,
				HomeUrl:             "https://github.com/stefanprodan/podinfo",
				DisplayName:         "podinfo",
				ShortDescription:    "Podinfo Helm chart for Kubernetes",
				SourceUrls:          []string{"https://github.com/stefanprodan/podinfo"},
				Maintainers: []*corev1.Maintainer{
					{Name: "stefanprodan", Email: "stefanprodan@users.noreply.github.com"},
				},
				Readme:        "Podinfo is a tiny web application made with Go",
				DefaultValues: "Default values for podinfo.\n\nreplicaCount: 1\n",
			},
		}
	}

	expected_detail_oci_stefanprodan_podinfo_2 = func(name, url string) *corev1.GetAvailablePackageDetailResponse {
		return &corev1.GetAvailablePackageDetailResponse{
			AvailablePackageDetail: &corev1.AvailablePackageDetail{
				AvailablePackageRef: availableRef(name+"/podinfo", "default"),
				Name:                "podinfo",
				Version:             pkgAppVersion("6.1.6"),
				RepoUrl:             url,
				HomeUrl:             "https://github.com/stefanprodan/podinfo",
				DisplayName:         "podinfo",
				ShortDescription:    "Podinfo Helm chart for Kubernetes",
				SourceUrls:          []string{"https://github.com/stefanprodan/podinfo"},
				Maintainers: []*corev1.Maintainer{
					{Name: "stefanprodan", Email: "stefanprodan@users.noreply.github.com"},
				},
				Readme:        "Podinfo is a tiny web application made with Go",
				DefaultValues: "Default values for podinfo.\n\nreplicaCount: 1\n",
			},
		}
	}

	newFakeRemoteOciRegistryData_1 = func() (*fakeRemoteOciRegistryData, error) {
		chartBytes, err := os.ReadFile(testTgz("podinfo-6.1.5.tgz"))
		if err != nil {
			return nil, err
		}
		return &fakeRemoteOciRegistryData{
			repositories: []fakeRepo{
				{
					name: "podinfo",
					chart: fakeChart{
						versions: []fakeChartVersion{
							{
								version:  "6.1.5",
								tgzBytes: chartBytes,
							},
						},
					},
				},
			},
		}, nil
	}

	oci_repo_available_package_summaries = []*corev1.AvailablePackageSummary{
		{
			Name:        "podinfo",
			DisplayName: "podinfo",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "6.1.5",
			},
			AvailablePackageRef: availableRef("repo-1/podinfo", "namespace-1"),
			Categories:          []string{""},
			ShortDescription:    "Podinfo Helm chart for Kubernetes",
		},
	}

	oci_repo_available_package_summaries_2 = []*corev1.AvailablePackageSummary{
		{
			Name:        "podinfo",
			DisplayName: "podinfo",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "6.1.5",
			},
			AvailablePackageRef: availableRef("repo-1/podinfo", "namespace-1"),
			Categories:          []string{""},
			ShortDescription:    "Podinfo Helm chart for Kubernetes",
		},
		{
			Name:        "airflow",
			DisplayName: "airflow",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "6.7.1",
			},
			IconUrl:             "https://bitnami.com/assets/stacks/airflow/img/airflow-stack-110x117.png",
			AvailablePackageRef: availableRef("repo-1/airflow", "namespace-1"),
			Categories:          []string{"WorkFlow"},
			ShortDescription:    "Apache Airflow is a platform to programmatically author, schedule and monitor workflows.",
		},
	}

	oci_podinfo_charts_spec = []testSpecChartWithUrl{
		{
			chartID:       "repo-1/podinfo",
			chartUrl:      "oci://localhost:54321/userX/charts/podinfo:6.1.5",
			chartRevision: "6.1.5",
			repoNamespace: "namespace-1",
		},
	}

	oci_podinfo_charts_spec_2 = []testSpecChartWithUrl{
		{
			chartID:       "repo-1/podinfo",
			chartUrl:      "oci://localhost:54321/userX/charts/podinfo:6.1.5",
			chartRevision: "6.1.5",
			repoNamespace: "namespace-1",
		},
		{
			chartID:       "repo-1/podinfo",
			chartUrl:      "oci://localhost:54321/userX/charts/podinfo:6.0.3",
			chartRevision: "6.0.3",
			repoNamespace: "namespace-1",
		},
		{
			chartID:       "repo-1/podinfo",
			chartUrl:      "oci://localhost:54321/userX/charts/podinfo:6.0.0",
			chartRevision: "6.0.0",
			repoNamespace: "namespace-1",
		},
	}

	expected_detail_podinfo_1 = &corev1.AvailablePackageDetail{
		AvailablePackageRef: availableRef("repo-1/podinfo", "namespace-1"),
		Name:                "podinfo",
		Version: &corev1.PackageAppVersion{
			PkgVersion: "6.1.5",
			AppVersion: "6.1.5",
		},
		RepoUrl:          "oci://localhost:54321/userX/charts",
		HomeUrl:          "https://github.com/stefanprodan/podinfo",
		DisplayName:      "podinfo",
		ShortDescription: "Podinfo Helm chart for Kubernetes",
		Readme:           "Podinfo is a tiny web application made with Go",
		DefaultValues:    "Default values for podinfo.",
		SourceUrls:       []string{"https://github.com/stefanprodan/podinfo"},
		Maintainers: []*corev1.Maintainer{
			{
				Name:  "stefanprodan",
				Email: "stefanprodan@users.noreply.github.com",
			},
		},
	}

	newFakeRemoteOciRegistryData_2 = func() (*fakeRemoteOciRegistryData, error) {
		chartBytes1, err := os.ReadFile(testTgz("podinfo-6.1.5.tgz"))
		if err != nil {
			return nil, err
		}
		chartBytes2, err := os.ReadFile(testTgz("podinfo-6.0.0.tgz"))
		if err != nil {
			return nil, err
		}
		chartBytes3, err := os.ReadFile(testTgz("podinfo-6.0.3.tgz"))
		if err != nil {
			return nil, err
		}
		return &fakeRemoteOciRegistryData{
			repositories: []fakeRepo{
				{
					name: "podinfo",
					chart: fakeChart{
						versions: []fakeChartVersion{
							{
								version:  "6.1.5",
								tgzBytes: chartBytes1,
							},
							{
								version:  "6.0.0",
								tgzBytes: chartBytes2,
							},
							{
								version:  "6.0.3",
								tgzBytes: chartBytes3,
							},
						},
					},
				},
			},
		}, nil
	}

	newFakeRemoteOciRegistryData_3 = func() (*fakeRemoteOciRegistryData, error) {
		chartBytes1, err := os.ReadFile(testTgz("podinfo-6.1.5.tgz"))
		if err != nil {
			return nil, err
		}
		chartBytes2, err := os.ReadFile(testTgz("airflow-6.7.1.tgz"))
		if err != nil {
			return nil, err
		}
		return &fakeRemoteOciRegistryData{
			repositories: []fakeRepo{
				{
					name: "podinfo",
					chart: fakeChart{
						versions: []fakeChartVersion{
							{
								version:  "6.1.5",
								tgzBytes: chartBytes1,
							},
						},
					},
				},
				{
					name: "airflow",
					chart: fakeChart{
						versions: []fakeChartVersion{
							{
								version:  "6.7.1",
								tgzBytes: chartBytes2,
							},
						},
					},
				},
			},
		}, nil
	}
)
