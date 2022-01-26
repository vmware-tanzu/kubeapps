// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	cmp "github.com/google/go-cmp/cmp"
	cmpopts "github.com/google/go-cmp/cmp/cmpopts"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	assetmanager "github.com/kubeapps/kubeapps/cmd/assetsvc/pkg/utils"
	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	pluginsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	pkghelmv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	paginate "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	pkgutils "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	helmagent "github.com/kubeapps/kubeapps/pkg/agent"
	chartutilsfake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	chartmodels "github.com/kubeapps/kubeapps/pkg/chart/models"
	dbutils "github.com/kubeapps/kubeapps/pkg/dbutils"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmchartutil "helm.sh/helm/v3/pkg/chartutil"
	helmkube "helm.sh/helm/v3/pkg/kube"
	helmkubefake "helm.sh/helm/v3/pkg/kube/fake"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8sauthorizationv1 "k8s.io/api/authorization/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8dynamicclient "k8s.io/client-go/dynamic"
	k8dynamicclientfake "k8s.io/client-go/dynamic/fake"
	k8stypedclient "k8s.io/client-go/kubernetes"
	k8stypedclientfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	log "k8s.io/klog/v2"
	k8syaml "sigs.k8s.io/yaml"
)

const (
	globalPackagingNamespace = "kubeapps"
	globalPackagingCluster   = "default"
	DefaultAppVersion        = "1.2.6"
	DefaultReleaseRevision   = 1
	DefaultChartDescription  = "default chart description"
	DefaultChartIconURL      = "https://example.com/helmchart.svg"
	DefaultChartHomeURL      = "https://helm.sh/helm"
	DefaultChartCategory     = "cat1"
)

func setMockManager(t *testing.T) (sqlmock.Sqlmock, func(), assetmanager.AssetManager) {
	var manager assetmanager.AssetManager
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	manager = &assetmanager.PostgresAssetManager{&dbutils.PostgresAssetManager{DB: db, GlobalReposNamespace: globalPackagingNamespace}}
	return mock, func() { db.Close() }, manager
}

