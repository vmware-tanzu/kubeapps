/*
Copyright (c) 2017 Bitnami

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

package response

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type resource struct {
	ID string `json:"id"`
}

func TestNewErrorResponse(t *testing.T) {
	type args struct {
		code    int
		message string
	}
	tests := []struct {
		name string
		args args
		want ErrorResponse
	}{
		{"404 response", args{http.StatusNotFound, "not found"}, ErrorResponse{http.StatusNotFound, "not found"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewErrorResponse(tt.args.code, tt.args.message))
		})
	}
}

func TestErrorResponse_Write(t *testing.T) {
	tests := []struct {
		name string
		e    ErrorResponse
	}{
		{"404 response", ErrorResponse{http.StatusNotFound, "not found"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.e.Write(w)
			assert.Equal(t, tt.e.Code, w.Code)
			var body ErrorResponse
			json.NewDecoder(w.Body).Decode(&body)
			assert.Equal(t, tt.e, body)
		})
	}
}

func TestNewDataResponse(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"single resource", resource{"test"}},
		{"multiple resources", []resource{{"one"}, {"two"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDataResponse(tt.data)
			assert.Equal(t, tt.data, d.Data)
		})
	}
}

func TestNewDataResponseWithMeta(t *testing.T) {
	tests := []struct {
		name string
		data interface{}
		meta interface{}
	}{
		{"single resource", resource{"test"}, resource{"foo"}},
		{"multiple resources", []resource{{"one"}, {"two"}}, resource{"foo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDataResponseWithMeta(tt.data, tt.meta)
			assert.Equal(t, tt.data, d.Data)
			assert.Equal(t, tt.meta, d.Meta)
		})
	}
}

func TestDataResponse_Write(t *testing.T) {
	tests := []struct {
		name string
		d    DataResponse
		want string
	}{
		{"single resource", DataResponse{http.StatusOK, resource{"test"}, nil}, `{"data":{"id":"test"}}`},
		{"multiple resources", DataResponse{http.StatusOK, []resource{{"one"}, {"two"}}, nil}, `{"data":[{"id":"one"},{"id":"two"}]}`},
		{"multiple resources with meta", DataResponse{http.StatusOK, []resource{{"one"}, {"two"}}, resource{"foo"}}, `{"data":[{"id":"one"},{"id":"two"}],"meta":{"id":"foo"}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			tt.d.Write(w)
			assert.Equal(t, http.StatusOK, w.Code)
			bytes, err := ioutil.ReadAll(w.Body)
			assert.NoError(t, err)
			body := string(bytes)
			assert.Equal(t, tt.want, body)
		})
	}
}

func TestDataResponse_WithCode(t *testing.T) {
	d := NewDataResponse(resource{"test"})
	assert.Equal(t, http.StatusOK, d.Code)
	d = d.WithCode(http.StatusBadRequest)
	assert.Equal(t, http.StatusBadRequest, d.Code)
}
