// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	resourcerefstest "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetInstalledPackageResourceRefs(t *testing.T) {

	// sanity check
	if len(resourcerefstest.TestCases1) < 11 {
		t.Fatalf("Expected array [resourcerefstest.TestCases1] size of at least 11")
	}

	type testCase struct {
		baseTestCase       resourcerefstest.TestCase
		request            *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse
		expectedStatusCode grpccodes.Code
	}

	// newTestCase is a function to take an existing test-case
	// (a so-called baseTestCase in pkg/resourcerefs module, which contains a LOT of useful data)
	// and "enrich" it with some new fields to create a different kind of test case
	// that tests server.GetInstalledPackageResourceRefs() func
	newTestCase := func(tc int, identifier string, response bool, code grpccodes.Code) testCase {
		newCase := testCase{
			baseTestCase: resourcerefstest.TestCases1[tc],
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: identifier,
				},
			},
		}
		if response {
			newCase.expectedResponse = &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				ResourceRefs: resourcerefstest.TestCases1[tc].ExpectedResourceRefs,
			}
		}
		newCase.expectedStatusCode = code
		return newCase
	}

	testCases := []testCase{
		newTestCase(0, "my-apache", true, grpccodes.OK),
		newTestCase(1, "my-apache", true, grpccodes.OK),
		newTestCase(2, "my-apache", true, grpccodes.OK),
		newTestCase(3, "my-apache", true, grpccodes.OK),
		newTestCase(4, "my-iis", false, grpccodes.NotFound),
		newTestCase(5, "my-apache", false, grpccodes.Internal),
		// See https://github.com/kubeapps/kubeapps/issues/632
		newTestCase(6, "my-apache", true, grpccodes.OK),
		newTestCase(7, "my-apache", true, grpccodes.OK),
		newTestCase(8, "my-apache", true, grpccodes.OK),
		// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/role-v1/#RoleList
		newTestCase(9, "my-apache", true, grpccodes.OK),
		// See https://kubernetes.io/docs/reference/kubernetes-api/authorization-resources/cluster-role-v1/#ClusterRoleList
		newTestCase(10, "my-apache", true, grpccodes.OK),
	}

	ignoredFields := cmpopts.IgnoreUnexported(
		pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{},
		pkgsGRPCv1alpha1.ResourceRef{},
		pkgsGRPCv1alpha1.Context{},
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

			server, _, cleanup := makeServer(t, authorized, actionConfig, &apprepov1alpha1.AppRepository{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()

			response, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
