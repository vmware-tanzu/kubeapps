/*
Copyright (c) 2018 Bitnami

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

import (
	"testing"

	"github.com/arschles/assert"
)

func Test_userAgent(t *testing.T) {
	tests := []struct {
		name             string
		version          string
		userAgentComment string
		expectedResult   string
	}{
		{
			name:             "Shows default User-Agent unless comment nor version provided",
			version:          "",
			userAgentComment: "",
			expectedResult:   "tiller-proxy/devel",
		},
		{
			name:             "Shows just custom version unless comment provided",
			version:          "v4.4.4",
			userAgentComment: "",
			expectedResult:   "tiller-proxy/v4.4.4",
		},
		{
			name:             "Shows custom version plus comment if provided",
			version:          "v4.4.4",
			userAgentComment: "Kubeapps/v2.3.4",
			expectedResult:   "tiller-proxy/v4.4.4 (Kubeapps/v2.3.4)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override global variables used to generate the userAgent
			if tt.version != "" {
				defer func(origVersion string) { version = origVersion }(version)
				version = tt.version
			}
			if tt.userAgentComment != "" {
				defer func(origAgent string) { userAgentComment = origAgent }(userAgentComment)
				userAgentComment = tt.userAgentComment
			}
			assert.Equal(t, tt.expectedResult, userAgent(), "expected user agent")
		})
	}
}
