// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	v1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	fakeapprepoclientset "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/fake"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakecoreclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	fakeRest "k8s.io/client-go/rest/fake"
	log "k8s.io/klog/v2"
)

type repoStub struct {
	name    string
	private bool
}

type secretStub struct {
	name string
}

type fakeHTTPCli struct {
	request  *http.Request
	response *http.Response
	err      error
}

func (f *fakeHTTPCli) Do(r *http.Request) (*http.Response, error) {
	f.request = r
	return f.response, f.err
}

const kubeappsNamespace = "kubeapps"

func makeAppRepoObjects(reposPerNamespace map[string][]repoStub) []k8sruntime.Object {
	objects := []k8sruntime.Object{}
	for namespace, repoStubs := range reposPerNamespace {
		for _, repoStub := range repoStubs {
			appRepo := &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoStub.name,
					Namespace: namespace,
				},
			}
			if repoStub.private {
				authHeader := &v1alpha1.AppRepositoryAuthHeader{}
				authHeader.SecretKeyRef.LocalObjectReference.Name = secretNameForRepo(repoStub.name)
				appRepo.Spec.Auth.Header = authHeader
			}
			objects = append(objects, k8sruntime.Object(appRepo))
		}
	}
	return objects
}

func makeSecretsForRepos(reposPerNamespace map[string][]repoStub, kubeappsNamespace string) []k8sruntime.Object {
	objects := []k8sruntime.Object{}
	for namespace, repoStubs := range reposPerNamespace {
		for _, repoStub := range repoStubs {
			// Only create secrets if it's a private repo.
			if !repoStub.private {
				continue
			}
			var appRepo k8sruntime.Object = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretNameForRepo(repoStub.name),
					Namespace: namespace,
				},
			}
			objects = append(objects, appRepo)

			// Only create a copy of the secret in the kubeapps namespace if the app repo
			// is in a user namespace.
			if namespace != kubeappsNamespace {
				var appRepo k8sruntime.Object = &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      KubeappsSecretNameForRepo(repoStub.name, namespace),
						Namespace: kubeappsNamespace,
					},
				}
				objects = append(objects, appRepo)
			}
		}
	}
	return objects
}

type fakeAppRepoClientset = fakeapprepoclientset.Clientset
type fakeCombinedClientset struct {
	*fakeAppRepoClientset
	*fakecoreclientset.Clientset
	rc *fakeRest.RESTClient
}

// Not sure why golang thinks this Discovery() is ambiguous on the fake but not on
// the real combinedClientset, but to satisfy:
func (f fakeCombinedClientset) Discovery() discovery.DiscoveryInterface {
	return f.Clientset.Discovery()
}

func (f fakeCombinedClientset) RestClient() rest.Interface {
	return f.rc
}

func (f fakeCombinedClientset) MaxWorkers() int {
	return 1
}

func checkErr(t *testing.T, err error, expectedError error) {
	if err == nil && expectedError != nil {
		t.Errorf("got: nil, want: %+v", expectedError)
	} else if err != nil {
		if expectedError == nil {
			t.Errorf("got: %+v, want: nil", err)
		} else if got, want := err.Error(), expectedError.Error(); got != want {
			t.Errorf("got: %q, want: %q", got, want)
		}
	}
}

func checkAppRepo(t *testing.T, requestData string, requestNamespace string, cs fakeCombinedClientset) (appRepositoryRequest, *v1alpha1.AppRepository, *v1alpha1.AppRepository) {
	var appRepoRequest appRepositoryRequest
	err := json.NewDecoder(strings.NewReader(requestData)).Decode(&appRepoRequest)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	// Ensure the expected AppRepository is stored
	expectedAppRepo := appRepositoryForRequest(&appRepoRequest)
	expectedAppRepo.ObjectMeta.Namespace = requestNamespace

	responseAppRepo, err := cs.KubeappsV1alpha1().AppRepositories(requestNamespace).Get(context.TODO(), expectedAppRepo.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("expected data %v not present: %+v", expectedAppRepo, err)
	}

	if got, want := responseAppRepo, expectedAppRepo; !cmp.Equal(want, got) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
	return appRepoRequest, expectedAppRepo, responseAppRepo
}

