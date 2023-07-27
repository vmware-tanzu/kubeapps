// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resourcerefs

import (
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
)

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	ignoredFields := cmpopts.IgnoreUnexported(
		corev1.ResourceRef{},
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
			expectedStatusCode := tc.ExpectedErrCode
			if got, want := connect.CodeOf(err), expectedStatusCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := resourceRefs, tc.ExpectedResourceRefs; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
