// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"
	"fmt"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	"google.golang.org/grpc/codes"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynfake "k8s.io/client-go/dynamic/fake"
	typfake "k8s.io/client-go/kubernetes/fake"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetClientProvider(t *testing.T) {

	clientGetter := &ClientProvider{ClientsFunc: func(ctx context.Context, cluster string) (*ClientGetter, error) {
		return &ClientGetter{
			Typed: func() (kubernetes.Interface, error) { return typfake.NewSimpleClientset(), nil },
			Dynamic: func() (dynamic.Interface, error) {
				return dynfake.NewSimpleDynamicClientWithCustomListKinds(
					runtime.NewScheme(),
					map[schema.GroupVersionResource]string{
						{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
					},
				), nil
			},
			ApiExt: func() (apiext.Interface, error) {
				return apiextfake.NewSimpleClientset(), nil
			},
			ControllerRuntime: func() (ctrlclient.WithWatch, error) {
				return ctrlfake.NewClientBuilder().Build(), nil
			},
		}, nil

	}}

	badClientGetter := &ClientProvider{ClientsFunc: func(ctx context.Context, cluster string) (*ClientGetter, error) {
		return nil, fmt.Errorf("Bang!")
	}}

	testCases := []struct {
		name         string
		clientGetter ClientProviderInterface
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
			clientGetter: clientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCode == codes.OK {
				dynamicClient, err := tc.clientGetter.Dynamic(context.Background(), "")
				if err != nil {
					t.Fatal(err)
				} else if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}

				typedClient, err := tc.clientGetter.Typed(context.Background(), "")
				if err != nil {
					t.Fatal(err)
				} else if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}

				apiExClient, err := tc.clientGetter.ApiExt(context.Background(), "")
				if err != nil {
					t.Fatal(err)
				} else if apiExClient == nil {
					t.Errorf("got: nil, want: clientset.Interface")
				}

				ctrlClient, err := tc.clientGetter.ControllerRuntime(context.Background(), "")
				if err != nil {
					t.Fatal(err)
				} else if ctrlClient == nil {
					t.Errorf("got: nil, want: client.WithWatch")
				}
			}
		})
	}
}
