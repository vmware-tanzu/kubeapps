// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"net/http"
	"os"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/client-go/rest"
)

func TestNewClusterConfig(t *testing.T) {
	testCases := []struct {
		name            string
		userToken       string
		cluster         string
		clustersConfig  ClustersConfig
		inClusterConfig *rest.Config
		expectedConfig  *rest.Config
		errorExpected   bool
		maxReq          int
	}{
		{
			name:      "returns an in-cluster with explicit token for the default cluster",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
				},
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a cluster config with explicit apiServiceURL and cert even for the kubeapps default cluster, when specified",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:                   "https://proxy.example.com:7890",
						CertificateAuthorityData:        "Y2EtZmlsZS1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-file-data",
						CAFile:                          "/tmp/ca-file-data",
					},
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://proxy.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
				TLSClientConfig: rest.TLSClientConfig{
					CAData: []byte("ca-file-data"),
					CAFile: "/tmp/ca-file-data",
				},
			},
		},
		{
			name:      "returns an in-cluster config when the global packaging cluster token is specified",
			userToken: "token-1",
			cluster:   KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN,
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "",
				Clusters: map[string]ClusterConfig{
					"cluster-1": {
						APIServiceURL:                   "https://cluster-1.example.com:7890",
						CertificateAuthorityData:        "Y2EtZmlsZS1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-file-data",
						CAFile:                          "/tmp/ca-file-data",
					},
				},
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config setup for an additional cluster",
			userToken: "token-1",
			cluster:   "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
					"cluster-1": {
						APIServiceURL:                   "https://cluster-1.example.com:7890",
						CertificateAuthorityData:        "Y2EtZmlsZS1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-file-data",
						CAFile:                          "/tmp/ca-file-data",
					},
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://cluster-1.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
				TLSClientConfig: rest.TLSClientConfig{
					CAData: []byte("ca-file-data"),
					CAFile: "/tmp/ca-file-data",
				},
			},
		},
		{
			name:      "assumes a public cert if no ca data provided",
			userToken: "token-1",
			cluster:   "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
					"cluster-1": {
						APIServiceURL: "https://cluster-1.example.com:7890",
					},
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://cluster-1.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:            "returns an error if the cluster does not exist",
			cluster:         "cluster-1",
			inClusterConfig: &rest.Config{},
			errorExpected:   true,
		},
		{
			name:      "returns a config to proxy via pinniped-proxy",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "https://kubernetes.default",
						CertificateAuthorityData: "SGVsbG8K",
						PinnipedConfig:           PinnipedConciergeConfig{Enabled: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config to proxy via pinniped-proxy using the deprecated flag enable",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "https://kubernetes.default",
						CertificateAuthorityData: "SGVsbG8K",
						PinnipedConfig:           PinnipedConciergeConfig{Enable: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config to proxy via pinniped-proxy without headers for kubernetes.default",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "",
						CertificateAuthorityData: "",
						PinnipedConfig:           PinnipedConciergeConfig{Enabled: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := NewClusterConfig(tc.inClusterConfig, tc.userToken, tc.cluster, tc.clustersConfig)
			if got, want := err != nil, tc.errorExpected; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got, cmpopts.IgnoreFields(rest.Config{}, "WrapTransport")) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
			// If the test case defined a pinniped proxy url, verify that the expected headers
			// are added to the request.
			if clusterConfig, ok := tc.clustersConfig.Clusters[tc.cluster]; ok && clusterConfig.PinnipedConfig.Enabled {
				if config.WrapTransport == nil {
					t.Errorf("expected config.WrapTransport to be set but it is nil")
				} else {
					req := http.Request{}
					roundTripper := config.WrapTransport(&fakeRoundTripper{})
					_, err := roundTripper.RoundTrip(&req)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					want := http.Header{}
					if clusterConfig.APIServiceURL != "" {
						want["Pinniped_proxy_api_server_url"] = []string{clusterConfig.APIServiceURL}
					}
					if clusterConfig.CertificateAuthorityData != "" {

						want["Pinniped_proxy_api_server_cert"] = []string{clusterConfig.CertificateAuthorityData}
					}
					if got := req.Header; !cmp.Equal(want, got) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
					}
				}
			}
		})
	}
}

func TestParseClusterConfig(t *testing.T) {
	defaultPinnipedURL := "http://kubeapps-internal-pinniped-proxy.kubeapps:3333"
	testCases := []struct {
		name           string
		configJSON     string
		expectedErr    bool
		expectedConfig ClustersConfig
	}{
		{
			name:       "parses a single cluster",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedProxyURL": "http://172.0.1.18:3333", "isKubeappsCluster": true}]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
						IsKubeappsCluster:               true,
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses multiple clusters",
			configJSON: `[
	{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "isKubeappsCluster": true},
	{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						IsKubeappsCluster:               true,
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
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-1",
				Clusters: map[string]ClusterConfig{
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
			name: "parses config not specifying an explicit Kubeapps cluster",
			configJSON: `[
				{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
				{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
			]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN,
				Clusters: map[string]ClusterConfig{
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
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedConfig": {"enabled": true}, "isKubeappsCluster": true}]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
						PinnipedConfig: PinnipedConciergeConfig{
							Enabled: true,
						},
						IsKubeappsCluster: true,
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
		{
			name: "errors if more than one cluster with isKubeappsCluster=true is configured",
			configJSON: `[
		       {"name": "cluster-1", isKubeappsCluster: true},
		       {"name": "cluster-2", isKubeappsCluster: true }
		]`,
			expectedErr: true,
		},
		{
			name: "errors if both no APIServiceURL and isKubeappsCluster=true are configured",
			configJSON: `[
		       {"name": "cluster-1",  },
		       {"name": "cluster-2", isKubeappsCluster: true }
		]`,
			expectedErr: true,
		},
	}

	ignoreCAFile := cmpopts.IgnoreFields(ClusterConfig{}, "CAFile")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			path := createConfigFile(t, tc.configJSON)
			defer os.Remove(path)

			config, deferFn, err := ParseClusterConfig(path, "/tmp", defaultPinnipedURL, "")
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
					fileCAData, err := os.ReadFile(config.Clusters[clusterName].CAFile)
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
	tmpfile, err := os.CreateTemp("", "")
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
