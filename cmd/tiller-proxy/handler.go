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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	log "github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"

	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	netClient  *http.Client
	enableAuth bool
)

func init() {
	netClient = &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
	}
	enableAuth, _ = strconv.ParseBool(os.Getenv("ENABLE_AUTH"))
}

// Params a key-value map of path params
type Params map[string]string

// WithAuth can be used to wrap handlers to take an extra arg for path params
type WithAuth func(http.ResponseWriter, *http.Request, Params, *auth.UserAuth)

func (h WithAuth) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	authHeader := strings.Split(req.Header.Get("Authorization"), "Bearer ")
	if len(authHeader) != 2 {
		response.NewErrorResponse(http.StatusUnauthorized, "Unauthorized").Write(w)
		return
	}
	userAuth, err := auth.NewAuth(authHeader[1])
	if err != nil {
		response.NewErrorResponse(http.StatusUnauthorized, err.Error()).Write(w)
		return
	}
	h(w, req, vars, userAuth)
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

func isAlreadyExists(err error) bool {
	return strings.Contains(err.Error(), "is still in use") || strings.Contains(err.Error(), "already exists")
}

func isForbidden(err error) bool {
	return strings.Contains(err.Error(), "Unauthorized")
}

func errorCode(err error) int {
	errCode := http.StatusInternalServerError
	if isAlreadyExists(err) {
		errCode = http.StatusConflict
	} else if isNotFound(err) {
		errCode = http.StatusNotFound
	} else if isForbidden(err) {
		errCode = http.StatusForbidden
	}
	return errCode
}

func deployRelease(w http.ResponseWriter, req *http.Request, params Params, user *auth.UserAuth) {
	log.Printf("Creating/updating Helm Release")
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	chartDetails, err := chartUtils.ParseDetails(body)
	if err != nil {
		response.NewErrorResponse(http.StatusUnprocessableEntity, err.Error()).Write(w)
		return
	}
	ch, err := chartUtils.GetChart(chartDetails, kubeClient, netClient, chartutil.LoadArchive)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	action := "create"
	if req.Method == "PUT" && params["releaseName"] != "" {
		action = "upgrade"
	}
	if enableAuth {
		manifest, err := proxy.ResolveManifest(params["namespace"], chartDetails.Values, ch)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		err = user.CanI(params["namespace"], action, manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
	}
	var rel *release.Release
	if action == "create" {
		rel, err = proxy.CreateRelease(
			chartDetails.ReleaseName,
			params["namespace"],
			chartDetails.Values,
			ch,
		)
	} else {
		rel, err = proxy.UpdateRelease(
			params["releaseName"],
			params["namespace"],
			chartDetails.Values,
			ch,
		)
	}
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	log.Printf("Installed/updated release %s", rel.Name)
	status, err := proxy.GetReleaseStatus(rel.Name)
	if err != nil {
		log.Printf("Unable to fecth release status of %s: %v", rel.Name, err)
	} else {
		log.Printf("Release status: %s", status)
	}
	response.NewDataResponse(*rel).Write(w)
}

func listAllReleases(w http.ResponseWriter, req *http.Request, params Params, user *auth.UserAuth) {
	apps, err := proxy.ListReleases("")
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func listReleases(w http.ResponseWriter, req *http.Request, params Params, user *auth.UserAuth) {
	apps, err := proxy.ListReleases(params["namespace"])
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func getRelease(w http.ResponseWriter, req *http.Request, params Params, user *auth.UserAuth) {
	rel, err := proxy.GetRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if enableAuth {
		manifest, err := proxy.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		err = user.CanI(params["namespace"], "get", manifest)
	}
	response.NewDataResponse(*rel).Write(w)
}

func deleteRelease(w http.ResponseWriter, req *http.Request, params Params, user *auth.UserAuth) {
	if enableAuth {
		rel, err := proxy.GetRelease(params["releaseName"], params["namespace"])
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		manifest, err := proxy.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		err = user.CanI(params["namespace"], "delete", manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
	}
	err := proxy.DeleteRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	w.Write([]byte("OK"))
}
