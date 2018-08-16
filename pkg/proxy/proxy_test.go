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
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
)

func newFakeProxy(existingTillerReleases []AppOverview) *Proxy {
	helmClient := helm.FakeClient{}
	// Populate Fake helm client with releases
	for _, r := range existingTillerReleases {
		status := release.Status_DEPLOYED
		if r.Status == "DELETED" {
			status = release.Status_DELETED
		}
		helmClient.Rels = append(helmClient.Rels, &release.Release{
			Name:      r.ReleaseName,
			Namespace: r.Namespace,
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version: r.Version,
					Icon:    r.Icon,
				},
			},
			Info: &release.Info{
				Status: &release.Status{
					Code: status,
				},
			},
		})
	}
	kubeClient := fake.NewSimpleClientset()
	return NewProxy(kubeClient, &helmClient)
}

func TestListAllReleases(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED"}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases("", 256)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	if len(releases) != 2 {
		t.Errorf("It should return both releases")
	}
	if !reflect.DeepEqual([]AppOverview{app1, app2}, releases) {
		t.Errorf("Unexpected list of releases %v", releases)
	}
}

func TestListNamespacedRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED"}
	proxy := newFakeProxy([]AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases(app1.Namespace, 256)
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
	app := AppOverview{rs, version, ns2, "icon.png", "DEPLOYED"}
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
	app := AppOverview{rs, version, ns, "icon.png", "DEPLOYED"}
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
	// Simulate the same app but in a different namespace
	ns2 := "other_ns"
	app := AppOverview{rs, version, ns2, "icon.png", "DEPLOYED"}
	proxy := newFakeProxy([]AppOverview{app})

	_, err := proxy.UpdateRelease(rs, ns, "", ch)
	if err == nil {
		t.Error("Update should fail, there is not a release in the namespace specified")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestGetHelmRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns", "icon2.png", "DELETED"}
	type testStruct struct {
		existingApps    []AppOverview
		shouldFail      bool
		targetApp       string
		tartegNamespace string
		expectedResult  string
	}
	tests := []testStruct{
		{[]AppOverview{app1, app2}, false, "foo", "my_ns", "foo"},
		{[]AppOverview{app1, app2}, true, "bar", "my_ns", ""},
		{[]AppOverview{app1, app2}, true, "foobar", "my_ns", ""},
		{[]AppOverview{app1, app2}, false, "foo", "", "foo"},
	}
	for _, test := range tests {
		proxy := newFakeProxy(test.existingApps)
		res, err := proxy.GetRelease(test.targetApp, test.tartegNamespace)
		if test.shouldFail && err == nil {
			t.Errorf("Get %s/%s should fail", test.tartegNamespace, test.targetApp)
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
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED"}
	proxy := newFakeProxy([]AppOverview{app})

	err := proxy.DeleteRelease(app.ReleaseName, app.Namespace)
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
	app := AppOverview{"foo", "1.0.0", "my_ns", "icon.png", "DEPLOYED"}
	proxy := newFakeProxy([]AppOverview{app})

	err := proxy.DeleteRelease(app.ReleaseName, "other_ns")
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
			err := proxy.DeleteRelease(rs, ns)
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
