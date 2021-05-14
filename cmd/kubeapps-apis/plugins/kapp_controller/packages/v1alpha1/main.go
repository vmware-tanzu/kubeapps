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
	v1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
)

// RegisterWithGRPCServer enables a plugin to register with a gRPC server.
func RegisterWithGRPCServer(s grpc.ServiceRegistrar) {
	v1alpha1.RegisterPackagesServiceServer(s, &Server{})
}

// RegisterHTTPHandlerFromEndpoint enables a plugin to register an http
// handler to translate to the gRPC request.
func RegisterHTTPHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	// TODO: uptohere: the gw pb files are not being generated so this does not exist.
	// But it exists for core/plugins/v1alpha1 ??
	return v1alpha1.RegisterPackagesServiceHandlerFromEndpoint(ctx, mux, endpoint, opts)
}

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer
}

// func (s *Server) GetAvailablePackages(request *v1.GetAvailablePackagesRequest, stream PackageRepositoriesService_GetAvailablePackagesServer) error {
// 	repo := &v1.PackageRepository{
// 		Name:      "bitnami",
// 		Namespace: "kubeapps",
// 	}
// 	availablePackages := []*v1.AvailablePackage{
// 		{
// 			Name:          "package-a",
// 			LatestVersion: "1.2.0",
// 			Repository:    repo,
// 			IconUrl:       "http://example.com/package-a.jpg",
// 		},
// 		{
// 			Name:          "package-b",
// 			Repository:    repo,
// 			LatestVersion: "1.4.0",
// 			IconUrl:       "http://example.com/package-b.jpg",
// 		},
// 	}
// 	for _, pkg := range availablePackages {
// 		stream.Send(pkg)
// 	}
// 	return nil
// }
