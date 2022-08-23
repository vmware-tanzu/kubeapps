// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/pkgutils"
	"github.com/vmware-tanzu/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	k8scache "k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"
)

const (
	// max number of retries due to transient errors
	maxChartCacheRetries = 5
	// number of background workers to process work queue items
	maxChartCacheWorkers = 2
)

var (
	// pretty much a constant, init pattern similar to that of asset-syncer
	verboseChartCacheQueue = os.Getenv("DEBUG_CHART_CACHE_QUEUE") == "true"
)

type ChartCache struct {
	// the redis client
	redisCli *redis.Client

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time and makes it easy to ensure we are never processing the same item
	// simultaneously in different workers.
	queue RateLimitingInterface

	// this is a transient (temporary) store only used to keep track of
	// state (chart url, etc) during the time window between AddRateLimited()
	// is called by the producer and runWorker consumer picks up
	// the corresponding item from the queue. Upon successful processing
	// of the item, the corresponding store entry is deleted
	processing k8scache.Store

	// I am using a Read/Write Mutex to gate access to cache's resync() operation, which is
	// significant in that it flushes the whole redis cache and re-populates the state from k8s.
	// When that happens we don't really want any concurrent access to the cache until the resync()
	// operation is complete. In other words, we want to:
	//  - be able to have multiple concurrent readers (goroutines doing Get())
	//  - only a single writer (goroutine doing a resync()) is allowed, and while its doing its job
	//    no readers are allowed
	resyncCond *sync.Cond

	// bi-directional channel used exclusively by unit tests
	resyncCh chan int
}

type DownloadChartFn func(chartID, chartUrl, chartVersion string) ([]byte, error)

// chartCacheStoreEntry is what we'll be storing in the processing store
// note that url and delete fields are mutually exclusive, you must either:
//  - set url to non-empty string or
//  - deleted flag to true
// setting both for a given entry does not make sense
type chartCacheStoreEntry struct {
	namespace  string
	id         string
	version    string
	url        string
	downloadFn DownloadChartFn
	deleted    bool
}

func NewChartCache(name string, redisCli *redis.Client, stopCh <-chan struct{}) (*ChartCache, error) {
	log.Infof("+NewChartCache(%s, %v)", name, redisCli)

	if redisCli == nil {
		return nil, fmt.Errorf("server not configured with redis client")
	}

	c := ChartCache{
		redisCli:   redisCli,
		queue:      NewRateLimitingQueue(name, verboseChartCacheQueue),
		processing: k8scache.NewStore(chartCacheKeyFunc),
		resyncCond: sync.NewCond(&sync.RWMutex{}),
	}

	// each loop iteration will launch a single worker that processes items on the work
	//  queue as they come in. runWorker will loop until "something bad" happens.
	// The .Until will then rekick the worker after one second
	for i := 0; i < maxChartCacheWorkers; i++ {
		// let's give each worker a unique name - easier to debug
		name := fmt.Sprintf("%s-worker-%d", c.queue.Name(), i)
		fn := func() {
			c.runWorker(name)
		}
		go wait.Until(fn, time.Second, stopCh)
	}

	return &c, nil
}

