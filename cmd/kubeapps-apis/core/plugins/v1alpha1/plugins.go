// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"reflect"
	"sort"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"google.golang.org/grpc"
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

// GRPCPluginRegistrationOptions defines the single argument that
// a plugin's RegisterWithGRPCServer function must accept. This allows
// the arguments to be defined (or modified) in the one place.
type GRPCPluginRegistrationOptions struct {
	ConfigGetter     core.KubernetesConfigGetter
	ClustersConfig   kube.ClustersConfig
	PluginConfigPath string
	// The QPS and Burst options that have been configured for any
	// clients of the K8s API server created by plugins.
	ClientQPS   float32
	ClientBurst int

	// The mux used for the connect gRPC routing
	Mux *http.ServeMux

	LocalPort int
}

// PluginWithServer keeps a record of a GRPC server and its plugin detail.
type PluginWithServer struct {
	Plugin *plugins.Plugin
	Server interface{}
}

// PluginsServer implements the API defined in "plugins.proto"
type PluginsServer struct {
	plugins.UnimplementedPluginsServiceServer

	// The slice of pluginsWithServers is initialised when registering pluginsWithServers during NewPluginsServer.
	pluginsWithServers []PluginWithServer

	// The parsed config for clusters in a multi-cluster setup.
	clustersConfig kube.ClustersConfig
}

func NewPluginsServer(serveOpts core.ServeOptions, gwArgs core.GatewayHandlerArgs, mux *http.ServeMux) (*PluginsServer, error) {
	// Store the serveOptions in the global 'pluginsServeOpts' variable

	// Find all .so plugins in the specified plugins directory.
	pluginPaths, err := listSOFiles(os.DirFS(pluginRootDir), serveOpts.PluginDirs)
	if err != nil {
		log.Fatalf("Failed to check for plugins: %v", err)
	}

	ps := &PluginsServer{}

	// get the parsed kube.ClustersConfig from the serveOpts
	clustersConfig, err := getClustersConfigFromServeOpts(serveOpts)
	if err != nil {
		return nil, err
	}
	ps.clustersConfig = clustersConfig

	err = ps.registerPlugins(pluginPaths, gwArgs, serveOpts, mux)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugins: %w", err)
	}

	return ps, nil
}

// sortPlugins returns a consistently ordered slice.
func sortPlugins(p []PluginWithServer) {
	sort.Slice(p, func(i, j int) bool { return ComparePlugin(p[i].Plugin, p[j].Plugin) })
}

func ComparePlugin(pluginA *plugins.Plugin, pluginB *plugins.Plugin) bool {
	return pluginA.Name < pluginB.Name || (pluginA.Name == pluginB.Name && pluginA.Version < pluginB.Version)
}

// GetConfiguredPlugins returns details for each configured plugin.
func (s *PluginsServer) GetConfiguredPlugins(ctx context.Context, in *connect.Request[plugins.GetConfiguredPluginsRequest]) (*connect.Response[plugins.GetConfiguredPluginsResponse], error) {
	// this gets logged twice (liveness and readiness checks) every 10 seconds and
	// really adds a lot of noise to the logs, so lowering verbosity
	log.V(4).Infof("+core GetConfiguredPlugins")
	pluginDetails := make([]*plugins.Plugin, len(s.pluginsWithServers))
	for i, p := range s.pluginsWithServers {
		pluginDetails[i] = p.Plugin
	}
	return connect.NewResponse(&plugins.GetConfiguredPluginsResponse{
		Plugins: pluginDetails,
	}), nil
}

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func (s *PluginsServer) registerPlugins(pluginPaths []string, gwArgs core.GatewayHandlerArgs, serveOpts core.ServeOptions, mux *http.ServeMux) error {
	pluginsWithServers := []PluginWithServer{}

	configGetter, err := createConfigGetter(serveOpts, s.clustersConfig)
	if err != nil {
		return fmt.Errorf("unable to create a ClientGetter: %w", err)
	}

	for _, pluginPath := range pluginPaths {
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return fmt.Errorf("unable to open plugin %q: %w", pluginPath, err)
		}

		var pluginDetail *plugins.Plugin
		if pluginDetail, err = getPluginDetail(p, pluginPath); err != nil {
			return err
		}

		if grpcServer, err := s.registerGRPC(p, pluginDetail, configGetter, serveOpts, mux); err != nil {
			return err
		} else {
			pluginsWithServers = append(pluginsWithServers, PluginWithServer{
				Plugin: pluginDetail,
				Server: grpcServer,
			})
		}

		if err = registerHTTP(p, pluginDetail, gwArgs); err != nil {
			return err
		}

		log.InfoS("Successfully registered plugin", "pluginPath", pluginPath)
	}

	sortPlugins(pluginsWithServers)

	s.pluginsWithServers = pluginsWithServers

	return nil
}

