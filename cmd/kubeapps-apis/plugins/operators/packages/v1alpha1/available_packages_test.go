// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"

	apimanifests "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"sigs.k8s.io/yaml"

	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
)

func TestGetAvailablePackageSummariesWithoutPagination(t *testing.T) {
	testCases := []struct {
		name              string
		request           *corev1.GetAvailablePackageSummariesRequest
		manifests         []string
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
		expectedErrorCode codes.Code
	}{
		{
			name:             "it returns a couple of operators packages from the cluster (no request ns specified)",
			manifests:        []string{"./testdata/etcd.yaml"},
			request:          &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}},
			expectedResponse: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifests := []apimanifests.PackageManifest{}
			for _, f := range tc.manifests {
				m, err := loadPackageManifest(f)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("manifest ===> %s", PrettyPrint(m))
				manifests = append(manifests, *m)
			}

			_, err := newServerWithPackageManifests(t, manifests)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}
		})
	}
}

func loadPackageManifest(file string) (*apimanifests.PackageManifest, error) {
	byteArray, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var m apimanifests.PackageManifest
	if err := yaml.Unmarshal(byteArray, &m); err != nil {
		return nil, err
	}

	return &m, err
}
