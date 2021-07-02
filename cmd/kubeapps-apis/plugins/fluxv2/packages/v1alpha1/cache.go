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

type fluxPlugInCache struct {
	// to prevent multiple watchers
	watcherStarted bool
	// this mutex guards watcherStarted var
	watcherMutex sync.Mutex
	redisCli     *redis.Client
	clientGetter server.KubernetesClientGetter
	// these waitGroup is used exclusively by unit tests to block until all expected repos have
	// been indexed by the go routine running in the background. The creation of the WaitGroup object
	// and .Add() is expected to be done by the unit test client. The server-side only signals
	// when indexing one repo is complete
	indexRepoWaitGroup *sync.WaitGroup
}

func newCache(clientGetter server.KubernetesClientGetter) (*fluxPlugInCache, error) {
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

	log.Infof("newCache: addr: [%s], password: [%s], DB=[%d]", REDIS_ADDR, REDIS_PASSWORD, REDIS_DB)

	return newCacheWithRedisClient(
		clientGetter,
		redis.NewClient(&redis.Options{
			Addr:     REDIS_ADDR,
			Password: REDIS_PASSWORD,
			DB:       REDIS_DB_NUM,
		}))
}

func newCacheWithRedisClient(clientGetter server.KubernetesClientGetter, redisCli *redis.Client) (*fluxPlugInCache, error) {
	log.Infof("+newCacheWithRedisClient")

	if redisCli == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with redis Client")
	}

	if clientGetter == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "server not configured with configGetter")
	}

	// sanity check that the redis client is connected
	pong, err := redisCli.Ping(redisCli.Context()).Result()
	if err != nil {
		return nil, err
	}
	log.Infof("[PING] -> [%s]", pong)

	c := fluxPlugInCache{
		watcherStarted: false,
		watcherMutex:   sync.Mutex{},
		redisCli:       redisCli,
		clientGetter:   clientGetter,
	}
	go c.startHelmRepositoryWatcher()
	return &c, nil
}

func (c *fluxPlugInCache) startHelmRepositoryWatcher() {
	log.Infof("+fluxv2 startHelmRepositoryWatcher")
	c.watcherMutex.Lock()
	// can't defer c.watcherMutex.Unlock() because when all is well,
	// we never return from this func

	if !c.watcherStarted {
		ch, err := c.newHelmRepositoryWatcherChan()
		if err != nil {
			c.watcherMutex.Unlock()
			log.Errorf("failed to start HelmRepository watcher due to: %v", err)
			return
		}
		c.watcherStarted = true
		c.watcherMutex.Unlock()
		log.Infof("watcher successfully started. waiting for events...")

		c.processEvents(ch)
	} else {
		c.watcherMutex.Unlock()
		log.Infof("watcher already started. exiting...")
	}
	// we should never reach here under normal usage
	log.Warningf("-fluxv2 startHelmRepositoryWatcher")
}

