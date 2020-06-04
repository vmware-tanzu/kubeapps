package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	chartFake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	helmTime "helm.sh/helm/v3/pkg/time"
	"k8s.io/client-go/rest"

	"helm.sh/helm/v3/pkg/release"
)

const (
	defaultListLimit        = 256
	KUBERNETES_SERVICE_HOST = "KUBERNETES_SERVICE_HOST"
	KUBERNETES_SERVICE_PORT = "KUBERNETES_SERVICE_PORT"
)

var (
	testingTime, _ = helmTime.Parse(time.RFC3339, "1977-09-02T22:04:05Z")
)

// newConfigFixture returns a Config with fake clients
// and memory storage.
func newConfigFixture(t *testing.T, k *kubefake.FailingKubeClient) *Config {
	t.Helper()

	return &Config{
		ActionConfig: &action.Configuration{
			Releases:     storage.Init(driver.NewMemory()),
			KubeClient:   k,
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

var releaseComparer = cmp.Comparer(func(x release.Release, y release.Release) bool {
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
		ExistingReleases []*release.Release
		Skip             bool //TODO: Remove this when the memory bug is fixed
		KubeError        error
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
			Description:      "Create a simple release",
			ExistingReleases: []*release.Release{},
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
			Description: "Create a conflicting release",
			ExistingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
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
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusDeployed),
			},
			ResponseBody: `{"data":{"name":"foobar","info":{"status":{"code":1}},"chart":{"metadata":{"name":"foo"},"values":{"raw":"{}\n"}},"config":{"raw":"{}\n"},"version":1,"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description: "Get a deleted release",
			ExistingReleases: []*release.Release{
				createRelease("foo", "foobar", "default", 1, release.StatusUninstalled),
			},
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
		{
			// Scenario params
			Description:      "Creates a release with missing permissions",
			ExistingReleases: []*release.Release{},
			KubeError:        errors.New(`Failed to create: secrets is forbidden: User "foo" cannot create resource "secrets" in API group "" in the namespace "default"`),
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
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
			// TODO Remove this `if` statement after the memory driver bug is fixed
			// Memory Driver Bug: https://github.com/helm/helm/pull/7372
			if test.Skip {
				t.SkipNow()
			}
			// Initialize environment for test
			req := httptest.NewRequest("GET", fmt.Sprintf("http://foo.bar%s", test.RequestQuery), strings.NewReader(test.RequestBody))
			response := httptest.NewRecorder()
			k := &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}}
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
			responseBody: `{"code":404,"message":"release: not found"}`,
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
			responseBody: `{"code":404,"message":"release: not found"}`,
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
			k := &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}}
			cfg := newConfigFixture(t, k)
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

func TestUpgradeAction(t *testing.T) {
	const releaseName = "my-release"
	testCases := []struct {
		name             string
		existingReleases []*release.Release
		queryString      string
		requestBody      string
		params           map[string]string
		statusCode       int
		expectedReleases []*release.Release
		responseBody     string
	}{
		{
			name: "upgrade a release",
			existingReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusDeployed),
			},
			queryString: "action=upgrade",
			requestBody: `{"chartName": "apache",	"releaseName":"my-release",	"version": "1.0.0"}`,
			params:     map[string]string{nameParam: releaseName},
			statusCode: http.StatusOK,
			expectedReleases: []*release.Release{
				createRelease("apache", releaseName, "default", 1, release.StatusSuperseded),
				createRelease("apache", releaseName, "default", 2, release.StatusDeployed),
			},
			responseBody: `{"data":{"name":"my-release","info":{"status":{"code":1}},"chart":{"metadata":{"name":"apache"},"values":{"raw":"{}\n"}},"config":{"raw":"{}\n"},"version":2,"namespace":"default"}}`,
		},
		{
			name:             "upgrade a missing release",
			existingReleases: []*release.Release{},
			queryString:      "action=upgrade",
			requestBody: `{"chartName": "apache",	"releaseName":"my-release",	"version": "1.0.0"}`,
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
			k := &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: ioutil.Discard}}
			cfg := newConfigFixture(t, k)
			createExistingReleases(t, cfg, tc.existingReleases)
			req := httptest.NewRequest("PUT", fmt.Sprintf("https://example.com/whatever?%s", tc.queryString), strings.NewReader(tc.requestBody))
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

func TestNewClusterConfig(t *testing.T) {
	testCases := []struct {
		name               string
		token              string
		cluster            string
		additionalClusters map[string]AdditionalClusterConfig
		inClusterConfig    *rest.Config
		expectedConfig     *rest.Config
		errorExpected      bool
	}{
		{
			name:    "returns an in-cluster with explicit token for the default cluster",
			token:   "token-1",
			cluster: "default",
			inClusterConfig: &rest.Config{
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
			},
			expectedConfig: &rest.Config{
				BearerToken:     "token-1",
				BearerTokenFile: "",
			},
		},
		{
			name:    "returns a config setup for an additional cluster",
			token:   "token-1",
			cluster: "cluster-1",
			additionalClusters: map[string]AdditionalClusterConfig{
				"cluster-1": AdditionalClusterConfig{
					ApiServiceURL:            "https://cluster-1.example.com:7890",
					CertificateAuthorityData: "ca-file-data",
				},
			},
			inClusterConfig: &rest.Config{
				Host:            "https://something-else.example.com:6443",
				BearerToken:     "something-else",
				BearerTokenFile: "/foo/bar",
				TLSClientConfig: rest.TLSClientConfig{
					CAFile: "/var/run/whatever/ca.crt",
				},
			},
			expectedConfig: &rest.Config{
				Host:            "https://cluster-1.example.com:7890",
				BearerToken:     "token-1",
				BearerTokenFile: "",
				TLSClientConfig: rest.TLSClientConfig{
					CAData: []byte("ca-file-data"),
				},
			},
		},
		{
			name:            "returns an error if the cluster does not exist",
			cluster:         "cluster-1",
			inClusterConfig: &rest.Config{},
			errorExpected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := NewClusterConfig(tc.inClusterConfig, tc.token, tc.cluster, tc.additionalClusters)
			if got, want := err != nil, tc.errorExpected; got != want {
				t.Fatalf("got: %t, want: %t. err: %+v", got, want, err)
			}

			if got, want := config, tc.expectedConfig; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
