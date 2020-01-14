package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/pkg/auth"
	authFake "github.com/kubeapps/kubeapps/pkg/auth/fake"
	chartFake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	helmTime "helm.sh/helm/v3/pkg/time"

	"helm.sh/helm/v3/pkg/release"
)

const defaultListLimit = 256

var (
	testingTime, _ = helmTime.Parse(time.RFC3339, "1977-09-02T22:04:05Z")
)

// newConfigFixture returns a Config with fake clients
// and memory storage.
func newConfigFixture(t *testing.T) *Config {
	t.Helper()

	return &Config{
		ActionConfig: &action.Configuration{
			Releases:     storage.Init(driver.NewMemory()),
			KubeClient:   &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}},
			Capabilities: chartutil.DefaultCapabilities,
			Log: func(format string, v ...interface{}) {
				t.Helper()
				t.Logf(format, v...)
			},
		},
		ChartClient: &chartFake.FakeChart{},
		Options: Options{
			ListLimit: defaultListLimit,
		},
	}
}

var releaseComparer = cmp.Comparer(func(x release.Release, y release.Release) bool {
	return x.Name == y.Name &&
		x.Version == y.Version &&
		x.Namespace == y.Namespace &&
		x.Info.Status == y.Info.Status &&
		x.Chart.Name() == y.Chart.Name() &&
		x.Manifest == y.Manifest &&
		cmp.Equal(x.Config, y.Config) &&
		cmp.Equal(x.Hooks, y.Hooks)
})

