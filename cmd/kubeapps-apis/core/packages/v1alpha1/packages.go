// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"
	"strconv"

	linq "github.com/ahmetb/go-linq/v3"
	pluginsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core/plugins/v1alpha1"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// pkgPluginsWithServer stores the plugin detail together with its implementation.
type pkgPluginsWithServer struct {
	plugin *pluginsGRPCv1alpha1.Plugin
	server pkgsGRPCv1alpha1.PackagesServiceServer
}

// packagesServer implements the API defined in proto/kubeappsapis/core/packages/v1alpha1/packages.proto
type packagesServer struct {
	pkgsGRPCv1alpha1.UnimplementedPackagesServiceServer

	// pluginsWithServers is a slice of all registered pluginsWithServers which satisfy the core.packages.v1alpha1
	// interface.
	pluginsWithServers []pkgPluginsWithServer
}

func NewPackagesServer(pkgingPlugins []pluginsv1alpha1.PluginWithServer) (*packagesServer, error) {
	// Verify that each plugin is indeed a packaging plugin while
	// casting.
	pluginsWithServer := make([]pkgPluginsWithServer, len(pkgingPlugins))
	for i, p := range pkgingPlugins {
		pkgsSrv, ok := p.Server.(pkgsGRPCv1alpha1.PackagesServiceServer)
		if !ok {
			return nil, fmt.Errorf("Unable to convert plugin %v to core PackagesServicesServer", p)
		}
		pluginsWithServer[i] = pkgPluginsWithServer{
			plugin: p.Plugin,
			server: pkgsSrv,
		}
		log.Infof("Registered %v for core.packaging.v1alpha1 aggregation.", p.Plugin)
	}
	return &packagesServer{
		pluginsWithServers: pluginsWithServer,
	}, nil
}

// GetAvailablePackages returns the packages based on the request.
func (s packagesServer) GetAvailablePackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+core GetAvailablePackageSummaries %s", contextMsg)

	pageOffset, err := pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	pageSize := request.GetPaginationOptions().GetPageSize()
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to intepret page token %q: %v", request.GetPaginationOptions().GetPageToken(), err)
	}

	// TODO(agamez): temporarily fetching all the results (size=0) and then paginate them
	// ideally, paginate each plugin request and then aggregate results.
	requestN := request
	requestN.PaginationOptions = &pkgsGRPCv1alpha1.PaginationOptions{
		PageToken: "0",
		PageSize:  0,
	}

	pkgs := []*pkgsGRPCv1alpha1.AvailablePackageSummary{}
	categories := []string{}

	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.pluginsWithServers {
		log.Infof("Items now: %d/%d", len(pkgs), (pageOffset*int(pageSize) + int(pageSize)))
		if pageSize == 0 || len(pkgs) <= (pageOffset*int(pageSize)+int(pageSize)) {
			log.Infof("Should enter")

			response, err := p.server.GetAvailablePackageSummaries(ctx, requestN)
			if err != nil {
				return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Invalid GetAvailablePackageSummaries response from the plugin %v: %v", p.plugin.Name, err)
			}

			categories = append(categories, response.Categories...)

			// Add the plugin for the pkgs
			pluginPkgs := response.AvailablePackageSummaries
			for _, r := range pluginPkgs {
				if r.AvailablePackageRef == nil {
					r.AvailablePackageRef = &pkgsGRPCv1alpha1.AvailablePackageReference{}
				}
				r.AvailablePackageRef.Plugin = p.plugin
			}
			pkgs = append(pkgs, pluginPkgs...)
		}
	}
	// Delete duplicate categories and sort by name
	linq.From(categories).Distinct().OrderBy(func(i interface{}) interface{} { return i }).ToSlice(&categories)

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 {
		// Using https://github.com/ahmetb/go-linq for simplicity
		linq.From(pkgs).
			// Order by package name, regardless of the plugin
			OrderBy(func(pkg interface{}) interface{} {
				return pkg.(*pkgsGRPCv1alpha1.AvailablePackageSummary).Name + pkg.(*pkgsGRPCv1alpha1.AvailablePackageSummary).AvailablePackageRef.Plugin.Name
			}).
			Skip(pageOffset * int(pageSize)).
			Take(int(pageSize)).
			ToSlice(&pkgs)

		if len(pkgs) == int(pageSize) {
			nextPageToken = fmt.Sprintf("%d", pageOffset+1)
		}
	} else {
		linq.From(pkgs).
			// Order by package name, regardless of the plugin
			OrderBy(func(pkg interface{}) interface{} {
				return pkg.(*pkgsGRPCv1alpha1.AvailablePackageSummary).Name + pkg.(*pkgsGRPCv1alpha1.AvailablePackageSummary).AvailablePackageRef.Plugin.Name
			}).ToSlice(&pkgs)
	}

	return &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: pkgs,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s packagesServer) GetAvailablePackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+core GetAvailablePackageDetail %s", contextMsg)

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageDetail(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to get the available package detail for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.GetAvailablePackageDetail().GetAvailablePackageRef() == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Invalid available package detail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &pkgsGRPCv1alpha1.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: response.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s packagesServer) GetInstalledPackageSummaries(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetContext().GetCluster(), request.GetContext().GetNamespace())
	log.Infof("+core GetInstalledPackageSummaries %s", contextMsg)

	// Aggregate the response for each plugin
	pkgs := []*pkgsGRPCv1alpha1.InstalledPackageSummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.pluginsWithServers {
		response, err := p.server.GetInstalledPackageSummaries(ctx, request)
		if err != nil {
			return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Invalid GetInstalledPackageSummaries response from the plugin %v: %v", p.plugin.Name, err)
		}

		// Add the plugin for the pkgs
		pluginPkgs := response.InstalledPackageSummaries
		for _, r := range pluginPkgs {
			if r.InstalledPackageRef == nil {
				r.InstalledPackageRef = &pkgsGRPCv1alpha1.InstalledPackageReference{}
			}
			r.InstalledPackageRef.Plugin = p.plugin
		}
		pkgs = append(pkgs, pluginPkgs...)
	}

	linq.From(pkgs).
		// Order by package name, regardless of the plugin
		OrderBy(func(pkg interface{}) interface{} {
			return pkg.(*pkgsGRPCv1alpha1.InstalledPackageSummary).Name + pkg.(*pkgsGRPCv1alpha1.InstalledPackageSummary).InstalledPackageRef.Plugin.Name
		}).
		ToSlice(&pkgs)

	// Build the response
	return &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: pkgs,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s packagesServer) GetInstalledPackageDetail(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetInstalledPackageRef().GetContext().GetCluster(), request.GetInstalledPackageRef().GetContext().GetNamespace())
	log.Infof("+core GetInstalledPackageDetail %s", contextMsg)

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageDetail(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to get the installed package detail for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.GetInstalledPackageDetail() == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Invalid GetInstalledPackageDetail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: response.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s packagesServer) GetAvailablePackageVersions(ctx context.Context, request *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest) (*pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetAvailablePackageRef().GetContext().GetCluster(), request.GetAvailablePackageRef().GetContext().GetNamespace())
	log.Infof("+core GetAvailablePackageVersions %s", contextMsg)

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageVersions(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to get the available package versions for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.PackageAppVersions == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Invalid GetAvailablePackageVersions response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: response.PackageAppVersions,
	}, nil
}

