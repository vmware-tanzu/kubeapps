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
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	watchutil "k8s.io/client-go/tools/watch"
	log "k8s.io/klog/v2"
)

// a type of cache that is based on watching for changes to specified kubernetes resources.
// The resource is assumed to be namespace-scoped. Cluster-wide resources are not
// supported at this time
type NamespacedResourceWatcherCache struct {
	// these expected to be provided by the caller when creating new cache
	config   cacheConfig
	redisCli *redis.Client
	// this WaitGroup is used exclusively by unit tests to block until all expected objects have
	// been 'processed' by the go routine running in the background. The creation of the WaitGroup object
	// and to call to .Add() is expected to be done by the unit test client. The server-side only signals
	// .Done() when processing one object is complete
	eventProcessedWaitGroup *sync.WaitGroup
}

type cacheValueSetter func(string, map[string]interface{}) (interface{}, bool, error)
type cacheValueGetter func(string, interface{}) (interface{}, error)
type cacheValueDeleter func(string, map[string]interface{}) (bool, error)

// TODO (gfichtenholt) rename this to just Config when caching is separated out into core server
// and/or caching-rleated code is moved into a separate package?
type cacheConfig struct {
	gvr schema.GroupVersionResource
	// this clientGetter is for running out-of-request interactions with the Kubernetes API server,
	// such as watching for resource changes
	clientGetter clientGetter
	// 'onAdd' and 'onModify' hooks are called when a new or modified object comes about and
	// allows the plug-in to return information about WHETHER OR NOT and WHAT is to be stored
	// in the cache for a given k8s object (passed in as a untyped/unstructured map)
	// the list of types actually supported be redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	onAdd    cacheValueSetter
	onModify cacheValueSetter
	// the semantics of 'onGet' hook is to convert or "reverse engineer" what was previously
	// stored in the cache (via onAdd/onModify hooks) to an object that the plug-in understands
	// and wishes to be returned as part of response to fetchCachedObjects() call
	onGet cacheValueGetter
	// onDelete hook is called on the plug-in when the corresponding object is deleted in k8s cluster
	onDelete cacheValueDeleter
}

func newCache(config cacheConfig) (*NamespacedResourceWatcherCache, error) {
	log.Infof("+newCache(%v)", config.gvr)
	// TODO (gfichtenholt) small preference for reading all config in the main.go
	// (whether from env vars or cmd-line options) only in the one spot and passing
	// explicitly to functions (so functions are less dependent on env state).
	REDIS_ADDR, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, "missing environment variable REDIS_ADDR")
	}
	REDIS_PASSWORD, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, "missing environment variable REDIS_PASSWORD")
	}
	REDIS_DB, ok := os.LookupEnv("REDIS_DB")
	if !ok {
		return nil, status.Errorf(codes.FailedPrecondition, "missing environment variable REDIS_DB")
	}

	REDIS_DB_NUM, err := strconv.Atoi(REDIS_DB)
	if err != nil {
		return nil, err
	}

	return newCacheWithRedisClient(
		config,
		redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Password: REDIS_PASSWORD,
			DB:       REDIS_DB_NUM,
		}),
		nil)
}

