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
		request            *corev1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *corev1.GetInstalledPackageResourceRefsResponse
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
  namespace: default
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
  namespace: default
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "v1",
						Name:       "apache-test",
						Namespace:  "default",
						Kind:       "Service",
					},
					{
						ApiVersion: "apps/v1",
						Name:       "apache-test",
						Namespace:  "default",
						Kind:       "Deployment",
					},
				},
			},
		},
		{
			name: "returns resource references for resources in other namespaces",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
apiVersion: v1
kind: ClusterRole
metadata:
  name: test-cluster-role
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-other-namespace
  namespace: some-other-namespace
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "v1",
						Name:       "test-cluster-role",
						Kind:       "ClusterRole",
					},
					{
						ApiVersion: "apps/v1",
						Name:       "test-other-namespace",
						Namespace:  "some-other-namespace",
						Kind:       "Deployment",
					},
				},
			},
		},
		{
			name: "skips resources that do not have a kind",
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
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Name:       "apache-test",
						Kind:       "Deployment",
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
			request: &corev1.GetInstalledPackageResourceRefsRequest{
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
			request: &corev1.GetInstalledPackageResourceRefsRequest{
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
		{
			name: "handles duplicate labels as helm does",
			// See https://github.com/kubeapps/kubeapps/issues/632
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: apache-test
  label:
    chart: "gitea-0.2.0"
    chart: "gitea-0.2.0"
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Name:       "apache-test",
						Kind:       "Deployment",
					},
				},
			},
		},
		{
			name: "supports manifests with YAML type casting",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
# Source: apache/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: !!string apache-test
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Name:       "apache-test",
						Kind:       "Deployment",
					},
				},
			},
		},
		{
			name: "renders a list of items",
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
apiVersion: v1
kind: List
items:
- apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: apache-test
    namespace: default
- apiVersion: v1
  kind: Service
  metadata:
    name: apache-test
    namespace: default
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Name:       "apache-test",
						Namespace:  "default",
						Kind:       "Deployment",
					},
					{
						ApiVersion: "v1",
						Name:       "apache-test",
						Namespace:  "default",
						Kind:       "Service",
					},
				},
			},
		},
		{
			name: "renders a rolelist of items",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
apiVersion: v1
kind: RoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-1
    namespace: default
- apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: role-2
    namespace: default
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "rbac.authorization.k8s.io/v1",
						Name:       "role-1",
						Namespace:  "default",
						Kind:       "Role",
					},
					{
						ApiVersion: "rbac.authorization.k8s.io/v1",
						Name:       "role-2",
						Namespace:  "default",
						Kind:       "Role",
					},
				},
			},
		},
		{
			name: "renders a ClusterRoleList of items",
			// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
			existingReleases: []releaseStub{
				{
					name:      "my-apache",
					namespace: "default",
					manifest: `
---
apiVersion: v1
kind: ClusterRoleList
items:
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-1
- apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: clusterrole-2
`,
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "rbac.authorization.k8s.io/v1",
						Name:       "clusterrole-1",
						Kind:       "ClusterRole",
					},
					{
						ApiVersion: "rbac.authorization.k8s.io/v1",
						Name:       "clusterrole-2",
						Kind:       "ClusterRole",
					},
				},
			},
		},
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageResourceRefsResponse{},
		corev1.ResourceRef{},
		corev1.Context{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)

			server, _, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()

			response, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
