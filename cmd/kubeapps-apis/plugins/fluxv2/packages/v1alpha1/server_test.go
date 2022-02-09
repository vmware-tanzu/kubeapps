// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func TestGetAvailablePackagesStatus(t *testing.T) {
	testCases := []struct {
		name       string
		repo       sourcev1.HelmRepository
		statusCode codes.Code
	}{
		{
			name: "returns without error if response status does not contain conditions",
			repo: newRepo("test", "default",
				&sourcev1.HelmRepositorySpec{
					URL:      "http://example.com",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				},
				nil),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response status does not contain conditions (2)",
			repo: newRepo("test", "default",
				&sourcev1.HelmRepositorySpec{
					URL:      "http://example.com",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				},
				&sourcev1.HelmRepositoryStatus{}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if response does not contain ready repos",
			repo: newRepo("test", "default",
				&sourcev1.HelmRepositorySpec{
					URL:      "http://example.com",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				},
				&sourcev1.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: "False",
							Reason: "IndexationFailed",
						},
					},
				}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain namespace",
			repo: newRepo("test", "",
				nil,
				&sourcev1.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: "True",
							Reason: "IndexationSucceed",
						},
					},
				}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object contains default spec",
			repo: newRepo("test", "default",
				nil,
				&sourcev1.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: "True",
							Reason: "IndexationSucceed",
						},
					},
				}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain spec url",
			repo: newRepo("test", "default",
				&sourcev1.HelmRepositorySpec{},
				&sourcev1.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: "True",
							Reason: "IndexationSucceed",
						},
					},
				}),
			statusCode: codes.OK,
		},
		{
			name: "returns without error if repo object does not contain status url",
			repo: newRepo("test", "default",
				&sourcev1.HelmRepositorySpec{
					URL:      "http://example.com",
					Interval: metav1.Duration{Duration: 1 * time.Minute},
				},
				&sourcev1.HelmRepositoryStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "Ready",
							Status: "True",
							Reason: "IndexationSucceed",
						},
					},
				}),
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, err := newServerWithRepos(t, []sourcev1.HelmRepository{tc.repo}, nil, nil)
			if err != nil {
				t.Fatalf("error instantiating the server: %v", err)
			}

			// these (negative) tests are all testing very unlikely scenarios which were kind of hard to fit into
			// redisMockBeforeCallToGetAvailablePackageSummaries() so I put special logic here instead
			if isRepoReady(tc.repo) {
				if key, err := redisKeyForRepo(tc.repo); err == nil {
					// TODO explain why 3 calls to ExpectGet. It has to do with the way
					// the caching internals work: a call such as GetAvailablePackageSummaries
					// will cause multiple fetches
					mock.ExpectGet(key).RedisNil()
					mock.ExpectGet(key).RedisNil()
					mock.ExpectGet(key).RedisNil()
					mock.ExpectDel(key).SetVal(0)
				}
			}

			response, err := s.GetAvailablePackageSummaries(
				context.Background(),
				&corev1.GetAvailablePackageSummariesRequest{Context: &corev1.Context{}})

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, error: %v, want: %+v", got, err, want)
			} else if got == codes.OK {
				if response == nil || len(response.AvailablePackageSummaries) != 0 {
					t.Fatalf("unexpected response: %v", response)
				}
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("%v", err)
			}
		})
	}
}

