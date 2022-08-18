// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func secretNameForRepo(repoName string) string {
	return fmt.Sprintf("apprepo-%s", repoName)
}

// Generate a suitable name for per-namespace repository secret
func namespacedSecretNameForRepo(repoName, namespace string) string {
	return fmt.Sprintf("%s-%s", namespace, secretNameForRepo(repoName))
}

func newLocalOpaqueSecret(ownerRepo types.NamespacedName) *k8scorev1.Secret {
	return &k8scorev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretNameForRepo(ownerRepo.Name),
		},
		Type: k8scorev1.SecretTypeOpaque,
		Data: map[string][]byte{},
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

func createKubeappsManagedRepoSecrets(
	ctx context.Context,
	typedClient kubernetes.Interface,
	namespace string,
	secret *k8scorev1.Secret, imagesPullSecret *k8scorev1.Secret) (newSecret *k8scorev1.Secret, newImgPullSecret *k8scorev1.Secret, err error) {

	if secret != nil {
		// create a secret first, if applicable
		secretName := secret.GetName()
		newSecret, err = typedClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, statuserror.FromK8sError("create", "secret", secretName, err)
		}
	}
	if imagesPullSecret != nil {
		secretName := imagesPullSecret.GetName()
		newImgPullSecret, err = typedClient.CoreV1().Secrets(namespace).Create(ctx, imagesPullSecret, metav1.CreateOptions{})
		if err != nil {
			return nil, nil, statuserror.FromK8sError("create", "secret", secretName, err)
		}
	}
	return newSecret, newImgPullSecret, nil
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
			Name:      namespacedSecretNameForRepo(repoName.Name, repoName.Namespace),
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

func updateKubeappsManagedImagesPullSecret(ctx context.Context, typedClient kubernetes.Interface,
	ownerRepo types.NamespacedName, credentials *corev1.DockerCredentials) (*k8scorev1.Secret, bool, error) {

	secretsInterface := typedClient.CoreV1().Secrets(ownerRepo.Namespace)
	secretName := imagesPullSecretName(ownerRepo.Name)
	existingSecret, _ := secretsInterface.Get(ctx, secretName, metav1.GetOptions{})

	// Check that the provided Docker credentials are ok
	if credentials != nil && (credentials.Server == "" || credentials.Username == "" || credentials.Password == "" || credentials.Email == "") {
		return nil, false, status.Errorf(codes.InvalidArgument, "Images pull secret Docker credentials are wrong")
	}

	// Remove any existing managed secret
	if existingSecret != nil {
		if err := deleteSecret(ctx, secretsInterface, secretName); err != nil {
			return nil, false, err
		}
	}

	if credentials != nil {
		newSecret, isSameSecret, err := newDockerImagePullSecret(ownerRepo, credentials)
		if err != nil {
			return nil, false, err
		}
		if !isSameSecret {
			createdSecret, err := secretsInterface.Create(ctx, newSecret, metav1.CreateOptions{})
			if err != nil {
				return nil, false, statuserror.FromK8sError("create", "secret", newSecret.GetGenerateName(), err)
			}
			return createdSecret, false, nil
		} else {
			return existingSecret, true, nil
		}
	}

	return nil, false, nil
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

func (s *Server) updateKubeappsManagedRepoSecret(ctx context.Context, repo *HelmRepository, existingSecretRef string) (secret *k8scorev1.Secret, updateRepo bool, err error) {
	secret, isSameSecret, err := newSecretFromTlsConfigAndAuth(repo.name, repo.tlsConfig, repo.auth)
	if err != nil {
		return nil, false, err
	} else if isSameSecret {
		// Do nothing if repo auth data came redacted
		return nil, false, nil
	}

	typedClient, err := s.clientGetter.Typed(ctx, repo.cluster)
	if err != nil {
		return nil, false, err
	}
	secretInterface := typedClient.CoreV1().Secrets(repo.name.Namespace)
	if secret != nil {
		if existingSecretRef == "" {
			// create a secret first
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return nil, false, statuserror.FromK8sError("create", "secret", secret.GetGenerateName(), err)
			}
			return newSecret, true, nil
		} else {
			// TODO (gfichtenholt) we should optimize this to somehow tell if the existing secret
			// is the same (data-wise) as the new one and if so skip all this
			if err = secretInterface.Delete(ctx, existingSecretRef, metav1.DeleteOptions{}); err != nil {
				return nil, false, statuserror.FromK8sError("get", "secret", existingSecretRef, err)
			}
			// create a new one
			newSecret, err := secretInterface.Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return nil, false, statuserror.FromK8sError("update", "secret", secret.GetGenerateName(), err)
			}
			return newSecret, true, nil
		}
	} else if existingSecretRef != "" {
		if err = secretInterface.Delete(ctx, existingSecretRef, metav1.DeleteOptions{}); err != nil {
			log.Errorf("Error deleting existing secret: [%s] due to %v", err)
		}
	}
	return secret, true, nil
}

