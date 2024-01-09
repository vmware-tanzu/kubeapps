// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	k8scorev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	log "k8s.io/klog/v2"
)

const (
	SecretCaKey         = "ca.crt"
	SecretAuthHeaderKey = "authorizationHeader"
	DockerConfigJsonKey = ".dockerconfigjson"

	Annotation_ManagedBy_Key   = "kubeapps.dev/managed-by"
	Annotation_ManagedBy_Value = "plugin:helm"
)

func newLocalOpaqueSecret(repoName string) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        helm.SecretNameForRepo(repoName),
			Annotations: map[string]string{Annotation_ManagedBy_Key: Annotation_ManagedBy_Value},
		},
		Type: k8scorev1.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
}

func handleAuthSecretForCreate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, bool, error) {

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
		secret, err := validateUserManagedRepoSecret(ctx, typedClient, repoName.Namespace, tlsConfig, auth)
		return secret, false, err
	} else if hasCaData || hasAuthData {
		secret, _, err := newSecretFromTlsConfigAndAuth(repoName.Name, nil, tlsConfig, auth)
		return secret, true, err
	} else {
		return nil, false, nil
	}
}

func handleImagesPullSecretForCreate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repoName types.NamespacedName,
	customDetail *v1alpha1.HelmPackageRepositoryCustomDetail) (*k8scorev1.Secret, bool, error) {

	hasRef := customDetail != nil && customDetail.ImagesPullSecret != nil && customDetail.ImagesPullSecret.GetSecretRef() != ""
	hasData := customDetail != nil && customDetail.ImagesPullSecret != nil && customDetail.ImagesPullSecret.GetCredentials() != nil

	// create/get secret
	if hasRef {
		secret, err := validateDockerImagePullSecret(ctx, typedClient, repoName.Namespace, customDetail.ImagesPullSecret.GetSecretRef())
		return secret, false, err
	} else if hasData {
		secret, _, err := newDockerImagePullSecret(repoName.Name, nil, customDetail.ImagesPullSecret.GetCredentials())
		return secret, true, err
	} else {
		return nil, false, nil
	}
}

func handleAuthSecretForUpdate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repo *apprepov1alpha1.AppRepository,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth,
	secret *k8scorev1.Secret) (updatedSecret *k8scorev1.Secret, secretIsKubeappsManaged bool, secretIsUpdated bool, err error) {

	hasCaRef := tlsConfig != nil && tlsConfig.GetSecretRef() != nil
	hasCaData := tlsConfig != nil && tlsConfig.GetCertAuthority() != ""
	hasAuthRef := auth != nil && auth.GetSecretRef() != nil
	hasAuthData := auth != nil && auth.GetSecretRef() == nil && auth.GetType() != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED

	// if we have both ref config and data config, it is an invalid mixed configuration
	if (hasCaRef || hasAuthRef) && (hasCaData || hasAuthData) {
		return nil, false, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository cannot mix referenced secrets and user provided secret data"))
	}

	// check we cannot change mode (per design spec)
	if secret != nil && (hasCaRef || hasCaData || hasAuthRef || hasAuthData) {
		if isAuthSecretKubeappsManaged(repo, secret) != (hasAuthData || hasCaData) {
			return nil, false, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth management mode cannot be changed"))
		}
	}

	// create/get secret
	if hasCaRef || hasAuthRef {
		// handle user managed secret
		updatedSecret, err := validateUserManagedRepoSecret(ctx, typedClient, repo.GetNamespace(), tlsConfig, auth)
		return updatedSecret, false, true, err

	} else if hasCaData || hasAuthData {
		// handle kubeapps managed secret
		updatedSecret, isSameSecret, err := newSecretFromTlsConfigAndAuth(repo.GetName(), secret, tlsConfig, auth)
		if err != nil {
			return nil, true, false, err
		} else if isSameSecret {
			// Do nothing if repo auth data came fully redacted
			return nil, true, false, nil
		} else {
			// secret has changed, we try to delete any existing secret
			if secret != nil {
				secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
				if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
					log.Errorf("Error deleting existing secret: [%s] due to %v", secret.GetName(), err)
				}
			}
			return updatedSecret, true, true, nil
		}

	} else {
		// no auth, delete existing secret if necessary
		if secret != nil {
			secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
			if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", secret.GetName(), err)
			}
		}
		return nil, false, true, nil
	}
}

