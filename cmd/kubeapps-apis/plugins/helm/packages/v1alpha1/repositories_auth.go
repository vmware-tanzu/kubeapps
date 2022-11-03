// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	"strings"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	log "k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/credentialprovider"
)

const (
	SecretCaKey         = "ca.crt"
	SecretAuthHeaderKey = "authorizationHeader"
	DockerConfigJsonKey = ".dockerconfigjson"
)

func newLocalOpaqueSecret(ownerRepo types.NamespacedName) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: helm.SecretNameForRepo(ownerRepo.Name),
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
		return nil, false, status.Errorf(codes.InvalidArgument, "Package repository cannot mix referenced secrets and user provided secret data")
	}

	// create/get secret
	if hasCaRef || hasAuthRef {
		secret, err := validateUserManagedRepoSecret(ctx, typedClient, repoName, tlsConfig, auth)
		return secret, false, err
	} else if hasCaData || hasAuthData {
		secret, _, err := newSecretFromTlsConfigAndAuth(repoName, tlsConfig, auth)
		return secret, true, err
	} else {
		return nil, false, nil
	}
}

func handleImagesPullSecretForCreate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repo *HelmRepository) (*k8scorev1.Secret, bool, error) {

	hasRef := repo.customDetail != nil && repo.customDetail.ImagesPullSecret != nil && repo.customDetail.ImagesPullSecret.GetSecretRef() != ""
	hasData := repo.customDetail != nil && repo.customDetail.ImagesPullSecret != nil && repo.customDetail.ImagesPullSecret.GetCredentials() != nil

	// create/get secret
	if hasRef {
		secret, err := validateDockerImagePullSecret(ctx, typedClient, repo.name, repo.customDetail.ImagesPullSecret.GetSecretRef())
		return secret, false, err
	} else if hasData {
		secret, _, err := newDockerImagePullSecret(repo.name, repo.customDetail.ImagesPullSecret.GetCredentials())
		return secret, true, err
	} else {
		return nil, false, nil
	}
}

func handleAuthSecretForUpdate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth,
	secret *k8scorev1.Secret) (updatedSecret *k8scorev1.Secret, secretIsKubeappsManaged bool, secretIsUpdated bool, err error) {

	hasCaRef := tlsConfig != nil && tlsConfig.GetSecretRef() != nil
	hasCaData := tlsConfig != nil && tlsConfig.GetCertAuthority() != ""
	hasAuthRef := auth != nil && auth.GetSecretRef() != nil
	hasAuthData := auth != nil && auth.GetSecretRef() == nil && auth.GetType() != corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED

	// if we have both ref config and data config, it is an invalid mixed configuration
	if (hasCaRef || hasAuthRef) && (hasCaData || hasAuthData) {
		return nil, false, false, status.Errorf(codes.InvalidArgument, "Package repository cannot mix referenced secrets and user provided secret data")
	}

	// check we cannot change mode (per design spec)
	if secret != nil && (hasCaRef || hasCaData || hasAuthRef || hasAuthData) {
		if isAuthSecretKubeappsManaged(repoName.Name, secret) != (hasAuthData || hasCaData) {
			return nil, false, false, status.Errorf(codes.InvalidArgument, "Auth management mode cannot be changed")
		}
	}

	// handle user managed secret
	if hasCaRef || hasAuthRef {
		updatedSecret, err := validateUserManagedRepoSecret(ctx, typedClient, repoName, tlsConfig, auth)
		return updatedSecret, false, true, err
	}

	// handle kubeapps managed secret
	updatedSecret, isSameSecret, err := newSecretFromTlsConfigAndAuth(repoName, tlsConfig, auth)
	if err != nil {
		return nil, true, false, err
	} else if isSameSecret {
		// Do nothing if repo auth data came redacted
		return nil, true, false, nil
	} else {
		// either we have no secret, or it has changed. in both cases, we try to delete any existing secret
		if secret != nil {
			secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
			if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", err)
			}
		}

		return updatedSecret, true, true, nil
	}
}

