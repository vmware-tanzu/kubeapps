// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package utils

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/emicklei/go-restful-swagger12"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	ktesting "k8s.io/client-go/testing"
)

type FakeDiscovery struct {
	Fake     *ktesting.Fake
	delegate discovery.SwaggerSchemaInterface
}

func NewFakeDiscovery(delegate discovery.SwaggerSchemaInterface) *FakeDiscovery {
	fakePtr := &ktesting.Fake{}
	return &FakeDiscovery{
		Fake:     fakePtr,
		delegate: delegate,
	}
}

var _ discovery.ServerResourcesInterface = &FakeDiscovery{}

func (c *FakeDiscovery) ServerResourcesForGroupVersion(gv string) (*metav1.APIResourceList, error) {
	parsedGv, err := schema.ParseGroupVersion(gv)
	if err != nil {
		return nil, err
	}
	action := ktesting.ActionImpl{}
	action.Verb = "get"
	action.Resource = parsedGv.WithResource("apiresources")
	c.Fake.Invokes(action, nil)

	var rsrcs []metav1.APIResource
	switch gv {
	case "v1":
		rsrcs = []metav1.APIResource{
			{
				Name:       "configmaps",
				Kind:       "ConfigMap",
				Namespaced: true,
			},
			{
				Name:       "namespaces",
				Kind:       "Namespace",
				Namespaced: false,
			},
			{
				Name:       "replicationcontrollers",
				Kind:       "ReplicationController",
				Namespaced: true,
			},
		}
	default:
		return nil, fmt.Errorf("gv %v not found in test implementation", gv)
	}
	rsrc := metav1.APIResourceList{
		GroupVersion: gv,
		APIResources: rsrcs,
	}
	return &rsrc, nil
}

func (c *FakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	panic("unimplemented")
}

func (c *FakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	panic("unimplemented")
}

func (c *FakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	panic("unimplemented")
}

var _ discovery.SwaggerSchemaInterface = &FakeDiscovery{}

func (c *FakeDiscovery) SwaggerSchema(gv schema.GroupVersion) (*swagger.ApiDeclaration, error) {
	log.Debugf("SwaggerSchema(%v) called", gv)
	action := ktesting.ActionImpl{}
	action.Verb = "get"
	action.Resource = gv.WithResource("schema")
	c.Fake.Invokes(action, nil)

	return c.delegate.SwaggerSchema(gv)
}

func TestDepSort(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	disco := NewFakeDiscovery(schemaFromFile{dir: filepath.FromSlash("../testdata")})

	newObj := func(apiVersion, kind string) *unstructured.Unstructured {
		return &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": apiVersion,
				"kind":       kind,
			},
		}
	}

	objs := []*unstructured.Unstructured{
		newObj("v1", "ReplicationController"),
		newObj("v1", "ConfigMap"),
		newObj("v1", "Namespace"),
		newObj("bogus/v1", "UnknownKind"),
		newObj("apiextensions/v1beta1", "CustomResourceDefinition"),
	}

	sorter, err := DependencyOrder(disco, objs)
	if err != nil {
		t.Fatalf("DependencyOrder error: %v", err)
	}
	sort.Sort(sorter)

	for i, o := range objs {
		t.Logf("obj[%d] after sort is %v", i, o.GroupVersionKind())
	}

	if objs[0].GetKind() != "CustomResourceDefinition" {
		t.Error("CRD should be sorted first")
	}
	if objs[1].GetKind() != "Namespace" {
		t.Error("Namespace should be sorted second")
	}
	if objs[4].GetKind() != "ReplicationController" {
		t.Error("RC should be sorted after other objects")
	}
}

func TestAlphaSort(t *testing.T) {
	newObj := func(ns, name, kind string) *unstructured.Unstructured {
		o := unstructured.Unstructured{}
		o.SetNamespace(ns)
		o.SetName(name)
		o.SetKind(kind)
		return &o
	}

	objs := []*unstructured.Unstructured{
		newObj("default", "mysvc", "Deployment"),
		newObj("", "default", "StorageClass"),
		newObj("", "default", "ClusterRole"),
		newObj("default", "mydeploy", "Deployment"),
		newObj("default", "mysvc", "Secret"),
	}

	expected := []*unstructured.Unstructured{
		objs[2],
		objs[1],
		objs[3],
		objs[0],
		objs[4],
	}

	sort.Sort(AlphabeticalOrder(objs))

	if !reflect.DeepEqual(objs, expected) {
		t.Errorf("actual != expected: %v != %v", objs, expected)
	}
}
