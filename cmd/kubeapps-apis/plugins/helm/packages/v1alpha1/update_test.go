// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *corev1.UpdateInstalledPackageRequest
		expectedResponse   *corev1.UpdateInstalledPackageResponse
		expectedStatusCode codes.Code
		expectedRelease    *release.Release
	}{
		{
			name: "updates the installed package from repo without credentials",
			existingReleases: []releaseStub{
				{
					name:           "my-apache",
					namespace:      "default",
					chartID:        "bitnami/apache",
					chartVersion:   "1.18.3",
					chartNamespace: globalPackagingNamespace,
					status:         release.StatusDeployed,
				},
			},
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.18.4",
				},
				Values: "{\"foo\": \"baz\"}",
			},
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
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
					Description: "Upgrade complete",
					Status:      release.StatusDeployed,
				},
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Name:    "apache",
						Version: "1.18.4",
					},
					Values: map[string]interface{}{},
				},
				Config:    map[string]interface{}{"foo": "baz"},
				Version:   1,
				Namespace: "default",
			},
		},
		{
			name: "populates the cluster in the returned context if not set in request",
			existingReleases: []releaseStub{
				{
					name:           "my-apache",
					namespace:      "default",
					chartID:        "bitnami/apache",
					chartVersion:   "1.18.3",
					chartNamespace: globalPackagingNamespace,
					status:         release.StatusDeployed,
				},
			},
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.18.4",
				},
				Values: "{\"foo\": \"baz\"}",
			},
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
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
					Description: "Upgrade complete",
					Status:      release.StatusDeployed,
				},
				Chart: &chart.Chart{
					Metadata: &chart.Metadata{
						Name:    "apache",
						Version: "1.18.4",
					},
					Values: map[string]interface{}{},
				},
				Config:    map[string]interface{}{"foo": "baz"},
				Version:   1,
				Namespace: "default",
			},
		},
		{
			name: "returns invalid if installed package doesn't exist",
			request: &corev1.UpdateInstalledPackageRequest{
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
		corev1.UpdateInstalledPackageResponse{},
		corev1.InstalledPackageReference{},
		corev1.Context{},
		plugins.Plugin{},
		chart.Chart{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)
			server, mockDB, cleanup := makeServer(t, authorized, actionConfig, &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()
			populateAssetDB(t, mockDB, tc.existingReleases)
			if tc.expectedRelease != nil {
				populateAssetForTarball(t, mockDB, fmt.Sprintf("bitnami%%%s", tc.expectedRelease.Chart.Metadata.Name), globalPackagingNamespace, tc.expectedRelease.Chart.Metadata.Version)
			}
			response, err := server.UpdateInstalledPackage(context.Background(), tc.request)

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
