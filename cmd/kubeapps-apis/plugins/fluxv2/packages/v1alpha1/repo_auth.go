// Copyright 2021-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	sourcev1beta2 "github.com/fluxcd/source-controller/api/v1beta2"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
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
	headers http.Header,
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
		return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository cannot mix referenced secrets and user provided secret data"))
	}

	// create/get secret
	if hasCaRef || hasAuthRef {
		// user-managed
		secret, err := s.validateUserManagedRepoSecret(ctx, headers, repoName, repoType, tlsConfig, auth)
		return secret, false, err
	} else if hasCaData || hasAuthData {
		// kubeapps managed
		secret, _, err := newSecretFromTlsConfigAndAuth(repoName, repoType, nil, tlsConfig, auth)
		if err != nil {
			return nil, false, err
		}
		// a bit of catch 22: I need to create a secret first, so that I can create a repo that references it
		// but then I need to set the owner reference on this secret to the repo. In has to be done
		// in that order because to set an owner ref you need object (i.e. repo) UID, which you only get
		// once the object's been created
		// create a secret first, if applicable
		if typedClient, err := s.clientGetter.Typed(headers, s.kubeappsCluster); err != nil {
			return nil, false, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, false, connecterror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
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
	headers http.Header,
	repo *sourcev1beta2.HelmRepository,
	newTlsConfig *corev1.PackageRepositoryTlsConfig,
	newAuth *corev1.PackageRepositoryAuth) (updatedSecret *apiv1.Secret, isKubeappsManagedSecret bool, isSecretUpdated bool, err error) {

	hasCaRef := newTlsConfig != nil && newTlsConfig.GetSecretRef() != nil
	hasCaData := newTlsConfig != nil && newTlsConfig.GetCertAuthority() != ""
	hasAuthRef := newAuth != nil && newAuth.GetSecretRef() != nil
	hasAuthData := newAuth != nil && newAuth.GetSecretRef() == nil && newAuth.GetType() != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED

	// if we have both ref config and data config, it is an invalid mixed configuration
	if (hasCaRef || hasAuthRef) && (hasCaData || hasAuthData) {
		return nil, false, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository cannot mix referenced secrets and user provided secret data"))
	}

	typedClient, err := s.clientGetter.Typed(headers, s.kubeappsCluster)
	if err != nil {
		return nil, false, false, err
	}
	secretInterface := typedClient.CoreV1().Secrets(repo.Namespace)

	var existingSecret *apiv1.Secret
	// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
	if repo.Spec.SecretRef != nil {
		if existingSecret, err = secretInterface.Get(ctx, repo.Spec.SecretRef.Name, metav1.GetOptions{}); err != nil {
			return nil, false, false, connecterror.FromK8sError("get", "secret", repo.Spec.SecretRef.Name, err)
		}
	}

	// check we cannot change mode (per design spec)
	if existingSecret != nil && (hasCaRef || hasCaData || hasAuthRef || hasAuthData) {
		if isSecretKubeappsManaged(existingSecret, repo) != (hasAuthData || hasCaData) {
			return nil, false, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth management mode cannot be changed"))
		}
	}

	repoName := types.NamespacedName{Name: repo.Name, Namespace: repo.Namespace}
	repoType := repo.Spec.Type

	// create/get secret
	if hasCaRef || hasAuthRef {
		// handle user managed secret
		updatedSecret, err := s.validateUserManagedRepoSecret(ctx, headers, repoName, repoType, newTlsConfig, newAuth)
		return updatedSecret, false, true, err

	} else if hasCaData || hasAuthData {
		// handle kubeapps managed secret
		updatedSecret, isSameSecret, err := newSecretFromTlsConfigAndAuth(repoName, repoType, existingSecret, newTlsConfig, newAuth)
		if err != nil {
			return nil, true, false, err
		} else if isSameSecret {
			// Do nothing if repo auth data came fully redacted
			return nil, true, false, nil
		} else {
			// secret has changed, we try to delete any existing secret
			if existingSecret != nil {
				if err = secretInterface.Delete(ctx, existingSecret.Name, metav1.DeleteOptions{}); err != nil {
					log.Errorf("Error deleting existing secret: [%s] due to %v", existingSecret.Name, err)
				}
			}
			// and we recreate the updated one
			if updatedSecret != nil {
				if updatedSecret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Create(ctx, updatedSecret, metav1.CreateOptions{}); err != nil {
					return nil, false, false, connecterror.FromK8sError("create", "secret", updatedSecret.GetGenerateName(), err)
				}
			}
			return updatedSecret, true, true, nil
		}

	} else {
		// no auth, delete existing secret if necessary
		if existingSecret != nil {
			if err = secretInterface.Delete(ctx, existingSecret.Name, metav1.DeleteOptions{}); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", existingSecret.Name, err)
			}
		}
		return nil, false, true, nil
	}
}

