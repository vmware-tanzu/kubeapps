// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugin_test"
)

var mockedRepoPlugin1 = makeDefaultTestRepositoriesPlugin("mock1")
var mockedRepoPlugin2 = makeDefaultTestRepositoriesPlugin("mock2")
var mockedNotFoundRepoPlugin = makeOnlyStatusTestRepositoriesPlugin("bad-plugin", connect.CodeNotFound)

var ignoreUnexportedRepoOpts = cmpopts.IgnoreUnexported(
	corev1.AddPackageRepositoryResponse{},
	corev1.Context{},
	plugins.Plugin{},
	corev1.PackageRepositoryReference{},
	corev1.GetPackageRepositoryDetailResponse{},
	corev1.PackageRepositoryDetail{},
	corev1.PackageRepositoryStatus{},
	corev1.GetPackageRepositorySummariesResponse{},
	corev1.PackageRepositorySummary{},
	corev1.UpdatePackageRepositoryResponse{},
	corev1.GetPackageRepositoryPermissionsResponse{},
	corev1.PackageRepositoriesPermissions{},
)

func makeDefaultTestRepositoriesPlugin(pluginName string) repoPluginsWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	repositoriesPluginServer := &plugin_test.TestRepositoriesPluginServer{Plugin: pluginDetails}

	repositoriesPluginServer.PackageRepositoryDetail =
		plugin_test.MakePackageRepositoryDetail("repo-1", pluginDetails)

	repositoriesPluginServer.PackageRepositorySummaries = []*corev1.PackageRepositorySummary{
		plugin_test.MakePackageRepositorySummary("repo-2", pluginDetails),
		plugin_test.MakePackageRepositorySummary("repo-1", pluginDetails),
	}

	return repoPluginsWithServer{
		plugin: pluginDetails,
		server: repositoriesPluginServer,
	}
}

func makeOnlyStatusTestRepositoriesPlugin(pluginName string, errorCode connect.Code) repoPluginsWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	repositoriesPluginServer := &plugin_test.TestRepositoriesPluginServer{Plugin: pluginDetails}

	repositoriesPluginServer.ErrorCode = errorCode

	return repoPluginsWithServer{
		plugin: pluginDetails,
		server: repositoriesPluginServer,
	}
}

