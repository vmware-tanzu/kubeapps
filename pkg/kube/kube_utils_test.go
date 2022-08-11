// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"crypto/x509"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"golang.org/x/net/http/httpproxy"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

			clientWithDefaultHeaders, ok := httpClient.(*httpclient.ClientWithDefaults)
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

func Test_getDataFromRegistrySecret(t *testing.T) {
	testCases := []struct {
		name     string
		secret   *corev1.Secret
		expected string
	}{
		{
			name: "retrieves username and password from a dockerconfigjson",
			secret: &corev1.Secret{
				Data: map[string][]byte{
					".dockerconfigjson": []byte(`{"auths":{"foo":{"username":"foo","password":"bar"}}}`),
				},
			},
			// Basic: base64(foo:bar)
			expected: "Basic Zm9vOmJhcg==",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			auth, err := getDataFromRegistrySecret(".dockerconfigjson", tc.secret)
			if err != nil {
				t.Fatalf("Unexpected error %v", err)
			}
			if auth != tc.expected {
				t.Errorf("Expecting %s, got %s", tc.expected, auth)
			}
		})
	}
}
