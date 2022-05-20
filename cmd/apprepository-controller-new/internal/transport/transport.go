/*
Copyright 2022 The Flux authors

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

package transport

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// TransportPool is a progressive and non-blocking pool
// for http.Transport objects, optimised for Gargabe Collection
// and without a hard limit on number of objects created.
//
// Its main purpose is to enable for transport objects to be
// used across helm chart download requests and helm/pkg/getter
// instances by leveraging the getter.WithTransport(t) construct.
//
// The use of this pool improves the default behaviour of helm getter
// which creates a new connection per request, or per getter instance,
// resulting on unnecessary TCP connections with the target.
//
// http.Transport objects may contain sensitive material and also have
// settings that may impact the security of HTTP operations using
// them (i.e. InsecureSkipVerify). Therefore, ensure that they are
// used in a thread-safe way, and also by reseting TLS specific state
// after each use.
//
// Calling the Release(t) function will reset TLS specific state whilst
// also releasing the transport back to the pool to be reused.
//
// xref: https://github.com/helm/helm/pull/10568
// xref2: https://github.com/fluxcd/source-controller/issues/578
type TransportPool struct {
}

var pool = &sync.Pool{
	New: func() interface{} {
		return &http.Transport{
			DisableCompression: true,
			Proxy:              http.ProxyFromEnvironment,

			// Due to the non blocking nature of this approach,
			// at peak usage a higher number of transport objects
			// may be created. sync.Pool will ensure they are
			// gargage collected when/if needed.
			//
			// By setting a low value to IdleConnTimeout the connections
			// will be closed after that period of inactivity, allowing the
			// transport to be garbage collected.
			IdleConnTimeout: 60 * time.Second,

			// use safe defaults based off http.DefaultTransport
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	},
}

// NewOrIdle tries to return an existing transport that is not currently being used.
// If none is found, creates a new Transport instead.
//
// tlsConfig can optionally set the TLSClientConfig for the transport.
func NewOrIdle(tlsConfig *tls.Config) *http.Transport {
	t := pool.Get().(*http.Transport)
	t.TLSClientConfig = tlsConfig

	return t
}

// Release releases the transport back to the TransportPool after
// sanitising its sensitive fields.
func Release(transport *http.Transport) error {
	if transport == nil {
		return fmt.Errorf("cannot release nil transport")
	}

	transport.TLSClientConfig = nil

	pool.Put(transport)
	return nil
}
