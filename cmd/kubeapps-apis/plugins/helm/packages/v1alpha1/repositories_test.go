// Copyright 2022-2024 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/bufbuild/connect-go"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appRepov1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	appRepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/pkg/helm"
	"google.golang.org/protobuf/types/known/anypb"
	authorizationv1 "k8s.io/api/authorization/v1"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	k8stesting "k8s.io/client-go/testing"
)

var plugin = &plugins.Plugin{
	Name:    "helm.packages",
	Version: "v1alpha1",
}

var repo1 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-1",
		Namespace:       "ns-1",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:         "https://test-repo",
		Type:        "helm",
		Description: "description 1",
	},
}

var repo2 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-2",
		Namespace:       "ns-2",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:         "https://test-repo2",
		Type:        "oci",
		Description: "description 2",
	},
}

var repo3 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-3",
		Namespace:       globalPackagingNamespace,
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:         "https://test-repo3",
		Type:        "helm",
		Description: "description 3",
		Auth: appRepov1alpha1.AppRepositoryAuth{
			Header: &appRepov1alpha1.AppRepositoryAuthHeader{
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: helm.SecretNameForRepo("repo-3")}, Key: "authorizationHeader"},
			},
		},
	},
}

var repo4 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-4",
		Namespace:       "ns-4",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:                   "https://test-repo4",
		Type:                  "oci",
		Description:           "description 4",
		TLSInsecureSkipVerify: true,
		Auth: appRepov1alpha1.AppRepositoryAuth{
			CustomCA: &appRepov1alpha1.AppRepositoryCustomCA{
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: helm.SecretNameForRepo("repo-4")}, Key: "ca.crt"},
			},
		},
		OCIRepositories: []string{"oci-repo-1", "oci-repo-2"},
		PassCredentials: true,
		FilterRule: appRepov1alpha1.FilterRuleSpec{
			JQ:        "jq-filter-$1",
			Variables: map[string]string{"$1": "value1"},
		},
	},
}

var repo5 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-5",
		Namespace:       "ns-5",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:                   "https://test-repo5",
		Type:                  "helm",
		Description:           "description 5",
		DockerRegistrySecrets: []string{imagesPullSecretName("repo-5")},
	},
}

var repo6 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-6",
		Namespace:       "ns-6",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:  "https://test-repo6",
		Type: "helm",
		Auth: appRepov1alpha1.AppRepositoryAuth{
			Header: &appRepov1alpha1.AppRepositoryAuthHeader{
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: helm.SecretNameForRepo("repo-6")}, Key: "authorizationHeader"},
			},
			CustomCA: &appRepov1alpha1.AppRepositoryCustomCA{
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: helm.SecretNameForRepo("repo-6")}, Key: "ca.crt"},
			},
		},
	},
}

var repo7 = &appRepov1alpha1.AppRepository{
	TypeMeta: metav1.TypeMeta{
		APIVersion: appReposAPIVersion,
		Kind:       AppRepositoryKind,
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:            "repo-7",
		Namespace:       "ns-7",
		ResourceVersion: "1",
	},
	Spec: appRepov1alpha1.AppRepositorySpec{
		URL:  "https://test-repo7",
		Type: "helm",
		Auth: appRepov1alpha1.AppRepositoryAuth{
			Header: &appRepov1alpha1.AppRepositoryAuthHeader{
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: helm.SecretNameForRepo("repo-7")}, Key: "authorizationHeader"},
			},
		},
	},
}

var repo1Summary = &corev1.PackageRepositorySummary{
	PackageRepoRef:  repoRef("repo-1", KubeappsCluster, "ns-1"),
	Name:            "repo-1",
	Description:     "description 1",
	NamespaceScoped: true,
	Type:            "helm",
	Url:             "https://test-repo",
	RequiresAuth:    false,
	Status:          &corev1.PackageRepositoryStatus{Ready: true},
}

var repo2Summary = &corev1.PackageRepositorySummary{
	PackageRepoRef:  repoRef("repo-2", KubeappsCluster, "ns-2"),
	Name:            "repo-2",
	Description:     "description 2",
	NamespaceScoped: true,
	Type:            "oci",
	Url:             "https://test-repo2",
	RequiresAuth:    false,
	Status:          &corev1.PackageRepositoryStatus{Ready: true},
}

var repo3Summary = &corev1.PackageRepositorySummary{
	PackageRepoRef:  repoRef("repo-3", KubeappsCluster, globalPackagingNamespace),
	Name:            "repo-3",
	Description:     "description 3",
	NamespaceScoped: false,
	Type:            "helm",
	Url:             "https://test-repo3",
	RequiresAuth:    true,
	Status:          &corev1.PackageRepositoryStatus{Ready: true},
}

var appReposAPIVersion = fmt.Sprintf("%s/%s", appRepov1alpha1.SchemeGroupVersion.Group, appRepov1alpha1.SchemeGroupVersion.Version)

