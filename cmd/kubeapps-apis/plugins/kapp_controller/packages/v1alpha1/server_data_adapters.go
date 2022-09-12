// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"

	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	kappctrlpackageinstall "github.com/vmware-tanzu/carvel-kapp-controller/pkg/packageinstall"
	"github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappcorev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
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

	Redacted = "REDACTED"

	Annotation_ManagedBy_Key   = "kubeapps.dev/managed-by"
	Annotation_ManagedBy_Value = "plugin:kapp-controller"

	SSHAuthKnownHosts = "ssh-knownhosts"
	BearerAuthToken   = "token"
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
		installedPackageDetail.ReconciliationOptions.Interval = pkgutils.FromDuration(app.Spec.SyncPeriod)
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
	// #nosec G101
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
		if pkgInstall.Spec.SyncPeriod, err = pkgutils.ToDuration(reconciliationOptions.Interval); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "The interval is invalid: %v", err)
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

func (s *Server) buildPackageRepositorySummary(pkgRepository *packagingv1alpha1.PackageRepository, cluster string) (*corev1.PackageRepositorySummary, error) {

	// base struct
	repository := &corev1.PackageRepositorySummary{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgRepository.Namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: pkgRepository.Name,
		},
		Name:            pkgRepository.Name,
		NamespaceScoped: s.pluginConfig.globalPackagingNamespace != pkgRepository.Namespace,
		RequiresAuth:    repositorySecretRef(pkgRepository) != nil,
	}

	// handle fetch-specific configuration
	fetch := pkgRepository.Spec.Fetch
	switch {
	case fetch.ImgpkgBundle != nil:
		repository.Type = Type_ImgPkgBundle
		repository.Url = fetch.ImgpkgBundle.Image
	case fetch.Image != nil:
		repository.Type = Type_Image
		repository.Url = fetch.Image.URL
	case fetch.Git != nil:
		repository.Type = Type_GIT
		repository.Url = fetch.Git.URL
	case fetch.HTTP != nil:
		repository.Type = Type_HTTP
		repository.Url = fetch.HTTP.URL
	case fetch.Inline != nil:
		repository.Type = Type_Inline
	default:
		return nil, fmt.Errorf("the package repository has a fetch directive that is not supported")
	}

	// extract status
	if len(pkgRepository.Status.Conditions) > 0 {
		repository.Status = &corev1.PackageRepositoryStatus{
			Ready:      pkgRepository.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReason(pkgRepository.Status.Conditions[0]),
			UserReason: statusUserReason(pkgRepository.Status.Conditions[0], pkgRepository.Status.UsefulErrorMessage),
		}
	}

	// result
	return repository, nil
}

