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
	"os"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	log "k8s.io/klog/v2"
)

// a type of cache that is based on watching for changes to specified kubernetes resources
type ResourceWatcherCache struct {
	// these expected to be provided by the caller when creating new cache
	config cacheConfig
	// internal state: prevent multiple watchers
	initOk bool
	// internal state: this mutex guards watcherStarted var
	initMutex sync.Mutex
	redisCli  *redis.Client
	// this WaitGroup is used exclusively by unit tests to block until all expected objects have
	// been 'processed' by the go routine running in the background. The creation of the WaitGroup object
	// and to call to .Add() is expected to be done by the unit test client. The server-side only signals
	// .Done() when processing one object is complete
	eventProcessingWaitGroup *sync.WaitGroup
}

// TODO (gfichtenholt) rename this to just Config when caching is separated out into core server
// and/or caching-rleated code is moved into a separate package?
type cacheConfig struct {
	gvr          schema.GroupVersionResource
	clientGetter server.KubernetesClientGetter
	// 'onAdd' and 'onModify' hooks are called when a new or modified object comes about and
	// allows the plug-in to return information about WHETHER OR NOT and WHAT is to be stored
	// in the cache for a given k8s object (passed in as a untyped/unstructured map)
	// the list of types actually supported be redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	onAdd    func(string, map[string]interface{}) (interface{}, bool, error)
	onModify func(string, map[string]interface{}) (interface{}, bool, error)
	// the semantics of 'onGet' hook is to convert or "reverse engineer" what was previously
	// stored in the cache (via onAdd/onModify hooks) to an object that the plug-in understands
	// and wishes to be returned as part of response to fetchCachedObjects() call
	onGet func(string, interface{}) (interface{}, error)
	// onDelete hook is called on the plug-in when the corresponding object is deleted in k8s cluster
	onDelete func(string, map[string]interface{}) (bool, error)
}

func newCache(config cacheConfig) (*ResourceWatcherCache, error) {
	log.Infof("+newCache")
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

	log.Infof("newCache: addr: [%s], password: [%s], DB=[%d]", REDIS_ADDR, REDIS_PASSWORD, REDIS_DB_NUM)

	return newCacheWithRedisClient(
		config,
		redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Password: REDIS_PASSWORD,
			DB:       REDIS_DB_NUM,
		}))
}

func newCacheWithRedisClient(config cacheConfig, redisCli *redis.Client) (*ResourceWatcherCache, error) {
	log.Infof("+newCacheWithRedisClient")

	if redisCli == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with redis Client")
	}

	if config.clientGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with configGetter")
	}

	if config.onAdd == nil || config.onModify == nil || config.onDelete == nil || config.onGet == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with expected cache hooks")
	}

	// sanity check that the redis client is connected
	pong, err := redisCli.Ping(redisCli.Context()).Result()
	if err != nil {
		return nil, err
	}
	log.Infof("[PING] -> [%s]", pong)

	c := ResourceWatcherCache{
		config:    config,
		redisCli:  redisCli,
		initOk:    false,
		initMutex: sync.Mutex{},
	}

	c.initMutex.Lock()
	go c.startResourceWatcher()
	return &c, nil
}

func (c *ResourceWatcherCache) startResourceWatcher() {
	log.Infof("+ResourceWatcherCache startResourceWatcher")
	// can't defer c.initMutex.Unlock() because when all is well,
	// we never return from this func

	if !c.initOk {
		for {
			ch, err := c.newResourceWatcherChan()
			if err != nil {
				c.initMutex.Unlock()
				log.Errorf("Failed to start resource watcher for [%s] due to: %v", c.config.gvr, err)
				return
			}
			c.initOk = true
			c.initMutex.Unlock()
			log.Infof("Watcher for [%s] successfully started. waiting for events...", c.config.gvr)

			c.processEvents(ch)
			// if we get here the watch needs to be re-started
			c.initMutex.Lock()
			c.initOk = false
		}
	} else {
		c.initMutex.Unlock()
		log.Infof("watcher already started. exiting...")
	}
	// we should never reach here under normal usage
	log.Warningf("-ResourceWatcherCache startResourceWatcher")
}

