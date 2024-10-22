// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/k8sutils"

	"carvel.dev/vendir/pkg/vendir/versions"
	vendirversions "carvel.dev/vendir/pkg/vendir/versions/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	kappctrlpackageinstall "github.com/vmware-tanzu/carvel-kapp-controller/pkg/packageinstall"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappcorev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"google.golang.org/protobuf/types/known/anypb"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

const (
	typeInline       = "inline"
	typeImage        = "image"
	typeImgPkgBundle = "imgpkgBundle"
	typeHTTP         = "http"
	typeGIT          = "git"

	redacted = "REDACTED"

	annotationManagedByKey   = "kubeapps.dev/managed-by"
	annotationManagedByValue = "plugin:kapp-controller"

	sshAuthKnownHosts = "ssh-knownhosts"
	bearerAuthToken   = "token"
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
		return nil, fmt.Errorf("cannot get the latest matching version for the pkg %q: %s", pkgMetadata.Name, err.Error())
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
		return nil, fmt.Errorf("cannot get the latest matching version for the pkg %q: %s", pkgMetadata.Name, err.Error())
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

	if len(pkgInstall.Status.Conditions) > 0 {
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
	eligibleVersion, err := versions.HighestConstrainedVersion([]string{pkgVersion}, vendirversions.VersionSelection{Semver: versionSelection})
	if eligibleVersion == "" || err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The selected version %q is not eligible to be installed: %w", pkgVersion, err))
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
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The interval is invalid: %w", err))
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
		Description:     k8sutils.GetDescription(&pkgRepository.ObjectMeta),
		NamespaceScoped: s.pluginConfig.globalPackagingNamespace != pkgRepository.Namespace,
		RequiresAuth:    repositorySecretRef(pkgRepository) != nil,
	}

	// handle fetch-specific configuration
	fetch := pkgRepository.Spec.Fetch
	switch {
	case fetch.ImgpkgBundle != nil:
		repository.Type = typeImgPkgBundle
		repository.Url = fetch.ImgpkgBundle.Image
	case fetch.Image != nil:
		repository.Type = typeImage
		repository.Url = fetch.Image.URL
	case fetch.Git != nil:
		repository.Type = typeGIT
		repository.Url = fetch.Git.URL
	case fetch.HTTP != nil:
		repository.Type = typeHTTP
		repository.Url = fetch.HTTP.URL
	case fetch.Inline != nil:
		repository.Type = typeInline
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
		Description:     k8sutils.GetDescription(&pkgRepository.ObjectMeta),
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
			repository.Type = typeImgPkgBundle
			repository.Url = fetch.ImgpkgBundle.Image

			customFetch = toFetchImgpkg(fetch.ImgpkgBundle)
		}
	case fetch.Image != nil:
		{
			repository.Type = typeImage
			repository.Url = fetch.Image.URL

			customFetch = toFetchImage(fetch.Image)
		}
	case fetch.Git != nil:
		{
			repository.Type = typeGIT
			repository.Url = fetch.Git.URL

			customFetch = toFetchGit(fetch.Git)
		}
	case fetch.HTTP != nil:
		{
			repository.Type = typeHTTP
			repository.Url = fetch.HTTP.URL

			customFetch = toFetchHttp(fetch.HTTP)
		}
	case fetch.Inline != nil:
		{
			repository.Type = typeInline
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
						Username: redacted,
						Password: redacted,
					},
				}
			case isSshAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SshCreds{
					SshCreds: &corev1.SshCredentials{
						PrivateKey: redacted,
						KnownHosts: redacted,
					},
				}
			case isDockerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Username: redacted,
						Password: redacted,
						Server:   redacted,
						Email:    redacted,
					},
				}
			case isBearerAuth(pkgSecret):
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: redacted,
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
			Name:      name,
			Namespace: namespace,
		},
	}
	repository.Spec = s.buildPkgRepositorySpec(request.Type, request.Interval, request.Url, request.Auth, pkgSecret, details)

	// description
	k8sutils.SetDescription(&repository.ObjectMeta, request.Description)

	return repository, nil
}

