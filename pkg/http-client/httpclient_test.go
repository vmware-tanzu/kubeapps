// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	errors "errors"
	"io"
	"github.com/google/go-cmp/cmp"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const pemCert = `
-----BEGIN CERTIFICATE-----
MIIDETCCAfkCFEY03BjOJGqOuIMoBewOEDORMewfMA0GCSqGSIb3DQEBCwUAMEUx
CzAJBgNVBAYTAkRFMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRl
cm5ldCBXaWRnaXRzIFB0eSBMdGQwHhcNMTkwODE5MDQxNzU5WhcNMTkxMDA4MDQx
NzU5WjBFMQswCQYDVQQGEwJERTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UE
CgwYSW50ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAzA+X6HcScuHxqxCc5gs68weW8i72qMjvcWvBG064SvpTuNDK
ECEGvug6f8SFJjpA+hWjlqR5+UPMdfjMKPUEg1CI8JZm6lyNiB54iY50qvhv+qQg
1STdAWNTzvqUXUMGIImzeXFnErxlq8WwwLGwPNT4eFxF8V8fzIhR8sqQKFLOqvpS
7sCQwF5QOhziGfS+zParDLFsBoXQpWyDKqxb/yBSPwqijKkuW7kF4jGfPHD0Re3+
rspXiq8+jWSwSJIPSIbya8DQqrMwFeLCAxABidPnlrwS0UUion557ylaBK6Cv0UB
MojA4SMfjm5xRdzrOcoE8EcabxqoQD5rCIBgFQIDAQABMA0GCSqGSIb3DQEBCwUA
A4IBAQCped08LTojPejkPqmp1edZa9rWWrCMviY5cvqb6t3P3erse+jVcBi9NOYz
8ewtDbR0JWYvSW6p3+/nwyDG4oVfG5TiooAZHYHmgg4x9+5h90xsnmgLhIsyopPc
Rltj86tRCl1YiuRpkWrOfRBGdYfkGEG4ihJzLHWRMCd1SmMwnmLliBctD7IeqBKw
UKt8wcroO8/sj/Xd1/LCtNZ79/FdQFa4l3HnzhOJOrlQyh4gyK05EKdg6vv3un17
l6NEPfiXd7dZvsWi9uY/PGBhu9EY/bdvuIOWDNNK262azk1A56HINpMrYBUcfti1
YrvYQHgOtHsqCB/hFHWfZp1lg2Sx
-----END CERTIFICATE-----
`

func TestNew(t *testing.T) {
	t.Run("test default client has expected defaults", func(t *testing.T) {
		client := New()
		if client.Timeout == 0 {
			t.Fatal("expected default timeout to be set")
		}
		if client.Transport == nil {
			t.Fatal("expected a Transport structure to be set")
		}

		transport, ok := client.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected a Transport structure to be set, but got type: %T", client.Transport)
		}
		if transport.Proxy == nil {
			t.Fatal("expected a Proxy to be set")
		}
	})
}

func TestSetClientProxy(t *testing.T) {
	t.Run("test SetClientProxy", func(t *testing.T) {
		client := New()
		transport := client.Transport.(*http.Transport)
		transport.Proxy = nil

		if transport.Proxy != nil {
			t.Fatal("expected proxy to have been nilled")
		}

		testerror := errors.New("Test Proxy Error")
		proxyFunc := func(r *http.Request) (*url.URL, error) { return nil, testerror }
		err := SetClientProxy(client, proxyFunc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if transport.Proxy == nil {
			t.Fatal("expected proxy to have been set")
		}
		if _, err := transport.Proxy(nil); err == nil || err != testerror {
			t.Fatalf("invocation of proxy function did not result in expected error: {%+v}", err)
		}
	})
}

func TestSetClientTls(t *testing.T) {
	t.Run("test SetClientTls", func(t *testing.T) {
		client := New()
		transport := client.Transport.(*http.Transport)

		if transport.TLSClientConfig != nil {
			t.Fatal("invalid initial state, TLS is not nil")
		}

		systemCertPool, err := x509.SystemCertPool()
		if err != nil {
			t.Fatalf("%+v", err)
		}

		tlsConf := &tls.Config{
			RootCAs:            systemCertPool,
			InsecureSkipVerify: true,
		}
		err = SetClientTLS(client, tlsConf)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if transport.TLSClientConfig == nil {
			t.Fatal("expected TLS config to have been set but it is nil")
		}
		if transport.TLSClientConfig.RootCAs != systemCertPool {
			t.Fatal("expected root CA to have been set to system cert pool")
		}
		if transport.TLSClientConfig.InsecureSkipVerify != true {
			t.Fatal("expected InsecureSkipVerify to have been set to true")
		}

		expectedPayload := []byte("Bob's your uncle")
		ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			_, err := w.Write(expectedPayload)
			if err != nil {
				t.Fatalf("%+v", err)
			}
		}))
		ts.TLS = tlsConf
		ts.StartTLS()
		defer ts.Close()

		resp, err := client.Get(ts.URL)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("expected OK, got: %d", resp.StatusCode)
		}

		payload, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		} else if got, want := payload, expectedPayload; !cmp.Equal(got, want) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	})
}

