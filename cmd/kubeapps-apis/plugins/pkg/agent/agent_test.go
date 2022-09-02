// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"io"
	"testing"
	"time"

	kubechart "github.com/vmware-tanzu/kubeapps/pkg/chart"
	chartFake "github.com/vmware-tanzu/kubeapps/pkg/chart/fake"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/client-go/rest"
)

// newActionConfigFixture returns an action.Configuration with fake clients
// and memory storage.
func newActionConfigFixture(t *testing.T) *action.Configuration {
	t.Helper()

	return &action.Configuration{
		Releases:     storage.Init(driver.NewMemory()),
		KubeClient:   &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}
}

type releaseStub struct {
	name         string
	namespace    string
	version      int
	chartVersion string
	status       release.Status
}

// makeReleases adds a slice of releases to the configured storage.
func makeReleases(t *testing.T, actionConfig *action.Configuration, rels []releaseStub) {
	t.Helper()
	storage := actionConfig.Releases
	for _, r := range rels {
		rel := &release.Release{
			Name:      r.name,
			Namespace: r.namespace,
			Version:   r.version,
			Info: &release.Info{
				Status: r.status,
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version: r.chartVersion,
					Icon:    "https://example.com/icon.png",
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
		desc              string
		releaseName       string
		namespace         string
		chartName         string
		values            string
		version           int
		existingReleases  []releaseStub
		remainingReleases int
		shouldFail        bool
	}{
		{
			desc:      "install new release",
			chartName: "mychart",
			values:    "",
			namespace: "default",
			version:   1,
			existingReleases: []releaseStub{
				{"otherchart", "default", 1, "1.0.0", release.StatusDeployed},
			},
			remainingReleases: 2,
			shouldFail:        false,
		},
		{
			desc:      "install with an existing name",
			chartName: "mychart",
			values:    "",
			namespace: "default",
			version:   1,
			existingReleases: []releaseStub{
				{"mychart", "default", 1, "1.0.0", release.StatusDeployed},
			},
			remainingReleases: 1,
			shouldFail:        true,
		},
		{
			desc:      "install with same name different version",
			chartName: "mychart",
			values:    "",
			namespace: "dev",
			version:   1,
			existingReleases: []releaseStub{
				{"mychart", "dev", 2, "1.0.0", release.StatusDeployed},
			},
			remainingReleases: 1,
			shouldFail:        true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Initialize environment for test
			actionConfig := newActionConfigFixture(t)
			makeReleases(t, actionConfig, tc.existingReleases)
			fakechart := chartFake.ChartClient{}
			ch, _ := fakechart.GetChart(&kubechart.Details{
				ChartName: tc.chartName,
			}, "")
			// Perform test
			rls, err := CreateRelease(actionConfig, tc.chartName, tc.namespace, tc.values, ch, nil, 0)
			// Check result
			if tc.shouldFail && err == nil {
				t.Errorf("Should fail with %v; instead got %s in %s", tc.desc, tc.releaseName, tc.namespace)
			}
			if !tc.shouldFail && rls == nil {
				t.Errorf("Should succeed with %v; instead got error %v", tc.desc, err)
			}
			rlss, err := actionConfig.Releases.ListReleases()
			if err != nil {
				t.Errorf("Unexpected err %v", err)
			}
			if len(rlss) != tc.remainingReleases {
				t.Errorf("Expecting %d remaining releases, got %d", tc.remainingReleases, len(rlss))
			}
		})
	}
}

func TestDeleteRelease(t *testing.T) {
	testCases := []struct {
		description     string
		releases        []releaseStub
		releaseToDelete string
		namespace       string
		shouldFail      bool
	}{
		{
			description: "Delete a release",
			releases: []releaseStub{
				{"airwatch", "default", 1, "1.0.0", release.StatusDeployed},
			},
			releaseToDelete: "airwatch",
		},
		{
			description: "Delete a non-existing release",
			releases: []releaseStub{
				{"airwatch", "default", 1, "1.0.0", release.StatusDeployed},
			},
			releaseToDelete: "apache",
			shouldFail:      true,
		},
		{
			description: "Delete a release in different namespace",
			releases: []releaseStub{
				{"airwatch", "default", 1, "1.0.0", release.StatusDeployed},
				{"apache", "dev", 1, "1.0.0", release.StatusDeployed},
			},
			releaseToDelete: "apache",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			cfg := newActionConfigFixture(t)
			makeReleases(t, cfg, tc.releases)
			err := DeleteRelease(cfg, tc.releaseToDelete, true, 0)
			t.Logf("error: %v", err)
			if didFail := err != nil; didFail != tc.shouldFail {
				t.Errorf("wanted fail = %v, got fail = %v", tc.shouldFail, err != nil)
			}
		})
	}
}