func checkSecrets(t *testing.T, requestNamespace string, appRepoRequest appRepositoryRequest, expectedAppRepo *v1alpha1.AppRepository, responseAppRepo *v1alpha1.AppRepository, handler userHandler) {
	// TODO(#1655)
	// The fake k8s API does not generate UID's for created object
	// (among other things). We would need to add reactors to the fake
	// to do so to test the UID being set on the secret.
	// https://github.com/kubernetes/client-go/issues/439
	// responseAppRepo.ObjectMeta.UID = "5ef40f28-3c69-460d-bd12-c00b944e6d1b"

	// When appropriate, ensure the expected secret is stored.
	if appRepoRequest.AppRepository.AuthHeader != "" {
		expectedSecret, err := handler.secretForRequest(&appRepoRequest, responseAppRepo, requestNamespace)
		if err != nil {
			t.Errorf("error getting the expected secret: %+v", err)
		}
		expectedSecret.ObjectMeta.Namespace = requestNamespace
		responseSecret, err := handler.clientset.CoreV1().Secrets(requestNamespace).Get(context.TODO(), expectedSecret.ObjectMeta.Name, metav1.GetOptions{})

		if err != nil {
			t.Errorf("expected data %v not present: %+v", expectedSecret, err)
		}

		if got, want := responseSecret, expectedSecret; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		// Verify the copy of the repo secret in kubeapps is
		// also stored if this is a per-namespace app repository.
		kubeappsSecretName := KubeappsSecretNameForRepo(expectedAppRepo.ObjectMeta.Name, expectedAppRepo.ObjectMeta.Namespace)
		expectedSecret.ObjectMeta.Name = kubeappsSecretName
		expectedSecret.ObjectMeta.Namespace = kubeappsNamespace
		// The owner ref cannot be present for the copy in the kubeapps namespace.
		expectedSecret.ObjectMeta.OwnerReferences = nil

		if requestNamespace != kubeappsNamespace {
			responseSecret, err = handler.clientset.CoreV1().Secrets(kubeappsNamespace).Get(context.TODO(), kubeappsSecretName, metav1.GetOptions{})
			if err != nil {
				t.Errorf("expected data %v not present: %+v", expectedSecret, err)
			}

			if got, want := responseSecret, expectedSecret; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		} else {
			// The copy of the secret should not be created when the request namespace is kubeapps.
			secret, err := handler.clientset.CoreV1().Secrets(kubeappsNamespace).Get(context.TODO(), kubeappsSecretName, metav1.GetOptions{})
			if err == nil {
				t.Fatalf("secret should not be created, found %+v", secret)
			}
			if got, want := errorCodeForK8sError(t, err), 404; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		}
	}

}

func TestAppRepositoryCreate(t *testing.T) {
	testCases := []struct {
		name             string
		requestNamespace string
		existingRepos    map[string][]repoStub
		requestData      string
		expectedError    error
	}{
		{
			name:             "it creates an app repository in the default kubeappsNamespace",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
		},
		{
			name:             "it creates an app repository in a specific namespace",
			requestNamespace: "my-namespace",
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
		},
		{
			name:             "it creates an app repository with an empty template",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "syncJobPodTemplate": {}}}`,
		},
		{
			name:             "it includes the docker registry secret names when provided",
			requestNamespace: "other-namespace",
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "registrySecrets": ["secret-one", "secret-two"]}}`,
		},
		{
			name:             "it errors if docker registry secrets are included for a global app repository",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "registrySecrets": ["secret-one", "secret-two"]}}`,
			expectedError:    ErrGlobalRepositoryWithSecrets,
		},
		{
			name:             "it errors if the repo exists in the kubeapps ns already",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "bitnami"}}`,
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "bitnami"}},
			},
			expectedError: fmt.Errorf(`apprepositories.kubeapps.com "bitnami" already exists`),
		},
		{
			name:             "it creates the repo even if the same repo exists in other namespaces",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "bitnami"}}`,
			existingRepos: map[string][]repoStub{
				"kubeapps-other-ns-1": {repoStub{name: "bitnami"}},
				"kubeapps-other-ns-2": {repoStub{name: "bitnami"}},
			},
		},
		{
			name:             "it results in a bad request if the json cannot be parsed",
			requestNamespace: kubeappsNamespace,
			requestData:      `not a { json object`,
			expectedError:    fmt.Errorf(`invalid character 'o' in literal null (expecting 'u')`),
		},
		{
			name:             "it creates a secret if the auth header is set",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it creates a copy of the namespaced repo secret in the kubeapps namespace",
			requestNamespace: "test-namespace",
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it creates an app repo with a description",
			requestNamespace: "test-namespace",
			requestData:      `{"appRepository": {"name": "test-repo-2", "url": "http://example.com/test-repo", "description": "test-me"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...),
				fakecoreclientset.NewSimpleClientset(),
				&fakeRest.RESTClient{},
			}
			handler := userHandler{
				kubeappsNamespace: kubeappsNamespace,
				svcClientset:      cs,
				clientset:         cs,
			}

			apprepo, err := handler.CreateAppRepository(io.NopCloser(strings.NewReader(tc.requestData)), tc.requestNamespace)
			checkErr(t, err, tc.expectedError)

			if apprepo != nil {
				appRepoRequest, expectedAppRepo, responseAppRepo := checkAppRepo(t, tc.requestData, tc.requestNamespace, cs)
				checkSecrets(t, tc.requestNamespace, appRepoRequest, expectedAppRepo, responseAppRepo, handler)
			}
		})
	}
}

func TestAppRepositoryList(t *testing.T) {
	testCases := []struct {
		name             string
		requestNamespace string
		existingRepos    map[string][]repoStub
	}{
		{
			name:             "it gets repos from the global namespace",
			requestNamespace: kubeappsNamespace,
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "test-repo"}},
			},
		},
		{
			name:             "it gets repos from a namespace",
			requestNamespace: "foo",
			existingRepos: map[string][]repoStub{
				"foo": {repoStub{name: "test-repo"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...),
				fakecoreclientset.NewSimpleClientset(),
				&fakeRest.RESTClient{},
			}
			// Depending on the namespace, we instantiate the svcClientset or the user clientset
			// to ensure that we are using the expected clientset.
			handler := userHandler{
				kubeappsNamespace: kubeappsNamespace,
				svcClientset:      cs,
			}
			if tc.requestNamespace != kubeappsNamespace {
				handler = userHandler{
					kubeappsNamespace: kubeappsNamespace,
					clientset:         cs,
				}
			}

			apprepos, err := handler.ListAppRepositories(tc.requestNamespace)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if got, want := len(apprepos.Items), len(tc.existingRepos[tc.requestNamespace]); got != want {
				t.Errorf("expected %d repos, got %d", want, got)
			}
		})
	}
}

