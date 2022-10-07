// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	errors "errors"
	"github.com/google/go-cmp/cmp"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"golang.org/x/net/http/httpproxy"
	"io"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
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

func TestInitNetClient(t *testing.T) {
	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		t.Fatalf("%+v", err)
	}

	const (
		authHeaderSecretName = "auth-header-secret-name"
		authHeaderSecretData = "really-secret-stuff"
		customCASecretName   = "custom-ca-secret-name"
		appRepoName          = "custom-repo"
		requestURL           = "https://request.example.com/foo/bar"
		proxyURL             = "https://proxy.example.com"
	)

	testCases := []struct {
		name             string
		customCAData     string
		appRepoSpec      v1alpha1.AppRepositorySpec
		errorExpected    bool
		numCertsExpected int
		expectedHeaders  http.Header
		expectProxied    bool
		expectSkipTLS    bool
	}{
		{
			name: "default cert pool without auth",
			//nolint:staticcheck
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "custom CA added when passed an AppRepository CRD",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					CustomCA: &v1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: customCASecretName},
							Key:                  "custom-secret-key",
						},
					},
				},
			},
			customCAData: pemCert,
			//nolint:staticcheck
			numCertsExpected: len(systemCertPool.Subjects()) + 1,
		},
		{
			name: "errors if custom CA key cannot be found in secret",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					CustomCA: &v1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: customCASecretName},
							Key:                  "some-other-secret-key",
						},
					},
				},
			},
			customCAData:  pemCert,
			errorExpected: true,
		},
		{
			name: "errors if custom CA cannot be parsed",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					CustomCA: &v1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: customCASecretName},
							Key:                  "custom-secret-key",
						},
					},
				},
			},
			customCAData:  "not a valid cert",
			errorExpected: true,
		},
		{
			name: "authorization header added when passed an AppRepository CRD",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					Header: &v1alpha1.AppRepositoryAuthHeader{
						SecretKeyRef: corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: authHeaderSecretName},
							Key:                  "custom-secret-key",
						},
					},
				},
			},
			//nolint:staticcheck
			numCertsExpected: len(systemCertPool.Subjects()),
			expectedHeaders:  http.Header{"Authorization": []string{authHeaderSecretData}},
		},
		{
			name: "http proxy added when passed an AppRepository CRD with an http_proxy env var",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				SyncJobPodTemplate: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Env: []corev1.EnvVar{
									{
										Name:  "https_proxy",
										Value: proxyURL,
									},
								},
							},
						},
					},
				},
			},
			expectProxied: true,
			//nolint:staticcheck
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "skip tls config",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				TLSInsecureSkipVerify: true,
			},
			expectSkipTLS: true,
			//nolint:staticcheck
			numCertsExpected: len(systemCertPool.Subjects()),
		},
	}

	for _, tc := range testCases {
		// The fake k8s client will contain secret for the CA and header respectively.
		caCertSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      customCASecretName,
				Namespace: metav1.NamespaceSystem,
			},
			StringData: map[string]string{
				"custom-secret-key": tc.customCAData,
			},
		}
		authSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      authHeaderSecretName,
				Namespace: metav1.NamespaceSystem,
			},
			Data: map[string][]byte{
				"custom-secret-key": []byte(authHeaderSecretData),
			},
		}

		appRepo := &v1alpha1.AppRepository{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: metav1.NamespaceSystem,
			},
			Spec: tc.appRepoSpec,
		}

		t.Run(tc.name, func(t *testing.T) {
			var testCASecret *corev1.Secret
			if tc.appRepoSpec.Auth.CustomCA != nil {
				testCASecret = caCertSecret
			}
			var testAuthSecret *corev1.Secret
			if tc.appRepoSpec.Auth.Header != nil {
				testAuthSecret = authSecret
			}
			httpClient, err := InitNetClient(appRepo, testCASecret, testAuthSecret, nil)
			if err != nil {
				if tc.errorExpected {
					return
				}
				t.Fatalf("%+v", err)
			} else {
				if tc.errorExpected {
					t.Fatalf("got: nil, want: error")
				}
			}

			clientWithDefaultHeaders, ok := httpClient.(*ClientWithDefaults)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			client, ok := clientWithDefaultHeaders.Client.(*http.Client)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			transport, ok := client.Transport.(*http.Transport)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			if tc.expectSkipTLS && !transport.TLSClientConfig.InsecureSkipVerify {
				t.Error("expecting to skip TLS verification")
			}
			certPool := transport.TLSClientConfig.RootCAs

			//nolint:staticcheck
			if got, want := len(certPool.Subjects()), tc.numCertsExpected; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// If the Auth header was set, secrets should be returned
			_, ok = clientWithDefaultHeaders.DefaultHeaders["Authorization"]
			if tc.expectedHeaders != nil {
				if !ok {
					t.Fatalf("expected Authorization header but found none")
				}
				if got, want := clientWithDefaultHeaders.DefaultHeaders.Get("Authorization"), authHeaderSecretData; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			} else {
				if ok {
					t.Errorf("Authorization header present when non included in app repo")
				}
			}

			// Verify that a URL is proxied or not, depending on the app repo configuration.
			u, err := url.Parse(requestURL)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			reqURL, err := transport.Proxy(&http.Request{URL: u})
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if tc.expectProxied {
				if reqURL == nil {
					t.Fatalf("Expecting the URL %s to be proxied", requestURL)
				}
				if got, want := reqURL.String(), proxyURL; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			} else {
				// The proxy function returns nil (with a nil error) if the
				// request should not be proxied
				if got := reqURL; got != nil {
					t.Errorf("got: %q, want: nil", got)
				}
			}
		})
	}
}

