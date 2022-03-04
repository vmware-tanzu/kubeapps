// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package k8sutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynfake "k8s.io/client-go/dynamic/fake"
)

func TestWaitForResource(t *testing.T) {
	testCases := []struct {
		name            string
		existingObjects []runtime.Object
		expectedErr     error
	}{
		{
			name: "the object exists",
			existingObjects: []runtime.Object{&k8scorev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-pod",
				},
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{{
						Name: "my-container",
					}},
				},
			}},
			expectedErr: nil,
		},
		{
			name: "the object does not exist",
			existingObjects: []runtime.Object{&k8scorev1.Pod{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Pod",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "im-not-here",
				},
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{{
						Name: "my-container",
					}},
				},
			}},
			expectedErr: fmt.Errorf("timed out waiting for the condition"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			dynClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
				runtime.NewScheme(),
				map[schema.GroupVersionResource]string{
					{Group: "", Version: "v1", Resource: "pods"}: "Pod" + "List",
				},
				unstructuredObjects...,
			)
			gvr := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
			resource := dynClient.Resource(gvr).Namespace("default")

			err := WaitForResource(context.Background(), resource, "my-pod", time.Second*1, time.Second*5)
			if tc.expectedErr != nil {
				if err == nil {
					t.Fatalf("expected an error but got none")
				}
				if got, want := err.Error(), tc.expectedErr.Error(); !cmp.Equal(want, got) {
					t.Errorf("in %s: mismatch (-want +got):\n%s", tc.name, cmp.Diff(want, got))
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
