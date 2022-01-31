// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3" // The usual "sigs.k8s.io/yaml" is not used because we're dealing with unstructured yaml directly
)

func TestNewDockerSecretsPostRenderer(t *testing.T) {
	testCases := []struct {
		name            string
		secrets         map[string]string
		expectedSecrets map[string]string
		expectErr       bool
	}{
		{
			name:            "it copies the secrets without changing the original",
			secrets:         map[string]string{"example.com": "secret-name"},
			expectedSecrets: map[string]string{"example.com": "secret-name"},
		},
		{
			name: "it reduces FQDNs to hosts for comparison",
			secrets: map[string]string{
				"https://example.com/": "secret-1",
			},
			expectedSecrets: map[string]string{
				"example.com": "secret-1",
			},
		},
		{
			name: "it includes both index.docker.io and docker.io",
			secrets: map[string]string{
				"https://index.docker.io/v1/": "dockerhub-secret",
			},
			expectedSecrets: map[string]string{
				"index.docker.io": "dockerhub-secret",
				"docker.io":       "dockerhub-secret",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewDockerSecretsPostRenderer(tc.secrets)
			if got, want := err != nil, tc.expectErr; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := r.secrets, tc.expectedSecrets; !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestDockerSecretsPostRenderer(t *testing.T) {
	testCases := []struct {
		name      string
		input     *bytes.Buffer
		secrets   map[string]string
		output    *bytes.Buffer
		expectErr bool
	}{
		{
			name:   "it returns the input without parsing when no secrets set",
			input:  bytes.NewBuffer([]byte(`anything at : all`)),
			output: bytes.NewBuffer([]byte(`anything at : all`)),
		},
		{
			name:      "it returns an error if the input cannot be parsed as yaml",
			input:     bytes.NewBuffer([]byte("v: [A,")),
			secrets:   map[string]string{"foo.example.com": "secret-name"},
			expectErr: true,
		},
		{
			name: "it re-renders the yaml with ordering and indent changes only",
			input: bytes.NewBuffer([]byte(`apiVersion: v1
kind: Pod
metadata:
  name: image-secret-test
  annotations:
    annotation-1: some-annotation
spec:
  containers:
    - command:
        - sh
        - -c
        - echo 'foo'
      env:
        - name: SOME_ENV
          value: env_value
      image: example.com/bitnami/nginx:1.16.1-debian-10-r42
      name: container-name
  restartPolicy: Never
---
kind: Unknown
other: doc
`)),
			output: bytes.NewBuffer([]byte(`apiVersion: v1
kind: Pod
metadata:
    annotations:
        annotation-1: some-annotation
    name: image-secret-test
spec:
    containers:
        - command:
            - sh
            - -c
            - echo 'foo'
          env:
            - name: SOME_ENV
              value: env_value
          image: example.com/bitnami/nginx:1.16.1-debian-10-r42
          name: container-name
    restartPolicy: Never
---
kind: Unknown
other: doc
`)),
			secrets: map[string]string{"foo.example.com": "secret-name"},
		},
		{
			name: "it appends relevant image pull secrets to pod specs",
			input: bytes.NewBuffer([]byte(`apiVersion: v1
kind: Pod
metadata:
  annotations:
    annotation-1: some-annotation
  name: image-secret-test
spec:
  containers:
  - command:
    - sh
    - -c
    - echo 'foo'
    env:
    - name: SOME_ENV
      value: env_value
    image: example.com/bitnami/nginx:1.16.1-debian-10-r42
    name: container-name
  restartPolicy: Never
---
kind: Unknown
other: doc
`)),
			output: bytes.NewBuffer([]byte(`apiVersion: v1
kind: Pod
metadata:
    annotations:
        annotation-1: some-annotation
    name: image-secret-test
spec:
    containers:
        - command:
            - sh
            - -c
            - echo 'foo'
          env:
            - name: SOME_ENV
              value: env_value
          image: example.com/bitnami/nginx:1.16.1-debian-10-r42
          name: container-name
    imagePullSecrets:
        - name: secret-1
    restartPolicy: Never
---
kind: Unknown
other: doc
`)),
			secrets: map[string]string{"example.com": "secret-1"},
		},
		{
			name: "it appends relevant image pull secret for nested lists of resources",
			input: bytes.NewBuffer([]byte(`apiVersion: v1
kind: PodTemplateList
metadata:
  annotations:
    annotation-1: some-annotation
  name: image-secret-test
items:
- kind: PodTemplate
  template:
    spec:
      containers:
      - command:
        - sh
        - -c
        - echo 'foo'
        env:
        - name: SOME_ENV
          value: env_value
        image: example.com/bitnami/nginx:1.16.1-debian-10-r42
        name: container-name
      restartPolicy: Never
- kind: PodTemplate
  template:
    spec:
      containers:
      - command:
        - sh
        - -c
        - echo 'bar'
        env:
        - name: SOME_ENV
          value: env_value
        image: example.com/bitnami/nginx:1.16.1-debian-10-r42
        name: container-name
      restartPolicy: Never
---
kind: Unknown
other: doc
`)),
			output: bytes.NewBuffer([]byte(`apiVersion: v1
items:
    - kind: PodTemplate
      template:
        spec:
            containers:
                - command:
                    - sh
                    - -c
                    - echo 'foo'
                  env:
                    - name: SOME_ENV
                      value: env_value
                  image: example.com/bitnami/nginx:1.16.1-debian-10-r42
                  name: container-name
            imagePullSecrets:
                - name: secret-1
            restartPolicy: Never
    - kind: PodTemplate
      template:
        spec:
            containers:
                - command:
                    - sh
                    - -c
                    - echo 'bar'
                  env:
                    - name: SOME_ENV
                      value: env_value
                  image: example.com/bitnami/nginx:1.16.1-debian-10-r42
                  name: container-name
            imagePullSecrets:
                - name: secret-1
            restartPolicy: Never
kind: PodTemplateList
metadata:
    annotations:
        annotation-1: some-annotation
    name: image-secret-test
---
kind: Unknown
other: doc
`)),
			secrets: map[string]string{"example.com": "secret-1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewDockerSecretsPostRenderer(tc.secrets)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			renderedManifests, err := r.Run(tc.input)
			if got, want := err != nil, tc.expectErr; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := renderedManifests.String(), tc.output.String(); !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestUpdatePodSpecWithPullSecrets(t *testing.T) {
	testCases := []struct {
		name                string
		podSpec             string
		secrets             map[string]string
		expectedPullSecrets interface{}
	}{
		{
			name: "it does not add image pull secrets when no secret matches",
			podSpec: `containers:
- image: "example.com/foobar:v1"`,
			secrets: map[string]string{
				"other.com": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it adds an image pull secret when one secret matches",
			podSpec: `containers:
- image: "example.com/foobar:v1"`,
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it adds multiple image pull secrets when multiple secrets matches",
			podSpec: `containers:
- image: "example.com/foobar:v1"
- image: "otherexample.com/foobar:v1"`,
			secrets: map[string]string{
				"example.com":      "secret-1",
				"otherexample.com": "secret-2",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
				{"name": "secret-2"},
			},
		},
		{
			name: "it appends to existing image pull secrets",
			podSpec: `containers:
- image: "example.com/foobar:v1"
- image: "otherexample.com/foobar:v1"
imagePullSecrets:
- name: "secret-1"`,
			secrets: map[string]string{
				"example.com":      "secret-2",
				"otherexample.com": "secret-3",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
				{"name": "secret-2"},
				{"name": "secret-3"},
			},
		},
		{
			name: "it does not duplicate existing image pull secrets",
			podSpec: `containers:
- image: "example.com/foobar:v1"
- image: "otherexample.com/foobar:v1"
imagePullSecrets:
- name: "secret-1"`,
			secrets: map[string]string{
				"example.com":      "secret-1",
				"otherexample.com": "secret-2",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
				{"name": "secret-2"},
			},
		},
		{
			name: "it does not mistake domainless image refs from dockerhub with a badly-named secret",
			podSpec: `containers:
- image: wordpress`,
			secrets: map[string]string{
				"wordpress": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it adds an explicit dockerhub secret when the registry server matches dockerhubs",
			podSpec: `containers:
- image: wordpress`,
			secrets: map[string]string{
				"https://index.docker.io/v1/": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it makes no changes if a containers key does not exist",
			podSpec: `notcontainers:
- image: example.com/foobar:v1`,
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name:                "it makes no changes if a containers value is not a slice",
			podSpec:             `containers: "not a slice"`,
			expectedPullSecrets: nil,
		},
		{
			name: "it ignores containers with non-map values while updating others",
			podSpec: `containers:
- "not a map"
- image: "example.com/foobar:v1"`,
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it ignores containers without an image key",
			podSpec: `containers:
- notimage: "somethingelse.com/foobar:v1"
- image: "example.com/foobar:v1"`,
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewDockerSecretsPostRenderer(tc.secrets)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			podSpec := make(map[string]interface{})

			// var podSpec interface{}
			err = yaml.Unmarshal([]byte(tc.podSpec), &podSpec)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			r.updatePodSpecWithPullSecrets(podSpec)

			if got, want := podSpec["imagePullSecrets"], tc.expectedPullSecrets; !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetResourcePodSpec(t *testing.T) {
	testCases := []struct {
		name     string
		kind     string
		resource map[string]interface{}
		result   map[string]interface{}
	}{
		{
			name: "it ignores an invalid doc with a non-map spec",
			kind: "Pod",
			resource: map[string]interface{}{
				"spec": "not a map",
			},
			result: nil,
		},
		{
			name: "it returns the pod spec from a pod",
			kind: "Pod",
			resource: map[string]interface{}{
				"spec": map[string]interface{}{"some": "spec"},
			},
			result: map[string]interface{}{
				"some": "spec",
			},
		},
		{
			name: "it returns the pod spec from a daemon set",
			kind: "DaemonSet",
			resource: map[string]interface{}{
				"kind": "DaemonSet",
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{"some": "spec"},
					},
				},
			},
			result: map[string]interface{}{
				"some": "spec",
			},
		},
		{
			name: "it returns the pod spec from a deployment",
			kind: "Deployment",
			resource: map[string]interface{}{
				"kind": "Deployment",
				"spec": map[string]interface{}{
					"template": map[string]interface{}{
						"spec": map[string]interface{}{"some": "spec"},
					},
				},
			},
			result: map[string]interface{}{
				"some": "spec",
			},
		},
		{
			name: "it returns the pod spec from a CronJob",
			kind: "CronJob",
			resource: map[string]interface{}{
				"kind": "CronJob",
				"spec": map[string]interface{}{
					"jobTemplate": map[string]interface{}{
						"spec": map[string]interface{}{
							"template": map[string]interface{}{
								"spec": map[string]interface{}{
									"some": "spec",
								},
							},
						},
					},
				},
			},
			result: map[string]interface{}{
				"some": "spec",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := getResourcePodSpec(tc.kind, tc.resource), tc.result; !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
