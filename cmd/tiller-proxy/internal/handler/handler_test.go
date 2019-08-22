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
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"

	"github.com/kubeapps/kubeapps/pkg/auth"
	authFake "github.com/kubeapps/kubeapps/pkg/auth/fake"
	chartFake "github.com/kubeapps/kubeapps/pkg/chart/fake"
	proxyFake "github.com/kubeapps/kubeapps/pkg/proxy/fake"
)

func TestErrorCodeWithDefault(t *testing.T) {
	type test struct {
		err          error
		defaultCode  int
		expectedCode int
	}
	tests := []test{
		{fmt.Errorf("a release named foo already exists"), http.StatusInternalServerError, http.StatusConflict},
		{fmt.Errorf("release foo not found"), http.StatusInternalServerError, http.StatusNotFound},
		{fmt.Errorf("Unauthorized to get release foo"), http.StatusInternalServerError, http.StatusForbidden},
		{fmt.Errorf("release \"Foo \" failed"), http.StatusInternalServerError, http.StatusUnprocessableEntity},
		{fmt.Errorf("This is an unexpected error"), http.StatusInternalServerError, http.StatusInternalServerError},
		{fmt.Errorf("This is an unexpected error"), http.StatusUnprocessableEntity, http.StatusUnprocessableEntity},
	}
	for _, s := range tests {
		code := errorCodeWithDefault(s.err, s.defaultCode)
		if code != s.expectedCode {
			t.Errorf("Expected '%v' to return code %v got %v", s.err, s.expectedCode, code)
		}
	}
}

