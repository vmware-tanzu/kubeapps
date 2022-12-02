// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type OperatorsPluginConfig struct {
}

func NewDefaultPluginConfig() *OperatorsPluginConfig {
	// If no config is provided, we default to the existing values for backwards compatibility.
	return &OperatorsPluginConfig{}
}

// ParsePluginConfig parses the input plugin configuration json file and returns the configuration options.
func ParsePluginConfig(pluginConfigPath string) (*OperatorsPluginConfig, error) {

	// Operators plugin config defines the following struct and json config
	type operatorsConfig struct {
		Operators struct {
			Packages struct {
				V1alpha1 struct {
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"resources"`
	}
	var config operatorsConfig

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
	return &OperatorsPluginConfig{}, nil
}
