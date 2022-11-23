// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"
)

const (
	redactedString = "REDACTED"
	// to be consistent with carvel and helm plug-in
	annotationManagedByKey   = "kubeapps.dev/managed-by"
	annotationManagedByValue = "plugin:flux"
)

func (s *Server) handleRepoSecretForCreate(
	ctx context.Context,
	repoName types.NamespacedName,
	repoType string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, bool, error) {

	hasCaRef := tlsConfig != nil && tlsConfig.GetSecretRef() != nil
	hasCaData := tlsConfig != nil && tlsConfig.GetCertAuthority() != ""
	hasAuthRef := auth != nil && auth.GetSecretRef() != nil
	hasAuthData := auth != nil && auth.GetSecretRef() == nil && auth.GetType() != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED

	// if we have both ref config and data config, it is an invalid mixed configuration
	if (hasCaRef || hasAuthRef) && (hasCaData || hasAuthData) {
		return nil, false, status.Errorf(codes.InvalidArgument, "Package repository cannot mix referenced secrets and user provided secret data")
	}

	// create/get secret
	if hasCaRef || hasAuthRef {
		// user-managed
		secret, err := s.validateUserManagedRepoSecret(ctx, repoName, repoType, tlsConfig, auth)
		return secret, false, err
	} else if hasCaData || hasAuthData {
		// kubeapps managed
		secret, _, err := newSecretFromTlsConfigAndAuth(repoName, repoType, tlsConfig, auth)
		if err != nil {
			return nil, false, err
		}
		// a bit of catch 22: I need to create a secret first, so that I can create a repo that references it
		// but then I need to set the owner reference on this secret to the repo. In has to be done
		// in that order because to set an owner ref you need object (i.e. repo) UID, which you only get
		// once the object's been created
		// create a secret first, if applicable
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, false, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, false, statuserror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
		} else {
			return secret, true, err
		}
	} else {
		return nil, false, nil
	}
}

// isSecretUpdated is a boolean indicating whether or not the secret ref for a repository
// has been updated as a result of this call.
func (s *Server) handleRepoSecretForUpdate(
	ctx context.Context,
	repo *sourcev1.HelmRepository,
	newTlsConfig *corev1.PackageRepositoryTlsConfig,
	newAuth *corev1.PackageRepositoryAuth) (updatedSecret *apiv1.Secret, isKubeappsManagedSecret bool, isSecretUpdated bool, err error) {

	hasCaRef := newTlsConfig != nil && newTlsConfig.GetSecretRef() != nil
	hasCaData := newTlsConfig != nil && newTlsConfig.GetCertAuthority() != ""
	hasAuthRef := newAuth != nil && newAuth.GetSecretRef() != nil
	hasAuthData := newAuth != nil && newAuth.GetSecretRef() == nil && newAuth.GetType() != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED

	// if we have both ref config and data config, it is an invalid mixed configuration
	if (hasCaRef || hasAuthRef) && (hasCaData || hasAuthData) {
		return nil, false, false, status.Errorf(codes.InvalidArgument, "Package repository cannot mix referenced secrets and user provided secret data")
	}

	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
	if err != nil {
		return nil, false, false, err
	}
	secretInterface := typedClient.CoreV1().Secrets(repo.Namespace)

	var existingSecret *apiv1.Secret
	if repo.Spec.SecretRef != nil {
		if existingSecret, err = secretInterface.Get(ctx, repo.Spec.SecretRef.Name, metav1.GetOptions{}); err != nil {
			return nil, false, false, statuserror.FromK8sError("get", "secret", repo.Spec.SecretRef.Name, err)
		}
	}

	// check we cannot change mode (per design spec)
	if existingSecret != nil && (hasCaRef || hasCaData || hasAuthRef || hasAuthData) {
		if isSecretKubeappsManaged(existingSecret, repo) != (hasAuthData || hasCaData) {
			return nil, false, false, status.Errorf(codes.InvalidArgument, "Auth management mode cannot be changed")
		}
	}

	repoName := types.NamespacedName{Name: repo.Name, Namespace: repo.Namespace}
	repoType := repo.Spec.Type

	// handle user managed secret
	if hasCaRef || hasAuthRef {
		updatedSecret, err := s.validateUserManagedRepoSecret(ctx, repoName, repoType, newTlsConfig, newAuth)
		return updatedSecret, false, true, err
	}

	// handle kubeapps managed secret
	var isSameSecret bool
	updatedSecret, isSameSecret, err = newSecretFromTlsConfigAndAuth(repoName, repoType, newTlsConfig, newAuth)
	if err != nil {
		return nil, true, false, err
	} else if isSameSecret {
		// Do nothing if repo auth data came redacted
		return nil, true, false, nil
	} else {
		// either we have no secret, or it has changed. in both cases, we try to delete any existing secret
		if existingSecret != nil {
			if err = secretInterface.Delete(ctx, existingSecret.Name, metav1.DeleteOptions{}); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", err)
			}
		}
		if updatedSecret != nil {
			if updatedSecret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, updatedSecret, metav1.CreateOptions{}); err != nil {
				return nil, false, false, statuserror.FromK8sError("create", "secret", updatedSecret.GetGenerateName(), err)
			}
		}
		return updatedSecret, true, true, nil
	}
}

