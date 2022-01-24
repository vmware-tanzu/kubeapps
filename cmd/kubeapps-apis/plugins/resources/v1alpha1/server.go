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
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	"github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/client/clientset/versioned/scheme"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)

// Currently just a stub unimplemented server. More to come in following PRs.
type Server struct {
	v1alpha1.UnimplementedResourcesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientGetter

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

func NewServer(configGetter core.KubernetesConfigGetter) (*Server, error) {
	mapper, err := createRESTMapper()
	if err != nil {
		return nil, err
	}
	return &Server{
		clientGetter: func(ctx context.Context, cluster string) (kubernetes.Interface, dynamic.Interface, error) {
			if configGetter == nil {
				return nil, nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get config : %v", err.Error())
			}
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get dynamic client : %s", err.Error())
			}
			typedClient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, nil, status.Errorf(codes.FailedPrecondition, "unable to get typed client: %s", err.Error())
			}
			return typedClient, dynamicClient, nil
		},
		corePackagesClientGetter: func() (pkgsGRPCv1alpha1.PackagesServiceClient, error) {
			port := os.Getenv("PORT")
			conn, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
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
	}, nil
}

// GetResources returns the resources for an installed package.
func (s *Server) GetResources(r *v1alpha1.GetResourcesRequest, stream v1alpha1.ResourcesService_GetResourcesServer) error {
	namespace := r.GetInstalledPackageRef().GetContext().GetNamespace()
	cluster := r.GetInstalledPackageRef().GetContext().GetCluster()
	log.Infof("+resources GetResources (cluster: %q, namespace=%q)", cluster, namespace)

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
		sendResourceData(e.ResourceRef, e.Object, stream)
	}

	return nil
}

// GetServiceAccountNames returns the list of service account names in a given cluster and namespace.
func (s *Server) GetServiceAccountNames(ctx context.Context, r *v1alpha1.GetServiceAccountNamesRequest) (*v1alpha1.GetServiceAccountNamesResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.Infof("+resources GetServiceAccountNames (cluster: %q, namespace=%q)", cluster, namespace)

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
	s.Send(&v1alpha1.GetResourcesResponse{
		ResourceRef: ref,
		Manifest:    string(resourceBytes),
	})

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
