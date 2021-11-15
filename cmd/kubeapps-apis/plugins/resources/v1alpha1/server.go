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
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog/v2"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/core"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
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
}

func NewServer(configGetter core.KubernetesConfigGetter) *Server {
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
	}
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

	// Then look up each referenced resource and send it down the stream.
	// TODO(minelson): Filter to the resources specified in the request, and 400
	// if they don't exist for the package.
	_, dynamicClient, err := s.clientGetter(stream.Context(), cluster)
	if err != nil {
		return err
	}
	for _, ref := range refsResponse.GetResourceRefs() {
		groupVersion, err := schema.ParseGroupVersion(ref.ApiVersion)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to parse group version from %q: %s", ref.ApiVersion, err.Error())
		}
		gvk := groupVersion.WithKind(ref.Kind)
		// TODO(minelson): Find alternative to UnsafeGuessKindToResource.
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		resource, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(stream.Context(), ref.GetName(), metav1.GetOptions{})
		if err != nil {
			return status.Errorf(codes.Internal, "unable to get resource referenced by %+v: %s", ref, err.Error())
		}

		resourceBytes, err := json.Marshal(resource)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to marshal json for resource: %s", err.Error())
		}
		stream.Send(&v1alpha1.GetResourcesResponse{
			ResourceRef: ref,
			Manifest: &anypb.Any{
				Value: resourceBytes,
			},
		})
	}

	return nil
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
