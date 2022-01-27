// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	chartutilsfake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	kubeutils "github.com/kubeapps/kubeapps/pkg/kube"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	helmchartutil "helm.sh/helm/v3/pkg/chartutil"
	helmkubefake "helm.sh/helm/v3/pkg/kube/fake"
	helmrelease "helm.sh/helm/v3/pkg/release"
	helmstorage "helm.sh/helm/v3/pkg/storage"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	helmtime "helm.sh/helm/v3/pkg/time"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultListLimit = 256
)

var (
	testingTime, _    = helmtime.Parse(time.RFC3339, "1977-09-02T22:04:05Z")
	lastDeployedRegex = regexp.MustCompile(`"last_deployed":\s*"[^"]+?[^\/"]+"`)
)

// newConfigFixture returns a Config with fake clients
// and memory helmstorage.
func newConfigFixture(t *testing.T, k *helmkubefake.FailingKubeClient) *Config {
	t.Helper()

	return &Config{
		ActionConfig: &helmaction.Configuration{
			Releases:     helmstorage.Init(helmstoragedriver.NewMemory()),
			KubeClient:   k,
			Capabilities: helmchartutil.DefaultCapabilities,
			Log: func(format string, v ...interface{}) {
				t.Helper()
				t.Logf(format, v...)
			},
		},
		KubeHandler: &kubeutils.FakeHandler{
			AppRepos: []*apprepov1alpha1.AppRepository{
				{ObjectMeta: k8smetav1.ObjectMeta{Name: "bitnami", Namespace: "default"},
					Spec: apprepov1alpha1.AppRepositorySpec{Type: "helm", URL: "http://foo.bar"},
				},
			},
		},
		ChartClientFactory: &chartutilsfake.ChartClientFactory{},
		Options: Options{
			ListLimit: defaultListLimit,
		},
	}
}

// See https://github.com/kubeapps/kubeapps/pull/1439/files#r365678777
// for discussion about cleaner long booleans.
func and(exps ...bool) bool {
	for _, exp := range exps {
		if !exp {
			return false
		}
	}
	return true
}

var releaseComparer = cmp.Comparer(func(x helmrelease.Release, y helmrelease.Release) bool {
	return and(
		x.Name == y.Name,
		x.Version == y.Version,
		x.Namespace == y.Namespace,
		x.Info.Status == y.Info.Status,
		x.Chart.Name() == y.Chart.Name(),
		x.Manifest == y.Manifest,
		cmp.Equal(x.Config, y.Config),
		cmp.Equal(x.Hooks, y.Hooks),
	)
})

