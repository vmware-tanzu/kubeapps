// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"strings"

	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/server"
	v1 "k8s.io/api/core/v1"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name        string
		args        []string
		conf        server.Config
		errExpected bool
	}{
		{
			"no arguments returns default flag values",
			[]string{},
			server.Config{
				Kubeconfig:               "",
				APIServerURL:             "",
				RepoSyncImage:            "docker.io/kubeapps/asset-syncer:latest",
				RepoSyncImagePullSecrets: nil,
				RepoSyncCommand:          "/chart-repo",
				KubeappsNamespace:        "kubeapps",
				GlobalPackagingNamespace: "kubeapps",
				ReposPerNamespace:        true,
				DBURL:                    "localhost",
				DBUser:                   "root",
				DBName:                   "charts",
				DBSecretName:             "kubeapps-db",
				DBSecretKey:              "postgresql-root-password",
				UserAgentComment:         "",
				Crontab:                  "*/10 * * * *",
				TTLSecondsAfterFinished:  "3600",
				CustomAnnotations:        []string{""},
				CustomLabels:             []string{""},
				ParsedCustomAnnotations:  map[string]string{},
				ParsedCustomLabels:       map[string]string{},
			},
			true,
		},
		{
			"pullSecrets with spaces",
			[]string{ // note trailing spaces
				"--repo-sync-image-pullsecrets=s1, s2",
				"--repo-sync-image-pullsecrets= s3",
			},
			server.Config{
				Kubeconfig:               "",
				APIServerURL:             "",
				RepoSyncImage:            "docker.io/kubeapps/asset-syncer:latest",
				RepoSyncImagePullSecrets: []string{"s1", " s2", " s3"},
				ImagePullSecretsRefs:     []v1.LocalObjectReference{{Name: "s1"}, {Name: " s2"}, {Name: " s3"}},
				RepoSyncCommand:          "/chart-repo",
				KubeappsNamespace:        "kubeapps",
				GlobalPackagingNamespace: "kubeapps",
				ReposPerNamespace:        true,
				DBURL:                    "localhost",
				DBUser:                   "root",
				DBName:                   "charts",
				DBSecretName:             "kubeapps-db",
				DBSecretKey:              "postgresql-root-password",
				UserAgentComment:         "",
				Crontab:                  "*/10 * * * *",
				TTLSecondsAfterFinished:  "3600",
				CustomAnnotations:        []string{""},
				CustomLabels:             []string{""},
				ParsedCustomAnnotations:  map[string]string{},
				ParsedCustomLabels:       map[string]string{},
			},
			true,
		},
		{
			"all arguments are captured",
			[]string{
				"--kubeconfig", "foo01",
				"--apiserver", "foo02",
				"--repo-sync-image", "foo03",
				"--repo-sync-image-pullsecrets", "s1,s2",
				"--repo-sync-image-pullsecrets", "s3",
				"--repo-sync-cmd", "foo04",
				"--namespace", "foo05",
				"--global-repos-namespace", "kubeapps-repos-global",
				"--repos-per-namespace=false",
				"--database-url", "foo06",
				"--database-user", "foo07",
				"--database-name", "foo08",
				"--database-secret-name", "foo09",
				"--database-secret-key", "foo10",
				"--user-agent-comment", "foo11",
				"--crontab", "foo12",
				"--custom-annotations", "foo13=bar13,foo13x=bar13x",
				"--custom-annotations", "extra13=extra13",
				"--custom-labels", "foo14=bar14,foo14x=bar14x",
			},
			server.Config{
				Kubeconfig:               "foo01",
				APIServerURL:             "foo02",
				RepoSyncImage:            "foo03",
				RepoSyncImagePullSecrets: []string{"s1", "s2", "s3"},
				ImagePullSecretsRefs:     []v1.LocalObjectReference{{Name: "s1"}, {Name: "s2"}, {Name: "s3"}},
				RepoSyncCommand:          "foo04",
				KubeappsNamespace:        "foo05",
				GlobalPackagingNamespace: "kubeapps-repos-global",
				ReposPerNamespace:        false,
				DBURL:                    "foo06",
				DBUser:                   "foo07",
				DBName:                   "foo08",
				DBSecretName:             "foo09",
				DBSecretKey:              "foo10",
				UserAgentComment:         "foo11",
				Crontab:                  "foo12",
				TTLSecondsAfterFinished:  "3600",
				CustomAnnotations:        []string{"foo13=bar13", "foo13x=bar13x", "extra13=extra13"},
				CustomLabels:             []string{"foo14=bar14", "foo14x=bar14x"},
				ParsedCustomAnnotations:  map[string]string{"foo13": "bar13", "foo13x": "bar13x", "extra13": "extra13"},
				ParsedCustomLabels:       map[string]string{"foo14": "bar14", "foo14x": "bar14x"},
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
