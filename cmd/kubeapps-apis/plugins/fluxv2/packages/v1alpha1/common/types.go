// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package common

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type HelmActionConfigGetterFunc func(ctx context.Context, namespace string) (*action.Configuration, error)
type ClientGetterFunc func(context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error)
