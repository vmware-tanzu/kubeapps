// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
)

const (
	HelmRepoType          = "helm"
	OCIRepoType           = "oci"
	RedactedString        = "REDACTED"
	AppRepositoryResource = "apprepositories"
	AppRepositoryKind     = "AppRepository"
)

type HelmRepository struct {
	cluster      string
	name         types.NamespacedName
	url          string
	repoType     string
	description  string
	interval     string
	tlsConfig    *corev1.PackageRepositoryTlsConfig
	auth         *corev1.PackageRepositoryAuth
	customDetail *v1alpha1.HelmPackageRepositoryCustomDetail
}

var ValidRepoTypes = []string{HelmRepoType, OCIRepoType}

func (s *Server) newRepo(ctx context.Context, repo *HelmRepository) (*corev1.PackageRepositoryReference, error) {
	if repo.url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}
	if repo.repoType == "" || !slices.Contains(ValidRepoTypes, repo.repoType) {
		return nil, status.Errorf(codes.InvalidArgument, "repository type [%s] not supported", repo.repoType)
	}
	// Helm repositories do not support intervals for reconciliation by now
	if repo.interval != "" {
		return nil, status.Errorf(codes.InvalidArgument, "interval is not supported for Helm repositories")
	}
	typedClient, err := s.clientGetter.Typed(ctx, repo.cluster)
	if err != nil {
		return nil, err
	}

	// Get or validate secret resource for auth,
	// not yet stored in K8s
	var secret *k8scorev1.Secret
	if s.pluginConfig.UserManagedSecrets {
		if secret, err = validateUserManagedRepoSecret(ctx, typedClient, repo.name, repo.tlsConfig, repo.auth); err != nil {
			return nil, err
		}
	} else {
		if secret, _, err = newSecretFromTlsConfigAndAuth(repo.name, repo.tlsConfig, repo.auth); err != nil {
			return nil, err
		}
	}
	// Copy secret to the namespace of asset syncer if needed. See issue #5129.
	if repo.name.Namespace != s.kubeappsNamespace && secret != nil {
		// Target namespace must be the same as the asset syncer job,
		// otherwise the job won't be able to access the secret
		if _, err = s.copyRepositorySecretToNamespace(typedClient, s.kubeappsNamespace, secret, repo.name); err != nil {
			return nil, err
		}
	}

	// Handle imagesPullSecret if any
	imagePullSecret, _, err := handleImagesPullSecret(ctx, typedClient, s.pluginConfig.UserManagedSecrets, repo, true)
	if err != nil {
		return nil, err
	}

	// Map data to a repository CRD
	helmRepoCrd, err := newHelmRepoCrd(repo, secret, imagePullSecret)
	if err != nil {
		return nil, err
	}

	// Repository validation
	if repo.customDetail != nil && repo.customDetail.PerformValidation {
		if err = s.ValidateRepository(helmRepoCrd, secret); err != nil {
			return nil, err
		}
	}

	// Store secret if Kubeapps manages secrets
	if !s.pluginConfig.UserManagedSecrets {
		// a bit of catch 22: I need to create a secret first, so that I can create a repo that references it,
		// but then I need to set the owner reference on this secret to the repo. In has to be done
		// in that order because to set an owner ref you need object (i.e. repo) UID, which you only get
		// once the object's been created
		if secret, imagePullSecret, err = createKubeappsManagedRepoSecrets(ctx, typedClient, repo.name.Namespace, secret, imagePullSecret); err != nil {
			return nil, err
		}
	}

	// Create repository CRD in K8s
	if client, err := s.getClient(ctx, repo.cluster, repo.name.Namespace); err != nil {
		return nil, err
	} else if err = client.Create(ctx, helmRepoCrd); err != nil {
		return nil, statuserror.FromK8sError("create", AppRepositoryKind, repo.name.String(), err)
	} else {
		if !s.pluginConfig.UserManagedSecrets {
			if err = s.setOwnerReferencesForRepoSecret(ctx, secret, repo.cluster, helmRepoCrd); err != nil {
				return nil, err
			}
			if err = s.setOwnerReferencesForRepoSecret(ctx, imagePullSecret, repo.cluster, helmRepoCrd); err != nil {
				return nil, err
			}
		}
		return &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: helmRepoCrd.Namespace,
				Cluster:   repo.cluster,
			},
			Identifier: helmRepoCrd.Name,
			Plugin:     GetPluginDetail(),
		}, nil
	}
}

