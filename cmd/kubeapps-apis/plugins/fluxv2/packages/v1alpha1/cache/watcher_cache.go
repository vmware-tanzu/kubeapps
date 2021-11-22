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
package cache

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchutil "k8s.io/client-go/tools/watch"
	log "k8s.io/klog/v2"
)

// a type of cache that is based on watching for changes to specified kubernetes resources.
// The resource is assumed to be namespace-scoped. Cluster-wide resources are not
// supported at this time
type NamespacedResourceWatcherCache struct {
	// these expected to be provided by the caller when creating new cache
	config   NamespacedResourceWatcherCacheConfig
	redisCli *redis.Client

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time and makes it easy to ensure we are never processing the same item
	// simultaneously in different workers. So we are using a workqueue.RateLimitingInterface
	// in order to satisfy these requirements:
	// - events for a given repo must be fully processed in the order they were received,
	//   e.g. in a scenario when a MODIFIED event was received and the processing of it has began
	//   that might take a long time to process due to indexing of a large index.yaml file.
	//   In the meantime, if a DELETED event happens to come in, the MODIFIED event
	//   needs to be fully processed before the DELETED event is, otherwise the cache will end up
	//   in invalid state (a entry in the cache for an object that doesn't exist)
	// as well as for other features:
	// - multiple rapid updates to a single resource will be collapsed into the latest version
	//   by the cache/queue
	// ref: https://pkg.go.dev/k8s.io/client-go/util/workqueue
	// ref: https://engineering.bitnami.com/articles/a-deep-dive-into-kubernetes-controllers.html#workqueue
	queue rateLimitingInterface

	// I am using a Read/Write Mutex to gate access to cache's resync() operation, which is
	// significant in that it flushes the whole redis cache and re-populates the state from k8s.
	// When that happens we don't really want any concurrent access to the cache until the resync()
	// operation is complete. In other words, we want to:
	//  - be able to have multiple concurrent readers (goroutines doing GetForOne()/GetForMultiple())
	//  - only a single writer (goroutine doing a resync()) is allowed, and while its doing its job
	//    no readers are allowed
	resyncCond *sync.Cond

	// used exclusively by unit tests
	expectResync bool
}

type ValueGetterFunc func(string, interface{}) (interface{}, error)
type ValueAdderFunc func(string, map[string]interface{}) (interface{}, bool, error)
type ValueModifierFunc func(string, map[string]interface{}, interface{}) (interface{}, bool, error)
type KeyDeleterFunc func(string) (bool, error)

type NamespacedResourceWatcherCacheConfig struct {
	Gvr schema.GroupVersionResource
	// this ClientGetter is for running out-of-request interactions with the Kubernetes API server,
	// such as watching for resource changes
	ClientGetter common.ClientGetterFunc
	// 'OnAddFunc' hook is called when an object comes about and the cache does not have a
	// corresponding entry. Note this maybe happen as a result of a newly created k8s object
	// or a modified object for which there was no entry in the cache
	// This allows the call site to return information about WHETHER OR NOT and WHAT is to be stored
	// in the cache for a given k8s object (passed in as a untyped/unstructured map).
	// The call site may return []byte, but it doesn't have to be that.
	// The list of all types actually supported by redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	OnAddFunc ValueAdderFunc
	// 'OnModifyFunc' hooks is called when an object for which there is a corresponding cache entry
	// is modified. This allows the call site to return information about WHETHER OR NOT and WHAT
	// is to be stored in the cache for a given k8s object (passed in as a untyped/unstructured map).
	// The call site may return []byte, but it doesn't have to be that.
	// The list of all types actually supported by redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	OnModifyFunc ValueModifierFunc
	// the semantics of 'OnGetFunc' hook is to convert or "reverse engineer" what was previously
	// stored in the cache (via onAdd/onModify hooks) to an object that the call site understands
	// and wishes to be returned as part of response to various flavors of 'fetch' call
	OnGetFunc ValueGetterFunc
	// OnDeleteFunc hook is called on the plug-in when the corresponding object is deleted in k8s cluster
	OnDeleteFunc KeyDeleterFunc
}

