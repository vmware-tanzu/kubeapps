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
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirversions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
	corev1.InstalledPackageSummary{},
	corev1.InstalledPackageReference{},
	corev1.InstalledPackageStatus{},
	corev1.InstalledPackageDetail{},
	corev1.ReconciliationOptions{},
	corev1.CreateInstalledPackageResponse{},
	corev1.DeleteInstalledPackageResponse{},
	corev1.UpdateInstalledPackageResponse{},
	corev1.VersionReference{},
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
				{Group: "foo", Version: "bar", Resource: "baz"}: "fooList",
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

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		existingObjects    []runtime.Object
		expectedPackages   []*corev1.AvailablePackageSummary
		expectedStatusCode codes.Code
	}{
		{
			name: "it returns a not found error status if a package meta does not contain spec.displayName",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						LongDescription: "Classic Tetris",
					},
				},
			},
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "it returns an not found error status if a package does not contain version",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "Classic Tetris",
					},
				},
			},
			expectedStatusCode: codes.NotFound,
		},
		{
			name: "it returns carvel package summaries with basic info from the cluster",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "Classic Tetris",
					},
				},
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tombi.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "Tombi!",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tombi.foo.example.com.1.2.5",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tombi.foo.example.com",
						Version: "1.2.5",
					},
				},
			},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:          "tetris.foo.example.com",
					DisplayName:   "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
				},
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tombi.foo.example.com",
					},
					Name:          "tombi.foo.example.com",
					DisplayName:   "Tombi!",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.5"},
				},
			},
		},
		{
			name: "it returns carvel package summaries with complete metadata",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
			},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:             "tetris.foo.example.com",
					DisplayName:      "Classic Tetris",
					LatestVersion:    &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					Categories:       []string{"logging", "daemon-set"},
				},
			},
		},
		{
			name: "it returns the latest semver version in the latest version field",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName: "Classic Tetris",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.7",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.4",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.4",
					},
				},
			},
			expectedPackages: []*corev1.AvailablePackageSummary{
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:          "tetris.foo.example.com",
					DisplayName:   "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.7"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: defaultContext})

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			if got, want := response.AvailablePackageSummaries, tc.expectedPackages; !cmp.Equal(got, want, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		existingObjects    []runtime.Object
		request            *corev1.GetAvailablePackageVersionsRequest
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name:               "it returns invalid argument if called without a package reference",
			request:            nil,
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without namespace",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{},
					Identifier: "package-one",
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without an identifier",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name: "it returns the package version summary",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.7",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.4",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.4",
					},
				},
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "default",
					},
					Identifier: "tetris.foo.example.com",
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{
					{
						PkgVersion: "1.2.7",
					},
					{
						PkgVersion: "1.2.4",
					},
					{
						PkgVersion: "1.2.3",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}: pkgResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		name            string
		existingObjects []runtime.Object
		expectedPackage *corev1.AvailablePackageDetail
		statusCode      codes.Code
		request         *corev1.GetAvailablePackageDetailRequest
	}{
		{
			name: "it returns an availablePackageDetail of the latest version",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:      "Classic Tetris",
						IconSVGBase64:    "Tm90IHJlYWxseSBTVkcK",
						ShortDescription: "A great game for arcade gamers",
						LongDescription:  "A few sentences but not really a readme",
						Categories:       []string{"logging", "daemon-set"},
						Maintainers:      []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
			},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:             "tetris.foo.example.com",
				DisplayName:      "Classic Tetris",
				IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
				Categories:       []string{"logging", "daemon-set"},
				ShortDescription: "A great game for arcade gamers",
				LongDescription:  "A few sentences but not really a readme",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
				},
				Readme: fmt.Sprintf(`## Details


### Capactiy requirements:
%s


### Release Notes:
%s


### Licenses:
%s


### ReleasedAt:
%s


`,
					"",
					"",
					[]string{""},
					&metav1.Time{},
				),
				Maintainers: []*corev1.Maintainer{
					{Name: "person1"},
					{Name: "person2"},
				},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginv1.Plugin{Name: "kapp_controller.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it combines long description and support description for readme field",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						LongDescription:    "A few sentences but not really a readme",
						SupportDescription: "Some support info",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
			},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:            "tetris.foo.example.com",
				DisplayName:     "Classic Tetris",
				LongDescription: "A few sentences but not really a readme",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
				},
				Maintainers: []*v1alpha1.Maintainer{},
				Readme: fmt.Sprintf(`## Details


### Capactiy requirements:
%s


### Release Notes:
%s


### Licenses:
%s


### ReleasedAt:
%s


`,
					"",
					"",
					[]string{""},
					&metav1.Time{},
				),
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginv1.Plugin{Name: "kapp_controller.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it returns an invalid arg error status if no context is provided",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "foo/bar",
				},
			},
			statusCode: codes.InvalidArgument,
		},
		{
			name: "it returns not found error status if the requested package version doesn't exist",
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersion: "1.2.4",
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						LongDescription:    "A few sentences but not really a readme",
						SupportDescription: "Some support info",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
			},
			statusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}
			availablePackageDetail, err := s.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := availablePackageDetail.AvailablePackageDetail, tc.expectedPackage; !cmp.Equal(got, want, ignoreUnexported) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
				}
			}
		})
	}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		existingObjects    []runtime.Object
		expectedPackages   []*corev1.InstalledPackageSummary
		expectedStatusCode codes.Code
	}{
		{
			name: "it returns carvel empty installed package summary when no package install is present",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
			},
			expectedPackages: []*corev1.InstalledPackageSummary{},
		},
		{
			name: "it returns carvel installed package summary with complete metadata",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
			},
			expectedPackages: []*corev1.InstalledPackageSummary{
				{
					InstalledPackageRef: &corev1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:                  "my-installation",
					PkgDisplayName:        "Classic Tetris",
					LatestVersion:         &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					IconUrl:               "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:      "A great game for arcade gamers",
					PkgVersionReference:   &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion:        &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					LatestMatchingVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Reconcile succeeded",
					},
				},
			},
		},
		{
			name: "it returns carvel installed package summary with a packageInstall without status",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
				},
			},
			expectedPackages: []*corev1.InstalledPackageSummary{
				{
					InstalledPackageRef: &corev1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:                  "my-installation",
					PkgDisplayName:        "Classic Tetris",
					LatestVersion:         &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					IconUrl:               "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:      "A great game for arcade gamers",
					PkgVersionReference:   &corev1.VersionReference{Version: ""},
					CurrentVersion:        &corev1.PackageAppVersion{PkgVersion: ""},
					LatestMatchingVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					Status: &corev1.InstalledPackageStatus{
						Ready:      false,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
						UserReason: "no status information yet",
					},
				},
			},
		},
		{
			name: "it returns the latest semver version in the latest version field",
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.3",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.7",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.4",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName: "tetris.foo.example.com",
						Version: "1.2.4",
					},
				},
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
			},
			expectedPackages: []*corev1.InstalledPackageSummary{
				{
					InstalledPackageRef: &corev1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:                  "my-installation",
					PkgDisplayName:        "Classic Tetris",
					LatestVersion:         &corev1.PackageAppVersion{PkgVersion: "1.2.7"},
					IconUrl:               "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:      "A great game for arcade gamers",
					PkgVersionReference:   &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion:        &corev1.PackageAppVersion{PkgVersion: "1.2.3"},
					LatestMatchingVersion: &corev1.PackageAppVersion{PkgVersion: "1.2.7"},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Reconcile succeeded",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return nil, dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetInstalledPackageSummaries(context.Background(), &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext})

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			if got, want := response.InstalledPackageSummaries, tc.expectedPackages; !cmp.Equal(got, want, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name                 string
		existingObjects      []runtime.Object
		existingTypedObjects []runtime.Object
		expectedPackage      *corev1.InstalledPackageDetail
		statusCode           codes.Code
		request              *corev1.GetInstalledPackageDetailRequest
	}{
		{
			name: "it returns carvel installed package detail",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Identifier: "my-installation",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
				&kappctrlv1alpha1.App{
					TypeMeta: metav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
					},
					Status: kappctrlv1alpha1.AppStatus{
						Deploy: &kappctrlv1alpha1.AppStatusDeploy{
							Stdout: "deployStdout",
							Stderr: "deployStderr",
						},
						Fetch: &kappctrlv1alpha1.AppStatusFetch{
							Stdout: "fetchStdout",
							Stderr: "fetchStderr",
						},
						Inspect: &kappctrlv1alpha1.AppStatusInspect{
							Stdout: "inspectStdout",
							Stderr: "inspectStderr",
						},
					},
				},
			},
			existingTypedObjects: []runtime.Object{
				&k8scorev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			statusCode: codes.OK,
			expectedPackage: &corev1.InstalledPackageDetail{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
				Name: "my-installation",
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.2.3",
				},
				CurrentVersion: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				ValuesApplied: "\n# values.yaml\nfoo: bar\n",
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
					Interval:           30,
					Suspend:            false,
				},
				Status: &corev1.InstalledPackageStatus{
					Ready:      true,
					Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "Reconcile succeeded",
				},
				PostInstallationNotes: fmt.Sprintf(`## Installation output


### Deploy:
%s


### Fetch:
%s


### Inspect:
%s


## Errors


### Deploy:
%s


### Fetch:
%s


### Inspect:
%s

`, "deployStdout", "fetchStdout", "inspectStdout", "deployStderr", "fetchStderr", "inspectStderr"),
				LatestMatchingVersion: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				LatestVersion: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typfake.NewSimpleClientset(tc.existingTypedObjects...),
						dynfake.NewSimpleDynamicClientWithCustomListKinds(
							runtime.NewScheme(),
							map[schema.GroupVersionResource]string{
								{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
								{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							},
							unstructuredObjects...,
						), nil
				},
			}
			installedPackageDetail, err := s.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPackageDetail.InstalledPackageDetail, tc.expectedPackage; !cmp.Equal(got, want, ignoreUnexported) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
				}
				// TODO(agamez): check the actual object being updated in the k8s fake
			}
		})
	}
}

func TestCreateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name                   string
		request                *corev1.CreateInstalledPackageRequest
		existingObjects        []runtime.Object
		expectedStatusCode     codes.Code
		expectedResponse       *corev1.CreateInstalledPackageResponse
		expectedPackageInstall *packagingv1alpha1.PackageInstall
	}{
		{
			name: "create installed package",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.2.3",
				},
				Name: "my-installation",
				TargetContext: &corev1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &packagingv1alpha1.PackageInstall{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: packagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &packagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversions.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []packagingv1alpha1.PackageInstallValues{{
						SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &metav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: packagingv1alpha1.PackageInstallStatus{
					GenericStatus: kappctrlv1alpha1.GenericStatus{
						ObservedGeneration: 1,
						Conditions: []kappctrlv1alpha1.AppCondition{{
							Type:    kappctrlv1alpha1.ReconcileSucceeded,
							Status:  k8scorev1.ConditionTrue,
							Reason:  "baz",
							Message: "qux",
						}},
						FriendlyDescription: "foo",
						UsefulErrorMessage:  "foo",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
		{
			name: "create installed package (with values)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.2.3",
				},
				Name:   "my-installation",
				Values: "foo: bar",
				TargetContext: &corev1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &packagingv1alpha1.PackageInstall{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: packagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &packagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversions.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []packagingv1alpha1.PackageInstallValues{{
						SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &metav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: packagingv1alpha1.PackageInstallStatus{
					GenericStatus: kappctrlv1alpha1.GenericStatus{
						ObservedGeneration: 1,
						Conditions: []kappctrlv1alpha1.AppCondition{{
							Type:    kappctrlv1alpha1.ReconcileSucceeded,
							Status:  k8scorev1.ConditionTrue,
							Reason:  "baz",
							Message: "qux",
						}},
						FriendlyDescription: "foo",
						UsefulErrorMessage:  "foo",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
		{
			name: "create installed package (with reconciliationOptions)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.2.3",
				},
				ReconciliationOptions: &corev1.ReconciliationOptions{
					Interval:           99,
					Suspend:            true,
					ServiceAccountName: "my-sa",
				},
				Name: "my-installation",
				TargetContext: &corev1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &packagingv1alpha1.PackageInstall{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: packagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &packagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversions.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []packagingv1alpha1.PackageInstallValues{{
						SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &metav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: packagingv1alpha1.PackageInstallStatus{
					GenericStatus: kappctrlv1alpha1.GenericStatus{
						ObservedGeneration: 1,
						Conditions: []kappctrlv1alpha1.AppCondition{{
							Type:    kappctrlv1alpha1.ReconcileSucceeded,
							Status:  k8scorev1.ConditionTrue,
							Reason:  "baz",
							Message: "qux",
						}},
						FriendlyDescription: "foo",
						UsefulErrorMessage:  "foo",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
		{
			name: "create installed package (version constraint)",
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: ">1",
				},
				Name: "my-installation",
				TargetContext: &corev1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &packagingv1alpha1.PackageInstall{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: packagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &packagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversions.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []packagingv1alpha1.PackageInstallValues{{
						SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &metav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: packagingv1alpha1.PackageInstallStatus{
					GenericStatus: kappctrlv1alpha1.GenericStatus{
						ObservedGeneration: 1,
						Conditions: []kappctrlv1alpha1.AppCondition{{
							Type:    kappctrlv1alpha1.ReconcileSucceeded,
							Status:  k8scorev1.ConditionTrue,
							Reason:  "baz",
							Message: "qux",
						}},
						FriendlyDescription: "foo",
						UsefulErrorMessage:  "foo",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			createInstalledPackageResponse, err := s.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}
			if tc.expectedPackageInstall != nil {
				if got, want := createInstalledPackageResponse, tc.expectedResponse; !cmp.Equal(want, got, ignoreUnexported) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
				}
			}
		})
	}
}

func TestUpdateInstalledPackage(t *testing.T) {
	testCases := []struct {
		name                   string
		request                *corev1.UpdateInstalledPackageRequest
		existingObjects        []runtime.Object
		existingTypedObjects   []runtime.Object
		expectedStatusCode     codes.Code
		expectedResponse       *corev1.UpdateInstalledPackageResponse
		expectedPackageInstall *packagingv1alpha1.PackageInstall
	}{
		{
			name: "update installed package",
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
				PkgVersionReference: &corev1.VersionReference{
					Version: "1.2.3",
				},
				Values: "foo: bar",
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
					Interval:           30,
					Suspend:            false,
				},
			},
			existingObjects: []runtime.Object{
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						Categories:         []string{"logging", "daemon-set"},
						LongDescription:    "A great game for arcade gamers",
						ProviderName:       "Tetris inc.",
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "foo"}},
						SupportDescription: "Block support team",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"foo license"},
						ReleasedAt:                      metav1.Time{},
						CapactiyRequirementsDescription: "foo capactiyRequirementsDescription",
						ReleaseNotes:                    "foo releaseNotes",
						Template:                        datapackagingv1alpha1.AppTemplateSpec{},
						ValuesSchema:                    datapackagingv1alpha1.ValuesSchema{},
					},
				},
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
			},
			existingTypedObjects: []runtime.Object{
				&k8scorev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &packagingv1alpha1.PackageInstall{
				TypeMeta: metav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: packagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &packagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversions.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []packagingv1alpha1.PackageInstallValues{{
						SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &metav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: packagingv1alpha1.PackageInstallStatus{
					GenericStatus: kappctrlv1alpha1.GenericStatus{
						ObservedGeneration: 1,
						Conditions: []kappctrlv1alpha1.AppCondition{{
							Type:    kappctrlv1alpha1.ReconcileSucceeded,
							Status:  k8scorev1.ConditionTrue,
							Reason:  "baz",
							Message: "qux",
						}},
						FriendlyDescription: "foo",
						UsefulErrorMessage:  "foo",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typfake.NewSimpleClientset(tc.existingTypedObjects...), dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			updateInstalledPackageResponse, err := s.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}
			if tc.expectedPackageInstall != nil {
				if got, want := updateInstalledPackageResponse, tc.expectedResponse; !cmp.Equal(want, got, ignoreUnexported) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
				}
				// TODO(agamez): check the actual object being updated in the k8s fake
			}
		})
	}
}

func TestDeleteInstalledPackage(t *testing.T) {
	testCases := []struct {
		name                 string
		request              *corev1.DeleteInstalledPackageRequest
		existingObjects      []runtime.Object
		existingTypedObjects []runtime.Object
		expectedStatusCode   codes.Code
		expectedResponse     *corev1.DeleteInstalledPackageResponse
	}{
		{
			name: "deletes installed package",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-install",
				},
			},
			existingObjects: []runtime.Object{
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-install",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-secret",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
			},
			existingTypedObjects: []runtime.Object{
				&k8scorev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-secret",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse:   &corev1.DeleteInstalledPackageResponse{},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "noy-my-install",
				},
			},
			existingObjects: []runtime.Object{
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-install",
					},
					Spec: packagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &packagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversions.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []packagingv1alpha1.PackageInstallValues{{
							SecretRef: &packagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-secret",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &metav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: packagingv1alpha1.PackageInstallStatus{
						GenericStatus: kappctrlv1alpha1.GenericStatus{
							ObservedGeneration: 1,
							Conditions: []kappctrlv1alpha1.AppCondition{{
								Type:    kappctrlv1alpha1.ReconcileSucceeded,
								Status:  k8scorev1.ConditionTrue,
								Reason:  "baz",
								Message: "qux",
							}},
							FriendlyDescription: "foo",
							UsefulErrorMessage:  "foo",
						},
						Version:              "1.2.3",
						LastAttemptedVersion: "1.2.3",
					},
				},
			},
			existingTypedObjects: []runtime.Object{
				&k8scorev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-secret",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: codes.NotFound,
			expectedResponse:   &corev1.DeleteInstalledPackageResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []runtime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &unstructured.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typfake.NewSimpleClientset(tc.existingTypedObjects...), dynfake.NewSimpleDynamicClientWithCustomListKinds(
						runtime.NewScheme(),
						map[schema.GroupVersionResource]string{
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			deleteInstalledPackageResponse, err := s.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}
			if got, want := deleteInstalledPackageResponse, tc.expectedResponse; !cmp.Equal(want, got, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}
}
