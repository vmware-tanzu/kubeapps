package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/pkg/kube"
)

func TestParseAdditionalClusterConfig(t *testing.T) {
	testCases := []struct {
		name           string
		configJSON     string
		expectedErr    bool
		expectedConfig kube.AdditionalClustersConfig
	}{
		{
			name:       "parses a single additional cluster",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "abcd"}]`,
			expectedConfig: kube.AdditionalClustersConfig{
				"cluster-2": {
					Name:                     "cluster-2",
					APIServiceURL:            "https://example.com",
					CertificateAuthorityData: "abcd",
				},
			},
		},
		{
			name: "parses multiple additional clusters",
			configJSON: `[
	{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "abcd"},
	{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "efgh"}
]`,
			expectedConfig: kube.AdditionalClustersConfig{
				"cluster-2": {
					Name:                     "cluster-2",
					APIServiceURL:            "https://example.com/cluster-2",
					CertificateAuthorityData: "abcd",
				},
				"cluster-3": {
					Name:                     "cluster-3",
					APIServiceURL:            "https://example.com/cluster-3",
					CertificateAuthorityData: "efgh",
				},
			},
		},
		{
			name:        "errors if the cluster configs cannot be parsed",
			configJSON:  `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "extracomma",}]`,
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := createConfigFile(t, tc.configJSON)
			defer os.Remove(path)

			config, err := parseAdditionalClusterConfig(path)
			if got, want := err != nil, tc.expectedErr; got != want {
				t.Errorf("got: %t, want: %t", got, want)
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func createConfigFile(t *testing.T, content string) string {
	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("%+v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("%+v", err)
	}
	return tmpfile.Name()
}
