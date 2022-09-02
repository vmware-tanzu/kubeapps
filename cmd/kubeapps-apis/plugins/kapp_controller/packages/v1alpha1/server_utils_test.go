// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPkgVersionsMap(t *testing.T) {
	version123, _ := semver.NewVersion("1.2.3")
	version124, _ := semver.NewVersion("1.2.4")
	version127, _ := semver.NewVersion("1.2.7")
	version1210, _ := semver.NewVersion("1.2.10")
	tests := []struct {
		name                   string
		packages               []*datapackagingv1alpha1.Package
		expectedPkgVersionsMap map[string][]pkgSemver
	}{
		{"empty packages", []*datapackagingv1alpha1.Package{}, map[string][]pkgSemver{}},
		{"multiple package versions", []*datapackagingv1alpha1.Package{
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.3",
				},
				Spec: datapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.3",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.7",
				},
				Spec: datapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.7",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.10",
				},
				Spec: datapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.10",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
			{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.4",
				},
				Spec: datapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.4",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
		}, map[string][]pkgSemver{
			"tetris.foo.example.com": {
				{
					version: version1210,
				},
				{
					version: version127,
				},
				{
					version: version124,
				},
				{
					version: version123,
				},
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgVersionsMap, err := getPkgVersionsMap(tt.packages)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// We want to compare the `version` field of pkgSemver.
			opts := cmp.Options{
				cmp.AllowUnexported(pkgSemver{}),
				cmpopts.IgnoreFields(pkgSemver{}, "pkg"),
			}
			if want, got := tt.expectedPkgVersionsMap, pkgVersionsMap; !cmp.Equal(want, got, opts) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestLatestMatchingVersion(t *testing.T) {
	version123, _ := semver.NewVersion("1.2.3")
	version124, _ := semver.NewVersion("1.2.4")
	version127, _ := semver.NewVersion("1.2.7")
	version200, _ := semver.NewVersion("2.0.0")
	tests := []struct {
		name                    string
		versions                []pkgSemver
		constraints             string
		expectedMatchingVersion *semver.Version
	}{
		{"simple constaint", []pkgSemver{
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version123,
			},
		},
			">1.0.0",
			version200,
		},
		{"complex constaint", []pkgSemver{
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version123,
			},
		},
			"1.2.3 || >1.0.0 <=1.2.4 || <2.0.0",
			version127,
		},
		{"unsatisfiable constaint", []pkgSemver{
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &datapackagingv1alpha1.Package{},
				version: version123,
			},
		},
			"9.9.9",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matchingVersion, err := latestMatchingVersion(tt.versions, tt.constraints)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectedMatchingVersion != nil {
				opts := cmpopts.IgnoreUnexported(pkgSemver{})
				if want, got := tt.expectedMatchingVersion, matchingVersion; !cmp.Equal(want, got, opts) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got, opts))
				}
			} else {
				if matchingVersion != nil {
					t.Errorf("in %s: mismatch expecting nil got %v", tt.name, matchingVersion)
				}
			}
		})
	}
}

