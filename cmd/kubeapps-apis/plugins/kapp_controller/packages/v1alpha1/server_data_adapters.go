// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"
	"time"

	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	kappctrlpackageinstall "github.com/vmware-tanzu/carvel-kapp-controller/pkg/packageinstall"
	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

const (
	Type_Inline       = "inline"
	Type_Image        = "image"
	Type_ImgPkgBundle = "imgpkgBundle"
	Type_HTTP         = "http"
	Type_GIT          = "git"
)

// available packages

func (s *Server) buildAvailablePackageSummary(pkgMetadata *datapackagingv1alpha1.PackageMetadata, latestVersion string, cluster string) *corev1.AvailablePackageSummary {
	var iconStringBuilder strings.Builder

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.

	// TODO(agamez): check if want to avoid sending this data over the wire
	// instead we could send a url (to another API endpoint) to retrieve the icon
	// See: https://github.com/vmware-tanzu/kubeapps/pull/3787#discussion_r754741255
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconStringBuilder.WriteString("data:image/svg+xml;base64,")
		iconStringBuilder.WriteString(pkgMetadata.Spec.IconSVGBase64)
	}

	// build package identifier based on the metadata
	identifier := buildPackageIdentifier(pkgMetadata)

	availablePackageSummary := &corev1.AvailablePackageSummary{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgMetadata.Namespace,
			},
			Plugin:     &pluginDetail,
			Identifier: identifier,
		},
		Name: pkgMetadata.Name,
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: latestVersion,
			AppVersion: latestVersion,
		},
		IconUrl:          iconStringBuilder.String(),
		DisplayName:      pkgMetadata.Spec.DisplayName,
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Categories:       pkgMetadata.Spec.Categories,
	}

	return availablePackageSummary
}

func (s *Server) buildAvailablePackageDetail(pkgMetadata *datapackagingv1alpha1.PackageMetadata, requestedPkgVersion string, foundPkgSemver *pkgSemver, cluster string) (*corev1.AvailablePackageDetail, error) {

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.

	// TODO(agamez): check if want to avoid sending this data over the wire
	// instead we could send a url (to another API endpoint) to retrieve the icon
	// See: https://github.com/vmware-tanzu/kubeapps/pull/3787#discussion_r754741255
	var iconStringBuilder strings.Builder
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconStringBuilder.WriteString("data:image/svg+xml;base64,")
		iconStringBuilder.WriteString(pkgMetadata.Spec.IconSVGBase64)
	}

	// build maintainers information
	maintainers := []*corev1.Maintainer{}
	for _, maintainer := range pkgMetadata.Spec.Maintainers {
		maintainers = append(maintainers, &corev1.Maintainer{
			Name: maintainer.Name,
		})
	}

	// build package identifier based on the metadata
	identifier := buildPackageIdentifier(pkgMetadata)

	// build readme
	readme := buildReadme(pkgMetadata, foundPkgSemver)

	// build default values
	defaultValues, err := pkgutils.DefaultValuesFromSchema(foundPkgSemver.pkg.Spec.ValuesSchema.OpenAPIv3.Raw, true)
	if err != nil {
		log.Warningf("Failed to parse default values from schema: %v", err)
		defaultValues = "# There is an error while parsing the schema."
	}

	availablePackageDetail := &corev1.AvailablePackageDetail{
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgMetadata.Namespace,
			},
			Plugin:     &pluginDetail,
			Identifier: identifier,
		},
		Name:             pkgMetadata.Name,
		IconUrl:          iconStringBuilder.String(),
		DisplayName:      pkgMetadata.Spec.DisplayName,
		ShortDescription: pkgMetadata.Spec.ShortDescription,
		Categories:       pkgMetadata.Spec.Categories,
		LongDescription:  pkgMetadata.Spec.LongDescription,
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		Version: &corev1.PackageAppVersion{
			PkgVersion: requestedPkgVersion,
			AppVersion: requestedPkgVersion,
		},
		Maintainers:   maintainers,
		Readme:        readme,
		ValuesSchema:  string(foundPkgSemver.pkg.Spec.ValuesSchema.OpenAPIv3.Raw),
		DefaultValues: defaultValues,
		// TODO(agamez): fields 'HomeUrl','RepoUrl' are not being populated right now,
		// but some fields (eg, release notes) have URLs (but not sure if in every pkg also happens)
		// HomeUrl: "",
		// RepoUrl:  "",
	}
	return availablePackageDetail, nil
}

