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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"sort"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/kube"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
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
	clustersCAFilesPrefix   = "/etc/additional-clusters-cafiles"
)

var (
	clustersConfigPath string
	pinnipedProxyURL   string
	//temporary flag while this component in under heavy development
	unsafeUseDemoSA bool
)

// coreServer implements the API defined in cmd/kubeapps-api-service/core/core.proto
type pluginsServer struct {
	plugins.UnimplementedPluginsServiceServer

	// The slice of plugins is initialised when registering plugins during NewPluginsServer.
	plugins []*plugins.Plugin
}

func NewPluginsServer(serveOpts ServeOptions, registrar grpc.ServiceRegistrar, gwArgs gwHandlerArgs) (*pluginsServer, error) {
	// Find all .so plugins in the specified plugins directory.
	pluginPaths, err := listSOFiles(os.DirFS(pluginRootDir), serveOpts.PluginDirs)
	if err != nil {
		log.Fatalf("failed to check for plugins: %v", err)
	}

	pluginDetails, err := registerPlugins(pluginPaths, serveOpts, registrar, gwArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to register plugins: %w", err)
	}

	sortPlugins(pluginDetails)

	return &pluginsServer{
		plugins: pluginDetails,
	}, nil
}

// sortPlugins returns a consistently ordered slice.
func sortPlugins(p []*plugins.Plugin) {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Name < p[j].Name || (p[i].Name == p[j].Name && p[i].Version < p[j].Version)
	})
}

// GetConfiguredPlugins returns details for each configured plugin.
func (s *pluginsServer) GetConfiguredPlugins(ctx context.Context, in *plugins.GetConfiguredPluginsRequest) (*plugins.GetConfiguredPluginsResponse, error) {
	log.Infof("+GetConfiguredPlugins")
	return &plugins.GetConfiguredPluginsResponse{
		Plugins: s.plugins,
	}, nil
}

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func registerPlugins(pluginPaths []string, serveOpts ServeOptions, grpcReg grpc.ServiceRegistrar, gwArgs gwHandlerArgs) ([]*plugins.Plugin, error) {
	pluginDetails := []*plugins.Plugin{}
	for _, pluginPath := range pluginPaths {
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return nil, fmt.Errorf("unable to open plugin %q: %w", pluginPath, err)
		}

		if err = registerGRPC(p, pluginPath, serveOpts, grpcReg); err != nil {
			return nil, err
		}

		if err = registerHTTP(p, pluginPath, gwArgs); err != nil {
			return nil, err
		}

		if pluginDetail, err := getPluginDetail(p, pluginPath); err != nil {
			return nil, err
		} else {
			pluginDetails = append(pluginDetails, pluginDetail)
		}

		log.Infof("Successfully registered plugin %q", pluginPath)
	}
	return pluginDetails, nil
}

