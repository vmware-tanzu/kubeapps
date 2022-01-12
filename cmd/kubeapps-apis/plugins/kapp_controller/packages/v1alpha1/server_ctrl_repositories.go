// /*
// Copyright © 2021 VMware
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
package main

import (
	"context"
	"fmt"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
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
