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

package httphandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
)

type FakeHandler struct {
	appRepo    *v1alpha1.AppRepository
	namespaces []corev1.Namespace
	err        error
}

func (c *FakeHandler) CreateAppRepository(req *http.Request, namespace string) (*v1alpha1.AppRepository, error) {
	return c.appRepo, c.err
}

func (c *FakeHandler) DeleteAppRepository(req *http.Request, name, namespace string) error {
	return c.err
}

func (c *FakeHandler) GetNamespaces(req *http.Request) ([]corev1.Namespace, error) {
	return c.namespaces, c.err
}

func TestCreateAppRepository(t *testing.T) {
	testCases := []struct {
		name         string
		appRepo      *v1alpha1.AppRepository
		err          error
		expectedCode int
	}{
		{
			name:         "it should return the repo and a 200 if the repo is created",
			appRepo:      &v1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			expectedCode: 201,
		},
		{
			name:         "it should return a 404 if not found",
			err:          k8sErrors.NewNotFound(schema.GroupResource{}, "foo"),
			expectedCode: 404,
		},
		{
			name:         "it should return a 409 when conflict",
			err:          k8sErrors.NewConflict(schema.GroupResource{}, "foo", fmt.Errorf("already exists")),
			expectedCode: 409,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createAppFunc := CreateAppRepository(&FakeHandler{appRepo: tc.appRepo, namespaces: []corev1.Namespace{}, err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories", strings.NewReader("data"))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps"})

			response := httptest.NewRecorder()
			createAppFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 201 {
				var appRepoResponse appRepositoryResponse
				err := json.NewDecoder(response.Body).Decode(&appRepoResponse)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				expectedResponse := appRepositoryResponse{AppRepository: *tc.appRepo}
				if got, want := appRepoResponse, expectedResponse; !cmp.Equal(want, got) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
			}
		})
	}
}

func TestDeleteAppRepository(t *testing.T) {
	testCases := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "it should return a 200 if the repo is deleted",
			expectedCode: 200,
		},
		{
			name:         "it should return a 404 if not found",
			err:          k8sErrors.NewNotFound(schema.GroupResource{}, "foo"),
			expectedCode: 404,
		},
		{
			name:         "it should return a 403 when forbidden",
			err:          k8sErrors.NewForbidden(schema.GroupResource{}, "foo", fmt.Errorf("nope")),
			expectedCode: 403,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deleteAppFunc := DeleteAppRepository(&FakeHandler{appRepo: nil, namespaces: []corev1.Namespace{}, err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories", strings.NewReader("data"))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps"})

			response := httptest.NewRecorder()
			deleteAppFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}
		})
	}
}

func TestGetNamespaces(t *testing.T) {
	testCases := []struct {
		name         string
		namespaces   []corev1.Namespace
		err          error
		expectedCode int
	}{
		{
			name:         "it should return the list of namespaces and a 200 if the repo is created",
			namespaces:   []corev1.Namespace{corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode: 200,
		},
		{
			name:         "it should return a 403 when forbidden",
			err:          k8sErrors.NewForbidden(schema.GroupResource{}, "foo", fmt.Errorf("nope")),
			expectedCode: 403,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getNSFunc := GetNamespaces(&FakeHandler{appRepo: nil, namespaces: tc.namespaces, err: tc.err})
			req := httptest.NewRequest("GET", "https://foo.bar/backend/v1/namespaces", nil)

			response := httptest.NewRecorder()
			getNSFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 200 {
				var nsResponse namespacesResponse
				err := json.NewDecoder(response.Body).Decode(&nsResponse)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				expectedResponse := namespacesResponse{Namespaces: tc.namespaces}
				if got, want := nsResponse, expectedResponse; !cmp.Equal(want, got) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
				}
			}
		})
	}
}
