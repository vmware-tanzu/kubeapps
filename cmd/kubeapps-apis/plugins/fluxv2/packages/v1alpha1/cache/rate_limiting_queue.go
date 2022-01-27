// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	k8sruntimeutil "k8s.io/apimachinery/pkg/util/runtime"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	k8swait "k8s.io/apimachinery/pkg/util/wait"
	k8sworkqueue "k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
)

// RateLimitingInterface is an interface that rate limits items being added to the queue.
type RateLimitingInterface interface {
	k8sworkqueue.RateLimitingInterface
	Name() string
	ExpectAdd(item string)
	IsProcessing(item string) bool
	WaitUntilForgotten(item string)
	Reset()
}

func NewRateLimitingQueue(name string, verbose bool) RateLimitingInterface {
	if verbose {
		log.Infof("+NewRateLimitingQueue(%s)", name)
	}
	queue := newQueue(name, verbose)
	return &rateLimitingType{
		queue:             queue,
		DelayingInterface: k8sworkqueue.NewDelayingQueueWithCustomQueue(queue, name),
		rateLimiter:       k8sworkqueue.DefaultControllerRateLimiter(),
	}
}

// rateLimitingType wraps an Interface and provides rateLimited re-enquing
type rateLimitingType struct {
	k8sworkqueue.DelayingInterface
	rateLimiter k8sworkqueue.RateLimiter
	queue       *Type
}

// AddRateLimited AddAfter's the item based on the time when the rate limiter says it's ok
func (q *rateLimitingType) AddRateLimited(item interface{}) {
	duration := q.rateLimiter.When(item)
	if q.queue.verbose {
		log.Infof("[%s]: AddRateLimited(%s) - in %d ms", q.queue.name, item, duration.Milliseconds())
	}
	if itemstr, ok := item.(string); !ok {
		// k8sworkqueue.Interface does not allow returning errors, so
		k8sruntimeutil.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
			reflect.TypeOf(item)))
	} else {
		q.DelayingInterface.AddAfter(itemstr, duration)
	}
}

func (q *rateLimitingType) Name() string {
	return q.queue.name
}

func (q *rateLimitingType) NumRequeues(item interface{}) int {
	return q.rateLimiter.NumRequeues(item)
}

func (q *rateLimitingType) Forget(item interface{}) {
	q.rateLimiter.Forget(item)
}

func (q *rateLimitingType) IsProcessing(item string) bool {
	return q.queue.isProcessing(item)
}

func (q *rateLimitingType) Reset() {
	log.Infof("+Reset(), [%s], delayingInterface queue size: [%d]",
		q.Name(), q.DelayingInterface.Len())

	q.queue.reset()

	// this way we "forget" about ratelimit failures
	q.rateLimiter = k8sworkqueue.DefaultControllerRateLimiter()

	// TODO (gfichtenholt) Also need to "forget" the items queued up via previous call(s)
	// to .AddRateLimited() (i.e. via q.DelayingInterface.AddAfter) ?
}

// only used in unit tests
func (q *rateLimitingType) ExpectAdd(item string) {
	q.queue.expectAdd(item)
}

// used in unit test and production code, when a repo/chart needs to be loaded on demand
func (q *rateLimitingType) WaitUntilForgotten(item string) {
	q.queue.waitUntilDone(item)
	// q.queue might be done with the item, but it may have been
	// re-added via AddRateLimited if there was an error processing the item
	// in which case, NumRequeues will be > 0, and will only become 0 after
	// a call to .Forget(item).
	// TODO: (gfichtenholt) don't do k8swait.PollInfinite() here, use some sensible
	// timeout instead, and then this func will need to return an error
	k8swait.PollInfinite(10*time.Millisecond, func() (bool, error) {
		return q.rateLimiter.NumRequeues(item) == 0, nil
	})
}

func newQueue(name string, verbose bool) *Type {
	return &Type{
		name:       name,
		verbose:    verbose,
		expected:   k8ssets.String{},
		dirty:      k8ssets.String{},
		processing: k8ssets.String{},
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

// Type is a work queue.
// Ref https://pkg.go.dev/k8s.io/client-go/util/workqueue
type Type struct {
	// just for debugging purposes
	name string

	// just for debugging purposes
	verbose bool

	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue []string

	// expected defines all of the items that are expected to be processed.
	// Used in unit tests only
	expected k8ssets.String

	// dirty defines all of the items that need to be processed.
	dirty k8ssets.String

	// Things that are currently being processed are in the processing set.
	// These things may be simultaneously in the dirty set. When we finish
	// processing something and remove it from this set, we'll check if
	// it's in the dirty set, and if so, add it to the queue.
	processing k8ssets.String

	cond *sync.Cond

	shuttingDown bool
}

// Add marks item as needing processing.
func (q *Type) Add(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}
	if itemstr, ok := item.(string); !ok {
		// k8sworkqueue.Interface does not allow returning errors, so
		k8sruntimeutil.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
			reflect.TypeOf(item)))
	} else {
		q.expected.Delete(itemstr)
		if q.dirty.Has(itemstr) {
			return
		}

		q.dirty.Insert(itemstr)
		if q.processing.Has(itemstr) {
			return
		}

		q.queue = append(q.queue, itemstr)
		if q.verbose {
			log.Infof("[%s]: Add(%s)%s", q.name, item, q.prettyPrintAll())
		}
		q.cond.Broadcast()
	}
}

