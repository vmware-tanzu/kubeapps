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
package main

import (
	"fmt"
	"time"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirVersions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

func (s *Server) getAvailablePackageSummary(pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, cluster string) (*corev1.AvailablePackageSummary, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgMetadata.Name]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name)
	}

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.
	iconUrl := ""
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconUrl = fmt.Sprintf("data:image/svg+xml;base64,%s", pkgMetadata.Spec.IconSVGBase64)
	}

	availablePackageSummary := &corev1.AvailablePackageSummary{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgMetadata.Namespace,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgMetadata.Name,
		},
		Name: pkgMetadata.Name,
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
		},
		IconUrl:          iconUrl,
		DisplayName:      pkgMetadata.Spec.DisplayName,
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Categories:       pkgMetadata.Spec.Categories,
	}

	return availablePackageSummary, nil
}

func (s *Server) getAvailablePackageDetail(pkgMetadata *datapackagingv1alpha1.PackageMetadata, requestedPkgVersion string, foundPkgSemver *pkgSemver, cluster string) (*corev1.AvailablePackageDetail, error) {

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.
	iconUrl := ""
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconUrl = fmt.Sprintf("data:image/svg+xml;base64,%s", pkgMetadata.Spec.IconSVGBase64)
	}

	maintainers := []*corev1.Maintainer{}
	for _, maintainer := range pkgMetadata.Spec.Maintainers {
		maintainers = append(maintainers, &corev1.Maintainer{
			Name: maintainer.Name,
		})
	}

	readme := fmt.Sprintf(`## Details


### Capactiy requirements:
%s


### Release Notes:
%s


### Licenses:
%s


### ReleasedAt:
%s


`,
		foundPkgSemver.pkg.Spec.CapactiyRequirementsDescription,
		foundPkgSemver.pkg.Spec.ReleaseNotes,
		foundPkgSemver.pkg.Spec.Licenses,
		foundPkgSemver.pkg.Spec.ReleasedAt,
	)
	availablePackageDetail := &corev1.AvailablePackageDetail{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgMetadata.Namespace,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgMetadata.Name,
		},
		Name:             pkgMetadata.Name,
		IconUrl:          iconUrl,
		DisplayName:      pkgMetadata.Spec.DisplayName,
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Categories:       pkgMetadata.Spec.Categories,
		LongDescription:  pkgMetadata.Spec.LongDescription,
		Version: &corev1.PackageAppVersion{
			PkgVersion: requestedPkgVersion,
		},
		Maintainers: maintainers,
		Readme:      readme,

		// TODO(agamez): we might need to have a default value (from the openapi schema?)
		// and/or perform some changes in the UI
		// DefaultValues: "",

		// TODO(agamez): pkgs have an OpenAPI Schema object,
		// but currently we aren't able to parse it from the UI
		// ValuesSchema: foundPkgSemver.pkg.Spec.ValuesSchema.String(),

		// TODO(agamez): fields 'HomeUrl','RepoUrl' are not being populated right now,
		// but some fields (eg, release notes) have URLs (but not sure if in every pkg also happens)
		// HomeUrl: "",
		// RepoUrl:  "",

	}

	return availablePackageDetail, nil
}

func (s *Server) getInstalledPackageSummary(pkgInstall *packagingv1alpha1.PackageInstall, pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, cluster string) (*corev1.InstalledPackageSummary, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgInstall.Spec.PackageRef.RefName]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgInstall.Spec.PackageRef.RefName)
	}

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.
	iconUrl := ""
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconUrl = fmt.Sprintf("data:image/svg+xml;base64,%s", pkgMetadata.Spec.IconSVGBase64)
	}

	installedPackageSummary := &corev1.InstalledPackageSummary{
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: pkgInstall.Status.LastAttemptedVersion,
		},
		IconUrl: iconUrl,
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: pkgMetadata.Namespace,
				Cluster:   cluster,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgInstall.Name,
		},
		LatestMatchingVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
		},
		Name:           pkgInstall.Name,
		PkgDisplayName: pkgMetadata.Spec.DisplayName,
		PkgVersionReference: &corev1.VersionReference{
			Version: pkgInstall.Status.LastAttemptedVersion,
		},
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Status: &corev1.InstalledPackageStatus{
			Ready:      false,
			Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
			UserReason: "no status information yet",
		},
	}
	if len(pkgInstall.Status.Conditions) > 0 {
		installedPackageSummary.Status = &corev1.InstalledPackageStatus{
			Ready:      pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
			UserReason: userReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
		}
	}

	return installedPackageSummary, nil
}

