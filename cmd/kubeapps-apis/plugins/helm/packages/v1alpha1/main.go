// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	pluginsgrpcv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"google.golang.org/grpc"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var (
	pluginDetail pluginsgrpcv1alpha1.Plugin
	// This version var is updated during the build (see the -ldflags option
	// in the cmd/kubeapps-apis/Dockerfile)
	version = "devel"
)

func init() {
	pluginDetail = pluginsgrpcv1alpha1.Plugin{
		Name:    "helm.packages",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
//
//nolint:deadcode
func RegisterWithGRPCServer(opts pluginsv1alpha1.GRPCPluginRegistrationOptions) (interface{}, error) {
	svr := NewServer(opts.ConfigGetter, opts.ClustersConfig.KubeappsClusterName, opts.ClustersConfig.GlobalPackagingNamespace, opts.PluginConfigPath)
	v1alpha1.RegisterHelmPackagesServiceServer(opts.Registrar, svr)
	v1alpha1.RegisterHelmRepositoriesServiceServer(opts.Registrar, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
//
//nolint:deadcode
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1alpha1.RegisterHelmPackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *pluginsgrpcv1alpha1.Plugin {
	return &pluginDetail
}
