// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	addRepoSimpleHelm = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			Type: "helm",
			URL:  "http://example.com",
		},
	}

	addRepoSimpleOci = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			Type: "oci",
			URL:  "http://example.com",
		},
	}

	addRepoWithTLSCA = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:  "http://example.com",
			Type: "helm",
			Auth: v1alpha1.AppRepositoryAuth{
				CustomCA: &v1alpha1.AppRepositoryCustomCA{
					SecretKeyRef: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{Name: "apprepo-bar"}, Key: "ca.crt"},
				},
			},
		},
	}

	addRepoTLSSecret = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:  "http://example.com",
			Type: "helm",
			Auth: v1alpha1.AppRepositoryAuth{
				CustomCA: &v1alpha1.AppRepositoryCustomCA{
					SecretKeyRef: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{Name: "secret-1"}, Key: "ca.crt"},
				},
			},
		},
	}

	addRepoAuthHeaderPassCredentials = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:  "http://example.com",
			Type: "helm",
			Auth: v1alpha1.AppRepositoryAuth{
				Header: &v1alpha1.AppRepositoryAuthHeader{
					SecretKeyRef: v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "apprepo-bar",
						},
						Key: "authorizationHeader",
					},
				},
			},
			PassCredentials: true,
		},
	}

	addRepoAuthHeaderWithSecretRef = func(secretName string) *v1alpha1.AppRepository {
		return &v1alpha1.AppRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       AppRepositoryKind,
				APIVersion: AppRepositoryApi,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       "foo",
				ResourceVersion: "1",
			},
			Spec: v1alpha1.AppRepositorySpec{
				URL:  "http://example.com",
				Type: "helm",
				Auth: v1alpha1.AppRepositoryAuth{
					Header: &v1alpha1.AppRepositoryAuthHeader{
						SecretKeyRef: v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: secretName,
							},
							Key: "authorizationHeader",
						},
					},
				},
			},
		}
	}

	addRepoOnlyPassCredentials = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "foo",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			URL:             "http://example.com",
			Type:            "helm",
			PassCredentials: true,
		},
	}

	addRepoGlobal = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       "kubeapps",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.AppRepositorySpec{
			Type: "helm",
			URL:  "http://example.com",
		},
	}

	addRepoReqWrongType = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "foobar",
		Url:     "http://example.com",
	}

	addRepoReqNoUrl = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
	}

	addRepoReqGlobal = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "kubeapps"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: false,
	}

	addRepoReqTlsSkipVerify = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			InsecureSkipVerify: true,
		},
	}

	addRepoReqSimple = func(repoType string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            repoType,
			Url:             "http://example.com",
			NamespaceScoped: true,
		}
	}

	addRepoReqTLSCA = func(ca []byte) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			TlsConfig: &corev1.PackageRepositoryTlsConfig{
				PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
					CertAuthority: string(ca),
				},
			},
		}
	}

	addRepoReqTLSSecretRef = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-1",
				},
			},
		},
	}

	addRepoReqBasicAuth = func(username, password string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
					UsernamePassword: &corev1.UsernamePassword{
						Username: username,
						Password: password,
					},
				},
				PassCredentials: true,
			},
		}
	}

	addRepoReqWrongBasicAuth = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
				UsernamePassword: &corev1.UsernamePassword{
					Username: "baz",
				},
			},
			PassCredentials: true,
		},
	}

	addRepoReqBearerToken = func(token string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
					Header: token,
				},
			},
		}
	}

	addRepoReqAuthWithSecret = func(authType corev1.PackageRepositoryAuth_PackageRepositoryAuthType, secretName string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth: &corev1.PackageRepositoryAuth{
				Type: authType,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
					SecretRef: &corev1.SecretKeyReference{
						Name: secretName,
					},
				},
			},
		}
	}

	addRepoReqCustomAuth = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
				Header: "foobarzot",
			},
		},
	}

	addRepoReqTLSDifferentSecretAuth = &corev1.AddPackageRepositoryRequest{
		Name:      "bar",
		Context:   &corev1.Context{Namespace: "foo"},
		Type:      "helm",
		Url:       "http://example.com",
		Auth:      packageRepoSecretBasicAuth("secret-1"),
		TlsConfig: packageRepoSecretTls("secret-2"),
	}

	addRepoReqOnlyPassCredentials = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			PassCredentials: true,
		},
	}

	addRepoExpectedResp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("bar", "foo"),
	}

	addRepoExpectedGlobalResp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("bar", "kubeapps"),
	}

	packageRepoSecretBasicAuth = func(secretName string) *corev1.PackageRepositoryAuth {
		return &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-1",
				},
			},
		}
	}

	packageRepoSecretTls = func(secretName string) *corev1.PackageRepositoryTlsConfig {
		return &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: secretName,
				},
			},
		}
	}
)
