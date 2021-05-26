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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
			name: "returns without error if response status does not contain conditions",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", map[string]interface{}{}, map[string]interface{}{}),
				), nil
			},
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain ready repos",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", map[string]interface{}{}, map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
								"reason": "IndexationFailed",
							},
						}}),
				), nil
			},
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain status url",
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return fake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
					},
					newRepo("test", map[string]interface{}{}, map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
								"reason": "IndexationSucceed",
							},
						}}),
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

func newRepo(name string, spec map[string]interface{}, status map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": fmt.Sprintf("%s/%s", fluxGroup, fluxVersion),
			"kind":       fluxHelmRepository,
			"metadata": map[string]interface{}{
				"name": name,
			},
			"spec":   spec,
			"status": status,
		},
	}
}

func TestGetAvailablePackages(t *testing.T) {
	testCases := []struct {
		testName         string
		repoName         string
		repoUrl          string
		repoIndex        string
		expectedPackages []*corev1.AvailablePackage
	}{
		{
			testName:  "it returns a couple of fluxv2 packages from the cluster",
			repoName:  "bitnami-1",
			repoUrl:   "https://example.repo.com/charts",
			repoIndex: "testdata/valid-index.yaml",
			expectedPackages: []*corev1.AvailablePackage{
				{
					Name:       "acs-engine-autoscaler",
					Version:    "2.1.1",
					IconUrl:    "https://github.com/kubernetes/kubernetes/blob/master/logo/logo.png",
					Repository: &corev1.AvailablePackage_PackageRepositoryReference{Name: "bitnami-1"},
				},
				{
					Name:       "wordpress",
					Version:    "0.7.5",
					IconUrl:    "https://bitnami.com/assets/stacks/wordpress/img/wordpress-stack-220x234.png",
					Repository: &corev1.AvailablePackage_PackageRepositoryReference{Name: "bitnami-1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			indexYAMLBytes, err := ioutil.ReadFile(tc.repoIndex)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			// stand up an http server just for the duration of this test
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, string(indexYAMLBytes))
			}))
			defer ts.Close()

			repoSpec := map[string]interface{}{
				"name": tc.repoName,
			}
			repoStatus := map[string]interface{}{
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
						"reason": "IndexationSucceed",
					},
				},
				"url": ts.URL,
			}
			repo := newRepo(tc.repoName, repoSpec, repoStatus)
			s := Server{
				clientGetter: func(context.Context) (dynamic.Interface, error) {
					return fake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: fluxGroup, Version: fluxVersion, Resource: fluxHelmRepositories}: fluxHelmRepositoryList,
						},
						repo,
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
	return a.Name == b.Name && a.Version == b.Version && a.IconUrl == b.IconUrl && a.Repository.Name == b.Repository.Name
}
