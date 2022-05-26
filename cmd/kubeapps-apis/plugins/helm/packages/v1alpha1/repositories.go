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
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
)

const (
	HelmRepoType   = "helm"
	OCIRepoType    = "oci"
	RedactedString = "REDACTED"
)

type HelmRepository struct {
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

	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
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
	if client, err := s.getClient(ctx, repo.name.Namespace); err != nil {
		return nil, err
	} else if err = client.Create(ctx, helmRepoCrd); err != nil {
		return nil, statuserror.FromK8sError("create", "AppRepository", repo.name.String(), err)
	} else {
		if !s.pluginConfig.UserManagedSecrets {
			if err = s.setOwnerReferencesForRepoSecret(ctx, secret, helmRepoCrd); err != nil {
				return nil, err
			}
		}
		return &corev1.PackageRepositoryReference{
			Context: &corev1.Context{
				Namespace: helmRepoCrd.Namespace,
				Cluster:   s.kubeappsCluster,
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

// Using owner references on the secret so that it can be
// (1) cleaned up automatically and/or
// (2) enable some control (ie. if I add a secret manually
//   via kubectl before running kubeapps, it won't get deleted just
//   because Kubeapps is deleting it)?
// See https://github.com/vmware-tanzu/kubeapps/pull/4630#discussion_r861446394 for details
func (s *Server) setOwnerReferencesForRepoSecret(
	ctx context.Context,
	secret *k8scorev1.Secret,
	repo *apprepov1alpha1.AppRepository) error {

	if (repo.Spec.Auth.Header != nil || repo.Spec.Auth.CustomCA != nil) && secret != nil {
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return err
		} else {
			secretsInterface := typedClient.CoreV1().Secrets(repo.Namespace)
			secret.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(
					repo,
					schema.GroupVersionKind{
						Group:   apprepov1alpha1.SchemeGroupVersion.Group,
						Version: apprepov1alpha1.SchemeGroupVersion.Version,
						Kind:    "AppRepository",
					}),
			}
			if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return statuserror.FromK8sError("update", "secrets", secret.Name, err)
			}
		}
	}
	return nil
}
