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
	"github.com/k14s/kapp/pkg/kapp/logger"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)
type kappFactoryGetter func(ctx context.Context, cluster string, appFlags kappcmdapp.Flags) (ctlapp.App, kappcmdapp.FactorySupportObjs, error)

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
	kappFactoryGetter        kappFactoryGetter
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster string) *Server {
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
		kappFactoryGetter: func(ctx context.Context, cluster string, appFlags kappcmdapp.Flags) (ctlapp.App, kappcmdapp.FactorySupportObjs, error) {
			if configGetter == nil {
				return nil, kappcmdapp.FactorySupportObjs{}, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, kappcmdapp.FactorySupportObjs{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
			}
			configFactory := NewConfigurableConfigFactoryImpl()
			configFactory.ConfigureRESTConfig(config)

			resourceTypesFlags := kappcmdapp.ResourceTypesFlags{
				IgnoreFailingAPIServices:         true,
				ScopeToFallbackAllowedNamespaces: true,
			}
			depsFactory := kappcmdcore.NewDepsFactoryImpl(configFactory, ui.NewNoopUI())

			app, supportObjs, err := kappcmdapp.Factory(depsFactory, appFlags, resourceTypesFlags, logger.NewNoopLogger())
			if err != nil {
				return nil, kappcmdapp.FactorySupportObjs{}, err
			}

			return app, supportObjs, nil
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

// GetKappFactory ensures a client getter is available and uses it to return a Kapp Factory.
func (s *Server) GetKappFactory(ctx context.Context, cluster string, appFlags kappcmdapp.Flags) (ctlapp.App, kappcmdapp.FactorySupportObjs, error) {
	if s.clientGetter == nil {
		return nil, kappcmdapp.FactorySupportObjs{}, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	app, supportObjs, err := s.kappFactoryGetter(ctx, cluster, appFlags)
	if err != nil {
		return nil, kappcmdapp.FactorySupportObjs{}, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get Kapp Factory : %v", err))
	}
	return app, supportObjs, nil
}