func (s *Server) validateUserManagedRepoSecret(
	ctx context.Context,
	repoName types.NamespacedName,
	repoType string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, error) {
	var secretRefTls, secretRefAuth string
	if tlsConfig != nil {
		if tlsConfig.GetCertAuthority() != "" {
			return nil, status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if tlsConfig.GetSecretRef().GetName() != "" {
			secretRefTls = tlsConfig.GetSecretRef().GetName()
		}
	}

	if auth != nil {
		if auth.GetDockerCreds() != nil ||
			auth.GetHeader() != "" ||
			auth.GetTlsCertKey() != nil ||
			auth.GetUsernamePassword() != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Secret Ref must be used with user managed secrets")
		} else if auth.GetSecretRef().GetName() != "" {
			secretRefAuth = auth.GetSecretRef().GetName()
		}
	}

	var secretRef string
	if secretRefTls != "" && secretRefAuth != "" && secretRefTls != secretRefAuth {
		// flux repo spec only allows one secret per HelmRepository CRD
		return nil, status.Errorf(
			codes.InvalidArgument, "TLS config secret and Auth secret must be the same")
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	var secret *apiv1.Secret
	if secretRef != "" {
		// check that the specified secret exists
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return nil, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, statuserror.FromK8sError("get", "secret", secretRef, err)
		} else {
			// also check that the data in the opaque secret corresponds
			// to specified auth type, e.g. if AuthType is
			// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			// check that the secret has "username" and "password" fields, etc.
			// it appears flux does not care about the k8s secret type (opaque vs tls vs basic-auth, etc.)
			// https://github.com/fluxcd/source-controller/blob/bc5a47e821562b1c4f9731acd929b8d9bd23b3a8/controllers/helmrepository_controller.go#L357
			if secretRefTls != "" && secret.Data["caFile"] == nil {
				return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing field 'caFile'", secretRef)
			}
			if secretRefAuth != "" {
				switch auth.Type {
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
					if secret.Data["username"] == nil || secret.Data["password"] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing fields 'username' and/or 'password'", secretRef)
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
					if secret.Data["keyFile"] == nil || secret.Data["certFile"] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing fields 'keyFile' and/or 'certFile'", secretRef)
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
					if repoType == sourcev1.HelmRepositoryTypeOCI {
						if secret.Data[apiv1.DockerConfigJsonKey] == nil {
							return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing field '%s'", secretRef, apiv1.DockerConfigJsonKey)
						}
					} else {
						return nil, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
					}
				default:
					return nil, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
				}
			}
		}

		// ref https://github.com/vmware-tanzu/kubeapps/pull/4353#discussion_r816332595
		// check whether flux supports typed secrets in addition to opaque secrets
		// https://kubernetes.io/docs/concepts/configuration/secret/#secret-types
		// update: flux currently does not care about secret type, just what is in the data map.
	}
	return secret, nil
}