func TestGetClient(t *testing.T) {
	dbConfig := dbutils.Config{URL: "localhost:5432", Database: "assetsvc", Username: "postgres", Password: "password"}
	manager, err := assetmanager.NewPGManager(dbConfig, globalPackagingNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}
	testClientGetter := func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		return k8stypedclientfake.NewSimpleClientset(), k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
			k8sruntime.NewScheme(),
			map[k8sschema.GroupVersionResource]string{
				{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
			},
		), nil
	}

	testCases := []struct {
		name              string
		manager           assetmanager.AssetManager
		clientGetter      clientgetter.ClientGetterFunc
		statusCodeClient  grpccodes.Code
		statusCodeManager grpccodes.Code
	}{
		{
			name:              "it returns internal error status when no clientGetter configured",
			manager:           manager,
			clientGetter:      nil,
			statusCodeClient:  grpccodes.Internal,
			statusCodeManager: grpccodes.OK,
		},
		{
			name:              "it returns internal error status when no manager configured",
			manager:           nil,
			clientGetter:      testClientGetter,
			statusCodeClient:  grpccodes.OK,
			statusCodeManager: grpccodes.Internal,
		},
		{
			name:              "it returns internal error status when no clientGetter/manager configured",
			manager:           nil,
			clientGetter:      nil,
			statusCodeClient:  grpccodes.Internal,
			statusCodeManager: grpccodes.Internal,
		},
		{
			name:    "it returns failed-precondition when configGetter itself errors",
			manager: manager,
			clientGetter: func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
				return nil, nil, fmt.Errorf("Bang!")
			},
			statusCodeClient:  grpccodes.FailedPrecondition,
			statusCodeManager: grpccodes.OK,
		},
		{
			name:         "it returns client without error when configured correctly",
			manager:      manager,
			clientGetter: testClientGetter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Server{clientGetter: tc.clientGetter, manager: tc.manager}

			typedClient, dynamicClient, errClient := s.GetClients(context.Background(), "")

			if got, want := grpcstatus.Code(errClient), tc.statusCodeClient; got != want {
				t.Errorf("got: %+v, want: %+v", got, want)
			}

			_, errManager := s.GetManager()

			if got, want := grpcstatus.Code(errManager), tc.statusCodeManager; got != want {
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

// makeChart makes a chart with specific input used in the test and default constants for other relevant data.
func makeChart(chart_name, repo_name, repo_url, namespace string, chart_versions []string, category string) *chartmodels.Chart {
	ch := &chartmodels.Chart{
		Name:        chart_name,
		ID:          fmt.Sprintf("%s/%s", repo_name, chart_name),
		Category:    category,
		Description: DefaultChartDescription,
		Home:        DefaultChartHomeURL,
		Icon:        DefaultChartIconURL,
		Maintainers: []helmchart.Maintainer{{Name: "me", Email: "me@me.me"}},
		Sources:     []string{"http://source-1"},
		Repo: &chartmodels.Repo{
			Name:      repo_name,
			Namespace: namespace,
			URL:       repo_url,
		},
	}
	versions := []chartmodels.ChartVersion{}
	for _, v := range chart_versions {
		versions = append(versions, chartmodels.ChartVersion{
			Version:    v,
			AppVersion: DefaultAppVersion,
			Readme:     "not-used",
			Values:     "not-used",
			Schema:     "not-used",
		})
	}
	ch.ChartVersions = versions
	return ch
}

// makeChartRowsJSON returns a slice of paginated JSON chart info data.
func makeChartRowsJSON(t *testing.T, charts []*chartmodels.Chart, pageToken string, pageSize int) []string {
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
		pageOffset, err := paginate.PageOffsetFromPageToken(pageToken)
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
func makeServer(t *testing.T, authorized bool, actionConfig *helmaction.Configuration, objects ...k8sruntime.Object) (*Server, sqlmock.Sqlmock, func()) {
	// Creating the dynamic client
	scheme := k8sruntime.NewScheme()
	err := apprepov1alpha1.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	dynamicClient := k8dynamicclientfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[k8sschema.GroupVersionResource]string{
			{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
		},
		objects...,
	)

	// Creating an authorized clientGetter
	clientSet := k8stypedclientfake.NewSimpleClientset()
	clientSet.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
		return true, &k8sauthorizationv1.SelfSubjectAccessReview{
			Status: k8sauthorizationv1.SubjectAccessReviewStatus{Allowed: authorized},
		}, nil
	})
	clientGetter := func(context.Context, string) (k8stypedclient.Interface, k8dynamicclient.Interface, error) {
		return clientSet, dynamicClient, nil
	}

	// Creating the SQL mock manager
	mock, cleanup, manager := setMockManager(t)

	return &Server{
		clientGetter:             clientGetter,
		manager:                  manager,
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		actionConfigGetter: func(context.Context, *pkgsGRPCv1alpha1.Context) (*helmaction.Configuration, error) {
			return actionConfig, nil
		},
		chartClientFactory: &chartutilsfake.ChartClientFactory{},
		versionsInSummary:  pkgutils.GetDefaultVersionsInSummary(),
		createReleaseFunc:  helmagent.CreateRelease,
	}, mock, cleanup
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name                   string
		charts                 []*chartmodels.Chart
		expectDBQueryNamespace string
		statusCode             grpccodes.Code
		request                *pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest
		expectedResponse       *pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse
		authorized             bool
		expectedCategories     []*chartmodels.ChartCategory
	}{
		{
			name:       "it returns a set of availablePackageSummary from the database (global ns)",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3-global", "repo-1", "http://chart-3", globalPackagingNamespace, []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-3-global",
						DisplayName: "chart-3-global",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: globalPackagingNamespace},
							Identifier: "repo-1/chart-3-global",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns a set of availablePackageSummary from the database (specific ns)",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: "my-ns",
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns a set of the global availablePackageSummary from the database (not the specific ns on other cluster)",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "other",
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns a unimplemented status if no namespaces is provided",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Namespace: "",
				},
			},
			charts:     []*chartmodels.Chart{},
			statusCode: grpccodes.Unimplemented,
		},
		{
			name:       "it returns an internal error status if response does not contain version",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts:                 []*chartmodels.Chart{makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{}, DefaultChartCategory)},
			statusCode:             grpccodes.Internal,
		},
		{
			name:       "it returns an unauthenticated status if the user doesn't have permissions",
			authorized: false,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Namespace: "my-ns",
				},
			},
			charts:     []*chartmodels.Chart{{Name: "foo"}},
			statusCode: grpccodes.Unauthenticated,
		},
		{
			name:       "it returns only the requested page of results and includes the next page token",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageToken: "2",
					PageSize:  1,
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						Categories:       []string{DefaultChartCategory},
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "3",
				Categories:    []string{"cat1"},
			},
		},
		{
			name:       "it returns the last page without a next page token",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				// Start on page two with two results per page, which in this input
				// corresponds only to the third helmchart.
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageToken: "2",
					PageSize:  2,
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-3",
						DisplayName: "chart-3",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-3",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "",
				Categories:    []string{"cat1"},
			},
		},
		{
			name:       "it returns an invalid argument error if the page token is invalid",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageToken: "this is not a page token",
					PageSize:  2,
				},
			},
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name:       "it returns the proper chart categories",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{
					Cluster:   "",
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: "my-ns",
			charts: []*chartmodels.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, "foo"),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, "bar"),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, "bar"),
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*pkgsGRPCv1alpha1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"foo"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"bar"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-3",
						DisplayName: "chart-3",
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"bar"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
							Context:    &pkgsGRPCv1alpha1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-3",
							Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"bar", "foo"},
			},
			statusCode: grpccodes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, mock, cleanup := makeServer(t, tc.authorized, nil)
			defer cleanup()

			// Simulate the pagination by reducing the rows of JSON based on the offset and limit.
			// TODO(mnelson): We should check the LIMIT and OFFSET in the actual query as well.
			rowsJSON := makeChartRowsJSON(t, tc.charts, tc.request.GetPaginationOptions().GetPageToken(), int(tc.request.GetPaginationOptions().GetPageSize()))

			rows := sqlmock.NewRows([]string{"info"})
			for _, row := range rowsJSON {
				rows.AddRow(row)
			}

			if tc.expectDBQueryNamespace != "" {
				// Checking if the WHERE condtion is properly applied

				// Check returned categories
				catrows := sqlmock.NewRows([]string{"name", "count"})

				// Generate the categories from the tc.charts input
				dict := make(map[string]int)
				for _, chart := range tc.charts {
					dict[chart.Category] = dict[chart.Category] + 1
				}
				// Ensure we've got a fixed order for the results.
				categories := []string{}
				for category := range dict {
					categories = append(categories, category)
				}
				sort.Strings(categories)
				for _, category := range categories {
					catrows.AddRow(category, dict[category])
				}

				mock.ExpectQuery("SELECT (info ->> 'category')*").
					WithArgs(tc.expectDBQueryNamespace, server.globalPackagingNamespace).
					WillReturnRows(catrows)

				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.expectDBQueryNamespace, server.globalPackagingNamespace).
					WillReturnRows(rows)

				if tc.request.GetPaginationOptions().GetPageSize() > 0 {
					mock.ExpectQuery("SELECT count").
						WithArgs(tc.request.Context.Namespace, server.globalPackagingNamespace).
						WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
				}
			}

			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				opt1 := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.GetAvailablePackageSummariesResponse{}, pkgsGRPCv1alpha1.AvailablePackageSummary{}, pkgsGRPCv1alpha1.AvailablePackageReference{}, pkgsGRPCv1alpha1.Context{}, pluginsGRPCv1alpha1.Plugin{}, pkgsGRPCv1alpha1.PackageAppVersion{})
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
		chart      *chartmodels.Chart
		chartFiles *chartmodels.ChartFiles
		expected   *pkgsGRPCv1alpha1.AvailablePackageDetail
		statusCode grpccodes.Code
	}{
		{
			name:  "it returns AvailablePackageDetail if the chart is correct",
			chart: makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
			chartFiles: &chartmodels.ChartFiles{
				Readme: "chart readme",
				Values: "chart values",
				Schema: "chart schema",
			},
			expected: &pkgsGRPCv1alpha1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				RepoUrl:          "http://foo",
				HomeUrl:          DefaultChartHomeURL,
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				Version: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*pkgsGRPCv1alpha1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns internal error if empty chart",
			chart:      &chartmodels.Chart{},
			statusCode: grpccodes.Internal,
		},
		{
			name:       "it returns internal error if chart is invalid",
			chart:      &chartmodels.Chart{Name: "foo"},
			statusCode: grpccodes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			availablePackageDetail, err := AvailablePackageDetailFromChart(tc.chart, tc.chartFiles)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				opt1 := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.AvailablePackageDetail{}, pkgsGRPCv1alpha1.AvailablePackageSummary{}, pkgsGRPCv1alpha1.AvailablePackageReference{}, pkgsGRPCv1alpha1.Context{}, pluginsGRPCv1alpha1.Plugin{}, pkgsGRPCv1alpha1.Maintainer{}, pkgsGRPCv1alpha1.PackageAppVersion{})
				if got, want := availablePackageDetail, tc.expected; !cmp.Equal(got, want, opt1) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opt1))
				}
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		name            string
		charts          []*chartmodels.Chart
		expectedPackage *pkgsGRPCv1alpha1.AvailablePackageDetail
		statusCode      grpccodes.Code
		request         *pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest
		authorized      bool
	}{
		{
			name:       "it returns an availablePackageDetail from the database (latest version)",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "repo-1%2Ffoo",
				},
			},
			charts: []*chartmodels.Chart{makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0"}, DefaultChartCategory)},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				HomeUrl:          DefaultChartHomeURL,
				RepoUrl:          "http://foo",
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				Version: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*pkgsGRPCv1alpha1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns an availablePackageDetail from the database (specific version)",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
				PkgVersion: "1.0.0",
			},
			charts: []*chartmodels.Chart{makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0", "2.0.0", "1.0.0"}, DefaultChartCategory)},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				HomeUrl:          DefaultChartHomeURL,
				RepoUrl:          "http://foo",
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				Version: &pkgsGRPCv1alpha1.PackageAppVersion{
					PkgVersion: "1.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*pkgsGRPCv1alpha1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &pluginsGRPCv1alpha1.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: grpccodes.OK,
		},
		{
			name:       "it returns an invalid arg error status if no context is provided",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Identifier: "foo/bar",
				},
			},
			charts:     []*chartmodels.Chart{{Name: "foo"}},
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name:       "it returns an invalid arg error status if cluster is not the global/kubeapps one",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "other-cluster", Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:     []*chartmodels.Chart{{Name: "foo"}},
			statusCode: grpccodes.InvalidArgument,
		},
		{
			name:       "it returns an internal error status if the chart is invalid",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:          []*chartmodels.Chart{{Name: "foo"}},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{},
			statusCode:      grpccodes.Internal,
		},
		{
			name:       "it returns an internal error status if the requested chart version doesn't exist",
			authorized: true,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
				PkgVersion: "9.9.9",
			},
			charts:          []*chartmodels.Chart{{Name: "foo"}},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{},
			statusCode:      grpccodes.Internal,
		},
		{
			name:       "it returns an unauthenticated status if the user doesn't have permissions",
			authorized: false,
			request: &pkgsGRPCv1alpha1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:          []*chartmodels.Chart{{Name: "foo"}},
			expectedPackage: &pkgsGRPCv1alpha1.AvailablePackageDetail{},
			statusCode:      grpccodes.Unauthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, mock, cleanup := makeServer(t, tc.authorized, nil)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			if tc.statusCode == grpccodes.OK {
				// Checking if the WHERE condition is properly applied
				chartIDUnescaped, err := url.QueryUnescape(tc.request.AvailablePackageRef.Identifier)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mock.ExpectQuery("SELECT info FROM charts").
					WithArgs(tc.request.AvailablePackageRef.Context.Namespace, chartIDUnescaped).
					WillReturnRows(rows)
				fileID := fileIDForChart(chartIDUnescaped, tc.expectedPackage.Version.PkgVersion)
				fileJSON, err := json.Marshal(chartmodels.ChartFiles{
					Readme: tc.expectedPackage.Readme,
					Values: tc.expectedPackage.DefaultValues,
					Schema: tc.expectedPackage.ValuesSchema,
				})
				if err != nil {
					t.Fatalf("%+v", err)
				}
				fileRows := sqlmock.NewRows([]string{"info"})
				fileRows.AddRow(string(fileJSON))
				mock.ExpectQuery("SELECT info FROM files").
					WithArgs(tc.request.GetAvailablePackageRef().GetContext().GetNamespace(), fileID).
					WillReturnRows(fileRows)
			}

			availablePackageDetails, err := server.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == grpccodes.OK {
				opt1 := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.AvailablePackageDetail{}, pkgsGRPCv1alpha1.AvailablePackageSummary{}, pkgsGRPCv1alpha1.AvailablePackageReference{}, pkgsGRPCv1alpha1.Context{}, pluginsGRPCv1alpha1.Plugin{}, pkgsGRPCv1alpha1.Maintainer{}, pkgsGRPCv1alpha1.PackageAppVersion{})
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
		charts             []*chartmodels.Chart
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
					Identifier: "bitnami/apache",
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
			name: "it returns invalid argument if called with a cluster other than the global/kubeapps one",
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context:    &pkgsGRPCv1alpha1.Context{Cluster: "other-cluster", Namespace: "kubeapps"},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: grpccodes.InvalidArgument,
		},
		{
			name:   "it returns the package version summary",
			charts: []*chartmodels.Chart{makeChart("apache", "bitnami", "http://apache", "kubeapps", []string{"3.0.0", "2.0.0", "1.0.0"}, DefaultChartCategory)},
			request: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: "kubeapps",
					},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*pkgsGRPCv1alpha1.PackageAppVersion{
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
			server, mock, cleanup := makeServer(t, authorized, nil)
			defer cleanup()

			rows := sqlmock.NewRows([]string{"info"})

			for _, chart := range tc.charts {
				chartJSON, err := json.Marshal(chart)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				rows.AddRow(string(chartJSON))
			}
			if tc.expectedStatusCode == grpccodes.OK {
				mock.ExpectQuery("SELECT info FROM").
					WithArgs(tc.request.AvailablePackageRef.Context.Namespace, tc.request.AvailablePackageRef.Identifier).
					WillReturnRows(rows)
			}

			response, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.GetAvailablePackageVersionsResponse{}, pkgsGRPCv1alpha1.PackageAppVersion{})
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

