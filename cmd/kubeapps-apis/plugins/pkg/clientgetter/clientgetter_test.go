// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynfake "k8s.io/client-go/dynamic/fake"
	typfake "k8s.io/client-go/kubernetes/fake"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClient(t *testing.T) {
	testClientGetter := func(context.Context, string) (ClientInterfaces, error) {
		return &clientInterfacesType{
			typfake.NewSimpleClientset(),
			dynfake.NewSimpleDynamicClientWithCustomListKinds(
				runtime.NewScheme(),
				map[schema.GroupVersionResource]string{
					{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
				},
			),
			apiextfake.NewSimpleClientset(),
			ctrlfake.NewClientBuilder().Build(),
		}, nil
	}
	badClientGetter := func(context.Context, string) (ClientInterfaces, error) {
		return &clientInterfacesType{nil, nil, nil, nil}, fmt.Errorf("Bang!")
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
			ifcs, err := tc.clientGetter(context.Background(), "")
			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCode == codes.OK {
				dynamicClient, err := ifcs.Dynamic()
				if err != nil {
					t.Fatal(err)
				} else if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}

				typedClient, err := ifcs.Typed()
				if err != nil {
					t.Fatal(err)
				} else if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}

				apiExClient, err := ifcs.ApiExt()
				if err != nil {
					t.Fatal(err)
				} else if apiExClient == nil {
					t.Errorf("got: nil, want: clientset.Interface")
				}

				ctrlClient, err := ifcs.ControllerRuntime()
				if err != nil {
					t.Fatal(err)
				} else if ctrlClient == nil {
					t.Errorf("got: nil, want: client.WithWatch")
				}
			}
		})
	}
}