func TestAddPackageRepository(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		errorCode         connect.Code
		request           *corev1.AddPackageRepositoryRequest
		expectedResponse  *corev1.AddPackageRepositoryResponse
	}{
		{
			name: "installs the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			request: &corev1.AddPackageRepositoryRequest{
				Plugin: &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "repo-1",
			},
			expectedResponse: &corev1.AddPackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Cluster:   "default",
						Namespace: "my-ns",
					},
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:      "returns invalid argument if plugin not specified in request",
			errorCode: connect.CodeInvalidArgument,
			request: &corev1.AddPackageRepositoryRequest{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "repo-1",
			},
		},
		{
			name:      "returns internal error if unable to find the plugin",
			errorCode: connect.CodeInternal,
			request: &corev1.AddPackageRepositoryRequest{
				Plugin: &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "repo-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []repoPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, repoPluginsWithServer{
					plugin: p,
					server: plugin_test.TestRepositoriesPluginServer{Plugin: p},
				})
			}

			server := &repositoriesServer{
				pluginsWithServers: configuredPluginServers,
			}

			addRepoResponse, err := server.AddPackageRepository(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.errorCode == 0 {
				if got, want := addRepoResponse.Msg, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}

func TestGetPackageRepositoryDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []repoPluginsWithServer
		errorCode         connect.Code
		request           *corev1.GetPackageRepositoryDetailRequest
		expectedResponse  *corev1.GetPackageRepositoryDetailResponse
	}{
		{
			name: "it should successfully call the core GetPackageRepositoryDetail operation",
			configuredPlugins: []repoPluginsWithServer{
				mockedRepoPlugin1,
				mockedRepoPlugin2,
			},
			request: &corev1.GetPackageRepositoryDetailRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Cluster:   plugin_test.GlobalPackagingCluster,
						Namespace: plugin_test.DefaultNamespace,
					},
					Identifier: "repo-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryDetailResponse{
				Detail: plugin_test.MakePackageRepositoryDetail("repo-1", mockedRepoPlugin1.plugin),
			},
		},
		{
			name: "it should fail when calling the core GetPackageRepositoryDetail operation when the package is not present in a plugin",
			configuredPlugins: []repoPluginsWithServer{
				mockedRepoPlugin1,
				mockedNotFoundRepoPlugin,
			},
			request: &corev1.GetPackageRepositoryDetailRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context: &corev1.Context{
						Cluster:   plugin_test.GlobalPackagingCluster,
						Namespace: plugin_test.DefaultNamespace,
					},
					Identifier: "repo-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},
			errorCode: connect.CodeNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &repositoriesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			packageRepoDetail, err := server.GetPackageRepositoryDetail(
				context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.errorCode == 0 {
				if got, want := packageRepoDetail.Msg, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}

func TestGetPackageRepositorySummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []repoPluginsWithServer
		errorCode         connect.Code
		request           *corev1.GetPackageRepositorySummariesRequest
		expectedResponse  *corev1.GetPackageRepositorySummariesResponse
	}{
		{
			name: "it should successfully call the core GetPackageRepositorySummaries operation",
			configuredPlugins: []repoPluginsWithServer{
				mockedRepoPlugin1,
				mockedRepoPlugin2,
			},
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{
					Cluster:   plugin_test.GlobalPackagingCluster,
					Namespace: plugin_test.DefaultNamespace,
				},
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					plugin_test.MakePackageRepositorySummary("repo-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakePackageRepositorySummary("repo-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakePackageRepositorySummary("repo-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakePackageRepositorySummary("repo-2", mockedPackagingPlugin2.plugin),
				},
			},
		},
		{
			name: "it should fail when calling the core GetPackageRepositorySummaries operation when the package is not present in a plugin",
			configuredPlugins: []repoPluginsWithServer{
				mockedRepoPlugin1,
				mockedNotFoundRepoPlugin,
			},
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{
					Cluster:   plugin_test.GlobalPackagingCluster,
					Namespace: plugin_test.DefaultNamespace,
				},
			},
			errorCode: connect.CodeNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &repositoriesServer{
				pluginsWithServers: tc.configuredPlugins,
			}
			repoSummaries, err := server.GetPackageRepositorySummaries(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.errorCode == 0 {
				if got, want := repoSummaries.Msg, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}

func TestUpdatePackageRepository(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		errorCode         connect.Code
		request           *corev1.UpdatePackageRepositoryRequest
		expectedResponse  *corev1.UpdatePackageRepositoryResponse
	}{
		{
			name: "updates the package repository using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			request: &corev1.UpdatePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
			expectedResponse: &corev1.UpdatePackageRepositoryResponse{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:      "returns invalid argument if plugin not specified in request",
			errorCode: connect.CodeInvalidArgument,
			request: &corev1.UpdatePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Identifier: "repo-1",
				},
			},
		},
		{
			name:      "returns internal error if unable to find the plugin",
			errorCode: connect.CodeInternal,
			request: &corev1.UpdatePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []repoPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, repoPluginsWithServer{
					plugin: p,
					server: plugin_test.TestRepositoriesPluginServer{Plugin: p},
				})
			}

			server := &repositoriesServer{
				pluginsWithServers: configuredPluginServers,
			}

			updatedRepoResponse, err := server.UpdatePackageRepository(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.errorCode == 0 {
				if got, want := updatedRepoResponse.Msg, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}

func TestDeletePackageRepository(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		errorCode         connect.Code
		request           *corev1.DeletePackageRepositoryRequest
	}{
		{
			name: "deletes the package repository",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:      "returns invalid argument if plugin not specified in request",
			errorCode: connect.CodeInvalidArgument,
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Identifier: "repo-1",
				},
			},
		},
		{
			name:      "returns internal error if unable to find the plugin",
			errorCode: connect.CodeInternal,
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Identifier: "repo-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []repoPluginsWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, repoPluginsWithServer{
					plugin: p,
					server: plugin_test.TestRepositoriesPluginServer{Plugin: p},
				})
			}

			server := &repositoriesServer{
				pluginsWithServers: configuredPluginServers,
			}

			_, err := server.DeletePackageRepository(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
		})
	}
}

func TestGetPackageRepositoryPermissions(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		errorCode         connect.Code
		request           *corev1.GetPackageRepositoryPermissionsRequest
		expectedResponse  *corev1.GetPackageRepositoryPermissionsResponse
	}{
		{
			name: "returns permissions for all plugins",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			request: &corev1.GetPackageRepositoryPermissionsRequest{},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
						Namespace: map[string]bool{
							"ns-verb": true,
						},
						Global: map[string]bool{
							"global-verb": true,
						},
					},
					{
						Plugin: &plugins.Plugin{Name: "plugin-1", Version: "v1alpha2"},
						Namespace: map[string]bool{
							"ns-verb": true,
						},
						Global: map[string]bool{
							"global-verb": true,
						},
					},
				},
			},
		},
		{
			name:             "returns empty set when no plugins",
			request:          &corev1.GetPackageRepositoryPermissionsRequest{},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var configuredPluginServers []repoPluginsWithServer
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, repoPluginsWithServer{
					plugin: p,
					server: plugin_test.TestRepositoriesPluginServer{Plugin: p},
				})
			}

			server := &repositoriesServer{
				pluginsWithServers: configuredPluginServers,
			}

			updatedRepoResponse, err := server.GetPackageRepositoryPermissions(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.errorCode == 0 {
				if got, want := updatedRepoResponse.Msg, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}