func (s *Server) buildPackageRepository(pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret, cluster string) (*corev1.PackageRepositoryDetail, error) {

	// base struct
	repository := &corev1.PackageRepositoryDetail{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: pkgRepository.Namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: pkgRepository.Name,
		},
		Name:            pkgRepository.Name,
		NamespaceScoped: s.pluginConfig.globalPackagingNamespace != pkgRepository.Namespace,
	}

	// synchronization
	repository.Interval = pkgutils.FromDuration(pkgRepository.Spec.SyncPeriod)

	// handle fetch-specific configuration
	var customFetch *kappcorev1.PackageRepositoryFetch

	fetch := pkgRepository.Spec.Fetch
	switch {
	case fetch.ImgpkgBundle != nil:
		{
			repository.Type = Type_ImgPkgBundle
			repository.Url = fetch.ImgpkgBundle.Image

			customFetch = toFetchImgpkg(fetch.ImgpkgBundle)
		}
	case fetch.Image != nil:
		{
			repository.Type = Type_Image
			repository.Url = fetch.Image.URL

			customFetch = toFetchImage(fetch.Image)
		}
	case fetch.Git != nil:
		{
			repository.Type = Type_GIT
			repository.Url = fetch.Git.URL

			customFetch = toFetchGit(fetch.Git)
		}
	case fetch.HTTP != nil:
		{
			repository.Type = Type_HTTP
			repository.Url = fetch.HTTP.URL

			customFetch = toFetchHttp(fetch.HTTP)
		}
	case fetch.Inline != nil:
		{
			repository.Type = Type_Inline
			customFetch = toFetchInline(fetch.Inline)
		}
	default:
		return nil, fmt.Errorf("the package repository has a fetch directive that is not supported")
	}

	if customFetch != nil {
		if customDetail, err := anypb.New(&kappcorev1.KappControllerPackageRepositoryCustomDetail{
			Fetch: customFetch,
		}); err != nil {
			return nil, err
		} else {
			repository.CustomDetail = customDetail
		}
	}

	// auth
	if pkgSecret != nil {
		auth := &corev1.PackageRepositoryAuth{}
		if isPluginManaged(pkgRepository, pkgSecret) {
			switch {
			case isBasicAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: Redacted,
						Password: Redacted,
					},
				}
			case isSshAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SshCreds{
					SshCreds: &corev1.SshCredentials{
						PrivateKey: Redacted,
						KnownHosts: Redacted,
					},
				}
			case isDockerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Username: Redacted,
						Password: Redacted,
						Server:   Redacted,
						Email:    Redacted,
					},
				}
			case isBearerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: Redacted,
				}
			default:
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
			}
		} else {
			switch {
			case isBasicAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			case isSshAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH
			case isDockerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
			case isBearerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
			default:
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
			}
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: pkgSecret.Name,
				},
			}
		}
		repository.Auth = auth
	}

	// extract status
	if len(pkgRepository.Status.Conditions) > 0 {
		repository.Status = &corev1.PackageRepositoryStatus{
			Ready:      pkgRepository.Status.Conditions[0].Type == kappctrlv1alpha1.ReconcileSucceeded,
			Reason:     statusReason(pkgRepository.Status.Conditions[0]),
			UserReason: statusUserReason(pkgRepository.Status.Conditions[0], pkgRepository.Status.UsefulErrorMessage),
		}
	}

	// result
	return repository, nil
}

func (s *Server) buildPkgRepositoryCreate(request *corev1.AddPackageRepositoryRequest, pkgSecret *k8scorev1.Secret) (*packagingv1alpha1.PackageRepository, error) {
	// identifier
	namespace := request.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.Name

	// custom details
	details := &kappcorev1.KappControllerPackageRepositoryCustomDetail{}
	if request.CustomDetail != nil {
		if err := request.CustomDetail.UnmarshalTo(details); err != nil {
			return nil, fmt.Errorf("custom details are invalid: %v", err)
		}
	}

	// repository
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
	}
	repository.Spec = s.buildPkgRepositorySpec(request.Type, request.Interval, request.Url, request.Auth, pkgSecret, details)

	return repository, nil
}

func (s *Server) buildPkgRepositoryUpdate(request *corev1.UpdatePackageRepositoryRequest, repository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) (*packagingv1alpha1.PackageRepository, error) {
	// existing type
	var rptype string
	switch {
	case repository.Spec.Fetch.ImgpkgBundle != nil:
		rptype = Type_ImgPkgBundle
	case repository.Spec.Fetch.Image != nil:
		rptype = Type_Image
	case repository.Spec.Fetch.Git != nil:
		rptype = Type_GIT
	case repository.Spec.Fetch.HTTP != nil:
		rptype = Type_HTTP
	}

	// custom details
	details := &kappcorev1.KappControllerPackageRepositoryCustomDetail{}
	if request.CustomDetail != nil {
		if err := request.CustomDetail.UnmarshalTo(details); err != nil {
			return nil, fmt.Errorf("custom details are invalid: %v", err)
		}
	}

	// repository
	repository.Spec = s.buildPkgRepositorySpec(rptype, request.Interval, request.Url, request.Auth, pkgSecret, details)

	return repository, nil
}

