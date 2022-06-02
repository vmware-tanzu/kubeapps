// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
		DockerRegistrySecrets: []string{"repo-4-secret"},
		OCIRepositories:       []string{"oci-repo-1", "oci-repo-2"},
		PassCredentials:       true,
		FilterRule: appRepov1alpha1.FilterRuleSpec{
			JQ:        "jq-filter-$1",
			Variables: map[string]string{"$1": "value1"},
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
}

var repo2Summary = &corev1.PackageRepositorySummary{
	PackageRepoRef:  repoRef("repo-2", KubeappsCluster, "ns-2"),
	Name:            "repo-2",
	Description:     "description 2",
	NamespaceScoped: true,
	Type:            "oci",
	Url:             "https://test-repo2",
}

var appReposAPIVersion = fmt.Sprintf("%s/%s", appRepov1alpha1.SchemeGroupVersion.Group, appRepov1alpha1.SchemeGroupVersion.Version)

func TestAddPackageRepository(t *testing.T) {
	// these will be used further on for TLS-related scenarios. Init
	// byte arrays up front so they can be re-used in multiple places later
	ca, _, _ := getCertsForTesting(t)

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
			statusCode:            codes.OK,
		},
		{
			name:       "errors when package repository with secret key reference (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository with secret key reference",
			request:            addRepoReqTLSSecretRef,
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       &addRepoTLSSecret,
			statusCode:         codes.OK,
			existingSecret:     newTlsSecret("secret-1", "foo", nil, nil, ca),
			userManagedSecrets: true,
		},
		{
			name:               "fails when package repository links to non-existing secret",
			request:            addRepoReqTLSSecretRef,
			statusCode:         codes.NotFound,
			userManagedSecrets: true,
		},
		{
			name:       "fails when package repository links to non-existing secret (kubeapps managed secrets)",
			request:    addRepoReqTLSSecretRef,
			statusCode: codes.InvalidArgument,
		},
		// BASIC AUTH
		{
			name:                  "package repository with basic auth and pass_credentials flag",
			request:               addRepoReqBasicAuth("baz", "zot"),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          &addRepoAuthHeaderPassCredentials,
			expectedCreatedSecret: setSecretOwnerRef("bar", newBasicAuthSecret("apprepo-bar", "foo", "baz", "zot")),
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
			name:               "package repository basic auth with existing secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-basic"),
			existingSecret:     newBasicAuthSecret("secret-basic", "foo", "baz-user", "zot-pwd"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
		},
		{
			name:       "package repository basic auth with existing secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"), //addRepoReq13,
			statusCode: codes.InvalidArgument,
		},
		// BEARER TOKEN
		{
			name:                  "package repository with bearer token",
			request:               addRepoReqBearerToken("the-token"),
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("apprepo-bar"),
			statusCode:            codes.OK,
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "Bearer the-token")),
		},
		{
			name:       "package repository with no bearer token",
			request:    addRepoReqBearerToken(""),
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository bearer token with secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "secret-bearer"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-bearer"),
			userManagedSecrets: true,
			existingSecret:     newAuthTokenSecret("secret-bearer", "foo", "Bearer the-token"),
			statusCode:         codes.OK,
		},
		{
			name:       "package repository bearer token with secret (kubeapps managed secrets)",
			request:    addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BEARER, "secret-bearer"),
			statusCode: codes.InvalidArgument,
		},
		{
			name:               "package repository bearer token (user managed secrets)",
			request:            addRepoReqBearerToken("the-token"),
			userManagedSecrets: true,
			statusCode:         codes.InvalidArgument,
		},
		// CUSTOM AUTH
		{
			name:                  "package repository with custom auth",
			request:               addRepoReqCustomAuth,
			expectedResponse:      addRepoExpectedResp,
			expectedRepo:          addRepoAuthHeaderWithSecretRef("apprepo-bar"),
			statusCode:            codes.OK,
			expectedCreatedSecret: setSecretOwnerRef("bar", newAuthTokenSecret("apprepo-bar", "foo", "foobarzot")),
		},
		{
			name:               "package repository custom auth with existing secret (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_CUSTOM, "secret-custom"),
			expectedResponse:   addRepoExpectedResp,
			expectedRepo:       addRepoAuthHeaderWithSecretRef("secret-custom"),
			existingSecret:     newBasicAuthSecret("secret-custom", "foo", "baz", "zot"),
			statusCode:         codes.OK,
			userManagedSecrets: true,
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
				newAuthDockerSecret("apprepo-bar",
					"foo",
					"{\"auths\":{\"https://docker-server\":{\"username\":\"the-user\",\"password\":\"the-password\",\"email\":\"foo@bar.com\",\"auth\":\"dGhlLXVzZXI6dGhlLXBhc3N3b3Jk\"}}}")),
			statusCode: codes.OK,
		},
		{
			name:               "package repository with Docker auth (user managed secrets)",
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_DOCKER_CONFIG_JSON, "secret-docker"),
			expectedResponse:   addRepoExpectedResp,
			userManagedSecrets: true,
			existingSecret: newAuthDockerSecret("secret-docker", "foo",
				"{\"auths\":{\"https://docker-server\":{\"username\":\"the-user\",\"password\":\"the-password\",\"email\":\"foo@bar.com\",\"auth\":\"dGhlLXVzZXI6dGhlLXBhc3N3b3Jk\"}}}"),
			expectedRepo: addRepoAuthDocker("secret-docker"),
			statusCode:   codes.OK,
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
			request:            addRepoReqAuthWithSecret(corev1.PackageRepositoryAuth_PACKAGE_REPOSITORY_AUTH_TYPE_BASIC_AUTH, "secret-basic"),
			existingSecret:     newTlsSecret("secret-basic", "foo", nil, nil, nil), // Creates empty secret
			userManagedSecrets: true,
			statusCode:         codes.Internal,
		},
		// Custom values
		{
			name:             "package repository with custom values",
			request:          addRepoReqCustomValues,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailsHelm,
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
			expectedRepo:     &addRepoCustomDetailsHelm,
			repoClientGetter: newRepoHttpClient(map[string]*http.Response{"https://example.com/index.yaml": {StatusCode: 200}}),
			statusCode:       codes.OK,
		},
		{
			name:             "package repository with validation success (OCI)",
			request:          addRepoReqCustomValuesOCIValid,
			expectedResponse: addRepoExpectedResp,
			expectedRepo:     &addRepoCustomDetailsOci,
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
					checkRepoSecrets(s, t, nsname.Namespace, tc.userManagedSecrets, &actualRepo, tc.expectedRepo, tc.expectedCreatedSecret)
				}
			}
		})
	}
}

