package main

import (
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	addRepo1 = v1alpha1.AppRepository{
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

	addRepo2 = v1alpha1.AppRepository{
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

	addRepo3 = v1alpha1.AppRepository{
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

	addRepo4 = v1alpha1.AppRepository{
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
					SecretKeyRef: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{
						Name: "apprepo-bar"},
						Key: "authorizationHeader",
					},
				},
			},
			PassCredentials: true,
		},
	}

	addRepoBearerToken = v1alpha1.AppRepository{
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
					SecretKeyRef: v1.SecretKeySelector{LocalObjectReference: v1.LocalObjectReference{
						Name: "apprepo-bar"},
						Key: "authorizationHeader",
					},
				},
			},
		},
	}

	addRepo5 = v1alpha1.AppRepository{
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

	addRepoReq5 = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			InsecureSkipVerify: true,
		},
	}

	addRepoReq4 = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
	}

	addRepoReq6 = func(ca []byte) *corev1.AddPackageRepositoryRequest {
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

	addRepoReq7 = &corev1.AddPackageRepositoryRequest{
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

	addRepoReq8 = &corev1.AddPackageRepositoryRequest{
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
					Password: "zot",
				},
			},
			PassCredentials: true,
		},
	}

	/*	addRepoReq9 = func(pub, priv []byte) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo"},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth:            tlsAuth(pub, priv),
		}
	}*/

	addRepoReq10 = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
				Header: "foobarzot",
			},
		},
	}

	addRepoReq11 = &corev1.AddPackageRepositoryRequest{
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

	addRepoReq12 = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
				DockerCreds: &corev1.DockerCredentials{
					Server:   "your.private.registry.example.com",
					Username: "janedoe",
					Password: "xxxxxxxx",
					Email:    "jdoe@example.com",
				},
			},
		},
	}

	addRepoReq13 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth:    secret1Auth,
	}

	addRepoReq14 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
		Auth:    secret1Auth,
		TlsConfig: &corev1.PackageRepositoryTlsConfig{
			PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-2",
				},
			},
		},
	}

	/*add_repo_req_18 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-4",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_basic_auth_repo_url,
		Auth:    secret1Auth,
	}

	add_repo_req_19 = &corev1.AddPackageRepositoryRequest{
		Name:    "my-podinfo-4",
		Context: &corev1.Context{Namespace: "default"},
		Type:    "helm",
		Url:     podinfo_tls_repo_url,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_TLS,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
				SecretRef: &corev1.SecretKeyReference{
					Name: "secret-2",
				},
			},
		},
	}*/

	addRepoReq20 = &corev1.AddPackageRepositoryRequest{
		Name:    "bar",
		Context: &corev1.Context{Namespace: "foo"},
		Type:    "helm",
		Url:     "http://example.com",
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

	secret1Auth = &corev1.PackageRepositoryAuth{
		Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
		PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
			SecretRef: &corev1.SecretKeyReference{
				Name: "secret-1",
			},
		},
	}
)
