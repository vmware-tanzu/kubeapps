// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
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

func (s *Server) newRepo(ctx context.Context, targetName types.NamespacedName, url string, repoType string, description string, interval uint32,
	tlsConfig *corev1.PackageRepositoryTlsConfig, auth *corev1.PackageRepositoryAuth) (*corev1.PackageRepositoryReference, error) {
	if url == "" {
		return nil, status.Errorf(codes.InvalidArgument, "repository url may not be empty")
	}
	if repoType == "" || !slices.Contains(getValidRepoTypes(), repoType) {
		return nil, status.Errorf(codes.InvalidArgument, "repository type [%s] not supported", repoType)
	}

	var secret *k8scorev1.Secret
	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
	if err != nil {
		return nil, err
	}
	if s.pluginConfig.UserManagedSecrets {
		if secret, err = validateUserManagedRepoSecret(ctx, typedClient, targetName, tlsConfig, auth); err != nil {
			return nil, err
		}
	} else {
		// a bit of catch 22: I need to create a secret first, so that I can create a repo that references it
		// but then I need to set the owner reference on this secret to the repo. In has to be done
		// in that order because to set an owner ref you need object (i.e. repo) UID, which you only get
		// once the object's been created
		if secret, err = createKubeappsManagedRepoSecret(ctx, typedClient, targetName, tlsConfig, auth); err != nil {
			return nil, err
		}
	}

	passCredentials := auth != nil && auth.PassCredentials
	insecureSkipVerify := tlsConfig != nil && tlsConfig.InsecureSkipVerify

	if helmRepoCrd, err := newHelmRepoCrd(targetName, url, repoType, description, interval, secret, tlsConfig, auth, passCredentials, insecureSkipVerify); err != nil {
		return nil, err
	} else if client, err := s.getClient(ctx, targetName.Namespace); err != nil {
		return nil, err
	} else if err = client.Create(ctx, helmRepoCrd); err != nil {
		return nil, statuserror.FromK8sError("create", "AppRepository", targetName.String(), err)
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

func getValidRepoTypes() []string {
	return []string{HelmRepoType, OCIRepoType}
}

func newHelmRepoCrd(targetName types.NamespacedName,
	url string, repoType string, description string,
	interval uint32,
	secret *k8scorev1.Secret,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth,
	passCredentials bool,
	tlsInsecureSkipVerify bool) (*apprepov1alpha1.AppRepository, error) {
	appRepoCrd := &apprepov1alpha1.AppRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      targetName.Name,
			Namespace: targetName.Namespace,
		},
		Spec: apprepov1alpha1.AppRepositorySpec{
			URL:                   url,
			Type:                  repoType,
			TLSInsecureSkipVerify: tlsInsecureSkipVerify,
			Description:           description,
			PassCredentials:       passCredentials,
		},
	}
	if auth != nil || tlsConfig != nil {
		if repoAuth, err := newAppRepositoryAuth(secret, tlsConfig, auth); err != nil {
			return nil, err
		} else if repoAuth != nil {
			appRepoCrd.Spec.Auth = *repoAuth
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
