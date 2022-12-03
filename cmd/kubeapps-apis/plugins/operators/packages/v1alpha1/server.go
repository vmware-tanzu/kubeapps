// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	apimanifests "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
	ctrlruntime "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/operators/packages/v1alpha1"
)

const (
	packageManifestResource = "PackageManifest"
)

type Server struct {
	v1alpha1.UnimplementedOperatorsPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientgetter.ClientProviderInterface

	// pluginConfig Operators plugin configuration values
	pluginConfig *OperatorsPluginConfig

	kubeappsCluster string
}

func NewServer(configGetter core.KubernetesConfigGetter, clientQPS float32, clientBurst int, pluginConfigPath string, clustersConfig kube.ClustersConfig) (*Server, error) {
	// If no config is provided, we default to the existing values for backwards compatibility.
	pluginConfig := NewDefaultPluginConfig()
	if pluginConfigPath != "" {
		var err error
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

	return &Server{
		// Get the client getter with context auth
		clientGetter:    clientGetter,
		pluginConfig:    pluginConfig,
		kubeappsCluster: clustersConfig.KubeappsClusterName,
	}, nil
}

func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	log.Infof("+operators GetAvailablePackageSummaries(request: [%v])", request)
	defer log.Info("-operators GetAvailablePackageSummaries")

	// grpc compiles in getters for you which automatically return a default (empty) struct
	// if the pointer was nil
	cluster := request.GetContext().GetCluster()
	if request != nil && cluster != "" && cluster != s.kubeappsCluster {
		return nil, status.Errorf(
			codes.Unimplemented,
			"not supported yet: request.Context.Cluster: [%v]",
			request.Context.Cluster)
	}

	_, err := paginate.ItemOffsetFromPageToken(request.GetPaginationOptions().GetPageToken())
	if err != nil {
		return nil, err
	}

	// Retrieve parameters from the request
	namespace := request.GetContext().GetNamespace()
	client, err := s.clientGetter.ControllerRuntime(ctx, cluster)
	if err != nil {
		return nil, err
	}
	var manifestList apimanifests.PackageManifestList
	if err := client.List(ctx, &manifestList, &ctrlruntime.ListOptions{Namespace: namespace}); err != nil {
		return nil, statuserror.FromK8sError("list", packageManifestResource, "", err)
	}

	summaries := []*corev1.AvailablePackageSummary{}
	for _, item := range manifestList.Items {
		summary, err := s.availablePackageSummaryFromPackageManifest(item)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}

	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: summaries,
		NextPageToken:             "0", // TODO
	}, nil
}

func (s *Server) availablePackageSummaryFromPackageManifest(manifest apimanifests.PackageManifest) (*corev1.AvailablePackageSummary, error) {
	summary := corev1.AvailablePackageSummary{}
	// TODO: add catalog prefix to identifier, e.g. "operatorhubio-catalog"
	// a little mute at the moment since there is only one catalog we care about
	// (that of operatorhub.io) but for the future. Maybe we'll want to support multiple
	// catalogs one day
	pkgIdentifier := manifest.Name
	summary.AvailablePackageRef = &corev1.AvailablePackageReference{
		Identifier: pkgIdentifier,
		Plugin:     &pluginDetail,
		Context: &corev1.Context{
			Namespace: manifest.Namespace,
			Cluster:   s.kubeappsCluster,
		},
	}
	summary.Name = manifest.Status.PackageName
	summary.DisplayName = manifest.Status.Channels[0].CurrentCSVDesc.DisplayName
	return &summary, nil
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
