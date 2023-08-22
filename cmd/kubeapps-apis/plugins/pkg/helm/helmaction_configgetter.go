// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/helm/agent"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type HelmActionConfigGetterFunc func(headers http.Header, namespace string) (*action.Configuration, error)

func NewHelmActionConfigGetter(configGetter core.KubernetesConfigGetter, cluster string) HelmActionConfigGetterFunc {
	return func(headers http.Header, namespace string) (*action.Configuration, error) {
		if configGetter == nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("The configGetter arg is required"))
		}
		config, err := configGetter(headers, cluster)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Unable to get config due to: %w", err))
		}

		restClientGetter := agent.NewConfigFlagsFromCluster(namespace, config)
		clientSet, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("Unable to create kubernetes client due to: %w", err))
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