// this func will enqueue work items into chart work queue and return.
// the charts will be synced worker threads running in the background
func (c *ChartCache) SyncCharts(charts []models.Chart, downloadFn DownloadChartFn) error {
	log.Info("+SyncCharts()")
	totalToSync := 0
	defer func() {
		log.Infof("-SyncCharts(): [%d] total charts to sync", totalToSync)
	}()

	// let's just cache the latest one for now. The chart versions array would
	// have already been sorted and the latest chart version will be at array index 0
	for _, chart := range charts {
		// add chart to temp store. It will be removed when processed by background
		// runWorker/syncHandler
		if len(chart.ChartVersions) == 0 {
			log.Warningf("Skipping chart [%s] due to empty version array", chart.ID)
			continue
		} else if len(chart.ChartVersions[0].URLs) == 0 {
			log.Warningf("Chart: [%s], version: [%s] has no URLs", chart.ID, chart.ChartVersions[0].Version)
			continue
		}

		// The tarball URL will always be the first URL in the repo.chartVersions.
		// So says the helm plugin :-)
		// however, not everybody agrees:
		// ref https://github.com/helm/helm/blob/65d8e72504652e624948f74acbba71c51ac2e342/pkg/downloader/chart_downloader.go#L296
		u, err := url.Parse(chart.ChartVersions[0].URLs[0])
		if err != nil {
			return fmt.Errorf("invalid URL format for chart [%s]: %v", chart.ID, err)
		}

		// If the URL is relative (no scheme), prepend the chart repo's base URL
		// ref https://github.com/vmware-tanzu/kubeapps/issues/4381
		// ref https://github.com/helm/helm/blob/65d8e72504652e624948f74acbba71c51ac2e342/pkg/downloader/chart_downloader.go#L303
		if !u.IsAbs() {
			repoURL, err := url.Parse(chart.Repo.URL)
			if err != nil {
				return fmt.Errorf("invalid URL format for chart repo [%s]: %v", chart.ID, err)
			}
			q := repoURL.Query()
			// We need a trailing slash for ResolveReference to work, but make sure there isn't already one
			repoURL.Path = strings.TrimSuffix(repoURL.Path, "/") + "/"
			u = repoURL.ResolveReference(u)
			u.RawQuery = q.Encode()
		}

		entry := chartCacheStoreEntry{
			namespace:  chart.Repo.Namespace,
			id:         chart.ID,
			version:    chart.ChartVersions[0].Version,
			url:        u.String(),
			downloadFn: downloadFn,
			deleted:    false,
		}
		if key, err := chartCacheKeyFunc(entry); err != nil {
			log.Errorf("Failed to get key for chart due to %+v", err)
		} else {
			err := c.processing.Add(entry)
			if err != nil {
				log.Errorf("Failed to sync chart due to %+v", err)
			}
			c.queue.AddRateLimited(key)
			totalToSync++
		}
	}
	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *ChartCache) runWorker(workerName string) {
	log.Infof("+runWorker(%s)", workerName)
	defer log.Infof("-runWorker(%s)", workerName)

	for c.processNextWorkItem(workerName) {
	}
}

// processNextWorkItem will read a single work item off the work queue and
// attempt to process it, by calling the syncHandler.

// ref: https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html
// ref: https://github.com/bitnami-labs/kubewatch/blob/master/pkg/controller/controller.go
func (c *ChartCache) processNextWorkItem(workerName string) bool {
	log.Infof("+processNextWorkItem(%s)", workerName)
	defer log.Infof("-processNextWorkItem(%s)", workerName)

	obj, shutdown := c.queue.Get()
	if shutdown {
		log.Infof("[%s] shutting down...", workerName)
		return false
	}

	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	// We must remember to call Done so the queue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the queue and attempted again after a back-off
	// period.
	key, ok := obj.(string)
	if !ok {
		c.queue.Done(obj)
		// As the item in the work queue is actually invalid, we call
		// Forget() here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		c.queue.Forget(obj)
		runtime.HandleError(fmt.Errorf("expected string in work queue but got %#v", obj))
		return true
	}
	if !c.queue.IsProcessing(key) {
		// This is the scenario where between the call to .Get() and
		// here there was a resync event, so we can discard this item
		return true
	}
	defer c.queue.Done(obj)
	if err := c.syncHandler(workerName, key); err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
		errDel := c.processing.Delete(key)
		if err != nil {
			runtime.HandleError(fmt.Errorf("error deleting key [%s] due to: %v", key, errDel))
		}
	} else if c.queue.NumRequeues(key) < maxChartCacheRetries {
		log.Errorf("Error processing [%s] (will retry [%d] times): %v",
			key, maxChartCacheRetries-c.queue.NumRequeues(key), err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		errDel := c.processing.Delete(key)
		if err != nil {
			runtime.HandleError(fmt.Errorf("error deleting key [%s] due to: %v", key, errDel))
		}
		runtime.HandleError(fmt.Errorf("error syncing key [%s] due to: %v", key, err))
	}
	return true
}

