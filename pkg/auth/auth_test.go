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

package auth

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"reflect"
	"strings"
	"testing"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	discovery "k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

type fakeK8sAuth struct {
	DiscoveryCli discovery.DiscoveryInterface
	canIResult   bool
	canIError    error
}

func (u fakeK8sAuth) Validate() error {
	return nil
}
func (u fakeK8sAuth) GetResourceList(groupVersion string) (*metav1.APIResourceList, error) {
	g, err := u.DiscoveryCli.ServerResourcesForGroupVersion(groupVersion)
	if err != nil && strings.Contains(err.Error(), "not found") {
		// Fake DiscoveryCli doesn't return a valid NotFound error so we need to forge it
		err = k8sErrors.NewNotFound(schema.GroupResource{}, groupVersion)
	}
	return g, err
}

func (u fakeK8sAuth) CanI(verb, group, resource, namespace string) (bool, error) {
	return u.canIResult, u.canIError
}

func newFakeUserAuth(canIResult bool, canIError error) *UserAuth {
	resourceListV1 := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{Name: "pods", Kind: "Pod", Namespaced: true},
		},
	}
	resourceListAppsV1Beta1 := metav1.APIResourceList{
		GroupVersion: "apps/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "deployments", Kind: "Deployment", Namespaced: true},
		},
	}
	resourceListExtensionsV1Beta1 := metav1.APIResourceList{
		GroupVersion: "extensions/v1beta1",
		APIResources: []metav1.APIResource{
			{Name: "deployments", Kind: "Deployment", Namespaced: true},
		},
	}
	resourceListClusterRoleRBAC := metav1.APIResourceList{
		GroupVersion: "rbac.authorization.k8s.io/v1",
		APIResources: []metav1.APIResource{
			{Name: "clusterrolebindings", Kind: "ClusterRoleBinding", Namespaced: false},
			{Name: "clusterroles", Kind: "ClusterRole", Namespaced: false},
		},
	}
	cli := fake.NewSimpleClientset()
	fakeDiscovery, _ := cli.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		&resourceListV1,
		&resourceListAppsV1Beta1,
		&resourceListExtensionsV1Beta1,
		&resourceListClusterRoleRBAC,
	}
	fakeK8sAuthCli := fakeK8sAuth{cli.Discovery(), canIResult, canIError}
	return &UserAuth{fakeK8sAuthCli}
}

func TestGetForbidden(t *testing.T) {
	const namespace = "test-namspace"
	type test struct {
		Name            string
		Action          string
		CanIResult      bool
		Manifest        string
		ExpectedActions []Action
	}
	testSuite := []test{
		{
			Name:       "it should be able to create pods",
			Action:     "create",
			CanIResult: true,
			Manifest: `---
apiVersion: v1
kind: Pod
`,
			ExpectedActions: []Action{},
		},
		{
			Name:   "it shouldn't be able to create deployments",
			Action: "create",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: namespace, Verbs: []string{"create"}},
			},
		},
		{
			Name:   "it should overwrite the default namespace",
			Action: "create",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  namespace: bar
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: "bar", Verbs: []string{"create"}},
			},
		},
		{
			Name:   "it should report the same resource in different resource groups",
			Action: "create",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
---
apiVersion: extensions/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: namespace, Verbs: []string{"create"}},
				{APIVersion: "extensions/v1beta1", Resource: "deployments", Namespace: namespace, Verbs: []string{"create"}},
			},
		},
		{
			Name:   "it should report the same resource with different verbs when upgrading",
			Action: "upgrade",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: namespace, Verbs: []string{"create", "update", "delete"}},
			},
		},
		{
			Name:       "it should allow unversioned clusterroles",
			Action:     "get",
			CanIResult: true,
			Manifest: `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
`,
			ExpectedActions: []Action{},
		},
		{
			Name:   "it should report if a resource is clusterWide",
			Action: "get",
			Manifest: `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
`,
			ExpectedActions: []Action{
				{APIVersion: "rbac.authorization.k8s.io/v1", Resource: "clusterrolebindings", ClusterWide: true, Verbs: []string{"get"}},
			},
		},
		{
			Name:   "it should allow an unrecognized resource, so that CRDs can be installed before CRs",
			Action: "get",
			Manifest: `---
apiVersion: foo.bar.io/v1
kind: FooBar
`,
			ExpectedActions: []Action{},
		},
	}
	for _, tt := range testSuite {
		t.Run(tt.Name, func(t *testing.T) {
			auth := newFakeUserAuth(tt.CanIResult, nil)
			res, err := auth.GetForbiddenActions(namespace, tt.Action, tt.Manifest)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(res, tt.ExpectedActions) {
				t.Errorf("Expecting %v, received %v", tt.ExpectedActions, res)
			}
		})
	}
}

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