func TestGetPackageRepositorySummaries(t *testing.T) {
	// some prep
	indexYAMLBytes, err := ioutil.ReadFile(testYaml("valid-index.yaml"))
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
	repo1Summary.Url = ts.URL
	repo2Summary.Url = ts.URL

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
		customDetails *v1alpha1.RepositoryCustomDetails) *corev1.GetPackageRepositoryDetailResponse {
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
			},
		}
		if customDetails != nil {
			response.Detail.CustomDetail = toProtoBufAny(customDetails)
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
				&v1alpha1.RepositoryCustomDetails{
					DockerRegistrySecrets: []string{"repo-4-secret"},
					OciRepositories:       []string{"oci-repo-1", "oci-repo-2"},
					FilterRule: &v1alpha1.RepositoryFilterRule{
						Jq:        "jq-filter-$1",
						Variables: map[string]string{"$1": "value1"},
					},
				}),
			existingSecret:     newTlsSecret("repo-4-secret", "ns-4", nil, nil, ca),
			expectedStatusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range []*appRepov1alpha1.AppRepository{repo1, repo2, repo3, repo4} {
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
				v1alpha1.RepositoryCustomDetails{},
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
			expectedRef:        defaultRef,
			expectedSecret:     setSecretOwnerRef("repo-1", newTlsSecret("apprepo-repo-1", "ns-1", nil, nil, ca)),
			expectedStatusCode: codes.OK,
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
			expectedRef:        defaultRef,
			expectedSecret:     setSecretOwnerRef("repo-1", newAuthTokenSecret("apprepo-repo-1", "ns-1", "Bearer foobarzot")),
			expectedStatusCode: codes.OK,
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
			expectedRef: defaultRef,
			//expectedSecret:     setSecretOwnerRef("repo-1", newAuthTokenSecret("apprepo-repo-1", "ns-1", "Bearer foobarzot")),
			expectedStatusCode: codes.OK,
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
	}

	repos := []*appRepov1alpha1.AppRepository{repo1, repo3, repo4}

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
				v1alpha1.RepositoryCustomDetails{},
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

			if tc.expectedSecret != nil {
				checkRepoSecrets(s, t, tc.expectedRef.Context.Namespace, tc.userManagedSecrets, appRepo, expectedRepository, tc.expectedSecret)
			}
		})
	}
}

