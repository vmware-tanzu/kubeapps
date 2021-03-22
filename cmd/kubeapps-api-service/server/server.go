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
	"fmt"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"
	"plugin"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/core"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	log "k8s.io/klog/v2"

	kapp_pkgsv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/plugins/kapp-controller/packages/v1"
)

const PLUGIN_ROOT_DIR = "/"
const GRPC_REGISTER_FUNCTION = "RegisterWithGRPCServer"

// TODO:
// * Implement plugins available
// * Add proto for creating a deployed package or displaying an installed app.
//   - identifier for the package (may include repository for some formats?)
//   - yaml string for
// * Add gateway to main, multiplexing on the port
//
// server is used to implement kubeapps.Packages

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(port int, pluginDirs []string) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	core.RegisterCoreServer(s, &coreServer{})

	plugins, err := listSOFiles(os.DirFS(PLUGIN_ROOT_DIR), pluginDirs)
	if err != nil {
		log.Fatalf("failed to check for plugins: %v", err)
	}

	err = registerPlugins(s, plugins)
	if err != nil {
		log.Fatalf("failed to register plugins: %v", err)
	}

	kapp_pkgsv1.RegisterPackagesServer(s, &kapp_pkgsv1.Server{})
	reflection.Register(s)
	log.Infof("Starting server on :%d", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// registerPlugins opens each plugin, looks up the register function and calls it with the registrar.
func registerPlugins(s grpc.ServiceRegistrar, pluginPaths []string) error {
	for _, pluginPath := range pluginPaths {
		p, err := plugin.Open(pluginPath)
		if err != nil {
			return err
		}
		regFn, err := p.Lookup(GRPC_REGISTER_FUNCTION)
		if err != nil {
			return err
		}
		regFn.(func(grpc.ServiceRegistrar))(s)
		log.Infof("Successfully registered plugin %q", pluginPath)
	}
	return nil
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
		relPluginDir, err := filepath.Rel(PLUGIN_ROOT_DIR, pluginDir)
		if err != nil {
			return nil, err
		}

		m, err := fs.Glob(fsys, path.Join(relPluginDir, "/", "*.so"))
		if err != nil {
			return nil, err
		}

		for _, match := range m {
			matches = append(matches, filepath.Join(PLUGIN_ROOT_DIR, match))
		}
	}
	return matches, nil
}
