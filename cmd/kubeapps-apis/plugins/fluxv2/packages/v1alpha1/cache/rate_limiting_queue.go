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

Inspired by https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go and
         by https://github.com/kubernetes/client-go/blob/v0.22.4/util/workqueue/rate_limiting_queue.go
	but adds a few funcs, like ExpectAdd() and WaitUntilDone()
*/
package cache

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/workqueue"
	log "k8s.io/klog/v2"
)

// RateLimitingInterface is an interface that rate limits items being added to the queue.
type RateLimitingInterface interface {
	workqueue.RateLimitingInterface
	Name() string
	ExpectAdd(item string)
	IsProcessing(item string) bool
	WaitUntilDone(item string)
	Reset()
}

func NewRateLimitingQueue(name string, debugEnabled bool) RateLimitingInterface {
	if debugEnabled {
		log.Infof("+NewRateLimitingQueue(%s)", name)
	}
	queue := newQueue(name, debugEnabled)
	return &rateLimitingType{
		queue:             queue,
		DelayingInterface: workqueue.NewDelayingQueueWithCustomQueue(queue, name),
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
		runtime.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
			reflect.TypeOf(item)))
	} else {
		// TODO (gfichtenholt) .ExpectAdd() was added here for
		// TestTransientHttpFailuresAreRetriedForChartCache scenario, but it doesn't belong here
		// it should be moved someplace into the unit test itself
		q.ExpectAdd(itemstr)
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

func (q *rateLimitingType) ExpectAdd(item string) {
	q.queue.expectAdd(item)
}

func (q *rateLimitingType) WaitUntilDone(item string) {
	q.queue.waitUntilDone(item)
}

func (q *rateLimitingType) IsProcessing(item string) bool {
	return q.queue.isProcessing(item)
}

func (q *rateLimitingType) Reset() {
	log.Infof("+Reset(), [%s], delayingInterface queue size: [%d]",
		q.Name(), q.DelayingInterface.Len())

	q.queue.reset()

	// this way we "forget" about ratelimit failures
	q.rateLimiter = workqueue.DefaultControllerRateLimiter()

	// TODO (gfichtenholt) Also need to "forget" the items queued up via previous call(s)
	// to .AddRateLimited() (i.e. via q.DelayingInterface.AddAfter)
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
		runtime.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
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
		if q.debugEnabled {
			if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
				log.Infof("[%s]: Add(%s)%s", q.name, item, q.prettyPrintAll())
			}
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
				if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
					log.Infof("[%s]: Get() returning [%s], %s", q.name, itemstr, q.prettyPrintAll())
				}
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
		// workqueue.Interface does not allow returning errors, so
		runtime.HandleError(fmt.Errorf("invalid argument: expected string, found: [%s]",
			reflect.TypeOf(item)))
	} else {
		q.processing.Delete(itemstr)
		if q.dirty.Has(itemstr) {
			q.queue = append(q.queue, itemstr)
		}
		if q.debugEnabled {
			if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
				log.Infof("[%s]: Done(%s) %s", q.name, item, q.prettyPrintAll())
			}
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
	if q.debugEnabled {
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
	if q.debugEnabled {
		if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
			log.Infof("[%s]: expectAdd(%s) %s",
				q.name, item, q.prettyPrintAll())
		}
	}
}

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
	if q.debugEnabled {
		if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
			log.Infof("[%s]: +waitUntilDone(%s)", q.name, item)
		}
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if q.debugEnabled {
		if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
			defer log.Infof("[%s]: -waitUntilDone(%s)", q.name, item)
		}
	}

	for q.expected.Has(item) || q.dirty.Has(item) || q.processing.Has(item) {
		q.cond.Wait()

		if q.debugEnabled {
			if item == "helmcharts:default:multitude-of-charts/redis-11:14.4.0" {
				log.Infof("[%s]: waitUntilDone(%s) %s", q.name, item, q.prettyPrintAll())
			}
		}
	}
}

// this func is the added feature that was missing in k8s workqueue
func (q *Type) reset() {
	if q.debugEnabled {
		log.Infof("[%s]: +reset()", q.name)
	}
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.queue = []string{}
	q.dirty = sets.String{}
	q.processing = sets.String{}
	// we are intentionally not resetting q.expected as we don't want to lose
	// those accross resync's
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