// GetInstalledPackageResourceRefs returns the references for the Kubernetes resources created by
// an installed package.
func (s *packagesServer) GetInstalledPackageResourceRefs(ctx context.Context, request *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest) (*pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", pkgRef.GetContext().GetCluster(), pkgRef.GetContext().GetNamespace())
	identifier := pkgRef.GetIdentifier()
	log.Infof("+core GetInstalledPackageResourceRefs %s %s", contextMsg, identifier)

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageResourceRefs(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to get the resource refs for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	return response, nil
}

// CreateInstalledPackage creates an installed package using configured plugins.
func (s packagesServer) CreateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.CreateInstalledPackageRequest) (*pkgsGRPCv1alpha1.CreateInstalledPackageResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetTargetContext().GetCluster(), request.GetTargetContext().GetNamespace())
	log.Infof("+core CreateInstalledPackage %s", contextMsg)

	if request.GetAvailablePackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.CreateInstalledPackage(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to create the installed package for the package %q using the plugin %q: %v", request.AvailablePackageRef.Identifier, request.AvailablePackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.InstalledPackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Invalid CreateInstalledPackage response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

// UpdateInstalledPackage updates an installed package using configured plugins.
func (s packagesServer) UpdateInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest) (*pkgsGRPCv1alpha1.UpdateInstalledPackageResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetInstalledPackageRef().GetContext().GetCluster(), request.GetInstalledPackageRef().GetContext().GetNamespace())
	log.Infof("+core UpdateInstalledPackage %s", contextMsg)

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.UpdateInstalledPackage(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to update the installed package for the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	// Validate the plugin response
	if response.InstalledPackageRef == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Invalid UpdateInstalledPackage response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	return response, nil
}

// DeleteInstalledPackage deletes an installed package using configured plugins.
func (s packagesServer) DeleteInstalledPackage(ctx context.Context, request *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest) (*pkgsGRPCv1alpha1.DeleteInstalledPackageResponse, error) {
	contextMsg := fmt.Sprintf("(cluster=%q, namespace=%q)", request.GetInstalledPackageRef().GetContext().GetCluster(), request.GetInstalledPackageRef().GetContext().GetNamespace())
	log.Infof("+core DeleteInstalledPackage %s", contextMsg)

	if request.GetInstalledPackageRef().GetPlugin() == nil {
		return nil, grpcstatus.Errorf(grpccodes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.DeleteInstalledPackage(ctx, request)
	if err != nil {
		return nil, grpcstatus.Errorf(grpcstatus.Convert(err).Code(), "Unable to delete the installed packagefor the package %q using the plugin %q: %v", request.InstalledPackageRef.Identifier, request.InstalledPackageRef.Plugin.Name, err)
	}

	return response, nil
}

// getPluginWithServer returns the *pkgPluginsWithServer from a given packagesServer
// matching the plugin name
func (s packagesServer) getPluginWithServer(plugin *pluginsGRPCv1alpha1.Plugin) *pkgPluginsWithServer {
	for _, p := range s.pluginsWithServers {
		if plugin.Name == p.plugin.Name {
			return &p
		}
	}
	return nil
}

// pageOffsetFromPageToken converts a page token to an integer offset
// representing the page of results.
// TODO(mnelson): When aggregating results from different plugins, we'll
// need to update the actual query in GetPaginatedChartListWithFilters to
// use a row offset rather than a page offset (as not all rows may be consumed
// for a specific plugin when combining).
func pageOffsetFromPageToken(pageToken string) (int, error) {
	if pageToken == "" {
		return 0, nil
	}
	offset, err := strconv.ParseUint(pageToken, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(offset), nil
}
