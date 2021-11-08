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

	corev1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"
)

type clientGetter func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error)

type Server struct {
	v1alpha1.UnimplementedResourcesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter clientGetter

	// TODO: use a field for the function to return resource refs,
	// so it can be replaced in prod with the actual? Non-test code can return
	// the grpc client function (using localhost?) while test code can return
	// something quicker?
	corePackagesClientGetter func() (corev1alpha1.PackagesServiceClient, error)
}

func NewServer(configGetter server.KubernetesConfigGetter) *Server {
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
		corePackagesClientGetter: func() (corev1alpha1.PackagesServiceClient, error) {
			port := os.Getenv("PORT")
			conn, err := grpc.Dial("localhost:"+port, grpc.WithInsecure())
			if err != nil {
				return nil, err
			}
			return corev1alpha1.NewPackagesServiceClient(conn), nil
		},
	}
}

// GetResources returns the resources for an installed package.
func (s *Server) GetResources(r *v1alpha1.GetResourcesRequest, stream v1alpha1.ResourcesService_GetResourcesServer) error {
	namespace := r.GetInstalledPackageRef().GetContext().GetNamespace()
	cluster := r.GetInstalledPackageRef().GetContext().GetCluster()
	log.Infof("+resources GetResources (cluster: %q, namespace=%q)", cluster, namespace)

	coreClient, err := s.corePackagesClientGetter()
	if err != nil {
		return err
	}
	// Need to explicitly copy the authz from the incoming to outgoing context metadata.
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return err
	}
	ctxOut := stream.Context()
	if len(md["authorization"]) > 0 {
		ctxOut = metadata.AppendToOutgoingContext(ctxOut, "authorization", md["authorization"][0])
	}
	refsResponse, err := coreClient.GetInstalledPackageResourceRefs(ctxOut, &corev1alpha1.GetInstalledPackageResourceRefsRequest{
		InstalledPackageRef: r.InstalledPackageRef,
	})
	if err != nil {
		return err
	}

	_, dynamicClient, err := s.clientGetter(stream.Context(), cluster)
	if err != nil {
		return err
	}

	var watchers []watch.Interface
	// TODO: Filter to the resources specified in the request, and 400 if they don't
	// exist for the package.
	for _, ref := range refsResponse.GetResourceRefs() {
		groupVersion, err := schema.ParseGroupVersion(ref.ApiVersion)
		if err != nil {
			// TODO: status code.
			return err
		}
		gvk := groupVersion.WithKind(ref.Kind)
		// Could use the UnsafeGuessKindToResource, but is there a better way
		// to use the default rest mapper?
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		resource, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(stream.Context(), ref.GetName(), metav1.GetOptions{})
		if err != nil {
			// TODO status code
			return err
		}

		resourceBytes, err := json.Marshal(resource)
		if err != nil {
			// TODO status code
			return err
		}
		stream.Send(&v1alpha1.GetResourcesResponse{
			ResourceRef: ref,
			Manifest: &anypb.Any{
				// TODO: URL for k8s resources?
				Value: resourceBytes,
			},
		})

		// If watching
		if r.GetWatch() {
			watcher, err := dynamicClient.Resource(gvr).Namespace(namespace).Watch(stream.Context(), metav1.ListOptions{})
			if err != nil {
				// TODO status code
				return err
			}
			watchers = append(watchers, watcher)
		}

	}

	// If we're not watching, we're done.
	if watchers == nil {
		return nil
	}

	watcher := mergeWatchers(watchers...)
	for e := range watcher.ResultChan() {
		resourceBytes, err := json.Marshal(e.Object)
		if err != nil {
			// TODO status code
			return err
		}
		stream.Send(&v1alpha1.GetResourcesResponse{
			ResourceRef: nil,
			Manifest: &anypb.Any{
				// TODO: URL for k8s resources?
				Value: resourceBytes,
			},
		})
	}

	return nil
}

// mergeWatchers returns a single watcher for many.
// Inspired by the fan-in merge at https://go.dev/blog/pipelines
func mergeWatchers(watchers ...watch.Interface) watch.Interface {
	var wg sync.WaitGroup
	out := make(chan watch.Event)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan watch.Event) {
		for n := range c {
			out <- n
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

	return &MultiWatcher{
		StopFn: func() {
			for _, w := range watchers {
				w.Stop()
			}
		},
		ResultChanFn: func() <-chan watch.Event {
			return out
		},
	}
}

type MultiWatcher struct {
	StopFn       func()
	ResultChanFn func() <-chan watch.Event
}

func (mw *MultiWatcher) Stop() {
	mw.StopFn()
}

func (mw *MultiWatcher) ResultChan() <-chan watch.Event {
	return mw.ResultChanFn()
}

// ResourceWatcher is a watcher that knows the resource its watching.
type ResourceWatcher struct {
	ResourceRef corev1alpha1.ResourceRef
	Watcher     watch.Interface
}

func (rw *ResourceWatcher) Stop() {
	rw.Watcher.Stop()
}

func (rw *ResourceWatcher) ResultChan() <-chan watch.Event {
	return rw.Watcher.ResultChan()
}
