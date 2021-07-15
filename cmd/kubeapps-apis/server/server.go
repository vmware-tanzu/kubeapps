/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	log "k8s.io/klog/v2"
)

type ServeOptions struct {
	Port               int
	PluginDirs         []string
	ClustersConfigPath string
	PinnipedProxyURL   string
	//temporary flags while this component in under heavy development
	UnsafeUseDemoSA          bool
	UnsafeLocalDevKubeconfig bool
}

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(serveOpts ServeOptions) {
	// Create the grpc server and register the reflection server (for now, useful for discovery
	// using grpcurl) or similar.
	grpcSrv := grpc.NewServer()
	reflection.Register(grpcSrv)

	// Create the http server, register our core service followed by any plugins.
	listenAddr := fmt.Sprintf(":%d", serveOpts.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gwArgs := gwHandlerArgs{
		ctx:         ctx,
		mux:         gatewayMux(),
		addr:        listenAddr,
		dialOptions: []grpc.DialOption{grpc.WithInsecure()},
	}
	httpSrv := &http.Server{
		Handler: gwArgs.mux,
	}

	// Create the core.plugins server which handles registration of plugins,
	// and register it for both grpc and http.
	pluginsServer, err := NewPluginsServer(serveOpts, grpcSrv, gwArgs)
	if err != nil {
		log.Fatalf("failed to initialize plugins server: %v", err)
	}
	plugins.RegisterPluginsServiceServer(grpcSrv, pluginsServer)
	err = plugins.RegisterPluginsServiceHandlerFromEndpoint(gwArgs.ctx, gwArgs.mux, gwArgs.addr, gwArgs.dialOptions)
	if err != nil {
		log.Fatalf("failed to register core.plugins handler for gateway: %v", err)
	}

	// Create the core.packages server and register it for both grpc and http.
	packages.RegisterPackagesServiceServer(grpcSrv, NewPackagesServer(pluginsServer.packagesPlugins))
	err = packages.RegisterPackagesServiceHandlerFromEndpoint(gwArgs.ctx, gwArgs.mux, gwArgs.addr, gwArgs.dialOptions)
	if err != nil {
		log.Fatalf("failed to register core.packages handler for gateway: %v", err)
	}

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
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

	httpSrv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if webrpcProxy.IsGrpcWebRequest(r) || webrpcProxy.IsAcceptableGrpcCorsRequest(r) || webrpcProxy.IsGrpcWebSocketRequest(r) {
			webrpcProxy.ServeHTTP(w, r)
		}
	})

	go grpcSrv.Serve(grpcLis)
	go grpcSrv.Serve(grpcwebLis)
	go httpSrv.Serve(httpLis)

	if serveOpts.UnsafeUseDemoSA {
		log.Warning("Using the demo Service Account for authenticating the requests. This is not recommended except for development purposes. Set `kubeappsapis.unsafeUseDemoSA: false` to remove this warning")
	}

	log.Infof("Starting server on :%d", serveOpts.Port)
	if err := mux.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// gwHandlerArgs is a helper struct just encapsulating all the args
// required when registering an HTTP handler for the gateway.
type gwHandlerArgs struct {
	ctx         context.Context
	mux         *runtime.ServeMux
	addr        string
	dialOptions []grpc.DialOption
}

// Create a gateway mux that does not emit unpopulated fields.
func gatewayMux() *runtime.ServeMux {
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
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
	err := gwmux.HandlePath(http.MethodGet, "/openapi.json", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/kubeapps-apis.swagger.json")
	}))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	err = gwmux.HandlePath(http.MethodGet, "/docs", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/index.html")
	}))
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return gwmux
}
