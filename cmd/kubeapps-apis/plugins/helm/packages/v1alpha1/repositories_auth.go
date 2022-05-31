// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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
					secret.Data[SecretAuthHeaderKey] = []byte("Bearer " + token)
				}
			} else {
				return nil, false, status.Errorf(codes.InvalidArgument, "Bearer token is missing")
			}
		case corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM:
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
			corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM:
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
	secret *k8scorev1.Secret) (*k8scorev1.Secret, error) {

	if secret != nil {
		// create a secret first, if applicable
		if secret, err := typedClient.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{}); err != nil {
			return nil, statuserror.FromK8sError("create", "secret", secret.GetName(), err)
		}
	}
	return secret, nil
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
					corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM,
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
