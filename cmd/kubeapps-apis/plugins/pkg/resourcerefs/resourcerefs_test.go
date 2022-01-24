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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/resourcerefs/resourcerefstest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			expectedStatusCode := tc.ExpectedErrStatusCode
			if expectedStatusCode == 0 {
				expectedStatusCode = codes.OK
			}
			if got, want := status.Code(err), expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if got, want := resourceRefs, tc.ExpectedResourceRefs; !cmp.Equal(want, got, ignoredFields) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredFields))
			}
		})
	}
}
