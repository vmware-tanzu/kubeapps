// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
)

var ignoreUnexported = cmpopts.IgnoreUnexported(
	PluginWithServer{},
	plugins.Plugin{},
)

func TestPluginsAvailable(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []PluginWithServer
		expectedPlugins   []*plugins.Plugin
	}{
		{
			name: "it returns the configured plugins verbatim",
			configuredPlugins: []PluginWithServer{
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "kapp_controller.packages",
						Version: "v1alpha1",
					},
				},
			},
			expectedPlugins: []*plugins.Plugin{
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "kapp_controller.packages",
					Version: "v1alpha1",
				},
			},
		},
		// We may later allow requesting just plugins for a specific service.
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ps := PluginsServer{
				pluginsWithServers: tc.configuredPlugins,
			}

			resp, err := ps.GetConfiguredPlugins(context.TODO(), &plugins.GetConfiguredPluginsRequest{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := resp.Plugins, tc.expectedPlugins; !cmp.Equal(want, got, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}
}

func pluginEqual(a, b PluginWithServer) bool {
	return a.Plugin.Name == b.Plugin.Name && a.Plugin.Version == b.Plugin.Version
}

func TestSortPlugins(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []PluginWithServer
		expectedPlugins   []PluginWithServer
	}{
		{
			name: "it sorts plugins by name",
			configuredPlugins: []PluginWithServer{
				{
					Plugin: &plugins.Plugin{
						Name:    "kapp_controller.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha1",
					},
				},
			},
			expectedPlugins: []PluginWithServer{
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "kapp_controller.packages",
						Version: "v1alpha1",
					},
				},
			},
		},
		{
			name: "it sorts plugins by version (alpha-ordering) when names equal",
			configuredPlugins: []PluginWithServer{
				{
					Plugin: &plugins.Plugin{
						Name:    "kapp_controller.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha2",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1beta1",
					},
				},
			},
			expectedPlugins: []PluginWithServer{
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1alpha2",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "fluxv2.packages",
						Version: "v1beta1",
					},
				},
				{
					Plugin: &plugins.Plugin{
						Name:    "kapp_controller.packages",
						Version: "v1alpha1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sortPlugins(tc.configuredPlugins)

			if got, want := tc.configuredPlugins, tc.expectedPlugins; !cmp.Equal(want, got, cmp.Comparer(pluginEqual)) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmp.Comparer(pluginEqual)))
			}
		})
	}
}

func TestListOSFiles(t *testing.T) {
	testCases := []struct {
		name            string
		filenames       []string
		pluginsDirs     []string
		pluginFilenames []string
	}{
		{
			name: "finds only so files in plugins directory",
			filenames: []string{
				"/tmp/plugins/foo.so",
				"/tmp/plugins/bar.so",
				"/tmp/plugins/not-an-so.txt",
			},
			pluginsDirs: []string{"/tmp/plugins"},
			pluginFilenames: []string{
				"/tmp/plugins/bar.so",
				"/tmp/plugins/foo.so",
			},
		},
		{
			name: "finds so files in multiple plugin directories",
			filenames: []string{
				"/tmp/plugins/foo.so",
				"/tmp/plugins/bar.so",
				"/tmp/plugins/not-an-so.txt",
				"/tmp/other/zap.so",
				"/tmp/other/not-an-so.woo",
			},
			pluginsDirs: []string{"/tmp/plugins", "/tmp/other"},
			pluginFilenames: []string{
				"/tmp/plugins/bar.so",
				"/tmp/plugins/foo.so",
				"/tmp/other/zap.so",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			fs := createTestFS(t, tc.filenames)

			got, err := listSOFiles(fs, tc.pluginsDirs)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := got, tc.pluginFilenames; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}

		})
	}
}

func createTestFS(t *testing.T, filenames []string) fstest.MapFS {
	fs := fstest.MapFS{
		"tmp":         {Mode: fs.ModeDir},
		"tmp/plugins": {Mode: fs.ModeDir},
		"tmp/other":   {Mode: fs.ModeDir},
	}

	for _, filename := range filenames {
		relFilename, err := filepath.Rel(pluginRootDir, filename)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		fs[relFilename] = &fstest.MapFile{
			Data: []byte("foo"),
			Mode: 0777,
		}
	}
	return fs
}

