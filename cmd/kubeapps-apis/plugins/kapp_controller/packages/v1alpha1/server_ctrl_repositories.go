// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"
)

// AddPackageRepository adds a package repository managed by the 'kapp_controller' plugin
func (s *Server) AddPackageRepository(ctx context.Context, request *corev1.AddPackageRepositoryRequest) (*corev1.AddPackageRepositoryResponse, error) {
	// context info
	cluster := request.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}

	// trace logging
	log.InfoS("+kapp-controller AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", request.GetName())

	// validation
	if cluster != s.globalPackagingCluster {
		return nil, status.Errorf(codes.InvalidArgument, "installing package repositories in other clusters in not supported yet")
	}
	if err := s.validatePackageRepositoryCreate(ctx, cluster, request); err != nil {
		return nil, err
	}

	// create secret (must be done first, to get the name)
	var err error
	var pkgSecret *k8scorev1.Secret
	if request.Auth != nil && request.Auth.GetSecretRef() == nil {
		pkgSecret, err = s.buildPkgRepositorySecretCreate(namespace, request.Name, request.Auth)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "unable to build the associated secret: %v", err)
		}

		pkgSecret, err = s.createSecret(ctx, cluster, pkgSecret)
		if err != nil {
			return nil, statuserror.FromK8sError("create", "Secret", request.Name, err)
		}
	}

	// create repository
	pkgRepository, err := s.buildPkgRepositoryCreate(request, pkgSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to build the PackageRepository: %v", err)
	}
	pkgRepository, err = s.createPkgRepository(ctx, cluster, namespace, pkgRepository)
	if err != nil {
		return nil, statuserror.FromK8sError("create", "PackageRepository", request.Name, err)
	}

	// update secret with owner reference if needed
	if pkgSecret != nil {
		setOwnerReference(pkgSecret, pkgRepository)
		_, err = s.updateSecret(ctx, cluster, pkgSecret)
		if err != nil {
			return nil, statuserror.FromK8sError("update", "Secret", request.Name, err)
		}
	}

	// response
	response := &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: request.Name,
		},
	}

	log.InfoS("-kapp-controller AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", request.GetName())
	return response, nil
}

// GetPackageRepositoryDetail returns the package repository metadata managed by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositoryDetail(ctx context.Context, request *corev1.GetPackageRepositoryDetailRequest) (*corev1.GetPackageRepositoryDetailResponse, error) {
	// context info
	cluster := request.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller GetPackageRepositoryDetail", "cluster", cluster, "namespace", namespace, "name", name)

	// fetch repository
	pkgRepository, err := s.getPkgRepository(ctx, cluster, namespace, name)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageRepository", name, err)
	}

	// fetch repository secret
	var pkgSecret *k8scorev1.Secret
	if pkgSecretRef := repositorySecretRef(pkgRepository); pkgSecretRef != nil {
		pkgSecret, err = s.getSecret(ctx, cluster, namespace, pkgSecretRef.Name)
		if err != nil {
			return nil, statuserror.FromK8sError("get", "Secret", pkgSecretRef.Name, err)
		}
	}

	// translate
	repository, err := s.buildPackageRepository(pkgRepository, pkgSecret, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to convert the PackageRepository: %v", err)
	}

	// response
	response := &corev1.GetPackageRepositoryDetailResponse{
		Detail: repository,
	}

	log.InfoS("-kapp-controller GetPackageRepositoryDetail", "cluster", cluster, "namespace", namespace, "name", name)
	return response, nil
}

// GetPackageRepositorySummaries returns the package repositories managed by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositorySummaries(ctx context.Context, request *corev1.GetPackageRepositorySummariesRequest) (*corev1.GetPackageRepositorySummariesResponse, error) {
	// context info
	cluster := request.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.GetContext().GetNamespace()

	// trace logging
	log.InfoS("+kapp-controller GetPackageRepositories", "cluster", cluster, "namespace", namespace)

	// retrieve the list of installed packages
	pkgRepositories, err := s.getPkgRepositories(ctx, cluster, namespace)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageRepository", "", err)
	}

	// convert the Carvel PackageRepository to our API PackageRepository struct
	repositories := []*corev1.PackageRepositorySummary{}
	for _, repo := range pkgRepositories {
		repo, err := s.buildPackageRepositorySummary(repo, cluster)
		if err != nil {
			// todo -> instead of failing the whole query, we should be able to log the error along with the response
			return nil, status.Errorf(codes.Internal, "unable to convert the PackageRepository: %v", err)
		}
		repositories = append(repositories, repo)
	}

	// response
	response := &corev1.GetPackageRepositorySummariesResponse{
		PackageRepositorySummaries: repositories,
	}

	log.InfoS("-kapp-controller GetPackageRepositories", "cluster", cluster, "namespace", namespace)
	return response, nil
}

