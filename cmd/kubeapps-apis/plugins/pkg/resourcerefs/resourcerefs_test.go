// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resourcerefs

import (
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	resourcerefstest "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	ignoredFields := cmpopts.IgnoreUnexported(
		pkgsGRPCv1alpha1.ResourceRef{},
	)

	testCases := []resourcerefstest.TestCase{}
	testCases = append(testCases, resourcerefstest.TestCases1...)
	testCases = append(testCases, resourcerefstest.TestCases2...)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			if len(tc.ExistingReleases) == 0 {
				t.Skip()
			}
			resourceRefs, err := ResourceRefsFromManifest(
				tc.ExistingReleases[0].Manifest,
				tc.ExistingReleases[0].Namespace)
			expectedStatusCode := tc.ExpectedErrStatusCode
			if expectedStatusCode == 0 {
				expectedStatusCode = grpccodes.OK
			}
			if got, want := grpcstatus.Code(err), expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := resourceRefs, tc.ExpectedResourceRefs; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
