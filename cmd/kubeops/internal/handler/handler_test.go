package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/kubeapps/kubeapps/pkg/agent"
	"github.com/kubeapps/kubeapps/pkg/auth"
	authFake "github.com/kubeapps/kubeapps/pkg/auth/fake"
	chartFake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"helm.sh/helm/v3/pkg/release"
)

const defaultListLimit = 256

// newConfigFixture returns an agent.Config with fake clients
// and memory storage.
func newConfigFixture(t *testing.T) *agent.Config {
	t.Helper()

	return &agent.Config{
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
		AgentOptions: agent.Options{
			ListLimit: defaultListLimit,
		},
	}
}

func TestActions(t *testing.T) {
	type testScenario struct {
		// Scenario params
		Description      string
		ExistingReleases []release.Release
		DisableAuth      bool
		// Request params
		RequestBody  string
		RequestQuery string
		Action       string
		Params       map[string]string
		// Expected result
		StatusCode        int
		RemainingReleases []release.Release
		ResponseBody      string //optional
	}

	tests := []testScenario{}

	for _, test := range tests {
		// Initialize environment for test
		req := httptest.NewRequest("GET", fmt.Sprintf("http://foo.bar%s", test.RequestQuery), strings.NewReader(test.RequestBody))
		if !test.DisableAuth {
			fauth := &authFake.FakeAuth{}
			ctx := context.WithValue(req.Context(), auth.UserKey, fauth)
			req = req.WithContext(ctx)
		}
		response := httptest.NewRecorder()
		cfg := newConfigFixture(t)
		for i := range test.ExistingReleases {
			err := cfg.ActionConfig.Releases.Create(&test.ExistingReleases[i])
			if err != nil {
				t.Errorf("Failed to initiate test: %v", err)
			}
		}
		// Perform request
		t.Log(test.Description)
		switch test.Action {
		default:
			t.Errorf("Unexpected action %s", test.Action)
		}
		// Check result
		if response.Code != test.StatusCode {
			t.Errorf("Expecting a StatusCode %d, received %d", test.StatusCode, response.Code)
		}
		releases := extractReleases(cfg.ActionConfig.Releases)
		if !reflect.DeepEqual(releases, test.RemainingReleases) {
			t.Errorf("Unexpected remaining releases. Expecting %v, found %v", test.RemainingReleases, releases)
		}
		if test.ResponseBody != "" {
			if test.ResponseBody != response.Body.String() {
				t.Errorf("Unexpected body response. Expecting %s, found %s", test.ResponseBody, response.Body)
			}
		}
	}
}

func extractReleases(storage *storage.Storage) []release.Release {
	rls, _ := storage.ListReleases()
	releases := make([]release.Release, len(rls))
	//cleanup unused properties
	for i := range rls {
		//dereference element
		releases[i] = *rls[i]
		//save status information only by ignoring timestamps
		releases[i].Info = &release.Info{}            //set all timestamps to zeroth value
		releases[i].SetStatus(rls[i].Info.Status, "") //copy only relevant `status` information
	}
	return releases
}

func createRelease(chartName, name, namespace string, version int, status release.Status) release.Release {
	return release.Release{
		Name:      name,
		Namespace: namespace,
		Version:   version,
		Info:      &release.Info{Status: status},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: chartName,
			},
			Values: make(map[string]interface{}),
		},
		Config: make(map[string]interface{}),
	}
}