// registerGRPC finds and calls the required function for registering the plugin for the GRPC server.
func registerGRPC(p *plugin.Plugin, pluginPath string, serveOpts ServeOptions, registrar grpc.ServiceRegistrar) error {
	grpcRegFn, err := p.Lookup(grpcRegisterFunction)
	if err != nil {
		return fmt.Errorf("unable to lookup %q for %q: %w", grpcRegisterFunction, pluginPath, err)
	}
	type grpcRegisterFunctionType = func(grpc.ServiceRegistrar, func(context.Context) (dynamic.Interface, error))
	grpcFn, ok := grpcRegFn.(grpcRegisterFunctionType)
	if !ok {
		var dummyFn grpcRegisterFunctionType = func(grpc.ServiceRegistrar, func(context.Context) (dynamic.Interface, error)) {}
		return fmt.Errorf("unable to use %q in plugin %q due to mismatched signature.\nwant: %T\ngot: %T", grpcRegisterFunction, pluginPath, dummyFn, grpcRegFn)
	}

	// setting these globals vars so tht they are accesible by the 'dynClientGetterForContext' function
	clustersConfigPath = serveOpts.ClustersConfigPath
	pinnipedProxyURL = serveOpts.PinnipedProxyURL
	unsafeUseDemoSA = serveOpts.UnsafeUseDemoSA

	grpcFn(registrar, dynClientGetterForContext)

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

// parseClusterConfig returns a kube.ClustersConfig struct after parsing the raw `clusters` object provided by the user
// TODO(agamez): this fn is the same as in kubeapps/cmd/kubeops/main.go, export it and use it instead
func parseClusterConfig(configPath, caFilesPrefix string, pinnipedProxyURL string) (kube.ClustersConfig, func(), error) {
	caFilesDir, err := ioutil.TempDir(caFilesPrefix, "")
	if err != nil {
		return kube.ClustersConfig{}, func() {}, err
	}
	deferFn := func() { os.RemoveAll(caFilesDir) }
	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return kube.ClustersConfig{}, deferFn, err
	}

	var clusterConfigs []kube.ClusterConfig
	if err = json.Unmarshal(content, &clusterConfigs); err != nil {
		return kube.ClustersConfig{}, deferFn, err
	}

	configs := kube.ClustersConfig{Clusters: map[string]kube.ClusterConfig{}}
	configs.PinnipedProxyURL = pinnipedProxyURL
	for _, c := range clusterConfigs {
		// Select the cluster in which Kubeapps in installed. We look for either
		// `isKubeappsCluster: true` or an empty `APIServiceURL`.
		isKubeappsClusterCandidate := c.IsKubeappsCluster || c.APIServiceURL == ""
		if isKubeappsClusterCandidate {
			if configs.KubeappsClusterName == "" {
				configs.KubeappsClusterName = c.Name
			} else {
				return kube.ClustersConfig{}, nil, fmt.Errorf("only one cluster can be configured using either 'isKubeappsCluster: true' or without an apiServiceURL to refer to the cluster on which Kubeapps is installed, two defined: %q, %q", configs.KubeappsClusterName, c.Name)
			}
		}

		// We need to decode the base64-encoded cadata from the input.
		if c.CertificateAuthorityData != "" {
			decodedCAData, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
			if err != nil {
				return kube.ClustersConfig{}, deferFn, err
			}
			c.CertificateAuthorityDataDecoded = string(decodedCAData)

			// We also need a CAFile field because Helm uses the genericclioptions.ConfigFlags
			// struct which does not support CAData.
			// https://github.com/kubernetes/cli-runtime/issues/8
			c.CAFile = filepath.Join(caFilesDir, c.Name)
			err = ioutil.WriteFile(c.CAFile, decodedCAData, 0644)
			if err != nil {
				return kube.ClustersConfig{}, deferFn, err
			}
		}
		configs.Clusters[c.Name] = c
	}
	return configs, deferFn, nil
}

// dynClientGetterForContext returns a k8s client for use during interactions with the cluster.
// It utilizes the user credential from the request context. The plugins just have to call this function
// passing the context in order to retrieve the configured k8s client
func dynClientGetterForContext(ctx context.Context) (dynamic.Interface, error) {
	token, err := extractToken(ctx)
	if err != nil {
		return nil, err
	}

	// If there is no clusters config, we default to the previous behaviour of a "default" cluster.
	config := kube.ClustersConfig{KubeappsClusterName: "default"}
	if clustersConfigPath != "" {
		var err error
		var cleanupCAFiles func()
		config, cleanupCAFiles, err = parseClusterConfig(clustersConfigPath, clustersCAFilesPrefix, pinnipedProxyURL)
		if err != nil {
			log.Fatalf("unable to parse additional clusters config: %+v", err)
		}
		defer cleanupCAFiles()
	}

	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get inClusterConfig: %w", err)
	}
	var client dynamic.Interface
	if !unsafeUseDemoSA {
		restConfig, err := kube.NewClusterConfig(inClusterConfig, token, "default", config)
		if err != nil {
			return nil, fmt.Errorf("unable to get clusterConfig: %w", err)
		}
		client, err = dynamic.NewForConfig(restConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to create dynamic client: %w", err)
		}
	} else {
		client, err = dynamic.NewForConfig(inClusterConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to create dynamic client: %w", err)
		}
	}
	return client, nil
}

// extractToken returns the token passed through the gRPC request in the "authorization" metadata
// It is equivalent to the A"uthorization" usual HTTP 1 header
// For instance: authorization="Bearer abc" will return "abc"
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "error reading request metadata/headers")
	}
	if len(md["authorization"]) > 0 {
		if strings.HasPrefix(md["authorization"][0], "Bearer ") {
			return strings.TrimPrefix(md["authorization"][0], "Bearer "), nil
		} else {
			return "", status.Errorf(codes.Unauthenticated, "malformed authorization metadata")
		}
	} else {
		// No authorization header found, no error here, we will delegate it to the RBAC
		return "", nil
	}
}
