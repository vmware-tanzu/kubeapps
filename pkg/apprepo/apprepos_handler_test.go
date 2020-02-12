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

package apprepo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakecoreclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	fakeapprepoclientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/fake"
)

func makeAppRepoObjects(repoNamesPerNamespace map[string][]string) []runtime.Object {
	objects := []runtime.Object{}
	for namespace, repoNames := range repoNamesPerNamespace {
		for _, repoName := range repoNames {
			var appRepo runtime.Object = &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      repoName,
					Namespace: namespace,
				},
			}
			objects = append(objects, appRepo)
		}
	}
	return objects
}

type fakeAppRepoClientset = fakeapprepoclientset.Clientset
type fakeCombinedClientset struct {
	*fakeAppRepoClientset
	*fakecoreclientset.Clientset
}

// Not sure why golang thinks this Discovery() is ambiguous on the fake but not on
// the real combinedClientset, but to satisfy:
func (f fakeCombinedClientset) Discovery() discovery.DiscoveryInterface {
	return f.Clientset.Discovery()
}

func TestAppRepositoryCreate(t *testing.T) {
	testCases := []struct {
		name              string
		requestNamespace  string
		kubeappsNamespace string
		// existingRepos is a map with the namespaces as the key
		// and a slice of repository names for that namespace as the value.
		existingRepos map[string][]string
		requestData   string
		expectedCode  int
	}{
		{
			name:              "it creates an app repository in the default kubeappsNamespace",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
			expectedCode:      http.StatusCreated,
		},
		{
			name:              "it creates an app repository in a specific namespace",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "my-namespace",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
			expectedCode:      http.StatusCreated,
		},
		{
			name:              "it creates an app repository with an empty template",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "syncJobPodTemplate": {}}}`,
			expectedCode:      http.StatusCreated,
		},
		{
			name:              "it errors if the repo exists in the kubeapps ns already",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `{"appRepository": {"name": "bitnami"}}`,
			existingRepos: map[string][]string{
				"kubeapps": []string{"bitnami"},
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:              "it creates the repo even if the same repo exists in other namespaces",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `{"appRepository": {"name": "bitnami"}}`,
			existingRepos: map[string][]string{
				"kubeapps-other-ns-1": []string{"bitnami"},
				"kubeapps-other-ns-2": []string{"bitnami"},
			},
			expectedCode: http.StatusCreated,
		},
		{
			name:              "it results in a bad request if the json cannot be parsed",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `not a { json object`,
			expectedCode:      http.StatusBadRequest,
		},
		{
			name:              "it creates a secret if the auth header is set",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "kubeapps",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
			expectedCode:      http.StatusCreated,
		},
		{
			name:              "it creates a copy of the namespaced repo secret in the kubeapps namespace",
			kubeappsNamespace: "kubeapps",
			requestNamespace:  "test-namespace",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo", "authHeader": "test-me"}}`,
			expectedCode:      http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...),
				fakecoreclientset.NewSimpleClientset(),
			}
			handler := appRepositoriesHandler{
				clientsetForConfig: func(*rest.Config) (combinedClientsetInterface, error) { return cs, nil },
				kubeappsNamespace:  tc.kubeappsNamespace,
				svcKubeClient:      fakecoreclientset.NewSimpleClientset(),
			}

			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories", strings.NewReader(tc.requestData))
			req = mux.SetURLVars(req, map[string]string{"namespace": tc.requestNamespace})

			response := httptest.NewRecorder()

			handler.Create(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 201 {
				var appRepoRequest appRepositoryRequest
				err := json.NewDecoder(strings.NewReader(tc.requestData)).Decode(&appRepoRequest)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				// Ensure the expected AppRepository is stored
				expectedAppRepo := appRepositoryForRequest(appRepoRequest)
				expectedAppRepo.ObjectMeta.Namespace = tc.requestNamespace

				responseAppRepo, err := cs.KubeappsV1alpha1().AppRepositories(tc.requestNamespace).Get(expectedAppRepo.ObjectMeta.Name, metav1.GetOptions{})
				if err != nil {
					t.Fatalf("expected data %v not present: %+v", expectedAppRepo, err)
				}

				if got, want := responseAppRepo, expectedAppRepo; !cmp.Equal(want, got) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}

				// Ensure the response contained the created app repository
				var appRepoResponse appRepositoryResponse
				err = json.NewDecoder(response.Body).Decode(&appRepoResponse)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				expectedResponse := appRepositoryResponse{AppRepository: *expectedAppRepo}
				if got, want := appRepoResponse, expectedResponse; !cmp.Equal(want, got) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}

				// When appropriate, ensure the expected secret is stored.
				if appRepoRequest.AppRepository.AuthHeader != "" {
					expectedSecret := secretForRequest(appRepoRequest, responseAppRepo)
					expectedSecret.ObjectMeta.Namespace = tc.requestNamespace
					responseSecret, err := cs.CoreV1().Secrets(tc.requestNamespace).Get(expectedSecret.ObjectMeta.Name, metav1.GetOptions{})

					if err != nil {
						t.Errorf("expected data %v not present: %+v", expectedSecret, err)
					}

					if got, want := responseSecret, expectedSecret; !cmp.Equal(want, got) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
					}

					// Verify the copy of the repo secret in in kubeapps is
					// also stored if this is a per-namespace app repository.
					kubeappsSecretName := kubeappsSecretNameForRepo(expectedAppRepo.ObjectMeta.Name, expectedAppRepo.ObjectMeta.Namespace)
					expectedSecret.ObjectMeta.Name = kubeappsSecretName
					expectedSecret.ObjectMeta.Namespace = tc.kubeappsNamespace

					if tc.requestNamespace != tc.kubeappsNamespace {
						responseSecret, err = handler.svcKubeClient.CoreV1().Secrets(tc.kubeappsNamespace).Get(kubeappsSecretName, metav1.GetOptions{})
						if err != nil {
							t.Errorf("expected data %v not present: %+v", expectedSecret, err)
						}

						if got, want := responseSecret, expectedSecret; !cmp.Equal(want, got) {
							t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
						}
					} else {
						// The copy of the secret should not be created when the request namespace is kubeapps.
						secret, err := handler.svcKubeClient.CoreV1().Secrets(tc.kubeappsNamespace).Get(kubeappsSecretName, metav1.GetOptions{})
						if err == nil {
							t.Fatalf("secret should not be created, found %+v", secret)
						}
						if statusErr, ok := err.(*errors.StatusError); ok {
							status := statusErr.ErrStatus
							if got, want := status.Code, int32(404); got != want {
								t.Errorf("got: %d, want: %d", got, want)
							}
						} else {
							t.Errorf("Unable to convert err to StatusError: %+v", err)
						}
					}
				}
			}
		})
	}
}