func newCacheWithRedisClient(config cacheConfig, redisCli *redis.Client, waitGroup *sync.WaitGroup) (*NamespacedResourceWatcherCache, error) {
	log.Infof("+newCacheWithRedisClient(%v)", redisCli)

	if redisCli == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with redis Client")
	} else if config.clientGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with clientGetter")
	} else if config.onAdd == nil || config.onModify == nil || config.onDelete == nil || config.onGet == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with expected cache hooks")
	}

	// sanity check that the redis client is connected
	if pong, err := redisCli.Ping(redisCli.Context()).Result(); err != nil {
		return nil, err
	} else {
		log.Infof("Redis [PING] -> [%s]", pong)
	}

	if maxmemory, err := redisCli.ConfigGet(redisCli.Context(), "maxmemory").Result(); err != nil {
		return nil, err
	} else if len(maxmemory) > 1 {
		log.Infof("Redis config maxmemory: %v", maxmemory[1])
	}

	c := NamespacedResourceWatcherCache{
		config:                  config,
		redisCli:                redisCli,
		eventProcessedWaitGroup: waitGroup,
	}

	// sanity check that the specified GVR is a valid registered CRD
	if err := c.isGvrValid(); err != nil {
		return nil, err
	}

	// let's do the initial re-sync and creating a new RetryWatcher here so
	// bootstrap errors, if any, are flagged early synchronously and the
	// caller does not end up with a partially initialized cache
	resourceVersion, err := c.resync()
	if err != nil {
		return nil, err
	}

	// RetryWatcher will take care of re-starting the watcher if the underlying channel
	// happens to close for some reason, as well as recover from other failures
	// at the same time ensuring not to replay events that have been processed
	watcher, err := watchutil.NewRetryWatcher(resourceVersion, c)
	if err != nil {
		return nil, err
	}

	go c.watchLoop(watcher)
	return &c, nil
}

func (c NamespacedResourceWatcherCache) isGvrValid() error {
	if c.config.gvr.Empty() {
		return status.Errorf(codes.FailedPrecondition, "server configured with empty GVR")
	}
	// sanity check that CRD for GVR has been registered
	ctx := context.Background()
	_, apiExt, err := c.config.clientGetter(ctx)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "clientGetter failed due to: %v", err)
	} else if apiExt == nil {
		return status.Errorf(codes.FailedPrecondition, "clientGetter returned invalid data")
	}

	name := fmt.Sprintf("%s.%s", c.config.gvr.Resource, c.config.gvr.Group)
	if crd, err := apiExt.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{}); err != nil {
		return err
	} else {
		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextv1.Established &&
				condition.Status == apiextv1.ConditionTrue {
				return nil
			}
		}
	}
	return status.Errorf(codes.FailedPrecondition, "CRD [%s] is not valid", c.config.gvr)
}

// note that I am not using pointer receivers on any the methods, because none
// of them need to modify the ResourceWatcherCache internal state.
// see https://golang.org/doc/faq#methods_on_values_or_pointers

func (c NamespacedResourceWatcherCache) watchLoop(watcher *watchutil.RetryWatcher) {
	for {
		c.receive(watcher.ResultChan())
		// if we are here, that means the RetryWatcher has stopped processing events
		// due to what it thinks is an un-retryable error (such as HTTP 410 GONE),
		// i.e. a pretty bad/unsual situation, we'll need to resync and restart the watcher
		watcher.Stop()
		// this should close the watcher channel
		<-watcher.Done()
		// per https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes
		log.Infof("Current watcher stopped. Will resync/create a new RetryWatcher...")
		resourceVersion, err := c.resync()
		if err != nil {
			log.Errorf("Failed to resync due to: %v", err)
			// TODO (gfichtenholt) retry some fixed number of times with exponential backoff?
			return
		}
		watcher, err = watchutil.NewRetryWatcher(resourceVersion, c)
		if err != nil {
			log.Errorf("Failed to create a new RetryWatcher due to: %v", err)
			// TODO (gfichtenholt) retry some fixed number of times with exponential backoff?
			return
		}
	}
}

// ResourceWatcherCache must implement cache.Watcher interface, which is this:
// https://pkg.go.dev/k8s.io/client-go@v0.20.8/tools/cache#Watcher
func (c NamespacedResourceWatcherCache) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := context.Background()

	dynamicClient, _, err := c.config.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// this will start a watcher on all namespaces
	return dynamicClient.Resource(c.config.gvr).Namespace(apiv1.NamespaceAll).Watch(ctx, options)
}