func TestParsePluginConfig(t *testing.T) {
	testCases := []struct {
		name                    string
		pluginYAMLConf          []byte
		exp_versions_in_summary pkgutils.VersionsInSummary
		exp_error_str           string
	}{
		{
			name:                    "non existing plugin-config file",
			pluginYAMLConf:          nil,
			exp_versions_in_summary: pkgutils.VersionsInSummary{0, 0, 0},
			exp_error_str:           "no such file or directory",
		},
		{
			name: "non-default plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{4, 2, 1},
			exp_error_str:           "",
		},
		{
			name: "partial params in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 1
        `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{1, 0, 0},
			exp_error_str:           "",
		},
		{
			name: "invalid plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1-IFC-123
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{},
			exp_error_str:           "json: cannot unmarshal",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
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
			versions_in_summary, _, goterr := parsePluginConfig(filename)
			if goterr != nil && !strings.Contains(goterr.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", goterr.Error(), tc.exp_error_str)
			}
			if got, want := versions_in_summary, tc.exp_versions_in_summary; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}
func TestParsePluginConfigTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		pluginYAMLConf []byte
		exp_timeout    int32
		exp_error_str  string
	}{
		{
			name:           "no timeout specified in plugin config",
			pluginYAMLConf: nil,
			exp_timeout:    0,
			exp_error_str:  "",
		},
		{
			name: "specific timeout in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      timeoutSeconds: 650
      `),
			exp_timeout:   650,
			exp_error_str: "",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
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
			_, timeoutSeconds, goterr := parsePluginConfig(filename)
			if goterr != nil && !strings.Contains(goterr.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", goterr.Error(), tc.exp_error_str)
			}
			if got, want := timeoutSeconds, tc.exp_timeout; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}
func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		request            *pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest
		existingReleases   []releaseStub
		expectedStatusCode grpccodes.Code
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "returns installed packages in a specific namespace",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       helmrelease.StatusDeployed,
					version:      2,
				},
				{
					name:      "my-release-2",
					namespace: "other-namespace",
					status:    helmrelease.StatusDeployed,
					version:   4,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-1",
					chartVersion: "4.5.6",
					status:       helmrelease.StatusDeployed,
					version:      6,
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-3",
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
			},
		},
		{
			name: "returns installed packages across all namespaces",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       helmrelease.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-2",
							},
							Identifier: "my-release-2",
						},
						Name:    "my-release-2",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "3.4.5",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "3.4.5",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.4.5",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-3",
							},
							Identifier: "my-release-3",
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
			},
		},
		{
			name: "returns limited results",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageSize: 2,
				},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       helmrelease.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-2",
							},
							Identifier: "my-release-2",
						},
						Name:    "my-release-2",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "3.4.5",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "3.4.5",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "3.4.5",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
				NextPageToken: "3",
			},
		},
		{
			name: "fetches results from an offset",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: ""},
				PaginationOptions: &pkgsGRPCv1alpha1.PaginationOptions{
					PageSize:  2,
					PageToken: "2",
				},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       helmrelease.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-3",
							},
							Identifier: "my-release-3",
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "includes a latest package version when available",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesRequest{
				Context: &pkgsGRPCv1alpha1.Context{Namespace: "namespace-1"},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       helmrelease.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: grpccodes.OK,
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*pkgsGRPCv1alpha1.InstalledPackageSummary{
					{
						InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
							Context: &pkgsGRPCv1alpha1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
							PkgVersion: "1.2.5",
						},
						Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
							Ready:      true,
							Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetContext().GetNamespace(), tc.existingReleases, nil)
			server, mock, cleanup := makeServer(t, authorized, actionConfig)
			defer cleanup()

			if tc.expectedStatusCode == grpccodes.OK {
				populateAssetDBWithSummaries(t, mock, tc.expectedResponse.InstalledPackageSummaries)
			}

			response, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.GetInstalledPackageSummariesResponse{}, pkgsGRPCv1alpha1.InstalledPackageSummary{}, pkgsGRPCv1alpha1.InstalledPackageReference{}, pkgsGRPCv1alpha1.Context{}, pkgsGRPCv1alpha1.VersionReference{}, pkgsGRPCv1alpha1.InstalledPackageStatus{}, pkgsGRPCv1alpha1.PackageAppVersion{})
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