func NewNamespacedResourceWatcherCache(config NamespacedResourceWatcherCacheConfig, redisCli *redis.Client) (*NamespacedResourceWatcherCache, error) {
	log.Infof("+NewNamespacedResourceWatcherCache(%v, %v)", config.Gvr, redisCli)

	if redisCli == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with redis Client")
	} else if config.ClientGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with clientGetter")
	} else if config.OnAddFunc == nil || config.OnModifyFunc == nil || config.OnDeleteFunc == nil || config.OnGetFunc == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with expected cache hooks")
	}

	c := NamespacedResourceWatcherCache{
		config:     config,
		redisCli:   redisCli,
		queue:      newRateLimitingQueue(),
		resyncCond: sync.NewCond(&sync.RWMutex{}),
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

	// dummy channel for now. Ideally, this would be passed in as an input argument
	// and the caller would indicate when to stop
	stopCh := make(chan struct{})
	// this will launch a single worker that processes items on the work queue as they come in
	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	// TODO (gfichtenholt) in theory, we should be able to launch multiple workers,
	// and the workqueue will make sure that only a single worker works on an item with a given key.
	// Test that actually is the case
	go wait.Until(c.runWorker, time.Second, stopCh)

	// RetryWatcher will take care of re-starting the watcher if the underlying channel
	// happens to close for some reason, as well as recover from other failures
	// at the same time ensuring not to replay events that have been processed
	watcher, err := watchutil.NewRetryWatcher(resourceVersion, &c)
	if err != nil {
		return nil, err
	}

	go c.watchLoop(watcher)
	return &c, nil
}

func (c *NamespacedResourceWatcherCache) isGvrValid() error {
	if c.config.Gvr.Empty() {
		return status.Errorf(codes.FailedPrecondition, "server configured with empty GVR")
	}
	// sanity check that CRD for GVR has been registered
	ctx := context.Background()
	_, apiExt, err := c.config.ClientGetter(ctx)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "clientGetter failed due to: %v", err)
	} else if apiExt == nil {
		return status.Errorf(codes.FailedPrecondition, "clientGetter returned invalid data")
	}

	name := fmt.Sprintf("%s.%s", c.config.Gvr.Resource, c.config.Gvr.Group)
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
	return status.Errorf(codes.FailedPrecondition, "CRD [%s] is not valid", c.config.Gvr)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *NamespacedResourceWatcherCache) runWorker() {
	log.Infof("+runWorker()")
	defer log.Infof("-runWorker()")

	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the work queue and
// attempt to process it, by calling the syncHandler.
func (c *NamespacedResourceWatcherCache) processNextWorkItem() bool {
	log.V(4).Infof("+processNextWorkItem()")
	defer log.V(4).Infof("-processNextWorkItem()")

	obj, shutdown := c.queue.Get()
	// ref https://go101.org/article/concurrent-synchronization-more.html
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

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
			if _, err := c.syncHandler(key); err != nil {
				return fmt.Errorf("error syncing key [%s] due to: %v", key, err)
			}
			// Finally, if no error occurs we Forget this item so it does not
			// get queued again until another change happens.
			c.queue.Forget(obj)
			log.Infof("Done syncing key [%s]", key)
			return nil
		}
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}
	return true
}

func (c *NamespacedResourceWatcherCache) watchLoop(watcher *watchutil.RetryWatcher) {
	for {
		c.processWatchEvents(watcher.ResultChan())
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
			runtime.HandleError(fmt.Errorf("failed to resync due to: %v", err))
			// TODO (gfichtenholt) retry some fixed number of times with exponential backoff?
			return
		}
		watcher, err = watchutil.NewRetryWatcher(resourceVersion, c)
		if err != nil {
			runtime.HandleError(fmt.Errorf("failed to create a new RetryWatcher due to: %v", err))
			// TODO (gfichtenholt) retry some fixed number of times with exponential backoff?
			return
		}
	}
}

// ResourceWatcherCache must implement cache.Watcher interface, which is this:
// https://pkg.go.dev/k8s.io/client-go@v0.20.8/tools/cache#Watcher
func (c *NamespacedResourceWatcherCache) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := context.Background()

	dynamicClient, _, err := c.config.ClientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// this will start a watcher on all namespaces
	return dynamicClient.Resource(c.config.Gvr).Namespace(apiv1.NamespaceAll).Watch(ctx, options)
}