func TestAddPackageRepository(t *testing.T) {
	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	ca, _, _ := getCertsForTesting(t)

	newPackageRepoRequestWithDetails := func(customDetail *v1alpha1.HelmPackageRepositoryCustomDetail) *corev1.AddPackageRepositoryRequest {
		return &corev1.AddPackageRepositoryRequest{
			Name:            "bar",
			Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
			Type:            "helm",
			Url:             "https://example.com",
			NamespaceScoped: true,
			CustomDetail:    toProtoBufAny(customDetail),
		}
	}

	testCases := []struct {
		name                        string
		request                     *corev1.AddPackageRepositoryRequest
		expectedResponse            *corev1.AddPackageRepositoryResponse
		expectedRepo                *appRepov1alpha1.AppRepository
		errorCode                   connect.Code
		existingAuthSecret          *apiv1.Secret
		existingDockerSecret        *apiv1.Secret
		expectedAuthCreatedSecret   *apiv1.Secret
		expectedDockerCreatedSecret *apiv1.Secret
		userManagedSecrets          bool
		testRepoServer              *httptest.Server
		expectedGlobalSecret        *apiv1.Secret
	}{
		{
			name:      "returns error if no namespace is provided",
			request:   &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{}},
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if no name is provided",
			request:   &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Namespace: "foo"}},
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if wrong repository type",
			request:   addRepoReqWrongType,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:      "returns error if no url",
			request:   addRepoReqNoUrl,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:             "check that interval is used",
			request:          addRepoReqWithInterval,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoWithInterval,
		},
		{
			name:             "simple add package repository scenario (HELM)",
			request:          addRepoReqSimple("helm"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleHelm,
		},
		{
			name:             "simple add package repository scenario (OCI)",
			request:          addRepoReqSimple("oci"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleOci,
		},
		{
			name:             "add package global repository",
			request:          addRepoReqGlobal,
			expectedResponse: addRepoExpectedGlobalResp,
			expectedRepo:     &addRepoGlobal,
		},
		// CUSTOM CA AUTH
		{
			name:                      "package repository with tls cert authority",
			request:                   addRepoReqTLSCA(ca),
			expectedResponse:          addRepoExpectedResp,
			expectedRepo:              &addRepoWithTLSCA,
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar", newTlsSecret("apprepo-bar", "foo", nil, nil, ca))),
			expectedGlobalSecret:      newTlsSecret("foo-apprepo-bar", kubeappsNamespace, nil, nil, ca), // Global secrets must reside in the Kubeapps (asset syncer) namespace
		},
		{
			name:                 "package repository with secret key reference",
			request:              addRepoReqTLSSecretRef,
			userManagedSecrets:   true,
			existingAuthSecret:   newTlsSecret("secret-1", "foo", nil, nil, ca),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         &addRepoTLSSecret,
			expectedGlobalSecret: newTlsSecret("foo-apprepo-bar", kubeappsNamespace, nil, nil, ca),
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            addRepoReqTLSSecretRef,
			userManagedSecrets: true,
			errorCode:          connect.CodeNotFound,
		},
		// BASIC AUTH
		{
			name:                      "[kubeapps managed secrets] package repository with basic auth and pass_credentials flag",
			request:                   addRepoReqBasicAuth("baz", "zot"),
			expectedResponse:          addRepoExpectedResp,
			expectedRepo:              addRepoAuthHeaderPassCredentials("foo"),
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", "foo", "baz", "zot"))),
			expectedGlobalSecret:      newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz", "zot"),
		},
		{
			name: "[kubeapps managed secrets] package repository with basic auth and pass_credentials flag in global namespace copies secret to kubeapps ns",
			request: &corev1.AddPackageRepositoryRequest{
				Name:            "bar",
				Context:         &corev1.Context{Cluster: KubeappsCluster, Namespace: globalPackagingNamespace},
				Type:            "helm",
				Url:             "http://example.com",
				NamespaceScoped: false,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: "the-user",
							Password: "the-pwd",
						},
					},
					PassCredentials: true,
				},
			},
			expectedResponse:          addRepoExpectedGlobalResp,
			expectedRepo:              addRepoAuthHeaderPassCredentials(globalPackagingNamespace),
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", globalPackagingNamespace, "the-user", "the-pwd"))),
			expectedGlobalSecret:      newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "the-user", "the-pwd"),
		},
		{
			name:      "package repository with wrong basic auth",
			request:   addRepoReqWrongBasicAuth,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:                 "[user managed secrets] package repository basic auth with existing secret",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "foo", "secret-basic"),
			userManagedSecrets:   true,
			existingAuthSecret:   newBasicAuthSecret("secret-basic", "foo", "baz-user", "zot-pwd"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-basic"),
			expectedGlobalSecret: newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz-user", "zot-pwd"),
		},
		{
			name: "[user managed secrets] add repository to global namespace creates secret in kubeapps namespace for syncer",
			request: &corev1.AddPackageRepositoryRequest{
				Name:            "bar",
				Context:         &corev1.Context{Namespace: globalPackagingNamespace, Cluster: KubeappsCluster},
				Type:            "helm",
				Url:             "http://example.com",
				NamespaceScoped: false,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
						SecretRef: &corev1.SecretKeyReference{
							Name: "secret-basic",
						},
					},
				},
			},
			userManagedSecrets:   true,
			existingAuthSecret:   newBasicAuthSecret("secret-basic", globalPackagingNamespace, "baz-user", "zot-pwd"),
			expectedResponse:     addRepoExpectedGlobalResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef(globalPackagingNamespace, "secret-basic"),
			expectedGlobalSecret: newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "baz-user", "zot-pwd"),
		},
		// BEARER TOKEN
		{
			name:                      "package repository with bearer token",
			request:                   addRepoReqBearerToken("the-token"),
			expectedResponse:          addRepoExpectedResp,
			expectedRepo:              addRepoAuthHeaderWithSecretRef("foo", "apprepo-bar"),
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar", newBearerAuthSecret("apprepo-bar", "foo", "the-token"))),
			expectedGlobalSecret:      newBearerAuthSecret("foo-apprepo-bar", kubeappsNamespace, "the-token"),
		},
		{
			name:      "package repository with no bearer token",
			request:   addRepoReqBearerToken(""),
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:                 "package repository bearer token with secret (user managed secrets)",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "foo", "secret-bearer"),
			userManagedSecrets:   true,
			existingAuthSecret:   newBearerAuthSecret("secret-bearer", "foo", "the-token"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-bearer"),
			expectedGlobalSecret: newBearerAuthSecret("foo-apprepo-bar", kubeappsNamespace, "the-token"),
		},
		// CUSTOM AUTH
		{
			name:                      "package repository with custom auth",
			request:                   addRepoReqCustomAuth,
			expectedResponse:          addRepoExpectedResp,
			expectedRepo:              addRepoAuthHeaderWithSecretRef("foo", "apprepo-bar"),
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar", newHeaderAuthSecret("apprepo-bar", "foo", "foobarzot"))),
			expectedGlobalSecret:      newHeaderAuthSecret("foo-apprepo-bar", kubeappsNamespace, "foobarzot"),
		},
		{
			name:                 "package repository custom auth with existing secret (user managed secrets)",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER, "foo", "secret-custom"),
			userManagedSecrets:   true,
			existingAuthSecret:   newBasicAuthSecret("secret-custom", "foo", "baz", "zot"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-custom"),
			expectedGlobalSecret: newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz", "zot"),
		},
		{
			name: "global package repository custom auth with existing secret does not generate copied global secret (user managed secrets)",
			request: &corev1.AddPackageRepositoryRequest{
				Name:            "bar",
				Context:         &corev1.Context{Namespace: globalPackagingNamespace, Cluster: KubeappsCluster},
				Type:            "helm",
				Url:             "http://example.com",
				NamespaceScoped: false,
				Auth: &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
						SecretRef: &corev1.SecretKeyReference{
							Name: "secret-custom",
						},
					},
				},
			},
			userManagedSecrets:   true,
			existingAuthSecret:   newBasicAuthSecret("secret-custom", globalPackagingNamespace, "baz", "zot"),
			expectedResponse:     addRepoExpectedGlobalResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef(globalPackagingNamespace, "secret-custom"),
			expectedGlobalSecret: newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "baz", "zot"),
		},
		// DOCKER AUTH
		{
			name: "package repository with Docker auth",
			request: addRepoReqDockerAuth(&corev1.DockerCredentials{
				Server:   "https://docker-server",
				Username: "the-user",
				Password: "the-password",
				Email:    "foo@bar.com",
			}),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     addRepoAuthDocker("apprepo-bar"),
			expectedAuthCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar",
				newAuthDockerSecret("apprepo-bar", "foo",
					dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")))),
			expectedGlobalSecret: newAuthDockerSecret("foo-apprepo-bar", kubeappsNamespace,
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
		},
		{
			name:               "package repository with Docker auth (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON, "foo", "secret-docker"),
			userManagedSecrets: true,
			existingAuthSecret: newAuthDockerSecret("secret-docker", "foo",
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     addRepoAuthDocker("secret-docker"),
			expectedGlobalSecret: newAuthDockerSecret("foo-apprepo-bar", kubeappsNamespace,
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
		},
		// Others
		{
			name:      "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:   addRepoReqTLSDifferentSecretAuth,
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            addRepoReqTLSDifferentSecretAuth,
			errorCode:          connect.CodeInvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          addRepoReqOnlyPassCredentials,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoOnlyPassCredentials,
		},
		{
			name:               "package repository with reference to malformed secret",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "foo", "secret-basic"),
			existingAuthSecret: newTlsSecret("secret-basic", "foo", nil, nil, nil), // Creates empty secret
			userManagedSecrets: true,
			errorCode:          connect.CodeInternal,
		},
		// Custom values
		{
			name:             "package repository with custom values",
			request:          addRepoReqCustomValues,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailHelm,
		},
		{
			name:             "package repository with invalid custom values",
			request:          addRepoReqWrongCustomValues,
			expectedResponse: addRepoExpectedResp,
			errorCode:        connect.CodeInvalidArgument,
		},
		{
			name:             "package repository with validation success (Helm)",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailHelm,
			testRepoServer:   newFakeRepoServer(t, map[string]*http.Response{"/index.yaml": {StatusCode: 200}}),
		},
		{
			name:             "package repository with validation success (OCI)",
			request:          addRepoReqCustomValuesOCIValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailOci,
			testRepoServer: newFakeRepoServer(t, map[string]*http.Response{
				"/v2/repo1/tags/list":      httpResponse(200, "{ \"name\":\"repo1\", \"tags\":[\"tag1\"] }"),
				"/v2/repo1/manifests/tag1": httpResponse(200, "{ \"config\":{ \"mediaType\":\"application/vnd.cncf.helm.config\" } }"),
			}),
		},
		{
			name:             "package repository with validation failing",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			testRepoServer: newFakeRepoServer(t,
				map[string]*http.Response{
					"/index.yaml": httpResponse(404, "It failed because of X and Y"),
				}),
			errorCode: connect.CodeFailedPrecondition,
		},
		{
			name: "[user managed secrets] create repository is ok with existing imagePullSecrets",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
						SecretRef: "secret-docker",
					},
				},
				OciRepositories: []string{"oci-repo-1", "oci-repo-2"},
			}),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			existingDockerSecret: newAuthDockerSecret("secret-docker", "foo",
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			expectedRepo: &appRepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       AppRepositoryKind,
					APIVersion: AppRepositoryApi,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "foo",
					ResourceVersion: "1",
				},
				Spec: appRepov1alpha1.AppRepositorySpec{
					URL:                   "https://example.com",
					Type:                  "helm",
					OCIRepositories:       []string{"oci-repo-1", "oci-repo-2"},
					DockerRegistrySecrets: []string{"secret-docker"},
				},
			},
		},
		{
			name: "[user managed secrets] create repository fails with non existing images pull secret",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
						SecretRef: "secret-docker",
					},
				},
				OciRepositories: []string{"oci-repo-1", "oci-repo-2"},
			}),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			errorCode:          connect.CodeNotFound,
		},
		{
			name: "[user managed secrets] create repository fails with existing images pull secret having wrong type",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
						SecretRef: "secret-docker",
					},
				},
			}),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			existingAuthSecret: newHeaderAuthSecret("secret-docker", "foo", ""),
			errorCode:          connect.CodeInvalidArgument,
		},
		{
			name: "[user managed secrets] create repository fails with existing images pull secret having wrong key",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
						SecretRef: "secret-docker",
					},
				},
			}),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			existingDockerSecret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret-docker",
					Namespace: "foo",
				},
				Type: apiv1.SecretTypeDockerConfigJson,
				Data: map[string][]byte{
					"wrong-key": []byte(""),
				},
			},
			errorCode: connect.CodeInvalidArgument,
		},
		{
			name: "[kubeapps managed secrets] create repository is ok",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
						Credentials: &corev1.DockerCredentials{
							Server:   "https://myfooserver.com",
							Username: "username",
							Password: "password",
							Email:    "foo@bar.com",
						},
					},
				},
			}),
			expectedResponse: addRepoExpectedResp,
			expectedRepo: &appRepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       AppRepositoryKind,
					APIVersion: AppRepositoryApi,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "foo",
					ResourceVersion: "1",
				},
				Spec: appRepov1alpha1.AppRepositorySpec{
					URL:                   "https://example.com",
					Type:                  "helm",
					DockerRegistrySecrets: []string{"pullsecret-bar"},
				},
			},
			expectedDockerCreatedSecret: setSecretAnnotations(setSecretOwnerRef("bar",
				newAuthDockerSecret("pullsecret-bar", "foo",
					dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")))),
		},
		{
			name: "[kubeapps managed secrets] create repository fails with existing imagePullSecret",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
						Credentials: &corev1.DockerCredentials{
							Server:   "https://myfooserver.com",
							Username: "username",
							Password: "password",
							Email:    "foo@bar.com",
						},
					},
				},
				OciRepositories: []string{"oci-repo-1", "oci-repo-2"},
			}),
			expectedResponse: addRepoExpectedResp,
			existingDockerSecret: newAuthDockerSecret("pullsecret-bar", "foo",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			errorCode: connect.CodeAlreadyExists,
		},
		{
			name: "[custom details] repository is created with proxy options",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ProxyOptions: &v1alpha1.ProxyOptions{
					Enabled:    true,
					HttpProxy:  "http://proxy.com",
					HttpsProxy: "https://proxy.com",
					NoProxy:    "no-proxy.com",
				},
			}),
			expectedResponse: addRepoExpectedResp,
			expectedRepo: &appRepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       AppRepositoryKind,
					APIVersion: AppRepositoryApi,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "foo",
					ResourceVersion: "1",
				},
				Spec: appRepov1alpha1.AppRepositorySpec{
					URL:  "https://example.com",
					Type: "helm",
					SyncJobPodTemplate: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							Containers: []apiv1.Container{
								{
									Env: []apiv1.EnvVar{
										{
											Name:  "HTTP_PROXY",
											Value: "http://proxy.com",
										},
										{
											Name:  "HTTPS_PROXY",
											Value: "https://proxy.com",
										},
										{
											Name:  "NO_PROXY",
											Value: "no-proxy.com",
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "[custom details] repository is created with pod template options",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				NodeSelector: map[string]string{
					"node": "node-1",
				},
				Tolerations: []*v1alpha1.Toleration{
					{
						Key:               Ptr("key"),
						Value:             Ptr("value"),
						Effect:            Ptr("NoSchedule"),
						Operator:          Ptr("Equal"),
						TolerationSeconds: Ptr(int64(3600)),
					},
				},
				SecurityContext: &v1alpha1.PodSecurityContext{
					RunAsUser:          Ptr(int64(1001)),
					RunAsGroup:         Ptr(int64(1002)),
					FSGroup:            Ptr(int64(1003)),
					RunAsNonRoot:       Ptr(true),
					SupplementalGroups: []int64{1004},
				},
			}),
			expectedResponse: addRepoExpectedResp,
			expectedRepo: &appRepov1alpha1.AppRepository{
				TypeMeta: metav1.TypeMeta{
					Kind:       AppRepositoryKind,
					APIVersion: AppRepositoryApi,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:            "bar",
					Namespace:       "foo",
					ResourceVersion: "1",
				},
				Spec: appRepov1alpha1.AppRepositorySpec{
					URL:  "https://example.com",
					Type: "helm",
					SyncJobPodTemplate: apiv1.PodTemplateSpec{
						Spec: apiv1.PodSpec{
							NodeSelector: map[string]string{
								"node": "node-1",
							},
							Tolerations: []apiv1.Toleration{
								{
									Key:               "key",
									Value:             "value",
									Effect:            "NoSchedule",
									Operator:          "Equal",
									TolerationSeconds: &[]int64{3600}[0],
								},
							},
							SecurityContext: &apiv1.PodSecurityContext{
								RunAsUser:          &[]int64{1001}[0],
								RunAsGroup:         &[]int64{1002}[0],
								FSGroup:            &[]int64{1003}[0],
								RunAsNonRoot:       &[]bool{true}[0],
								SupplementalGroups: []int64{1004},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var secrets []k8sruntime.Object
			if tc.existingAuthSecret != nil {
				secrets = append(secrets, tc.existingAuthSecret)
			}
			if tc.existingDockerSecret != nil {
				secrets = append(secrets, tc.existingDockerSecret)
			}
			s := newServerWithSecretsAndRepos(t, secrets, nil)
			if tc.testRepoServer != nil {
				defer tc.testRepoServer.Close()
				s.repoClientGetter = func(_ *appRepov1.AppRepository, _ *apiv1.Secret) (*http.Client, error) {
					return tc.testRepoServer.Client(), nil
				}
				tc.request.Url = tc.testRepoServer.URL
				if tc.expectedRepo != nil {
					tc.expectedRepo.Spec.URL = tc.testRepoServer.URL
				}
			}

			nsname := types.NamespacedName{Namespace: tc.request.Context.Namespace, Name: tc.request.Name}
			ctx := context.Background()
			response, err := s.AddPackageRepository(ctx, connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.errorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the expectedResponse for OK status.
			if tc.errorCode == 0 {
				if response == nil {
					t.Fatalf("got: nil, want: expectedResponse")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.AddPackageRepositoryResponse{},
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
					)
					if got, want := response.Msg, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// purposefully not calling mock.ExpectationsWereMet() here because
			// AddPackageRepository will trigger an ADD event that will be processed
			// asynchronously, so it may or may not have enough time to get to the
			// point where the cache worker does a GET

			// We don't need to check anything else for non-OK codes.
			if tc.errorCode != 0 {
				return
			}

			// check expected HelmRelease CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(http.Header{}, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				var actualRepo appRepov1alpha1.AppRepository
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				} else {
					checkRepoSecrets(s, t, tc.userManagedSecrets, &actualRepo, tc.expectedRepo, tc.expectedAuthCreatedSecret, tc.expectedDockerCreatedSecret, tc.expectedGlobalSecret)
				}
			}
		})
	}
}

func TestGetPackageRepositorySummaries(t *testing.T) {
	// some prep
	indexYAMLBytes, err := os.ReadFile(testYaml("valid-index.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(indexYAMLBytes))
	}))
	defer ts.Close()

	repo1b := *repo1
	repo1b.Spec.URL = ts.URL

	repo2b := *repo2
	repo2b.Spec.URL = ts.URL

	repo3 := *repo3
	repo3.Spec.URL = ts.URL

	repo1Summary.Url = ts.URL
	repo2Summary.Url = ts.URL
	repo3Summary.Url = ts.URL

	testCases := []struct {
		name               string
		request            *corev1.GetPackageRepositorySummariesRequest
		existingRepos      []k8sruntime.Object
		existingNamespaces []*apiv1.Namespace
		expectedErrorCode  connect.Code
		expectedResponse   *corev1.GetPackageRepositorySummariesResponse
		reactors           []*ClientReaction
	}{
		{
			name: "returns all package summaries when namespace not specified",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster},
			},
			existingRepos: []k8sruntime.Object{
				&repo1b,
				&repo2b,
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					repo1Summary,
					repo2Summary,
				},
			},
		},
		{
			name: "returns actual accessible package summaries when namespace not specified and no cluster level access",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster},
			},
			existingNamespaces: []*apiv1.Namespace{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ns-accessible",
					},
					Status: apiv1.NamespaceStatus{
						Phase: apiv1.NamespaceActive,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "ns-inaccessible",
					},
					Status: apiv1.NamespaceStatus{
						Phase: apiv1.NamespaceActive,
					},
				},
			},
			reactors: []*ClientReaction{
				{
					verb:     "list",
					resource: "apprepositories",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						switch action.GetNamespace() {
						// Forbidden cluster-wide listing and a specific namespace
						case "":
							return true, nil, k8sErrors.NewForbidden(authorizationv1.Resource("AppRepository"), "", errors.New("bang"))
						case "ns-inaccessible":
							return true, nil, k8sErrors.NewForbidden(authorizationv1.Resource("AppRepository"), "", errors.New("bang"))
						case "ns-accessible":
							return true, &appRepov1alpha1.AppRepositoryList{
								Items: []appRepov1alpha1.AppRepository{
									{
										TypeMeta: metav1.TypeMeta{
											APIVersion: appReposAPIVersion,
											Kind:       AppRepositoryKind,
										},
										ObjectMeta: metav1.ObjectMeta{
											Name:            "repo-accessible-1",
											Namespace:       "ns-accessible",
											ResourceVersion: "1",
										},
										Spec: appRepov1alpha1.AppRepositorySpec{
											URL:         "https://test-repo",
											Type:        "helm",
											Description: "description 1",
										},
									},
								},
							}, nil
						default:
							return true, &appRepov1alpha1.AppRepositoryList{}, nil
						}
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					{
						PackageRepoRef:  repoRef("repo-accessible-1", KubeappsCluster, "ns-accessible"),
						Name:            "repo-accessible-1",
						Description:     "description 1",
						NamespaceScoped: true,
						Type:            "helm",
						Url:             "https://test-repo",
						RequiresAuth:    false,
						Status:          &corev1.PackageRepositoryStatus{Ready: true},
					},
				},
			},
		},
		{
			name: "returns package summaries when namespace is specified",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster, Namespace: "ns-2"},
			},
			existingRepos: []k8sruntime.Object{
				&repo1b,
				&repo2b,
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					repo2Summary,
				},
			},
		},
		{
			name: "returns package summaries with auth",
			request: &corev1.GetPackageRepositorySummariesRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster, Namespace: globalPackagingNamespace},
			},
			existingRepos: []k8sruntime.Object{
				&repo3,
			},
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					repo3Summary,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			if tc.existingRepos != nil {
				for _, repo := range tc.existingRepos {
					unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(repo)
					unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
				}
			}

			var typedObjects []k8sruntime.Object
			if tc.existingNamespaces != nil {
				for _, ns := range tc.existingNamespaces {
					typedObjects = append(typedObjects, ns)
				}
			}

			s := newServerWithAppRepoReactors(unstructuredObjects, nil, typedObjects, nil, tc.reactors)

			response, err := s.GetPackageRepositorySummaries(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedErrorCode != 0 {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositorySummariesResponse{},
				corev1.PackageRepositorySummary{},
				corev1.PackageRepositoryReference{},
				corev1.PackageRepositoryStatus{},
				corev1.PackageRepositoryAuth{},
			)
			opts2 := cmpopts.SortSlices(lessPackageRepositorySummaryFunc)
			if got, want := response.Msg, tc.expectedResponse; !cmp.Equal(want, got, opts, opts2) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts, opts2))
			}
		})
	}
}

