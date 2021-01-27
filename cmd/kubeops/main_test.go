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
			name:       "parses a single cluster",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedProxyURL": "http://172.0.1.18:3333"}]`,
			expectedConfig: kube.ClustersConfig{
				Clusters: map[string]kube.ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses multiple clusters",
			configJSON: `[
	{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
	{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: kube.ClustersConfig{
				Clusters: map[string]kube.ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
					},
					"cluster-3": {
						Name:                            "cluster-3",
						APIServiceURL:                   "https://example.com/cluster-3",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK",
						CertificateAuthorityDataDecoded: "ca-cert-data-additional\n",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses a cluster without a service URL as the Kubeapps cluster",
			configJSON: `[
       {"name": "cluster-1" },
       {"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
       {"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: kube.ClustersConfig{
				KubeappsClusterName: "cluster-1",
				Clusters: map[string]kube.ClusterConfig{
					"cluster-1": {
						Name: "cluster-1",
					},
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
					},
					"cluster-3": {
						Name:                            "cluster-3",
						APIServiceURL:                   "https://example.com/cluster-3",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK",
						CertificateAuthorityDataDecoded: "ca-cert-data-additional\n",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name:       "parses a cluster with pinniped token exchange",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedConfig": {"enable": true}}]`,
			expectedConfig: kube.ClustersConfig{
				Clusters: map[string]kube.ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
						PinnipedConfig: kube.PinnipedConciergeConfig{
							Enable: true,
						},
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
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
		{
			name: "errors if more than one cluster without an api service URL is configured",
			configJSON: `[
       {"name": "cluster-1" },
       {"name": "cluster-2" }
]`,
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
				if clusterConfig.CertificateAuthorityDataDecoded != "" {
					fileCAData, err := ioutil.ReadFile(config.Clusters[clusterName].CAFile)
					if err != nil {
						t.Fatalf("error opening %s: %+v", config.Clusters[clusterName].CAFile, err)
					}
					if got, want := string(fileCAData), clusterConfig.CertificateAuthorityDataDecoded; got != want {
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