func (s *Server) buildPkgRepositorySpec(rptype string, interval string, url string, auth *corev1.PackageRepositoryAuth, pkgSecret *k8scorev1.Secret, details *kappcorev1.KappControllerPackageRepositoryCustomDetail) packagingv1alpha1.PackageRepositorySpec {
	// spec stub
	spec := packagingv1alpha1.PackageRepositorySpec{
		Fetch: &packagingv1alpha1.PackageRepositoryFetch{},
	}

	// synchronization
	spec.SyncPeriod, _ = pkgutils.ToDuration(interval)

	// auth
	var secretRef *kappctrlv1alpha1.AppFetchLocalRef
	if auth != nil {
		if auth.GetSecretRef() != nil {
			secretRef = &kappctrlv1alpha1.AppFetchLocalRef{Name: auth.GetSecretRef().GetName()}
		} else if pkgSecret != nil {
			secretRef = &kappctrlv1alpha1.AppFetchLocalRef{Name: pkgSecret.GetName()}
		}
	}

	// fetch
	switch rptype {
	case Type_ImgPkgBundle:
		{
			imgpkg := &kappctrlv1alpha1.AppFetchImgpkgBundle{
				Image:     url,
				SecretRef: secretRef,
			}
			if details.Fetch != nil && details.Fetch.ImgpkgBundle != nil {
				toPkgFetchImgpkg(details.Fetch.ImgpkgBundle, imgpkg)
			}
			spec.Fetch.ImgpkgBundle = imgpkg
		}
	case Type_Image:
		{
			image := &kappctrlv1alpha1.AppFetchImage{
				URL:       url,
				SecretRef: secretRef,
			}
			if details.Fetch != nil && details.Fetch.Image != nil {
				toPkgFetchImage(details.Fetch.Image, image)
			}
			spec.Fetch.Image = image
		}
	case Type_GIT:
		{
			git := &kappctrlv1alpha1.AppFetchGit{
				URL:       url,
				SecretRef: secretRef,
			}
			if details.Fetch != nil && details.Fetch.Git != nil {
				toPkgFetchGit(details.Fetch.Git, git)
			}
			spec.Fetch.Git = git
		}
	case Type_HTTP:
		{
			http := &kappctrlv1alpha1.AppFetchHTTP{
				URL:       url,
				SecretRef: secretRef,
			}
			if details.Fetch != nil && details.Fetch.Http != nil {
				toPkgFetchHttp(details.Fetch.Http, http)
			}
			spec.Fetch.HTTP = http
		}
	}

	return spec
}

// package repositories validation

func (s *Server) validatePackageRepositoryCreate(ctx context.Context, cluster string, request *corev1.AddPackageRepositoryRequest) error {
	namespace := request.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}

	if request.Description != "" {
		return status.Errorf(codes.InvalidArgument, "Description is not supported")
	}
	if request.TlsConfig != nil {
		return status.Errorf(codes.InvalidArgument, "TLS Config is not supported")
	}

	if request.Name == "" {
		return status.Errorf(codes.InvalidArgument, "no request Name provided")
	}
	if request.NamespaceScoped != (namespace != s.pluginConfig.globalPackagingNamespace) {
		return status.Errorf(codes.InvalidArgument, "Namespace Scope is inconsistent with the provided Namespace")
	}

	switch request.Type {
	case Type_ImgPkgBundle, Type_Image, Type_GIT, Type_HTTP:
		// valid types
	case Type_Inline:
		return status.Errorf(codes.InvalidArgument, "inline repositories are not supported")
	case "":
		return status.Errorf(codes.InvalidArgument, "no repository Type provided")
	default:
		return status.Errorf(codes.InvalidArgument, "invalid repository Type")
	}

	if _, err := pkgutils.ToDuration(request.Interval); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid interval: %v", err)
	}
	if request.Url == "" {
		return status.Errorf(codes.InvalidArgument, "no request Url provided")
	}
	if request.Auth != nil {
		if err := s.validatePackageRepositoryAuth(ctx, cluster, namespace, request.Type, request.Auth, nil, nil); err != nil {
			return err
		}
	}
	if request.CustomDetail != nil {
		if err := s.validatePackageRepositoryDetails(request.Type, request.CustomDetail); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) validatePackageRepositoryUpdate(ctx context.Context, cluster string, request *corev1.UpdatePackageRepositoryRequest, pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) error {
	var rptype string
	switch {
	case pkgRepository.Spec.Fetch.ImgpkgBundle != nil:
		rptype = Type_ImgPkgBundle
	case pkgRepository.Spec.Fetch.Image != nil:
		rptype = Type_Image
	case pkgRepository.Spec.Fetch.Git != nil:
		rptype = Type_GIT
	case pkgRepository.Spec.Fetch.HTTP != nil:
		rptype = Type_HTTP
	case pkgRepository.Spec.Fetch.Inline != nil:
		return status.Errorf(codes.FailedPrecondition, "inline repositories are not supported")
	default:
		return status.Errorf(codes.Internal, "the package repository has a fetch directive that is not supported")
	}

	if request.Description != "" {
		return status.Errorf(codes.InvalidArgument, "Description is not supported")
	}
	if request.TlsConfig != nil {
		return status.Errorf(codes.InvalidArgument, "TLS Config is not supported")
	}

	if _, err := pkgutils.ToDuration(request.Interval); err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid interval: %v", err)
	}
	if request.Url == "" {
		return status.Errorf(codes.InvalidArgument, "no request Url provided")
	}
	if request.Auth != nil {
		if err := s.validatePackageRepositoryAuth(ctx, cluster, pkgRepository.GetNamespace(), rptype, request.Auth, pkgRepository, pkgSecret); err != nil {
			return err
		}
	}
	if request.CustomDetail != nil {
		if err := s.validatePackageRepositoryDetails(rptype, request.CustomDetail); err != nil {
			return err
		}
	}

	if len(pkgRepository.Status.Conditions) > 0 {
		switch statusReason(pkgRepository.Status.Conditions[0]) {
		case corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS:
		case corev1.PackageRepositoryStatus_STATUS_REASON_FAILED:
		default:
			return status.Errorf(codes.FailedPrecondition, "the repository is not in a stable state, wait for the repository to reconcile")
		}
	}

	return nil
}

