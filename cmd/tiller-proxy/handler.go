/*
Copyright (c) 2018 Bitnami

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
	"io/ioutil"
	"k8s.io/helm/pkg/proto/hapi/release"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	log "github.com/sirupsen/logrus"
)

// Params a key-value map of path params
type Params map[string]string

// WithParams can be used to wrap handlers to take an extra arg for path params
type WithParams func(http.ResponseWriter, *http.Request, Params)

func (h WithParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(w, req, vars)
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

func isAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "is still in use") || strings.Contains(err.Error(), "already exists")
}

func deployRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Creating/updating Helm Release")
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
		return
	}
	var rel *release.Release
	if req.Method == "POST" && params["releaseName"] == "" {
		rel, err = proxy.CreateRelease(params["namespace"], body)
	} else {
		rel, err = proxy.UpdateRelease(params["releaseName"], params["namespace"], body)
	}
	if err != nil {
		errCode := http.StatusInternalServerError
		if isAlreadyExists(err) {
			errCode = http.StatusConflict
		} else if isNotFound(err) {
			errCode = http.StatusNotFound
		}
		response.NewErrorResponse(errCode, err.Error()).Write(w)
		return
	}
	log.Printf("Installed/updated release %s", rel.Name)
	proxy.LogReleaseStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

func listAllReleases(w http.ResponseWriter, req *http.Request) {
	log.Printf("Listing All Helm Releases")
	apps, err := proxy.ListReleases("")
	if err != nil {
		response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func listReleases(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Listing Helm Releases of the namespace %s", params["namespace"])
	apps, err := proxy.ListReleases(params["namespace"])
	if err != nil {
		response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func getRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Getting Helm Release %s", params["releaseName"])
	rel, err := proxy.GetRelease(params["releaseName"], params["namespace"])
	if err != nil {
		errCode := http.StatusInternalServerError
		if isNotFound(err) {
			errCode = http.StatusNotFound
		}
		response.NewErrorResponse(errCode, err.Error()).Write(w)
		return
	}
	response.NewDataResponse(*rel).Write(w)
}

func deleteRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Deleting Helm Release %s", params["releaseName"])
	err := proxy.DeleteRelease(params["releaseName"], params["namespace"])
	if err != nil {
		errCode := http.StatusInternalServerError
		if isNotFound(err) {
			errCode = http.StatusNotFound
		}
		response.NewErrorResponse(errCode, err.Error()).Write(w)
		return
	}
	w.Write([]byte("OK"))
}