func (c *ChartCache) DeleteChartsForRepo(repo *types.NamespacedName) error {
	log.Infof("+DeleteChartsForRepo(%s)", repo)
	defer log.Infof("-DeleteChartsForRepo(%s)", repo)

	// need to get a list of all charts/versions for this repo that are either:
	//   a. already in the cache OR
	//   b. being processed

	// this loop should take care of (a)
	// glob-style pattern, you can use https://www.digitalocean.com/community/tools/glob to test
	// also ref. https://stackoverflow.com/questions/4006324/how-to-atomically-delete-keys-matching-a-pattern-using-redis
	match := fmt.Sprintf("helmcharts%s%s%s%s/*%s*",
		KeySegmentsSeparator,
		repo.Namespace,
		KeySegmentsSeparator,
		repo.Name,
		KeySegmentsSeparator)
	redisKeysToDelete := sets.String{}
	// https://redis.io/commands/scan An iteration starts when the cursor is set to 0,
	// and terminates when the cursor returned by the server is 0
	cursor := uint64(0)
	for {
		var keys []string
		var err error
		keys, cursor, err = c.redisCli.Scan(c.redisCli.Context(), cursor, match, 0).Result()
		if err != nil {
			return err
		}
		for _, k := range keys {
			redisKeysToDelete.Insert(k)
		}
		if cursor == 0 {
			break
		}
	}

	// we still need to take care of (b)
	for _, k := range c.processing.ListKeys() {
		if namespace, chartID, _, err := c.fromKey(k); err != nil {
			log.Errorf("%+v", err)
		} else {
			if parts := strings.Split(chartID, "/"); len(parts) != 2 {
				log.Errorf("unexpected chartID format: [%s]", chartID)
			} else if repo.Namespace == namespace && repo.Name == parts[0] {
				redisKeysToDelete.Insert(k)
			}
		}
	}

	for k := range redisKeysToDelete {
		if namespace, chartID, chartVersion, err := c.fromKey(k); err != nil {
			log.Errorf("%+v", err)
		} else {
			entry := chartCacheStoreEntry{
				namespace: namespace,
				id:        chartID,
				version:   chartVersion,
				deleted:   true,
			}
			err = c.processing.Add(entry)
			if err != nil {
				log.Errorf("Failed to delete chart due to %+v", err)
			}
			log.V(4).Infof("Marked key [%s] to be deleted", k)
			c.queue.Add(k)
		}
	}
	return nil
}

func (c *ChartCache) OnResync() error {
	log.Infof("+OnResync(), queue: [%s], size: [%d]", c.queue.Name(), c.queue.Len())
	c.resyncCond.L.Lock()
	defer func() {
		if c.resyncCh != nil {
			close(c.resyncCh)
			c.resyncCh = nil
		}
		c.resyncCond.L.Unlock()
		c.resyncCond.Broadcast()
		log.Info("-OnResync()")
	}()

	if c.resyncCh != nil {
		c.resyncCh <- c.queue.Len()
		// now let's wait for the client (unit test code) that it's ok to proceed
		// to re-build the whole cache. Presumably the client will have set up the
		// right expectations for redis mock. Don't care what the client sends,
		// just need an indication its ok to proceed
		<-c.resyncCh
	}

	log.Infof("Resetting work queue [%s] and store...", c.queue.Name())
	c.queue.Reset()
	c.processing = k8scache.NewStore(chartCacheKeyFunc)
	return nil
}

// this is what we store in the cache for each cached repo
// all struct fields are capitalized so they're exported by gob encoding
type chartCacheEntryValue struct {
	ChartTarball []byte
}

