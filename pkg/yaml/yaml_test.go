/*
Copyright (c) 2018 Bitnami

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

package yaml

import (
	"testing"
)

func TestParseObjectsSuccess(t *testing.T) {
	testCases := []struct {
		desc            string
		manifest        string
		numberResources int
		apiVersions     []string
		kinds           []string
	}{
		{
			"returns nothing if manifest is empty",
			"",
			0, nil, nil,
		},
		{
			"returns a single resource",
			`
apiVersion: v1
kind: Namespace
metadata:
  name: kubeapps`,
			1, []string{"v1"}, []string{"Namespace"},
		},
		{
			"returns multiple resources",
			`
apiVersion: v1
kind: Namespace
metadata:
  name: kubeapps
---
apiVersion: extensions/v1beta1       
kind: Deployment
metadata:
  name: kubeapps`,
			2, []string{"v1", "extensions/v1beta1"}, []string{"Namespace", "Deployment"},
		},
		{
			"ignores files with just comments",
			`
apiVersion: v1
kind: LonelyNamespace
metadata:
  name: kubeapps
---
# This is a comment in yaml`,
			1, []string{"v1"}, []string{"LonelyNamespace"},
		},
		{
			"ignores empty files",
			`
apiVersion: v1
kind: LonelyNamespace
metadata:
  name: kubeapps
---
---
`,
			1, []string{"v1"}, []string{"LonelyNamespace"},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.desc, func(t *testing.T) {
			resources, err := ParseObjects(tt.manifest)
			if err != nil {
				t.Error(err)
			}

			if got, want := len(resources), tt.numberResources; got != want {
				t.Errorf("Expected %d yaml element, got %v", want, got)
			}

			for i, resource := range resources {
				if got, want := resource.GetAPIVersion(), tt.apiVersions[i]; got != want {
					t.Errorf("got %q, want %q", got, want)

				}
				if got, want := resource.GetAPIVersion(), tt.apiVersions[i]; got != want {
					t.Errorf("got %q, want %q", got, want)
				}
				if got, want := resource.GetKind(), tt.kinds[i]; got != want {
					t.Errorf("got %q, want %q", got, want)
				}
			}
		})
	}
}

func TestParseObjectFailure(t *testing.T) {
	m2 := `apiVersion: v1
kind: Namespace
metadata:
    annotations: {}
  labels:
    name: kubeless
  name: kubeless`
	_, err := ParseObjects(m2)
	if err == nil {
		t.Error("Expected parse fail, got success")
	}
}