// installed packages

func (s *Server) buildInstalledPackageSummary(pkgInstall *packagingv1alpha1.PackageInstall, pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, cluster string) (*corev1.InstalledPackageSummary, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgInstall.Spec.PackageRef.RefName]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgInstall.Spec.PackageRef.RefName)
	}

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.

	// TODO(agamez): check if want to avoid sending this data over the wire
	// instead we could send a url (to another API endpoint) to retrieve the icon
	// See: https://github.com/vmware-tanzu/kubeapps/pull/3787#discussion_r754741255
	var iconStringBuilder strings.Builder
	if pkgMetadata.Spec.IconSVGBase64 != "" {
		iconStringBuilder.WriteString("data:image/svg+xml;base64,")
		iconStringBuilder.WriteString(pkgMetadata.Spec.IconSVGBase64)
	}

	latestMatchingVersion, err := latestMatchingVersion(versions, pkgInstall.Spec.PackageRef.VersionSelection.Constraints)
	if err != nil {
		return nil, fmt.Errorf("Cannot get the latest matching version for the pkg %q: %s", pkgMetadata.Name, err.Error())
	}

	installedPackageSummary := &corev1.InstalledPackageSummary{
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		CurrentVersion: &corev1.PackageAppVersion{
			PkgVersion: pkgInstall.Status.LastAttemptedVersion,
			AppVersion: pkgInstall.Status.LastAttemptedVersion,
		},
		IconUrl: iconStringBuilder.String(),
		InstalledPackageRef: &corev1.InstalledPackageReference{
			Context: &corev1.Context{
				Namespace: pkgInstall.Namespace,
				Cluster:   cluster,
			},
			Plugin:     &pluginDetail,
			Identifier: pkgInstall.Name,
		},
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
			AppVersion: versions[0].version.String(),
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
			UserReason: simpleUserReasonForKappStatus(""),
		},
	}

	if latestMatchingVersion != nil {
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		installedPackageSummary.LatestMatchingVersion = &corev1.PackageAppVersion{
			PkgVersion: latestMatchingVersion.String(),
			AppVersion: latestMatchingVersion.String(),
		}
	}

	if len(pkgInstall.Status.Conditions) > 0 {
		installedPackageSummary.Status = &corev1.InstalledPackageStatus{
			Ready:      pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
			UserReason: simpleUserReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
		}
	}

	return installedPackageSummary, nil
}