func (c *NamespacedResourceWatcherCache) resync() (string, error) {
	c.resyncCond.L.Lock()
	defer func() {
		c.expectResync = false
		c.resyncCond.L.Unlock()
		c.resyncCond.Broadcast()
		log.Infof("-resync()")
	}()

	log.Infof("+resync()")

	// clear the entire cache in one call
	if result, err := c.redisCli.FlushDB(c.redisCli.Context()).Result(); err != nil {
		return "", err
	} else {
		log.Infof("Redis [FLUSHDB]: %s", result)
	}

	ctx := context.Background()
	dynamicClient, _, err := c.config.ClientGetter(ctx)
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
	listItems, err := dynamicClient.Resource(c.config.Gvr).Namespace(apiv1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	// for debug only, will remove later
	log.Infof("List(%s) returned list with [%d] items, object:\n%s",
		c.config.Gvr.Resource, len(listItems.Items), common.PrettyPrintMap(listItems.Object))

	rv := listItems.GetResourceVersion()
	if rv == "" {
		// fail fast, without a valid resource version the whole workflow breaks down
		return "", status.Errorf(codes.Internal, "List() call response does not contain resource version")
	}

	// re-populate the cache with current state from k8s
	c.populateWith(listItems.Items)
	return rv, nil
}

// this is loop that waits for new events and processes them when they happen
func (c *NamespacedResourceWatcherCache) processWatchEvents(ch <-chan watch.Event) {
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
			log.Warning("Channel was closed unexpectedly")
			return
		}
		if event.Type == "" {
			// not quite sure why this happens (the docs don't say), but it seems to happen quite often
			continue
		}
		log.Infof("Got event: type: [%v] object:\n[%s]", event.Type, common.PrettyPrintObject(event.Object))
		switch event.Type {
		case watch.Added, watch.Modified, watch.Deleted:
			if unstructuredObj, ok := event.Object.(*unstructured.Unstructured); !ok {
				runtime.HandleError(fmt.Errorf("could not cast %s to unstructured.Unstructured", reflect.TypeOf(event.Object)))
			} else {
				var key string
				var err error
				if event.Type == watch.Added || event.Type == watch.Modified {
					key, err = cache.MetaNamespaceKeyFunc(unstructuredObj)
				} else {
					key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(unstructuredObj)
				}
				if err != nil {
					runtime.HandleError(err)
				} else {
					c.queue.AddRateLimited(key)
				}
			}
		case watch.Error:
			// will let caller (RetryWatcher) deal with it
			continue

		default:
			// TODO (gfichtenholt) handle other kinds of events?
			runtime.HandleError(fmt.Errorf("got unexpected event: %v", event))
		}
	}
}

// syncs the current state of the given resource in k8s with that in the cache
func (c *NamespacedResourceWatcherCache) syncHandler(key string) (interface{}, error) {
	log.Infof("+syncHandler(%s)", key)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil || namespace == "" || name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource key: %s", key)
	}

	// Get the resource with this namespace/name
	ctx := context.Background()
	dynamicClient, _, err := c.config.ClientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// If an error occurs during Get/Create/Update/Delete, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a temporary network
	// failure, or any other transient reason.
	unstructuredObj, err := dynamicClient.Resource(c.config.Gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		// The resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			return nil, c.onDelete(namespace, name)
		} else {
			return nil, status.Errorf(codes.Internal, "error fetching object with key [%s]: %v", key, err)
		}
	}
	return c.onAddOrModify(true, unstructuredObj.Object)
}

// this is effectively a cache SET operation
func (c *NamespacedResourceWatcherCache) onAddOrModify(checkOldValue bool, unstructuredObj map[string]interface{}) (newValue interface{}, err error) {
	log.V(4).Infof("+onAddOrModify")
	defer log.V(4).Infof("-onAddOrModify")

	key, err := c.keyFor(unstructuredObj)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis key due to: %v", err)
	}

	var oldValue []byte
	if checkOldValue {
		if oldValue, err = c.redisCli.Get(c.redisCli.Context(), key).Bytes(); err != redis.Nil && err != nil {
			return nil, fmt.Errorf("onAddOrModify() failed to get value for key [%s] in cache due to: %v", key, err)
		}
	}

	var setVal bool
	var funcName string
	if oldValue == nil {
		funcName = "onAdd"
		newValue, setVal, err = c.config.OnAddFunc(key, unstructuredObj)
	} else {
		funcName = "onModify"
		newValue, setVal, err = c.config.OnModifyFunc(key, unstructuredObj, oldValue)
	}

	if err != nil {
		log.Errorf("Invocation of [%s] for object %s\nfailed due to: %v", funcName, common.PrettyPrintMap(unstructuredObj), err)
		// clear that key so cache doesn't contain any stale info for this object
		keysremoved, err2 := c.redisCli.Del(c.redisCli.Context(), key).Result()
		if err2 != nil {
			log.Errorf("failed to delete value for object [%s] from cache due to: %v", key, err2)
		} else {
			// debugging an intermittent failure
			log.Infof("Redis [DEL %s]: %d", key, keysremoved)
		}
		return nil, err
	} else if setVal {
		// Zero expiration means the key has no expiration time.
		// However, cache entries may be evicted by redis in order to make room for new ones,
		// if redis is limited by maxmemory constraint
		result, err := c.redisCli.Set(c.redisCli.Context(), key, newValue, 0).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to set value for object with key [%s] in cache due to: %v", key, err)
		} else {
			// debugging an intermittent issue
			usedMemory, totalMemory := c.memoryStats()
			log.Infof("Redis [SET %s]: %s. Redis [INFO memory]: [%s/%s]", key, result, usedMemory, totalMemory)
		}
	}
	return newValue, nil
}

