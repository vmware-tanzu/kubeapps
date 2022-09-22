// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	errorutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	watchutil "k8s.io/client-go/tools/watch"
	log "k8s.io/klog/v2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// max number of retries to process one cache entry due to transient errors
	maxWatcherCacheRetries = 5
	// max number of attempts to resync before giving up
	maxWatcherCacheResyncBackoff = 2
	KeySegmentsSeparator         = ":"
	// max number of concurrent workers computing or retrieving cache values at
	// the same time
	maxWorkers = 10
)

var (
	// pretty much a constant, init pattern similar to that of asset-syncer
	verboseWatcherCacheQueue = os.Getenv("DEBUG_WATCHER_CACHE_QUEUE") == "true"
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
	queue RateLimitingInterface

	// I am using a Read/Write Mutex to gate access to cache's resync() operation, which is
	// significant in that it flushes the whole redis cache and re-populates the state from k8s.
	// When that happens we don't really want any concurrent access to the cache until the resync()
	// operation is complete. In other words, we want to:
	//  - be able to have multiple concurrent readers (goroutines doing Get()/GetMultiple())
	//  - only a single writer (goroutine doing a resync()) is allowed, and while its doing its job
	//    no readers are allowed
	resyncCond *sync.Cond

	// bi-directional channel used exclusively by unit tests
	resyncCh chan int
}

type ValueGetterFunc func(key string, cachedValue interface{}) (rawValue interface{}, err error)
type ValueAdderFunc func(key string, obj ctrlclient.Object) (cachedValue interface{}, setValue bool, err error)
type ValueModifierFunc func(key string, obj ctrlclient.Object, oldCachedVal interface{}) (newCachedValue interface{}, setValue bool, err error)
type KeyDeleterFunc func(key string) (deleteValue bool, err error)
type ResyncFunc func() error

type NewObjectFunc func() ctrlclient.Object
type NewObjectListFunc func() ctrlclient.ObjectList
type GetListItemsFunc func(ctrlclient.ObjectList) []ctrlclient.Object

type NamespacedResourceWatcherCacheConfig struct {
	Gvr schema.GroupVersionResource
	// this ClientGetter is for running out-of-request interactions with the Kubernetes API server,
	// such as watching for resource changes
	ClientGetter clientgetter.BackgroundClientGetterFunc
	// 'OnAddFunc' hook is called when an object comes about and the cache does not have a
	// corresponding entry. Note this maybe happen as a result of a newly created k8s object
	// or a modified object for which there was no entry in the cache
	// This allows the call site to return information about WHETHER OR NOT and WHAT is to be stored
	// in the cache for a given k8s object (passed in as a ctrlclient.Object).
	// ref https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/client#Object
	// The call site may return []byte, but it doesn't have to be that.
	// The list of all types actually supported by redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	OnAddFunc ValueAdderFunc
	// 'OnModifyFunc' hook is called when an object for which there is a corresponding cache entry
	// is modified. This allows the call site to return information about WHETHER OR NOT and WHAT
	// in the cache for a given k8s object (passed in as a ctrlclient.Object).
	// ref https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/client#Object
	// The call site may return []byte, but it doesn't have to be that.
	// The list of all types actually supported by redis you can find in
	// https://github.com/go-redis/redis/blob/v8.10.0/internal/proto/writer.go#L61
	OnModifyFunc ValueModifierFunc
	// the semantics of 'OnGetFunc' hook is to convert or "reverse engineer" what was previously
	// stored in the cache (via onAdd/onModify hooks) to an object that the call site understands
	// and wishes to be returned as part of response to various flavors of 'fetch' call
	OnGetFunc ValueGetterFunc
	// OnDeleteFunc hook is called when the corresponding object is deleted in k8s cluster
	OnDeleteFunc KeyDeleterFunc
	// OnResync hook is called when the cache is resynced
	OnResyncFunc ResyncFunc

	// These funcs are needed to manipulate API-specific objects, such as flux's
	// sourcev1.HelmRepository, in a generic fashion
	NewObjFunc    NewObjectFunc
	NewListFunc   NewObjectListFunc
	ListItemsFunc GetListItemsFunc
}

