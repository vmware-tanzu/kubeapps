// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"

	. "github.com/ahmetb/go-linq/v3"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packages "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// pkgPluginWithServer stores the plugin detail together with its implementation.
type pkgPluginWithServer struct {
	plugin *v1alpha1.Plugin
	server packages.PackagesServiceServer
}

// packagesServer implements the API defined in proto/kubeappsapis/core/packages/v1alpha1/packages.proto
type packagesServer struct {
	packages.UnimplementedPackagesServiceServer

	// pluginsWithServers is a slice of all registered pluginsWithServers which satisfy the core.packages.v1alpha1
	// interface.
	pluginsWithServers []pkgPluginWithServer
}

func NewPackagesServer(pkgingPlugins []pluginsv1alpha1.PluginWithServer) (*packagesServer, error) {
	// Verify that each plugin is indeed a packaging plugin while
	// casting.
	pluginsWithServer := make([]pkgPluginWithServer, len(pkgingPlugins))
	for i, p := range pkgingPlugins {
		pkgsSrv, ok := p.Server.(packages.PackagesServiceServer)
		if !ok {
			return nil, fmt.Errorf("Unable to convert plugin %v to core PackagesServicesServer", p)
		}
		pluginsWithServer[i] = pkgPluginWithServer{
			plugin: p.Plugin,
			server: pkgsSrv,
		}
		log.Infof("Registered %v for core.packaging.v1alpha1 packages aggregation.", p.Plugin)
	}
	return &packagesServer{
		pluginsWithServers: pluginsWithServer,
	}, nil
}

