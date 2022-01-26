// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	grpcgwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	grpcweb "github.com/improbable-eng/grpc-web/go/grpcweb"
	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core/packages/v1alpha1"
	pluginsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	cmux "github.com/soheilhy/cmux"
	grpc "google.golang.org/grpc"
	reflection "google.golang.org/grpc/reflection"
	grpcstatus "google.golang.org/grpc/status"
	protojson "google.golang.org/protobuf/encoding/protojson"
	log "k8s.io/klog/v2"
)

func getLogLevelOfEndpoint(endpoint string) log.Level {

	// Add all endpoint function names which you want to suppress in interceptor logging
	supressLoggingOfEndpoints := []string{"GetConfiguredPlugins"}
	var level log.Level

	// level=3 is default logging level
	level = 3
	for i := 0; i < len(supressLoggingOfEndpoints); i++ {
		if strings.Contains(endpoint, supressLoggingOfEndpoints[i]) {
			level = 4
			break
		}
	}

	return level
}

// LogRequest is a gRPC UnaryServerInterceptor that will log the API call
func LogRequest(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (response interface{}, err error) {

	start := time.Now()
	res, err := handler(ctx, req)

	level := getLogLevelOfEndpoint(info.FullMethod)

	// Format string : [status code] [duration] [full path]
	// OK 97.752µs /kubeappsapis.apiscore.packages.v1alpha1.PackagesService/GetAvailablePackageSummaries
	log.V(level).Infof("%v %s %s\n",
		grpcstatus.Code(err),
		time.Since(start),
		info.FullMethod)

	return res, err
}

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(serveOpts apiscore.ServeOptions) error {
	// Create the grpc server and register the reflection server (for now, useful for discovery
	// using grpcurl) or similar.

	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(LogRequest))
	reflection.Register(grpcSrv)

	// Create the http server, register our core service followed by any plugins.
	listenAddr := fmt.Sprintf(":%d", serveOpts.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gw, err := gatewayMux()
	if err != nil {
		return fmt.Errorf("Failed to create gateway: %v", err)
	}
	gwArgs := apiscore.GatewayHandlerArgs{
		Ctx:         ctx,
		Mux:         gw,
		Addr:        listenAddr,
		DialOptions: []grpc.DialOption{grpc.WithInsecure()},
	}

	// Create the apiscore.plugins.v1alpha1 server which handles registration of
	// plugins, and register it for both grpc and http.
	pluginsServer, err := pluginsv1alpha1.NewPluginsServer(serveOpts, grpcSrv, gwArgs)
	if err != nil {
		return fmt.Errorf("failed to initialize plugins server: %v", err)
	}
	pluginsGRPCv1alpha1.RegisterPluginsServiceServer(grpcSrv, pluginsServer)
	err = pluginsGRPCv1alpha1.RegisterPluginsServiceHandlerFromEndpoint(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)
	if err != nil {
		return fmt.Errorf("failed to register apiscore.plugins handler for gateway: %v", err)
	}

	// Ask the plugins server for plugins with GRPC servers that fulfil the core
	// packaging v1alpha1 API, then pass to the constructor below.
	// The argument for the reflect.TypeOf is based on what grpc-go
	// does itself at:
	// https://github.com/grpc/grpc-go/blob/v1.38.0/server.go#L621
	packagingPlugins := pluginsServer.GetPluginsSatisfyingInterface(reflect.TypeOf((*pkgsGRPCv1alpha1.PackagesServiceServer)(nil)).Elem())

	// Create the apiscore.packages server and register it for both grpc and http.
	packagesServer, err := pkgsv1alpha1.NewPackagesServer(packagingPlugins)
	if err != nil {
		return fmt.Errorf("failed to create apiscore.packages.v1alpha1 server: %w", err)
	}
	pkgsGRPCv1alpha1.RegisterPackagesServiceServer(grpcSrv, packagesServer)
	err = pkgsGRPCv1alpha1.RegisterPackagesServiceHandlerFromEndpoint(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)
	if err != nil {
		return fmt.Errorf("failed to register apiscore.packages handler for gateway: %v", err)
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Multiplex the connection between grpc and http.
	// Note: due to a change in the grpc protocol, it's no longer possible to just match
	// on the simpler cmux.HTTP2HeaderField("content-type", "application/grpc"). More details
	// at https://github.com/soheilhy/cmux/issues/64
	mux := cmux.New(lis)
	grpcLis := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	grpcwebLis := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc-web"))
	httpLis := mux.Match(cmux.Any())

	webrpcProxy := grpcweb.WrapServer(grpcSrv,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
	)

	httpSrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if webrpcProxy.IsGrpcWebRequest(r) || webrpcProxy.IsAcceptableGrpcCorsRequest(r) || webrpcProxy.IsGrpcWebSocketRequest(r) {
				webrpcProxy.ServeHTTP(w, r)
			} else {
				gwArgs.Mux.ServeHTTP(w, r)
			}
		}),
	}

	go func() {
		err := grpcSrv.Serve(grpcLis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		err := grpcSrv.Serve(grpcwebLis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		err := httpSrv.Serve(httpLis)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	if serveOpts.UnsafeLocalDevKubeconfig {
		log.Warning("Using the local Kubeconfig file instead of the actual in-cluster's config. This is not recommended except for development purposes.")
	}

	log.Infof("Starting server on :%d", serveOpts.Port)
	if err := mux.Serve(); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// Create a gateway mux that does not emit unpopulated fields.
func gatewayMux() (*grpcgwruntime.ServeMux, error) {
	gwmux := grpcgwruntime.NewServeMux(
		grpcgwruntime.WithMarshalerOption(grpcgwruntime.MIMEWildcard, &grpcgwruntime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				EmitUnpopulated: false,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	)

	// TODO(agamez): remove these '/openapi.json' and '/docs' paths. They are serving a
	// static 'swagger-ui' dashboard with hardcoded values just intended for develoment purposes.
	// This docs will eventually converge into the docs already (properly) served by the dashboard
	err := gwmux.HandlePath(http.MethodGet, "/openapi.json", grpcgwruntime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/kubeapps-apis.swagger.json")
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to serve: %v", err)
	}

	err = gwmux.HandlePath(http.MethodGet, "/docs", grpcgwruntime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/index.html")
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to serve: %v", err)
	}

	return gwmux, nil
}