func TestStatusReasonForKappStatus(t *testing.T) {
	tests := []struct {
		name                 string
		status               kappctrlv1alpha1.ConditionType
		expectedStatusReason corev1.InstalledPackageStatus_StatusReason
	}{
		{"ReconcileSucceeded", kappctrlv1alpha1.ConditionType("ReconcileSucceeded"), corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.ConditionType("ValuesSchemaCheckFailed"), corev1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"ReconcileFailed", kappctrlv1alpha1.ConditionType("ReconcileFailed"), corev1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"Reconciling", kappctrlv1alpha1.ConditionType("Reconciling"), corev1.InstalledPackageStatus_STATUS_REASON_PENDING},
		{"Unknown", kappctrlv1alpha1.ConditionType("foo"), corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReason := statusReasonForKappStatus(tt.status)
			if want, got := tt.expectedStatusReason, userReason; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestUserReasonForKappStatus(t *testing.T) {
	tests := []struct {
		name               string
		status             kappctrlv1alpha1.ConditionType
		expectedUserReason string
	}{
		{"ReconcileSucceeded", kappctrlv1alpha1.ConditionType("ReconcileSucceeded"), "Deployed"},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.ConditionType("ValuesSchemaCheckFailed"), "Reconcile failed"},
		{"ReconcileFailed", kappctrlv1alpha1.ConditionType("ReconcileFailed"), "Reconcile failed"},
		{"Reconciling", kappctrlv1alpha1.ConditionType("Reconciling"), "Reconciling"},
		{"Unknown", kappctrlv1alpha1.ConditionType("foo"), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReason := simpleUserReasonForKappStatus(tt.status)
			if want, got := tt.expectedUserReason, userReason; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestBuildPostInstallationNotes(t *testing.T) {
	tests := []struct {
		name     string
		app      *kappctrlv1alpha1.App
		expected string
	}{
		{"renders the expected notes (full)", &kappctrlv1alpha1.App{
			TypeMeta: metav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &metav1.Duration{Duration: (time.Second * 30)},
			},
			Status: kappctrlv1alpha1.AppStatus{
				Deploy: &kappctrlv1alpha1.AppStatusDeploy{
					Stdout: "deployStdout",
					Stderr: "deployStderr",
				},
				Fetch: &kappctrlv1alpha1.AppStatusFetch{
					Stdout: "fetchStdout",
					Stderr: "fetchStderr",
				},
				Inspect: &kappctrlv1alpha1.AppStatusInspect{
					Stdout: "inspectStdout",
					Stderr: "inspectStderr",
				},
			},
		}, `#### Deploy

<x60><x60><x60>
deployStdout
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStdout
<x60><x60><x60>

### Errors

#### Deploy

<x60><x60><x60>
deployStderr
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStderr
<x60><x60><x60>

`},
		{"renders the expected notes (no stderr)", &kappctrlv1alpha1.App{
			TypeMeta: metav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &metav1.Duration{Duration: (time.Second * 30)},
			},
			Status: kappctrlv1alpha1.AppStatus{
				Deploy: &kappctrlv1alpha1.AppStatusDeploy{
					Stdout: "deployStdout",
				},
				Fetch: &kappctrlv1alpha1.AppStatusFetch{
					Stdout: "fetchStdout",
				},
				Inspect: &kappctrlv1alpha1.AppStatusInspect{
					Stdout: "inspectStdout",
				},
			},
		}, `#### Deploy

<x60><x60><x60>
deployStdout
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStdout
<x60><x60><x60>

`},
		{"renders the expected notes (no stdout)", &kappctrlv1alpha1.App{
			TypeMeta: metav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &metav1.Duration{Duration: (time.Second * 30)},
			},
			Status: kappctrlv1alpha1.AppStatus{
				Deploy: &kappctrlv1alpha1.AppStatusDeploy{
					Stderr: "deployStderr",
				},
				Fetch: &kappctrlv1alpha1.AppStatusFetch{
					Stderr: "fetchStderr",
				},
				Inspect: &kappctrlv1alpha1.AppStatusInspect{
					Stderr: "inspectStderr",
				},
			},
		}, `### Errors

#### Deploy

<x60><x60><x60>
deployStderr
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStderr
<x60><x60><x60>

`},
		{"renders the expected notes (missing field)", &kappctrlv1alpha1.App{
			TypeMeta: metav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &metav1.Duration{Duration: (time.Second * 30)},
			},
			Status: kappctrlv1alpha1.AppStatus{
				Fetch: &kappctrlv1alpha1.AppStatusFetch{
					Stdout: "fetchStdout",
				},
				Inspect: &kappctrlv1alpha1.AppStatusInspect{
					Stdout: "inspectStdout",
				},
			},
		}, `#### Fetch

<x60><x60><x60>
fetchStdout
<x60><x60><x60>

`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readme := buildPostInstallationNotes(tt.app)
			// when using `` we cannot escape the ` char itself, so let's replace it later
			expected := strings.ReplaceAll(tt.expected, "<x60>", "`")
			if want, got := expected, readme; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestBuildReadme(t *testing.T) {
	tests := []struct {
		name           string
		pkgMetadata    *datapackagingv1alpha1.PackageMetadata
		foundPkgSemver *pkgSemver
		expected       string
	}{
		{"empty", &datapackagingv1alpha1.PackageMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind:       pkgMetadataResource,
				APIVersion: datapackagingAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "tetris.foo.example.com",
			},
			Spec: datapackagingv1alpha1.PackageMetadataSpec{
				DisplayName:        "Classic Tetris",
				IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
				ShortDescription:   "A great game for arcade gamers",
				LongDescription:    "A few sentences but not really a readme",
				Categories:         []string{"logging", "daemon-set"},
				Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				SupportDescription: "Some support information",
				ProviderName:       "Tetris inc.",
			},
		}, &pkgSemver{
			pkg: &datapackagingv1alpha1.Package{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.3",
				},
				Spec: datapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.3",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
					ReleasedAt:                      metav1.Time{Time: time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
				},
			},
			version: &semver.Version{},
		}, `## Description

A few sentences but not really a readme

## Capactiy requirements

capacity description

## Release notes

release notes

Released at: June, 6 1984

## Support

Some support information

## Licenses

- my-license

`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readme := buildReadme(tt.pkgMetadata, tt.foundPkgSemver)
			if want, got := tt.expected, readme; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestBuildPackageIdentifier(t *testing.T) {
	tests := []struct {
		name        string
		pkgMetadata *datapackagingv1alpha1.PackageMetadata
		expected    string
	}{
		{"empty", &datapackagingv1alpha1.PackageMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind:       pkgMetadataResource,
				APIVersion: datapackagingAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "tetris.foo.example.com",
			},
			Spec: datapackagingv1alpha1.PackageMetadataSpec{
				DisplayName:        "Classic Tetris",
				IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
				ShortDescription:   "A great game for arcade gamers",
				LongDescription:    "A few sentences but not really a readme",
				Categories:         []string{"logging", "daemon-set"},
				Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				SupportDescription: "Some support information",
				ProviderName:       "Tetris inc.",
			},
		}, "unknown/tetris.foo.example.com",
		},
		{"a valid annotation", &datapackagingv1alpha1.PackageMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind:       pkgMetadataResource,
				APIVersion: datapackagingAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "tetris.foo.example.com",
				Annotations: map[string]string{
					REPO_REF_ANNOTATION: "default/tce-repo",
				},
			},
			Spec: datapackagingv1alpha1.PackageMetadataSpec{
				DisplayName:        "Classic Tetris",
				IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
				ShortDescription:   "A great game for arcade gamers",
				LongDescription:    "A few sentences but not really a readme",
				Categories:         []string{"logging", "daemon-set"},
				Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				SupportDescription: "Some support information",
				ProviderName:       "Tetris inc.",
			},
		}, "tce-repo/tetris.foo.example.com",
		},
		{"an invalid annotation", &datapackagingv1alpha1.PackageMetadata{
			TypeMeta: metav1.TypeMeta{
				Kind:       pkgMetadataResource,
				APIVersion: datapackagingAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "tetris.foo.example.com",
				Annotations: map[string]string{
					REPO_REF_ANNOTATION: "default/foo/tce-repo",
				},
			},
			Spec: datapackagingv1alpha1.PackageMetadataSpec{
				DisplayName:        "Classic Tetris",
				IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
				ShortDescription:   "A great game for arcade gamers",
				LongDescription:    "A few sentences but not really a readme",
				Categories:         []string{"logging", "daemon-set"},
				Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				SupportDescription: "Some support information",
				ProviderName:       "Tetris inc.",
			},
		}, "unknown/tetris.foo.example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packageIdentifier := buildPackageIdentifier(tt.pkgMetadata)
			if want, got := tt.expected, packageIdentifier; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}
func TestGetRepoNameFromAnnotation(t *testing.T) {
	tests := []struct {
		name              string
		repoRefAnnotation string
		expected          string
	}{
		{"empty", "", "unknown"},
		{"a valid annotation", "default/tce-repo", "tce-repo"},
		{"an invalid annotation", "default/foo/tce-repo", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoName := getRepoNameFromAnnotation(tt.repoRefAnnotation)
			if want, got := tt.expected, repoName; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestPrereleasesVersionSelection(t *testing.T) {
	tests := []struct {
		name                        string
		prereleasesVersionSelection []string
		expected                    *vendirversions.VersionSelectionSemverPrereleases
	}{
		{"disallow prereleases", nil, nil},
		{"allow prereleases - matching any identifier", []string{}, &vendirversions.VersionSelectionSemverPrereleases{}},
		{"allow prereleases - only matching one identifier", []string{"rc"}, &vendirversions.VersionSelectionSemverPrereleases{Identifiers: []string{"rc"}}},
		{"allow prereleases - only matching anyof identifier", []string{"rc", "pre"}, &vendirversions.VersionSelectionSemverPrereleases{Identifiers: []string{"rc", "pre"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := prereleasesVersionSelection(tt.prereleasesVersionSelection)
			if !cmp.Equal(tt.expected, values) {
				t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
			}
		})
	}
}

func TestFilterMetadatas(t *testing.T) {
	testCases := []struct {
		name              string
		metadatas         []*datapackagingv1alpha1.PackageMetadata
		filterOptions     corev1.FilterOptions
		expectedMetadatas []*datapackagingv1alpha1.PackageMetadata
	}{
		{
			name: "filters out non-matching metadatas",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my foo game",
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my bar game",
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my zed game",
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query: "bar",
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my bar game",
					},
				},
			},
		},
		{
			name: "matches display name",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my foo game",
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query: "foo",
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my foo game",
					},
				},
			},
		},
		{
			name: "matches short description",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						ShortDescription: "A game with a lot of foo",
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query: "foo",
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						ShortDescription: "A game with a lot of foo",
					},
				},
			},
		},
		{
			name: "matches long description",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						LongDescription: "A game with a lot of foo and other stuff",
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query: "foo",
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						LongDescription: "A game with a lot of foo and other stuff",
					},
				},
			},
		},
		{
			name: "matches case insensitively",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my fOo game",
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query: "FoO",
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "my fOo game",
					},
				},
			},
		},
		{
			name: "matches a single category",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category1", "category2", "category3"},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category4", "category5", "category6"},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Categories: []string{"category5"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category4", "category5", "category6"},
					},
				},
			},
		},
		{
			name: "matches multiple categories",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category1", "category2", "category3"},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category4", "category5", "category6"},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Categories: []string{"category2", "category3"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category1", "category2", "category3"},
					},
				},
			},
		},
		{
			name: "does not match unless metadata has all categories from filter",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category1", "category2", "category3"},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category4", "category5", "category6"},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Categories: []string{"category6", "category3"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{},
		},
		{
			name: "does not match unless metadata has all categories from filter",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category1", "category2", "category3"},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						Categories: []string{"category4", "category5", "category6"},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Categories: []string{"category6", "category3"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{},
		},
		{
			name: "does match if metadata has a repo from filter",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/tce-repo",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/another-repo",
						},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Repositories: []string{"tce-repo"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/tce-repo",
						},
					},
				},
			},
		},
		{
			name: "does not match if metadata has an unknown repo",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/another-repo",
						},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Repositories: []string{"tce-repo"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{},
		},
		{
			name: "does match if every filterOptions is met",
			metadatas: []*datapackagingv1alpha1.PackageMetadata{
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "match-pkg",
						Categories:  []string{"match-category"},
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/match-repo",
						},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "non-match-by-display-name",
						Categories:  []string{"match-category"},
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/match-repo",
						},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "match-pkg",
						Categories:  []string{"non-match-by-category"},
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/match-repo",
						},
					},
				},
				{
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "match-pkg",
						Categories:  []string{"match-category"},
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
						Annotations: map[string]string{
							REPO_REF_ANNOTATION: "default/non-match-by-repo",
						},
					},
				},
			},
			filterOptions: corev1.FilterOptions{
				Query:        "match-pkg",
				Categories:   []string{"match-category"},
				Repositories: []string{"match-repo"},
			},
			expectedMetadatas: []*datapackagingv1alpha1.PackageMetadata{{
				Spec: datapackagingv1alpha1.PackageMetadataSpec{
					DisplayName: "match-pkg",
					Categories:  []string{"match-category"},
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com",
					Annotations: map[string]string{
						REPO_REF_ANNOTATION: "default/match-repo",
					},
				},
			},
			},
		},
	}

	//nolint:govet
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := FilterMetadatas(tc.metadatas, &tc.filterOptions), tc.expectedMetadatas; !cmp.Equal(want, got) {
				t.Errorf("mismatch in '%s': %s", tc.name, cmp.Diff(want, got))
			}
		})
	}

}
