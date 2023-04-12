// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"context"
	"net/http"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/helm/agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type HelmActionConfigGetterFunc func(headers http.Header, namespace string) (*action.Configuration, error)

func NewHelmActionConfigGetter(configGetter core.KubernetesConfigGetter, cluster string) HelmActionConfigGetterFunc {
	return func(headers http.Header, namespace string) (*action.Configuration, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// TODO(minelson): Remove context from signature of configGetter.
		config, err := configGetter(context.TODO(), headers, cluster)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
		}

		restClientGetter := agent.NewConfigFlagsFromCluster(namespace, config)
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to create kubernetes client due to: %v", err)
		}
		// TODO(mnelson): Update to allow different helm storage options.
		storage := agent.StorageForSecrets(namespace, clientSet)
		return &action.Configuration{
			RESTClientGetter: restClientGetter,
			KubeClient:       kube.New(restClientGetter),
			Releases:         storage,
			Log:              log.Infof,
		}, nil
	}
}