func handleImagesPullSecretForUpdate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repo *apprepov1alpha1.AppRepository,
	customDetail *v1alpha1.HelmPackageRepositoryCustomDetail,
	secret *k8scorev1.Secret) (updatedSecret *k8scorev1.Secret, secretIsKubeappsManaged bool, secretIsUpdated bool, err error) {

	var imagesPullSecrets *v1alpha1.ImagesPullSecret
	if customDetail != nil && customDetail.ImagesPullSecret != nil {
		imagesPullSecrets = customDetail.ImagesPullSecret
	} else {
		imagesPullSecrets = &v1alpha1.ImagesPullSecret{}
	}

	hasRef := imagesPullSecrets != nil && imagesPullSecrets.GetSecretRef() != ""
	hasData := imagesPullSecrets != nil && imagesPullSecrets.GetCredentials() != nil

	// check we are not changing mode
	if secret != nil && (hasRef || hasData) {
		if isImagesPullSecretKubeappsManaged(repo, secret) != hasData {
			return nil, false, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Auth management mode cannot be changed"))
		}
	}

	// create/get secret
	if hasRef {
		// handle user managed secret
		updatedSecret, err := validateDockerImagePullSecret(ctx, typedClient, repo.GetNamespace(), imagesPullSecrets.GetSecretRef())
		return updatedSecret, false, true, err

	} else if hasData {
		// handle kubeapps managed secret
		updatedSecret, isSameSecret, err := newDockerImagePullSecret(repo.GetName(), secret, imagesPullSecrets.GetCredentials())
		if err != nil {
			return nil, true, false, err
		} else if isSameSecret {
			// Do nothing if repo auth data came fully redacted
			return nil, true, false, nil
		} else {
			// secret has changed, we try to delete any existing secret
			if secret != nil {
				secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
				if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
					log.Errorf("Error deleting existing secret: [%s] due to %v", secret.GetName(), err)
				}
			}
			return updatedSecret, true, true, nil
		}

	} else {
		// no image pull secret, delete existing secret if necessary
		if secret != nil {
			secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
			if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", secret.GetName(), err)
			}
		}
		return nil, false, true, nil
	}
}

