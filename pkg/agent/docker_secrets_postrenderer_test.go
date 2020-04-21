package agent

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
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
		podSpec             map[interface{}]interface{}
		secrets             map[string]string
		expectedPullSecrets interface{}
	}{
		{
			name: "it does not add image pull secrets when no secret matches",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
				},
			},
			secrets: map[string]string{
				"other.com": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it adds an image pull secret when one secret matches",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
				},
			},
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it adds multiple image pull secrets when multiple secrets matches",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
					map[interface{}]interface{}{
						"image": "otherexample.com/foobar:v1",
					},
				},
			},
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
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
					map[interface{}]interface{}{
						"image": "otherexample.com/foobar:v1",
					},
				},
				"imagePullSecrets": []map[string]interface{}{
					map[string]interface{}{
						"name": "secret-1",
					},
				},
			},
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
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
					map[interface{}]interface{}{
						"image": "otherexample.com/foobar:v1",
					},
				},
				"imagePullSecrets": []map[string]interface{}{
					map[string]interface{}{
						"name": "secret-1",
					},
				},
			},
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
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "wordpress",
					},
				},
			},
			secrets: map[string]string{
				"wordpress": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it adds an explicit dockerhub secret when the registry server matches dockerhubs",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"image": "wordpress",
					},
				},
			},
			secrets: map[string]string{
				"https://index.docker.io/v1/": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it makes no changes if a containers key does not exist",
			podSpec: map[interface{}]interface{}{
				"notcontainers": []interface{}{
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
				},
			},
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it makes no changes if a containers value is not a slice",
			podSpec: map[interface{}]interface{}{
				"containers": "not a slice",
			},
			expectedPullSecrets: nil,
		},
		{
			name: "it ignores containers with non-map values while updating others",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					"not a map",
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
				},
			},
			secrets: map[string]string{
				"example.com": "secret-1",
			},
			expectedPullSecrets: []map[string]interface{}{
				{"name": "secret-1"},
			},
		},
		{
			name: "it ignores containers without an image key",
			podSpec: map[interface{}]interface{}{
				"containers": []interface{}{
					map[interface{}]interface{}{
						"notimage": "somethingelse.com/foobar:v1",
					},
					map[interface{}]interface{}{
						"image": "example.com/foobar:v1",
					},
				},
			},
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

			r.updatePodSpecWithPullSecrets(tc.podSpec)

			if got, want := tc.podSpec["imagePullSecrets"], tc.expectedPullSecrets; !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetResourcePodSpec(t *testing.T) {
	testCases := []struct {
		name     string
		resource map[interface{}]interface{}
		result   map[interface{}]interface{}
	}{
		{
			name: "it ignores an invalid doc without a kind",
			resource: map[interface{}]interface{}{
				"notkind": "Pod",
				"spec":    map[string]interface{}{"some": "spec"},
			},
			result: nil,
		},
		{
			name: "it ignores an invalid doc with a non-string kind",
			resource: map[interface{}]interface{}{
				"kind": map[string]interface{}{"not": "string"},
				"spec": map[interface{}]interface{}{"some": "spec"},
			},
			result: nil,
		},
		{
			name: "it ignores an invalid doc with a non-map spec",
			resource: map[interface{}]interface{}{
				"kind": "Pod",
				"spec": "not a map",
			},
			result: nil,
		},
		{
			name: "it returns the pod spec from a pod",
			resource: map[interface{}]interface{}{
				"kind": "Pod",
				"spec": map[interface{}]interface{}{"some": "spec"},
			},
			result: map[interface{}]interface{}{
				"some": "spec",
			},
		},
		{
			name: "it returns the pod spec from a daemon set",
			resource: map[interface{}]interface{}{
				"kind": "DaemonSet",
				"spec": map[interface{}]interface{}{
					"template": map[interface{}]interface{}{
						"spec": map[interface{}]interface{}{"some": "spec"},
					},
				},
			},
			result: map[interface{}]interface{}{
				"some": "spec",
			},
		},
		{
			name: "it returns the pod spec from a deployment",
			resource: map[interface{}]interface{}{
				"kind": "Deployment",
				"spec": map[interface{}]interface{}{
					"template": map[interface{}]interface{}{
						"spec": map[interface{}]interface{}{"some": "spec"},
					},
				},
			},
			result: map[interface{}]interface{}{
				"some": "spec",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := getResourcePodSpec(tc.resource), tc.result; !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
