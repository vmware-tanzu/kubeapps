/*
Copyright (c) 2019 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakecoreclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	fakeRest "k8s.io/client-go/rest/fake"
	k8stesting "k8s.io/client-go/testing"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	fakeapprepoclientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/fake"
)

type repoStub struct {
	name    string
	private bool
}

type secretStub struct {
	name string
}

type fakeHTTPCli struct {
	response *http.Response
	err      error
}

func (f *fakeHTTPCli) Do(*http.Request) (*http.Response, error) {
	return f.response, f.err
}

const kubeappsNamespace = "kubeapps"

func makeAppRepoObjects(reposPerNamespace map[string][]repoStub) []runtime.Object {
	objects := []runtime.Object{}
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
			objects = append(objects, runtime.Object(appRepo))
		}
	}
	return objects
}

func makeSecretObjects(secretsPerNamespace map[string][]secretStub) []runtime.Object {
	objects := []runtime.Object{}
	for namespace, secretsStubs := range secretsPerNamespace {
		for _, secretStub := range secretsStubs {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretStub.name,
					Namespace: namespace,
				},
			}
			objects = append(objects, runtime.Object(secret))
		}
	}
	return objects
}

func makeSecretsForRepos(reposPerNamespace map[string][]repoStub, kubeappsNamespace string) []runtime.Object {
	objects := []runtime.Object{}
	for namespace, repoStubs := range reposPerNamespace {
		for _, repoStub := range repoStubs {
			// Only create secrets if it's a private repo.
			if !repoStub.private {
				continue
			}
			var appRepo runtime.Object = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      secretNameForRepo(repoStub.name),
					Namespace: namespace,
				},
			}
			objects = append(objects, appRepo)

			// Only create a copy of the secret in the kubeapps namespace if the app repo
			// is in a user namespace.
			if namespace != kubeappsNamespace {
				var appRepo runtime.Object = &corev1.Secret{
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
	// responseAppRepo.ObjectMeta.UID = "dead-beef"

	// When appropriate, ensure the expected secret is stored.
	if appRepoRequest.AppRepository.AuthHeader != "" {
		expectedSecret := secretForRequest(&appRepoRequest, responseAppRepo)
		expectedSecret.ObjectMeta.Namespace = requestNamespace
		responseSecret, err := handler.clientset.CoreV1().Secrets(requestNamespace).Get(context.TODO(), expectedSecret.ObjectMeta.Name, metav1.GetOptions{})

		if err != nil {
			t.Errorf("expected data %v not present: %+v", expectedSecret, err)
		}

		if got, want := responseSecret, expectedSecret; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		// Verify the copy of the repo secret in in kubeapps is
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

			apprepo, err := handler.CreateAppRepository(ioutil.NopCloser(strings.NewReader(tc.requestData)), tc.requestNamespace)
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

			apprepo, err := handler.UpdateAppRepository(ioutil.NopCloser(strings.NewReader(tc.requestData)), tc.requestNamespace)
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
				clientsetForConfig: func(*rest.Config) (combinedClientsetInterface, error) { return cs, nil },
				kubeappsNamespace:  kubeappsNamespace,
				svcClientset:       cs,
			}

			err := handler.AsSVC().DeleteAppRepository(tc.repoName, tc.requestNamespace)

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := secretForRequest(&appRepositoryRequest{tc.request}, &appRepo), tc.secret; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetNamespaces(t *testing.T) {
	testCases := []struct {
		name               string
		existingNamespaces []string
		allowed            bool
		userClientErr      error
		svcClientErr       error
		expectedNamespaces []string
	}{
		{
			name:               "it lists namespaces if the user client returns the namespaces",
			existingNamespaces: []string{"foo", "bar", "zed"},
			expectedNamespaces: []string{"foo", "bar", "zed"},
			allowed:            true,
		},
		{
			name:               "it does not filter the namespaces if returned by the user client",
			existingNamespaces: []string{"foo", "bar", "zed"},
			expectedNamespaces: []string{"foo", "bar", "zed"},
			allowed:            false,
		},
		{
			name:               "it lists namespaces if the userclient fails but the service client succeeds",
			existingNamespaces: []string{"foo"},
			userClientErr:      k8sErrors.NewForbidden(schema.GroupResource{}, "bang", fmt.Errorf("Bang")),
			expectedNamespaces: []string{"foo"},
			allowed:            true,
		},
		{
			name:               "it filters the namespaces if the userclient fails but the service client succeeds",
			existingNamespaces: []string{"foo"},
			userClientErr:      k8sErrors.NewForbidden(schema.GroupResource{}, "bang", fmt.Errorf("Bang")),
			expectedNamespaces: []string{},
			allowed:            false,
		},
		{
			name:               "it returns an empty list if both the user and service account forbidden",
			existingNamespaces: []string{"foo"},
			userClientErr:      k8sErrors.NewForbidden(schema.GroupResource{}, "bang", fmt.Errorf("Bang")),
			svcClientErr:       k8sErrors.NewForbidden(schema.GroupResource{}, "bang", fmt.Errorf("Bang")),
			expectedNamespaces: []string{},
			allowed:            true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userClientSet := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(),
				fakecoreclientset.NewSimpleClientset(),
				&fakeRest.RESTClient{},
			}

			svcClientSet := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(),
				fakecoreclientset.NewSimpleClientset(),
				&fakeRest.RESTClient{},
			}

			setClientsetData(userClientSet, tc.existingNamespaces, tc.userClientErr)
			setClientsetData(svcClientSet, tc.existingNamespaces, tc.svcClientErr)

			// Set whether the userClientSet is allowed to create self subject access reviews.
			// The handler for the reactor has only the action as input, which does not convey
			// enough info to be able to decide whether an individual namespace is allowed
			// (as action.GetNamespace() is always empty as self subject access reviews are
			// *not* themselves namespaced - the spec includes the namespace but is not contained
			// in the action). As a result, we can only filter everything or nothing in tests.
			userClientSet.Clientset.Fake.PrependReactor(
				"create",
				"selfsubjectaccessreviews",
				func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					mysar := &authorizationv1.SelfSubjectAccessReview{
						Status: authorizationv1.SubjectAccessReviewStatus{
							Allowed: tc.allowed,
							Reason:  "I want to test it",
						},
					}
					return true, mysar, nil
				},
			)

			handler := kubeHandler{
				clientsetForConfig: func(*rest.Config) (combinedClientsetInterface, error) { return userClientSet, nil },
				kubeappsNamespace:  "kubeapps",
				svcClientset:       svcClientSet,
				clustersConfig:     ClustersConfig{KubeappsClusterName: "default"},
			}

			userHandler, err := handler.AsUser("token", "default")
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			namespaces, err := userHandler.GetNamespaces()
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			namespaceNames := []string{}
			for _, ns := range namespaces {
				namespaceNames = append(namespaceNames, ns.ObjectMeta.Name)
			}
			if !cmp.Equal(namespaceNames, tc.expectedNamespaces) {
				t.Errorf("Unexpected response: %s", cmp.Diff(namespaces, tc.expectedNamespaces))
			}
		})
	}
}

// setClientsetData configures the fake clientset with the return and error.
func setClientsetData(cs fakeCombinedClientset, namespaceNames []string, err error) {
	namespaces := []corev1.Namespace{}
	for _, ns := range namespaceNames {
		namespaces = append(namespaces, corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		})
	}
	cs.Clientset.Fake.PrependReactor(
		"list",
		"namespaces",
		func(action k8stesting.Action) (bool, runtime.Object, error) {
			return true, &corev1.NamespaceList{Items: namespaces}, err
		},
	)
}

func TestValidateAppRepository(t *testing.T) {
	const kubeappsNamespace = "kubeapps"
	getValidationCliAndReqTests := []struct {
		name             string
		requestData      string
		requestNamespace string
		expectedURL      string
		expectedHeaders  http.Header
		expectedError    error
	}{
		{
			name:             "it parses the repo URL",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "repoURL": "http://example.com/test-repo"}}`,
			expectedURL:      "http://example.com/test-repo/index.yaml",
		},
		{
			name:             "it includes the auth creds",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "repoURL": "http://example.com/test-repo", "authHeader": "test-me"}}`,
			expectedURL:      "http://example.com/test-repo/index.yaml",
			expectedHeaders:  http.Header{"Authorization": []string{"test-me"}},
		},
		{
			name:             "validation fails if docker registry secrets included for a global repo",
			requestNamespace: kubeappsNamespace,
			requestData:      `{"appRepository": {"name": "test-repo", "repoURL": "http://example.com/test-repo", "registrySecrets": ["secret-1"]}}`,
			expectedError:    ErrGlobalRepositoryWithSecrets,
		},
	}

	for _, tc := range getValidationCliAndReqTests {
		t.Run(tc.name, func(t *testing.T) {
			cli, req, err := getValidationCliAndReq(ioutil.NopCloser(strings.NewReader(tc.requestData)), tc.requestNamespace, kubeappsNamespace)
			if (err != nil || tc.expectedError != nil) && !errors.Is(err, tc.expectedError) {
				t.Fatalf("got: %+v, want: %+v", err, tc.expectedError)
			}
			if tc.expectedError != nil {
				return
			}
			if tc.expectedURL != req.URL.String() {
				t.Errorf("Expected %v got %v", tc.expectedURL, req.URL.String())
			}
			if tc.expectedHeaders != nil && !cmp.Equal(tc.expectedHeaders, cli.(*clientWithDefaultHeaders).defaultHeaders) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.expectedHeaders, cli.(*clientWithDefaultHeaders).defaultHeaders))
			}
		})
	}

	doValidationRequestTests := []struct {
		name           string
		err            error
		response       *http.Response
		expectedResult *ValidationResponse
	}{
		{
			name:           "returns nil if there is no error and the response is okay",
			err:            nil,
			response:       &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte("OK")))},
			expectedResult: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:           "returns an error",
			err:            fmt.Errorf("Boom"),
			response:       &http.Response{},
			expectedResult: &ValidationResponse{Code: 400, Message: "Boom"},
		},
		{
			name:           "returns an error from the response",
			err:            nil,
			response:       &http.Response{StatusCode: 401, Body: ioutil.NopCloser(bytes.NewReader([]byte("Boom")))},
			expectedResult: &ValidationResponse{Code: 401, Message: "Boom"},
		},
	}
	for _, tc := range doValidationRequestTests {
		t.Run(tc.name, func(t *testing.T) {
			cli := &fakeHTTPCli{
				response: tc.response,
				err:      tc.err,
			}
			got, err := doValidationRequest(cli, &http.Request{})
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if want := tc.expectedResult; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestNewClusterConfig(t *testing.T) {
	testCases := []struct {
		name            string
		token           string
		cluster         string
		clustersConfig  ClustersConfig
		inClusterConfig *rest.Config
		expectedConfig  *rest.Config
		errorExpected   bool
	}{
		{
			name:           "returns an in-cluster with explicit token for the default cluster",
			token:          "token-1",
			cluster:        "default",
			clustersConfig: ClustersConfig{KubeappsClusterName: "default"},
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
			name:    "returns a config setup for an additional cluster",
			token:   "token-1",
			cluster: "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
					"cluster-1": {
						APIServiceURL:            "https://cluster-1.example.com:7890",
						CertificateAuthorityData: "ca-file-data",
						CAFile:                   "/tmp/ca-file-data",
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
			name:    "assumes a public cert if no ca data provided",
			token:   "token-1",
			cluster: "cluster-1",
			clustersConfig: ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]ClusterConfig{
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := NewClusterConfig(tc.inClusterConfig, tc.token, tc.cluster, tc.clustersConfig)
			if got, want := err != nil, tc.errorExpected; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
