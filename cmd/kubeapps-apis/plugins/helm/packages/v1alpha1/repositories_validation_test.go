// Copyright 2022-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	ocicatalog "github.com/vmware-tanzu/kubeapps/cmd/oci-catalog/gen/catalog/v1alpha1"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"github.com/vmware-tanzu/kubeapps/pkg/ocicatalogtest"
	corev1 "k8s.io/api/core/v1"
	log "k8s.io/klog/v2"
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
	validRequest, err := http.NewRequest("GET", "https://example.com/index.yaml", nil)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	testCases := []struct {
		name             string
		fakeHttpError    error
		fakeRepoResponse *http.Response
		expectedResponse *ValidationResponse
	}{
		{
			name:             "it returns 200 OK validation response if there is no error and the external response is 200",
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("OK")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it does not include the body of the upstream response when validation succeeds",
			fakeRepoResponse: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewReader([]byte("10 Mb of data")))},
			expectedResponse: &ValidationResponse{Code: 200, Message: "OK"},
		},
		{
			name:             "it returns an error from the response with the body text if validation fails",
			fakeRepoResponse: &http.Response{StatusCode: 401, Body: io.NopCloser(bytes.NewReader([]byte("It failed because of X and Y")))},
			expectedResponse: &ValidationResponse{Code: 401, Message: "It failed because of X and Y"},
		},
		{
			name:             "it returns a 400 error if the validation cannot be run",
			fakeHttpError:    fmt.Errorf("client.Do returns an error"),
			expectedResponse: &ValidationResponse{Code: 400, Message: "client.Do returns an error"},
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

			httpValidator := HelmNonOCIValidator{
				ClientGetter: func(*v1alpha1.AppRepository, *corev1.Secret) (httpclient.Client, error) {
					return fakeClient, nil
				},
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type: "oci",
						URL:  "https://example.com",
					},
				},
			}

			response, err := httpValidator.Validate(context.TODO())
			if err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if got, want := fakeClient.request, validRequest; !cmp.Equal(want, got, cmpOpts...) {
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
		{
			name: "it returns an EmptyOCIRegistry response if no repos listed and no catalog",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{},
					},
				},
			},
			repos: map[string]fakeOCIRepo{},
			expectedResponse: &ValidationResponse{
				Code:    400,
				Message: "unable to determine the OCI catalog, you need to specify at least one repository",
			},
		},
		{
			name: "it returns a valid response if no repos listed but VAC catalog index is available",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{},
					},
				},
				OCICatalogAddr: "localhost:9876",
			},
			repos: map[string]fakeOCIRepo{
				"charts-index": {
					tags: repoTagsList{
						Tags: []string{"latest"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.vmware.charts.index.config.v1+json",
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
			name: "it returns a valid response if no repos listed but VAC catalog index is available, even if oci catalog address is set",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type:            "oci",
						OCIRepositories: []string{},
					},
				},
			},
			repos: map[string]fakeOCIRepo{
				"charts-index": {
					tags: repoTagsList{
						Tags: []string{"latest"},
					},
					manifest: repoManifest{
						Config: repoConfig{
							MediaType: "application/vnd.vmware.charts.index.config.v1+json",
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

			// Use the actual client getter since we're using a test double.
			tc.validator.ClientGetter = newRepositoryClient

			response, err := tc.validator.Validate(context.TODO())
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func TestOCIValidateWithCatalogServer(t *testing.T) {
	ociCatalogAddr, ociCatalogDouble, cleanup := ocicatalogtest.SetupTestDouble(t)
	defer cleanup()

	testCases := []struct {
		name             string
		repos            []ocicatalog.Repository
		validator        HelmOCIValidator
		expectedResponse *ValidationResponse
		expectError      bool
	}{
		{
			name: "it returns valid if the oci catalog service finds repositories",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type: "oci",
					},
				},
				OCICatalogAddr: ociCatalogAddr,
			},
			repos: []ocicatalog.Repository{
				{
					Name: "apache",
				},
				{
					Name: "kubeapps",
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name: "it returns valid if the oci catalog service finds just a single repository",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type: "oci",
					},
				},
				OCICatalogAddr: ociCatalogAddr,
			},
			repos: []ocicatalog.Repository{
				{
					Name: "apache",
				},
			},
			expectedResponse: &ValidationResponse{
				Code:    200,
				Message: "OK",
			},
		},
		{
			name: "it returns an error if the oci catalog service is unavailable or does not find any repos",
			validator: HelmOCIValidator{
				AppRepo: &v1alpha1.AppRepository{
					Spec: v1alpha1.AppRepositorySpec{
						Type: "oci",
					},
				},
				OCICatalogAddr: ociCatalogAddr,
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ociCatalogDouble.Repositories = tc.repos

			response, err := tc.validator.Validate(context.TODO())
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected err, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if got, want := response, tc.expectedResponse; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