func handleImagesPullSecretForUpdate(
	ctx context.Context,
	typedClient kubernetes.Interface,
	repo *HelmRepository,
	secret *k8scorev1.Secret) (updatedSecret *k8scorev1.Secret, secretIsKubeappsManaged bool, secretIsUpdated bool, err error) {

	var imagesPullSecrets *v1alpha1.ImagesPullSecret
	if repo.customDetail != nil && repo.customDetail.ImagesPullSecret != nil {
		imagesPullSecrets = repo.customDetail.ImagesPullSecret
	} else {
		imagesPullSecrets = &v1alpha1.ImagesPullSecret{}
	}

	hasRef := imagesPullSecrets != nil && imagesPullSecrets.GetSecretRef() != ""
	hasData := imagesPullSecrets != nil && imagesPullSecrets.GetCredentials() != nil

	// check we are not changing mode
	if secret != nil && (hasRef || hasData) {
		if isImagesPullSecretKubeappsManaged(repo.name.Name, secret) != hasData {
			return nil, false, false, status.Errorf(codes.InvalidArgument, "Auth management mode cannot be changed")
		}
	}

	// handle user managed secret
	if hasRef {
		updatedSecret, err := validateDockerImagePullSecret(ctx, typedClient, repo.name, imagesPullSecrets.GetSecretRef())
		return updatedSecret, false, true, err
	}

	// handle kubeapps managed secret
	updatedSecret, isSameSecret, err := newDockerImagePullSecret(repo.name, imagesPullSecrets.GetCredentials())
	if err != nil {
		return nil, true, false, err
	} else if isSameSecret {
		// Do nothing if repo credential data came redacted
		return nil, true, false, nil
	} else {
		// either we have no secret, or it has changed. in both cases, we try to delete any existing secret
		if secret != nil {
			secretsInterface := typedClient.CoreV1().Secrets(secret.GetNamespace())
			if err := deleteSecret(ctx, secretsInterface, secret.GetName()); err != nil {
				log.Errorf("Error deleting existing secret: [%s] due to %v", err)
			}
		}

		return updatedSecret, true, true, nil
	}
}

