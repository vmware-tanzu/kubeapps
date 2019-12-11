/*
Copyright (c) 2019 Bitnami

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

package handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// AppRepositories handles http requests for operating on app repositories
// in Kubeapps, without exposing implementation details to 3rd party integrations.
type AppRepositories struct{}

func (a *AppRepositories) Create(w http.ResponseWriter, req *http.Request) {
	log.Printf("Creating AppRepository")
	// TODO: Create AppRepository using the k8s client.
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("OK"))
}
