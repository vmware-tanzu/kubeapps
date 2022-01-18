// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var (
	pluginDetail plugins.Plugin
	// This version var is updated during the build (see the -ldflags option
	// in the cmd/kubeapps-apis/Dockerfile)
	version = "devel"
)

func init() {
	pluginDetail = plugins.Plugin{
		Name:    "helm.packages",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
func RegisterWithGRPCServer(s grpc.ServiceRegistrar, configGetter core.KubernetesConfigGetter,
	clustersConfig kube.ClustersConfig, pluginConfigPath string) (interface{}, error) {
	svr := NewServer(configGetter, clustersConfig.KubeappsClusterName, clustersConfig.GlobalReposNamespace, pluginConfigPath)
	v1alpha1.RegisterHelmPackagesServiceServer(s, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1alpha1.RegisterHelmPackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *plugins.Plugin {
	return &pluginDetail
}
