// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"net/http"
	"reflect"
	"testing"
	"unsafe"

	"github.com/kubeapps/kubeapps/pkg/helm"
)

// GetUnexportedField returns a reflect.Value representing a (new) pointer to a value of the specified field
// using unsafe.Pointer as that pointer.
// In short, it's a way to get the value of an unexported field.
// Note that 'unsafe' is not always available in every system, but, as this code is solely run in test cases, it's ok
// Taken from: https://stackoverflow.com/a/60598827
func GetUnexportedField(field reflect.Value) interface{} {
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface()
}

// CheckHeader verifies that the given puller contains the given header
func CheckHeader(t *testing.T, ociClient helm.IOCIClient, key, value string) {
	// The header property is private so we need to use reflect to get its value
	resolver := GetUnexportedField(reflect.ValueOf(ociClient).Elem().FieldByName("resolver"))
	header := GetUnexportedField(reflect.Indirect(reflect.ValueOf(resolver)).FieldByName("header"))

	// type assertion to http.Header
	// See https://github.com/containerd/containerd/blob/main/remotes/docker/resolver.go
	headerMap := header.(http.Header)
	got := headerMap.Get(key)

	if got != value {
		t.Errorf("Expecting %s to contain %s", got, value)
	}

}