func TestGetPackageRepositoryDetail(t *testing.T) {
	buildRequest := func(namespace, name string) *corev1.GetPackageRepositoryDetailRequest {
		return &corev1.GetPackageRepositoryDetailRequest{
			PackageRepoRef: &corev1.PackageRepositoryReference{
				Plugin:     plugin,
				Context:    &corev1.Context{Namespace: namespace, Cluster: KubeappsCluster},
				Identifier: name,
			},
		}
	}
	buildResponse := func(namespace, name, repoType, url, description string,
		auth *corev1.PackageRepositoryAuth, tlsConfig *corev1.PackageRepositoryTlsConfig,
		customDetail *v1alpha1.HelmPackageRepositoryCustomDetail) *corev1.GetPackageRepositoryDetailResponse {
		response := &corev1.GetPackageRepositoryDetailResponse{
			Detail: &corev1.PackageRepositoryDetail{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Plugin:     plugin,
					Context:    &corev1.Context{Namespace: namespace, Cluster: KubeappsCluster},
					Identifier: name,
				},
				Name:            name,
				Description:     description,
				NamespaceScoped: namespace != globalPackagingNamespace,
				Type:            repoType,
				Url:             url,
				Auth:            auth,
				TlsConfig:       tlsConfig,
				Status:          &corev1.PackageRepositoryStatus{Ready: true},
			},
		}
		if customDetail != nil {
			response.Detail.CustomDetail = toProtoBufAny(customDetail)
		}
		return response
	}

	ca, _, _ := getCertsForTesting(t)

	testCases := []struct {
		name                 string
		request              *corev1.GetPackageRepositoryDetailRequest
		repositoryCustomizer func(repository *appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository
		expectedResponse     *corev1.GetPackageRepositoryDetailResponse
		expectedErrorCode    connect.Code
		existingSecret       *apiv1.Secret
	}{
		{
			name:              "not found",
			request:           buildRequest("ns-1", "foo"),
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name:    "check ref",
			request: buildRequest("ns-1", "repo-1"),
			expectedResponse: buildResponse("ns-1", "repo-1", "helm",
				"https://test-repo", "description 1", nil, nil, nil),
		},
		{
			name:    "check values with auth",
			request: buildRequest(globalPackagingNamespace, "repo-3"),
			expectedResponse: buildResponse(globalPackagingNamespace, "repo-3", "helm", "https://test-repo3", "description 3",
				&corev1.PackageRepositoryAuth{
					PassCredentials: false,
					Type:            corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: "REDACTED",
							Password: "REDACTED",
						},
					},
				},
				nil,
				nil),
			existingSecret: newBasicAuthSecret(helm.SecretNameForRepo("repo-3"), globalPackagingNamespace, "baz-user", "zot-pwd"),
		},
		{
			name:    "check values without auth",
			request: buildRequest("ns-4", "repo-4"),
			expectedResponse: buildResponse("ns-4", "repo-4", "oci", "https://test-repo4", "description 4",
				&corev1.PackageRepositoryAuth{
					PassCredentials: true,
				},
				&corev1.PackageRepositoryTlsConfig{
					InsecureSkipVerify: true,
					PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
						CertAuthority: "REDACTED",
					},
				},
				&v1alpha1.HelmPackageRepositoryCustomDetail{
					OciRepositories: []string{"oci-repo-1", "oci-repo-2"},
					FilterRule: &v1alpha1.RepositoryFilterRule{
						Jq:        "jq-filter-$1",
						Variables: map[string]string{"$1": "value1"},
					},
				}),
			existingSecret: newTlsSecret(helm.SecretNameForRepo("repo-4"), "ns-4", nil, nil, ca),
		},
		{
			name:    "check values with imagesPullSecret",
			request: buildRequest("ns-5", "repo-5"),
			expectedResponse: buildResponse("ns-5", "repo-5", "helm", "https://test-repo5", "description 5",
				nil, nil,
				&v1alpha1.HelmPackageRepositoryCustomDetail{
					ImagesPullSecret: &v1alpha1.ImagesPullSecret{
						DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
							Credentials: &corev1.DockerCredentials{
								Server:   "REDACTED",
								Username: "REDACTED",
								Password: "REDACTED",
								Email:    "REDACTED",
							},
						},
					},
				}),
			existingSecret: newAuthDockerSecret(imagesPullSecretName("repo-5"), "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var secrets []k8sruntime.Object
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}

			s := newServerWithSecretsAndRepos(t, secrets, []*appRepov1alpha1.AppRepository{repo1, repo2, repo3, repo4, repo5})

			actualResponse, err := s.GetPackageRepositoryDetail(context.Background(), connect.NewRequest(tc.request))

			// checks
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != 0 {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositoryDetailResponse{},
				corev1.PackageRepositoryDetail{},
				corev1.GetPackageRepositorySummariesResponse{},
				corev1.PackageRepositorySummary{},
				corev1.PackageRepositoryReference{},
				corev1.PackageRepositoryStatus{},
				corev1.PackageRepositoryTlsConfig{},
				corev1.PackageRepositoryAuth{},
				v1alpha1.HelmPackageRepositoryCustomDetail{},
				anypb.Any{},
				corev1.PackageRepositoryAuth_UsernamePassword{},
				corev1.UsernamePassword{},
			)

			if got, want := tc.request.PackageRepoRef, actualResponse.Msg.Detail.PackageRepoRef; !cmp.Equal(want, got, opts) {
				t.Errorf("ref mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			if got, want := actualResponse.Msg, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestUpdatePackageRepository(t *testing.T) {
	defaultRef := &corev1.PackageRepositoryReference{
		Plugin:     &pluginDetail,
		Context:    &corev1.Context{Namespace: "ns-1", Cluster: KubeappsCluster},
		Identifier: "repo-1",
	}
	commonRequest := func() *corev1.UpdatePackageRepositoryRequest {
		return &corev1.UpdatePackageRepositoryRequest{
			PackageRepoRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-1", Cluster: KubeappsCluster},
				Identifier: "repo-1",
			},
			Url: "https://new-repo-url",
		}
	}

	ca, pub, _ := getCertsForTesting(t)

	testCases := []struct {
		name                   string
		requestCustomizer      func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest
		expectedRepoCustomizer func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository
		expectedErrorCode      connect.Code
		expectedRef            *corev1.PackageRepositoryReference
		existingAuthSecret     *apiv1.Secret
		existingDockerSecret   *apiv1.Secret
		expectedAuthSecret     *apiv1.Secret
		expectedDockerSecret   *apiv1.Secret
		userManagedSecrets     bool
		expectedGlobalSecret   *apiv1.Secret
	}{
		{
			name: "invalid package repo ref",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = nil
				return request
			},
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name: "repository not found",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef.Context = &corev1.Context{Cluster: "other", Namespace: globalPackagingNamespace}
				return request
			},
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name: "validate name",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef.Identifier = ""
				return request
			},
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name: "validate url",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Url = ""
				return request
			},
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name: "check that interval is used",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Interval = "15m"
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Spec.Interval = "15m"
				// other changes inherent to how the customizer works
				repository.ResourceVersion = "2"
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},

			expectedRef: defaultRef,
		},
		{
			name: "update the proxy options",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					ProxyOptions: &v1alpha1.ProxyOptions{
						Enabled:    true,
						HttpProxy:  "http://proxy",
						HttpsProxy: "https://proxy",
						NoProxy:    "no-proxy",
					},
				})
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Spec.SyncJobPodTemplate.Spec.Containers = []apiv1.Container{
					{
						Env: []apiv1.EnvVar{
							{Name: "HTTP_PROXY", Value: "http://proxy"},
							{Name: "HTTPS_PROXY", Value: "https://proxy"},
							{Name: "NO_PROXY", Value: "no-proxy"},
						},
					},
				}
				// other changes inherent to how the customizer works
				repository.ResourceVersion = "2"
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},

			expectedRef: defaultRef,
		},
		{
			name: "update the pod template options",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					NodeSelector: map[string]string{
						"node": "node-1",
					},
					Tolerations: []*v1alpha1.Toleration{
						{
							Key:               Ptr("key"),
							Value:             Ptr("value"),
							Effect:            Ptr("NoSchedule"),
							Operator:          Ptr("Equal"),
							TolerationSeconds: Ptr(int64(3600)),
						},
					},
					SecurityContext: &v1alpha1.PodSecurityContext{
						RunAsUser:          Ptr(int64(1001)),
						RunAsGroup:         Ptr(int64(1002)),
						FSGroup:            Ptr(int64(1003)),
						RunAsNonRoot:       Ptr(true),
						SupplementalGroups: []int64{1004},
					},
				})
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Spec.SyncJobPodTemplate.Spec.NodeSelector = map[string]string{
					"node": "node-1",
				}
				repository.Spec.SyncJobPodTemplate.Spec.Tolerations = []apiv1.Toleration{
					{
						Key:               "key",
						Value:             "value",
						Effect:            "NoSchedule",
						Operator:          "Equal",
						TolerationSeconds: &[]int64{3600}[0],
					},
				}
				repository.Spec.SyncJobPodTemplate.Spec.SecurityContext = &apiv1.PodSecurityContext{
					RunAsUser:          &[]int64{1001}[0],
					RunAsGroup:         &[]int64{1002}[0],
					FSGroup:            &[]int64{1003}[0],
					RunAsNonRoot:       &[]bool{true}[0],
					SupplementalGroups: []int64{1004},
				}
				// other changes inherent to how the customizer works
				repository.ResourceVersion = "2"
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef: defaultRef,
		},
		{
			name: "update tls config",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.TlsConfig = &corev1.PackageRepositoryTlsConfig{
					PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
						CertAuthority: string(ca),
					},
				}
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.Auth = appRepov1alpha1.AppRepositoryAuth{
					CustomCA: &appRepov1alpha1.AppRepositoryCustomCA{
						SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: "apprepo-repo-1"}, Key: "ca.crt"},
					},
				}
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef:          defaultRef,
			expectedAuthSecret:   setSecretAnnotations(setSecretOwnerRef("repo-1", newTlsSecret("apprepo-repo-1", "ns-1", nil, nil, ca))),
			expectedGlobalSecret: newTlsSecret("ns-1-apprepo-repo-1", kubeappsNamespace, nil, nil, ca),
		},
		{
			name:               "update removing tsl config",
			existingAuthSecret: newTlsSecret(helm.SecretNameForRepo("repo-4"), "ns-4", nil, nil, ca),
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-4", Cluster: KubeappsCluster},
					Identifier: "repo-4",
				}
				request.TlsConfig = nil
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Name = "repo-4"
				repository.Namespace = "ns-4"
				repository.ResourceVersion = "2"
				repository.Spec.Type = "oci"
				repository.Spec.Auth = appRepov1alpha1.AppRepositoryAuth{}
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-4", Cluster: KubeappsCluster},
				Identifier: "repo-4",
			},
		},
		{
			name: "update adding auth",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
						Header: "foobarzot",
					},
				}
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.Auth = appRepov1alpha1.AppRepositoryAuth{
					Header: &appRepov1alpha1.AppRepositoryAuthHeader{
						SecretKeyRef: apiv1.SecretKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{Name: "apprepo-repo-1"},
							Key:                  "authorizationHeader",
						},
					},
				}
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef:          defaultRef,
			expectedAuthSecret:   setSecretAnnotations(setSecretOwnerRef("repo-1", newBearerAuthSecret("apprepo-repo-1", "ns-1", "foobarzot"))),
			expectedGlobalSecret: newBearerAuthSecret("ns-1-apprepo-repo-1", kubeappsNamespace, "foobarzot"),
		},
		{
			name: "update adding auth (user managed secrets)",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_SecretRef{
						SecretRef: &corev1.SecretKeyReference{
							Name: "my-own-secret",
						},
					},
				}
				return request
			},
			userManagedSecrets: true,
			existingAuthSecret: newBearerAuthSecret("my-own-secret", "ns-1", "foobarzot"),
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.Auth = appRepov1alpha1.AppRepositoryAuth{
					Header: &appRepov1alpha1.AppRepositoryAuthHeader{
						SecretKeyRef: apiv1.SecretKeySelector{
							LocalObjectReference: apiv1.LocalObjectReference{Name: "my-own-secret"},
							Key:                  "authorizationHeader",
						},
					},
				}
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef:          defaultRef,
			expectedGlobalSecret: newBearerAuthSecret("ns-1-apprepo-repo-1", kubeappsNamespace, "foobarzot"),
		},
		{
			name:               "update removing auth",
			existingAuthSecret: newBearerAuthSecret(helm.SecretNameForRepo("repo-3"), globalPackagingNamespace, "token-value"),
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: globalPackagingNamespace, Cluster: KubeappsCluster},
					Identifier: "repo-3",
				}
				request.Auth = nil
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Name = "repo-3"
				repository.Namespace = globalPackagingNamespace
				repository.ResourceVersion = "2"
				repository.Spec.Auth = appRepov1alpha1.AppRepositoryAuth{}
				repository.Spec.Description = ""
				repository.Spec.URL = "https://new-repo-url"
				return &repository
			},
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: globalPackagingNamespace, Cluster: KubeappsCluster},
				Identifier: "repo-3",
			},
		},
		{
			name: "updated with new url",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Url = "foo"
				request.Description = "description 1b"
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.URL = "foo"
				repository.Spec.Description = "description 1b"
				return &repository
			},
			expectedRef: defaultRef,
		},
		{
			name: "updated with custom details",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Url = "foo"
				request.Description = "description 1b"
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.URL = "foo"
				repository.Spec.Description = "description 1b"
				return &repository
			},
			expectedRef: defaultRef,
		},
		{
			name: "[kubeapps managed secrets] update repo with image pull secret credentials",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Description = "description"
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					ImagesPullSecret: &v1alpha1.ImagesPullSecret{
						DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
							Credentials: &corev1.DockerCredentials{
								Server:   "https://myfooserver.com",
								Username: "username",
								Password: "password",
								Email:    "foo@bar.com",
							},
						},
					},
				})
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.URL = "https://new-repo-url"
				repository.Spec.Description = "description"
				repository.Spec.DockerRegistrySecrets = []string{"pullsecret-repo-1"}
				return &repository
			},
			expectedDockerSecret: setSecretAnnotations(setSecretOwnerRef("repo-1",
				newAuthDockerSecret("pullsecret-repo-1", "ns-1",
					dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")))),
			expectedRef: defaultRef,
		},
		{
			name: "[kubeapps managed secrets] update repo with image pull secret redacted",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
					Identifier: "repo-5",
				}
				request.Description = "description"
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					ImagesPullSecret: &v1alpha1.ImagesPullSecret{
						DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
							Credentials: &corev1.DockerCredentials{
								Server:   "REDACTED",
								Username: "REDACTED",
								Password: "REDACTED",
								Email:    "REDACTED",
							},
						},
					},
				})
				return request
			},
			existingDockerSecret: newAuthDockerSecret(imagesPullSecretName("repo-5"), "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.Name = "repo-5"
				repository.Namespace = "ns-5"
				repository.ResourceVersion = "2"
				repository.Spec.URL = "https://new-repo-url"
				repository.Spec.Description = "description"
				repository.Spec.DockerRegistrySecrets = []string{imagesPullSecretName("repo-5")}
				return &repository
			},
			expectedDockerSecret: newAuthDockerSecret(imagesPullSecretName("repo-5"), "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
				Identifier: "repo-5",
			},
		},
		{
			name: "[user managed secrets] update repo with image pull secret ref",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Description = "description"
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					ImagesPullSecret: &v1alpha1.ImagesPullSecret{
						DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
							SecretRef: "test-pull-secret",
						},
					},
				})
				return request
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.URL = "https://new-repo-url"
				repository.Spec.Description = "description"
				repository.Spec.DockerRegistrySecrets = []string{"test-pull-secret"}
				return &repository
			},
			userManagedSecrets: true,
			existingDockerSecret: newAuthDockerSecret("test-pull-secret", "ns-1",
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			expectedRef: defaultRef,
		},
		{
			name: "[kubeapps managed secrets] update repo removing images pull secret",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
					Identifier: "repo-5",
				}
				request.Description = "description"
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{})
				return request
			},
			existingDockerSecret: newAuthDockerSecret("pullsecret-repo-5", "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = "ns-5"
				repository.Name = "repo-5"
				repository.Spec.URL = "https://new-repo-url"
				repository.Spec.Description = "description"
				repository.Spec.DockerRegistrySecrets = nil
				return &repository
			},
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
				Identifier: "repo-5",
			},
		},
		{
			name: "[kubeapps managed secrets] update repo auth basic to token",
			existingAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-7",
				newBasicAuthSecret("apprepo-repo-7", "ns-7", "foo", "bar"))),

			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-7", Cluster: KubeappsCluster},
					Identifier: "repo-7",
				}
				request.Url = repo7.Spec.URL
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_Header{
						Header: "zot",
					},
				}
				return request
			},

			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = "ns-7"
				repository.Name = "repo-7"
				repository.Spec = repo7.Spec
				return &repository
			},
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-7", Cluster: KubeappsCluster},
				Identifier: "repo-7",
			},
			expectedAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-7",
				newBearerAuthSecret("apprepo-repo-7", "ns-7", "zot"))),
			expectedGlobalSecret: newBearerAuthSecret("ns-7-apprepo-repo-7", kubeappsNamespace, "zot"),
		},
		{
			name: "[kubeapps managed secrets] update repo auth basic to docker",
			existingAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-7",
				newBasicAuthSecret("apprepo-repo-7", "ns-7", "foo", "bar"))),

			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-7", Cluster: KubeappsCluster},
					Identifier: "repo-7",
				}
				request.Url = repo7.Spec.URL
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_DockerCreds{
						DockerCreds: &corev1.DockerCredentials{
							Server:   "sample.com",
							Username: "foo",
							Password: "bar",
							Email:    "user@sample.com",
						},
					},
				}
				return request
			},

			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = "ns-7"
				repository.Name = "repo-7"
				repository.Spec = repo7.Spec
				repository.Spec.Auth.Header.SecretKeyRef.Key = DockerConfigJsonKey
				return &repository
			},
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-7", Cluster: KubeappsCluster},
				Identifier: "repo-7",
			},
			expectedAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-7",
				newAuthDockerSecret("apprepo-repo-7", "ns-7",
					dockerAuthJson("sample.com", "foo", "bar", "user@sample.com", "Zm9vOmJhcg==")))),
			expectedGlobalSecret: newAuthDockerSecret("ns-7-apprepo-repo-7", kubeappsNamespace,
				dockerAuthJson("sample.com", "foo", "bar", "user@sample.com", "Zm9vOmJhcg==")),
		},
		{
			name: "[issue 5746] secret updates ignored if not all credentials are provided - auth updates",
			existingAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-6",
				addTlsToSecret(newBasicAuthSecret("apprepo-repo-6", "ns-6", "foo", "bar"), nil, nil, ca))),

			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-6", Cluster: KubeappsCluster},
					Identifier: "repo-6",
				}
				request.Url = repo6.Spec.URL
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: RedactedString,
							Password: "zot",
						},
					},
				}
				request.TlsConfig = &corev1.PackageRepositoryTlsConfig{
					PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
						CertAuthority: RedactedString,
					},
				}
				return request
			},

			expectedAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-6",
				addTlsToSecret(newBasicAuthSecret("apprepo-repo-6", "ns-6", "foo", "zot"), nil, nil, ca))),
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-6", Cluster: KubeappsCluster},
				Identifier: "repo-6",
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = repo6.Namespace
				repository.Name = repo6.Name
				repository.Spec = repo6.Spec
				return &repository
			},
			expectedGlobalSecret: addTlsToSecret(newBasicAuthSecret("ns-6-apprepo-repo-6", kubeappsNamespace, "foo", "zot"), nil, nil, ca),
		},
		{
			name: "[issue 5746] secret updates ignored if not all credentials are provided - tls updates",
			existingAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-6",
				addTlsToSecret(newBasicAuthSecret("apprepo-repo-6", "ns-6", "foo", "bar"), nil, nil, ca))),

			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-6", Cluster: KubeappsCluster},
					Identifier: "repo-6",
				}
				request.Url = repo6.Spec.URL
				request.Auth = &corev1.PackageRepositoryAuth{
					Type: corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH,
					PackageRepoAuthOneOf: &corev1.PackageRepositoryAuth_UsernamePassword{
						UsernamePassword: &corev1.UsernamePassword{
							Username: RedactedString,
							Password: RedactedString,
						},
					},
				}
				request.TlsConfig = &corev1.PackageRepositoryTlsConfig{
					PackageRepoTlsConfigOneOf: &corev1.PackageRepositoryTlsConfig_CertAuthority{
						CertAuthority: string(pub),
					},
				}
				return request
			},

			expectedAuthSecret: setSecretAnnotations(setSecretOwnerRef("repo-6",
				addTlsToSecret(newBasicAuthSecret("apprepo-repo-6", "ns-6", "foo", "bar"), nil, nil, pub))),
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-6", Cluster: KubeappsCluster},
				Identifier: "repo-6",
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = repo6.Namespace
				repository.Name = repo6.Name
				repository.Spec = repo6.Spec
				return &repository
			},
			expectedGlobalSecret: addTlsToSecret(newBasicAuthSecret("ns-6-apprepo-repo-6", kubeappsNamespace, "foo", "bar"), nil, nil, pub),
		},
		{
			name: "[issue 5746] secret updates ignored if not all credentials are provided - image pull secrets updates",
			existingDockerSecret: setSecretAnnotations(setSecretOwnerRef("repo-5",
				newAuthDockerSecret("pullsecret-repo-5", "ns-5",
					dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")))),

			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = &corev1.PackageRepositoryReference{
					Plugin:     &pluginDetail,
					Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
					Identifier: "repo-5",
				}
				request.Url = repo5.Spec.URL
				request.Description = repo5.Spec.Description
				request.CustomDetail = toProtoBufAny(&v1alpha1.HelmPackageRepositoryCustomDetail{
					ImagesPullSecret: &v1alpha1.ImagesPullSecret{
						DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_Credentials{
							Credentials: &corev1.DockerCredentials{
								Server:   RedactedString,
								Username: RedactedString,
								Password: RedactedString,
								Email:    "newemail",
							},
						},
					},
				})
				return request
			},

			expectedDockerSecret: setSecretAnnotations(setSecretOwnerRef("repo-5",
				newAuthDockerSecret("pullsecret-repo-5", "ns-5",
					dockerAuthJson("https://myfooserver.com", "username", "password", "newemail", "dXNlcm5hbWU6cGFzc3dvcmQ=")))),
			expectedRef: &corev1.PackageRepositoryReference{
				Plugin:     &pluginDetail,
				Context:    &corev1.Context{Namespace: "ns-5", Cluster: KubeappsCluster},
				Identifier: "repo-5",
			},
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Namespace = repo5.Namespace
				repository.Name = repo5.Name
				repository.Spec = repo5.Spec
				return &repository
			},
		},
	}

	repos := []*appRepov1alpha1.AppRepository{repo1, repo3, repo4, repo5, repo6, repo7}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var secrets []k8sruntime.Object
			if tc.existingAuthSecret != nil {
				secrets = append(secrets, tc.existingAuthSecret)
			}
			if tc.existingDockerSecret != nil {
				secrets = append(secrets, tc.existingDockerSecret)
			}

			s := newServerWithSecretsAndRepos(t, secrets, repos)

			request := tc.requestCustomizer(commonRequest())
			response, err := s.UpdatePackageRepository(context.Background(), connect.NewRequest(request))

			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != 0 {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositoryDetailResponse{},
				corev1.PackageRepositoryDetail{},
				corev1.GetPackageRepositorySummariesResponse{},
				corev1.PackageRepositorySummary{},
				corev1.PackageRepositoryReference{},
				corev1.PackageRepositoryStatus{},
				corev1.PackageRepositoryTlsConfig{},
				corev1.PackageRepositoryAuth{},
				v1alpha1.HelmPackageRepositoryCustomDetail{},
				anypb.Any{},
				corev1.PackageRepositoryAuth_UsernamePassword{},
				corev1.UsernamePassword{},
			)

			// check ref
			if got, want := response.Msg.GetPackageRepoRef(), tc.expectedRef; !cmp.Equal(want, got, opts) {
				t.Errorf("response mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// check repository
			appRepo, _, _, err := s.getPkgRepository(context.Background(), http.Header{}, tc.expectedRef.Context.Cluster, tc.expectedRef.Context.Namespace, tc.expectedRef.Identifier)
			if err != nil {
				t.Fatalf("unexpected error retrieving repository: %+v", err)
			}
			expectedRepository := repo1
			if tc.expectedRepoCustomizer != nil {
				expectedRepository = tc.expectedRepoCustomizer(*repo1)
			}

			if got, want := appRepo, expectedRepository; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			checkRepoSecrets(s, t, tc.userManagedSecrets, appRepo, expectedRepository, tc.expectedAuthSecret, tc.expectedDockerSecret, tc.expectedGlobalSecret)
		})
	}
}

