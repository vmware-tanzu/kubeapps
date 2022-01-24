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
	"sync"
	"testing"
	"time"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctlapp "github.com/k14s/kapp/pkg/kapp/app"
	kappcmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	kappcmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	"github.com/k14s/kapp/pkg/kapp/logger"
	ctlres "github.com/k14s/kapp/pkg/kapp/resources"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginv1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
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
	disfake "k8s.io/client-go/discovery/fake"
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
	corev1.CreateInstalledPackageResponse{},
	corev1.CreateInstalledPackageResponse{},
	corev1.DeleteInstalledPackageResponse{},
	corev1.GetAvailablePackageVersionsResponse{},
	corev1.GetAvailablePackageVersionsResponse{},
	corev1.GetInstalledPackageResourceRefsResponse{},
	corev1.GetInstalledPackageResourceRefsResponse{},
	corev1.InstalledPackageDetail{},
	corev1.InstalledPackageReference{},
	corev1.InstalledPackageStatus{},
	corev1.InstalledPackageSummary{},
	corev1.Maintainer{},
	corev1.PackageAppVersion{},
	corev1.ReconciliationOptions{},
	corev1.ResourceRef{},
	corev1.UpdateInstalledPackageResponse{},
	corev1.VersionReference{},
	pluginv1.Plugin{},
	v1alpha1.PackageRepository{},
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
		clientGetter      clientgetter.ClientGetterFunc
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
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						DisplayName:        "Tombi!",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "An awesome game from the 90's",
						LongDescription:    "Tombi! is an open world platform-adventure game with RPG elements.",
						Categories:         []string{"platforms", "rpg"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tombi!",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tombi.foo.example.com",
						Version:                         "1.2.5",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					Categories:       []string{"logging", "daemon-set"},
				},
				{
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tombi.foo.example.com",
					},
					Name:        "tombi.foo.example.com",
					DisplayName: "Tombi!",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.5",
						AppVersion: "1.2.5",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "An awesome game from the 90's",
					Categories:       []string{"platforms", "rpg"},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
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
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.4",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.7",
						AppVersion: "1.2.7",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					Categories:       []string{"logging", "daemon-set"},
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.4",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						AppVersion: "1.2.7",
					},
					{
						PkgVersion: "1.2.4",
						AppVersion: "1.2.4",
					},
					{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
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

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
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
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
			},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:            "tetris.foo.example.com",
				DisplayName:     "Classic Tetris",
				LongDescription: "A few sentences but not really a readme",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				Maintainers:      []*corev1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
				ShortDescription: "A great game for arcade gamers",
				Categories:       []string{"logging", "daemon-set"},
				Readme: fmt.Sprintf(`## Details


### Description:
%s


### Capactiy requirements:
%s


### Release Notes:
%s


### Support:
%s


### Licenses:
%s


### ReleasedAt:
%s


`,
					"A few sentences but not really a readme",
					"capacity description",
					"release notes",
					"Some support information",
					[]string{"my-license"},
					&metav1.Time{},
				),
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginDetail,
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
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
			},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:            "tetris.foo.example.com",
				DisplayName:     "Classic Tetris",
				LongDescription: "A few sentences but not really a readme",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				Maintainers:      []*corev1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
				ShortDescription: "A great game for arcade gamers",
				Categories:       []string{"logging", "daemon-set"},
				Readme: fmt.Sprintf(`## Details


### Description:
%s


### Capactiy requirements:
%s


### Release Notes:
%s


### Support:
%s


### Licenses:
%s


### ReleasedAt:
%s


`,
					"A few sentences but not really a readme",
					"capacity description",
					"release notes",
					"Some support information",
					[]string{"my-license"},
					&metav1.Time{},
				),
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginDetail,
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
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
		request            *corev1.GetInstalledPackageSummariesRequest
		existingObjects    []runtime.Object
		expectedPackages   []*corev1.InstalledPackageSummary
		expectedStatusCode codes.Code
	}{
		{
			name:    "it returns carvel empty installed package summary when no package install is present",
			request: &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
			},
			expectedPackages: []*corev1.InstalledPackageSummary{},
		},
		{
			name:    "it returns carvel installed package summary with complete metadata",
			request: &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
							UsefulErrorMessage:  "Deployed",
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
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name: "it returns carvel installed package from different namespaces if context.namespace=='' ",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "",
					Cluster:   defaultContext.Cluster,
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&datapackagingv1alpha1.PackageMetadata{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "another-ns",
						Name:      "tetris.foo.example.com",
					},
					Spec: datapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "another-ns",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
				&packagingv1alpha1.PackageInstall{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "another-ns",
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
						Context:    &corev1.Context{Namespace: "another-ns", Cluster: defaultContext.Cluster},
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
				{
					InstalledPackageRef: &corev1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name:    "it returns carvel installed package summary with a packageInstall without status",
			request: &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{
						Version: "",
					},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "",
						AppVersion: "",
					},
					LatestMatchingVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      false,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_PENDING,
						UserReason: "No status information yet",
					},
				},
			},
		},
		{
			name:    "it returns the latest semver version in the latest version field with the latest matching version",
			request: &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.2.0.0",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "2.0.0",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
								Constraints: ">1.0.0 <2.0.0",
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
							UsefulErrorMessage:  "Deployed",
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
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "2.0.0",
						AppVersion: "2.0.0",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.7",
						AppVersion: "1.2.7",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name:    "it returns the latest semver version in the latest version field with no latest matching version if constraint is not satisfied ",
			request: &corev1.GetInstalledPackageSummariesRequest{Context: defaultContext},
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
								Constraints: "9.9.9",
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
							UsefulErrorMessage:  "Deployed",
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
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.7",
						AppVersion: "1.2.7",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &corev1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
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

			response, err := s.GetInstalledPackageSummaries(context.Background(), tc.request)

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
			name: "it returns carvel installed package detail with the latest matching version",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&datapackagingv1alpha1.Package{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.2.0.0",
					},
					Spec: datapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "2.0.0",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
								Constraints: ">1.0.0 <2.0.0",
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
							UsefulErrorMessage:  "Deployed",
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
					UserReason: "Deployed",
				},
				PostInstallationNotes: fmt.Sprintf(`## Installation output


### Deploy:
%s


### Fetch:
%s


## Errors


### Deploy:
%s


### Fetch:
%s


`, "deployStdout", "fetchStdout", "deployStderr", "fetchStderr"),
				LatestMatchingVersion: &corev1.PackageAppVersion{
					PkgVersion: "1.2.7",
					AppVersion: "1.2.7",
				},
				LatestVersion: &corev1.PackageAppVersion{
					PkgVersion: "2.0.0",
					AppVersion: "2.0.0",
				},
			},
		},
		{
			name: "it returns carvel installed package detail with no latest matching version if constraint is not satisfied",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
								Constraints: "9.9.9",
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
							UsefulErrorMessage:  "Deployed",
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
					UserReason: "Deployed",
				},
				PostInstallationNotes: fmt.Sprintf(`## Installation output


### Deploy:
%s


### Fetch:
%s


## Errors


### Deploy:
%s


### Fetch:
%s


`, "deployStdout", "fetchStdout", "deployStderr", "fetchStderr"),
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
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						UsefulErrorMessage:  "Deployed",
					},
					Version:              "1.2.3",
					LastAttemptedVersion: "1.2.3",
				},
			},
		},
		{
			name: "create installed package with error (kapp App not being created)",
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
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
			},
			expectedStatusCode: codes.Internal,
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
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						UsefulErrorMessage:  "Deployed",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						UsefulErrorMessage:  "Deployed",
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
				ReconciliationOptions: &corev1.ReconciliationOptions{
					ServiceAccountName: "default",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
						UsefulErrorMessage:  "Deployed",
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
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []datapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
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
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
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
							UsefulErrorMessage:  "Deployed",
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
						UsefulErrorMessage:  "Deployed",
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
					Identifier: "my-installation",
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
							UsefulErrorMessage:  "Deployed",
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
			expectedResponse:   &corev1.DeleteInstalledPackageResponse{},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "noy-my-installation",
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
							UsefulErrorMessage:  "Deployed",
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

func TestGetInstalledPackageResourceRefs(t *testing.T) {
	testCases := []struct {
		name                 string
		request              *corev1.GetInstalledPackageResourceRefsRequest
		existingObjects      []runtime.Object
		existingTypedObjects []runtime.Object
		expectedStatusCode   codes.Code
		expectedResponse     *corev1.GetInstalledPackageResourceRefsResponse
	}{
		{
			name: "fetch the resources from an installed package",
			request: &corev1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
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
				// Although it's a typical k8s object, it is retrieved with the dynamic client
				&k8scorev1.Pod{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-pod",
						Labels:    map[string]string{"kapp.k14s.io/app": "my-id"},
					},
					Spec: k8scorev1.PodSpec{
						Containers: []k8scorev1.Container{{
							Name: "my-installation-container",
						}},
					},
				},
			},
			existingTypedObjects: []runtime.Object{
				&k8scorev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-ctrl",
					},
					Data: map[string]string{
						"spec": "{\"labelKey\":\"kapp.k14s.io/app\",\"labelValue\":\"my-id\"}",
					},
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageResourceRefsResponse{
				ResourceRefs: []*corev1.ResourceRef{
					{
						ApiVersion: "v1",
						Kind:       "Pod",
						Name:       "my-installation-pod",
						Namespace:  "default",
					},
				},
				Context: defaultContext,
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
			// If more resources types are added, this will need to be updated accordingly
			apiResources := []*metav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []metav1.APIResource{
						{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get"}},
					},
				},
			}

			typedClient := typfake.NewSimpleClientset(tc.existingTypedObjects...)

			// We cast the dynamic client to a fake client, so we can set the response
			fakeDiscovery, _ := typedClient.Discovery().(*disfake.FakeDiscovery)
			fakeDiscovery.Fake.Resources = apiResources

			dynClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
				runtime.NewScheme(),
				map[schema.GroupVersionResource]string{
					{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
					{Group: datapackagingv1alpha1.SchemeGroupVersion.Group, Version: datapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
					{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
					// If more resources types are added, this will need to be updated accordingly
					{Group: "", Version: "v1", Resource: "pods"}: "Pod" + "List",
				},
				unstructuredObjects...,
			)

			s := Server{
				clientGetter: func(context.Context, string) (kubernetes.Interface, dynamic.Interface, error) {
					return typedClient, dynClient, nil
				},
				kappClientsGetter: func(ctx context.Context, cluster, namespace string) (ctlapp.Apps, ctlres.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, ctlres.ResourceFilter, error) {
					// Create a fake Kapp DepsFactory and configure there the fake k8s clients the hereinbefore created
					depsFactory := NewFakeDepsFactoryImpl()
					depsFactory.SetCoreClient(typedClient)
					depsFactory.SetDynamicClient(dynClient)
					// The rest of the logic remain unchanged as in the real server.go file (DRY it up?)
					resourceFilterFlags := kappcmdtools.ResourceFilterFlags{}
					resourceFilter, err := resourceFilterFlags.ResourceFilter()
					if err != nil {
						return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					resourceTypesFlags := kappcmdapp.ResourceTypesFlags{
						IgnoreFailingAPIServices:         true,
						ScopeToFallbackAllowedNamespaces: true,
					}
					failingAPIServicesPolicy := resourceTypesFlags.FailingAPIServicePolicy()
					supportingNsObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: namespace}, resourceTypesFlags, logger.NewNoopLogger())
					if err != nil {
						return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					supportingObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: ""}, resourceTypesFlags, logger.NewNoopLogger())
					if err != nil {
						return ctlapp.Apps{}, ctlres.IdentifiedResources{}, nil, ctlres.ResourceFilter{}, status.Errorf(codes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					return supportingNsObjs.Apps, supportingObjs.IdentifiedResources, failingAPIServicesPolicy, resourceFilter, nil
				},
			}

			getInstalledPackageResourceRefsResponse, err := s.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != codes.OK {
				return
			}
			if got, want := getInstalledPackageResourceRefsResponse, tc.expectedResponse; !cmp.Equal(want, got, ignoreUnexported) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
			}
		})
	}
}

