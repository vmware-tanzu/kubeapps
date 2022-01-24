// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/cmd/kubeops/server"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf server.ServeOptions
	}{
		{
			"all arguments are captured",
			[]string{

				"--assetsvc-url", "foo01",
				"--helm-driver", "foo02",
				"--list-max", "901",
				"--user-agent-comment", "foo03",
				"--timeout", "902",
				"--clusters-config-path", "foo04",
				"--pinniped-proxy-url", "foo05",
				"--burst", "903",
				"--qps", "904",
				"--namespace-header-name", "foo06",
				"--namespace-header-pattern", "foo07",
				"--global-repos-namespace", "kubeapps-global",
			},
			server.ServeOptions{
				AssetsvcURL:            "foo01",
				HelmDriverArg:          "foo02",
				ListLimit:              901,
				UserAgentComment:       "foo03",
				Timeout:                902,
				ClustersConfigPath:     "foo04",
				PinnipedProxyURL:       "foo05",
				Burst:                  903,
				Qps:                    904,
				NamespaceHeaderName:    "foo06",
				NamespaceHeaderPattern: "foo07",
				UserAgent:              "kubeops/devel (foo03)",
				GlobalReposNamespace:   "kubeapps-global",
			},
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
			cmd.Execute()
			serveOpts.UserAgent = getUserAgent(version, serveOpts.UserAgentComment)
			if got, want := serveOpts, tt.conf; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetUserAgent(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		comment  string
		expected string
	}{
		{
			name:     "creates a user agent without a comment",
			version:  "2.1.6",
			expected: "kubeops/2.1.6",
		},
		{
			name:     "creates a user agent with comment",
			version:  "2.1.6",
			comment:  "foobar",
			expected: "kubeops/2.1.6 (foobar)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := getUserAgent(tc.version, tc.comment), tc.expected; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}

}
