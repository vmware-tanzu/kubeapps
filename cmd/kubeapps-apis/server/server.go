// Copyright 2021 the Kubeapps contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