func TestAppRepositoryUpdate(t *testing.T) {
	const kubeappsNamespace = "kubeapps"
	testCases := []struct {
		name             string
		requestNamespace string
		existingRepos    map[string][]repoStub
		existingSecrets  map[string][]secretStub
		requestData      string
		expectedError    error
	}{
		{
			name:             "it updates an app repository in the default kubeappsNamespace",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "test-repo"}},
			},
		},
		{
			name:             "it errors if the repo doesn't exist",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo"}}`,
			expectedError:    errors.New("apprepositories.kubeapps.com \"test-repo\" not found"),
		},
		{
			name:             "it updates an app repository in the correct namespace",
			requestNamespace: "default",
			requestData:      `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
			existingRepos: map[string][]repoStub{
				"default": {repoStub{name: "test-repo"}},
			},
		},
		{
			name:             "it creates a secret if the auth header is set",
			requestNamespace: kubeappsNamespace,
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "test-repo"}},
			},
			requestData: `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it creates a secret if the auth header is set in different namespaces",
			requestNamespace: "default",
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "test-repo"}},
				"default":  {repoStub{name: "test-repo"}},
			},
			requestData: `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it updates a secret if the auth header is set",
			requestNamespace: kubeappsNamespace,
			existingRepos: map[string][]repoStub{
				"kubeapps": {repoStub{name: "test-repo"}},
			},
			existingSecrets: map[string][]secretStub{
				"kubeapps": {secretStub{name: "apprepo-test-repo"}},
			},
			requestData: `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it updates a secret if the auth header is set in both default and kubeapps namespace",
			requestNamespace: "default",
			existingRepos: map[string][]repoStub{
				"default": {repoStub{name: "test-repo"}},
			},
			existingSecrets: map[string][]secretStub{
				"kubeapps": {secretStub{name: "default-apprepo-test-repo"}},
				"default":  {secretStub{name: "apprepo-test-repo"}},
			},
			requestData: `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
		},
		{
			name:             "it updates a description for a repo",
			requestNamespace: "default",
			existingRepos: map[string][]repoStub{
				"default": {repoStub{name: "test-repo"}},
			},
			requestData: `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "description": "updated"}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...),
				fakecoreclientset.NewSimpleClientset(),
				&fakeRest.RESTClient{},
			}
			handler := userHandler{
				kubeappsNamespace: kubeappsNamespace,
				svcClientset:      cs,
				clientset:         cs,
			}

			apprepo, err := handler.UpdateAppRepository(io.NopCloser(strings.NewReader(tc.requestData)), tc.requestNamespace)
			checkErr(t, err, tc.expectedError)

			if apprepo != nil {
				appRepoRequest, expectedAppRepo, responseAppRepo := checkAppRepo(t, tc.requestData, tc.requestNamespace, cs)
				checkSecrets(t, tc.requestNamespace, appRepoRequest, expectedAppRepo, responseAppRepo, handler)
			}
		})
	}
}

func TestDeleteAppRepository(t *testing.T) {
	const kubeappsNamespace = "kubeapps"
	testCases := []struct {
		name              string
		repoName          string
		requestNamespace  string
		existingRepos     map[string][]repoStub
		expectedErrorCode int
	}{
		{
			name:             "it deletes an existing repo from a namespace",
			repoName:         "my-repo",
			requestNamespace: "my-namespace",
			existingRepos:    map[string][]repoStub{"my-namespace": {repoStub{name: "my-repo"}}},
		},
		{
			name:             "it deletes an existing repo with credentials from a namespace",
			repoName:         "my-repo",
			requestNamespace: "my-namespace",
			existingRepos:    map[string][]repoStub{"my-namespace": {repoStub{name: "my-repo", private: true}}},
		},
		{
			name:              "it returns not found when repo does not exist in specified namespace",
			repoName:          "my-repo",
			requestNamespace:  "other-namespace",
			existingRepos:     map[string][]repoStub{"my-namespace": {repoStub{name: "my-repo"}}},
			expectedErrorCode: 404,
		},
		{
			name:             "it deletes an existing repo from kubeapps' namespace",
			repoName:         "my-repo",
			requestNamespace: kubeappsNamespace,
			existingRepos:    map[string][]repoStub{kubeappsNamespace: {repoStub{name: "my-repo"}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...),
				fakecoreclientset.NewSimpleClientset(makeSecretsForRepos(tc.existingRepos, kubeappsNamespace)...),
				&fakeRest.RESTClient{},
			}
			handler := kubeHandler{
				clientsetForConfig:   func(*rest.Config) (combinedClientsetInterface, error) { return cs, nil },
				kubeappsNamespace:    kubeappsNamespace,
				kubeappsSvcClientset: cs,
				clustersConfig: ClustersConfig{
					KubeappsClusterName: "cluster",
					Clusters: map[string]ClusterConfig{
						"cluster": {
							Name:                     "cluster",
							APIServiceURL:            "fake",
							CertificateAuthorityData: "fake",
							CAFile:                   "",
							ServiceToken:             "fake",
							Insecure:                 true,
						},
					},
				},
			}

			cli, err := handler.AsSVC("cluster")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			err = cli.DeleteAppRepository(tc.repoName, tc.requestNamespace)

			if got, want := errorCodeForK8sError(t, err), tc.expectedErrorCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			if err == nil {
				// Ensure the repo has been deleted, so expecting a 404.
				_, err = cs.KubeappsV1alpha1().AppRepositories(tc.requestNamespace).Get(context.TODO(), tc.repoName, metav1.GetOptions{})
				if got, want := errorCodeForK8sError(t, err), 404; got != want {
					t.Errorf("got: %d, want: %d", got, want)
				}

				// We cannot ensure that the deletion of any owned secret was propagated
				// because the fake client does not handle finalizers but verified in real life.

				// Ensure any copy of the repo credentials has been deleted from the kubeapps namespace.
				_, err = cs.CoreV1().Secrets(kubeappsNamespace).Get(context.TODO(), KubeappsSecretNameForRepo(tc.repoName, tc.requestNamespace), metav1.GetOptions{})
				if got, want := errorCodeForK8sError(t, err), 404; got != want {
					t.Errorf("got: %d, want: %d", got, want)
				}
			}
		})
	}
}

