// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes/fake"

	"context"
	"testing"
)

func TestGetRegistrySecretsPerDomain(t *testing.T) {
	const (
		userAuthToken = "ignored"
		namespace     = "user-namespace"
		// Secret created with
		// k create secret docker-registry test-secret --dry-run --docker-email=a@b.com --docker-password='password' --docker-username='username' --docker-server='https://index.docker.io/v1/' -o yaml
		indexDockerIOCred   = `{"auths":{"https://index.docker.io/v1/":{"username":"username","password":"password","email":"a@b.com","auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`
		otherExampleComCred = `{"auths":{"other.example.com":{"username":"username","password":"password","email":"a@b.com","auth":"dXNlcm5hbWU6cGFzc3dvcmQ="}}}`
	)

	testCases := []struct {
		name             string
		secretNames      []string
		existingSecrets  []runtime.Object
		secretsPerDomain map[string]string
		expectError      bool
	}{
		{
			name:             "it returns an empty map if there are no secret names",
			secretNames:      nil,
			secretsPerDomain: map[string]string{},
		},
		{
			name:        "it returns an error if a secret does not exist",
			secretNames: []string{"should-exist"},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret is not a dockerConfigJSON type",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte("whatevs"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret data does not have .dockerconfigjson key",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"custom-secret-key": []byte("whatevs"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns an error if the secret .dockerconfigjson value is not json decodable",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: "Opaque",
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte("not json"),
					},
				},
			},
			expectError: true,
		},
		{
			name:        "it returns the registry secrets per domain",
			secretNames: []string{"bitnami-repo"},
			existingSecrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(indexDockerIOCred),
					},
				},
			},
			secretsPerDomain: map[string]string{
				"https://index.docker.io/v1/": "bitnami-repo",
			},
		},
		{
			name:        "it includes secrets for multiple servers",
			secretNames: []string{"bitnami-repo1", "bitnami-repo2"},
			existingSecrets: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo1",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(indexDockerIOCred),
					},
				},
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bitnami-repo2",
						Namespace: namespace,
					},
					Type: dockerConfigJSONType,
					Data: map[string][]byte{
						dockerConfigJSONKey: []byte(otherExampleComCred),
					},
				},
			},
			secretsPerDomain: map[string]string{
				"https://index.docker.io/v1/": "bitnami-repo1",
				"other.example.com":           "bitnami-repo2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tc.existingSecrets...)

			secretsPerDomain, err := RegistrySecretsPerDomain(context.Background(), tc.secretNames, namespace, client)
			if got, want := err != nil, tc.expectError; !cmp.Equal(got, want) {
				t.Fatalf("got: %t, want: %t, err was: %+v", got, want, err)
			}
			if err != nil {
				return
			}

			if got, want := secretsPerDomain, tc.secretsPerDomain; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