func (s *Server) buildPkgRepositoryUpdate(request *corev1.UpdatePackageRepositoryRequest, repository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) (*packagingv1alpha1.PackageRepository, error) {
	// existing type
	var rptype string
	switch {
	case repository.Spec.Fetch.ImgpkgBundle != nil:
		rptype = typeImgPkgBundle
	case repository.Spec.Fetch.Image != nil:
		rptype = typeImage
	case repository.Spec.Fetch.Git != nil:
		rptype = typeGIT
	case repository.Spec.Fetch.HTTP != nil:
		rptype = typeHTTP
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

	// description
	k8sutils.SetDescription(&repository.ObjectMeta, request.Description)

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
	case typeImgPkgBundle:
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
	case typeImage:
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
	case typeGIT:
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
	case typeHTTP:
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

func (s *Server) validatePackageRepositoryCreate(ctx context.Context, cluster string, request *connect.Request[corev1.AddPackageRepositoryRequest]) error {
	namespace := request.Msg.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}

	if request.Msg.TlsConfig != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TLS Config is not supported"))
	}

	if request.Msg.Name == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Name provided"))
	}
	if request.Msg.NamespaceScoped != (namespace != s.pluginConfig.globalPackagingNamespace) {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Namespace Scope is inconsistent with the provided Namespace"))
	}

	switch request.Msg.Type {
	case typeImgPkgBundle, typeImage, typeGIT, typeHTTP:
		// valid types
	case typeInline:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Inline repositories are not supported"))
	case "":
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No repository Type provided"))
	default:
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid repository Type"))
	}

	if _, err := pkgutils.ToDuration(request.Msg.Interval); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid interval: %w", err))
	}
	if request.Msg.Url == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Url provided"))
	}
	if request.Msg.Auth != nil {
		if err := s.validatePackageRepositoryAuth(ctx, request.Header(), cluster, namespace, request.Msg.Type, request.Msg.Auth, nil, nil); err != nil {
			return err
		}
	}
	if request.Msg.CustomDetail != nil {
		if err := s.validatePackageRepositoryDetails(request.Msg.Type, request.Msg.CustomDetail); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) validatePackageRepositoryUpdate(ctx context.Context, cluster string, request *connect.Request[corev1.UpdatePackageRepositoryRequest], pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) error {
	var rptype string
	switch {
	case pkgRepository.Spec.Fetch.ImgpkgBundle != nil:
		rptype = typeImgPkgBundle
	case pkgRepository.Spec.Fetch.Image != nil:
		rptype = typeImage
	case pkgRepository.Spec.Fetch.Git != nil:
		rptype = typeGIT
	case pkgRepository.Spec.Fetch.HTTP != nil:
		rptype = typeHTTP
	case pkgRepository.Spec.Fetch.Inline != nil:
		return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Inline repositories are not supported"))
	default:
		return connect.NewError(connect.CodeInternal, fmt.Errorf("The package repository has a fetch directive that is not supported"))
	}

	if request.Msg.TlsConfig != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TLS Config is not supported"))
	}

	if _, err := pkgutils.ToDuration(request.Msg.Interval); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid interval: %w", err))
	}
	if request.Msg.Url == "" {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Url provided"))
	}
	if request.Msg.Auth != nil {
		if err := s.validatePackageRepositoryAuth(ctx, request.Header(), cluster, pkgRepository.GetNamespace(), rptype, request.Msg.Auth, pkgRepository, pkgSecret); err != nil {
			return err
		}
	}
	if request.Msg.CustomDetail != nil {
		if err := s.validatePackageRepositoryDetails(rptype, request.Msg.CustomDetail); err != nil {
			return err
		}
	}

	if len(pkgRepository.Status.Conditions) > 0 {
		switch statusReason(pkgRepository.Status.Conditions[0]) {
		case corev1.PackageRepositoryStatus_STATUS_REASON_SUCCESS:
		case corev1.PackageRepositoryStatus_STATUS_REASON_FAILED:
		default:
			return connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("The repository is not in a stable state, wait for the repository to reconcile"))
		}
	}

	return nil
}

