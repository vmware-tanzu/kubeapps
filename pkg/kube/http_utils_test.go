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

import (
	"crypto/x509"
	"net/http"
	"testing"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
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
	)

	testCases := []struct {
		name             string
		customCAData     string
		appRepoSpec      v1alpha1.AppRepositorySpec
		errorExpected    bool
		numCertsExpected int
		expectedHeaders  http.Header
	}{
		{
			name:             "default cert pool without auth",
			numCertsExpected: len(systemCertPool.Subjects()),
		},
		{
			name: "custom CA added when passed an AppRepository CRD",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					CustomCA: &v1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{customCASecretName},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			customCAData:     pemCert,
			numCertsExpected: len(systemCertPool.Subjects()) + 1,
		},
		{
			name: "errors if custom CA key cannot be found in secret",
			appRepoSpec: v1alpha1.AppRepositorySpec{
				Auth: v1alpha1.AppRepositoryAuth{
					CustomCA: &v1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: corev1.SecretKeySelector{
							corev1.LocalObjectReference{customCASecretName},
							"some-other-secret-key",
							nil,
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
							corev1.LocalObjectReference{customCASecretName},
							"custom-secret-key",
							nil,
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
							corev1.LocalObjectReference{authHeaderSecretName},
							"custom-secret-key",
							nil,
						},
					},
				},
			},
			numCertsExpected: len(systemCertPool.Subjects()),
			expectedHeaders:  http.Header{"Authorization": []string{authHeaderSecretData}},
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

			clientWithDefaultHeaders, ok := httpClient.(*clientWithDefaultHeaders)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			client, ok := clientWithDefaultHeaders.client.(*http.Client)
			if !ok {
				t.Fatalf("unable to assert expected type")
			}
			transport, ok := client.Transport.(*http.Transport)
			certPool := transport.TLSClientConfig.RootCAs

			if got, want := len(certPool.Subjects()), tc.numCertsExpected; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// If the Auth header was set, secrets should be returned
			if tc.appRepoSpec.Auth.Header != nil {
				_, ok := clientWithDefaultHeaders.defaultHeaders["Authorization"]
				if !ok {
					t.Fatalf("expected Authorization header but found none")
				}
				if got, want := clientWithDefaultHeaders.defaultHeaders.Get("Authorization"), authHeaderSecretData; got != want {
					t.Errorf("got: %q, want: %q", got, want)
				}
			}
		})
	}
}