func (s *Server) validatePackageRepositoryDetails(rptype string, any *anypb.Any) error {
	details := &kappcorev1.KappControllerPackageRepositoryCustomDetail{}
	if err := any.UnmarshalTo(details); err != nil {
		return status.Errorf(codes.InvalidArgument, "custom details are invalid: %v", err)
	}
	if fetch := details.Fetch; fetch != nil {
		switch {
		case fetch.ImgpkgBundle != nil:
			if rptype != Type_ImgPkgBundle {
				return status.Errorf(codes.InvalidArgument, "custom details do not match the expected type %s", rptype)
			}
		case fetch.Image != nil:
			if rptype != Type_Image {
				return status.Errorf(codes.InvalidArgument, "custom details do not match the expected type %s", rptype)
			}
		case fetch.Git != nil:
			if rptype != Type_GIT {
				return status.Errorf(codes.InvalidArgument, "custom details do not match the expected type %s", rptype)
			}
		case fetch.Http != nil:
			if rptype != Type_HTTP {
				return status.Errorf(codes.InvalidArgument, "custom details do not match the expected type %s", rptype)
			}
		case fetch.Inline != nil:
			if rptype != Type_Inline {
				return status.Errorf(codes.InvalidArgument, "custom details do not match the expected type %s", rptype)
			}
		}
	}
	return nil
}

