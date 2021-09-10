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

var mockedPackagingPlugin1 = makeDefaultTestPackagingPlugin("mock1")
var mockedPackagingPlugin2 = makeDefaultTestPackagingPlugin("mock2")
var mockedFailingPackagingPlugin = makeFailingTestPackagingPlugin("bad-plugin")

var ignoreUnexportedOpts = cmpopts.IgnoreUnexported(
	corev1.AvailablePackageDetail{},
	corev1.AvailablePackageReference{},
	corev1.AvailablePackageSummary{},
	corev1.Context{},
	corev1.GetAvailablePackageDetailResponse{},
	corev1.GetAvailablePackageSummariesResponse{},
	corev1.GetAvailablePackageVersionsResponse{},
	corev1.GetInstalledPackageDetailResponse{},
	corev1.GetInstalledPackageSummariesResponse{},
	corev1.InstalledPackageDetail{},
	corev1.InstalledPackageReference{},
	corev1.InstalledPackageStatus{},
	corev1.InstalledPackageSummary{},
	corev1.Maintainer{},
	corev1.PackageAppVersion{},
	corev1.VersionReference{},
	plugins.Plugin{},
)

func makeDefaultTestPackagingPlugin(pluginName string) *PkgsPluginWithServer {
	plugin := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	availablePackageSummaries := []*corev1.AvailablePackageSummary{
		MakeAvailablePackageSummary("pkg-2", plugin),
		MakeAvailablePackageSummary("pkg-1", plugin),
	}
	availablePackageDetail := MakeAvailablePackageDetail("pkg-1", plugin)
	installedPackageSummaries := []*corev1.InstalledPackageSummary{
		MakeInstalledPackageSummary("pkg-2", plugin),
		MakeInstalledPackageSummary("pkg-1", plugin),
	}
	installedPackageDetail := MakeInstalledPackageDetail("pkg-1", plugin)
	packageAppVersions := []*corev1.PackageAppVersion{
		MakePackageAppVersion(defaultAppVersion, defaultPkgUpdateVersion),
		MakePackageAppVersion(defaultAppVersion, defaultPkgVersion),
	}
	nextPageToken := "1"
	categories := []string{defaultCategory}
	return MakeTestPackagingPlugin(plugin, availablePackageSummaries, availablePackageDetail, installedPackageSummaries, installedPackageDetail, packageAppVersions, nextPageToken, categories, codes.OK)
}

func makeFailingTestPackagingPlugin(pluginName string) *PkgsPluginWithServer {
	plugin := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	return MakeTestPackagingPlugin(plugin, nil, nil, nil, nil, nil, "", nil, codes.NotFound)
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*PkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageSummariesRequest
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories: []string{"cat-1"},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (first page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "0", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (proper PageSize) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "0", PageSize: 4},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (last page+1) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "5", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				Categories:                []string{"cat-1"},
				NextPageToken:             "",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (last page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "4", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "5",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageSummaries operation with one of the plugin failing",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedFailingPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				Categories:                []string{""},
			},
			statusCode: codes.NotFound,
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
				opt1 := ignoreUnexportedOpts
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
		configuredPlugins []*PkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageDetailRequest
		expectedResponse  *corev1.GetAvailablePackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageDetail operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &corev1.GetAvailablePackageDetailResponse{
				AvailablePackageDetail: MakeAvailablePackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageDetail operation with one of the plugin failing",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedFailingPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedFailingPackagingPlugin.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &corev1.GetAvailablePackageDetailResponse{},
			statusCode:       codes.NotFound,
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
				opt1 := ignoreUnexportedOpts
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
		configuredPlugins []*PkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageSummariesRequest
		expectedResponse  *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageSummaries operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageSummaries operation with one of the plugin failing",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedFailingPackagingPlugin,
			},
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{},
			},
			statusCode: codes.NotFound,
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
				opt1 := ignoreUnexportedOpts
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
		configuredPlugins []*PkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageDetailRequest
		expectedResponse  *corev1.GetInstalledPackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageDetail operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: MakeInstalledPackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageDetail operation with one of the plugin failing",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedFailingPackagingPlugin,
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedFailingPackagingPlugin.plugin,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{},
			statusCode:       codes.NotFound,
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
				opt1 := ignoreUnexportedOpts
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
		configuredPlugins []*PkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageVersionsRequest
		expectedResponse  *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageVersions operation",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{
					MakePackageAppVersion(defaultAppVersion, defaultPkgUpdateVersion),
					MakePackageAppVersion(defaultAppVersion, defaultPkgVersion),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageSummaGetAvailablePackageVersionsries operation with one of the plugin failing",
			configuredPlugins: []*PkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedFailingPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedFailingPackagingPlugin.plugin,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{},
			},
			statusCode: codes.NotFound,
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
				opt1 := ignoreUnexportedOpts
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