func TestDeletePackageRepository(t *testing.T) {

	repos := []*appRepov1alpha1.AppRepository{repo1}

	testCases := []struct {
		name                       string
		existingObjects            []k8sruntime.Object
		request                    *corev1.DeletePackageRepositoryRequest
		expectedErrorCode          connect.Code
		expectedNonExistingSecrets []metav1.ObjectMeta
	}{
		{
			name: "no context provided",
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Plugin:     plugin,
					Identifier: "repo-1",
				},
			},
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name: "delete - success",
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Plugin:     plugin,
					Context:    &corev1.Context{Namespace: "ns-1", Cluster: KubeappsCluster},
					Identifier: "repo-1",
				},
			},
		},
		{
			name: "delete - not found (empty)",
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Plugin:     plugin,
					Context:    &corev1.Context{Namespace: "unknown-1", Cluster: KubeappsCluster},
					Identifier: "repo-1",
				},
			},
			expectedErrorCode: connect.CodeNotFound,
		},
		{
			name: "delete - deletes associated secrets",
			request: &corev1.DeletePackageRepositoryRequest{
				PackageRepoRef: &corev1.PackageRepositoryReference{
					Plugin:     plugin,
					Context:    &corev1.Context{Namespace: "ns-1", Cluster: KubeappsCluster},
					Identifier: "repo-1",
				},
			},
			expectedNonExistingSecrets: []metav1.ObjectMeta{
				{
					Name:      "apprepo-repo-1",
					Namespace: "ns-1",
				},
				{
					Name:      "ns-1-apprepo-repo-1",
					Namespace: kubeappsNamespace,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newServerWithSecretsAndRepos(t, nil, repos)

			_, err := s.DeletePackageRepository(context.Background(), connect.NewRequest(tc.request))

			// checks
			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != 0 {
				return
			}

			if tc.expectedNonExistingSecrets != nil {
				ctx := context.Background()
				typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster)
				if err != nil {
					t.Fatal(err)
				}
				for _, deletedSecret := range tc.expectedNonExistingSecrets {
					secret, err := typedClient.CoreV1().Secrets(deletedSecret.Namespace).Get(ctx, deletedSecret.Name, metav1.GetOptions{})
					if err != nil && !k8sErrors.IsNotFound(err) {
						t.Fatal(err)
					}
					if secret != nil {
						t.Fatalf("found existing secret '%s' in namespace '%s'", deletedSecret.Name, deletedSecret.Namespace)
					}
				}
			}
		})
	}
}