// syncs the current state of the given resource in k8s with that in the cache
func (c *ChartCache) syncHandler(workerName, key string) error {
	log.Infof("+syncHandler(%s, %s)", workerName, key)
	defer log.Infof("-syncHandler(%s, %s)", workerName, key)

	entry, exists, err := c.processing.GetByKey(key)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("no object exists in cache store for key: [%s]", key)
	}

	chart, ok := entry.(chartCacheStoreEntry)
	if !ok {
		return fmt.Errorf("unexpected object in cache store: [%s]", reflect.TypeOf(entry))
	}

	if chart.deleted {
		// TODO: (gfichtenholt) DEL has the capability to delete multiple keys in one
		// atomic operation. Ref https://www.redis.io/commands/del/
		// It would be nice to come up with a way to utilize that here
		// the problem is the queue is designed to work on one item at a time. In other words,
		// each queue entry has to be associated with some unique key. How to do that when
		// you're trying to delete 50 keys at once? One way to solve
		// it *might* be to add a .GetAll() method to RateLimitingInterface,
		// which will be a little tricky to make sure to get the logic right to be atomic and
		// also when *SOME* of the items fail and some succeed
		keysRemoved, _ := c.redisCli.Del(c.redisCli.Context(), key).Result()
		log.Infof("Redis [DEL %s]: %d", key, keysRemoved)
	} else {
		// unlike helm repositories, specific version chart tarball contents never changes
		// so before embarking on expensive operation such as getting chart tarball
		// via HTTP/S, first see if the cache already's got this entry
		if keysExist, err := c.redisCli.Exists(c.redisCli.Context(), key).Result(); err != nil {
			return fmt.Errorf("error checking whether key [%s] exists in redis: %+v", key, err)
		} else {
			log.Infof("Redis [EXISTS %s]: %d", key, keysExist)
			if keysExist == 1 {
				// nothing to do
				return nil
			}
		}
		byteArray, err := ChartCacheComputeValue(chart.id, chart.url, chart.version, chart.downloadFn)
		if err != nil {
			return err
		}
		startTime := time.Now()
		result, err := c.redisCli.Set(c.redisCli.Context(), key, byteArray, 0).Result()
		if err != nil {
			return fmt.Errorf("failed to set value for object with key [%s] in cache due to: %v", key, err)
		} else {
			duration := time.Since(startTime)
			usedMemory, totalMemory := common.RedisMemoryStats(c.redisCli)
			log.Infof("Redis [SET %s]: %s in [%d] ms. Redis [INFO memory]: [%s/%s]",
				key, result, duration.Milliseconds(), usedMemory, totalMemory)
		}
	}
	return err
}

// this is effectively a cache GET operation
func (c *ChartCache) Fetch(key string) ([]byte, error) {
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	log.Infof("+Fetch(%s)", key)

	// read back from cache: should be either:
	//  - what we previously wrote OR
	//  - redis.Nil if the key does  not exist or has been evicted due to memory pressure/TTL expiry
	//
	byteArray, err := c.redisCli.Get(c.redisCli.Context(), key).Bytes()
	// debugging an intermittent issue
	if err == redis.Nil {
		log.Infof("Redis [GET %s]: Nil", key)
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetch() failed to get value for key [%s] from cache due to: %v", key, err)
	}
	log.Infof("Redis [GET %s]: %d bytes read", key, len(byteArray))

	dec := gob.NewDecoder(bytes.NewReader(byteArray))
	var entryValue chartCacheEntryValue
	if err := dec.Decode(&entryValue); err != nil {
		return nil, err
	}
	return entryValue.ChartTarball, nil
}

/*
 Get() is like Fetch() but if there is a cache miss, it will then get chart data based on
 the corresponding repo object, process it and then add it to the cache and return the
 result.
 This func should:

 • return an error if the entry could not be computed due to not being able to read
 repos secretRef.

 • return nil for any invalid chart name.

 • otherwise return the bytes stored in the
 chart cache for the given entry
*/
func (c *ChartCache) Get(key string, chart *models.Chart, downloadFn DownloadChartFn) ([]byte, error) {
	// TODO (gfichtenholt) it'd be nice to get rid of all arguments except for the key, similar to that of
	// NamespacedResourceWatcherCache.Get()
	log.Infof("+Get(%s)", key)
	var value []byte
	var err error
	if value, err = c.Fetch(key); err != nil {
		return nil, err
	} else if value == nil {
		// cache miss
		namespace, chartID, version, err := c.fromKey(key)
		if err != nil {
			return nil, err
		}
		if namespace != chart.Repo.Namespace || chartID != chart.ID {
			return nil, fmt.Errorf("unexpected state for chart with key [%s]", key)
		}
		var entry *chartCacheStoreEntry
		for _, v := range chart.ChartVersions {
			if v.Version == version {
				if len(v.URLs) == 0 {
					log.Warningf("chart: [%s], version: [%s] has no URLs", chart.ID, v.Version)
				} else {
					entry = &chartCacheStoreEntry{
						namespace:  namespace,
						id:         chartID,
						version:    v.Version,
						url:        v.URLs[0],
						downloadFn: downloadFn,
					}
				}
				break
			}
		}
		if entry != nil {
			err = c.processing.Add(*entry)
			if err != nil {
				log.Errorf("Failed to get chart due to %+v", err)
			}
			c.queue.Add(key)
			// now need to wait until this item has been processed by runWorker().
			c.queue.WaitUntilForgotten(key)
			return c.Fetch(key)
		}
	}
	return value, nil
}

func (c *ChartCache) KeyFor(namespace, chartID, chartVersion string) (string, error) {
	return chartCacheKeyFor(namespace, chartID, chartVersion)
}