func TestGetInstalledPackageDetail(t *testing.T) {
	customDetailRevision2, err := anypb.New(&pkghelmv1alpha1.InstalledPackageDetailCustomDataHelm{
		ReleaseRevision: 2,
	})
	if err != nil {
		t.Fatalf("%+v", err)
	}
	const (
		releaseNamespace = "my-namespace-1"
		releaseName      = "my-release-1"
		releaseVersion   = "1.2.3"
		releaseValues    = "{\"value\":\"new\"}"
		releaseNotes     = "some notes"
	)
	testCases := []struct {
		name               string
		existingReleases   []releaseStub
		request            *pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest
		expectedResponse   *pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse
		expectedStatusCode grpccodes.Code
	}{
		{
			name: "returns an installed package detail",
			existingReleases: []releaseStub{
				{
					name:           releaseName,
					namespace:      releaseNamespace,
					chartVersion:   releaseVersion,
					chartNamespace: releaseNamespace,
					values:         releaseValues,
					notes:          releaseNotes,
					status:         helmrelease.StatusSuperseded,
					version:        1,
				},
				{
					name:           releaseName,
					namespace:      releaseNamespace,
					chartVersion:   releaseVersion,
					chartNamespace: releaseNamespace,
					values:         releaseValues,
					notes:          releaseNotes,
					status:         helmrelease.StatusDeployed,
					version:        2,
				},
			},
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: releaseNamespace,
						Cluster:   globalPackagingCluster,
					},
					Identifier: releaseName,
				},
			},
			expectedResponse: &pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: &pkgsGRPCv1alpha1.InstalledPackageDetail{
					InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
						Context: &pkgsGRPCv1alpha1.Context{
							Namespace: releaseNamespace,
							Cluster:   globalPackagingCluster,
						},
						Identifier: releaseName,
					},
					PkgVersionReference: &pkgsGRPCv1alpha1.VersionReference{
						Version: releaseVersion,
					},
					Name: releaseName,
					CurrentVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: releaseVersion,
						AppVersion: DefaultAppVersion,
					},
					LatestVersion: &pkgsGRPCv1alpha1.PackageAppVersion{
						PkgVersion: releaseVersion,
						AppVersion: DefaultAppVersion,
					},
					ValuesApplied:         releaseValues,
					PostInstallationNotes: releaseNotes,
					Status: &pkgsGRPCv1alpha1.InstalledPackageStatus{
						Ready:      true,
						Reason:     pkgsGRPCv1alpha1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "deployed",
					},
					AvailablePackageRef: &pkgsGRPCv1alpha1.AvailablePackageReference{
						Context: &pkgsGRPCv1alpha1.Context{
							Namespace: releaseNamespace,
							Cluster:   globalPackagingCluster,
						},
						Identifier: "myrepo/" + releaseName,
						Plugin:     GetPluginDetail(),
					},
					CustomDetail: customDetailRevision2,
				},
			},
			expectedStatusCode: grpccodes.OK,
		},
		{
			name: "returns a 404 if the installed package is not found",
			request: &pkgsGRPCv1alpha1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &pkgsGRPCv1alpha1.InstalledPackageReference{
					Context: &pkgsGRPCv1alpha1.Context{
						Namespace: releaseNamespace,
					},
					Identifier: releaseName,
				},
			},
			expectedStatusCode: grpccodes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)
			server, mock, cleanup := makeServer(t, authorized, actionConfig)
			defer cleanup()

			if tc.expectedStatusCode == grpccodes.OK {
				populateAssetDBWithDetail(t, mock, tc.expectedResponse.InstalledPackageDetail)
			}

			response, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := grpcstatus.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK grpccodes.
			if tc.expectedStatusCode != grpccodes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(pkgsGRPCv1alpha1.GetInstalledPackageDetailResponse{}, pkgsGRPCv1alpha1.InstalledPackageDetail{}, pkgsGRPCv1alpha1.InstalledPackageReference{}, pkgsGRPCv1alpha1.Context{}, pkgsGRPCv1alpha1.VersionReference{}, pkgsGRPCv1alpha1.InstalledPackageStatus{}, pkgsGRPCv1alpha1.AvailablePackageReference{}, pluginsGRPCv1alpha1.Plugin{}, pkgsGRPCv1alpha1.PackageAppVersion{}, anypb.Any{})
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

