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

func TestCheckNamespaceExists(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.CheckNamespaceExistsResponse{},
	)

	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.CheckNamespaceExistsRequest
		k8sError          error
		expectedResponse  *resourcesGRPCv1alpha1.CheckNamespaceExistsResponse
		expectedErrorCode grpccodes.Code
		existingObjects   []k8sruntime.Object
	}{
		{
			name: "returns true if namespace exists",
			request: &resourcesGRPCv1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&k8scorev1.Namespace{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "default",
					},
				},
			},
			expectedResponse: &resourcesGRPCv1alpha1.CheckNamespaceExistsResponse{
				Exists: true,
			},
		},
		{
			name: "returns false if namespace does not exist",
			request: &resourcesGRPCv1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			expectedResponse: &resourcesGRPCv1alpha1.CheckNamespaceExistsResponse{
				Exists: false,
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &resourcesGRPCv1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &resourcesGRPCv1alpha1.CheckNamespaceExistsRequest{
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
				fakeClient.CoreV1().(*k8stypedfakecorev1.FakeCoreV1).PrependReactor("get", "namespaces", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
					return true, &k8scorev1.Namespace{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CheckNamespaceExists(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}

func TestCreateNamespace(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.CreateNamespaceResponse{},
	)

	emptyResponse := &resourcesGRPCv1alpha1.CreateNamespaceResponse{}
	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.CreateNamespaceRequest
		k8sError          error
		expectedResponse  *resourcesGRPCv1alpha1.CreateNamespaceResponse
		expectedErrorCode grpccodes.Code
		existingObjects   []k8sruntime.Object
	}{
		{
			name: "creates a new namespace",
			request: &resourcesGRPCv1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			expectedResponse: emptyResponse,
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &resourcesGRPCv1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "returns already exists if k8s returns an already exists error",
			request: &resourcesGRPCv1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewAlreadyExists(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default"),
			expectedErrorCode: grpccodes.AlreadyExists,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &resourcesGRPCv1alpha1.CreateNamespaceRequest{
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
				fakeClient.CoreV1().(*k8stypedfakecorev1.FakeCoreV1).PrependReactor("create", "namespaces", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
					return true, &k8scorev1.Namespace{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CreateNamespace(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}

func TestGetNamespaceNames(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.GetNamespaceNamesResponse{},
	)

	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.GetNamespaceNamesRequest
		k8sError          error
		expectedResponse  *resourcesGRPCv1alpha1.GetNamespaceNamesResponse
		expectedErrorCode grpccodes.Code
		existingObjects   []k8sruntime.Object
	}{
		{
			name: "returns existing namespaces if user has RBAC",
			request: &resourcesGRPCv1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			existingObjects: []k8sruntime.Object{
				&k8scorev1.Namespace{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "default",
					},
				},
				&k8scorev1.Namespace{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "kubeapps",
					},
				},
			},
			expectedResponse: &resourcesGRPCv1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"default",
					"kubeapps",
				},
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &resourcesGRPCv1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			k8sError: k8serrors.NewForbidden(k8sschema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &resourcesGRPCv1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: grpccodes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := k8stypedclientfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*k8stypedfakecorev1.FakeCoreV1).PrependReactor("list", "namespaces", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
					return true, &k8scorev1.NamespaceList{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.GetNamespaceNames(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}