// this func is only used with kubeapps-managed secrets
func newSecretFromTlsConfigAndAuth(repoName string,
	existingSecret *k8scorev1.Secret,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (secret *k8scorev1.Secret, isSameSecret bool, err error) {

	var hadSecretCa, hadSecretHeader, hadSecretDocker bool
	if existingSecret != nil {
		_, hadSecretCa = existingSecret.Data[SecretCaKey]
		_, hadSecretHeader = existingSecret.Data[SecretAuthHeaderKey]
		_, hadSecretDocker = existingSecret.Data[DockerConfigJsonKey]
	}

	isSameSecret = true
	secret = newLocalOpaqueSecret(repoName)

	if tlsConfig != nil && tlsConfig.GetCertAuthority() != "" {
		caCert := tlsConfig.GetCertAuthority()
		if caCert == RedactedString {
			if hadSecretCa {
				secret.Data[SecretCaKey] = existingSecret.Data[SecretCaKey]
			} else {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}
		} else {
			secret.Data[SecretCaKey] = []byte(caCert)
			isSameSecret = false
		}
	} else {
		if hadSecretCa {
			isSameSecret = false
		}
	}

	if auth != nil && auth.Type != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED {
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			unp := auth.GetUsernamePassword()
			if unp == nil || unp.GetUsername() == "" || unp.GetPassword() == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Username/Password configuration is missing"))
			}
			if (unp.GetUsername() == RedactedString || unp.GetPassword() == RedactedString) && !hadSecretHeader {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}

			if unp.GetUsername() == RedactedString && unp.GetPassword() == RedactedString {
				secret.Data[SecretAuthHeaderKey] = existingSecret.Data[SecretAuthHeaderKey]
			} else if unp.GetUsername() == RedactedString || unp.GetPassword() == RedactedString {
				username, password, ok := decodeBasicAuth(string(existingSecret.Data[SecretAuthHeaderKey]))
				if !ok {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, the existing repository does not have username/password authentication"))
				}
				if unp.GetUsername() != RedactedString {
					username = unp.GetUsername()
				}
				if unp.GetPassword() != RedactedString {
					password = unp.GetPassword()
				}
				isSameSecret = false
				secret.Data[SecretAuthHeaderKey] = []byte(encodeBasicAuth(username, password))
			} else {
				isSameSecret = false
				secret.Data[SecretAuthHeaderKey] = []byte(encodeBasicAuth(unp.GetUsername(), unp.GetPassword()))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
			token := auth.GetHeader()
			if token == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Bearer token is missing"))
			}

			if token == RedactedString {
				if hadSecretHeader {
					secret.Data[SecretAuthHeaderKey] = existingSecret.Data[SecretAuthHeaderKey]
				} else {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
				}
			} else {
				isSameSecret = false
				secret.Data[SecretAuthHeaderKey] = []byte(encodeBearerAuth(token))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
			header := auth.GetHeader()
			if header == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Authentication header value is missing"))
			}

			if header == RedactedString {
				if hadSecretHeader {
					secret.Data[SecretAuthHeaderKey] = existingSecret.Data[SecretAuthHeaderKey]
				} else {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
				}
			} else {
				isSameSecret = false
				secret.Data[SecretAuthHeaderKey] = []byte(header)
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			creds := auth.GetDockerCreds()
			if creds == nil || creds.Server == "" || creds.Username == "" || creds.Password == "" {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Docker credentials are missing"))
			}
			if (creds.Server == RedactedString || creds.Username == RedactedString || creds.Password == RedactedString || creds.Email == RedactedString) && !hadSecretDocker {
				return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
			}

			secret.Type = k8scorev1.SecretTypeDockerConfigJson
			if creds.Server == RedactedString && creds.Username == RedactedString && creds.Password == RedactedString && creds.Email == RedactedString {
				secret.Data[DockerConfigJsonKey] = existingSecret.Data[DockerConfigJsonKey]
			} else if creds.Server == RedactedString || creds.Username == RedactedString || creds.Password == RedactedString || creds.Email == RedactedString {
				newcreds, err := decodeDockerAuth(existingSecret.Data[DockerConfigJsonKey])
				if err != nil {
					return nil, false, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid configuration, the existing repository does not have valid docker authentication"))
				}

				if creds.Server != RedactedString {
					newcreds.Server = creds.Server
				}
				if creds.Username != RedactedString {
					newcreds.Username = creds.Username
				}
				if creds.Password != RedactedString {
					newcreds.Password = creds.Password
				}
				if creds.Email != RedactedString {
					newcreds.Email = creds.Email
				}

				isSameSecret = false
				if configjson, err := encodeDockerAuth(newcreds); err != nil {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
				} else {
					secret.Data[DockerConfigJsonKey] = configjson
				}
			} else {
				isSameSecret = false
				if configjson, err := encodeDockerAuth(creds); err != nil {
					return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
				} else {
					secret.Data[DockerConfigJsonKey] = configjson
				}
			}
		default:
			return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
		}
	} else {
		// no authentication, check if it was removed
		if hadSecretHeader || hadSecretDocker {
			isSameSecret = false
		}
	}
	return secret, isSameSecret, nil
}

func newAppRepositoryAuth(secret *k8scorev1.Secret,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*apprepov1alpha1.AppRepositoryAuth, error) {

	var appRepoAuth = &apprepov1alpha1.AppRepositoryAuth{}

	if tlsConfig != nil {
		if secret == nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Secret for AppRepository auth is missing"))
		}
		appRepoAuth.CustomCA = &apprepov1alpha1.AppRepositoryCustomCA{
			SecretKeyRef: k8scorev1.SecretKeySelector{
				Key: SecretCaKey,
				LocalObjectReference: k8scorev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		}
	}

	if auth != nil {
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
			corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
			if secret == nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Secret for AppRepository auth is missing"))
			}
			if _, ok := secret.Data[SecretAuthHeaderKey]; ok {
				appRepoAuth.Header = &apprepov1alpha1.AppRepositoryAuthHeader{
					SecretKeyRef: k8scorev1.SecretKeySelector{
						Key: SecretAuthHeaderKey,
						LocalObjectReference: k8scorev1.LocalObjectReference{
							Name: secret.Name,
						},
					},
				}
			} else {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Authentication header is missing"))
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if secret == nil {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Secret for AppRepository auth is missing"))
			}
			if _, ok := secret.Data[DockerConfigJsonKey]; ok {
				appRepoAuth.Header = &apprepov1alpha1.AppRepositoryAuthHeader{
					SecretKeyRef: k8scorev1.SecretKeySelector{
						Key: DockerConfigJsonKey,
						LocalObjectReference: k8scorev1.LocalObjectReference{
							Name: secret.Name,
						},
					},
				}
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, nil
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unexpected package repository authentication type: %q", auth.Type))
		}
	}

	return appRepoAuth, nil
}

