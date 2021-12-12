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

Inspired by https://github.com/kubernetes/client-go/blob/v0.22.4/util/workqueue/queue.go and
         by https://github.com/kubernetes/client-go/blob/v0.22.4/util/workqueue/rate_limiting_queue.go
	but adds a couple of funcs: Expect() and WaitUntilGone()
*/
package cache

import (
	"fmt"
	"reflect"
	"sync"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
)

// RateLimitingInterface is an interface that rate limits items being added to the queue.
type RateLimitingInterface interface {
	workqueue.RateLimitingInterface
	ExpectAdd(item string)
	WaitUntilGone(item string)
}

func NewRateLimitingQueue(name string, debugEnabled bool) RateLimitingInterface {
	if debugEnabled {
		log.Infof("+NewRateLimitingQueue[%s]", name)
	}
	queue := newQueue(name, debugEnabled)
	return &rateLimitingType{
		queue:             queue,
		DelayingInterface: workqueue.NewDelayingQueueWithCustomQueue(queue, ""),
		rateLimiter:       workqueue.DefaultControllerRateLimiter(),
	}
}

// rateLimitingType wraps an Interface and provides rateLimited re-enquing
type rateLimitingType struct {
	workqueue.DelayingInterface
	rateLimiter workqueue.RateLimiter
	queue       *Type
}

// AddRateLimited AddAfter's the item based on the time when the rate limiter says it's ok
func (q *rateLimitingType) AddRateLimited(item interface{}) {
	duration := q.rateLimiter.When(item)
	if q.queue.debugEnabled {
		log.Infof("[%s]: AddRateLimited(%s) - in %d ms", q.queue.name, item, duration.Milliseconds())
	}
	if itemstr, ok := item.(string); !ok {
		// workqueue.Interface does not allow returning errors, so
		runtime.HandleError(fmt.Errorf("unexpected item in queue: expected string, found: [%s]", reflect.TypeOf(item)))
	} else {
		q.ExpectAdd(itemstr)
		q.DelayingInterface.AddAfter(itemstr, duration)
	}
}

func (q *rateLimitingType) NumRequeues(item interface{}) int {
	return q.rateLimiter.NumRequeues(item)
}

func (q *rateLimitingType) Forget(item interface{}) {
	q.rateLimiter.Forget(item)
}

func (q *rateLimitingType) ExpectAdd(item string) {
	q.queue.expectAdd(item)
}

func (q *rateLimitingType) WaitUntilGone(item string) {
	// this neeeds to take into account both the rateLimiter and the queue
	q.queue.waitUntilGone(item, q.rateLimiter)
}

func newQueue(name string, debugEnabled bool) *Type {
	return &Type{
		name:         name,
		debugEnabled: debugEnabled,
		expected:     sets.String{},
		dirty:        sets.String{},
		processing:   sets.String{},
		cond:         sync.NewCond(&sync.Mutex{}),
	}
}

// Type is a work queue.
// Ref https://pkg.go.dev/k8s.io/client-go/util/workqueue
type Type struct {
	// just for debugging purposes
	name string

	// just for debugging purposes
	debugEnabled bool

	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue []string

	// expected defines all of the items that are expected to be processed.
	// Used in unit tests only
	expected sets.String

	// dirty defines all of the items that need to be processed.
	dirty sets.String

	// Things that are currently being processed are in the processing set.
	// These things may be simultaneously in the dirty set. When we finish
	// processing something and remove it from this set, we'll check if
	// it's in the dirty set, and if so, add it to the queue.
	processing sets.String

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
		// workqueue.Interface does not allow returning errors, so
		runtime.HandleError(fmt.Errorf("unexpected item in queue: expected string, found: [%s]", reflect.TypeOf(item)))
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
		if q.debugEnabled {
			log.Infof("[%s]: Add(%s) expected: %s, dirty: %s, processing %s, queue: %s", q.name, item, q.expected.List(), q.dirty.List(), q.processing.List(), q.queue)
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
			if q.debugEnabled {
				log.Infof("[%s]: Get() returning %s expected: %s, dirty: %s, processing %s, queue: %s", q.name, itemstr, q.expected.List(), q.dirty.List(), q.processing.List(), q.queue)
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

	if itemstr, ok := item.(string); !ok {
		// workqueue.Interface does not allow returning errors, so
		runtime.HandleError(fmt.Errorf("unexpected item in queue: expected string, found: [%s]", reflect.TypeOf(item)))
	} else {
		q.processing.Delete(itemstr)
		if q.dirty.Has(itemstr) {
			q.queue = append(q.queue, itemstr)
		}
		if q.debugEnabled {
			log.Infof("[%s]: Done(%s) expected: %s, dirty: %s, processing %s, queue: %s", q.name, item, q.expected.List(), q.dirty.List(), q.processing.List(), q.queue)
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
	q.cond.Broadcast()
}

func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}

// Add marks item as expected to be processed in the near future
// Used in unit tests only
func (q *Type) expectAdd(item string) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.shuttingDown {
		return
	}

	q.expected.Insert(item)
	if q.debugEnabled {
		log.Infof("[%s]: expectAdd(%s) expected: %s, dirty: %s, processing %s, queue: %s", q.name, item, q.expected.List(), q.dirty.List(), q.processing.List(), q.queue)
	}
}

// this func is the whole reason for the existence of this queue
func (q *Type) waitUntilGone(item string, rateLimiter workqueue.RateLimiter) {
	if q.debugEnabled {
		log.Infof("[%s]: +waitUntilGone(%s)", q.name, item)
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.debugEnabled {
		defer log.Infof("[%s]: -waitUntilGone(%s)", q.name, item)
	}

	for q.expected.Has(item) || q.dirty.Has(item) || q.processing.Has(item) {
		q.cond.Wait()

		if q.debugEnabled {
			log.Infof("[%s]: waitUntilGone expected: %s, dirty: %s, processing %s, queue: %s", q.name, q.expected.List(), q.dirty.List(), q.processing.List(), q.queue)
		}
	}
}