func errorCodeForK8sError(t *testing.T, err error) int {
	if err == nil {
		return 0
	}
	if statusErr, ok := err.(*k8sErrors.StatusError); ok {
		return int(statusErr.ErrStatus.Code)
	}
	t.Fatalf("unable to convert error to status error")
	return 0
}

func TestAppRepositoryForRequest(t *testing.T) {
	testCases := []struct {
		name    string
		request appRepositoryRequestDetails
		appRepo v1alpha1.AppRepository
	}{
		{
			name: "it creates an app repo without auth",
			request: appRepositoryRequestDetails{
				Name:    "test-repo",
				Type:    "helm",
				RepoURL: "http://example.com/test-repo",
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:  "http://example.com/test-repo",
					Type: "helm",
				},
			},
		},
		{
			name: "it creates an app repo with auth header",
			request: appRepositoryRequestDetails{
				Name:       "test-repo",
				Type:       "helm",
				RepoURL:    "http://example.com/test-repo",
				AuthHeader: "testing",
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:  "http://example.com/test-repo",
					Type: "helm",
					Auth: v1alpha1.AppRepositoryAuth{
						Header: &v1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "apprepo-test-repo",
								},
								Key: "authorizationHeader",
							},
						},
					},
				},
			},
		},
		{
			name: "it creates an app repo with custom CA",
			request: appRepositoryRequestDetails{
				Name:     "test-repo",
				Type:     "helm",
				RepoURL:  "http://example.com/test-repo",
				CustomCA: "test-me",
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:  "http://example.com/test-repo",
					Type: "helm",
					Auth: v1alpha1.AppRepositoryAuth{
						CustomCA: &v1alpha1.AppRepositoryCustomCA{
							SecretKeyRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "apprepo-test-repo",
								},
								Key: "ca.crt",
							},
						},
					},
				},
			},
		},
		{
			name: "it creates an app repo with a sync job",
			request: appRepositoryRequestDetails{
				Name:    "test-repo",
				Type:    "helm",
				RepoURL: "http://example.com/test-repo",
				SyncJobPodTemplate: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-sync-job",
					},
				},
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:  "http://example.com/test-repo",
					Type: "helm",
					SyncJobPodTemplate: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-sync-job",
						},
					},
				},
			},
		},
		{
			name: "it creates an app repo with a resync requests",
			request: appRepositoryRequestDetails{
				Name:           "test-repo",
				Type:           "helm",
				RepoURL:        "http://example.com/test-repo",
				ResyncRequests: 99,
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:            "http://example.com/test-repo",
					Type:           "helm",
					ResyncRequests: 99,
				},
			},
		},
		{
			name: "it defaults type to helm",
			request: appRepositoryRequestDetails{
				Name:    "test-repo",
				RepoURL: "http://example.com/test-repo",
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:  "http://example.com/test-repo",
					Type: "helm",
				},
			},
		},
		{
			name: "it creates an OCI app repo",
			request: appRepositoryRequestDetails{
				Name:            "test-repo",
				Type:            "oci",
				RepoURL:         "http://example.com/test-repo",
				OCIRepositories: []string{"apache", "jenkins"},
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:             "http://example.com/test-repo",
					Type:            "oci",
					OCIRepositories: []string{"apache", "jenkins"},
				},
			},
		},
		{
			name: "it creates an app repo with a description",
			request: appRepositoryRequestDetails{
				Name:        "test-repo",
				Type:        "oci",
				RepoURL:     "http://example.com/test-repo",
				Description: "testing 1 2 3",
			},
			appRepo: v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:         "http://example.com/test-repo",
					Type:        "oci",
					Description: "testing 1 2 3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := appRepositoryForRequest(&appRepositoryRequest{tc.request}), &tc.appRepo; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestSecretForRequest(t *testing.T) {
	// Reuse the same app repo metadata for each test.
	appRepo := v1alpha1.AppRepository{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AppRepository",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			UID:       "abcd1234",
			Namespace: "repo-namespace",
		},
	}
	// And the same owner references expectation.
	blockOwnerDeletion := true
	ownerRefs := []metav1.OwnerReference{
		{
			APIVersion:         "kubeapps.com/v1alpha1",
			Kind:               "AppRepository",
			Name:               "test-repo",
			UID:                "abcd1234",
			BlockOwnerDeletion: &blockOwnerDeletion,
		},
	}
	dockerCredsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
	}

	testCases := []struct {
		name    string
		request appRepositoryRequestDetails
		secret  *corev1.Secret
	}{
		{
			name: "it does not create a secret without auth",
			request: appRepositoryRequestDetails{
				Name:    "test-repo",
				RepoURL: "http://example.com/test-repo",
			},
			secret: nil,
		},
		{
			name: "it creates a secret with an auth header",
			request: appRepositoryRequestDetails{
				Name:       "test-repo",
				RepoURL:    "http://example.com/test-repo",
				AuthHeader: "testing",
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "apprepo-test-repo",
					OwnerReferences: ownerRefs,
				},
				StringData: map[string]string{
					"authorizationHeader": "testing",
				},
			},
		},
		{
			name: "it creates a secret with custom CA",
			request: appRepositoryRequestDetails{
				Name:     "test-repo",
				RepoURL:  "http://example.com/test-repo",
				CustomCA: "test-me",
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "apprepo-test-repo",
					OwnerReferences: ownerRefs,
				},
				StringData: map[string]string{
					"ca.crt": "test-me",
				},
			},
		},
		{
			name: "uses the given secret for docker creds",
			request: appRepositoryRequestDetails{
				Name:         "test-repo",
				RepoURL:      "http://example.com/test-repo",
				AuthRegCreds: dockerCredsSecret.Name,
			},
			secret: dockerCredsSecret,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(),
				fakecoreclientset.NewSimpleClientset(dockerCredsSecret),
				&fakeRest.RESTClient{},
			}
			handler := userHandler{
				kubeappsNamespace: kubeappsNamespace,
				svcClientset:      cs,
				clientset:         cs,
			}

			secret, err := handler.secretForRequest(&appRepositoryRequest{tc.request}, &appRepo, "default")
			if err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !cmp.Equal(secret, tc.secret) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.secret, secret))
			}
		})
	}
}

