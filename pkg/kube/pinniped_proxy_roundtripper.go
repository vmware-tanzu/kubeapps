/*
Copyright (c) 2020 Bitnami

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
