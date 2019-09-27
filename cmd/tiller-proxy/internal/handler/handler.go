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

package handler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	proxy "github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// Context key type for request contexts
type contextKey int

// userKey is the context key for the User data in the request context
const userKey contextKey = 0

// AuthGate implements middleware to check if the user is logged in before continuing
func AuthGate() negroni.HandlerFunc {
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
		ctx := context.WithValue(req.Context(), userKey, userAuth)
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

func isUnprocessable(err error) bool {
	re := regexp.MustCompile(`release.*failed`)
	return re.MatchString(err.Error())
}

func errorCode(err error) int {
	return errorCodeWithDefault(err, http.StatusInternalServerError)
}

func errorCodeWithDefault(err error, defaultCode int) int {
	errCode := defaultCode
	if isAlreadyExists(err) {
		errCode = http.StatusConflict
	} else if isNotFound(err) {
		errCode = http.StatusNotFound
	} else if isForbidden(err) {
		errCode = http.StatusForbidden
	} else if isUnprocessable(err) {
		errCode = http.StatusUnprocessableEntity
	}
	return errCode
}

func getChart(req *http.Request, cu chartUtils.Resolver) (*chartUtils.Details, *chart.Chart, error) {
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}
	chartDetails, err := cu.ParseDetails(body)
	if err != nil {
		return nil, nil, err
	}
	netClient, err := cu.InitNetClient(chartDetails)
	if err != nil {
		return nil, nil, err
	}
	ch, err := cu.GetChart(chartDetails, netClient)
	if err != nil {
		return nil, nil, err
	}
	return chartDetails, ch, nil
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

// TillerProxy client and configuration
type TillerProxy struct {
	DisableAuth bool
	ListLimit   int
	ChartClient chartUtils.Resolver
	ProxyClient proxy.TillerClient
}

func (h *TillerProxy) logStatus(name string) {
	status, err := h.ProxyClient.GetReleaseStatus(name)
	if err != nil {
		log.Printf("Unable to fetch release status of %s: %v", name, err)
	} else {
		log.Printf("Release status: %s", status)
	}
}

// CreateRelease creates a new release in the namespace given as Param
func (h *TillerProxy) CreateRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Creating Helm Release")
	chartDetails, ch, err := getChart(req, h.ChartClient)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !h.DisableAuth {
		manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], chartDetails.Values, ch)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.Checker)
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
	rel, err := h.ProxyClient.CreateRelease(chartDetails.ReleaseName, params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(errorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
		return
	}
	log.Printf("Installed release %s", rel.Name)
	h.logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

// OperateRelease decides which method to call depending in the "action" query param
func (h *TillerProxy) OperateRelease(w http.ResponseWriter, req *http.Request, params Params) {
	switch req.FormValue("action") {
	case "upgrade":
		h.UpgradeRelease(w, req, params)
	case "rollback":
		h.RollbackRelease(w, req, params)
	default:
		// By default, for maintaining compatibility, we call upgrade
		h.UpgradeRelease(w, req, params)
	}
}

// RollbackRelease performs an action over a release
func (h *TillerProxy) RollbackRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Rolling back %s", params["releaseName"])
	revision := req.FormValue("revision")
	if revision == "" {
		response.NewErrorResponse(http.StatusUnprocessableEntity, "Missing revision to rollback in request").Write(w)
		return
	}
	revisionInt, err := strconv.ParseInt(revision, 10, 64)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !h.DisableAuth {
		manifest, err := h.ProxyClient.ResolveManifestFromRelease(params["releaseName"], int32(revisionInt))
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.Checker)
		// Using "upgrade" action since the concept is the same
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
	rel, err := h.ProxyClient.RollbackRelease(params["releaseName"], params["namespace"], int32(revisionInt))
	if err != nil {
		response.NewErrorResponse(errorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
		return
	}
	log.Printf("Rollback release for %s to %d", rel.Name, revisionInt)
	h.logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

// UpgradeRelease upgrades a release in the namespace given as Param
func (h *TillerProxy) UpgradeRelease(w http.ResponseWriter, req *http.Request, params Params) {
	log.Printf("Upgrading Helm Release")
	chartDetails, ch, err := getChart(req, h.ChartClient)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !h.DisableAuth {
		manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], chartDetails.Values, ch)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.Checker)
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
	rel, err := h.ProxyClient.UpdateRelease(params["releaseName"], params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(errorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
		return
	}
	log.Printf("Upgraded release %s", rel.Name)
	h.logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

// ListAllReleases list all releases that Tiller stores
func (h *TillerProxy) ListAllReleases(w http.ResponseWriter, req *http.Request) {
	apps, err := h.ProxyClient.ListReleases("", h.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

// ListReleases in the namespace given as Param
func (h *TillerProxy) ListReleases(w http.ResponseWriter, req *http.Request, params Params) {
	apps, err := h.ProxyClient.ListReleases(params["namespace"], h.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

// GetRelease returns the release info
func (h *TillerProxy) GetRelease(w http.ResponseWriter, req *http.Request, params Params) {
	rel, err := h.ProxyClient.GetRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	if !h.DisableAuth {
		manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.Checker)
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

// DeleteRelease removes a release from a namespace
func (h *TillerProxy) DeleteRelease(w http.ResponseWriter, req *http.Request, params Params) {
	if !h.DisableAuth {
		rel, err := h.ProxyClient.GetRelease(params["releaseName"], params["namespace"])
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
		if err != nil {
			response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
			return
		}
		userAuth := req.Context().Value(userKey).(auth.Checker)
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
	purge := false
	if req.URL.Query().Get("purge") == "1" || req.URL.Query().Get("purge") == "true" {
		purge = true
	}
	err := h.ProxyClient.DeleteRelease(params["releaseName"], params["namespace"], purge)
	if err != nil {
		response.NewErrorResponse(errorCode(err), err.Error()).Write(w)
		return
	}
	w.Header().Set("Status-Code", "200")
	w.Write([]byte("OK"))
}
