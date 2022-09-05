// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	fluxmeta "github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/go-redis/redis/v8"
	redismock "github.com/go-redis/redismock/v8"
	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	kubefake "helm.sh/helm/v3/pkg/kube/fake"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	authorizationv1 "k8s.io/api/authorization/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/storage/names"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
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
							Type:   fluxmeta.ReadyCondition,
							Status: metav1.ConditionFalse,
							Reason: fluxmeta.FailedReason,
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
							Type:   fluxmeta.ReadyCondition,
							Status: metav1.ConditionTrue,
							Reason: fluxmeta.SucceededReason,
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
							Type:   fluxmeta.ReadyCondition,
							Status: metav1.ConditionTrue,
							Reason: fluxmeta.SucceededReason,
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
							Type:   fluxmeta.ReadyCondition,
							Status: metav1.ConditionTrue,
							Reason: fluxmeta.SucceededReason,
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
							Type:   fluxmeta.ReadyCondition,
							Status: metav1.ConditionTrue,
							Reason: fluxmeta.SucceededReason,
						},
					},
				}),
			statusCode: codes.OK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, mock, err := newSimpleServerWithRepos(t, []sourcev1.HelmRepository{tc.repo})
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

// utilities
type testSpecChartWithUrl struct {
	chartID       string
	chartRevision string
	chartUrl      string
	// only applicable for HTTP charts
	opts          *common.HttpClientOptions
	repoNamespace string
	// this is for a negative test TestTransientHttpFailuresAreRetriedForChartCache
	numRetries int
}

func newSimpleServerWithRepos(t *testing.T, repos []sourcev1.HelmRepository) (*Server, redismock.ClientMock, error) {
	return newServerWithRepos(t, repos, nil, nil)
}

func newServerWithRepos(t *testing.T, repos []sourcev1.HelmRepository, charts []testSpecChartWithUrl, secrets []runtime.Object) (*Server, redismock.ClientMock, error) {
	typedClient := typfake.NewSimpleClientset(secrets...)

	// ref https://stackoverflow.com/questions/68794562/kubernetes-fake-client-doesnt-handle-generatename-in-objectmeta/68794563#68794563
	typedClient.PrependReactor(
		"create", "*",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			ret = action.(k8stesting.CreateAction).GetObject()
			meta, ok := ret.(metav1.Object)
			if !ok {
				return
			}
			if meta.GetName() == "" && meta.GetGenerateName() != "" {
				meta.SetName(names.SimpleNameGenerator.GenerateName(meta.GetGenerateName()))
			}
			return
		})

	// Creating an authorized clientGetter
	typedClient.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})

	apiextIfc := apiextfake.NewSimpleClientset(fluxHelmRepositoryCRD)
	ctrlClient := newCtrlClient(repos, nil, nil)
	clientGetter := func(context.Context, string) (clientgetter.ClientInterfaces, error) {
		return clientgetter.
			NewBuilder().
			WithTyped(typedClient).
			WithApiExt(apiextIfc).
			WithControllerRuntime(&ctrlClient).
			Build(), nil
	}
	return newServer(t, clientGetter, nil, repos, charts)
}

func newServerWithChartsAndReleases(t *testing.T, actionConfig *action.Configuration, charts []sourcev1.HelmChart, releases []helmv2.HelmRelease) (*Server, redismock.ClientMock, error) {
	typedClient := typfake.NewSimpleClientset()
	// Creating an authorized clientGetter
	typedClient.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})

	apiextIfc := apiextfake.NewSimpleClientset(fluxHelmRepositoryCRD)
	ctrlClient := newCtrlClient(nil, charts, releases)
	clientGetter := func(context.Context, string) (clientgetter.ClientInterfaces, error) {
		return clientgetter.
			NewBuilder().
			WithApiExt(apiextIfc).
			WithTyped(typedClient).
			WithControllerRuntime(&ctrlClient).
			Build(), nil
	}
	return newServer(t, clientGetter, actionConfig, nil, nil)
}