func createKubeappsManagedRepoSecret(
	ctx context.Context,
	typedClient kubernetes.Interface,
	namespace string,
	secret *k8scorev1.Secret) (newSecret *k8scorev1.Secret, err error) {

	if secret != nil {
		// create a secret first, if applicable
		secretName := secret.GetName()
		newSecret, err = typedClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, connecterror.FromK8sError("create", "secret", secretName, err)
		}
	}
	return newSecret, nil
}

func validateDockerImagePullSecret(ctx context.Context,
	typedClient kubernetes.Interface,
	namespace string,
	secretName string) (*k8scorev1.Secret, error) {

	if secret, err := typedClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{}); err != nil {
		return nil, connecterror.FromK8sError("get", "secret", secretName, err)
	} else if secret.Type != k8scorev1.SecretTypeDockerConfigJson {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Images Docker pull secret %s does not have valid type", secretName))
	} else if _, ok := secret.Data[k8scorev1.DockerConfigJsonKey]; !ok {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Images Docker pull secret %s does not have valid data", secretName))
	} else {
		return secret, nil
	}
}

func imagesPullSecretName(repoName string) string {
	return fmt.Sprintf("pullsecret-%s", repoName)
}

func newDockerImagePullSecret(repoName string, existingSecret *k8scorev1.Secret, creds *corev1.DockerCredentials) (secret *k8scorev1.Secret, isSameSecret bool, err error) {
	if creds.Server == "" || creds.Username == "" || creds.Password == "" {
		return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Docker credentials are missing"))
	}
	if (creds.Server == RedactedString || creds.Username == RedactedString || creds.Password == RedactedString || creds.Email == RedactedString) && existingSecret == nil {
		return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid configuration, unexpected REDACTED content"))
	}

	secret = &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        imagesPullSecretName(repoName),
			Annotations: map[string]string{Annotation_ManagedBy_Key: Annotation_ManagedBy_Value},
		},
		Type: k8scorev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{},
	}

	if creds.Server == RedactedString && creds.Username == RedactedString && creds.Password == RedactedString && creds.Email == RedactedString {
		// same
		return nil, true, nil

	} else if creds.Server == RedactedString || creds.Username == RedactedString || creds.Password == RedactedString || creds.Email == RedactedString {
		// merge
		newcreds, err := decodeDockerAuth(existingSecret.Data[DockerConfigJsonKey])
		if err != nil {
			return nil, false, connect.NewError(connect.CodeInternal, fmt.Errorf("Invalid configuration, the existing repository does not have valid docker authentication"))
		}

		if creds.Server != RedactedString {
			newcreds.Server = creds.Server
		}
		if creds.Username != RedactedString {
			newcreds.Username = creds.Username
		}
		if creds.Password != RedactedString {
			newcreds.Password = creds.Password
		}
		if creds.Email != RedactedString {
			newcreds.Email = creds.Email
		}

		if configjson, err := encodeDockerAuth(newcreds); err != nil {
			return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
		} else {
			secret.Data[DockerConfigJsonKey] = configjson
			return secret, false, nil
		}

	} else {
		// new
		if configjson, err := encodeDockerAuth(creds); err != nil {
			return nil, false, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Invalid Docker credentials"))
		} else {
			secret.Data[DockerConfigJsonKey] = configjson
			return secret, false, nil
		}
	}
}

func deleteSecret(ctx context.Context, secretsInterface v1.SecretInterface, secretName string) error {
	// Ignore action if secret didn't exist
	if _, err := secretsInterface.Get(ctx, secretName, metav1.GetOptions{}); err == nil {
		if err := secretsInterface.Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
			return connecterror.FromK8sError("delete", "secret", secretName, err)
		}
	}
	return nil
}