// this func is only used with kubeapps-managed secrets
func newSecretFromTlsConfigAndAuth(repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (secret *k8scorev1.Secret, isSameSecret bool, err error) {
	if tlsConfig != nil {
		if tlsConfig.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		caCert := tlsConfig.GetCertAuthority()
		if caCert == RedactedString {
			isSameSecret = true
		} else if caCert != "" {
			secret = newLocalOpaqueSecret(repoName)
			secret.Data[SecretCaKey] = []byte(caCert)
		}
	}
	if auth != nil {
		if auth.GetSecretRef() != nil {
			return nil, false, status.Errorf(codes.InvalidArgument, "SecretRef may not be used with kubeapps managed secrets")
		}
		if secret == nil {
			secret = newLocalOpaqueSecret(repoName)
		}
		switch auth.Type {
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
			if unp := auth.GetUsernamePassword(); unp != nil {
				if unp.Username == "" || unp.Password == "" {
					return nil, false, status.Errorf(codes.InvalidArgument, "Wrong combination of username and password")
				} else if unp.Username == RedactedString && unp.Password == RedactedString {
					isSameSecret = true
				} else {
					authString := fmt.Sprintf("%s:%s", unp.Username, unp.Password)
					authHeader := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authString)))
					secret.Data[SecretAuthHeaderKey] = []byte(authHeader)
				}
			} else {
				return nil, false, status.Errorf(codes.Internal, "Username/Password configuration is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER:
			if token := auth.GetHeader(); token != "" {
				if token == RedactedString {
					isSameSecret = true
				} else {
					secret.Data[SecretAuthHeaderKey] = []byte("Bearer " + strings.TrimPrefix(token, "Bearer "))
				}
			} else {
				return nil, false, status.Errorf(codes.InvalidArgument, "Bearer token is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER:
			if authHeaderValue := auth.GetHeader(); authHeaderValue != "" {
				if authHeaderValue == RedactedString {
					isSameSecret = true
				} else {
					secret.Data[SecretAuthHeaderKey] = []byte(authHeaderValue)
				}
			} else {
				return nil, false, status.Errorf(codes.InvalidArgument, "Authentication header value is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if dockerCreds := auth.GetDockerCreds(); dockerCreds != nil {
				if dockerCreds.Password == RedactedString {
					isSameSecret = true
				} else {
					secret.Type = k8scorev1.SecretTypeDockerConfigJson
					dockerConfig := &credentialprovider.DockerConfigJSON{
						Auths: map[string]credentialprovider.DockerConfigEntry{
							dockerCreds.Server: {
								Username: dockerCreds.Username,
								Password: dockerCreds.Password,
								Email:    dockerCreds.Email,
							},
						},
					}
					dockerConfigJson, err := json.Marshal(dockerConfig)
					if err != nil {
						return nil, false, status.Errorf(codes.InvalidArgument, "Docker credentials are wrong")
					}
					secret.Data[DockerConfigJsonKey] = dockerConfigJson
				}
			} else {
				return nil, false, status.Errorf(codes.InvalidArgument, "Docker credentials are missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS:
			return nil, false, status.Errorf(codes.Unimplemented, "Package repository authentication type %q is not supported", auth.Type)
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, true, nil
		default:
			return nil, false, status.Errorf(codes.Internal, "Unexpected package repository authentication type: %q", auth.Type)
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
			return nil, status.Errorf(codes.InvalidArgument, "Secret for AppRepository auth is missing")
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
				return nil, status.Errorf(codes.InvalidArgument, "Secret for AppRepository auth is missing")
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
				return nil, status.Errorf(codes.InvalidArgument, "Authentication header is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
			if secret == nil {
				return nil, status.Errorf(codes.InvalidArgument, "Secret for AppRepository auth is missing")
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
			return nil, status.Errorf(codes.Unimplemented, "Package repository authentication type %q is not supported", auth.Type)
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED:
			return nil, nil
		default:
			return nil, status.Errorf(codes.Internal, "Unexpected package repository authentication type: %q", auth.Type)
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
			return nil, statuserror.FromK8sError("create", "secret", secretName, err)
		}
	}
	return newSecret, nil
}

func validateDockerImagePullSecret(ctx context.Context,
	typedClient kubernetes.Interface,
	repoName types.NamespacedName,
	secretName string) (*k8scorev1.Secret, error) {

	if secret, err := typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretName, metav1.GetOptions{}); err != nil {
		return nil, statuserror.FromK8sError("get", "secret", secretName, err)
	} else if secret.Type != k8scorev1.SecretTypeDockerConfigJson {
		return nil, status.Errorf(codes.InvalidArgument, "Images Docker pull secret %s does not have valid type", secretName)
	} else if _, ok := secret.Data[k8scorev1.DockerConfigJsonKey]; !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Images Docker pull secret %s does not have valid data", secretName)
	} else {
		return secret, nil
	}
}

func imagesPullSecretName(repoName string) string {
	return fmt.Sprintf("pullsecret-%s", repoName)
}

func newDockerImagePullSecret(ownerRepo types.NamespacedName, credentials *corev1.DockerCredentials) (secret *k8scorev1.Secret, isSameSecret bool, err error) {
	if credentials != nil {
		if credentials.Server == "" || credentials.Username == "" || credentials.Password == "" || credentials.Email == "" {
			return nil, false, status.Errorf(codes.InvalidArgument, "Images pull secret Docker credentials are wrong")
		}

		secret = &k8scorev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: imagesPullSecretName(ownerRepo.Name),
			},
			Type: k8scorev1.SecretTypeDockerConfigJson,
			Data: map[string][]byte{},
		}

		if credentials.Password == RedactedString {
			isSameSecret = true
		} else {
			dockerConfig := &credentialprovider.DockerConfigJSON{
				Auths: map[string]credentialprovider.DockerConfigEntry{
					credentials.Server: {
						Username: credentials.Username,
						Password: credentials.Password,
						Email:    credentials.Email,
					},
				},
			}
			dockerConfigJson, err := json.Marshal(dockerConfig)
			if err != nil {
				return nil, false, status.Errorf(codes.InvalidArgument, "Docker credentials are malformed")
			}
			secret.Data[k8scorev1.DockerConfigJsonKey] = dockerConfigJson
		}
	}
	return secret, isSameSecret, nil
}

func deleteSecret(ctx context.Context, secretsInterface v1.SecretInterface, secretName string) error {
	// Ignore action if secret didn't exist
	if _, err := secretsInterface.Get(ctx, secretName, metav1.GetOptions{}); err == nil {
		if err := secretsInterface.Delete(ctx, secretName, metav1.DeleteOptions{}); err != nil {
			return statuserror.FromK8sError("delete", "secret", secretName, err)
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
	repoName types.NamespacedName,
	tlsConfig *corev1.PackageRepositoryTlsConfig,
	auth *corev1.PackageRepositoryAuth) (*k8scorev1.Secret, error) {
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
		return nil, status.Errorf(
			codes.InvalidArgument, "TLS config secret and Auth secret must be the same")
	} else if secretRefTls != "" {
		secretRef = secretRefTls
	} else if secretRefAuth != "" {
		secretRef = secretRefAuth
	}

	var secret *k8scorev1.Secret
	if secretRef != "" {
		var err error
		// check that the specified secret exists
		if secret, err = typedClient.CoreV1().Secrets(repoName.Namespace).Get(ctx, secretRef, metav1.GetOptions{}); err != nil {
			return nil, statuserror.FromK8sError("get", "secret", secretRef, err)
		} else {
			// also check that the data in the opaque secret corresponds
			// to specified auth type, e.g. if AuthType is
			// PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			// check that the secret data has valid fields
			if secretRefTls != "" && secret.Data[SecretCaKey] == nil {
				return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing key '%s'", secretRef, SecretCaKey)
			}
			if secretRefAuth != "" {
				switch auth.Type {
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
					corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
					corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH:
					if secret.Data[SecretAuthHeaderKey] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing key '%s'", secretRef, SecretAuthHeaderKey)
					}
				case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON:
					if secret.Data[DockerConfigJsonKey] == nil {
						return nil, status.Errorf(codes.Internal, "Specified secret [%s] missing key '%s'", secretRef, DockerConfigJsonKey)
					}
				default:
					return nil, status.Errorf(codes.Internal, "Package repository authentication type %q is not supported", auth.Type)
				}
			}
		}
	}
	return secret, nil
}

func getRepoImagesPullSecret(source *apprepov1alpha1.AppRepository, imagesPullSecret *k8scorev1.Secret) *v1alpha1.ImagesPullSecret {
	if imagesPullSecret == nil {
		return nil
	} else if isImagesPullSecretKubeappsManaged(source.GetName(), imagesPullSecret) {
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

			if isAuthSecretKubeappsManaged(source.GetName(), caSecret) {
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
			log.Warning("Unrecognized type of secret for auth [%s]", authSecret.Name)
		}

		// create data
		if isAuthSecretKubeappsManaged(source.GetName(), authSecret) {
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

// note: for now, checking based on name pattern, for backward compatibility
func isAuthSecretKubeappsManaged(repoName string, secret *k8scorev1.Secret) bool {
	return secret.GetName() == helm.SecretNameForRepo(repoName)
}

// note: for now, checking based on name pattern, for backward compatibility
func isImagesPullSecretKubeappsManaged(repoName string, secret *k8scorev1.Secret) bool {
	return secret.GetName() == imagesPullSecretName(repoName)
}
