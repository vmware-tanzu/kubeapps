/*
Copyright 2021 VMware.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import "testing"

func TestGetUserAgent(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		comment  string
		expected string
	}{
		{
			name:     "creates a user agent without a comment",
			version:  "2.1.6",
			expected: "kubeops/2.1.6",
		},
		{
			name:     "creates a user agent with comment",
			version:  "2.1.6",
			comment:  "foobar",
			expected: "kubeops/2.1.6 (foobar)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := getUserAgent(tc.version, tc.comment), tc.expected; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}

}
