// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/asset-syncer/server"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name        string
		args        []string
		conf        server.Config
		errExpected bool
	}{
		{
			"all arguments are captured (invalidate command)",
			[]string{
				"invalidate-cache",
				"--database-url", "foo01",
				"--database-name", "foo02",
				"--database-user", "foo03",
				"--namespace", "foo04",
				"--global-repos-namespace", "kubeapps-global",
				"--user-agent-comment", "foo05",
				"--debug", "true",
				"--tls-insecure-skip-verify", "true",
				"--filter-rules", "foo06",
				"--pass-credentials", "true",
			},
			server.Config{
				DatabaseURL:              "foo01",
				DatabaseName:             "foo02",
				DatabaseUser:             "foo03",
				Debug:                    true,
				Namespace:                "foo04",
				GlobalPackagingNamespace: "kubeapps-global",
				OciRepositories:          []string{},
				TlsInsecureSkipVerify:    true,
				FilterRules:              "foo06",
				PassCredentials:          true,
				UserAgent:                "asset-syncer/devel (foo05)",
				UserAgentComment:         "foo05",
			},
			true,
		},
		{
			"all arguments are captured (sync command)",
			[]string{
				"sync",
				"--database-url", "foo01",
				"--database-name", "foo02",
				"--database-user", "foo03",
				"--namespace", "foo04",
				"--global-repos-namespace", "kubeapps-global",
				"--user-agent-comment", "foo05",
				"--debug", "true",
				"--tls-insecure-skip-verify", "true",
				"--filter-rules", "foo06",
				"--pass-credentials", "true",
				"--oci-repositories", "foo07",
			},
			server.Config{
				DatabaseURL:              "foo01",
				DatabaseName:             "foo02",
				DatabaseUser:             "foo03",
				Debug:                    true,
				Namespace:                "foo04",
				GlobalPackagingNamespace: "kubeapps-global",
				OciRepositories:          []string{"foo07"},
				TlsInsecureSkipVerify:    true,
				FilterRules:              "foo06",
				PassCredentials:          true,
				UserAgent:                "asset-syncer/devel (foo05)",
				UserAgentComment:         "foo05",
			},
			true,
		},
		{
			"all arguments are captured (delete command)",
			[]string{
				"delete",
				"--database-url", "foo01",
				"--database-name", "foo02",
				"--database-user", "foo03",
				"--namespace", "foo04",
				"--global-repos-namespace", "kubeapps-global",
				"--user-agent-comment", "foo05",
				"--debug", "true",
				"--tls-insecure-skip-verify", "true",
				"--filter-rules", "foo06",
				"--pass-credentials", "true",
			},
			server.Config{
				DatabaseURL:              "foo01",
				DatabaseName:             "foo02",
				DatabaseUser:             "foo03",
				Debug:                    true,
				Namespace:                "foo04",
				GlobalPackagingNamespace: "kubeapps-global",
				OciRepositories:          []string{},
				TlsInsecureSkipVerify:    true,
				FilterRules:              "foo06",
				PassCredentials:          true,
				UserAgent:                "asset-syncer/devel (foo05)",
				UserAgentComment:         "foo05",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetErr(b)
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
