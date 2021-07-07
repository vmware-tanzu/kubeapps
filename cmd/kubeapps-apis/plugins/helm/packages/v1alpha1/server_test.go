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

const (
	globalPackagingNamespace = "kubeapps"
	DefaultAppVersion        = "1.2.6"
	DefaultChartDescription  = "default chart description"
	DefaultChartIconURL      = "https://example.com/chart.svg"
)

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
			name: "it returns true if the chart name, ID, repo and versions are specified",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: true,
		},
		{
			name: "it returns false if the chart name is missing",
			in: &models.Chart{
				ID: "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart ID is missing",
			in: &models.Chart{
				Name: "foo",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the chart repo is missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: false,
		},
		{
			name: "it returns false if the ChartVersions are missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
			},
			expected: false,
		},
		{
			name: "it returns false if a ChartVersions.Version is missing",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				ChartVersions: []models.ChartVersion{
					{Version: "3.0.0"},
					{AppVersion: DefaultAppVersion},
				},
			},
			expected: false,
		},
		{
			name: "it returns true if the minimum (+maintainer) chart is correct",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
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
				ID:   "foo/bar",
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
			name: "it returns a complete AvailablePackageSummary for a complete chart",
			in: &models.Chart{
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
					{Version: "3.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "2.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
					{Version: "1.0.0", AppVersion: DefaultAppVersion, Readme: "chart readme", Values: "chart values", Schema: "chart schema"},
				},
			},
			expected: &corev1.AvailablePackageSummary{
				DisplayName:      "foo",
				LatestPkgVersion: "3.0.0",
				IconUrl:          "foo.bar/icon.svg",
				ShortDescription: "best chart",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it returns a valid AvailablePackageSummary if the minimal chart is correct",
			in: &models.Chart{
				Name: "foo",
				ID:   "foo/bar",
				Repo: &models.Repo{
					Name:      "bar",
					Namespace: "my-ns",
				},
				ChartVersions: []models.ChartVersion{
					{
						Version: "3.0.0",
					},
				},
			},
			expected: &corev1.AvailablePackageSummary{
				DisplayName:      "foo",
				LatestPkgVersion: "3.0.0",
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
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

// makeChart makes a chart with specific input used in the test and default constants for other relevant data.
func makeChart(chart_name, repo_name, namespace string, chart_versions []string) *models.Chart {
	ch := &models.Chart{
		Name:        chart_name,
		ID:          fmt.Sprintf("%s/%s", repo_name, chart_name),
		Category:    "cat1",
		Description: DefaultChartDescription,
		Icon:        DefaultChartIconURL,
		Maintainers: []chart.Maintainer{{Name: "me", Email: "me@me.me"}},
		Repo: &models.Repo{
			Name:      repo_name,
			Namespace: namespace,
		},
	}
	versions := []models.ChartVersion{}
	for _, v := range chart_versions {
		versions = append(versions, models.ChartVersion{
			Version:    v,
			AppVersion: DefaultAppVersion,
			Readme:     "chart readme",
			Values:     "chart values",
			Schema:     "chart schema",
		})
	}
	ch.ChartVersions = versions
	return ch
}

// makeChartRowsJSON returns a slice of paginated JSON chart info data.
func makeChartRowsJSON(t *testing.T, charts []*models.Chart, pageToken string, pageSize int) []string {
	// Simulate the pagination by reducing the rows of JSON based on the offset and limit.
	rowsJSON := []string{}
	for _, chart := range charts {
		chartJSON, err := json.Marshal(chart)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		rowsJSON = append(rowsJSON, string(chartJSON))
	}
	if len(rowsJSON) == 0 {
		return rowsJSON
	}

	if pageToken != "" {
		pageOffset, err := pageOffsetFromPageToken(pageToken)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if pageSize == 0 {
			t.Fatalf("pagesize must be > 0 when using a page token")
		}
		rowsJSON = rowsJSON[((pageOffset - 1) * pageSize):]
	}
	if pageSize > 0 && pageSize < len(rowsJSON) {
		rowsJSON = rowsJSON[0:pageSize]
	}
	return rowsJSON
}

// makeServer returns a server backed with an sql mock and a cleanup function
func makeServer(t *testing.T, authorized bool) (*Server, sqlmock.Sqlmock, func()) {
	// Creating the dynamic client
	dynamicClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{
			{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
		},
	)

	// Creating an authorized clientGetter
	clientSet := typfake.NewSimpleClientset()
	clientSet.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: authorized},
		}, nil
	})
	clientGetter := func(context.Context) (kubernetes.Interface, dynamic.Interface, error) {
		return clientSet, dynamicClient, nil
	}

	// Creating the SQL mock manager
	mock, cleanup, manager := setMockManager(t)

	return &Server{
		clientGetter:             clientGetter,
		manager:                  manager,
		globalPackagingNamespace: globalPackagingNamespace,
	}, mock, cleanup
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name             string
		charts           []*models.Chart
		expectDBQuery    bool
		statusCode       codes.Code
		request          *corev1.GetAvailablePackageSummariesRequest
		expectedResponse *corev1.GetAvailablePackageSummariesResponse
		authorized       bool
	}{
		{
			name:       "it returns a set of availablePackageSummary from the database (global ns)",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},
			expectDBQuery: true,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "my-ns", []string{"3.0.0"}),
				makeChart("chart-2", "repo-1", "my-ns", []string{"2.0.0"}),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "chart-1",
						LatestPkgVersion: "3.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "chart-2",
						LatestPkgVersion: "2.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns a set of availablePackageSummary from the database (specific ns)",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: "my-ns",
				},
			},
			expectDBQuery: true,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "my-ns", []string{"3.0.0"}),
				makeChart("chart-2", "repo-1", "my-ns", []string{"2.0.0"}),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "chart-1",
						LatestPkgVersion: "3.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						DisplayName:      "chart-2",
						LatestPkgVersion: "2.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns a unimplemented status if no namespaces is provided",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "",
				},
			},
			expectDBQuery: false,
			charts:        []*models.Chart{},
			statusCode:    codes.Unimplemented,
		},
		{
			name:       "it returns an internal error status if response does not contain version",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},
			expectDBQuery: true,
			charts:        []*models.Chart{makeChart("chart-1", "repo-1", "my-ns", []string{})},
			statusCode:    codes.Internal,
		},
		{
			name:       "it returns an unauthenticated status if the user doesn't have permissions",
			authorized: false,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "my-ns",
				},
			},
			expectDBQuery: false,
			charts:        []*models.Chart{{Name: "foo"}},
			statusCode:    codes.Unauthenticated,
		},
		{
			name:       "it returns only the requested page of results and includes the next page token",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "2",
					PageSize:  1,
				},
			},
			expectDBQuery: true,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "my-ns", []string{"3.0.0"}),
				makeChart("chart-2", "repo-1", "my-ns", []string{"2.0.0"}),
				makeChart("chart-3", "repo-1", "my-ns", []string{"1.0.0"}),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "chart-2",
						LatestPkgVersion: "2.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "3",
			},
		},
		{
			name:       "it returns the last page without a next page token",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				// Start on page two with two results per page, which in this input
				// corresponds only to the third chart.
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "2",
					PageSize:  2,
				},
			},
			expectDBQuery: true,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "my-ns", []string{"3.0.0"}),
				makeChart("chart-2", "repo-1", "my-ns", []string{"2.0.0"}),
				makeChart("chart-3", "repo-1", "my-ns", []string{"1.0.0"}),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackagesSummaries: []*corev1.AvailablePackageSummary{
					{
						DisplayName:      "chart-3",
						LatestPkgVersion: "1.0.0",
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Namespace: "my-ns"},
							Identifier: "repo-1/chart-3",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name:       "it returns an invalid argument error if the page token is invalid",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{
					PageToken: "this is not a page token",
					PageSize:  2,
				},
			},
			expectDBQuery: false,
			statusCode:    codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, mock, cleanup := makeServer(t, tc.authorized)
			defer cleanup()

			// Simulate the pagination by reducing the rows of JSON based on the offset and limit.
			// TODO(mnelson): We should check the LIMIT and OFFSET in the actual query as well.
			rowsJSON := makeChartRowsJSON(t, tc.charts, tc.request.GetPaginationOptions().GetPageToken(), int(tc.request.GetPaginationOptions().GetPageSize()))

			rows := sqlmock.NewRows([]string{"info"})
			for _, row := range rowsJSON {
				rows.AddRow(row)
			}

			if tc.expectDBQuery {
				// Checking if the WHERE condtion is properly applied
				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.request.Context.Namespace, server.globalPackagingNamespace).
					WillReturnRows(rows)
				if tc.request.GetPaginationOptions().GetPageSize() > 0 {
					mock.ExpectQuery("SELECT count").
						WithArgs(tc.request.Context.Namespace, server.globalPackagingNamespace).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				}
			}
			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageSummariesResponse{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{})
				if got, want := availablePackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
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
			name:  "it returns AvailablePackageDetail if the chart is correct",
			chart: makeChart("foo", "repo-1", "my-ns", []string{"3.0.0"}),
			expected: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				IconUrl:          DefaultChartIconURL,
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				PkgVersion:       "3.0.0",
				AppVersion:       DefaultAppVersion,
				Readme:           "chart readme",
				DefaultValues:    "chart values",
				ValuesSchema:     "chart schema",
				Maintainers:      []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
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
	testCases := []struct {
		name             string
		charts           []*models.Chart
		requestedVersion string
		expectedPackage  *corev1.AvailablePackageDetail
		statusCode       codes.Code
		request          *corev1.GetAvailablePackageDetailRequest
		authorized       bool
	}{
		{
			name:       "it returns an availablePackageDetail from the database (latest version)",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts: []*models.Chart{makeChart("foo", "repo-1", "my-ns", []string{"3.0.0"})},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				IconUrl:          DefaultChartIconURL,
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				PkgVersion:       "3.0.0",
				AppVersion:       DefaultAppVersion,
				Readme:           "chart readme",
				DefaultValues:    "chart values",
				ValuesSchema:     "chart schema",
				Maintainers:      []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns an availablePackageDetail from the database (specific version)",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			requestedVersion: "1.0.0",
			charts:           []*models.Chart{makeChart("foo", "repo-1", "my-ns", []string{"3.0.0", "2.0.0", "1.0.0"})},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				IconUrl:          DefaultChartIconURL,
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				PkgVersion:       "1.0.0",
				AppVersion:       DefaultAppVersion,
				Readme:           "chart readme",
				DefaultValues:    "chart values",
				ValuesSchema:     "chart schema",
				Maintainers:      []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns an internal error status if the chart is invalid",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:          []*models.Chart{{Name: "foo"}},
			expectedPackage: &corev1.AvailablePackageDetail{},
			statusCode:      codes.Internal,
		},
		{
			name:       "it returns an internal error status if the requested chart version doesn't exist",
			authorized: true,
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
			name:       "it returns an unauthenticated status if the user doesn't have permissions",
			authorized: false,
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
			server, mock, cleanup := makeServer(t, tc.authorized)
			defer cleanup()

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
				PkgVersion: tc.requestedVersion,
			}
			availablePackageDetails, err := server.GetAvailablePackageDetail(context.Background(), req)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{})
				if got, want := availablePackageDetails.AvailablePackageDetail, tc.expectedPackage; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}

			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name               string
		charts             []*models.Chart
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
					Identifier: "bitnami/apache",
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
			name:   "it returns the package version summary",
			charts: []*models.Chart{makeChart("apache", "bitnami", "kubeapps", []string{"3.0.0", "2.0.0", "1.0.0"})},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Namespace: "kubeapps",
					},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
					{
						PkgVersion: "3.0.0",
						AppVersion: DefaultAppVersion,
					},
					{
						PkgVersion: "2.0.0",
						AppVersion: DefaultAppVersion,
					},
					{
						PkgVersion: "1.0.0",
						AppVersion: DefaultAppVersion,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			server, mock, cleanup := makeServer(t, authorized)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			if tc.expectedStatusCode == codes.OK {
				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.request.AvailablePackageRef.Context.Namespace, tc.request.AvailablePackageRef.Identifier).
					WillReturnRows(rows)
			}

			response, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{})
			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestPackageAppVersionsSummary(t *testing.T) {
	testCases := []struct {
		name            string
		chart_versions  []models.ChartVersion
		version_summary []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion
	}{
		{
			name: "it includes the latest three major versions only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "7.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "5.5.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "7.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
			},
		},
		{
			name: "it includes the latest three minor versions for each major version only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.6", AppVersion: DefaultAppVersion},
			},
		},
		{
			name: "it includes the latest three patch versions for each minor version only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
			},
		},
		{
			name: "it includes the latest three patch versions of the latest three minor versions of the latest three major versions only",
			chart_versions: []models.ChartVersion{
				{Version: "8.5.6", AppVersion: DefaultAppVersion},
				{Version: "8.5.5", AppVersion: DefaultAppVersion},
				{Version: "8.5.4", AppVersion: DefaultAppVersion},
				{Version: "8.5.3", AppVersion: DefaultAppVersion},
				{Version: "8.4.6", AppVersion: DefaultAppVersion},
				{Version: "8.4.5", AppVersion: DefaultAppVersion},
				{Version: "8.4.4", AppVersion: DefaultAppVersion},
				{Version: "8.4.3", AppVersion: DefaultAppVersion},
				{Version: "8.3.6", AppVersion: DefaultAppVersion},
				{Version: "8.3.5", AppVersion: DefaultAppVersion},
				{Version: "8.3.4", AppVersion: DefaultAppVersion},
				{Version: "8.3.3", AppVersion: DefaultAppVersion},
				{Version: "8.2.6", AppVersion: DefaultAppVersion},
				{Version: "8.2.5", AppVersion: DefaultAppVersion},
				{Version: "8.2.4", AppVersion: DefaultAppVersion},
				{Version: "8.2.3", AppVersion: DefaultAppVersion},
				{Version: "6.5.6", AppVersion: DefaultAppVersion},
				{Version: "6.5.5", AppVersion: DefaultAppVersion},
				{Version: "6.5.4", AppVersion: DefaultAppVersion},
				{Version: "6.5.3", AppVersion: DefaultAppVersion},
				{Version: "6.4.6", AppVersion: DefaultAppVersion},
				{Version: "6.4.5", AppVersion: DefaultAppVersion},
				{Version: "6.4.4", AppVersion: DefaultAppVersion},
				{Version: "6.4.3", AppVersion: DefaultAppVersion},
				{Version: "6.3.6", AppVersion: DefaultAppVersion},
				{Version: "6.3.5", AppVersion: DefaultAppVersion},
				{Version: "6.3.4", AppVersion: DefaultAppVersion},
				{Version: "6.3.3", AppVersion: DefaultAppVersion},
				{Version: "6.2.6", AppVersion: DefaultAppVersion},
				{Version: "6.2.5", AppVersion: DefaultAppVersion},
				{Version: "6.2.4", AppVersion: DefaultAppVersion},
				{Version: "6.2.3", AppVersion: DefaultAppVersion},
				{Version: "4.5.6", AppVersion: DefaultAppVersion},
				{Version: "4.5.5", AppVersion: DefaultAppVersion},
				{Version: "4.5.4", AppVersion: DefaultAppVersion},
				{Version: "4.5.3", AppVersion: DefaultAppVersion},
				{Version: "4.4.6", AppVersion: DefaultAppVersion},
				{Version: "4.4.5", AppVersion: DefaultAppVersion},
				{Version: "4.4.4", AppVersion: DefaultAppVersion},
				{Version: "4.4.3", AppVersion: DefaultAppVersion},
				{Version: "4.3.6", AppVersion: DefaultAppVersion},
				{Version: "4.3.5", AppVersion: DefaultAppVersion},
				{Version: "4.3.4", AppVersion: DefaultAppVersion},
				{Version: "4.3.3", AppVersion: DefaultAppVersion},
				{Version: "4.2.6", AppVersion: DefaultAppVersion},
				{Version: "4.2.5", AppVersion: DefaultAppVersion},
				{Version: "4.2.4", AppVersion: DefaultAppVersion},
				{Version: "4.2.3", AppVersion: DefaultAppVersion},
				{Version: "2.5.6", AppVersion: DefaultAppVersion},
				{Version: "2.5.5", AppVersion: DefaultAppVersion},
				{Version: "2.5.4", AppVersion: DefaultAppVersion},
				{Version: "2.5.3", AppVersion: DefaultAppVersion},
				{Version: "2.4.6", AppVersion: DefaultAppVersion},
				{Version: "2.4.5", AppVersion: DefaultAppVersion},
				{Version: "2.4.4", AppVersion: DefaultAppVersion},
				{Version: "2.4.3", AppVersion: DefaultAppVersion},
				{Version: "2.3.6", AppVersion: DefaultAppVersion},
				{Version: "2.3.5", AppVersion: DefaultAppVersion},
				{Version: "2.3.4", AppVersion: DefaultAppVersion},
				{Version: "2.3.3", AppVersion: DefaultAppVersion},
				{Version: "2.2.6", AppVersion: DefaultAppVersion},
				{Version: "2.2.5", AppVersion: DefaultAppVersion},
				{Version: "2.2.4", AppVersion: DefaultAppVersion},
				{Version: "2.2.3", AppVersion: DefaultAppVersion},
			},
			version_summary: []*corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{
				{PkgVersion: "8.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "8.3.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "6.3.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.5.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.4.4", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.6", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.5", AppVersion: DefaultAppVersion},
				{PkgVersion: "4.3.4", AppVersion: DefaultAppVersion},
			},
		},
	}

	opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse_PackageAppVersion{})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got, want := packageAppVersionsSummary(tc.chart_versions), tc.version_summary; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}
