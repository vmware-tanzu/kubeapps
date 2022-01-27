// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"testing"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	resourcesGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
	k8stypedfakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestCreateSecret(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.CreateSecretResponse{},
	)

	emptyResponse := &resourcesGRPCv1alpha1.CreateSecretResponse{}
	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.CreateSecretRequest
		k8sError          error
		expectedResponse  *resourcesGRPCv1alpha1.CreateSecretResponse
		expectedErrorCode grpccodes.Code
		existingObjects   []k8sruntime.Object
	}{
		{
			name: "creates an opaque secret by default",
			request: &resourcesGRPCv1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				StringData: map[string]string{
					"foo": "bar",
				},
			},
			expectedResponse: emptyResponse,
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &resourcesGRPCv1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default", errors.New("Bang")),
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "returns already exists if k8s returns an already exists error",
			request: &resourcesGRPCv1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewAlreadyExists(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default"),
			expectedErrorCode: grpccodes.AlreadyExists,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &resourcesGRPCv1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: grpccodes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := k8stypedclientfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*k8stypedfakecorev1.FakeCoreV1).PrependReactor("create", "secrets", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
					return true, &k8scorev1.Secret{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CreateSecret(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}

			if tc.expectedErrorCode != grpccodes.OK {
				return
			}
			secret, err := fakeClient.CoreV1().Secrets(tc.request.GetContext().GetNamespace()).Get(context.Background(), tc.request.GetName(), k8smetav1.GetOptions{})
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if got, want := secret.StringData, tc.request.GetStringData(); !cmp.Equal(got, want) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetSecretNames(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.GetSecretNamesResponse{},
	)

	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.GetSecretNamesRequest
		k8sError          error
		expectedResponse  *resourcesGRPCv1alpha1.GetSecretNamesResponse
		expectedErrorCode grpccodes.Code
		existingObjects   []k8sruntime.Object
	}{
		{
			name: "returns existing namespaces from the context namespace only if user has RBAC",
			request: &resourcesGRPCv1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "secret-1",
						Namespace: "default",
					},
					Type: k8scorev1.SecretTypeOpaque,
					StringData: map[string]string{
						"ignored": "we don't use it",
					},
				},
				&k8scorev1.Secret{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "secret-2",
						Namespace: "default",
					},
					Type: k8scorev1.SecretTypeDockerConfigJson,
				},
				&k8scorev1.Secret{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "secret-other-namespace",
						Namespace: "other-namespace",
					},
					Type: k8scorev1.SecretTypeDockerConfigJson,
				},
			},
			expectedResponse: &resourcesGRPCv1alpha1.GetSecretNamesResponse{
				SecretNames: map[string]resourcesGRPCv1alpha1.SecretType{
					"secret-1": resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED,
					"secret-2": resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG_JSON,
				},
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &resourcesGRPCv1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default", errors.New("Bang")),
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &resourcesGRPCv1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: grpccodes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := k8stypedclientfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*k8stypedfakecorev1.FakeCoreV1).PrependReactor("list", "secrets", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
					return true, &k8scorev1.SecretList{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.GetSecretNames(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}
