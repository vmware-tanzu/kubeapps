// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/credentialprovider"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/Masterminds/semver/v3"
	kappcmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappcorev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
)

const REPO_REF_ANNOTATION = "packaging.carvel.dev/package-repository-ref"
const DEFAULT_REPO_NAME = "unknown"

type pkgSemver struct {
	pkg     *datapackagingv1alpha1.Package
	version *semver.Version
}

// pkgVersionsMap recturns a map of packages keyed by the packagemetadataName.
//
// A Package CR in carvel is really a particular version of a package, so we need
// to sort them by the package metadata name, since this is what they share in common.
// The packages are then sorted by version.
func getPkgVersionsMap(packages []*datapackagingv1alpha1.Package) (map[string][]pkgSemver, error) {
	pkgVersionsMap := map[string][]pkgSemver{}
	for _, pkg := range packages {
		semverVersion, err := semver.NewVersion(pkg.Spec.Version)
		if err != nil {
			return nil, fmt.Errorf("required field spec.version was not semver compatible on kapp-controller Package: %v\n%v", err, pkg)
		}
		pkgVersionsMap[pkg.Spec.RefName] = append(pkgVersionsMap[pkg.Spec.RefName], pkgSemver{pkg, semverVersion})
	}

	for _, pkgVersions := range pkgVersionsMap {
		sort.Slice(pkgVersions, func(i, j int) bool {
			return pkgVersions[i].version.GreaterThan(pkgVersions[j].version)
		})
	}

	return pkgVersionsMap, nil
}

// latestMatchingVersion returns the latest version of a package that matches the given version constraint.
func latestMatchingVersion(versions []pkgSemver, constraints string) (*semver.Version, error) {
	// constraints can be a single one (e.g., ">1.2.3") or a range (e.g., ">1.0.0 <2.0.0 || 3.0.0")
	constraint, err := semver.NewConstraint(constraints)
	if err != nil {
		return nil, fmt.Errorf("the version in the constraint ('%s') is not semver-compatible: %v", constraints, err)
	}

	// assuming 'versions' is sorted,
	// get the first version that satisfies the constraint
	for _, v := range versions {
		if constraint.Check(v.version) {
			return v.version, nil
		}
	}
	return nil, nil
}

// statusReasonForKappStatus returns the reason for a given status
func statusReasonForKappStatus(status kappctrlv1alpha1.ConditionType) corev1.InstalledPackageStatus_StatusReason {
	switch status {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED
	case "ValuesSchemaCheckFailed", kappctrlv1alpha1.ReconcileFailed:
		return corev1.InstalledPackageStatus_STATUS_REASON_FAILED
	case kappctrlv1alpha1.Reconciling:
		return corev1.InstalledPackageStatus_STATUS_REASON_PENDING
	}
	// Fall back to unknown/unspecified.
	return corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED
}

// simpleUserReasonForKappStatus returns the simplified reason for a given status
func simpleUserReasonForKappStatus(status kappctrlv1alpha1.ConditionType) string {
	switch status {
	case kappctrlv1alpha1.ReconcileSucceeded:
		return "Deployed"
	case "ValuesSchemaCheckFailed", kappctrlv1alpha1.ReconcileFailed:
		return "Reconcile failed"
	case kappctrlv1alpha1.Reconciling:
		return "Reconciling"
	case "":
		return "No status information yet"
	}
	// Fall back to unknown/unspecified.
	return "Unknown"
}

