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
	"net/http"
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

func checkAppResponse(t *testing.T, response *httptest.ResponseRecorder, expectedResponse appRepositoryResponse) {
	var appRepoResponse appRepositoryResponse
	err := json.NewDecoder(response.Body).Decode(&appRepoResponse)
	if err != nil {
		t.Fatalf("%+v", err)
	}
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

func TestListAppRepositories(t *testing.T) {
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

func TestGetAppRepository(t *testing.T) {
	testCases := []struct {
		name         string
		appRepo      *v1alpha1.AppRepository
		secret       *corev1.Secret
		err          error
		expectedCode int
	}{
		{
			name:         "it should return a 200 if the repo is found",
			appRepo:      &v1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kubeapps"}},
			expectedCode: 200,
		},
		{
			name: "it should return a corresponding secret if present",
			appRepo: &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kubeapps"},
				Spec: v1alpha1.AppRepositorySpec{
					Auth: v1alpha1.AppRepositoryAuth{
						Header: &v1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "repo-secret",
								},
								Key: "authorizationHeader",
							},
						},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "repo-secret", Namespace: "kubeapps"},
				StringData: map[string]string{
					"authorizationHeader": "someheader",
				},
			},
			expectedCode: 200,
		},
		{
			name:         "it should return a 404 if app repository not found",
			appRepo:      &v1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "kubeapps"}},
			err:          k8sErrors.NewNotFound(schema.GroupResource{}, "foo"),
			expectedCode: 404,
		},
		{
			name: "it should return a 404 if related secret is not found",
			appRepo: &v1alpha1.AppRepository{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "kubeapps"},
				Spec: v1alpha1.AppRepositorySpec{
					Auth: v1alpha1.AppRepositoryAuth{
						Header: &v1alpha1.AppRepositoryAuthHeader{
							SecretKeyRef: corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "repo-secret",
								},
								Key: "authorizationHeader",
							},
						},
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "another-repo-secret", Namespace: "kubeapps"},
				StringData: map[string]string{
					"authorizationHeader": "someheader",
				},
			},
			expectedCode: 404,
		},
		{
			name:         "it should return a 403 when forbidden",
			appRepo:      &v1alpha1.AppRepository{ObjectMeta: metav1.ObjectMeta{Name: "bar", Namespace: "kubeapps"}},
			err:          k8sErrors.NewForbidden(schema.GroupResource{}, "foo", fmt.Errorf("nope")),
			expectedCode: 403,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getAppFunc := GetAppRepository(&kube.FakeHandler{
				AppRepos: []*v1alpha1.AppRepository{tc.appRepo},
				Secrets:  []*corev1.Secret{tc.secret},
				Err:      tc.err,
			})
			req := httptest.NewRequest("GET", "https://foo.bar/backend/v1/namespaces/kubeapps/apprepositories/foo", strings.NewReader(""))
			req = mux.SetURLVars(req, map[string]string{"namespace": "kubeapps", "name": "foo"})

			response := httptest.NewRecorder()
			getAppFunc(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}
			expectedResponse := appRepositoryResponse{AppRepository: *tc.appRepo}
			if tc.secret != nil {
				expectedResponse.Secret = *tc.secret
			}
			if response.Code == 200 {
				checkAppResponse(t, response, expectedResponse)
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
				checkAppResponse(t, response, appRepositoryResponse{AppRepository: *tc.appRepo})
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
				checkAppResponse(t, response, appRepositoryResponse{AppRepository: *tc.appRepo})
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
		name                   string
		existingNamespaces     []corev1.Namespace
		expectedNamespaces     []corev1.Namespace
		err                    error
		expectedCode           int
		additionalHeader       http.Header
		namespaceHeaderOptions kube.KubeOptions
	}{
		{
			name:               "it should return the list of namespaces and a 200 if the repo is created",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode:       200,
		},
		{
			name:         "it should return a 403 when forbidden",
			err:          k8sErrors.NewForbidden(schema.GroupResource{}, "foo", fmt.Errorf("nope")),
			expectedCode: 403,
		},
		{
			name:               "it should return the list of namespaces from the header and a 200 if the repo is created",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{
				{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
				{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
			},
			expectedCode:     200,
			additionalHeader: http.Header{"X-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "X-Consumer-Groups",
				NamespaceHeaderPattern: "^namespace:(\\w+)$",
			},
		},
		{
			name:               "it should return the existing list of namespaces and a 200 when header does not match kubeops arg namespace-header-name",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode:       200,
			additionalHeader:   http.Header{"X-Consumer-Groups": []string{"nspace:ns1", "nspace:ns2"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "X-Consumer-Groups",
				NamespaceHeaderPattern: "^namespace:(\\w+)$",
			},
		},
		{
			name:               "it should return the existing list of namespaces and a 200 when header does not match kubeops arg namespace-header-pattern",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode:       200,
			additionalHeader:   http.Header{"Y-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "X-Consumer-Groups",
				NamespaceHeaderPattern: "^namespace:(\\w+)$",
			},
		},
		{
			name:               "it should return the existing list of namespaces and a 200 when kubeops arg namespace-header-name is empty",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode:       200,
			additionalHeader:   http.Header{"Y-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "",
				NamespaceHeaderPattern: "^namespace:(\\w+)$",
			},
		},
		{
			name:               "it should return the existing list of namespaces and a 200 when kubeops arg namespace-header-pattern is empty",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedCode:       200,
			additionalHeader:   http.Header{"Y-Consumer-Groups": []string{"namespace:ns1", "namespace:ns2"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "X-Consumer-Groups",
				NamespaceHeaderPattern: "",
			},
		},
		{
			name:               "it should return some of the namespaces from header and a 200 when not all match namespace-header-pattern",
			existingNamespaces: []corev1.Namespace{{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}},
			expectedNamespaces: []corev1.Namespace{
				{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
				{ObjectMeta: metav1.ObjectMeta{Name: "ns4"}, Status: corev1.NamespaceStatus{Phase: corev1.NamespaceActive}},
			},
			expectedCode:     200,
			additionalHeader: http.Header{"X-Consumer-Groups": []string{"namespace:ns1:read", "namespace:ns2", "ns3", "namespace:ns4", "ns:ns5:write"}},
			namespaceHeaderOptions: kube.KubeOptions{
				NamespaceHeaderName:    "X-Consumer-Groups",
				NamespaceHeaderPattern: "^namespace:(\\w+)$",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getNSFunc := GetNamespaces(&kube.FakeHandler{Namespaces: tc.existingNamespaces, Err: tc.err, Options: tc.namespaceHeaderOptions})
			req := httptest.NewRequest("GET", "https://foo.bar/backend/v1/namespaces", nil)

			for headerName, headerValue := range tc.additionalHeader {
				req.Header.Set(headerName, strings.Join(headerValue, ","))
			}

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
				expectedResponse := namespacesResponse{Namespaces: tc.expectedNamespaces}
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

func TestCanI(t *testing.T) {
	testCases := []struct {
		name         string
		body         string
		allowed      bool
		err          error
		expectedCode int
	}{
		{
			name:         "it should return an allowed response",
			body:         `{"resource":"namespaces","verb":"create"}`,
			allowed:      true,
			expectedCode: 200,
		},
		{
			name:         "it should return a forbidden response",
			body:         `{"resource":"namespaces","verb":"create"}`,
			allowed:      false,
			expectedCode: 200,
		},
		{
			name:         "it should return an error for wrong input",
			body:         "nope",
			err:          fmt.Errorf("invalid character 'o' in literal null (expecting 'u')"),
			expectedCode: 500,
		},
		{
			name:         "it returns a json 500 error as a plain string for internal backend errors",
			body:         `{"resource":"namespaces","verb":"create"}`,
			err:          fmt.Errorf("bang"),
			expectedCode: 500,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			function := CanI(&kube.FakeHandler{Can: tc.allowed, Err: tc.err})
			req := httptest.NewRequest("POST", "https://foo.bar/backend/v1/", strings.NewReader(tc.body))
			req = mux.SetURLVars(req, map[string]string{"cluster": "default"})

			response := httptest.NewRecorder()
			function(response, req)

			if got, want := response.Code, tc.expectedCode; got != want {
				t.Errorf("got: %d, want: %d\nBody: %s", got, want, response.Body)
			}

			if response.Code == 200 {
				allowedRes := allowedResponse{}
				err := json.NewDecoder(response.Body).Decode(&allowedRes)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				if allowedRes.Allowed != tc.allowed {
					t.Errorf("got: %v, want: %v", allowedRes.Allowed, tc.allowed)
				}
			} else {
				checkError(t, response, tc.err)
			}
		})
	}
}