func TestGetValidator(t *testing.T) {
	testCases := []struct {
		name              string
		appRepo           *v1alpha1.AppRepository
		expectedValidator HttpValidator
		expectedError     error
	}{
		{
			name: "it returns a validator with a request to the repo index.yaml",
			appRepo: &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL: "http://example.com/test-repo",
				},
			},
			expectedValidator: HelmNonOCIValidator{
				Req: &http.Request{
					Method: "GET",
					URL: &url.URL{
						Scheme: "http",
						Host:   "example.com",
						Path:   "/test-repo/index.yaml",
					},
					Header: http.Header{},
					Host:   "example.com",
				},
			},
		},
		{
			name: "it returns an OCI validator for an OCI repo",
			appRepo: &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:             "http://example.com/test-repo",
					Type:            "oci",
					OCIRepositories: []string{"apache", "jenkins"},
				},
			},
			expectedValidator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-repo",
					},
					Spec: v1alpha1.AppRepositorySpec{
						URL:             "http://example.com/test-repo",
						Type:            "oci",
						OCIRepositories: []string{"apache", "jenkins"},
					},
				},
			},
		},
		{
			name: "it returns an error for an OCI repo if no repositories are given",
			appRepo: &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-repo",
				},
				Spec: v1alpha1.AppRepositorySpec{
					URL:             "http://example.com/test-repo",
					Type:            "oci",
					OCIRepositories: []string{},
				},
			},
			expectedError: ErrEmptyOCIRegistry,
		},
	}

	cmpOpts := []cmp.Option{cmpopts.IgnoreUnexported(http.Request{})}
	cmpOpts = append(cmpOpts, cmpopts.IgnoreFields(http.Request{}, "Proto", "ProtoMajor", "ProtoMinor"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			httpValidator, err := getValidator(tc.appRepo)
			if got, want := err, tc.expectedError; got != want {
				t.Fatalf("got: %+v, want: %+v", err, tc.expectedError)
			} else if tc.expectedError != nil {
				return
			}

			if got, want := httpValidator, tc.expectedValidator; !cmp.Equal(got, want, cmpOpts...) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpOpts...))
			}
		})
	}
}

