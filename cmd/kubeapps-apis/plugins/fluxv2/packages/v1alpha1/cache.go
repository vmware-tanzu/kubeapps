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
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	log "k8s.io/klog/v2"
)

type FluxPlugInCache struct {
	watcherStarted bool
	// to prevent multiple watchers
	mutex        sync.Mutex
	redisCli     *redis.Client
	clientGetter server.KubernetesClientGetter
}

func NewCache(clientGetter server.KubernetesClientGetter) (*FluxPlugInCache, error) {
	log.Infof("+NewCache")
	REDIS_ADDR, ok := os.LookupEnv("REDIS_ADDR")
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing environment variable REDIS_ADDR")
	}
	REDIS_PASSWORD, ok := os.LookupEnv("REDIS_PASSWORD")
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing environment variable REDIS_PASSWORD")
	}
	REDIS_DB, ok := os.LookupEnv("REDIS_DB")
	if !ok {
		return nil, status.Errorf(codes.Internal, "missing environment variable REDIS_DB")
	}

	REDIS_DB_NUM, err := strconv.Atoi(REDIS_DB)
	if err != nil {
		return nil, err
	}

	log.Infof("NewCache: addr: [%s], password: [%s], DB=[%d]", REDIS_ADDR, REDIS_PASSWORD, REDIS_DB)

	return NewCacheWithRedisClient(
		clientGetter,
		redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Password: REDIS_PASSWORD,
			DB:       REDIS_DB_NUM,
		}))
}

func NewCacheWithRedisClient(clientGetter server.KubernetesClientGetter, redisCli *redis.Client) (*FluxPlugInCache, error) {
	log.Infof("+NewCacheWithRedisClient")

	if redisCli == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with redis Client")
	}

	if clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}

	// sanity check that the redis client is connected
	pong, err := redisCli.Ping(redisCli.Context()).Result()
	if err != nil {
		return nil, err
	}
	log.Infof("[PING] -> [%s]", pong)

	c := FluxPlugInCache{
		watcherStarted: false,
		mutex:          sync.Mutex{},
		redisCli:       redisCli,
		clientGetter:   clientGetter,
	}
	go c.startHelmRepositoryWatcher()
	return &c, nil
}

func (c *FluxPlugInCache) startHelmRepositoryWatcher() {
	log.Infof("+fluxv2 startHelmRepositoryWatcher")
	c.mutex.Lock()
	// can't defer c.mutex.Unlock() because in all is well we never
	// return from this func

	if !c.watcherStarted {
		ch, err := c.newHelmRepositoryWatcherChan()
		if err != nil {
			c.mutex.Unlock()
			log.Errorf("failed to start HelmRepository watcher due to: %v", err)
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
	ctx := context.Background()
	// TODO this is a temp hack to get around the fact that the only clientGetter we've got today
	// always expects authorization Bearer token in the context
	ctx = metadata.NewIncomingContext(ctx, metadata.MD{
		"authorization": []string{"Bearer kaka"},
	})

	_, dynamicClient, err := c.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to get client due to: %v", err)
	}

	repositoriesResource := schema.GroupVersionResource{
		Group:    fluxGroup,
		Version:  fluxVersion,
		Resource: fluxHelmRepositories,
	}

	watcher, err := dynamicClient.Resource(repositoriesResource).Namespace("").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return watcher.ResultChan(), nil
}

func (c *FluxPlugInCache) processEvents(ch <-chan watch.Event) {
	for {
		event := <-ch
		prettyBytes, _ := json.MarshalIndent(event.Object, "", "  ")
		log.Infof("got event: type: [%v] object:\n[%v]", event.Type, string(prettyBytes))
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

// this is effectively a cache PUT operation
func (c *FluxPlugInCache) processNewRepo(unstructuredRepo map[string]interface{}) {
	startTime := time.Now()
	packages, err := indexOneRepo(unstructuredRepo)
	if err != nil {
		log.Errorf("Failed to processRepo %v due to: %v", unstructuredRepo, err)
		return
	}
	name, _, _ := unstructured.NestedString(unstructuredRepo, "metadata", "name")

	protoMsg := corev1.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: packages,
	}
	bytes, err := proto.Marshal(&protoMsg)
	if err != nil {
		log.Errorf("Failed to marshal due to: %v", err)
		return
	}
	err = c.redisCli.Set(c.redisCli.Context(), name, bytes, 0).Err()
	if err != nil {
		log.Errorf("Failed to set value for repository [%s] in cache due to: %v", name, err)
		return
	}
	duration := time.Since(startTime)
	log.Infof("Indexed [%d] packages in repo [%s] in [%d] ms", len(packages), name, duration.Milliseconds())
}

// this is effectively a cache GET operation
func (c *FluxPlugInCache) packageSummariesForRepo(name string) []*corev1.AvailablePackageSummary {
	// read back from cache, should be sane as what we wrote
	bytes, err := c.redisCli.Get(c.redisCli.Context(), name).Bytes()
	if err != nil {
		log.Errorf("Failed to get value for repository [%s] from cache due to: %v", name, err)
		return []*corev1.AvailablePackageSummary{}
	}

	var protoMsg corev1.GetAvailablePackageSummariesResponse
	err = proto.Unmarshal(bytes, &protoMsg)
	if err != nil {
		log.Errorf("Failed to unmarshal bytes for repository [%s] in cache due to: %v", name, err)
		return []*corev1.AvailablePackageSummary{}
	}
	log.Infof("Unmarshalled [%d] packages in repo: [%s]", len(protoMsg.AvailablePackagesSummaries), name)

	return protoMsg.AvailablePackagesSummaries
}