func (s *Server) buildInstalledPackageDetail(pkgInstall *packagingv1alpha1.PackageInstall, pkgMetadata *datapackagingv1alpha1.PackageMetadata, pkgVersionsMap map[string][]pkgSemver, app *kappctrlv1alpha1.App, valuesApplied, cluster string) (*corev1.InstalledPackageDetail, error) {
	// get the versions associated with the package
	versions := pkgVersionsMap[pkgMetadata.Name]
	if len(versions) == 0 {
		return nil, fmt.Errorf("no package versions for the package %q", pkgMetadata.Name)
	}

	// build postInstallationNotes
	postInstallationNotes := buildPostInstallationNotes(app)

	if len(pkgInstall.Status.Conditions) > 1 {
		log.Warningf("The package install %s has more than one status conditions. Using the first one: %s", pkgInstall.Name, pkgInstall.Status.Conditions[0])
	}

	latestMatchingVersion, err := latestMatchingVersion(versions, pkgInstall.Spec.PackageRef.VersionSelection.Constraints)
	if err != nil {
		return nil, fmt.Errorf("Cannot get the latest matching version for the pkg %q: %s", pkgMetadata.Name, err.Error())
	}

	// build package availablePackageIdentifier based on the metadata
	availablePackageIdentifier := buildPackageIdentifier(pkgMetadata)

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
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
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
			Plugin:     &pluginDetail,
			Identifier: availablePackageIdentifier,
		},
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		LatestVersion: &corev1.PackageAppVersion{
			PkgVersion: versions[0].version.String(),
			AppVersion: versions[0].version.String(),
		},
	}

	if latestMatchingVersion != nil {
		// Currently, PkgVersion and AppVersion are the same
		// https://kubernetes.slack.com/archives/CH8KCCKA5/p1636386358322000?thread_ts=1636371493.320900&cid=CH8KCCKA5
		installedPackageDetail.LatestMatchingVersion = &corev1.PackageAppVersion{
			PkgVersion: latestMatchingVersion.String(),
			AppVersion: latestMatchingVersion.String(),
		}
	}

	// Some fields would require an extra nil check before being populated
	if app.Spec.SyncPeriod != nil {
		installedPackageDetail.ReconciliationOptions.Interval = int32(app.Spec.SyncPeriod.Seconds())
	}

	if pkgInstall.Status.Conditions != nil && len(pkgInstall.Status.Conditions) > 0 {
		installedPackageDetail.Status = &corev1.InstalledPackageStatus{
			Ready:      pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReasonForKappStatus(pkgInstall.Status.Conditions[0].Type),
			UserReason: pkgInstall.Status.UsefulErrorMessage, // long message, instead of the simpleUserReasonForKappStatus
		}
		installedPackageDetail.ReconciliationOptions.Suspend = pkgInstall.Status.Conditions[0].Type == kappctrlv1alpha1.Reconciling
	}

	return installedPackageDetail, nil
}

func (s *Server) buildSecret(installedPackageName, values, targetNamespace string) (*k8scorev1.Secret, error) {
	// Using this pattern as per:
	// https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.36.1/cli/pkg/kctrl/cmd/package/installed/created_resource_annotations.go#L19
	kappctrlSecretName := "%s-%s-values"

	return &k8scorev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       k8scorev1.ResourceSecrets.String(),
			APIVersion: k8scorev1.SchemeGroupVersion.WithResource(k8scorev1.ResourceSecrets.String()).String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf(kappctrlSecretName, installedPackageName, targetNamespace),
			Namespace: targetNamespace,
		},
		Data: map[string][]byte{
			// Using "values.yaml" as per:
			// https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.32.0/cli/pkg/kctrl/cmd/package/installed/create_or_update.go#L32
			"values.yaml": []byte(values),
		},
		Type: "Opaque",
	}, nil
}

