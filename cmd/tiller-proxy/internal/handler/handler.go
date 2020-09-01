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
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kubeapps/common/response"
	"github.com/kubeapps/kubeapps/pkg/auth"
	chartUtils "github.com/kubeapps/kubeapps/pkg/chart"
	"github.com/kubeapps/kubeapps/pkg/handlerutil"
	proxy "github.com/kubeapps/kubeapps/pkg/proxy"
	log "github.com/sirupsen/logrus"
)

const requireV1Support = true

func returnForbiddenActions(forbiddenActions []auth.Action, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(forbiddenActions)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewErrorResponse(http.StatusForbidden, string(body)).Write(w)
}

// TillerProxy client and configuration
type TillerProxy struct {
	CheckerForRequest auth.CheckerForRequest
	ListLimit         int
	ChartClient       chartUtils.Resolver
	ProxyClient       proxy.TillerClient
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
func (h *TillerProxy) CreateRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	log.Printf("Creating Helm Release")
	chartDetails, chartMulti, err := handlerutil.ParseAndGetChart(req, h.ChartClient, requireV1Support)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	ch := chartMulti.Helm2Chart
	manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	if !h.isAuthorizedForManifest(w, req, params["namespace"], "create", manifest) {
		return
	}

	rel, err := h.ProxyClient.CreateRelease(chartDetails.ReleaseName, params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
		return
	}
	log.Printf("Installed release %s", rel.Name)
	h.logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

// OperateRelease decides which method to call depending in the "action" query param
func (h *TillerProxy) OperateRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	switch req.FormValue("action") {
	case "upgrade":
		h.UpgradeRelease(w, req, params)
	case "rollback":
		h.RollbackRelease(w, req, params)
	case "test":
		h.TestRelease(w, req, params)
	default:
		// By default, for maintaining compatibility, we call upgrade
		h.UpgradeRelease(w, req, params)
	}
}

// RollbackRelease performs an action over a release
func (h *TillerProxy) RollbackRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	log.Printf("Rolling back %s", params["releaseName"])
	revision := req.FormValue("revision")
	if revision == "" {
		response.NewErrorResponse(http.StatusUnprocessableEntity, "Missing revision to rollback in request").Write(w)
		return
	}
	revisionInt, err := strconv.ParseInt(revision, 10, 64)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	manifest, err := h.ProxyClient.ResolveManifestFromRelease(params["releaseName"], int32(revisionInt))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}

	// Using "upgrade" action since the concept is the same
	if !h.isAuthorizedForManifest(w, req, params["namespace"], "upgrade", manifest) {
		return
	}

	rel, err := h.ProxyClient.RollbackRelease(params["releaseName"], params["namespace"], int32(revisionInt))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
		return
	}
	log.Printf("Rollback release for %s to %d", rel.Name, revisionInt)
	h.logStatus(rel.Name)
	response.NewDataResponse(*rel).Write(w)
}

// UpgradeRelease upgrades a release in the namespace given as Param
func (h *TillerProxy) UpgradeRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	log.Printf("Upgrading Helm Release")
	chartDetails, chartMulti, err := handlerutil.ParseAndGetChart(req, h.ChartClient, requireV1Support)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	ch := chartMulti.Helm2Chart
	manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	if !h.isAuthorizedForManifest(w, req, params["namespace"], "upgrade", manifest) {
		return
	}

	rel, err := h.ProxyClient.UpdateRelease(params["releaseName"], params["namespace"], chartDetails.Values, ch)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCodeWithDefault(err, http.StatusUnprocessableEntity), err.Error()).Write(w)
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
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

// ListReleases in the namespace given as Param
func (h *TillerProxy) ListReleases(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	apps, err := h.ProxyClient.ListReleases(params["namespace"], h.ListLimit, req.URL.Query().Get("statuses"))
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(apps).Write(w)
}

// TestRelease in the namespace given as Param
func (h *TillerProxy) TestRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	// helm tests only create pods so we only need to check that
	manifest := "apiVersion: v1\nkind: Pod"
	if !h.isAuthorizedForManifest(w, req, params["namespace"], "create", manifest) {
		return
	}

	testResult, err := h.ProxyClient.TestRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	response.NewDataResponse(testResult).Write(w)
}

// GetRelease returns the release info
func (h *TillerProxy) GetRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	rel, err := h.ProxyClient.GetRelease(params["releaseName"], params["namespace"])
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	manifest, err := h.ProxyClient.ResolveManifest(params["namespace"], rel.Config.Raw, rel.Chart)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	if !h.isAuthorizedForManifest(w, req, params["namespace"], "get", manifest) {
		return
	}

	response.NewDataResponse(*rel).Write(w)
}

// DeleteRelease removes a release from a namespace
func (h *TillerProxy) DeleteRelease(w http.ResponseWriter, req *http.Request, params handlerutil.Params) {
	if !h.isAuthorizedForRelease(w, req, params["namespace"], "get", params["releaseName"]) {
		return
	}
	purge := handlerutil.QueryParamIsTruthy("purge", req)
	err := h.ProxyClient.DeleteRelease(params["releaseName"], params["namespace"], purge)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return
	}
	w.Header().Set("Status-Code", "200")
	w.Write([]byte("OK"))
}

func (h *TillerProxy) isAuthorizedForRelease(w http.ResponseWriter, req *http.Request, namespace, verb, releaseName string) bool {
	forbiddenActions, err := h.forbiddenActionsForRelease(req, namespace, verb, releaseName)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return false
	}
	if len(forbiddenActions) > 0 {
		returnForbiddenActions(forbiddenActions, w)
		return false
	}
	return true
}

func (h *TillerProxy) isAuthorizedForManifest(w http.ResponseWriter, req *http.Request, namespace, verb, manifest string) bool {
	forbiddenActions, err := h.forbiddenActionsForManifest(req, namespace, verb, manifest)
	if err != nil {
		response.NewErrorResponse(handlerutil.ErrorCode(err), err.Error()).Write(w)
		return false
	}
	if len(forbiddenActions) > 0 {
		returnForbiddenActions(forbiddenActions, w)
		return false
	}
	return true
}

func (h *TillerProxy) forbiddenActionsForRelease(req *http.Request, namespace, verb, releaseName string) ([]auth.Action, error) {
	rel, err := h.ProxyClient.GetRelease(releaseName, namespace)
	if err != nil {
		return nil, err
	}
	manifest, err := h.ProxyClient.ResolveManifest(namespace, rel.Config.Raw, rel.Chart)
	if err != nil {
		return nil, err
	}
	return h.forbiddenActionsForManifest(req, namespace, verb, manifest)
}

func (h *TillerProxy) forbiddenActionsForManifest(req *http.Request, namespace, verb, manifest string) ([]auth.Action, error) {
	userAuth, err := h.CheckerForRequest(req)
	if err != nil {
		return nil, err
	}
	return userAuth.GetForbiddenActions(namespace, verb, manifest)
}
