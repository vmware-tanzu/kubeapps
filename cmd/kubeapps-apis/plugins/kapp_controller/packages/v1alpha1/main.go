// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	grpc "google.golang.org/grpc"

	grpcgwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	pkgkappv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	kubeutils "github.com/kubeapps/kubeapps/pkg/kube"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var pluginDetail pluginsGRPCv1alpha1.Plugin

func init() {
	pluginDetail = pluginsGRPCv1alpha1.Plugin{
		Name:    "kapp_controller.packages",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
func RegisterWithGRPCServer(s grpc.ServiceRegistrar, configGetter apiscore.KubernetesConfigGetter,
	clustersConfig kubeutils.ClustersConfig, pluginConfigPath string) (interface{}, error) {
	svr := NewServer(configGetter, clustersConfig.KubeappsClusterName, pluginConfigPath)
	pkgkappv1alpha1.RegisterKappControllerPackagesServiceServer(s, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *grpcgwruntime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return pkgkappv1alpha1.RegisterKappControllerPackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// GetPluginDetail returns a apiscore.plugins.Plugin describing itself.
func GetPluginDetail() *pluginsGRPCv1alpha1.Plugin {
	return &pluginDetail
}