// UpdatePackageRepository updates a package repository managed by the 'kapp_controller' plugin
func (s *Server) UpdatePackageRepository(ctx context.Context, request *corev1.UpdatePackageRepositoryRequest) (*corev1.UpdatePackageRepositoryResponse, error) {
	// context info
	cluster := request.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller UpdatePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)

	// identity validation
	if cluster != s.globalPackagingCluster {
		return nil, status.Errorf(codes.InvalidArgument, "updating package repositories in other clusters in not supported yet")
	}
	if name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "no request Name provided")
	}

	// fetch existing repository
	pkgRepository, err := s.getPkgRepository(ctx, cluster, namespace, name)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "PackageRepository", name, err)
	}

	// fetch existing secret
	var pkgSecret *k8scorev1.Secret
	if pkgSecretRef := repositorySecretRef(pkgRepository); pkgSecretRef != nil {
		pkgSecret, err = s.getSecret(ctx, cluster, namespace, pkgSecretRef.Name)
		if err != nil {
			return nil, statuserror.FromK8sError("get", "Secret", pkgSecretRef.Name, err)
		}
	}

	// validate for update
	if err := s.validatePackageRepositoryUpdate(ctx, cluster, request, pkgRepository, pkgSecret); err != nil {
		return nil, err
	}

	// handle managed secret, there 3 cases to consider:
	//    create the secret if auth was not previously configured
	//    update the secret if auth has been updated
	//    delete the secret if auth has been removed
	if request.Auth == nil {
		// delete existing secret, if plugin managed
		if pkgSecret != nil && isPluginManaged(pkgRepository, pkgSecret) {
			if err := s.deleteSecret(ctx, cluster, pkgSecret.GetNamespace(), pkgSecret.GetName()); err != nil {
				return nil, statuserror.FromK8sError("delete", "Secret", pkgSecret.GetName(), err)
			}
		}
		pkgSecret = nil
	} else if request.Auth.GetSecretRef() == nil {
		if pkgSecret == nil {
			// create
			pkgSecret, err = s.buildPkgRepositorySecretCreate(namespace, pkgRepository.GetName(), request.Auth)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to build the associated secret: %v", err)
			}
			// repository already exist, we can set the owner reference as part of creation
			setOwnerReference(pkgSecret, pkgRepository)

			pkgSecret, err = s.createSecret(ctx, cluster, pkgSecret)
			if err != nil {
				return nil, statuserror.FromK8sError("create", "Secret", pkgRepository.GetName(), err)
			}
		} else {
			// update
			updated, err := s.buildPkgRepositorySecretUpdate(pkgSecret, request.Auth)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to build the associated secret: %v", err)
			}
			if updated {
				pkgSecret, err = s.updateSecret(ctx, cluster, pkgSecret)
				if err != nil {
					return nil, statuserror.FromK8sError("update", "Secret", pkgRepository.GetName(), err)
				}
			}
		}
	}

	// update repository
	pkgRepository, err = s.buildPkgRepositoryUpdate(request, pkgRepository, pkgSecret)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to build the PackageRepository: %v", err)
	}
	_, err = s.updatePkgRepository(ctx, cluster, namespace, pkgRepository)
	if err != nil {
		return nil, statuserror.FromK8sError("update", "PackageRepository", name, err)
	}

	// response
	response := &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: request.GetPackageRepoRef().GetIdentifier(),
		},
	}

	log.InfoS("-kapp-controller UpdatePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)
	return response, nil
}

// DeletePackageRepository deletes a package repository managed by the 'kapp_controller' plugin
func (s *Server) DeletePackageRepository(ctx context.Context, request *corev1.DeletePackageRepositoryRequest) (*corev1.DeletePackageRepositoryResponse, error) {
	// context info
	cluster := request.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller DeletePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)

	// delete
	err := s.deletePkgRepository(ctx, cluster, namespace, name)
	if err != nil {
		return nil, statuserror.FromK8sError("delete", "PackageRepository", name, err)
	}

	// response
	response := &corev1.DeletePackageRepositoryResponse{}

	log.InfoS("-kapp-controller DeletePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)
	return response, nil
}
