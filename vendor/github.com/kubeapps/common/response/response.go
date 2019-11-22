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

// Package response implements helpers for JSON responses.
package response

import (
	"net/http"

	"github.com/unrolled/render"
)

var renderer *render.Render

func init() {
	renderer = render.New(render.Options{})
}

/*
ErrorResponse describes a JSON error response with the following body:
	{
		"code": 404,
		"message": "not found"
	}
*/
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse returns a new ErrorResponse
func NewErrorResponse(code int, message string) ErrorResponse {
	return ErrorResponse{code, message}
}

func (e ErrorResponse) Write(w http.ResponseWriter) {
	renderer.JSON(w, e.Code, e)
}

/*
DataResponse describes a JSON response containing resource data:
	{
		data: {...}
	}
If resource data is an array of objects, the data key will be an array:
	{
		data: [...]
	}
*/
type DataResponse struct {
	Code int         `json:"-"`
	Data interface{} `json:"data"`
	Meta interface{} `json:"meta,omitempty"`
}

// NewDataResponse returns a new DataResponse
func NewDataResponse(resources interface{}) DataResponse {
	return DataResponse{http.StatusOK, resources, nil}
}

// NewDataResponseWithMeta returns a new DataResponse
func NewDataResponseWithMeta(resources, meta interface{}) DataResponse {
	return DataResponse{http.StatusOK, resources, meta}
}

// WithCode sets the code for the response and returns the DataResponse
func (r DataResponse) WithCode(code int) DataResponse {
	r.Code = code
	return r
}

func (r DataResponse) Write(w http.ResponseWriter) {
	renderer.JSON(w, r.Code, r)
}