func TestNonOCIValidate(t *testing.T) {
	validRequest, err := http.NewRequest("GET", "http://example.com/index.yaml", strings.NewReader(""))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testCases := []struct {
		name             string
		httpValidator    HelmNonOCIValidator
		fakeHttpError    error
		fakeRepoResponse *http.Response
		expectedResponse *ValidationResponse
	}{
		{
			name:             "it returns 200 OK validation response if there is no error and the external response is 200",
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("OK")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it does not include the body of the upstream response when validation succeeds",
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("10 Mb of data")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it returns an error from the response with the body text if validation fails",
			fakeRepoResponse: &http.Response{StatusCode: 401, Body: io.NopCloser(bytes.NewReader([]byte("It failed because of X and Y")))},
			expectedResponse: &ValidationResponse{Code: 401, Message: "It failed because of X and Y"},
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
		},
		{
			name:             "it returns a 400 error if the validation cannot be run",
			fakeHttpError:    fmt.Errorf("client.Do returns an error"),
			expectedResponse: &ValidationResponse{Code: 400, Message: "client.Do returns an error"},
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
		},
	}

	cmpOpts := []cmp.Option{cmpopts.IgnoreUnexported(http.Request{}, strings.Reader{})}
	cmpOpts = append(cmpOpts, cmpopts.IgnoreFields(http.Request{}, "GetBody"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &fakeHTTPCli{
				response: tc.fakeRepoResponse,
				err:      tc.fakeHttpError,
			}

			response, err := tc.httpValidator.Validate(fakeClient)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if got, want := fakeClient.request, tc.httpValidator.Req; !cmp.Equal(want, got, cmpOpts...) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpOpts...))
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

type fakeOCIRepo struct {
	tags     repoTagsList
	manifest repoManifest
}

// makeTestOCIServer returns a small test double for an OCI server that handles requests
// for tags/list and manifest only.
func makeTestOCIServer(t *testing.T, registryName string, repos map[string]fakeOCIRepo, requiredAuthHeader string) *httptest.Server {
	// Define a map of valid request/responses based on the fake repos passed in.
	responses := map[string]string{}
	for repoName, repo := range repos {
		tags, err := json.Marshal(repo.tags)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		responses[path.Join("/v2", registryName, repoName, "tags", "list")] = string(tags)

		manifest, err := json.Marshal(repo.manifest)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(repo.tags.Tags) > 0 {
			responses[path.Join("/v2", registryName, repoName, "manifests", repo.tags.Tags[0])] = string(manifest)
		}
	}

	// Return a test server that responds with these canned responses only.
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Required authorization when set.
		authHeader := r.Header.Get("Authorization")
		if authHeader != requiredAuthHeader {
			w.WriteHeader(401)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
		if response, ok := responses[r.URL.Path]; !ok {
			w.WriteHeader(404)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		} else {
			_, err := w.Write([]byte(response))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
	}))
}