func (s *Server) copyRepositorySecretToNamespace(typedClient kubernetes.Interface, targetNamespace string, secret *k8scorev1.Secret, repoName types.NamespacedName) (copiedSecret *k8scorev1.Secret, err error) {
	newSecret := &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helm.SecretNameForNamespacedRepo(repoName.Name, repoName.Namespace),
			Namespace: targetNamespace,
		},
		Type: secret.Type,
		Data: secret.Data,
	}
	copiedSecret, err = typedClient.CoreV1().Secrets(targetNamespace).Create(context.TODO(), newSecret, metav1.CreateOptions{})
	if err != nil && k8sErrors.IsAlreadyExists(err) {
		copiedSecret, err = typedClient.CoreV1().Secrets(targetNamespace).Update(context.TODO(), newSecret, metav1.UpdateOptions{})
	}
	if err != nil {
		err = connect.NewError(connect.CodeInternal, err)
	}
	return copiedSecret, err
}

func (s *Server) deleteRepositorySecretFromNamespace(typedClient kubernetes.Interface, targetNamespace, secretName string) error {
	secretsInterface := typedClient.CoreV1().Secrets(targetNamespace)
	if deleteErr := deleteSecret(context.TODO(), secretsInterface, secretName); deleteErr != nil {
		return deleteErr
	}
	return nil
}

func validateUserManagedRepoSecret(
	ctx context.Context,
	typedClient kubernetes.Interface,
	namespace string,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, error) {
	var secretRefTls, secretRefAuth string

	if tlsConfig != nil && tlsConfig.GetSecretRef() != nil {
		secretRefTls = tlsConfig.GetSecretRef().GetName()
	}
	if auth != nil && auth.GetSecretRef() != nil {
		secretRefAuth = auth.GetSecretRef().GetName()
	}

	var secretRef string
	if secretRefTls != "" && secretRefAuth != "" && secretRefTls != secretRefAuth {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("TLS config secret and Auth secret must be the same"))
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	var secret *k8scorev1.Secret
	if secretRef != "" {
		var err error
		// check that the specified secret exists
		if secret, err = typedClient.CoreV1().Secrets(namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, connecterror.FromK8sError("get", "secret", secretRef, err)
		} else {
			// also check that the data in the opaque secret corresponds
			// to specified auth type, e.g. if AuthType is
			// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			// check that the secret data has valid fields
			if secretRefTls != "" && secret.Data[SecretCaKey] == nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing key '%s'", secretRef, SecretCaKey))
			}
			if secretRefAuth != "" {
				switch auth.Type {
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
					if data := secret.Data[SecretAuthHeaderKey]; data == nil {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing key '%s'", secretRef, SecretAuthHeaderKey))
					} else if _, _, ok := decodeBasicAuth(string(data)); !ok {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] does not represent a valid Basic Auth secret'", secretRef))
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
					if data := secret.Data[SecretAuthHeaderKey]; data == nil {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing key '%s'", secretRef, SecretAuthHeaderKey))
					} else if _, ok := decodeBearerAuth(string(data)); !ok {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] does not represent a valid Bearer Auth secret'", secretRef))
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
					if data := secret.Data[SecretAuthHeaderKey]; data == nil {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing key '%s'", secretRef, SecretAuthHeaderKey))
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
					if secret.Type != k8scorev1.SecretTypeDockerConfigJson {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] does not have expected dockerconfig type", secretRef))
					} else if _, ok := secret.Data[k8scorev1.DockerConfigJsonKey]; !ok {
						return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Specified secret [%s] missing key '%s'", secretRef, DockerConfigJsonKey))
					}
				default:
					return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Package repository authentication type %q is not supported", auth.Type))
				}
			}
		}
	}
	return secret, nil
}

func getRepoImagesPullSecret(source *apprepov1alpha1.AppRepository, imagesPullSecret *k8scorev1.Secret) *v1alpha1.ImagesPullSecret {
	if imagesPullSecret == nil {
		return nil
	} else if isImagesPullSecretKubeappsManaged(source, imagesPullSecret) {
		return &v1alpha1.ImagesPullSecret{
			DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
				Credentials: &corev1.DockerCredentials{
					Server:   RedactedString,
					Username: RedactedString,
					Password: RedactedString,
					Email:    RedactedString,
				},
			},
		}
	} else {
		return &v1alpha1.ImagesPullSecret{
			DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
				SecretRef: imagesPullSecret.Name,
			},
		}
	}
}

