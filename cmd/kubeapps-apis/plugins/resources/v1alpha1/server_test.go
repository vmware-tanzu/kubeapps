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
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"
)

const bufSize = 1024 * 1024

var fakePkgsPlugin *pluginsGRPCv1alpha1.Plugin = &pluginsGRPCv1alpha1.Plugin{Name: "fake.packages", Version: "v1alpha1"}

func resourceRefsForObjects(t *testing.T, objects ...runtime.Object) []*pkgsGRPCv1alpha1.ResourceRef {
	refs := []*pkgsGRPCv1alpha1.ResourceRef{}
	for _, obj := range objects {
		k8sObjValue := reflect.ValueOf(obj).Elem()
		objMeta, ok := k8sObjValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
		if !ok {
			t.Fatalf("failed to retrieve object metadata for: %+v", objMeta)
		}
		gvk := obj.GetObjectKind().GroupVersionKind()
		refs = append(refs, &pkgsGRPCv1alpha1.ResourceRef{
			ApiVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       objMeta.Name,
		})
	}
	return refs
}

// getResourcesClient starts a GRPC server, serving both the resources service,
// and a test core packages service, but using a buf connection (ie. no need for
// slow network port etc.). More at
// https://stackoverflow.com/a/52080545
func getResourcesClient(t *testing.T, objects ...runtime.Object) (v1alpha1.ResourcesServiceClient, *dynfake.FakeDynamicClient, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	// Create and register a fake packages plugin returning resourcerefs for objects.
	fakePkgsPluginServer := &plugin_test.TestPackagingPluginServer{
		Plugin:       fakePkgsPlugin,
		ResourceRefs: resourceRefsForObjects(t, objects...),
	}
	pkgsGRPCv1alpha1.RegisterPackagesServiceServer(s, fakePkgsPluginServer)

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynfake.NewSimpleDynamicClient(
		scheme,
		objects...,
	)
	// Create the resources service server.
	v1alpha1.RegisterResourcesServiceServer(s, &Server{
		// Use a client getter that returns a dynamic client prepped with the
		// specified objects.
		clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
			return nil, fakeDynamicClient, nil
		},
		// Use a corePackagesClientGetter that returns a client connected to our
		// running test service.
		corePackagesClientGetter: func() (pkgsGRPCv1alpha1.PackagesServiceClient, error) {
			return pkgsGRPCv1alpha1.NewPackagesServiceClient(conn), nil
		},
	})

	go func() {
		if err := s.Serve(lis); err != nil {
			// Only valid error should be when the listener is closed.
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				return
			}
		}
	}()

	return v1alpha1.NewResourcesServiceClient(conn), fakeDynamicClient, func() {
		conn.Close()
		lis.Close()
	}
}

func TestGetResources(t *testing.T) {
	testCases := []struct {
		name              string
		request           *v1alpha1.GetResourcesRequest
		withoutAuthz      bool
		clusterObjects    []runtime.Object
		expectedErrorCode codes.Code
		expectedResources []*v1alpha1.GetResourcesResponse
	}{
		{
			name: "it returns permission denied for a request without auth",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
			},
			withoutAuthz:      true,
			expectedErrorCode: codes.PermissionDenied,
		},
		{
			name: "it gets all resources for an installed app when the filter is empty",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
			},
			clusterObjects: []runtime.Object{
				&apps.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
				&core.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: codes.OK,
			expectedResources: []*v1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
				},
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
					},
				},
			},
		},
		{
			name: "it gets only requested resources for an installed app when the filter is specified",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
				ResourceRefs: []*pkgsGRPCv1alpha1.ResourceRef{
					{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
					},
				},
			},
			clusterObjects: []runtime.Object{
				&apps.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
				&core.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: codes.OK,
			expectedResources: []*v1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
					},
				},
			},
		},
		{
			name: "it returns invalid argument if a requested resource isn't part of the installed package",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
				ResourceRefs: []*pkgsGRPCv1alpha1.ResourceRef{
					{
						ApiVersion: "core/v1",
						Kind:       "Secret",
						Name:       "some-secret",
					},
				},
			},
			clusterObjects: []runtime.Object{
				&apps.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if request is to watch all packages implicitly (ie. empty resource refs filter in request)",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
				Watch: true,
			},
			expectedErrorCode: codes.InvalidArgument,
		},
		// TODO(minelson): test a watch request also. I've spent quite a bit of
		// time trying to do so by putting the call to `GetResources` in a go
		// routine (and passing the results out via response and error channels)
		// and then deleting the resources via the fake k8s client's object
		// tracker. From the source code, this should trigger the watch event,
		// but I didn't succeed (yet).
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		v1alpha1.GetResourcesResponse{},
		pkgsGRPCv1alpha1.ResourceRef{},
	)
	ignoreManifest := cmpopts.IgnoreFields(v1alpha1.GetResourcesResponse{}, "Manifest")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, _, cleanup := getResourcesClient(t, tc.clusterObjects...)
			defer cleanup()

			// Use a context with a timeout to ensure that if a test unexpectedly
			// waits beyond expectations, we'll see the failure.
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if !tc.withoutAuthz {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "some-auth-token")
			}

			responseStream, err := client.GetResources(ctx, tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			var resources []*v1alpha1.GetResourcesResponse
			for numResponses := 0; numResponses < len(tc.expectedResources); numResponses++ {
				resource, err := responseStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if got, want := status.Code(err), tc.expectedErrorCode; got != want {
						t.Fatalf("got: %s, want: %s, err: %+v", got, want, err)
					}
					// If it was an expected error, we continue to the next test.
					return
				}
				resources = append(resources, resource)
			}
			responseStream.CloseSend()

			if got, want := resources, tc.expectedResources; !cmp.Equal(got, want, ignoredUnexported, ignoreManifest) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoreManifest))
			}
		})
	}
}
