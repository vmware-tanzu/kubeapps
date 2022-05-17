// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"sigs.k8s.io/yaml"
)

func TestRWMutexUtils(t *testing.T) {
	rw := &sync.RWMutex{}

	writeLocked := RWMutexWriteLocked(rw)
	readLocked := RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.RLock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || !readLocked {
		t.Fatalf("expected write/read lock: [false, true], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.RUnlock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.Lock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if !writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [true, false], got: [%t, %t]", writeLocked, readLocked)
	}

	rw.Unlock()
	writeLocked = RWMutexWriteLocked(rw)
	readLocked = RWMutexReadLocked(rw)
	if writeLocked || readLocked {
		t.Fatalf("expected write/read lock: [false, false], got: [%t, %t]", writeLocked, readLocked)
	}
}

func TestParsePluginConfig(t *testing.T) {
	testCases := []struct {
		name                    string
		pluginYAMLConf          []byte
		exp_versions_in_summary pkgutils.VersionsInSummary
		exp_error_str           string
	}{
		{
			name:                    "non existing plugin-config file",
			pluginYAMLConf:          nil,
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 0, Minor: 0, Patch: 0},
			exp_error_str:           "no such file or directory",
		},
		{
			name: "non-default plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 4, Minor: 2, Patch: 1},
			exp_error_str:           "",
		},
		{
			name: "partial params in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 1
        `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 1, Minor: 0, Patch: 0},
			exp_error_str:           "",
		},
		{
			name: "invalid plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1-IFC-123
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{},
			exp_error_str:           "json: cannot unmarshal",
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
			config, err := ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.exp_error_str)
			}
			if err == nil {
				if got, want := config.VersionsInSummary, tc.exp_versions_in_summary; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}

func TestParsePluginConfigTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		pluginYAMLConf []byte
		exp_timeout    int32
		exp_error_str  string
	}{
		{
			name:           "no timeout specified in plugin config",
			pluginYAMLConf: nil,
			exp_timeout:    0,
			exp_error_str:  "",
		},
		{
			name: "specific timeout in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      timeoutSeconds: 650
      `),
			exp_timeout:   650,
			exp_error_str: "",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			config, err := ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.exp_error_str)
			}
			if err == nil {
				if got, want := config.TimeoutSeconds, tc.exp_timeout; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}

func TestParsePluginConfigDefaultUpgradePolicy(t *testing.T) {
	testCases := []struct {
		name           string
		pluginYAMLConf []byte
		exp_policy_str string
		exp_error_str  string
	}{
		{
			name:           "no policy specified in plugin config",
			pluginYAMLConf: nil,
			exp_policy_str: "none",
			exp_error_str:  "",
		},
		{
			name: "specific policy in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      timeoutSeconds: 650
flux:
  packages:
    v1alpha1:
      defaultUpgradePolicy: minor
      `),
			exp_policy_str: "minor",
			exp_error_str:  "",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			config, err := ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.exp_error_str)
			}
			if err == nil {
				exp_policy, err := pkgutils.UpgradePolicyFromString(tc.exp_policy_str)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := config.DefaultUpgradePolicy, exp_policy; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}