func TestActions(t *testing.T) {
	type testScenario struct {
		// Scenario params
		Description      string
		ExistingReleases []*release.Release
		DisableAuth      bool
		Skip             bool //TODO: Remove this when the memory bug is fixed
		// Request params
		RequestBody  string
		RequestQuery string
		Action       string
		Params       map[string]string
		// Expected result
		StatusCode        int
		RemainingReleases []*release.Release
		ResponseBody      string //optional
	}

	tests := []testScenario{
		{
			// Scenario params
			Description:      "Create a simple release without auth",
			ExistingReleases: []*release.Release{},
			DisableAuth:      true,
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Create a simple release with auth",
			ExistingReleases: []*release.Release{},
			DisableAuth:      true,
			// Request params
			RequestBody:  `{"chartName":"foo","releaseName":"foobar","version":"1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description: "Create a conflicting release",
			ExistingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
			DisableAuth: false,
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 409,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Get a non-existing release",
			ExistingReleases: []*release.Release{},
			DisableAuth:      true,
			Skip:             true,
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: []*release.Release{},
			ResponseBody:      "",
		},
		{
			Description: "Delete a simple release",
			ExistingReleases: []*release.Release{
				createRelease("foobarchart", "foobar", "default", 1, release.StatusDeployed),
			},
			DisableAuth: true,
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foobarchart", "foobar", "default", 1, release.StatusUninstalled),
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description: "Delete and purge a simple release with purge=true",
			ExistingReleases: []*release.Release{
				createRelease("foobarchart", "foobar", "default", 1, release.StatusDeployed),
			},
			DisableAuth: true,
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
			ExistingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
				createRelease("oof", "oofbar", "dev", 1, release.StatusDeployed),
			},
			DisableAuth: true,
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
				createRelease("oof", "oofbar", "dev", 1, release.StatusDeployed),
			},
			ResponseBody: `{"data":{"name":"foobar","info":{"status":{"code":1}},"chart":{"metadata":{"name":"foo"},"values":{"raw":"{}\n"}},"config":{"raw":"{}\n"},"version":1,"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description: "Get a deleted release",
			ExistingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusUninstalled),
			},
			DisableAuth: true,
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusUninstalled),
			},
			ResponseBody: `{"data":{"name":"foobar","info":{"status":{"code":2},"deleted":{"seconds":242085845}},"chart":{"metadata":{"name":"foo"},"values":{"raw":"{}\n"}},"config":{"raw":"{}\n"},"version":1,"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description: "Delete and purge a simple release with purge=1",
			ExistingReleases: []*release.Release{
				createRelease("foobarchart", "foobar", "default", 1, release.StatusDeployed),
			},
			DisableAuth: true,
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
			ExistingReleases: []*release.Release{},
			DisableAuth:      true,
			Skip:             true,
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
	}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			// TODO Remove this `if` statement after the memory driver bug is fixed
			// Memory Driver Bug: https://github.com/helm/helm/pull/7372
			if test.Skip {
				t.SkipNow()
			}
			// Initialize environment for test
			req := httptest.NewRequest("GET", fmt.Sprintf("http://foo.bar%s", test.RequestQuery), strings.NewReader(test.RequestBody))
			if !test.DisableAuth {
				fauth := &authFake.FakeAuth{}
				ctx := context.WithValue(req.Context(), auth.UserKey, fauth)
				req = req.WithContext(ctx)
			}
			response := httptest.NewRecorder()
			cfg := newConfigFixture(t)
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
			if !cmp.Equal(releases, test.RemainingReleases, releaseComparer) {
				t.Errorf("Unexpected remaining releases. Diff:\n%s", cmp.Diff(releases, test.RemainingReleases, releaseComparer))
			}
			if test.ResponseBody != "" {
				if test.ResponseBody != response.Body.String() {
					t.Errorf("Unexpected body response. Diff %s", cmp.Diff(test.ResponseBody, response.Body))
				}
			}
		})
	}
}

func createRelease(chartName, name, namespace string, version int, status release.Status) *release.Release {
	deleted := helmTime.Time{}
	if status == release.StatusUninstalled {
		deleted = testingTime
	}
	return &release.Release{
		Name:      name,
		Namespace: namespace,
		Version:   version,
		Info:      &release.Info{Status: status, Deleted: deleted},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: chartName,
			},
			Values: make(map[string]interface{}),
		},
		Config: make(map[string]interface{}),
	}
}

func createExistingReleases(t *testing.T, cfg *Config, releases []*release.Release) {
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
		existingReleases []*release.Release
		queryString      string
		params           map[string]string
		statusCode       int
		expectedReleases []*release.Release
		responseBody     string
	}{
		{
			name: "rolls back a release",
			existingReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			queryString: "action=rollback&revision=1",
			params:      map[string]string{nameParam: "my-release"},
			statusCode:  http.StatusOK,
			expectedReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 3, release.StatusDeployed),
			},
			responseBody: `{"data":{"name":"my-release","info":{"status":{"code":1}},"chart":{"metadata":{"name":"apache"},"values":{"raw":"{}\n"}},"config":{"raw":"{}\n"},"version":3,"namespace":"default"}}`,
		},
		{
			name: "errors if the release does not exist",
			existingReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			queryString: "action=rollback&revision=1",
			params:      map[string]string{nameParam: "does-not-exist"},
			statusCode:  http.StatusNotFound,
			expectedReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			responseBody: `{"code":404,"message":"no revision for release \"does-not-exist\""}`,
		},
		{
			name: "errors if the revision does not exist",
			existingReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			queryString: "action=rollback&revision=3",
			params:      map[string]string{nameParam: "apache"},
			statusCode:  http.StatusNotFound,
			expectedReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			responseBody: `{"code":404,"message":"no revision for release \"apache\""}`,
		},
		{
			name: "errors if the revision is not specified",
			existingReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			queryString: "action=rollback",
			params:      map[string]string{nameParam: "apache"},
			statusCode:  http.StatusUnprocessableEntity,
			expectedReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			responseBody: `{"code":422,"message":"Missing revision to rollback in request"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newConfigFixture(t)
			createExistingReleases(t, cfg, tc.existingReleases)
			req := httptest.NewRequest("PUT", fmt.Sprintf("https://example.com/whatever?%s", tc.queryString), strings.NewReader(""))
			response := httptest.NewRecorder()

			OperateRelease(*cfg, response, req, tc.params)

			if got, want := response.Code, tc.statusCode; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := response.Body.String(), tc.responseBody; got != want {
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
