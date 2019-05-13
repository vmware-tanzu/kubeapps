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
	// Fake write permissions for pods
	if resource == "pods" {
		return true, nil
	}
	// Fake permissions for clusterroles in any version
	if resource == "clusterroles" && group == "rbac.authorization.k8s.io" {
		return true, nil
	}
	return false, nil
}

func newFakeUserAuth() *UserAuth {
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
	fakeK8sAuthCli := fakeK8sAuth{cli.Discovery()}
	return &UserAuth{fakeK8sAuthCli}
}

func TestGetForbidden(t *testing.T) {
	type test struct {
		Action          string
		Namespace       string
		Manifest        string
		ExpectedActions []Action
	}
	testSuite := []test{
		// It should be able to create pods
		{
			Action:    "create",
			Namespace: "foo",
			Manifest: `---
apiVersion: v1
kind: Pod
`,
			ExpectedActions: []Action{},
		},
		// It shouldn't be able to create deployments
		{
			Action:    "create",
			Namespace: "foo",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: "foo", Verbs: []string{"create"}},
			},
		},
		// It should overwrite the default namespace
		{
			Action:    "create",
			Namespace: "foo",
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
		// It should report the same resource in different resource groups
		{
			Action:    "create",
			Namespace: "foo",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
---
apiVersion: extensions/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: "foo", Verbs: []string{"create"}},
				{APIVersion: "extensions/v1beta1", Resource: "deployments", Namespace: "foo", Verbs: []string{"create"}},
			},
		},
		// It should report the same resource with different verbs when upgrading
		{
			Action:    "upgrade",
			Namespace: "foo",
			Manifest: `---
apiVersion: apps/v1beta1
kind: Deployment
`,
			ExpectedActions: []Action{
				{APIVersion: "apps/v1beta1", Resource: "deployments", Namespace: "foo", Verbs: []string{"create", "update", "delete"}},
			},
		},
		// It should allow unversioned clusterroles
		{
			Action:    "get",
			Namespace: "foo",
			Manifest: `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
`,
			ExpectedActions: []Action{},
		},
		// It should report if a resource is clusterWide
		{
			Action:    "get",
			Namespace: "foo",
			Manifest: `---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
`,
			ExpectedActions: []Action{
				{APIVersion: "rbac.authorization.k8s.io/v1", Resource: "clusterrolebindings", ClusterWide: true, Verbs: []string{"get"}},
			},
		},
		// It should allow an unrecognized resource, so that CRDs can be installed before CRs
		{
			Action:    "get",
			Namespace: "foo",
			Manifest: `---
apiVersion: foo.bar.io/v1
kind: FooBar
`,
			ExpectedActions: []Action{},
		},
	}
	for _, tt := range testSuite {
		auth := newFakeUserAuth()
		res, err := auth.GetForbiddenActions(tt.Namespace, tt.Action, tt.Manifest)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !reflect.DeepEqual(res, tt.ExpectedActions) {
			t.Errorf("Expecting %v, received %v", tt.ExpectedActions, res)
		}
	}
}
