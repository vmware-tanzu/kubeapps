// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/resources"
	"k8s.io/apimachinery/pkg/runtime/schema"

	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	k8scorev1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"
)

// AddPackageRepository adds a package repository managed by the 'kapp_controller' plugin
func (s *Server) AddPackageRepository(ctx context.Context, request *connect.Request[corev1.AddPackageRepositoryRequest]) (*connect.Response[corev1.AddPackageRepositoryResponse], error) {
	// context info
	cluster := request.Msg.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.Msg.GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}

	// trace logging
	log.InfoS("+kapp-controller AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", request.Msg.GetName())

	// validation
	if cluster != s.globalPackagingCluster {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Installing package repositories in other clusters in not supported yet"))
	}
	if err := s.validatePackageRepositoryCreate(ctx, cluster, request); err != nil {
		return nil, err
	}

	// create secret (must be done first, to get the name)
	var err error
	var pkgSecret *k8scorev1.Secret
	if request.Msg.Auth != nil && request.Msg.Auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED && request.Msg.Auth.GetSecretRef() == nil {
		pkgSecret, err = s.buildPkgRepositorySecretCreate(namespace, request.Msg.Name, request.Msg.Auth)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to build the associated secret: %w", err))
		}

		pkgSecret, err = s.createSecret(ctx, request.Header(), cluster, pkgSecret)
		if err != nil {
			return nil, connecterror.FromK8sError("create", "Secret", request.Msg.Name, err)
		}
	}

	// create repository
	pkgRepository, err := s.buildPkgRepositoryCreate(request.Msg, pkgSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to build the PackageRepository: %w", err))
	}
	pkgRepository, err = s.createPkgRepository(ctx, request.Header(), cluster, namespace, pkgRepository)
	if err != nil {
		return nil, connecterror.FromK8sError("create", "PackageRepository", request.Msg.Name, err)
	}

	// update secret with owner reference if needed
	if pkgSecret != nil {
		setOwnerReference(pkgSecret, pkgRepository)
		_, err = s.updateSecret(ctx, request.Header(), cluster, pkgSecret)
		if err != nil {
			return nil, connecterror.FromK8sError("update", "Secret", request.Msg.Name, err)
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
			Identifier: request.Msg.Name,
		},
	}

	log.InfoS("-kapp-controller AddPackageRepository", "cluster", cluster, "namespace", namespace, "name", request.Msg.GetName())
	return connect.NewResponse(response), nil
}

// GetPackageRepositoryDetail returns the package repository metadata managed by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositoryDetail(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryDetailRequest]) (*connect.Response[corev1.GetPackageRepositoryDetailResponse], error) {
	// context info
	cluster := request.Msg.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.Msg.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.Msg.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller GetPackageRepositoryDetail", "cluster", cluster, "namespace", namespace, "name", name)

	// fetch repository
	pkgRepository, err := s.getPkgRepository(ctx, request.Header(), cluster, namespace, name)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageRepository", name, err)
	}

	// fetch repository secret
	var pkgSecret *k8scorev1.Secret
	if pkgSecretRef := repositorySecretRef(pkgRepository); pkgSecretRef != nil {
		pkgSecret, err = s.getSecret(ctx, request.Header(), cluster, namespace, pkgSecretRef.Name)
		if err != nil {
			return nil, connecterror.FromK8sError("get", "Secret", pkgSecretRef.Name, err)
		}
	}

	// translate
	repository, err := s.buildPackageRepository(pkgRepository, pkgSecret, cluster)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to convert the PackageRepository: %w", err))
	}

	// response
	response := &corev1.GetPackageRepositoryDetailResponse{
		Detail: repository,
	}

	log.InfoS("-kapp-controller GetPackageRepositoryDetail", "cluster", cluster, "namespace", namespace, "name", name)
	return connect.NewResponse(response), nil
}