func TestParsePluginConfig(t *testing.T) {
	testCases := []struct {
		name                    string
		pluginYAMLConf          []byte
		exp_versions_in_summary pkgutils.VersionsInSummary
		exp_error_str           string
	}{
		{
			name:                    "non existing plugin-config file",
			pluginYAMLConf:          nil,
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 0, Minor: 0, Patch: 0},
			exp_error_str:           "no such file or directory",
		},
		{
			name: "non-default plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 4, Minor: 2, Patch: 1},
			exp_error_str:           "",
		},
		{
			name: "partial params in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 1
        `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{Major: 1, Minor: 0, Patch: 0},
			exp_error_str:           "",
		},
		{
			name: "invalid plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      versionsInSummary:
        major: 4
        minor: 2
        patch: 1-IFC-123
      `),
			exp_versions_in_summary: pkgutils.VersionsInSummary{},
			exp_error_str:           "json: cannot unmarshal",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := ""
			if tc.pluginYAMLConf != nil {
				pluginJSONConf, err := yaml.YAMLToJSON(tc.pluginYAMLConf)
				if err != nil {
					log.Fatalf("%s", err)
				}
				f, err := os.CreateTemp(".", "plugin_json_conf")
				if err != nil {
					log.Fatalf("%s", err)
				}
				defer os.Remove(f.Name()) // clean up
				if _, err := f.Write(pluginJSONConf); err != nil {
					log.Fatalf("%s", err)
				}
				if err := f.Close(); err != nil {
					log.Fatalf("%s", err)
				}
				filename = f.Name()
			}
			versions_in_summary, _, goterr := parsePluginConfig(filename)
			if goterr != nil && !strings.Contains(goterr.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", goterr.Error(), tc.exp_error_str)
			}
			if got, want := versions_in_summary, tc.exp_versions_in_summary; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

func TestParsePluginConfigTimeout(t *testing.T) {
	testCases := []struct {
		name           string
		pluginYAMLConf []byte
		exp_timeout    int32
		exp_error_str  string
	}{
		{
			name:           "no timeout specified in plugin config",
			pluginYAMLConf: nil,
			exp_timeout:    0,
			exp_error_str:  "",
		},
		{
			name: "specific timeout in plugin config",
			pluginYAMLConf: []byte(`
core:
  packages:
    v1alpha1:
      timeoutSeconds: 650
      `),
			exp_timeout:   650,
			exp_error_str: "",
		},
	}
	opts := cmpopts.IgnoreUnexported(pkgutils.VersionsInSummary{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := ""
			if tc.pluginYAMLConf != nil {
				pluginJSONConf, err := yaml.YAMLToJSON(tc.pluginYAMLConf)
				if err != nil {
					log.Fatalf("%s", err)
				}
				f, err := os.CreateTemp(".", "plugin_json_conf")
				if err != nil {
					log.Fatalf("%s", err)
				}
				defer os.Remove(f.Name()) // clean up
				if _, err := f.Write(pluginJSONConf); err != nil {
					log.Fatalf("%s", err)
				}
				if err := f.Close(); err != nil {
					log.Fatalf("%s", err)
				}
				filename = f.Name()
			}
			_, timeoutSeconds, goterr := parsePluginConfig(filename)
			if goterr != nil && !strings.Contains(goterr.Error(), tc.exp_error_str) {
				t.Errorf("err got %q, want to find %q", goterr.Error(), tc.exp_error_str)
			}
			if got, want := timeoutSeconds, tc.exp_timeout; !cmp.Equal(want, got, opts) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, opts))
			}
		})
	}
}

//
// utilities
//
type testSpecChartWithUrl struct {
	chartID       string
	chartRevision string
	chartUrl      string
	opts          *common.ClientOptions
	repoNamespace string
	// this is for a negative test TestTransientHttpFailuresAreRetriedForChartCache
	numRetries int
}

