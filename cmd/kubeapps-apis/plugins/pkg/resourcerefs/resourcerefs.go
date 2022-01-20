/*
Copyright © 2022 VMware
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
package resourcerefs

import (
	goerrs "errors"
	"io"
	"strings"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type yamlMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type yamlResource struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   yamlMetadata   `json:"metadata"`
	Items      []yamlResource `json:"items"`
}

// resourceRefsFromManifest returns the resource refs for a given yaml manifest.
func ResourceRefsFromManifest(m, pkgNamespace string) ([]*corev1.ResourceRef, error) {
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(m))
	refs := []*corev1.ResourceRef{}
	doc := yamlResource{}
	for {
		err := decoder.Decode(&doc)
		if err != nil {
			if goerrs.Is(err, io.EOF) {
				break
			}
			return nil, status.Errorf(codes.Internal, "Unable to decode yaml manifest: %v", err)
		}
		if doc.Kind == "" {
			continue
		}
		if doc.Kind == "List" || doc.Kind == "RoleList" || doc.Kind == "ClusterRoleList" {
			for _, i := range doc.Items {
				namespace := i.Metadata.Namespace
				if namespace == "" {
					namespace = pkgNamespace
				}
				refs = append(refs, &corev1.ResourceRef{
					ApiVersion: i.APIVersion,
					Kind:       i.Kind,
					Name:       i.Metadata.Name,
					Namespace:  namespace,
				})
			}
			continue
		}
		// Helm does not require that the rendered manifest specifies the
		// resource namespace so some charts do not do so (ldap).  We explicitly
		// set the namespace for the resource ref so that it can be used as part
		// of the key for the resource ref.
		// TODO(minelson): At the moment we do not distinguish between
		// cluster-scoped and namespace-scoped resources for the refs.  This
		// does not affect the resources plugin fetching them correctly, but
		// would be better if we only set the namespace in the reference if (a)
		// it was not set in the manifest, and (b) it is a namespace-scoped
		// resource.
		namespace := doc.Metadata.Namespace
		if namespace == "" {
			namespace = pkgNamespace
		}
		refs = append(refs, &corev1.ResourceRef{
			ApiVersion: doc.APIVersion,
			Kind:       doc.Kind,
			Name:       doc.Metadata.Name,
			Namespace:  namespace,
		})
	}

	return refs, nil
}

// this is done so that test scenarios can be re-used in another package (helm and flux plug-ins)
// ref: https://stackoverflow.com/questions/28476307/how-to-get-test-environment-at-run-time
type TestReleaseStub struct {
	Name      string
	Namespace string
	Manifest  string
}

type TestCase struct {
	Name                  string
	ExistingReleases      []TestReleaseStub
	ExpectedResourceRefs  []*corev1.ResourceRef
	ExpectedErrStatusCode codes.Code
}

var (
	// will be properly initialized in resourcerefs_test.go init()
	TestCases1, TestCases2 = []TestCase(nil), []TestCase(nil)
)