func TestConfigForToken(t *testing.T) {
	handler := appRepositoriesHandler{
		config: rest.Config{},
	}
	token := "abcd"

	configWithToken := handler.ConfigForToken(token)

	// The returned config has the token set.
	if got, want := configWithToken.BearerToken, token; got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}

	// The handler config's BearerToken is still blank.
	if got, want := handler.config.BearerToken, ""; got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
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
									Name: "apprepo-test-repo-secrets",
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
									Name: "apprepo-test-repo-secrets",
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
			if got, want := appRepositoryForRequest(appRepositoryRequest{tc.request}), &tc.appRepo; !cmp.Equal(want, got) {
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
		metav1.OwnerReference{
			APIVersion:         "kubeapps.com/v1alpha1",
			Kind:               "AppRepository",
			Name:               "test-repo",
			UID:                "abcd1234",
			BlockOwnerDeletion: &blockOwnerDeletion,
		},
	}

	testCases := []struct {
		name             string
		requestNamespace string
		request          appRepositoryRequestDetails
		secret           *corev1.Secret
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
					Name:            "apprepo-test-repo-secrets",
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
					Name:            "apprepo-test-repo-secrets",
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
			if got, want := secretForRequest(appRepositoryRequest{tc.request}, &appRepo), tc.secret; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestGetNamespaces(t *testing.T) {
	testCases := []struct {
		name              string
		kubeappsNamespace string
		// existingRepos is a map with the namespaces as the key
		// and a slice of repository names for that namespace as the value.
		existingNS       []string
		expectedResponse []corev1.Namespace
		allowed          bool
	}{
		{
			name:              "it list namespaces",
			kubeappsNamespace: "kubeapps",
			existingNS:        []string{"foo"},
			expectedResponse: []corev1.Namespace{
				corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foo",
					},
				},
			},
			allowed: true,
		},
		{
			name:              "it returns an empty list if not allowed",
			kubeappsNamespace: "kubeapps",
			existingNS:        []string{"foo"},
			expectedResponse:  []corev1.Namespace{},
			allowed:           false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeCombinedClientset{
				fakeapprepoclientset.NewSimpleClientset(),
				fakecoreclientset.NewSimpleClientset(),
			}

			for _, ns := range tc.existingNS {
				cs.Clientset.CoreV1().Namespaces().Create(&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: ns,
					},
				})
			}

			cs.Clientset.Fake.PrependReactor(
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

			handler := appRepositoriesHandler{
				clientsetForConfig: func(*rest.Config) (combinedClientsetInterface, error) { return cs, nil },
				kubeappsNamespace:  tc.kubeappsNamespace,
			}

			req := httptest.NewRequest("GET", "https://foo.bar/backend/v1/namespaces", nil)

			response := httptest.NewRecorder()

			handler.GetNamespaces(response, req)

			var responseNS []corev1.Namespace
			err := json.NewDecoder(response.Body).Decode(&responseNS)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if !cmp.Equal(responseNS, tc.expectedResponse) {
				t.Errorf("Unexpected response: %s", cmp.Diff(responseNS, tc.expectedResponse))
			}
		})
	}
}
