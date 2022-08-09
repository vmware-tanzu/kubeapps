// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	helmv1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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

	addRepoAuthHeaderPassCredentials = func(namespace string) *v1alpha1.AppRepository {
		return &v1alpha1.AppRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       AppRepositoryKind,
				APIVersion: AppRepositoryApi,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       namespace,
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
	}

	addRepoAuthHeaderWithSecretRef = func(namespace, secretName string) *v1alpha1.AppRepository {
		return &v1alpha1.AppRepository{
			TypeMeta: metav1.TypeMeta{
				Kind:       AppRepositoryKind,
				APIVersion: AppRepositoryApi,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            "bar",
				Namespace:       namespace,
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

	addRepoAuthDocker = func(secretName string) *v1alpha1.AppRepository {
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
							Key: ".dockerconfigjson",
						},
					},
				},
			},
		}
	}

	addRepoCustomDetailHelm = v1alpha1.AppRepository{
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
			URL:             "https://example.com",
			Type:            "helm",
			OCIRepositories: []string{"repo1", "repo2"},
			FilterRule: v1alpha1.FilterRuleSpec{
				JQ:        ".name == $var0 or .name == $var1",
				Variables: map[string]string{"$var0": "package1", "$var1": "package2"},
			},
		},
	}

	addRepoCustomDetailOci = v1alpha1.AppRepository{
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
			URL:             "https://example.com",
			Type:            "oci",
			OCIRepositories: []string{"repo1"},
		},
	}

	addRepoGlobal = v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       AppRepositoryKind,
			APIVersion: AppRepositoryApi,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            "bar",
			Namespace:       globalPackagingNamespace,
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
		Context:         &corev1.Context{Namespace: globalPackagingNamespace, Cluster: KubeappsCluster},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: false,
	}

	addRepoReqSimple = func(repoType string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
			Type:            repoType,
			Url:             "http://example.com",
			NamespaceScoped: true,
		}
	}

	addRepoReqTLSCA = func(ca []byte) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
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
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
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
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
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

	addRepoReqBearerToken = func(token string, withPrefix bool) *corev1.AddPackageRepositoryRequest {
		prefixedToken := token
		if withPrefix {
			prefixedToken = "Bearer " + token
		}

		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
					Header: prefixedToken,
				},
			},
		}
	}

	addRepoReqAuthWithSecret = func(authType corev1.PackageRepositoryAuth_PackageRepositoryAuthType, namespace, secretName string) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: namespace, Cluster: KubeappsCluster},
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
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
			PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
				Header: "foobarzot",
			},
		},
	}

	addRepoReqDockerAuth = func(credentials *corev1.DockerCredentials) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
			Type:            "helm",
			Url:             "http://example.com",
			NamespaceScoped: true,
			Auth: &corev1.PackageRepositoryAuth{
				Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
				PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
					DockerCreds: credentials,
				},
			},
		}
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
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		Auth: &corev1.PackageRepositoryAuth{
			PassCredentials: true,
		},
	}

	addRepoReqCustomValues = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
		Type:            "helm",
		Url:             "https://example.com",
		NamespaceScoped: true,
		CustomDetail: toProtoBufAny(&helmv1.HelmPackageRepositoryCustomDetail{
			OciRepositories: []string{"repo1", "repo2"},
			FilterRule: &helmv1.RepositoryFilterRule{
				Jq:        ".name == $var0 or .name == $var1",
				Variables: map[string]string{"$var0": "package1", "$var1": "package2"},
			},
		}),
	}

	addRepoReqCustomValuesHelmValid = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
		Type:            "helm",
		Url:             "https://example.com",
		NamespaceScoped: true,
		CustomDetail: toProtoBufAny(&helmv1.HelmPackageRepositoryCustomDetail{
			OciRepositories: []string{"repo1", "repo2"},
			FilterRule: &helmv1.RepositoryFilterRule{
				Jq:        ".name == $var0 or .name == $var1",
				Variables: map[string]string{"$var0": "package1", "$var1": "package2"},
			},
			PerformValidation: true,
		}),
	}

	addRepoReqCustomValuesOCIValid = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
		Type:            "oci",
		Url:             "https://example.com",
		NamespaceScoped: true,
		CustomDetail: toProtoBufAny(&helmv1.HelmPackageRepositoryCustomDetail{
			OciRepositories:   []string{"repo1"},
			PerformValidation: true,
		}),
	}

	addRepoReqWrongCustomValues = &corev1.AddPackageRepositoryRequest{
		Name:            "bar",
		Context:         &corev1.Context{Namespace: "foo"},
		Type:            "helm",
		Url:             "http://example.com",
		NamespaceScoped: true,
		CustomDetail: toProtoBufAny(&helmv1.RepositoryFilterRule{
			Jq: "wrong-struct",
		}),
	}

	toProtoBufAny = func(src proto.Message) *anypb.Any {
		if anyObj, err := anypb.New(src); err != nil {
			return nil
		} else {
			return anyObj
		}
	}

	addRepoExpectedResp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("bar", KubeappsCluster, "foo"),
	}

	addRepoExpectedGlobalResp = &corev1.AddPackageRepositoryResponse{
		PackageRepoRef: repoRef("bar", KubeappsCluster, globalPackagingNamespace),
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
