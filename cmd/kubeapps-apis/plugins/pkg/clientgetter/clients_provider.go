// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TypedClientFunc func() (kubernetes.Interface, error)
type DynamicClientFunc func() (dynamic.Interface, error)
type ApiExtFunc func() (apiext.Interface, error)

// ControllerRuntimeFunc returns an instance of https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/client#Client
// that also supports Watch operations
type ControllerRuntimeFunc func() (client.WithWatch, error)

// ClientGetter holds the functions that will actually return the client independently
type ClientGetter struct {
	Typed             TypedClientFunc
	Dynamic           DynamicClientFunc
	ControllerRuntime ControllerRuntimeFunc
	ApiExt            ApiExtFunc
}

// GetClientsFunc is a function that provides a ClientGetter per cluster
type GetClientsFunc func(ctx context.Context, cluster string) (*ClientGetter, error)

type ClientProviderInterface interface {
	Typed(ctx context.Context, cluster string) (kubernetes.Interface, error)
	Dynamic(ctx context.Context, cluster string) (dynamic.Interface, error)
	ControllerRuntime(ctx context.Context, cluster string) (client.WithWatch, error)
	ApiExt(ctx context.Context, cluster string) (apiext.Interface, error)
	GetClients(ctx context.Context, cluster string) (*ClientGetter, error)
}

// ClientProvider provides a real implementation of the ClientProviderInterface interface
type ClientProvider struct {
	ClientsFunc GetClientsFunc
}

func (cp ClientProvider) Typed(ctx context.Context, cluster string) (kubernetes.Interface, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.Typed()
}

func (cp ClientProvider) Dynamic(ctx context.Context, cluster string) (dynamic.Interface, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.Dynamic()
}

func (cp ClientProvider) ControllerRuntime(ctx context.Context, cluster string) (client.WithWatch, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.ControllerRuntime()
}

func (cp ClientProvider) ApiExt(ctx context.Context, cluster string) (apiext.Interface, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.ApiExt()
}

func (cp ClientProvider) GetClients(ctx context.Context, cluster string) (*ClientGetter, error) {
	if cp.ClientsFunc == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "clients provider function is not set")
	}
	return cp.ClientsFunc(ctx, cluster)
}

// buildClientsProviderFunction Creates the default function for obtaining a ClientGetter
func buildClientsProviderFunction(configGetter core.KubernetesConfigGetter, options Options) (GetClientsFunc, error) {
	return func(ctx context.Context, cluster string) (*ClientGetter, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		config, err := configGetter(ctx, cluster)
		if err != nil {
			code := codes.FailedPrecondition
			if status.Code(err) == codes.Unauthenticated {
				// want to make sure we return same status in this case
				code = codes.Unauthenticated
			}
			return nil, status.Errorf(code, "unable to get in cluster config due to: %v", err)
		}

		var typedClientFunc TypedClientFunc = func() (kubernetes.Interface, error) {
			typedClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "unable to get typed client due to: %v", err)
			}
			return typedClient, nil
		}

		var dynamicClientFunc DynamicClientFunc = func() (dynamic.Interface, error) {
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client due to: %v", err)
			}
			return dynamicClient, nil
		}

		var controllerRuntimeClientFunc ControllerRuntimeFunc = func() (client.WithWatch, error) {
			ctrlOpts := client.Options{}
			if options.Scheme != nil {
				ctrlOpts.Scheme = options.Scheme
			}
			if options.Mapper != nil {
				ctrlOpts.Mapper = options.Mapper
			}

			ctrlClient, err := client.NewWithWatch(config, ctrlOpts)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "unable to get controller runtime client due to: %v", err)
			}
			return ctrlClient, nil
		}

		var apiExtClientFunc ApiExtFunc = func() (apiext.Interface, error) {
			apiExtensions, err := apiext.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "unable to get api extensions client due to: %v", err)
			}
			return apiExtensions, nil
		}
		return &ClientGetter{typedClientFunc, dynamicClientFunc, controllerRuntimeClientFunc, apiExtClientFunc}, nil
	}, nil
}

func NewClientProvider(configGetter core.KubernetesConfigGetter, options Options) (ClientProviderInterface, error) {
	clientsGetFunc, err := buildClientsProviderFunction(configGetter, options)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client provider functions due to: %v", err)
	}
	return &ClientProvider{ClientsFunc: clientsGetFunc}, nil
}

func NewFixedClientProvider(clientsGetter *ClientGetter) ClientProviderInterface {
	return &ClientProvider{ClientsFunc: func(ctx context.Context, cluster string) (*ClientGetter, error) {
		return clientsGetter, nil
	}}
}
