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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/repo"
)

// Fake server for repositories and charts
type fakeHTTPClient struct {
	repoURLs  []string
	chartURLs []string
	index     *repo.IndexFile
}

func (f *fakeHTTPClient) Do(h *http.Request) (*http.Response, error) {
	for _, repoURL := range f.repoURLs {
		if h.URL.String() == fmt.Sprintf("%sindex.yaml", repoURL) {
			// Return fake chart index (not customizable per repo)
			body, err := json.Marshal(*f.index)
			if err != nil {
				fmt.Printf("Error! %v", err)
			}
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader(body))}, nil
		}
	}
	for _, chartURL := range f.chartURLs {
		if h.URL.String() == chartURL {
			// Simulate download time
			time.Sleep(100 * time.Millisecond)
			// Fake chart response
			return &http.Response{StatusCode: 200, Body: ioutil.NopCloser(bytes.NewReader([]byte{}))}, nil
		}
	}
	// Unexpected path
	return &http.Response{StatusCode: 404}, fmt.Errorf("Unexpected path")
}

func fakeLoadChart(in io.Reader) (*chart.Chart, error) {
	// Fake I/O time
	time.Sleep(100 * time.Millisecond)
	return &chart.Chart{}, nil
}

func newFakeProxy(hrs []helmRelease, existingTillerReleases []AppOverview) *Proxy {
	var repoURLs []string
	var chartURLs []string
	entries := map[string]repo.ChartVersions{}
	// Populate Chart registry with content of the given helmReleases
	for _, hr := range hrs {
		repoURLs = append(repoURLs, hr.RepoURL)
		chartMeta := chart.Metadata{Name: hr.ChartName, Version: hr.Version}
		chartURL := fmt.Sprintf("%s%s-%s.tgz", hr.RepoURL, hr.ChartName, hr.Version)
		chartURLs = append(chartURLs, chartURL)
		chartVersion := repo.ChartVersion{Metadata: &chartMeta, URLs: []string{chartURL}}
		chartVersions := []*repo.ChartVersion{&chartVersion}
		entries[hr.ChartName] = chartVersions
	}
	index := &repo.IndexFile{APIVersion: "v1", Generated: time.Now(), Entries: entries}
	netClient := fakeHTTPClient{repoURLs, chartURLs, index}
	helmClient := helm.FakeClient{}
	// Populate Fake helm client with releases
	for _, r := range existingTillerReleases {
		helmClient.Rels = append(helmClient.Rels, &release.Release{
			Name:      r.ReleaseName,
			Namespace: r.Namespace,
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version: r.Version,
				},
			},
		})
	}
	kubeClient := fake.NewSimpleClientset()
	return NewProxy(kubeClient, &helmClient, &netClient, fakeLoadChart)
}

