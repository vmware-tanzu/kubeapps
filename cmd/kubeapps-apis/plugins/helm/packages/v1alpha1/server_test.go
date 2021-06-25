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
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"github.com/kubeapps/kubeapps/pkg/dbutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	dynfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	typfake "k8s.io/client-go/kubernetes/fake"
	log "k8s.io/klog/v2"
)

func setMockManager(t *testing.T) (sqlmock.Sqlmock, func(), utils.AssetManager) {
	var manager utils.AssetManager
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	manager = &utils.PostgresAssetManager{&dbutils.PostgresAssetManager{DB: db, KubeappsNamespace: "kubeappsNamespace"}}
	return mock, func() { db.Close() }, manager
}

func TestGetClient(t *testing.T) {
	kubeappsNamespace := "kubeapps"
	dbConfig := datastore.Config{URL: "localhost:5432", Database: "assetsvc", Username: "postgres", Password: "password"}
	manager, err := utils.NewPGManager(dbConfig, kubeappsNamespace)
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

func TestAvailablePackageSummaryFromChart(t *testing.T) {
	chartOK := &models.Chart{
		Name:        "foo",
		ID:          "foo",
		Category:    "cat1",
		Description: "best chart",
		Icon:        "foo.bar/icon.svg",
		Repo: &models.Repo{
			Name:      "bar",
			Namespace: "my-ns",
		},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0", AppVersion: "0.1.0"},
			{Version: "1.0.0", AppVersion: "whatever"},
			{Version: "whatever", AppVersion: "1.0.0"},
		},
	}

	availablePackageSummaryOK := &corev1.AvailablePackageSummary{
		DisplayName:      "foo",
		LatestVersion:    "1.0.0",
		IconUrl:          "foo.bar/icon.svg",
		ShortDescription: "best chart",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Namespace: "my-ns"},
			Identifier: "foo",
		},
	}

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
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{})
				if got, want := availablePackageSummary, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return typfake.NewSimpleClientset(), dynfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
			nil,
		), nil
	}
	chartOK := &models.Chart{
		Name:        "foo",
		ID:          "foo",
		Category:    "cat1",
		Description: "best chart",
		Icon:        "foo.bar/icon.svg",
		Repo: &models.Repo{
			Name:      "bar",
			Namespace: "my-ns",
		},
		ChartVersions: []models.ChartVersion{
			{Version: "1.0.0", AppVersion: "0.1.0"},
			{Version: "1.0.0", AppVersion: "whatever"},
			{Version: "whatever", AppVersion: "1.0.0"},
		},
	}
	availablePackageSummaryOK := &corev1.AvailablePackageSummary{
		DisplayName:      "foo",
		LatestVersion:    "1.0.0",
		IconUrl:          "foo.bar/icon.svg",
		ShortDescription: "best chart",
		AvailablePackageRef: &corev1.AvailablePackageReference{
			Context:    &corev1.Context{Namespace: "my-ns"},
			Identifier: "foo",
		},
	}
	testCases := []struct {
		name             string
		charts           []*models.Chart
		expectedPackages []*corev1.AvailablePackageSummary
		statusCode       codes.Code
	}{
		{
			name:             "it returns a set of availablePackageSummary from the database",
			charts:           []*models.Chart{chartOK},
			expectedPackages: []*corev1.AvailablePackageSummary{availablePackageSummaryOK},
			statusCode:       codes.OK,
		},
		{
			name:             "it returns an internal error status if response does not contain version",
			charts:           []*models.Chart{{Name: "foo"}},
			expectedPackages: []*corev1.AvailablePackageSummary{},
			statusCode:       codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock, cleanup, manager := setMockManager(t)
			defer cleanup()
			s := Server{
				clientGetter: clientGetter,
				manager:      manager,
			}

			rows := sqlmock.NewRows([]string{"info"})
			rowCount := sqlmock.NewRows([]string{"count"}).AddRow(len(tc.charts))

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			mock.ExpectQuery("SELECT info FROM").
				WillReturnRows(rows)

			mock.ExpectQuery("^SELECT count(.+) FROM").
				WillReturnRows(rowCount)

			availablePackageSummaries, err := s.GetAvailablePackageSummaries(context.Background(), &corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{})
				if got, want := availablePackageSummaries.AvailablePackagesSummaries, tc.expectedPackages; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}
