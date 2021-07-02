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
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/common/datastore"
	"github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authorizationv1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/helm/pkg/proto/hapi/chart"
	log "k8s.io/klog/v2"
)

const globalPackagingNamespace = "kubeapps"

var chartOK = &models.Chart{
	Name:        "foo",
	ID:          "foo/bar",
	Category:    "cat1",
	Description: "best chart",
	Icon:        "foo.bar/icon.svg",
	Repo: &models.Repo{
		Name:      "bar",
		Namespace: "my-ns",
	},
	Maintainers: []chart.Maintainer{{Name: "me", Email: "me@me.me"}},
	ChartVersions: []models.ChartVersion{
		{Version: "3.0.0", AppVersion: "1.0.0", Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
		{Version: "2.0.0", AppVersion: "1.0.0", Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
		{Version: "1.0.0", AppVersion: "1.0.0", Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
	},
}

var availablePackageSummaryOK = &corev1.AvailablePackageSummary{
	DisplayName:      "foo",
	LatestPkgVersion: "3.0.0",
	IconUrl:          "foo.bar/icon.svg",
	ShortDescription: "best chart",
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Context:    &corev1.Context{Namespace: "my-ns"},
		Identifier: "foo/bar",
		Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
	},
}

var availablePackageDetailOK = &corev1.AvailablePackageDetail{
	Name:             "foo",
	DisplayName:      "foo",
	IconUrl:          "foo.bar/icon.svg",
	ShortDescription: "best chart",
	LongDescription:  "",
	PkgVersion:       "3.0.0",
	AppVersion:       "1.0.0",
	Readme:           "chart readme",
	DefaultValues:    "chart values",
	ValuesSchema:     "chart schema",
	Maintainers:      []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
	AvailablePackageRef: &corev1.AvailablePackageReference{
		Context:    &corev1.Context{Namespace: "my-ns"},
		Identifier: "foo/bar",
		Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
	},
}

func setMockManager(t *testing.T) (sqlmock.Sqlmock, func(), utils.AssetManager) {
	var manager utils.AssetManager
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	manager = &utils.PostgresAssetManager{&dbutils.PostgresAssetManager{DB: db, KubeappsNamespace: globalPackagingNamespace}}
	return mock, func() { db.Close() }, manager
}

func TestGetClient(t *testing.T) {
	dbConfig := datastore.Config{URL: "localhost:5432", Database: "assetsvc", Username: "postgres", Password: "password"}
	manager, err := utils.NewPGManager(dbConfig, globalPackagingNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}

	testCases := []struct {
		name              string
		manager           utils.AssetManager
		clientGetter      server.KubernetesClientGetter
		statusCodeClient  codes.Code
		statusCodeManager codes.Code
	}{
		{
			name:              "it returns internal error status when no clientGetter configured",
			manager:           manager,
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.OK,
		},
		{
			name:              "it returns internal error status when no manager configured",
			manager:           nil,
			clientGetter:      clientGetter,
			statusCodeClient:  codes.OK,
			statusCodeManager: codes.Internal,
		},
		{
			name:              "it returns internal error status when no clientGetter/manager configured",
			manager:           nil,
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.Internal,
		},
		{
			name:    "it returns failed-precondition when configGetter itself errors",
			manager: manager,
			clientGetter: func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCodeClient:  codes.FailedPrecondition,
			statusCodeManager: codes.OK,
		},
		{
			name:         "it returns client without error when configured correctly",
			manager:      manager,
			clientGetter: clientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter, manager: tc.manager}

			typedClient, dynamicClient, errClient := s.GetClients(context.Background())

			if got, want := status.Code(errClient), tc.statusCodeClient; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			_, errManager := s.GetManager()

			if got, want := status.Code(errManager), tc.statusCodeManager; got != want {
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

func TestIsValidChart(t *testing.T) {
	testCases := []struct {
		name     string
		in       *models.Chart
		expected bool
	}{
		{
			name:     "it returns true if the chart is correct",
			in:       chartOK,
			expected: true,
		},
		{
			name: "it returns true if the minimum chart is correct",
			in: &models.Chart{
				Name: "foo",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: true,
		},
		{
			name: "it returns false if the ChartVersions are missing",
			in: &models.Chart{
				Name: "foo",
			},
			expected: false,
		},
		{
			name: "it returns false if a ChartVersions.Version is missing",
			in: &models.Chart{
				Name: "foo",
				ChartVersions: []models.ChartVersion{
					{Version: "3.0.0"},
					{AppVersion: "3.0.0"},
				},
			},
			expected: false,
		},
		{
			name: "it returns true if the minimum (+maintainer) chart is correct",
			in: &models.Chart{
				Name: "foo",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []chart.Maintainer{{Name: "me"}},
			},
			expected: true,
		},
		{
			name: "it returns false if a Maintainer.Name is missing",
			in: &models.Chart{
				Name: "foo",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
				Maintainers: []chart.Maintainer{{Name: "me"}, {Email: "you"}},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := isValidChart(tc.in)
			if got, want := res, tc.expected; got != want {
				t.Fatalf("got: %+v, want: %+v, res: %+v (%+v)", got, want, res, err)
			}
		})
	}
}

func TestAvailablePackageSummaryFromChart(t *testing.T) {
	invalidChart := &models.Chart{Name: "foo"}

	testCases := []struct {
		name       string
		in         *models.Chart
		expected   *corev1.AvailablePackageSummary
		statusCode codes.Code
	}{
		{
			name:       "it returns AvailablePackageSummary if the chart is correct",
			in:         chartOK,
			expected:   availablePackageSummaryOK,
			statusCode: codes.OK,
		},
		{
			name:       "it returns internal error if empty chart",
			in:         &models.Chart{},
			statusCode: codes.Internal,
		},
		{
			name:       "it returns internal error if chart is invalid",
			in:         invalidChart,
			statusCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			availablePackageSummary, err := AvailablePackageSummaryFromChart(tc.in)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
				if got, want := availablePackageSummary, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	// Creating the dynamic client
	dynamicClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
		},
	)

	// Creating an authorized clientGetter
	authorizedClientSet := typfake.NewSimpleClientset()
	authorizedClientSet.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})
	authorizedClientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return authorizedClientSet, dynamicClient, nil
	}

	// Creating a unauthorized clientGetter
	unauthorizedClientSet := typfake.NewSimpleClientset()
	unauthorizedClientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return unauthorizedClientSet, dynamicClient, nil
	}

	// Creating the SQL mock manager
	mock, cleanup, manager := setMockManager(t)
	defer cleanup()

	testCases := []struct {
		name             string
		charts           []*models.Chart
		expectedPackages []*corev1.AvailablePackageSummary
		statusCode       codes.Code
		request          *corev1.GetAvailablePackageSummariesRequest
		server           *Server
	}{
		{
			name: "it returns a set of availablePackageSummary from the database (global ns)",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				FilterOptions: &corev1.FilterOptions{
					Query:        "",
					AppVersion:   "",
					PkgVersion:   "",
					Categories:   nil,
					Repositories: nil,
				},
			},
			charts:           []*models.Chart{chartOK},
			expectedPackages: []*corev1.AvailablePackageSummary{availablePackageSummaryOK},
			statusCode:       codes.OK,
		},
		{
			name: "it returns a set of availablePackageSummary from the database (specific ns)",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: "my-ns",
				},
				FilterOptions: &corev1.FilterOptions{
					Query:        "",
					AppVersion:   "",
					PkgVersion:   "",
					Categories:   nil,
					Repositories: nil,
				},
			},
			charts:           []*models.Chart{chartOK},
			expectedPackages: []*corev1.AvailablePackageSummary{availablePackageSummaryOK},
			statusCode:       codes.OK,
		},
		{
			name: "it returns a unimplemented status if no namespaces is provided",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "",
				},
			},
			charts:           []*models.Chart{{Name: "foo"}},
			expectedPackages: []*corev1.AvailablePackageSummary{},
			statusCode:       codes.Unimplemented,
		},
		{
			name: "it returns an internal error status if response does not contain version",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: globalPackagingNamespace,
				},
			},
			charts:           []*models.Chart{{Name: "foo"}},
			expectedPackages: []*corev1.AvailablePackageSummary{},
			statusCode:       codes.Internal,
		},
		{
			name: "it returns an unauthenticated status if the user doesn't have permissions",
			server: &Server{
				clientGetter:             unauthorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "my-ns",
				},
			},
			charts:           []*models.Chart{{Name: "foo"}},
			expectedPackages: []*corev1.AvailablePackageSummary{},
			statusCode:       codes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"info"})

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			if tc.statusCode != codes.Unauthenticated {
				// Checking if the WHERE condtion is properly applied
				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.request.Context.Namespace, tc.server.globalPackagingNamespace).
					WillReturnRows(rows)
			}
			availablePackageSummaries, err := tc.server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
				if got, want := availablePackageSummaries.AvailablePackagesSummaries, tc.expectedPackages; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestAvailablePackageDetailFromChart(t *testing.T) {
	testCases := []struct {
		name       string
		chart      *models.Chart
		expected   *corev1.AvailablePackageDetail
		statusCode codes.Code
	}{
		{
			name:       "it returns AvailablePackageDetail if the chart is correct",
			chart:      chartOK,
			expected:   availablePackageDetailOK,
			statusCode: codes.OK,
		},
		{
			name:       "it returns internal error if empty chart",
			chart:      &models.Chart{},
			statusCode: codes.Internal,
		},
		{
			name:       "it returns internal error if chart is invalid",
			chart:      &models.Chart{Name: "foo"},
			statusCode: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			availablePackageDetail, err := AvailablePackageDetailFromChart(tc.chart)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
				if got, want := availablePackageDetail, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	// Creating the dynamic client
	dynamicClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
		},
	)

	// Creating an authorized clientGetter
	authorizedClientSet := typfake.NewSimpleClientset()
	authorizedClientSet.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})
	authorizedClientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return authorizedClientSet, dynamicClient, nil
	}

	// Creating a unauthorized clientGetter
	unauthorizedClientSet := typfake.NewSimpleClientset()
	unauthorizedClientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return unauthorizedClientSet, dynamicClient, nil
	}

	// Creating the SQL mock manager
	mock, cleanup, manager := setMockManager(t)
	defer cleanup()

	testCases := []struct {
		name             string
		charts           []*models.Chart
		requestedVersion string
		expectedVersion  string
		expectedPackage  *corev1.AvailablePackageDetail
		statusCode       codes.Code
		request          *corev1.GetAvailablePackageDetailRequest
		server           *Server
	}{
		{
			name: "it returns an availablePackageDetail from the database (latest version)",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			expectedVersion: availablePackageDetailOK.PkgVersion,
			charts:          []*models.Chart{chartOK},
			expectedPackage: availablePackageDetailOK,
			statusCode:      codes.OK,
		},
		{
			name: "it returns an availablePackageDetail from the database (specific version)",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			requestedVersion: "1.0.0",
			expectedVersion:  "1.0.0",
			charts:           []*models.Chart{chartOK},
			expectedPackage:  availablePackageDetailOK,
			statusCode:       codes.OK,
		},
		{
			name: "it returns an internal error status if the chart is invalid",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			expectedVersion: availablePackageDetailOK.PkgVersion,
			charts:          []*models.Chart{{Name: "foo"}},
			expectedPackage: &corev1.AvailablePackageDetail{},
			statusCode:      codes.Internal,
		},
		{
			name: "it returns an internal error status if the requested chart version doesn't exist",
			server: &Server{
				clientGetter:             authorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			requestedVersion: "9.9.9",
			charts:           []*models.Chart{{Name: "foo"}},
			expectedPackage:  &corev1.AvailablePackageDetail{},
			statusCode:       codes.Internal,
		},
		{
			name: "it returns an unauthenticated status if the user doesn't have permissions",
			server: &Server{
				clientGetter:             unauthorizedClientGetter,
				manager:                  manager,
				globalPackagingNamespace: globalPackagingNamespace,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:          []*models.Chart{{Name: "foo"}},
			expectedPackage: &corev1.AvailablePackageDetail{},
			statusCode:      codes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			currentExpectedPackage := tc.expectedPackage
			currentExpectedPackage.PkgVersion = tc.expectedVersion

			rows := sqlmock.NewRows([]string{"info"})

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			if tc.statusCode != codes.Unauthenticated {
				// Checking if the WHERE condition is properly applied
				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.request.AvailablePackageRef.Context.Namespace, tc.request.AvailablePackageRef.Identifier).
					WillReturnRows(rows)
			}
			req := &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			}
			if tc.requestedVersion != "" {
				req.PkgVersion = tc.requestedVersion
			}
			availablePackageDetails, err := tc.server.GetAvailablePackageDetail(context.Background(), req)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
				if got, want := availablePackageDetails.AvailablePackageDetail, currentExpectedPackage; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}