func TestOCIValidate(t *testing.T) {
	registryName := "bitnami"
	testCases := []struct {
		name             string
		repos            map[string]fakeOCIRepo
		validator        HelmOCIValidator
		expectedResponse *ValidationResponse
	}{
		{
			name: "it returns a valid response if all the OCI repos are of the helm type",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name: "it returns an invalid response if just one of OCI repos is of the wrong type",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.docker.container.image.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "nginx is not a Helm OCI Repo. mediaType starting with \"application/vnd.cncf.helm.config\" expected, found \"application/vnd.docker.container.image.v1+json\"",
			},
		},
		{
			name: "it returns an invalid response if a repo does not exist",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"notnginx": {},
			},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "Unexpected status code when querying \"nginx\": 404",
			},
		},
		{
			name: "it returns an invalid response if a manifest does not exist",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := makeTestOCIServer(t, registryName, tc.repos, "")
			defer ts.Close()
			// Use the test servers host/port as repo url.
			tc.validator.AppRepo.Spec.URL = fmt.Sprintf("%s/%s", ts.URL, registryName)

			response, err := tc.validator.Validate(httpclient.New())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestValidateAppRepository(t *testing.T) {
	registryName := "bitnami"
	testCases := []struct {
		name               string
		repos              map[string]fakeOCIRepo
		requiredAuthHeader string
		requestNamespace   string
		requestDetails     appRepositoryRequestDetails
		expectedResponse   *ValidationResponse
	}{
		{
			name: "it returns a valid response if all the OCI repos are of the helm type",
			requestDetails: appRepositoryRequestDetails{
				Name:            "testrepo",
				Type:            "oci",
				OCIRepositories: []string{"apache"},
			},
			requestNamespace: kubeappsNamespace,
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name:               "it returns a valid response for an authenticated OCI repo",
			requiredAuthHeader: "Bearer: some-token",
			requestDetails: appRepositoryRequestDetails{
				Name:            "testrepo",
				Type:            "oci",
				OCIRepositories: []string{"apache"},
				AuthHeader:      "Bearer: some-token",
			},
			requestNamespace: kubeappsNamespace,
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name:               "it returns a validation error for an authenticated OCI repo with incorrect credentials",
			requiredAuthHeader: "Bearer: some-token",
			requestDetails: appRepositoryRequestDetails{
				Name:            "testrepo",
				Type:            "oci",
				OCIRepositories: []string{"apache"},
				AuthHeader:      "Bearer: some-other-token",
			},
			requestNamespace: kubeappsNamespace,
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "Unexpected status code when querying \"apache\": 401",
			},
		},
		{
			name: "it returns a validation error if docker registry secrets included for a global repo",
			requestDetails: appRepositoryRequestDetails{
				Name:            "testrepo",
				Type:            "oci",
				OCIRepositories: []string{"apache"},
				RegistrySecrets: []string{"some-secret"},
			},
			requestNamespace: kubeappsNamespace,
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: ErrGlobalRepositoryWithSecrets.Error(),
			},
		},
		{
			name: "it returns a valid response if docker registry secrets included for a non-global repo",
			requestDetails: appRepositoryRequestDetails{
				Name:            "testrepo",
				Type:            "oci",
				OCIRepositories: []string{"apache"},
				RegistrySecrets: []string{"some-secret"},
			},
			requestNamespace: "other-namespace",
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := makeTestOCIServer(t, registryName, tc.repos, tc.requiredAuthHeader)
			defer ts.Close()
			// Use the test servers host/port as repo url.
			tc.requestDetails.RepoURL = fmt.Sprintf("%s/%s", ts.URL, registryName)
			appRepoJson, err := json.Marshal(appRepositoryRequest{tc.requestDetails})
			if err != nil {
				t.Fatalf("%+v", err)
			}

			handler := userHandler{kubeappsNamespace: kubeappsNamespace}
			response, err := handler.ValidateAppRepository(io.NopCloser(bytes.NewReader(appRepoJson)), tc.requestNamespace)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestNewClusterConfig(t *testing.T) {
	testCases := []struct {
		name            string
		userToken       string
		cluster         string
		clustersConfig  ClustersConfig
		inClusterConfig *rest.Config
		expectedConfig  *rest.Config
		errorExpected   bool
		maxReq          int
	}{
		{
			name:      "returns an in-cluster with explicit token for the default cluster",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
				},
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns an in-cluster config when no cluster is specified",
			userToken: "token-1",
			cluster:   "",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "",
				Clusters: map[string]ClusterConfig{
					"cluster-1": {
						APIServiceURL:                   "https://cluster-1.example.com:7890",
						CertificateAuthorityData:        "Y2EtZmlsZS1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-file-data",
						CAFile:                          "/tmp/ca-file-data",
					},
				},
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config setup for an additional cluster",
			userToken: "token-1",
			cluster:   "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
					"cluster-1": {
						APIServiceURL:                   "https://cluster-1.example.com:7890",
						CertificateAuthorityData:        "Y2EtZmlsZS1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-file-data",
						CAFile:                          "/tmp/ca-file-data",
					},
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://cluster-1.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
				TLSClientConfig: rest.TLSClientConfig{
					CAData: []byte("ca-file-data"),
					CAFile: "/tmp/ca-file-data",
				},
			},
		},
		{
			name:      "assumes a public cert if no ca data provided",
			userToken: "token-1",
			cluster:   "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {},
					"cluster-1": {
						APIServiceURL: "https://cluster-1.example.com:7890",
					},
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://cluster-1.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:            "returns an error if the cluster does not exist",
			cluster:         "cluster-1",
			inClusterConfig: &rest.Config{},
			errorExpected:   true,
		},
		{
			name:      "returns a config to proxy via pinniped-proxy",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "https://kubernetes.default",
						CertificateAuthorityData: "SGVsbG8K",
						PinnipedConfig:           PinnipedConciergeConfig{Enabled: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config to proxy via pinniped-proxy using the deprecated flag enable",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "https://kubernetes.default",
						CertificateAuthorityData: "SGVsbG8K",
						PinnipedConfig:           PinnipedConciergeConfig{Enable: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:      "returns a config to proxy via pinniped-proxy without headers for kubernetes.default",
			userToken: "token-1",
			cluster:   "default",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"default": {
						APIServiceURL:            "",
						CertificateAuthorityData: "",
						PinnipedConfig:           PinnipedConciergeConfig{Enabled: true},
					},
				},
				PinnipedProxyURL: "https://172.0.1.18:3333",
			},
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				Host:            "https://172.0.1.18:3333",
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := NewClusterConfig(tc.inClusterConfig, tc.userToken, tc.cluster, tc.clustersConfig)
			if got, want := err != nil, tc.errorExpected; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got, cmpopts.IgnoreFields(rest.Config{}, "WrapTransport")) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
			// If the test case defined a pinniped proxy url, verify that the expected headers
			// are added to the request.
			if clusterConfig, ok := tc.clustersConfig.Clusters[tc.cluster]; ok && clusterConfig.PinnipedConfig.Enabled {
				if config.WrapTransport == nil {
					t.Errorf("expected config.WrapTransport to be set but it is nil")
				} else {
					req := http.Request{}
					roundTripper := config.WrapTransport(&fakeRoundTripper{})
					_, err := roundTripper.RoundTrip(&req)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}
					want := http.Header{}
					if clusterConfig.APIServiceURL != "" {
						want["Pinniped_proxy_api_server_url"] = []string{clusterConfig.APIServiceURL}
					}
					if clusterConfig.CertificateAuthorityData != "" {

						want["Pinniped_proxy_api_server_cert"] = []string{clusterConfig.CertificateAuthorityData}
					}
					if got := req.Header; !cmp.Equal(want, got) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
					}
				}
			}
		})
	}
}

