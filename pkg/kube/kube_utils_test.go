// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func Test_getDataFromRegistrySecret(t *testing.T) {
	testCases := []struct {
		name     string
		secret   *corev1.Secret
		expected string
	}{
		{
			name: "retrieves username and password from a dockerconfigjson",
			secret: &corev1.Secret{
				Data: map[string][]byte{
					".dockerconfigjson": []byte(`{"auths":{"foo":{"username":"foo","password":"bar"}}}`),
				},
			},
			// Basic: base64(foo:bar)
			expected: "Basic Zm9vOmJhcg==",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth, err := getDataFromRegistrySecret(".dockerconfigjson", tc.secret)
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			if auth != tc.expected {
				t.Errorf("Expecting %s, got %s", tc.expected, auth)
			}
		})
	}
}
