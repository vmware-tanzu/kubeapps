package agent

import (
	"io/ioutil"
	"sort"
	"testing"

	kubechart "github.com/kubeapps/kubeapps/pkg/chart"
	chartFake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"

	"github.com/google/go-cmp/cmp"
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
		ChartClient: &chartFake.FakeChart{},
		AgentOptions: Options{
			ListLimit: defaultListLimit,
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

func TestCreateReleases(t *testing.T) {
	testCases := []struct {
		desc             string
		releaseName      string
		namespace        string
		chartName        string
		values           string
		version          int
		existingReleases []releaseStub
		shouldFail       bool
	}{
		{
			desc:      "install new release",
			chartName: "mychart",
			values:    "",
			namespace: "default",
			version:   1,
			existingReleases: []releaseStub{
				releaseStub{"otherchart", "default", 1},
			},
			shouldFail: false,
		},
		{
			desc:      "install with an existing name",
			chartName: "mychart",
			values:    "",
			namespace: "default",
			version:   1,
			existingReleases: []releaseStub{
				releaseStub{"mychart", "default", 1},
			},
			shouldFail: true,
		},
		{
			desc:      "install with same name different version",
			chartName: "mychart",
			values:    "",
			namespace: "dev",
			version:   1,
			existingReleases: []releaseStub{
				releaseStub{"mychart", "dev", 2},
			},
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Initialize environment for test
			config := newConfigFixture(t)
			makeReleases(t, config, tc.existingReleases)
			fakechart := chartFake.FakeChart{}
			ch, _ := fakechart.GetChart(&kubechart.Details{
				ChartName: tc.chartName,
			}, nil, false)
			// Perform test
			rls, err := CreateRelease(*config, tc.chartName, tc.namespace, tc.values, ch.Helm3Chart)
			// Check result
			if tc.shouldFail && err == nil {
				t.Errorf("Should fail with %v; instead got %s in %s", tc.desc, tc.releaseName, tc.namespace)
			}
			if !tc.shouldFail && rls == nil {
				t.Errorf("Should succeed with %v; instead got error %v", tc.desc, err)
			}
		})
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
				releaseStub{"airwatch", "default", 1},
				releaseStub{"wordpress", "default", 1},
				releaseStub{"not-in-default-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
				proxy.AppOverview{
					ReleaseName: "not-in-default-namespace",
					Namespace:   "other",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
				proxy.AppOverview{
					ReleaseName: "wordpress",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
			},
		},
		{
			name:      "returns apps for the given namespace",
			namespace: "default",
			listLimit: defaultListLimit,
			releases: []releaseStub{
				releaseStub{"airwatch", "default", 1},
				releaseStub{"wordpress", "default", 1},
				releaseStub{"not-in-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
				proxy.AppOverview{
					ReleaseName: "wordpress",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
			},
		},
		{
			name:      "returns at most listLimit apps",
			namespace: "default",
			listLimit: 1,
			releases: []releaseStub{
				releaseStub{"airwatch", "default", 1},
				releaseStub{"wordpress", "default", 1},
				releaseStub{"not-in-namespace", "other", 1},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "airwatch",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
			},
		},
		{
			name:      "returns two apps with same name but different namespaces and versions",
			namespace: "",
			listLimit: defaultListLimit,
			releases: []releaseStub{
				releaseStub{"wordpress", "default", 1},
				releaseStub{"wordpress", "dev", 2},
			},
			expectedApps: []proxy.AppOverview{
				proxy.AppOverview{
					ReleaseName: "wordpress",
					Namespace:   "default",
					Version:     "1",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
				},
				proxy.AppOverview{
					ReleaseName: "wordpress",
					Namespace:   "dev",
					Version:     "2",
					Status:      "deployed",
					Icon:        "https://example.com/icon.png",
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

			// Check for size of returned apps
			if got, want := len(apps), len(tc.expectedApps); got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}

			// The Helm memory driver does not appear to have consistent ordering.
			// See https://github.com/helm/helm/issues/7263
			// Just sort by version which is good enough here.
			sort.Slice(apps, func(i, j int) bool { return apps[i].Version < apps[j].Version })

			//Deep equality check of expected against attained result
			if !cmp.Equal(apps, tc.expectedApps) {
				t.Errorf(cmp.Diff(apps, tc.expectedApps))
			}
		})
	}
}

func TestParseDriverType(t *testing.T) {
	validTestCases := []struct {
		input      string
		driverName string
	}{
		{
			input:      "secret",
			driverName: "Secret",
		},
		{
			input:      "secrets",
			driverName: "Secret",
		},
		{
			input:      "configmap",
			driverName: "ConfigMap",
		},
		{
			input:      "configmaps",
			driverName: "ConfigMap",
		},
		{
			input:      "memory",
			driverName: "Memory",
		},
	}

	for _, tc := range validTestCases {
		t.Run(tc.input, func(t *testing.T) {
			storageForDriver, err := ParseDriverType(tc.input)
			if err != nil {
				t.Fatalf("%v", err)
			}
			storage := storageForDriver("default", &kubernetes.Clientset{})
			if got, want := storage.Name(), tc.driverName; got != want {
				t.Errorf("expected: %s, actual: %s", want, got)
			}
		})
	}

	invalidTestCase := "andresmgot"
	t.Run(invalidTestCase, func(t *testing.T) {
		storageForDriver, err := ParseDriverType(invalidTestCase)
		if err == nil {
			t.Errorf("Expected \"%s\" to be an invalid driver type, but it was parsed as %v", invalidTestCase, storageForDriver)
		}
		if storageForDriver != nil {
			t.Errorf("got: %#v, want: nil", storageForDriver)
		}
	})
}
