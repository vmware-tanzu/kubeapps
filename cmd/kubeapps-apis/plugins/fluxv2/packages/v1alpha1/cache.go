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

	"k8s.io/apimachinery/pkg/runtime/schema"
	log "k8s.io/klog/v2"
)

// global variable to prevent multiple watchers
var watcherStarted bool = false
var watcherMutex = &sync.Mutex{}

func startHelmRepositoryWatcher() {
	log.Infof("+fluxv2 startHelmRepositoryWatcher")
	watcherMutex.Lock()
	if !watcherStarted {
		watcherStarted = true
		watcherMutex.Unlock()

		repositoriesResource := schema.GroupVersionResource{
			Group:    fluxGroup,
			Version:  fluxVersion,
			Resource: fluxHelmRepositories,
		}

		client := context.Background()

		log.Infof("repositoriesResource: [%v], client: [%v]", repositoriesResource, client)

		// infinite loop that does not waste CPU cycles for now
		select {}
	} else {
		watcherMutex.Unlock()
		log.Infof("watcher already started. exiting...")
	}
	log.Infof("-fluxv2 startHelmRepositoryWatcher")
}
