// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	plugintest "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

const (
	globalPackagingNamespace = "kubeapps"
)

var mockedPackagingPlugin1 = makeDefaultTestPackagingPlugin("mock1")
var mockedPackagingPlugin2 = makeDefaultTestPackagingPlugin("mock2")
var mockedNotFoundPackagingPlugin = makeOnlyStatusTestPackagingPlugin("bad-plugin", grpccodes.NotFound)

var ignoreUnexportedOpts = cmpopts.IgnoreUnexported(
	pkgsGRPCv1alpha1.AvailablePackageDetail{},
	pkgsGRPCv1alpha1.AvailablePackageReference{},
	pkgsGRPCv1alpha1.AvailablePackageSummary{},
	pkgsGRPCv1alpha1.Context{},
	pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{},
	pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{},
	pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{},
	pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{},
	pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{},
	pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{},
	pkgsGRPCv1alpha1.CreateInstalledPackageResponse{},
	pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{},
	pkgsGRPCv1alpha1.InstalledPackageDetail{},
	pkgsGRPCv1alpha1.InstalledPackageReference{},
	pkgsGRPCv1alpha1.InstalledPackageStatus{},
	pkgsGRPCv1alpha1.InstalledPackageSummary{},
	pkgsGRPCv1alpha1.Maintainer{},
	pkgsGRPCv1alpha1.PackageAppVersion{},
	pkgsGRPCv1alpha1.VersionReference{},
	pkgsGRPCv1alpha1.ResourceRef{},
	pluginsGRPCv1alpha1.Plugin{},
)

