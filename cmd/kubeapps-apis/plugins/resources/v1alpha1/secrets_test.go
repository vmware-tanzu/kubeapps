// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
	fakecorev1 "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	clientGoTesting "k8s.io/client-go/testing"
)

func TestCreateSecret(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		v1alpha1.CreateSecretResponse{},
	)

	emptyResponse := &v1alpha1.CreateSecretResponse{}
	testCases := []struct {
		name              string
		request           *v1alpha1.CreateSecretRequest
		k8sError          error
		expectedResponse  *v1alpha1.CreateSecretResponse
		expectedErrorCode codes.Code
		existingObjects   []runtime.Object
	}{
		{
			name: "creates an opaque secret by default",
			request: &v1alpha1.CreateSecretRequest{
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
			request: &v1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "returns already exists if k8s returns an already exists error",
			request: &v1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewAlreadyExists(schema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default"),
			expectedErrorCode: codes.AlreadyExists,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &v1alpha1.CreateSecretRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := typfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("create", "secrets", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.Secret{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CreateSecret(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}

			if tc.expectedErrorCode != codes.OK {
				return
			}
			secret, err := fakeClient.CoreV1().Secrets(tc.request.GetContext().GetNamespace()).Get(context.Background(), tc.request.GetName(), metav1.GetOptions{})
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
		v1alpha1.GetSecretNamesResponse{},
	)

	testCases := []struct {
		name              string
		request           *v1alpha1.GetSecretNamesRequest
		k8sError          error
		expectedResponse  *v1alpha1.GetSecretNamesResponse
		expectedErrorCode codes.Code
		existingObjects   []runtime.Object
	}{
		{
			name: "returns existing namespaces from the context namespace only if user has RBAC",
			request: &v1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []runtime.Object{
				&core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-1",
						Namespace: "default",
					},
					Type: core.SecretTypeOpaque,
					StringData: map[string]string{
						"ignored": "we don't use it",
					},
				},
				&core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-2",
						Namespace: "default",
					},
					Type: core.SecretTypeDockerConfigJson,
				},
				&core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "secret-other-namespace",
						Namespace: "other-namespace",
					},
					Type: core.SecretTypeDockerConfigJson,
				},
			},
			expectedResponse: &v1alpha1.GetSecretNamesResponse{
				SecretNames: map[string]v1alpha1.SecretType{
					"secret-1": v1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED,
					"secret-2": v1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG_JSON,
				},
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &v1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "secrets",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &v1alpha1.GetSecretNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := typfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "secrets", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.SecretList{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.GetSecretNames(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}