// registerGRPC finds and calls the required function for registering the plugin for the GRPC server.
func (s *PluginsServer) registerGRPC(p *plugin.Plugin, pluginDetail *plugins.Plugin, configGetter core.KubernetesConfigGetter, serveOpts core.ServeOptions, mux *http.ServeMux) (interface{}, error) {
	grpcRegFn, err := p.Lookup(grpcRegisterFunction)
	if err != nil {
		return nil, fmt.Errorf("unable to lookup %q for %v: %w", grpcRegisterFunction, pluginDetail, err)
	}
	type grpcRegisterFunctionType = func(GRPCPluginRegistrationOptions) (interface{}, error)

	grpcFn, ok := grpcRegFn.(grpcRegisterFunctionType)
	if !ok {
		var stubFn = func(GRPCPluginRegistrationOptions) (interface{}, error) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to use %q in plugin %v due to mismatched signature.\nwant: %T\ngot: %T", grpcRegisterFunction, pluginDetail, stubFn, grpcRegFn)
	}

	server, err := grpcFn(GRPCPluginRegistrationOptions{
		ConfigGetter:     configGetter,
		ClustersConfig:   s.clustersConfig,
		PluginConfigPath: serveOpts.PluginConfigPath,
		ClientQPS:        serveOpts.QPS,
		ClientBurst:      serveOpts.Burst,
		Mux:              mux,
		LocalPort:        serveOpts.Port,
	})
	if err != nil {
		return nil, fmt.Errorf("plug-in %q failed to register due to: %v", pluginDetail, err)
	} else if server == nil {
		return nil, fmt.Errorf("registration for plug-in %v failed due to: %T returned nil when non-nil value was expected", pluginDetail, grpcFn)
	}

	return server, nil
}

// GetPluginsSatisfyingInterface returns the registered plugins which satisfy a
// particular interface. Currently this is used to return the plugins that satisfy
// the core.packaging interface for the core packaging server.
func (s *PluginsServer) GetPluginsSatisfyingInterface(targetInterface reflect.Type) []PluginWithServer {
	satisfiedPlugins := []PluginWithServer{}
	for _, pluginSrv := range s.pluginsWithServers {
		// The following check if the service implements an interface is what
		// grpc-go itself does, see:
		// https://github.com/grpc/grpc-go/blob/v1.38.0/server.go#L621
		serverType := reflect.TypeOf(pluginSrv.Server)

		if serverType.Implements(targetInterface) {
			satisfiedPlugins = append(satisfiedPlugins, pluginSrv)
		}
	}
	return satisfiedPlugins
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
		var stubFn = func() *plugins.Plugin { return &plugins.Plugin{} }
		return nil, fmt.Errorf("unable to use %q in plugin %q due to a mismatched signature. \nwant: %T\ngot: %T", pluginDetailFunction, pluginPath, stubFn, pluginDetailFn)
	}

	return fn(), nil
}

