/*
Copyright 2021 VMware. All Rights Reserved.

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

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/cmd/asset-syncer/server"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf server.Config
	}{
		{
			"all arguments are captured (root command)",
			[]string{
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
				DatabaseURL:           "foo01",
				DatabaseName:          "foo02",
				DatabaseUser:          "foo03",
				Debug:                 true,
				Namespace:             "foo04",
				GlobalReposNamespace:  "kubeapps-global",
				OciRepositories:       []string{"foo07"},
				TlsInsecureSkipVerify: true,
				FilterRules:           "foo06",
				PassCredentials:       true,
				UserAgent:             "asset-syncer/devel (foo05)",
			},
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
				DatabaseURL:           "foo01",
				DatabaseName:          "foo02",
				DatabaseUser:          "foo03",
				Debug:                 true,
				Namespace:             "foo04",
				GlobalReposNamespace:  "kubeapps-global",
				OciRepositories:       []string{"foo07"},
				TlsInsecureSkipVerify: true,
				FilterRules:           "foo06",
				PassCredentials:       true,
				UserAgent:             "asset-syncer/devel (foo05)",
			},
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
				"--oci-repositories", "foo07",
			},
			server.Config{
				DatabaseURL:           "foo01",
				DatabaseName:          "foo02",
				DatabaseUser:          "foo03",
				Debug:                 true,
				Namespace:             "foo04",
				GlobalReposNamespace:  "kubeapps-global",
				OciRepositories:       []string{"foo07"},
				TlsInsecureSkipVerify: true,
				FilterRules:           "foo06",
				PassCredentials:       true,
				UserAgent:             "asset-syncer/devel (foo05)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetErr(b)
			setRootFlags(cmd)
			setSyncFlags(cmd)
			cmd.SetArgs(tt.args)
			cmd.Execute()
			serveOpts.UserAgent = server.GetUserAgent(version, serveOpts.UserAgent)
			if got, want := serveOpts, tt.conf; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