func TestGetProxyConfig(t *testing.T) {
	proxyVars := []string{"http_proxy", "https_proxy", "no_proxy", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"}
	testCases := []struct {
		name             string
		appRepoEnvVars   []corev1.EnvVar
		containerEnvVars map[string]string
		expectedConfig   *httpproxy.Config
	}{
		{
			name: "configures when http_proxy specified",
			appRepoEnvVars: []corev1.EnvVar{
				{
					Name:  "http_proxy",
					Value: "http://proxied.example.com:8888",
				},
			},
			expectedConfig: &httpproxy.Config{
				HTTPProxy: "http://proxied.example.com:8888",
			},
		},
		{
			name: "configures when https_proxy specified",
			appRepoEnvVars: []corev1.EnvVar{
				{
					Name:  "https_proxy",
					Value: "https://proxied.example.com:8888",
				},
			},
			expectedConfig: &httpproxy.Config{
				HTTPSProxy: "https://proxied.example.com:8888",
			},
		},
		{
			name: "configures all three when specified",
			appRepoEnvVars: []corev1.EnvVar{
				{
					Name:  "http_proxy",
					Value: "http://proxied.example.com:8888",
				},
				{
					Name:  "https_proxy",
					Value: "https://proxied.example.com:8888",
				},
				{
					Name:  "no_proxy",
					Value: "http://some.example.com https://other.example.com",
				},
			},
			expectedConfig: &httpproxy.Config{
				HTTPSProxy: "https://proxied.example.com:8888",
				HTTPProxy:  "http://proxied.example.com:8888",
				NoProxy:    "http://some.example.com https://other.example.com",
			},
		},
		{
			name:           "returns a nil config when none specified in app repo or container",
			expectedConfig: &httpproxy.Config{},
		},
		{
			name: "defaults to the container environment proxy vars when set",
			containerEnvVars: map[string]string{
				"http_proxy": "http://container.example.com:9999",
			},
			expectedConfig: &httpproxy.Config{
				HTTPProxy: "http://container.example.com:9999",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			// Set the env for the test ensuring to restore after.
			originalValues := map[string]string{}
			for _, key := range proxyVars {
				originalVal, ok := os.LookupEnv(key)
				if ok {
					originalValues[key] = originalVal
					os.Unsetenv(key)
				}

				value, ok := tc.containerEnvVars[key]
				if ok {
					os.Setenv(key, value)
				}
			}
			defer func() {
				for key, val := range originalValues {
					os.Setenv(key, val)
				}
			}()

			appRepo := &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: metav1.NamespaceSystem,
				},
				Spec: v1alpha1.AppRepositorySpec{
					SyncJobPodTemplate: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Env: tc.appRepoEnvVars,
								},
							},
						},
					},
				},
			}
			if got, want := getProxyConfig(appRepo), tc.expectedConfig; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