// Len returns the current queue length, for informational purposes only. You
// shouldn't e.g. gate a call to Add() or Get() on Len() being a particular
// value, that can't be synchronized properly.
func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return len(q.queue)
}

// Get blocks until it can return an item to be processed. If shutdown = true,
// the caller should end their goroutine. You must call Done with item when you
// have finished processing it.
func (q *Type) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for {
		for len(q.queue) == 0 && !q.shuttingDown {
			q.cond.Wait()
		}
		if q.shuttingDown {
			// We must be shutting down.
			return nil, true
		} else if len(q.queue) > 0 {
			var itemstr string
			itemstr, q.queue = q.queue[0], q.queue[1:]
			q.processing.Insert(itemstr)
			q.dirty.Delete(itemstr)
			if q.verbose {
				log.Infof("[%s]: Get() returning [%s], %s", q.name, itemstr, q.prettyPrintAll())
			}
			return itemstr, false
		}
	}
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *Type) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	if itemstr, ok := item.(string); !ok {
		// k8sworkqueue.Interface does not allow returning errors, so
		k8sruntimeutil.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
			reflect.TypeOf(item)))
	} else {
		q.processing.Delete(itemstr)
		if q.dirty.Has(itemstr) {
			q.queue = append(q.queue, itemstr)
		}
		if q.verbose {
			log.Infof("[%s]: Done(%s) %s", q.name, item, q.prettyPrintAll())
		}
		q.cond.Broadcast()
	}
}

// ShutDown will cause q to ignore all new items added to it. As soon as the
// worker goroutines have drained the existing items in the queue, they will be
// instructed to exit.
func (q *Type) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.shuttingDown = true
	if q.verbose {
		log.Infof("[%s]: Queue shutdown, sizes [empty=%d, dirty=%d, processing=%d, queue=%d]",
			q.name, q.expected.Len(), q.dirty.Len(), q.processing.Len(), len(q.queue))
	}
	q.cond.Broadcast()
}

func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}

// expectAdd marks item as expected to be processed in the near future
// Used in unit tests only
func (q *Type) expectAdd(item string) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	q.expected.Insert(item)
	if q.verbose {
		log.Infof("[%s]: expectAdd(%s) %s", q.name, item, q.prettyPrintAll())
	}
}

// this func is the added feature that was missing in k8s workqueue
func (q *Type) isProcessing(item string) bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return false
	}

	return q.processing.Has(item)
}

// this func is the added feature that was missing in k8s workqueue
func (q *Type) waitUntilDone(item string) {
	if q.verbose {
		log.Infof("[%s]: +waitUntilDone(%s)", q.name, item)
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.verbose {
		defer log.Infof("[%s]: -waitUntilDone(%s)", q.name, item)
	}

	for q.expected.Has(item) || q.dirty.Has(item) || q.processing.Has(item) {
		q.cond.Wait()

		if q.verbose {
			log.Infof("[%s]: waitUntilDone(%s) %s", q.name, item, q.prettyPrintAll())
		}
	}
}

// this func is the added feature that was missing in k8s workqueue
func (q *Type) reset() {
	if q.verbose {
		log.Infof("[%s]: +reset()", q.name)
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.queue = []string{}
	q.dirty = k8ssets.String{}
	q.processing = k8ssets.String{}
	// we are intentionally not resetting q.expected as we don't want to lose
	// those across resync's
}

// for easier reading of debug output
func (q *Type) prettyPrintAll() string {
	return fmt.Sprintf("\n\texpected: %s\n\tdirty: %s\n\tprocessing: %s\n\tqueue: %s",
		printOneItemPerLine(q.expected.List()),
		printOneItemPerLine(q.dirty.List()),
		printOneItemPerLine(q.processing.List()),
		printOneItemPerLine(q.queue))
}

func printOneItemPerLine(strs []string) string {
	if len(strs) == 0 {
		return "[]"
	} else {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("[%d] [\n", len(strs)))
		for _, s := range strs {
			sb.WriteString("\t\t" + s + "\n")
		}
		sb.WriteString("\t]")
		return sb.String()
	}
}
