// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"testing"
)

func TestParseGrpcEndpoint(t *testing.T) {
	testCases := []struct {
		name             string
		url              string
		expectedEndpoint string
		expectedErr      bool
	}{
		{
			name:             "url without scheme or port is correctly converted defaulting to https",
			url:              "thehost",
			expectedEndpoint: "thehost:443",
		},
		{
			name:             "url with http scheme but no port is correctly converted",
			url:              "http://thehost",
			expectedEndpoint: "thehost:80",
		},
		{
			name:             "url with https scheme but no port is correctly converted",
			url:              "https://thehost",
			expectedEndpoint: "thehost:443",
		},
		{
			name:             "url without scheme but with port is correctly processed",
			url:              "thehost:7823",
			expectedEndpoint: "thehost:7823",
		},
		{
			name:             "url with scheme and port is correctly converted with preference to port",
			url:              "https://thehost:7823",
			expectedEndpoint: "thehost:7823",
		},
		{
			name:        "url with unknown scheme fails",
			url:         "customScheme://thehost",
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := parseGrpcEndpoint(tc.url)

			if got, want := err != nil, tc.expectedErr; got != want {
				t.Fatalf("got: %+v, want: %+v, error: %v", got, want, err)
			}

			if got, want := s, tc.expectedEndpoint; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
		})
	}
}

func TestIsValidClusterName(t *testing.T) {
	testCases := []struct {
		name        string
		clusterName string
		valid       bool
	}{
		{
			name:        "correct cluster name is valid",
			clusterName: "maincluster",
			valid:       true,
		},
		{
			name:        "cluster name with dashes is valid",
			clusterName: "main-cluster",
			valid:       true,
		},
		{
			name:        "cluster name with numbers is valid",
			clusterName: "main-cluster-2",
			valid:       true,
		},
		{
			name:        "cluster name with dots is valid",
			clusterName: "clusters.main",
			valid:       true,
		},
		{
			name:        "any other char in cluster name is not valid",
			clusterName: "*notallowedchars",
			valid:       false,
		},
		{
			name:        "break lines are not valid in cluster names",
			clusterName: "line1\nline2",
			valid:       false,
		},
		{
			name:        "spaces are disallowed in cluster name",
			clusterName: "my cluster",
			valid:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := isValidClusterName(tc.clusterName)

			if got, want := isValid, tc.valid; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
		})
	}
}