// using owner references on the secret so that it can be
// (1) cleaned up automatically and/or
// (2) enable some control (ie. if I add a secret manually
//
//	via kubectl before running kubeapps, it won't get deleted just
//
// because Kubeapps is deleting it)?
// see https://github.com/vmware-tanzu/kubeapps/pull/4630#discussion_r861446394 for details
func (s *Server) setOwnerReferencesForRepoSecret(
	ctx context.Context,
	secret *apiv1.Secret,
	repo *sourcev1.HelmRepository) error {

	if repo.Spec.SecretRef != nil && secret != nil {
		if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
			return err
		} else {
			secretsInterface := typedClient.CoreV1().Secrets(repo.Namespace)
			secret.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(
					repo,
					schema.GroupVersionKind{
						Group:   sourcev1.GroupVersion.Group,
						Version: sourcev1.GroupVersion.Version,
						Kind:    sourcev1.HelmRepositoryKind,
					}),
			}
			if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return statuserror.FromK8sError("update", "secret", secret.Name, err)
			}
		}
	}
	return nil
}

func (s *Server) getRepoTlsConfigAndAuth(ctx context.Context, repo sourcev1.HelmRepository) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	var auth *corev1.PackageRepositoryAuth

	if repo.Spec.SecretRef != nil {
		secretName := repo.Spec.SecretRef.Name
		if s == nil || s.clientGetter == nil {
			return nil, nil, status.Errorf(codes.Internal, "unexpected state in clientGetterHolder instance")
		}
		typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
		if err != nil {
			return nil, nil, err
		}
		secret, err := typedClient.CoreV1().Secrets(repo.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, statuserror.FromK8sError("get", "secret", secretName, err)
		}

		if isSecretKubeappsManaged(secret, &repo) {
			if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(secret); err != nil {
				return nil, nil, err
			}
		} else {
			if tlsConfig, auth, err = getRepoTlsConfigAndAuthWithUserManagedSecrets(secret); err != nil {
				return nil, nil, err
			}
		}
	} else {
		auth = &corev1.PackageRepositoryAuth{
			Type:            corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
			PassCredentials: repo.Spec.PassCredentials,
		}
	}

	if repo.Spec.PassCredentials {
		if auth == nil {
			auth = &corev1.PackageRepositoryAuth{
				PassCredentials: repo.Spec.PassCredentials,
			}
		}
	}

	return tlsConfig, auth, nil
}