func (c *ChartCache) String() string {
	return fmt.Sprintf("ChartCache[queue size: [%d]]", c.queue.Len())
}

// the opposite of keyFor
// the goal is to keep the details of what exactly the key looks like localized to one piece of code
func (c *ChartCache) fromKey(key string) (namespace, chartID, chartVersion string, err error) {
	parts := strings.Split(key, KeySegmentsSeparator)
	if len(parts) != 4 || parts[0] != "helmcharts" || len(parts[1]) == 0 || len(parts[2]) == 0 || len(parts[3]) == 0 {
		return "", "", "", status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return parts[1], parts[2], parts[3], nil
}

// this func is used by unit tests only
func (c *ChartCache) ExpectAdd(key string) {
	c.queue.ExpectAdd(key)
}

// this func is used by unit tests only
func (c *ChartCache) WaitUntilForgotten(key string) {
	c.queue.WaitUntilForgotten(key)
}

func (c *ChartCache) Shutdown() {
	c.queue.ShutDown()
}

// this func is used by unit tests only
// returns birectional channel where the number of items in the work queue will be sent
// at the time of the resync() call and guarantees no more work items will be processed
// until resync() finishes
func (c *ChartCache) ExpectResync() (chan int, error) {
	log.Info("+ExpectResync()")
	c.resyncCond.L.Lock()
	defer func() {
		c.resyncCond.L.Unlock()
		log.Info("-ExpectResync()")
	}()

	if c.resyncCh != nil {
		return nil, status.Errorf(codes.Internal, "ExpectSync() already called")
	} else {
		c.resyncCh = make(chan int, 1)
		return c.resyncCh, nil
	}
}

// this func is used by unit tests only
// By the end of the call the work queue should be empty
func (c *ChartCache) WaitUntilResyncComplete() {
	log.Info("+WaitUntilResyncComplete()")
	c.resyncCond.L.Lock()
	defer func() {
		c.resyncCond.L.Unlock()
		log.Info("-WaitUntilResyncComplete()")
	}()

	for c.resyncCh != nil {
		c.resyncCond.Wait()
	}
}

func chartCacheKeyFunc(obj interface{}) (string, error) {
	if entry, ok := obj.(chartCacheStoreEntry); !ok {
		return "", fmt.Errorf("unexpected object in chartCacheKeyFunc: [%s]", reflect.TypeOf(obj))
	} else {
		return chartCacheKeyFor(entry.namespace, entry.id, entry.version)
	}
}

func chartCacheKeyFor(namespace, chartID, chartVersion string) (string, error) {
	if namespace == "" || chartID == "" || chartVersion == "" {
		return "", fmt.Errorf("invalid chart in chartCacheKeyFor: [%s,%s,%s]", namespace, chartID, chartVersion)
	}

	var err error
	if chartID, err = pkgutils.GetUnescapedPackageID(chartID); err != nil {
		return "", fmt.Errorf("invalid chart ID in chartCacheKeyFor: [%s]: %v", chartID, err)
	}

	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmcharts:ns:chartID:chartVersion"
	// notice that chartID is of the form "repoName/id", so it includes the repo name
	return fmt.Sprintf("helmcharts%s%s%s%s%s%s",
		KeySegmentsSeparator,
		namespace,
		KeySegmentsSeparator,
		chartID,
		KeySegmentsSeparator,
		chartVersion), nil
}

// FYI: The work queue is able to retry transient HTTP errors that occur while invoking downloadFn
func ChartCacheComputeValue(chartID, chartUrl, chartVersion string, downloadFn DownloadChartFn) ([]byte, error) {
	chartTgz, err := downloadFn(chartID, chartUrl, chartVersion)
	if err != nil {
		return nil, err
	}

	log.Infof("Successfully fetched details for chart: [%s], version: [%s], url: [%s], details: [%d] bytes",
		chartID, chartVersion, chartUrl, len(chartTgz))

	cacheEntryValue := chartCacheEntryValue{
		ChartTarball: chartTgz,
	}

	// use gob encoding instead of json, it peforms much better
	var gobBuf bytes.Buffer
	enc := gob.NewEncoder(&gobBuf)
	if err := enc.Encode(cacheEntryValue); err != nil {
		return nil, err
	} else {
		return gobBuf.Bytes(), nil
	}
}
