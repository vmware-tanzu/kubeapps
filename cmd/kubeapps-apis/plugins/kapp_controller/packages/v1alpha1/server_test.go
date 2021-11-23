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
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
)

var ignoreUnexported = cmpopts.IgnoreUnexported(
	corev1.AvailablePackageDetail{},
	corev1.AvailablePackageReference{},
	corev1.AvailablePackageSummary{},
	corev1.Context{},
	corev1.Maintainer{},
	corev1.PackageAppVersion{},
	pluginv1.Plugin{},
)

var defaultContext = &corev1.Context{Cluster: "default", Namespace: "default"}

var datapackagingAPIVersion = fmt.Sprintf("%s/%s", datapackagingv1alpha1.SchemeGroupVersion.Group, datapackagingv1alpha1.SchemeGroupVersion.Version)
var packagingAPIVersion = fmt.Sprintf("%s/%s", packagingv1alpha1.SchemeGroupVersion.Group, packagingv1alpha1.SchemeGroupVersion.Version)
var kappctrlAPIVersion = fmt.Sprintf("%s/%s", kappctrlv1alpha1.SchemeGroupVersion.Group, kappctrlv1alpha1.SchemeGroupVersion.Version)

func TestGetClient(t *testing.T) {
	testClientGetter := func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
		return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}

	testCases := []struct {
		name              string
		clientGetter      clientGetter
		statusCodeClient  codes.Code
		statusCodeManager codes.Code
	}{
		{
			name:              "it returns internal error status when no clientGetter configured",
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.OK,
		},
		{
			name:              "it returns internal error status when no manager configured",
			clientGetter:      testClientGetter,
			statusCodeClient:  codes.OK,
			statusCodeManager: codes.Internal,
		},
		{
			name:              "it returns internal error status when no clientGetter/manager configured",
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.Internal,
		},
		{
			name: "it returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCodeClient:  codes.FailedPrecondition,
			statusCodeManager: codes.OK,
		},
		{
			name:         "it returns client without error when configured correctly",
			clientGetter: testClientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter}

			typedClient, dynamicClient, errClient := s.GetClients(context.Background(), "")

			if got, want := status.Code(errClient), tc.statusCodeClient; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCodeClient == codes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: dynamic.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: kubernetes.Interface")
				}
			}
		})
	}
}
