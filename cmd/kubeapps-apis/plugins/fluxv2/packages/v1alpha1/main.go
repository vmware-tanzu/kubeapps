// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/fluxv2/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	log "k8s.io/klog/v2"
)

// RegisterWithGRPCServer enables a plugin to register with a gRPC server
// returning the server implementation.
func RegisterWithGRPCServer(s grpc.ServiceRegistrar, configGetter core.KubernetesConfigGetter,
	clustersConfig kube.ClustersConfig, pluginConfigPath string) (interface{}, error) {
	log.Infof("+fluxv2 RegisterWithGRPCServer")

	// TODO (gfichtenholt) stub channel for now. Ideally, the caller (kubeappsapis-server)
	// passes that in and closes when is being gracefully shut down. That, or provide a
	// 'Shutdown' hook
	stopCh := make(chan struct{})

	svr, err := NewServer(configGetter, clustersConfig.KubeappsClusterName, stopCh, pluginConfigPath)
	if err != nil {
		return nil, err
	}
	v1alpha1.RegisterFluxV2PackagesServiceServer(s, svr)
	return svr, nil
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	log.Infof("+fluxv2 RegisterHTTPHandlerFromEndpoint")
	return v1alpha1.RegisterFluxV2PackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}