// this func is only used with kubeapps-managed secrets
func newSecretFromTlsConfigAndAuth(repoName types.NamespacedName,
	typ string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (secret *apiv1.Secret, isSameSecret bool, err error) {
	if tlsConfig != nil {
		if tlsConfig.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		caCert := tlsConfig.GetCertAuthority()
		if caCert == redactedString {
			isSameSecret = true
		} else if caCert != "" {
			secret = newLocalOpaqueSecret(repoName)
			secret.Data["caFile"] = []byte(caCert)
		}
	}
	if auth != nil {
		if auth.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		if secret == nil {
			if auth.Type == corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON {
				secret = newLocalDockerConfigJsonSecret(repoName)
			} else {
				secret = newLocalOpaqueSecret(repoName)
			}
		}
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if unp := auth.GetUsernamePassword(); unp != nil {
				if unp.Username == redactedString && unp.Password == redactedString {
					isSameSecret = true
				} else {
					secret.Data["username"] = []byte(unp.Username)
					secret.Data["password"] = []byte(unp.Password)
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "Username/Password configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			if ck := auth.GetTlsCertKey(); ck != nil {
				if ck.Cert == redactedString && ck.Key == redactedString {
					isSameSecret = true
				} else {
					secret.Data["certFile"] = []byte(ck.Cert)
					secret.Data["keyFile"] = []byte(ck.Key)
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "TLS Cert/Key configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if typ == sourcev1.HelmRepositoryTypeOCI {
				if dc := auth.GetDockerCreds(); dc != nil {
					if dc.Username == redactedString && dc.Password == redactedString && dc.Server == redactedString {
						isSameSecret = true
					} else {
						secret.Data = map[string][]byte{
							apiv1.DockerConfigJsonKey: []byte(`{"auths":{"` +
								dc.Server + `":{"` +
								`auth":"` + base64.StdEncoding.EncodeToString([]byte(dc.Username+":"+dc.Password)) + `"}}}`),
						}
					}
				} else {
					return nil, false, status.Errorf(codes.Internal, "Docker credentials configuration is missing")
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "Unsupported package repository authentication type: %q", auth.Type)
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
			corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
			return nil, false, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, true, nil
		default:
			return nil, false, status.Errorf(codes.Internal, "Unsupported package repository authentication type: %q", auth.Type)
		}
	}
	return secret, isSameSecret, nil
}

func getRepoTlsConfigAndAuthWithUserManagedSecrets(secret *apiv1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	auth := &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
	}

	if _, ok := secret.Data["caFile"]; ok {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			// flux plug in doesn't support this option
			InsecureSkipVerify: false,
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: secret.Name,
					Key:  "caFile",
				},
			},
		}
	}
	if _, ok := secret.Data["certFile"]; ok {
		if _, ok = secret.Data["keyFile"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{Name: secret.Name},
			}
		}
	} else if _, ok := secret.Data["username"]; ok {
		if _, ok = secret.Data["password"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{Name: secret.Name},
			}
		}
	} else {
		log.Warning("Unrecognized type of secret [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}

// TODO (gfichtenolt) Per slack discussion
// In fact, keeping the existing API might mean we could return exactly what it already does today
// (i.e. all secrets) if called with an extra explicit option (includeSecrets=true in the request
// message, not sure, similar to kubectl  config view --raw) and by default the secrets are REDACTED
// as you mention? This would mean clients will by default see only REDACTED secrets,
// but can request the full sensitive data when necessary?
func getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(secret *apiv1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	auth := &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED,
	}

	if _, ok := secret.Data["caFile"]; ok {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			// flux plug in doesn't support InsecureSkipVerify option
			InsecureSkipVerify: false,
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
				CertAuthority: redactedString,
			},
		}
	}

	if _, ok := secret.Data["certFile"]; ok {
		if _, ok := secret.Data["keyFile"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_TlsCertKey{
				TlsCertKey: &corev1.TlsCertKey{
					Cert: redactedString,
					Key:  redactedString,
				},
			}
		}
	} else if _, ok := secret.Data["username"]; ok {
		if _, ok := secret.Data["password"]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: redactedString,
					Password: redactedString,
				},
			}
		}
	} else if _, ok := secret.Data[apiv1.DockerConfigJsonKey]; ok {
		auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
		auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
			DockerCreds: &corev1.DockerCredentials{
				Username: redactedString,
				Password: redactedString,
				Server:   redactedString,
			},
		}
	} else {
		log.Warning("Unrecognized type of secret: [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}

func isSecretKubeappsManaged(secret *apiv1.Secret, repo *sourcev1.HelmRepository) bool {
	if !metav1.IsControlledBy(secret, repo) {
		return false
	}
	if managedby := secret.GetAnnotations()[annotationManagedByKey]; managedby != annotationManagedByValue {
		return false
	}
	return true
}

// "Local" in the sense of no namespace is specified
func newLocalOpaqueSecret(ownerRepo types.NamespacedName) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: ownerRepo.Name + "-",
			Annotations: map[string]string{
				annotationManagedByKey: annotationManagedByValue,
			},
		},
		Type: apiv1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}

// "Local" in the sense of no namespace is specified
func newLocalDockerConfigJsonSecret(ownerRepo types.NamespacedName) *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: ownerRepo.Name + "-",
			Annotations: map[string]string{
				annotationManagedByKey: annotationManagedByValue,
			},
		},
		Type: apiv1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{},
	}
}
