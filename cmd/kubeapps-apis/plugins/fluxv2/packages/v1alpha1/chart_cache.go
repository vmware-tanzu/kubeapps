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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/cache"
	"github.com/kubeapps/kubeapps/pkg/chart/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	log "k8s.io/klog/v2"
)

// For now:
// unlike NamespacedResourceWatcherCache this is not a general purpose cache meant to
// be re-used. It is written specifically for one purpose and has ties into the internals of
// repo and chart.go. So it exists outside the cache package

type ChartCache struct {
	redisCli *redis.Client

	// queue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time and makes it easy to ensure we are never processing the same item
	// simultaneously in different workers.
	queue cache.RateLimitingInterface
}

func NewChartCache(redisCli *redis.Client) (*ChartCache, error) {
	log.Infof("+NewChartCache(%v)", redisCli)

	if redisCli == nil {
		return nil, fmt.Errorf("server not configured with redis client")
	}

	c := ChartCache{
		redisCli: redisCli,
		queue:    cache.NewRateLimitingQueue(),
	}

	// dummy channel for now. Ideally, this would be passed in as an input argument
	// and the caller would indicate when to stop
	stopCh := make(chan struct{})

	// this will launch a single worker that processes items on the work queue as they come in
	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	go wait.Until(c.runWorker, time.Second, stopCh)

	return &c, nil
}

func (c *ChartCache) wrapOnAddFunc(fnAdd cache.ValueAdderFunc, fnGet cache.ValueGetterFunc) cache.ValueAdderFunc {
	return func(key string, obj map[string]interface{}) (interface{}, bool, error) {
		value, setValue, err := fnAdd(key, obj)
		if err == nil && setValue {
			if untypedValue, err2 := fnGet(key, value); err2 != nil {
				log.Errorf("%+v", err2)
			} else if typedValue, ok := untypedValue.(repoCacheEntry); !ok {
				log.Errorf("unexpected value fetched from cache: type: [%s], value: [%v]",
					reflect.TypeOf(untypedValue), value)
			} else {
				c.syncCharts(typedValue.Charts)
			}
		}
		return value, setValue, err
	}
}

// this will enqueue work items into chart work queue and return
// the charts will be synced by a worker thread running in the background
func (c *ChartCache) syncCharts(charts []models.Chart) {
	log.Infof("+syncCharts()")
	defer log.Infof("-syncCharts()")

	// let's just cache the latest one for now. The chart versions array would
	// have already been sorted and the latest chart version will be at array index 0
	for _, chart := range charts {
		c.queue.AddRateLimited(c.keyFor(chart))
	}
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
			log.Infof("Done syncing key [%s]", key)
			return nil
		}
	}(obj)

	if err != nil {
		runtime.HandleError(err)
	}
	return true
}

// syncs the current state of the given resource in k8s with that in the cache
func (c *ChartCache) syncHandler(key string) error {
	log.Infof("+syncHandler(%s)", key)
	defer log.Infof("-syncHandler(%s)", key)

	/* pulling this work on hold for the time being to get this PR in
	_, chartID, _, err := c.fromKey(key)
	if err != nil {
		return err
	}

	chartDetail, err := tar.FetchChartDetailFromTarball(chartID, tarUrl, "", "", httpclient.New())
	if err != nil {
		return err
	}
	*/

	return nil
}

func (c *ChartCache) keyFor(chart models.Chart) string {
	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmcharts:ns:chartID:chartVersion"
	// notice that chartID is of the form "repoName/id", so it includes the repo name
	return fmt.Sprintf("helmcharts:%s:%s:%s", chart.Repo.Namespace, chart.ID, chart.ChartVersions[0].Version)
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