func TestActions(t *testing.T) {
	type testScenario struct {
		// Scenario params
		Description      string
		ExistingReleases []release.Release
		DisableAuth      bool
		ForbiddenActions []auth.Action
		// Request params
		RequestBody  string
		RequestQuery string
		Action       string
		Params       map[string]string
		// Expected result
		StatusCode        int
		RemainingReleases []release.Release
		ResponseBody      string // Optional
	}
	tests := []testScenario{
		{
			// Scenario params
			Description:      "Create a simple release without auth",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Create a simple release with auth",
			ExistingReleases: []release.Release{},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Create a conflicting release",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 409,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
			},
			ResponseBody: "",
		},
		{
			// Scenario params
			Description:      "Create a simple release with forbidden actions",
			ExistingReleases: []release.Release{},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{
				auth.Action{APIVersion: "v1", Resource: "pods", Namespace: "default", ClusterWide: false, Verbs: []string{"create"}},
			},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "create",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode:        403,
			RemainingReleases: []release.Release{},
			ResponseBody:      `{"code":403,"message":"[{\"apiGroup\":\"v1\",\"resource\":\"pods\",\"namespace\":\"default\",\"clusterWide\":false,\"verbs\":[\"create\"]}]"}`,
		},
		{
			// Scenario params
			Description:      "Upgrade a simple release",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "upgrade",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Upgrade a missing release",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "upgrade",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Upgrade a simple release with forbidden actions",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{
				auth.Action{APIVersion: "v1", Resource: "pods", Namespace: "default", ClusterWide: false, Verbs: []string{"upgrade"}},
			},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "upgrade",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode:        403,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			ResponseBody:      `{"code":403,"message":"[{\"apiGroup\":\"v1\",\"resource\":\"pods\",\"namespace\":\"default\",\"clusterWide\":false,\"verbs\":[\"upgrade\"]}]"}`,
		},
		{
			// Scenario params
			Description:      "Delete a simple release",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DELETED}}}},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Delete and purge a simple release",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "?purge=true",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Delete a missing release",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Delete a release with forbidden actions",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default", Config: &chart.Config{Raw: ""}}},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{
				auth.Action{APIVersion: "v1", Resource: "pods", Namespace: "default", ClusterWide: false, Verbs: []string{"delete"}},
			},
			// Request params
			RequestBody: `{"chartName": "foo", "releaseName": "foobar",	"version": "1.0.0"}`,
			RequestQuery: "",
			Action:       "delete",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        403,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default", Config: &chart.Config{Raw: ""}}},
			ResponseBody:      `{"code":403,"message":"[{\"apiGroup\":\"v1\",\"resource\":\"pods\",\"namespace\":\"default\",\"clusterWide\":false,\"verbs\":[\"delete\"]}]"}`,
		},
		{
			// Scenario params
			Description:      "Get a simple release",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        200,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default"}},
			ResponseBody:      `{"data":{"name":"foobar","namespace":"default"}}`,
		},
		{
			// Scenario params
			Description:      "Get a missing release",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Get a release with forbidden actions",
			ExistingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default", Config: &chart.Config{Raw: ""}}},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{
				auth.Action{APIVersion: "v1", Resource: "pods", Namespace: "default", ClusterWide: false, Verbs: []string{"get"}},
			},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "get",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        403,
			RemainingReleases: []release.Release{release.Release{Name: "foobar", Namespace: "default", Config: &chart.Config{Raw: ""}}},
			ResponseBody:      `{"code":403,"message":"[{\"apiGroup\":\"v1\",\"resource\":\"pods\",\"namespace\":\"default\",\"clusterWide\":false,\"verbs\":[\"get\"]}]"}`,
		},
		{
			// Scenario params
			Description: "List all releases",
			ExistingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
				release.Release{Name: "foo", Namespace: "not-default"},
			},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "listall",
			Params:       map[string]string{},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
				release.Release{Name: "foo", Namespace: "not-default"},
			},
			ResponseBody: `{"data":[{"releaseName":"foobar","version":"","namespace":"default","status":"DEPLOYED","chart":"","chartMetadata":{}},{"releaseName":"foo","version":"","namespace":"not-default","status":"DEPLOYED","chart":"","chartMetadata":{}}]}`,
		},
		{
			// Scenario params
			Description: "List releases in a namespace",
			ExistingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
				release.Release{Name: "foo", Namespace: "not-default"},
			},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "list",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default"},
				release.Release{Name: "foo", Namespace: "not-default"},
			},
			ResponseBody: `{"data":[{"releaseName":"foobar","version":"","namespace":"default","status":"DEPLOYED","chart":"","chartMetadata":{}}]}`,
		},
		{
			// Scenario params
			Description: "Filter releases based on status when listing",
			ExistingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DEPLOYED}}},
				release.Release{Name: "foo", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DELETED}}},
			},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "?statuses=deployed",
			Action:       "list",
			Params:       map[string]string{"namespace": "default"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foobar", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DEPLOYED}}},
				release.Release{Name: "foo", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DELETED}}},
			},
			ResponseBody: `{"data":[{"releaseName":"foobar","version":"","namespace":"default","status":"DEPLOYED","chart":"","chartMetadata":{}}]}`,
		},
		{
			// Scenario params
			Description: "Rolls back a release",
			ExistingReleases: []release.Release{
				release.Release{Name: "foo", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DEPLOYED}}},
			},
			DisableAuth:      false,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "?revision=1",
			Action:       "rollback",
			Params:       map[string]string{"namespace": "default", "releaseName": "foo"},
			// Expected result
			StatusCode: 200,
			RemainingReleases: []release.Release{
				release.Release{Name: "foo", Namespace: "default", Info: &release.Info{Status: &release.Status{Code: release.Status_DEPLOYED}}},
			},
			ResponseBody: `{"data":{"name":"foo","info":{"status":{"code":1}},"namespace":"default"}}`,
		},
		{
			// Scenario params
			Description:      "Rollsback a missing release",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "?revision=1",
			Action:       "rollback",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        404,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
		{
			// Scenario params
			Description:      "Rollback without a revision",
			ExistingReleases: []release.Release{},
			DisableAuth:      true,
			ForbiddenActions: []auth.Action{},
			// Request params
			RequestBody:  "",
			RequestQuery: "",
			Action:       "rollback",
			Params:       map[string]string{"namespace": "default", "releaseName": "foobar"},
			// Expected result
			StatusCode:        422,
			RemainingReleases: []release.Release{},
			ResponseBody:      "",
		},
	}
	for _, test := range tests {
		// Prepare environment
		proxy := &proxyFake.FakeProxy{
			Releases: test.ExistingReleases,
		}
		handler := TillerProxy{
			DisableAuth: test.DisableAuth,
			ListLimit:   255,
			ChartClient: &chartFake.FakeChart{},
			ProxyClient: proxy,
		}
		req := httptest.NewRequest("GET", fmt.Sprintf("http://foo.bar%s", test.RequestQuery), strings.NewReader(test.RequestBody))
		if !test.DisableAuth {
			fauth := &authFake.FakeAuth{
				ForbiddenActions: test.ForbiddenActions,
			}
			ctx := context.WithValue(req.Context(), userKey, fauth)
			req = req.WithContext(ctx)
		}
		response := httptest.NewRecorder()
		// Perform request
		t.Log(test.Description)
		switch test.Action {
		case "create":
			handler.CreateRelease(response, req, test.Params)
		case "upgrade":
			handler.UpgradeRelease(response, req, test.Params)
		case "delete":
			handler.DeleteRelease(response, req, test.Params)
		case "get":
			handler.GetRelease(response, req, test.Params)
		case "rollback":
			handler.RollbackRelease(response, req, test.Params)
		case "list":
			handler.ListReleases(response, req, test.Params)
		case "listall":
			handler.ListAllReleases(response, req)
		default:
			t.Errorf("Unexpected action %s", test.Action)
		}
		// Check result
		if response.Code != test.StatusCode {
			t.Errorf("Expecting a StatusCode %d, received %d", test.StatusCode, response.Code)
		}
		if !reflect.DeepEqual(proxy.Releases, test.RemainingReleases) {
			t.Errorf("Unexpected remaining releases. Expecting %v, found %v", test.RemainingReleases, proxy.Releases)
		}
		if test.ResponseBody != "" {
			if test.ResponseBody != response.Body.String() {
				t.Errorf("Unexpected body response. Expecting %s, found %s", test.ResponseBody, response.Body)
			}
		}
	}
}
