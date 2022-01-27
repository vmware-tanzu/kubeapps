// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"
	"fmt"
	"testing"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8dynamicclientfake "k8s.io/client-go/dynamic/fake"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
)

func TestGetClient(t *testing.T) {
	testClientGetter := func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		return k8stypedclientfake.NewSimpleClientset(), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
			k8sruntime.NewScheme(),
			map[k8sschema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}
	badClientGetter := func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		return nil, nil, fmt.Errorf("Bang!")
	}

	testCases := []struct {
		name         string
		clientGetter ClientGetterFunc
		statusCode   grpccodes.Code
	}{
		{
			name:         "it returns failed-precondition when configGetter itself errors",
			statusCode:   grpccodes.Unknown,
			clientGetter: badClientGetter,
		},
		{
			name:         "it returns client without error when configured correctly",
			statusCode:   grpccodes.OK,
			clientGetter: testClientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			typedClient, dynamicClient, err := tc.clientGetter(context.Background(), "")
			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a k8dynamicclient.Interface implementation.
			if tc.statusCode == grpccodes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: k8dynamicclient.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: k8stypedclient.Interface")
				}
			}
		})
	}
}
