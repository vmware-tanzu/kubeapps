package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/kubeapps/pkg/kube"
)

func TestParseClusterConfig(t *testing.T) {
	testCases := []struct {
		name           string
		configJSON     string
		expectedErr    bool
		expectedConfig kube.ClustersConfig
	}{
		{
			name:       "defaults the kubeapps cluster when passed an empty list of clusters",
			configJSON: `[]`,
			expectedConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters:            map[string]kube.ClusterConfig{},
			},
		},
		{
			name:       "parses a single cluster",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd"}]`,
			expectedConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"cluster-2": {
						Name:                     "cluster-2",
						APIServiceURL:            "https://example.com",
						CertificateAuthorityData: "ca-cert-data\n",
						ServiceToken:             "abcd",
					},
				},
			},
		},
		{
			name: "parses multiple clusters",
			configJSON: `[
	{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
	{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"cluster-2": {
						Name:                     "cluster-2",
						APIServiceURL:            "https://example.com/cluster-2",
						CertificateAuthorityData: "ca-cert-data\n",
					},
					"cluster-3": {
						Name:                     "cluster-3",
						APIServiceURL:            "https://example.com/cluster-3",
						CertificateAuthorityData: "ca-cert-data-additional\n",
					},
				},
			},
		},
		{
			name:        "errors if the cluster configs cannot be parsed",
			configJSON:  `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "extracomma",}]`,
			expectedErr: true,
		},
		{
			name:        "errors if any CAData cannot be decoded",
			configJSON:  `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "not-base64-encoded"}]`,
			expectedErr: true,
		},
	}

	ignoreCAFile := cmpopts.IgnoreFields(kube.ClusterConfig{}, "CAFile")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			path := createConfigFile(t, tc.configJSON)
			defer os.Remove(path)

			config, deferFn, err := parseClusterConfig(path, "/tmp")
			if got, want := err != nil, tc.expectedErr; got != want {
				t.Errorf("got: %t, want: %t: err: %+v", got, want, err)
			}
			if !tc.expectedErr {
				defer deferFn()
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got, ignoreCAFile) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreCAFile))
			}

			for clusterName, clusterConfig := range tc.expectedConfig.Clusters {
				if clusterConfig.CertificateAuthorityData != "" {
					fileCAData, err := ioutil.ReadFile(config.Clusters[clusterName].CAFile)
					if err != nil {
						t.Fatalf("%+v", err)
					}
					if got, want := string(fileCAData), clusterConfig.CertificateAuthorityData; got != want {
						t.Errorf("got: %q, want: %q", got, want)
					}
				}
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
