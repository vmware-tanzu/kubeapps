// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	pluginsgrpcv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/operators/packages/v1alpha1"
)

// Set the pluginDetail once during a module init function so the single struct
// can be used throughout the plugin.
var pluginDetail pluginsgrpcv1alpha1.Plugin

func init() {
	pluginDetail = pluginsgrpcv1alpha1.Plugin{
		Name:    "operators",
		Version: "v1alpha1",
	}
}

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
//
//nolint:deadcode
func RegisterWithGRPCServer(opts pluginsv1alpha1.GRPCPluginRegistrationOptions) (interface{}, error) {
	svr, err := NewServer(opts.ConfigGetter, opts.ClientQPS, opts.ClientBurst, opts.PluginConfigPath, opts.ClustersConfig)
	if err != nil {
		return nil, err
	}
	v1alpha1.RegisterOperatorsPackagesServiceServer(opts.Registrar, svr)
	v1alpha1.RegisterOperatorsRepositoriesServiceServer(opts.Registrar, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
//
//nolint:deadcode
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	err := v1alpha1.RegisterOperatorsPackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	} else {
		return v1alpha1.RegisterOperatorsRepositoriesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}
}

// GetPluginDetail returns a core.plugins.Plugin describing itself.
//
//nolint:deadcode
func GetPluginDetail() *pluginsgrpcv1alpha1.Plugin {
	return &pluginDetail
}