func TestGetCertPool(t *testing.T) {
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testCases := []struct {
		name                 string
		cert                 []byte
		expectError          bool
		expectedSubjectCount int
	}{
		{
			name: "invocation with nil cert",
			cert: nil,
			//nolint:staticcheck
			expectedSubjectCount: len(systemCertPool.Subjects()),
		},
		{
			name: "invocation with empty cert",
			cert: []byte{},
			//nolint:staticcheck
			expectedSubjectCount: len(systemCertPool.Subjects()),
		},
		{
			name: "invocation with valid cert",
			cert: []byte(pemCert),
			//nolint:staticcheck
			expectedSubjectCount: len(systemCertPool.Subjects()) + 1,
		},
		{
			name:        "invocation with invalid cert",
			cert:        []byte("not valid cert"),
			expectError: true,
			//nolint:staticcheck
			expectedSubjectCount: len(systemCertPool.Subjects()) + 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			caCertPool, err := GetCertPool(tc.cert)

			// no creation case
			if tc.expectError {
				if err == nil {
					t.Fatalf("pool creation was expcted to fail")
				}
				return
			}

			// creation case
			if err != nil {
				t.Fatalf("error creating the cert pool: {%+v}", err)
			}

			//nolint:staticcheck
			if got, want := len(caCertPool.Subjects()), tc.expectedSubjectCount; got != want {
				t.Fatalf("cert pool subjects is not as expected, got {%d} instead of {%d}", got, want)
			}
		})
	}
}

type testClient struct {
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Header: req.Header,
	}, nil
}
func TestClientWithDefaults(t *testing.T) {
	initialHdrName := "TestHeader"
	initialHdrValue := "TestHeaderValue"
	initialHeaders := http.Header{initialHdrName: {initialHdrValue}}
	extraHdrName := "TestNewHeader"
	extraHdrValue := "TestNewHeaderValue"

	testCases := []struct {
		name            string
		headers         http.Header
		initialHeaders  http.Header
		expectedHeaders http.Header
	}{
		{
			name:            "nil headers",
			headers:         nil,
			initialHeaders:  initialHeaders,
			expectedHeaders: initialHeaders,
		},
		{
			name:            "empty headers",
			headers:         http.Header{},
			initialHeaders:  initialHeaders,
			expectedHeaders: initialHeaders,
		},
		{
			name:            "new header",
			headers:         http.Header{extraHdrName: {extraHdrValue}},
			initialHeaders:  initialHeaders,
			expectedHeaders: http.Header{initialHdrName: {initialHdrValue}, extraHdrName: {extraHdrValue}},
		},
		{
			name:            "no override",
			headers:         http.Header{initialHdrName: {extraHdrValue}},
			initialHeaders:  initialHeaders,
			expectedHeaders: initialHeaders,
		},
		{
			name:            "new and no override",
			headers:         http.Header{initialHdrName: {extraHdrValue}, extraHdrName: {extraHdrValue}},
			initialHeaders:  initialHeaders,
			expectedHeaders: http.Header{initialHdrName: {initialHdrValue}, extraHdrName: {extraHdrValue}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// test client to capture headers
			testclient := &testClient{}

			// init invocation
			client := &ClientWithDefaults{
				Client:         testclient,
				DefaultHeaders: tc.headers,
			}

			requestHeaders := http.Header{}
			for k, v := range tc.initialHeaders {
				requestHeaders[k] = v
			}
			request := &http.Request{
				Header: requestHeaders,
			}

			// invocation
			response, err := client.Do(request)
			if err != nil || response == nil || response.Header == nil {
				t.Fatal("unexpected error during invocation")
			}

			// check
			if len(response.Header) != len(tc.expectedHeaders) {
				t.Fatalf("response header length differs from expected, got {%+v} when expecting {%+v}", response.Header, tc.expectedHeaders)
			}
			for k := range tc.expectedHeaders {
				got := response.Header.Get(k)
				expected := tc.expectedHeaders.Get(k)
				if got != expected {
					t.Fatalf("response header differs from expected, got {%s} when expecting {%s}", got, expected)
				}
			}
		})
	}
}
