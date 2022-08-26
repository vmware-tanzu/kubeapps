// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeops/server"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name        string
		args        []string
		conf        server.ServeOptions
		errExpected bool
	}{
		{
			"all arguments are captured",
			[]string{
				"--clusters-config-path", "foo04",
				"--pinniped-proxy-url", "foo05",
				"--pinniped-proxy-ca-cert", "/etc/foo/my-ca.crt",
				"--burst", "903",
				"--qps", "904",
				"--namespace-header-name", "foo06",
				"--namespace-header-pattern", "foo07",
			},
			server.ServeOptions{
				ClustersConfigPath:  "foo04",
				PinnipedProxyURL:    "foo05",
				PinnipedProxyCACert: "/etc/foo/my-ca.crt",
				Burst:               903,
				Qps:                 904,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
