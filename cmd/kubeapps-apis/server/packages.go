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

	pkgs := []*packages.AvailablePackageSummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.plugins {
		response, err := p.server.GetAvailablePackageSummaries(ctx, request)
		if err != nil {
			return nil, err
		}

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

	// TODO: Sort via default sort order or that specified in request.
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: pkgs,
	}, nil
}

// GetAvailablePackageDetail returns the package details based on the request.
func (s packagesServer) GetAvailablePackageDetail(ctx context.Context, request *packages.GetAvailablePackageDetailRequest) (*packages.GetAvailablePackageDetailResponse, error) {
	contextMsg := ""
	if request.AvailablePackageRef != nil && request.AvailablePackageRef.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.AvailablePackageRef.Context.Cluster, request.AvailablePackageRef.Context.Namespace)
	}

	log.Infof("+core GetAvailablePackageDetail %s", contextMsg)

	// Check prerequsites
	if request.AvailablePackageRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the available package reference (missing AvailablePackageRef)")
	}
	if request.AvailablePackageRef.Context == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the context (missing AvailablePackageRef.Context)")
	}
	if request.AvailablePackageRef.Identifier == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the identifier (missing AvailablePackageRef.Identifier)")
	}
	if request.AvailablePackageRef.Plugin == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Unable to retrieve the plugin (missing AvailablePackageRef.Plugin)")
	}

	// Retrieve the plugin with server matching the requested plugin name
	pluginWithServer := s.getPluginWithServer(request.AvailablePackageRef.Plugin)
	if pluginWithServer == nil {
		return nil, status.Errorf(codes.Internal, "Unable get the plugin %v", pluginWithServer.plugin.Name)
	}

	// Get the response from the requested plugin
	response, err := pluginWithServer.server.GetAvailablePackageDetail(ctx, request)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable get the GetAvailablePackageDetail from the plugin %v: %v", pluginWithServer.plugin.Name, err)
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
