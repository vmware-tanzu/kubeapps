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
	"strings"

	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/server"
	v1 "k8s.io/api/core/v1"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf server.Config
	}{
		{
			"no arguments returns default flag values",
			[]string{},
			server.Config{
				Kubeconfig:               "",
				MasterURL:                "",
				RepoSyncImage:            "docker.io/kubeapps/asset-syncer:latest",
				RepoSyncImagePullSecrets: nil,
				RepoSyncCommand:          "/chart-repo",
				KubeappsNamespace:        "kubeapps",
				ReposPerNamespace:        true,
				DBURL:                    "localhost",
				DBUser:                   "root",
				DBName:                   "charts",
				DBSecretName:             "kubeapps-db",
				DBSecretKey:              "postgresql-root-password",
				UserAgentComment:         "",
				Crontab:                  "*/10 * * * *",
				TTLSecondsAfterFinished:  "3600",
			},
		},
		{
			"pullSecrets with spaces",
			[]string{ // note trailing spaces
				"--repo-sync-image-pullsecrets=s1, s2",
				"--repo-sync-image-pullsecrets= s3",
			},
			server.Config{
				Kubeconfig:               "",
				MasterURL:                "",
				RepoSyncImage:            "docker.io/kubeapps/asset-syncer:latest",
				RepoSyncImagePullSecrets: []string{"s1", " s2", " s3"},
				ImagePullSecretsRefs:     []v1.LocalObjectReference{{Name: "s1"}, {Name: " s2"}, {Name: " s3"}},
				RepoSyncCommand:          "/chart-repo",
				KubeappsNamespace:        "kubeapps",
				ReposPerNamespace:        true,
				DBURL:                    "localhost",
				DBUser:                   "root",
				DBName:                   "charts",
				DBSecretName:             "kubeapps-db",
				DBSecretKey:              "postgresql-root-password",
				UserAgentComment:         "",
				Crontab:                  "*/10 * * * *",
				TTLSecondsAfterFinished:  "3600",
			},
		},
		{
			"all arguments are captured",
			[]string{
				"--kubeconfig", "foo01",
				"--master", "foo02",
				"--repo-sync-image", "foo03",
				"--repo-sync-image-pullsecrets", "s1,s2",
				"--repo-sync-image-pullsecrets", "s3",
				"--repo-sync-cmd", "foo04",
				"--namespace", "foo05",
				"--repos-per-namespace=false",
				"--database-url", "foo06",
				"--database-user", "foo07",
				"--database-name", "foo08",
				"--database-secret-name", "foo09",
				"--database-secret-key", "foo10",
				"--user-agent-comment", "foo11",
				"--crontab", "foo12",
			},
			server.Config{
				Kubeconfig:               "foo01",
				MasterURL:                "foo02",
				RepoSyncImage:            "foo03",
				RepoSyncImagePullSecrets: []string{"s1", "s2", "s3"},
				ImagePullSecretsRefs:     []v1.LocalObjectReference{{Name: "s1"}, {Name: "s2"}, {Name: "s3"}},
				RepoSyncCommand:          "foo04",
				KubeappsNamespace:        "foo05",
				ReposPerNamespace:        false,
				DBURL:                    "foo06",
				DBUser:                   "foo07",
				DBName:                   "foo08",
				DBSecretName:             "foo09",
				DBSecretKey:              "foo10",
				UserAgentComment:         "foo11",
				Crontab:                  "foo12",
				TTLSecondsAfterFinished:  "3600",
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
			serveOpts.ImagePullSecretsRefs = getImagePullSecretsRefs(serveOpts.RepoSyncImagePullSecrets)
			if got, want := serveOpts, tt.conf; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestParseFlagsError(t *testing.T) {
	var tests = []struct {
		name   string
		args   []string
		errstr string
	}{
		{
			"non-existent flag",
			[]string{"--foo"},
			"unknown flag: --foo",
		},
		{
			"flag with worng value type",
			[]string{"--repos-per-namespace=3"},
			"flag: strconv.ParseBool",
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
			if !strings.Contains(err.Error(), tt.errstr) {
				t.Errorf("err got %q, want to find %q", err.Error(), tt.errstr)
			}
		})
	}
}
