// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name        string
		args        []string
		conf        core.ServeOptions
		errExpected bool
	}{
		{
			"all arguments are captured",
			[]string{
				"--config", "file",
				"--port", "901",
				"--plugin-dir", "foo01",
				"--clusters-config-path", "foo02",
				"--pinniped-proxy-url", "foo03",
				"--pinniped-proxy-ca-cert", "foo06",
				"--global-repos-namespace", "kubeapps-global",
				"--unsafe-local-dev-kubeconfig", "true",
				"--plugin-config-path", "foo05",
				"--kube-api-qps", "1.0",
				"--kube-api-burst", "1",
			},
			core.ServeOptions{
				Port:                     901,
				PluginDirs:               []string{"foo01"},
				ClustersConfigPath:       "foo02",
				PinnipedProxyURL:         "foo03",
				PinnipedProxyCACert:      "foo06",
				UnsafeLocalDevKubeconfig: true,
				GlobalHelmReposNamespace: "kubeapps-global",
				PluginConfigPath:         "foo05",
				QPS:                      1.0,
				Burst:                    1,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			cmd := newRootCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetErr(b)
			setFlags(cmd)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			if !tt.errExpected && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got, want := serveOpts, tt.conf; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
