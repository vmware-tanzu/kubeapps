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
	"bytes"
	"reflect"
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"
)

func TestGetPkgVersionsMap(t *testing.T) {
	version123, _ := semver.NewVersion("1.2.3")
	version124, _ := semver.NewVersion("1.2.4")
	version127, _ := semver.NewVersion("1.2.7")
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
					pkg:     &datapackagingv1alpha1.Package{},
					version: version123,
				},
				{
					pkg:     &datapackagingv1alpha1.Package{},
					version: version124,
				},
				{
					pkg:     &datapackagingv1alpha1.Package{},
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

func TestStatusReasonForKappStatus(t *testing.T) {
	tests := []struct {
		name                 string
		status               kappctrlv1alpha1.AppConditionType
		expectedStatusReason corev1.InstalledPackageStatus_StatusReason
	}{
		{"ReconcileSucceeded", kappctrlv1alpha1.AppConditionType("ReconcileSucceeded"), corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.AppConditionType("ValuesSchemaCheckFailed"), corev1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"ReconcileFailed", kappctrlv1alpha1.AppConditionType("ReconcileFailed"), corev1.InstalledPackageStatus_STATUS_REASON_FAILED},
		{"Reconciling", kappctrlv1alpha1.AppConditionType("Reconciling"), corev1.InstalledPackageStatus_STATUS_REASON_PENDING},
		{"Unknown", kappctrlv1alpha1.AppConditionType("foo"), corev1.InstalledPackageStatus_STATUS_REASON_UNSPECIFIED},
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
		{"ReconcileSucceeded", kappctrlv1alpha1.AppConditionType("ReconcileSucceeded"), "Reconcile succeeded"},
		{"ValuesSchemaCheckFailed", kappctrlv1alpha1.AppConditionType("ValuesSchemaCheckFailed"), "Reconcile failed"},
		{"ReconcileFailed", kappctrlv1alpha1.AppConditionType("ReconcileFailed"), "Reconcile failed"},
		{"Reconciling", kappctrlv1alpha1.AppConditionType("Reconciling"), "Reconciling"},
		{"Unknown", kappctrlv1alpha1.AppConditionType("foo"), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userReason := userReasonForKappStatus(tt.status)
			if want, got := tt.expectedUserReason, userReason; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tt.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestExtractValue(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		key      string
		expected string
	}{
		{"retrieves a top level key", `{"a": "foo"}`, "a", " foo"},
		{"returns '' in nested keys", `{"a": { b: "foo"}}`, "b", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := extractValue(tt.body, tt.key)
			if !cmp.Equal(tt.expected, values) {
				t.Errorf("mismatch in '%s': %s", tt.name, cmp.Diff(tt.expected, values))
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

// From https://github.com/kubernetes/apiextensions-apiserver/blob/release-1.21/pkg/apiserver/schema/defaulting/algorithm_test.go
// With a new test case ("object without default values")
func TestDefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		schema   *structuralschema.Structural
		expected string
	}{
		{"empty", "null", nil, "null"},
		{"scalar", "4", &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Default: structuralschema.JSON{"foo"},
			},
		}, "4"},
		{"scalar array", "[1,2]", &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{"foo"},
				},
			},
		}, "[1,2]"},
		{"object array", `[{"a":1},{"b":1},{"c":1}]`, &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{"A"},
						},
					},
					"b": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{"B"},
						},
					},
					"c": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{"C"},
						},
					},
				},
			},
		}, `[{"a":1,"b":"B","c":"C"},{"a":"A","b":1,"c":"C"},{"a":"A","b":"B","c":1}]`},
		// New test case checking our tweaks
		{"object without default values", `{}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Type: "string",
					},
				},
				"b": {
					Generic: structuralschema.Generic{
						Type: "boolean",
					},
				},
				"c": {
					Generic: structuralschema.Generic{
						Type: "array",
					},
				},
				"d": {
					Generic: structuralschema.Generic{
						Type: "number",
					},
				},
				"e": {
					Generic: structuralschema.Generic{
						Type: "integer",
					},
				},
				"f": {
					Generic: structuralschema.Generic{
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
		{"object array object", `{"array":[{"a":1},{"b":2}],"object":{"a":1},"additionalProperties":{"x":{"a":1},"y":{"b":2}}}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"array": {
					Items: &structuralschema.Structural{
						Properties: map[string]structuralschema.Structural{
							"a": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{"A"},
								},
							},
							"b": {
								Generic: structuralschema.Generic{
									Default: structuralschema.JSON{"B"},
								},
							},
						},
					},
				},
				"object": {
					Properties: map[string]structuralschema.Structural{
						"a": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{"N"},
							},
						},
						"b": {
							Generic: structuralschema.Generic{
								Default: structuralschema.JSON{"O"},
							},
						},
					},
				},
				"additionalProperties": {
					Generic: structuralschema.Generic{
						AdditionalProperties: &structuralschema.StructuralOrBool{
							Structural: &structuralschema.Structural{
								Properties: map[string]structuralschema.Structural{
									"a": {
										Generic: structuralschema.Generic{
											Default: structuralschema.JSON{"alpha"},
										},
									},
									"b": {
										Generic: structuralschema.Generic{
											Default: structuralschema.JSON{"beta"},
										},
									},
								},
							},
						},
					},
				},
				"foo": {
					Generic: structuralschema.Generic{
						Default: structuralschema.JSON{"bar"},
					},
				},
			},
		}, `{"array":[{"a":1,"b":"B"},{"a":"A","b":2}],"object":{"a":1,"b":"O"},"additionalProperties":{"x":{"a":1,"b":"beta"},"y":{"a":"alpha","b":2}},"foo":"bar"}`},
		{"empty and null", `[{},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":null},{"a":[]},{"a":{}}]`, &structuralschema.Structural{
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `[{"a":"A"},{"a":1},{"a":0},{"a":0.0},{"a":""},{"a":"A"},{"a":[]},{"a":{}}]`},
		{"null in nullable list", `[null]`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Nullable: true,
			},
			Items: &structuralschema.Structural{
				Properties: map[string]structuralschema.Structural{
					"a": {
						Generic: structuralschema.Generic{
							Default: structuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `[null]`},
		{"null in non-nullable list", `[null]`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				Nullable: false,
			},
			Items: &structuralschema.Structural{
				Generic: structuralschema.Generic{
					Default: structuralschema.JSON{"A"},
				},
			},
		}, `["A"]`},
		{"null in nullable object", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{},
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Nullable: true,
						Default:  structuralschema.JSON{"A"},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object", `{"a": null}`, &structuralschema.Structural{
			Properties: map[string]structuralschema.Structural{
				"a": {
					Generic: structuralschema.Generic{
						Nullable: false,
						Default:  structuralschema.JSON{"A"},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null in nullable object with additionalProperties", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Structural: &structuralschema.Structural{
						Generic: structuralschema.Generic{
							Nullable: true,
							Default:  structuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `{"a": null}`},
		{"null in non-nullable object with additionalProperties", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Structural: &structuralschema.Structural{
						Generic: structuralschema.Generic{
							Nullable: false,
							Default:  structuralschema.JSON{"A"},
						},
					},
				},
			},
		}, `{"a": "A"}`},
		{"null unknown field", `{"a": null}`, &structuralschema.Structural{
			Generic: structuralschema.Generic{
				AdditionalProperties: &structuralschema.StructuralOrBool{
					Bool: true,
				},
			},
		}, `{"a": null}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var in interface{}
			if err := json.Unmarshal([]byte(tt.json), &in); err != nil {
				t.Fatal(err)
			}

			var expected interface{}
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatal(err)
			}

			defaultValues(in, tt.schema)
			if !reflect.DeepEqual(in, expected) {
				var buf bytes.Buffer
				enc := json.NewEncoder(&buf)
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
