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
package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"plugin"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/core"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	log "k8s.io/klog/v2"
)

const (
	pluginRootDir           = "/"
	grpcRegisterFunction    = "RegisterWithGRPCServer"
	gatewayRegisterFunction = "RegisterHTTPHandlerFromEndpoint"
)

// TODO:
// * Implement plugins available
// * Add proto for creating a deployed package or displaying an installed app.
//   - identifier for the package (may include repository for some formats?)
//   - yaml string for
//
// server is used to implement kubeapps.Packages

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(port int, pluginDirs []string) {
	// Create the grpc server, register the standard services (reflection and our core service).
	grpcSrv := grpc.NewServer()
	reflection.Register(grpcSrv)
	core.RegisterCoreServer(grpcSrv, &coreServer{})

	// Create the http server, register our core service followed by any plugins.
	listenAddr := fmt.Sprintf(":%d", port)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	gwArgs := gwHandlerArgs{
		ctx:         ctx,
		mux:         runtime.NewServeMux(),
		addr:        listenAddr,
		dialOptions: []grpc.DialOption{grpc.WithInsecure()},
	}
	err := core.RegisterCoreHandlerFromEndpoint(gwArgs.ctx, gwArgs.mux, gwArgs.addr, gwArgs.dialOptions)
	if err != nil {
		log.Fatalf("Failed to register core handler for gateway: %v", err)
	}
	httpSrv := &http.Server{
		Handler: gwArgs.mux,
	}

	// Find and register the plugins both for gRPC and the http gateway.
	plugins, err := listSOFiles(os.DirFS(pluginRootDir), pluginDirs)
	if err != nil {
		log.Fatalf("failed to check for plugins: %v", err)
	}
	err = registerPlugins(plugins, grpcSrv, gwArgs)
	if err != nil {
		log.Fatalf("failed to register plugins: %v", err)
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
	httpLis := mux.Match(cmux.Any())

	// TODO: handle errors, perhaps graceful shutdowns etc.
	go grpcSrv.Serve(grpcLis)
	go httpSrv.Serve(httpLis)

	log.Infof("Starting server on :%d", port)
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

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func registerPlugins(pluginPaths []string, grpcReg grpc.ServiceRegistrar, gwArgs gwHandlerArgs) error {
	for _, pluginPath := range pluginPaths {
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return fmt.Errorf("unable to open plugin %q: %w", pluginPath, err)
		}

		if err = registerGRPC(p, pluginPath, grpcReg); err != nil {
			return err
		}

		if err = registerHTTP(p, pluginPath, gwArgs); err != nil {
			return err
		}

		log.Infof("Successfully registered plugin %q", pluginPath)
	}
	return nil
}

// registerGRPC finds and calls the required function for registering the plugin for the GRPC server.
func registerGRPC(p *plugin.Plugin, pluginPath string, registrar grpc.ServiceRegistrar) error {
	grpcRegFn, err := p.Lookup(grpcRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %q: %w", grpcRegisterFunction, pluginPath, err)
	}
	type grpcRegisterFunctionType = func(grpc.ServiceRegistrar)
	grpcFn, ok := grpcRegFn.(grpcRegisterFunctionType)
	if !ok {
		var dummyFn grpcRegisterFunctionType = func(grpc.ServiceRegistrar) {}
		return fmt.Errorf("unable to use %q in plugin %q due to mismatched signature.\nwant: %T\ngot: %T", grpcRegisterFunction, pluginPath, dummyFn, grpcRegFn)
	}
	grpcFn(registrar)
	return nil
}

// registerHTTP finds and calls the required function for registering the plugin for the HTTP gateway server.
func registerHTTP(p *plugin.Plugin, pluginPath string, gwArgs gwHandlerArgs) error {
	gwRegFn, err := p.Lookup(gatewayRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %q: %w", gatewayRegisterFunction, pluginPath, err)
	}
	type gatewayRegisterFunctionType = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	gwfn, ok := gwRegFn.(gatewayRegisterFunctionType)
	if !ok {
		// Create a dummyFn only so we can ensure the correct type is shown in case
		// of an error.
		var dummyFn gatewayRegisterFunctionType = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error { return nil }
		return fmt.Errorf("unable to use %q in plugin %q due to mismatched signature.\nwant: %T\ngot: %T", gatewayRegisterFunction, pluginPath, dummyFn, gwRegFn)
	}
	return gwfn(gwArgs.ctx, gwArgs.mux, gwArgs.addr, gwArgs.dialOptions)
}

// listSOFiles returns the absolute paths of all .so files found in any of the provided plugin directories.
//
// pluginDirs can be relative to the current directory or absolute.
func listSOFiles(fsys fs.FS, pluginDirs []string) ([]string, error) {
	matches := []string{}

	for _, pluginDir := range pluginDirs {
		if !filepath.IsAbs(pluginDir) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, err
			}
			pluginDir = filepath.Join(cwd, pluginDir)
		}
		relPluginDir, err := filepath.Rel(pluginRootDir, pluginDir)
		if err != nil {
			return nil, err
		}

		m, err := fs.Glob(fsys, path.Join(relPluginDir, "/", "*.so"))
		if err != nil {
			return nil, err
		}

		for _, match := range m {
			matches = append(matches, filepath.Join(pluginRootDir, match))
		}
	}
	return matches, nil
}
