// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	resourcesGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	plugintest "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"
	grpc "google.golang.org/grpc"
	grpccodes "google.golang.org/grpc/codes"
	grpcmetadata "google.golang.org/grpc/metadata"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	k8sappsv1 "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8srbacv1 "k8s.io/api/rbac/v1"
	k8smeta "k8s.io/apimachinery/pkg/api/meta"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8dynamicclientfake "k8s.io/client-go/dynamic/fake"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
)

const bufSize = 1024 * 1024

var fakePkgsPlugin *pluginsGRPCv1alpha1.Plugin = &pluginsGRPCv1alpha1.Plugin{Name: "fake.packages", Version: "v1alpha1"}

func resourceRefsForObjects(t *testing.T, objects ...k8sruntime.Object) []*pkgsGRPCv1alpha1.ResourceRef {
	refs := []*pkgsGRPCv1alpha1.ResourceRef{}
	for _, obj := range objects {
		k8sObjValue := reflect.ValueOf(obj).Elem()
		objMeta, ok := k8sObjValue.FieldByName("ObjectMeta").Interface().(k8smetav1.ObjectMeta)
		if !ok {
			t.Fatalf("failed to retrieve object metadata for: %+v", objMeta)
		}
		gvk := obj.GetObjectKind().GroupVersionKind()
		refs = append(refs, &pkgsGRPCv1alpha1.ResourceRef{
			ApiVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       objMeta.Name,
			Namespace:  objMeta.Namespace,
		})
	}
	return refs
}

