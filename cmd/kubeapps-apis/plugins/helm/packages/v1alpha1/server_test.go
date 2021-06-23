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

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kubeapps/common/datastore"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	log "k8s.io/klog/v2"
)

func TestGetClient(t *testing.T) {
	kubeappsNamespace := "kubeapps"
	dbConfig := datastore.Config{URL: "localhost:5432", Database: "assetsvc", Username: "postgres", Password: "password"}
	manager, err := newPGManager(dbConfig, kubeappsNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	getter := func(context.Context) (dynamic.Interface, error) {
		return fake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}

	testCases := []struct {
		name              string
		manager           assetManager
		clientGetter      func(context.Context) (dynamic.Interface, error)
		statusCodeClient  codes.Code
		statusCodeManager codes.Code
	}{
		{
			name:              "returns internal error status when no getter configured",
			manager:           manager,
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.OK,
		},
		{
			name:              "returns internal error status when no manager configured",
			manager:           nil,
			clientGetter:      getter,
			statusCodeClient:  codes.OK,
			statusCodeManager: codes.Internal,
		},
		{
			name:              "returns internal error status when no getter/manager configured",
			manager:           nil,
			clientGetter:      nil,
			statusCodeClient:  codes.Internal,
			statusCodeManager: codes.Internal,
		},
		{
			name:    "returns failed-precondition when configGetter itself errors",
			manager: manager,
			clientGetter: func(context.Context) (dynamic.Interface, error) {
				return nil, fmt.Errorf("Bang!")
			},
			statusCodeClient:  codes.FailedPrecondition,
			statusCodeManager: codes.OK,
		},
		{
			name:         "returns client without error when configured correctly",
			manager:      manager,
			clientGetter: getter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter, manager: tc.manager}

			client, errClient := s.GetClient(context.Background())

			if got, want := status.Code(errClient), tc.statusCodeClient; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			_, errManager := s.GetManager()

			if got, want := status.Code(errManager), tc.statusCodeManager; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			// If there is no error, the client should be a dynamic.Interface implementation.
			if tc.statusCodeClient == codes.OK {
				if _, ok := client.(dynamic.Interface); !ok {
					t.Errorf("got: %T, want: dynamic.Interface", client)
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
			name:       "returns AvailablePackageSummary if the chart is correct",
			in:         chartOK,
			expected:   availablePackageSummaryOK,
			statusCode: codes.OK,
		},
		{
			name:       "returns internal error if empty chart",
			in:         &models.Chart{},
			statusCode: codes.Internal,
		},
		{
			name:       "returns internal error if chart is invalid",
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
