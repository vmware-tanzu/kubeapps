/*
Copyright © 2021 VMware
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
package main

import (
	"context"
	"fmt"

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)

const (
	globalPackagingNamespace = "kapp-controller-packaging-global"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedKappControllerPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientGetter
	globalPackagingNamespace string
	globalPackagingCluster   string

	// We keep a restmapper to cache discovery of REST mappings from GVK->GVR.
	restMapper meta.RESTMapper

	// kindToResource is a function to convert a GVK to GVR with
	// namespace/cluster scope information. Can be replaced in tests with a
	// dummy version using the unsafe helpers while the real implementation
	// queries the k8s API for a REST mapper.
	kindToResource func(meta.RESTMapper, schema.GroupVersionKind) (schema.GroupVersionResource, meta.RESTScopeName, error)
}

// createRESTMapper returns a rest mapper configured with the APIs of the
// local k8s API server. This is used to convert between the GroupVersionKinds
// of the resource references to the GroupVersionResource used by the API server.
func createRESTMapper() (meta.RESTMapper, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// To use the config with RESTClientFor, extra fields are required.
	// See https://github.com/kubernetes/client-go/issues/657#issuecomment-842960258
	config.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	config.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
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

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster string) *Server {
	mapper, err := createRESTMapper()
	if err != nil {
		return nil
	}
	return &Server{
		clientGetter: func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
			if configGetter == nil {
				return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
			}
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get dynamic client : %v", err))
			}
			typedClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get typed client : %v", err))
			}
			return typedClient, dynamicClient, nil
		},
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		restMapper:               mapper,
		kindToResource: func(mapper meta.RESTMapper, gvk schema.GroupVersionKind) (schema.GroupVersionResource, meta.RESTScopeName, error) {
			mapping, err := mapper.RESTMapping(gvk.GroupKind())
			if err != nil {
				return schema.GroupVersionResource{}, "", err
			}
			return mapping.Resource, mapping.Scope.Name(), nil
		},
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}
