package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
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
				version = tt.version
			}
			if tt.userAgentComment != "" {
				userAgentComment = tt.userAgentComment
			}
			assert.Equal(t, tt.expectedResult, userAgent(), "expected user agent")
		})
	}
}

func Test_clientWithDefaultUserAgentOverride(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "tiller-proxy/devel", req.Header.Get("User-Agent"), "expected user agent")
	}))
	// Close the server when test finishes
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	netClient = &clientWithDefaultUserAgent{}
	_, err := netClient.Do(&http.Request{
		URL:    serverURL,
		Header: map[string][]string{},
	})

	assert.NoErr(t, err)
}
