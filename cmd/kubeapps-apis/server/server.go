// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/net/http2/hpack"
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
	pluginsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
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
	listenAddr := fmt.Sprintf(":%d", serveOpts.Port)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gw, err := gatewayMux()
	if err != nil {
		return fmt.Errorf("failed to create gateway: %v", err)
	}
	gwArgs := core.GatewayHandlerArgs{
		Ctx:         ctx,
		Mux:         gw,
		Addr:        listenAddr,
		DialOptions: []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	}

	// Create the core.plugins.v1alpha1 server which handles registration of
	// plugins, and register it for both grpc and http.

	pluginsServer, err := pluginsv1alpha1.NewPluginsServer(serveOpts, grpcSrv, gwArgs)
	if err != nil {
		return fmt.Errorf("failed to initialize plugins server: %v", err)
	}

	// The connect service handler automatically handles grpc-web, connect and
	// grpc for us, so we won't need all the extra code below once all services
	// have been transitioned to the new mux (and we can remove the use of cmux
	// once connect is used for all requests).

	// For now, we collect all the gRPC paths used by the services that have
	// been transitioned, and use those to determine which handler to use.
	paths_for_connect_grpc := []string{}
	plugins_path, handler := pluginsConnect.NewPluginsServiceHandler(pluginsServer)
	paths_for_connect_grpc = append(paths_for_connect_grpc, plugins_path)

	mux_connect := http.NewServeMux()
	mux_connect.Handle(plugins_path, handler)

	// The gRPC Health checker reports on all connected services.
	checker := grpchealth.NewStaticChecker(
		pluginsConnect.PluginsServiceName,
	)
	checker_path, handler := grpchealth.NewHandler(checker)
	mux_connect.Handle(checker_path, handler)
	paths_for_connect_grpc = append(paths_for_connect_grpc, checker_path)

	if err = registerPackagesServiceServer(grpcSrv, pluginsServer, gwArgs); err != nil {
		return err
	} else if err = registerRepositoriesServiceServer(grpcSrv, pluginsServer, gwArgs); err != nil {
		return err
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

	// Want to match anything (grpc, connect, grpc-web that is for the
	// transitioned plugins/handlers)
	connectListener := mux.MatchWithWriters(match_transitioned_paths(paths_for_connect_grpc))

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
		err := http.Serve(connectListener, h2c.NewHandler(mux_connect, &http2.Server{}))
		if err != nil {
			klogv2.Fatalf("failed to server: %+v", err)
		}
	}()

	if serveOpts.UnsafeLocalDevKubeconfig {
		klogv2.Warning("Using the local Kubeconfig file instead of the actual in-cluster's config. This is not recommended except for development purposes.")
	}

	klogv2.Infof("Starting server on :%d", serveOpts.Port)
	if err := mux.Serve(); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func registerPluginsServiceServer(gwArgs core.GatewayHandlerArgs) error {
	err := pluginsGRPCv1alpha1.RegisterPluginsServiceHandlerFromEndpoint(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)

	if err != nil {
		return fmt.Errorf("failed to register core.plugins handler for gateway: %v", err)
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

// match_transitioned_paths is a mux that matches if an http2 request path
// matches any of the configured paths.
//
// This can be removed once all paths are transitioned to the connect mux.
func match_transitioned_paths(paths []string) cmux.MatchWriter {
	return func(w io.Writer, r io.Reader) bool {
		if !has_HTTP2_preface(r) {
			return false
		}

		done := false
		matched := false

		framer := http2.NewFramer(w, r)

		// We're only interested in the :path header.
		hdec := hpack.NewDecoder(uint32(4<<10), func(hf hpack.HeaderField) {
			if hf.Name == ":path" {
				done = true
				klogv2.Errorf("Found :path=%q", hf.Value)
				for _, path := range paths {
					if strings.HasPrefix(hf.Value, path) {
						matched = true
						klogv2.Errorf("Matched a transitioned path")
					}
				}
			}
		})

		for {
			f, err := framer.ReadFrame()
			if err != nil {
				return false
			}

			switch f := f.(type) {
			case *http2.SettingsFrame:
				// Sender acknoweldged the SETTINGS frame. No need to write
				// SETTINGS again.
				if f.IsAck() {
					break
				}
				if err := framer.WriteSettings(); err != nil {
					return false
				}
			case *http2.ContinuationFrame:
				if _, err := hdec.Write(f.HeaderBlockFragment()); err != nil {
					return false
				}
				done = done || f.FrameHeader.Flags&http2.FlagHeadersEndHeaders != 0
			case *http2.HeadersFrame:
				if _, err := hdec.Write(f.HeaderBlockFragment()); err != nil {
					return false
				}
				done = done || f.FrameHeader.Flags&http2.FlagHeadersEndHeaders != 0
			}

			if done {
				return matched
			}
		}
	}
}

// has_HTTP2_preface returns true if the request includes an http2 preface
func has_HTTP2_preface(r io.Reader) bool {
	var b [len(http2.ClientPreface)]byte
	last := 0

	for {
		n, err := r.Read(b[last:])
		if err != nil {
			return false
		}

		last += n
		eq := string(b[:last]) == http2.ClientPreface[:last]
		if last == len(http2.ClientPreface) {
			return eq
		}
		if !eq {
			return false
		}
	}
}