func TestGetPackageRepositories(t *testing.T) {
	testCases := []struct {
		name               string
		request            *v1alpha1.GetPackageRepositoriesRequest
		existingObjects    []runtime.Object
		expectedResponse   []*v1alpha1.PackageRepository
		expectedStatusCode codes.Code
	}{
		{
			name: "returns expected repositories",
			request: &v1alpha1.GetPackageRepositoriesRequest{
				Context: &corev1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []runtime.Object{
				&packagingv1alpha1.PackageRepository{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgRepositoryResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "repo-1",
						Namespace: "default",
					},
					Spec: packagingv1alpha1.PackageRepositorySpec{
						Fetch: &packagingv1alpha1.PackageRepositoryFetch{
							ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
								Image: "projects.registry.example.com/repo-1/main@sha256:abcd",
							},
						},
					},
					Status: packagingv1alpha1.PackageRepositoryStatus{},
				},
				&packagingv1alpha1.PackageRepository{
					TypeMeta: metav1.TypeMeta{
						Kind:       pkgRepositoryResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "repo-2",
						Namespace: "default",
					},
					Spec: packagingv1alpha1.PackageRepositorySpec{
						Fetch: &packagingv1alpha1.PackageRepositoryFetch{
							ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
								Image: "projects.registry.example.com/repo-2/main@sha256:abcd",
							},
						},
					},
					Status: packagingv1alpha1.PackageRepositoryStatus{},
				},
			},
			expectedResponse: []*v1alpha1.PackageRepository{
				{
					Name:      "repo-1",
					Url:       "projects.registry.example.com/repo-1/main@sha256:abcd",
					Namespace: "default",
					Plugin:    &pluginDetail,
				},
				{
					Name:      "repo-2",
					Url:       "projects.registry.example.com/repo-2/main@sha256:abcd",
					Namespace: "default",
					Plugin:    &pluginDetail,
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
							{Group: packagingv1alpha1.SchemeGroupVersion.Group, Version: packagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgRepositoriesResource}: pkgRepositoryResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			getPackageRepositoriesResponse, err := s.GetPackageRepositories(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			// Only check the response for OK status.
			if tc.expectedStatusCode == codes.OK {
				if getPackageRepositoriesResponse == nil {
					t.Fatalf("got: nil, want: response")
				} else {
					if got, want := getPackageRepositoriesResponse.Repositories, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexported) {
						t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexported))
					}
				}
			}
		})
	}
}

