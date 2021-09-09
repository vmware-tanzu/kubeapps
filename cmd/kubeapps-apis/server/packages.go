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
	"strconv"

	. "github.com/ahmetb/go-linq/v3"
	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	log "k8s.io/klog/v2"
)

// packagesServer implements the API defined in proto/kubeappsapis/core/packages/v1alpha1/packages.proto
type packagesServer struct {
	packages.UnimplementedPackagesServiceServer

	// plugins is a slice of all registered plugins which satisfy the core.packages.v1alpha1
	// interface.
	plugins []*pkgsPluginWithServer
}

func NewPackagesServer(plugins []*pkgsPluginWithServer) *packagesServer {
	return &packagesServer{
		plugins: plugins,
	}
}

// GetAvailablePackages returns the packages based on the request.
func (s packagesServer) GetAvailablePackageSummaries(ctx context.Context, request *packages.GetAvailablePackageSummariesRequest) (*packages.GetAvailablePackageSummariesResponse, error) {
	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+core GetAvailablePackageSummaries %s", contextMsg)

	pageOffset, err := pageOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	pageSize := request.GetPaginationOptions().GetPageSize()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to intepret page token %q: %v", request.GetPaginationOptions().GetPageToken(), err)
	}

	// TODO(agamez): temporarily fetching all the results (size=0) and then paginate them
	// ideally, paginate each plugin request and then aggregate results.
	requestN := request
	requestN.PaginationOptions = &packages.PaginationOptions{
		PageToken: "0",
		PageSize:  0,
	}

	pkgs := []*packages.AvailablePackageSummary{}
	categories := []string{}

	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.plugins {
		response, err := p.server.GetAvailablePackageSummaries(ctx, requestN)
		if err != nil {
			return nil, err
		}

		categories = append(categories, response.Categories...)

		// Add the plugin for the pkgs
		pluginPkgs := response.AvailablePackageSummaries
		for _, r := range pluginPkgs {
			if r.AvailablePackageRef == nil {
				r.AvailablePackageRef = &packages.AvailablePackageReference{}
			}
			r.AvailablePackageRef.Plugin = p.plugin
		}
		pkgs = append(pkgs, pluginPkgs...)
	}

	pkgsR := []*packages.AvailablePackageSummary{}

	if pageSize > 0 {
		// Using https://github.com/ahmetb/go-linq for simplicity
		From(pkgs).Skip(pageOffset*int(pageSize) - 1).Take(int(pageSize)).ToSlice(&pkgsR)
	}

	// Only return a next page token if the request was for pagination and
	// the results are a full page.
	nextPageToken := ""
	if pageSize > 0 && len(pkgsR) == int(pageSize) {
		nextPageToken = fmt.Sprintf("%d", pageOffset+1)
	}

	// TODO: Sort via default sort order or that specified in request.
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: pkgsR,
		Categories:                categories,
		NextPageToken:             nextPageToken,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s packagesServer) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	contextMsg := ""
	if request.AvailablePackageRef != nil && request.AvailablePackageRef.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.AvailablePackageRef.Context.Cluster, request.AvailablePackageRef.Context.Namespace)
	}

	log.Infof("+core GetAvailablePackageDetail %s", contextMsg)

	if request.AvailablePackageRef == nil || request.AvailablePackageRef.Plugin == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable get the GetAvailablePackageDetail from the plugin %v: %v", request.AvailablePackageRef.Plugin, err)
	}

	// Validate the plugin response
	if response.AvailablePackageDetail == nil || response.AvailablePackageDetail.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.Internal, "Invalid GetAvailablePackageDetail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetAvailablePackageDetailResponse{
		AvailablePackageDetail: response.AvailablePackageDetail,
	}, nil
}

// GetInstalledPackageSummaries returns the installed package summaries based on the request.
func (s packagesServer) GetInstalledPackageSummaries(ctx context.Context, request *packages.GetInstalledPackageSummariesRequest) (*packages.GetInstalledPackageSummariesResponse, error) {
	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+core GetInstalledPackageSummaries %s", contextMsg)

	// Aggregate the response for each plugin
	pkgs := []*packages.InstalledPackageSummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.plugins {
		response, err := p.server.GetInstalledPackageSummaries(ctx, request)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Invalid GetInstalledPackageSummaries response from the plugin %v: %v", p.plugin.Name, err)
		}

		// Add the plugin for the pkgs
		pluginPkgs := response.InstalledPackageSummaries
		for _, r := range pluginPkgs {
			if r.InstalledPackageRef == nil {
				r.InstalledPackageRef = &packages.InstalledPackageReference{}
			}
			r.InstalledPackageRef.Plugin = p.plugin
		}
		pkgs = append(pkgs, pluginPkgs...)
	}

	// Build the response
	// TODO: Sort via default sort order or that specified in request.
	return &packages.GetInstalledPackageSummariesResponse{
		InstalledPackageSummaries: pkgs,
	}, nil
}

// GetInstalledPackageDetail returns the package versions based on the request.
func (s packagesServer) GetInstalledPackageDetail(ctx context.Context, request *packages.GetInstalledPackageDetailRequest) (*packages.GetInstalledPackageDetailResponse, error) {
	contextMsg := ""
	if request.InstalledPackageRef != nil && request.InstalledPackageRef.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.InstalledPackageRef.Context.Cluster, request.InstalledPackageRef.Context.Namespace)
	}

	log.Infof("+core GetInstalledPackageDetail %s", contextMsg)

	if request.InstalledPackageRef == nil || request.InstalledPackageRef.Plugin == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing InstalledPackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.InstalledPackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable get the plugin %v", request.InstalledPackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetInstalledPackageDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable get the GetInstalledPackageDetail from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Validate the plugin response
	if response.InstalledPackageDetail == nil {
		return nil, status.Errorf(codes.Internal, "Invalid GetInstalledPackageDetail response from the plugin %v: %v", pluginWithServer.plugin.Name, err)
	}

	// Build the response
	return &packages.GetInstalledPackageDetailResponse{
		InstalledPackageDetail: response.InstalledPackageDetail,
	}, nil
}

// GetAvailablePackageVersions returns the package versions based on the request.
func (s packagesServer) GetAvailablePackageVersions(ctx context.Context, request *packages.GetAvailablePackageVersionsRequest) (*packages.GetAvailablePackageVersionsResponse, error) {
	contextMsg := ""
	if request.AvailablePackageRef != nil && request.AvailablePackageRef.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.AvailablePackageRef.Context.Cluster, request.AvailablePackageRef.Context.Namespace)
	}

	log.Infof("+core GetAvailablePackageVersions %s", contextMsg)

	if request.AvailablePackageRef == nil || request.AvailablePackageRef.Plugin == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable get the plugin %v", request.AvailablePackageRef.Plugin)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageVersions(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable get the GetAvailablePackageVersions from the plugin %v: %v", pluginWithServer.plugin.Name, err)
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

// TODO(agamez): implement CreateInstalledPackage core api operation too
// func (s packagesServer) CreateInstalledPackage(ctx context.Context, request *packages.CreateInstalledPackageRequest) (*packages.CreateInstalledPackageResponse, error) {
// 	return nil, nil
// }

// getPluginWithServer returns the *pkgsPluginWithServer from a given packagesServer
// matching the plugin name
func (s packagesServer) getPluginWithServer(plugin *v1alpha1.Plugin) *pkgsPluginWithServer {
	for _, p := range s.plugins {
		if plugin.Name == p.plugin.Name {
			return p
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
