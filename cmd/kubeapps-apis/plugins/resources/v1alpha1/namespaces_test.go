/*
Copyright © 2021 VMware
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

package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
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

func TestCheckNamespaceExists(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		v1alpha1.CheckNamespaceExistsResponse{},
	)

	testCases := []struct {
		name              string
		request           *v1alpha1.CheckNamespaceExistsRequest
		k8sError          error
		expectedResponse  *v1alpha1.CheckNamespaceExistsResponse
		expectedErrorCode codes.Code
		existingObjects   []runtime.Object
	}{
		{
			name: "returns true if namespace exists",
			request: &v1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				},
			},
			expectedResponse: &v1alpha1.CheckNamespaceExistsResponse{
				Exists: true,
			},
		},
		{
			name: "returns false if namespace does not exist",
			request: &v1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			expectedResponse: &v1alpha1.CheckNamespaceExistsResponse{
				Exists: false,
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &v1alpha1.CheckNamespaceExistsRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &v1alpha1.CheckNamespaceExistsRequest{
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
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("get", "namespaces", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.Namespace{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CheckNamespaceExists(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
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
		v1alpha1.CreateNamespaceResponse{},
	)

	emptyResponse := &v1alpha1.CreateNamespaceResponse{}
	testCases := []struct {
		name              string
		request           *v1alpha1.CreateNamespaceRequest
		k8sError          error
		expectedResponse  *v1alpha1.CreateNamespaceResponse
		expectedErrorCode codes.Code
		existingObjects   []runtime.Object
	}{
		{
			name: "creates a new namespace",
			request: &v1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			expectedResponse: emptyResponse,
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &v1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "returns already exists if k8s returns an already exists error",
			request: &v1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			k8sError: k8serrors.NewAlreadyExists(schema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default"),
			expectedErrorCode: codes.AlreadyExists,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &v1alpha1.CreateNamespaceRequest{
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
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("create", "namespaces", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.Namespace{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.CreateNamespace(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
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
		v1alpha1.GetNamespaceNamesResponse{},
	)

	testCases := []struct {
		name              string
		request           *v1alpha1.GetNamespaceNamesRequest
		k8sError          error
		expectedResponse  *v1alpha1.GetNamespaceNamesResponse
		expectedErrorCode codes.Code
		existingObjects   []runtime.Object
	}{
		{
			name: "returns existing namespaces if user has RBAC",
			request: &v1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
				},
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "kubeapps",
					},
				},
			},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"default",
					"kubeapps",
				},
			},
		},
		{
			name: "returns permission denied if k8s returns a forbidden error",
			request: &v1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "returns an internal error if k8s returns an unexpected error",
			request: &v1alpha1.GetNamespaceNamesRequest{
				Cluster: "default",
			},
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := typfake.NewSimpleClientset(tc.existingObjects...)
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "namespaces", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.NamespaceList{}, tc.k8sError
				})
			}
			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
			}

			response, err := s.GetNamespaceNames(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}
