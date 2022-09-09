// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/resources/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	log "k8s.io/klog/v2"

	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)

// Currently just a stub unimplemented server. More to come in following PRs.
type Server struct {
	v1alpha1.UnimplementedResourcesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientGetter

	// clusterServiceAccountClientGetter gets a client getter with service account for additional clusters
	clusterServiceAccountClientGetter clientGetter

	// for interactions with k8s API server in the context of
	// kubeapps-internal-kubeappsapis service account
	localServiceAccountClientGetter clientgetter.BackgroundClientGetterFunc

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

	// pluginConfig Resources plugin configuration values
	pluginConfig *common.ResourcesPluginConfig

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
	pluginConfig := common.NewDefaultPluginConfig()
	if pluginConfigPath != "" {
		pluginConfig, err = common.ParsePluginConfig(pluginConfigPath)
		if err != nil {
			log.Fatalf("%s", err)
		}
		log.Infof("+resources using custom config: [%v]", *pluginConfig)
	} else {
		log.Info("+resources using default config since pluginConfigPath is empty")
	}

	return &Server{
		// Get the client getter with context auth
		clientGetter: newClientGetter(configGetter, false, clustersConfig),
		// Get the additional cluster client getter with service account
		clusterServiceAccountClientGetter: newClientGetter(configGetter, true, clustersConfig),
		// Get the "in-cluster" client getter
		localServiceAccountClientGetter: clientgetter.NewBackgroundClientGetter(configGetter, clientgetter.Options{}),
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

func newClientGetter(configGetter core.KubernetesConfigGetter, useServiceAccount bool, clustersConfig kube.ClustersConfig) clientGetter {
	return func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
		if configGetter == nil {
			return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
		}
		restConfig, err := configGetter(ctx, cluster)
		if err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get config : %v", err.Error())
		}
		if err := setupRestConfigForCluster(restConfig, cluster, useServiceAccount, clustersConfig); err != nil {
			return nil, nil, err
		}
		dynamicClient, err := dynamic.NewForConfig(restConfig)
		if err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client : %s", err.Error())
		}
		typedClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get typed client: %s", err.Error())
		}
		return typedClient, dynamicClient, nil
	}
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