func newHelmRepoCrd(repo *HelmRepository, secret *k8scorev1.Secret, imagePullSecret *k8scorev1.Secret) (*apprepov1alpha1.AppRepository, error) {
	appRepoCrd := &apprepov1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repo.name.Name,
			Namespace: repo.name.Namespace,
		},
		Spec: apprepov1alpha1.AppRepositorySpec{
			URL:                   repo.url,
			Type:                  repo.repoType,
			TLSInsecureSkipVerify: repo.tlsConfig != nil && repo.tlsConfig.InsecureSkipVerify,
			Description:           repo.description,
			PassCredentials:       repo.auth != nil && repo.auth.PassCredentials,
		},
	}
	if repo.auth != nil || repo.tlsConfig != nil {
		if repoAuth, err := newAppRepositoryAuth(secret, repo.tlsConfig, repo.auth); err != nil {
			return nil, err
		} else if repoAuth != nil {
			appRepoCrd.Spec.Auth = *repoAuth
		}
	}
	if imagePullSecret != nil {
		appRepoCrd.Spec.DockerRegistrySecrets = []string{imagePullSecret.Name}
	} else {
		appRepoCrd.Spec.DockerRegistrySecrets = nil
	}
	if repo.customDetail != nil {
		if repo.customDetail.FilterRule != nil {
			appRepoCrd.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{
				JQ:        repo.customDetail.FilterRule.Jq,
				Variables: repo.customDetail.FilterRule.Variables,
			}
		}
		if repo.customDetail.OciRepositories != nil {
			appRepoCrd.Spec.OCIRepositories = repo.customDetail.OciRepositories
		}
	}
	return appRepoCrd, nil
}

func (s *Server) mapToPackageRepositoryDetail(source *apprepov1alpha1.AppRepository,
	cluster, namespace string,
	caSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret,
	imagesPullSecret *k8scorev1.Secret) (*corev1.PackageRepositoryDetail, error) {

	// Auth
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	var auth *corev1.PackageRepositoryAuth
	var err error
	if s.pluginConfig.UserManagedSecrets {
		if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithUserManagedSecrets(source, caSecret, authSecret); err != nil {
			return nil, err
		}
	} else {
		if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(source, caSecret, authSecret); err != nil {
			return nil, err
		}
	}

	target := &corev1.PackageRepositoryDetail{
		PackageRepoRef: &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Cluster:   cluster,
				Namespace: namespace,
			},
			Plugin:     GetPluginDetail(),
			Identifier: source.Name,
		},
		Name:            source.Name,
		Description:     source.Spec.Description,
		NamespaceScoped: s.GetGlobalPackagingNamespace() != namespace,
		Type:            source.Spec.Type,
		Url:             source.Spec.URL,
		Auth:            auth,
		TlsConfig:       tlsConfig,
		// TODO(agamez): check if we can get the status from the repo somehow
		// https://github.com/vmware-tanzu/kubeapps/issues/153
		Status: &corev1.PackageRepositoryStatus{
			Ready: true,
		},
	}

	// Custom details
	if source.Spec.DockerRegistrySecrets != nil || source.Spec.FilterRule.JQ != "" || source.Spec.OCIRepositories != nil {
		var customDetail = &v1alpha1.HelmPackageRepositoryCustomDetail{}

		if source.Spec.DockerRegistrySecrets != nil {
			// Set Docker image pull secrets
			customDetail.ImagesPullSecret = &v1alpha1.ImagesPullSecret{}
			if s.pluginConfig.UserManagedSecrets {
				customDetail.ImagesPullSecret.DockerRegistryCredentialOneOf = getRepoImagesPullSecretWithUserManagedSecrets(imagesPullSecret)
			} else {
				customDetail.ImagesPullSecret.DockerRegistryCredentialOneOf = getRepoImagesPullSecretWithKubeappsManagedSecrets(imagesPullSecret)
			}
		}
		if source.Spec.FilterRule.JQ != "" {
			customDetail.FilterRule = &v1alpha1.RepositoryFilterRule{
				Jq:        source.Spec.FilterRule.JQ,
				Variables: source.Spec.FilterRule.Variables,
			}
		}
		customDetail.OciRepositories = source.Spec.OCIRepositories
		target.CustomDetail, err = anypb.New(customDetail)
		if err != nil {
			return nil, err
		}
	}

	return target, nil
}

