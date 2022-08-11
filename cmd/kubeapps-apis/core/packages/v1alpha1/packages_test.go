// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugin_test"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	globalPackagingNamespace = "kubeapps"
)

var mockedPackagingPlugin1 = makeDefaultTestPackagingPlugin("mock1")
var mockedPackagingPlugin2 = makeDefaultTestPackagingPlugin("mock2")
var mockedPackagingPlugin3 = makeDefaultTestPackagingPlugin("mock2")
var mockedNotFoundPackagingPlugin = makeOnlyStatusTestPackagingPlugin("bad-plugin", codes.NotFound)

var ignoreUnexportedOpts = cmpopts.IgnoreUnexported(
	corev1.AvailablePackageDetail{},
	corev1.AvailablePackageReference{},
	corev1.AvailablePackageSummary{},
	corev1.Context{},
	corev1.GetAvailablePackageDetailResponse{},
	corev1.GetAvailablePackageSummariesResponse{},
	corev1.GetAvailablePackageVersionsResponse{},
	corev1.GetInstalledPackageResourceRefsResponse{},
	corev1.GetInstalledPackageDetailResponse{},
	corev1.GetInstalledPackageSummariesResponse{},
	corev1.CreateInstalledPackageResponse{},
	corev1.UpdateInstalledPackageResponse{},
	corev1.InstalledPackageDetail{},
	corev1.InstalledPackageReference{},
	corev1.InstalledPackageStatus{},
	corev1.InstalledPackageSummary{},
	corev1.Maintainer{},
	corev1.PackageAppVersion{},
	corev1.VersionReference{},
	corev1.ResourceRef{},
	plugins.Plugin{},
)

func makeDefaultTestPackagingPlugin(pluginName string) pkgPluginWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugin_test.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.AvailablePackageSummaries = []*corev1.AvailablePackageSummary{
		plugin_test.MakeAvailablePackageSummary("pkg-1", pluginDetails),
		plugin_test.MakeAvailablePackageSummary("pkg-2", pluginDetails),
	}
	packagingPluginServer.AvailablePackageDetail = plugin_test.MakeAvailablePackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.InstalledPackageSummaries = []*corev1.InstalledPackageSummary{
		plugin_test.MakeInstalledPackageSummary("pkg-1", pluginDetails),
		plugin_test.MakeInstalledPackageSummary("pkg-2", pluginDetails),
	}
	packagingPluginServer.InstalledPackageDetail = plugin_test.MakeInstalledPackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.PackageAppVersions = []*corev1.PackageAppVersion{
		plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgUpdateVersion),
		plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgVersion),
	}
	packagingPluginServer.NextPageToken = ""
	packagingPluginServer.Categories = []string{plugin_test.DefaultCategory}

	return pkgPluginWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func makeOnlyStatusTestPackagingPlugin(pluginName string, statusCode codes.Code) pkgPluginWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugin_test.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.Status = statusCode

	return pkgPluginWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageSummariesRequest
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginWithServer{
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
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories: []string{"cat-1"},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate one page the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageSize: 2},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: `{"mock1":1,"mock2":1}`,
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate with proper PageSize the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "", PageSize: 3},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: `{"mock1":2,"mock2":1}`,
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate last page of the core GetAvailablePackageSummaries operation exhausting the results",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: `{"mock1":2,"mock2":1}`, PageSize: 2},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate the last page of the core GetAvailablePackageSummaries operation without exhausting the results",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: `{"mock1":2,"mock2":1}`, PageSize: 1},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: `{"mock1":-1,"mock2":2}`,
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate beyond the last page of the core GetAvailablePackageSummaries operation when not exhausted",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: `{"mock1":-1,"mock2":2}`,
					PageSize:  1,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				Categories:                []string{},
				NextPageToken:             "",
			},
			statusCode: codes.OK,
		},
		{
			name: "it maintains the offset of a plugin even if that plugin did not contribute to the result",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
				mockedPackagingPlugin3,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: `{"mock1":1,"mock2":1,"mock3":1}`,
					PageSize:  2,
				},
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: `{"mock1":-1,"mock2":2,"mock3":1}`,
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageSummaries operation when the plugin returns a 404 for the api call",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
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
				pluginsWithServers: tc.configuredPlugins,
			}
			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := availablePackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageDetailRequest
		expectedResponse  *corev1.GetAvailablePackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageDetail operation",
			configuredPlugins: []pkgPluginWithServer{
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
				AvailablePackageDetail: plugin_test.MakeAvailablePackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
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
				pluginsWithServers: tc.configuredPlugins,
			}
			availablePackageDetail, err := server.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := availablePackageDetail, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageSummariesRequest
		expectedResponse  *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageSummaries operation",
			configuredPlugins: []pkgPluginWithServer{
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
					plugin_test.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageSummaries operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
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
				pluginsWithServers: tc.configuredPlugins,
			}
			installedPackageSummaries, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageDetailRequest
		expectedResponse  *corev1.GetInstalledPackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageDetail operation",
			configuredPlugins: []pkgPluginWithServer{
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
				InstalledPackageDetail: plugin_test.MakeInstalledPackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{},
			statusCode:       codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			installedPackageDetail, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPackageDetail, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageVersionsRequest
		expectedResponse  *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageVersions operation",
			configuredPlugins: []pkgPluginWithServer{
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
					plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgUpdateVersion),
					plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgVersion),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageVersions operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
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
				pluginsWithServers: tc.configuredPlugins,
			}
			AvailablePackageVersions, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := AvailablePackageVersions, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestCreateInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.CreateInstalledPackageRequest
		expectedResponse  *corev1.CreateInstalledPackageResponse
	}{
		{
			name: "installs the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			installedPkgResponse, err := server.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPkgResponse, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestUpdateInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.UpdateInstalledPackageRequest
		expectedResponse  *corev1.UpdateInstalledPackageResponse
	}{
		{
			name: "updates the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			updatedPkgResponse, err := server.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := updatedPkgResponse, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestDeleteInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.DeleteInstalledPackageRequest
	}{
		{
			name: "deletes the package",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			_, err := server.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
		})
	}
}

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	installedPlugin := &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"}

	testCases := []struct {
		name               string
		statusCode         codes.Code
		pluginResourceRefs []*corev1.ResourceRef
		request            *corev1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *corev1.GetInstalledPackageResourceRefsResponse
	}{
		{
			name: "it should successfully call the plugins GetInstalledPackageResourceRefs endpoint",
			pluginResourceRefs: []*corev1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "some-deployment",
				},
			},
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     installedPlugin,
				},
			},
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				Context: &corev1.Context{Cluster: "default", Namespace: "my-ns"},
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should return an invalid argument if the plugin is not specified",
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
				},
			},
			statusCode: codes.InvalidArgument,
		},
		{
			name: "it should return an invalid argument if the plugin cannot be found",
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "other-plugin.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: []pkgPluginWithServer{
					{
						plugin: installedPlugin,
						server: &plugin_test.TestPackagingPluginServer{
							Plugin:       installedPlugin,
							ResourceRefs: tc.pluginResourceRefs,
						},
					},
				},
			}

			resourceRefs, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := resourceRefs, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}