func TestParseClusterConfig(t *testing.T) {
	defaultPinnipedURL := "http://kubeapps-internal-pinniped-proxy.kubeapps:3333"
	testCases := []struct {
		name           string
		configJSON     string
		expectedErr    bool
		expectedConfig ClustersConfig
	}{
		{
			name:       "parses a single cluster",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedProxyURL": "http://172.0.1.18:3333", "isKubeappsCluster": true}]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
						IsKubeappsCluster:               true,
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses multiple clusters",
			configJSON: `[
	{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "isKubeappsCluster": true},
	{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						IsKubeappsCluster:               true,
					},
					"cluster-3": {
						Name:                            "cluster-3",
						APIServiceURL:                   "https://example.com/cluster-3",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK",
						CertificateAuthorityDataDecoded: "ca-cert-data-additional\n",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses a cluster without a service URL as the Kubeapps cluster",
			configJSON: `[
       {"name": "cluster-1" },
       {"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
       {"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-1",
				Clusters: map[string]ClusterConfig{
					"cluster-1": {
						Name: "cluster-1",
					},
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
					},
					"cluster-3": {
						Name:                            "cluster-3",
						APIServiceURL:                   "https://example.com/cluster-3",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK",
						CertificateAuthorityDataDecoded: "ca-cert-data-additional\n",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name: "parses config not specifying an explicit Kubeapps cluster",
			configJSON: `[
				{"name": "cluster-2", "apiServiceURL": "https://example.com/cluster-2", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg=="},
				{"name": "cluster-3", "apiServiceURL": "https://example.com/cluster-3", "certificateAuthorityData": "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK"}
			]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com/cluster-2",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
					},
					"cluster-3": {
						Name:                            "cluster-3",
						APIServiceURL:                   "https://example.com/cluster-3",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhLWFkZGl0aW9uYWwK",
						CertificateAuthorityDataDecoded: "ca-cert-data-additional\n",
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name:       "parses a cluster with pinniped token exchange",
			configJSON: `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "Y2EtY2VydC1kYXRhCg==", "serviceToken": "abcd", "pinnipedConfig": {"enabled": true}, "isKubeappsCluster": true}]`,
			expectedConfig: ClustersConfig{
				KubeappsClusterName: "cluster-2",
				Clusters: map[string]ClusterConfig{
					"cluster-2": {
						Name:                            "cluster-2",
						APIServiceURL:                   "https://example.com",
						CertificateAuthorityData:        "Y2EtY2VydC1kYXRhCg==",
						CertificateAuthorityDataDecoded: "ca-cert-data\n",
						ServiceToken:                    "abcd",
						PinnipedConfig: PinnipedConciergeConfig{
							Enabled: true,
						},
						IsKubeappsCluster: true,
					},
				},
				PinnipedProxyURL: "http://kubeapps-internal-pinniped-proxy.kubeapps:3333",
			},
		},
		{
			name:        "errors if the cluster configs cannot be parsed",
			configJSON:  `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "extracomma",}]`,
			expectedErr: true,
		},
		{
			name:        "errors if any CAData cannot be decoded",
			configJSON:  `[{"name": "cluster-2", "apiServiceURL": "https://example.com", "certificateAuthorityData": "not-base64-encoded"}]`,
			expectedErr: true,
		},
		{
			name: "errors if more than one cluster without an api service URL is configured",
			configJSON: `[
       {"name": "cluster-1" },
       {"name": "cluster-2" }
]`,
			expectedErr: true,
		},
		{
			name: "errors if more than one cluster with isKubeappsCluster=true is configured",
			configJSON: `[
		       {"name": "cluster-1", isKubeappsCluster: true},
		       {"name": "cluster-2", isKubeappsCluster: true }
		]`,
			expectedErr: true,
		},
		{
			name: "errors if both no APIServiceURL and isKubeappsCluster=true are configured",
			configJSON: `[
		       {"name": "cluster-1",  },
		       {"name": "cluster-2", isKubeappsCluster: true }
		]`,
			expectedErr: true,
		},
	}

	ignoreCAFile := cmpopts.IgnoreFields(ClusterConfig{}, "CAFile")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			path := createConfigFile(t, tc.configJSON)
			defer os.Remove(path)

			config, deferFn, err := ParseClusterConfig(path, "/tmp", defaultPinnipedURL, "")
			if got, want := err != nil, tc.expectedErr; got != want {
				t.Errorf("got: %t, want: %t: err: %+v", got, want, err)
			}
			if !tc.expectedErr {
				defer deferFn()
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got, ignoreCAFile) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreCAFile))
			}

			for clusterName, clusterConfig := range tc.expectedConfig.Clusters {
				if clusterConfig.CertificateAuthorityDataDecoded != "" {
					fileCAData, err := os.ReadFile(config.Clusters[clusterName].CAFile)
					if err != nil {
						t.Fatalf("error opening %s: %+v", config.Clusters[clusterName].CAFile, err)
					}
					if got, want := string(fileCAData), clusterConfig.CertificateAuthorityDataDecoded; got != want {
						t.Errorf("got: %q, want: %q", got, want)
					}
				}
			}
		})
	}
}

func createConfigFile(t *testing.T, content string) string {
	tmpfile, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("%+v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("%+v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("%+v", err)
	}
	return tmpfile.Name()
}
