// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// Options are creation options for a Client.
type Options struct {
	// Scheme, if provided, will be used to map go structs to GroupVersionKinds
	Scheme *runtime.Scheme

	// Mapper, if provided, will be used to map GroupVersionKinds to Resources
	Mapper meta.RESTMapper
}

type ClientProviderInterface interface {
	// Typed returns "typed" API client for k8s that works with strongly-typed objects
	Typed(ctx context.Context, cluster string) (kubernetes.Interface, error)

	// Dynamic returns "untyped" API client for k8s that works with
	// k8s.io/apimachinery/pkg/apis/meta/v1/unstructured objects
	Dynamic(ctx context.Context, cluster string) (dynamic.Interface, error)

	// ControllerRuntime returns an instance of https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/client#Client
	// that also supports Watch operations
	ControllerRuntime(ctx context.Context, cluster string) (client.WithWatch, error)

	// ApiExt returns k8s API Extensions client interface, that can be used to query the
	// status of particular CRD in a cluster
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
		code := codes.FailedPrecondition
		if status.Code(err) == codes.Unauthenticated {
			// want to make sure we return same status in this case
			code = codes.Unauthenticated
		}
		return nil, status.Errorf(code, "unable to build clients due to: %v", err)
	}
	return clientGetter.Typed()
}

func (cp ClientProvider) Dynamic(ctx context.Context, cluster string) (dynamic.Interface, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		code := codes.FailedPrecondition
		if status.Code(err) == codes.Unauthenticated {
			// want to make sure we return same status in this case
			code = codes.Unauthenticated
		}
		return nil, status.Errorf(code, "unable to build clients due to: %v", err)
	}
	return clientGetter.Dynamic()
}

func (cp ClientProvider) ControllerRuntime(ctx context.Context, cluster string) (client.WithWatch, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		code := codes.FailedPrecondition
		if status.Code(err) == codes.Unauthenticated {
			// want to make sure we return same status in this case
			code = codes.Unauthenticated
		}
		return nil, status.Errorf(code, "unable to build clients due to: %v", err)
	}
	return clientGetter.ControllerRuntime()
}

func (cp ClientProvider) ApiExt(ctx context.Context, cluster string) (apiext.Interface, error) {
	clientGetter, err := cp.GetClients(ctx, cluster)
	if err != nil {
		code := codes.FailedPrecondition
		if status.Code(err) == codes.Unauthenticated {
			// want to make sure we return same status in this case
			code = codes.Unauthenticated
		}
		return nil, status.Errorf(code, "unable to build clients due to: %v", err)
	}
	return clientGetter.ApiExt()
}

func (cp ClientProvider) GetClients(ctx context.Context, cluster string) (*ClientGetter, error) {
	if cp.ClientsFunc == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "clients provider function is not set")
	}
	return cp.ClientsFunc(ctx, cluster)
}

type FixedClusterClientGetterFunc func(ctx context.Context) (*ClientGetter, error)

type FixedClusterClientProviderInterface interface {
	Typed(ctx context.Context) (kubernetes.Interface, error)
	Dynamic(ctx context.Context) (dynamic.Interface, error)
	ControllerRuntime(ctx context.Context) (client.WithWatch, error)
	ApiExt(ctx context.Context) (apiext.Interface, error)
	GetClients(ctx context.Context) (*ClientGetter, error)
}

type FixedClusterClientProvider struct {
	ClientsFunc FixedClusterClientGetterFunc
}

func (bcp FixedClusterClientProvider) Typed(ctx context.Context) (kubernetes.Interface, error) {
	clientGetter, err := bcp.GetClients(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.Typed()
}

func (bcp FixedClusterClientProvider) Dynamic(ctx context.Context) (dynamic.Interface, error) {
	clientGetter, err := bcp.GetClients(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.Dynamic()
}

func (bcp FixedClusterClientProvider) ControllerRuntime(ctx context.Context) (client.WithWatch, error) {
	clientGetter, err := bcp.GetClients(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.ControllerRuntime()
}

func (bcp FixedClusterClientProvider) ApiExt(ctx context.Context) (apiext.Interface, error) {
	clientGetter, err := bcp.GetClients(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to build clients due to: %v", err)
	}
	return clientGetter.ApiExt()
}

func (bcp FixedClusterClientProvider) GetClients(ctx context.Context) (*ClientGetter, error) {
	if bcp.ClientsFunc == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "clients provider function is not set")
	}
	return bcp.ClientsFunc(ctx)
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

		return buildClientGetter(config, options)
	}, nil
}

func buildClientGetter(config *rest.Config, options Options) (*ClientGetter, error) {
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
}

func NewClientProvider(configGetter core.KubernetesConfigGetter, options Options) (ClientProviderInterface, error) {
	clientsGetFunc, err := buildClientsProviderFunction(configGetter, options)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client provider functions due to: %v", err)
	}
	return &ClientProvider{ClientsFunc: clientsGetFunc}, nil
}

// NewBackgroundClientProvider returns an "out-of-band" or "in-cluster" client getter that returns various client interfaces
// with the context of the current cluster it is executing on and the service account
// configured for "kubeapps-apis" deployment
// https://github.com/vmware-tanzu/kubeapps/issues/3560
// flux plug-in runs out-of-request interactions with the Kubernetes API server.
// Although we've already ensured that if the flux plugin is selected, that the service account
// will be granted additional read privileges, we also need to ensure that the plugin can get a
// config based on the service account rather than the request context
func NewBackgroundClientProvider(options Options, clientQPS float32, clientBurst int) FixedClusterClientProviderInterface {
	return &FixedClusterClientProvider{ClientsFunc: func(ctx context.Context) (*ClientGetter, error) {
		// Some plugins currently support interactions with the default (kubeapps) cluster only
		if config, err := rest.InClusterConfig(); err != nil {
			code := codes.FailedPrecondition
			if status.Code(err) == codes.Unauthenticated {
				// want to make sure we return same status in this case
				code = codes.Unauthenticated
			}
			return nil, status.Errorf(code, "unable to get in cluster config due to: %v", err)
		} else {
			config.QPS = clientQPS
			config.Burst = clientBurst
			return buildClientGetter(config, options)
		}
	}}
}

// Builder builds a ClientProviderInterface or FixedClusterClientProviderInterface instance.
// convenience functions exported only for unit tests in plugins
type Builder struct {
	ClientGetter
}

// NewBuilder returns a new builder
func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) WithDynamic(i dynamic.Interface) *Builder {
	b.Dynamic = func() (dynamic.Interface, error) {
		return i, nil
	}
	return b
}

func (b *Builder) WithTyped(i kubernetes.Interface) *Builder {
	b.Typed = func() (kubernetes.Interface, error) {
		return i, nil
	}
	return b
}

func (b *Builder) WithApiExt(a apiext.Interface) *Builder {
	b.ApiExt = func() (apiext.Interface, error) {
		return a, nil
	}
	return b
}

func (b *Builder) WithControllerRuntime(c client.WithWatch) *Builder {
	b.ControllerRuntime = func() (client.WithWatch, error) {
		return c, nil
	}
	return b
}

// Build builds and returns a new instance of ClientProviderInterface.
func (b *Builder) Build() ClientProviderInterface {
	return &ClientProvider{ClientsFunc: func(ctx context.Context, cluster string) (*ClientGetter, error) {
		return &b.ClientGetter, nil
	}}
}

func (b *Builder) BuildFixedCluster() FixedClusterClientProviderInterface {
	return &FixedClusterClientProvider{ClientsFunc: func(ctx context.Context) (*ClientGetter, error) {
		return &b.ClientGetter, nil
	}}
}