// Using owner references on the secret so that it can be
// (1) cleaned up automatically and/or
// (2) enable some control (ie. if I add a secret manually
//
//	via kubectl before running kubeapps, it won't get deleted just
//	because Kubeapps is deleting it)?
//
// See https://github.com/vmware-tanzu/kubeapps/pull/4630#discussion_r861446394 for details
func (s *Server) setOwnerReferencesForRepoSecret(
	ctx context.Context,
	secret *k8scorev1.Secret,
	cluster string,
	repo *apprepov1alpha1.AppRepository) error {

	if secret != nil {
		if typedClient, err := s.clientGetter.Typed(ctx, cluster); err != nil {
			return err
		} else {
			secretsInterface := typedClient.CoreV1().Secrets(repo.Namespace)
			secret.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(
					repo,
					schema.GroupVersionKind{
						Group:   apprepov1alpha1.SchemeGroupVersion.Group,
						Version: apprepov1alpha1.SchemeGroupVersion.Version,
						Kind:    AppRepositoryKind,
					}),
			}
			if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return statuserror.FromK8sError("update", "secrets", secret.Name, err)
			}
		}
	}
	return nil
}

func (s *Server) updateRepo(ctx context.Context, repo *HelmRepository) (*corev1.PackageRepositoryReference, error) {
	if repo.url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}
	if repo.name.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository name may not be empty")
	}
	// Helm repositories do not support intervals for reconciliation by now
	if repo.interval != "" {
		return nil, status.Errorf(codes.InvalidArgument, "interval is not supported for Helm repositories")
	}
	typedClient, err := s.clientGetter.Typed(ctx, repo.cluster)
	if err != nil {
		return nil, err
	}

	appRepo, caSecret, authSecret, err := s.getPkgRepository(ctx, repo.cluster, repo.name.Namespace, repo.name.Name)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "unable to retrieve repository '%s/%s' due to [%v]", repo.name.Namespace, repo.name.Namespace, err)
	}

	if authSecret != nil && caSecret != nil && authSecret.Name != caSecret.Name {
		return nil, status.Errorf(codes.Internal, "inconsistent state. auth secret and ca secret must be the same.")
	}
	var secretRef string
	if authSecret != nil {
		secretRef = authSecret.Name
	} else if caSecret != nil {
		secretRef = caSecret.Name
	}

	appRepo.Spec.URL = repo.url
	appRepo.Spec.Description = repo.description
	appRepo.Spec.TLSInsecureSkipVerify = repo.tlsConfig != nil && repo.tlsConfig.InsecureSkipVerify

	// Update secret if needed
	var secret *k8scorev1.Secret
	var updateRepoSecret bool
	if s.pluginConfig.UserManagedSecrets {
		if secret, err = validateUserManagedRepoSecret(ctx, typedClient, repo.name, repo.tlsConfig, repo.auth); err != nil {
			return nil, err
		}
	} else {
		if secret, updateRepoSecret, err = s.updateKubeappsManagedRepoSecret(ctx, repo, secretRef); err != nil {
			return nil, err
		}
	}
	// Copy secret to the namespace of asset syncer if needed. See issue #5129.
	if repo.name.Namespace != s.kubeappsNamespace && secret != nil {
		// Target namespace must be the same as the asset syncer job,
		// otherwise the job won't be able to access the secret
		if _, err = s.copyRepositorySecretToNamespace(typedClient, s.kubeappsNamespace, secret, repo.name); err != nil {
			return nil, err
		}
	}

	// Handle imagesPullSecret if any
	imagePullSecret, updateImgPullSecret, err := handleImagesPullSecret(ctx, typedClient, s.pluginConfig.UserManagedSecrets, repo, false)
	if err != nil {
		return nil, err
	}

	if s.pluginConfig.UserManagedSecrets || updateRepoSecret {
		if secret != nil {
			if repoAuth, err := newAppRepositoryAuth(secret, repo.tlsConfig, repo.auth); err != nil {
				return nil, err
			} else if repoAuth != nil {
				appRepo.Spec.Auth = *repoAuth
			}
		} else {
			appRepo.Spec.Auth.Header = nil
			appRepo.Spec.Auth.CustomCA = nil
		}
	}

	appRepo.Spec.PassCredentials = repo.auth != nil && repo.auth.PassCredentials

	if imagePullSecret != nil || (imagePullSecret != nil && !updateImgPullSecret) {
		appRepo.Spec.DockerRegistrySecrets = []string{imagePullSecret.Name}
	} else {
		appRepo.Spec.DockerRegistrySecrets = nil
	}

	// Custom details
	if repo.customDetail != nil {
		if repo.customDetail.FilterRule != nil {
			appRepo.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{
				JQ:        repo.customDetail.FilterRule.Jq,
				Variables: repo.customDetail.FilterRule.Variables,
			}
		} else {
			appRepo.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{}
		}
		appRepo.Spec.OCIRepositories = repo.customDetail.OciRepositories
	} else {
		appRepo.Spec.DockerRegistrySecrets = nil
		appRepo.Spec.OCIRepositories = nil
		appRepo.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{}
	}

	// persist repository
	err = s.updatePkgRepository(ctx, repo.cluster, repo.name.Namespace, appRepo)
	if err != nil {
		return nil, statuserror.FromK8sError("update", AppRepositoryKind, repo.name.String(), err)
	}

	if updateRepoSecret && secret != nil {
		// new secret => will need to set the owner
		if err = s.setOwnerReferencesForRepoSecret(ctx, secret, repo.cluster, appRepo); err != nil {
			return nil, err
		}
	}
	if imagePullSecret != nil && updateImgPullSecret {
		if err = s.setOwnerReferencesForRepoSecret(ctx, imagePullSecret, repo.cluster, appRepo); err != nil {
			return nil, err
		}
	}

	log.V(4).Infof("Updated AppRepository '%s' in namespace '%s' of cluster '%s'", repo.name.Name, repo.name.Namespace, repo.cluster)

	return &corev1.PackageRepositoryReference{
		Context: &corev1.Context{
			Namespace: repo.name.Namespace,
			Cluster:   repo.cluster,
		},
		Identifier: repo.name.Name,
		Plugin:     GetPluginDetail(),
	}, nil
}

