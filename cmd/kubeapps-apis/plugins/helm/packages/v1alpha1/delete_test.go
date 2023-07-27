// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteInstalledPackage(t *testing.T) {
	testCases := []struct {
		name              string
		existingReleases  []releaseStub
		request           *corev1.DeleteInstalledPackageRequest
		expectedErrorCode connect.Code
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
					status:         release.StatusDeployed,
				},
			},
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "my-apache",
				},
			},
		},
		{
			name: "returns invalid if installed package doesn't exist",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "default",
					},
					Identifier: "not-a-valid-identifier",
				},
			},
			expectedErrorCode: connect.CodeNotFound,
		},
	}

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

			_, err := server.DeleteInstalledPackage(context.Background(), connect.NewRequest(tc.request))

			if err == nil && tc.expectedErrorCode == 0 {
				return
			}
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

		})
	}
}
