// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	log "k8s.io/klog/v2"
	"os"
	"runtime"
	"sigs.k8s.io/yaml"
	"strings"
	"testing"
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
			expectedConfig: &ResourcesPluginConfig{},
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
					log.Fatalf("%s", err)
				}
				f, err := os.CreateTemp(".", "plugin_json_conf")
				if err != nil {
					log.Fatalf("%s", err)
				}
				defer os.Remove(f.Name()) // clean up
				if _, err := f.Write(pluginJSONConf); err != nil {
					log.Fatalf("%s", err)
				}
				if err := f.Close(); err != nil {
					log.Fatalf("%s", err)
				}
				filename = f.Name()
			}
			pluginConfig, err := ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.expectedError) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.expectedError)
			} else if pluginConfig != nil {
				if got, want := pluginConfig, tc.expectedConfig; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}
