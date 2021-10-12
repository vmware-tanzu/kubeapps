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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetResources(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *corev1.GetResourceRefsRequest
		expectedResponse   *corev1.GetResourceRefsResponse
		expectedStatusCode codes.Code
	}{
		{
			name: "returns resource references for helm installation",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
# Source: apache/templates/svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: apache-test
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
`,
				},
			},
			request: &corev1.GetResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetResourceRefsResponse{
				ResourceRefs: []*corev1.ResourceRef{
					{
						Name: "apache-test",
						Kind: "Service",
						Context: &corev1.Context{
							Cluster:   "default",
							Namespace: "default",
						},
					},
					{
						Name: "apache-test",
						Kind: "Deployment",
						Context: &corev1.Context{
							Cluster:   "default",
							Namespace: "default",
						},
					},
				},
			},
		},
		{
			name: "skips resources that do not have a kind (such as resource-lists)",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
# Source: apache/templates/svc.yaml
apiVersion: v1
otherstuff: ignored
metadata:
  name: apache-test
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
`,
				},
			},
			request: &corev1.GetResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetResourceRefsResponse{
				ResourceRefs: []*corev1.ResourceRef{
					{
						Name: "apache-test",
						Kind: "Deployment",
						Context: &corev1.Context{
							Cluster:   "default",
							Namespace: "default",
						},
					},
				},
			},
		},
		{
			name: "returns a not found error if the release is not found",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
				},
			},
			request: &corev1.GetResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-iis",
				},
			},
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "returns internal error if the yaml manifest cannot be parsed",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
apiVersion: v1
should not be :! parsed as yaml$
`,
				},
			},
			request: &corev1.GetResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedStatusCode: codes.Internal,
		},
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.GetResourceRefsResponse{},
		corev1.ResourceRef{},
		corev1.Context{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases)

			server, _, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()

			response, err := server.GetResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
