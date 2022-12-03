// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"testing"

	apimanifests "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"sigs.k8s.io/yaml"

	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
				manifests = append(manifests, *m)
			}

			s, err := newServerWithPackageManifests(t, manifests)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), tc.request)
			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %v, want: %v", got, want)
			}
			// If an error code was expected, then no need to continue checking
			// the response.
			if tc.expectedErrorCode != codes.OK {
				return
			}

			t.Logf("response ===> %s", PrettyPrint(response))

			// TODO
			// compareAvailablePackageSummaries(t, response, tc.expectedResponse)
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

/*
func compareAvailablePackageSummaries(t *testing.T, actual *corev1.GetAvailablePackageSummariesResponse, expected *corev1.GetAvailablePackageSummariesResponse) {
	// these are helpers to compare slices ignoring order
	lessAvailablePackageFunc := func(p1, p2 *corev1.AvailablePackageSummary) bool {
		return p1.DisplayName < p2.DisplayName
	}
	opt1 := cmpopts.IgnoreUnexported(
		corev1.GetAvailablePackageSummariesResponse{},
		corev1.AvailablePackageSummary{},
		corev1.AvailablePackageReference{},
		corev1.Context{},
		plugins.Plugin{},
		corev1.PackageAppVersion{})
	opt2 := cmpopts.SortSlices(lessAvailablePackageFunc)

	if !cmp.Equal(actual, expected, opt1, opt2) {
		t.Fatalf("mismatch (-want +got):\n%s", cmp.Diff(expected, actual, opt1, opt2))
	}
}
*/
