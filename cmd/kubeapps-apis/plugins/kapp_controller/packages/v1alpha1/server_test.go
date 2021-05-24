/*
Copyright © 2021 VMware
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

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name         string
		clientGetter func(context.Context) (dynamic.Interface, error)
		statusCode   codes.Code
	}{
		{
			name:         "returns internal error status when no getter configured",
			clientGetter: nil,
			statusCode:   codes.Internal,
		},
		{
			name: "returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return nil, fmt.Errorf("Bang!")
			},
			statusCode: codes.FailedPrecondition,
		},
		{
			name: "returns an internal error status if response does not contain publicName",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"version": "1.2.3",
					}),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns an internal error status if response does not contain version",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"publicName": "someName",
					}),
				), nil
			},
			statusCode: codes.Internal,
		},
		{
			name: "returns without error if items contain required fields",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
					},
					packageFromSpec(map[string]interface{}{
						"publicName": "someName",
						"version":    "1.2.3",
					}),
				), nil
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			_, err := s.GetAvailablePackages(context.Background(), &corev1.GetAvailablePackagesRequest{})

			if err == nil && tc.statusCode != codes.OK {
				t.Fatalf("got: nil, want: error")
			}

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}
		})
	}

}

func packageFromSpec(spec map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", packageGroup, packageVersion),
			"kind":       "Package",
			"metadata": map[string]interface{}{
				"name": fmt.Sprintf("%s.%s", spec["publicName"], spec["version"]),
			},
			"spec": spec,
		},
	}
}

func packagesFromSpecs(specs []map[string]interface{}) []runtime.Object {
	pkgs := []runtime.Object{}
	for _, s := range specs {
		pkgs = append(pkgs, packageFromSpec(s))
	}
	return pkgs
}

func TestGetAvailablePackages(t *testing.T) {
	testCases := []struct {
		name             string
		packageSpecs     []map[string]interface{}
		expectedPackages []*corev1.AvailablePackage
	}{
		{
			name: "it returns carvel packages from the cluster",
			packageSpecs: []map[string]interface{}{
				{
					"publicName": "tetris.foo.example.com",
					"version":    "1.2.3",
				},
				{
					"publicName": "another.foo.example.com",
					"version":    "1.2.5",
				},
			},
			expectedPackages: []*corev1.AvailablePackage{
				{
					Name:    "another.foo.example.com",
					Version: "1.2.5",
				},
				{
					Name:    "tetris.foo.example.com",
					Version: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pkgs := packagesFromSpecs(tc.packageSpecs)
			s := Server{
				clientGetter: func(context.Context) (dynamic.Interface, error) {
					return fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: packageGroup, Version: packageVersion, Resource: packagesResource}: "PackageList",
						},
						pkgs...,
					), nil
				},
			}

			response, err := s.GetAvailablePackages(context.Background(), &corev1.GetAvailablePackagesRequest{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response.Packages, tc.expectedPackages; !cmp.Equal(got, want, cmp.Comparer(pkgEqual)) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmp.Comparer(pkgEqual)))
			}
		})
	}

}

func pkgEqual(a, b *corev1.AvailablePackage) bool {
	return a.Name == b.Name && a.Version == b.Version
}
