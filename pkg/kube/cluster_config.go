// Copyright 2019-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
)

const (
	// KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN
	// Kubeapps can be configured such that users cannot target the cluster
	// on which Kubeapps is itself installed (ie. it's not listed in the
	// clusters config). In this specific case, there is no way to refer
	// to a configured name for the global packaging cluster, so we define
	// one to be used that does not clash with user-configurable names.
	KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN = "-"
)

// ClusterConfig contains required info to talk to additional clusters.
type ClusterConfig struct {
	Name                     string `json:"name"`
	APIServiceURL            string `json:"apiServiceURL"`
	CertificateAuthorityData string `json:"certificateAuthorityData,omitempty"`
	// When parsing config we decode the cert auth data to ensure it is valid
	// and store it since it's required when using the data.
	CertificateAuthorityDataDecoded string
	// The genericclioptions.ConfigFlags struct includes only a CAFile field, not
	// a CAData field.
	// https://github.com/kubernetes/cli-runtime/issues/8
	// Embedding genericclioptions.ConfigFlags in a struct which includes the actual rest.Config
	// and returning that for ToRESTConfig() isn't enough, so we each configured cert out and
	// include a CAFile field in the config.
	CAFile string
	// ServiceToken can be configured so that the Kubeapps application itself
	// has access to get all namespaces on additional clusters, for example. It
	// should *not* be for reading secrets or similar, but limited to the
	// required functionality.
	ServiceToken string

	// Insecure should only be used in test or development environments and enables
	// TLS requests without requiring the cert authority validation.
	Insecure bool `json:"insecure"`

	// PinnipedConfig is an optional per-cluster configuration specifying
	// the pinniped namespace, authenticator type and authenticator name
	// that should be used for any credential exchange.
	PinnipedConfig PinnipedConciergeConfig `json:"pinnipedConfig,omitempty"`

	// IsKubeappsCluster is an optional per-cluster configuration specifying
	// that this cluster is the one in which Kubeapps is being installed.
	// Often this is inferred as the cluster without an explicit APIServiceURL, but
	// if every cluster defines an APIServiceURL, we can no longer infer the cluster
	// on which Kubeapps is installed.
	IsKubeappsCluster bool `json:"isKubeappsCluster,omitempty"`
}

// PinnipedConciergeConfig enables each cluster configuration to specify the
// pinniped-concierge installation to use for any credential exchange.
type PinnipedConciergeConfig struct {
	// Enabled flags whether this cluster should use
	// pinniped to exchange credentials.
	Enabled bool `json:"enabled"`
	// Enable is deprecated and will be removed in a future release.
	Enable bool `json:"enable"`
	// The Namespace, AuthenticatorType and Authenticator name to use
	// when exchanging credentials.
	Namespace         string `json:"namespace,omitempty"`
	AuthenticatorType string `json:"authenticatorType,omitempty"`
	AuthenticatorName string `json:"authenticatorName,omitempty"`
}

// ClustersConfig is an alias for a map of additional cluster configs.
type ClustersConfig struct {
	KubeappsClusterName      string
	GlobalPackagingNamespace string
	PinnipedProxyURL         string
	PinnipedProxyCACert      string
	Clusters                 map[string]ClusterConfig
}

