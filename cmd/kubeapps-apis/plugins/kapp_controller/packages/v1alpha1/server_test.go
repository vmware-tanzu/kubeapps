// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	ui "github.com/cppforlife/go-cli-ui/ui"
	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	kappapp "github.com/k14s/kapp/pkg/kapp/app"
	kappcmdapp "github.com/k14s/kapp/pkg/kapp/cmd/app"
	kappcmdcore "github.com/k14s/kapp/pkg/kapp/cmd/core"
	kappcmdtools "github.com/k14s/kapp/pkg/kapp/cmd/tools"
	kapplogger "github.com/k14s/kapp/pkg/kapp/logger"
	kappresources "github.com/k14s/kapp/pkg/kapp/resources"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	pkgkappv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappctrlpackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kappctrldatapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	vendirversionsv1alpha1 "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smetaunstructuredv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8discoveryclientfake "k8s.io/client-go/discovery/fake"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8dynamicclientfake "k8s.io/client-go/dynamic/fake"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
	log "k8s.io/klog/v2"
	k8syaml "sigs.k8s.io/yaml"
)

var ignoreUnexported = cmpopts.IgnoreUnexported(
	pkgsGRPCv1alpha1.AvailablePackageDetail{},
	pkgsGRPCv1alpha1.AvailablePackageReference{},
	pkgsGRPCv1alpha1.AvailablePackageSummary{},
	pkgsGRPCv1alpha1.Context{},
	pkgsGRPCv1alpha1.CreateInstalledPackageResponse{},
	pkgsGRPCv1alpha1.CreateInstalledPackageResponse{},
	pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{},
	pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{},
	pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{},
	pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{},
	pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{},
	pkgsGRPCv1alpha1.InstalledPackageDetail{},
	pkgsGRPCv1alpha1.InstalledPackageReference{},
	pkgsGRPCv1alpha1.InstalledPackageStatus{},
	pkgsGRPCv1alpha1.InstalledPackageSummary{},
	pkgsGRPCv1alpha1.Maintainer{},
	pkgsGRPCv1alpha1.PackageAppVersion{},
	pkgsGRPCv1alpha1.ReconciliationOptions{},
	pkgsGRPCv1alpha1.ResourceRef{},
	pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{},
	pkgsGRPCv1alpha1.VersionReference{},
	pluginsGRPCv1alpha1.Plugin{},
	pkgkappv1alpha1.PackageRepository{},
)

var defaultContext = &pkgsGRPCv1alpha1.Context{Cluster: "default", Namespace: "default"}

var datapackagingAPIVersion = fmt.Sprintf("%s/%s", kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version)
var packagingAPIVersion = fmt.Sprintf("%s/%s", kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, kappctrlpackagingv1alpha1.SchemeGroupVersion.Version)
var kappctrlAPIVersion = fmt.Sprintf("%s/%s", kappctrlv1alpha1.SchemeGroupVersion.Group, kappctrlv1alpha1.SchemeGroupVersion.Version)

