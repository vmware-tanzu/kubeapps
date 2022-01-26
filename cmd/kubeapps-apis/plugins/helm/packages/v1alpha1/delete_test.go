// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmrelease "helm.sh/helm/v3/pkg/release"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteInstalledPackage(t *testing.T) {
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest
		expectedStatusCode grpccodes.Code
	}{
		{
			name: "deletes the installed package",
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
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
			expectedStatusCode: grpccodes.OK,
		},
		{
			name: "returns invalid if installed package doesn't exist",
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
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

			_, err := server.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

		})
	}
}
