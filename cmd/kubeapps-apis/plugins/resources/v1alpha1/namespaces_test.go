// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/resources/v1alpha1/common"
	"google.golang.org/grpc/metadata"

	authorizationv1 "k8s.io/api/authorization/v1"

	"net/http"
	"strings"
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
		validator         func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error)
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
			validator: func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
				createAction := action.(clientGoTesting.CreateActionImpl)
				createNamespace := createAction.GetObject().(*v1.Namespace)
				assert.Nil(t, createNamespace.ObjectMeta.Labels)
				return false, nil, err
			},
		},
		{
			name: "creates a new namespace with labels",
			request: &v1alpha1.CreateNamespaceRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
				Labels: map[string]string{
					"label1": "value1",
					"label2": "value2",
				},
			},
			expectedResponse: emptyResponse,
			validator: func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
				createAction := action.(clientGoTesting.CreateActionImpl)
				createNamespace := createAction.GetObject().(*v1.Namespace)
				assert.Contains(t, createNamespace.ObjectMeta.Labels, "label1")
				assert.Contains(t, createNamespace.ObjectMeta.Labels, "label2")
				return false, nil, err
			},
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
			if tc.validator != nil {
				fakeClient.PrependReactor("create", "namespaces", tc.validator)
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

	defaultRequest := &v1alpha1.GetNamespaceNamesRequest{
		Cluster: "default",
	}

	testCases := []struct {
		name                    string
		request                 *v1alpha1.GetNamespaceNamesRequest
		trustedNamespacesConfig common.TrustedNamespaces
		k8sError                error
		requestHeaders          http.Header
		expectedResponse        *v1alpha1.GetNamespaceNamesResponse
		expectedErrorCode       codes.Code
		existingObjects         []runtime.Object
	}{
		{
			name:    "returns existing namespaces if user has RBAC",
			request: defaultRequest,
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "default",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
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
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
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
			name:    "returns permission denied if k8s returns a forbidden error",
			request: defaultRequest,
			k8sError: k8serrors.NewForbidden(schema.GroupResource{
				Group:    "v1",
				Resource: "namespaces",
			}, "default", errors.New("Bang")),
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name:              "returns an internal error if k8s returns an unexpected error",
			request:           defaultRequest,
			k8sError:          k8serrors.NewInternalError(errors.New("Bang")),
			expectedErrorCode: codes.Internal,
		},
		{
			name: "it should return the list of only active namespaces if accessible",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "terminating-ns",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceTerminating,
					},
				},
			},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should return the list of namespaces matching the trusted namespaces header",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"ns1",
					"ns2",
				},
			},
		},
		{
			name: "it should return the existing list of namespaces when trusted namespaces header does not match pattern",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"nspace:ns1", "nspace:ns2"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should return the existing list of namespaces when trusted namespaces header does not match name",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			requestHeaders: http.Header{"Y-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should return the existing list of namespaces when trusted namespaces header name is empty",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should return the existing list of namespaces when trusted namespaces pattern is empty",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "",
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should return some of the namespaces from trusted namespaces header when not all match the pattern",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"namespace:ns1:read", "namespace:ns2", "ns3", "namespace:ns4", "ns:ns5:write"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"ns2",
					"ns4",
				},
			},
		},
		{
			name: "it should return existing namespaces if no trusted ns header but trusted configuration is in place",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			trustedNamespacesConfig: common.TrustedNamespaces{
				HeaderName:    "X-Consumer-Groups",
				HeaderPattern: "^namespace:(\\w+)$",
			},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
		},
		{
			name: "it should ignore incoming trusted namespaces header when no trusted configuration is in place",
			existingObjects: []runtime.Object{
				&core.Namespace{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Namespace",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
					Status: v1.NamespaceStatus{
						Phase: v1.NamespaceActive,
					},
				},
			},
			requestHeaders: http.Header{"X-Consumer-Groups": []string{"namespace:ns1:read", "namespace:ns2", "ns3", "namespace:ns4", "ns:ns5:write"}},
			expectedResponse: &v1alpha1.GetNamespaceNamesResponse{
				NamespaceNames: []string{
					"foo",
				},
			},
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

			backgroundClientGetter := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
				return clientgetter.
					NewBuilder().
					WithTyped(fakeClient).
					Build(), nil
			}

			pluginConfig := &common.ResourcesPluginConfig{}
			if (tc.trustedNamespacesConfig != common.TrustedNamespaces{}) {
				pluginConfig.TrustedNamespaces = tc.trustedNamespacesConfig
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
				clusterServiceAccountClientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
				localServiceAccountClientGetter: backgroundClientGetter,
				clientQPS:                       5,
				pluginConfig:                    pluginConfig,
			}

			ctx := context.Background()
			for headerName, headerValue := range tc.requestHeaders {
				ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(headerName, strings.Join(headerValue, ",")))
			}

			response, err := s.GetNamespaceNames(ctx, tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}

func TestCanI(t *testing.T) {

	ignoredUnexported := cmpopts.IgnoreUnexported(
		v1alpha1.CanIResponse{},
	)

	testCases := []struct {
		name              string
		isAllowed         bool
		request           *v1alpha1.CanIRequest
		expectedResponse  *v1alpha1.CanIResponse
		k8sError          error
		expectedErrorCode codes.Code
	}{
		{
			name:      "returns allowed",
			isAllowed: true,
			request: &v1alpha1.CanIRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster: "default",
				},
			},
			expectedResponse: &v1alpha1.CanIResponse{
				Allowed: true,
			},
		},
		{
			name:      "returns forbidden",
			isAllowed: false,
			request: &v1alpha1.CanIRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster: "default",
				},
			},
			expectedResponse: &v1alpha1.CanIResponse{
				Allowed: false,
			},
		},
		{
			name:              "requires context parameter",
			request:           &v1alpha1.CanIRequest{},
			expectedErrorCode: codes.InvalidArgument,
		},
		{
			name: "requires cluster parameter",
			request: &v1alpha1.CanIRequest{
				Context: &pkgsGRPCv1alpha1.Context{},
			},
			expectedErrorCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			fakeClient := typfake.NewSimpleClientset()
			if tc.k8sError != nil {
				fakeClient.CoreV1().(*fakecorev1.FakeCoreV1).PrependReactor("list", "namespaces", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.NamespaceList{}, tc.k8sError
				})
			}

			// Creating an authorized clientGetter
			fakeClient.PrependReactor("create", "selfsubjectaccessreviews", func(action clientGoTesting.Action) (handled bool, ret runtime.Object, err error) {
				return true, &authorizationv1.SelfSubjectAccessReview{
					Status: authorizationv1.SubjectAccessReviewStatus{Allowed: tc.isAllowed},
				}, nil
			})

			backgroundClientGetter := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
				return clientgetter.
					NewBuilder().
					WithTyped(fakeClient).
					Build(), nil
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return fakeClient, nil, nil
				},
				localServiceAccountClientGetter: backgroundClientGetter,
				clientQPS:                       5,
			}

			response, err := s.CanI(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedErrorCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, ignoredUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported))
			}
		})
	}
}
