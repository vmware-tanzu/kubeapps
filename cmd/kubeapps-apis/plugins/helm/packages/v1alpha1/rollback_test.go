// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	helmv1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRollbackInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *helmv1.RollbackInstalledPackageRequest
		expectedResponse   *helmv1.RollbackInstalledPackageResponse
		expectedStatusCode codes.Code
		expectedRelease    *release.Release
	}{
		{
			name: "rolls back the installed package from revision 2 to 1, creating rev 3",
			existingReleases: []releaseStub{
				{
					name:           "my-apache",
					namespace:      "default",
					chartID:        "bitnami/apache",
					chartVersion:   "1.18.3",
					chartNamespace: globalPackagingNamespace,
					status:         release.StatusDeployed,
					version:        2,
				},
				{
					name:           "my-apache",
					namespace:      "default",
					chartID:        "bitnami/apache",
					chartVersion:   "1.18.2",
					chartNamespace: globalPackagingNamespace,
					status:         release.StatusSuperseded,
					version:        1,
				},
			},
			request: &helmv1.RollbackInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				ReleaseRevision: 1,
			},
			expectedResponse: &helmv1.RollbackInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
					Plugin:     GetPluginDetail(),
				},
			},
			expectedStatusCode: codes.OK,
			expectedRelease: &release.Release{
				Name: "my-apache",
				Info: &release.Info{
					Description: "Rollback to 1",
					Status:      release.StatusDeployed,
				},
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Version:    "1.18.2",
						AppVersion: "1.2.6",
						Icon:       "https://example.com/icon.png",
					},
				},
				Config:    map[string]interface{}{},
				Version:   3,
				Namespace: "default",
			},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &helmv1.RollbackInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "default",
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedStatusCode: codes.NotFound,
		},
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		helmv1.RollbackInstalledPackageResponse{},
		corev1.InstalledPackageReference{},
		corev1.Context{},
		plugins.Plugin{},
		chart.Chart{},
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

			response, err := server.RollbackInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Verify the expected response (our contract to the caller).
			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}

			if tc.expectedRelease != nil {
				// Verify the expected request was made to Helm (our contract to the helm lib).
				deployedFilter := func(r *release.Release) bool {
					return r.Info.Status == release.StatusDeployed
				}
				releases, err := actionConfig.Releases.Driver.List(deployedFilter)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				if got, want := len(releases), 1; got != want {
					t.Fatalf("got: %d, want: %d", got, want)
				}

				ignoredFields := cmpopts.IgnoreFields(release.Info{}, "FirstDeployed", "LastDeployed")
				if got, want := releases[0], tc.expectedRelease; !cmp.Equal(got, want, ignoredUnexported, ignoredFields) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoredFields))
				}
			}
		})
	}
}
