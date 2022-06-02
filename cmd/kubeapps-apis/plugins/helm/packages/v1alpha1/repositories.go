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
	cluster       string
	name          types.NamespacedName
	url           string
	repoType      string
	description   string
	interval      uint32
	tlsConfig     *corev1.PackageRepositoryTlsConfig
	auth          *corev1.PackageRepositoryAuth
	customDetails *v1alpha1.RepositoryCustomDetails
}

var ValidRepoTypes = []string{HelmRepoType, OCIRepoType}

func (s *Server) newRepo(ctx context.Context, repo *HelmRepository) (*corev1.PackageRepositoryReference, error) {
	if repo.url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}
	if repo.repoType == "" || !slices.Contains(ValidRepoTypes, repo.repoType) {
		return nil, status.Errorf(codes.InvalidArgument, "repository type [%s] not supported", repo.repoType)
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

	// Map data to a repository CRD
	helmRepoCrd, err := newHelmRepoCrd(repo, secret)
	if err != nil {
		return nil, err
	}

	// Repository validation
	if repo.customDetails != nil && repo.customDetails.PerformValidation {
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
		if secret, err = createKubeappsManagedRepoSecret(ctx, typedClient, repo.name.Namespace, secret); err != nil {
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

func newHelmRepoCrd(repo *HelmRepository, secret *k8scorev1.Secret) (*apprepov1alpha1.AppRepository, error) {
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
	if repo.customDetails != nil {
		if repo.customDetails.DockerRegistrySecrets != nil {
			appRepoCrd.Spec.DockerRegistrySecrets = repo.customDetails.DockerRegistrySecrets
		}
		if repo.customDetails.FilterRule != nil {
			appRepoCrd.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{
				JQ:        repo.customDetails.FilterRule.Jq,
				Variables: repo.customDetails.FilterRule.Variables,
			}
		}
		if repo.customDetails.OciRepositories != nil {
			appRepoCrd.Spec.OCIRepositories = repo.customDetails.OciRepositories
		}
	}
	return appRepoCrd, nil
}

func (s *Server) mapToPackageRepositoryDetail(source *apprepov1alpha1.AppRepository,
	cluster, namespace string,
	caSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret) (*corev1.PackageRepositoryDetail, error) {

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
		NamespaceScoped: s.globalPackagingNamespace != namespace,
		Type:            source.Spec.Type,
		Url:             source.Spec.URL,
		Auth:            auth,
		TlsConfig:       tlsConfig,
	}

	// Custom details
	if source.Spec.DockerRegistrySecrets != nil || source.Spec.FilterRule.JQ != "" || source.Spec.OCIRepositories != nil {
		var customDetails = &v1alpha1.RepositoryCustomDetails{}
		customDetails.DockerRegistrySecrets = source.Spec.DockerRegistrySecrets
		if source.Spec.FilterRule.JQ != "" {
			customDetails.FilterRule = &v1alpha1.RepositoryFilterRule{
				Jq:        source.Spec.FilterRule.JQ,
				Variables: source.Spec.FilterRule.Variables,
			}
		}
		customDetails.OciRepositories = source.Spec.OCIRepositories
		target.CustomDetail, err = anypb.New(customDetails)
		if err != nil {
			return nil, err
		}
	}

	return target, nil
}

// Using owner references on the secret so that it can be
// (1) cleaned up automatically and/or
// (2) enable some control (ie. if I add a secret manually
//   via kubectl before running kubeapps, it won't get deleted just
//   because Kubeapps is deleting it)?
// See https://github.com/vmware-tanzu/kubeapps/pull/4630#discussion_r861446394 for details
func (s *Server) setOwnerReferencesForRepoSecret(
	ctx context.Context,
	secret *k8scorev1.Secret,
	cluster string,
	repo *apprepov1alpha1.AppRepository) error {

	if (repo.Spec.Auth.Header != nil || repo.Spec.Auth.CustomCA != nil) && secret != nil {
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

	// Custom details
	if repo.customDetails != nil {
		appRepo.Spec.DockerRegistrySecrets = repo.customDetails.DockerRegistrySecrets
		if repo.customDetails.FilterRule != nil {
			appRepo.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{
				JQ:        repo.customDetails.FilterRule.Jq,
				Variables: repo.customDetails.FilterRule.Variables,
			}
		} else {
			appRepo.Spec.FilterRule = apprepov1alpha1.FilterRuleSpec{}
		}
		appRepo.Spec.OCIRepositories = repo.customDetails.OciRepositories
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
			NamespaceScoped: s.globalPackagingNamespace != repo.Namespace,
			Type:            repo.Spec.Type,
			Url:             repo.Spec.URL,
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
		return nil
	}
}
