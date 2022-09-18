// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
)

const (
	DefaultTimeoutSeconds           int32 = 300
	DefaultGlobalPackagingNamespace       = ""
)

type HelmPluginConfig struct {
	VersionsInSummary pkgutils.VersionsInSummary
	TimeoutSeconds    int32
	// Whether secrets are fully managed by user or Kubeapps
	// see comments in design spec under AddPackageRepository.
	// false (i.e. Kubeapps manages secrets) by default
	UserManagedSecrets       bool
	GlobalPackagingNamespace string
}

func NewDefaultPluginConfig() *HelmPluginConfig {
	// If no config is provided, we default to the existing values for backwards
	// compatibility.
	return &HelmPluginConfig{
		VersionsInSummary:        pkgutils.GetDefaultVersionsInSummary(),
		TimeoutSeconds:           DefaultTimeoutSeconds,
		UserManagedSecrets:       false,
		GlobalPackagingNamespace: DefaultGlobalPackagingNamespace,
	}
}

// ParsePluginConfig parses the input plugin configuration json file and returns the configuration options.
func ParsePluginConfig(pluginConfigPath string) (*HelmPluginConfig, error) {

	// Note at present VersionsInSummary is the only configurable option for this plugin,
	// and if required this func can be enhanced to return helmConfig struct

	// In the helm plugin, for example, we are interested in config for the
	// core.packages.v1alpha1 only. So the plugin defines the following struct and parses the config.
	type (
		helmPluginConfig struct {
			Core struct {
				Packages struct {
					V1alpha1 struct {
						VersionsInSummary pkgutils.VersionsInSummary
						TimeoutSeconds    int32 `json:"timeoutSeconds"`
					} `json:"v1alpha1"`
				} `json:"packages"`
			} `json:"core"`

			Helm struct {
				Packages struct {
					V1alpha1 struct {
						GlobalPackagingNamespace string `json:"globalPackagingNamespace"`
						UserManagedSecrets       bool   `json:"userManagedSecrets"`
					} `json:"v1alpha1"`
				} `json:"packages"`
			} `json:"helm"`
		}
	)
	var config helmPluginConfig

	// #nosec G304
	pluginConfig, err := os.ReadFile(pluginConfigPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	// return configured value
	return &HelmPluginConfig{
		VersionsInSummary:        config.Core.Packages.V1alpha1.VersionsInSummary,
		TimeoutSeconds:           config.Core.Packages.V1alpha1.TimeoutSeconds,
		UserManagedSecrets:       config.Helm.Packages.V1alpha1.UserManagedSecrets,
		GlobalPackagingNamespace: config.Helm.Packages.V1alpha1.GlobalPackagingNamespace,
	}, nil
}