// invokeExpectResync arg is only set to true for by unit tests only
func NewNamespacedResourceWatcherCache(name string, config NamespacedResourceWatcherCacheConfig, redisCli *redis.Client, stopCh <-chan struct{}, invokeExpectResync bool) (*NamespacedResourceWatcherCache, error) {
	log.Infof("+NewNamespacedResourceWatcherCache(%s, %v, %v)", name, config.Gvr, redisCli)

	if redisCli == nil {
		return nil, fmt.Errorf("server not configured with redis Client")
	} else if config.ClientGetter == nil {
		return nil, fmt.Errorf("server not configured with clientGetter")
	} else if config.OnAddFunc == nil || config.OnModifyFunc == nil ||
		config.OnDeleteFunc == nil || config.OnGetFunc == nil ||
		config.OnResyncFunc == nil || config.NewObjFunc == nil ||
		config.NewListFunc == nil || config.ListItemsFunc == nil {
		return nil, fmt.Errorf("server not configured with expected cache hooks")
	}

	c := NamespacedResourceWatcherCache{
		config:     config,
		redisCli:   redisCli,
		queue:      NewRateLimitingQueue(name, verboseWatcherCacheQueue),
		resyncCond: sync.NewCond(&sync.RWMutex{}),
	}

	// sanity check that the specified GVR is a valid registered CRD
	if err := c.isGvrValid(); err != nil {
		return nil, err
	}

	// this will launch a single worker that processes items on the work queue as they
	// come in runWorker will loop until "something bad" happens.  The .Until() func will
	// then rekick the worker after one second
	go wait.Until(c.runWorker, time.Second, stopCh)

	// this is needed by unit tests only. Since the potential lengthy bootstrap is done
	// asynchronously (see below), the this func will set a condition before returning and
	// the unit test will wait for for this condition to complete WaitUntilResyncComplete().
	// That's how it knows when the bootstrap is done
	if invokeExpectResync {
		if _, err := c.ExpectResync(); err != nil {
			return nil, err
		}
	}

	// per https://github.com/vmware-tanzu/kubeapps/issues/4329
	// we want to do this asynchronously, so that having to parse existing large repos in the cluster
	// doesn't block the kubeapps apis pod start-up
	go c.syncAndStartWatchLoop(stopCh)
	return &c, nil
}