func (c *ResourceWatcherCache) newResourceWatcherChan() (<-chan watch.Event, error) {
	ctx := context.Background()

	_, dynamicClient, err := c.config.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// this will start a watcher on all namespaces
	// notice, we are not setting resourceVersion in ListOptions, which means
	// per https://kubernetes.io/docs/reference/using-api/api-concepts/
	// Get State and Start at Most Recent: to establish initial state, the watch
	// begins with synthetic "Added" events of all resources instances that exist
	// at the starting resource version. All following watch events are for all changes
	// that occurred after the resource version the watch started at
	watcher, err := dynamicClient.Resource(c.config.gvr).Namespace("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return watcher.ResultChan(), nil
}

// this is an infinite loop that waits for new events and processes them when they happen
func (c *ResourceWatcherCache) processEvents(ch <-chan watch.Event) {
	for {
		event, ok := <-ch
		if !ok {
			log.Errorf("Channel already closed. Will attempt to restart the watcher")
			// this may happen and we will need to restart the watcher
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
		default:
			// TODO (gfichtenholt) handle other kinds of events?
			log.Errorf("got unexpected event: %v", event)
		}
	}
}

// this is effectively a cache PUT operation
func (c *ResourceWatcherCache) onAddOrModify(add bool, unstructuredObj map[string]interface{}) {
	defer func() {
		if c.eventProcessingWaitGroup != nil {
			c.eventProcessingWaitGroup.Done()
		}
	}()

	key, err := c.redisKeyFor(unstructuredObj)
	if err != nil {
		log.Errorf("Failed to get redis key due to: %v", err)
		return
	}

	// clear that key so cache doesn't contain any stale info for this object if not ready or
	// indexing or marshalling fails for whatever reason
	c.redisCli.Del(c.redisCli.Context(), *key)

	var funcName string
	var value interface{}
	var setVal bool
	// Define an actual type so you can use it in your interface earlier also, as well as below:
	type CacheSetter func(string, map[string]interface{}) (interface{}, bool, error)

	var addOrModify CacheSetter
	if add {
		funcName = "OnAdd"
		addOrModify = c.config.onAdd
	} else {
		funcName = "OnModify"
		addOrModify = c.config.onModify
	}
	value, setVal, err = addOrModify(*key, unstructuredObj)
	if err != nil {
		log.Errorf("Invokation of [%s] for object %s\nfailed due to: %v", funcName, prettyPrintMap(unstructuredObj), err)
		return
	}

	if setVal {
		// Zero expiration means the key has no expiration time.
		err = c.redisCli.Set(c.redisCli.Context(), *key, value, 0).Err()
		if err != nil {
			log.Errorf("Failed to set value for object with key [%s] in cache due to: %v", *key, err)
			return
		} else {
			log.Infof("Set value for object with key [%s] in cache", *key)
		}
	}
}

// this is effectively a cache DEL operation
func (c *ResourceWatcherCache) onDelete(unstructuredObj map[string]interface{}) {
	defer func() {
		if c.eventProcessingWaitGroup != nil {
			c.eventProcessingWaitGroup.Done()
		}
	}()

	key, err := c.redisKeyFor(unstructuredObj)
	if err != nil {
		log.Errorf("Failed to get redis key due to: %v", err)
		return
	}

	delete, err := c.config.onDelete(*key, unstructuredObj)
	if err != nil {
		log.Errorf("Invokation of 'onDelete' for object %s\nfailed due to: %v", prettyPrintMap(unstructuredObj), err)
		return
	}

	if delete {
		err = c.redisCli.Del(c.redisCli.Context(), *key).Err()
		if err != nil {
			log.Errorf("Failed to delete value for object [%s] from cache due to: %v", *key, err)
		}
	}
}

// this is effectively a cache GET operation
func (c *ResourceWatcherCache) fetchForOne(key string) (interface{}, error) {
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
		log.Errorf("Invokation of 'onGet' for object with key [%s]\nfailed due to: %v", key, err)
		return nil, err
	}

	//log.Infof("Fetched value for key [%s]: %v", key, val)
	return val, nil
}

// return all keys, optionally matching a given filter (repository list)
// currently we're caching the index of a repo using the repo name as the key
func (c *ResourceWatcherCache) listKeys(filters []string) ([]string, error) {
	if err := c.checkInit(); err != nil {
		return nil, err
	}
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
			log.Infof("listKeys: SCAN returned keys: %s, cursor: [%d]", keys, cursor)
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

func (c *ResourceWatcherCache) fetchForMultiple(keys []string) (map[string]interface{}, error) {
	if err := c.checkInit(); err != nil {
		return nil, err
	}

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
func (c *ResourceWatcherCache) redisKeyFor(unstructuredObj map[string]interface{}) (*string, error) {
	name, found, err := unstructured.NestedString(unstructuredObj, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field metadata.name not found on: %v:\n%s",
			err,
			prettyPrintMap(unstructuredObj))
	}

	namespace, found, err := unstructured.NestedString(unstructuredObj, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field metadata.namespace not found on: %v:\n%s",
			err,
			prettyPrintMap(unstructuredObj))
	}

	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	s := fmt.Sprintf("%s:%s:%s", c.config.gvr.Resource, namespace, name)
	return &s, nil
}

// this func is meant to be called to make sure cache client waits
// for the cache to be fully initialized before attempting read the data out
// of the cache
func (c *ResourceWatcherCache) checkInit() error {
	c.initMutex.Lock()
	defer c.initMutex.Unlock()
	if !c.initOk {
		return status.Errorf(
			codes.FailedPrecondition,
			"server cache has not been properly initialized")
	} else {
		return nil
	}
}
