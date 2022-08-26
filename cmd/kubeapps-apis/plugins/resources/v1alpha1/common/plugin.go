// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"
	"os"
)

type ResourcesPluginConfig struct {
	TrustedNamespaces TrustedNamespaces
}

type TrustedNamespaces struct {
	HeaderName    string
	HeaderPattern string
}

func NewDefaultPluginConfig() *ResourcesPluginConfig {
	// If no config is provided, we default to the existing values for backwards compatibility.
	return &ResourcesPluginConfig{}
}

// ParsePluginConfig parses the input plugin configuration json file and returns the configuration options.
func ParsePluginConfig(pluginConfigPath string) (*ResourcesPluginConfig, error) {

	// Resources plugin config defines the following struct and json config
	type resourcesConfig struct {
		Resources struct {
			Packages struct {
				V1alpha1 struct {
					TrustedNamespaces struct {
						HeaderName    string `json:"headerName"`
						HeaderPattern string `json:"headerPattern"`
					} `json:"trustedNamespaces"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"resources"`
	}
	var config resourcesConfig

	// #nosec G304
	pluginConfig, err := os.ReadFile(pluginConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal(pluginConfig, &config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	// return configured value
	return &ResourcesPluginConfig{
		TrustedNamespaces: TrustedNamespaces{
			HeaderName:    config.Resources.Packages.V1alpha1.TrustedNamespaces.HeaderName,
			HeaderPattern: config.Resources.Packages.V1alpha1.TrustedNamespaces.HeaderPattern,
		},
	}, nil
}