func (c *NamespacedResourceWatcherCache) isGvrValid() error {
	if c.config.Gvr.Empty() {
		return fmt.Errorf("server configured with empty GVR")
	}
	// sanity check that CRD for GVR has been registered
	ctx := context.Background()
	apiExt, err := c.config.ClientGetter.ApiExt(ctx)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s.%s", c.config.Gvr.Resource, c.config.Gvr.Group)
	if crd, err := apiExt.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{}); err != nil {
		return err
	} else {
		// check version
		ok := false
		for _, v := range crd.Status.StoredVersions {
			if v == c.config.Gvr.Version {
				ok = true
				break
			}
		}
		if ok {
			for _, condition := range crd.Status.Conditions {
				if condition.Type == apiextv1.Established &&
					condition.Status == apiextv1.ConditionTrue {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("CRD [%s] is not valid", c.config.Gvr)
}

func (c *NamespacedResourceWatcherCache) syncAndStartWatchLoop(stopCh <-chan struct{}) {
	// RetryWatcher will take care of re-starting the watcher if the underlying channel
	// happens to close for some reason, as well as recover from other failures
	// at the same time ensuring not to replay events that have been processed
	watcher, err := c.resyncAndNewRetryWatcher(true)
	if err != nil {
		err = fmt.Errorf(
			"[%s]: Initial resync failed after [%d] retries were exhausted, last error: %v",
			c.queue.Name(), maxWatcherCacheRetries, err)
		// yes, I really want this to panic. Something is seriously wrong and
		// possibly restarting kubeapps-apis server is needed...
		runtime.Must(err)
	}
	c.watchLoop(watcher, stopCh)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *NamespacedResourceWatcherCache) runWorker() {
	log.Info("+runWorker()")
	defer log.Info("-runWorker()")

	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the work queue and
// attempt to process it, by calling the syncHandler.
func (c *NamespacedResourceWatcherCache) processNextWorkItem() bool {
	log.Info("+processNextWorkItem()")
	defer log.Info("-processNextWorkItem()")

	var obj interface{}
	var shutdown bool
	if obj, shutdown = c.queue.Get(); shutdown {
		log.Infof("[%s] worker shutting down...", c.queue.Name())
		return false
	}

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
	err := c.syncHandler(key)
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(key)
	} else if c.queue.NumRequeues(key) < maxWatcherCacheRetries {
		log.Errorf("Error processing [%s] (will retry [%d] times): %v", key, maxWatcherCacheRetries-c.queue.NumRequeues(key), err)
		c.queue.AddRateLimited(key)
	} else {
		// err != nil and too many retries
		log.Errorf("Error processing %s (giving up): %v", key, err)
		c.queue.Forget(key)
		runtime.HandleError(fmt.Errorf("error syncing key [%s] due to: %v", key, err))
	}
	return true
}

func (c *NamespacedResourceWatcherCache) watchLoop(watcher *watchutil.RetryWatcher, stopCh <-chan struct{}) {
	for {
		shutdown := c.processEvents(watcher.ResultChan(), stopCh)

		// If we are here, that means either:
		// a) stopCh was closed (i.e. graceful shutdown of the pod) OR
		// b) the RetryWatcher has stopped processing events due to what it thinks is an
		//    un-retryable error (such as HTTP 410 GONE),
		// i.e. a pretty bad/unsual situation, we'll need to resync and create a new watcher
		// In either case we should reset this resource cache work queue and
		// chart cache work queue in here, because in case of (b) we are going to completely
		// re-build both caches from scratch according to latest state in k8s and thus any
		// current pending work is unnecessary and would be duplication of effort

		watcher.Stop()
		// this should close the watcher channel
		<-watcher.Done()

		if shutdown {
			return
		}

		// per https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes
		log.Info("Current watcher stopped. Will try resync/create a new RetryWatcher...")

		var err error
		if watcher, err = c.resyncAndNewRetryWatcher(false); err != nil {
			// this is a catastrophic error for the watcher. We've tried to create a new watcher a number
			// of times and failed.

			err = fmt.Errorf(
				"[%s]: Watch loop has been stopped after [%d] retries were exhausted, last error: %v",
				c.queue.Name(), maxWatcherCacheRetries, err)
			// yes, I really want this to panic. Something is seriously wrong
			// possibly restarting plugin/kubeapps-apis server is needed...
			defer runtime.Must(err)
			break
		}
	}
}

func (c *NamespacedResourceWatcherCache) resyncAndNewRetryWatcher(bootstrap bool) (watcher *watchutil.RetryWatcher, eror error) {
	log.Info("+resyncAndNewRetryWatcher()")
	c.resyncCond.L.Lock()
	defer func() {
		if c.resyncCh != nil {
			close(c.resyncCh)
			c.resyncCh = nil
		}
		c.resyncCond.L.Unlock()
		c.resyncCond.Broadcast()
		log.Info("-resyncAndNewRetryWatcher()")
	}()

	var err error
	var resourceVersion string

	// max backoff is 2^(NamespacedResourceWatcherCacheMaxResyncBackoff) seconds
	for i := 0; i < maxWatcherCacheResyncBackoff; i++ {
		if resourceVersion, err = c.resync(bootstrap); err != nil {
			runtime.HandleError(fmt.Errorf("failed to resync due to: %v", err))
		} else if watcher, err = watchutil.NewRetryWatcher(resourceVersion, c); err != nil {
			runtime.HandleError(fmt.Errorf("failed to create a new RetryWatcher due to: %v", err))
		} else {
			break
		}
		waitTime := math.Pow(2, float64(i))
		log.Infof("Waiting [%d] seconds before retrying to resync()...", int(waitTime))
		time.Sleep(time.Duration(waitTime) * time.Second)
	}

	if err != nil {
		return nil, err
	} else {
		return watcher, nil
	}
}

// ResourceWatcherCache must implement cache.Watcher interface, which is this:
// https://pkg.go.dev/k8s.io/client-go@v0.20.8/tools/cache#Watcher
func (c *NamespacedResourceWatcherCache) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := context.Background()

	ctrlClient, err := c.config.ClientGetter.ControllerRuntime(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// this will start a watcher on all namespaces
	return ctrlClient.Watch(ctx, c.config.NewListFunc(), &ctrlclient.ListOptions{
		Namespace: apiv1.NamespaceAll,
		Raw:       &options,
	})
}

// it is expected that the caller will perform lock/unlock of c.resyncCond as there maybe
// multiple calls to resync() due to transient failure
func (c *NamespacedResourceWatcherCache) resync(bootstrap bool) (string, error) {
	log.Infof("+resync(bootstrap=%t), queue: [%s], size: [%d]", bootstrap, c.queue.Name(), c.queue.Len())
	defer log.Info("-resync()")

	// confidence test: I'd like to make sure this is called within the context
	// of resync, i.e. resync.Cond.L is locked by this goroutine.
	if !common.RWMutexWriteLocked(c.resyncCond.L.(*sync.RWMutex)) {
		return "", status.Errorf(codes.Internal, "Invalid state of the cache in resync()")
	}

	// no need to do any of this on bootstrap, queue should be empty
	if !bootstrap {
		if c.resyncCh != nil {
			c.resyncCh <- c.queue.Len()
			// now let's wait for the client (unit test code) that it's ok to proceed
			// to re-build the whole cache. Presumably the client will now set up the
			// right expectations for redis mock. Don't care what the client sends back,
			// just need an indication its ok to proceed
			<-c.resyncCh
		}
		log.Infof("Resetting work queue [%s]...", c.queue.Name())
		c.queue.Reset()

		if err := c.config.OnResyncFunc(); err != nil {
			return "", status.Errorf(codes.Internal, "invocation of [OnResync] failed due to: %v", err)
		}
	}

	// clear the entire cache in one call
	if result, err := c.redisCli.FlushDB(c.redisCli.Context()).Result(); err != nil {
		return "", err
	} else {
		log.Infof("Redis [FLUSHDB]: %s", result)
	}

	ctx := context.Background()
	ctrlClient, err := c.config.ClientGetter.ControllerRuntime(ctx)
	if err != nil {
		return "", status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// This code runs in the background, i.e. not in a context of any specific user request.
	// As such, it requires RBAC to be set up properly during install to be able to list specified GVR
	// (e.g. flux CRDs). For further details, see https://github.com/vmware-tanzu/kubeapps/pull/3551 and
	// see helm chart templates/kubeappsapis/rbac_fluxv2.yaml

	// Notice, we are not setting resourceVersion in ListOptions, which means
	// per https://kubernetes.io/docs/reference/using-api/api-concepts/
	// For Get() and List(), the semantics of resource version unset are to get the most recent
	// version
	listObj := c.config.NewListFunc()
	err = ctrlClient.List(ctx, listObj, &ctrlclient.ListOptions{Namespace: apiv1.NamespaceAll})
	if err != nil {
		return "", err
	}

	listItems := c.config.ListItemsFunc(listObj)

	// for debug only, will remove later
	log.Infof("List(%s) returned list with [%d] items, object:\n%s",
		c.config.Gvr.Resource, len(listItems), common.PrettyPrint(listObj))

	rv := listObj.GetResourceVersion()
	if rv == "" {
		// fail fast, without a valid resource version the whole workflow breaks down
		return "", status.Errorf(codes.Internal, "List() call response does not contain resource version")
	}

	// re-populate the cache with current state from k8s
	if err = c.populateWith(listItems); err != nil {
		// we don't want to fail the whole re-sync process and trigger retries
		// if, for example, just one of the repos fails to sync to cache, so
		// for now log the error(s)
		runtime.HandleError(fmt.Errorf("populateWith failed due to: %+v", err))
	}
	return rv, nil
}

// this is loop that waits for new events and processes them when they happen
func (c *NamespacedResourceWatcherCache) processEvents(watchCh <-chan watch.Event, stopCh <-chan struct{}) (shuttingDown bool) {
	for {
		select {
		case event, ok := <-watchCh:
			if !ok {
				// This may happen due to
				//   HTTP 410 (HTTP_GONE) "message": "too old resource version: 1 (2200654)"
				// which according to https://kubernetes.io/docs/reference/using-api/api-concepts/
				// "...means clients must handle the case by recognizing the status code 410 Gone,
				// clearing their local cache, performing a list operation, and starting the watch
				// from the resourceVersion returned by that new list operation
				// OR it may also happen due to "cancel-able" context being canceled for whatever reason
				log.Warning("Channel was closed unexpectedly")
				return false
			}
			c.processOneEvent(event)

		case <-stopCh:
			return true
		}
	}
}

func (c *NamespacedResourceWatcherCache) processOneEvent(event watch.Event) {
	if event.Type == "" {
		// not quite sure why this happens (the docs don't say), but it seems to happen quite often
		return
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Got event: type: [%v], object: ", event.Type))
	if event.Object == nil {
		sb.WriteString("<nil>")
	} else if event.Type == watch.Deleted {
		// when the object is deleted, we rarely care to look at all of its state,
		// so save some log space
		sb.WriteString(fmt.Sprintf("[%s]", common.PreferObjectName(event.Object)))
	} else {
		sb.WriteString(fmt.Sprintf("\n[%s]", common.PrettyPrint(event.Object)))
	}
	log.Infof(sb.String())

	switch event.Type {
	case watch.Added, watch.Modified, watch.Deleted:
		if obj, ok := event.Object.(ctrlclient.Object); !ok {
			runtime.HandleError(fmt.Errorf("could not cast %s to *ctrlclient.Object", reflect.TypeOf(event.Object)))
		} else if key, err := c.keyFor(obj); err != nil {
			runtime.HandleError(err)
		} else {
			c.queue.AddRateLimited(key)
		}
	case watch.Error:
		// will let RetryWatcher deal with it, which will close the channel
		return

	default:
		// TODO (gfichtenholt) handle other kinds of events?
		runtime.HandleError(fmt.Errorf("got unexpected event: %v", event))
	}
}

// syncs the current state of the given resource in k8s with that in the cache
func (c *NamespacedResourceWatcherCache) syncHandler(key string) error {
	log.Infof("+syncHandler(%s)", key)
	defer log.Infof("-syncHandler(%s)", key)

	// Convert the namespace/name string into a distinct namespace and name
	name, err := c.NamespacedNameFromKey(key)
	if err != nil {
		return err
	}

	// Get the resource with this namespace/name
	ctx := context.Background()
	ctrlClient, err := c.config.ClientGetter.ControllerRuntime(ctx)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	// TODO: (gfichtenholt) confidence test: I'd like to make sure the caller has the read lock,
	// i.e. we are not in the middle of a cache resync() operation. To do that, I need to
	// find a reliable alternative to common.RWMutexReadLocked which doesn't always work

	// If an error occurs during Get/Create/Update/Delete, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a temporary network
	// failure, or any other transient reason.
	obj := c.config.NewObjFunc()
	err = ctrlClient.Get(ctx, *name, obj)
	if err != nil {
		// The resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			return c.onDelete(key)
		} else {
			return status.Errorf(codes.Internal, "error fetching object with key [%s]: %v", key, err)
		}
	}
	return c.onAddOrModify(obj)
}

// this is effectively a cache GET followed by SET operation
func (c *NamespacedResourceWatcherCache) onAddOrModify(obj ctrlclient.Object) (err error) {
	log.V(4).Infof("+onAddOrModify")
	defer log.V(4).Infof("-onAddOrModify")

	var key string
	if key, err = c.keyFor(obj); err != nil {
		return err
	}

	var oldValue []byte
	if oldValue, err = c.redisCli.Get(c.redisCli.Context(), key).Bytes(); err != redis.Nil && err != nil {
		return fmt.Errorf("onAddOrModify() failed to get value for key [%s] in cache due to: %v", key, err)
	} else {
		log.V(4).Infof("Redis [GET %s]: %d bytes read", key, len(oldValue))
	}

	var setVal bool
	var funcName string
	var newValue interface{}
	if oldValue == nil {
		funcName = "onAdd"
		newValue, setVal, err = c.config.OnAddFunc(key, obj)
	} else {
		funcName = "onModify"
		newValue, setVal, err = c.config.OnModifyFunc(key, obj, oldValue)
	}

	if err != nil {
		log.Errorf("Invocation of [%s] for object %s\nfailed due to: %v", funcName, common.PrettyPrint(obj), err)
		// clear that key so cache doesn't contain any stale info for this object
		keysremoved, err2 := c.redisCli.Del(c.redisCli.Context(), key).Result()
		if err2 != nil {
			log.Errorf("failed to delete value for object [%s] from cache due to: %v", key, err2)
		} else {
			// debugging an intermittent failure
			log.Infof("Redis [DEL %s]: %d", key, keysremoved)
		}
		return nil
	} else if setVal {
		// Zero expiration means the key has no expiration time.
		// However, cache entries may be evicted by redis in order to make room for new ones,
		// if redis is limited by maxmemory constraint
		startTime := time.Now()
		result, err := c.redisCli.Set(c.redisCli.Context(), key, newValue, 0).Result()
		if err != nil {
			return fmt.Errorf("failed to set value for object with key [%s] in cache due to: %v", key, err)
		} else {
			duration := time.Since(startTime)
			// debugging an intermittent issue
			usedMemory, totalMemory := common.RedisMemoryStats(c.redisCli)
			log.Infof("Redis [SET %s]: %s in [%d] ms. Redis [INFO memory]: [%s/%s]",
				key, result, duration.Milliseconds(), usedMemory, totalMemory)
		}
	}
	return nil
}

// this is effectively a cache DEL operation
func (c *NamespacedResourceWatcherCache) onDelete(key string) error {
	log.V(4).Infof("+onDelete(%s, %s)")
	defer log.V(4).Infof("-onDelete")

	delete, err := c.config.OnDeleteFunc(key)
	if err != nil {
		log.Errorf("Invocation of 'onDelete' for object with key [%s] failed due to: %v", key, err)
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
func (c *NamespacedResourceWatcherCache) fetch(key string) (interface{}, error) {
	log.InfoS("+fetch", "key", key)
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
		return nil, fmt.Errorf("fetch() failed to get value for key [%s] from cache due to: %v", key, err)
	}
	log.V(4).Infof("Redis [GET %s]: %d bytes read", key, len(byteArray))

	// TODO (gfichtenholt) See if there might be a cleaner way than to have onGet() take []byte as
	// a 2nd argument. In theory, I would have liked to pass in an interface{}, just like onAdd/onModify.
	// The limitation here is caused by the fact that redis go client does not offer a
	// generic Get() method that would work with interface{}. Instead, all results are returned as
	// strings which can be converted to desired types as needed, e.g.
	// redisCli.Get(ctx, key).Bytes() first gets the string and then converts it to bytes.
	val, err := c.config.OnGetFunc(key, byteArray)
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

// parallelize the process of value retrieval because fetch() calls
// c.config.onGet() which will de-code the data from bytes into expected struct, which
// may be computationally expensive and thus benefit from multiple threads of execution
func (c *NamespacedResourceWatcherCache) fetchMultiple(keys sets.String) (map[string]interface{}, error) {
	response := make(map[string]interface{})

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
				result, err := c.fetch(job.key)
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
		for key := range keys {
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

// the difference between 'fetchMultiple' and 'GetMultiple' is that 'fetch' will only
// get the value from the cache for a given or return nil if one is missing, whereas
// 'GetMultiple' will first call 'fetch' but then for any cache misses it will force
// a re-computation of the value, if available, based on the input argument itemList and load
// that result into the cache. So, 'GetMultiple' provides a guarantee that if a key exists,
// it's value will be returned,
// whereas 'fetchMultiple' does not guarantee that.
// The keys are expected to be in the format of the cache (the caller does that)
func (c *NamespacedResourceWatcherCache) GetMultiple(keys sets.String) (map[string]interface{}, error) {
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	log.Infof("+GetMultiple(%s)", keys)
	// at any given moment, the redis cache may only have a subset of the entire set of existing keys.
	// Some key may have been evicted due to memory pressure and LRU eviction policy.
	// ref: https://redis.io/topics/lru-cache
	// so, first, let's fetch the entries that are still cached at this moment
	// before redis maybe forced to evict those in order to make room for new ones
	chartsUntyped, err := c.fetchMultiple(keys)
	if err != nil {
		return nil, err
	}

	// now, re-compute and fetch the ones that are left over from the previous operation
	keysLeft := sets.String{}

	for key, value := range chartsUntyped {
		if value == nil {
			// this cache miss may have happened due to one of these reasons:
			// 1) key truly does not exist in k8s (e.g. there is no repo with
			//    the given name in the "Ready" state)
			// 2) key exists and the "Ready" repo currently being indexed but
			//    has not yet completed
			// 3) key exists in k8s but the corresponding cache entry has been
			//    evicted by redis due to LRU maxmemory policies or entry TTL
			//    expiry (doesn't apply currently, cuz we use TTL=0 for all entries)
			// In the 3rd case we want to re-compute the key and add it to the cache,
			// which may potentially cause other entries to be evicted in order to
			// make room for the ones being added
			keysLeft.Insert(key)
		}
	}

	if chartsUntypedLeft, err := c.computeAndFetchValuesForKeys(keysLeft); err != nil {
		return nil, err
	} else {

		for k, v := range chartsUntypedLeft {
			chartsUntyped[k] = v
		}
	}
	return chartsUntyped, nil
}

// This func is only called in the context of a resync() operation,
// after emptying the cache via FLUSHDB, i.e. on startup or after
// some major (network) failure.
// Computing a value for a key maybe expensive, e.g. indexing a repo takes a while,
// so we will do this in a concurrent fashion to minimize the time window and performance
// impact of doing so
func (c *NamespacedResourceWatcherCache) populateWith(items []ctrlclient.Object) error {
	// confidence test: I'd like to make sure this is called within the context
	// of resync, i.e. resync.Cond.L is locked by this goroutine.
	if !common.RWMutexWriteLocked(c.resyncCond.L.(*sync.RWMutex)) {
		return status.Errorf(codes.Internal, "Invalid state of the cache in populateWith()")
	}

	keys := sets.String{}
	for _, item := range items {
		if key, err := c.keyFor(item); err != nil {
			return status.Errorf(codes.Internal, "%v", err)
		} else {
			keys.Insert(key)
		}
	}

	// wait until all all items have been processed
	c.computeValuesForKeys(keys)
	return nil
}

func (c *NamespacedResourceWatcherCache) computeValuesForKeys(keys sets.String) {
	var wg sync.WaitGroup
	numWorkers := int(math.Min(float64(len(keys)), float64(maxWorkers)))
	requestChan := make(chan string, numWorkers)

	// Process only at most maxWorkers at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			// The following loop will only terminate when the request channel is
			// closed (and there are no more items)
			for key := range requestChan {
				// see Get() for explanation of what is happening below
				c.forceKey(key)
			}
			wg.Done()
		}()
	}

	go func() {
		for key := range keys {
			requestChan <- key
		}
		close(requestChan)
	}()

	// wait until all all items have been processed
	wg.Wait()
}

func (c *NamespacedResourceWatcherCache) computeAndFetchValuesForKeys(keys sets.String) (map[string]interface{}, error) {
	type computeValueJob struct {
		key string
	}
	type computeValueJobResult struct {
		computeValueJob
		value interface{}
		err   error
	}

	var wg sync.WaitGroup
	numWorkers := int(math.Min(float64(len(keys)), float64(maxWorkers)))
	requestChan := make(chan computeValueJob, numWorkers)
	responseChan := make(chan computeValueJobResult, numWorkers)

	// Process only at most maxWorkers at a time
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			// The following loop will only terminate when the request channel is
			// closed (and there are no more items)
			for job := range requestChan {
				// see Get() for explanation of what is happening below
				value, err := c.ForceAndFetch(job.key)
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
		for key := range keys {
			requestChan <- computeValueJob{key}
		}
		close(requestChan)
	}()

	// Start receiving results
	// The following loop will only terminate when the response channel is closed, i.e.
	// after the all the requests have been processed
	chartsUntyped := make(map[string]interface{})
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
// some kind of 'KeyFunc(obj) string'
func (c *NamespacedResourceWatcherCache) keyFor(obj ctrlclient.Object) (string, error) {
	if n, err := common.NamespacedName(obj); err != nil {
		return "", err
	} else {
		return c.KeyForNamespacedName(*n), nil
	}
}

func (c *NamespacedResourceWatcherCache) KeyForNamespacedName(name types.NamespacedName) string {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepositories:ns:repoName"
	return fmt.Sprintf("%s%s%s%s%s",
		c.config.Gvr.Resource,
		KeySegmentsSeparator,
		name.Namespace,
		KeySegmentsSeparator,
		name.Name)
}

// the opposite of keyFor()
// the goal is to keep the details of what exactly the key looks like localized to one piece of code
func (c *NamespacedResourceWatcherCache) NamespacedNameFromKey(key string) (*types.NamespacedName, error) {
	parts := strings.Split(key, KeySegmentsSeparator)
	if len(parts) != 3 || parts[0] != c.config.Gvr.Resource || len(parts[1]) == 0 || len(parts[2]) == 0 {
		return nil, status.Errorf(codes.Internal, "invalid key [%s]", key)
	}
	return &types.NamespacedName{Namespace: parts[1], Name: parts[2]}, nil
}

// Get() is like fetch() but if there is a cache miss, it will also check the
// k8s for the corresponding object, process it and then add it to the cache and return the
// result.
func (c *NamespacedResourceWatcherCache) Get(key string) (interface{}, error) {
	c.resyncCond.L.(*sync.RWMutex).RLock()
	defer c.resyncCond.L.(*sync.RWMutex).RUnlock()

	log.Infof("+Get(%s)", key)
	var value interface{}
	var err error
	if value, err = c.fetch(key); err != nil {
		return nil, err
	} else if value == nil {
		// cache miss
		return c.ForceAndFetch(key)
	}
	return value, nil
}

func (c *NamespacedResourceWatcherCache) forceKey(key string) {
	c.queue.Add(key)
	// now need to wait until this item has been processed by runWorker().
	// a little bit in-efficient: syncHandler() will eventually call config.onAdd()
	// which encode the data as []byte before storing it in the cache. That part is fine.
	// But to get back the original data we have to decode it via config.onGet().
	// It'd nice if there was a shortcut and skip the cycles spent decoding data from
	// []byte to repoCacheEntry
	c.queue.WaitUntilForgotten(key)
}

func (c *NamespacedResourceWatcherCache) ForceAndFetch(key string) (interface{}, error) {
	c.forceKey(key)
	// TODO (gfichtenholt): if there was an error while processing the cache entry, such as
	// E0903 09:07:17.660753       1 watcher_cache.go:595] Invocation of [onAdd] for object {
	//  ...
	// } failed due to: rpc error: code = Internal desc = No repository lister found for OCI registry with URL: [oci://...]
	// it would be nice to surface that error here to the caller rather than
	// have redis return Nil for the key, same as a regular cache miss.
	// That requires saving more state in the cache

	// yes, there is a small time window here between after we are done with WaitUntilForgotten()
	// and the following fetch, where another concurrent goroutine may force the newly added
	// cache entry out, but that is an edge case and I am willing to overlook it for now
	// To fix it, would somehow require WaitUntilForgotten() returning a value from a cache, so
	// the whole thing would be atomic. Don't know how to do this yet
	return c.fetch(key)
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) ExpectAdd(key string) {
	c.queue.ExpectAdd(key)
}

// this func is used by unit tests only
func (c *NamespacedResourceWatcherCache) WaitUntilForgotten(key string) {
	c.queue.WaitUntilForgotten(key)
}

// this func is used by unit tests only
// returns birectional channel where the number of items in the work queue will be sent
// at the time of the resync() call and guarantees no more work items will be processed
// until resync() finishes
func (c *NamespacedResourceWatcherCache) ExpectResync() (chan int, error) {
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
		// this channel will be closed and nil'ed out at the end of resync()
		return c.resyncCh, nil
	}
}

// this func is used by unit tests only
// By the end of the call the work queue should be empty
func (c *NamespacedResourceWatcherCache) WaitUntilResyncComplete() {
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

func (c *NamespacedResourceWatcherCache) Shutdown() {
	c.queue.ShutDown()
}
