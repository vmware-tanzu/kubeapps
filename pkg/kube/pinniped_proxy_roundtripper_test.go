// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type fakeRoundTripper struct{}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}

func TestPinnipedProxyRoundTripper(t *testing.T) {
	testCases := []struct {
		name            string
		existingHeaders http.Header
		pinnipedHeaders http.Header
	}{
		{
			name:            "it creates the headers if they don't exist",
			pinnipedHeaders: http.Header{},
		},
		{
			name: "it sets the pinniped headers",
			pinnipedHeaders: http.Header{
				"Some_header_1": []string{"some value 1"},
				"Some_header_2": []string{"some value 2"},
			},
		},
		{
			name: "it leaves existing headers intact",
			existingHeaders: http.Header{
				"Existing_header_1": []string{"existing value 1"},
				"Existing_header_2": []string{"existing value 2"},
			},
			pinnipedHeaders: http.Header{
				"Some_header_1": []string{"some value 1"},
				"Some_header_2": []string{"some value 2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pprt := pinnipedProxyRoundTripper{
				headers: tc.pinnipedHeaders,
				rt:      &fakeRoundTripper{},
			}
			req := http.Request{Header: tc.existingHeaders.Clone()}

			_, _ = pprt.RoundTrip(&req)

			if got, want := req.Header, mergeHeaders(tc.existingHeaders, tc.pinnipedHeaders); !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func mergeHeaders(h1, h2 http.Header) http.Header {
	result := http.Header{}
	for k, vv := range h1 {
		for _, v := range vv {
			result.Add(k, v)
		}
	}
	for k, vv := range h2 {
		for _, v := range vv {
			result.Add(k, v)
		}
	}
	return result
}
