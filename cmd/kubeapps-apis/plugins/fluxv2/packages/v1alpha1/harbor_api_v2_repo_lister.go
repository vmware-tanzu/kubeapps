// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	"helm.sh/helm/v3/pkg/registry"
	log "k8s.io/klog/v2"
	"oras.land/oras-go/v2/errdef"
	orasregistryv2 "oras.land/oras-go/v2/registry"
	orasregistryauthv2 "oras.land/oras-go/v2/registry/remote/auth"
)

// This flavor of OCI repository lister works with respect to an instance of
// CNCF Harbor container registry. The OpenAPI Specification (Swagger) for the
// API can be found on a running server in the API explorer "About" link, e.g.
//   https://demo.goharbor.io/devcenter-api-2.0
// Why is this needed? Or why doesn't dockerRegistryApiV2RepositoryLister just
// take care of this? The answer is harbor robot accounts are not able to list
// repositories using the generic API. But it works using harbor-specific REST API
// ref https://github.com/vmware-tanzu/kubeapps/issues/5219

func NewHarborRegistryApiV2RepositoryLister() OCIChartRepositoryLister {
	return &harborRegistryApiV2RepositoryLister{}
}

type harborRegistryApiV2RepositoryLister struct {
}

func (l *harborRegistryApiV2RepositoryLister) IsApplicableFor(ociRepo *OCIChartRepository) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRepo.url.String())
	ref := strings.TrimPrefix(ociRepo.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	parsedRef, err := orasregistryv2.ParseReference(ref)
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	cred, err := ociRepo.registryCredentialFn(ctx, parsedRef.Host())
	if err != nil {
		return false, err
	}

	pong, err := pingHarbor(ctx, parsedRef, cred)
	if err != nil {
		pong = fmt.Sprintf("%v", err)
	}
	log.Infof("Harbor v2 Registry [%s] Ping: %s", ociRepo.url.String(), pong)
	if err != nil {
		return false, err
	}

	projectName, err := harborProjectNameFromURL(ociRepo.url)
	if err != nil {
		return false, err
	}

	err = harborProjectExists(ctx, projectName, parsedRef, cred)
	return err == nil, err
}

func (l *harborRegistryApiV2RepositoryLister) ListRepositoryNames(ociRepo *OCIChartRepository) ([]string, error) {
	log.Infof("+ListRepositoryNames(%s)", ociRepo.url.String())
	ref := strings.TrimPrefix(ociRepo.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	parsedRef, err := orasregistryv2.ParseReference(ref)
	if err != nil {
		return nil, err
	}

	projectName, err := harborProjectNameFromURL(ociRepo.url)
	if err != nil {
		return nil, err
	}
	repos := []string{}
	ctx := context.Background()
	cred, err := ociRepo.registryCredentialFn(ctx, parsedRef.Host())
	if err != nil {
		return nil, err
	}
	for page, pageSize, more := 1, 10, true; more; page++ {
		onePage, err := listHarborProjectRepositories(
			ctx, projectName, parsedRef, cred, page, pageSize)
		if err != nil {
			return nil, err
		}
		repos = append(repos, onePage...)
		if len(onePage) < pageSize {
			more = false
		}
	}
	return repos, nil
}

// ref https://demo.goharbor.io/#/ping/getPing
func pingHarbor(ctx context.Context, ref orasregistryv2.Reference, cred orasregistryauthv2.Credential) (string, error) {
	log.Infof("+pingHarbor")
	url := fmt.Sprintf("%s/ping", buildHarborRegistryBaseURL(false, ref))
	resp, err := harborDoHttpRequest(ctx, http.MethodGet, cred, url)
	if err != nil {
		return "", err
	}
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/plain") {
		return "", fmt.Errorf("unexpected response content type: [%s]", contentType)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		lr := io.LimitReader(resp.Body, 100)
		if pong, err := io.ReadAll(lr); err != nil {
			return "", err
		} else if string(pong) != "Pong" {
			return "", fmt.Errorf("unexpected response body: %s", string(pong))
		} else {
			return string(pong), nil
		}

	case http.StatusNotFound:
		return "", errdef.ErrNotFound

	default:
		err = parseHarborErrorResponse(resp)
		return err.Error(), err
	}
}

