// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"

	. "github.com/ahmetb/go-linq/v3"

	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packages "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

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
		log.Infof("Registered %v for core.packaging.v1alpha1 repositories aggregation.", p.Plugin)
	}
	return &repositoriesServer{
		pluginsWithServers: pluginsWithServer,
	}, nil
}

func (s repositoriesServer) AddPackageRepository(ctx context.Context, request *packages.AddPackageRepositoryRequest) (*packages.AddPackageRepositoryResponse, error) {
	log.InfoS("+core AddPackageRepository", "name", request.GetName(), "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

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
	log.InfoS("+core GetPackageRepositoryDetail", "identifier", request.GetPackageRepoRef().GetIdentifier(), "cluster", request.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.GetPackageRepoRef().GetContext().GetNamespace())

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

// GetPackageRepositorySummaries returns the package repository summaries based on the request.
func (s repositoriesServer) GetPackageRepositorySummaries(ctx context.Context, request *packages.GetPackageRepositorySummariesRequest) (*packages.GetPackageRepositorySummariesResponse, error) {
	log.InfoS("+core GetPackageRepositorySummaries", "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

	// Aggregate the response for each plugin
	summaries := []*packages.PackageRepositorySummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.pluginsWithServers {
		response, err := p.server.GetPackageRepositorySummaries(ctx, request)
		if err != nil {
			return nil, status.Errorf(status.Convert(err).Code(), "Invalid GetPackageRepositorySummaries response from the plugin %v: %v", p.plugin.Name, err)
		}
		if response == nil {
			log.Infof("core GetPackageRepositorySummaries received nil response from plugin %s / %s", p.plugin.GetName(), p.plugin.GetVersion())
			continue
		}

		// Add the plugin for the pkgs
		pluginSummaries := response.PackageRepositorySummaries
		for _, r := range pluginSummaries {
			if r.PackageRepoRef == nil {
				r.PackageRepoRef = &packages.PackageRepositoryReference{}
			}
			r.PackageRepoRef.Plugin = p.plugin
		}
		summaries = append(summaries, pluginSummaries...)
	}

	From(summaries).
		// Order by repo name, regardless of the plugin
		OrderBy(func(repo interface{}) interface{} {
			return repo.(*packages.PackageRepositorySummary).Name + repo.(*packages.PackageRepositorySummary).PackageRepoRef.Plugin.Name
		}).ToSlice(&summaries)

	// Build the response
	return &packages.GetPackageRepositorySummariesResponse{
		PackageRepositorySummaries: summaries,
	}, nil
}

// UpdatePackageRepository updates a package repository using configured plugins.
func (s repositoriesServer) UpdatePackageRepository(ctx context.Context, request *packages.UpdatePackageRepositoryRequest) (*packages.UpdatePackageRepositoryResponse, error) {
	log.InfoS("+core UpdatePackageRepository", "cluster", request.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.GetPackageRepoRef().GetContext().GetNamespace(), "id", request.GetPackageRepoRef().GetIdentifier())

	if request.GetPackageRepoRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing PackageRepoRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.PackageRepoRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.UpdatePackageRepository(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to update the package repository %q using the plugin %q: %v",
			request.PackageRepoRef.Identifier, request.PackageRepoRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.PackageRepoRef == nil {
		return nil, status.Errorf(codes.Internal, "Invalid UpdatePackageRepository response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

// DeletePackageRepository deletes a package repository using configured plugins.
func (s repositoriesServer) DeletePackageRepository(ctx context.Context, request *packages.DeletePackageRepositoryRequest) (*packages.DeletePackageRepositoryResponse, error) {
	log.InfoS("+core DeletePackageRepository", "cluster", request.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.GetPackageRepoRef().GetContext().GetNamespace(), "id", request.GetPackageRepoRef().GetIdentifier())

	if request.GetPackageRepoRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing PackageRepoRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.PackageRepoRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.DeletePackageRepository(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to delete the package repository %q using the plugin %q: %v",
			request.PackageRepoRef.Identifier, request.PackageRepoRef.Plugin.Name, err)
	}

	return response, nil
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
