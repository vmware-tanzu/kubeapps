// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/ahmetb/go-linq/v3"
	"github.com/bufbuild/connect-go"
	pluginsv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	packages "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	connectpackages "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1/v1alpha1connect"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"

	"google.golang.org/grpc/metadata"
	log "k8s.io/klog/v2"
)

// pkgPluginWithServer stores the plugin detail together with its implementation.
type pkgPluginWithServer struct {
	plugin *v1alpha1.Plugin
	server connectpackages.PackagesServiceHandler
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
		pkgsSrv, ok := p.Server.(connectpackages.PackagesServiceHandler)
		if !ok {
			return nil, fmt.Errorf("unable to convert plugin %v to core PackagesServicesServer", p)
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
func (s packagesServer) GetAvailablePackageSummaries(ctx context.Context, request *connect.Request[packages.GetAvailablePackageSummariesRequest]) (*connect.Response[packages.GetAvailablePackageSummariesResponse], error) {
	log.InfoS("+core GetAvailablePackageSummaries", "cluster", request.Msg.GetContext().GetCluster(), "namespace", request.Msg.GetContext().GetNamespace())

	pageSize := request.Msg.GetPaginationOptions().GetPageSize()

	summariesWithOffsets, err := fanInAvailablePackageSummaries(ctx, s.pluginsWithServers, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to request results from registered plugins: %w", err))
	}

	pkgs := []*packages.AvailablePackageSummary{}
	categories := []string{}
	var pkgWithOffsets availableSummaryWithOffsets
	for pkgWithOffsets = range summariesWithOffsets {
		if pkgWithOffsets.err != nil {
			return nil, err
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
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to marshal next item offsets %v: %w", pkgWithOffsets.nextItemOffsets, err))
			}
			nextPageToken = string(token)
			break
		}
	}

	// Delete duplicate categories and sort by name
	From(categories).Distinct().OrderBy(func(i interface{}) interface{} { return i }).ToSlice(&categories)

	return connect.NewResponse(&packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: pkgs,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}), nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s packagesServer) GetAvailablePackageDetail(ctx context.Context, request *connect.Request[packages.GetAvailablePackageDetailRequest]) (*connect.Response[packages.GetAvailablePackageDetailResponse], error) {
	log.InfoS("+core GetAvailablePackageDetail", "cluster", request.Msg.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetAvailablePackageRef().GetContext().GetNamespace())

	if request.Msg.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.AvailablePackageRef.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageDetail(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to get the available package detail for the package %q using the plugin %q: %w", request.Msg.AvailablePackageRef.Identifier, request.Msg.AvailablePackageRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.GetAvailablePackageDetail().GetAvailablePackageRef() == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid available package detail response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	// Build the response
	return connect.NewResponse(&packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: response.Msg.AvailablePackageDetail,
	}), nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s packagesServer) GetInstalledPackageSummaries(ctx context.Context, request *connect.Request[packages.GetInstalledPackageSummariesRequest]) (*connect.Response[packages.GetInstalledPackageSummariesResponse], error) {
	log.InfoS("+core GetInstalledPackageSummaries", "cluster", request.Msg.GetContext().GetCluster(), "namespace", request.Msg.GetContext().GetNamespace())

	pageSize := request.Msg.GetPaginationOptions().GetPageSize()

	summariesWithOffsets, err := fanInInstalledPackageSummaries(ctx, s.pluginsWithServers, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to request results from registered plugins: %w", err))
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
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to marshal next item offsets %v: %w", pkgWithOffsets.nextItemOffsets, err))
			}
			nextPageToken = string(token)
			break
		}
	}

	return connect.NewResponse(&packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: pkgs,
		NextPageToken:             nextPageToken,
	}), nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s packagesServer) GetInstalledPackageDetail(ctx context.Context, request *connect.Request[packages.GetInstalledPackageDetailRequest]) (*connect.Response[packages.GetInstalledPackageDetailResponse], error) {
	log.InfoS("+core GetInstalledPackageDetail", "cluster", request.Msg.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.Msg.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.InstalledPackageRef.Plugin))
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageDetail(ctx, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to get the installed package detail for the package %q using the plugin %q: %w", request.Msg.InstalledPackageRef.Identifier, request.Msg.InstalledPackageRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.GetInstalledPackageDetail() == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid GetInstalledPackageDetail response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	// Build the response
	return connect.NewResponse(&packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: response.Msg.InstalledPackageDetail,
	}), nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s packagesServer) GetAvailablePackageVersions(ctx context.Context, request *connect.Request[packages.GetAvailablePackageVersionsRequest]) (*connect.Response[packages.GetAvailablePackageVersionsResponse], error) {
	log.InfoS("+core GetAvailablePackageVersions", "cluster", request.Msg.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetAvailablePackageRef().GetContext().GetNamespace())

	if request.Msg.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.AvailablePackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.GetAvailablePackageVersions(ctxForPlugin, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to get the available package versions for the package %q using the plugin %q: %w", request.Msg.AvailablePackageRef.Identifier, request.Msg.AvailablePackageRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.PackageAppVersions == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid GetAvailablePackageVersions response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	// Build the response
	return connect.NewResponse(&packages.GetAvailablePackageVersionsResponse{
		PackageAppVersions: response.Msg.PackageAppVersions,
	}), nil
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
// an installed package.
func (s *packagesServer) GetInstalledPackageResourceRefs(ctx context.Context, request *connect.Request[packages.GetInstalledPackageResourceRefsRequest]) (*connect.Response[packages.GetInstalledPackageResourceRefsResponse], error) {
	pkgRef := request.Msg.GetInstalledPackageRef()
	identifier := pkgRef.GetIdentifier()
	log.InfoS("+core GetInstalledPackageResourceRefs", "cluster", pkgRef.GetContext().GetCluster(), "namespace", pkgRef.GetContext().GetNamespace(), "identifier", identifier)

	if request.Msg.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin %v", request.Msg.InstalledPackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.GetInstalledPackageResourceRefs(ctxForPlugin, request)
	if err != nil {
		log.Errorf("Unable to get the resource refs for the package %q using the plugin %q: %v", request.Msg.InstalledPackageRef.Identifier, request.Msg.InstalledPackageRef.Plugin.Name, err)

		errCode := connect.CodeOf(err)
		connectError := connect.NewError(errCode, fmt.Errorf("Unable to get the resource refs for the package %q using the plugin %q: %v", request.Msg.InstalledPackageRef.Identifier, request.Msg.InstalledPackageRef.Plugin.Name, errCode))

		// Plugins are still using gRPC here, not connect:
		return nil, connect.NewError(errCode, connectError)
	}

	return response, nil
}

// CreateInstalledPackage creates an installed package using configured plugins.
func (s packagesServer) CreateInstalledPackage(ctx context.Context, request *connect.Request[packages.CreateInstalledPackageRequest]) (*connect.Response[packages.CreateInstalledPackageResponse], error) {
	log.InfoS("+core CreateInstalledPackage", "cluster", request.Msg.GetTargetContext().GetCluster(), "namespace", request.Msg.GetTargetContext().GetNamespace())

	if request.Msg.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.AvailablePackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.CreateInstalledPackage(ctxForPlugin, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to create the installed package for the package %q using the plugin %q: %w", request.Msg.AvailablePackageRef.Identifier, request.Msg.AvailablePackageRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.InstalledPackageRef == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid CreateInstalledPackage response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	return response, nil
}

// UpdateInstalledPackage updates an installed package using configured plugins.
func (s packagesServer) UpdateInstalledPackage(ctx context.Context, request *connect.Request[packages.UpdateInstalledPackageRequest]) (*connect.Response[packages.UpdateInstalledPackageResponse], error) {
	log.InfoS("+core UpdateInstalledPackage", "cluster", request.Msg.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.Msg.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.InstalledPackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.UpdateInstalledPackage(ctxForPlugin, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to update the installed package for the package %q using the plugin %q: %w", request.Msg.InstalledPackageRef.Identifier, request.Msg.InstalledPackageRef.Plugin.Name, err))
	}

	// Validate the plugin response
	if response.Msg.InstalledPackageRef == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid UpdateInstalledPackage response from the plugin %v: %w", pluginWithServer.plugin.Name, err))
	}

	return response, nil
}

// DeleteInstalledPackage deletes an installed package using configured plugins.
func (s packagesServer) DeleteInstalledPackage(ctx context.Context, request *connect.Request[packages.DeleteInstalledPackageRequest]) (*connect.Response[packages.DeleteInstalledPackageResponse], error) {
	log.InfoS("+core DeleteInstalledPackage", "cluster", request.Msg.GetInstalledPackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetInstalledPackageRef().GetContext().GetNamespace())

	if request.Msg.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.InstalledPackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.DeleteInstalledPackage(ctxForPlugin, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to delete the installed packagefor the package %q using the plugin %q: %w", request.Msg.InstalledPackageRef.Identifier, request.Msg.InstalledPackageRef.Plugin.Name, err))
	}

	return response, nil
}

func (s packagesServer) GetAvailablePackageMetadatas(ctx context.Context, request *connect.Request[packages.GetAvailablePackageMetadatasRequest]) (*connect.Response[packages.GetAvailablePackageMetadatasResponse], error) {
	log.InfoS("+core GetAvailablePackageVersions", "cluster", request.Msg.GetAvailablePackageRef().GetContext().GetCluster(), "namespace", request.Msg.GetAvailablePackageRef().GetContext().GetNamespace())

	if request.Msg.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)"))
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.Msg.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the plugin %v", request.Msg.AvailablePackageRef.Plugin))
	}

	// Get the response from the requested plugin
	ctxForPlugin := updateContextWithAuthz(ctx, request.Header())
	response, err := pluginWithServer.server.GetAvailablePackageMetadatas(ctxForPlugin, request)
	if err != nil {
		return nil, connect.NewError(connect.CodeOf(err), fmt.Errorf("Unable to get the available package metadatas for the package %q using the plugin %q: %w", request.Msg.AvailablePackageRef.Identifier, request.Msg.AvailablePackageRef.Plugin.Name, err))
	}

	// Build the response
	return connect.NewResponse(&packages.GetAvailablePackageMetadatasResponse{
		PackageMetadata: response.Msg.PackageMetadata,
	}), nil
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

func updateContextWithAuthz(ctx context.Context, h http.Header) context.Context {
	// Add authz to context metadata for untransitioned plugins.
	// TODO: Remove once plugins transitioned.
	token := h.Get("Authorization")
	ctxWithToken := ctx
	if token != "" {
		ctxWithToken = metadata.NewIncomingContext(ctx, metadata.MD{
			"authorization": []string{token},
		})
	}
	return ctxWithToken
}