func TestChartTarballURLBuild(t *testing.T) {
	testCases := []struct {
		name         string
		repo         *chartmodels.Repo
		chartVersion *chartmodels.ChartVersion
		expectedUrl  string
	}{
		{
			name:         "tarball url with relative URL without leading slash in chart",
			repo:         &chartmodels.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &chartmodels.ChartVersion{URLs: []string{"chart/test"}},
			expectedUrl:  "https://demo.repo/repo1/chart/test",
		},
		{
			name:         "tarball url with relative URL with leading slash in chart",
			repo:         &chartmodels.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &chartmodels.ChartVersion{URLs: []string{"/chart/test"}},
			expectedUrl:  "https://demo.repo/repo1/chart/test",
		},
		{
			name:         "tarball url with absolute URL",
			repo:         &chartmodels.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &chartmodels.ChartVersion{URLs: []string{"https://demo.repo/repo1/chart/test"}},
			expectedUrl:  "https://demo.repo/repo1/chart/test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tarballUrl := chartTarballURL(tc.repo, *tc.chartVersion)

			if got, want := tarballUrl, tc.expectedUrl; got != want {
				t.Fatalf("got: %+v, want: %+v", got, want)
			}
		})
	}
}

// newActionConfigFixture returns an helmaction.Configuration with fake clients
// and memory helmstorage.
func newActionConfigFixture(t *testing.T, namespace string, rels []releaseStub, kubeClient helmkube.Interface) *helmaction.Configuration {
	t.Helper()

	memDriver := helmstoragedriver.NewMemory()

	if kubeClient == nil {
		kubeClient = &helmkubefake.FailingKubeClient{PrintingKubeClient: helmkubefake.PrintingKubeClient{Out: ioutil.Discard}}
	}

	actionConfig := &helmaction.Configuration{
		// Create the Releases storage explicitly so we can set the
		// internal log function used to see data in test output.
		Releases: &helmstorage.Storage{
			Driver: memDriver,
			Log: func(format string, v ...interface{}) {
				t.Logf(format, v...)
			},
		},
		KubeClient:   kubeClient,
		Capabilities: helmchartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}

	for _, r := range rels {
		rel := releaseForStub(t, r)
		err := actionConfig.Releases.Create(rel)
		if err != nil {
			t.Fatal(err)
		}
	}
	// It is the namespace of the driver which determines the results. In the prod code,
	// the actionConfigGetter sets this using StorageForSecrets(namespace, clientset).
	memDriver.SetNamespace(namespace)

	return actionConfig
}

