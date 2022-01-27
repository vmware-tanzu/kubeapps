// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
	"time"

	semver "github.com/Masterminds/semver/v3"
	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappctrldatapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	k8sstructuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sjsonutil "k8s.io/apimachinery/pkg/util/json"
)

func TestGetPkgVersionsMap(t *testing.T) {
	version123, _ := semver.NewVersion("1.2.3")
	version124, _ := semver.NewVersion("1.2.4")
	version127, _ := semver.NewVersion("1.2.7")
	tests := []struct {
		name                   string
		packages               []*kappctrldatapackagingv1alpha1.Package
		expectedPkgVersionsMap map[string][]pkgSemver
	}{
		{"empty packages", []*kappctrldatapackagingv1alpha1.Package{}, map[string][]pkgSemver{}},
		{"multiple package versions", []*kappctrldatapackagingv1alpha1.Package{
			{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.3",
				},
				Spec: kappctrldatapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.3",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
			{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.7",
				},
				Spec: kappctrldatapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.7",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
				},
			},
			{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.4",
				},
				Spec: kappctrldatapackagingv1alpha1.PackageSpec{
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
					pkg:     &kappctrldatapackagingv1alpha1.Package{},
					version: version123,
				},
				{
					pkg:     &kappctrldatapackagingv1alpha1.Package{},
					version: version124,
				},
				{
					pkg:     &kappctrldatapackagingv1alpha1.Package{},
					version: version127,
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
			opts := cmpopts.IgnoreUnexported(pkgSemver{})
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
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version123,
			},
		},
			">1.0.0",
			version200,
		},
		{"complex constaint", []pkgSemver{
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version123,
			},
		},
			"1.2.3 || >1.0.0 <=1.2.4 || <2.0.0",
			version127,
		},
		{"unsatisfiable constaint", []pkgSemver{
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version200,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version127,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
				version: version124,
			},
			{
				pkg:     &kappctrldatapackagingv1alpha1.Package{},
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
		status               kappctrlv1alpha1.AppConditionType
		expectedStatusReason pkgsGRPCv1alpha1.InstalledPackageStatus_StatusReason
	}{
		{"ReconcileSucceeded", kappctrlv1alpha1.AppConditionType("ReconcileSucceeded"), pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.AppConditionType("ValuesSchemaCheckFailed"), pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"ReconcileFailed", kappctrlv1alpha1.AppConditionType("ReconcileFailed"), pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"Reconciling", kappctrlv1alpha1.AppConditionType("Reconciling"), pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING},
		{"Unknown", kappctrlv1alpha1.AppConditionType("foo"), pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED},
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
		status             kappctrlv1alpha1.AppConditionType
		expectedUserReason string
	}{
		{"ReconcileSucceeded", kappctrlv1alpha1.AppConditionType("ReconcileSucceeded"), "Deployed"},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.AppConditionType("ValuesSchemaCheckFailed"), "Reconcile failed"},
		{"ReconcileFailed", kappctrlv1alpha1.AppConditionType("ReconcileFailed"), "Reconcile failed"},
		{"Reconciling", kappctrlv1alpha1.AppConditionType("Reconciling"), "Reconciling"},
		{"Unknown", kappctrlv1alpha1.AppConditionType("foo"), "Unknown"},
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
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: k8smetav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: k8smetav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: k8smetav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       appResource,
				APIVersion: kappctrlAPIVersion,
			},
			ObjectMeta: k8smetav1.ObjectMeta{
				Namespace: "default",
				Name:      "my-installation",
			},
			Spec: kappctrlv1alpha1.AppSpec{
				SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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

func TestDefaultValuesFromSchema(t *testing.T) {
	tests := []struct {
		name           string
		isCommentedOut bool
		schema         []byte
		expected       string
		expectedErr    error
	}{
		{"schema with defaults", false, []byte(`properties:
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`valueWithDefault: 80
`, nil},
		{"schema with without defaults (integer)", false, []byte(`properties:
  missingDefaultInteger:
    description: Missing default
    type: integer`,
		),
			`missingDefaultInteger: 0
`, nil},
		{"schema with without defaults (number)", false, []byte(`properties:
  missingDefaultNumber:
    description: Missing default
    type: number`,
		),
			`missingDefaultNumber: 0
`, nil},
		{"schema with without defaults (string)", false, []byte(`properties:
  missingDefaultString:
    description: Missing default
    type: string`,
		),
			`missingDefaultString: ""
`, nil},
		{"schema with without defaults (boolean)", false, []byte(`properties:
  missingDefaultBoolean:
    description: Missing default
    type: boolean`,
		),
			`missingDefaultBoolean: false
`, nil},
		{"schema with without defaults (array)", false, []byte(`properties:
  missingDefaultArray:
    description: Missing default
    type: array`,
		),
			`missingDefaultArray: []
`, nil},
		{"schema with without defaults (object)", false, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object`,
		),
			`missingDefaultObject: {}
`, nil},
		{"schema (mixed) with isCommentedOut=false", false, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`missingDefaultObject: {}
valueWithDefault: 80
`, nil},
		{"schema (mixed) with isCommentedOut=true", true, []byte(`properties:
  missingDefaultObject:
    description: Missing default
    type: object
  valueWithDefault:
    default: 80
    description: Value with default
    type: integer`,
		),
			`# missingDefaultObject: {}
# valueWithDefault: 80
`, nil},
		{"good schema (w/ additionalProperties: true, as per jsonschema draft 4)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties: true
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
		{"good schema (w/ additionalProperties: <schema>)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties:
      type: string
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
		{"bad schema (w/ additionalProperties: string)", true, []byte(`properties:
  myAdditionalPropertiesProp:
    type: object
    additionalProperties: string
`,
		),
			`# myAdditionalPropertiesProp: {}
`, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := defaultValuesFromSchema(tt.schema, tt.isCommentedOut)
			if err != nil && tt.expectedErr == nil {
				t.Errorf("unexpected error = %v", err)
			}
			if tt.expectedErr != nil {
				if want, got := tt.expectedErr.Error(), err.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
				}
			} else {
				if want, got := tt.expected, values; !cmp.Equal(want, got) {
					t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
				}
			}
		})
	}
}

func TestBuildReadme(t *testing.T) {
	tests := []struct {
		name           string
		pkgMetadata    *kappctrldatapackagingv1alpha1.PackageMetadata
		foundPkgSemver *pkgSemver
		expected       string
	}{
		{"empty", &kappctrldatapackagingv1alpha1.PackageMetadata{
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       pkgMetadataResource,
				APIVersion: datapackagingAPIVersion,
			},
			ObjectMeta: k8smetav1.ObjectMeta{
				Namespace: "default",
				Name:      "tetris.foo.example.com",
			},
			Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
				DisplayName:        "Classic Tetris",
				IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
				ShortDescription:   "A great game for arcade gamers",
				LongDescription:    "A few sentences but not really a readme",
				Categories:         []string{"logging", "daemon-set"},
				Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				SupportDescription: "Some support information",
				ProviderName:       "Tetris inc.",
			},
		}, &pkgSemver{
			pkg: &kappctrldatapackagingv1alpha1.Package{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgResource,
					APIVersion: datapackagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "tetris.foo.example.com.1.2.3",
				},
				Spec: kappctrldatapackagingv1alpha1.PackageSpec{
					RefName:                         "tetris.foo.example.com",
					Version:                         "1.2.3",
					Licenses:                        []string{"my-license"},
					ReleaseNotes:                    "release notes",
					CapactiyRequirementsDescription: "capacity description",
					ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
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

// From https://github.com/kubernetes/apiextensions-apiserver/blob/release-1.21/pkg/apiserver/schema/defaulting/algorithm_test.go
// With a new test case ("object without default values")
func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		schema   *k8sstructuralschema.Structural
		expected string
	}{
		{"empty", "null", nil, "null"},
		{"scalar", "4", &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				Default: k8sstructuralschema.JSON{"foo"},
			},
		}, "4"},
		{"scalar array", "[1,2]", &k8sstructuralschema.Structural{
			Items: &k8sstructuralschema.Structural{
				Generic: k8sstructuralschema.Generic{
					Default: k8sstructuralschema.JSON{"foo"},
				},
			},
		}, "[1,2]"},
		{"object array", `[{"a":1},{"b":1},{"c":1}]`, &k8sstructuralschema.Structural{
			Items: &k8sstructuralschema.Structural{
				Properties: map[string]k8sstructuralschema.Structural{
					"a": {
						Generic: k8sstructuralschema.Generic{
							Default: k8sstructuralschema.JSON{"A"},
						},
					},
					"b": {
						Generic: k8sstructuralschema.Generic{
							Default: k8sstructuralschema.JSON{"B"},
						},
					},
					"c": {
						Generic: k8sstructuralschema.Generic{
							Default: k8sstructuralschema.JSON{"C"},
						},
					},
				},
			},
		}, `[{"a":1,"b":"B","c":"C"},{"a":"A","b":1,"c":"C"},{"a":"A","b":"B","c":1}]`},
		// New test case checking our tweaks
		{"object without default values", `{}`, &k8sstructuralschema.Structural{
			Properties: map[string]k8sstructuralschema.Structural{
				"a": {
					Generic: k8sstructuralschema.Generic{
						Type: "string",
					},
				},
				"b": {
					Generic: k8sstructuralschema.Generic{
						Type: "boolean",
					},
				},
				"c": {
					Generic: k8sstructuralschema.Generic{
						Type: "array",
					},
				},
				"d": {
					Generic: k8sstructuralschema.Generic{
						Type: "number",
					},
				},
				"e": {
					Generic: k8sstructuralschema.Generic{
						Type: "integer",
					},
				},
				"f": {
					Generic: k8sstructuralschema.Generic{
						Type: "object",
					},
				},
			},
		}, `{
  "a": "",
  "b": false,
  "c": [],
  "d": 0,
  "e": 0,
  "f": {}
}
`},
		{"object array object", `{"array":[{"a":1},{"b":2}],"object":{"a":1},"additionalProperties":{"x":{"a":1},"y":{"b":2}}}`, &k8sstructuralschema.Structural{
			Properties: map[string]k8sstructuralschema.Structural{
				"array": {
					Items: &k8sstructuralschema.Structural{
						Properties: map[string]k8sstructuralschema.Structural{
							"a": {
								Generic: k8sstructuralschema.Generic{
									Default: k8sstructuralschema.JSON{"A"},
								},
							},
							"b": {
								Generic: k8sstructuralschema.Generic{
									Default: k8sstructuralschema.JSON{"B"},
								},
							},
						},
					},
				},
				"object": {
					Properties: map[string]k8sstructuralschema.Structural{
						"a": {
							Generic: k8sstructuralschema.Generic{
								Default: k8sstructuralschema.JSON{"N"},
							},
						},
						"b": {
							Generic: k8sstructuralschema.Generic{
								Default: k8sstructuralschema.JSON{"O"},
							},
						},
					},
				},
				"additionalProperties": {
					Generic: k8sstructuralschema.Generic{
						AdditionalProperties: &k8sstructuralschema.StructuralOrBool{
							Structural: &k8sstructuralschema.Structural{
								Properties: map[string]k8sstructuralschema.Structural{
									"a": {
										Generic: k8sstructuralschema.Generic{
											Default: k8sstructuralschema.JSON{"alpha"},
										},
									},
									"b": {
										Generic: k8sstructuralschema.Generic{
											Default: k8sstructuralschema.JSON{"beta"},
										},
									},
								},
							},
						},
					},
				},
				"foo": {
					Generic: k8sstructuralschema.Generic{
						Default: k8sstructuralschema.JSON{"bar"},
					},
				},
			},
		}, `{"array":[{"a":1,"b":"B"},{"a":"A","b":2}],"object":{"a":1,"b":"O"},"additionalProperties":{"x":{"a":1,"b":"beta"},"y":{"a":"alpha","b":2}},"foo":"bar"}`},
		{"empty and null", `[{},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":null},{"a":[]},{"a":{}}]`, &k8sstructuralschema.Structural{
			Items: &k8sstructuralschema.Structural{
				Properties: map[string]k8sstructuralschema.Structural{
					"a": {
						Generic: k8sstructuralschema.Generic{
							Default: k8sstructuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `[{"a":"A"},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":"A"},{"a":[]},{"a":{}}]`},
		{"null in nullable list", `[null]`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				Nullable: true,
			},
			Items: &k8sstructuralschema.Structural{
				Properties: map[string]k8sstructuralschema.Structural{
					"a": {
						Generic: k8sstructuralschema.Generic{
							Default: k8sstructuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `[null]`},
		{"null in non-nullable list", `[null]`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				Nullable: false,
			},
			Items: &k8sstructuralschema.Structural{
				Generic: k8sstructuralschema.Generic{
					Default: k8sstructuralschema.JSON{"A"},
				},
			},
		}, `["A"]`},
		{"null in nullable object", `{"a": null}`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{},
			Properties: map[string]k8sstructuralschema.Structural{
				"a": {
					Generic: k8sstructuralschema.Generic{
						Nullable: true,
						Default:  k8sstructuralschema.JSON{"A"},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object", `{"a": null}`, &k8sstructuralschema.Structural{
			Properties: map[string]k8sstructuralschema.Structural{
				"a": {
					Generic: k8sstructuralschema.Generic{
						Nullable: false,
						Default:  k8sstructuralschema.JSON{"A"},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null in nullable object with additionalProperties", `{"a": null}`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				AdditionalProperties: &k8sstructuralschema.StructuralOrBool{
					Structural: &k8sstructuralschema.Structural{
						Generic: k8sstructuralschema.Generic{
							Nullable: true,
							Default:  k8sstructuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object with additionalProperties", `{"a": null}`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				AdditionalProperties: &k8sstructuralschema.StructuralOrBool{
					Structural: &k8sstructuralschema.Structural{
						Generic: k8sstructuralschema.Generic{
							Nullable: false,
							Default:  k8sstructuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null unknown field", `{"a": null}`, &k8sstructuralschema.Structural{
			Generic: k8sstructuralschema.Generic{
				AdditionalProperties: &k8sstructuralschema.StructuralOrBool{
					Bool: true,
				},
			},
		}, `{"a": null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in interface{}
			if err := k8sjsonutil.Unmarshal([]byte(tt.json), &in); err != nil {
				t.Fatal(err)
			}

			var expected interface{}
			if err := k8sjsonutil.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatal(err)
			}

			defaultValues(in, tt.schema)
			if !reflect.DeepEqual(in, expected) {
				var buf bytes.Buffer
				enc := k8sjsonutil.NewEncoder(&buf)
				enc.SetIndent("", "  ")
				err := enc.Encode(in)
				if err != nil {
					t.Fatalf("unexpected result mashalling error: %v", err)
				}
				if tt.expected != buf.String() {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tt.expected, buf.String()))
				}
			}
		})
	}
}

func TestVersionConstraintWithUpgradePolicy(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		upgradePolicy upgradePolicy
		expected      string
	}{
		{"get constraints with upgradePolicy 'major'", "1.2.3", major, ">=1.2.3"},
		{"get constraints with upgradePolicy 'minor'", "1.2.3", minor, ">=1.2.3 <2.0.0"},
		{"get constraints with upgradePolicy 'patch'", "1.2.3", patch, ">=1.2.3 <1.3.0"},
		{"get constraints with upgradePolicy 'none'", "1.2.3", none, "1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := versionConstraintWithUpgradePolicy(tt.version, tt.upgradePolicy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !cmp.Equal(tt.expected, values) {
				t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
			}
		})
	}
}