func makeDefaultTestPackagingPlugin(pluginName string) pkgPluginsWithServer {
	pluginDetails := &pluginsGRPCv1alpha1.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugintest.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.AvailablePackageSummaries = []*pkgsGRPCv1alpha1.AvailablePackageSummary{
		plugintest.MakeAvailablePackageSummary("pkg-2", pluginDetails),
		plugintest.MakeAvailablePackageSummary("pkg-1", pluginDetails),
	}
	packagingPluginServer.AvailablePackageDetail = plugintest.MakeAvailablePackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.InstalledPackageSummaries = []*pkgsGRPCv1alpha1.InstalledPackageSummary{
		plugintest.MakeInstalledPackageSummary("pkg-2", pluginDetails),
		plugintest.MakeInstalledPackageSummary("pkg-1", pluginDetails),
	}
	packagingPluginServer.InstalledPackageDetail = plugintest.MakeInstalledPackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.PackageAppVersions = []*pkgsGRPCv1alpha1.PackageAppVersion{
		plugintest.MakePackageAppVersion(plugintest.DefaultAppVersion, plugintest.DefaultPkgUpdateVersion),
		plugintest.MakePackageAppVersion(plugintest.DefaultAppVersion, plugintest.DefaultPkgVersion),
	}
	packagingPluginServer.NextPageToken = "1"
	packagingPluginServer.Categories = []string{plugintest.DefaultCategory}

	return pkgPluginsWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func makeOnlyStatusTestPackagingPlugin(pluginName string, statusCode grpccodes.Code) pkgPluginsWithServer {
	pluginDetails := &pluginsGRPCv1alpha1.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugintest.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.Status = statusCode

	return pkgPluginsWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []pkgPluginsWithServer
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest
		expectedResponse  *pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					plugintest.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories: []string{"cat-1"},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should successfully call and paginate (first page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{PageToken: "0", PageSize: 1},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					plugintest.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should successfully call and paginate (proper PageSize) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{PageToken: "0", PageSize: 4},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					plugintest.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should successfully call and paginate (last page - 1) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{PageToken: "3", PageSize: 1},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "4",
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should successfully call and paginate (last page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{PageToken: "3", PageSize: 1},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					plugintest.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "4",
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should successfully call and paginate (last page + 1) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{PageToken: "4", PageSize: 1},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{},
				Categories:                []string{"cat-1"},
				NextPageToken:             "",
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageSummaries operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{},
				Categories:                []string{""},
			},
			statusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []pkgPluginsWithServer
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest
		expectedResponse  *pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageDetail operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{
				AvailablePackageDetail: plugintest.MakeAvailablePackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{},
			statusCode:       grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			availablePackageDetail, err := server.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []pkgPluginsWithServer
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest
		expectedResponse  *pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageSummaries operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					plugintest.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugintest.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugintest.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugintest.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageSummaries operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{},
			},
			statusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			installedPackageSummaries, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []pkgPluginsWithServer
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest
		expectedResponse  *pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageDetail operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: plugintest.MakeInstalledPackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{},
			statusCode:       grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			installedPackageDetail, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []pkgPluginsWithServer
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest
		expectedResponse  *pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageVersions operation",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*pkgsGRPCv1alpha1.PackageAppVersion{
					plugintest.MakePackageAppVersion(plugintest.DefaultAppVersion, plugintest.DefaultPkgUpdateVersion),
					plugintest.MakePackageAppVersion(plugintest.DefaultAppVersion, plugintest.DefaultPkgVersion),
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageVersions operation when the package is not present in a plugin",
			configuredPlugins: []pkgPluginsWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},

			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*pkgsGRPCv1alpha1.PackageAppVersion{},
			},
			statusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			AvailablePackageVersions, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []*pluginsGRPCv1alpha1.Plugin
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.CreateInstalledPackageRequest
		expectedResponse  *pkgsGRPCv1alpha1.CreateInstalledPackageResponse
	}{
		{
			name: "installs the package using the correct plugin",
			configuredPlugins: []*pluginsGRPCv1alpha1.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: grpccodes.OK,
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
			expectedResponse: &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: grpccodes.InvalidArgument,
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Identifier: "available-pkg-1",
				},
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: grpccodes.Internal,
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginsWithServer{
					plugin: p,
					server: plugintest.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			installedPkgResponse, err := server.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []*pluginsGRPCv1alpha1.Plugin
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest
		expectedResponse  *pkgsGRPCv1alpha1.UpdateInstalledPackageResponse
	}{
		{
			name: "updates the package using the correct plugin",
			configuredPlugins: []*pluginsGRPCv1alpha1.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: grpccodes.OK,
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
			expectedResponse: &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: grpccodes.InvalidArgument,
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: grpccodes.Internal,
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginsWithServer{
					plugin: p,
					server: plugintest.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			updatedPkgResponse, err := server.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		configuredPlugins []*pluginsGRPCv1alpha1.Plugin
		statusCode        grpccodes.Code
		request           *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest
	}{
		{
			name: "deletes the package",
			configuredPlugins: []*pluginsGRPCv1alpha1.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: grpccodes.OK,
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: grpccodes.InvalidArgument,
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: grpccodes.Internal,
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []pkgPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, pkgPluginsWithServer{
					plugin: p,
					server: plugintest.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				pluginsWithServers: configuredPluginServers,
			}

			_, err := server.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
		})
	}
}

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	installedPlugin := &pluginsGRPCv1alpha1.Plugin{Name: "plugin-1", Version: "v1alpha1"}

	testCases := []struct {
		name               string
		statusCode         grpccodes.Code
		pluginResourceRefs []*pkgsGRPCv1alpha1.ResourceRef
		request            *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse
	}{
		{
			name: "it should successfully call the plugins GetInstalledPackageResourceRefs endpoint",
			pluginResourceRefs: []*pkgsGRPCv1alpha1.ResourceRef{
				{
					ApiVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "some-deployment",
				},
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     installedPlugin,
				},
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
				Context: &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
				ResourceRefs: []*pkgsGRPCv1alpha1.ResourceRef{
					{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it should return an invalid argument if the plugin is not specified",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
				},
			},
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name: "it should return an invalid argument if the plugin cannot be found",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "other-plugin.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				pluginsWithServers: []pkgPluginsWithServer{
					{
						plugin: installedPlugin,
						server: &plugintest.TestPackagingPluginServer{
							Plugin:       installedPlugin,
							ResourceRefs: tc.pluginResourceRefs,
						},
					},
				},
			}

			resourceRefs, err := server.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				if got, want := resourceRefs, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}