func releaseForStub(t *testing.T, r releaseStub) *helmrelease.Release {
	config := map[string]interface{}{}
	if r.values != "" {
		err := json.Unmarshal([]byte(r.values), &config)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
	return &helmrelease.Release{
		Name:      r.name,
		Namespace: r.namespace,
		Manifest:  r.manifest,
		Version:   r.version,
		Info: &helmrelease.Info{
			Status: r.status,
			Notes:  r.notes,
		},
		Chart: &helmchart.Chart{
			Metadata: &helmchart.Metadata{
				Version:    r.chartVersion,
				Icon:       "https://example.com/icon.png",
				AppVersion: DefaultAppVersion,
			},
		},
		Config: config,
	}
}

func chartAssetForPackage(pkg *pkgsGRPCv1alpha1.InstalledPackageSummary) *chartmodels.Chart {
	chartVersions := []chartmodels.ChartVersion{}
	if pkg.LatestVersion.PkgVersion != "" {
		chartVersions = append(chartVersions, chartmodels.ChartVersion{
			Version: pkg.LatestVersion.PkgVersion,
		})
	}
	chartVersions = append(chartVersions, chartmodels.ChartVersion{
		Version: pkg.CurrentVersion.PkgVersion,
	})

	return &chartmodels.Chart{
		Name:          pkg.Name,
		ChartVersions: chartVersions,
	}
}

func chartAssetForReleaseStub(rel *releaseStub) *chartmodels.Chart {
	chartVersions := []chartmodels.ChartVersion{}
	if rel.latestVersion != "" {
		chartVersions = append(chartVersions, chartmodels.ChartVersion{
			Version: rel.latestVersion,
			URLs:    []string{fmt.Sprintf("https://example.com/%s-%s.tgz", rel.chartID, rel.latestVersion)},
		})
	}
	chartVersions = append(chartVersions, chartmodels.ChartVersion{
		Version:    rel.chartVersion,
		AppVersion: DefaultAppVersion,
	})

	return &chartmodels.Chart{
		Name: rel.name,
		ID:   rel.chartID,
		Repo: &chartmodels.Repo{
			Namespace: rel.chartNamespace,
		},
		ChartVersions: chartVersions,
	}
}

func populateAssetDBWithSummaries(t *testing.T, mock sqlmock.Sqlmock, pkgs []*pkgsGRPCv1alpha1.InstalledPackageSummary) {
	// The code currently executes one query per release in the paginated
	// results and should receive a single row response.
	rels := []releaseStub{}
	for _, pkg := range pkgs {
		rels = append(rels, releaseStub{
			name:          pkg.Name,
			namespace:     pkg.GetInstalledPackageRef().GetContext().GetNamespace(),
			chartVersion:  pkg.CurrentVersion.PkgVersion,
			latestVersion: pkg.LatestVersion.PkgVersion,
			version:       DefaultReleaseRevision,
		})
	}
	populateAssetDB(t, mock, rels)
}

func populateAssetDBWithDetail(t *testing.T, mock sqlmock.Sqlmock, pkg *pkgsGRPCv1alpha1.InstalledPackageDetail) {
	// The code currently executes one query per release in the paginated
	// results and should receive a single row response.
	rel := releaseStub{
		name:           pkg.Name,
		namespace:      pkg.GetInstalledPackageRef().GetContext().GetNamespace(),
		chartVersion:   pkg.GetCurrentVersion().GetPkgVersion(),
		chartID:        pkg.GetAvailablePackageRef().GetIdentifier(),
		chartNamespace: pkg.GetAvailablePackageRef().GetContext().GetNamespace(),
		version:        DefaultReleaseRevision,
	}
	populateAssetDB(t, mock, []releaseStub{rel})
}

func populateAssetForTarball(t *testing.T, mock sqlmock.Sqlmock, chartId, namespace, version string) {
	chart := &chartmodels.Chart{
		Name: chartId,
		ID:   chartId,
		Repo: &chartmodels.Repo{
			Namespace: globalPackagingNamespace,
		},
		ChartVersions: []chartmodels.ChartVersion{{
			Version: version,
			URLs:    []string{fmt.Sprintf("https://example.com/%s-%s.tgz", chartId, version)}}},
	}
	chartJSON, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	rows := sqlmock.NewRows([]string{"info"})
	rows.AddRow(string(chartJSON))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT info FROM charts WHERE repo_namespace = $1 AND chart_id ILIKE $2")).
		WithArgs(chart.Repo.Namespace, chart.ID).
		WillReturnRows(rows)
}

func populateAssetDB(t *testing.T, mock sqlmock.Sqlmock, rels []releaseStub) {
	// The code currently executes one query per release in the paginated
	// results and should receive a single row response.
	for _, rel := range rels {
		chartJSON, err := json.Marshal(chartAssetForReleaseStub(&rel))
		if err != nil {
			t.Fatalf("%+v", err)
		}
		rows := sqlmock.NewRows([]string{"info"})
		rows.AddRow(string(chartJSON))
		mock.ExpectQuery("SELECT info FROM").
			WillReturnRows(rows)
	}
}

type releaseStub struct {
	name           string
	namespace      string
	version        int
	chartVersion   string
	chartID        string
	chartNamespace string
	latestVersion  string
	values         string
	notes          string
	status         helmrelease.Status
	manifest       string
}
