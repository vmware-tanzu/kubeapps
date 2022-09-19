// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	helmv1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/helm/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/helm/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/helm/packages/v1alpha1/utils"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/agent"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/paginate"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/fake"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"github.com/vmware-tanzu/kubeapps/pkg/dbutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/kube"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	authorizationv1 "k8s.io/api/authorization/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/storage/names"
	dynfake "k8s.io/client-go/dynamic/fake"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

const (
	globalPackagingNamespace = "kubeapps-repos-global"
	kubeappsNamespace        = "kubeapps"
	globalPackagingCluster   = "default"
	DefaultAppVersion        = "1.2.6"
	DefaultReleaseRevision   = 1
	DefaultChartDescription  = "default chart description"
	DefaultChartIconURL      = "https://example.com/chart.svg"
	DefaultChartHomeURL      = "https://helm.sh/helm"
	DefaultChartCategory     = "cat1"
)

func setMockManager(t *testing.T) (sqlmock.Sqlmock, func(), utils.AssetManager) {
	var manager utils.AssetManager
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("%+v", err)
	}
	manager = &utils.PostgresAssetManager{PostgresAssetManagerIface: &dbutils.PostgresAssetManager{DB: db, GlobalPackagingNamespace: globalPackagingNamespace}}
	return mock, func() { db.Close() }, manager
}

func TestGetClient(t *testing.T) {
	dbConfig := dbutils.Config{URL: "localhost:5432", Database: "assets", Username: "postgres", Password: "password"}
	manager, err := utils.NewPGManager(dbConfig, globalPackagingNamespace)
	if err != nil {
		log.Fatalf("%s", err)
	}

	clientGetter := clientgetter.NewFixedClientProvider(&clientgetter.ClientGetter{
		Typed: func() (kubernetes.Interface, error) { return typfake.NewSimpleClientset(), nil },
		Dynamic: func() (dynamic.Interface, error) {
			return dynfake.NewSimpleDynamicClientWithCustomListKinds(
				k8sruntime.NewScheme(),
				map[schema.GroupVersionResource]string{
					{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
				},
			), nil
		},
	})

	testCases := []struct {
		name              string
		manager           utils.AssetManager
		clientGetter      clientgetter.ClientProviderInterface
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
			name:    "it returns failed-precondition when clients getter function itself errors",
			manager: manager,
			clientGetter: &clientgetter.ClientProvider{ClientsFunc: func(ctx context.Context, cluster string) (*clientgetter.ClientGetter, error) {
				return nil, fmt.Errorf("Bang!")
			}},
			statusCodeClient:  codes.FailedPrecondition,
			statusCodeManager: codes.OK,
		},
		{
			name:              "it returns failed-precondition when clients getter function is not set",
			manager:           manager,
			clientGetter:      &clientgetter.ClientProvider{ClientsFunc: nil},
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

			typedClient, dynamicClient, errClient := s.GetClients(context.Background(), "")

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

// makeChart makes a chart with specific input used in the test and default constants for other relevant data.
func makeChart(chart_name, repo_name, repo_url, namespace string, chart_versions []string, category string) *models.Chart {
	ch := &models.Chart{
		Name:        chart_name,
		ID:          fmt.Sprintf("%s/%s", repo_name, chart_name),
		Category:    category,
		Description: DefaultChartDescription,
		Home:        DefaultChartHomeURL,
		Icon:        DefaultChartIconURL,
		Maintainers: []chart.Maintainer{{Name: "me", Email: "me@me.me"}},
		Sources:     []string{"http://source-1"},
		Repo: &models.Repo{
			Name:      repo_name,
			Namespace: namespace,
			URL:       repo_url,
		},
	}
	versions := []models.ChartVersion{}
	for _, v := range chart_versions {
		versions = append(versions, models.ChartVersion{
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
		itemOffset, err := paginate.ItemOffsetFromPageToken(pageToken)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		if pageSize == 0 {
			t.Fatalf("pagesize must be > 0 when using a page token")
		}
		rowsJSON = rowsJSON[itemOffset:]
	}
	if pageSize > 0 && pageSize < len(rowsJSON) {
		rowsJSON = rowsJSON[0:pageSize]
	}
	return rowsJSON
}

// makeServer returns a server backed with an sql mock and a cleanup function
func makeServer(t *testing.T, authorized bool, actionConfig *action.Configuration, objects ...k8sruntime.Object) (*Server, sqlmock.Sqlmock, func()) {
	// Creating the dynamic client
	scheme := k8sruntime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("%+v", err)
	}
	dynamicClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			{Group: "foo", Version: "bar", Resource: "baz"}: "PackageList",
		},
		objects...,
	)

	// Creating an authorized clientGetter
	clientSet := typfake.NewSimpleClientset()
	clientSet.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: authorized},
		}, nil
	})

	clientGetter := clientgetter.NewFixedClientProvider(&clientgetter.ClientGetter{
		Typed:   func() (kubernetes.Interface, error) { return clientSet, nil },
		Dynamic: func() (dynamic.Interface, error) { return dynamicClient, nil },
	})

	// Creating the SQL mock manager
	mock, cleanup, manager := setMockManager(t)

	return &Server{
		clientGetter:             clientGetter,
		manager:                  manager,
		kubeappsNamespace:        kubeappsNamespace,
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		actionConfigGetter: func(context.Context, *corev1.Context) (*action.Configuration, error) {
			return actionConfig, nil
		},
		chartClientFactory: &fake.ChartClientFactory{},
		pluginConfig:       common.NewDefaultPluginConfig(),
		createReleaseFunc:  agent.CreateRelease,
	}, mock, cleanup
}

