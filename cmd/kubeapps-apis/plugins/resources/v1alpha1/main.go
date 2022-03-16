// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pluginsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	pluginsgrpcv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
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
		Name:    "resources",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
func RegisterWithGRPCServer(opts pluginsv1alpha1.GRPCPluginRegistrationOptions) (interface{}, error) {
	svr, err := NewServer(opts.ConfigGetter, opts.ClientQPS, opts.ClientBurst)
	if err != nil {
		return nil, err
	}
	v1alpha1.RegisterResourcesServiceServer(opts.Registrar, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return v1alpha1.RegisterResourcesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
func GetPluginDetail() *pluginsgrpcv1alpha1.Plugin {
	return &pluginDetail
}
