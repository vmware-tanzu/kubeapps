// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
)

type HelmActionConfigGetterFunc func(ctx context.Context, namespace string) (*action.Configuration, error)

type ClientGetterFunc func(ctx context.Context, cluster string) (ClientInterfaces, error)
type BackgroundClientGetterFunc func(context.Context) (ClientInterfaces, error)

type ClientInterfaces interface {
	// returns "typed" API client for k8s that works with strongly-typed objects
	Typed() (kubernetes.Interface, error)
	// returns "untyped" API client for k8s that works with
	// k8s.io/apimachinery/pkg/apis/meta/v1/unstructured objects
	Dynamic() (dynamic.Interface, error)
	// returns k8s API Extensions client interface, that can be used to query the
	// status of particular CRD in a cluster
	ApiExt() (apiext.Interface, error)
}

// very basic implementation to start with. Will enhance later as needed
// such as lazy/on-demand loading of clients, caching when possible, etc.
type clientInterfacesType struct {
	typed kubernetes.Interface
	dyn   dynamic.Interface
	apiex apiext.Interface
}

func (c *clientInterfacesType) Typed() (kubernetes.Interface, error) {
	return c.typed, nil
}

func (c *clientInterfacesType) Dynamic() (dynamic.Interface, error) {
	return c.dyn, nil
}

func (c *clientInterfacesType) ApiExt() (apiext.Interface, error) {
	return c.apiex, nil
}

func NewHelmActionConfigGetter(configGetter core.KubernetesConfigGetter, cluster string) HelmActionConfigGetterFunc {
	return func(ctx context.Context, namespace string) (*action.Configuration, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		config, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
		}

		restClientGetter := agent.NewConfigFlagsFromCluster(namespace, config)
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to create kubernetes client due to: %v", err)
		}
		// TODO(mnelson): Update to allow different helm storage options.
		storage := agent.StorageForSecrets(namespace, clientSet)
		return &action.Configuration{
			RESTClientGetter: restClientGetter,
			KubeClient:       kube.New(restClientGetter),
			Releases:         storage,
			Log:              log.Infof,
		}, nil
	}
}

func NewClientGetter(configGetter core.KubernetesConfigGetter) ClientGetterFunc {
	return func(ctx context.Context, cluster string) (ClientInterfaces, error) {
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
		return clientGetterHelper(config)
	}
}

// returns an "out-of-band" or "in-cluster" client getter that returns various client interfaces
// with the context of the current cluster it is executing on and the service account
// configured for "kubeapps-apis" deployment
// https://github.com/kubeapps/kubeapps/issues/3560
// flux plug-in runs out-of-request interactions with the Kubernetes API server.
// Although we've already ensured that if the flux plugin is selected, that the service account
// will be granted additional read privileges, we also need to ensure that the plugin can get a
// config based on the service account rather than the request context
func NewBackgroundClientGetter(configGetter core.KubernetesConfigGetter) BackgroundClientGetterFunc {
	return func(ctx context.Context) (ClientInterfaces, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		if config, err := rest.InClusterConfig(); err != nil {
			code := codes.FailedPrecondition
			if status.Code(err) == codes.Unauthenticated {
				// want to make sure we return same status in this case
				code = codes.Unauthenticated
			}
			return nil, status.Errorf(code, "unable to get in cluster config due to: %v", err)
		} else {
			return clientGetterHelper(config)
		}
	}
}

// just a convenience func as a shortcut to get API Extension client in one line
func (cg BackgroundClientGetterFunc) ApiExt(ctx context.Context) (apiext.Interface, error) {
	if clientInterfaces, err := cg(ctx); err != nil {
		return nil, err
	} else if apiExt, err := clientInterfaces.ApiExt(); err != nil {
		return nil, err
	} else {
		return apiExt, nil
	}
}

// just a convenience func as a shortcut to get dynamic.Interface client in one line
func (cg BackgroundClientGetterFunc) Dynamic(ctx context.Context) (dynamic.Interface, error) {
	if clientInterfaces, err := cg(ctx); err != nil {
		return nil, err
	} else if dyn, err := clientInterfaces.Dynamic(); err != nil {
		return nil, err
	} else {
		return dyn, nil
	}
}

// just a convenience func as a shortcut to get kubernetes.Interface client in one line
func (cg ClientGetterFunc) Typed(ctx context.Context, cluster string) (kubernetes.Interface, error) {
	if clientInterfaces, err := cg(ctx, cluster); err != nil {
		return nil, err
	} else if typed, err := clientInterfaces.Typed(); err != nil {
		return nil, err
	} else {
		return typed, nil
	}
}

// just a convenience func as a shortcut to get dynamic.Interface client in one line
func (cg ClientGetterFunc) Dynamic(ctx context.Context, cluster string) (dynamic.Interface, error) {
	if clientInterfaces, err := cg(ctx, cluster); err != nil {
		return nil, err
	} else if dyn, err := clientInterfaces.Dynamic(); err != nil {
		return nil, err
	} else {
		return dyn, nil
	}
}

// just a convenience func as a shortcut to get kubernetes.Interface client in one line
func (cg BackgroundClientGetterFunc) Typed(ctx context.Context) (kubernetes.Interface, error) {
	if clientInterfaces, err := cg(ctx); err != nil {
		return nil, err
	} else if typed, err := clientInterfaces.Typed(); err != nil {
		return nil, err
	} else {
		return typed, nil
	}
}

func clientGetterHelper(config *rest.Config) (ClientInterfaces, error) {
	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get typed client due to: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client due to: %v", err)
	}
	apiExtensions, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get api extensions client due to: %v", err)
	}
	return NewClientInterfaces(typedClient, dynamicClient, apiExtensions), nil
}

// convenience func exported only for unit tests in plugins
func NewClientInterfaces(typedClient kubernetes.Interface, dynamicClient dynamic.Interface, apiExtensions apiext.Interface) ClientInterfaces {
	return &clientInterfacesType{
		typedClient,
		dynamicClient,
		apiExtensions}
}
