// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// global vars
// why define these here? see https://github.com/kubeapps/kubeapps/pull/3736#discussion_r745246398
// plus I am putting them in a separate file, since they take up so much space they distract from
// overall test logic
var (
	create_request_basic = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-1/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			// note that Namespace is just the prefix - the actual name will
			// have a random string appended at the end, e.g. "test-1-h23r"
			// this will happen during the running of the test
			Namespace: "test-1",
			Cluster:   KubeappsCluster,
		},
	}

	// specify just the fields that cannot be easily computed based on the request
	expected_detail_basic = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
	}

	expected_resource_refs_basic = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo",
		},
	}

	create_request_semver_constraint = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-2/podinfo", "default"),
		Name:                "my-podinfo-2",
		TargetContext: &corev1.Context{
			Namespace: "test-2",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
	}

	expected_detail_semver_constraint = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "> 5",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-2 8080:9898\n",
	}

	expected_resource_refs_semver_constraint = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-2",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-2",
		},
	}

	create_request_reconcile_options = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-3/podinfo", "default"),
		Name:                "my-podinfo-3",
		TargetContext: &corev1.Context{
			Namespace: "test-3",
			Cluster:   KubeappsCluster,
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            false,
			ServiceAccountName: "foo",
		},
	}

	expected_detail_reconcile_options = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval:           60,
			Suspend:            false,
			ServiceAccountName: "foo",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-3 8080:9898\n",
	}

	expected_resource_refs_reconcile_options = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-3",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-3",
		},
	}

	create_request_with_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-4/podinfo", "default"),
		Name:                "my-podinfo-4",
		TargetContext: &corev1.Context{
			Namespace: "test-4",
			Cluster:   KubeappsCluster,
		},
		Values: "{\"ui\": { \"message\": \"what we do in the shadows\" } }",
	}

	expected_detail_with_values = &corev1.InstalledPackageDetail{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-4 8080:9898\n",
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_resource_refs_with_values = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-4",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-4",
		},
	}

	create_request_install_fails = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-5/podinfo", "default"),
		Name:                "my-podinfo-5",
		TargetContext: &corev1.Context{
			Namespace: "test-5",
			Cluster:   KubeappsCluster,
		},
		Values: "{\"replicaCount\": \"what we do in the shadows\" }",
	}

	expected_detail_install_fails = &corev1.InstalledPackageDetail{
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

	create_request_podinfo_5_2_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-6/podinfo", "default"),
		Name:                "my-podinfo-6",
		TargetContext: &corev1.Context{
			Namespace: "test-6",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-6 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-6",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-6",
		},
	}

	expected_detail_podinfo_6_0_0 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "6.0.0",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-6 8080:9898\n",
	}

	create_request_podinfo_5_2_1_no_values = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-7/podinfo", "default"),
		Name:                "my-podinfo-7",
		TargetContext: &corev1.Context{
			Namespace: "test-7",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_5_2_1_no_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-7 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_no_values = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-7",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-7",
		},
	}

	expected_detail_podinfo_5_2_1_values = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-7 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-8/podinfo", "default"),
		Name:                "my-podinfo-8",
		TargetContext: &corev1.Context{
			Namespace: "test-8",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-8 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_2 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-8",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-8",
		},
	}

	expected_detail_podinfo_5_2_1_values_3 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"Le Bureau des Légendes\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-8 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_4 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-9/podinfo", "default"),
		Name:                "my-podinfo-9",
		TargetContext: &corev1.Context{
			Namespace: "test-9",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_4 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-9 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_4 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-9",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-9",
		},
	}

	expected_detail_podinfo_5_2_1_values_5 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-9 8080:9898\n",
	}

	create_request_podinfo_5_2_1_values_6 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-10/podinfo", "default"),
		Name:                "my-podinfo-10",
		TargetContext: &corev1.Context{
			Namespace: "test-10",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		Values: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
	}

	expected_detail_podinfo_5_2_1_values_6 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		ValuesApplied: "{\"ui\":{\"message\":\"what we do in the shadows\"}}",
		Status:        statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-10 8080:9898\n",
	}

	expected_resource_refs_podinfo_5_2_1_values_6 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-10",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-10",
		},
	}

	create_request_podinfo_7 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-11/podinfo", "default"),
		Name:                "my-podinfo-11",
		TargetContext: &corev1.Context{
			Namespace: "test-11",
			Cluster:   KubeappsCluster,
		},
	}

	expected_detail_podinfo_7 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-11 8080:9898\n",
	}

	expected_resource_refs_podinfo_7 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-11",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-11",
		},
	}

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

	create_request_podinfo_for_delete_1 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-12/podinfo", "default"),
		Name:                "my-podinfo-12",
		TargetContext: &corev1.Context{
			Namespace: "test-12",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_for_delete_1 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-12 8080:9898\n",
	}

	expected_resource_refs_for_delete_1 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-12",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-12",
		},
	}

	create_request_podinfo_for_delete_2 = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-13/podinfo", "default"),
		Name:                "my-podinfo-13",
		TargetContext: &corev1.Context{
			Namespace: "test-13",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
	}

	expected_detail_podinfo_for_delete_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "=5.2.1",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "5.2.1",
			AppVersion: "5.2.1",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-13 8080:9898\n",
	}

	expected_resource_refs_for_delete_2 = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-13",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-13",
		},
	}

	create_request_wrong_cluster = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-14/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-14",
			Cluster:   "this is not the cluster you're looking for",
		},
	}

	create_request_target_ns_doesnt_exist = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-15/podinfo", "default"),
		Name:                "my-podinfo",
		TargetContext: &corev1.Context{
			Namespace: "test-15",
			Cluster:   KubeappsCluster,
		},
	}

	create_request_auto_update = &corev1.CreateInstalledPackageRequest{
		AvailablePackageRef: availableRef("podinfo-16/podinfo", "default"),
		Name:                "my-podinfo-16",
		TargetContext: &corev1.Context{
			Namespace: "test-16",
			Cluster:   KubeappsCluster,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
	}

	expected_detail_auto_update = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-16 8080:9898\n",
	}

	expected_detail_auto_update_2 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: ">= 6",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.3",
			AppVersion: "6.0.3",
		},
		Name:   "my-podinfo-16",
		Status: statusInstalled,
		ReconciliationOptions: &corev1.ReconciliationOptions{
			Interval: 30,
		},
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   KubeappsCluster,
				Namespace: "default",
			},
			Identifier: "podinfo-16/podinfo",
			Plugin:     fluxPlugin,
		},
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo-16 8080:9898\n",
	}

	expected_resource_refs_auto_update = []*corev1.ResourceRef{
		{
			ApiVersion: "v1",
			Kind:       "Service",
			Name:       "my-podinfo-16",
		},
		{
			ApiVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "my-podinfo-16",
		},
	}

	expected_detail_test_release_rbac = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
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
				CurrentVersion: &corev1.PackageAppVersion{
					PkgVersion: "6.0.0",
					AppVersion: "6.0.0",
				},
				PkgDisplayName:   "podinfo",
				ShortDescription: "Podinfo Helm chart for Kubernetes",
				LatestVersion: &corev1.PackageAppVersion{
					PkgVersion: "6.0.0",
					AppVersion: "6.0.0",
				},
			},
		},
	}

	expected_detail_test_release_rbac_2 = &corev1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: &corev1.InstalledPackageDetail{
			InstalledPackageRef: installedRef("my-podinfo", "@TARGET_NS@"),
			PkgVersionReference: &corev1.VersionReference{Version: "*"},
			Name:                "my-podinfo",
			CurrentVersion: &corev1.PackageAppVersion{
				PkgVersion: "6.0.0",
				AppVersion: "6.0.0",
			},
			ReconciliationOptions: &corev1.ReconciliationOptions{
				Interval: 60,
			},
			Status: statusInstalled,
			PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
				"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
				"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
			AvailablePackageRef: availableRef("podinfo-1/podinfo", "@SOURCE_NS@"),
		},
	}

	expected_detail_test_release_rbac_3 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
	}

	expected_detail_test_release_rbac_4 = &corev1.InstalledPackageDetail{
		PkgVersionReference: &corev1.VersionReference{
			Version: "*",
		},
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: "6.0.0",
			AppVersion: "6.0.0",
		},
		Status: statusInstalled,
		PostInstallationNotes: "1. Get the application URL by running these commands:\n  " +
			"echo \"Visit http://127.0.0.1:8080 to use your application\"\n  " +
			"kubectl -n @TARGET_NS@ port-forward deploy/my-podinfo 8080:9898\n",
	}

	available_package_summaries_podinfo_basic_auth = &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
			{
				Name:                "podinfo",
				AvailablePackageRef: availableRef("podinfo-basic-auth/podinfo", "default"),
				LatestVersion:       &corev1.PackageAppVersion{PkgVersion: "6.0.0", AppVersion: "6.0.0"},
				DisplayName:         "podinfo",
				ShortDescription:    "Podinfo Helm chart for Kubernetes",
				Categories:          []string{""},
			},
		},
	}

	expected_detail_podinfo_basic_auth = &corev1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: &corev1.AvailablePackageDetail{
			AvailablePackageRef: availableRef("podinfo-basic-auth/podinfo", "default"),
			Name:                "podinfo",
			Version:             &corev1.PackageAppVersion{PkgVersion: "6.0.0", AppVersion: "6.0.0"},
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

	valid_index_package_summaries = []*corev1.AvailablePackageSummary{
		{
			Name:        "acs-engine-autoscaler",
			DisplayName: "acs-engine-autoscaler",
			LatestVersion: &corev1.PackageAppVersion{
				PkgVersion: "2.1.1",
				AppVersion: "2.1.1",
			},
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
		Name:        "cert-manager",
		DisplayName: "cert-manager",
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: "v1.4.0",
			AppVersion: "v1.4.0",
		},
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

	add_repo_expected_resp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: "foo",
				Cluster:   KubeappsCluster,
			},
			Identifier: "bar",
			Plugin:     fluxPlugin,
		},
	}

	statusInstalled = &corev1.InstalledPackageStatus{
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
		Status:           statusInstalled,
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
		Status:           statusInstalled,
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
		Status:           statusInstalled,
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
		Status:           statusInstalled,
	}

	lastTransitionTime, _ = time.Parse(time.RFC3339, "2021-08-11T08:46:03Z")

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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
					Type:               "Ready",
					Status:             "False",
					Reason:             helmv2.InstallFailedReason,
					Message:            "install retries exhausted",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "False",
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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
					Type:               "Ready",
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
					Type:               "Ready",
					Status:             "False",
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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
			Interval: 60,
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
			Interval: 60,
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
			Interval: 60,
		},
		Status:                statusInstalled,
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
			Interval:           60,
			Suspend:            true,
			ServiceAccountName: "foo",
		},
		Status:                statusInstalled,
		ValuesApplied:         "{\"replica\": { \"replicaCount\":  \"1\", \"configuration\": \"xyz\"    }}",
		AvailablePackageRef:   availableRef("bitnami-1/redis", "default"),
		PostInstallationNotes: "some notes",
	}

	flux_helm_release_basic = &helmv2.HelmRelease{
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
		},
	}

	flux_helm_release_semver_constraint = &helmv2.HelmRelease{
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
					Version: "> 5",
				},
			},
			Interval: metav1.Duration{Duration: 1 * time.Minute},
		},
	}

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
					Type:               "Ready",
					Status:             "True",
					Reason:             "ReconciliationSucceeded",
					Message:            "Release reconciliation succeeded",
				},
				{
					LastTransitionTime: metav1.Time{Time: lastTransitionTime},
					Type:               "Released",
					Status:             "True",
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
)
