/*
Copyright 2021 VMware. All Rights Reserved.

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

package server

import (
	"fmt"
	"html"
	"net/http"

	log "k8s.io/klog/v2"
)

// Serve is the root command that is run when no other sub-commands are present.
// It runs the gRPC service, registering the configured plugins.
func Serve(port int, pluginDirs []string) {
	// Stub echo http server
	listenAddr := fmt.Sprintf(":%d", port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Infof("Handling %q", r.URL.Path)
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	log.Infof("Starting server on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
