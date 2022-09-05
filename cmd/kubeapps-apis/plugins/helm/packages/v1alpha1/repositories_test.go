// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appRepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	apiv1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
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
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: "repo-3-secret"}, Key: "AuthorizationHeader"},
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
				SecretKeyRef: apiv1.SecretKeySelector{LocalObjectReference: apiv1.LocalObjectReference{Name: "repo-4-secret"}, Key: "ca.crt"},
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
		DockerRegistrySecrets: []string{"repo-5-secret"},
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
		name                  string
		request               *corev1.AddPackageRepositoryRequest
		expectedResponse      *corev1.AddPackageRepositoryResponse
		expectedRepo          *appRepov1alpha1.AppRepository
		statusCode            codes.Code
		existingSecret        *apiv1.Secret
		expectedCreatedSecret *apiv1.Secret
		userManagedSecrets    bool
		repoClientGetter      newRepoClient
		expectedGlobalSecret  *apiv1.Secret
	}{
		{
			name:       "returns error if no namespace is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if no name is provided",
			request:    &corev1.AddPackageRepositoryRequest{Context: &corev1.Context{Namespace: "foo"}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if wrong repository type",
			request:    addRepoReqWrongType,
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "returns error if no url",
			request:    addRepoReqNoUrl,
			statusCode: codes.InvalidArgument,
		},
		{
			name: "check that interval is not used",
			request: &corev1.AddPackageRepositoryRequest{
				Name:            "bar",
				Context:         &corev1.Context{Namespace: "foo", Cluster: KubeappsCluster},
				Type:            "helm",
				Url:             "http://example.com",
				NamespaceScoped: true,
				Interval:        "1s",
			},
			statusCode: codes.InvalidArgument,
		},
		{
			name:             "simple add package repository scenario (HELM)",
			request:          addRepoReqSimple("helm"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleHelm,
			statusCode:       codes.OK,
		},
		{
			name:             "simple add package repository scenario (OCI)",
			request:          addRepoReqSimple("oci"),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoSimpleOci,
			statusCode:       codes.OK,
		},
		{
			name:             "add package global repository",
			request:          addRepoReqGlobal,
			expectedResponse: addRepoExpectedGlobalResp,
			expectedRepo:     &addRepoGlobal,
			statusCode:       codes.OK,
		},
		// CUSTOM CA AUTH
		{
			name:                  "package repository with tls cert authority",
			request:               addRepoReqTLSCA(ca),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepoWithTLSCA,
			expectedCreatedSecret: setSecretOwnerRef("bar", newTlsSecret("apprepo-bar", "foo", nil, nil, ca)),
			expectedGlobalSecret:  newTlsSecret("foo-apprepo-bar", kubeappsNamespace, nil, nil, ca), // Global secrets must reside in the Kubeapps (asset syncer) namespace
			statusCode:            codes.OK,
		},
		{
			name:       "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		{
			name:                 "package repository with secret key reference",
			request:              addRepoReqTLSSecretRef,
			userManagedSecrets:   true,
			existingSecret:       newTlsSecret("secret-1", "foo", nil, nil, ca),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         &addRepoTLSSecret,
			expectedGlobalSecret: newTlsSecret("foo-apprepo-bar", kubeappsNamespace, nil, nil, ca),
			statusCode:           codes.OK,
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            addRepoReqTLSSecretRef,
			userManagedSecrets: true,
			statusCode:         codes.NotFound,
		},
		{
			name:       "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		// BASIC AUTH
		{
			name:                  "[kubeapps managed secrets] package repository with basic auth and pass_credentials flag",
			request:               addRepoReqBasicAuth("baz", "zot"),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderPassCredentials("foo"),
			expectedCreatedSecret: setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", "foo", "baz", "zot")),
			expectedGlobalSecret:  newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz", "zot"),
			statusCode:            codes.OK,
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
			expectedResponse:      addRepoExpectedGlobalResp,
			expectedRepo:          addRepoAuthHeaderPassCredentials(globalPackagingNamespace),
			expectedCreatedSecret: setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", globalPackagingNamespace, "the-user", "the-pwd")),
			expectedGlobalSecret:  newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "the-user", "the-pwd"),
			statusCode:            codes.OK,
		},
		{
			name:       "package repository with wrong basic auth",
			request:    addRepoReqWrongBasicAuth,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "fails for package repository passing basic auth (user managed secrets)",
			request:            addRepoReqBasicAuth("kermit", "frog"),
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
		},
		{
			name:                 "[user managed secrets] package repository basic auth with existing secret",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "foo", "secret-basic"),
			userManagedSecrets:   true,
			existingSecret:       newBasicAuthSecret("secret-basic", "foo", "baz-user", "zot-pwd"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-basic"),
			expectedGlobalSecret: newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz-user", "zot-pwd"),
			statusCode:           codes.OK,
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
			existingSecret:       newBasicAuthSecret("secret-basic", globalPackagingNamespace, "baz-user", "zot-pwd"),
			expectedResponse:     addRepoExpectedGlobalResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef(globalPackagingNamespace, "secret-basic"),
			expectedGlobalSecret: newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "baz-user", "zot-pwd"),
			statusCode:           codes.OK,
		},
		{
			name:       "package repository basic auth with existing secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "foo", "secret-basic"), //addRepoReq13,
			statusCode: codes.InvalidArgument,
		},
		// BEARER TOKEN
		{
			name:                  "package repository with bearer token w/o prefix",
			request:               addRepoReqBearerToken("the-token", false),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("foo", "apprepo-bar"),
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "Bearer the-token")),
			expectedGlobalSecret:  newAuthTokenSecret("foo-apprepo-bar", kubeappsNamespace, "Bearer the-token"),
			statusCode:            codes.OK,
		},
		{
			name:                  "package repository with bearer token w/ prefix",
			request:               addRepoReqBearerToken("the-token", true),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("foo", "apprepo-bar"),
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "Bearer the-token")),
			expectedGlobalSecret:  newAuthTokenSecret("foo-apprepo-bar", kubeappsNamespace, "Bearer the-token"),
			statusCode:            codes.OK,
		},
		{
			name:       "package repository with no bearer token",
			request:    addRepoReqBearerToken("", false),
			statusCode: codes.InvalidArgument,
		},
		{
			name:                 "package repository bearer token with secret (user managed secrets)",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "foo", "secret-bearer"),
			userManagedSecrets:   true,
			existingSecret:       newAuthTokenSecret("secret-bearer", "foo", "Bearer the-token"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-bearer"),
			expectedGlobalSecret: newAuthTokenSecret("foo-apprepo-bar", kubeappsNamespace, "Bearer the-token"),
			statusCode:           codes.OK,
		},
		{
			name:       "package repository bearer token with secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "foo", "secret-bearer"),
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository bearer token (user managed secrets)",
			request:            addRepoReqBearerToken("the-token", false),
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
		},
		// CUSTOM AUTH
		{
			name:                  "package repository with custom auth",
			request:               addRepoReqCustomAuth,
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("foo", "apprepo-bar"),
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "foobarzot")),
			expectedGlobalSecret:  newAuthTokenSecret("foo-apprepo-bar", kubeappsNamespace, "foobarzot"),
			statusCode:            codes.OK,
		},
		{
			name:                 "package repository custom auth with existing secret (user managed secrets)",
			request:              addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_AUTHORIZATION_HEADER, "foo", "secret-custom"),
			userManagedSecrets:   true,
			existingSecret:       newBasicAuthSecret("secret-custom", "foo", "baz", "zot"),
			expectedResponse:     addRepoExpectedResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef("foo", "secret-custom"),
			expectedGlobalSecret: newBasicAuthSecret("foo-apprepo-bar", kubeappsNamespace, "baz", "zot"),
			statusCode:           codes.OK,
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
			existingSecret:       newBasicAuthSecret("secret-custom", globalPackagingNamespace, "baz", "zot"),
			expectedResponse:     addRepoExpectedGlobalResp,
			expectedRepo:         addRepoAuthHeaderWithSecretRef(globalPackagingNamespace, "secret-custom"),
			expectedGlobalSecret: newBasicAuthSecret("kubeapps-repos-global-apprepo-bar", kubeappsNamespace, "baz", "zot"),
			statusCode:           codes.OK,
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
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newAuthDockerSecret("apprepo-bar", "foo",
					dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk"))),
			expectedGlobalSecret: newAuthDockerSecret("foo-apprepo-bar", kubeappsNamespace,
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			statusCode: codes.OK,
		},
		{
			name:               "package repository with Docker auth (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON, "foo", "secret-docker"),
			userManagedSecrets: true,
			existingSecret: newAuthDockerSecret("secret-docker", "foo",
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     addRepoAuthDocker("secret-docker"),
			expectedGlobalSecret: newAuthDockerSecret("foo-apprepo-bar", kubeappsNamespace,
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			statusCode: codes.OK,
		},
		// Others
		{
			name:       "errors when package repository with 1 secret for TLS CA and a different secret for basic auth (kubeapps managed secrets)",
			request:    addRepoReqTLSDifferentSecretAuth,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "errors when package repository with 1 secret for TLS CA and a different secret for basic auth",
			request:            addRepoReqTLSDifferentSecretAuth,
			statusCode:         codes.InvalidArgument,
			userManagedSecrets: true,
		},
		{
			name:             "package repository with just pass_credentials flag",
			request:          addRepoReqOnlyPassCredentials,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoOnlyPassCredentials,
			statusCode:       codes.OK,
		},
		{
			name:               "package repository with reference to malformed secret",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "foo", "secret-basic"),
			existingSecret:     newTlsSecret("secret-basic", "foo", nil, nil, nil), // Creates empty secret
			userManagedSecrets: true,
			statusCode:         codes.Internal,
		},
		// Custom values
		{
			name:             "package repository with custom values",
			request:          addRepoReqCustomValues,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailHelm,
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with invalid custom values",
			request:          addRepoReqWrongCustomValues,
			expectedResponse: addRepoExpectedResp,
			statusCode:       codes.InvalidArgument,
		},
		{
			name:             "package repository with validation success (Helm)",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailHelm,
			repoClientGetter: newRepoHttpClient(map[string]*http.Response{"https://example.com/index.yaml": {StatusCode: 200}}),
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with validation success (OCI)",
			request:          addRepoReqCustomValuesOCIValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailOci,
			repoClientGetter: newRepoHttpClient(map[string]*http.Response{
				"https://example.com/v2/repo1/tags/list?n=1":  httpResponse(200, "{ \"name\":\"repo1\", \"tags\":[\"tag1\"] }"),
				"https://example.com/v2/repo1/manifests/tag1": httpResponse(200, "{ \"config\":{ \"mediaType\":\"application/vnd.cncf.helm.config\" } }"),
			}),
			statusCode: codes.OK,
		},
		{
			name:             "package repository with validation failing",
			request:          addRepoReqCustomValuesHelmValid,
			expectedResponse: addRepoExpectedResp,
			repoClientGetter: newRepoHttpClient(
				map[string]*http.Response{
					"https://example.com/index.yaml": httpResponse(404, "It failed because of X and Y"),
				}),
			statusCode: codes.FailedPrecondition,
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
			existingSecret: newAuthDockerSecret("secret-docker", "foo",
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
			statusCode: codes.OK,
		},
		{
			name: "[user managed secrets] create repository fails with Docker credentials",
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
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
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
			statusCode:         codes.NotFound,
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
			existingSecret:     newAuthTokenSecret("secret-docker", "foo", ""),
			statusCode:         codes.InvalidArgument,
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
			existingSecret: &apiv1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "secret-docker",
					Namespace: "foo",
				},
				Type: apiv1.SecretTypeDockerConfigJson,
				Data: map[string][]byte{
					"wrong-key": []byte(""),
				},
			},
			statusCode: codes.InvalidArgument,
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
			expectedCreatedSecret: setSecretOwnerRef("bar",
				newAuthDockerSecret("pullsecret-bar", "foo",
					dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ="))),
			statusCode: codes.OK,
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
			existingSecret: newAuthDockerSecret("pullsecret-bar", "foo",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			statusCode: codes.AlreadyExists,
		},
		{
			name: "[kubeapps managed secrets] create repository fails with pull secret ref",
			request: newPackageRepoRequestWithDetails(&v1alpha1.HelmPackageRepositoryCustomDetail{
				ImagesPullSecret: &v1alpha1.ImagesPullSecret{
					DockerRegistryCredentialOneOf: &v1alpha1.ImagesPullSecret_SecretRef{
						SecretRef: "secret-docker",
					},
				},
			}),
			statusCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var secrets []k8sruntime.Object
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}
			s := newServerWithSecretsAndRepos(t, secrets, nil, nil)
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets
			if tc.repoClientGetter != nil {
				s.repoClientGetter = tc.repoClientGetter
			}

			nsname := types.NamespacedName{Namespace: tc.request.Context.Namespace, Name: tc.request.Name}
			ctx := context.Background()
			response, err := s.AddPackageRepository(ctx, tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// Only check the expectedResponse for OK status.
			if tc.statusCode == codes.OK {
				if response == nil {
					t.Fatalf("got: nil, want: expectedResponse")
				} else {
					opt1 := cmpopts.IgnoreUnexported(
						corev1.AddPackageRepositoryResponse{},
						corev1.Context{},
						corev1.PackageRepositoryReference{},
						plugins.Plugin{},
					)
					if got, want := response, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
					}
				}
			}

			// purposefully not calling mock.ExpectationsWereMet() here because
			// AddPackageRepository will trigger an ADD event that will be processed
			// asynchronously, so it may or may not have enough time to get to the
			// point where the cache worker does a GET

			// We don't need to check anything else for non-OK codes.
			if tc.statusCode != codes.OK {
				return
			}

			// check expected HelmRelease CRD has been created
			if ctrlClient, err := s.clientGetter.ControllerRuntime(ctx, s.kubeappsCluster); err != nil {
				t.Fatal(err)
			} else {
				var actualRepo appRepov1alpha1.AppRepository
				if err = ctrlClient.Get(ctx, nsname, &actualRepo); err != nil {
					t.Fatal(err)
				} else {
					checkRepoSecrets(s, t, tc.userManagedSecrets, &actualRepo, tc.expectedRepo, tc.expectedCreatedSecret, tc.expectedGlobalSecret)
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
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetPackageRepositorySummariesResponse
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
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetPackageRepositorySummariesResponse{
				PackageRepositorySummaries: []*corev1.PackageRepositorySummary{
					repo1Summary,
					repo2Summary,
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.OK,
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
			s := newServerWithSecretsAndRepos(t, nil, unstructuredObjects, nil)

			response, err := s.GetPackageRepositorySummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
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
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts, opts2) {
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
		expectedStatusCode   codes.Code
		existingSecret       *apiv1.Secret
	}{
		{
			name:               "not found",
			request:            buildRequest("ns-1", "foo"),
			expectedStatusCode: codes.NotFound,
		},
		{
			name:    "check ref",
			request: buildRequest("ns-1", "repo-1"),
			expectedResponse: buildResponse("ns-1", "repo-1", "helm",
				"https://test-repo", "description 1", nil, nil, nil),
			expectedStatusCode: codes.OK,
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
			existingSecret:     newBasicAuthSecret("repo-3-secret", globalPackagingNamespace, "baz-user", "zot-pwd"),
			expectedStatusCode: codes.OK,
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
			existingSecret:     newTlsSecret("repo-4-secret", "ns-4", nil, nil, ca),
			expectedStatusCode: codes.OK,
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
			existingSecret: newAuthDockerSecret("repo-5-secret", "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedStatusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range []*appRepov1alpha1.AppRepository{repo1, repo2, repo3, repo4, repo5} {
				repository := obj
				if tc.repositoryCustomizer != nil {
					repository = tc.repositoryCustomizer(obj)
				}
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(repository)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			var secrets []k8sruntime.Object
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}

			s := newServerWithSecretsAndRepos(t, secrets, unstructuredObjects, nil)

			actualResponse, err := s.GetPackageRepositoryDetail(context.Background(), tc.request)

			// checks
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != codes.OK {
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

			if got, want := tc.request.PackageRepoRef, actualResponse.Detail.PackageRepoRef; !cmp.Equal(want, got, opts) {
				t.Errorf("ref mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			if got, want := actualResponse, tc.expectedResponse; !cmp.Equal(want, got, opts) {
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

	ca, _, _ := getCertsForTesting(t)

	testCases := []struct {
		name                   string
		requestCustomizer      func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest
		expectedRepoCustomizer func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository
		expectedStatusCode     codes.Code
		expectedRef            *corev1.PackageRepositoryReference
		existingSecret         *apiv1.Secret
		expectedSecret         *apiv1.Secret
		userManagedSecrets     bool
		expectedGlobalSecret   *apiv1.Secret
	}{
		{
			name: "invalid package repo ref",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef = nil
				return request
			},
			expectedStatusCode: codes.InvalidArgument,
		}, {
			name: "repository not found",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef.Context = &corev1.Context{Cluster: "other", Namespace: globalPackagingNamespace}
				return request
			},
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "validate name",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.PackageRepoRef.Identifier = ""
				return request
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "validate url",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Url = ""
				return request
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "check that interval is not used",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
				request.Interval = "1s"
				return request
			},
			expectedStatusCode: codes.InvalidArgument,
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
			expectedSecret:       setSecretOwnerRef("repo-1", newTlsSecret("apprepo-repo-1", "ns-1", nil, nil, ca)),
			expectedGlobalSecret: newTlsSecret("ns-1-apprepo-repo-1", kubeappsNamespace, nil, nil, ca),
			expectedStatusCode:   codes.OK,
		},
		{
			name:           "update removing tsl config",
			existingSecret: newTlsSecret("repo-4-secret", "ns-4", nil, nil, ca),
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
			expectedStatusCode: codes.OK,
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
			expectedSecret:       setSecretOwnerRef("repo-1", newAuthTokenSecret("apprepo-repo-1", "ns-1", "Bearer foobarzot")),
			expectedGlobalSecret: newAuthTokenSecret("ns-1-apprepo-repo-1", kubeappsNamespace, "Bearer foobarzot"),
			expectedStatusCode:   codes.OK,
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
			existingSecret:     newAuthTokenSecret("my-own-secret", "ns-1", "Bearer foobarzot"),
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
			expectedGlobalSecret: newAuthTokenSecret("ns-1-apprepo-repo-1", kubeappsNamespace, "Bearer foobarzot"),
			expectedStatusCode:   codes.OK,
		},
		{
			name:           "update removing auth",
			existingSecret: newAuthTokenSecret("repo-3-secret", globalPackagingNamespace, "token-value"),
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.OK,
			expectedRef:        defaultRef,
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
			expectedStatusCode: codes.OK,
			expectedRef:        defaultRef,
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
			expectedSecret: setSecretOwnerRef("repo-1",
				newAuthDockerSecret("pullsecret-repo-1", "ns-1",
					dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ="))),
			expectedStatusCode: codes.OK,
			expectedRef:        defaultRef,
		},
		{
			name: "[kubeapps managed secrets] update repo with image pull secret redacted",
			requestCustomizer: func(request *corev1.UpdatePackageRepositoryRequest) *corev1.UpdatePackageRepositoryRequest {
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
			existingSecret: newAuthDockerSecret("pullsecret-repo-1", "ns-1",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedRepoCustomizer: func(repository appRepov1alpha1.AppRepository) *appRepov1alpha1.AppRepository {
				repository.ResourceVersion = "2"
				repository.Spec.URL = "https://new-repo-url"
				repository.Spec.Description = "description"
				repository.Spec.DockerRegistrySecrets = []string{"pullsecret-repo-1"}
				return &repository
			},
			expectedStatusCode: codes.OK,
			expectedRef:        defaultRef,
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
			existingSecret: newAuthDockerSecret("test-pull-secret", "ns-1",
				dockerAuthJson("https://docker-server", "the-user", "the-password", "foo@bar.com", "dGhlLXVzZXI6dGhlLXBhc3N3b3Jk")),
			expectedStatusCode: codes.OK,
			expectedRef:        defaultRef,
		},
		{
			name: "[user managed secrets] update repo with Docker credentials should fail",
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
			userManagedSecrets: true,
			expectedStatusCode: codes.InvalidArgument,
			expectedRef:        defaultRef,
		},
		{
			name: "[kubeapps managed secrets] update repo with secret ref should fail",
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
			expectedStatusCode: codes.InvalidArgument,
			expectedRef:        defaultRef,
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
			existingSecret: newAuthDockerSecret("pullsecret-repo-5", "ns-5",
				dockerAuthJson("https://myfooserver.com", "username", "password", "foo@bar.com", "dXNlcm5hbWU6cGFzc3dvcmQ=")),
			expectedStatusCode: codes.OK,
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
	}

	repos := []*appRepov1alpha1.AppRepository{repo1, repo3, repo4, repo5}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range repos {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			var secrets []k8sruntime.Object
			if tc.existingSecret != nil {
				secrets = append(secrets, tc.existingSecret)
			}

			s := newServerWithSecretsAndRepos(t, secrets, unstructuredObjects, repos)
			s.pluginConfig.UserManagedSecrets = tc.userManagedSecrets

			request := tc.requestCustomizer(commonRequest())
			response, err := s.UpdatePackageRepository(context.Background(), request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != codes.OK {
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
			if got, want := response.GetPackageRepoRef(), tc.expectedRef; !cmp.Equal(want, got, opts) {
				t.Errorf("response mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}

			// check repository
			appRepo, _, _, err := s.getPkgRepository(context.Background(), tc.expectedRef.Context.Cluster, tc.expectedRef.Context.Namespace, tc.expectedRef.Identifier)
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

			checkRepoSecrets(s, t, tc.userManagedSecrets, appRepo, expectedRepository, tc.expectedSecret, tc.expectedGlobalSecret)
		})
	}
}

func TestDeletePackageRepository(t *testing.T) {

	repos := []*appRepov1alpha1.AppRepository{repo1}

	testCases := []struct {
		name                       string
		existingObjects            []k8sruntime.Object
		request                    *corev1.DeletePackageRepositoryRequest
		expectedStatusCode         codes.Code
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
			expectedStatusCode: codes.InvalidArgument,
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
			expectedStatusCode: codes.OK,
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
			expectedStatusCode: codes.NotFound,
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
			expectedStatusCode: codes.OK,
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range repos {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := newServerWithSecretsAndRepos(t, nil, unstructuredObjects, repos)

			_, err := s.DeletePackageRepository(context.Background(), tc.request)

			// checks
			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got error: %d, want: %d, err: %+v", got, want, err)
			} else if got != codes.OK {
				return
			}

			if tc.expectedNonExistingSecrets != nil {
				ctx := context.Background()
				typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
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

func checkRepoSecrets(s *Server, t *testing.T, userManagedSecrets bool,
	actualRepo *appRepov1alpha1.AppRepository, expectedRepo *appRepov1alpha1.AppRepository,
	expectedSecret *apiv1.Secret, expectedGlobalSecret *apiv1.Secret) {
	ctx := context.Background()
	if userManagedSecrets {
		if expectedSecret != nil {
			t.Fatalf("Error: unexpected state")
		}
		if got, want := actualRepo, expectedRepo; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	} else {
		if got, want := actualRepo, expectedRepo; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		if expectedSecret != nil {
			if actualRepo.Spec.Auth.Header == nil && actualRepo.Spec.Auth.CustomCA == nil && len(actualRepo.Spec.DockerRegistrySecrets) == 0 {
				t.Errorf("Error: Repository secrets were expected but auth header, CA and imagePullSecret are empty")
			}
			typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
			if err != nil {
				t.Fatal(err)
			}
			if actualRepo.Spec.Auth.Header != nil {
				checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.Auth.Header.SecretKeyRef.Name, expectedSecret)
			} else if actualRepo.Spec.Auth.CustomCA != nil {
				checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, expectedSecret)
			} else {
				// Docker image pull secret check
				checkSecrets(t, ctx, typedClient, actualRepo.Namespace, actualRepo.Spec.DockerRegistrySecrets[0], expectedSecret)
			}
		} else if actualRepo.Spec.Auth.Header != nil {
			t.Fatalf("Expected no secret, but found Header: [%v]", actualRepo.Spec.Auth.Header.SecretKeyRef)
		} else if actualRepo.Spec.Auth.CustomCA != nil {
			t.Fatalf("Expected no secret, but found CustomCA: [%v]", actualRepo.Spec.Auth.CustomCA.SecretKeyRef)
		} else if expectedRepo.Spec.Auth.Header != nil {
			t.Fatalf("Error: unexpected state")
		}
	}
	checkGlobalSecret(s, t, expectedRepo, expectedGlobalSecret, expectedSecret != nil || expectedRepo.Spec.Auth.Header != nil || expectedRepo.Spec.Auth.CustomCA != nil)
}

func checkGlobalSecret(s *Server, t *testing.T, expectedRepo *appRepov1alpha1.AppRepository, expectedGlobalSecret *apiv1.Secret, checkNoGlobalSecret bool) {
	ctx := context.Background()
	typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster)
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
