// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetClient(t *testing.T) {
	testClientGetter := func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
		return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}
	badClientGetter := func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
		return nil, nil, fmt.Errorf("Bang!")
	}

	testCases := []struct {
		name         string
		clientGetter ClientGetterFunc
		statusCode   codes.Code
	}{
		{
			name:         "it returns failed-precondition when configGetter itself errors",
			statusCode:   codes.Unknown,
			clientGetter: badClientGetter,
		},
		{
			name:         "it returns client without error when configured correctly",
			statusCode:   codes.OK,
			clientGetter: testClientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			typedClient, dynamicClient, err := tc.clientGetter(context.Background(), "")
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCode == codes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}
			}
		})
	}
}