// buildReadme generates a readme based on the information there is available
func buildReadme(pkgMetadata *datapackagingv1alpha1.PackageMetadata, foundPkgSemver *pkgSemver) string {
	var readmeSB strings.Builder
	if txt := pkgMetadata.Spec.LongDescription; txt != "" {
		readmeSB.WriteString(fmt.Sprintf("## Description\n\n%s\n\n", txt))
	}
	if txt := foundPkgSemver.pkg.Spec.CapactiyRequirementsDescription; txt != "" {
		readmeSB.WriteString(fmt.Sprintf("## Capactiy requirements\n\n%s\n\n", txt))
	}
	if txt := foundPkgSemver.pkg.Spec.ReleaseNotes; txt != "" {
		readmeSB.WriteString(fmt.Sprintf("## Release notes\n\n%s\n\n", txt))
		if date := foundPkgSemver.pkg.Spec.ReleasedAt.Time; date != (time.Time{}) {
			txt := date.UTC().Format("January, 1 2006")
			readmeSB.WriteString(fmt.Sprintf("Released at: %s\n\n", txt))
		}
	}
	if txt := pkgMetadata.Spec.SupportDescription; txt != "" {
		readmeSB.WriteString(fmt.Sprintf("## Support\n\n%s\n\n", txt))
	}
	if len(foundPkgSemver.pkg.Spec.Licenses) > 0 {
		readmeSB.WriteString("## Licenses\n\n")
		for _, license := range foundPkgSemver.pkg.Spec.Licenses {
			if license != "" {
				readmeSB.WriteString(fmt.Sprintf("- %s\n", license))
			}
		}
		readmeSB.WriteString("\n")
	}
	return readmeSB.String()
}

// buildPackageIdentifier generates the package identifier (repoName/pkgName) for a given package
func buildPackageIdentifier(pkgMetadata *datapackagingv1alpha1.PackageMetadata) string {
	repoName := getRepoNameFromAnnotation(pkgMetadata.Annotations[REPO_REF_ANNOTATION])
	return fmt.Sprintf("%s/%s", repoName, pkgMetadata.Name)
}

// getRepoNameFromAnnotation gets the repo name from a string with the format "namespace/repoName",
// for instance "default/tce-repo", and returns just the "repoName" part, e.g., "tce-repo"
func getRepoNameFromAnnotation(repoRefAnnotation string) string {
	// falling back to a "default" repo name if using kapp controller < v0.36.1
	// See https://github.com/vmware-tanzu/carvel-kapp-controller/pull/532
	repoName := DEFAULT_REPO_NAME
	if repoRefAnnotation != "" {
		splitRepoRefAnnotation := strings.Split(repoRefAnnotation, "/")
		// just change the repo name if we have a valid annotation
		if len(splitRepoRefAnnotation) == 2 {
			repoName = strings.Split(repoRefAnnotation, "/")[1]
		}
	}
	return repoName
}

// buildPostInstallationNotes generates the installation notes based on the application status
func buildPostInstallationNotes(app *kappctrlv1alpha1.App) string {
	var postInstallNotesSB strings.Builder
	deployStdout := ""
	deployStderr := ""
	fetchStdout := ""
	fetchStderr := ""

	if app.Status.Deploy != nil {
		deployStdout = app.Status.Deploy.Stdout
		deployStderr = app.Status.Deploy.Stderr
	}
	if app.Status.Fetch != nil {
		fetchStdout = app.Status.Fetch.Stdout
		fetchStderr = app.Status.Fetch.Stderr
	}

	if deployStdout != "" || fetchStdout != "" {
		if deployStdout != "" {
			postInstallNotesSB.WriteString(fmt.Sprintf("#### Deploy\n\n```\n%s\n```\n\n", deployStdout))
		}
		if fetchStdout != "" {
			postInstallNotesSB.WriteString(fmt.Sprintf("#### Fetch\n\n```\n%s\n```\n\n", fetchStdout))
		}
	}
	if deployStderr != "" || fetchStderr != "" {
		postInstallNotesSB.WriteString("### Errors\n\n")
		if deployStderr != "" {
			postInstallNotesSB.WriteString(fmt.Sprintf("#### Deploy\n\n```\n%s\n```\n\n", deployStderr))
		}
		if fetchStderr != "" {
			postInstallNotesSB.WriteString(fmt.Sprintf("#### Fetch\n\n```\n%s\n```\n\n", fetchStderr))
		}
	}
	return postInstallNotesSB.String()
}