func TestGetPackageRepositoryPermissions(t *testing.T) {

	testCases := []struct {
		name              string
		request           *corev1.GetPackageRepositoryPermissionsRequest
		expectedErrorCode connect.Code
		expectedResponse  *corev1.GetPackageRepositoryPermissionsResponse
		reactors          []*ClientReaction
	}{
		{
			name: "returns permissions for global package repositories",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster},
			},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						createAction := action.(k8stesting.CreateActionImpl)
						accessReview := createAction.Object.(*authorizationv1.SelfSubjectAccessReview)
						if accessReview.Spec.ResourceAttributes.Namespace != globalPackagingNamespace {
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
						switch accessReview.Spec.ResourceAttributes.Verb {
						case "list", "delete":
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						default:
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: GetPluginDetail(),
						Global: map[string]bool{
							"create": false,
							"delete": true,
							"get":    false,
							"list":   true,
							"update": false,
							"watch":  false,
						},
						Namespace: nil,
					},
				},
			},
		},
		{
			name:    "returns local permissions when no cluster specified",
			request: &corev1.GetPackageRepositoryPermissionsRequest{},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: GetPluginDetail(),
						Global: map[string]bool{
							"create": true,
							"delete": true,
							"get":    true,
							"list":   true,
							"update": true,
							"watch":  true,
						},
						Namespace: nil,
					},
				},
			},
		},
		{
			name: "fails when namespace is specified but not the cluster",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Namespace: "my-ns"},
			},
			expectedErrorCode: connect.CodeInvalidArgument,
		},
		{
			name: "returns permissions for namespaced package repositories",
			request: &corev1.GetPackageRepositoryPermissionsRequest{
				Context: &corev1.Context{Cluster: KubeappsCluster, Namespace: "my-ns"},
			},
			reactors: []*ClientReaction{
				{
					verb:     "create",
					resource: "selfsubjectaccessreviews",
					reaction: func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
						createAction := action.(k8stesting.CreateActionImpl)
						accessReview := createAction.Object.(*authorizationv1.SelfSubjectAccessReview)
						if accessReview.Spec.ResourceAttributes.Namespace == globalPackagingNamespace {
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						}
						switch accessReview.Spec.ResourceAttributes.Verb {
						case "list", "delete":
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true}}, nil
						default:
							return true, &authorizationv1.SelfSubjectAccessReview{Status: authorizationv1.SubjectAccessReviewStatus{Allowed: false}}, nil
						}
					},
				},
			},
			expectedResponse: &corev1.GetPackageRepositoryPermissionsResponse{
				Permissions: []*corev1.PackageRepositoriesPermissions{
					{
						Plugin: GetPluginDetail(),
						Global: map[string]bool{
							"create": true,
							"delete": true,
							"get":    true,
							"list":   true,
							"update": true,
							"watch":  true,
						},
						Namespace: map[string]bool{
							"create": false,
							"delete": true,
							"get":    false,
							"list":   true,
							"update": false,
							"watch":  false,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := newServerWithAppRepoReactors(nil, nil, nil, tc.reactors, nil)

			response, err := s.GetPackageRepositoryPermissions(context.Background(), connect.NewRequest(tc.request))

			if got, want := connect.CodeOf(err), tc.expectedErrorCode; err != nil && got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedErrorCode != 0 {
				return
			}

			opts := cmpopts.IgnoreUnexported(
				corev1.Context{},
				plugins.Plugin{},
				corev1.GetPackageRepositoryPermissionsResponse{},
				corev1.PackageRepositoriesPermissions{},
			)
			if got, want := response.Msg, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func checkRepoSecrets(s *Server, t *testing.T, userManagedSecrets bool,
	actualRepo *appRepov1alpha1.AppRepository, expectedRepo *appRepov1alpha1.AppRepository,
	expectedAuthSecret *apiv1.Secret, expectedDockerSecret *apiv1.Secret,
	expectedGlobalSecret *apiv1.Secret) {
	ctx := context.Background()

	// Manually setting TypeMeta, as the fakeclient doesn't do it anymore:
	// https://github.com/kubernetes-sigs/controller-runtime/pull/2633
	actualRepo.TypeMeta = expectedRepo.TypeMeta

	if userManagedSecrets {
		if expectedAuthSecret != nil || expectedDockerSecret != nil {
			t.Fatalf("Error: unexpected state")
		}
		if got, want := actualRepo, expectedRepo; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	} else {
		if got, want := actualRepo, expectedRepo; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		if expectedAuthSecret != nil {
			if actualRepo.Spec.Auth.Header == nil && actualRepo.Spec.Auth.CustomCA == nil {
				t.Errorf("Error: Repository auth secret was expected but auth header and CA are empty")
			}
			typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster)
			if err != nil {
				t.Fatal(err)
			}
			if actualRepo.Spec.Auth.Header != nil {
				checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.Auth.Header.SecretKeyRef.Name, expectedAuthSecret)
			} else if actualRepo.Spec.Auth.CustomCA != nil {
				checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, expectedAuthSecret)
			}
		} else if actualRepo.Spec.Auth.Header != nil {
			t.Fatalf("Expected no secret, but found Header: [%v]", actualRepo.Spec.Auth.Header.SecretKeyRef)
		} else if actualRepo.Spec.Auth.CustomCA != nil {
			t.Fatalf("Expected no secret, but found CustomCA: [%v]", actualRepo.Spec.Auth.CustomCA.SecretKeyRef)
		} else if expectedRepo.Spec.Auth.Header != nil {
			t.Fatalf("Error: unexpected state")
		}

		if expectedDockerSecret != nil {
			if len(actualRepo.Spec.DockerRegistrySecrets) == 0 {
				t.Errorf("Error: Repository docker secret was expected but imagePullSecrets are empty")
			}
			typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster)
			if err != nil {
				t.Fatal(err)
			}
			checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.DockerRegistrySecrets[0], expectedDockerSecret)
		} else if len(actualRepo.Spec.DockerRegistrySecrets) > 0 {
			t.Fatalf("Expected no secret, but found image pull secrets: [%v]", actualRepo.Spec.DockerRegistrySecrets[0])
		} else if len(expectedRepo.Spec.DockerRegistrySecrets) > 0 {
			t.Fatalf("Error: unexpected state")
		}
	}
	checkGlobalSecret(s, t, expectedRepo, expectedGlobalSecret, expectedAuthSecret != nil || expectedRepo.Spec.Auth.Header != nil || expectedRepo.Spec.Auth.CustomCA != nil)
}

