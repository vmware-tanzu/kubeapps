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
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/proto/hapi/chart"

	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
)

const (
	defaultTimeoutSeconds = 180
)

var (
	netClient *http.Client
)

func init() {
	netClient = &http.Client{
		Timeout: time.Second * defaultTimeoutSeconds,
	}
}

// Context key type for request contexts
type contextKey int

// userKey is the context key for the User data in the request context
const userKey contextKey = 0

// authGate implements middleware to check if the user is logged in before continuing
func authGate() negroni.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		authHeader := strings.Split(req.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			response.NewErrorResponse(http.StatusUnauthorized, "Unauthorized").Write(w)
			return
		}
		userAuth, err := auth.NewAuth(authHeader[1])
		if err != nil {
			response.NewErrorResponse(http.StatusInternalServerError, err.Error()).Write(w)
			return
		}
		err = userAuth.Validate()
		if err != nil {
			response.NewErrorResponse(http.StatusUnauthorized, err.Error()).Write(w)
			return
		}
		ctx := context.WithValue(req.Context(), userKey, *userAuth)
		next(w, req.WithContext(ctx))
	}
}

// Params a key-value map of path params
type Params map[string]string

// WithParams can be used to wrap handlers to take an extra arg for path params
type WithParams func(http.ResponseWriter, *http.Request, Params)

func (h WithParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	h(w, req, vars)
}

// WithoutParams can be used to wrap handlers that doesn't take params
type WithoutParams func(http.ResponseWriter, *http.Request)

func (h WithoutParams) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h(w, req)
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

func getChart(req *http.Request) (*chartUtils.Details, *chart.Chart, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}
	chartDetails, err := chartUtils.ParseDetails(body)
	if err != nil {
		return nil, nil, err
	}
	ch, err := chartUtils.GetChart(chartDetails, kubeClient, netClient, chartutil.LoadArchive)
	if err != nil {
		return nil, nil, err
	}
	return chartDetails, ch, nil
}

func logStatus(name string) {
	status, err := proxy.GetReleaseStatus(name)
	if err != nil {
		log.Printf("Unable to fetch release status of %s: %v", name, err)
	} else {
		log.Printf("Release status: %s", status)
	}
}

func returnForbiddenActions(forbiddenActions []auth.Action, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(forbiddenActions)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewErrorResponse(http.StatusForbidden, string(body)).Write(w)
}

func createRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Creating Helm Release")
	chartDetails, ch, err := getChart(req)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !disableAuth {
		manifest, err := proxy.ResolveManifest(params["namespace"], chartDetails.Values, ch)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.UserAuth)
		forbiddenActions, err := userAuth.GetForbiddenActions(params["namespace"], "create", manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		if len(forbiddenActions) > 0 {
			returnForbiddenActions(forbiddenActions, w)
			return
		}
	}
	rel, err := proxy.CreateRelease(chartDetails.ReleaseName, params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	log.Printf("Installed release %s", rel.Name)
	logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

func upgradeRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Upgrading Helm Release")
	chartDetails, ch, err := getChart(req)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !disableAuth {
		manifest, err := proxy.ResolveManifest(params["namespace"], chartDetails.Values, ch)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.UserAuth)
		forbiddenActions, err := userAuth.GetForbiddenActions(params["namespace"], "upgrade", manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		if len(forbiddenActions) > 0 {
			returnForbiddenActions(forbiddenActions, w)
			return
		}
	}
	rel, err := proxy.UpdateRelease(params["releaseName"], params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	log.Printf("Upgraded release %s", rel.Name)
	logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

func listAllReleases(w http.ResponseWriter, req *http.Request) {
	apps, err := proxy.ListReleases("", listLimit)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func listReleases(w http.ResponseWriter, req *http.Request, params Params) {
	apps, err := proxy.ListReleases(params["namespace"], listLimit)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

func getRelease(w http.ResponseWriter, req *http.Request, params Params) {
	rel, err := proxy.GetRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !disableAuth {
		manifest, err := proxy.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.UserAuth)
		forbiddenActions, err := userAuth.GetForbiddenActions(params["namespace"], "get", manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		if len(forbiddenActions) > 0 {
			returnForbiddenActions(forbiddenActions, w)
			return
		}
	}
	response.NewDataResponse(*rel).Write(w)
}

func deleteRelease(w http.ResponseWriter, req *http.Request, params Params) {
	if !disableAuth {
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
		userAuth := req.Context().Value(userKey).(auth.UserAuth)
		forbiddenActions, err := userAuth.GetForbiddenActions(params["namespace"], "delete", manifest)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		if len(forbiddenActions) > 0 {
			returnForbiddenActions(forbiddenActions, w)
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
