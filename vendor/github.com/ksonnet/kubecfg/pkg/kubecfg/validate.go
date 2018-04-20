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

package kubecfg

import (
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"

	"github.com/ksonnet/kubecfg/utils"
)

// ValidateCmd represents the validate subcommand
type ValidateCmd struct {
	Discovery     discovery.DiscoveryInterface
	IgnoreUnknown bool
}

func (c ValidateCmd) Run(apiObjects []*unstructured.Unstructured, out io.Writer) error {
	knownGVKs := sets.NewString()
	gvkExists := func(gvk schema.GroupVersionKind) bool {
		if knownGVKs.Has(gvk.String()) {
			return true
		}
		gv := gvk.GroupVersion()
		rls, err := c.Discovery.ServerResourcesForGroupVersion(gv.String())
		if err != nil {
			if !errors.IsNotFound(err) {
				log.Debugf("ServerResourcesForGroupVersion(%q) returned unexpected error %v", gv, err)
			}
			return false
		}
		for _, rl := range rls.APIResources {
			knownGVKs.Insert(gv.WithKind(rl.Kind).String())
		}
		return knownGVKs.Has(gvk.String())
	}

	hasError := false

	for _, obj := range apiObjects {
		desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(c.Discovery, obj), utils.FqName(obj))
		log.Info("Validating ", desc)

		gvk := obj.GroupVersionKind()
		gv := gvk.GroupVersion()

		var allErrs []error

		schema, err := utils.NewSwaggerSchemaFor(c.Discovery, gv)
		if err != nil {
			isNotFound := errors.IsNotFound(err) ||
				strings.Contains(err.Error(), "is not supported by the server")
			if isNotFound && (c.IgnoreUnknown || gvkExists(gvk)) {
				log.Infof(" No schema found for %s, skipping validation", gvk)
				continue
			}
			allErrs = append(allErrs, fmt.Errorf("Unable to fetch schema: %v", err))
		} else {
			// Validate obj
			for _, err := range schema.Validate(obj) {
				_, isNotFound := err.(utils.TypeNotFoundError)
				if isNotFound && (c.IgnoreUnknown || gvkExists(gvk)) {
					log.Infof(" Found apiGroup, but it did not contain a schema for %s, ignoring", gvk)
					continue
				}
				allErrs = append(allErrs, err)
			}
		}

		for _, err := range allErrs {
			log.Errorf("Error in %s: %v", desc, err)
			hasError = true
		}
	}

	if hasError {
		return fmt.Errorf("Validation failed")
	}

	return nil
}
