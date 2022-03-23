// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"k8s.io/client-go/rest"
)

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
func statusReasonForKappStatus(status kappctrlv1alpha1.AppConditionType) corev1.InstalledPackageStatus_StatusReason {
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
func simpleUserReasonForKappStatus(status kappctrlv1alpha1.AppConditionType) string {
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
	}
)

var defaultPluginConfig = &kappControllerPluginParsedConfig{
	versionsInSummary:                  pkgutils.GetDefaultVersionsInSummary(),
	timeoutSeconds:                     fallbackTimeoutSeconds,
	defaultUpgradePolicy:               fallbackDefaultUpgradePolicy,
	defaultPrereleasesVersionSelection: fallbackDefaultPrereleasesVersionSelection(),
	defaultAllowDowngrades:             fallbackDefaultAllowDowngrades,
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
		for _, prereleaseVersionSelectionId := range prereleasesVersionSelection {
			prereleases.Identifiers = append(prereleases.Identifiers, prereleaseVersionSelectionId)
		}
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