func checkGlobalSecret(s *Server, t *testing.T, expectedRepo *appRepov1alpha1.AppRepository, expectedGlobalSecret *apiv1.Secret, checkNoGlobalSecret bool) {
	ctx := context.Background()
	typedClient, err := s.clientGetter.Typed(http.Header{}, s.kubeappsCluster)
	if err != nil {
		t.Fatal(err)
	}
	repoGlobalSecretName := fmt.Sprintf("%s-apprepo-%s", expectedRepo.Namespace, expectedRepo.Name)
	if expectedGlobalSecret != nil {
		// Check for copied secret to global namespace
		actualGlobalSecret, err := typedClient.CoreV1().Secrets(s.kubeappsNamespace).Get(ctx, repoGlobalSecretName, metav1.GetOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := actualGlobalSecret, expectedGlobalSecret; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	} else if checkNoGlobalSecret {
		// Check that global secret does not exist
		secret, err := typedClient.CoreV1().Secrets(s.kubeappsNamespace).Get(ctx, repoGlobalSecretName, metav1.GetOptions{})
		if err != nil && !k8sErrors.IsNotFound(err) {
			t.Fatal(err)
		}
		if secret != nil {
			t.Errorf("global secret was found: %v", secret)
		}
	}
}

func checkSecrets(t *testing.T, ctx context.Context, typedClient kubernetes.Interface, namespace, secretName string, expectedSecret *apiv1.Secret) {
	opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
	if secretName != expectedSecret.GetName() {
		t.Errorf("Secret [%s] was expected to be named [%s]", expectedSecret.GetName(), secretName)
	}
	if secret, err := typedClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{}); err != nil {
		t.Fatal(err)
	} else if got, want := secret, expectedSecret; !cmp.Equal(want, got, opt2) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
	} else if !strings.HasPrefix(secret.Name, expectedSecret.Name) {
		t.Errorf("Secret Name [%s] was expected to start with [%s]",
			secret.Name, expectedSecret.Name)
	}
}

// see https://stackoverflow.com/a/30716481
func Ptr[T any](v T) *T {
	return &v
}