func (s *Server) validatePackageRepositoryAuth(ctx context.Context, cluster, namespace string, rptype string, auth *corev1.PackageRepositoryAuth, pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) error {
	// validate type compatibility
	switch rptype {
	case Type_ImgPkgBundle, Type_Image:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER {
			return status.Errorf(codes.InvalidArgument, "Auth Type is incompatible with the repository Type")
		}
	case Type_GIT:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH {
			return status.Errorf(codes.InvalidArgument, "Auth Type is incompatible with the repository Type")
		}
	case Type_HTTP:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH {
			return status.Errorf(codes.InvalidArgument, "Auth Type is incompatible with the repository Type")
		}
	}

	// validate mode compatibility (applies to updates only)
	if pkgRepository != nil && pkgSecret != nil {
		if isPluginManaged(pkgRepository, pkgSecret) != (auth.GetSecretRef() == nil) {
			return status.Errorf(codes.InvalidArgument, "Auth management mode cannot be changed")
		}
	}

	// validate the type is not changed
	if pkgRepository != nil && pkgSecret != nil && isPluginManaged(pkgRepository, pkgSecret) {
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if !isBasicAuth(pkgSecret) {
				return status.Errorf(codes.InvalidArgument, "auth type cannot be changed")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
			if !isSshAuth(pkgSecret) {
				return status.Errorf(codes.InvalidArgument, "auth type cannot be changed")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if !isDockerAuth(pkgSecret) {
				return status.Errorf(codes.InvalidArgument, "auth type cannot be changed")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
			if !isBearerAuth(pkgSecret) {
				return status.Errorf(codes.InvalidArgument, "auth type cannot be changed")
			}
		}
	}

	// validate referenced secret matches type
	if auth.GetSecretRef() != nil {
		name := auth.GetSecretRef().Name
		if name == "" {
			return status.Errorf(codes.InvalidArgument, "invalid auth, the secret name is not provided")
		}

		secret, err := s.getSecret(ctx, cluster, namespace, name)
		if err != nil {
			err = statuserror.FromK8sError("get", "Secret", name, err)
			return status.Errorf(codes.InvalidArgument, "invalid auth, the secret could not be accessed: %v", err)
		}

		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if !isBasicAuth(secret) {
				return status.Errorf(codes.InvalidArgument, "invalid auth, the secret does not match the expected Type")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
			if !isSshAuth(secret) {
				return status.Errorf(codes.InvalidArgument, "invalid auth, the secret does not match the expected Type")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if !isDockerAuth(secret) {
				return status.Errorf(codes.InvalidArgument, "invalid auth, the secret does not match the expected Type")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
			if !isBearerAuth(secret) {
				return status.Errorf(codes.InvalidArgument, "invalid auth, the secret does not match the expected Type")
			}
		}

		return nil
	}

	// validate auth data
	//    ensures the expected credential struct is provided
	//    for new auth, credentials can't have Redacted content
	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		up := auth.GetUsernamePassword()
		if up == nil || up.Username == "" || up.Password == "" {
			return status.Errorf(codes.InvalidArgument, "missing basic auth credentials")
		}
		if pkgSecret == nil {
			if up.Username == Redacted || up.Password == Redacted {
				return status.Errorf(codes.InvalidArgument, "invalid auth, unexpected REDACTED content")
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		ssh := auth.GetSshCreds()
		if ssh == nil || ssh.PrivateKey == "" {
			return status.Errorf(codes.InvalidArgument, "missing SSH auth credentials")
		}
		if pkgSecret == nil {
			if ssh.PrivateKey == Redacted || ssh.KnownHosts == Redacted {
				return status.Errorf(codes.InvalidArgument, "invalid auth, unexpected REDACTED content")
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		docker := auth.GetDockerCreds()
		if docker == nil || docker.Username == "" || docker.Password == "" || docker.Server == "" {
			return status.Errorf(codes.InvalidArgument, "missing Docker Config auth credentials")
		}
		if pkgSecret == nil {
			if docker.Username == Redacted || docker.Password == Redacted || docker.Server == Redacted || docker.Email == Redacted {
				return status.Errorf(codes.InvalidArgument, "invalid auth, unexpected REDACTED content")
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		token := auth.GetHeader()
		if token == "" {
			return status.Errorf(codes.InvalidArgument, "missing Token auth credentials")
		}
		if pkgSecret == nil {
			if token == Redacted {
				return status.Errorf(codes.InvalidArgument, "invalid auth, unexpected REDACTED content")
			}
		}
	}
	return nil
}

// package repositories secrets

func (s *Server) buildPkgRepositorySecretCreate(namespace, name string, auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, error) {
	pkgSecret := &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: name + "-",
			Annotations:  map[string]string{Annotation_ManagedBy_Key: Annotation_ManagedBy_Value},
		},
		StringData: map[string]string{},
	}

	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		pkgSecret.Type = k8scorev1.SecretTypeBasicAuth

		up := auth.GetUsernamePassword()
		pkgSecret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
		pkgSecret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		pkgSecret.Type = k8scorev1.SecretTypeSSHAuth

		ssh := auth.GetSshCreds()
		pkgSecret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
		pkgSecret.StringData[SSHAuthKnownHosts] = ssh.KnownHosts

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		pkgSecret.Type = k8scorev1.SecretTypeDockerConfigJson

		docker := auth.GetDockerCreds()
		if dockerjson, err := toDockerConfig(docker); err != nil {
			return nil, err
		} else {
			pkgSecret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		pkgSecret.Type = k8scorev1.SecretTypeOpaque
		if token := auth.GetHeader(); token != "" {
			pkgSecret.StringData[BearerAuthToken] = "Bearer " + strings.TrimPrefix(token, "Bearer ")
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "Bearer token is missing")
		}
	}

	return pkgSecret, nil
}

func (s *Server) buildPkgRepositorySecretUpdate(pkgSecret *k8scorev1.Secret, auth *corev1.PackageRepositoryAuth) (bool, error) {
	pkgSecret.StringData = map[string]string{}

	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		up := auth.GetUsernamePassword()
		if isBasicAuth(pkgSecret) {
			if up.Username == Redacted && up.Password == Redacted {
				return false, nil
			}
			if up.Username != Redacted {
				pkgSecret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
			}
			if up.Password != Redacted {
				pkgSecret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password
			}
		} else {
			pkgSecret.Type = k8scorev1.SecretTypeBasicAuth
			pkgSecret.Data = nil
			pkgSecret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
			pkgSecret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		ssh := auth.GetSshCreds()
		if isSshAuth(pkgSecret) {
			if ssh.PrivateKey == Redacted && ssh.KnownHosts == Redacted {
				return false, nil
			}
			if ssh.KnownHosts != Redacted {
				pkgSecret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
				pkgSecret.StringData[SSHAuthKnownHosts] = ssh.KnownHosts
			}
		} else {
			pkgSecret.Type = k8scorev1.SecretTypeSSHAuth
			pkgSecret.Data = nil
			pkgSecret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
			pkgSecret.StringData[SSHAuthKnownHosts] = ssh.KnownHosts
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		docker := auth.GetDockerCreds()
		if isDockerAuth(pkgSecret) {
			if docker.Username == Redacted && docker.Password == Redacted && docker.Server == Redacted && docker.Email == Redacted {
				return false, nil
			}

			pkgdocker, err := fromDockerConfig(pkgSecret.Data[k8scorev1.DockerConfigJsonKey])
			if err != nil {
				return false, err
			}
			if docker.Username != Redacted {
				pkgdocker.Username = docker.Username
			}
			if docker.Password != Redacted {
				pkgdocker.Password = docker.Password
			}
			if docker.Server != Redacted {
				pkgdocker.Server = docker.Server
			}
			if docker.Email != Redacted {
				pkgdocker.Email = docker.Email
			}

			if dockerjson, err := toDockerConfig(pkgdocker); err != nil {
				return false, err
			} else {
				pkgSecret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
			}
		} else {
			pkgSecret.Type = k8scorev1.SecretTypeDockerConfigJson
			pkgSecret.Data = nil
			if dockerjson, err := toDockerConfig(docker); err != nil {
				return false, err
			} else {
				pkgSecret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
			}
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		token := auth.GetHeader()
		if isBearerAuth(pkgSecret) {
			if token == Redacted {
				return false, nil
			} else {
				pkgSecret.StringData[BearerAuthToken] = "Bearer " + strings.TrimPrefix(token, "Bearer ")
			}
		} else {
			pkgSecret.Type = k8scorev1.SecretTypeOpaque
			pkgSecret.Data = nil
			pkgSecret.StringData[BearerAuthToken] = "Bearer " + strings.TrimPrefix(token, "Bearer ")
		}
	}

	return true, nil
}

// status utils

func statusReason(status kappctrlv1alpha1.Condition) corev1.PackageRepositoryStatus_StatusReason {
	switch status.Type {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS
	case kappctrlv1alpha1.Reconciling, kappctrlv1alpha1.Deleting:
		return corev1.PackageRepositoryStatus_STATUS_REASON_PENDING
	case kappctrlv1alpha1.ReconcileFailed, kappctrlv1alpha1.DeleteFailed:
		return corev1.PackageRepositoryStatus_STATUS_REASON_FAILED
	}
	// Fall back to unknown/unspecified.
	return corev1.PackageRepositoryStatus_STATUS_REASON_UNSPECIFIED
}

func statusUserReason(status kappctrlv1alpha1.Condition, usefulerror string) string {
	switch status.Type {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return status.Message
	case kappctrlv1alpha1.Reconciling:
		if status.Message == "" {
			return "Reconciling"
		}
		return status.Message
	case kappctrlv1alpha1.Deleting:
		if status.Message == "" {
			return "Deleting"
		}
		return status.Message
	}

	if strings.Contains(status.Message, ".status.usefulErrorMessage") {
		return usefulerror
	} else {
		return status.Message
	}
}
