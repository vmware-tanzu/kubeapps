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

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	fakecoreclientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

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
		kubeappsNamespace string
		// existingRepos is a map with the namespaces as the key
		// and a slice of repository names for that namespace as the value.
		existingRepos map[string][]string
		requestData   string
		expectedCode  int
	}{
		{
			name:              "it creates an app repository",
			kubeappsNamespace: "kubeapps",
			requestData:       `{"appRepository": {"name": "test-repo", "url": "http://example.com/test-repo"}}`,
			expectedCode:      http.StatusCreated,
		},
		{
			name:              "it errors if the repo exists in the kubeapps ns already",
			kubeappsNamespace: "kubeapps",
			requestData:       `{"appRepository": {"name": "bitnami"}}`,
			existingRepos: map[string][]string{
				"kubeapps": []string{"bitnami"},
			},
			expectedCode: http.StatusConflict,
		},
		{
			name:              "it creates the repo even if the same repo exists in other namespaces",
			kubeappsNamespace: "kubeapps",
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
			requestData:       `not a { json object`,
			expectedCode:      http.StatusBadRequest,
		},
		{
			name:              "it results in an Unauthorized response if the kubeapps namespace is not set",
			requestData:       `{"appRepository": {"name": "bitnami"}}`,
			kubeappsNamespace: "",
			expectedCode:      http.StatusUnauthorized,
		},
		{
			name:              "it creates a secret if the auth header is set",
			kubeappsNamespace: "kubeapps",
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
			}

			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/apprepositories", strings.NewReader(tc.requestData))

			response := httptest.NewRecorder()

			handler.Create(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			if response.Code == 201 {
				var appRepoRequest appRepositoryRequest
				err := json.NewDecoder(strings.NewReader(tc.requestData)).Decode(&appRepoRequest)
				if err != nil {
					t.Fatalf("%+v", err)
				}

				// Ensure the expected AppRepository is stored
				requestAppRepo := appRepositoryForRequest(&appRepoRequest)
				requestAppRepo.ObjectMeta.Namespace = tc.kubeappsNamespace

				responseAppRepo, err := cs.KubeappsV1alpha1().AppRepositories(tc.kubeappsNamespace).Get(requestAppRepo.ObjectMeta.Name, metav1.GetOptions{})
				if err != nil {
					t.Errorf("expected data %v not present: %+v", requestAppRepo, err)
				}

				if got, want := responseAppRepo, requestAppRepo; !cmp.Equal(want, got) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}

				// When appropriate, ensure the expected secret is stored
				if appRepoRequest.AppRepository.AuthHeader != "" {
					requestSecret := secretForRequest(appRepoRequest, *responseAppRepo)
					requestSecret.ObjectMeta.Namespace = tc.kubeappsNamespace

					responseSecret, err := cs.CoreV1().Secrets(tc.kubeappsNamespace).Get(requestSecret.ObjectMeta.Name, metav1.GetOptions{})
					if err != nil {
						t.Errorf("expected data %v not present: %+v", requestSecret, err)
					}

					if got, want := responseSecret, requestSecret; !cmp.Equal(want, got) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
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
			name: "it creates an app repo witha resync requests",
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
			Name: "test-repo",
			UID:  "abcd1234",
		},
	}
	// And the same owner references expectation.
	blockOwnerDeletion := true
	ownerRefs := []metav1.OwnerReference{
		metav1.OwnerReference{
			APIVersion:         "v1",
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
			name: "it creates a nil secret without auth",
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
			if got, want := secretForRequest(appRepositoryRequest{tc.request}, appRepo), tc.secret; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