func getRepoTlsConfigAndAuth(
	source *apprepov1alpha1.AppRepository,
	caSecret *k8scorev1.Secret,
	authSecret *k8scorev1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {

	var tlsConfig *corev1.PackageRepositoryTlsConfig
	var auth *corev1.PackageRepositoryAuth

	if source.Spec.TLSInsecureSkipVerify {
		tlsConfig = &corev1.PackageRepositoryTlsConfig{
			InsecureSkipVerify: source.Spec.TLSInsecureSkipVerify,
		}
	}
	if source.Spec.PassCredentials {
		auth = &corev1.PackageRepositoryAuth{
			PassCredentials: source.Spec.PassCredentials,
		}
	}

	if caSecret != nil {
		if _, ok := caSecret.Data[SecretCaKey]; ok {
			if tlsConfig == nil {
				tlsConfig = &corev1.PackageRepositoryTlsConfig{}
			}

			if isAuthSecretKubeappsManaged(source, caSecret) {
				tlsConfig.PackageRepoTlsConfigOneOf = &corev1.PackageRepositoryTlsConfig_CertAuthority{
					CertAuthority: RedactedString,
				}
			} else {
				tlsConfig.PackageRepoTlsConfigOneOf = &corev1.PackageRepositoryTlsConfig_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: caSecret.Name,
						Key:  SecretCaKey,
					},
				}
			}
		}
	}

	if authSecret != nil {
		if auth == nil {
			auth = &corev1.PackageRepositoryAuth{}
		}

		// find type
		if authHeader, ok := authSecret.Data[SecretAuthHeaderKey]; ok {
			if strings.HasPrefix(string(authHeader), "Basic") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			} else if strings.HasPrefix(string(authHeader), "Bearer") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
			} else {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
			}
		} else if _, ok := authSecret.Data[DockerConfigJsonKey]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
		} else {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
			log.Warningf("Unrecognized type of secret for auth [%s]", authSecret.Name)
		}

		// create data
		if isAuthSecretKubeappsManaged(source, authSecret) {
			switch auth.Type {
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: RedactedString,
						Password: RedactedString,
					},
				}
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: RedactedString,
				}
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: RedactedString,
				}
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: &corev1.DockerCredentials{
						Username: RedactedString,
						Password: RedactedString,
						Email:    RedactedString,
						Server:   RedactedString,
					},
				}
			}
		} else {
			switch auth.Type {
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
				corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
				corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: authSecret.Name,
						Key:  SecretAuthHeaderKey,
					},
				}
			case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: authSecret.Name,
						Key:  DockerConfigJsonKey,
					},
				}
			}
		}
	}

	return tlsConfig, auth, nil
}

func isAuthSecretKubeappsManaged(repo *apprepov1alpha1.AppRepository, secret *k8scorev1.Secret) bool {
	if isSecretKubeappsManaged(repo, secret) {
		return true
	}

	// note: until fully deprecated, we also check based on name pattern for backward compatibility
	return secret.GetName() == helm.SecretNameForRepo(repo.GetName())
}

func isImagesPullSecretKubeappsManaged(repo *apprepov1alpha1.AppRepository, secret *k8scorev1.Secret) bool {
	if isSecretKubeappsManaged(repo, secret) {
		return true
	}

	// note: until fully deprecated, we also check based on name pattern for backward compatibility
	return secret.GetName() == imagesPullSecretName(repo.GetName())
}

func isSecretKubeappsManaged(repo *apprepov1alpha1.AppRepository, secret *k8scorev1.Secret) bool {
	if !metav1.IsControlledBy(secret, repo) {
		return false
	}
	if managedby := secret.GetAnnotations()[Annotation_ManagedBy_Key]; managedby != Annotation_ManagedBy_Value {
		return false
	}
	return true
}

// utilities for secrets encoding/decocing

func encodeBasicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
	return auth
}

func decodeBasicAuth(auth string) (username string, password string, ok bool) {
	if strings.HasPrefix(auth, "Basic ") {
		auth = strings.TrimPrefix(auth, "Basic ")
	} else {
		return "", "", false
	}
	if bytes, err := base64.StdEncoding.DecodeString(auth); err != nil {
		return "", "", false
	} else {
		auth = string(bytes)
	}
	return strings.Cut(auth, ":")
}

func encodeBearerAuth(token string) string {
	return "Bearer " + token
}

func decodeBearerAuth(auth string) (token string, ok bool) {
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer "), true
	} else {
		return "", false
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