func TestExtractToken(t *testing.T) {
	testCases := []struct {
		name          string
		contextKey    string
		contextValue  string
		expectedToken string
		expectedErr   error
	}{
		{
			name:          "it returns the expected token without error for a valid 'authorization' metadata value",
			contextKey:    "authorization",
			contextValue:  "Bearer abc",
			expectedToken: "abc",
			expectedErr:   nil,
		},
		{
			name:          "it returns no token with an error if the 'authorization' metadata value is invalid",
			contextKey:    "authorization",
			contextValue:  "Bla",
			expectedToken: "",
			expectedErr:   fmt.Errorf("malformed authorization metadata"),
		},
		{
			name:          "it returns no token and expected error if the 'authorization' is empty",
			contextKey:    "",
			contextValue:  "",
			expectedToken: "",
			expectedErr:   fmt.Errorf("missing authorization metadata"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			context := context.Background()
			context = metadata.NewIncomingContext(context, metadata.New(map[string]string{
				tc.contextKey: tc.contextValue,
			}))

			token, err := extractToken(context)

			if tc.expectedErr != nil && err != nil {
				if got, want := err.Error(), tc.expectedErr.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
				}
			} else if err != nil {
				t.Fatalf("in %s: %+v", tc.name, err)
			}

			if got, want := token, tc.expectedToken; !cmp.Equal(want, got) {
				t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
			}
		})
	}
}

func TestCreateConfigGetterWithParams(t *testing.T) {
	const (
		DefaultClusterName = "default"
		DefaultK8sAPI      = "http://example.com/default/"
		OtherClusterName   = "other"
		OtherK8sAPI        = "http://example.com/other/"
	)
	inClusterConfig := &rest.Config{
		Host: DefaultK8sAPI,
	}
	clustersConfig := kube.ClustersConfig{
		KubeappsClusterName: "default",
		Clusters: map[string]kube.ClusterConfig{
			DefaultClusterName: {
				Name:              "default",
				IsKubeappsCluster: true,
			},
			OtherClusterName: {
				Name:          "other",
				APIServiceURL: OtherK8sAPI,
			},
		},
	}
	testCases := []struct {
		name            string
		cluster         string
		contextKey      string
		contextValue    string
		expectedAPIHost string
		expectedErrMsg  error
	}{
		{
			name:            "it creates the config for the default cluster when passing a valid value for the authorization metadata",
			contextKey:      "authorization",
			contextValue:    "Bearer abc",
			expectedAPIHost: DefaultK8sAPI,
			expectedErrMsg:  nil,
		},
		{
			name:           "it doesn't create the config and throws a grpc error when passing an invalid authorization metadata",
			contextKey:     "authorization",
			contextValue:   "Bla",
			expectedErrMsg: status.Errorf(codes.Unauthenticated, "invalid authorization metadata: malformed authorization metadata"),
		},
		{
			name:            "it doesn't create the config and throws a grpc error for the default cluster when no authorization metadata is passed",
			contextKey:      "",
			contextValue:    "",
			expectedAPIHost: DefaultK8sAPI,
			expectedErrMsg:  status.Errorf(codes.Unauthenticated, "invalid authorization metadata: missing authorization metadata"),
		},
		{
			name:            "it doesn't create the config and throws a grpc error for the other cluster",
			contextKey:      "",
			contextValue:    "",
			cluster:         OtherClusterName,
			expectedAPIHost: OtherK8sAPI,
			expectedErrMsg:  status.Errorf(codes.Unauthenticated, "invalid authorization metadata: missing authorization metadata"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				tc.contextKey: tc.contextValue,
			}))

			serveOpts := core.ServeOptions{
				ClustersConfigPath: "/config.yaml",
				PinnipedProxyURL:   "http://example.com",
			}
			configGetter, err := createConfigGetterWithParams(inClusterConfig, serveOpts, clustersConfig)
			if err != nil {
				t.Fatalf("in %s: fail creating the configGetter:  %+v", tc.name, err)
			}

			restConfig, err := configGetter(ctx, tc.cluster)
			if tc.expectedErrMsg != nil && err != nil {
				if got, want := err.Error(), tc.expectedErrMsg.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
				}
			} else if err != nil {
				t.Fatalf("in %s: %+v", tc.name, err)
			}

			if tc.expectedErrMsg == nil {
				if restConfig == nil {
					t.Errorf("got: nil, want: rest.Config")
				} else if got, want := restConfig.Host, tc.expectedAPIHost; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			}
		})
	}
}