func TestRollbackRelease(t *testing.T) {
	const (
		revisionBeingSuperseded = 2
		targetRevision          = 1
	)

	testCases := []struct {
		name      string
		releases  []releaseStub
		release   string
		namespace string
		revision  int
		err       error
	}{
		{
			name: "rolls back a release",
			releases: []releaseStub{
				{"airwatch", "default", targetRevision, "1.0.0", release.StatusSuperseded},
				{"airwatch", "default", revisionBeingSuperseded, "1.0.0", release.StatusDeployed},
			},
			release:  "airwatch",
			revision: targetRevision,
		},
		{
			name: "errors when rolling back to a release revision which does not exist",
			releases: []releaseStub{
				{"airwatch", "default", revisionBeingSuperseded, "1.0.0", release.StatusDeployed},
			},
			release:  "airwatch",
			revision: targetRevision,
			err:      driver.ErrReleaseNotFound,
		},
		{
			name: "rolls back a release in non-default namespace",
			releases: []releaseStub{
				{"otherrelease", "default", 1, "1.0.0", release.StatusDeployed},
				{"airwatch", "othernamespace", targetRevision, "1.0.0", release.StatusSuperseded},
				{"airwatch", "othernamespace", revisionBeingSuperseded, "1.0.0", release.StatusDeployed},
			},
			release:  "airwatch",
			revision: targetRevision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newActionConfigFixture(t)
			makeReleases(t, cfg, tc.releases)

			newRelease, err := RollbackRelease(cfg, tc.release, tc.revision, 0)
			if got, want := err, tc.err; got != want {
				t.Errorf("got: %v, want: %v", got, want)
			}
			if tc.err != nil {
				return
			}

			// Previously deployed revision gets superseded
			rel, err := cfg.Releases.Get(tc.release, revisionBeingSuperseded)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if got, want := rel.Info.Status, release.StatusSuperseded; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}

			// Target revision is deployed as a new revision
			rel, err = cfg.Releases.Get(tc.release, revisionBeingSuperseded+1)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if got, want := newRelease.Version, revisionBeingSuperseded+1; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := rel.Info.Status, release.StatusDeployed; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}

func TestUpgradeRelease(t *testing.T) {
	const revisionBeingUpdated = 1
	testCases := []struct {
		description string
		releases    []releaseStub
		release     string
		valuesYaml  string
		chartName   string
		shouldFail  bool
	}{
		{
			description: "upgrade a release with chart",
			releases: []releaseStub{
				{"myrls", "default", revisionBeingUpdated, "mychart", release.StatusDeployed},
			},
			valuesYaml: "IsValidYaml: true",
			release:    "myrls",
			chartName:  "mynewchart",
		},
		{
			description: "upgrade a release with invalid values",
			releases: []releaseStub{
				{"myrls", "default", revisionBeingUpdated, "mychart", release.StatusDeployed},
			},
			valuesYaml: "\\-xx-@myval:\"test value\"\\\n", // <- invalid yaml
			release:    "myrls",
			chartName:  "mynewchart",
			shouldFail: true,
		},
		{
			description: "upgrade a deleted release",
			releases: []releaseStub{
				{"myrls", "default", revisionBeingUpdated, "mychart", release.StatusUninstalled},
			},
			release:    "myrls",
			chartName:  "mynewchart",
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			cfg := newActionConfigFixture(t)
			makeReleases(t, cfg, tc.releases)
			fakechart := chartFake.ChartClient{}
			ch, _ := fakechart.GetChart(&kubechart.Details{
				ChartName: tc.chartName,
			}, "")
			newRelease, err := UpgradeRelease(cfg, tc.release, tc.valuesYaml, ch, nil, 0)
			// Check for errors
			if got, want := err != nil, tc.shouldFail; got != want {
				t.Errorf("Failure: got: %v, want: %v", got, want)
			}
			if tc.shouldFail {
				return
			}
			// Target revision is deployed as a new revision
			rel, err := cfg.Releases.Get(tc.release, revisionBeingUpdated+1)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if got, want := newRelease.Version, revisionBeingUpdated+1; got != want {
				t.Errorf("got: %d, want: %d", got, want)
			}
			if got, want := rel.Info.Status, release.StatusDeployed; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
			//check original version is superseded
			rel, err = cfg.Releases.Get(tc.release, revisionBeingUpdated)
			if err != nil {
				t.Fatalf("%+v", err)
			}
			if got, want := rel.Info.Status, release.StatusSuperseded; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}

func TestNewConfigFlagsFromCluster(t *testing.T) {
	testCases := []struct {
		name   string
		config rest.Config
	}{
		{
			name: "bearer token remains for an https host",
			config: rest.Config{
				Host:        "https://example.com/",
				APIPath:     "",
				BearerToken: "foo",
			},
		},
		{
			name: "bearer token remains for an http host",
			config: rest.Config{
				Host:        "http://example.com/",
				APIPath:     "",
				BearerToken: "foo",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			restClientGetter := NewConfigFlagsFromCluster("namespace-a", &tc.config)

			config, err := restClientGetter.ToRESTConfig()
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := config.BearerToken, tc.config.BearerToken; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}

// Make sure that Helm command gets the timeout
func TestNewInstallCommand(t *testing.T) {
	testCases := []struct {
		name     string
		timeout  int32
		expected time.Duration
	}{
		{
			name:     "install command with no timeout",
			timeout:  0,
			expected: time.Duration(0),
		},
		{
			name:     "install command with invalid timeout",
			timeout:  -10,
			expected: time.Duration(0),
		},
		{
			name:     "install command with timeout",
			timeout:  33,
			expected: time.Duration(33) * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := newActionConfigFixture(t)
			cmd, err := newInstallCommand(cfg, "", "", nil, tc.timeout)

			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := cmd.Timeout, tc.expected; got != want {
				t.Errorf("got: %q, want: %q", got, want)
			}
		})
	}
}
