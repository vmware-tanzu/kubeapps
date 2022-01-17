/*
Copyright © 2021 VMware
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
