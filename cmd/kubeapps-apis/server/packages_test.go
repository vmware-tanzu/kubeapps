/*
Copyright © 2021 VMware
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
package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	globalPackagingNamespace = "kubeapps"
)

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageSummariesRequest
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				makeTestPackagingPlugin("mock1"),
				makeTestPackagingPlugin("mock2"),
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					makeAvailablePackageSummary("pkg-1", makeTestPackagingPlugin("mock1").plugin),
					makeAvailablePackageSummary("pkg-2", makeTestPackagingPlugin("mock1").plugin),
					makeAvailablePackageSummary("pkg-1", makeTestPackagingPlugin("mock2").plugin),
					makeAvailablePackageSummary("pkg-2", makeTestPackagingPlugin("mock2").plugin),
				},
				Categories: []string{"cat1", "cat1"},
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.InstalledPackageStatus{}, corev1.VersionReference{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.Maintainer{}, corev1.GetAvailablePackageDetailResponse{}, corev1.AvailablePackageReference{}, corev1.AvailablePackageSummary{}, corev1.InstalledPackageDetail{}, corev1.GetInstalledPackageDetailResponse{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.AvailablePackageDetail{}, corev1.GetAvailablePackageDetailResponse{}, corev1.GetAvailablePackageSummariesResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
				if got, want := availablePackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageDetailRequest
		expectedResponse  *corev1.GetAvailablePackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageDetail operation",
			configuredPlugins: []*pkgsPluginWithServer{
				makeTestPackagingPlugin("mock1"),
				makeTestPackagingPlugin("mock2"),
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin: &plugins.Plugin{
						Name:    "mock1",
						Version: "v1alpha1",
					},
				},
				PkgVersion: "",
			},

			expectedResponse: &corev1.GetAvailablePackageDetailResponse{
				AvailablePackageDetail: makeAvailablePackageDetail("pkg-1", makeTestPackagingPlugin("mock1").plugin),
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			availablePackageDetail, err := server.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.InstalledPackageStatus{}, corev1.VersionReference{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.Maintainer{}, corev1.GetAvailablePackageDetailResponse{}, corev1.AvailablePackageReference{}, corev1.AvailablePackageSummary{}, corev1.InstalledPackageDetail{}, corev1.GetInstalledPackageDetailResponse{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.AvailablePackageDetail{}, corev1.GetAvailablePackageDetailResponse{}, corev1.GetAvailablePackageSummariesResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
				if got, want := availablePackageDetail, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageSummariesRequest
		expectedResponse  *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				makeTestPackagingPlugin("mock1"),
				makeTestPackagingPlugin("mock2"),
			},
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					makeInstalledPackageSummary("pkg-1", makeTestPackagingPlugin("mock1").plugin),
					makeInstalledPackageSummary("pkg-2", makeTestPackagingPlugin("mock1").plugin),
					makeInstalledPackageSummary("pkg-1", makeTestPackagingPlugin("mock2").plugin),
					makeInstalledPackageSummary("pkg-2", makeTestPackagingPlugin("mock2").plugin),
				},
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			installedPackageSummaries, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.InstalledPackageStatus{}, corev1.VersionReference{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.Maintainer{}, corev1.GetAvailablePackageDetailResponse{}, corev1.AvailablePackageReference{}, corev1.AvailablePackageSummary{}, corev1.InstalledPackageDetail{}, corev1.GetInstalledPackageDetailResponse{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.AvailablePackageDetail{}, corev1.GetAvailablePackageDetailResponse{}, corev1.GetAvailablePackageSummariesResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
				if got, want := installedPackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageDetailRequest
		expectedResponse  *corev1.GetInstalledPackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageDetail operation",
			configuredPlugins: []*pkgsPluginWithServer{
				makeTestPackagingPlugin("mock1"),
				makeTestPackagingPlugin("mock2"),
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin: &plugins.Plugin{
						Name:    "mock1",
						Version: "v1alpha1",
					},
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: makeInstalledPackageDetail("pkg-1", makeTestPackagingPlugin("mock1").plugin),
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			installedPackageDetail, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.InstalledPackageStatus{}, corev1.VersionReference{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.Maintainer{}, corev1.GetAvailablePackageDetailResponse{}, corev1.AvailablePackageReference{}, corev1.AvailablePackageSummary{}, corev1.InstalledPackageDetail{}, corev1.GetInstalledPackageDetailResponse{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.AvailablePackageDetail{}, corev1.GetAvailablePackageDetailResponse{}, corev1.GetAvailablePackageSummariesResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
				if got, want := installedPackageDetail, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageVersionsRequest
		expectedResponse  *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageVersions operation",
			configuredPlugins: []*pkgsPluginWithServer{
				makeTestPackagingPlugin("mock1"),
				makeTestPackagingPlugin("mock2"),
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin: &plugins.Plugin{
						Name:    "mock1",
						Version: "v1alpha1",
					},
				},
			},

			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{
					{
						PkgVersion: "3.0.0/mock1",
						AppVersion: defaultAppVersion,
					},
					{
						PkgVersion: "2.0.0/mock1",
						AppVersion: defaultAppVersion,
					},
					{
						PkgVersion: "1.0.0/mock1",
						AppVersion: defaultAppVersion,
					},
				},
			},
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			AvailablePackageVersions, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.InstalledPackageStatus{}, corev1.VersionReference{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.Maintainer{}, corev1.GetAvailablePackageDetailResponse{}, corev1.AvailablePackageReference{}, corev1.AvailablePackageSummary{}, corev1.InstalledPackageDetail{}, corev1.GetInstalledPackageDetailResponse{}, corev1.GetInstalledPackageSummariesResponse{}, corev1.AvailablePackageDetail{}, corev1.GetAvailablePackageDetailResponse{}, corev1.GetAvailablePackageSummariesResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
				if got, want := AvailablePackageVersions, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

// TODO: implement this test once we implement the core operation
// func TestCreateInstalledPackage(t *testing.T) {
// }
