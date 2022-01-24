// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

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
