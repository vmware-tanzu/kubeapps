// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	log "k8s.io/klog/v2"

	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/operators/packages/v1alpha1"
)

type Server struct {
	v1alpha1.UnimplementedOperatorsPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientgetter.ClientProviderInterface

	// clusterServiceAccountClientGetter gets a client getter with service account for additional clusters
	clusterServiceAccountClientGetter clientgetter.ClientProviderInterface

	// for interactions with k8s API server in the context of
	// kubeapps-internal-kubeappsapis service account
	localServiceAccountClientGetter clientgetter.FixedClusterClientProviderInterface

	// corePackagesClientGetter holds a function to obtain the core.packages.v1alpha1
	// client. It is similarly initialised in NewServer() below.
	corePackagesClientGetter func() (pkgsGRPCv1alpha1.PackagesServiceClient, error)

	// We keep a restmapper to cache discovery of REST mappings from GVK->GVR.
	restMapper meta.RESTMapper

	// kindToResource is a function to convert a GVK to GVR with
	// namespace/cluster scope information. Can be replaced in tests with a
	// stub version using the unsafe helpers while the real implementation
	// queries the k8s API for a REST mapper.
	kindToResource func(meta.RESTMapper, schema.GroupVersionKind) (schema.GroupVersionResource, meta.RESTScopeName, error)

	// pluginConfig Operators plugin configuration values
	pluginConfig *OperatorsPluginConfig

	clientQPS float32

	kubeappsCluster string
}

// createRESTMapper returns a rest mapper configured with the APIs of the
// local k8s API server. This is used to convert between the GroupVersionKinds
// of the resource references to the GroupVersionResource used by the API server.
func createRESTMapper(clientQPS float32, clientBurst int) (meta.RESTMapper, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// To use the config with RESTClientFor, extra fields are required.
	// See https://github.com/kubernetes/client-go/issues/657#issuecomment-842960258
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}

	// Avoid client-side throttling while the rest mapper discovers the
	// available APIs on the K8s api server.  Note that this is only used for
	// the discovery client below to return the rest mapper. The configured
	// values for QPS and Burst are used for the client used for user requests.
	config.QPS = clientQPS
	config.Burst = clientBurst

	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, err
	}
	discoveryClient := discovery.NewDiscoveryClient(client)
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDiscoveryRESTMapper(groupResources), nil
}

func NewServer(configGetter core.KubernetesConfigGetter, clientQPS float32, clientBurst int, pluginConfigPath string, clustersConfig kube.ClustersConfig) (*Server, error) {
	mapper, err := createRESTMapper(clientQPS, clientBurst)
	if err != nil {
		return nil, err
	}

	// If no config is provided, we default to the existing values for backwards compatibility.
	pluginConfig := NewDefaultPluginConfig()
	if pluginConfigPath != "" {
		pluginConfig, err = ParsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+operators using custom config: [%v]", *pluginConfig)
	} else {
		log.Info("+operators using default config since pluginConfigPath is empty")
	}

	clientGetter, err := newClientGetter(configGetter, false, clustersConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}

	clusterServiceAccountClientGetter, err := newClientGetter(configGetter, true, clustersConfig)
	if err != nil {
		log.Fatalf("%s", err)
	}

	return &Server{
		// Get the client getter with context auth
		clientGetter: clientGetter,
		// Get the additional cluster client getter with service account
		clusterServiceAccountClientGetter: clusterServiceAccountClientGetter,
		// Get the "in-cluster" client getter
		localServiceAccountClientGetter: clientgetter.NewBackgroundClientProvider(clientgetter.Options{}, clientQPS, clientBurst),
		corePackagesClientGetter: func() (pkgsGRPCv1alpha1.PackagesServiceClient, error) {
			port := os.Getenv("PORT")
			conn, err := grpc.Dial("localhost:"+port, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "unable to dial to localhost grpc service: %s", err.Error())
			}
			return pkgsGRPCv1alpha1.NewPackagesServiceClient(conn), nil
		},
		restMapper: mapper,
		kindToResource: func(mapper meta.RESTMapper, gvk schema.GroupVersionKind) (schema.GroupVersionResource, meta.RESTScopeName, error) {
			mapping, err := mapper.RESTMapping(gvk.GroupKind())
			if err != nil {
				return schema.GroupVersionResource{}, "", err
			}
			return mapping.Resource, mapping.Scope.Name(), nil
		},
		clientQPS:       clientQPS,
		pluginConfig:    pluginConfig,
		kubeappsCluster: clustersConfig.KubeappsClusterName,
	}, nil
}

func newClientGetter(configGetter core.KubernetesConfigGetter, useServiceAccount bool, clustersConfig kube.ClustersConfig) (clientgetter.ClientProviderInterface, error) {

	customConfigGetter := func(ctx context.Context, cluster string) (*rest.Config, error) {
		restConfig, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "unable to get config : %v", err.Error())
		}
		if err := setupRestConfigForCluster(restConfig, cluster, useServiceAccount, clustersConfig); err != nil {
			return nil, err
		}
		return restConfig, nil
	}

	clientProvider, err := clientgetter.NewClientProvider(customConfigGetter, clientgetter.Options{})
	if err != nil {
		return nil, err
	}
	return clientProvider, nil
}

func setupRestConfigForCluster(restConfig *rest.Config, cluster string, useServiceAccount bool, clustersConfig kube.ClustersConfig) error {
	// Override client config with the service token for additional cluster
	// Added from #5034 after deprecation of "kubeops"
	if cluster != clustersConfig.KubeappsClusterName && useServiceAccount {
		additionalCluster, ok := clustersConfig.Clusters[cluster]
		if !ok {
			return status.Errorf(codes.Internal, "cluster %q has no configuration", cluster)
		}
		if additionalCluster.ServiceToken != "" {
			restConfig.BearerToken = additionalCluster.ServiceToken
		}
	}
	return nil
}
