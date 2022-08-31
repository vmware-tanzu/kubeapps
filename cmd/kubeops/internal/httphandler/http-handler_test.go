// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package httphandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func TestExtractToken(t *testing.T) {
	testSuite := []struct {
		Name          string
		TokenRaw      string
		ExpectedToken string
	}{
		{
			"Token ok",
			"Bearer foo",
			"foo",
		},
		{
			"Token nok",
			"foo bar",
			"",
		},
	}
	for _, test := range testSuite {
		t.Run(test.Name, func(t *testing.T) {
			if got, want := extractToken(test.TokenRaw), test.ExpectedToken; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
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
