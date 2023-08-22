// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"
	"sort"
	"sync"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/bufbuild/connect-go"

	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packages "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	packagesconnect "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1/v1alpha1connect"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

	log "k8s.io/klog/v2"
)

// repoPluginsWithServer stores the plugin detail together with its implementation.
type repoPluginsWithServer struct {
	plugin *v1alpha1.Plugin
	server packagesconnect.RepositoriesServiceHandler
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
		pkgsSrv, ok := p.Server.(packagesconnect.RepositoriesServiceHandler)
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

func (s repositoriesServer) AddPackageRepository(ctx context.Context, request *connect.Request[packages.AddPackageRepositoryRequest]) (*connect.Response[packages.AddPackageRepositoryResponse], error) {
	log.InfoS("+core AddPackageRepository", "name", request.Msg.GetName(), "cluster", request.Msg.GetContext().GetCluster(), "namespace", request.Msg.GetContext().GetNamespace())

	if request.Msg.GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing request.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.AddPackageRepository(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to add package repository %q using the plugin %q: %w", request.Msg.Name, request.Msg.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.PackageRepoRef == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid AddPackageRepository response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	return response, nil
}

func (s repositoriesServer) GetPackageRepositoryDetail(ctx context.Context, request *connect.Request[packages.GetPackageRepositoryDetailRequest]) (*connect.Response[packages.GetPackageRepositoryDetailResponse], error) {
	log.InfoS("+core GetPackageRepositoryDetail", "identifier", request.Msg.GetPackageRepoRef().GetIdentifier(), "cluster", request.Msg.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.Msg.GetPackageRepoRef().GetContext().GetNamespace())

	if request.Msg.GetPackageRepoRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing PackageRepoRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.PackageRepoRef.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetPackageRepositoryDetail(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to get the package repository detail for the repository %q using the plugin %q: %w", request.Msg.PackageRepoRef.Identifier, request.Msg.PackageRepoRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.GetDetail().GetPackageRepoRef() == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid package reposirtory detail response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	// Build the response
	return connect.NewResponse(&packages.GetPackageRepositoryDetailResponse{
		Detail: response.Msg.Detail,
	}), nil
}

// GetPackageRepositorySummaries returns the package repository summaries based on the request.
func (s repositoriesServer) GetPackageRepositorySummaries(ctx context.Context, request *connect.Request[packages.GetPackageRepositorySummariesRequest]) (*connect.Response[packages.GetPackageRepositorySummariesResponse], error) {
	log.InfoS("+core GetPackageRepositorySummaries", "cluster", request.Msg.GetContext().GetCluster(), "namespace", request.Msg.GetContext().GetNamespace())

	// Aggregate the response for each plugin
	summaries := []*packages.PackageRepositorySummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.pluginsWithServers {
		response, err := p.server.GetPackageRepositorySummaries(ctx, request)
		if err != nil {
			return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Invalid GetPackageRepositorySummaries response from the plugin %v: %w", p.plugin.Name, err))
		}
		if response == nil {
			log.Infof("Core GetPackageRepositorySummaries received nil response from plugin %s / %s", p.plugin.GetName(), p.plugin.GetVersion())
			continue
		}

		// Add the plugin for the pkgs
		pluginSummaries := response.Msg.PackageRepositorySummaries
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
	return connect.NewResponse(&packages.GetPackageRepositorySummariesResponse{
		PackageRepositorySummaries: summaries,
	}), nil
}

// UpdatePackageRepository updates a package repository using configured plugins.
func (s repositoriesServer) UpdatePackageRepository(ctx context.Context, request *connect.Request[packages.UpdatePackageRepositoryRequest]) (*connect.Response[packages.UpdatePackageRepositoryResponse], error) {
	log.InfoS("+core UpdatePackageRepository", "cluster", request.Msg.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.Msg.GetPackageRepoRef().GetContext().GetNamespace(), "id", request.Msg.GetPackageRepoRef().GetIdentifier())

	if request.Msg.GetPackageRepoRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing PackageRepoRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.PackageRepoRef.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.UpdatePackageRepository(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to update the package repository %q using the plugin %q: %w",
			request.Msg.PackageRepoRef.Identifier, request.Msg.PackageRepoRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.PackageRepoRef == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid UpdatePackageRepository response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	return response, nil
}

// DeletePackageRepository deletes a package repository using configured plugins.
func (s repositoriesServer) DeletePackageRepository(ctx context.Context, request *connect.Request[packages.DeletePackageRepositoryRequest]) (*connect.Response[packages.DeletePackageRepositoryResponse], error) {
	log.InfoS("+core DeletePackageRepository", "cluster", request.Msg.GetPackageRepoRef().GetContext().GetCluster(), "namespace", request.Msg.GetPackageRepoRef().GetContext().GetNamespace(), "id", request.Msg.GetPackageRepoRef().GetIdentifier())

	if request.Msg.GetPackageRepoRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing PackageRepoRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.PackageRepoRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.PackageRepoRef.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.DeletePackageRepository(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to delete the package repository %q using the plugin %q: %w",
			request.Msg.PackageRepoRef.Identifier, request.Msg.PackageRepoRef.Plugin.Name, err))
	}

	return response, nil
}

func (s repositoriesServer) GetPackageRepositoryPermissions(ctx context.Context, request *connect.Request[packages.GetPackageRepositoryPermissionsRequest]) (*connect.Response[packages.GetPackageRepositoryPermissionsResponse], error) {
	log.InfoS("+core GetPackageRepositoryPermissions", "cluster", request.Msg.GetContext().GetCluster(), "namespace", request.Msg.GetContext().GetNamespace())
	resultsChannel := make(chan *connect.Response[packages.GetPackageRepositoryPermissionsResponse], len(s.pluginsWithServers))
	var wg sync.WaitGroup

	for _, p := range s.pluginsWithServers {
		wg.Add(1)
		go func(repoPlugin repoPluginsWithServer) {
			defer wg.Done()

			response, err := repoPlugin.server.GetPackageRepositoryPermissions(ctx, request)
			if err != nil {
				log.Errorf("+core error finding repository permissions in plugin %s: [%v]", repoPlugin.plugin.Name, err)
				return
			}
			resultsChannel <- response
		}(p)
	}
	go func() {
		wg.Wait()
		close(resultsChannel)
	}()

	var permissions []*packages.PackageRepositoriesPermissions
	for pluginResult := range resultsChannel {
		permissions = append(permissions, pluginResult.Msg.Permissions...)
	}
	sort.Slice(permissions, func(i, j int) bool {
		return pluginsv1alpha1.ComparePlugin(permissions[i].Plugin, permissions[j].Plugin)
	})

	return connect.NewResponse(&packages.GetPackageRepositoryPermissionsResponse{
		Permissions: permissions,
	}), nil
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