// this is effectively a cache DEL operation
func (c *NamespacedResourceWatcherCache) onDelete(namespace, name string) error {
	log.V(4).Infof("+onDelete(%s, %s)", namespace, name)
	defer log.V(4).Infof("-onDelete")

	key := c.KeyForNamespacedName(types.NamespacedName{Namespace: namespace, Name: name})
	delete, err := c.config.OnDeleteFunc(key)
	if err != nil {
		log.Errorf("Invocation of 'onDelete' for object with key [%s] failed due to: %v", err)
		return err
	}

	if delete {
		keysremoved, err := c.redisCli.Del(c.redisCli.Context(), key).Result()
		if err != nil {
			return fmt.Errorf("failed to delete value for object [%s] from cache due to: %v", key, err)
		} else {
			// debugging an intermittent failure
			log.Infof("Redis [DEL %s]: %d", key, keysremoved)
		}
	}
	return nil
}

// this is effectively a cache GET operation
func (c *NamespacedResourceWatcherCache) fetchForOne(key string) (interface{}, error) {
	log.Infof("+fetchForOne(%s)", key)
	// read back from cache: should be either:
	//  - what we previously wrote OR
	//  - redis.Nil if the key does  not exist or has been evicted due to memory pressure/TTL expiry
	//
	bytes, err := c.redisCli.Get(c.redisCli.Context(), key).Bytes()
	// debugging an intermittent issue
	if err == redis.Nil {
		log.V(4).Infof("Redis [GET %s]: Nil", key)
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("fetchForOne() failed to get value for key [%s] from cache due to: %v", key, err)
	}
	log.V(4).Infof("Redis [GET %s]: %d bytes read", key, len(bytes))

	// TODO (gfichtenholt) See if there might be a cleaner way than to have onGet() take []byte as
	// a 2nd argument. In theory, I would have liked to pass in an interface{}, just like onAdd/onModify.
	// The limitation here is caused by the fact that redis go client does not offer a
	// generic Get() method that would work with interface{}. Instead, all results are returned as
	// strings which can be converted to desired types as needed, e.g.
	// redisCli.Get(ctx, key).Bytes() first gets the string and then converts it to bytes.
	val, err := c.config.OnGetFunc(key, bytes)
	if err != nil {
		log.Errorf("Invocation of 'onGet' for object with key [%s]\nfailed due to: %v", key, err)
		return nil, err
	}
	return val, nil
}

// it is worth noting that a method such as
//   func (c NamespacedResourceWatcherCache) listKeys(filters []string) ([]string, error)
// has proven to be of no use today. The problem is that such a function can
// only return the set of keys in the cache *AT A GIVEN MONENT IN TIME*, which maybe a subset
// of all existing keys (due to memory pressure and eviction policies) and therefore cannot
// be relied upon to be the "source of truth". So I removed it for now as I found it
// of no use

