/*
Copyright (c) 2018 Bitnami

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

package proxy

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

type AppOverviewTest struct {
	AppOverview
	Manifest string
}

func newFakeProxyWithManifest(existingTillerReleases []AppOverviewTest) *Proxy {
	helmClient := helm.FakeClient{}
	// Populate Fake helm client with releases
	for _, r := range existingTillerReleases {
		status := release.Status_DEPLOYED
		if r.Status == "DELETED" {
			status = release.Status_DELETED
		} else if r.Status == "FAILED" {
			status = release.Status_FAILED
		}
		version := int32(1)
		// Increment version number (helm revision counter)
		// if the same release name has been already added
		for _, versionAdded := range helmClient.Rels {
			if r.ReleaseName == versionAdded.GetName() {
				version++
			}
		}
		helmClient.Rels = append(helmClient.Rels, &release.Release{
			Name:      r.ReleaseName,
			Namespace: r.Namespace,
			Version:   version,
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version: r.Version,
					Icon:    r.Icon,
					Name:    r.Chart,
				},
			},
			Info: &release.Info{
				Status: &release.Status{
					Code: status,
				},
			},
			Manifest: r.Manifest,
		})
	}
	kubeClient := fake.NewSimpleClientset()
	return NewProxy(kubeClient, &helmClient)
}

func newFakeProxy(existingTillerReleases []AppOverview) *Proxy {
	releasesWithManifest := []AppOverviewTest{}
	for _, r := range existingTillerReleases {
		releasesWithManifest = append(releasesWithManifest, AppOverviewTest{r, ""})
	}
	return newFakeProxyWithManifest(releasesWithManifest)
}

func TestListAllReleases(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon2.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases("", 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 2 {
		t.Errorf("It should return both releases")
	}
	if !reflect.DeepEqual([]AppOverview{app1, app2}, releases) {
		t.Log(releases[0].ChartMetadata)
		t.Log(app1.ChartMetadata)
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestListNamespacedRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon2.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases(app1.Namespace, 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 1 {
		t.Errorf("It should return both releases")
	}
	if !reflect.DeepEqual([]AppOverview{app1}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestListOldRelease(t *testing.T) {
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	appUpgraded := AppOverview{"foo", "1.0.1", "my_ns", "icon.png", "FAILED", "wordpress", chart.Metadata{
		Version: "1.0.1",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app, appUpgraded})

	// Should avoid old release versions
	releases, err := proxy.ListReleases(app.Namespace, 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 1 {
		t.Errorf("It should return a single release")
	}
	if releases[0].ReleaseName != "foo" && releases[0].Status != "FAILED" {
		t.Errorf("It should group releases by release name")
	}
	if !reflect.DeepEqual([]AppOverview{appUpgraded}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestMultipleOldReleases(t *testing.T) {
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	appUpgraded := AppOverview{"foo", "1.0.1", "my_ns", "icon.png", "FAILED", "wordpress", chart.Metadata{
		Version: "1.0.1",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2 := AppOverview{"bar", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2Outdated := AppOverview{"bar", "1.0.2", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.2",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2Upgraded := AppOverview{"bar", "1.0.2", "my_ns", "icon.png", "FAILED", "wordpress", chart.Metadata{
		Version: "1.0.2",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app, appUpgraded, app2, app2Outdated, app2Upgraded})

	// Should avoid old release versions
	releases, err := proxy.ListReleases(app.Namespace, 256, "")
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 2 {
		t.Errorf("It should return two unique releases")
	}
	if releases[0].ReleaseName != "foo" && releases[0].Status != "FAILED" {
		t.Errorf("It should group releases by release name")
	}
	if releases[1].ReleaseName != "bar" && releases[1].Status != "FAILED" {
		t.Errorf("It should group releases by release name")
	}
	if !reflect.DeepEqual([]AppOverview{appUpgraded, app2Upgraded}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestGetStatuses(t *testing.T) {
	// TODO: We should test that the helm client receive the correct params
	// but right now the fake implementation ignores the status filter option
	type test struct {
		input          string
		expectedResult []release.Status_Code
	}
	tests := []test{
		{"", []release.Status_Code{release.Status_DEPLOYED, release.Status_FAILED}},
		{"all", allReleaseStatuses},
		{"deleted,deployed", []release.Status_Code{release.Status_DELETED, release.Status_DEPLOYED}},
		{"deleted,none,deployed", []release.Status_Code{release.Status_DELETED, release.Status_DEPLOYED}},
	}
	for _, tt := range tests {
		res := getStatuses(tt.input)
		if !reflect.DeepEqual(res, tt.expectedResult) {
			t.Errorf("Expecting %v, received %v", tt.expectedResult, res)
		}
	}
}

func TestResolveManifest(t *testing.T) {
	ns := "myns"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}
	proxy := newFakeProxy([]AppOverview{})

	manifest, err := proxy.ResolveManifest(ns, "", ch)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if !strings.Contains(manifest, "apiVersion") || !strings.Contains(manifest, "kind") {
		t.Errorf("%s doesn't contain a manifest", manifest)
	}
	if strings.HasPrefix(manifest, "\n") {
		t.Error("The manifest should not contain new lines at the beginning")
	}
}

func TestResolveManifestFromRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon2.png",
		Name:    "wordpress",
	}}
	type testStruct struct {
		description      string
		existingApps     []AppOverviewTest
		releaseName      string
		shouldFail       bool
		expectedManifest string
	}
	tests := []testStruct{
		{
			"should return the right manifest",
			[]AppOverviewTest{{app1, "foo: bar"}, {app2, "bar: foo"}},
			app2.ReleaseName,
			false,
			"bar: foo",
		},
		{
			"should trim initial empty lines",
			[]AppOverviewTest{{app1, "\nfoo: bar"}, {app2, "bar: foo"}},
			app1.ReleaseName,
			false,
			"foo: bar",
		},
		{
			"should fail if the app doesn't exists",
			[]AppOverviewTest{{app1, "foo: bar"}, {app2, "bar: foo"}},
			"foobar",
			true,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			proxy := newFakeProxyWithManifest(test.existingApps)

			manifest, err := proxy.ResolveManifestFromRelease(test.releaseName, 1)
			if test.shouldFail {
				if err == nil {
					t.Error("Test should have failed")
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error %v", err)
				}
				if test.expectedManifest != manifest {
					t.Errorf("manifest doesn't match. Want %s got %s", test.expectedManifest, manifest)
				}
				if strings.HasPrefix(manifest, "\n") {
					t.Error("The manifest should not contain new lines at the beginning")
				}
			}
		})
	}
}

func TestCreateHelmRelease(t *testing.T) {
	ns := "myns"
	rs := "foo"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}
	proxy := newFakeProxy([]AppOverview{})

	result, err := proxy.CreateRelease(rs, ns, "", ch)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if result.Name != rs {
		t.Errorf("Expected release named %s received %s", rs, result.Name)
	}
	if result.Namespace != ns {
		t.Errorf("Expected release in namespace %s received %s", ns, result.Namespace)
	}
	// We cannot check that the rest of the chart properties as properly set
	// because the fake InstallReleaseFromChart ignores the given chart
}

func TestCreateConflictingHelmRelease(t *testing.T) {
	ns := "myns"
	rs := "foo"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}
	ns2 := "other_ns"
	app := AppOverview{rs, version, ns2, "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	_, err := proxy.CreateRelease(rs, ns, "", ch)
	if err == nil {
		t.Error("Release should fail, an existing release in a different namespace already exists")
	}
	if !strings.Contains(err.Error(), "name that is still in use") {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestHelmReleaseUpdated(t *testing.T) {
	ns := "myns"
	rs := "foo"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}
	app := AppOverview{rs, version, ns, "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	result, err := proxy.UpdateRelease(rs, ns, "", ch)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if result.Name != rs {
		t.Errorf("Expected release named %s received %s", rs, result.Name)
	}
	if result.Namespace != ns {
		t.Errorf("Expected release in namespace %s received %s", ns, result.Namespace)
	}
	rels, err := proxy.helmClient.ListReleases()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	// We cannot test that the release content changes because fake UpdateReleaseResponse
	// does not modify the release
	if len(rels.Releases) != 1 {
		t.Errorf("Unexpected amount of releases %d, it should update the existing one", len(rels.Releases))
	}
}

func TestUpdateMissingHelmRelease(t *testing.T) {
	ns := "myns"
	rs := "foo"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}

	ns2 := "other_ns"
	rs2 := "not_foo"
	app := AppOverview{rs2, version, ns2, "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	_, err := proxy.UpdateRelease(rs, ns, "", ch)
	if err == nil {
		t.Error("Update should fail, there is not a release in the namespace specified")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("release: \"%s\" not found", rs)) {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestRollbackRelease(t *testing.T) {
	ns := "myns"
	rs := "foo"
	version := "v1.0.0"

	app := AppOverview{rs, version, ns, "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	_, err := proxy.RollbackRelease(rs, ns, 1)
	if err != nil {
		t.Errorf("Update should not fail %v", err)
	}
}

func TestRollbackMissingRelease(t *testing.T) {
	ns := "myns"
	rs := "foo"
	version := "v1.0.0"

	ns2 := "other_ns"
	rs2 := "not_foo"
	app := AppOverview{rs2, version, ns2, "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	_, err := proxy.RollbackRelease(rs, ns, 1)
	if err == nil {
		t.Error("Update should fail, there is not a release in the namespace specified")
	}
	if !strings.Contains(err.Error(), fmt.Sprintf("release: \"%s\" not found", rs)) {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestGetHelmRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon2.png",
		Name:    "wordpress",
	}}
	type testStruct struct {
		existingApps    []AppOverview
		shouldFail      bool
		targetApp       string
		targetNamespace string
		expectedResult  string
	}
	tests := []testStruct{
		{[]AppOverview{app1, app2}, false, "foo", "my_ns", "foo"},
		{[]AppOverview{app1, app2}, false, "foo", "", "foo"},
		// Can't retrieve release from another namespace
		{[]AppOverview{app1, app2}, true, "foo", "other_ns", "foo"},
	}
	for _, test := range tests {
		proxy := newFakeProxy(test.existingApps)
		res, err := proxy.GetRelease(test.targetApp, test.targetNamespace)
		if test.shouldFail && err == nil {
			t.Errorf("Get %s/%s should fail", test.targetNamespace, test.targetApp)
		}
		if !test.shouldFail {
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if res.Name != test.expectedResult {
				t.Errorf("Expecting app %s, received %s", test.expectedResult, res.Name)
			}
		}
	}
}

func TestHelmReleaseDeleted(t *testing.T) {
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	// TODO: Add a test for a non-purged release when the fake helm cli supports it
	err := proxy.DeleteRelease(app.ReleaseName, app.Namespace, true)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	rels, err := proxy.helmClient.ListReleases()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if len(rels.Releases) != 0 {
		t.Errorf("Unexpected amount of releases %d, it should be empty", len(rels.Releases))
	}
}

func TestDeleteMissingHelmRelease(t *testing.T) {
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED", "wordpress", chart.Metadata{
		Version: "1.0.0",
		Icon:    "icon.png",
		Name:    "wordpress",
	}}
	proxy := newFakeProxy([]AppOverview{app})

	err := proxy.DeleteRelease("not_foo", "other_ns", true)
	if err == nil {
		t.Error("Delete should fail, there is not a release in the namespace specified")
	}
	rels, err := proxy.helmClient.ListReleases()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if len(rels.Releases) != 1 {
		t.Errorf("Unexpected amount of releases %d, it contain a release", len(rels.Releases))
	}
}

func TestEnsureThreadSafety(t *testing.T) {
	ns := "myns"
	rs := "foo"
	chartName := "bar"
	version := "v1.0.0"
	ch := &chart.Chart{
		Metadata: &chart.Metadata{Name: chartName, Version: version},
	}
	proxy := newFakeProxy([]AppOverview{})
	finish := make(chan struct{})
	type test func()
	phases := []test{
		func() {
			// Create first element
			result, err := proxy.CreateRelease(rs, ns, "", ch)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if result.Name != rs {
				t.Errorf("Expected release named %s received %s", rs, result.Name)
			}
		},
		func() {
			// Try to create it again
			_, err := proxy.CreateRelease(rs, ns, "", ch)
			if err == nil {
				t.Errorf("Should fail with 'already exists'")
			}
		},
		func() {
			_, err := proxy.UpdateRelease(rs, ns, "", ch)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
		},
		func() {
			err := proxy.DeleteRelease(rs, ns, true)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			finish <- struct{}{}
		},
	}
	for _, phase := range phases {
		// Run all phases in parallel
		go phase()
		// Give minimum time for phase to block
		time.Sleep(1 * time.Millisecond)
	}
	<-finish
}
