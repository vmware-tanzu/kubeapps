package agent

import (
	"io/ioutil"
	"testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/kubeapps/kubeapps/pkg/proxy"
)

const defaultListLimit = 256

// newConfigFixture returns an agent.Config with fake clients
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
	}
}

type releaseStub struct {
	name      string
	namespace string
	version   int
}

// makeReleases adds a slice of releases to the configured storage.
func makeReleases(t *testing.T, config *Config, rels []releaseStub) {
	t.Helper()
	storage := config.ActionConfig.Releases
	for _, r := range rels {
		rel := &release.Release{
			Name:      r.name,
			Namespace: r.namespace,
			Version:   r.version,
			Info: &release.Info{
				Status: release.StatusDeployed,
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Icon: "https://example.com/icon.png",
				},
			},
		}
		err := storage.Create(rel)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestListReleases(t *testing.T) {
	testCases := []struct {
		name         string
		namespace    string
		listLimit    int
		releases     []releaseStub
		expectedApps []proxy.AppOverview
	}{
		{
			name:      "returns all apps across namespaces",
			namespace: "",
			listLimit: defaultListLimit,
			releases: []releaseStub{
				releaseStub{"wordpress", "default", 1},
				releaseStub{"airwatch", "default", 1},
				releaseStub{"not-in-default-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
				},
				proxy.AppOverview{
					ReleaseName: "wordpress",
				},
				proxy.AppOverview{
					ReleaseName: "not-in-default-namespace",
				},
			},
		},
		{
			name:      "returns apps for the given namespace",
			namespace: "default",
			listLimit: defaultListLimit,
			releases: []releaseStub{
				releaseStub{"wordpress", "default", 1},
				releaseStub{"airwatch", "default", 1},
				releaseStub{"not-in-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
				},
				proxy.AppOverview{
					ReleaseName: "wordpress",
				},
			},
		},
		{
			name:      "returns at most listLimit apps",
			namespace: "default",
			listLimit: 1,
			releases: []releaseStub{
				releaseStub{"wordpress", "default", 1},
				releaseStub{"airwatch", "default", 1},
				releaseStub{"not-in-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := newConfigFixture(t)
			makeReleases(t, config, tc.releases)

			apps, err := ListReleases(config.ActionConfig, tc.namespace, tc.listLimit, "ignored?")
			if err != nil {
				t.Errorf("%v", err)
			}

			if got, want := len(apps), len(tc.expectedApps); got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
		})
	}
}

func TestParseDriverType(t *testing.T) {
	validTestCases := []struct {
		input  string
		output DriverType
	}{
		{
			input:  "secret",
			output: Secret,
		},
		{
			input:  "secrets",
			output: Secret,
		},
		{
			input:  "configmap",
			output: ConfigMap,
		},
		{
			input:  "configmaps",
			output: ConfigMap,
		},
		{
			input:  "memory",
			output: Memory,
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.input, func(t *testing.T) {
			driverType, err := ParseDriverType(tc.input)
			if err != nil {
				t.Errorf("%v", err)
			} else if driverType != tc.output {
				t.Errorf("expected: %s, actual: %s", tc.output, driverType)
			}
		})
	}

	invalidTestCase := "andresmgot"
	t.Run(invalidTestCase, func(t *testing.T) {
		driverType, err := ParseDriverType(invalidTestCase)
		if err == nil {
			t.Errorf("Expected \"%s\" to be an invalid driver type, but it was parsed as %v", invalidTestCase, driverType)
		}
	})
}
