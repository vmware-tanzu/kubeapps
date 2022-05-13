package main

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appRepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"testing"
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
			name:             "simple add package repository scenario",
			request:          addRepoReq4,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepo1,
			statusCode:       codes.OK,
		},
		{
			name:             "add package global repository",
			request:          addRepoReqGlobal,
			expectedResponse: addRepoExpectedGlobalResp,
			expectedRepo:     &addRepoGlobal,
			statusCode:       codes.OK,
		},
		{
			name:                  "package repository with tls cert authority",
			request:               addRepoReq6(ca),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepo2,
			expectedCreatedSecret: setSecretOwnerRef("bar", newTlsSecret("apprepo-bar", "foo", nil, nil, ca)),
			statusCode:            codes.OK,
		},
		{
			name:       "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:    addRepoReq7,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository with secret key reference",
			request:            addRepoReq7,
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       &addRepo3,
			statusCode:         codes.OK,
			existingSecret:     newTlsSecret("secret-1", "foo", nil, nil, ca),
			userManagedSecrets: true,
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            addRepoReq7,
			statusCode:         codes.NotFound,
			userManagedSecrets: true,
		},
		{
			name:       "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:    addRepoReq7,
			statusCode: codes.InvalidArgument,
		},
		{
			name:                  "package repository with basic auth and pass_credentials flag",
			request:               addRepoReq8,
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepo4,
			expectedCreatedSecret: setSecretOwnerRef("bar", newBasicAuthSecret("bar-", "foo", "baz", "zot")),
			statusCode:            codes.OK,
		},
		/*		{
				name:                  "package repository with TLS authentication",
				request:               addRepoReq9(pub, priv),
				expectedResponse:      addRepoExpectedResp,
				expectedRepo:          &addRepo2,
				expectedCreatedSecret: setSecretOwnerRef("bar", newTlsSecret("apprepo-bar", "foo", pub, priv, nil)),
				statusCode:            codes.OK,
			},*/
		{
			name:             "errors for package repository with bearer token",
			request:          addRepoReq10,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoBearerToken,
			//existingSecret:   newBasicAuthSecret("apprepo-bar", "foo", "baz", "zot"),
			statusCode: codes.OK,
		},
		{
			name:             "errors for package repository with bearer token  (user managed secrets)",
			request:          addRepoReq10,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoBearerToken,
			existingSecret:   newBasicAuthSecret("secret-1", "foo", "baz", "zot"),
			statusCode:       codes.OK,
		},
		{
			name:       "errors for package repository with custom auth token",
			request:    addRepoReq11,
			statusCode: codes.Unimplemented,
		},
		{
			name:               "package repository with basic auth and existing secret",
			request:            addRepoReq13,
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       &addRepo3,
			existingSecret:     newBasicAuthSecret("secret-1", "foo", "baz", "zot"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
		},
		{
			name:       "package repository with basic auth and existing secret (kubeapps managed secrets)",
			request:    addRepoReq13,
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:    addRepoReq14,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            addRepoReq14,
			statusCode:         codes.InvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          addRepoReq20,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepo5,
			statusCode:       codes.OK,
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
						// TODO(rcastelblanq) Remove
						//opt1 := cmpopts.IgnoreFields(appRepov1alpha1.AppRepositorySpec{}, "SecretRef")

						if got, want := &actualRepo, tc.expectedRepo; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}

						/*if tc.expectedCreatedSecret != nil {
							if !strings.HasPrefix(actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name) {
								t.Errorf("SecretRef [%s] was expected to start with [%s]",
									actualRepo.Spec.SecretRef.Name, tc.expectedRepo.Spec.SecretRef.Name)
							}
							opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
							// check expected secret has been created
							if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
								t.Fatal(err)
							} else if secret, err := typedClient.CoreV1().Secrets(nsname.Namespace).Get(ctx, actualRepo.Spec.SecretRef.Name, metav1.GetOptions{}); err != nil {
								t.Fatal(err)
							} else if got, want := secret, tc.expectedCreatedSecret; !cmp.Equal(want, got, opt2) {
								t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
							} else if !strings.HasPrefix(secret.Name, tc.expectedCreatedSecret.Name) {
								t.Errorf("Secret Name [%s] was expected to start with [%s]",
									secret.Name, tc.expectedCreatedSecret.Name)
							}
						} else if actualRepo.Spec.SecretRef != nil {
							t.Fatalf("Expected no secret, but found: [%q]", actualRepo.Spec.SecretRef.Name)
						} else if tc.expectedRepo.Spec.SecretRef != nil {
							t.Fatalf("Error: unexpected state")
						}*/
					}
				}
			}
		})
	}
}
