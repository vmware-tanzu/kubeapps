/*
Copyright 2022 The Flux authors

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

package object

import (
	"errors"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	ErrObservedGenerationNotFound     = errors.New("observed generation not found")
	ErrLastHandledReconcileAtNotFound = errors.New("last handled reconcile at not found")
	ErrRequeueIntervalNotFound        = errors.New("requeue interval not found")
)

// toUnstructured converts a runtime object into Unstructured.
// Based on https://github.com/fluxcd/pkg/blob/b4a14854c75753ea9431693b39c4be672f246552/runtime/patch/utils.go#L55.
func toUnstructured(obj runtime.Object) (*unstructured.Unstructured, error) {
	// If the incoming object is already unstructured, perform a deep copy first
	// otherwise DefaultUnstructuredConverter ends up returning the inner map without
	// making a copy.
	if _, ok := obj.(runtime.Unstructured); ok {
		obj = obj.DeepCopyObject()
	}
	rawMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: rawMap}, nil
}

// GetStatusLastHandledReconcileAt returns the status.lastHandledReconcileAt
// value of a given runtime object, if present.
func GetStatusLastHandledReconcileAt(obj runtime.Object) (string, error) {
	u, err := toUnstructured(obj)
	if err != nil {
		return "", err
	}
	ra, found, err := unstructured.NestedString(u.Object, "status", "lastHandledReconcileAt")
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrLastHandledReconcileAtNotFound
	}
	return ra, nil
}

// SetStatusLastHandledReconcileAt sets the status.lastHandledReconcileAt value
// of a given runtime object.
func SetStatusLastHandledReconcileAt(obj runtime.Object, val string) error {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return err
	}
	u := unstructured.Unstructured{}
	u.SetUnstructuredContent(content)
	if err := unstructured.SetNestedField(u.Object, val, "status", "lastHandledReconcileAt"); err != nil {
		return err
	}
	return runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, obj)
}

// GetStatusObservedGeneration returns the status.observedGeneration of a given
// runtime object.
func GetStatusObservedGeneration(obj runtime.Object) (int64, error) {
	u, err := toUnstructured(obj)
	if err != nil {
		return 0, err
	}
	og, found, err := unstructured.NestedInt64(u.Object, "status", "observedGeneration")
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, ErrObservedGenerationNotFound
	}
	return og, nil
}

// GetRequeueInterval returns the spec.interval of a given runtime object, if
// present.
func GetRequeueInterval(obj runtime.Object) (time.Duration, error) {
	period := time.Second
	u, err := toUnstructured(obj)
	if err != nil {
		return period, err
	}
	interval, found, err := unstructured.NestedString(u.Object, "spec", "interval")
	if err != nil {
		return period, err
	}
	if !found {
		return period, ErrRequeueIntervalNotFound
	}
	return time.ParseDuration(interval)
}