// GetResources returns the resources for an installed package.
func (s *Server) GetResources(r *v1alpha1.GetResourcesRequest, stream v1alpha1.ResourcesService_GetResourcesServer) error {
	namespace := r.GetInstalledPackageRef().GetContext().GetNamespace()
	cluster := r.GetInstalledPackageRef().GetContext().GetCluster()
	log.InfoS("+resources GetResources ", "cluster", cluster, "namespace", namespace)

	ctx, err := copyAuthorizationMetadataForOutgoing(stream.Context())
	if err != nil {
		return err
	}

	// First we grab the resource references for the specified installed package.
	coreClient, err := s.corePackagesClientGetter()
	if err != nil {
		return err
	}
	refsResponse, err := coreClient.GetInstalledPackageResourceRefs(ctx, &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
		InstalledPackageRef: r.InstalledPackageRef,
	})
	if err != nil {
		return err
	}
	var resourcesToReturn []*pkgsGRPCv1alpha1.ResourceRef
	// If the request didn't specify a filter of resource refs,
	// we return all those found for the installed package. Otherwise
	// we only return the requested ones.
	if len(r.GetResourceRefs()) == 0 {
		if r.GetWatch() {
			return status.Errorf(codes.InvalidArgument, "resource refs must be specified in request when watching resources")
		}
		resourcesToReturn = refsResponse.GetResourceRefs()
	} else {
		for _, requestedRef := range r.GetResourceRefs() {
			found := false
			for _, pkgRef := range refsResponse.GetResourceRefs() {
				if resourceRefsEqual(pkgRef, requestedRef) {
					found = true
					break
				}
			}
			if !found {
				return status.Errorf(codes.InvalidArgument, "requested resource %+v does not belong to installed package %+v", requestedRef, r.GetInstalledPackageRef())
			}
		}
		resourcesToReturn = r.GetResourceRefs()
	}

	// Then look up each referenced resource and send it down the stream.
	_, dynamicClient, err := s.clientGetter(stream.Context(), cluster)
	if err != nil {
		return err
	}
	var watchers []*ResourceWatcher
	for _, ref := range resourcesToReturn {
		groupVersion, err := schema.ParseGroupVersion(ref.ApiVersion)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to parse group version from %q: %s", ref.ApiVersion, err.Error())
		}
		gvk := groupVersion.WithKind(ref.Kind)

		// We need to get or watch a different endpoint depending on
		// the scope of the resource (namespaced or not).
		gvr, scopeName, err := s.kindToResource(s.restMapper, gvk)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to map group-kind %v to resource: %s", gvk.GroupKind(), err.Error())
		}

		if !r.GetWatch() {
			var resource interface{}
			if scopeName == meta.RESTScopeNameNamespace {
				resource, err = dynamicClient.Resource(gvr).Namespace(ref.Namespace).Get(stream.Context(), ref.GetName(), metav1.GetOptions{})
			} else {
				resource, err = dynamicClient.Resource(gvr).Get(stream.Context(), ref.GetName(), metav1.GetOptions{})
			}
			if err != nil {
				return status.Errorf(codes.Internal, "unable to get resource referenced by %+v: %s", ref, err.Error())
			}
			err = sendResourceData(ref, resource, stream)
			if err != nil {
				return err
			}

			continue
		}

		listOptions := metav1.ListOptions{
			FieldSelector: fmt.Sprintf("metadata.name=%s", ref.GetName()),
		}
		var watcher watch.Interface
		if scopeName == meta.RESTScopeNameNamespace {
			watcher, err = dynamicClient.Resource(gvr).Namespace(ref.Namespace).Watch(stream.Context(), listOptions)
		} else {
			watcher, err = dynamicClient.Resource(gvr).Watch(stream.Context(), listOptions)
		}
		if err != nil {
			log.Errorf("unable to watch resource %v: %v", ref, err)
			return status.Errorf(codes.Internal, "unable to watch resource %v", ref)
		}
		watchers = append(watchers, &ResourceWatcher{
			ResourceRef: ref,
			Watcher:     watcher,
		})
	}

	// If we're not watching, we're done.
	if watchers == nil {
		return nil
	}

	// Otherwise merge the watchers into a single resourceWatcher and stream the
	// data as it arrives.
	resourceWatcher := mergeWatchers(watchers)
	for e := range resourceWatcher.ResultChan() {
		err = sendResourceData(e.ResourceRef, e.Object, stream)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetServiceAccountNames returns the list of service account names in a given cluster and namespace.
func (s *Server) GetServiceAccountNames(ctx context.Context, r *v1alpha1.GetServiceAccountNamesRequest) (*v1alpha1.GetServiceAccountNamesResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.InfoS("+resources GetServiceAccountNames ", "cluster", cluster, "namespace", namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	saList, err := typedClient.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("list", "ServiceAccounts", "", err)
	}

	// We only need to send the list of SA names (ie, not sending secret names)
	saStringList := []string{}
	for _, sa := range saList.Items {
		saStringList = append(saStringList, sa.Name)
	}

	return &v1alpha1.GetServiceAccountNamesResponse{
		ServiceaccountNames: saStringList,
	}, nil

}

// sendResourceData just DRYs up this functionality shared between requests to
// watch or get resources.
func sendResourceData(ref *pkgsGRPCv1alpha1.ResourceRef, obj interface{}, s v1alpha1.ResourcesService_GetResourcesServer) error {
	resourceBytes, err := json.Marshal(obj)
	if err != nil {
		return status.Errorf(codes.Internal, "unable to marshal json for resource: %s", err.Error())
	}

	// Note, a string in Go is effectively a read-only slice of bytes.
	// See https://stackoverflow.com/a/50880408 for interesting links.
	err = s.Send(&v1alpha1.GetResourcesResponse{
		ResourceRef: ref,
		Manifest:    string(resourceBytes),
	})
	if err != nil {
		return status.Errorf(codes.Internal, "unable send GetResourcesResponse: %s", err.Error())
	}

	return nil
}

// ResourceEvent embeds a watch.Event and adds the resource ref.
type ResourceEvent struct {
	watch.Event
	ResourceRef *pkgsGRPCv1alpha1.ResourceRef
}

// mergeWatchers returns a single watcher for many.
// Inspired by the fan-in merge at https://go.dev/blog/pipelines
func mergeWatchers(watchers []*ResourceWatcher) *MultiResourceWatcher {
	var wg sync.WaitGroup
	out := make(chan ResourceEvent)

	// Start an output goroutine for each input channel in watchers.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan ResourceEvent) {
		for e := range c {
			log.Errorf("Event received: %+v", e)
			out <- e
		}
		wg.Done()
	}
	wg.Add(len(watchers))
	for _, w := range watchers {
		go output(w.ResultChan())
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return &MultiResourceWatcher{
		StopFn: func() {
			for _, w := range watchers {
				w.Stop()
			}
		},
		ResultChanFn: func() <-chan ResourceEvent {
			return out
		},
	}
}

type MultiResourceWatcher struct {
	StopFn       func()
	ResultChanFn func() <-chan ResourceEvent
}

func (mw *MultiResourceWatcher) Stop() {
	mw.StopFn()
}

func (mw *MultiResourceWatcher) ResultChan() <-chan ResourceEvent {
	return mw.ResultChanFn()
}

// ResourceWatcher is a watcher that knows the resource its watching.
type ResourceWatcher struct {
	Watcher     watch.Interface
	ResourceRef *pkgsGRPCv1alpha1.ResourceRef
	resultChan  chan ResourceEvent
}

func (rw *ResourceWatcher) Stop() {
	// Calling Watcher.Stop() will close the watcher's result channel
	// which will cascade below in ResultChan() to close the
	rw.Watcher.Stop()
}

func (rw *ResourceWatcher) ResultChan() <-chan ResourceEvent {
	if rw.resultChan != nil {
		return rw.resultChan
	}

	rw.resultChan = make(chan ResourceEvent)

	// Start a go-routine that copies the actual watcher events but with
	// the extra resource ref.
	go func() {
		for e := range rw.Watcher.ResultChan() {
			rw.resultChan <- ResourceEvent{e, rw.ResourceRef}
		}
		close(rw.resultChan)
	}()

	return rw.resultChan
}

// copyAuthorizationMetadataForOutgoing explicitly copies the authz from the
// incoming context to the outgoing context when making the outgoing call the
// core packaging API.
func copyAuthorizationMetadataForOutgoing(ctx context.Context) (context.Context, error) {
	notAllowedErr := status.Errorf(codes.PermissionDenied, "unable to get authorization from request context")

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, notAllowedErr
	}
	if len(md["authorization"]) == 0 {
		return nil, notAllowedErr
	}

	return metadata.AppendToOutgoingContext(ctx, "authorization", md["authorization"][0]), nil
}

func resourceRefsEqual(r1, r2 *pkgsGRPCv1alpha1.ResourceRef) bool {
	return r1.ApiVersion == r2.ApiVersion &&
		r1.Kind == r2.Kind &&
		r1.Namespace == r2.Namespace &&
		r1.Name == r2.Name
}