func TestGetClient(t *testing.T) {
	testClientGetter := func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		return k8stypedclientfake.NewSimpleClientset(), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
			k8sruntime.NewScheme(),
			map[k8sschema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "fooList",
			},
		), nil
	}

	testCases := []struct {
		name              string
		clientGetter      clientgetter.ClientGetterFunc
		statusCodeClient  grpccodes.Code
		statusCodeManager grpccodes.Code
	}{
		{
			name:              "it returns internal error status when no clientGetter configured",
			clientGetter:      nil,
			statusCodeClient:  grpccodes.Internal,
			statusCodeManager: grpccodes.OK,
		},
		{
			name: "it returns failed-precondition when configGetter itself errors",
			clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCodeClient:  grpccodes.FailedPrecondition,
			statusCodeManager: grpccodes.OK,
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

			if got, want := grpcstatus.Code(errClient), tc.statusCodeClient; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a k8dynamicclient.Interface implementation.
			if tc.statusCodeClient == grpccodes.OK {
				if dynamicClient == nil {
					t.Errorf("got: nil, want: k8dynamicclient.Interface")
				}
				if typedClient == nil {
					t.Errorf("got: nil, want: k8stypedclient.Interface")
				}
			}
		})
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		existingObjects    []k8sruntime.Object
		expectedPackages   []*pkgsGRPCv1alpha1.AvailablePackageSummary
		expectedStatusCode grpccodes.Code
	}{
		{
			name: "it returns a not found error status if a package meta does not contain spec.displayName",
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
			},
			expectedStatusCode: grpccodes.NotFound,
		},
		{
			name: "it returns an not found error status if a package does not contain version",
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
			},
			expectedStatusCode: grpccodes.NotFound,
		},
		{
			name: "it returns carvel package summaries with basic info from the cluster",
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tombi.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Tombi!",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "An awesome game from the 90's",
						LongDescription:    "Tombi! is an open world platform-adventure game with RPG elements.",
						Categories:         []string{"platforms", "rpg"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tombi!",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tombi.foo.example.com.1.2.5",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tombi.foo.example.com",
						Version:                         "1.2.5",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1997, time.December, 25, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackages: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
				{
					AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					Categories:       []string{"logging", "daemon-set"},
				},
				{
					AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tombi.foo.example.com",
					},
					Name:        "tombi.foo.example.com",
					DisplayName: "Tombi!",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
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
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackages: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
				{
					AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
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
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.4",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.4",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackages: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
				{
					AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "tetris.foo.example.com",
					},
					Name:        "tetris.foo.example.com",
					DisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return nil, k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageSummaries(context.Background(), &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{Context: defaultContext})

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		existingObjects    []k8sruntime.Object
		request            *pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse
	}{
		{
			name:               "it returns invalid argument if called without a package reference",
			request:            nil,
			expectedStatusCode: grpccodes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without namespace",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{},
					Identifier: "package-one",
				},
			},
			expectedStatusCode: grpccodes.InvalidArgument,
		},
		{
			name: "it returns invalid argument if called without an identifier",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "kubeapps",
					},
				},
			},
			expectedStatusCode: grpccodes.InvalidArgument,
		},
		{
			name: "it returns the package version summary",
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.4",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.4",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
					},
					Identifier: "tetris.foo.example.com",
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*pkgsGRPCv1alpha1.PackageAppVersion{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return nil, k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}: pkgResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
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
		existingObjects []k8sruntime.Object
		expectedPackage *pkgsGRPCv1alpha1.AvailablePackageDetail
		statusCode      grpccodes.Code
		request         *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest
	}{
		{
			name: "it returns an availablePackageDetail of the latest version",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{
				Name:            "tetris.foo.example.com",
				DisplayName:     "Classic Tetris",
				LongDescription: "A few sentences but not really a readme",
				Version: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				Maintainers:      []*pkgsGRPCv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
				ShortDescription: "A great game for arcade gamers",
				Categories:       []string{"logging", "daemon-set"},
				Readme: `## Description

A few sentences but not really a readme

## Capactiy requirements

capacity description

## Release notes

release notes

Released at: June, 6 1984

## Support

Some support information

## Licenses

- my-license

`,
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginDetail,
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it combines long description and support description for readme field",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{
				Name:            "tetris.foo.example.com",
				DisplayName:     "Classic Tetris",
				LongDescription: "A few sentences but not really a readme",
				Version: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				Maintainers:      []*pkgsGRPCv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
				IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
				ShortDescription: "A great game for arcade gamers",
				Categories:       []string{"logging", "daemon-set"},
				Readme: `## Description

A few sentences but not really a readme

## Capactiy requirements

capacity description

## Release notes

release notes

Released at: June, 6 1984

## Support

Some support information

## Licenses

- my-license

`,
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
					Plugin:     &pluginDetail,
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name: "it returns an invalid arg error status if no context is provided",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Identifier: "foo/bar",
				},
			},
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name: "it returns not found error status if the requested package version doesn't exist",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersion: "1.2.4",
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			statusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return nil, k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}
			availablePackageDetail, err := s.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		request            *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest
		existingObjects    []k8sruntime.Object
		expectedPackages   []*pkgsGRPCv1alpha1.InstalledPackageSummary
		expectedStatusCode grpccodes.Code
	}{
		{
			name:    "it returns carvel empty installed package summary when no package install is present",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{Context: defaultContext},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{},
		},
		{
			name:    "it returns carvel installed package summary with complete metadata",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{Context: defaultContext},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name: "it returns carvel installed package from different namespaces if context.namespace=='' ",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Namespace: "",
					Cluster:   defaultContext.Cluster,
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "another-ns",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "another-ns",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "another-ns",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    &pkgsGRPCv1alpha1.Context{Namespace: "another-ns", Cluster: defaultContext.Cluster},
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name:    "it returns carvel installed package summary with a packageInstall without status",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{Context: defaultContext},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
				},
			},
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					IconUrl:          "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription: "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
						Version: "",
					},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "",
						AppVersion: "",
					},
					LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      false,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_PENDING,
						UserReason: "No status information yet",
					},
				},
			},
		},
		{
			name:    "it returns the latest semver version in the latest version field with the latest matching version",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{Context: defaultContext},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.2.0.0",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "2.0.0",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: ">1.0.0 <2.0.0",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "2.0.0",
						AppVersion: "2.0.0",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.7",
						AppVersion: "1.2.7",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
		{
			name:    "it returns the latest semver version in the latest version field with no latest matching version if constraint is not satisfied ",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{Context: defaultContext},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "9.9.9",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			expectedPackages: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
				{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context:    defaultContext,
						Plugin:     &pluginDetail,
						Identifier: "my-installation",
					},
					Name:           "my-installation",
					PkgDisplayName: "Classic Tetris",
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.7",
						AppVersion: "1.2.7",
					},
					IconUrl:             "data:image/svg+xml;base64,Tm90IHJlYWxseSBTVkcK",
					ShortDescription:    "A great game for arcade gamers",
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{Version: "1.2.3"},
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: "1.2.3",
						AppVersion: "1.2.3",
					},
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "Deployed",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return nil, k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			response, err := s.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		existingObjects      []k8sruntime.Object
		existingTypedObjects []k8sruntime.Object
		expectedPackage      *pkgsGRPCv1alpha1.InstalledPackageDetail
		statusCode           grpccodes.Code
		request              *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest
	}{
		{
			name: "it returns carvel installed package detail with the latest matching version",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Identifier: "my-installation",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.7",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.7",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.2.0.0",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "2.0.0",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: ">1.0.0 <2.0.0",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			statusCode: grpccodes.OK,
			expectedPackage: &pkgsGRPCv1alpha1.InstalledPackageDetail{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
				Name: "my-installation",
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				ValuesApplied: "\n# values.yaml\nfoo: bar\n",
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
					Interval:           30,
					Suspend:            false,
				},
				Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
					Ready:      true,
					Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "Deployed",
				},
				PostInstallationNotes: strings.ReplaceAll(`#### Deploy

<x60><x60><x60>
deployStdout
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStdout
<x60><x60><x60>

### Errors

#### Deploy

<x60><x60><x60>
deployStderr
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStderr
<x60><x60><x60>

`, "<x60>", "`"),
				LatestMatchingVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.7",
					AppVersion: "1.2.7",
				},
				LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "2.0.0",
					AppVersion: "2.0.0",
				},
			},
		},
		{
			name: "it returns carvel installed package detail with no latest matching version if constraint is not satisfied",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Identifier: "my-installation",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "9.9.9",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			statusCode: grpccodes.OK,
			expectedPackage: &pkgsGRPCv1alpha1.InstalledPackageDetail{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
				Name: "my-installation",
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
				ValuesApplied: "\n# values.yaml\nfoo: bar\n",
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
					Interval:           30,
					Suspend:            false,
				},
				Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
					Ready:      true,
					Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
					UserReason: "Deployed",
				},
				PostInstallationNotes: strings.ReplaceAll(`#### Deploy

<x60><x60><x60>
deployStdout
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStdout
<x60><x60><x60>

### Errors

#### Deploy

<x60><x60><x60>
deployStderr
<x60><x60><x60>

#### Fetch

<x60><x60><x60>
fetchStderr
<x60><x60><x60>

`, "<x60>", "`"),
				LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.2.3",
					AppVersion: "1.2.3",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return k8stypedclientfake.NewSimpleClientset(tc.existingTypedObjects...),
						k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
							k8sruntime.NewScheme(),
							map[k8sschema.GroupVersionResource]string{
								{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
								{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							},
							unstructuredObjects...,
						), nil
				},
			}
			installedPackageDetail, err := s.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
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
		request                *pkgsGRPCv1alpha1.CreateInstalledPackageRequest
		existingObjects        []k8sruntime.Object
		expectedStatusCode     grpccodes.Code
		expectedResponse       *pkgsGRPCv1alpha1.CreateInstalledPackageResponse
		expectedPackageInstall *kappctrlpackagingv1alpha1.PackageInstall
	}{
		{
			name: "create installed package",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				Name: "my-installation",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlv1alpha1.App{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &kappctrlpackagingv1alpha1.PackageInstall{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
						SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				Name: "my-installation",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
			},
			expectedStatusCode: grpccodes.Internal,
		},
		{
			name: "create installed package (with values)",
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				Name:   "my-installation",
				Values: "foo: bar",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlv1alpha1.App{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &kappctrlpackagingv1alpha1.PackageInstall{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
						SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					Interval:           99,
					Suspend:            true,
					ServiceAccountName: "my-sa",
				},
				Name: "my-installation",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlv1alpha1.App{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &kappctrlpackagingv1alpha1.PackageInstall{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
						SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			request: &pkgsGRPCv1alpha1.CreateInstalledPackageRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "tetris.foo.example.com",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1",
				},
				Name: "my-installation",
				TargetContext: &pkgsGRPCv1alpha1.Context{
					Namespace: "default",
					Cluster:   "default",
				},
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlv1alpha1.App{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       appResource,
						APIVersion: kappctrlAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlv1alpha1.AppSpec{
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
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
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.CreateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &kappctrlpackagingv1alpha1.PackageInstall{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
						SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return k8stypedclientfake.NewSimpleClientset(), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			createInstalledPackageResponse, err := s.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		request                *pkgsGRPCv1alpha1.UpdateInstalledPackageRequest
		existingObjects        []k8sruntime.Object
		existingTypedObjects   []k8sruntime.Object
		expectedStatusCode     grpccodes.Code
		expectedResponse       *pkgsGRPCv1alpha1.UpdateInstalledPackageResponse
		expectedPackageInstall *kappctrlpackagingv1alpha1.PackageInstall
	}{
		{
			name: "update installed package",
			request: &pkgsGRPCv1alpha1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "default",
						Cluster:   "default",
					},
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
				PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
					Version: "1.2.3",
				},
				Values: "foo: bar",
				ReconciliationOptions: &pkgsGRPCv1alpha1.ReconciliationOptions{
					ServiceAccountName: "default",
					Interval:           30,
					Suspend:            false,
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrldatapackagingv1alpha1.PackageMetadata{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgMetadataResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageMetadataSpec{
						DisplayName:        "Classic Tetris",
						IconSVGBase64:      "Tm90IHJlYWxseSBTVkcK",
						ShortDescription:   "A great game for arcade gamers",
						LongDescription:    "A few sentences but not really a readme",
						Categories:         []string{"logging", "daemon-set"},
						Maintainers:        []kappctrldatapackagingv1alpha1.Maintainer{{Name: "person1"}, {Name: "person2"}},
						SupportDescription: "Some support information",
						ProviderName:       "Tetris inc.",
					},
				},
				&kappctrldatapackagingv1alpha1.Package{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgResource,
						APIVersion: datapackagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "tetris.foo.example.com.1.2.3",
					},
					Spec: kappctrldatapackagingv1alpha1.PackageSpec{
						RefName:                         "tetris.foo.example.com",
						Version:                         "1.2.3",
						Licenses:                        []string{"my-license"},
						ReleaseNotes:                    "release notes",
						CapactiyRequirementsDescription: "capacity description",
						ReleasedAt:                      k8smetav1.Time{time.Date(1984, time.June, 6, 0, 0, 0, 0, time.UTC)},
					},
				},
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			expectedPackageInstall: &kappctrlpackagingv1alpha1.PackageInstall{
				TypeMeta: k8smetav1.TypeMeta{
					Kind:       pkgInstallResource,
					APIVersion: packagingAPIVersion,
				},
				ObjectMeta: k8smetav1.ObjectMeta{
					Namespace: "default",
					Name:      "my-installation",
				},
				Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
					ServiceAccountName: "default",
					PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
						RefName: "tetris.foo.example.com",
						VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
							Constraints: "1.2.3",
						},
					},
					Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
						SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
							Name: "my-installation-values",
						},
					},
					},
					Paused:     false,
					Canceled:   false,
					SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
					NoopDelete: false,
				},
				Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return k8stypedclientfake.NewSimpleClientset(tc.existingTypedObjects...), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			updateInstalledPackageResponse, err := s.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		request              *pkgsGRPCv1alpha1.DeleteInstalledPackageRequest
		existingObjects      []k8sruntime.Object
		existingTypedObjects []k8sruntime.Object
		expectedStatusCode   grpccodes.Code
		expectedResponse     *pkgsGRPCv1alpha1.DeleteInstalledPackageResponse
	}{
		{
			name: "deletes installed package",
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse:   &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{},
		},
		{
			name: "returns not found if installed package doesn't exist",
			request: &pkgsGRPCv1alpha1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "noy-my-installation",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.Secret{
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-values",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"values.yaml": []byte("foo: bar"),
					},
				},
			},
			expectedStatusCode: grpccodes.NotFound,
			expectedResponse:   &pkgsGRPCv1alpha1.DeleteInstalledPackageResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return k8stypedclientfake.NewSimpleClientset(tc.existingTypedObjects...), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
							{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
							{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			deleteInstalledPackageResponse, err := s.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		request              *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest
		existingObjects      []k8sruntime.Object
		existingTypedObjects []k8sruntime.Object
		expectedStatusCode   grpccodes.Code
		expectedResponse     *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse
	}{
		{
			name: "fetch the resources from an installed package",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context:    defaultContext,
					Plugin:     &pluginDetail,
					Identifier: "my-installation",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrlpackagingv1alpha1.PackageInstall{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgInstallResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation",
					},
					Spec: kappctrlpackagingv1alpha1.PackageInstallSpec{
						ServiceAccountName: "default",
						PackageRef: &kappctrlpackagingv1alpha1.PackageRef{
							RefName: "tetris.foo.example.com",
							VersionSelection: &vendirversionsv1alpha1.VersionSelectionSemver{
								Constraints: "1.2.3",
							},
						},
						Values: []kappctrlpackagingv1alpha1.PackageInstallValues{{
							SecretRef: &kappctrlpackagingv1alpha1.PackageInstallValuesSecretRef{
								Name: "my-installation-values",
							},
						},
						},
						Paused:     false,
						Canceled:   false,
						SyncPeriod: &k8smetav1.Duration{(time.Second * 30)},
						NoopDelete: false,
					},
					Status: kappctrlpackagingv1alpha1.PackageInstallStatus{
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
					TypeMeta: k8smetav1.TypeMeta{
						APIVersion: "v1",
						Kind:       "Pod",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
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
			existingTypedObjects: []k8sruntime.Object{
				&k8scorev1.ConfigMap{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Namespace: "default",
						Name:      "my-installation-ctrl",
					},
					Data: map[string]string{
						"spec": "{\"labelKey\":\"kapp.k14s.io/app\",\"labelValue\":\"my-id\"}",
					},
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
				ResourceRefs: []*pkgsGRPCv1alpha1.ResourceRef{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}
			// If more resources types are added, this will need to be updated accordingly
			apiResources := []*k8smetav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []k8smetav1.APIResource{
						{Name: "pods", Namespaced: true, Kind: "Pod", Verbs: []string{"list", "get"}},
					},
				},
			}

			typedClient := k8stypedclientfake.NewSimpleClientset(tc.existingTypedObjects...)

			// We cast the dynamic client to a fake client, so we can set the response
			fakeDiscovery, _ := typedClient.Discovery().(*k8discoveryclientfake.FakeDiscovery)
			fakeDiscovery.Fake.Resources = apiResources

			dynClient := k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
				k8sruntime.NewScheme(),
				map[k8sschema.GroupVersionResource]string{
					{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgsResource}:         pkgResource + "List",
					{Group: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrldatapackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgMetadatasResource}: pkgMetadataResource + "List",
					{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgInstallsResource}:          pkgInstallResource + "List",
					// If more resources types are added, this will need to be updated accordingly
					{Group: "", Version: "v1", Resource: "pods"}: "Pod" + "List",
				},
				unstructuredObjects...,
			)

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return typedClient, dynClient, nil
				},
				kappClientsGetter: func(ctx context.Context, cluster, namespace string) (kappapp.Apps, kappresources.IdentifiedResources, *kappcmdapp.FailingAPIServicesPolicy, kappresources.ResourceFilter, error) {
					// Create a fake Kapp DepsFactory and configure there the fake k8s clients the hereinbefore created
					depsFactory := NewFakeDepsFactoryImpl()
					depsFactory.SetCoreClient(typedClient)
					depsFactory.SetDynamicClient(dynClient)
					// The rest of the logic remain unchanged as in the real server.go file (DRY it up?)
					resourceFilterFlags := kappcmdtools.ResourceFilterFlags{}
					resourceFilter, err := resourceFilterFlags.ResourceFilter()
					if err != nil {
						return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					resourceTypesFlags := kappcmdapp.ResourceTypesFlags{
						IgnoreFailingAPIServices:         true,
						ScopeToFallbackAllowedNamespaces: true,
					}
					failingAPIServicesPolicy := resourceTypesFlags.FailingAPIServicePolicy()
					supportingNsObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: namespace}, resourceTypesFlags, kapplogger.NewNoopLogger())
					if err != nil {
						return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					supportingObjs, err := kappcmdapp.FactoryClients(depsFactory, kappcmdcore.NamespaceFlags{Name: ""}, resourceTypesFlags, kapplogger.NewNoopLogger())
					if err != nil {
						return kappapp.Apps{}, kappresources.IdentifiedResources{}, nil, kappresources.ResourceFilter{}, grpcstatus.Errorf(grpccodes.FailedPrecondition, "unable to get config due to: %v", err)
					}
					return supportingNsObjs.Apps, supportingObjs.IdentifiedResources, failingAPIServicesPolicy, resourceFilter, nil
				},
			}

			getInstalledPackageResourceRefsResponse, err := s.GetInstalledPackageResourceRefs(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}
			// If we were expecting an error, continue to the next test.
			if tc.expectedStatusCode != grpccodes.OK {
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
		request            *pkgkappv1alpha1.GetPackageRepositoriesRequest
		existingObjects    []k8sruntime.Object
		expectedResponse   []*pkgkappv1alpha1.PackageRepository
		expectedStatusCode grpccodes.Code
	}{
		{
			name: "returns expected repositories",
			request: &pkgkappv1alpha1.GetPackageRepositoriesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "default",
					Namespace: "default",
				},
			},
			existingObjects: []k8sruntime.Object{
				&kappctrlpackagingv1alpha1.PackageRepository{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgRepositoryResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "repo-1",
						Namespace: "default",
					},
					Spec: kappctrlpackagingv1alpha1.PackageRepositorySpec{
						Fetch: &kappctrlpackagingv1alpha1.PackageRepositoryFetch{
							ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
								Image: "projects.registry.example.com/repo-1/main@sha256:abcd",
							},
						},
					},
					Status: kappctrlpackagingv1alpha1.PackageRepositoryStatus{},
				},
				&kappctrlpackagingv1alpha1.PackageRepository{
					TypeMeta: k8smetav1.TypeMeta{
						Kind:       pkgRepositoryResource,
						APIVersion: packagingAPIVersion,
					},
					ObjectMeta: k8smetav1.ObjectMeta{
						Name:      "repo-2",
						Namespace: "default",
					},
					Spec: kappctrlpackagingv1alpha1.PackageRepositorySpec{
						Fetch: &kappctrlpackagingv1alpha1.PackageRepositoryFetch{
							ImgpkgBundle: &kappctrlv1alpha1.AppFetchImgpkgBundle{
								Image: "projects.registry.example.com/repo-2/main@sha256:abcd",
							},
						},
					},
					Status: kappctrlpackagingv1alpha1.PackageRepositoryStatus{},
				},
			},
			expectedResponse: []*pkgkappv1alpha1.PackageRepository{
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
			var unstructuredObjects []k8sruntime.Object
			for _, obj := range tc.existingObjects {
				unstructuredContent, _ := k8sruntime.DefaultUnstructuredConverter.ToUnstructured(obj)
				unstructuredObjects = append(unstructuredObjects, &k8smetaunstructuredv1.Unstructured{Object: unstructuredContent})
			}

			s := Server{
				clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
					return nil, k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
						k8sruntime.NewScheme(),
						map[k8sschema.GroupVersionResource]string{
							{Group: kappctrlpackagingv1alpha1.SchemeGroupVersion.Group, Version: kappctrlpackagingv1alpha1.SchemeGroupVersion.Version, Resource: pkgRepositoriesResource}: pkgRepositoryResource + "List",
						},
						unstructuredObjects...,
					), nil
				},
			}

			getPackageRepositoriesResponse, err := s.GetPackageRepositories(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %d, want: %d, err: %+v", got, want, err)
			}

			// Only check the response for OK grpcstatus.
			if tc.expectedStatusCode == grpccodes.OK {
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

func TestParsePluginConfig(t *testing.T) {
	testCases := []struct {
		name                         string
		pluginYAMLConf               []byte
		expectedDefaultUpgradePolicy upgradePolicy
		expectedErrorStr             string
	}{
		{
			name:                         "non existing plugin-config file",
			pluginYAMLConf:               nil,
			expectedDefaultUpgradePolicy: none,
			expectedErrorStr:             "no such file or directory",
		},
		{
			name: "defaultUpgradePolicy not set",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      `),
			expectedDefaultUpgradePolicy: none,
			expectedErrorStr:             "",
		},
		{
			name: "defaultUpgradePolicy: major",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: major
        `),
			expectedDefaultUpgradePolicy: major,
			expectedErrorStr:             "",
		},
		{
			name: "defaultUpgradePolicy: minor",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: minor
        `),
			expectedDefaultUpgradePolicy: minor,
			expectedErrorStr:             "",
		},
		{
			name: "defaultUpgradePolicy: patch",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: patch
        `),
			expectedDefaultUpgradePolicy: patch,
			expectedErrorStr:             "",
		},
		{
			name: "defaultUpgradePolicy: none",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: none
        `),
			expectedDefaultUpgradePolicy: none,
			expectedErrorStr:             "",
		},
		{
			name: "invalid defaultUpgradePolicy",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: foo
      `),
			expectedDefaultUpgradePolicy: none,
			expectedErrorStr:             "json: cannot unmarshal",
		},
		{
			name: "invalid defaultUpgradePolicy",
			pluginYAMLConf: []byte(`
kappController:
  packages:
    v1alpha1:
      defaultUpgradePolicy: 10.09
      `),
			expectedDefaultUpgradePolicy: none,
			expectedErrorStr:             "json: cannot unmarshal",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := ""
			if tc.pluginYAMLConf != nil {
				pluginJSONConf, err := k8syaml.YAMLToJSON(tc.pluginYAMLConf)
				if err != nil {
					log.Fatalf("%s", err)
				}
				f, err := os.CreateTemp(".", "plugin_json_conf")
				if err != nil {
					log.Fatalf("%s", err)
				}
				defer os.Remove(f.Name()) // clean up
				if _, err := f.Write(pluginJSONConf); err != nil {
					log.Fatalf("%s", err)
				}
				if err := f.Close(); err != nil {
					log.Fatalf("%s", err)
				}
				filename = f.Name()
			}
			defaultUpgradePolicy, goterr := parsePluginConfig(filename)
			if goterr != nil && !strings.Contains(goterr.Error(), tc.expectedErrorStr) {
				t.Errorf("err got %q, want to find %q", goterr.Error(), tc.expectedErrorStr)
			}
			if got, want := defaultUpgradePolicy, tc.expectedDefaultUpgradePolicy; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

// Implementing a FakeDepsFactoryImpl for injecting the typed and dynamic k8s clients
type FakeDepsFactoryImpl struct {
	kappcmdcore.DepsFactoryImpl
	coreClient    k8stypedclient.Interface
	dynamicClient k8dynamicclient.Interface

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

func (f *FakeDepsFactoryImpl) SetCoreClient(coreClient k8stypedclient.Interface) {
	f.coreClient = coreClient
}

func (f *FakeDepsFactoryImpl) SetDynamicClient(dynamicClient k8dynamicclient.Interface) {
	f.dynamicClient = dynamicClient
}

func (f *FakeDepsFactoryImpl) CoreClient() (k8stypedclient.Interface, error) {
	return f.coreClient, nil
}

func (f *FakeDepsFactoryImpl) DynamicClient(opts kappcmdcore.DynamicClientOpts) (k8dynamicclient.Interface, error) {
	return f.dynamicClient, nil
}
