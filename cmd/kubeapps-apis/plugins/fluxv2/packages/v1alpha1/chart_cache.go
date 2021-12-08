/*
Copyright © 2021 VMware
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
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/getter"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	k8scache "k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"
)

const (
	// copied from helm plug-in
	UserAgentPrefix = "kubeapps-apis/plugins"
)

// For now:
// unlike NamespacedResourceWatcherCache this is not a general purpose cache meant to
// be re-used. It is written specifically for one purpose of caching helm chart details
// and has ties into the internals of repo and chart.go. So, it exists outside the cache
// package
// TODO (gfichtenholt): clean this up if possible, and move into cache package

type ChartCache struct {
	redisCli *redis.Client

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time and makes it easy to ensure we are never processing the same item
	// simultaneously in different workers.
	queue cache.RateLimitingInterface

	// this is a transient (temporary) store only used to keep track of
	// state (chart url, etc) during the time window between AddRateLimited()
	// is called by the producer and runWorker consumer picks up
	// the corresponding item from the queue. Upon successful processing
	// of the item, the corresponding store entry is deleted
	processing k8scache.Store
}

// chartCacheStoreEntry is what we'll be storing in the processing store
// note that url and delete fields are mutually exclusive, you must either:
//  - set url to non-empty string or
//  - deleted flag to true
// setting both does not make sense
type chartCacheStoreEntry struct {
	namespace     string
	id            string
	version       string
	url           string
	clientOptions []getter.Option
	deleted       bool
}

func NewChartCache(redisCli *redis.Client) (*ChartCache, error) {
	log.Infof("+NewChartCache(%v)", redisCli)

	if redisCli == nil {
		return nil, fmt.Errorf("server not configured with redis client")
	}

	c := ChartCache{
		redisCli:   redisCli,
		queue:      cache.NewRateLimitingQueue(),
		processing: k8scache.NewStore(chartCacheKeyFunc),
	}

	// dummy channel for now. Ideally, this would be passed in as an input argument
	// and the caller would indicate when to stop
	stopCh := make(chan struct{})

	// this will launch a single worker that processes items on the work queue as they come in
	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	// TODO (gfichtenholt) we should have multiple workers
	go wait.Until(c.runWorker, time.Second, stopCh)

	return &c, nil
}

// this will enqueue work items into chart work queue and return.
// the charts will be synced by a worker thread running in the background
func (c *ChartCache) syncCharts(charts []models.Chart, clientOptions []getter.Option) error {
	log.Infof("+syncCharts()")
	totalToSync := 0
	defer func() {
		log.Infof("-syncCharts(): [%d] total charts to sync", totalToSync)
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
		entry := chartCacheStoreEntry{
			namespace:     chart.Repo.Namespace,
			id:            chart.ID,
			version:       chart.ChartVersions[0].Version,
			url:           chart.ChartVersions[0].URLs[0],
			clientOptions: clientOptions,
			deleted:       false,
		}
		c.processing.Add(entry)
		if key, err := chartCacheKeyFunc(entry); err != nil {
			log.Errorf("Failed to get key for chart due to %+v", err)
		} else {
			c.queue.AddRateLimited(key)
			totalToSync++
		}
	}
	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *ChartCache) runWorker() {
	log.Infof("+runWorker()")
	defer log.Infof("-runWorker()")

	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the work queue and
// attempt to process it, by calling the syncHandler.
func (c *ChartCache) processNextWorkItem() bool {
	log.V(4).Infof("+processNextWorkItem()")
	defer log.V(4).Infof("-processNextWorkItem()")

	obj, shutdown := c.queue.Get()
	if shutdown {
		log.Info("Shutting down...")
		return false
	}

	// We wrap this block in a func so we can defer c.queue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the queue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the queue and attempted again after a back-off
		// period.
		defer c.queue.Done(obj)
		// We expect strings to come off the work queue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// work queue means the items in k8s may actually be more up to date
		// that when the item was initially put onto the work queue.
		if key, ok := obj.(string); !ok {
			// As the item in the work queue is actually invalid, we call
			// Forget() here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.queue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in work queue but got %#v", obj))
			return nil
		} else {
			// Run the syncHandler, passing it the namespace/name string of the
			// resource to be synced.
			if err := c.syncHandler(key); err != nil {
				return fmt.Errorf("error syncing key [%s] due to: %v", key, err)
			}
			// Finally, if no error occurs we Forget this item so it does not
			// get queued again until another change happens.
			c.queue.Forget(obj)
			log.V(4).Infof("Done syncing key [%s]", key)
			return nil
		}
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}
	return true
}

func (c *ChartCache) deleteChartsForRepo(key string) error {
	log.Infof("+deleteChartsFor(%s)", key)
	defer log.Infof("-deleteChartsFor(%s)", key)

	// need to get a list of all charts/versions for this repo that are either:
	//   a. already in the cache OR
	//   b. being processed

	// TODO HACK this is a copied from NamespacedResourceCache.fromKey
	parts := strings.Split(key, ":")
	if len(parts) != 3 || parts[0] != fluxHelmRepositories || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	repo := &types.NamespacedName{Namespace: parts[1], Name: parts[2]}
	redisKeysToDelete := sets.String{}

	// this loop should take care of (a)
	// glob-style pattern, you can use https://www.digitalocean.com/community/tools/glob to test
	// also ref. https://stackoverflow.com/questions/4006324/how-to-atomically-delete-keys-matching-a-pattern-using-redis
	match := fmt.Sprintf("helmcharts:%s:%s/*:*", repo.Namespace, repo.Name)
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
			parts := strings.Split(chartID, "/")
			if repo.Namespace == namespace && repo.Name == parts[0] {
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
			c.processing.Add(entry)
			log.V(4).Infof("Marked key [%s] to be deleted", k)
			c.queue.Add(k)
		}
	}
	return nil
}

// this is what we store in the cache for each cached repo
// all struct fields are capitalized so they're exported by gob encoding
type chartCacheEntryValue struct {
	ChartTarball []byte
}

// syncs the current state of the given resource in k8s with that in the cache
func (c *ChartCache) syncHandler(key string) (err error) {
	log.Infof("+syncHandler(%s)", key)
	defer log.Infof("-syncHandler(%s)", key)

	var entry interface{}
	var exists bool
	entry, exists, err = c.processing.GetByKey(key)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("no object exists in cache store for key: [%s]", key)
	}

	// we need to keep track whether or not this func returns an error,
	// because if it does, we still need the entry in store for when retry kicks in
	defer func() {
		if err != nil {
			c.processing.Delete(entry)
		}
	}()

	chart, ok := entry.(chartCacheStoreEntry)
	if !ok {
		err = fmt.Errorf("unexpected object in cache store: [%s]", reflect.TypeOf(entry))
		return err
	}

	if chart.deleted {
		// TODO: (gfichtenholt) DEL has the capability to delete multiple keys in one
		// atomic operation. It would be nice to come up with a way to utilize that here
		var keysremoved int64
		keysremoved, err = c.redisCli.Del(c.redisCli.Context(), key).Result()
		log.Infof("Redis [DEL %s]: %d", key, keysremoved)
	} else {
		var byteArray []byte
		byteArray, err = chartCacheComputeValue(chart.id, chart.url, chart.version, chart.clientOptions)
		if err == nil {
			var result string
			startTime := time.Now()
			result, err = c.redisCli.Set(c.redisCli.Context(), key, byteArray, 0).Result()
			if err != nil {
				err = fmt.Errorf("failed to set value for object with key [%s] in cache due to: %v", key, err)
			} else {
				duration := time.Since(startTime)
				usedMemory, totalMemory := common.RedisMemoryStats(c.redisCli)
				log.Infof("Redis [SET %s]: %s in [%d] ms. Redis [INFO memory]: [%s/%s]",
					key, result, duration.Milliseconds(), usedMemory, totalMemory)
			}
		}
	}
	return err
}

// this is effectively a cache GET operation
func (c *ChartCache) fetchForOne(key string) ([]byte, error) {
	log.Infof("+fetchForOne(%s)", key)
	// read back from cache: should be either:
	//  - what we previously wrote OR
	//  - redis.Nil if the key does  not exist or has been evicted due to memory pressure/TTL expiry
	//
	byteArray, err := c.redisCli.Get(c.redisCli.Context(), key).Bytes()
	// debugging an intermittent issue
	if err == redis.Nil {
		log.V(4).Infof("Redis [GET %s]: Nil", key)
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetchForOne() failed to get value for key [%s] from cache due to: %v", key, err)
	}
	log.V(4).Infof("Redis [GET %s]: %d bytes read", key, len(byteArray))

	dec := gob.NewDecoder(bytes.NewReader(byteArray))
	var entryValue chartCacheEntryValue
	if err := dec.Decode(&entryValue); err != nil {
		return nil, err
	}
	return entryValue.ChartTarball, nil
}

// GetForOne() is like fetchForOne() but if there is a cache miss, it will also check the
// k8s for the corresponding object, process it and then add it to the cache and return the
// result.
// TODO (gfichtenholt) get rid of chartModel models.Chart argument. This is breaking an abstraction
// TODO (gfichtenholt) I promised Michael this would return an error if the entry could not be
// computed due to not being able to read repo's secretRef. This is actually hard to do due to async
// nature of how entries are added to the cache. Currently, it returns nil, same
// as any invalid chart name
func (c *ChartCache) GetForOne(key string, chart *models.Chart) ([]byte, error) {
	log.Infof("+GetForOne(%s)", key)
	var value []byte
	var err error
	if value, err = c.fetchForOne(key); err != nil {
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
						namespace: namespace,
						id:        chartID,
						version:   v.Version,
						url:       v.URLs[0],
					}
				}
				break
			}
		}
		if entry != nil {
			c.processing.Add(*entry)
			c.queue.Add(key)
			// now need to wait until this item has been processed by runWorker().
			c.queue.WaitUntilDoneWith(key)
			return c.fetchForOne(key)
		}
	}
	return value, nil
}

func (c *ChartCache) keyFor(namespace, chartID, chartVersion string) (string, error) {
	return chartCacheKeyFor(namespace, chartID, chartVersion)
}

func (c *ChartCache) String() string {
	return fmt.Sprintf("ChartCache[queue size: [%d]", c.queue.Len())
}

// the opposite of keyFor
// the goal is to keep the details of what exactly the key looks like localized to one piece of code
func (c *ChartCache) fromKey(key string) (namespace, chartID, chartVersion string, err error) {
	parts := strings.Split(key, ":")
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
func (c *ChartCache) WaitUntilDoneWith(key string) {
	c.queue.WaitUntilDoneWith(key)
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
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmcharts:ns:chartID:chartVersion"
	// notice that chartID is of the form "repoName/id", so it includes the repo name
	return fmt.Sprintf("helmcharts:%s:%s:%s", namespace, chartID, chartVersion), nil
}

// The work queue should be able to retry transient HTTP errors
// so I shouldn't have to do retries here.
// TODO (gfichtenholt) verify this is actually the case. Gasp! it appears NOT to be the
// case:
//	I1207 03:25:03.694088       1 chart_cache.go:337] +syncHandler(helmcharts:default:bitnami-0/wavefront-hpa-adapter:1.0.4)
//	I1207 03:25:25.494913       1 chart_cache.go:386] -syncHandler(helmcharts:default:bitnami-0/wavefront-hpa-adapter:1.0.4)
//	E1207 03:25:25.495069       1 chart_cache.go:256] error syncing key [helmcharts:default:bitnami-0/wavefront-hpa-adapter:1.0.4] due to: Get "https://charts.bitnami.com/bitnami/wavefront-hpa-adapter-1.0.4.tgz": EOF
// which breaks a fundamental assumption I rely on.
// I am putting the retries back in here just as a workaround until I figure out why the queue is broken

func chartCacheComputeValue(chartID, chartUrl, chartVersion string, clientOptions []getter.Option) ([]byte, error) {
	// TODO (gfichtenholt) what about HTTP_PROXY

	// userAgent string is always the same for all requests
	userAgentString := fmt.Sprintf("%s/%s/%s/%s", UserAgentPrefix, pluginDetail.Name, pluginDetail.Version, version)
	// I wish I could use the code in pkg/chart/chart.go and pkg/kube_utils/kube_utils.go
	// InitHTTPClient(), etc. but alas, it's all built around AppRepository CRD, which I don't have.
	// *Sigh*
	clientOptions = append(clientOptions, getter.WithUserAgent(userAgentString))
	clientOptions = append(clientOptions, getter.WithURL(chartUrl))

	// In theory, the work queue should be able to retry transient errors
	// so I shouldn't have to do retries here. See above comment for explanation
	const maxRetries = 5
	var buf *bytes.Buffer
	ok := false
	for i := 0; i < maxRetries && !ok; i++ {
		getter, err := getter.NewHTTPGetter(clientOptions...)
		if err == nil {
			if buf, err = getter.Get(chartUrl); err != nil {
				log.Warningf("%+v", err)
			} else if buf != nil && buf.Len() > 0 {
				ok = true
				break
			}
		} else {
			log.Warningf("%+v", err)
		}
		waitTime := math.Pow(2, float64(i))
		log.Infof("Waiting [%d] seconds before retry...", int(waitTime))
		time.Sleep(time.Duration(waitTime) * time.Second)
	}

	if !ok {
		return nil, status.Errorf(codes.Internal, "failed to retrieve chart [%s] from [%s]", chartID, chartUrl)
	}

	log.Infof("Successfully fetched details for chart: [%s], version: [%s], url: [%s], details: [%d] bytes",
		chartID, chartVersion, chartUrl, buf.Len())

	cacheEntryValue := chartCacheEntryValue{
		ChartTarball: buf.Bytes(),
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
