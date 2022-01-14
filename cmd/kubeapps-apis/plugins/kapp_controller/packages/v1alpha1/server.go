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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)

const (
	globalPackagingNamespace               = "kapp-controller-packaging-global"
	DefaulltUpgradePolicy    versionPolicy = none
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
	upgradePolicy            versionPolicy
}

// parsePluginConfig parses the input plugin configuration json file and return the configuration options.
func parsePluginConfig(pluginConfigPath string) (versionPolicy, error) {
	type kappControllerPluginConfig struct {
		KappController struct {
			Packages struct {
				V1alpha1 struct {
					UpgradePolicy string `json:"upgradePolicy"`
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

	versionPolicy := versionPolicyMapping[config.KappController.Packages.V1alpha1.UpgradePolicy]

	// return configured value
	return versionPolicy, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter core.KubernetesConfigGetter, globalPackagingCluster, pluginConfigPath string) *Server {
	var err error
	upgradePolicy := DefaulltUpgradePolicy
	if pluginConfigPath != "" {
		upgradePolicy, err = parsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+kapp-controller using custom packages config with upgradePolicy: %v\n", upgradePolicy.string())
	} else {
		log.Infof("+kapp-controller using default config since pluginConfigPath is empty")
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