func getRepoImagesPullSecretWithUserManagedSecrets(imagesPullSecret *k8scorev1.Secret) *v1alpha1.ImagesPullSecret_SecretRef {
	if imagesPullSecret != nil {
		return &v1alpha1.ImagesPullSecret_SecretRef{
			SecretRef: imagesPullSecret.Name,
		}
	}
	return nil
}

func getRepoImagesPullSecretWithKubeappsManagedSecrets(imagesPullSecret *k8scorev1.Secret) *v1alpha1.ImagesPullSecret_Credentials {
	if imagesPullSecret != nil {
		return &v1alpha1.ImagesPullSecret_Credentials{
			Credentials: &corev1.DockerCredentials{
				Server:   RedactedString,
				Username: RedactedString,
				Password: RedactedString,
				Email:    RedactedString,
			},
		}
	}
	return nil
}

func getRepoTlsConfigAndAuthWithUserManagedSecrets(source *apprepov1alpha1.AppRepository,
	caSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
	tlsConfig := &corev1.PackageRepositoryTlsConfig{
		InsecureSkipVerify: source.Spec.TLSInsecureSkipVerify,
	}
	auth := &corev1.PackageRepositoryAuth{
		PassCredentials: source.Spec.PassCredentials,
	}

	if caSecret != nil {
		if _, ok := caSecret.Data[SecretCaKey]; ok {
			tlsConfig.PackageRepoTlsConfigOneOf = &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: caSecret.Name,
					Key:  SecretCaKey,
				},
			}
		}
	}
	if authSecret != nil {
		if authHeader, ok := authSecret.Data[SecretAuthHeaderKey]; ok {
			if strings.HasPrefix(string(authHeader), "Basic") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
			} else if strings.HasPrefix(string(authHeader), "Bearer") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
			} else {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
			}
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: authSecret.Name,
					Key:  SecretAuthHeaderKey,
				},
			}
		} else if _, ok := authSecret.Data[DockerConfigJsonKey]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: authSecret.Name,
					Key:  DockerConfigJsonKey,
				},
			}
		} else {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
			log.Warning("Unrecognized type of secret for auth [%s]", authSecret.Name)
		}
	}
	return tlsConfig, auth, nil
}

func getRepoTlsConfigAndAuthWithKubeappsManagedSecrets(source *apprepov1alpha1.AppRepository,
	caSecret *k8scorev1.Secret, authSecret *k8scorev1.Secret) (*corev1.PackageRepositoryTlsConfig, *corev1.PackageRepositoryAuth, error) {
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
			tlsConfig.PackageRepoTlsConfigOneOf = &corev1.PackageRepositoryTlsConfig_CertAuthority{
				CertAuthority: RedactedString,
			}
		}
	}

	if authSecret != nil {
		if auth == nil {
			auth = &corev1.PackageRepositoryAuth{}
		}
		if authHeader, ok := authSecret.Data[SecretAuthHeaderKey]; ok {
			if strings.HasPrefix(string(authHeader), "Basic") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: RedactedString,
						Password: RedactedString,
					},
				}
			} else if strings.HasPrefix(string(authHeader), "Bearer") {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: RedactedString,
				}
			} else {
				auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER
				auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_Header{
					Header: RedactedString,
				}
			}

		} else if _, ok := authSecret.Data[DockerConfigJsonKey]; ok {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON
			auth.PackageRepoAuthOneOf = &corev1.PackageRepositoryAuth_DockerCreds{
				DockerCreds: &corev1.DockerCredentials{
					Username: RedactedString,
					Password: RedactedString,
					Email:    RedactedString,
					Server:   RedactedString,
				},
			}
		} else {
			auth.Type = corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_UNSPECIFIED
			log.Warning("Unrecognized type of secret for auth [%s]", authSecret.Name)
		}
	}
	return tlsConfig, auth, nil
}
