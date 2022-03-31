// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	log "k8s.io/klog/v2"
)

// GetPackageRepositories returns the package repositories based on the request managed by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositories(ctx context.Context, request *v1alpha1.GetPackageRepositoriesRequest) (*v1alpha1.GetPackageRepositoriesResponse, error) {
	log.Infof("+kapp-controller GetPackageRepositories")

	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}

	// retrieve the list of installed packages
	pkgRepositories, err := s.getPkgRepositories(ctx, cluster, namespace)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageRepository", "", err)
	}

	// convert the Carvel PackageRepository to our API PackageRepository struct
	responseRepos := []*v1alpha1.PackageRepository{}
	for _, repo := range pkgRepositories {
		repo, err := getPackageRepository(repo)
		if err != nil {
			return nil, fmt.Errorf("unable to create the PackageRepository: %v", err)
		}
		responseRepos = append(responseRepos, repo)
	}

	return &v1alpha1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}
