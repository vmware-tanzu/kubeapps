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
	"log"
	"net"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"

	pkgsv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"
	coreserver "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
)

const bufSize = 1024 * 1024

var fakePkgsPlugin *plugins.Plugin = &plugins.Plugin{Name: "fake.packages", Version: "v1alpha1"}

func resourceRefsForObjects(t *testing.T, objects ...runtime.Object) []*pkgsv1alpha1.ResourceRef {
	refs := []*pkgsv1alpha1.ResourceRef{}
	for _, obj := range objects {
		k8sObjValue := reflect.ValueOf(obj).Elem()
		objMeta, ok := k8sObjValue.FieldByName("ObjectMeta").Interface().(metav1.ObjectMeta)
		if !ok {
			t.Fatalf("failed to retrieve object metadata for: %+v", objMeta)
		}
		gvk := obj.GetObjectKind().GroupVersionKind()
		refs = append(refs, &pkgsv1alpha1.ResourceRef{
			ApiVersion: gvk.GroupVersion().String(),
			Kind:       gvk.Kind,
			Name:       objMeta.Name,
		})
	}
	return refs
}

func NewTestPluginsServer(t *testing.T, pkgsPlugins []*coreserver.PkgsPluginWithServer) *coreserver.PluginsServer {
	plugins := []*plugins.Plugin{}
	for _, p := range pkgsPlugins {
		plugins = append(plugins, p.Plugin)
	}
	return &coreserver.PluginsServer{
		Plugins:         plugins,
		PackagesPlugins: pkgsPlugins,
	}
}

// getResourcesClient starts a GRPC server, serving the resources service, but
// using the buf connection (ie. no need for slow network port etc.). More at
// https://stackoverflow.com/a/52080545
func getResourcesClient(t *testing.T, objects ...runtime.Object) (v1alpha1.ResourcesServiceClient, func()) {
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

	// Create fake packaging plugin returning resourcerefs for objects.
	fakePkgsPluginServer := &plugin_test.TestPackagingPluginServer{
		Plugin:       fakePkgsPlugin,
		ResourceRefs: resourceRefsForObjects(t, objects...),
	}

	// Create a fake core packaging service using the fake packages plugin.
	pkgsPlugins := []*coreserver.PkgsPluginWithServer{
		{
			Plugin: fakePkgsPlugin,
			Server: fakePkgsPluginServer,
		},
	}
	plugins.RegisterPluginsServiceServer(s, NewTestPluginsServer(t, pkgsPlugins))
	corePkgsServer := coreserver.NewPackagesServer(pkgsPlugins)
	pkgsv1alpha1.RegisterPackagesServiceServer(s, corePkgsServer)

	v1alpha1.RegisterResourcesServiceServer(s, &Server{
		// Create a client getter that returns a dynamic client prepped with the
		// specified objects.
		clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
			scheme := runtime.NewScheme()
			fakeDynamicClient := dynfake.NewSimpleDynamicClient(
				scheme,
				objects...,
			)
			return nil, fakeDynamicClient, nil
		},
		corePackagesClientGetter: func() (pkgsv1alpha1.PackagesServiceClient, error) {
			return pkgsv1alpha1.NewPackagesServiceClient(conn), nil
		},
	})

	go func() {
		if err := s.Serve(lis); err != nil {
			// Only valid error should be when the listener is closed.
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				return
			}
			log.Printf("Server exited with error: %+v", err)
		}
	}()

	return v1alpha1.NewResourcesServiceClient(conn), func() {
		conn.Close()
		lis.Close()
	}
}

func TestGetResources(t *testing.T) {
	testCases := []struct {
		name              string
		request           *v1alpha1.GetResourcesRequest
		clusterObjects    []runtime.Object
		expectedResources []*v1alpha1.GetResourcesResponse
	}{
		{
			name: "it returns resources for an installed app",
			request: &v1alpha1.GetResourcesRequest{
				InstalledPackageRef: &pkgsv1alpha1.InstalledPackageReference{
					Context: &pkgsv1alpha1.Context{
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
			expectedResources: []*v1alpha1.GetResourcesResponse{
				{
					ResourceRef: &pkgsv1alpha1.ResourceRef{
						ApiVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "some-deployment",
					},
				},
				{
					ResourceRef: &pkgsv1alpha1.ResourceRef{
						ApiVersion: "core/v1",
						Kind:       "Service",
						Name:       "some-service",
					},
				},
			},
		},
	}

	ignoredUnexported := cmpopts.IgnoreUnexported(
		v1alpha1.GetResourcesResponse{},
		pkgsv1alpha1.ResourceRef{},
	)
	ignoreManifest := cmpopts.IgnoreFields(v1alpha1.GetResourcesResponse{}, "Manifest")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, cleanup := getResourcesClient(t, tc.clusterObjects...)
			defer cleanup()

			responseStream, err := client.GetResources(context.Background(), tc.request)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			resources := []*v1alpha1.GetResourcesResponse{}
			for {
				resource, err := responseStream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Unexpected error: %+v", err)
				}
				resources = append(resources, resource)
			}

			if got, want := resources, tc.expectedResources; !cmp.Equal(got, want, ignoredUnexported, ignoreManifest) {
				t.Errorf("Did comparison, printing mismatch")
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoredUnexported, ignoreManifest))
			}
		})
	}
}
