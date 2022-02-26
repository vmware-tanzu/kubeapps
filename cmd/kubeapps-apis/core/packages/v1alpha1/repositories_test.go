// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ignoreUnexportedRepoOpts = cmpopts.IgnoreUnexported(
	corev1.AddPackageRepositoryRequest{},
	corev1.AddPackageRepositoryResponse{},
	corev1.Context{},
	plugins.Plugin{},
)

func TestAddPackageRepository(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.AddPackageRepositoryRequest
		expectedResponse  *corev1.AddPackageRepositoryResponse
	}{
		{
			name: "installs the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.AddPackageRepositoryRequest{
				Plugin: &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "repo-1",
			},
			expectedResponse: &corev1.AddPackageRepositoryResponse{},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.AddPackageRepositoryRequest{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "repo-1",
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
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

			addRepoResponse, err := server.AddPackageRepository(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := addRepoResponse, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedRepoOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedRepoOpts))
				}
			}
		})
	}
}
