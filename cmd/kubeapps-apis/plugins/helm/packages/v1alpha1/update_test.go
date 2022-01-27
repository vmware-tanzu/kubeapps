// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrelease "helm.sh/helm/v3/pkg/release"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest
		expectedResponse   *pkgsGRPCv1alpha1.UpdateInstalledPackageResponse
		expectedStatusCode grpccodes.Code
		expectedRelease    *helmrelease.Release
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
					status:         helmrelease.StatusDeployed,
				},
			},
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.18.4",
				},
				Values: "{\"foo\": \"baz\"}",
			},
			expectedResponse: &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
					Plugin:     GetPluginDetail(),
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedRelease: &helmrelease.Release{
				Name: "my-apache",
				Info: &helmrelease.Info{
					Description: "Upgrade complete",
					Status:      helmrelease.StatusDeployed,
				},
				Chart: &helmchart.Chart{
					Metadata: &helmchart.Metadata{
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
					status:         helmrelease.StatusDeployed,
				},
			},
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.18.4",
				},
				Values: "{\"foo\": \"baz\"}",
			},
			expectedResponse: &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
					Plugin:     GetPluginDetail(),
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedRelease: &helmrelease.Release{
				Name: "my-apache",
				Info: &helmrelease.Info{
					Description: "Upgrade complete",
					Status:      helmrelease.StatusDeployed,
				},
				Chart: &helmchart.Chart{
					Metadata: &helmchart.Metadata{
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
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedStatusCode: grpccodes.NotFound,
		},
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{},
		pkgsGRPCv1alpha1.InstalledPackageReference{},
		pkgsGRPCv1alpha1.Context{},
		pluginsGRPCv1alpha1.Plugin{},
		helmchart.Chart{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)
			server, mockDB, cleanup := makeServer(t, authorized, actionConfig, &apprepov1alpha1.AppRepository{
				ObjectMeta: k8smetav1.ObjectMeta{
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

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Verify the expected response (our contract to the caller).
			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}

			if tc.expectedRelease != nil {
				// Verify the expected request was made to Helm (our contract to the helm lib).
				deployedFilter := func(r *helmrelease.Release) bool {
					return r.Info.Status == helmrelease.StatusDeployed
				}
				releases, err := actionConfig.Releases.Driver.List(deployedFilter)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				if got, want := len(releases), 1; got != want {
					t.Fatalf("got: %d, want: %d", got, want)
				}

				ignoredFields := cmpopts.IgnoreFields(helmrelease.Info{}, "FirstDeployed", "LastDeployed")
				if got, want := releases[0], tc.expectedRelease; !cmp.Equal(got, want, ignoredUnexported, ignoredFields) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoredFields))
				}
			}
		})
	}
}
