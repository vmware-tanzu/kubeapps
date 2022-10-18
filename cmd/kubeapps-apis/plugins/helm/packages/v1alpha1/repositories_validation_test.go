// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"io"
	log "k8s.io/klog/v2"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"
)

type fakeHTTPCli struct {
	request  *http.Request
	response *http.Response
	err      error
}

func (f *fakeHTTPCli) Do(r *http.Request) (*http.Response, error) {
	f.request = r
	return f.response, f.err
}

func TestNonOCIValidate(t *testing.T) {
	validRequest, err := http.NewRequest("GET", "http://example.com/index.yaml", strings.NewReader(""))
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testCases := []struct {
		name             string
		httpValidator    HelmNonOCIValidator
		fakeHttpError    error
		fakeRepoResponse *http.Response
		expectedResponse *ValidationResponse
	}{
		{
			name:             "it returns 200 OK validation response if there is no error and the external response is 200",
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("OK")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it does not include the body of the upstream response when validation succeeds",
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("10 Mb of data")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it returns an error from the response with the body text if validation fails",
			fakeRepoResponse: &http.Response{StatusCode: 401, Body: io.NopCloser(bytes.NewReader([]byte("It failed because of X and Y")))},
			expectedResponse: &ValidationResponse{Code: 401, Message: "It failed because of X and Y"},
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
		},
		{
			name:             "it returns a 400 error if the validation cannot be run",
			fakeHttpError:    fmt.Errorf("client.Do returns an error"),
			expectedResponse: &ValidationResponse{Code: 400, Message: "client.Do returns an error"},
			httpValidator:    HelmNonOCIValidator{Req: validRequest},
		},
	}

	cmpOpts := []cmp.Option{cmpopts.IgnoreUnexported(http.Request{}, strings.Reader{})}
	cmpOpts = append(cmpOpts, cmpopts.IgnoreFields(http.Request{}, "GetBody"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClient := &fakeHTTPCli{
				response: tc.fakeRepoResponse,
				err:      tc.fakeHttpError,
			}

			response, err := tc.httpValidator.Validate(fakeClient)
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if got, want := fakeClient.request, tc.httpValidator.Req; !cmp.Equal(want, got, cmpOpts...) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, cmpOpts...))
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

type fakeOCIRepo struct {
	tags     repoTagsList
	manifest repoManifest
}

// makeTestOCIServer returns a small test double for an OCI server that handles requests
// for tags/list and manifest only.
func makeTestOCIServer(t *testing.T, registryName string, repos map[string]fakeOCIRepo, requiredAuthHeader string) *httptest.Server {
	// Define a map of valid request/responses based on the fake repos passed in.
	responses := map[string]string{}
	for repoName, repo := range repos {
		tags, err := json.Marshal(repo.tags)
		if err != nil {
			t.Fatalf("%+v", err)
		}
		responses[path.Join("/v2", registryName, repoName, "tags", "list")] = string(tags)

		manifest, err := json.Marshal(repo.manifest)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(repo.tags.Tags) > 0 {
			responses[path.Join("/v2", registryName, repoName, "manifests", repo.tags.Tags[0])] = string(manifest)
		}
	}

	// Return a test server that responds with these canned responses only.
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Required authorization when set.
		authHeader := r.Header.Get("Authorization")
		if authHeader != requiredAuthHeader {
			w.WriteHeader(401)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
		if response, ok := responses[r.URL.Path]; !ok {
			w.WriteHeader(404)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		} else {
			_, err := w.Write([]byte(response))
			if err != nil {
				log.Fatalf("%+v", err)
			}
		}
	}))
}

func TestOCIValidate(t *testing.T) {
	registryName := "bitnami"
	testCases := []struct {
		name             string
		repos            map[string]fakeOCIRepo
		validator        HelmOCIValidator
		expectedResponse *ValidationResponse
	}{
		{
			name: "it returns a valid response if all the OCI repos are of the helm type",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name: "it returns an invalid response if just one of OCI repos is of the wrong type",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.docker.container.image.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "nginx is not a Helm OCI Repo. mediaType starting with \"application/vnd.cncf.helm.config\" expected, found \"application/vnd.docker.container.image.v1+json\"",
			},
		},
		{
			name: "it returns an invalid response if a repo does not exist",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"notnginx": {},
			},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "Unexpected status code when querying \"nginx\": 404",
			},
		},
		{
			name: "it returns an invalid response if a manifest does not exist",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{"apache", "nginx"},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"apache": {
					tags: repoTagsList{
						Tags: []string{"1.1", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
				"nginx": {
					tags: repoTagsList{
						Tags: []string{"2.0", "1.0"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.cncf.helm.config.v1+json",
						},
					},
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := makeTestOCIServer(t, registryName, tc.repos, "")
			defer ts.Close()
			// Use the test servers host/port as repo url.
			tc.validator.AppRepo.Spec.URL = fmt.Sprintf("%s/%s", ts.URL, registryName)

			response, err := tc.validator.Validate(httpclient.New())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
