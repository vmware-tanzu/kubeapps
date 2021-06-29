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
	"fmt"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
)

func TestPluginsAvailable(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		expectedPlugins   []*plugins.Plugin
	}{
		{
			name: "it returns the configured plugins verbatim",
			configuredPlugins: []*plugins.Plugin{
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "kapp_controller.packages",
					Version: "v1alpha1",
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
			ps := pluginsServer{
				plugins: tc.configuredPlugins,
			}

			resp, err := ps.GetConfiguredPlugins(context.TODO(), &plugins.GetConfiguredPluginsRequest{})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := resp.Plugins, tc.expectedPlugins; !cmp.Equal(want, got, cmp.Comparer(pluginEqual)) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmp.Comparer(pluginEqual)))
			}
		})
	}
}

func pluginEqual(a, b *plugins.Plugin) bool {
	return a.Name == b.Name && a.Version == b.Version
}

func TestSortPlugins(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		expectedPlugins   []*plugins.Plugin
	}{
		{
			name: "it sorts plugins by name",
			configuredPlugins: []*plugins.Plugin{
				{
					Name:    "kapp_controller.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha1",
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
		{
			name: "it sorts plugins by version (alpha-ordering) when names equal",
			configuredPlugins: []*plugins.Plugin{
				{
					Name:    "kapp_controller.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha2",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1beta1",
				},
			},
			expectedPlugins: []*plugins.Plugin{
				{
					Name:    "fluxv2.packages",
					Version: "v1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha1",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1alpha2",
				},
				{
					Name:    "fluxv2.packages",
					Version: "v1beta1",
				},
				{
					Name:    "kapp_controller.packages",
					Version: "v1alpha1",
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
			name:          "it returns no token and no error if the 'authorization' is empty",
			contextKey:    "",
			contextValue:  "",
			expectedToken: "",
			expectedErr:   nil,
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

// TODO(agamez): this test is just testing that the clients (typed and dynamic)
// are created, but nothing else.
// As per the PR #2908' comments, we could:
// use the http_test package to create a fake http server and use it's address as the endpoint you expect,
// then you could actually use the client to request something (anything),
// and verify that the token was sent with the request to the expected address
func TestCreateClientGetterWithParams(t *testing.T) {
	testCases := []struct {
		name           string
		contextKey     string
		contextValue   string
		shouldCreate   bool
		expectedErrMsg error
	}{
		{
			name:           "it creates the clients when passing a valid value for the authorization metadata",
			contextKey:     "authorization",
			contextValue:   "Bearer abc",
			shouldCreate:   true,
			expectedErrMsg: nil,
		},
		{
			name:           "it doesn't create the clients and throws a grpc error when passing an invalid authorization metadata",
			contextKey:     "authorization",
			contextValue:   "Bla",
			shouldCreate:   false,
			expectedErrMsg: status.Errorf(codes.Unauthenticated, "invalid authorization metadata: malformed authorization metadata"),
		},
		{
			name:           "it creates the clients when no authorization metadata is passed",
			contextKey:     "",
			contextValue:   "",
			shouldCreate:   true,
			expectedErrMsg: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
				tc.contextKey: tc.contextValue,
			}))

			inClusterConfig := &rest.Config{}
			serveOpts := ServeOptions{
				ClustersConfigPath: "/config.yaml",
				PinnipedProxyURL:   "http://example.com",
				UnsafeUseDemoSA:    true,
			}
			config := kube.ClustersConfig{
				KubeappsClusterName: "default",
				PinnipedProxyURL:    serveOpts.PinnipedProxyURL,
				Clusters: map[string]kube.ClusterConfig{
					"default": {
						Name: "default",
						PinnipedConfig: kube.PinnipedConciergeConfig{
							Enable: true,
						},
						IsKubeappsCluster: true,
					},
				},
			}
			clientGetter, err := createClientGetterWithParams(inClusterConfig, serveOpts, config)
			if err != nil {
				t.Fatalf("in %s: fail creating the clientGetter:  %+v", tc.name, err)
			}

			typedClient, dynamicClient, err := clientGetter(ctx)
			if tc.expectedErrMsg != nil && err != nil {
				if got, want := err.Error(), tc.expectedErrMsg.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
				}
			} else if err != nil {
				t.Fatalf("in %s: %+v", tc.name, err)
			}

			if tc.shouldCreate {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}
			}
		})
	}
}
