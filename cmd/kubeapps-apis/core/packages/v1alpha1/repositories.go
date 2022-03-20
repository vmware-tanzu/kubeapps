// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"

	pluginsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// repoPluginsWithServer stores the plugin detail together with its implementation.
type repoPluginsWithServer struct {
	plugin *v1alpha1.Plugin
	server packages.RepositoriesServiceServer
}

// repositoriesServer implements the API defined in proto/kubeappsapis/core/packages/v1alpha1/repositories.proto
type repositoriesServer struct {
	packages.UnimplementedRepositoriesServiceServer

	// pluginsWithServers is a slice of all registered pluginsWithServers which satisfy the core.packages.v1alpha1
	// interface.
	pluginsWithServers []repoPluginsWithServer
}

func NewRepositoriesServer(pkgingPlugins []pluginsv1alpha1.PluginWithServer) (*repositoriesServer, error) {
	// Verify that each plugin is indeed a packaging plugin while
	// casting.
	pluginsWithServer := make([]repoPluginsWithServer, len(pkgingPlugins))
	for i, p := range pkgingPlugins {
		pkgsSrv, ok := p.Server.(packages.RepositoriesServiceServer)
		if !ok {
			return nil, fmt.Errorf("unable to convert plugin %v to core RepositoriesServiceServer", p)
		}
		pluginsWithServer[i] = repoPluginsWithServer{
			plugin: p.Plugin,
			server: pkgsSrv,
		}
		log.Infof("Registered %v for core.packaging.v1alpha1 aggregation.", p.Plugin)
	}
	return &repositoriesServer{
		pluginsWithServers: pluginsWithServer,
	}, nil
}

func (s repositoriesServer) AddPackageRepository(ctx context.Context, request *packages.AddPackageRepositoryRequest) (*packages.AddPackageRepositoryResponse, error) {
	contextMsg := fmt.Sprintf("(name=%q, cluster=%q, namespace=%q)", request.GetName(), request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+core AddPackageRepository %s", contextMsg)

	if request.GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing request.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.AddPackageRepository(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to add package repository %q using the plugin %q: %v", request.Name, request.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.PackageRepoRef == nil {
		return nil, status.Errorf(codes.Internal, "Invalid AddPackageRepository response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

func (s repositoriesServer) GetPackageRepositoryDetail(ctx context.Context, request *packages.GetPackageRepositoryDetailRequest) (*packages.GetPackageRepositoryDetailResponse, error) {
	contextMsg := fmt.Sprintf("(identifier=%q, cluster=%q, namespace=%q)", request.GetPackageRepoRef().GetIdentifier(), request.GetPackageRepoRef().GetContext().GetCluster(), request.GetPackageRepoRef().GetContext().GetNamespace())
	log.Infof("+core GetPackageRepositoryDetail %s", contextMsg)

	if request.GetPackageRepoRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing PackageRepoRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.PackageRepoRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetPackageRepositoryDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to get the package repository detail for the repository %q using the plugin %q: %v", request.PackageRepoRef.Identifier, request.PackageRepoRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.GetDetail().GetPackageRepoRef() == nil {
		return nil, status.Errorf(codes.Internal, "Invalid package reposirtory detail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetPackageRepositoryDetailResponse{
		Detail: response.Detail,
	}, nil
}

// getPluginWithServer returns the *pkgPluginsWithServer from a given packagesServer
// matching the plugin name
func (s repositoriesServer) getPluginWithServer(plugin *v1alpha1.Plugin) *repoPluginsWithServer {
	for _, p := range s.pluginsWithServers {
		if plugin.Name == p.plugin.Name {
			return &p
		}
	}
	return nil
}