func (s *Server) validatePackageRepositoryDetails(rptype string, any *anypb.Any) error {
	details := &kappcorev1.KappControllerPackageRepositoryCustomDetail{}
	if err := any.UnmarshalTo(details); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details are invalid: %w", err))
	}
	if fetch := details.Fetch; fetch != nil {
		switch {
		case fetch.ImgpkgBundle != nil:
			if rptype != typeImgPkgBundle {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details do not match the expected type %s", rptype))
			}
		case fetch.Image != nil:
			if rptype != typeImage {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details do not match the expected type %s", rptype))
			}
		case fetch.Git != nil:
			if rptype != typeGIT {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details do not match the expected type %s", rptype))
			}
		case fetch.Http != nil:
			if rptype != typeHTTP {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details do not match the expected type %s", rptype))
			}
		case fetch.Inline != nil:
			if rptype != typeInline {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("The custom details do not match the expected type %s", rptype))
			}
		}
	}
	return nil
}

func (s *Server) validatePackageRepositoryAuth(ctx context.Context, headers http.Header, cluster, namespace string, rptype string, auth *corev1.PackageRepositoryAuth, pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) error {
	// ignore auth if type is not specified
	if auth.Type == corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		if auth.GetPackageRepoAuthOneOf() != nil {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth Type is not specified but auth configuration data were provided"))
		}
		return nil
	}

	// validate type compatibility
	// see https://carvel.dev/kapp-controller/docs/v0.43.2/app-overview/#specfetch
	switch rptype {
	case typeImgPkgBundle, typeImage:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth Type is incompatible with the repository Type"))
		}
	case typeGIT:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH &&
			auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth Type is incompatible with the repository Type"))
		}
	case typeHTTP:
		if auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth Type is incompatible with the repository Type"))
		}
	}

	// validate mode compatibility (applies to updates only)
	if pkgRepository != nil && pkgSecret != nil {
		if isPluginManaged(pkgRepository, pkgSecret) != (auth.GetSecretRef() == nil) {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth management mode cannot be changed"))
		}
	}

	// validate referenced secret matches type
	if auth.GetSecretRef() != nil {
		name := auth.GetSecretRef().Name
		if name == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret name is not provided"))
		}

		secret, err := s.getSecret(ctx, headers, cluster, namespace, name)
		if err != nil {
			err = connecterror.FromK8sError("get", "Secret", name, err)
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret could not be accessed: %w", err))
		}

		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if !isBasicAuth(secret) {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret does not match the expected Type"))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
			if !isSshAuth(secret) {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret does not match the expected Type"))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if !isDockerAuth(secret) {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret does not match the expected Type"))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
			if !isBearerAuth(secret) {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, the secret does not match the expected Type"))
			}
		}

		return nil
	}

	// validate auth data
	//    ensures the expected credential struct is provided
	//    for new auth or new auth type, credentials can't have redacted content
	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		up := auth.GetUsernamePassword()
		if up == nil || up.Username == "" || up.Password == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Missing basic auth credentials"))
		}
		if pkgSecret == nil || !isBasicAuth(pkgSecret) {
			if up.Username == redacted || up.Password == redacted {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, unexpected REDACTED content"))
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		ssh := auth.GetSshCreds()
		if ssh == nil || ssh.PrivateKey == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Missing SSH auth credentials"))
		}
		if pkgSecret == nil || !isSshAuth(pkgSecret) {
			if ssh.PrivateKey == redacted || ssh.KnownHosts == redacted {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, unexpected REDACTED content"))
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		docker := auth.GetDockerCreds()
		if docker == nil || docker.Username == "" || docker.Password == "" || docker.Server == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Missing Docker Config auth credentials"))
		}
		if pkgSecret == nil || !isDockerAuth(pkgSecret) {
			if docker.Username == redacted || docker.Password == redacted || docker.Server == redacted || docker.Email == redacted {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, unexpected REDACTED content"))
			}
		}
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		token := auth.GetHeader()
		if token == "" {
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Missing Token auth credentials"))
		}
		if pkgSecret == nil || !isBearerAuth(pkgSecret) {
			if token == redacted {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid auth, unexpected REDACTED content"))
			}
		}
	}
	return nil
}

// package repositories secrets