// Implementing a FakeDepsFactoryImpl for injecting the typed and dynamic k8s clients
type FakeDepsFactoryImpl struct {
	kappcmdcore.DepsFactoryImpl
	coreClient    kubernetes.Interface
	dynamicClient dynamic.Interface

	configFactory   kappcmdcore.ConfigFactory
	ui              ui.UI
	printTargetOnce *sync.Once
	Warnings        bool
}

var _ kappcmdcore.DepsFactory = &FakeDepsFactoryImpl{}

func NewFakeDepsFactoryImpl() *FakeDepsFactoryImpl {
	return &FakeDepsFactoryImpl{
		configFactory:   &ConfigurableConfigFactoryImpl{},
		ui:              ui.NewNoopUI(),
		printTargetOnce: &sync.Once{},
	}
}

func (f *FakeDepsFactoryImpl) SetCoreClient(coreClient kubernetes.Interface) {
	f.coreClient = coreClient
}

func (f *FakeDepsFactoryImpl) SetDynamicClient(dynamicClient dynamic.Interface) {
	f.dynamicClient = dynamicClient
}

func (f *FakeDepsFactoryImpl) CoreClient() (kubernetes.Interface, error) {
	return f.coreClient, nil
}

func (f *FakeDepsFactoryImpl) DynamicClient(opts kappcmdcore.DynamicClientOpts) (dynamic.Interface, error) {
	return f.dynamicClient, nil
}