func (s *Server) buildPkgInstall(installedPackageName, targetCluster, targetNamespace, packageRefName, pkgVersion string, reconciliationOptions *corev1.ReconciliationOptions, secret *k8scorev1.Secret) (*packagingv1alpha1.PackageInstall, error) {
	// Calculate the constraints and prerelease fields
	versionConstraints, err := pkgutils.VersionConstraintWithUpgradePolicy(pkgVersion, s.pluginConfig.defaultUpgradePolicy)
	if err != nil {
		return nil, err
	}
	prereleases := prereleasesVersionSelection(s.pluginConfig.defaultPrereleasesVersionSelection)

	versionSelection := &vendirversions.VersionSelectionSemver{
		Constraints: versionConstraints,
		Prereleases: prereleases,
	}

	// Ensure the selected version can be, actually installed to let the user know before installing
	elegibleVersion, err := versions.HighestConstrainedVersion([]string{pkgVersion}, vendirversions.VersionSelection{Semver: versionSelection})
	if elegibleVersion == "" || err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "The selected version %q is not elegible to be installed: %v", pkgVersion, err)
	}

	pkgInstall := &packagingv1alpha1.PackageInstall{
		TypeMeta: metav1.TypeMeta{
			Kind:       pkgInstallResource,
			APIVersion: fmt.Sprintf("%s/%s", packagingv1alpha1.SchemeGroupVersion.Group, packagingv1alpha1.SchemeGroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        installedPackageName,
			Namespace:   targetNamespace,
			Annotations: map[string]string{},
		},
		Spec: packagingv1alpha1.PackageInstallSpec{
			// This is the Carvel's way of supporting deployments across clusters
			// without having kapp-controller on those other clusters
			// We, currently, don't support deploying to another cluster without kapp-controller
			// See https://github.com/vmware-tanzu/kubeapps/pull/3789#discussion_r754786633
			// Cluster: &kappctrlv1alpha1.AppCluster{
			// 	Namespace:           targetNamespace,
			// 	KubeconfigSecretRef: &kappctrlv1alpha1.AppClusterKubeconfigSecretRef{},
			// },
			PackageRef: &packagingv1alpha1.PackageRef{
				RefName:          packageRefName,
				VersionSelection: versionSelection,
			},
		},
	}

	// Allow this PackageInstall to be downgraded
	// https://carvel.dev/kapp-controller/docs/v0.32.0/package-consumer-concepts/#downgrading
	if s.pluginConfig.defaultAllowDowngrades {
		pkgInstall.ObjectMeta.Annotations[kappctrlpackageinstall.DowngradableAnnKey] = ""
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

	if secret != nil {
		// Similar logic as in https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.32.0/cli/pkg/kctrl/cmd/package/installed/create_or_update.go#L505
		pkgInstall.Spec.Values = []packagingv1alpha1.PackageInstallValues{{
			SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
				// The secret name should have the format: <name>-<namespace> as per:
				// https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.32.0/cli/pkg/kctrl/cmd/package/installed/created_resource_annotations.go#L19
				Name: secret.Name,
			},
		}}
	}

	return pkgInstall, nil
}

// package repositories

func (s *Server) buildPackageRepositorySummary(pr *packagingv1alpha1.PackageRepository, cluster string) (*corev1.PackageRepositorySummary, error) {

	// base struct
	repository := &corev1.PackageRepositorySummary{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pr.Namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: pr.Name,
		},
		Name:            pr.Name,
		NamespaceScoped: s.globalPackagingNamespace == pr.Namespace,
	}

	// handle fetch-specific configuration
	if pr.Spec.Fetch.Inline != nil {
		repository.Type = Type_Inline
	} else if pr.Spec.Fetch.Image != nil {
		repository.Type = Type_Image
		repository.Url = pr.Spec.Fetch.Image.URL
	} else if pr.Spec.Fetch.ImgpkgBundle != nil {
		repository.Type = Type_ImgPkgBundle
		repository.Url = pr.Spec.Fetch.ImgpkgBundle.Image
	} else if pr.Spec.Fetch.HTTP != nil {
		repository.Type = Type_HTTP
		repository.Url = pr.Spec.Fetch.HTTP.URL
	} else if pr.Spec.Fetch.Git != nil {
		repository.Type = Type_GIT
		repository.Url = pr.Spec.Fetch.Git.URL
	} else {
		return nil, fmt.Errorf("the package repository has a fetch directive that is not supported")
	}

	// extract status
	if len(pr.Status.Conditions) > 0 {
		repository.Status = &corev1.PackageRepositoryStatus{
			Ready:      pr.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReason(pr.Status.Conditions[0]),
			UserReason: statusUserReason(pr.Status.Conditions[0], pr.Status.UsefulErrorMessage),
		}
	}

	// result
	return repository, nil
}