//  https://demo.goharbor.io/#/project/headProject
func harborProjectExists(ctx context.Context, projectName string, ref orasregistryv2.Reference, cred orasregistryauthv2.Credential) error {
	log.Infof("+projectExists(%s)", projectName)
	url := fmt.Sprintf("%s/projects?project_name=%s",
		buildHarborRegistryBaseURL(false, ref), projectName)
	resp, err := harborDoHttpRequest(ctx, http.MethodHead, cred, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		return errdef.ErrNotFound

	default:
		err = parseHarborErrorResponse(resp)
		return err
	}
}

// https://demo.goharbor.io/#/repository/listRepositories
func listHarborProjectRepositories(ctx context.Context, projectName string, ref orasregistryv2.Reference, cred orasregistryauthv2.Credential, page, pageSize int) ([]string, error) {
	log.Infof("+listProjectRepositories(%s)", projectName)
	url := fmt.Sprintf("%s/projects/%s/repositories?page=%d&page_size=%d",
		buildHarborRegistryBaseURL(false, ref), projectName, page, pageSize)
	resp, err := harborDoHttpRequest(ctx, http.MethodGet, cred, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return parseHarborProjectRepositoriesResponse(resp)

	case http.StatusNotFound:
		return nil, errdef.ErrNotFound

	default:
		err = parseHarborErrorResponse(resp)
		return nil, err
	}
}

// buildHarborRegistryBaseURL builds the URL for accessing the base API.
func buildHarborRegistryBaseURL(plainHTTP bool, ref orasregistryv2.Reference) string {
	scheme := "https"
	if plainHTTP {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/api/v2.0", scheme, ref.Host())
}

func harborProjectNameFromURL(url url.URL) (string, error) {
	path := strings.TrimPrefix(url.Path, "/")
	segments := strings.SplitN(path, "/", 2)
	if len(segments) > 0 {
		return segments[0], nil
	} else {
		return "", fmt.Errorf("unexpected URL format: [%s]", url.String())
	}
}

func harborDoHttpRequest(ctx context.Context, method string, cred orasregistryauthv2.Credential, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if cred.Username != "" && cred.Password != "" {
		req.SetBasicAuth(cred.Username, cred.Password)
	}
	return httpclient.New().Do(req)
}

// maxErrorBytes specifies the default limit on how many response bytes are
// allowed in the server's error response.
// A typical error message is around 200 bytes. Hence, 8 KiB should be
// sufficient.
var maxErrorBytes int64 = 8 * 1024 // 8 KiB

// requestError contains a single error.
type requestError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// requestErrors is a bundle of requestError.
type requestErrors []requestError

// Error returns a error string describing the error.
func (e requestError) Error() string {
	code := strings.Map(func(r rune) rune {
		if r == '_' {
			return ' '
		}
		return unicode.ToLower(r)
	}, e.Code)
	if e.Message == "" {
		return code
	}
	return fmt.Sprintf("%s: %s", code, e.Message)
}

// Error returns a error string describing the error.
func (errs requestErrors) Error() string {
	switch len(errs) {
	case 0:
		return "<nil>"
	case 1:
		return errs[0].Error()
	}
	var errmsgs []string
	for _, err := range errs {
		errmsgs = append(errmsgs, err.Error())
	}
	return strings.Join(errmsgs, "; ")
}

type harborRepoModel struct {
	Name string `json:"name"`
}

func parseHarborProjectRepositoriesResponse(resp *http.Response) ([]string, error) {
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		return nil, fmt.Errorf("unexpected response content type: [%s]", contentType)
	}
	var respRepos []harborRepoModel
	if err := json.NewDecoder(resp.Body).Decode(&respRepos); err != nil {
		return nil, err
	} else {
		repos := []string{}
		for _, r := range respRepos {
			repos = append(repos, r.Name)
		}
		return repos, nil
	}
}

// parseHarborErrorResponse parses the error returned by the remote registry.
func parseHarborErrorResponse(resp *http.Response) error {
	var errmsg string
	var body struct {
		Errors requestErrors `json:"errors"`
	}
	lr := io.LimitReader(resp.Body, maxErrorBytes)
	if err := json.NewDecoder(lr).Decode(&body); err == nil && len(body.Errors) > 0 {
		errmsg = body.Errors.Error()
	} else {
		errmsg = http.StatusText(resp.StatusCode)
	}
	return fmt.Errorf("%s %q: unexpected status code %d: %s", resp.Request.Method, resp.Request.URL, resp.StatusCode, errmsg)
}