func handleImagesPullSecret(ctx context.Context, typedClient kubernetes.Interface,
	isUserManagedSecrets bool, repo *HelmRepository, newRepo bool) (imgPullSecret *k8scorev1.Secret, updateImgPullSecret bool, err error) {
	if repo.customDetail == nil || repo.customDetail.ImagesPullSecret == nil {
		if !isUserManagedSecrets {
			managedSecretName := imagesPullSecretName(repo.name.Name)
			secretsInterface := typedClient.CoreV1().Secrets(repo.name.Namespace)
			if deleteErr := deleteSecret(ctx, secretsInterface, managedSecretName); deleteErr != nil {
				return nil, false, deleteErr
			}
		}
		return nil, false, nil
	}
	if isUserManagedSecrets {
		if repo.customDetail.ImagesPullSecret.GetSecretRef() != "" {
			// Validate existing images pull secret managed by user
			if validSecret, err := validateDockerImagePullSecret(ctx, typedClient, repo.name, repo.customDetail.ImagesPullSecret.GetSecretRef()); err != nil {
				return nil, false, err
			} else {
				return validSecret, false, nil
			}
		} else if repo.customDetail.ImagesPullSecret.GetCredentials() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "full credentials are not valid having user managed secrets")
		}
	} else {
		if repo.customDetail.ImagesPullSecret.GetCredentials() != nil {
			if newRepo {
				// Create a new secret due to new repo
				if newSecret, _, err := newDockerImagePullSecret(repo.name, repo.customDetail.ImagesPullSecret.GetCredentials()); err != nil {
					return nil, false, err
				} else {
					return newSecret, false, nil
				}
			} else {
				// When updating repo check if secret needs creation, update or removal
				if updatedSecret, isSameSecret, err := updateKubeappsManagedImagesPullSecret(ctx, typedClient, repo.name, repo.customDetail.ImagesPullSecret.GetCredentials()); err != nil {
					return nil, false, err
				} else {
					return updatedSecret, !isSameSecret, nil
				}
			}
		} else if repo.customDetail.ImagesPullSecret.GetSecretRef() != "" {
			return nil, false, status.Errorf(codes.InvalidArgument, "secret ref is not valid having kubeapps managed secrets")
		}
	}
	return nil, false, nil
}

