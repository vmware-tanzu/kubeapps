// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	ctlapp "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/app"
	kappcmdapp "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/app"
	kappcmdcore "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/core"
	kappcmdtools "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/cmd/tools"
	"github.com/vmware-tanzu/carvel-kapp/pkg/kapp/logger"
	ctlres "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/resources"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type kappClientsGetter func(ctx context.Context, cluster, namespace string) (ctlapp.Apps, ctlres.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, ctlres.ResourceFilter, error)

const (
	fallbackGlobalPackagingNamespace                        = "kapp-controller-packaging-global"
	fallbackDefaultUpgradePolicy     pkgutils.UpgradePolicy = pkgutils.UpgradePolicyNone
	fallbackDefaultAllowDowngrades                          = false
	fallbackTimeoutSeconds                                  = 300
)

func fallbackDefaultPrereleasesVersionSelection() []string {
	return nil
}

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)
var _ corev1.RepositoriesServiceServer = (*Server)(nil)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedKappControllerPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter           clientgetter.ClientGetterFunc
	globalPackagingCluster string
	// TODO (gfichtenholt) it should now be possible to add this into clientgetter pkg,
	// and thus just have a single clientGetter field. Only *if* it makes sense to do so
	// (i.e. code is re-usable by multiple components)
	kappClientsGetter kappClientsGetter
	pluginConfig      *kappControllerPluginParsedConfig
}

// parsePluginConfig parses the input plugin configuration json file and return the configuration options.
func parsePluginConfig(pluginConfigPath string) (*kappControllerPluginParsedConfig, error) {
	// default configuration
	config := defaultPluginConfig

	// load the configuration file and unmarshall the values
	// #nosec G304
	pluginConfigFile, err := os.ReadFile(pluginConfigPath)
	if err != nil {
		return config, fmt.Errorf("unable to open plugin config at %q: %w", pluginConfigPath, err)
	}
	var pluginConfig kappControllerPluginConfig
	err = json.Unmarshal([]byte(pluginConfigFile), &pluginConfig)
	if err != nil {
		return config, fmt.Errorf("unable to unmarshal pluginconfig: %q error: %w", string(pluginConfigFile), err)
	}

	defaultUpgradePolicy, err := pkgutils.UpgradePolicyFromString(pluginConfig.KappController.Packages.V1alpha1.DefaultUpgradePolicy)
	if err != nil {
		return config, err
	}
	// override the defaults with the loaded configuration
	config.timeoutSeconds = pluginConfig.Core.Packages.V1alpha1.TimeoutSeconds
	config.versionsInSummary = pluginConfig.Core.Packages.V1alpha1.VersionsInSummary
	config.defaultUpgradePolicy = defaultUpgradePolicy
	config.defaultPrereleasesVersionSelection = pluginConfig.KappController.Packages.V1alpha1.DefaultPrereleasesVersionSelection
	config.defaultAllowDowngrades = pluginConfig.KappController.Packages.V1alpha1.DefaultAllowDowngrades
	config.globalPackagingNamespace = pluginConfig.KappController.Packages.V1alpha1.GlobalPackagingNamespace

	return config, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster, pluginConfigPath string) *Server {
	var err error
	pluginConfig := defaultPluginConfig
	if pluginConfigPath != "" {
		pluginConfig, err = parsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.InfoS("+kapp-controller using custom config", "pluginConfig", pluginConfig)
	} else {
		log.Info("+kapp-controller using default config since pluginConfigPath is empty")
	}
	return &Server{
		clientGetter:           clientgetter.NewClientGetter(configGetter, clientgetter.Options{}),
		globalPackagingCluster: globalPackagingCluster,
		pluginConfig:           pluginConfig,
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
	// TODO (gfichtenholt) Today this function returns 2 different
	// clients (typed and dynamic). Now if one looks at the callers, it is clear that
	// only one client is actually needed for a given scenario.
	// So for now, in order not to make too many changes, I am going to do more work than
	// is actually needed by getting *all* clients and returning them.
	// But we should think about refactoring the callers to ask for only what's needed
	dynamicClient, err := s.clientGetter.Dynamic(ctx, cluster)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	typedClient, err := s.clientGetter.Typed(ctx, cluster)
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
