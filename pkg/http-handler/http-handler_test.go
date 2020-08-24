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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/kubeapps/kubeapps/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
)

func checkAppResponse(t *testing.T, response *httptest.ResponseRecorder, expectedRepo *v1alpha1.AppRepository) {
	var appRepoResponse appRepositoryResponse
	err := json.NewDecoder(response.Body).Decode(&appRepoResponse)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	expectedResponse := appRepositoryResponse{AppRepository: *expectedRepo}
	if got, want := appRepoResponse, expectedResponse; !cmp.Equal(want, got) {
		t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
	}
}

func checkError(t *testing.T, response *httptest.ResponseRecorder, expectedError error) {
	if response.Code == 500 {
		// If the error is a 500 we simply retunr a string (encoded in JSON)
		var errMsg string
		err := json.NewDecoder(response.Body).Decode(&errMsg)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if got, want := errMsg, expectedError.Error(); got != want {
			t.Errorf("got: %q, want: %q", got, want)
		}
	} else {
		// The error should be a kubernetes error response.
		var status metav1.Status
		err := json.NewDecoder(response.Body).Decode(&status)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if got, want := status, expectedError.(*k8sErrors.StatusError).ErrStatus; !cmp.Equal(want, got) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	}
}

func TestCreateListAppRepositories(t *testing.T) {
	testCases := []struct {
		name         string
		appRepos     []*v1alpha1.AppRepository
		err          error
		expectedCode int
	}{
		{
			name:         "it should return the list of repos",
			expectedCode: 200,
		},
		{
			name:         "it should return an error",
			err:          fmt.Errorf("boom"),
			expectedCode: 500,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			listFunc := ListAppRepositories(&kube.FakeHandler{AppRepos: []*v1alpha1.AppRepository{
				{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			}, Err: tc.err})
			req := httptest.NewRequest("GET", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories", strings.NewReader("data"))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps"})

			response := httptest.NewRecorder()
			listFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code != 200 {
				checkError(t, response, tc.err)
			}
		})
	}
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
		{
			name:         "it returns a json 500 error as a plain string for internal backend errors",
			err:          fmt.Errorf("bang"),
			expectedCode: 500,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createAppFunc := CreateAppRepository(&kube.FakeHandler{CreatedRepo: tc.appRepo, Err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories", strings.NewReader("data"))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps"})

			response := httptest.NewRecorder()
			createAppFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 201 {
				checkAppResponse(t, response, tc.appRepo)
			} else {
				checkError(t, response, tc.err)
			}
		})
	}
}

func TestUpdateAppRepository(t *testing.T) {
	testCases := []struct {
		name         string
		appRepo      *v1alpha1.AppRepository
		err          error
		expectedCode int
	}{
		{
			name:         "it should return the repo and a 200 if the repo is updated",
			appRepo:      &v1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "foo"}},
			expectedCode: 200,
		},
		{
			name:         "it should return a 404 if not found",
			err:          k8sErrors.NewNotFound(schema.GroupResource{}, "foo"),
			expectedCode: 404,
		},
		{
			name:         "it returns a json 500 error as a plain string for internal backend errors",
			err:          fmt.Errorf("bang"),
			expectedCode: 500,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createAppFunc := UpdateAppRepository(&kube.FakeHandler{UpdatedRepo: tc.appRepo, Err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories/foo", strings.NewReader("data"))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps"})

			response := httptest.NewRecorder()
			createAppFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 200 {
				checkAppResponse(t, response, tc.appRepo)
			} else {
				checkError(t, response, tc.err)
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
			deleteAppFunc := DeleteAppRepository(&kube.FakeHandler{Err: tc.err})
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
			namespaces:   []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
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
			getNSFunc := GetNamespaces(&kube.FakeHandler{Namespaces: tc.namespaces, Err: tc.err})
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

func TestValidateAppRepository(t *testing.T) {
	testCases := []struct {
		name               string
		err                error
		validationResponse kube.ValidationResponse
		expectedCode       int
		expectedBody       string
	}{
		{
			name:               "it should return OK if no error is detected",
			validationResponse: kube.ValidationResponse{Code: 200, Message: "OK"},
			expectedCode:       200,
			expectedBody:       `{"code":200,"message":"OK"}`,
		},
		{
			name:               "it should return the error code if given",
			err:                fmt.Errorf("Boom"),
			validationResponse: kube.ValidationResponse{},
			expectedCode:       500,
			expectedBody:       "\"Boom\"\n",
		},
		{
			name:               "it should return an error in the validation response",
			validationResponse: kube.ValidationResponse{Code: 401, Message: "Forbidden"},
			expectedCode:       200,
			expectedBody:       `{"code":401,"message":"Forbidden"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validateAppRepoFunc := ValidateAppRepository(&kube.FakeHandler{ValRes: &tc.validationResponse, Err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories/validate", strings.NewReader("data"))

			response := httptest.NewRecorder()
			validateAppRepoFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			responseBody, _ := ioutil.ReadAll(response.Body)
			if got, want := string(responseBody), tc.expectedBody; got != want {
				t.Errorf("got: %s, want: %s\n", got, want)
			}
		})
	}
}

func TestGetOperatorLogo(t *testing.T) {
	testCases := []struct {
		name                string
		logo                []byte
		expectedContentType string
		err                 error
	}{
		{
			name:                "it should return a SVG logo",
			logo:                []byte("<svg viewBox=\"0 0 658 270\"></svg>"),
			expectedContentType: "image/svg+xml",
		},
		// TODO(andresmgot): Add test for PNG scenario
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getOpLogo := GetOperatorLogo(&kube.FakeHandler{Err: tc.err})
			req := httptest.NewRequest("Get", "https://foo.bar/backend/v1/namespaces/kubeapps/operator/foo", bytes.NewReader(tc.logo))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps", "name": "foo"})

			response := httptest.NewRecorder()
			getOpLogo(response, req)

			if got := response.Header().Get("Content-Type"); tc.expectedContentType != got {
				t.Errorf("Expecting content-type %s got %s", tc.expectedContentType, got)
			}
		})
	}
}
