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
	"errors"
	"fmt"
	"sort"

	"github.com/emicklei/go-restful-swagger12"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
)

var (
	gkTpr = schema.GroupKind{Group: "extensions", Kind: "ThirdPartyResource"}
	gkCrd = schema.GroupKind{Group: "apiextensions", Kind: "CustomResourceDefinition"}
)

// Heh: Swagger. Walk. :P
func swaggerWalk(api *swagger.ApiDeclaration, typeName string, seen sets.String, visitor func(string) error) error {
	if !versionRegexp.MatchString(typeName) {
		return nil
	}

	// Prevent possible schema loops
	if seen.Has(typeName) {
		return nil
	}
	seen.Insert(typeName)

	if err := visitor(typeName); err != nil {
		return err
	}

	// is an API object of some sort
	models := api.Models
	model, ok := models.At(typeName)
	if !ok {
		return fmt.Errorf("type %q not found in schema", typeName)
	}

	properties := model.Properties
	for _, namedproperty := range properties.List {
		details := namedproperty.Property
		var fieldType string
		if details.Type != nil {
			fieldType = *details.Type
		} else if details.Ref != nil {
			fieldType = *details.Ref
		} else {
			return fmt.Errorf("type of field %q in %q undefined in schema", namedproperty.Name, typeName)
		}

		if fieldType == "array" {
			if details.Items.Ref != nil {
				fieldType = *details.Items.Ref
			} else if details.Items.Type != nil {
				fieldType = *details.Items.Type
			} else {
				return fmt.Errorf("array type of %q undefined in schema", fieldType)
			}
		}

		if err := swaggerWalk(api, fieldType, seen, visitor); err != nil {
			return err
		}
	}

	return nil
}

var podSpecCache = map[string]bool{}

func containsPodSpec(disco discovery.SwaggerSchemaInterface, gvk schema.GroupVersionKind) bool {
	if result, ok := podSpecCache[gvk.String()]; ok {
		return result
	}

	swagger, err := disco.SwaggerSchema(gvk.GroupVersion())
	if err != nil {
		// Indeterminate result.
		log.Debugf("Unable to fetch schema for %s (%v), assuming not a PodSpec", gvk, err)
		return false
	}

	foundErr := errors.New("Found") // not really an error ...
	visitor := func(typeName string) error {
		if typeName == "v1.PodSpec" {
			return foundErr
		}
		return nil
	}

	var result bool

	typeName := fmt.Sprintf("%s.%s", gvk.Version, gvk.Kind)
	err = swaggerWalk(swagger, typeName, sets.NewString(), visitor)
	switch err {
	case nil:
		result = false
	case foundErr:
		result = true
	default:
		// Indeterminate result, but repeatable so may as well cache it
		log.Debugf("Error walking swagger schema for %q: %v", typeName, err)
		result = false
	}

	podSpecCache[gvk.String()] = result
	return result
}

// Arbitrary numbers used to do a simple topological sort of resources.
func depTier(disco ServerResourcesSwaggerSchema, o schema.ObjectKind) (int, error) {
	gvk := o.GroupVersionKind()
	if gk := gvk.GroupKind(); gk == gkTpr || gk == gkCrd {
		// Special case: these create other types
		return 10, nil
	}

	rsrc, err := serverResourceForGroupVersionKind(disco, gvk)
	if err != nil {
		log.Debugf("unable to fetch resource for %s (%v), continuing", gvk, err)
		return 50, nil
	}

	if !rsrc.Namespaced {
		// Place global before namespaced
		return 20, nil
	} else if containsPodSpec(disco, gvk) {
		// (Potentially) starts a pod, so place last
		return 100, nil
	} else {
		// Everything else
		return 50, nil
	}
}

// A subset of discovery.DiscoveryInterface
type ServerResourcesSwaggerSchema interface {
	discovery.ServerResourcesInterface
	discovery.SwaggerSchemaInterface
}

// DependencyOrder is a `sort.Interface` that *best-effort* sorts the
// objects so that known dependencies appear earlier in the list.  The
// idea is to prevent *some* of the "crash-restart" loops when
// creating inter-dependent resources.
func DependencyOrder(disco ServerResourcesSwaggerSchema, list []*unstructured.Unstructured) (sort.Interface, error) {
	sortKeys := make([]int, len(list))
	for i, item := range list {
		var err error
		sortKeys[i], err = depTier(disco, item.GetObjectKind())
		if err != nil {
			return nil, err
		}
	}
	log.Debugf("sortKeys is %v", sortKeys)
	return &mappedSort{sortKeys: sortKeys, items: list}, nil
}

type mappedSort struct {
	sortKeys []int
	items    []*unstructured.Unstructured
}

func (l *mappedSort) Len() int { return len(l.items) }
func (l *mappedSort) Swap(i, j int) {
	l.sortKeys[i], l.sortKeys[j] = l.sortKeys[j], l.sortKeys[i]
	l.items[i], l.items[j] = l.items[j], l.items[i]
}
func (l *mappedSort) Less(i, j int) bool {
	if l.sortKeys[i] != l.sortKeys[j] {
		return l.sortKeys[i] < l.sortKeys[j]
	}
	// Fall back to alpha sort, to give persistent order
	return AlphabeticalOrder(l.items).Less(i, j)
}

// AlphabeticalOrder is a `sort.Interface` that sorts the
// objects by namespace/name/kind alphabetical order
type AlphabeticalOrder []*unstructured.Unstructured

func (l AlphabeticalOrder) Len() int      { return len(l) }
func (l AlphabeticalOrder) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l AlphabeticalOrder) Less(i, j int) bool {
	a, b := l[i], l[j]

	if a.GetNamespace() != b.GetNamespace() {
		return a.GetNamespace() < b.GetNamespace()
	}
	if a.GetName() != b.GetName() {
		return a.GetName() < b.GetName()
	}
	return a.GetKind() < b.GetKind()
}