// getResourcesClient starts a GRPC server, serving both the resources service,
// and a test core packages service, but using a buf connection (ie. no need for
// slow network port etc.). More at
// https://stackoverflow.com/a/52080545
func getResourcesClient(t *testing.T, objects ...k8sruntime.Object) (resourcesGRPCv1alpha1.ResourcesServiceClient, *k8dynamicclientfake.FakeDynamicClient, func()) {
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
	fakePkgsPluginServer := &plugintest.TestPackagingPluginServer{
		Plugin:       fakePkgsPlugin,
		ResourceRefs: resourceRefsForObjects(t, objects...),
	}
	pkgsGRPCv1alpha1.RegisterPackagesServiceServer(s, fakePkgsPluginServer)

	scheme := k8sruntime.NewScheme()
	t.Logf("loading fake client with objects: %+v", objects)
	fakeDynamicClient := k8dynamicclientfake.NewSimpleDynamicClient(
		scheme,
		objects...,
	)
	// Create the resources service server.
	resourcesGRPCv1alpha1.RegisterResourcesServiceServer(s, &Server{
		// Use a client getter that returns a dynamic client prepped with the
		// specified objects.
		clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
			return nil, fakeDynamicClient, nil
		},
		// Use a corePackagesClientGetter that returns a client connected to our
		// running test service.
		corePackagesClientGetter: func() (pkgsGRPCv1alpha1.PackagesServiceClient, error) {
			return pkgsGRPCv1alpha1.NewPackagesServiceClient(conn), nil
		},
		// For testing, define a kindToResource converter that doesn't require
		// a rest mapper.
		kindToResource: func(mapper k8smeta.RESTMapper, gvk k8sschema.GroupVersionKind) (k8sschema.GroupVersionResource, k8smeta.RESTScopeName, error) {
			gvr, _ := k8smeta.UnsafeGuessKindToResource(gvk)
			return gvr, k8smeta.RESTScopeNameNamespace, nil
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

	return resourcesGRPCv1alpha1.NewResourcesServiceClient(conn), fakeDynamicClient, func() {
		conn.Close()
		lis.Close()
	}
}

func TestGetResources(t *testing.T) {
	testCases := []struct {
		name              string
		request           *resourcesGRPCv1alpha1.GetResourcesRequest
		withoutAuthz      bool
		clusterObjects    []k8sruntime.Object
		expectedErrorCode grpccodes.Code
		expectedResources []*resourcesGRPCv1alpha1.GetResourcesResponse
	}{
		{
			name: "it returns permission denied for a request without auth",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
			expectedErrorCode: grpccodes.PermissionDenied,
		},
		{
			name: "it gets all resources for an installed app when the filter is empty",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Cluster:   "default",
						Namespace: "default",
					},
					Identifier: "some-package",
					Plugin:     fakePkgsPlugin,
				},
			},
			clusterObjects: []k8sruntime.Object{
				&k8sappsv1.Deployment{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
				&k8scorev1.Service{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: grpccodes.OK,
			expectedResources: []*resourcesGRPCv1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
						Namespace:  "default",
					},
				},
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
						Namespace:  "default",
					},
				},
			},
		},
		{
			name: "it gets only requested resources for an installed app when the filter is specified",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
						Namespace:  "default",
					},
				},
			},
			clusterObjects: []k8sruntime.Object{
				&k8sappsv1.Deployment{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
				&k8scorev1.Service{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: grpccodes.OK,
			expectedResources: []*resourcesGRPCv1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
						Namespace:  "default",
					},
				},
			},
		},
		{
			name: "it returns invalid argument if a requested resource isn't part of the installed package",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
			clusterObjects: []k8sruntime.Object{
				&k8sappsv1.Deployment{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-deployment",
						Namespace: "default",
					},
				},
			},
			expectedErrorCode: grpccodes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if request is to watch all packages implicitly (ie. empty resource refs filter in request)",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
			expectedErrorCode: grpccodes.InvalidArgument,
		},
		{
			name: "it gets requested resources from different namespaces when they belong to the installed package",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
						Namespace:  "other-namespace",
					},
				},
			},
			clusterObjects: []k8sruntime.Object{
				&k8scorev1.Service{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "other-namespace",
					},
				},
			},
			expectedErrorCode: grpccodes.OK,
			expectedResources: []*resourcesGRPCv1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
						Namespace:  "other-namespace",
					},
				},
			},
		},
		{
			name: "it gets non-namespaced requested resources when they belong to the installed package",
			request: &resourcesGRPCv1alpha1.GetResourcesRequest{
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
						ApiVersion: "rbac/v1",
						Kind:       "ClusterRole",
						Name:       "some-cluster-role",
					},
				},
			},
			clusterObjects: []k8sruntime.Object{
				&k8srbacv1.ClusterRole{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "ClusterRole",
						APIVersion: "rbac/v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "some-cluster-role",
					},
				},
			},
			expectedErrorCode: grpccodes.OK,
			expectedResources: []*resourcesGRPCv1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsGRPCv1alpha1.ResourceRef{
						ApiVersion: "rbac/v1",
						Kind:       "ClusterRole",
						Name:       "some-cluster-role",
					},
				},
			},
		},
		// TODO(minelson): test a watch request also. I've spent quite a bit of
		// time trying to do so by putting the call to `GetResources` in a go
		// routine (and passing the results out via response and error channels)
		// and then deleting the resources via the fake k8s client's object
		// tracker. From the source code, this should trigger the watch event,
		// but I didn't succeed (yet).
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		resourcesGRPCv1alpha1.GetResourcesResponse{},
		pkgsGRPCv1alpha1.ResourceRef{},
	)
	ignoreManifest := cmpopts.IgnoreFields(resourcesGRPCv1alpha1.GetResourcesResponse{}, "Manifest")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, _, cleanup := getResourcesClient(t, tc.clusterObjects...)
			defer cleanup()

			// Use a context with a timeout to ensure that if a test unexpectedly
			// waits beyond expectations, we'll see the failure.
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if !tc.withoutAuthz {
				ctx = grpcmetadata.AppendToOutgoingContext(ctx, "authorization", "some-auth-token")
			}

			responseStream, err := client.GetResources(ctx, tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			var resources []*resourcesGRPCv1alpha1.GetResourcesResponse
			for numResponses := 0; numResponses < len(tc.expectedResources); numResponses++ {
				resource, err := responseStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					if got, want := grpcstatus.Code(err), tc.expectedErrorCode; got != want {
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

func TestGetServiceAccountNames(t *testing.T) {
	testCases := []struct {
		name               string
		request            *resourcesGRPCv1alpha1.GetServiceAccountNamesRequest
		existingObjects    []k8sruntime.Object
		expectedResponse   []string
		expectedStatusCode grpccodes.Code
	}{
		{
			name: "returns expected SAs",
			request: &resourcesGRPCv1alpha1.GetServiceAccountNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&k8scorev1.ServiceAccount{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "some-service-account",
						Namespace: "default",
					},
				},
				&k8scorev1.ServiceAccount{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "another-service-account",
						Namespace: "default",
					},
				},
				&k8scorev1.ServiceAccount{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "unwanted-service-account",
						Namespace: "other-ns",
					},
				},
			},
			expectedResponse: []string{"another-service-account", "some-service-account"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return k8stypedclientfake.NewSimpleClientset(tc.existingObjects...), nil, nil
				},
			}

			GetServiceAccountNamesResponse, err := s.GetServiceAccountNames(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			// Only check the response for OK grpcstatus.
			if tc.expectedStatusCode == grpccodes.OK {
				if GetServiceAccountNamesResponse == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					if got, want := GetServiceAccountNamesResponse.ServiceaccountNames, tc.expectedResponse; !cmp.Equal(got, want, nil) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, nil))
					}
				}
			}
		})
	}
}
