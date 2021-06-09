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

	packages "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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
	pkgs := []*packages.AvailablePackageSummary{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.plugins {
		response, err := p.server.GetAvailablePackageSummaries(ctx, request)
		if err != nil {
			return nil, err
		}

		// Add the plugin for the pkgs
		pluginPkgs := response.AvailablePackagesSummaries
		for _, r := range pluginPkgs {
			r.AvailablePackageRef.Plugin = p.plugin
		}

		pkgs = append(pkgs, pluginPkgs...)
	}

	// TODO: Sort via default sort order or that specified in request.
	return &packages.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: pkgs,
	}, nil
}

// GetPackageRepositories returns the package repositories based on the request.
func (s packagesServer) GetPackageRepositories(ctx context.Context, request *packages.GetPackageRepositoriesRequest) (*packages.GetPackageRepositoriesResponse, error) {
	repos := []*packages.PackageRepository{}
	// TODO: We can do these in parallel in separate go routines.
	for _, p := range s.plugins {
		response, err := p.server.GetPackageRepositories(ctx, request)
		if err != nil {
			return nil, err
		}

		// Add the plugin for the repos
		pluginRepos := response.Repositories
		for _, r := range pluginRepos {
			r.Plugin = p.plugin
		}

		repos = append(repos, pluginRepos...)
	}

	// TODO: Sort via default sort order or that specified in request.
	return &packages.GetPackageRepositoriesResponse{
		Repositories: repos,
	}, nil
}
