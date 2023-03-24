// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/soheilhy/cmux"

	grpchealth "github.com/bufbuild/connect-grpchealth-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	packagesv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/packages/v1alpha1"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packagesGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsConnect "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1/v1alpha1connect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	klogv2 "k8s.io/klog/v2"
)

func getLogLevelOfEndpoint(endpoint string) klogv2.Level {

	// Add all endpoint function names which you want to suppress in interceptor logging
	suppressLoggingOfEndpoints := []string{"GetConfiguredPlugins"}
	var level klogv2.Level

	// level=3 is default logging level
	level = 3
	for i := 0; i < len(suppressLoggingOfEndpoints); i++ {
		if strings.Contains(endpoint, suppressLoggingOfEndpoints[i]) {
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
	// OK 97.752µs /kubeappsapis.core.packages.v1alpha1.PackagesService/GetAvailablePackageSummaries
	klogv2.V(level).Infof("%v %s %s\n",
		status.Code(err),
		time.Since(start),
		info.FullMethod)

	return res, err
}

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(serveOpts core.ServeOptions) error {
	// Create the grpc server and register the reflection server (for now, useful for discovery
	// using grpcurl) or similar.

	grpcSrv := grpc.NewServer(grpc.ChainUnaryInterceptor(LogRequest))
	reflection.Register(grpcSrv)

	// Create the http server, register our core service followed by any plugins.
	// The cmux listen address will be a random port. We'll send traffic through to this
	// port from the main http.mux.
	listenAddrCMux := ":0"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gw, err := gatewayMux()
	if err != nil {
		return fmt.Errorf("failed to create gateway: %v", err)
	}
	gwArgs := core.GatewayHandlerArgs{
		Ctx:         ctx,
		Mux:         gw,
		Addr:        listenAddrCMux,
		DialOptions: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	}

	// The connect service handler automatically handles grpc-web, connect and
	// grpc for us, so we won't need all the extra code below once all services
	// have been transitioned to the new mux (and we can remove the use of cmux
	// once connect is used for all requests).

	// For now, we use the connect grpc mux by default and any unhandled paths
	// are routed to the old cmux handler's listener. I originally tried to
	// use the cmux as the default, but checking http2 header frames requires
	// writing settings, which appears to upset the handler.
	mux_connect := http.NewServeMux()

	// Create the core.plugins.v1alpha1 server which handles registration of
	// plugins, and register it for both grpc and http.
	pluginsServer, err := pluginsv1alpha1.NewPluginsServer(serveOpts, grpcSrv, gwArgs, mux_connect)
	if err != nil {
		return fmt.Errorf("failed to initialize plugins server: %v", err)
	}

	// The gRPC Health checker reports on all connected services.
	checker := grpchealth.NewStaticChecker(
		pluginsConnect.PluginsServiceName,
	)
	mux_connect.Handle(grpchealth.NewHandler(checker))

	if err = registerPackagesServiceServer(grpcSrv, pluginsServer, gwArgs); err != nil {
		return err
	} else if err = registerRepositoriesServiceServer(grpcSrv, pluginsServer, gwArgs); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", listenAddrCMux)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Multiplex the connection between grpc and http.
	// Note: due to a change in the grpc protocol, it's no longer possible to just match
	// on the simpler cmux.HTTP2HeaderField("content-type", "application/grpc"). More details
	// at https://github.com/soheilhy/cmux/issues/64
	mux := cmux.New(lis)

	// The non-transitioned services continue as normal for now.
	grpcListener := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	grpcWebListener := mux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc-web"))
	httpListener := mux.Match(cmux.Any())

	webRpcProxy := grpcweb.WrapServer(grpcSrv,
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithWebsockets(true),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
	)

	httpSrv := &http.Server{
		ReadHeaderTimeout: 60 * time.Second, // mitigate slowloris attacks, set to nginx's default
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if webRpcProxy.IsGrpcWebRequest(r) || webRpcProxy.IsAcceptableGrpcCorsRequest(r) || webRpcProxy.IsGrpcWebSocketRequest(r) {
				webRpcProxy.ServeHTTP(w, r)
			} else {
				gwArgs.Mux.ServeHTTP(w, r)
			}
		},
		),
	}

	go func() {
		err := grpcSrv.Serve(grpcListener)
		if err != nil {
			klogv2.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		err := grpcSrv.Serve(grpcWebListener)
		if err != nil {
			klogv2.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		err := httpSrv.Serve(httpListener)
		if err != nil {
			klogv2.Fatalf("failed to serve: %v", err)
		}
	}()
	go func() {
		if err := mux.Serve(); err != nil {
			klogv2.Fatalf("failed to serve: %v", err)
		}
	}()

	if serveOpts.UnsafeLocalDevKubeconfig {
		klogv2.Warning("Using the local Kubeconfig file instead of the actual in-cluster's config. This is not recommended except for development purposes.")
	}

	// Finally, link the new mux so that all other requests are routed to the old cmux's listen
	// address. cmux requires a listener.
	mux_connect.Handle("/", &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			// Update the port (only) of the url
			parts := strings.SplitAfter(lis.Addr().String(), ":")
			port, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				klogv2.Fatalf("unable to extract port from listen address %q", lis.Addr().String())
			}
			r.URL.Host = fmt.Sprintf("127.0.0.1:%d", port)
			klogv2.Errorf("proxied URL is: %+v", r.URL)
		},
	})

	listenPort := fmt.Sprintf(":%d", serveOpts.Port)
	klogv2.Infof("Starting server on %q", listenPort)
	if err := http.ListenAndServe(listenPort, h2c.NewHandler(mux_connect, &http2.Server{})); err != nil {
		klogv2.Fatalf("failed to server: %+v", err)
	}

	return nil
}