func (s *Server) repoSummaries(ctx context.Context, cluster string, namespace string) ([]*corev1.PackageRepositorySummary, error) {
	var summaries []*corev1.PackageRepositorySummary

	repos, err := s.GetPkgRepositories(ctx, cluster, namespace)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "AppRepository", "", err)
	}

	for _, repo := range repos {
		summary := &corev1.PackageRepositorySummary{
			PackageRepoRef: &corev1.PackageRepositoryReference{
				Context: &corev1.Context{
					Namespace: repo.Namespace,
					Cluster:   cluster,
				},
				Identifier: repo.Name,
				Plugin:     GetPluginDetail(),
			},
			Name:            repo.Name,
			Description:     repo.Spec.Description,
			NamespaceScoped: s.GetGlobalPackagingNamespace() != repo.Namespace,
			Type:            repo.Spec.Type,
			Url:             repo.Spec.URL,
			RequiresAuth:    repo.Spec.Auth.Header != nil,
			// TODO(agamez): check if we can get the status from the repo somehow
			// https://github.com/vmware-tanzu/kubeapps/issues/153
			Status: &corev1.PackageRepositoryStatus{
				Ready: true,
			},
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// GetPkgRepositories returns the list of package repositories for the given cluster and namespace
func (s *Server) GetPkgRepositories(ctx context.Context, cluster, namespace string) ([]*apprepov1alpha1.AppRepository, error) {
	resource, err := s.getPkgRepositoryResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	// TODO(agamez): handle permission denied scenario when listing w/o namespace, which would need a ClusterRole
	unstructured, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var pkgRepositories []*apprepov1alpha1.AppRepository
	for _, item := range unstructured.Items {
		pkgRepository := &apprepov1alpha1.AppRepository{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, pkgRepository)
		if err != nil {
			return nil, err
		}
		pkgRepositories = append(pkgRepositories, pkgRepository)
	}
	return pkgRepositories, nil
}

func (s *Server) deleteRepo(ctx context.Context, cluster string, repoRef *corev1.PackageRepositoryReference) error {
	client, err := s.getClient(ctx, cluster, repoRef.Context.Namespace)
	if err != nil {
		return err
	}

	log.V(4).Infof("Deleting AppRepository: [%s]", repoRef.Identifier)

	// For kubeapps-managed secrets environment secrets will be deleted (garbage-collected)
	// when the owner repo is deleted

	repo := &apprepov1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      repoRef.Identifier,
			Namespace: repoRef.Context.Namespace,
		},
	}
	if err = client.Delete(ctx, repo); err != nil {
		return statuserror.FromK8sError("delete", AppRepositoryKind, repoRef.Identifier, err)
	} else {
		// Cross-namespace owner references are disallowed by design.
		// We need to explicitly delete the repo secret from the namespace of the asset syncer.
		typedClient, err := s.clientGetter.Typed(ctx, cluster)
		if err != nil {
			return err
		}
		namespacedSecretName := namespacedSecretNameForRepo(repoRef.Identifier, repoRef.Context.Namespace)
		if deleteErr := s.deleteRepositorySecretFromNamespace(typedClient, s.kubeappsNamespace, namespacedSecretName); deleteErr != nil {
			return statuserror.FromK8sError("delete", "Secret", namespacedSecretName, deleteErr)
		}
		return nil
	}
}