func (s *Server) buildPkgRepositorySecretCreate(namespace, name string, auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, error) {
	secret := &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: name + "-",
			Annotations:  map[string]string{annotationManagedByKey: annotationManagedByValue},
		},
		Type:       k8scorev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}

	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		up := auth.GetUsernamePassword()
		secret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
		secret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		ssh := auth.GetSshCreds()
		secret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
		secret.StringData[sshAuthKnownHosts] = ssh.KnownHosts

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		secret.Type = k8scorev1.SecretTypeDockerConfigJson
		docker := auth.GetDockerCreds()
		if dockerjson, err := toDockerConfig(docker); err != nil {
			return nil, err
		} else {
			secret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		token := auth.GetHeader()
		secret.StringData[bearerAuthToken] = token
	}

	return secret, nil
}

func (s *Server) buildPkgRepositorySecretUpdate(pkgSecret *k8scorev1.Secret, namespace, name string, auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, error) {
	secret := &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace,
			GenerateName: name + "-",
			Annotations:  map[string]string{annotationManagedByKey: annotationManagedByValue},
		},
		Type:       k8scorev1.SecretTypeOpaque,
		StringData: map[string]string{},
	}

	switch auth.Type {
	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
		up := auth.GetUsernamePassword()
		if isBasicAuth(pkgSecret) {
			if up.Username == redacted && up.Password == redacted {
				return nil, nil
			}
			if up.Username != redacted {
				secret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
			} else {
				secret.StringData[k8scorev1.BasicAuthUsernameKey] = string(pkgSecret.Data[k8scorev1.BasicAuthUsernameKey])
			}
			if up.Password != redacted {
				secret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password
			} else {
				secret.StringData[k8scorev1.BasicAuthPasswordKey] = string(pkgSecret.Data[k8scorev1.BasicAuthPasswordKey])
			}
		} else {
			secret.StringData[k8scorev1.BasicAuthUsernameKey] = up.Username
			secret.StringData[k8scorev1.BasicAuthPasswordKey] = up.Password
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_SSH:
		ssh := auth.GetSshCreds()
		if isSshAuth(pkgSecret) {
			if ssh.PrivateKey == redacted && ssh.KnownHosts == redacted {
				return nil, nil
			}
			if ssh.PrivateKey != redacted {
				secret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
			} else {
				secret.StringData[k8scorev1.SSHAuthPrivateKey] = string(pkgSecret.Data[k8scorev1.SSHAuthPrivateKey])
			}
			if ssh.KnownHosts != redacted {
				secret.StringData[sshAuthKnownHosts] = ssh.KnownHosts
			} else {
				secret.StringData[sshAuthKnownHosts] = string(pkgSecret.Data[sshAuthKnownHosts])
			}
		} else {
			secret.StringData[k8scorev1.SSHAuthPrivateKey] = ssh.PrivateKey
			secret.StringData[sshAuthKnownHosts] = ssh.KnownHosts
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
		secret.Type = k8scorev1.SecretTypeDockerConfigJson
		docker := auth.GetDockerCreds()
		if isDockerAuth(pkgSecret) {
			if docker.Username == redacted && docker.Password == redacted && docker.Server == redacted && docker.Email == redacted {
				return nil, nil
			}

			pkgdocker, err := fromDockerConfig(pkgSecret.Data[k8scorev1.DockerConfigJsonKey])
			if err != nil {
				return nil, err
			}
			if docker.Username != redacted {
				pkgdocker.Username = docker.Username
			}
			if docker.Password != redacted {
				pkgdocker.Password = docker.Password
			}
			if docker.Server != redacted {
				pkgdocker.Server = docker.Server
			}
			if docker.Email != redacted {
				pkgdocker.Email = docker.Email
			}

			if dockerjson, err := toDockerConfig(pkgdocker); err != nil {
				return nil, err
			} else {
				secret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
			}
		} else {
			if dockerjson, err := toDockerConfig(docker); err != nil {
				return nil, err
			} else {
				secret.StringData[k8scorev1.DockerConfigJsonKey] = string(dockerjson)
			}
		}

	case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
		token := auth.GetHeader()
		if isBearerAuth(pkgSecret) {
			if token == redacted {
				return nil, nil
			} else {
				secret.StringData[bearerAuthToken] = token
			}
		} else {
			secret.StringData[bearerAuthToken] = token
		}
	}

	return secret, nil
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
