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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	log "k8s.io/klog/v2"
)

const (
	pluginRootDir           = "/"
	grpcRegisterFunction    = "RegisterWithGRPCServer"
	gatewayRegisterFunction = "RegisterHTTPHandlerFromEndpoint"
	pluginDetailFunction    = "GetPluginDetail"
	clustersCAFilesPrefix   = "/etc/additional-clusters-cafiles"
)

// KubernetesClientGetter is a function type used by plugins to get a k8s client
type KubernetesClientGetter func(context.Context) (kubernetes.Interface, dynamic.Interface, error)

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

func NewPluginsServer(serveOpts ServeOptions, registrar grpc.ServiceRegistrar, gwArgs gwHandlerArgs) (*pluginsServer, error) {
	// Store the serveOptions in the global 'pluginsServeOpts' variable

	// Find all .so plugins in the specified plugins directory.
	pluginPaths, err := listSOFiles(os.DirFS(pluginRootDir), serveOpts.PluginDirs)
	if err != nil {
		log.Fatalf("failed to check for plugins: %v", err)
	}

	ps := &pluginsServer{}

	pluginDetails, err := ps.registerPlugins(pluginPaths, registrar, gwArgs, serveOpts)
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
	log.Infof("+core GetConfiguredPlugins")
	return &plugins.GetConfiguredPluginsResponse{
		Plugins: s.plugins,
	}, nil
}

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func (s *pluginsServer) registerPlugins(pluginPaths []string, grpcReg grpc.ServiceRegistrar, gwArgs gwHandlerArgs, serveOpts ServeOptions) ([]*plugins.Plugin, error) {
	pluginDetails := []*plugins.Plugin{}

	clientGetter, err := createClientGetter(serveOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to create a ClientGetter: %w", err)
	}

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

		if err = s.registerGRPC(p, pluginDetail, grpcReg, clientGetter); err != nil {
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
func (s *pluginsServer) registerGRPC(p *plugin.Plugin, pluginDetail *plugins.Plugin, registrar grpc.ServiceRegistrar, clientGetter KubernetesClientGetter) error {
	grpcRegFn, err := p.Lookup(grpcRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %v: %w", grpcRegisterFunction, pluginDetail, err)
	}
	type grpcRegisterFunctionType = func(grpc.ServiceRegistrar, KubernetesClientGetter) interface{}

	grpcFn, ok := grpcRegFn.(grpcRegisterFunctionType)
	if !ok {
		var dummyFn grpcRegisterFunctionType = func(grpc.ServiceRegistrar, KubernetesClientGetter) interface{} { return nil }
		return fmt.Errorf("unable to use %q in plugin %v due to mismatched signature.\nwant: %T\ngot: %T", grpcRegisterFunction, pluginDetail, dummyFn, grpcRegFn)
	}

	server := grpcFn(registrar, clientGetter)

	return s.registerPluginsSatisfyingCoreAPIs(server, pluginDetail)
}

// registerPluginsImplementingCoreAPIs checks a plugin implementation to see
// if it implements a core api (such as `packages.v1alpha1`) and if so,
// keeps a (typed) reference to the implementation for use on aggregate APIs.
func (s *pluginsServer) registerPluginsSatisfyingCoreAPIs(pluginSrv interface{}, pluginDetail *plugins.Plugin) error {
	// The following check if the service implements an interface is what
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
		log.Infof("Plugin %v implements core.packages.v1alpha1. Registered for aggregation.", pluginDetail)
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

// createClientGetter returns a function closure for creating the k8s client to interact with the cluster.
// The returned function utilizes the user credential present in the request context.
// The plugins just have to call this function passing the context in order to retrieve the configured k8s client
func createClientGetter(serveOpts ServeOptions) (KubernetesClientGetter, error) {
	var restConfig *rest.Config
	var clustersConfig kube.ClustersConfig
	var err error

	if !serveOpts.UnsafeLocalDevKubeconfig {
		// get the default rest inCluster config for the kube.NewClusterConfig function
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to get inClusterConfig: %w", err)
		}
	} else {
		// using the local kubeconfig instead of the inCluster config
		log.Warningf("Using the local kubeconfig configuration (in KUBECONFIG='%s' envar) since you passed --unsafe-local-dev-kubeconfig=true", os.Getenv("KUBECONFIG"))
		kubeconfigBytes, err := ioutil.ReadFile(os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, fmt.Errorf("unable to read the file in KUBECONFIG envar: %w", err)
		}
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			return nil, fmt.Errorf("unable to get local KUBECONFIG='%s' file: %w", os.Getenv("KUBECONFIG"), err)
		}
	}

	if !serveOpts.UnsafeUseDemoSA {
		// get the parsed kube.ClustersConfig from the serveOpts
		clustersConfig, err = getClustersConfigFromServeOpts(serveOpts)
		if err != nil {
			return nil, err
		}
	} else {
		// Just using the created SA, no user account nor clustersConfig is used here
		clustersConfig = kube.ClustersConfig{}
	}

	// return the closure fuction that takes the context, but preserving the required scope,
	// 'inClusterConfig' and 'config'
	return createClientGetterWithParams(restConfig, serveOpts, clustersConfig)
}

// createClientGetter takes the required params and returns the closure fuction.
// it's splitted for testing this fn separately
func createClientGetterWithParams(inClusterConfig *rest.Config, serveOpts ServeOptions, clustersConfig kube.ClustersConfig) (KubernetesClientGetter, error) {

	// return the closure fuction that takes the context, but preserving the required scope,
	// 'inClusterConfig' and 'config'
	return func(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		var err error
		token, err := extractToken(ctx)
		if err != nil {
			return nil, nil, status.Errorf(codes.Unauthenticated, "invalid authorization metadata: %v", err)
		}

		var config *rest.Config
		if !serveOpts.UnsafeUseDemoSA {
			// We are using the KubeappsClusterName, but if the endpoint was cluster-scoped,
			// we should pass the cluster name instead
			config, err = kube.NewClusterConfig(inClusterConfig, token, clustersConfig.KubeappsClusterName, clustersConfig)
			if err != nil {
				return nil, nil, fmt.Errorf("unable to get clusterConfig: %w", err)
			}
		} else {
			// Just using the created SA, no user account is used
			config = inClusterConfig
		}
		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create dynamic client: %w", err)
		}
		typedClient, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to create typed client: %w", err)
		}
		return typedClient, dynamicClient, nil
	}, nil
}

