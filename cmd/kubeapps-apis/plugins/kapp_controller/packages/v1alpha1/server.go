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

	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	kappcmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	kappcmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type kappClientsGetter func(ctx context.Context, cluster, namespace string) (ctlapp.Apps, ctlres.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, ctlres.ResourceFilter, error)

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
	clientGetter             clientgetter.ClientGetterFunc
	globalPackagingNamespace string
	globalPackagingCluster   string
	kappClientsGetter        kappClientsGetter
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster string) *Server {
	return &Server{
		clientGetter:             clientgetter.NewClientGetter(configGetter),
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		kappClientsGetter: func(ctx context.Context, cluster, namespace string) (ctlapp.Apps, ctlres.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, ctlres.ResourceFilter, error) {
			if configGetter == nil {
				return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.Internal, "configGetter arg required")
			}
			// Retrieve the k8s REST client from the configGetter
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}
			// Pass the REST client to the (custom) kapp factory
			configFactory := NewConfigurableConfigFactoryImpl()
			configFactory.ConfigureRESTConfig(config)
			depsFactory := kappcmdcore.NewDepsFactoryImpl(configFactory, ui.NewNoopUI())

			// Create an empty resource filter
			resourceFilterFlags := kappcmdtools.ResourceFilterFlags{}
			resourceFilter, err := resourceFilterFlags.ResourceFilter()
			if err != nil {
				return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}

			// Create the preconfigured resource types flags and a failing policy
			resourceTypesFlags := kappcmdapp.ResourceTypesFlags{
				// Allow to ignore failing APIServices
				IgnoreFailingAPIServices: true,
				// Scope resource searching to fallback allowed namespaces
				ScopeToFallbackAllowedNamespaces: true,
			}
			failingAPIServicesPolicy := resourceTypesFlags.FailingAPIServicePolicy()

			// Getting namespaced clients (e.g., for fetching an App)
			supportingNsObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: namespace}, resourceTypesFlags, logger.NewNoopLogger())
			if err != nil {
				return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}

			// Getting non-namespaced clients (e.g., for fetching every k8s object in the cluster)
			supportingObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: ""}, resourceTypesFlags, logger.NewNoopLogger())
			if err != nil {
				return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}

			return supportingNsObjs.Apps, supportingObjs.IdentifiedResources, failingAPIServicesPolicy, resourceFilter, nil
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

// GetKappClients ensures a client getter is available and uses it to return a Kapp Factory.
func (s *Server) GetKappClients(ctx context.Context, cluster, namespace string) (ctlapp.Apps, ctlres.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, ctlres.ResourceFilter, error) {
	if s.clientGetter == nil {
		return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	appsClient, resourcesClient, failingAPIServicesPolicy, resourceFilter, err := s.kappClientsGetter(ctx, cluster, namespace)
	if err != nil {
		return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get Kapp Factory : %v", err))
	}
	return appsClient, resourcesClient, failingAPIServicesPolicy, resourceFilter, nil
}
