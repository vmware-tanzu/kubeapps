// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	log "k8s.io/klog/v2"
)

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
//nolint:deadcode
func RegisterWithGRPCServer(opts pluginsv1alpha1.GRPCPluginRegistrationOptions) (interface{}, error) {
	log.Info("+fluxv2 RegisterWithGRPCServer")

	// TODO (gfichtenholt) stub channel for now. Ideally, the caller (kubeappsapis-server)
	// passes that in and closes when is being gracefully shut down. That, or provide a
	// 'Shutdown' hook
	stopCh := make(chan struct{})

	svr, err := NewServer(opts.ConfigGetter, opts.ClustersConfig.KubeappsClusterName, stopCh, opts.PluginConfigPath)
	if err != nil {
		return nil, err
	}
	v1alpha1.RegisterFluxV2PackagesServiceServer(opts.Registrar, svr)
	v1alpha1.RegisterFluxV2RepositoriesServiceServer(opts.Registrar, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
//nolint:deadcode
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	log.Info("+fluxv2 RegisterHTTPHandlerFromEndpoint")
	err := v1alpha1.RegisterFluxV2PackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	if err != nil {
		return err
	} else {
		return v1alpha1.RegisterFluxV2RepositoriesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
	}
}
