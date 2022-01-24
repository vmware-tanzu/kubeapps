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
	"bufio"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/runtime"
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

// extracts the value for a key from a JSON-formatted string
// body - the JSON-response as a string. Usually retrieved via the request body
// key - the key for which the value should be extracted
// returns - the value for the given key
// https://stackoverflow.com/questions/17452722/how-to-get-the-key-value-from-a-json-string-in-go/37332972
func extractValue(body string, key string) string {
	keystr := "\"" + key + "\":[^,;\\]}]*"
	r, _ := regexp.Compile(keystr)
	match := r.FindString(body)
	keyValMatch := strings.Split(match, ":")
	value := ""
	if len(keyValMatch) > 1 {
		value = strings.ReplaceAll(keyValMatch[1], "\"", "")
	}
	return value
}

// defaultValuesFromSchema returns a yaml string with default values generated from an OpenAPI v3 Schema
func defaultValuesFromSchema(schema []byte, isCommentedOut bool) (string, error) {
	if len(schema) == 0 {
		return "", nil
	}
	// Deserialize the schema passed into the function
	jsonSchemaProps := &apiextensions.JSONSchemaProps{}
	if err := yaml.Unmarshal(schema, jsonSchemaProps); err != nil {
		return "", err
	}
	structural, err := structuralschema.NewStructural(jsonSchemaProps)
	if err != nil {
		return "", err
	}

	// Generate the default values
	unstructuredDefaultValues := make(map[string]interface{})
	defaultValues(unstructuredDefaultValues, structural)
	yamlDefaultValues, err := yaml.Marshal(unstructuredDefaultValues)
	if err != nil {
		return "", err
	}
	strYamlDefaultValues := string(yamlDefaultValues)

	// If isCommentedOut, add a yaml comment character '#' to the beginning of each line
	if isCommentedOut {
		var sb strings.Builder
		scanner := bufio.NewScanner(strings.NewReader(strYamlDefaultValues))
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			sb.WriteString("# ")
			sb.WriteString(fmt.Sprintln(scanner.Text()))
		}
		strYamlDefaultValues = sb.String()
	}
	return strYamlDefaultValues, nil
}

// Default does defaulting of x depending on default values in s.
// Based upon https://github.com/kubernetes/apiextensions-apiserver/blob/release-1.21/pkg/apiserver/schema/defaulting/algorithm.go
// Plus modifications from https://github.com/vmware-tanzu/tanzu-framework/pull/1422
// In short, it differs from upstream in that:
// -- 1. Prevent deep copy of int as it panics
// -- 2. For type object scan the first level properties for any defaults to create an empty map to populate
// -- 3. If the property does not have a default, add one based on the type ("", false, etc)
func defaultValues(x interface{}, s *structuralschema.Structural) {
	if s == nil {
		return
	}

	switch x := x.(type) {
	case map[string]interface{}:
		for k, prop := range s.Properties { //nolint
			// if Default for object is nil, scan first level of properties for any defaults to create an empty default
			if prop.Default.Object == nil {
				createDefault := false
				if prop.Properties != nil {
					for _, v := range prop.Properties { //nolint
						if v.Default.Object != nil {
							createDefault = true
							break
						}
					}
				}
				if createDefault {
					prop.Default.Object = make(map[string]interface{})
					// If not generating an empty object, fall back to the data type's defaults
				} else {
					switch prop.Type {
					case "string":
						prop.Default.Object = ""
					case "number":
						prop.Default.Object = 0
					case "integer":
						prop.Default.Object = 0
					case "boolean":
						prop.Default.Object = false
					case "array":
						prop.Default.Object = []interface{}{}
					case "object":
						prop.Default.Object = make(map[string]interface{})
					}
				}
			}
			if _, found := x[k]; !found || isNonNullableNull(x[k], &prop) {
				if isKindInt(prop.Default.Object) {
					x[k] = prop.Default.Object
				} else {
					x[k] = runtime.DeepCopyJSONValue(prop.Default.Object)
				}
			}
		}
		for k := range x {
			if prop, found := s.Properties[k]; found {
				defaultValues(x[k], &prop)
			} else if s.AdditionalProperties != nil {
				if isNonNullableNull(x[k], s.AdditionalProperties.Structural) {
					if isKindInt(s.AdditionalProperties.Structural.Default.Object) {
						x[k] = s.AdditionalProperties.Structural.Default.Object
					} else {
						x[k] = runtime.DeepCopyJSONValue(s.AdditionalProperties.Structural.Default.Object)
					}
				}
				defaultValues(x[k], s.AdditionalProperties.Structural)
			}
		}
	case []interface{}:
		for i := range x {
			if isNonNullableNull(x[i], s.Items) {
				if isKindInt(s.Items.Default.Object) {
					x[i] = s.Items.Default.Object
				} else {
					x[i] = runtime.DeepCopyJSONValue(s.Items.Default.Object)
				}
			}
			defaultValues(x[i], s.Items)
		}
	default:
		// scalars, do nothing
	}
}

// isNonNullalbeNull returns true if the item is nil AND it's nullable
func isNonNullableNull(x interface{}, s *structuralschema.Structural) bool {
	return x == nil && s != nil && !s.Generic.Nullable
}

// isKindInt returns true if the item is an int
func isKindInt(src interface{}) bool {
	return src != nil && reflect.TypeOf(src).Kind() == reflect.Int
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
