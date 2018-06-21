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
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCanI(t *testing.T) {
	resourceList := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{Name: "pods", Kind: "Pod"},
		},
	}
	cli := fake.NewSimpleClientset()
	fakeDiscovery, ok := cli.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
	}
	fakeDiscovery.Resources = []*metav1.APIResourceList{&resourceList}
	auth := UserAuth{
		authCli:      cli.AuthorizationV1(),
		discoveryCli: cli.Discovery(),
	}
	manifest := `---
apiVersion: v1
kind: Pod
`
	err := auth.CanI("foo", "create", manifest)
	// Fake client returns an empty result so it will deny any request
	if !strings.Contains(err.Error(), "Unauthorized to create v1/pods in the foo namespace") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestNamespacedCanI(t *testing.T) {
	resourceList := metav1.APIResourceList{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{
			{Name: "pods", Kind: "Pod"},
		},
	}
	cli := fake.NewSimpleClientset()
	fakeDiscovery, ok := cli.Discovery().(*fakediscovery.FakeDiscovery)
	if !ok {
		t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
	}
	fakeDiscovery.Resources = []*metav1.APIResourceList{&resourceList}
	auth := UserAuth{
		authCli:      cli.AuthorizationV1(),
		discoveryCli: cli.Discovery(),
	}
	manifest := `---
apiVersion: v1
kind: Pod
metadata:
  namespace: bar
`
	err := auth.CanI("foo", "create", manifest)
	// Fake client returns an empty result so it will deny any request
	if !strings.Contains(err.Error(), "Unauthorized to create v1/pods in the bar namespace") {
		t.Errorf("Unexpected error: %v", err)
	}
}
