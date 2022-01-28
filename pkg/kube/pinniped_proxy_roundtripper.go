// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import "net/http"

// pinnipedProxyRoundTripper is a round tripper that additionally sets
// any required headers for the exchange of credentials and request proxying.
// See https://github.com/kubernetes/client-go/issues/407
type pinnipedProxyRoundTripper struct {
	headers map[string][]string

	rt http.RoundTripper
}

func (p *pinnipedProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header == nil {
		req.Header = http.Header{}
	}
	for k, vv := range p.headers {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	return p.rt.RoundTrip(req)
}
