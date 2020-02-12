package handlerutil

import (
	"fmt"
	"net/http"
	"testing"
)

func TestErrorCodeWithDefault(t *testing.T) {
	type test struct {
		err          error
		defaultCode  int
		expectedCode int
	}
	tests := []test{
		{fmt.Errorf("a release named foo already exists"), http.StatusInternalServerError, http.StatusConflict},
		{fmt.Errorf("release foo not found"), http.StatusInternalServerError, http.StatusNotFound},
		{fmt.Errorf("Unauthorized to get release foo"), http.StatusInternalServerError, http.StatusForbidden},
		{fmt.Errorf("release \"Foo \" failed"), http.StatusInternalServerError, http.StatusUnprocessableEntity},
		{fmt.Errorf("Release \"Foo \" failed"), http.StatusInternalServerError, http.StatusUnprocessableEntity},
		{fmt.Errorf("This is an unexpected error"), http.StatusInternalServerError, http.StatusInternalServerError},
		{fmt.Errorf("This is an unexpected error"), http.StatusUnprocessableEntity, http.StatusUnprocessableEntity},
	}
	for _, s := range tests {
		code := ErrorCodeWithDefault(s.err, s.defaultCode)
		if code != s.expectedCode {
			t.Errorf("Expected '%v' to return code %v got %v", s.err, s.expectedCode, code)
		}
	}
}