// newHelmActionConfig returns an action.Configuration with fake clients and memory storage.
func newHelmActionConfig(t *testing.T, namespace string, rels []helmReleaseStub) *action.Configuration {
	t.Helper()

	memDriver := driver.NewMemory()

	actionConfig := &action.Configuration{
		Releases:     storage.Init(memDriver),
		KubeClient:   &kubefake.FailingKubeClient{PrintingKubeClient: kubefake.PrintingKubeClient{Out: io.Discard}},
		Capabilities: chartutil.DefaultCapabilities,
		Log: func(format string, v ...interface{}) {
			t.Helper()
			t.Logf(format, v...)
		},
	}

	for _, r := range rels {
		config := map[string]interface{}{}
		rel := &release.Release{
			Name:      r.name,
			Namespace: r.namespace,
			Info: &release.Info{
				Status: r.status,
				Notes:  r.notes,
			},
			Chart: &chart.Chart{
				Metadata: &chart.Metadata{
					Version:    r.chartVersion,
					Icon:       "https://example.com/icon.png",
					AppVersion: "1.2.3",
				},
			},
			Config:   config,
			Manifest: r.manifest,
		}
		err := actionConfig.Releases.Create(rel)
		if err != nil {
			t.Fatal(err)
		}
	}
	// It is the namespace of the driver which determines the results. In the prod code,
	// the actionConfigGetter sets this using StorageForSecrets(namespace, clientset).
	memDriver.SetNamespace(namespace)

	return actionConfig
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

	okRepos := seedRepoCacheWithRepos(t, mock, sink, repos)

	chartCache, waitTilChartCacheSyncComplete, err :=
		seedChartCacheWithCharts(t, redisCli, mock, sink, stopCh, okRepos, charts)
	if err != nil {
		return nil, mock, err
	} else {
		sink.chartCache = chartCache
	}

	cacheConfig := cache.NamespacedResourceWatcherCacheConfig{
		Gvr:          common.GetRepositoriesGvr(),
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
		"repoCacheTest", cacheConfig, redisCli, stopCh, true)
	if err != nil {
		return nil, mock, err
	}
	t.Cleanup(func() { repoCache.Shutdown() })

	// need to wait until repoCache has finished syncing
	repoCache.WaitUntilResyncComplete()

	// need to wait until chartCache has finished syncing
	waitTilChartCacheSyncComplete()

	if err := mock.ExpectationsWereMet(); err != nil {
		return nil, mock, err
	}

	s := &Server{
		clientGetter:               clientGetter,
		serviceAccountClientGetter: backgroundClientGetter,
		actionConfigGetter: func(context.Context, string) (*action.Configuration, error) {
			return actionConfig, nil
		},
		repoCache:       repoCache,
		chartCache:      chartCache,
		kubeappsCluster: KubeappsCluster,
		pluginConfig:    common.NewDefaultPluginConfig(),
	}
	return s, mock, nil
}

func seedRepoCacheWithRepos(t *testing.T, mock redismock.ClientMock, sink repoEventSink, repos []sourcev1.HelmRepository) map[string]sourcev1.HelmRepository {
	okRepos := make(map[string]sourcev1.HelmRepository)
	for _, r := range repos {
		key, err := redisKeyForRepo(r)
		if err != nil {
			t.Logf("Skipping repo [%s] due to %+v", key, err)
			continue
		}
		if isRepoReady(r) {
			// we are willfully just logging any errors coming from redisMockSetValueForRepo()
			// here and just skipping over to next repo. This is done for test
			// TestGetAvailablePackagesStatus where we make sure that even if the flux CRD happens
			// to be invalid flux plug in can still operate
			_, _, err = sink.redisMockSetValueForRepo(mock, r, nil)
			if err != nil {
				t.Logf("Skipping repo [%s] due to %+v", key, err)
			} else {
				okRepos[key] = r
			}
		} else {
			mock.ExpectGet(key).RedisNil()
		}
	}
	return okRepos
}

func seedChartCacheWithCharts(t *testing.T, redisCli *redis.Client, mock redismock.ClientMock, sink repoEventSink, stopCh <-chan struct{}, repos map[string]sourcev1.HelmRepository, charts []testSpecChartWithUrl) (*cache.ChartCache, func(), error) {
	t.Logf("+seedChartCacheWithCharts(%v)", charts)

	var chartCache *cache.ChartCache
	var err error
	cachedChartKeys := sets.String{}
	cachedChartIds := sets.String{}

	if charts != nil {
		chartCache, err = cache.NewChartCache("chartCacheTest", redisCli, stopCh)
		if err != nil {
			return nil, nil, err
		}
		t.Cleanup(func() { chartCache.Shutdown() })

		// for now we only cache latest version of each chart
		for _, c := range charts {
			// very simple logic for now, relies on the order of elements in the array
			// to pick the latest version.
			if !cachedChartIds.Has(c.chartID) {
				cachedChartIds.Insert(c.chartID)
				key, err := chartCache.KeyFor(c.repoNamespace, c.chartID, c.chartRevision)
				if err != nil {
					return nil, nil, err
				}
				repoName := types.NamespacedName{
					Name:      strings.Split(c.chartID, "/")[0],
					Namespace: c.repoNamespace}

				repoKey, err := redisKeyForRepoNamespacedName(repoName)
				if err == nil {
					if r, ok := repos[repoKey]; ok {
						for i := 0; i < c.numRetries; i++ {
							mock.ExpectExists(key).SetVal(0)
						}
						isOci := strings.HasPrefix(c.chartUrl, "oci://")
						if isOci {
							ociChartRepo, err := sink.newOCIChartRepositoryAndLogin(context.Background(), r)
							if err != nil {
								return nil, nil, err
							}
							err = redisMockSetValueForOciChart(mock, key, c.chartUrl, ociChartRepo)
							if err != nil {
								return nil, nil, err
							}
						} else {
							err = redisMockSetValueForHttpChart(mock, key, c.chartUrl, c.opts)
							if err != nil {
								return nil, nil, err
							}
						}
					}
					cachedChartKeys.Insert(key)
					chartCache.ExpectAdd(key)
				}
			}
		}
	}

	waitTilFn := func() {
		for key := range cachedChartKeys {
			chartCache.WaitUntilForgotten(key)
		}
	}

	return chartCache, waitTilFn, err
}