func (s *Server) buildPackageRepository(pr *packagingv1alpha1.PackageRepository, cluster string) (*corev1.PackageRepositoryDetail, error) {

	// base struct
	repository := &corev1.PackageRepositoryDetail{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pr.Namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: pr.Name,
		},
		Name:            pr.Name,
		NamespaceScoped: s.globalPackagingNamespace == pr.Namespace,
	}

	// synchronization
	if pr.Spec.SyncPeriod != nil {
		repository.Interval = uint32(pr.Spec.SyncPeriod.Seconds())
	}

	// handle fetch-specific configuration (todo -> handle custom configuration)
	fetch := pr.Spec.Fetch
	if fetch.Inline != nil {
		repository.Type = Type_Inline
	} else if fetch.Image != nil {
		repository.Type = Type_Image
		repository.Url = fetch.Image.URL
	} else if fetch.ImgpkgBundle != nil {
		repository.Type = Type_ImgPkgBundle
		repository.Url = pr.Spec.Fetch.ImgpkgBundle.Image
	} else if fetch.HTTP != nil {
		repository.Type = Type_HTTP
		repository.Url = fetch.HTTP.URL
	} else if fetch.Git != nil {
		repository.Type = Type_GIT
		repository.Url = fetch.Git.URL
	} else {
		return nil, fmt.Errorf("the package repository has a fetch directive that is not supported")
	}

	// auth

	// extract status
	if len(pr.Status.Conditions) > 0 {
		repository.Status = &corev1.PackageRepositoryStatus{
			Ready:      pr.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReason(pr.Status.Conditions[0]),
			UserReason: statusUserReason(pr.Status.Conditions[0], pr.Status.UsefulErrorMessage),
		}
	}

	// result
	return repository, nil
}

func (s *Server) buildPkgRepositoryCreate(request *corev1.AddPackageRepositoryRequest) (*packagingv1alpha1.PackageRepository, error) {
	// identifier
	namespace := request.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.globalPackagingNamespace
	}
	name := request.GetName()

	// repository stub
	repository := &packagingv1alpha1.PackageRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       pkgRepositoryResource,
			APIVersion: fmt.Sprintf("%s/%s", packagingv1alpha1.SchemeGroupVersion.Group, packagingv1alpha1.SchemeGroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   namespace,
			Annotations: map[string]string{},
		},
		Spec: packagingv1alpha1.PackageRepositorySpec{},
	}

	// synchronization
	interval := request.GetInterval()
	if interval > 0 {
		repository.Spec.SyncPeriod = &metav1.Duration{Duration: time.Duration(interval) * time.Second}
	}

	// fetch (todo -> add support for other directives than just imgkpg)
	repository.Spec.Fetch = &packagingv1alpha1.PackageRepositoryFetch{
		ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
			Image: request.GetUrl(),
		},
	}

	return repository, nil
}

func (s *Server) buildPkgRepositoryUpdate(request *corev1.UpdatePackageRepositoryRequest, repository *packagingv1alpha1.PackageRepository) (*packagingv1alpha1.PackageRepository, error) {
	// repository stub
	repository.Spec = packagingv1alpha1.PackageRepositorySpec{}

	// synchronization
	interval := request.GetInterval()
	if interval > 0 {
		repository.Spec.SyncPeriod = &metav1.Duration{Duration: time.Duration(interval) * time.Second}
	}

	// fetch (todo -> add support for other directives than just imgkpg)
	repository.Spec.Fetch = &packagingv1alpha1.PackageRepositoryFetch{
		ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
			Image: request.GetUrl(),
		},
	}

	return repository, nil
}

// utils

func statusReason(status kappctrlv1alpha1.Condition) corev1.PackageRepositoryStatus_StatusReason {
	switch status.Type {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS
	case kappctrlv1alpha1.ReconcileFailed:
		return corev1.PackageRepositoryStatus_STATUS_REASON_FAILED
	case kappctrlv1alpha1.Reconciling:
		return corev1.PackageRepositoryStatus_STATUS_REASON_PENDING
	}
	// Fall back to unknown/unspecified.
	return corev1.PackageRepositoryStatus_STATUS_REASON_UNSPECIFIED
}

func statusUserReason(status kappctrlv1alpha1.Condition, usefulerror string) string {
	if status.Type == kappctrlv1alpha1.ReconcileSucceeded {
		return status.Message
	}

	if strings.Contains(status.Message, ".status.usefulErrorMessage") {
		return usefulerror
	} else {
		return status.Message
	}
}
