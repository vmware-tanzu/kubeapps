// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"github.com/vmware-tanzu/kubeapps/pkg/kube"
	"io"
	"k8s.io/client-go/rest"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"

	pkgsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugin_test"
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
			Namespace:  objMeta.Namespace,
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
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
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
	t.Logf("loading fake client with objects: %+v", objects)
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
		// For testing, define a kindToResource converter that doesn't require
		// a rest mapper.
		kindToResource: func(mapper meta.RESTMapper, gvk schema.GroupVersionKind) (schema.GroupVersionResource, meta.RESTScopeName, error) {
			gvr, _ := meta.UnsafeGuessKindToResource(gvk)
			return gvr, meta.RESTScopeNameNamespace, nil
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
						Namespace:  "default",
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
						Namespace:  "default",
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
		{
			name: "it gets requested resources from different namespaces when they belong to the installed package",
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
						Namespace:  "other-namespace",
					},
				},
			},
			clusterObjects: []runtime.Object{
				&core.Service{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Service",
						APIVersion: "core/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-service",
						Namespace: "other-namespace",
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
						Namespace:  "other-namespace",
					},
				},
			},
		},
		{
			name: "it gets non-namespaced requested resources when they belong to the installed package",
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
						ApiVersion: "rbac/v1",
						Kind:       "ClusterRole",
						Name:       "some-cluster-role",
					},
				},
			},
			clusterObjects: []runtime.Object{
				&rbac.ClusterRole{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ClusterRole",
						APIVersion: "rbac/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "some-cluster-role",
					},
				},
			},
			expectedErrorCode: codes.OK,
			expectedResources: []*v1alpha1.GetResourcesResponse{
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
			err = responseStream.CloseSend()
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := resources, tc.expectedResources; !cmp.Equal(got, want, ignoredUnexported, ignoreManifest) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoreManifest))
			}
		})
	}
}

func TestGetServiceAccountNames(t *testing.T) {
	testCases := []struct {
		name               string
		request            *v1alpha1.GetServiceAccountNamesRequest
		existingObjects    []runtime.Object
		expectedResponse   []string
		expectedStatusCode codes.Code
	}{
		{
			name: "returns expected SAs",
			request: &v1alpha1.GetServiceAccountNamesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []runtime.Object{
				&core.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "some-service-account",
						Namespace: "default",
					},
				},
				&core.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "another-service-account",
						Namespace: "default",
					},
				},
				&core.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
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
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typfake.NewSimpleClientset(tc.existingObjects...), nil, nil
				},
			}

			GetServiceAccountNamesResponse, err := s.GetServiceAccountNames(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.expectedStatusCode == codes.OK {
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

func TestSetupConfigForCluster(t *testing.T) {
	testCases := []struct {
		name               string
		restConfig         *rest.Config
		cluster            string
		useServiceAccount  bool
		clustersConfig     kube.ClustersConfig
		expectedRestConfig *rest.Config
		expectedErrorCode  codes.Code
	}{
		{
			name:    "config is not modified for kubeapps cluster",
			cluster: "default",
			clustersConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"default": {},
				},
			},
			restConfig:         &rest.Config{},
			expectedRestConfig: &rest.Config{},
		},
		{
			name:    "config is not modified for additional clusters and no service account",
			cluster: "additional-1",
			clustersConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"default":      {},
					"additional-1": {},
				},
			},
			restConfig:         &rest.Config{},
			expectedRestConfig: &rest.Config{},
		},
		{
			name:              "config setup fails for additional clusters with no cluster config data",
			cluster:           "additional-1",
			useServiceAccount: true,
			clustersConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"default": {},
				},
			},
			restConfig:         &rest.Config{},
			expectedRestConfig: &rest.Config{},
			expectedErrorCode:  codes.Internal,
		},
		{
			name:              "config is not modified for additional clusters with no configured service token",
			cluster:           "additional-1",
			useServiceAccount: true,
			clustersConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"default":      {},
					"additional-1": {},
				},
			},
			restConfig:         &rest.Config{},
			expectedRestConfig: &rest.Config{},
		},
		{
			name:              "config is modified for additional clusters when configured service token",
			cluster:           "additional-1",
			useServiceAccount: true,
			clustersConfig: kube.ClustersConfig{
				KubeappsClusterName: "default",
				Clusters: map[string]kube.ClusterConfig{
					"default": {},
					"additional-1": {
						ServiceToken: "service-token-1",
					},
				},
			},
			restConfig: &rest.Config{},
			expectedRestConfig: &rest.Config{
				BearerToken: "service-token-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			err := setupRestConfigForCluster(tc.restConfig, tc.cluster, tc.useServiceAccount, tc.clustersConfig)

			if got, want := status.Code(err), tc.expectedErrorCode; !cmp.Equal(got, want, nil) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, nil))
			}

			if got, want := tc.restConfig, tc.expectedRestConfig; !cmp.Equal(got, want, nil) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, nil))
			}
		})
	}
}
