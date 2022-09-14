// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"k8s.io/client-go/rest"
)

// ServeOptions encapsulates the available command-line options.
type ServeOptions struct {
	Port                     int
	PluginDirs               []string
	ClustersConfigPath       string
	PluginConfigPath         string
	PinnipedProxyURL         string
	PinnipedProxyCACert      string
	GlobalHelmReposNamespace string
	UnsafeLocalDevKubeconfig bool
	QPS                      float32
	Burst                    int
}

// GatewayHandlerArgs is a helper struct just encapsulating all the args
// required when registering an HTTP handler for the gateway.
type GatewayHandlerArgs struct {
	Ctx         context.Context
	Mux         *runtime.ServeMux
	Addr        string
	DialOptions []grpc.DialOption
}

// KubernetesConfigGetter is a function type used throughout the apis server so
// that call-sites don't need to know how to obtain an authenticated client, but
// rather can just pass the request context and the cluster to get one.
type KubernetesConfigGetter func(ctx context.Context, cluster string) (*rest.Config, error)