// extractToken returns the token passed through the gRPC request in the "authorization" metadata in the context
// It is equivalent to the "Authorization" usual HTTP 1 header
// For instance: authorization="Bearer abc" will return "abc"
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("error reading request metadata/headers")
	}

	// metadata is always lowercased
	if len(md["authorization"]) > 0 {
		if strings.HasPrefix(md["authorization"][0], "Bearer ") {
			return strings.TrimPrefix(md["authorization"][0], "Bearer "), nil
		} else {
			return "", fmt.Errorf("malformed authorization metadata")
		}
	} else {
		// No authorization header found, no error here, we will delegate it to the RBAC
		return "", nil
	}
}

// getClustersConfigFromServeOpts get the serveOptions and calls parseClusterConfig with the proper values
// returning a kube.ClustersConfig
func getClustersConfigFromServeOpts(serveOpts ServeOptions) (kube.ClustersConfig, error) {
	if serveOpts.ClustersConfigPath == "" {
		return kube.ClustersConfig{}, fmt.Errorf("unable to parse clusters config, no config path passed")
	}
	var cleanupCAFiles func()
	config, cleanupCAFiles, err := kube.ParseClusterConfig(serveOpts.ClustersConfigPath, clustersCAFilesPrefix, serveOpts.PinnipedProxyURL)
	if err != nil {
		return kube.ClustersConfig{}, fmt.Errorf("unable to parse additional clusters config: %+v", err)
	}
	defer cleanupCAFiles()
	return config, nil
}
