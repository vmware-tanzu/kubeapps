// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package clientgetter

import (
	"context"

	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	helmagent "github.com/kubeapps/kubeapps/pkg/agent"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmkube "helm.sh/helm/v3/pkg/kube"
	k8sapiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8srest "k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
)

type HelmActionConfigGetterFunc func(ctx context.Context, namespace string) (*helmaction.Configuration, error)
type ClientGetterFunc func(ctx context.Context, cluster string) (k8stypedclient.Interface, k8dynamicclient.Interface, error)
type ClientGetterWithApiExtFunc func(context.Context) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error)

func NewHelmActionConfigGetter(configGetter apiscore.KubernetesConfigGetter, cluster string) HelmActionConfigGetterFunc {
	return func(ctx context.Context, namespace string) (*helmaction.Configuration, error) {
		if configGetter == nil {
			return nil, grpcstatus.Errorf(grpccodes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		config, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
		}

		restClientGetter := helmagent.NewConfigFlagsFromCluster(namespace, config)
		clientSet, err := k8stypedclient.NewForConfig(config)
		if err != nil {
			return nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to create kubernetes client due to: %v", err)
		}
		// TODO(mnelson): Update to allow different helm storage options.
		storage := helmagent.StorageForSecrets(namespace, clientSet)
		return &helmaction.Configuration{
			RESTClientGetter: restClientGetter,
			KubeClient:       helmkube.New(restClientGetter),
			Releases:         storage,
			Log:              log.Infof,
		}, nil
	}
}

func NewClientGetter(configGetter apiscore.KubernetesConfigGetter) ClientGetterFunc {
	return func(ctx context.Context, cluster string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		if configGetter == nil {
			return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "configGetter arg required")
		}
		config, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config : %v", err)
		}
		typedClient, dynamicClient, _, err := clientGetterHelper(config)
		if err != nil {
			return nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get client due to: %v", err)
		}
		return typedClient, dynamicClient, nil
	}
}

func NewClientGetterWithApiExt(configGetter apiscore.KubernetesConfigGetter, cluster string) ClientGetterWithApiExtFunc {
	return func(ctx context.Context) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error) {
		if configGetter == nil {
			return nil, nil, nil, grpcstatus.Errorf(grpccodes.Internal, "configGetter arg required")
		}
		// The Flux plugin currently supports interactions with the default (kubeapps)
		// cluster only:
		if config, err := configGetter(ctx, cluster); err != nil {
			if grpcstatus.Code(err) == grpccodes.Unauthenticated {
				// want to make sure we return same status in this case
				return nil, nil, nil, grpcstatus.Errorf(grpccodes.Unauthenticated, "unable to get config due to: %v", err)
			} else {
				return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
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
	return func(ctx context.Context) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error) {
		if config, err := k8srest.InClusterConfig(); err != nil {
			return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get in cluster config due to: %v", err)
		} else {
			return clientGetterHelper(config)
		}
	}
}

func clientGetterHelper(config *k8srest.Config) (k8stypedclient.Interface, k8dynamicclient.Interface, k8sapiextensionsclient.Interface, error) {
	typedClient, err := k8stypedclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get typed client : %v", err)
	}
	dynamicClient, err := k8dynamicclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get dynamic client due to: %v", err)
	}
	apiExtensions, err := k8sapiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get api extensions client due to: %v", err)
	}
	return typedClient, dynamicClient, apiExtensions, nil
}
