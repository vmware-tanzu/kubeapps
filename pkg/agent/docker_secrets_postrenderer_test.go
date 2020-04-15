package agent

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDockerSecretsPostRenderer(t *testing.T) {
	testCases := []struct {
		name        string
		input       *bytes.Buffer
		secrets     map[string]string
		output      *bytes.Buffer
		expectedErr error
	}{
		{
			name:   "it returns the input without parsing when no secrets set",
			input:  bytes.NewBuffer([]byte(`anything at : all`)),
			output: bytes.NewBuffer([]byte(`anything at : all`)),
		},
		{
			name:        "it returns an error if the input cannot be parsed as yaml",
			input:       bytes.NewBuffer([]byte("v: [A,")),
			secrets:     map[string]string{"foo.example.com": "secret-name"},
			expectedErr: fmt.Errorf("yaml: line 1: did not find expected node content"),
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
other: doc
`)),
			secrets: map[string]string{"foo.example.com": "secret-name"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewDockerSecretsPostRenderer(tc.secrets)

			got, err := r.Run(tc.input)
			if err != nil {
				if got, want := err.Error(), tc.expectedErr.Error(); got != want {
					t.Fatalf("got: %q, want: %q", got, want)
				}
			} else if want := tc.expectedErr; nil != want {
				t.Fatalf("got: nil, want: %+v", want)
			}
			if tc.expectedErr != nil {
				return
			}

			if got, want := got.String(), tc.output.String(); !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
