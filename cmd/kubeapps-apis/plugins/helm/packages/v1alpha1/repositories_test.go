// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appRepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

func TestAddPackageRepository(t *testing.T) {
	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	//ca, pub, priv := getCertsForTesting(t)
	ca, _, _ := getCertsForTesting(t)

	testCases := []struct {
		name                  string
		request               *corev1.AddPackageRepositoryRequest
		expectedResponse      *corev1.AddPackageRepositoryResponse
		expectedRepo          *appRepov1alpha1.AppRepository
		statusCode            codes.Code
		existingSecret        *apiv1.Secret
		expectedCreatedSecret *apiv1.Secret
		userManagedSecrets    bool
		repoClientGetter      newRepoClient
	}{
		{
			name:       "returns error if no namespace is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if no name is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Namespace: "foo"}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if invalid cluster",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Cluster: "wrongCluster", Namespace: "foo"}},
			statusCode: codes.Unimplemented,
		},
		{
			name:       "returns error if wrong repository type",
			request:    addRepoReqWrongType,
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if no url",
			request:    addRepoReqNoUrl,
			statusCode: codes.InvalidArgument,
		},
		{
			name:             "simple add package repository scenario (HELM)",
			request:          addRepoReqSimple("helm"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleHelm,
			statusCode:       codes.OK,
		},
		{
			name:             "simple add package repository scenario (OCI)",
			request:          addRepoReqSimple("oci"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleOci,
			statusCode:       codes.OK,
		},
		{
			name:             "add package global repository",
			request:          addRepoReqGlobal,
			expectedResponse: addRepoExpectedGlobalResp,
			expectedRepo:     &addRepoGlobal,
			statusCode:       codes.OK,
		},
		// CUSTOM CA AUTH
		{
			name:                  "package repository with tls cert authority",
			request:               addRepoReqTLSCA(ca),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepoWithTLSCA,
			expectedCreatedSecret: setSecretOwnerRef("bar", newTlsSecret("apprepo-bar", "foo", nil, nil, ca)),
			statusCode:            codes.OK,
		},
		{
			name:       "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository with secret key reference",
			request:            addRepoReqTLSSecretRef,
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       &addRepoTLSSecret,
			statusCode:         codes.OK,
			existingSecret:     newTlsSecret("secret-1", "foo", nil, nil, ca),
			userManagedSecrets: true,
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            addRepoReqTLSSecretRef,
			statusCode:         codes.NotFound,
			userManagedSecrets: true,
		},
		{
			name:       "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		// BASIC AUTH
		{
			name:                  "package repository with basic auth and pass_credentials flag",
			request:               addRepoReqBasicAuth("baz", "zot"),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepoAuthHeaderPassCredentials,
			expectedCreatedSecret: setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", "foo", "baz", "zot")),
			statusCode:            codes.OK,
		},
		{
			name:       "package repository with wrong basic auth",
			request:    addRepoReqWrongBasicAuth,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "fails for package repository passing basic auth (user managed secrets)",
			request:            addRepoReqBasicAuth("kermit", "frog"),
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
		},
		{
			name:               "package repository basic auth with existing secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-basic"),
			existingSecret:     newBasicAuthSecret("secret-basic", "foo", "baz-user", "zot-pwd"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
		},
		{
			name:       "package repository basic auth with existing secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"), //addRepoReq13,
			statusCode: codes.InvalidArgument,
		},
		// BEARER TOKEN
		{
			name:                  "package repository with bearer token",
			request:               addRepoReqBearerToken("the-token"),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("apprepo-bar"),
			statusCode:            codes.OK,
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "Bearer the-token")),
		},
		{
			name:       "package repository with no bearer token",
			request:    addRepoReqBearerToken(""),
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository bearer token with secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "secret-bearer"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-bearer"),
			userManagedSecrets: true,
			existingSecret:     newAuthTokenSecret("secret-bearer", "foo", "Bearer the-token"),
			statusCode:         codes.OK,
		},
		{
			name:       "package repository bearer token with secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "secret-bearer"),
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository bearer token (user managed secrets)",
			request:            addRepoReqBearerToken("the-token"),
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
		},
		// CUSTOM AUTH
		{
			name:                  "package repository with custom auth",
			request:               addRepoReqCustomAuth,
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("apprepo-bar"),
			statusCode:            codes.OK,
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "foobarzot")),
		},
		{
			name:               "package repository custom auth with existing secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM, "secret-custom"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-custom"),
			existingSecret:     newBasicAuthSecret("secret-custom", "foo", "baz", "zot"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
		},
		// DOCKER AUTH
		{
			name: "package repository with Docker auth",
			request: addRepoReqDockerAuth(&corev1.DockerCredentials{
				Server:   "https://docker-server",
				Username: "the-user",
				Password: "the-password",
				Email:    "foo@bar.com",
			}),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     addRepoAuthDocker("apprepo-bar"),
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newAuthDockerSecret("apprepo-bar",
					"foo",
					"{\"auths\":{\"https://docker-server\":{\"username\":\"the-user\",\"password\":\"the-password\",\"email\":\"foo@bar.com\",\"auth\":\"dGhlLXVzZXI6dGhlLXBhc3N3b3Jk\"}}}")),
			statusCode: codes.OK,
		},
		{
			name:               "package repository with Docker auth (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON, "secret-docker"),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			existingSecret: newAuthDockerSecret("secret-docker", "foo",
				"{\"auths\":{\"https://docker-server\":{\"username\":\"the-user\",\"password\":\"the-password\",\"email\":\"foo@bar.com\",\"auth\":\"dGhlLXVzZXI6dGhlLXBhc3N3b3Jk\"}}}"),
			expectedRepo: addRepoAuthDocker("secret-docker"),
			statusCode:   codes.OK,
		},
		// Others
		{
			name:       "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:    addRepoReqTLSDifferentSecretAuth,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            addRepoReqTLSDifferentSecretAuth,
			statusCode:         codes.InvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          addRepoReqOnlyPassCredentials,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoOnlyPassCredentials,
			statusCode:       codes.OK,
		},
		{
			name:               "package repository with reference to malformed secret",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"),
			existingSecret:     newTlsSecret("secret-basic", "foo", nil, nil, nil), // Creates empty secret
			userManagedSecrets: true,
			statusCode:         codes.Internal,
		},
		// Custom values
		{
			name:             "package repository with custom values",
			request:          addRepoReqCustomValues,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailsHelm,
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with invalid custom values",
			request:          addRepoReqWrongCustomValues,
			expectedResponse: addRepoExpectedResp,
			statusCode:       codes.InvalidArgument,
		},
		{
			name:             "package repository with validation success (Helm)",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailsHelm,
			repoClientGetter: newRepoHttpClient(map[string]*http.Response{"https://example.com/index.yaml": {StatusCode: 200}}),
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with validation success (OCI)",
			request:          addRepoReqCustomValuesOCIValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailsOci,
			repoClientGetter: newRepoHttpClient(map[string]*http.Response{
				"https://example.com/v2/repo1/tags/list?n=1":  httpResponse(200, "{ \"name\":\"repo1\", \"tags\":[\"tag1\"] }"),
				"https://example.com/v2/repo1/manifests/tag1": httpResponse(200, "{ \"config\":{ \"mediaType\":\"application/vnd.cncf.helm.config\" } }"),
			}),
			statusCode: codes.OK,
		},
		{
			name:             "package repository with validation failing",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			repoClientGetter: newRepoHttpClient(
				map[string]*http.Response{
					"https://example.com/index.yaml": httpResponse(404, "It failed because of X and Y"),
				}),
			statusCode: codes.FailedPrecondition,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var secrets []runtime.Object
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}
			s := newServerWithSecrets(t, secrets)
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets
			if tc.repoClientGetter != nil {
				s.repoClientGetter = tc.repoClientGetter
			}

			nsname := types.NamespacedName{Namespace: tc.request.Context.Namespace, Name: tc.request.Name}
			ctx := context.Background()
			response, err := s.AddPackageRepository(ctx, tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.AddPackageRepositoryResponse{},
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
					)
					if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// purposefully not calling mock.ExpectationsWereMet() here because
			// AddPackageRepository will trigger an ADD event that will be processed
			// asynchronously, so it may or may not have enough time to get to the
			// point where the cache worker does a GET

			// We don't need to check anything else for non-OK codes.
			if tc.statusCode != codes.OK {
				return
			}

			// check expected HelmRelease CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				var actualRepo appRepov1alpha1.AppRepository
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				} else {
					if tc.userManagedSecrets {
						if tc.expectedCreatedSecret != nil {
							t.Fatalf("Error: unexpected state")
						}
						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}
					} else {
						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}

						if tc.expectedCreatedSecret != nil {
							opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
							if actualRepo.Spec.Auth.Header == nil && actualRepo.Spec.Auth.CustomCA == nil {
								t.Errorf("Error: Repository secrets were expected but auth header and CA are empty")
							} else if actualRepo.Spec.Auth.Header != nil {
								if !strings.HasPrefix(actualRepo.Spec.Auth.Header.SecretKeyRef.Name, tc.expectedRepo.Spec.Auth.Header.SecretKeyRef.Name) {
									t.Errorf("Auth header SecretKeyRef [%s] was expected to start with [%s]",
										actualRepo.Spec.Auth.Header.SecretKeyRef.Name, tc.expectedRepo.Spec.Auth.Header.SecretKeyRef.Name)
								}
								// check expected secret has been created
								if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
									t.Fatal(err)
								} else if secret, err := typedClient.CoreV1().Secrets(nsname.Namespace).Get(ctx, actualRepo.Spec.Auth.Header.SecretKeyRef.Name, metav1.GetOptions{}); err != nil {
									t.Fatal(err)
								} else if got, want := secret, tc.expectedCreatedSecret; !cmp.Equal(want, got, opt2) {
									t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
								} else if !strings.HasPrefix(secret.Name, tc.expectedCreatedSecret.Name) {
									t.Errorf("Secret Name [%s] was expected to start with [%s]",
										secret.Name, tc.expectedCreatedSecret.Name)
								}
							} else {
								if !strings.HasPrefix(actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, tc.expectedRepo.Spec.Auth.CustomCA.SecretKeyRef.Name) {
									t.Errorf("CustomCA SecretKeyRef [%s] was expected to start with [%s]",
										actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, tc.expectedRepo.Spec.Auth.CustomCA.SecretKeyRef.Name)
								}
								// check expected secret has been created
								if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
									t.Fatal(err)
								} else if secret, err := typedClient.CoreV1().Secrets(nsname.Namespace).Get(ctx, actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, metav1.GetOptions{}); err != nil {
									t.Fatal(err)
								} else if got, want := secret, tc.expectedCreatedSecret; !cmp.Equal(want, got, opt2) {
									t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
								} else if !strings.HasPrefix(secret.Name, tc.expectedCreatedSecret.Name) {
									t.Errorf("Secret Name [%s] was expected to start with [%s]",
										secret.Name, tc.expectedCreatedSecret.Name)
								}
							}
						} else if actualRepo.Spec.Auth.Header != nil {
							t.Fatalf("Expected no secret, but found Header: [%v]", actualRepo.Spec.Auth.Header.SecretKeyRef)
						} else if actualRepo.Spec.Auth.CustomCA != nil {
							t.Fatalf("Expected no secret, but found CustomCA: [%v]", actualRepo.Spec.Auth.CustomCA.SecretKeyRef)
						} else if tc.expectedRepo.Spec.Auth.Header != nil {
							t.Fatalf("Error: unexpected state")
						}
					}
				}
			}
		})
	}
}