// NewClusterConfig returns a copy of an in-cluster config with a user token
// and/or custom cluster host
func NewClusterConfig(inClusterConfig *rest.Config, userToken string, cluster string, clustersConfig ClustersConfig) (*rest.Config, error) {
	config := rest.CopyConfig(inClusterConfig)
	config.BearerToken = userToken
	config.BearerTokenFile = ""

	// If the cluster name is the Kubeapps global packaging cluster then the
	// inClusterConfig is already correct. This can be the case when the cluster
	// on which Kubeapps is installed is not one presented in the UI as a target
	// (hence not in the `clusters` configuration).
	if IsKubeappsClusterRef(cluster) {
		return config, nil
	}

	clusterConfig, ok := clustersConfig.Clusters[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster %q has no configuration", cluster)
	}

	if userToken != "" && (clusterConfig.PinnipedConfig.Enabled || clusterConfig.PinnipedConfig.Enable) {
		// Create a config for routing requests via the pinniped-proxy for credential
		// exchange.
		config.Host = clustersConfig.PinnipedProxyURL
		// set roundtripper.
		// https://github.com/kubernetes/client-go/issues/407
		existingWrapTransport := config.WrapTransport
		config.WrapTransport = func(rt http.RoundTripper) http.RoundTripper {
			if existingWrapTransport != nil {
				rt = existingWrapTransport(rt)
			}
			headers := map[string][]string{}
			if clusterConfig.APIServiceURL != "" {
				headers["PINNIPED_PROXY_API_SERVER_URL"] = []string{clusterConfig.APIServiceURL}
			}
			if clusterConfig.CertificateAuthorityData != "" {
				headers["PINNIPED_PROXY_API_SERVER_CERT"] = []string{clusterConfig.CertificateAuthorityData}
			}
			return &pinnipedProxyRoundTripper{
				headers: headers,
				rt:      rt,
			}
		}

		// If pinniped-proxy is configured with TLS, we need to set the
		// CACert.
		config.CAFile = clustersConfig.PinnipedProxyCACert

		return config, nil
	}

	// We cannot assume that if the cluster is the kubeapps cluster that we simply return
	// the incluster config, because some users set proxies in front of their clusters in
	// which case the incluster kubernetes.default will skip the proxy.
	if cluster == clustersConfig.KubeappsClusterName && clusterConfig.APIServiceURL == "" {
		return config, nil
	}

	config.Host = clusterConfig.APIServiceURL
	config.TLSClientConfig = rest.TLSClientConfig{}
	config.TLSClientConfig.Insecure = clusterConfig.Insecure
	if clusterConfig.CertificateAuthorityDataDecoded != "" {
		config.TLSClientConfig.CAData = []byte(clusterConfig.CertificateAuthorityDataDecoded)
		config.CAFile = clusterConfig.CAFile
	}
	return config, nil
}

func ParseClusterConfig(configPath, caFilesPrefix string, pinnipedProxyURL, PinnipedProxyCACert string) (ClustersConfig, func(), error) {
	caFilesDir, err := os.MkdirTemp(caFilesPrefix, "")
	if err != nil {
		return ClustersConfig{}, func() {}, err
	}
	deferFn := func() {
		err = os.RemoveAll(caFilesDir)
	}

	// #nosec G304
	content, err := os.ReadFile(configPath)
	if err != nil {
		return ClustersConfig{}, deferFn, err
	}

	var clusterConfigs []ClusterConfig
	if err = json.Unmarshal(content, &clusterConfigs); err != nil {
		return ClustersConfig{}, deferFn, err
	}

	configs := ClustersConfig{Clusters: map[string]ClusterConfig{}}
	configs.PinnipedProxyURL = pinnipedProxyURL
	configs.PinnipedProxyCACert = PinnipedProxyCACert
	for _, c := range clusterConfigs {
		// Select the cluster in which Kubeapps in installed. We look for either
		// `isKubeappsCluster: true` or an empty `APIServiceURL`.
		isKubeappsClusterCandidate := c.IsKubeappsCluster || c.APIServiceURL == ""
		if isKubeappsClusterCandidate {
			if configs.KubeappsClusterName == "" {
				configs.KubeappsClusterName = c.Name
			} else {
				return ClustersConfig{}, nil, fmt.Errorf("only one cluster can be configured using either 'isKubeappsCluster: true' or without an apiServiceURL to refer to the cluster on which Kubeapps is installed, two defined: %q, %q", configs.KubeappsClusterName, c.Name)
			}
		}

		// We need to decode the base64-encoded cadata from the input.
		if c.CertificateAuthorityData != "" {
			decodedCAData, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
			if err != nil {
				return ClustersConfig{}, deferFn, err
			}
			c.CertificateAuthorityDataDecoded = string(decodedCAData)

			// We also need a CAFile field because Helm uses the genericclioptions.ConfigFlags
			// struct which does not support CAData.
			// https://github.com/kubernetes/cli-runtime/issues/8
			c.CAFile = filepath.Join(caFilesDir, c.Name)
			// #nosec G306
			// TODO(agamez): check if we can set perms to 0600 instead of 0644.
			err = os.WriteFile(c.CAFile, decodedCAData, 0644)
			if err != nil {
				return ClustersConfig{}, deferFn, err
			}
		}
		configs.Clusters[c.Name] = c
	}
	// If the cluster on which Kubeapps is installed was not present in
	// the clusters config, we explicitly use a token to identify this
	// cluster when needed (such as for global available packages).
	if configs.KubeappsClusterName == "" {
		configs.KubeappsClusterName = KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN
	}
	return configs, deferFn, nil
}

// IsKubeappsClusterRef checks if the provided cluster name references the global packaging Kubeapps cluster
func IsKubeappsClusterRef(cluster string) bool {
	return cluster == "" || cluster == KUBEAPPS_GLOBAL_PACKAGING_CLUSTER_TOKEN
}