func TestDeletePackageRepository(t *testing.T) {

	repos := []*appRepov1alpha1.AppRepository{repo1}

	testCases := []struct {
		name               string
		existingObjects    []k8sruntime.Object
		request            *corev1.DeletePackageRepositoryRequest
		expectedStatusCode codes.Code
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
		})
	}
}

func checkRepoSecrets(s *Server, t *testing.T, namespace string, userManagedSecrets bool,
	actualRepo *appRepov1alpha1.AppRepository, expectedRepo *appRepov1alpha1.AppRepository, expectedSecret *apiv1.Secret) {
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
			opt2 := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "Name", "GenerateName")
			if actualRepo.Spec.Auth.Header == nil && actualRepo.Spec.Auth.CustomCA == nil {
				t.Errorf("Error: Repository secrets were expected but auth header and CA are empty")
			} else if actualRepo.Spec.Auth.Header != nil {
				if !strings.HasPrefix(actualRepo.Spec.Auth.Header.SecretKeyRef.Name, expectedRepo.Spec.Auth.Header.SecretKeyRef.Name) {
					t.Errorf("Auth header SecretKeyRef [%s] was expected to start with [%s]",
						actualRepo.Spec.Auth.Header.SecretKeyRef.Name, expectedRepo.Spec.Auth.Header.SecretKeyRef.Name)
				}
				// check expected secret has been created
				if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
					t.Fatal(err)
				} else if secret, err := typedClient.CoreV1().Secrets(namespace).Get(ctx, actualRepo.Spec.Auth.Header.SecretKeyRef.Name, metav1.GetOptions{}); err != nil {
					t.Fatal(err)
				} else if got, want := secret, expectedSecret; !cmp.Equal(want, got, opt2) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
				} else if !strings.HasPrefix(secret.Name, expectedSecret.Name) {
					t.Errorf("Secret Name [%s] was expected to start with [%s]",
						secret.Name, expectedSecret.Name)
				}
			} else {
				if !strings.HasPrefix(actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, expectedRepo.Spec.Auth.CustomCA.SecretKeyRef.Name) {
					t.Errorf("CustomCA SecretKeyRef [%s] was expected to start with [%s]",
						actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, expectedRepo.Spec.Auth.CustomCA.SecretKeyRef.Name)
				}
				// check expected secret has been created
				if typedClient, err := s.clientGetter.Typed(ctx, s.kubeappsCluster); err != nil {
					t.Fatal(err)
				} else if secret, err := typedClient.CoreV1().Secrets(namespace).Get(ctx, actualRepo.Spec.Auth.CustomCA.SecretKeyRef.Name, metav1.GetOptions{}); err != nil {
					t.Fatal(err)
				} else if got, want := secret, expectedSecret; !cmp.Equal(want, got, opt2) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt2))
				} else if !strings.HasPrefix(secret.Name, expectedSecret.Name) {
					t.Errorf("Secret Name [%s] was expected to start with [%s]",
						secret.Name, expectedSecret.Name)
				}
			}
		} else if actualRepo.Spec.Auth.Header != nil {
			t.Fatalf("Expected no secret, but found Header: [%v]", actualRepo.Spec.Auth.Header.SecretKeyRef)
		} else if actualRepo.Spec.Auth.CustomCA != nil {
			t.Fatalf("Expected no secret, but found CustomCA: [%v]", actualRepo.Spec.Auth.CustomCA.SecretKeyRef)
		} else if expectedRepo.Spec.Auth.Header != nil {
			t.Fatalf("Error: unexpected state")
		}
	}
}
