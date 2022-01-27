// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	grpcgwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	resourcesGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	kubeutils "github.com/kubeapps/kubeapps/pkg/kube"
	grpc "google.golang.org/grpc"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var (
	pluginDetail pluginsGRPCv1alpha1.Plugin
	// This version var is updated during the build (see the -ldflags option
	// in the cmd/kubeapps-apis/Dockerfile)
	version = "devel"
)

func init() {
	pluginDetail = pluginsGRPCv1alpha1.Plugin{
		Name:    "resources",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
func RegisterWithGRPCServer(s grpc.ServiceRegistrar, configGetter apiscore.KubernetesConfigGetter, clustersConfig kubeutils.ClustersConfig, pluginConfigPath string) (interface{}, error) {
	svr, err := NewServer(configGetter)
	if err != nil {
		return nil, err
	}
	resourcesGRPCv1alpha1.RegisterResourcesServiceServer(s, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *grpcgwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return resourcesGRPCv1alpha1.RegisterResourcesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// GetPluginDetail returns a apiscore.plugins.Plugin describing itself.
func GetPluginDetail() *pluginsGRPCv1alpha1.Plugin {
	return &pluginDetail
}