func (s *Server) getInstalledPackageDetail(pkgInstall *packagingv1alpha1.PackageInstall, pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, app *kappctrlv1alpha1.App, valuesApplied, cluster string) (*corev1.InstalledPackageDetail, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgMetadata.Name]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name)
	}

	deployStdout := ""
	deployStderr := ""
	fetchStdout := ""
	fetchStderr := ""
	inspectStdout := ""
	inspectStderr := ""

	if app.Status.Deploy != nil {
		deployStdout = app.Status.Deploy.Stdout
		deployStderr = app.Status.Deploy.Stderr
	}
	if app.Status.Fetch != nil {
		fetchStdout = app.Status.Fetch.Stdout
		fetchStderr = app.Status.Fetch.Stderr
	}
	if app.Status.Inspect != nil {
		inspectStdout = app.Status.Inspect.Stdout
		inspectStderr = app.Status.Inspect.Stderr
	}

	// Build some custom installation notes based on the available stdout + stderr
	// TODO(agamez): this is just a temporary solution until come up with a better UX solution
	// short-term improvement is to just display those values != ""
	postInstallationNotes := fmt.Sprintf(`## Installation output


### Deploy:
%s


### Fetch:
%s


### Inspect:
%s


## Errors


### Deploy:
%s


### Fetch:
%s


### Inspect:
%s

`, deployStdout, fetchStdout, inspectStdout, deployStderr, fetchStderr, inspectStderr)

	if len(pkgInstall.Status.Conditions) > 1 {
		log.Warningf("The package install %s has more than one status conditions. Using the first one: %s", pkgInstall.Name, pkgInstall.Status.Conditions[0])
	}

	installedPackageDetail := &corev1.InstalledPackageDetail{
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: pkgMetadata.Namespace,
				Cluster:   cluster,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgInstall.Name,
		},
		PkgVersionReference: &corev1.VersionReference{
			Version: pkgInstall.Status.LastAttemptedVersion,
		},
		Name: pkgInstall.Name,
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: pkgInstall.Status.LastAttemptedVersion,
			AppVersion: pkgInstall.Status.LastAttemptedVersion,
		},
		ValuesApplied: valuesApplied,

		ReconciliationOptions: &corev1.ReconciliationOptions{
			ServiceAccountName: pkgInstall.Spec.ServiceAccountName,
		},
		PostInstallationNotes: postInstallationNotes,
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Namespace: pkgMetadata.Namespace,
				Cluster:   cluster,
			},
			Identifier: pkgInstall.Spec.PackageRef.RefName,
			Plugin:     &pluginDetail,
		},
		LatestMatchingVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
			AppVersion: versions[0].version.String(),
		},
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
			AppVersion: versions[0].version.String(),
		},
	}

	// Some fields would require an extra nil check before being populated
	if app.Spec.SyncPeriod != nil {
		installedPackageDetail.ReconciliationOptions.Interval = int32(app.Spec.SyncPeriod.Seconds())
	}

	if pkgInstall.Status.Conditions != nil && &pkgInstall.Status.Conditions[0].Type != nil {
		installedPackageDetail.Status = &corev1.InstalledPackageStatus{
			Ready:      pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
			UserReason: userReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
		}
		installedPackageDetail.ReconciliationOptions.Suspend = pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.Reconciling
	}

	return installedPackageDetail, nil
}

func (s *Server) newSecret(installedPackageName, values, targetNamespace string) (*k8scorev1.Secret, error) {
	return &k8scorev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       k8scorev1.ResourceSecrets.String(),
			APIVersion: k8scorev1.SchemeGroupVersion.WithResource(k8scorev1.ResourceSecrets.String()).String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			// TODO(agamez): think about name collisions
			Name:      fmt.Sprintf("%s-values", installedPackageName),
			Namespace: targetNamespace,
		},
		Data: map[string][]byte{
			// TODO(agamez): check the actual value for the key.
			// Assuming "values.yaml" perhaps is not always true.
			// Perhaos this info is in the "package" object?
			"values.yaml": []byte(values),
		},
		Type: "Opaque",
	}, nil
}

func (s *Server) newPkgInstall(installedPackageName, targetCluster, targetNamespace, packageRefName, pkgVersion string, reconciliationOptions *corev1.ReconciliationOptions) (*packagingv1alpha1.PackageInstall, error) {
	pkgInstall := &packagingv1alpha1.PackageInstall{
		TypeMeta: metav1.TypeMeta{
			Kind:       pkgInstallResource,
			APIVersion: fmt.Sprintf("%s/%s", packagingv1alpha1.SchemeGroupVersion.Group, packagingv1alpha1.SchemeGroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      installedPackageName,
			Namespace: targetNamespace,
		},
		Spec: packagingv1alpha1.PackageInstallSpec{
			// TODO(agamez): remove this when we have a real service account in the UI
			ServiceAccountName: "default",
			// ServiceAccountName: "default-ns-sa",
			// TODO(agamez): check what this "Cluster" is for
			// Cluster: &kappctrlv1alpha1.AppCluster{
			// 	Namespace:           targetNamespace,
			// 	KubeconfigSecretRef: &kappctrlv1alpha1.AppClusterKubeconfigSecretRef{},
			// },
			Values: []packagingv1alpha1.PackageInstallValues{
				{
					SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
						Name: fmt.Sprintf("%s-values", installedPackageName),
						Key:  "values.yaml",
					},
				},
			},
			PackageRef: &packagingv1alpha1.PackageRef{
				RefName: packageRefName,
				VersionSelection: &vendirVersions.VersionSelectionSemver{
					Constraints: pkgVersion,
					// https://github.com/vmware-tanzu/carvel-kapp-controller/issues/116
					// This is to allow prereleases to be also installed
					Prereleases: &vendirVersions.VersionSelectionSemverPrereleases{},
				},
			},
		},
	}

	if reconciliationOptions != nil {
		if reconciliationOptions.Interval > 0 {
			pkgInstall.Spec.SyncPeriod = &metav1.Duration{
				Duration: time.Duration(reconciliationOptions.Interval) * time.Second,
			}
		}
		pkgInstall.Spec.ServiceAccountName = reconciliationOptions.ServiceAccountName
		pkgInstall.Spec.Paused = reconciliationOptions.Suspend
	}
	return pkgInstall, nil
}