// GetAvailablePackageSummaries returns the packages based on the request.
func (s packagesServer) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	log.InfoS("+core GetAvailablePackageSummaries", "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

	pageSize := request.GetPaginationOptions().GetPageSize()

	summariesWithOffsets, err := fanInAvailablePackageSummaries(ctx, s.pluginsWithServers, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to request results from registered plugins: %v", err)
	}

	pkgs := []*packages.AvailablePackageSummary{}
	categories := []string{}
	var pkgWithOffsets availableSummaryWithOffsets
	for pkgWithOffsets = range summariesWithOffsets {
		if pkgWithOffsets.err != nil {
			return nil, pkgWithOffsets.err
		}
		pkgs = append(pkgs, pkgWithOffsets.availablePackageSummary)
		categories = append(categories, pkgWithOffsets.categories...)
		if pageSize > 0 && len(pkgs) >= int(pageSize) {
			break
		}
	}

	// Only return a next page token of the combined plugin offsets if at least one
	// plugin is not completely exhausted.
	nextPageToken := ""
	for _, v := range pkgWithOffsets.nextItemOffsets {
		if v != CompleteToken {
			token, err := json.Marshal(pkgWithOffsets.nextItemOffsets)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Unable to marshal next item offsets %v: %s", pkgWithOffsets.nextItemOffsets, err)
			}
			nextPageToken = string(token)
			break
		}
	}

	// Delete duplicate categories and sort by name
	From(categories).Distinct().OrderBy(func(i interface{}) interface{} { return i }).ToSlice(&categories)

	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: pkgs,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s packagesServer) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	log.InfoS("+core GetAvailablePackageDetail", "cluster", request.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.GetAvailablePackageRef().GetContext().GetNamespace())

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to get the available package detail for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.GetAvailablePackageDetail().GetAvailablePackageRef() == nil {
		return nil, status.Errorf(codes.Internal, "Invalid available package detail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: response.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s packagesServer) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	log.InfoS("+core GetInstalledPackageSummaries", "cluster", request.GetContext().GetCluster(), "namespace", request.GetContext().GetNamespace())

	pageSize := request.GetPaginationOptions().GetPageSize()

	summariesWithOffsets, err := fanInInstalledPackageSummaries(ctx, s.pluginsWithServers, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to request results from registered plugins: %v", err)
	}
	pkgs := []*packages.InstalledPackageSummary{}
	var pkgWithOffsets installedSummaryWithOffsets
	for pkgWithOffsets = range summariesWithOffsets {
		if pkgWithOffsets.err != nil {
			return nil, pkgWithOffsets.err
		}
		pkgs = append(pkgs, pkgWithOffsets.installedPackageSummary)
		if pageSize > 0 && len(pkgs) >= int(pageSize) {
			break
		}
	}

	// Only return a next page token of the combined plugin offsets if at least one
	// plugin is not completely exhausted.
	nextPageToken := ""
	for _, v := range pkgWithOffsets.nextItemOffsets {
		if v != CompleteToken {
			token, err := json.Marshal(pkgWithOffsets.nextItemOffsets)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Unable to marshal next item offsets %v: %s", pkgWithOffsets.nextItemOffsets, err)
			}
			nextPageToken = string(token)
			break
		}
	}

	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: pkgs,
		NextPageToken:             nextPageToken,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s packagesServer) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	log.InfoS("+core GetInstalledPackageDetail", "cluster", request.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to get the installed package detail for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.GetInstalledPackageDetail() == nil {
		return nil, status.Errorf(codes.Internal, "Invalid GetInstalledPackageDetail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: response.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s packagesServer) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	log.InfoS("+core GetAvailablePackageVersions %s", "cluster", request.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.GetAvailablePackageRef().GetContext().GetNamespace())

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageVersions(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to get the available package versions for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.PackageAppVersions == nil {
		return nil, status.Errorf(codes.Internal, "Invalid GetAvailablePackageVersions response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: response.PackageAppVersions,
	}, nil
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
// an installed package.
func (s *packagesServer) GetInstalledPackageResourceRefs(ctx context.Context, request *packages.GetInstalledPackageResourceRefsRequest) (*packages.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	identifier := pkgRef.GetIdentifier()
	log.InfoS("+core GetInstalledPackageResourceRefs", "cluster", pkgRef.GetContext().GetCluster(), "namespace", pkgRef.GetContext().GetNamespace(), "identifier", identifier)

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageResourceRefs(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to get the resource refs for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	return response, nil
}

// CreateInstalledPackage creates an installed package using configured plugins.
func (s packagesServer) CreateInstalledPackage(ctx context.Context, request *packages.CreateInstalledPackageRequest) (*packages.CreateInstalledPackageResponse, error) {
	log.InfoS("+core CreateInstalledPackage", "cluster", request.GetTargetContext().GetCluster(), "namespace", request.GetTargetContext().GetNamespace())

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.CreateInstalledPackage(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to create the installed package for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.InstalledPackageRef == nil {
		return nil, status.Errorf(codes.Internal, "Invalid CreateInstalledPackage response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

// UpdateInstalledPackage updates an installed package using configured plugins.
func (s packagesServer) UpdateInstalledPackage(ctx context.Context, request *packages.UpdateInstalledPackageRequest) (*packages.UpdateInstalledPackageResponse, error) {
	log.InfoS("+core UpdateInstalledPackage", "cluster", request.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.UpdateInstalledPackage(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to update the installed package for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.InstalledPackageRef == nil {
		return nil, status.Errorf(codes.Internal, "Invalid UpdateInstalledPackage response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

// DeleteInstalledPackage deletes an installed package using configured plugins.
func (s packagesServer) DeleteInstalledPackage(ctx context.Context, request *packages.DeleteInstalledPackageRequest) (*packages.DeleteInstalledPackageResponse, error) {
	log.InfoS("+core DeleteInstalledPackage", "cluster", request.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.DeleteInstalledPackage(ctx, request)
	if err != nil {
		return nil, status.Errorf(status.Convert(err).Code(), "Unable to delete the installed packagefor the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	return response, nil
}

// getPluginWithServer returns the *pkgPluginsWithServer from a given packagesServer
// matching the plugin name
func (s packagesServer) getPluginWithServer(plugin *v1alpha1.Plugin) *pkgPluginWithServer {
	for _, p := range s.pluginsWithServers {
		if plugin.Name == p.plugin.Name {
			return &p
		}
	}
	return nil
}
