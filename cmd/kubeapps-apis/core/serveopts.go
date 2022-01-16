/*
Copyright 2021 VMware. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	GlobalReposNamespace     string
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