// This func does not create a kubernetes dynamic client. It is meant to work in conjunction with
// a call to fake.NewSimpleDynamicClientWithCustomListKinds. The reason for argument repos
// (unlike charts or releases) is that repos are treated special because
// a new instance of a Server object is only returned once the cache has been synced with indexed repos
func newServer(t *testing.T,
	clientGetter clientgetter.ClientGetterFunc,
	actionConfig *action.Configuration,
	repos []sourcev1.HelmRepository,
	charts []testSpecChartWithUrl) (*Server, redismock.ClientMock, error) {

	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	redisCli, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(false)

	if clientGetter != nil {
		// if client getter returns an error, FLUSHDB call does not take place, because
		// newCacheWithRedisClient() raises an error before redisCli.FlushDB() call
		if _, err := clientGetter(context.TODO(), ""); err == nil {
			mock.ExpectFlushDB().SetVal("OK")
		}
	}

	backgroundClientGetter := func(ctx context.Context) (clientgetter.ClientInterfaces, error) {
		return clientGetter(ctx, KubeappsCluster)
	}

	sink := repoEventSink{
		clientGetter: backgroundClientGetter,
		chartCache:   nil,
	}

	okRepos := sets.String{}
	for _, r := range repos {
		if isRepoReady(r) {
			// we are willfully just logging any errors coming from redisMockSetValueForRepo()
			// here and just skipping over to next repo. This is done for test
			// TestGetAvailablePackagesStatus where we make sure that even if the flux CRD happens
			// to be invalid flux plug in can still operate
			key, _, err := sink.redisMockSetValueForRepo(mock, r)
			if err != nil {
				t.Logf("Skipping repo [%s] due to %+v", key, err)
			} else {
				okRepos.Insert(key)
			}
		}
	}

	var chartCache *cache.ChartCache
	var err error
	cachedChartKeys := sets.String{}
	cachedChartIds := sets.String{}

	if charts != nil {
		chartCache, err = cache.NewChartCache("chartCacheTest", redisCli, stopCh)
		if err != nil {
			return nil, mock, err
		}
		t.Cleanup(func() { chartCache.Shutdown() })

		sink.chartCache = chartCache

		// for now we only cache latest version of each chart
		for _, c := range charts {
			// very simple logic for now, relies on the order of elements in the array
			// to pick the latest version.
			if !cachedChartIds.Has(c.chartID) {
				cachedChartIds.Insert(c.chartID)
				key, err := chartCache.KeyFor(c.repoNamespace, c.chartID, c.chartRevision)
				if err != nil {
					return nil, mock, err
				}
				repoName := types.NamespacedName{
					Name:      strings.Split(c.chartID, "/")[0],
					Namespace: c.repoNamespace}

				repoKey, err := redisKeyForRepoNamespacedName(repoName)
				if err == nil && okRepos.Has(repoKey) {
					for i := 0; i < c.numRetries; i++ {
						mock.ExpectExists(key).SetVal(0)
					}
					err = sink.redisMockSetValueForChart(mock, key, c.chartUrl, c.opts)
					if err != nil {
						return nil, mock, err
					}
					cachedChartKeys.Insert(key)
					chartCache.ExpectAdd(key)
				}
			}
		}
	}

	cacheConfig := cache.NamespacedResourceWatcherCacheConfig{
		Gvr:          repositoriesGvr,
		ClientGetter: backgroundClientGetter,
		OnAddFunc:    sink.onAddRepo,
		OnModifyFunc: sink.onModifyRepo,
		OnGetFunc:    sink.onGetRepo,
		OnDeleteFunc: sink.onDeleteRepo,
		OnResyncFunc: sink.onResync,
		NewObjFunc:   func() ctrlclient.Object { return &sourcev1.HelmRepository{} },
		NewListFunc:  func() ctrlclient.ObjectList { return &sourcev1.HelmRepositoryList{} },
		ListItemsFunc: func(ol ctrlclient.ObjectList) []ctrlclient.Object {
			if hl, ok := ol.(*sourcev1.HelmRepositoryList); !ok {
				t.Fatalf("Expected: *sourcev1.HelmRepositoryList, got: %s", reflect.TypeOf(ol))
				return nil
			} else {
				ret := make([]ctrlclient.Object, len(hl.Items))
				for i, hr := range hl.Items {
					ret[i] = hr.DeepCopy()
				}
				return ret
			}
		},
	}

	repoCache, err := cache.NewNamespacedResourceWatcherCache(
		"repoCacheTest", cacheConfig, redisCli, stopCh)
	if err != nil {
		return nil, mock, err
	}
	t.Cleanup(func() { repoCache.Shutdown() })

	// need to wait until ChartCache has finished syncing
	for key := range cachedChartKeys {
		chartCache.WaitUntilForgotten(key)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		return nil, mock, err
	}

	s := &Server{
		clientGetter: clientGetter,
		actionConfigGetter: func(context.Context, string) (*action.Configuration, error) {
			return actionConfig, nil
		},
		repoCache:         repoCache,
		chartCache:        chartCache,
		kubeappsCluster:   KubeappsCluster,
		versionsInSummary: pkgutils.GetDefaultVersionsInSummary(),
	}
	return s, mock, nil
}