func registerPackagesServiceServer(grpcSrv *grpc.Server, pluginsServer *pluginsv1alpha1.PluginsServer, gwArgs core.GatewayHandlerArgs) error {
	// Ask the plugins server for plugins with GRPC servers that fulfil the core
	// packaging v1alpha1 API, then pass to the constructor below.
	// The argument for the reflect.TypeOf is based on what grpc-go
	// does itself at:
	// https://github.com/grpc/grpc-go/blob/v1.38.0/server.go#L621
	packagingPlugins := pluginsServer.GetPluginsSatisfyingInterface(reflect.TypeOf((*packagesGRPCv1alpha1.PackagesServiceServer)(nil)).Elem())

	// Create the core.packages server and register it for both grpc and http.
	packagesServer, err := packagesv1alpha1.NewPackagesServer(packagingPlugins)
	if err != nil {
		return fmt.Errorf("failed to create core.packages.v1alpha1 server: %w", err)
	}
	packagesGRPCv1alpha1.RegisterPackagesServiceServer(grpcSrv, packagesServer)
	err = packagesGRPCv1alpha1.RegisterPackagesServiceHandlerFromEndpoint(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)
	if err != nil {
		return fmt.Errorf("failed to register core.packages handler for gateway: %v", err)
	}
	return nil
}

func registerRepositoriesServiceServer(grpcSrv *grpc.Server, pluginsServer *pluginsv1alpha1.PluginsServer, gwArgs core.GatewayHandlerArgs) error {
	// see comment in registerPackagesServiceServer
	repositoriesPlugins := pluginsServer.GetPluginsSatisfyingInterface(reflect.TypeOf((*packagesGRPCv1alpha1.RepositoriesServiceServer)(nil)).Elem())

	// Create the core.packages server and register it for both grpc and http.
	repoServer, err := packagesv1alpha1.NewRepositoriesServer(repositoriesPlugins)
	if err != nil {
		return fmt.Errorf("failed to create core.packages.v1alpha1 server: %w", err)
	}
	packagesGRPCv1alpha1.RegisterRepositoriesServiceServer(grpcSrv, repoServer)
	err = packagesGRPCv1alpha1.RegisterRepositoriesServiceHandlerFromEndpoint(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)
	if err != nil {
		return fmt.Errorf("failed to register core.packages handler for gateway: %v", err)
	}
	return nil
}

// Create a gateway mux that does not emit unpopulated fields.
func gatewayMux() (*runtime.ServeMux, error) {
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
	// static 'swagger-ui' dashboard with hardcoded values just intended for development purposes.
	// This docs will eventually converge into the docs already (properly) served by the dashboard
	err := gwmux.HandlePath(http.MethodGet, "/openapi.json", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/kubeapps-apis.swagger.json")
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to serve: %v", err)
	}

	err = gwmux.HandlePath(http.MethodGet, "/docs", runtime.HandlerFunc(func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		http.ServeFile(w, r, "docs/index.html")
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to serve: %v", err)
	}

	svcRestConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve in cluster configuration: %v", err)
	}
	coreClientSet, err := kubernetes.NewForConfig(svcRestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve clientset: %v", err)
	}

	// TODO(rcastelblanq) Move this endpoint to the Operators plugin when implementing #4920
	// Proxies the operator icon request to K8s
	err = gwmux.HandlePath(http.MethodGet, "/operators/namespaces/{namespace}/operator/{name}/logo", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		namespace := pathParams["namespace"]
		name := pathParams["name"]

		logoBytes, err := coreClientSet.RESTClient().Get().AbsPath(fmt.Sprintf("/apis/packages.operators.coreos.com/v1/namespaces/%s/packagemanifests/%s/icon", namespace, name)).Do(context.TODO()).Raw()
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to retrieve operator logo: %v", err), http.StatusInternalServerError)
			return
		}

		contentType := http.DetectContentType(logoBytes)
		if strings.Contains(contentType, "text/") {
			// DetectContentType is unable to return svg icons since they are in fact text
			contentType = "image/svg+xml"
		}
		w.Header().Set("Content-Type", contentType)
		_, err = w.Write(logoBytes)
		if err != nil {
			return
		}
	})
	if err != nil {
		return nil, fmt.Errorf("failed to serve: %v", err)
	}

	return gwmux, nil
}
