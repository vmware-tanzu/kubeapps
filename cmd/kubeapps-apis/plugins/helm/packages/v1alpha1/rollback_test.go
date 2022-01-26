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
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	pkghelmv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmrelease "helm.sh/helm/v3/pkg/release"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRollbackInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *pkghelmv1alpha1.RollbackInstalledPackageRequest
		expectedResponse   *pkghelmv1alpha1.RollbackInstalledPackageResponse
		expectedStatusCode grpccodes.Code
		expectedRelease    *helmrelease.Release
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
					status:         helmrelease.StatusDeployed,
					version:        2,
				},
				{
					name:           "my-apache",
					namespace:      "default",
					chartID:        "bitnami/apache",
					chartVersion:   "1.18.2",
					chartNamespace: globalPackagingNamespace,
					status:         helmrelease.StatusSuperseded,
					version:        1,
				},
			},
			request: &pkghelmv1alpha1.RollbackInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
				ReleaseRevision: 1,
			},
			expectedResponse: &pkghelmv1alpha1.RollbackInstalledPackageResponse{
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
					Description: "Rollback to 1",
					Status:      helmrelease.StatusDeployed,
				},
				Chart: &helmchart.Chart{
					Metadata: &helmchart.Metadata{
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
			request: &pkghelmv1alpha1.RollbackInstalledPackageRequest{
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
		pkghelmv1alpha1.RollbackInstalledPackageResponse{},
		pkgsGRPCv1alpha1.InstalledPackageReference{},
		pkgsGRPCv1alpha1.Context{},
		pluginsGRPCv1alpha1.Plugin{},
		helmchart.Chart{},
	)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)

			server, _, cleanup := makeServer(t, authorized, actionConfig, &apprepov1alpha1.AppRepository{
				ObjectMeta: k8smetav1.ObjectMeta{
					Name:      "bitnami",
					Namespace: globalPackagingNamespace,
				},
			})
			defer cleanup()

			response, err := server.RollbackInstalledPackage(context.Background(), tc.request)

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