func TestListAllReleases(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns"}
	proxy := newFakeProxy([]helmRelease{}, []AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases("")
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
	app1 := AppOverview{"foo", "1.0.0", "my_ns"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns"}
	proxy := newFakeProxy([]helmRelease{}, []AppOverview{app1, app2})

	// Should return all the releases if no namespace is given
	releases, err := proxy.ListReleases(app1.Namespace)
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

func TestCreateHelmRelease(t *testing.T) {
	ns := "myns"
	h := helmRelease{
		ReleaseName: "not-foo",
		RepoURL:     "http://charts.example.com/repo/",
		ChartName:   "foo",
		Version:     "v1.0.0",
	}
	rawRelease, _ := json.Marshal(h)
	proxy := newFakeProxy([]helmRelease{h}, []AppOverview{})

	result, err := proxy.CreateRelease(ns, rawRelease)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if result.Name != h.ReleaseName {
		t.Errorf("Expected release named %s received %s", h.ReleaseName, result.Name)
	}
	if result.Namespace != ns {
		t.Errorf("Expected release in namespace %s received %s", ns, result.Namespace)
	}
	// We cannot check that the rest of the chart properties as properly set
	// because the fake InstallReleaseFromChart ignores the given chart
}

func TestCreateConflictingHelmRelease(t *testing.T) {
	ns1 := "myns"
	h := helmRelease{
		ReleaseName: "not-foo",
		RepoURL:     "http://charts.example.com/repo/",
		ChartName:   "foo",
		Version:     "v1.0.0",
	}
	ns2 := "other_ns"
	app := AppOverview{h.ReleaseName, h.Version, ns2}
	rawRelease, _ := json.Marshal(h)
	proxy := newFakeProxy([]helmRelease{h}, []AppOverview{app})

	_, err := proxy.CreateRelease(ns1, rawRelease)
	if err == nil {
		t.Error("Release should fail, an existing release in a different namespace already exists")
	}
	if !strings.Contains(err.Error(), "name that is still in use") {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestHelmReleaseUpdated(t *testing.T) {
	ns := "myns"
	h := helmRelease{
		ReleaseName: "not-foo",
		RepoURL:     "http://charts.example.com/repo/",
		ChartName:   "foo",
		Version:     "v1.0.0",
	}
	rawRelease, _ := json.Marshal(h)
	app := AppOverview{h.ReleaseName, h.Version, ns}
	proxy := newFakeProxy([]helmRelease{h}, []AppOverview{app})

	result, err := proxy.UpdateRelease(h.ReleaseName, ns, rawRelease)
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if result.Name != h.ReleaseName {
		t.Errorf("Expected release named %s received %s", h.ReleaseName, result.Name)
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
	h := helmRelease{
		ReleaseName: "not-foo",
		RepoURL:     "http://charts.example.com/repo/",
		ChartName:   "foo",
		Version:     "v1.0.0",
	}
	rawRelease, _ := json.Marshal(h)
	// Simulate the same app but in a different namespace
	ns2 := "other_ns"
	app := AppOverview{h.ReleaseName, h.Version, ns2}
	proxy := newFakeProxy([]helmRelease{h}, []AppOverview{app})

	_, err := proxy.UpdateRelease(h.ReleaseName, ns, rawRelease)
	if err == nil {
		t.Error("Update should fail, there is not a release in the namespace specified")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Unexpected error %v", err)
	}
}

func TestGetHelmRelease(t *testing.T) {
	app1 := AppOverview{"foo", "1.0.0", "my_ns"}
	app2 := AppOverview{"bar", "1.0.0", "other_ns"}
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
		proxy := newFakeProxy([]helmRelease{}, test.existingApps)
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
	app := AppOverview{"foo", "1.0.0", "my_ns"}
	proxy := newFakeProxy([]helmRelease{}, []AppOverview{app})

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
	app := AppOverview{"foo", "1.0.0", "my_ns"}
	proxy := newFakeProxy([]helmRelease{}, []AppOverview{app})

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
	h := helmRelease{
		ReleaseName: "not-foo",
		RepoURL:     "http://charts.example.com/repo/",
		ChartName:   "foo",
		Version:     "v1.0.0",
	}
	rawRelease, _ := json.Marshal(h)
	proxy := newFakeProxy([]helmRelease{h}, []AppOverview{})
	finish := make(chan struct{})
	type test func()
	phases := []test{
		func() {
			// Create first element
			result, err := proxy.CreateRelease(ns, rawRelease)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
			if result.Name != h.ReleaseName {
				t.Errorf("Expected release named %s received %s", h.ReleaseName, result.Name)
			}
		},
		func() {
			// Try to create it again
			_, err := proxy.CreateRelease(ns, rawRelease)
			if err == nil {
				t.Errorf("Should fail with 'already exists'")
			}
		},
		func() {
			_, err := proxy.UpdateRelease(h.ReleaseName, ns, rawRelease)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}
		},
		func() {
			err := proxy.DeleteRelease(h.ReleaseName, ns)
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