type (
	kappControllerPluginConfig struct {
		Core struct {
			Packages struct {
				V1alpha1 struct {
					VersionsInSummary pkgutils.VersionsInSummary
					TimeoutSeconds    int32 `json:"timeoutSeconds"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"core"`

		KappController struct {
			Packages struct {
				V1alpha1 struct {
					DefaultUpgradePolicy               string   `json:"defaultUpgradePolicy"`
					DefaultPrereleasesVersionSelection []string `json:"defaultPrereleasesVersionSelection"`
					DefaultAllowDowngrades             bool     `json:"defaultAllowDowngrades"`
					GlobalPackagingNamespace           string   `json:"globalPackagingNamespace"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"kappController"`
	}
	kappControllerPluginParsedConfig struct {
		versionsInSummary                  pkgutils.VersionsInSummary
		timeoutSeconds                     int32
		defaultUpgradePolicy               pkgutils.UpgradePolicy
		defaultPrereleasesVersionSelection []string
		defaultAllowDowngrades             bool
		globalPackagingNamespace           string
	}
)

var defaultPluginConfig = &kappControllerPluginParsedConfig{
	versionsInSummary:                  pkgutils.GetDefaultVersionsInSummary(),
	timeoutSeconds:                     fallbackTimeoutSeconds,
	defaultUpgradePolicy:               fallbackDefaultUpgradePolicy,
	defaultPrereleasesVersionSelection: fallbackDefaultPrereleasesVersionSelection(),
	defaultAllowDowngrades:             fallbackDefaultAllowDowngrades,
	globalPackagingNamespace:           fallbackGlobalPackagingNamespace,
}

// prereleasesVersionSelection returns the proper value to the prereleases used in kappctrl from the selection
func prereleasesVersionSelection(prereleasesVersionSelection []string) *vendirversions.VersionSelectionSemverPrereleases {
	// More info on the semantics of prereleases
	// https://kubernetes.slack.com/archives/CH8KCCKA5/p1643376802571959

	if prereleasesVersionSelection == nil {
		// Exception: if the version constraint is a prerelease itself:
		// Current behavior (as of v0.32.0): error, it won't install a prerelease if `prereleases: nil`:
		// Future behavior: install it, it won't be required to set `prereleases:  {}` if the version constraint is a prerelease
		return nil
	}
	// `prereleases: {}`: allow any prerelease if the version constraint allows it
	if len(prereleasesVersionSelection) == 0 {
		return &vendirversions.VersionSelectionSemverPrereleases{}
	} else {
		// `prereleases: {Identifiers: []string{"foo"}}`: allow only prerelease with "foo" as part of the name if the version constraint allows it
		prereleases := &vendirversions.VersionSelectionSemverPrereleases{Identifiers: []string{}}
		prereleases.Identifiers = append(prereleases.Identifiers, prereleasesVersionSelection...)
		return prereleases
	}
}

// implementing a custom ConfigFactory to allow for customizing the *rest.Config
// https://kubernetes.slack.com/archives/CH8KCCKA5/p1642015047046200
type ConfigurableConfigFactoryImpl struct {
	kappcmdcore.ConfigFactoryImpl
	config *rest.Config
}

var _ kappcmdcore.ConfigFactory = &ConfigurableConfigFactoryImpl{}

func NewConfigurableConfigFactoryImpl() *ConfigurableConfigFactoryImpl {
	return &ConfigurableConfigFactoryImpl{}
}

func (f *ConfigurableConfigFactoryImpl) ConfigureRESTConfig(config *rest.Config) {
	f.config = config
}

func (f *ConfigurableConfigFactoryImpl) RESTConfig() (*rest.Config, error) {
	return f.config, nil
}

// FilterMetadatas returns a slice where the content has been filtered
// according to the provided filter options.
func FilterMetadatas(metadatas []*datapackagingv1alpha1.PackageMetadata, filterOptions *corev1.FilterOptions) []*datapackagingv1alpha1.PackageMetadata {
	filteredMetadatas := metadatas
	if filterOptions != nil {
		filteredMetadatas = []*datapackagingv1alpha1.PackageMetadata{}

		skipQueryFilter := filterOptions.Query == ""
		skipCategoriesFilter := len(filterOptions.Categories) == 0
		skipRepositoriesFilter := len(filterOptions.Repositories) == 0

		for _, metadata := range metadatas {
			matchesQuery := skipQueryFilter || testMetadataMatchesQuery(metadata, filterOptions.Query)
			matchesCategories := skipCategoriesFilter || testMetadataMatchesCategories(metadata, filterOptions.Categories)
			matchesRepos := skipRepositoriesFilter || testMetadataMatchesRepos(metadata, filterOptions.Repositories)
			if matchesRepos && matchesQuery && matchesCategories {
				filteredMetadatas = append(filteredMetadatas, metadata)
			}
		}
	}
	return filteredMetadatas
}

func testMetadataMatchesQuery(metadata *datapackagingv1alpha1.PackageMetadata, query string) bool {
	stringToMatch := strings.ToLower(fmt.Sprintf("%s %s %s", metadata.Spec.DisplayName, metadata.Spec.ShortDescription, metadata.Spec.LongDescription))
	return strings.Contains(stringToMatch, strings.ToLower(query))
}

func testMetadataMatchesCategories(metadata *datapackagingv1alpha1.PackageMetadata, categories []string) bool {
	metadataCategoriesHash := map[string]interface{}{}
	for _, category := range metadata.Spec.Categories {
		metadataCategoriesHash[category] = nil
	}

	// We only match if every category in the filter options is present.
	intersection := true
	for _, category := range categories {
		if _, ok := metadataCategoriesHash[category]; !ok {
			return false
		}
	}

	return intersection
}

// testMetadataMatchesRepos returns true if the metadata matches the provided repositories.
func testMetadataMatchesRepos(metadata *datapackagingv1alpha1.PackageMetadata, repositories []string) bool {
	metadataRepoName := getRepoNameFromAnnotation(metadata.Annotations[REPO_REF_ANNOTATION])
	// if the package is from one of the given repositories, it matches
	for _, repo := range repositories {
		if metadataRepoName == repo {
			return true
		}
	}
	return false
}

//
//		Utils for repositories

// translation utils for custom details. todo -> revisit once we can reuse existing proto files

func toFetchImgpkg(pkgfetch *kappctrlv1alpha1.AppFetchImgpkgBundle) *kappcorev1.PackageRepositoryFetch {
	if pkgfetch.TagSelection == nil {
		return nil
	}
	fetch := &kappcorev1.PackageRepositoryFetch{
		ImgpkgBundle: &kappcorev1.PackageRepositoryImgpkg{
			TagSelection: toVersionSelection(pkgfetch.TagSelection),
		},
	}
	return fetch
}

func toFetchImage(pkgfetch *kappctrlv1alpha1.AppFetchImage) *kappcorev1.PackageRepositoryFetch {
	if pkgfetch.SubPath == "" && pkgfetch.TagSelection == nil {
		return nil
	}
	fetch := &kappcorev1.PackageRepositoryFetch{
		Image: &kappcorev1.PackageRepositoryImage{
			SubPath:      pkgfetch.SubPath,
			TagSelection: toVersionSelection(pkgfetch.TagSelection),
		},
	}
	return fetch
}

func toFetchGit(pkgfetch *kappctrlv1alpha1.AppFetchGit) *kappcorev1.PackageRepositoryFetch {
	if pkgfetch.Ref == "" && pkgfetch.RefSelection == nil && pkgfetch.SubPath == "" && !pkgfetch.LFSSkipSmudge {
		return nil
	}
	fetch := &kappcorev1.PackageRepositoryFetch{
		Git: &kappcorev1.PackageRepositoryGit{
			Ref:           pkgfetch.Ref,
			RefSelection:  toVersionSelection(pkgfetch.RefSelection),
			SubPath:       pkgfetch.SubPath,
			LfsSkipSmudge: pkgfetch.LFSSkipSmudge,
		},
	}
	return fetch
}

func toFetchHttp(pkgfetch *kappctrlv1alpha1.AppFetchHTTP) *kappcorev1.PackageRepositoryFetch {
	if pkgfetch.SubPath == "" && pkgfetch.SHA256 == "" {
		return nil
	}
	fetch := &kappcorev1.PackageRepositoryFetch{
		Http: &kappcorev1.PackageRepositoryHttp{
			SubPath: pkgfetch.SubPath,
			Sha256:  pkgfetch.SHA256,
		},
	}
	return fetch
}

func toFetchInline(pkgfetch *kappctrlv1alpha1.AppFetchInline) *kappcorev1.PackageRepositoryFetch {
	if len(pkgfetch.Paths) == 0 && len(pkgfetch.PathsFrom) == 0 {
		return nil
	}

	fetch := &kappcorev1.PackageRepositoryFetch{
		Inline: &kappcorev1.PackageRepositoryInline{
			Paths: pkgfetch.Paths,
		},
	}
	if len(pkgfetch.PathsFrom) > 0 {
		paths := []*kappcorev1.PackageRepositoryInline_Source{}
		for _, pf := range pkgfetch.PathsFrom {
			pathfrom := &kappcorev1.PackageRepositoryInline_Source{}
			if pf.SecretRef != nil {
				pathfrom.SecretRef = &kappcorev1.PackageRepositoryInline_SourceRef{
					Name:          pf.SecretRef.Name,
					DirectoryPath: pf.SecretRef.DirectoryPath,
				}
			}
			if pf.ConfigMapRef != nil {
				pathfrom.ConfigMapRef = &kappcorev1.PackageRepositoryInline_SourceRef{
					Name:          pf.ConfigMapRef.Name,
					DirectoryPath: pf.ConfigMapRef.DirectoryPath,
				}
			}
			paths = append(paths, pathfrom)
		}
		fetch.Inline.PathsFrom = paths
	}

	return fetch
}

func toVersionSelection(pkgversion *vendirversions.VersionSelection) *kappcorev1.VersionSelection {
	if pkgversion == nil || pkgversion.Semver == nil {
		return nil
	}

	version := &kappcorev1.VersionSelection{
		Semver: &kappcorev1.VersionSelectionSemver{
			Constraints: pkgversion.Semver.Constraints,
		},
	}
	if pkgversion.Semver.Prereleases != nil && len(pkgversion.Semver.Prereleases.Identifiers) > 0 {
		version.Semver.Prereleases = &kappcorev1.VersionSelectionSemverPrereleases{
			Identifiers: pkgversion.Semver.Prereleases.Identifiers,
		}
	}

	return version
}

func toPkgFetchImgpkg(from *kappcorev1.PackageRepositoryImgpkg, to *kappctrlv1alpha1.AppFetchImgpkgBundle) {
	to.TagSelection = toPkgVersionSelection(from.TagSelection)
}

func toPkgFetchImage(from *kappcorev1.PackageRepositoryImage, to *kappctrlv1alpha1.AppFetchImage) {
	to.SubPath = from.SubPath
	to.TagSelection = toPkgVersionSelection(from.TagSelection)
}

func toPkgFetchGit(from *kappcorev1.PackageRepositoryGit, to *kappctrlv1alpha1.AppFetchGit) {
	to.Ref = from.Ref
	to.RefSelection = toPkgVersionSelection(from.RefSelection)
	to.SubPath = from.SubPath
	to.LFSSkipSmudge = from.LfsSkipSmudge
}

func toPkgFetchHttp(from *kappcorev1.PackageRepositoryHttp, to *kappctrlv1alpha1.AppFetchHTTP) {
	to.SubPath = from.SubPath
	to.SHA256 = from.Sha256
}

func toPkgVersionSelection(version *kappcorev1.VersionSelection) *vendirversions.VersionSelection {
	if version == nil || version.Semver == nil {
		return nil
	}

	pkgversion := &vendirversions.VersionSelection{
		Semver: &vendirversions.VersionSelectionSemver{
			Constraints: version.Semver.Constraints,
		},
	}
	if version.Semver.Prereleases != nil && len(version.Semver.Prereleases.Identifiers) > 0 {
		pkgversion.Semver.Prereleases = &vendirversions.VersionSelectionSemverPrereleases{
			Identifiers: version.Semver.Prereleases.Identifiers,
		}
	}

	return pkgversion
}

// secret state

func repositorySecretRef(pkgRepository *packagingv1alpha1.PackageRepository) *kappctrlv1alpha1.AppFetchLocalRef {
	fetch := pkgRepository.Spec.Fetch
	switch {
	case fetch.ImgpkgBundle != nil:
		return fetch.ImgpkgBundle.SecretRef
	case fetch.Image != nil:
		return fetch.Image.SecretRef
	case fetch.Git != nil:
		return fetch.Git.SecretRef
	case fetch.HTTP != nil:
		return fetch.HTTP.SecretRef
	}
	return nil
}

func isPluginManaged(pkgRepository *packagingv1alpha1.PackageRepository, pkgSecret *k8scorev1.Secret) bool {
	if !metav1.IsControlledBy(pkgSecret, pkgRepository) {
		return false
	}
	if managedby := pkgSecret.GetAnnotations()[Annotation_ManagedBy_Key]; managedby != Annotation_ManagedBy_Value {
		return false
	}
	return true
}

func isBasicAuth(secret *k8scorev1.Secret) bool {
	return secret.Data != nil && secret.Data[k8scorev1.BasicAuthUsernameKey] != nil && secret.Data[k8scorev1.BasicAuthPasswordKey] != nil
}

func isBearerAuth(secret *k8scorev1.Secret) bool {
	return secret.Data != nil && secret.Data[BearerAuthToken] != nil
}

func isSshAuth(secret *k8scorev1.Secret) bool {
	return secret.Data != nil && secret.Data[k8scorev1.SSHAuthPrivateKey] != nil
}

func isDockerAuth(secret *k8scorev1.Secret) bool {
	return secret.Data != nil && secret.Data[k8scorev1.DockerConfigJsonKey] != nil
}

func toDockerConfig(docker *corev1.DockerCredentials) ([]byte, error) {
	dockerConfig := &credentialprovider.DockerConfigJSON{
		Auths: map[string]credentialprovider.DockerConfigEntry{
			docker.Server: {
				Username: docker.Username,
				Password: docker.Password,
				Email:    docker.Email,
			},
		},
	}
	if dockerjson, err := json.Marshal(dockerConfig); err != nil {
		return nil, err
	} else {
		return dockerjson, nil
	}
}

func fromDockerConfig(dockerjson []byte) (*corev1.DockerCredentials, error) {
	dockerConfig := &credentialprovider.DockerConfigJSON{}
	if err := json.Unmarshal(dockerjson, dockerConfig); err != nil {
		return nil, err
	}
	for server, entry := range dockerConfig.Auths {
		docker := &corev1.DockerCredentials{
			Server:   server,
			Username: entry.Username,
			Password: entry.Password,
			Email:    entry.Email,
		}
		return docker, nil
	}
	return nil, fmt.Errorf("invalid dockerconfig, no Auths entries were found")
}

func setOwnerReference(pkgSecret *k8scorev1.Secret, pkgRepository *packagingv1alpha1.PackageRepository) {
	pkgSecret.OwnerReferences = []metav1.OwnerReference{
		*metav1.NewControllerRef(
			pkgRepository,
			schema.GroupVersionKind{
				Group:   packagingv1alpha1.SchemeGroupVersion.Group,
				Version: packagingv1alpha1.SchemeGroupVersion.Version,
				Kind:    pkgRepositoryResource,
			}),
	}
}
