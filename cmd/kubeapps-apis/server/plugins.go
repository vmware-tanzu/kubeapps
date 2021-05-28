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
	"os"
	"path"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	// packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

	"google.golang.org/grpc"
	log "k8s.io/klog/v2"
)

// Add plugin message to proto. then include in struct here with client?
// How to create client dynamically? Perhaps return client when registering each?
// Perhaps define an interface that client must implement (based on server?)

const (
	pluginRootDir           = "/"
	grpcRegisterFunction    = "RegisterWithGRPCServer"
	gatewayRegisterFunction = "RegisterHTTPHandlerFromEndpoint"
	pluginDetailFunction    = "GetPluginDetail"
)

// pkgsPluginWithServer stores the plugin detail together with its implementation.
type pkgsPluginWithServer struct {
	plugin *plugins.Plugin
	server packages.PackagesServiceServer
}

// coreServer implements the API defined in cmd/kubeapps-api-service/core/core.proto
type pluginsServer struct {
	plugins.UnimplementedPluginsServiceServer

	// The slice of plugins is initialised when registering plugins during NewPluginsServer.
	plugins []*plugins.Plugin

	// packagesPlugins contains plugin server implementations which satisfy
	// the core server packages.v1alpha1 interface.
	// TODO: Update the plugins server to be able to register different versions
	// of core plugins.
	packagesPlugins []*pkgsPluginWithServer
}

func NewPluginsServer(pluginDirs []string, registrar grpc.ServiceRegistrar, gwArgs gwHandlerArgs) (*pluginsServer, error) {
	// Find all .so plugins in the specified plugins directory.
	pluginPaths, err := listSOFiles(os.DirFS(pluginRootDir), pluginDirs)
	if err != nil {
		log.Fatalf("failed to check for plugins: %v", err)
	}

	ps := &pluginsServer{}

	pluginDetails, err := ps.registerPlugins(pluginPaths, registrar, gwArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugins: %w", err)
	}

	sortPlugins(pluginDetails)

	ps.plugins = pluginDetails

	return ps, nil
}

// sortPlugins returns a consistently ordered slice.
func sortPlugins(p []*plugins.Plugin) {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Name < p[j].Name || (p[i].Name == p[j].Name && p[i].Version < p[j].Version)
	})
}

// GetConfiguredPlugins returns details for each configured plugin.
func (s *pluginsServer) GetConfiguredPlugins(ctx context.Context, in *plugins.GetConfiguredPluginsRequest) (*plugins.GetConfiguredPluginsResponse, error) {
	return &plugins.GetConfiguredPluginsResponse{
		Plugins: s.plugins,
	}, nil
}

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func (s *pluginsServer) registerPlugins(pluginPaths []string, grpcReg grpc.ServiceRegistrar, gwArgs gwHandlerArgs) ([]*plugins.Plugin, error) {
	pluginDetails := []*plugins.Plugin{}
	for _, pluginPath := range pluginPaths {
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("unable to open plugin %q: %w", pluginPath, err)
		}

		var pluginDetail *plugins.Plugin
		if pluginDetail, err = getPluginDetail(p, pluginPath); err != nil {
			return nil, err
		} else {
			pluginDetails = append(pluginDetails, pluginDetail)
		}

		if err = s.registerGRPC(p, pluginDetail, grpcReg); err != nil {
			return nil, err
		}

		if err = registerHTTP(p, pluginDetail, gwArgs); err != nil {
			return nil, err
		}

		log.Infof("Successfully registered plugin %q", pluginPath)
	}
	return pluginDetails, nil
}

// registerGRPC finds and calls the required function for registering the plugin for the GRPC server.
func (s *pluginsServer) registerGRPC(p *plugin.Plugin, pluginDetail *plugins.Plugin, registrar grpc.ServiceRegistrar) error {
	grpcRegFn, err := p.Lookup(grpcRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %v: %w", grpcRegisterFunction, pluginDetail, err)
	}
	type grpcRegisterFunctionType = func(grpc.ServiceRegistrar) interface{}
	grpcFn, ok := grpcRegFn.(grpcRegisterFunctionType)
	if !ok {
		var dummyFn grpcRegisterFunctionType = func(grpc.ServiceRegistrar) interface{} { return nil }
		return fmt.Errorf("unable to use %q in plugin %v due to mismatched signature.\nwant: %T\ngot: %T", grpcRegisterFunction, pluginDetail, dummyFn, grpcRegFn)
	}
	server := grpcFn(registrar)

	return s.registerPluginsSatisfyingCoreAPIs(server, pluginDetail)
}

// registerPluginsImplementingCoreAPIs checks a plugin implementation to see
// if it implements a core api (such as `packages.v1alpha1`) and if so,
// keeps a (typed) reference to the implementation for use on aggregate APIs.
func (s *pluginsServer) registerPluginsSatisfyingCoreAPIs(pluginSrv interface{}, pluginDetail *plugins.Plugin) error {
	// The following heck if the service implements an interface is what
	// grpc-go itself does, see:
	// https://github.com/grpc/grpc-go/blob/v1.38.0/server.go#L621
	serverType := reflect.TypeOf(pluginSrv)
	corePackagesType := reflect.TypeOf((*packages.PackagesServiceServer)(nil)).Elem()

	if serverType.Implements(corePackagesType) {
		pkgsSrv, ok := pluginSrv.(packages.PackagesServiceServer)
		if !ok {
			return fmt.Errorf("Unable to convert plugin %v to core PackagesServicesServer although it implements the same.", pluginDetail)
		}
		s.packagesPlugins = append(s.packagesPlugins, &pkgsPluginWithServer{
			plugin: pluginDetail,
			server: pkgsSrv,
		})
		log.Errorf("Plugin %v implements core.packages.v1alpha1. Registered for aggregation.", pluginDetail)
	}
	return nil
}

// getPluginDetail returns a core.plugins.Plugin as defined by the plugin itself.
func getPluginDetail(p *plugin.Plugin, pluginPath string) (*plugins.Plugin, error) {
	pluginDetailFn, err := p.Lookup(pluginDetailFunction)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup %q for %q: %w", pluginDetailFunction, pluginPath, err)
	}

	type pluginDetailFunctionType = func() *plugins.Plugin

	fn, ok := pluginDetailFn.(pluginDetailFunctionType)
	if !ok {
		var dummyFn pluginDetailFunctionType = func() *plugins.Plugin { return &plugins.Plugin{} }
		return nil, fmt.Errorf("unable to use %q in plugin %q due to a mismatched signature. \nwant: %T\ngot: %T", pluginDetailFunction, pluginPath, dummyFn, pluginDetailFn)
	}

	return fn(), nil
}

// registerHTTP finds and calls the required function for registering the plugin for the HTTP gateway server.
func registerHTTP(p *plugin.Plugin, pluginDetail *plugins.Plugin, gwArgs gwHandlerArgs) error {
	gwRegFn, err := p.Lookup(gatewayRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %v: %w", gatewayRegisterFunction, pluginDetail, err)
	}
	type gatewayRegisterFunctionType = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	gwfn, ok := gwRegFn.(gatewayRegisterFunctionType)
	if !ok {
		// Create a dummyFn only so we can ensure the correct type is shown in case
		// of an error.
		var dummyFn gatewayRegisterFunctionType = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error { return nil }
		return fmt.Errorf("unable to use %q in plugin %v due to mismatched signature.\nwant: %T\ngot: %T", gatewayRegisterFunction, pluginDetail, dummyFn, gwRegFn)
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
