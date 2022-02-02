// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"net/http"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	helmregistry "helm.sh/helm/v3/pkg/registry"
)

type (
	// Derived from https://github.com/helm/helm/blob/v3.8.0/pkg/registry/client.go
	IOCIClient interface {
		Login(host string, options ...helmregistry.LoginOption) error
		Logout(host string, opts ...helmregistry.LogoutOption) error
		Pull(ref string, options ...helmregistry.PullOption) (*helmregistry.PullResult, error)
		Push(data []byte, ref string, options ...helmregistry.PushOption) (*helmregistry.PushResult, error)
		Tags(ref string) ([]string, error)
	}

	// Factory and wrapper for creating OCI clients (so that we an fake it for testing)
	IOCIClientFactory interface {
		BuildOCIClient(resolver remotes.Resolver) (IOCIClient, error)
	}
	OCIClientFactory        struct{}
	OCIClientFactoryWrapper struct{ OCIClientFactory }
)

// GetResolver returns a containerd resolver configured for the given headers and http client
func GetResolver(headers http.Header, netClient *http.Client) remotes.Resolver {
	return docker.NewResolver(docker.ResolverOptions{
		Hosts:   docker.ConfigureDefaultRegistries(docker.WithClient(netClient)),
		Headers: headers,
	})
}

// BuildOCIClient is the main function to call every time we need an OCICient
// Depending on the factory passed, we'll create the normal or the mocked one
func BuildOCIClient(cf IOCIClientFactory, resolver remotes.Resolver) (IOCIClient, error) {
	return cf.BuildOCIClient(resolver)
}

// BuildOCIClient of this wrapper is for avoiding type convertion erros
func (cf OCIClientFactoryWrapper) BuildOCIClient(resolver remotes.Resolver) (IOCIClient, error) {
	return cf.OCIClientFactory.BuildOCIClient(resolver)
}

// BuildOCIClient of this factory will initialize a helm registry client with a custom resolver
func (cf OCIClientFactory) BuildOCIClient(resolver remotes.Resolver) (IOCIClient, error) {
	var ociClient *helmregistry.Client
	var err error

	if resolver == nil {
		ociClient, err = helmregistry.NewClient()
	} else {
		ociClient, err = helmregistry.NewClient(helmregistry.ClientOptResolver(resolver))
	}

	if err != nil {
		return nil, err
	}
	return ociClient, nil
}