// TODO (gfichtenholt) we may need to introduce a mutex to guard against the scenario
// where someone is trying to fetch something out of the cache while its in the middle
// of a re-sync. It seems the current requirements of kubeapps catalog are pretty loose
// when it comes to consistency at any given point, as long as EVENTUALLY consistent
// state is reached, which will be the case
func (c NamespacedResourceWatcherCache) resync() (string, error) {
	ctx := context.Background()

	dynamicClient, _, err := c.config.clientGetter(ctx)
	if err != nil {
		return "", status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// This code runs in the background, i.e. not in a context of any specific user request.
	// As such, it requires RBAC to be set up properly during install to be able to list specified GVR
	// (e.g. flux CRDs). For further details, see https://github.com/kubeapps/kubeapps/pull/3551 and
	// see helm chart templates/kubeappsapis/rbac_fluxv2.yaml

	// Notice, we are not setting resourceVersion in ListOptions, which means
	// per https://kubernetes.io/docs/reference/using-api/api-concepts/
	// For Get() and List(), the semantics of resource version unset are to get the most recent
	// version
	listItems, err := dynamicClient.Resource(c.config.gvr).Namespace(apiv1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	// for debug only, will remove later
	log.Infof("List(%s) returned list with [%d] items, object:\n%s",
		c.config.gvr.Resource, len(listItems.Items), prettyPrintMap(listItems.Object))

	rv := listItems.GetResourceVersion()
	if rv == "" {
		// fail fast, without a valid resource version the whole workflow breaks down
		return "", status.Errorf(codes.Internal, "List() call response does not contain resource version")
	}

	// clear the entire cache in one call
	c.redisCli.FlushDB(c.redisCli.Context()).Err()
	if err != nil {
		return "", err
	}

	// re-populate the cache with current state from k8s
	c.populateWith(listItems.Items)
	return rv, nil
}

// this is loop that waits for new events and processes them when they happen
func (c NamespacedResourceWatcherCache) receive(ch <-chan watch.Event) {
	for {
		event, ok := <-ch
		if !ok {
			// This may happen due to
			//   HTTP 410 (HTTP_GONE) "message": "too old resource version: 1 (2200654)"
			// which according to https://kubernetes.io/docs/reference/using-api/api-concepts/
			// "...means clients must handle the case by recognizing the status code 410 Gone,
			// clearing their local cache, performing a list operation, and starting the watch
			// from the resourceVersion returned by that new list operation
			// OR it may also happen due to "cancel-able" context being canceled for whatever reason
			log.Errorf("Channel was closed unexpectedly")
			return
		}
		if event.Type == "" {
			// not quite sure why this happens (the docs don't say), but it seems to happen quite often
			continue
		}
		log.Infof("Got event: type: [%v] object:\n[%s]", event.Type, prettyPrintObject(event.Object))
		switch event.Type {
		case watch.Added, watch.Modified, watch.Deleted:
			unstructuredRepo, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				log.Errorf("Could not cast to unstructured.Unstructured")
			} else {
				if event.Type == watch.Added {
					go c.onAddOrModify(true, unstructuredRepo.Object)
				} else if event.Type == watch.Modified {
					go c.onAddOrModify(false, unstructuredRepo.Object)
				} else {
					go c.onDelete(unstructuredRepo.Object)
				}
			}
		case watch.Error:
			// will let caller (RetryWatcher) deal with it
			continue

		default:
			// TODO (gfichtenholt) handle other kinds of events?
			log.Errorf("got unexpected event: %v", event)
		}
	}
}

// this is effectively a cache PUT operation
// TODO (gfichtenholt) this func should return error if it happens, (see below for)
func (c NamespacedResourceWatcherCache) onAddOrModify(add bool, unstructuredObj map[string]interface{}) error {
	defer func() {
		if c.eventProcessedWaitGroup != nil {
			c.eventProcessedWaitGroup.Done()
		}
	}()

	key, err := c.keyFor(unstructuredObj)
	if err != nil {
		log.Errorf("Failed to get redis key due to: %v", err)
		return nil
	}

	var funcName string
	var addOrModify cacheValueSetter
	if add {
		funcName = "onAdd"
		addOrModify = c.config.onAdd
	} else {
		funcName = "onModify"
		addOrModify = c.config.onModify
	}
	value, setVal, err := addOrModify(key, unstructuredObj)
	if err != nil {
		log.Errorf("Invocation of [%s] for object %s\nfailed due to: %v", funcName, prettyPrintMap(unstructuredObj), err)
		// clear that key so cache doesn't contain any stale info for this object
		c.redisCli.Del(c.redisCli.Context(), key)
		return err
	}

	if setVal {
		// Zero expiration means the key has no expiration time.
		// TODO (gfichtenholt) currently this results in a global in-process cache that stores a potentially
		// unbound number of entries that never expire. Hence the potential of running out of memory. Fix this ASAP.
		log.Infof("about to call redisCli.Set()")
		err = c.redisCli.Set(c.redisCli.Context(), key, value, 0).Err()
		log.Infof("done with call to redisCli.Set()")
		if err != nil {
			log.Errorf("Failed to set value for object with key [%s] in cache due to: %v", key, err)
			return err
		} else {
			if log.V(2).Enabled() {
				usedMemory, totalMemory := c.memoryStats()
				log.V(2).Infof("Set value for key [%s] in cache. Redis memory usage: [%s/%s]", key, usedMemory, totalMemory)
			}
		}
	}
	return nil
}

// this is effectively a cache DEL operation
func (c NamespacedResourceWatcherCache) onDelete(unstructuredObj map[string]interface{}) error {
	defer func() {
		if c.eventProcessedWaitGroup != nil {
			c.eventProcessedWaitGroup.Done()
		}
	}()

	key, err := c.keyFor(unstructuredObj)
	if err != nil {
		log.Errorf("Failed to get redis key due to: %v", err)
		return err
	}

	delete, err := c.config.onDelete(key, unstructuredObj)
	if err != nil {
		log.Errorf("Invocation of 'onDelete' for object %s\nfailed due to: %v", prettyPrintMap(unstructuredObj), err)
		return err
	}

	if delete {
		err = c.redisCli.Del(c.redisCli.Context(), key).Err()
		if err != nil {
			log.Errorf("Failed to delete value for object [%s] from cache due to: %v", key, err)
			return err
		}
	}

	return nil
}

// this is effectively a cache GET operation
func (c NamespacedResourceWatcherCache) fetchForOne(key string) (interface{}, error) {
	// read back from cache: should be what we previously wrote or Redis.Nil
	// TODO (gfichtenholt) See if there might be a cleaner way than to have onGet() take []byte as
	// a 2nd argument. In theory, I would have liked to pass in an interface{}, just like onAdd/onModify.
	// The limitation here is caused by the fact that redis go client does not offer a
	// generic Get() method that would work with interface{}. Instead, all results are returned as
	// strings which can be converted to desired types as needed, e.g.
	// redisCli.Get(ctx, key).Bytes() first gets the string and then converts it to bytes.
	bytes, err := c.redisCli.Get(c.redisCli.Context(), key).Bytes()
	if err == redis.Nil {
		// this is normal if the key does not exist
		return nil, nil
	} else if err != nil {
		log.Errorf("Failed to get value for key [%s] from cache due to: %v", key, err)
		return nil, err
	}

	val, err := c.config.onGet(key, bytes)
	if err != nil {
		log.Errorf("Invocation of 'onGet' for object with key [%s]\nfailed due to: %v", key, err)
		return nil, err
	}

	//log.Infof("Fetched value for key [%s]: %v", key, val)
	return val, nil
}

// return all keys, optionally matching a given filter (repository list)
// currently we're caching the index of a repo using the repo name as the key
func (c NamespacedResourceWatcherCache) listKeys(filters []string) ([]string, error) {
	// see https://github.com/redis/redis/issues/3627:
	// 1) we don't want to use KEYS command
	// 2) match pattern does not support 'OR'
	// 3) simulate a HashSet in go to make sure we have no duplicates, as SCAN may
	// return duplicates
	redisKeys := map[string]struct{}{}
	match := []string{""} // everything by default

	if len(filters) > 0 {
		match = make([]string, len(filters))
		for i, f := range filters {
			match[i] = fmt.Sprintf("%s:*:%s", c.config.gvr.Resource, f)
		}
	}

	for _, m := range match {
		// https://redis.io/commands/scan An iteration starts when the cursor is set to 0,
		// and terminates when the cursor returned by the server is 0
		cursor := uint64(0)
		for {
			// glob-style pattern, you can use https://www.digitalocean.com/community/tools/glob to test
			keys, cursor, err := c.redisCli.Scan(c.redisCli.Context(), cursor, m, 0).Result()
			if err != nil {
				return nil, err
			}
			log.V(4).Infof("listKeys: SCAN returned keys: %s, cursor: [%d]", keys, cursor)
			for _, key := range keys {
				redisKeys[key] = struct{}{}
			}
			if cursor == 0 {
				break
			}
		}
	}

	resultKeys := make([]string, len(redisKeys))
	i := 0
	for k := range redisKeys {
		resultKeys[i] = k
		i++
	}
	return resultKeys, nil
}

func (c NamespacedResourceWatcherCache) fetchForMultiple(keys []string) (map[string]interface{}, error) {
	response := make(map[string]interface{})
	for _, key := range keys {
		result, err := c.fetchForOne(key)
		if err != nil {
			return nil, err
		}
		response[key] = result
	}
	return response, nil
}

// TODO (gfichtenholt) give the plug-ins the ability to override this (default) implementation
// for generating a cache key given an object
func (c NamespacedResourceWatcherCache) keyFor(unstructuredObj map[string]interface{}) (string, error) {
	name, err := namespacedName(unstructuredObj)
	if err != nil {
		return "", err
	}
	return c.keyForNamespacedName(*name), nil
}

func (c NamespacedResourceWatcherCache) keyForNamespacedName(name types.NamespacedName) string {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return fmt.Sprintf("%s:%s:%s", c.config.gvr.Resource, name.Namespace, name.Name)
}

// the opposite of keyFor
// the goal is to keep the details of what exactly the key looks like localized to one piece of code
func (c NamespacedResourceWatcherCache) fromKey(key string) (*types.NamespacedName, error) {
	parts := strings.Split(key, ":")
	if len(parts) != 3 {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	if parts[0] != c.config.gvr.Resource {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return &types.NamespacedName{Namespace: parts[1], Name: parts[2]}, nil
}

// computing a value for a key maybe expensive, e.g. indexing a repo takes a while,
// so we will do this in a concurrent fashion to minimize the time window and performance
// impact of doing so
// returns true if all was ok, false and logs any error(s) otherwise
func (c NamespacedResourceWatcherCache) populateWith(items []unstructured.Unstructured) {
	// max number of concurrent workers computing cache values at the same time
	const maxWorkers = 10

	type syncValueJob struct {
		item map[string]interface{}
	}

	var wg sync.WaitGroup
	numWorkers := int(math.Min(float64(len(items)), float64(maxWorkers)))
	requestChan := make(chan syncValueJob, numWorkers)

	// Process only at most maxWorkers at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			for job := range requestChan {
				// The following loop will only terminate when the request channel is
				// closed (and there are no more items)
				c.onAddOrModify(true, job.item)
			}
			wg.Done()
		}()
	}

	go func() {
		for _, item := range items {
			requestChan <- syncValueJob{item.Object}
		}
		close(requestChan)
	}()

	wg.Wait()
}

func (c NamespacedResourceWatcherCache) memoryStats() (used, total string) {
	used, total = "?", "?"
	// ref: https://redis.io/commands/info
	if meminfo, err := c.redisCli.Info(c.redisCli.Context(), "memory").Result(); err == nil {
		for _, l := range strings.Split(meminfo, "\r\n") {
			if used == "?" && strings.HasPrefix(l, "used_memory_rss_human:") {
				used = strings.Split(l, ":")[1]
			} else if total == "?" && strings.HasPrefix(l, "maxmemory_human:") {
				total = strings.Split(l, ":")[1]
			}
			if used != "?" && total != "?" {
				break
			}
		}
	}
	return used, total
}
