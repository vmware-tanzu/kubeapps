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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	clientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned"
	fakeclientset "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/fake"
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cs := fakeclientset.NewSimpleClientset(makeAppRepoObjects(tc.existingRepos)...)
			handler := appRepositoriesHandler{
				clientsetForConfig: func(*rest.Config) (clientset.Interface, error) { return cs, nil },
				kubeappsNamespace:  tc.kubeappsNamespace,
			}

			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/apprepositories", strings.NewReader(tc.requestData))

			response := httptest.NewRecorder()

			handler.Create(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			if response.Code == 201 {
				assertRepoPresent(t, tc.kubeappsNamespace, tc.requestData, cs)
			}
		})
	}
}

func assertRepoPresent(t *testing.T, namespace, requestData string, cs clientset.Interface) {
	requestAppRepo, err := appRepositoryForRequestData(ioutil.NopCloser(strings.NewReader(requestData)))
	if err != nil {
		t.Fatalf("%+v", err)
	}
	requestAppRepo.ObjectMeta.Namespace = namespace

	responseAppRepo, err := cs.KubeappsV1alpha1().AppRepositories(namespace).Get(requestAppRepo.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("expected data %v not present: %+v", requestAppRepo, err)
	}

	if got, want := responseAppRepo, requestAppRepo; !cmp.Equal(want, got) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
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
