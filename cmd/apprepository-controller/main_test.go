package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf Config
	}{
		{
			"no arguments returns default flag values",
			[]string{},
			Config{
				Kubeconfig:               "",
				MasterURL:                "",
				RepoSyncImage:            "quay.io/helmpack/chart-repo:latest",
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
				Args:                     []string{},
			},
		},
		{
			"pullSecrets with spaces",
			[]string{ // note trailing spaces
				"--repo-sync-image-pullsecrets=s1, s2",
				"--repo-sync-image-pullsecrets= s3",
			},
			Config{
				Kubeconfig:               "",
				MasterURL:                "",
				RepoSyncImage:            "quay.io/helmpack/chart-repo:latest",
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
				Args:                     []string{},
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
			Config{
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
				Args:                     []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			conf, output, err := parseFlags("program", tt.args)
			conf.ImagePullSecretsRefs = getImagePullSecretsRefs(conf.RepoSyncImagePullSecrets)

			if err != nil {
				t.Errorf("err got:\n%v\nwant nil", err)
			}
			if output != "" {
				t.Errorf("output got:\n%q\nwant empty", output)
			}
			if got, want := *conf, tt.conf; !cmp.Equal(want, got) {
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
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			conf, _, err := parseFlags("prog", tt.args)
			if conf != nil {
				t.Errorf("conf got %v, want nil", conf)
			}
			if strings.Index(err.Error(), tt.errstr) < 0 {
				t.Errorf("err got %q, want to find %q", err.Error(), tt.errstr)
			}
		})
	}
}
