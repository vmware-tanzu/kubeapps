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
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetInstalledPackageResourceRefs(t *testing.T) {

	// sanity check
	if len(resourcerefstest.TestCases1) < 11 {
		t.Fatalf("Expected array [resourcerefstest.TestCases1] size of at least 11")
	}

	type TestCase struct {
		baseTestCase       resourcerefstest.TestCase
		request            *corev1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *corev1.GetInstalledPackageResourceRefsResponse
		expectedStatusCode codes.Code
	}

	newTestCase := func(tc int, identifier string, response bool, code codes.Code) TestCase {
		newCase := TestCase{
			baseTestCase: resourcerefstest.TestCases1[tc],
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: identifier,
				},
			},
		}
		if response {
			newCase.expectedResponse = &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: resourcerefstest.TestCases1[tc].ExpectedResourceRefs,
			}
		}
		newCase.expectedStatusCode = code
		return newCase
	}

	testCases := []TestCase{
		newTestCase(0, "my-apache", true, codes.OK),
		newTestCase(1, "my-apache", true, codes.OK),
		newTestCase(2, "my-apache", true, codes.OK),
		newTestCase(3, "my-apache", true, codes.OK),
		newTestCase(4, "my-iis", false, codes.NotFound),
		newTestCase(5, "my-apache", false, codes.Internal),
		// See https://github.com/kubeapps/kubeapps/issues/632
		newTestCase(6, "my-apache", true, codes.OK),
		newTestCase(7, "my-apache", true, codes.OK),
		newTestCase(8, "my-apache", true, codes.OK),
		// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
		newTestCase(9, "my-apache", true, codes.OK),
		// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
		newTestCase(10, "my-apache", true, codes.OK),
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.GetInstalledPackageResourceRefsResponse{},
		corev1.ResourceRef{},
		corev1.Context{},
	)

	toHelmReleaseStubs := func(in []resourcerefstest.TestReleaseStub) []releaseStub {
		out := []releaseStub{}
		for _, r := range in {
			out = append(out, releaseStub{name: r.Name, namespace: r.Namespace, manifest: r.Manifest})
		}
		return out
	}

	for _, tc := range testCases {
		t.Run(tc.baseTestCase.Name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(
				t,
				tc.request.GetInstalledPackageRef().GetContext().GetNamespace(),
				toHelmReleaseStubs(tc.baseTestCase.ExistingReleases),
				nil)

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
