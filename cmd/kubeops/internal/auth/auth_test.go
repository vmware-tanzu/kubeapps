// Copyright 2018-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestParseForbiddenActions(t *testing.T) {
	testSuite := []struct {
		Description     string
		Error           string
		ExpectedActions []Action
	}{
		{
			"parses an error with a single resource",
			`User "foo" cannot create resource "secrets" in API group "" in the namespace "default"`,
			[]Action{
				{APIVersion: "", Resource: "secrets", Namespace: "default", Verbs: []string{"create"}},
			},
		},
		{
			"parses an error with a cluster-wide resource",
			`User "foo" cannot create resource "clusterroles" in API group "v1"`,
			[]Action{
				{APIVersion: "v1", Resource: "clusterroles", Namespace: "", Verbs: []string{"create"}, ClusterWide: true},
			},
		},
		{
			"parses several resources",
			`User "foo" cannot create resource "secrets" in API group "" in the namespace "default";
			User "foo" cannot create resource "pods" in API group "" in the namespace "default"`,
			[]Action{
				{APIVersion: "", Resource: "secrets", Namespace: "default", Verbs: []string{"create"}},
				{APIVersion: "", Resource: "pods", Namespace: "default", Verbs: []string{"create"}},
			},
		},
		{
			"includes different verbs and remove duplicates",
			`User "foo" cannot create resource "secrets" in API group "" in the namespace "default";
			User "foo" cannot create resource "secrets" in API group "" in the namespace "default";
			User "foo" cannot delete resource "secrets" in API group "" in the namespace "default"`,
			[]Action{
				{APIVersion: "", Resource: "secrets", Namespace: "default", Verbs: []string{"create", "delete"}},
			},
		},
	}
	for _, test := range testSuite {
		t.Run(test.Description, func(t *testing.T) {
			actions := ParseForbiddenActions(test.Error)
			// order actions by resource
			less := func(x, y Action) bool { return strings.Compare(x.Resource, y.Resource) < 0 }
			if !cmp.Equal(actions, test.ExpectedActions, cmpopts.SortSlices(less)) {
				t.Errorf("Unexpected forbidden actions: %v", cmp.Diff(actions, test.ExpectedActions))
			}
		})
	}
}