// parallelize the process of value retrieval because fetchForOne() calls
// c.config.onGet() which will de-code the data from bytes into expected struct, which
// may be computationally expensive and thus benefit from multiple threads of execution
func (c *NamespacedResourceWatcherCache) fetchForMultiple(keys []string) (map[string]interface{}, error) {
	response := make(map[string]interface{})

	// max number of concurrent workers retrieving cache values at the same time
	const maxWorkers = 10

	type fetchValueJob struct {
		key string
	}
	type fetchValueJobResult struct {
		fetchValueJob
		value interface{}
		err   error
	}

	var wg sync.WaitGroup
	numWorkers := int(math.Min(float64(len(keys)), float64(maxWorkers)))
	requestChan := make(chan fetchValueJob, numWorkers)
	responseChan := make(chan fetchValueJobResult, numWorkers)

	// Process only at most maxWorkers at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			// The following loop will only terminate when the request channel is
			// closed (and there are no more items)
			for job := range requestChan {
				result, err := c.fetchForOne(job.key)
				responseChan <- fetchValueJobResult{job, result, err}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	go func() {
		for _, key := range keys {
			requestChan <- fetchValueJob{key}
		}
		close(requestChan)
	}()

	// Start receiving results
	// The following loop will only terminate when the response channel is closed, i.e.
	// after the all the requests have been processed
	errs := []error{}
	for resp := range responseChan {
		if resp.err == nil {
			response[resp.key] = resp.value
		} else {
			errs = append(errs, resp.err)
		}
	}
	return response, errorutil.NewAggregate(errs)
}

// the difference between 'fetchForMultiple' and 'GetForMultiple' is that 'fetch' will only
// get the value from the cache for a given or return nil if one is missing, whereas
// 'GetForMultiple' will first call 'fetch' but then for any cache misses it will force
// a re-computation of the value, if available, based on the input argument itemList and load
// that result into the cache. So, 'GetForMultiple' provides a guarantee that if a key exists,
// it's value will be returned,
// whereas 'fetchForMultiple' does not guarantee that.
// The keys are expected to be in the format of the cache (the caller does that)
func (c *NamespacedResourceWatcherCache) GetForMultiple(keys []string) (map[string]interface{}, error) {
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	log.Infof("+GetForMultiple(%s)", keys)
	// at any given moment, the redis cache may only have a subset of the entire set of existing keys.
	// Some key may have been evicted due to memory pressure and LRU eviction policy.
	// ref: https://redis.io/topics/lru-cache
	// so, first, let's fetch the entries that are still cached at this moment
	// before redis maybe forced to evict those in order to make room for new ones
	chartsUntyped, err := c.fetchForMultiple(keys)
	if err != nil {
		return nil, err
	}

	// now, re-compute and fetch the ones that are left over from the previous operation
	keysLeft := []string{}

	for key, value := range chartsUntyped {
		if value == nil {
			// this cache miss may have happened due to one of these reasons:
			// 1) key truly does not exist in k8s (e.g. there is no repo with the given name in the "Ready" state)
			// 2) key exists and the "Ready" repo currently being indexed but has not yet completed
			// 3) key exists in k8s but the corresponding cache entry has been evicted by redis due to
			//    LRU maxmemory policies or entry TTL expiry (doesn't apply currently, cuz we use TTL=0
			//    for all entries)
			// In the 3rd case we want to re-compute the key and add it to the cache, which may potentially
			// cause other entries to be evicted in order to make room for the ones being added
			keysLeft = append(keysLeft, key)
		}
	}

	// this functionality is similar to that of populateWith() func,
	// but different enough so I did not see the value of re-using the code

	// max number of concurrent workers retrieving cache values at the same time
	const maxWorkers = 10

	type computeValueJob struct {
		key string
	}
	type computeValueJobResult struct {
		computeValueJob
		value interface{}
		err   error
	}

	var wg sync.WaitGroup
	numWorkers := int(math.Min(float64(len(keysLeft)), float64(maxWorkers)))
	requestChan := make(chan computeValueJob, numWorkers)
	responseChan := make(chan computeValueJobResult, numWorkers)

	// Process only at most maxWorkers at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			// The following loop will only terminate when the request channel is
			// closed (and there are no more items)
			for job := range requestChan {
				value, err := func() (interface{}, error) {
					if name, err := c.fromKey(job.key); err != nil {
						return nil, err
					} else {
						// see GetForOne() for explanation of what is happening below
						c.queue.Add(name.String())
						c.queue.WaitUntilDoneWith(name.String())
						return c.fetchForOne(job.key)
					}
				}()
				responseChan <- computeValueJobResult{job, value, err}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(responseChan)
	}()

	go func() {
		for _, key := range keysLeft {
			requestChan <- computeValueJob{key}
		}
		close(requestChan)
	}()

	// Start receiving results
	// The following loop will only terminate when the response channel is closed, i.e.
	// after the all the requests have been processed
	errs := []error{}
	for resp := range responseChan {
		if resp.err == nil {
			chartsUntyped[resp.key] = resp.value
		} else {
			errs = append(errs, resp.err)
		}
	}
	return chartsUntyped, errorutil.NewAggregate(errs)
}

// TODO (gfichtenholt) give the plug-ins the ability to override this (default) implementation
// for generating a cache key given an object
// some kind of 'KeyFunc(unstructuredObj) string'
func (c *NamespacedResourceWatcherCache) keyFor(unstructuredObj map[string]interface{}) (string, error) {
	name, err := common.NamespacedName(unstructuredObj)
	if err != nil {
		return "", err
	}
	return c.KeyForNamespacedName(*name), nil
}

func (c *NamespacedResourceWatcherCache) KeyForNamespacedName(name types.NamespacedName) string {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	return fmt.Sprintf("%s:%s:%s", c.config.Gvr.Resource, name.Namespace, name.Name)
}

// the opposite of keyFor
// the goal is to keep the details of what exactly the key looks like localized to one piece of code
func (c *NamespacedResourceWatcherCache) fromKey(key string) (*types.NamespacedName, error) {
	parts := strings.Split(key, ":")
	if len(parts) != 3 || parts[0] != c.config.Gvr.Resource || len(parts[1]) == 0 || len(parts[1]) == 0 {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return &types.NamespacedName{Namespace: parts[1], Name: parts[2]}, nil
}

// This func is only called in the context of a resync() operation,
// after emptying the cache via FLUSHDB, i.e. on startup or after
// some major (network) failure.
// Computing a value for a key maybe expensive, e.g. indexing a repo takes a while,
// so we will do this in a concurrent fashion to minimize the time window and performance
// impact of doing so
func (c *NamespacedResourceWatcherCache) populateWith(items []unstructured.Unstructured) {
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
			// The following loop will only terminate when the request channel is
			// closed (and there are no more items)
			for job := range requestChan {
				// don't need to check old value since we just flushed the whole cache
				if _, err := c.onAddOrModify(false, job.item); err != nil {
					// log an error and move on
					log.Errorf("populateWith: %+v", err)
				}
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

// GetForOne() is like fetchForOne() but if there is a cache miss, it will also check the
// k8s for the corresponding object, process it and then add it to the cache and return the
// result.
func (c *NamespacedResourceWatcherCache) GetForOne(key string) (interface{}, error) {
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	log.Infof("+GetForOne(%s)", key)
	var value interface{}
	var err error
	if value, err = c.fetchForOne(key); err != nil {
		return nil, err
	} else if value == nil {
		// cache miss
		if name, err := c.fromKey(key); err != nil {
			return nil, err
		} else {
			// shortcut: typespacedName.String() happens to return the same format
			// strings as cache.MetaNamespaceKeyFunc(unstructuredObj)
			// place the item in the work queue
			c.queue.Add(name.String())
			// now need to wait until this item has been processed by runWorker().
			// a little bit in-efficient: syncHandler() will eventually call config.onAdd()
			// which encode the data as []byte before storing it in the cache. That part is fine.
			// But to get back the original data we have to decode it via config.onGet().
			// It'd nice if there was a shortcut and skip the cycles spent decoding data from
			// []byte to repoCacheEntry
			c.queue.WaitUntilDoneWith(name.String())
			// yes, there is a small time window here between after we are done with WaitUntilDoneWith
			// and the following fetch, where another concurrent goroutine may force the newly added
			// cache entry out, but that is an edge case and I am willing to overlook it for now
			// To fix it, would somehow require WaitUntilDoneWith returning a value from a cache, so
			// the whole thing would be atomic. Don't know how to do this yet
			return c.fetchForOne(key)
		}
	}
	return value, nil
}

func (c *NamespacedResourceWatcherCache) memoryStats() (used, total string) {
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
	} else {
		log.Warningf("Failed to get redis memory stats due to: %v", err)
	}
	return used, total
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) ExpectAdd(key string) error {
	if name, err := c.fromKey(key); err != nil {
		return err
	} else {
		c.queue.ExpectAdd(name.String())
		return nil
	}
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) WaitUntilDoneWith(key string) error {
	if name, err := c.fromKey(key); err != nil {
		return err
	} else {
		c.queue.WaitUntilDoneWith(name.String())
		return nil
	}
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) ExpectResync() {
	c.resyncCond.L.Lock()
	defer c.resyncCond.L.Unlock()

	c.expectResync = true
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) WaitUntilResyncComplete() {
	c.resyncCond.L.Lock()
	defer c.resyncCond.L.Unlock()

	for c.expectResync {
		c.resyncCond.Wait()
	}
}