// registerHTTP finds and calls the required function for registering the plugin for the HTTP gateway server.
func registerHTTP(p *plugin.Plugin, pluginDetail *plugins.Plugin, gwArgs core.GatewayHandlerArgs) error {
	gwRegFn, err := p.Lookup(gatewayRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %v: %w", gatewayRegisterFunction, pluginDetail, err)
	}
	type gatewayRegisterFunctionType = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error
	gwfn, ok := gwRegFn.(gatewayRegisterFunctionType)
	if !ok {
		// Create a stubFn only so we can ensure the correct type is shown in case
		// of an error.
		var stubFn = func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error { return nil }
		return fmt.Errorf("unable to use %q in plugin %v due to mismatched signature.\nwant: %T\ngot: %T", gatewayRegisterFunction, pluginDetail, stubFn, gwRegFn)
	}
	return gwfn(gwArgs.Ctx, gwArgs.Mux, gwArgs.Addr, gwArgs.DialOptions)
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

// createConfigGetter returns a function closure for creating the k8s config to interact with the cluster.
// The returned function utilizes the user credential present in the request context.
// The plugins just have to call this function passing the context in order to retrieve the configured k8s client
func createConfigGetter(serveOpts core.ServeOptions, clustersConfig kube.ClustersConfig) (core.KubernetesConfigGetter, error) {
	var restConfig *rest.Config
	var err error

	if serveOpts.UnsafeLocalDevKubeconfig {
		// if using the local kubeconfig, read it from the KUBECONFIG path and
		// create the restConfig
		log.Warningf("Using the local kubeconfig configuration (in KUBECONFIG='%s' envar) since you passed --unsafe-local-dev-kubeconfig=true", os.Getenv("KUBECONFIG"))
		kubeconfigBytes, err := os.ReadFile(os.Getenv("KUBECONFIG"))
		if err != nil {
			return nil, fmt.Errorf("unable to read the file in KUBECONFIG envar: %w", err)
		}
		restConfig, err = clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
		if err != nil {
			return nil, fmt.Errorf("unable to get local KUBECONFIG='%s' file: %w", os.Getenv("KUBECONFIG"), err)
		}
	} else {
		// otherwise, get the default rest inCluster config for the kube.NewClusterConfig function
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to get inClusterConfig: %w", err)
		}
	}

	// return the closure function that takes the context, but preserving the required scope,
	// 'inClusterConfig' and 'config'
	return createConfigGetterWithParams(restConfig, serveOpts, clustersConfig)
}

// createClientGetter takes the required params and returns the closure function.
// it's split for testing this fn separately
func createConfigGetterWithParams(inClusterConfig *rest.Config, serveOpts core.ServeOptions, clustersConfig kube.ClustersConfig) (core.KubernetesConfigGetter, error) {
	// return the closure function that takes the context, but preserving the required scope,
	// 'inClusterConfig' and 'config'
	return func(headers http.Header, cluster string) (*rest.Config, error) {
		log.V(4).Infof("+clientGetter.GetClient")
		var err error
		token, err := extractToken(headers)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("Invalid authorization metadata: %w", err))
		}

		var config *rest.Config

		// Enable existing plugins to pass an empty cluster name to get the
		// kubeapps cluster
		if kube.IsKubeappsClusterRef(cluster) {
			cluster = clustersConfig.KubeappsClusterName
		}

		config, err = kube.NewClusterConfig(inClusterConfig, token, cluster, clustersConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to get clusterConfig: %w", err)
		}

		if serveOpts.QPS > 0.0 {
			config.QPS = serveOpts.QPS
		}

		if serveOpts.Burst > 0 {
			config.Burst = serveOpts.Burst
		}

		return config, nil
	}, nil
}

// extractToken returns the token passed through the gRPC request in the
// "authorization" metadata, in the headers for connect gRPC.
// For instance: authorization="Bearer abc" will return "abc"
func extractToken(headers http.Header) (string, error) {
	bearerToken := headers.Get("Authorization")

	if len(bearerToken) > 0 {
		if strings.HasPrefix(bearerToken, "Bearer ") {
			return strings.TrimPrefix(bearerToken, "Bearer "), nil
		} else {
			return "", fmt.Errorf("malformed authorization metadata")
		}
	} else {
		// No authorization header found, see comment above
		return "", fmt.Errorf("missing authorization metadata")
	}
}

// getClustersConfigFromServeOpts get the serveOptions and calls parseClusterConfig with the proper values
// returning a kube.ClustersConfig
func getClustersConfigFromServeOpts(serveOpts core.ServeOptions) (kube.ClustersConfig, error) {
	if serveOpts.ClustersConfigPath == "" {
		if serveOpts.UnsafeLocalDevKubeconfig {
			// if using a local kubeconfig (dev purposes), this ClusterConfig file is not strictly required
			return kube.ClustersConfig{}, nil
		} else {
			return kube.ClustersConfig{}, fmt.Errorf("unable to parse clusters config, no config path passed")
		}
	}

	var cleanupCAFiles func()
	config, cleanupCAFiles, err := kube.ParseClusterConfig(serveOpts.ClustersConfigPath, clustersCAFilesPrefix, serveOpts.PinnipedProxyURL, serveOpts.PinnipedProxyCACert)
	if err != nil {
		return kube.ClustersConfig{}, fmt.Errorf("unable to parse additional clusters config: %+v", err)
	}
	config.GlobalPackagingNamespace = serveOpts.GlobalHelmReposNamespace
	defer cleanupCAFiles()
	return config, nil
}
