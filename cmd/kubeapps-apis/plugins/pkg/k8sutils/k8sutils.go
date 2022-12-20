// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package k8sutils

import (
	"context"
	"time"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
)

const (
	// description as annotation
	AnnotationDescriptionKey = "kubeapps.dev/description"
)

func WaitForResource(ctx context.Context, ri dynamic.ResourceInterface, name string, interval, timeout time.Duration) error {
	err := wait.PollImmediateWithContext(ctx, interval, timeout, func(ctx context.Context) (bool, error) {
		_, err := ri.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				// the resource hasn't been created yet
				return false, nil
			} else {
				// any other real error
				return false, statuserror.FromK8sError("wait", "resource", name, err)
			}
		}
		// the resource is created now
		return true, nil
	})
	return err
}

// description
func SetDescription(metadata *metav1.ObjectMeta, description string) {
	if description != "" {
		if metadata.Annotations == nil {
			metadata.Annotations = make(map[string]string)
		}
		metadata.Annotations[AnnotationDescriptionKey] = description
	} else {
		delete(metadata.Annotations, AnnotationDescriptionKey)
	}
}

func GetDescription(metadata *metav1.ObjectMeta) string {
	return metadata.Annotations[AnnotationDescriptionKey]
}
