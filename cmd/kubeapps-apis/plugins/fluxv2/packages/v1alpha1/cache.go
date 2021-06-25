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
	"sync"

	"github.com/go-redis/redis"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
)

type FluxPlugInCache struct {
	watcherStarted bool
	// to prevent multiple watchers
	mutex    sync.Mutex
	redisCli *redis.Client
}

func NewCache() *FluxPlugInCache {
	log.Infof("+fluxv2 New FluxPlugInCache")
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "kubeapps-redis-master.kubeapps.svc.cluster.local:6379",
		Password: "5hcCT6srwP", // TODO found that value thru kubeapps UX, need to obtain programmatically
		DB:       0,            // use default DB
	})
	c := FluxPlugInCache{
		watcherStarted: false,
		mutex:          sync.Mutex{},
		redisCli:       redisCli,
	}
	go c.startHelmRepositoryWatcher()
	return &c
}

func (c *FluxPlugInCache) startHelmRepositoryWatcher() {
	log.Infof("+fluxv2 startHelmRepositoryWatcher")
	c.mutex.Lock()

	if !c.watcherStarted {
		ch, err := c.newHelmRepositoryWatcherChan()
		if err != nil {
			c.mutex.Unlock()
			log.Errorf("failed to start watcher %w", err)
			return
		}
		c.watcherStarted = true
		c.mutex.Unlock()
		log.Infof("watcher successfully started. waiting for events...")

		c.processEvents(ch)
	} else {
		c.mutex.Unlock()
		log.Infof("watcher already started. exiting...")
	}
	// we should never reach here
	log.Infof("-fluxv2 startHelmRepositoryWatcher")
}

func (c *FluxPlugInCache) newHelmRepositoryWatcherChan() (<-chan watch.Event, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	repositoriesResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	ctx := context.Background()

	watcher, err := client.Resource(repositoriesResource).Namespace("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return watcher.ResultChan(), nil
}

func (c *FluxPlugInCache) processEvents(ch <-chan watch.Event) {
	for {
		event := <-ch
		log.Infof("got event: type: [%v] object: [%v]", event.Type, event.Object)
		if event.Type == watch.Added {
			unstructuredRepo, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				log.Errorf("Could not cast to unstructured.Unstructured")
			}
			go c.processNewRepo(unstructuredRepo.Object)
		} else {
			// TODO handle other kinds of events
			log.Errorf("got unexpected event: %v", event)
		}
	}
}

func (c *FluxPlugInCache) processNewRepo(unstructuredRepo map[string]interface{}) {
	packages, err := readPackageSummariesFromOneRepo(unstructuredRepo)
	if err != nil {
		log.Errorf("Failed to processRepo: %v due to %v", unstructuredRepo, err)
		return
	}
	name, _, _ := unstructured.NestedString(unstructuredRepo, "metadata", "name")
	log.Infof("Found [%d] packages in repo: [%s]", len(packages), name)

	protoMsg := corev1.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: packages,
	}
	bytes, err := proto.Marshal(&protoMsg)
	if err != nil {
		log.Errorf("Failed to marshal due to: %v", err)
		return
	}
	err = c.redisCli.Set(name, bytes, 0).Err()
	if err != nil {
		log.Errorf("Failed to set value for repository [%s] in cache due to %v", name, err)
		return
	}

	// read back from cache, should be sane as what we wrote
	bytes, err = c.redisCli.Get(name).Bytes()
	if err != nil {
		log.Errorf("Failed to get value for repository [%s] from cache due to %v", name, err)
		return
	}

	err = proto.Unmarshal(bytes, &protoMsg)
	if err != nil {
		log.Errorf("Failed to unmarshal bytes for repository [%s] in cache due to %v", name, err)
		return
	}
	log.Infof("Unmarshalled [%d] packages in repo: [%s]", len(protoMsg.AvailablePackagesSummaries), name)
}