func newServerWithSecretsAndRepos(t *testing.T, secrets []k8sruntime.Object, unstructuredObjs []k8sruntime.Object, repos []*v1alpha1.AppRepository) *Server {
	typedClient := typfake.NewSimpleClientset(secrets...)

	// ref https://stackoverflow.com/questions/68794562/kubernetes-fake-client-doesnt-handle-generatename-in-objectmeta/68794563#68794563
	typedClient.PrependReactor(
		"create", "*",
		func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
			ret = action.(k8stesting.CreateAction).GetObject()
			meta, ok := ret.(metav1.Object)
			if !ok {
				return
			}
			if meta.GetName() == "" && meta.GetGenerateName() != "" {
				meta.SetName(names.SimpleNameGenerator.GenerateName(meta.GetGenerateName()))
			}
			return
		})

	// Creating an authorized clientGetter
	typedClient.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret k8sruntime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})

	apiExtIfc := apiextfake.NewSimpleClientset(helmAppRepositoryCRD)
	ctrlClient := newCtrlClient(repos)
	scheme := k8sruntime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("%s", err)
	}

	dynClient := dynfake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			{
				Group:    v1alpha1.SchemeGroupVersion.Group,
				Version:  v1alpha1.SchemeGroupVersion.Version,
				Resource: AppRepositoryResource,
			}: AppRepositoryResource + "List",
		},
		unstructuredObjs...,
	)

	clientGetter := clientgetter.NewFixedClientProvider(&clientgetter.ClientGetter{
		Typed:             func() (kubernetes.Interface, error) { return typedClient, nil },
		Dynamic:           func() (dynamic.Interface, error) { return dynClient, nil },
		ControllerRuntime: func() (client.WithWatch, error) { return ctrlClient, nil },
		ApiExt:            func() (apiext.Interface, error) { return apiExtIfc, nil },
	})

	return &Server{
		clientGetter:             clientGetter,
		kubeappsNamespace:        kubeappsNamespace,
		globalPackagingNamespace: globalPackagingNamespace,
		globalPackagingCluster:   globalPackagingCluster,
		chartClientFactory:       &fake.ChartClientFactory{},
		createReleaseFunc:        agent.CreateRelease,
		kubeappsCluster:          KubeappsCluster,
		pluginConfig:             common.NewDefaultPluginConfig(),
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name                   string
		charts                 []*models.Chart
		expectDBQueryNamespace string
		statusCode             codes.Code
		request                *corev1.GetAvailablePackageSummariesRequest
		expectedResponse       *corev1.GetAvailablePackageSummariesResponse
		authorized             bool
		expectedCategories     []*models.ChartCategory
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
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3-global", "repo-1", "http://chart-3", globalPackagingNamespace, []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-3-global",
						DisplayName: "chart-3-global",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: globalPackagingNamespace},
							Identifier: "repo-1/chart-3-global",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns a set of availablePackageSummary from the database (specific ns)",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: "my-ns",
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns a set of the global availablePackageSummary from the database (not the specific ns on other cluster)",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "other",
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"cat1"},
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
			charts:     []*models.Chart{},
			statusCode: codes.Unimplemented,
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
			expectDBQueryNamespace: globalPackagingNamespace,
			charts:                 []*models.Chart{makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{}, DefaultChartCategory)},
			statusCode:             codes.Internal,
		},
		{
			name:       "it returns a permissionDenied status if the user doesn't have permissions",
			authorized: false,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Namespace: "my-ns",
				},
			},
			charts:     []*models.Chart{{Name: "foo"}},
			statusCode: codes.PermissionDenied,
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
					PageToken: "1",
					PageSize:  1,
				},
			},
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						ShortDescription: DefaultChartDescription,
						Categories:       []string{DefaultChartCategory},
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				NextPageToken: "2",
				Categories:    []string{"cat1"},
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
			expectDBQueryNamespace: globalPackagingNamespace,
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, DefaultChartCategory),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, DefaultChartCategory),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-3",
						DisplayName: "chart-3",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{DefaultChartCategory},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-3",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
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
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "it returns the proper chart categories",
			authorized: true,
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: "my-ns",
				},
			},
			expectDBQueryNamespace: "my-ns",
			charts: []*models.Chart{
				makeChart("chart-1", "repo-1", "http://chart-1", "my-ns", []string{"3.0.0"}, "foo"),
				makeChart("chart-2", "repo-1", "http://chart-2", "my-ns", []string{"2.0.0"}, "bar"),
				makeChart("chart-3", "repo-1", "http://chart-3", "my-ns", []string{"1.0.0"}, "bar"),
			},
			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					{
						Name:        "chart-1",
						DisplayName: "chart-1",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"foo"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-1",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-2",
						DisplayName: "chart-2",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "2.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"bar"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-2",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
					{
						Name:        "chart-3",
						DisplayName: "chart-3",
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.0.0",
							AppVersion: DefaultAppVersion,
						},
						IconUrl:          DefaultChartIconURL,
						Categories:       []string{"bar"},
						ShortDescription: DefaultChartDescription,
						AvailablePackageRef: &corev1.AvailablePackageReference{
							Context:    &corev1.Context{Cluster: globalPackagingCluster, Namespace: "my-ns"},
							Identifier: "repo-1/chart-3",
							Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
						},
					},
				},
				Categories: []string{"bar", "foo"},
			},
			statusCode: codes.OK,
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
			}

			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageSummariesResponse{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.PackageAppVersion{})
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
		chartFiles *models.ChartFiles
		expected   *corev1.AvailablePackageDetail
		statusCode codes.Code
	}{
		{
			name:  "it returns AvailablePackageDetail if the chart is correct",
			chart: makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0"}, DefaultChartCategory),
			chartFiles: &models.ChartFiles{
				Readme: "chart readme",
				Values: "chart values",
				Schema: "chart schema",
			},
			expected: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				RepoUrl:          "http://foo",
				HomeUrl:          DefaultChartHomeURL,
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
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
			availablePackageDetail, err := AvailablePackageDetailFromChart(tc.chart, tc.chartFiles)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
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
		charts          []*models.Chart
		expectedPackage *corev1.AvailablePackageDetail
		statusCode      codes.Code
		request         *corev1.GetAvailablePackageDetailRequest
		authorized      bool
	}{
		{
			name:       "it returns an availablePackageDetail from the database (latest version)",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "repo-1%2Ffoo",
				},
			},
			charts: []*models.Chart{makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0"}, DefaultChartCategory)},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				HomeUrl:          DefaultChartHomeURL,
				RepoUrl:          "http://foo",
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				Version: &corev1.PackageAppVersion{
					PkgVersion: "3.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
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
				PkgVersion: "1.0.0",
			},
			charts: []*models.Chart{makeChart("foo", "repo-1", "http://foo", "my-ns", []string{"3.0.0", "2.0.0", "1.0.0"}, DefaultChartCategory)},
			expectedPackage: &corev1.AvailablePackageDetail{
				Name:             "foo",
				DisplayName:      "foo",
				HomeUrl:          DefaultChartHomeURL,
				RepoUrl:          "http://foo",
				IconUrl:          DefaultChartIconURL,
				Categories:       []string{DefaultChartCategory},
				ShortDescription: DefaultChartDescription,
				LongDescription:  "",
				Version: &corev1.PackageAppVersion{
					PkgVersion: "1.0.0",
					AppVersion: DefaultAppVersion,
				},
				Readme:        "chart readme",
				DefaultValues: "chart values",
				ValuesSchema:  "chart schema",
				SourceUrls:    []string{"http://source-1"},
				Maintainers:   []*corev1.Maintainer{{Name: "me", Email: "me@me.me"}},
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "repo-1/foo",
					Plugin:     &plugins.Plugin{Name: "helm.packages", Version: "v1alpha1"},
				},
			},
			statusCode: codes.OK,
		},
		{
			name:       "it returns an invalid arg error status if no context is provided",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "foo/bar",
				},
			},
			charts:     []*models.Chart{{Name: "foo"}},
			statusCode: codes.InvalidArgument,
		},
		{
			name:       "it returns an invalid arg error status if cluster is not the global/kubeapps one",
			authorized: true,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Cluster: "other-cluster", Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:     []*models.Chart{{Name: "foo"}},
			statusCode: codes.InvalidArgument,
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
				PkgVersion: "9.9.9",
			},
			charts:          []*models.Chart{{Name: "foo"}},
			expectedPackage: &corev1.AvailablePackageDetail{},
			statusCode:      codes.Internal,
		},
		{
			name:       "it returns a permissionDenied status if the user doesn't have permissions",
			authorized: false,
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Namespace: "my-ns"},
					Identifier: "foo/bar",
				},
			},
			charts:          []*models.Chart{{Name: "foo"}},
			expectedPackage: &corev1.AvailablePackageDetail{},
			statusCode:      codes.PermissionDenied,
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
			if tc.statusCode == codes.OK {
				// Checking if the WHERE condition is properly applied
				chartIDUnescaped, err := url.QueryUnescape(tc.request.AvailablePackageRef.Identifier)
				if err != nil {
					t.Fatalf("%+v", err)
				}
				mock.ExpectQuery("SELECT info FROM charts").
					WithArgs(tc.request.AvailablePackageRef.Context.Namespace, chartIDUnescaped).
					WillReturnRows(rows)
				fileID := fileIDForChart(chartIDUnescaped, tc.expectedPackage.Version.PkgVersion)
				fileJSON, err := json.Marshal(models.ChartFiles{
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

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				opt1 := cmpopts.IgnoreUnexported(corev1.AvailablePackageDetail{}, corev1.AvailablePackageSummary{}, corev1.AvailablePackageReference{}, corev1.Context{}, plugins.Plugin{}, corev1.Maintainer{}, corev1.PackageAppVersion{})
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
			name: "it returns invalid argument if called with a cluster other than the global/kubeapps one",
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context:    &corev1.Context{Cluster: "other-cluster", Namespace: "kubeapps"},
					Identifier: "bitnami/apache",
				},
			},
			expectedStatusCode: codes.InvalidArgument,
		},
		{
			name:   "it returns the package version summary",
			charts: []*models.Chart{makeChart("apache", "bitnami", "http://apache", "kubeapps", []string{"3.0.0", "2.0.0", "1.0.0"}, DefaultChartCategory)},
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
				PackageAppVersions: []*corev1.PackageAppVersion{
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

			opts := cmpopts.IgnoreUnexported(corev1.GetAvailablePackageVersionsResponse{}, corev1.PackageAppVersion{})
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
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 0, Minor: 0, Patch: 0},
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
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 4, Minor: 2, Patch: 1},
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
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 1, Minor: 0, Patch: 0},
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
			// TODO(agamez): env vars and file paths should be handled properly for Windows operating system
			if runtime.GOOS == "windows" {
				t.Skip("Skipping in a Windows OS")
			}
			filename := ""
			if tc.pluginYAMLConf != nil {
				pluginJSONConf, err := yaml.YAMLToJSON(tc.pluginYAMLConf)
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
			pluginConfig, err := common.ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.exp_error_str)
			} else if pluginConfig != nil {
				if got, want := pluginConfig.VersionsInSummary, tc.exp_versions_in_summary; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
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
				pluginJSONConf, err := yaml.YAMLToJSON(tc.pluginYAMLConf)
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
			pluginConfig, err := common.ParsePluginConfig(filename)
			if err != nil && !strings.Contains(err.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", err.Error(), tc.exp_error_str)
			} else if pluginConfig != nil {
				if got, want := pluginConfig.TimeoutSeconds, tc.exp_timeout; !cmp.Equal(want, got, opts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
				}
			}
		})
	}
}
func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name               string
		request            *corev1.GetInstalledPackageSummariesRequest
		existingReleases   []releaseStub
		expectedStatusCode codes.Code
		expectedResponse   *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "returns installed packages in a specific namespace",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       release.StatusDeployed,
					version:      2,
				},
				{
					name:      "my-release-2",
					namespace: "other-namespace",
					status:    release.StatusDeployed,
					version:   4,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-1",
					chartVersion: "4.5.6",
					status:       release.StatusDeployed,
					version:      6,
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-3",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
			},
		},
		{
			name: "returns installed packages across all namespaces",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       release.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       release.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       release.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-2",
							},
							Identifier: "my-release-2",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-2",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "3.4.5",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "3.4.5",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.4.5",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-3",
							},
							Identifier: "my-release-3",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
			},
		},
		{
			name: "returns limited results",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
				PaginationOptions: &corev1.PaginationOptions{
					PageSize: 2,
				},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       release.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       release.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       release.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.2.3",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-2",
							},
							Identifier: "my-release-2",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-2",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "3.4.5",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "3.4.5",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "3.4.5",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
				NextPageToken: "2",
			},
		},
		{
			name: "fetches results from an offset",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: ""},
				PaginationOptions: &corev1.PaginationOptions{
					PageSize:  2,
					PageToken: "2",
				},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       release.StatusDeployed,
					version:      1,
				},
				{
					name:         "my-release-2",
					namespace:    "namespace-2",
					status:       release.StatusDeployed,
					chartVersion: "3.4.5",
					version:      1,
				},
				{
					name:         "my-release-3",
					namespace:    "namespace-3",
					chartVersion: "4.5.6",
					status:       release.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-3",
							},
							Identifier: "my-release-3",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-3",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "4.5.6",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "4.5.6",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "4.5.6",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
							UserReason: "deployed",
						},
					},
				},
				NextPageToken: "",
			},
		},
		{
			name: "includes a latest package version when available",
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{Namespace: "namespace-1"},
			},
			existingReleases: []releaseStub{
				{
					name:         "my-release-1",
					namespace:    "namespace-1",
					chartVersion: "1.2.3",
					status:       release.StatusDeployed,
					version:      1,
				},
			},
			expectedStatusCode: codes.OK,
			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					{
						InstalledPackageRef: &corev1.InstalledPackageReference{
							Context: &corev1.Context{
								Cluster:   globalPackagingCluster,
								Namespace: "namespace-1",
							},
							Identifier: "my-release-1",
							Plugin:     GetPluginDetail(),
						},
						Name:    "my-release-1",
						IconUrl: "https://example.com/icon.png",
						PkgVersionReference: &corev1.VersionReference{
							Version: "1.2.3",
						},
						CurrentVersion: &corev1.PackageAppVersion{

							PkgVersion: "1.2.3",
							AppVersion: DefaultAppVersion,
						},
						LatestVersion: &corev1.PackageAppVersion{
							PkgVersion: "1.2.5",
						},
						Status: &corev1.InstalledPackageStatus{
							Ready:      true,
							Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
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

			if tc.expectedStatusCode == codes.OK {
				populateAssetDBWithSummaries(t, mock, tc.expectedResponse.InstalledPackageSummaries)
			}

			response, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetInstalledPackageSummariesResponse{}, corev1.InstalledPackageSummary{}, corev1.InstalledPackageReference{}, corev1.Context{}, corev1.VersionReference{}, corev1.InstalledPackageStatus{}, corev1.PackageAppVersion{}, plugins.Plugin{})
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
	customDetailRevision2, err := anypb.New(&helmv1.InstalledPackageDetailCustomDataHelm{
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
		request            *corev1.GetInstalledPackageDetailRequest
		expectedResponse   *corev1.GetInstalledPackageDetailResponse
		expectedStatusCode codes.Code
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
					status:         release.StatusSuperseded,
					version:        1,
				},
				{
					name:           releaseName,
					namespace:      releaseNamespace,
					chartVersion:   releaseVersion,
					chartNamespace: releaseNamespace,
					values:         releaseValues,
					notes:          releaseNotes,
					status:         release.StatusDeployed,
					version:        2,
				},
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: releaseNamespace,
						Cluster:   globalPackagingCluster,
					},
					Identifier: releaseName,
				},
			},
			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: &corev1.InstalledPackageDetail{
					InstalledPackageRef: &corev1.InstalledPackageReference{
						Context: &corev1.Context{
							Namespace: releaseNamespace,
							Cluster:   globalPackagingCluster,
						},
						Identifier: releaseName,
					},
					PkgVersionReference: &corev1.VersionReference{
						Version: releaseVersion,
					},
					Name: releaseName,
					CurrentVersion: &corev1.PackageAppVersion{
						PkgVersion: releaseVersion,
						AppVersion: DefaultAppVersion,
					},
					LatestVersion: &corev1.PackageAppVersion{
						PkgVersion: releaseVersion,
						AppVersion: DefaultAppVersion,
					},
					ValuesApplied:         releaseValues,
					PostInstallationNotes: releaseNotes,
					Status: &corev1.InstalledPackageStatus{
						Ready:      true,
						Reason:     corev1.InstalledPackageStatus_STATUS_REASON_INSTALLED,
						UserReason: "deployed",
					},
					AvailablePackageRef: &corev1.AvailablePackageReference{
						Context: &corev1.Context{
							Namespace: releaseNamespace,
							Cluster:   globalPackagingCluster,
						},
						Identifier: "myrepo/" + releaseName,
						Plugin:     GetPluginDetail(),
					},
					CustomDetail: customDetailRevision2,
				},
			},
			expectedStatusCode: codes.OK,
		},
		{
			name: "returns a 404 if the installed package is not found",
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Namespace: releaseNamespace,
					},
					Identifier: releaseName,
				},
			},
			expectedStatusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			authorized := true
			actionConfig := newActionConfigFixture(t, tc.request.GetInstalledPackageRef().GetContext().GetNamespace(), tc.existingReleases, nil)
			server, mock, cleanup := makeServer(t, authorized, actionConfig)
			defer cleanup()

			if tc.expectedStatusCode == codes.OK {
				populateAssetDBWithDetail(t, mock, tc.expectedResponse.InstalledPackageDetail)
			}

			response, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.expectedStatusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			// We don't need to check anything else for non-OK codes.
			if tc.expectedStatusCode != codes.OK {
				return
			}

			opts := cmpopts.IgnoreUnexported(corev1.GetInstalledPackageDetailResponse{}, corev1.InstalledPackageDetail{}, corev1.InstalledPackageReference{}, corev1.Context{}, corev1.VersionReference{}, corev1.InstalledPackageStatus{}, corev1.AvailablePackageReference{}, plugins.Plugin{}, corev1.PackageAppVersion{}, anypb.Any{})
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
		repo         *models.Repo
		chartVersion *models.ChartVersion
		expectedUrl  string
	}{
		{
			name:         "tarball url with relative URL without leading slash in chart",
			repo:         &models.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &models.ChartVersion{URLs: []string{"chart/test"}},
			expectedUrl:  "https://demo.repo/repo1/chart/test",
		},
		{
			name:         "tarball url with relative URL with leading slash in chart",
			repo:         &models.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &models.ChartVersion{URLs: []string{"/chart/test"}},
			expectedUrl:  "https://demo.repo/repo1/chart/test",
		},
		{
			name:         "tarball url with absolute URL",
			repo:         &models.Repo{URL: "https://demo.repo/repo1"},
			chartVersion: &models.ChartVersion{URLs: []string{"https://demo.repo/repo1/chart/test"}},
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

// newActionConfigFixture returns an action.Configuration with fake clients
// and memory storage.
func newActionConfigFixture(t *testing.T, namespace string, rels []releaseStub, kubeClient kube.Interface) *action.Configuration {
	t.Helper()

	memDriver := driver.NewMemory()

	if kubeClient == nil {
		kubeClient = &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}}
	}

	actionConfig := &action.Configuration{
		// Create the Releases storage explicitly so we can set the
		// internal log function used to see data in test output.
		Releases: &storage.Storage{
			Driver: memDriver,
			Log: func(format string, v ...interface{}) {
				t.Logf(format, v...)
			},
		},
		KubeClient:   kubeClient,
		Capabilities: chartutil.DefaultCapabilities,
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

func releaseForStub(t *testing.T, r releaseStub) *release.Release {
	config := map[string]interface{}{}
	if r.values != "" {
		err := json.Unmarshal([]byte(r.values), &config)
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
	return &release.Release{
		Name:      r.name,
		Namespace: r.namespace,
		Manifest:  r.manifest,
		Version:   r.version,
		Info: &release.Info{
			Status: r.status,
			Notes:  r.notes,
		},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Version:    r.chartVersion,
				Icon:       "https://example.com/icon.png",
				AppVersion: DefaultAppVersion,
			},
		},
		Config: config,
	}
}

func chartAssetForReleaseStub(rel *releaseStub) *models.Chart {
	chartVersions := []models.ChartVersion{}
	if rel.latestVersion != "" {
		chartVersions = append(chartVersions, models.ChartVersion{
			Version: rel.latestVersion,
			URLs:    []string{fmt.Sprintf("https://example.com/%s-%s.tgz", rel.chartID, rel.latestVersion)},
		})
	}
	chartVersions = append(chartVersions, models.ChartVersion{
		Version:    rel.chartVersion,
		AppVersion: DefaultAppVersion,
	})

	return &models.Chart{
		Name: rel.name,
		ID:   rel.chartID,
		Repo: &models.Repo{
			Namespace: rel.chartNamespace,
		},
		ChartVersions: chartVersions,
	}
}

func populateAssetDBWithSummaries(t *testing.T, mock sqlmock.Sqlmock, pkgs []*corev1.InstalledPackageSummary) {
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

func populateAssetDBWithDetail(t *testing.T, mock sqlmock.Sqlmock, pkg *corev1.InstalledPackageDetail) {
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
	chart := &models.Chart{
		Name: chartId,
		ID:   chartId,
		Repo: &models.Repo{
			Namespace: globalPackagingNamespace,
		},
		ChartVersions: []models.ChartVersion{{
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
	status         release.Status
	manifest       string
}