// GetPackageRepositorySummaries returns the package repositories managed by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositorySummaries(ctx context.Context, request *connect.Request[corev1.GetPackageRepositorySummariesRequest]) (*connect.Response[corev1.GetPackageRepositorySummariesResponse], error) {
	// context info
	cluster := request.Msg.GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.Msg.GetContext().GetNamespace()

	// trace logging
	log.InfoS("+kapp-controller GetPackageRepositories", "cluster", cluster, "namespace", namespace)

	// retrieve the list of repositories
	var pkgRepositories []*packagingv1alpha1.PackageRepository
	if namespace == "" {
		// find globally, either via cluster access or by enumerating through namespaces
		if repos, err := s.getPkgRepositories(ctx, request.Header(), cluster, ""); err == nil {
			pkgRepositories = append(pkgRepositories, repos...)
		} else {
			log.Warningf("+kapp-controller unable to list package repositories at the cluster scope in '%s' due to [%v]", cluster, err)
			if repos, err = s.getAccessiblePackageRepositories(ctx, request.Header(), cluster); err == nil {
				pkgRepositories = append(pkgRepositories, repos...)
			} else {
				return nil, err
			}
		}
	} else {
		// include namespace specific  repositories
		if repos, err := s.getPkgRepositories(ctx, request.Header(), cluster, namespace); err == nil {
			pkgRepositories = append(pkgRepositories, repos...)
		} else {
			return nil, err
		}

		// try to also include global repositories
		if namespace != s.pluginConfig.globalPackagingNamespace {
			if repos, err := s.getPkgRepositories(ctx, request.Header(), cluster, s.pluginConfig.globalPackagingNamespace); err == nil {
				pkgRepositories = append(pkgRepositories, repos...)
			}
		}
	}

	// convert the Carvel PackageRepository to our API PackageRepository struct
	var repositories []*corev1.PackageRepositorySummary
	for _, repo := range pkgRepositories {
		repo, err := s.buildPackageRepositorySummary(repo, cluster)
		if err != nil {
			// todo -> instead of failing the whole query, we should be able to log the error along with the response
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to convert the PackageRepository: %w", err))
		}
		repositories = append(repositories, repo)
	}

	// response
	response := &corev1.GetPackageRepositorySummariesResponse{
		PackageRepositorySummaries: repositories,
	}

	log.InfoS("-kapp-controller GetPackageRepositories", "cluster", cluster, "namespace", namespace)
	return connect.NewResponse(response), nil
}

// UpdatePackageRepository updates a package repository managed by the 'kapp_controller' plugin
func (s *Server) UpdatePackageRepository(ctx context.Context, request *connect.Request[corev1.UpdatePackageRepositoryRequest]) (*connect.Response[corev1.UpdatePackageRepositoryResponse], error) {
	// context info
	cluster := request.Msg.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.Msg.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.Msg.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller UpdatePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)

	// identity validation
	if cluster != s.globalPackagingCluster {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Updating package repositories in other clusters in not supported yet"))
	}
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("No request Name provided"))
	}

	// fetch existing repository
	pkgRepository, err := s.getPkgRepository(ctx, request.Header(), cluster, namespace, name)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "PackageRepository", name, err)
	}

	// fetch existing secret
	var pkgSecret *k8scorev1.Secret
	if pkgSecretRef := repositorySecretRef(pkgRepository); pkgSecretRef != nil {
		pkgSecret, err = s.getSecret(ctx, request.Header(), cluster, namespace, pkgSecretRef.Name)
		if err != nil {
			return nil, connecterror.FromK8sError("get", "Secret", pkgSecretRef.Name, err)
		}
	}

	// validate for update
	if err := s.validatePackageRepositoryUpdate(ctx, cluster, request, pkgRepository, pkgSecret); err != nil {
		return nil, err
	}

	// handle managed secret, there are 3 cases to consider:
	//    create the secret if auth was not previously configured
	//    update the secret if auth has been updated
	//    delete the secret if auth has been removed
	if request.Msg.Auth == nil || request.Msg.Auth.Type == corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		// delete existing secret, if plugin managed
		if pkgSecret != nil && isPluginManaged(pkgRepository, pkgSecret) {
			if err := s.deleteSecret(ctx, request.Header(), cluster, pkgSecret.GetNamespace(), pkgSecret.GetName()); err != nil {
				return nil, connecterror.FromK8sError("delete", "Secret", pkgSecret.GetName(), err)
			}
		}
		pkgSecret = nil
	} else if request.Msg.Auth.GetSecretRef() == nil {
		// build new secret
		var newSecret *k8scorev1.Secret
		if pkgSecret == nil {
			if newSecret, err = s.buildPkgRepositorySecretCreate(namespace, name, request.Msg.Auth); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to build the associated secret: %w", err))
			}
		} else {
			if newSecret, err = s.buildPkgRepositorySecretUpdate(pkgSecret, namespace, name, request.Msg.Auth); err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to build the associated secret: %w", err))
			}
		}

		// secret was updated, perform update via delete+create
		if newSecret != nil {
			// delete old one
			if pkgSecret != nil {
				if err := s.deleteSecret(ctx, request.Header(), cluster, pkgSecret.GetNamespace(), pkgSecret.GetName()); err != nil {
					log.Errorf("Error deleting existing secret: [%s] due to %v", pkgSecret.GetName(), err)
				}
				pkgSecret = nil
			}

			// create new one
			setOwnerReference(newSecret, pkgRepository)
			if pkgSecret, err = s.createSecret(ctx, request.Header(), cluster, newSecret); err != nil {
				return nil, connecterror.FromK8sError("create", "Secret", name, err)
			}
		}
	}

	// update repository
	pkgRepository, err = s.buildPkgRepositoryUpdate(request.Msg, pkgRepository, pkgSecret)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to build the PackageRepository: %w", err))
	}
	_, err = s.updatePkgRepository(ctx, request.Header(), cluster, namespace, pkgRepository)
	if err != nil {
		return nil, connecterror.FromK8sError("update", "PackageRepository", name, err)
	}

	// response
	response := &corev1.UpdatePackageRepositoryResponse{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: request.Msg.GetPackageRepoRef().GetIdentifier(),
		},
	}

	log.InfoS("-kapp-controller UpdatePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)
	return connect.NewResponse(response), nil
}

