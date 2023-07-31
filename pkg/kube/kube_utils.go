// Copyright 2020-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

func GetAuthHeaderFromDockerConfig(dockerConfig *DockerConfigJSON) (string, error) {
	if len(dockerConfig.Auths) > 1 {
		return "", fmt.Errorf("the given config should include one auth entry")
	}
	// This is a simplified handler of a Docker config which only looks for the username:password
	// of the first entry.
	for _, entry := range dockerConfig.Auths {
		auth := fmt.Sprintf("%s:%s", entry.Username, entry.Password)
		authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
		return authHeader, nil
	}
	return "", fmt.Errorf("the given config doesn't include an auth entry")
}

// getDataFromRegistrySecret retrieves the given key from the secret as a string
func getDataFromRegistrySecret(key string, s *corev1.Secret) (string, error) {
	dockerConfigJson, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("secret %q did not contain key %q", s.Name, key)
	}

	dockerConfig := &DockerConfigJSON{}
	err := json.Unmarshal(dockerConfigJson, dockerConfig)
	if err != nil {
		return "", fmt.Errorf("unable to parse secret %s as a Docker config. Got: %v", s.Name, err)
	}

	return GetAuthHeaderFromDockerConfig(dockerConfig)
}

// GetDataFromSecret retrieves the given key from the secret as a string
func GetDataFromSecret(key string, s *corev1.Secret) (string, error) {
	if key == ".dockerconfigjson" {
		// Parse the secret as a docker registry secret
		return getDataFromRegistrySecret(key, s)
	}
	// Parse the secret as a plain secret
	auth, ok := s.StringData[key]
	if !ok {
		authBytes, ok := s.Data[key]
		if !ok {
			return "", fmt.Errorf("secret %q did not contain key %q", s.Name, key)
		}
		auth = string(authBytes)
	}
	return auth, nil
}