func (c *fluxPlugInCache) newHelmRepositoryWatcherChan() (<-chan watch.Event, error) {
	// TODO this is a temp hack to get around the fact that the only clientGetter we've got today
	// always expects authorization Bearer token in the context
	ctx := metadata.NewIncomingContext(context.TODO(), metadata.MD{
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

// this is an infinite loop that waits for new events and processes them when they happen
func (c *fluxPlugInCache) processEvents(ch <-chan watch.Event) {
	for {
		event := <-ch
		if event.Type == "" {
			// not quite sure why this happens (the docs don't say), but it seems to happen quite often
			continue
		}
		log.Infof("got event: type: [%v] object:\n[%s]", event.Type, prettyPrintObject(event.Object))
		if event.Type == watch.Added || event.Type == watch.Modified {
			unstructuredRepo, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				log.Errorf("Could not cast to unstructured.Unstructured")
			} else {
				go c.onAddOrModifyRepo(unstructuredRepo.Object)
			}
		} else {
			// TODO handle other kinds of events
			log.Errorf("got unexpected event: %v", event)
		}
	}
}

// this is effectively a cache PUT operation
func (c *fluxPlugInCache) onAddOrModifyRepo(unstructuredRepo map[string]interface{}) {
	defer func() {
		if c.indexRepoWaitGroup != nil {
			c.indexRepoWaitGroup.Done()
		}
	}()

	startTime := time.Now()
	ready, err := isRepoReady(unstructuredRepo)
	if err != nil {
		log.Errorf("Failed to process repo %s\ndue to: %v", prettyPrintMap(unstructuredRepo), err)
		return
	}

	key, err := helmRepoRedisKey(unstructuredRepo)
	if err != nil {
		log.Errorf("Failed to get redis key due to: %v", err)
		return
	}

	if ready {
		packages, err := indexOneRepo(unstructuredRepo)
		if err != nil {
			log.Errorf("Failed to process repo %s\ndue to: %v", prettyPrintMap(unstructuredRepo), err)
			return
		}
		protoMsg := corev1.GetAvailablePackageSummariesResponse{
			AvailablePackagesSummaries: packages,
		}
		bytes, err := proto.Marshal(&protoMsg)
		if err != nil {
			log.Errorf("Failed to marshal due to: %v", err)
			return
		}

		err = c.redisCli.Set(c.redisCli.Context(), *key, bytes, 0).Err()
		if err != nil {
			log.Errorf("Failed to set value for repository [%s] in cache due to: %v", *key, err)
			return
		}
		duration := time.Since(startTime)
		log.Infof("Indexed [%d] packages in repo [%s] in [%d] ms", len(packages), *key, duration.Milliseconds())
	} else {
		// repo is not quite ready to be indexed - not really an error condition,
		// just skip it eventually there will be another event when it is in ready state
		log.Infof("Skipping packages for repository [%s] because it is not in 'Ready' state", *key)

		// clear that key so cache doesn't contain any stale info for this repo
		c.redisCli.Del(c.redisCli.Context(), *key)
	}
}

// this is effectively a cache GET operation
func (c *fluxPlugInCache) packageSummariesForRepo(unstucturedRepo map[string]interface{}) ([]*corev1.AvailablePackageSummary, error) {
	startTime := time.Now()

	key, err := helmRepoRedisKey(unstucturedRepo)
	if err != nil {
		return nil, err
	}

	// read back from cache: should be bytes we previously wrote or Redis.Nil
	bytes, err := c.redisCli.Get(c.redisCli.Context(), *key).Bytes()
	if err == redis.Nil {
		// this is normal if the key does not exist
		return []*corev1.AvailablePackageSummary{}, nil
	} else if err != nil {
		log.Errorf("Failed to get value for [%s] from cache due to: %v", key, err)
		return []*corev1.AvailablePackageSummary{}, nil
	}

	var protoMsg corev1.GetAvailablePackageSummariesResponse
	err = proto.Unmarshal(bytes, &protoMsg)
	if err != nil {
		log.Errorf("Failed to unmarshal bytes for [%s] in cache due to: %v", key, err)
		return []*corev1.AvailablePackageSummary{}, nil
	}
	duration := time.Since(startTime)
	log.Infof("Unmarshalled [%d] packages for: [%s] in [%d] ms",
		len(protoMsg.AvailablePackagesSummaries), key, duration.Milliseconds())
	return protoMsg.AvailablePackagesSummaries, nil
}

const (
	// max number of concurrent workers reading repo index at the same time
	maxWorkers = 10
)

type fetchRepoJob struct {
	unstructuredRepo map[string]interface{}
}

type fetchRepoJobResult struct {
	packages []*corev1.AvailablePackageSummary
	Error    error
}

// each repo is read in a separate go routine (lightweight thread of execution)
func (c *fluxPlugInCache) fetchPackageSummaries(repoItems []unstructured.Unstructured) ([]*corev1.AvailablePackageSummary, error) {
	responsePackages := []*corev1.AvailablePackageSummary{}
	var wg sync.WaitGroup
	workers := int(math.Min(float64(len(repoItems)), float64(maxWorkers)))
	requestChan := make(chan fetchRepoJob, workers)
	responseChan := make(chan fetchRepoJobResult, workers)

	// Process only at most maxWorkers at a time
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			for job := range requestChan {
				packages, err := c.packageSummariesForRepo(job.unstructuredRepo)
				responseChan <- fetchRepoJobResult{packages, err}
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(responseChan)
	}()

	go func() {
		for _, repoItem := range repoItems {
			requestChan <- fetchRepoJob{repoItem.Object}
		}
		close(requestChan)
	}()

	// Start receiving results
	for res := range responseChan {
		if res.Error == nil {
			responsePackages = append(responsePackages, res.packages...)
		} else {
			log.Errorf("%v", res.Error)
		}
	}
	return responsePackages, nil
}

func helmRepoRedisKey(unstructuredRepo map[string]interface{}) (*string, error) {
	name, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "name")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field metadata.name not found on HelmRepository: %v:\n%s",
			err,
			prettyPrintMap(unstructuredRepo))
	}

	namespace, found, err := unstructured.NestedString(unstructuredRepo, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field metadata.namespace not found on HelmRepository: %v:\n%s",
			err,
			prettyPrintMap(unstructuredRepo))
	}

	// redis convention on key format
	// https://redis.io/topics/data-types-intro
	// Try to stick with a schema. For instance "object-type:id" is a good idea, as in "user:1000".
	// We will use "helmrepository:ns:repoName"
	s := fmt.Sprintf("%s:%s:%s", fluxHelmRepository, namespace, name)
	return &s, nil
}
