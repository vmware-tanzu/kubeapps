// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	ui "github.com/cppforlife/go-cli-ui/ui"
	kappapp "github.com/k14s/kapp/pkg/kapp/app"
	kappcmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	kappcmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	kapplogger "github.com/k14s/kapp/pkg/kapp/logger"
	kappresources "github.com/k14s/kapp/pkg/kapp/resources"
	apiscore "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pkgkappv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8stypedclient "k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type kappClientsGetter func(ctx context.Context, cluster, namespace string) (kappapp.Apps, kappresources.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, kappresources.ResourceFilter, error)

const (
	globalPackagingNamespace                   = "kapp-controller-packaging-global"
	fallbackDefaultUpgradePolicy upgradePolicy = none
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ pkgsGRPCv1alpha1.PackagesServiceServer = (*Server)(nil)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	pkgkappv1alpha1.UnimplementedKappControllerPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter             clientgetter.ClientGetterFunc
	globalPackagingNamespace string
	globalPackagingCluster   string
	kappClientsGetter        kappClientsGetter
	defaultUpgradePolicy     upgradePolicy
}

// parsePluginConfig parses the input plugin configuration json file and return the configuration options.
func parsePluginConfig(pluginConfigPath string) (upgradePolicy, error) {
	type kappControllerPluginConfig struct {
		KappController struct {
			Packages struct {
				V1alpha1 struct {
					DefaultUpgradePolicy string `json:"defaultUpgradePolicy"`
				} `json:"v1alpha1"`
			} `json:"packages"`
		} `json:"kappController"`
	}
	var config kappControllerPluginConfig

	pluginConfig, err := ioutil.ReadFile(pluginConfigPath)
	if err != nil {
		return none, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	err = json.Unmarshal([]byte(pluginConfig), &config)
	if err != nil {
		return none, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfig), err)
	}

	defaultUpgradePolicy := upgradePolicyMapping[config.KappController.Packages.V1alpha1.DefaultUpgradePolicy]

	// return configured value
	return defaultUpgradePolicy, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter apiscore.KubernetesConfigGetter, globalPackagingCluster, pluginConfigPath string) *Server {
	var err error
	defaultUpgradePolicy := fallbackDefaultUpgradePolicy
	if pluginConfigPath != "" {
		defaultUpgradePolicy, err = parsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+kapp-controller using custom packages config with defaultUpgradePolicy: %v\n", defaultUpgradePolicy.string())
	} else {
		log.Infof("+kapp-controller using default config since pluginConfigPath is empty")
	}
	return &Server{
		clientGetter:             clientgetter.NewClientGetter(configGetter),
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		kappClientsGetter: func(ctx context.Context, cluster, namespace string) (kappapp.Apps, kappresources.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, kappresources.ResourceFilter, error) {
			if configGetter == nil {
				return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.Internal, "configGetter arg required")
			}
			// Retrieve the k8s REST client from the configGetter
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
			}
			// Pass the REST client to the (custom) kapp factory
			configFactory := NewConfigurableConfigFactoryImpl()
			configFactory.ConfigureRESTConfig(config)
			depsFactory := kappcmdcore.NewDepsFactoryImpl(configFactory, ui.NewNoopUI())

			// Create an empty resource filter
			resourceFilterFlags := kappcmdtools.ResourceFilterFlags{}
			resourceFilter, err := resourceFilterFlags.ResourceFilter()
			if err != nil {
				return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
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
			supportingNsObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: namespace}, resourceTypesFlags, kapplogger.NewNoopLogger())
			if err != nil {
				return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
			}

			// Getting non-namespaced clients (e.g., for fetching every k8s object in the cluster)
			supportingObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: ""}, resourceTypesFlags, kapplogger.NewNoopLogger())
			if err != nil {
				return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
			}

			return supportingNsObjs.Apps, supportingObjs.IdentifiedResources, failingAPIServicesPolicy, resourceFilter, nil
		},
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context, cluster string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, nil, grpcstatus.Errorf(grpccodes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}

// GetKappClients ensures a client getter is available and uses it to return a Kapp Factory.
func (s *Server) GetKappClients(ctx context.Context, cluster, namespace string) (kappapp.Apps, kappresources.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, kappresources.ResourceFilter, error) {
	if s.clientGetter == nil {
		return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.Internal, "server not configured with configGetter")
	}
	appsClient, resourcesClient, failingAPIServicesPolicy, resourceFilter, err := s.kappClientsGetter(ctx, cluster, namespace)
	if err != nil {
		return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, fmt.Sprintf("unable to get Kapp Factory : %v", err))
	}
	return appsClient, resourcesClient, failingAPIServicesPolicy, resourceFilter, nil
}