// DeletePackageRepository deletes a package repository managed by the 'kapp_controller' plugin
func (s *Server) DeletePackageRepository(ctx context.Context, request *connect.Request[corev1.DeletePackageRepositoryRequest]) (*connect.Response[corev1.DeletePackageRepositoryResponse], error) {
	// context info
	cluster := request.Msg.GetPackageRepoRef().GetContext().GetCluster()
	if cluster == "" {
		cluster = s.globalPackagingCluster
	}
	namespace := request.Msg.GetPackageRepoRef().GetContext().GetNamespace()
	if namespace == "" {
		namespace = s.pluginConfig.globalPackagingNamespace
	}
	name := request.Msg.GetPackageRepoRef().GetIdentifier()

	// trace logging
	log.InfoS("+kapp-controller DeletePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)

	// delete
	err := s.deletePkgRepository(ctx, request.Header(), cluster, namespace, name)
	if err != nil {
		return nil, connecterror.FromK8sError("delete", "PackageRepository", name, err)
	}

	// response
	response := &corev1.DeletePackageRepositoryResponse{}

	log.InfoS("-kapp-controller DeletePackageRepository", "cluster", cluster, "namespace", namespace, "name", name)
	return connect.NewResponse(response), nil
}

// GetPackageRepositoryPermissions provides permissions available to manage package repository by the 'kapp_controller' plugin
func (s *Server) GetPackageRepositoryPermissions(ctx context.Context, request *connect.Request[corev1.GetPackageRepositoryPermissionsRequest]) (*connect.Response[corev1.GetPackageRepositoryPermissionsResponse], error) {
	log.Infof("+kapp-controller GetPackageRepositoryPermissions [%v]", request)

	cluster := request.Msg.GetContext().GetCluster()
	namespace := request.Msg.GetContext().GetNamespace()
	if cluster == "" && namespace != "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Cluster must be specified when namespace is present: %s", namespace))
	}
	typedClient, err := s.clientGetter.Typed(request.Header(), cluster)
	if err != nil {
		return nil, err
	}

	resource := schema.GroupResource{
		Group:    packagingv1alpha1.SchemeGroupVersion.Group,
		Resource: pkgRepositoriesResource,
	}

	permissions := &corev1.PackageRepositoriesPermissions{
		Plugin: GetPluginDetail(),
	}

	// Global permissions
	permissions.Global, err = resources.GetPermissionsOnResource(ctx, typedClient, resource, s.pluginConfig.globalPackagingNamespace)
	if err != nil {
		return nil, err
	}

	// Namespace permissions
	if namespace != "" {
		permissions.Namespace, err = resources.GetPermissionsOnResource(ctx, typedClient, resource, request.Msg.GetContext().GetNamespace())
		if err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&corev1.GetPackageRepositoryPermissionsResponse{
		Permissions: []*corev1.PackageRepositoriesPermissions{permissions},
	}), nil
}