func TestActions(t *testing.T) {
	type testScenario struct {
		// Scenario params
		Description      string
		ExistingReleases []*helmrelease.Release
		KubeError        error
		// Request params
		RequestBody  string
		RequestQuery string
		Action       string
		Params       map[string]string
		// Expected result
		StatusCode        int
		RemainingReleases []*helmrelease.Release
		ResponseBody      string //optional
	}

	tests := []testScenario{
		{
			// Scenario params
			Description:      "Create a simple release",
			ExistingReleases: []*helmrelease.Release{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0", "appRepositoryResourceName": "bitnami", "appRepositoryResourceNamespace": "default"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description: "Create a conflicting release",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0", "appRepositoryResourceName": "bitnami", "appRepositoryResourceNamespace": "default"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 409,
			RemainingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Get a non-existing release",
			ExistingReleases: []*helmrelease.Release{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: nil,
			ResponseBody:      "",
		},
		{
			Description: "Delete a simple release",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foobarchart", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*helmrelease.Release{
				createRelease("foobarchart", "foobar", "default", 1, helmrelease.StatusUninstalled),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description: "Delete and purge a simple release with purge=true",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foobarchart", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "?purge=true",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: nil,
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description: "Get a simple release",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			ResponseBody: `{"data":{"name":"foobar","info":{"first_deployed":"1977-09-02T22:04:05Z","last_deployed":"1977-09-02T22:04:05Z","deleted":"","status":"deployed"},"chart":{"metadata":{"name":"foo"},"lock":null,"templates":null,"values":{},"schema":null,"files":null},"version":1,"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description: "Get a deleted release",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusUninstalled),
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*helmrelease.Release{
				createRelease("foo", "foobar", "default", 1, helmrelease.StatusUninstalled),
			},
			ResponseBody: `{"data":{"name":"foobar","info":{"first_deployed":"1977-09-02T22:04:05Z","last_deployed":"1977-09-02T22:04:05Z","deleted":"1977-09-02T22:04:05Z","status":"uninstalled"},"chart":{"metadata":{"name":"foo"},"lock":null,"templates":null,"values":{},"schema":null,"files":null},"version":1,"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description: "Delete and purge a simple release with purge=1",
			ExistingReleases: []*helmrelease.Release{
				createRelease("foobarchart", "foobar", "default", 1, helmrelease.StatusDeployed),
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "?purge=1",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: nil,
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Delete a missing release",
			ExistingReleases: []*helmrelease.Release{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: nil,
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Creates a release with missing permissions",
			ExistingReleases: []*helmrelease.Release{},
			KubeError:        errors.New(`Failed to create: secrets is forbidden: User "foo" cannot create resource "secrets" in API group "" in the namespace "default"`),
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0", "appRepositoryResourceName": "bitnami", "appRepositoryResourceNamespace": "default"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode:        403,
			RemainingReleases: nil,
			ResponseBody:      `{"code":403,"message":"[{\"apiGroup\":\"\",\"resource\":\"secrets\",\"namespace\":\"default\",\"clusterWide\":false,\"verbs\":[\"create\"]}]"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			// Initialize environment for test
			req := httptest.NewRequest("GET", fmt.Sprintf("http://foo.bar%s", test.RequestQuery), strings.NewReader(test.RequestBody))
			response := httptest.NewRecorder()
			k := &helmkubefake.FailingKubeClient{PrintingKubeClient: helmkubefake.PrintingKubeClient{Out: ioutil.Discard}}
			if test.KubeError != nil {
				// The helm fake Kube Client runs build before
				// create/install/upgrade. It also stores a release in storage
				// even if there were no resources to create so we need to error
				// before the release is saved.
				k.BuildError = test.KubeError
			}
			cfg := newConfigFixture(t, k)
			createExistingReleases(t, cfg, test.ExistingReleases)

			// Perform request
			switch test.Action {
			case "get":
				GetRelease(*cfg, response, req, test.Params)
			case "create":
				CreateRelease(*cfg, response, req, test.Params)
			case "delete":
				DeleteRelease(*cfg, response, req, test.Params)
			default:
				t.Errorf("Unexpected action %s", test.Action)
			}
			// Check result
			if response.Code != test.StatusCode {
				t.Errorf("Expecting a StatusCode %d, received %d", test.StatusCode, response.Code)
			}
			releases, err := cfg.ActionConfig.Releases.ListReleases()
			if err != nil {
				t.Fatalf("%+v", err)
			}
			// The Helm memory driver does not appear to have consistent ordering.
			// See https://github.com/helm/helm/issues/7263
			// Just sort by "name.version.namespace" which is good enough here.
			sort.Slice(releases, func(i, j int) bool {
				iKey := fmt.Sprintf("%s.%d.%s", releases[i].Name, releases[i].Version, releases[i].Namespace)
				jKey := fmt.Sprintf("%s.%d.%s", releases[j].Name, releases[j].Version, releases[j].Namespace)
				return iKey < jKey
			})
			if !cmp.Equal(releases, test.RemainingReleases, releaseComparer) {
				t.Errorf("Unexpected remaining releases. Diff:\n%s", cmp.Diff(test.RemainingReleases, releases, releaseComparer))
			}
			if test.ResponseBody != "" {
				if test.ResponseBody != response.Body.String() {
					t.Errorf("Unexpected body response. Diff %s", cmp.Diff(test.ResponseBody, response.Body))
				}
			}
		})
	}
}

func createRelease(chartName, name, namespace string, version int, status helmrelease.Status) *helmrelease.Release {
	deleted := helmtime.Time{}
	if status == helmrelease.StatusUninstalled {
		deleted = testingTime
	}
	return &helmrelease.Release{
		Name:      name,
		Namespace: namespace,
		Version:   version,
		Info:      &helmrelease.Info{Status: status, Deleted: deleted, FirstDeployed: testingTime, LastDeployed: testingTime},
		Chart: &helmchart.Chart{
			Metadata: &helmchart.Metadata{
				Name: chartName,
			},
			Values: make(map[string]interface{}),
		},
		Config: make(map[string]interface{}),
	}
}

func createExistingReleases(t *testing.T, cfg *Config, releases []*helmrelease.Release) {
	for i := range releases {
		err := cfg.ActionConfig.Releases.Create(releases[i])
		if err != nil {
			t.Fatalf("%+v", err)
		}
	}
}

func TestRollbackAction(t *testing.T) {
	const releaseName = "my-release"
	testCases := []struct {
		name             string
		existingReleases []*helmrelease.Release
		queryString      string
		params           map[string]string
		statusCode       int
		expectedReleases []*helmrelease.Release
		responseBody     string
	}{
		{
			name: "rolls back a release",
			existingReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			queryString: "action=rollback&revision=1",
			params:      map[string]string{nameParam: "my-release"},
			statusCode:  http.StatusOK,
			expectedReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 3, helmrelease.StatusDeployed),
			},
			responseBody: `{"data":{"name":"my-release","info":{"first_deployed":"1977-09-02T22:04:05Z","last_deployed":"1977-09-02T22:04:05Z","deleted":"","description":"Rollback to 1","status":"deployed"},"chart":{"metadata":{"name":"apache"},"lock":null,"templates":null,"values":{},"schema":null,"files":null},"version":3,"namespace":"default"}}`,
		},
		{
			name: "errors if the release does not exist",
			existingReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			queryString: "action=rollback&revision=1",
			params:      map[string]string{nameParam: "does-not-exist"},
			statusCode:  http.StatusNotFound,
			expectedReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			responseBody: `{"code":404,"message":"release: not found"}`,
		},
		{
			name: "errors if the revision does not exist",
			existingReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			queryString: "action=rollback&revision=3",
			params:      map[string]string{nameParam: "apache"},
			statusCode:  http.StatusNotFound,
			expectedReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			responseBody: `{"code":404,"message":"release: not found"}`,
		},
		{
			name: "errors if the revision is not specified",
			existingReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			queryString: "action=rollback",
			params:      map[string]string{nameParam: "apache"},
			statusCode:  http.StatusUnprocessableEntity,
			expectedReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			responseBody: `{"code":422,"message":"Missing revision to rollback in request"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k := &helmkubefake.FailingKubeClient{PrintingKubeClient: helmkubefake.PrintingKubeClient{Out: ioutil.Discard}}
			cfg := newConfigFixture(t, k)
			createExistingReleases(t, cfg, tc.existingReleases)
			req := httptest.NewRequest("PUT", fmt.Sprintf("https://example.com/whatever?%s", tc.queryString), strings.NewReader(""))
			response := httptest.NewRecorder()

			OperateRelease(*cfg, response, req, tc.params)

			if got, want := response.Code, tc.statusCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// using a fixed date to avoid time-dependant tests
			// ideally, we should mutate the date as we are doing for deleted releases
			fixedDateResponse := lastDeployedRegex.ReplaceAllString(response.Body.String(), `"last_deployed":"1977-09-02T22:04:05Z"`)
			if got, want := fixedDateResponse, tc.responseBody; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}

			actualReleases, err := cfg.ActionConfig.Releases.ListReleases()
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := actualReleases, tc.expectedReleases; !cmp.Equal(want, got, releaseComparer) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, releaseComparer))
			}
		})
	}
}

func TestUpgradeAction(t *testing.T) {
	const releaseName = "my-release"
	testCases := []struct {
		name             string
		existingReleases []*helmrelease.Release
		queryString      string
		requestBody      string
		params           map[string]string
		statusCode       int
		expectedReleases []*helmrelease.Release
		responseBody     string
	}{
		{
			name: "upgrade a release",
			existingReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusDeployed),
			},
			queryString: "action=upgrade",
			requestBody: `{"chartName": "apache",	"releaseName":"my-release",	"version": "1.0.0", "appRepositoryResourceName": "bitnami", "appRepositoryResourceNamespace": "default"}`,
			params:     map[string]string{nameParam: releaseName},
			statusCode: http.StatusOK,
			expectedReleases: []*helmrelease.Release{
				createRelease("apache", releaseName, "default", 1, helmrelease.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, helmrelease.StatusDeployed),
			},
			responseBody: `{"data":{"name":"my-release","info":{"first_deployed":"1977-09-02T22:04:05Z","last_deployed":"1977-09-02T22:04:05Z","deleted":"","description":"Upgrade complete","status":"deployed"},"chart":{"metadata":{"name":"apache","version":"1.0.0"},"lock":null,"templates":null,"values":{},"schema":null,"files":null},"version":2,"namespace":"default"}}`,
		},
		{
			name:             "upgrade a missing release",
			existingReleases: []*helmrelease.Release{},
			queryString:      "action=upgrade",
			requestBody: `{"chartName": "apache",	"releaseName":"my-release",	"version": "1.0.0", "appRepositoryResourceName": "bitnami", "appRepositoryResourceNamespace": "default"}`,
			params:     map[string]string{nameParam: releaseName},
			statusCode: http.StatusNotFound,
			// expectedReleases is `nil` because nil slice != empty slice
			// sotrage.ListReleases() returns a nil slice if no releases are found
			expectedReleases: nil,
			responseBody:     `{"code":404,"message":"release: not found"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k := &helmkubefake.FailingKubeClient{PrintingKubeClient: helmkubefake.PrintingKubeClient{Out: ioutil.Discard}}
			cfg := newConfigFixture(t, k)
			createExistingReleases(t, cfg, tc.existingReleases)
			req := httptest.NewRequest("PUT", fmt.Sprintf("https://example.com/whatever?%s", tc.queryString), strings.NewReader(tc.requestBody))
			response := httptest.NewRecorder()

			OperateRelease(*cfg, response, req, tc.params)

			if got, want := response.Code, tc.statusCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// using a fixed date to avoid time-dependant tests
			// ideally, we should mutate the date as we are doing for deleted releases
			fixedDateResponse := lastDeployedRegex.ReplaceAllString(response.Body.String(), `"last_deployed":"1977-09-02T22:04:05Z"`)
			if got, want := fixedDateResponse, tc.responseBody; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}

			actualReleases, err := cfg.ActionConfig.Releases.ListReleases()
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := actualReleases, tc.expectedReleases; !cmp.Equal(want, got, releaseComparer) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, releaseComparer))
			}
		})
	}
}
