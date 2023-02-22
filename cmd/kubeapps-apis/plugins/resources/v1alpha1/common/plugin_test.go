// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"sigs.k8s.io/yaml"
)

func TestParsePluginConfig(t *testing.T) {
	testCases := []struct {
		name           string
		pluginYAMLConf []byte
		expectedConfig *ResourcesPluginConfig
		expectedError  string
	}{
		{
			name:           "non existing plugin-config file",
			pluginYAMLConf: nil,
			expectedConfig: nil,
			expectedError:  "",
		},
		{
			name: "invalid plugin config",
			pluginYAMLConf: []byte(`
resources:
  packages:
    v1alpha1:
      trustedNamespaces:
        headerName: true
      `),
			expectedConfig: nil,
			expectedError:  "json: cannot unmarshal",
		},
		{
			name: "non-default, valid plugin config",
			pluginYAMLConf: []byte(`
resources:
  packages:
    v1alpha1:
      trustedNamespaces:
        headerName: "X-Consumer-Groups"
        headerPattern: "^namespace:([\\w-]+)$"
      `),
			expectedConfig: &ResourcesPluginConfig{
				TrustedNamespaces: TrustedNamespaces{
					HeaderName:    "X-Consumer-Groups",
					HeaderPattern: "^namespace:([\\w-]+)$",
				},
			},
			expectedError: "",
		},
		{
			name: "parses forwarded headers config",
			pluginYAMLConf: []byte(`
resources:
  packages:
    v1alpha1:
      forwardedHeaders:
      - X-Consumer-Username
      - X-Consumer-Permissions
`),
			expectedConfig: &ResourcesPluginConfig{
				ForwardedHeaders: []string{
					"X-Consumer-Username",
					"X-Consumer-Permissions",
				},
			},
			expectedError: "",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			filename := ""
			if tc.pluginYAMLConf != nil {
				pluginJSONConf, err := yaml.YAMLToJSON(tc.pluginYAMLConf)
				if err != nil {
					t.Fatalf("%s", err)
				}
				f, err := os.CreateTemp(".", "plugin_json_conf")
				if err != nil {
					t.Fatalf("%s", err)
				}
				defer os.Remove(f.Name()) // clean up
				if _, err := f.Write(pluginJSONConf); err != nil {
					t.Fatalf("%s", err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("%s", err)
				}
				filename = f.Name()
			}
			pluginConfig, err := ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.expectedError)
			} else if tc.expectedConfig != nil {
				if got, want := pluginConfig, tc.expectedConfig; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}