func (s *Server) validateUserManagedRepoSecret(
	ctx context.Context,
	headers http.Header,
	repoName types.NamespacedName,
	repoType string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apiv1.Secret, error) {
	var secretRefTls, secretRefAuth string

	if tlsConfig != nil && tlsConfig.GetSecretRef() != nil {
		secretRefTls = tlsConfig.GetSecretRef().GetName()
	}
	if auth != nil && auth.GetSecretRef() != nil {
		secretRefAuth = auth.GetSecretRef().GetName()
	}

	var secretRef string
	if secretRefTls != "" && secretRefAuth != "" && secretRefTls != secretRefAuth {
		// flux repo spec only allows one secret per HelmRepository CRD
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TLS config secret and Auth secret must be the same"))
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	var secret *apiv1.Secret
	if secretRef != "" {
		// check that the specified secret exists
		if typedClient, err := s.clientGetter.Typed(headers, s.kubeappsCluster); err != nil {
			return nil, err
		} else if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, connecterror.FromK8sError("get", "secret", secretRef, err)
		} else {
			// also check that the data in the opaque secret corresponds
			// to specified auth type, e.g. if AuthType is
			// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			// check that the secret has "username" and "password" fields, etc.
			// it appears flux does not care about the k8s secret type (opaque vs tls vs basic-auth, etc.)
			// https://github.com/fluxcd/source-controller/blob/bc5a47e821562b1c4f9731acd929b8d9bd23b3a8/controllers/helmrepository_controller.go#L357
			if secretRefTls != "" && secret.Data["caFile"] == nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing field 'caFile'", secretRef))
			}
			if secretRefAuth != "" {
				switch auth.Type {
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
					if secret.Data["username"] == nil || secret.Data["password"] == nil {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing fields 'username' and/or 'password'", secretRef))
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
					if repoType == sourcev1beta2.HelmRepositoryTypeOCI {
						// ref https://fluxcd.io/flux/components/source/helmrepositories/#tls-authentication
						// Note: TLS authentication is not yet supported by OCI Helm repositories.
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Package repository authentication type %q is not supported for OCI repositories", auth.Type))
					} else {
						if secret.Data["keyFile"] == nil || secret.Data["certFile"] == nil {
							return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing fields 'keyFile' and/or 'certFile'", secretRef))
						}
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
					if repoType == sourcev1beta2.HelmRepositoryTypeOCI {
						if secret.Data[apiv1.DockerConfigJsonKey] == nil {
							return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing field '%s'", secretRef, apiv1.DockerConfigJsonKey))
						}
					} else {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
					}
				default:
					return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
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
	headers http.Header,
	secret *apiv1.Secret,
	repo *sourcev1beta2.HelmRepository) error {

	// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
	if repo.Spec.SecretRef != nil && secret != nil {
		if typedClient, err := s.clientGetter.Typed(headers, s.kubeappsCluster); err != nil {
			return err
		} else {
			secretsInterface := typedClient.CoreV1().Secrets(repo.Namespace)
			secret.OwnerReferences = []metav1.OwnerReference{
				*metav1.NewControllerRef(
					repo,
					schema.GroupVersionKind{
						Group:   sourcev1beta2.GroupVersion.Group,
						Version: sourcev1beta2.GroupVersion.Version,
						Kind:    sourcev1beta2.HelmRepositoryKind,
					}),
			}
			if _, err := secretsInterface.Update(ctx, secret, metav1.UpdateOptions{}); err != nil {
				return connecterror.FromK8sError("update", "secret", secret.Name, err)
			}
		}
	}
	return nil
}

func (s *Server) getRepoTlsConfigAndAuth(ctx context.Context, headers http.Header, repo sourcev1beta2.HelmRepository) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	var tlsConfig *corev1.PackageRepositoryTlsConfig
	var auth *corev1.PackageRepositoryAuth

	// TODO(agamez): flux upgrade - migrate to CertSecretRef, see https://github.com/fluxcd/flux2/releases/tag/v2.1.0
	if repo.Spec.SecretRef != nil {
		secretName := repo.Spec.SecretRef.Name
		if s == nil || s.clientGetter == nil {
			return nil, nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unexpected state in clientGetterHolder instance"))
		}
		typedClient, err := s.clientGetter.Typed(headers, s.kubeappsCluster)
		if err != nil {
			return nil, nil, err
		}
		secret, err := typedClient.CoreV1().Secrets(repo.Namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, connecterror.FromK8sError("get", "secret", secretName, err)
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
	repoType string,
	existingSecret *apiv1.Secret,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (secret *apiv1.Secret, isSameSecret bool, err error) {

	var hadSecretTlsCa, hadSecretTlsCert, hadSecretTlsKey, hadSecretUsername, hadSecretPassword, hadSecretDocker bool
	if existingSecret != nil {
		_, hadSecretTlsCa = existingSecret.Data["caFile"]
		_, hadSecretTlsCert = existingSecret.Data["certFile"]
		_, hadSecretTlsKey = existingSecret.Data["keyFile"]
		_, hadSecretUsername = existingSecret.Data["username"]
		_, hadSecretPassword = existingSecret.Data["password"]
		_, hadSecretDocker = existingSecret.Data[apiv1.DockerConfigJsonKey]
	}

	isSameSecret = true
	secret = newLocalOpaqueSecret(repoName)

	if tlsConfig != nil && tlsConfig.GetCertAuthority() != "" {
		caCert := tlsConfig.GetCertAuthority()
		if caCert == redactedString {
			if hadSecretTlsCa {
				secret.Data["caFile"] = existingSecret.Data["caFile"]
			} else {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}
		} else {
			secret.Data["caFile"] = []byte(caCert)
			isSameSecret = false
		}
	} else {
		if hadSecretTlsCa {
			isSameSecret = false
		}
	}

	if auth != nil && auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			unp := auth.GetUsernamePassword()
			if unp == nil || unp.Username == "" || unp.Password == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Username/Password configuration is missing"))
			}
			if (unp.Username == redactedString && !hadSecretUsername) || (unp.Password == redactedString && !hadSecretPassword) {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}

			if existingSecret != nil {
				secret.Data["username"] = existingSecret.Data["username"]
				secret.Data["password"] = existingSecret.Data["password"]
			}
			if unp.Username != redactedString || unp.Password != redactedString {
				isSameSecret = false
				if unp.Username != redactedString {
					secret.Data["username"] = []byte(unp.Username)
				}
				if unp.Password != redactedString {
					secret.Data["password"] = []byte(unp.Password)
				}
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			if repoType == sourcev1beta2.HelmRepositoryTypeOCI {
				// ref https://fluxcd.io/flux/components/source/helmrepositories/#tls-authentication
				// Note: TLS authentication is not yet supported by OCI Helm repositories.
				return nil, false, connect.NewError(connect.CodeInternal, fmt.Errorf("Package repository authentication type %q is not supported for OCI repositories", auth.Type))
			}

			ck := auth.GetTlsCertKey()
			if ck == nil || ck.Cert == "" || ck.Key == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TLS Cert/Key configuration is missing"))
			}
			if (ck.Cert == redactedString && !hadSecretTlsCert) || (ck.Key == redactedString && !hadSecretTlsKey) {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}

			if existingSecret != nil {
				secret.Data["certFile"] = existingSecret.Data["certFile"]
				secret.Data["keyFile"] = existingSecret.Data["keyFile"]
			}
			if ck.Cert != redactedString || ck.Key != redactedString {
				isSameSecret = false
				if ck.Cert != redactedString {
					secret.Data["certFile"] = []byte(ck.Cert)
				}
				if ck.Key != redactedString {
					secret.Data["keyFile"] = []byte(ck.Key)
				}
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if repoType != sourcev1beta2.HelmRepositoryTypeOCI {
				return nil, false, connect.NewError(connect.CodeInternal, fmt.Errorf("Unsupported package repository authentication type: %q", auth.Type))
			}

			creds := auth.GetDockerCreds()
			if creds == nil || creds.Server == "" || creds.Username == "" || creds.Password == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Docker credentials are missing"))
			}
			if (creds.Server == redactedString || creds.Username == redactedString || creds.Password == redactedString || creds.Email == redactedString) && !hadSecretDocker {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}

			secret.Type = apiv1.SecretTypeDockerConfigJson
			if creds.Server == redactedString && creds.Username == redactedString && creds.Password == redactedString && creds.Email == redactedString {
				secret.Data[apiv1.DockerConfigJsonKey] = existingSecret.Data[apiv1.DockerConfigJsonKey]
			} else if creds.Server == redactedString || creds.Username == redactedString || creds.Password == redactedString || creds.Email == redactedString {
				newcreds, err := decodeDockerAuth(existingSecret.Data[apiv1.DockerConfigJsonKey])
				if err != nil {
					return nil, false, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid configuration, the existing repository does not have valid docker authentication"))
				}

				if creds.Server != redactedString {
					newcreds.Server = creds.Server
				}
				if creds.Username != redactedString {
					newcreds.Username = creds.Username
				}
				if creds.Password != redactedString {
					newcreds.Password = creds.Password
				}
				if creds.Email != redactedString {
					newcreds.Email = creds.Email
				}

				isSameSecret = false
				if configjson, err := encodeDockerAuth(newcreds); err != nil {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
				} else {
					secret.Data[apiv1.DockerConfigJsonKey] = configjson
				}
			} else {
				isSameSecret = false
				if configjson, err := encodeDockerAuth(creds); err != nil {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
				} else {
					secret.Data[apiv1.DockerConfigJsonKey] = configjson
				}
			}
		default:
			return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
		}
	} else {
		// no authentication, check if it was removed
		if hadSecretTlsCert || hadSecretTlsKey || hadSecretUsername || hadSecretPassword || hadSecretDocker {
			isSameSecret = false
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
		log.Warningf("Unrecognized type of secret [%s]", secret.Name)
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
				Email:    redactedString,
			},
		}
	} else {
		log.Warningf("Unrecognized type of secret: [%s]", secret.Name)
	}
	return tlsConfig, auth, nil
}

func isSecretKubeappsManaged(secret *apiv1.Secret, repo *sourcev1beta2.HelmRepository) bool {
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

func encodeDockerAuth(credentials *corev1.DockerCredentials) ([]byte, error) {
	config := &kube.DockerConfigJSON{
		Auths: map[string]kube.DockerConfigEntry{
			credentials.Server: {
				Username: credentials.Username,
				Password: credentials.Password,
				Email:    credentials.Email,
			},
		},
	}
	return json.Marshal(config)
}

func decodeDockerAuth(dockerjson []byte) (*corev1.DockerCredentials, error) {
	config := &kube.DockerConfigJSON{}
	if err := json.Unmarshal(dockerjson, config); err != nil {
		return nil, err
	}
	// note: by design, this method is used when the secret is kubeapps managed, hence we expect only one item in the map
	for server, entry := range config.Auths {
		docker := &corev1.DockerCredentials{
			Server:   server,
			Username: entry.Username,
			Password: entry.Password,
			Email:    entry.Email,
		}
		return docker, nil
	}
	return nil, fmt.Errorf("invalid dockerconfig, no Auths entries were found")
}
