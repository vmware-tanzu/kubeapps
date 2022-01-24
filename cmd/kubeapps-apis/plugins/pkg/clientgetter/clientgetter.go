/*
Copyright © 2022 VMware
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
package clientgetter

import (
	"context"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	"github.com/kubeapps/kubeapps/pkg/agent"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
)

type HelmActionConfigGetterFunc func(ctx context.Context, namespace string) (*action.Configuration, error)
type ClientGetterFunc func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error)
type ClientGetterWithApiExtFunc func(context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error)

func NewHelmActionConfigGetter(configGetter core.KubernetesConfigGetter, cluster string) HelmActionConfigGetterFunc {
	return func(ctx context.Context, namespace string) (*action.Configuration, error) {
		if configGetter == nil {
			return nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		config, err := configGetter(ctx, cluster)
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

func NewClientGetter(configGetter core.KubernetesConfigGetter) ClientGetterFunc {
	return func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
		if configGetter == nil {
			return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		config, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get config : %v", err)
		}
		typedClient, dynamicClient, _, err := clientGetterHelper(config)
		if err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
		}
		return typedClient, dynamicClient, nil
	}
}

func NewClientGetterWithApiExt(configGetter core.KubernetesConfigGetter, cluster string) ClientGetterWithApiExtFunc {
	return func(ctx context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error) {
		if configGetter == nil {
			return nil, nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		if config, err := configGetter(ctx, cluster); err != nil {
			if status.Code(err) == codes.Unauthenticated {
				// want to make sure we return same status in this case
				return nil, nil, nil, status.Errorf(codes.Unauthenticated, "unable to get config due to: %v", err)
			} else {
				return nil, nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}
		} else {
			return clientGetterHelper(config)
		}
	}
}

// https://github.com/kubeapps/kubeapps/issues/3560
// flux plug-in runs out-of-request interactions with the Kubernetes API server.
// Although we've already ensured that if the flux plugin is selected, that the service account
// will be granted additional read privileges, we also need to ensure that the plugin can get a
// config based on the service account rather than the request context
func NewBackgroundClientGetter() ClientGetterWithApiExtFunc {
	return func(ctx context.Context) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error) {
		if config, err := rest.InClusterConfig(); err != nil {
			return nil, nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get in cluster config due to: %v", err)
		} else {
			return clientGetterHelper(config)
		}
	}
}

func clientGetterHelper(config *rest.Config) (kubernetes.Interface, dynamic.Interface, apiext.Interface, error) {
	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get typed client : %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client due to: %v", err)
	}
	apiExtensions, err := apiext.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get api extensions client due to: %v", err)
	}
	return typedClient, dynamicClient, apiExtensions, nil
}
