// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"

	log "k8s.io/klog/v2"
	"oras.land/oras-go/v2/errdef"
	orasregistryv2 "oras.land/oras-go/v2/registry"
	orasregistryremotev2 "oras.land/oras-go/v2/registry/remote"
)

// This flavor of OCI repsitory ister works with respect to to VMware Harbor
// container registry. The Swagger for the API can be found on an running server
// in the API explorer "About" link, e.g.
// https://demo.goharbor.io/devcenter-api-2.0
// Why is this needed? Or why doesn't dockerRegistryApiV2RepositoryLister just
// take care of this? The answer is harbor robot accounts are not able to list
// repositories using the generic API. But it works using harbor REST API
// ref https://github.com/vmware-tanzu/kubeapps/issues/5219

func NewHarborRegistryApiV2RepositoryLister() OCIChartRepositoryLister {
	return &harborRegistryApiV2RepositoryLister{}
}

type harborRegistryApiV2RepositoryLister struct {
}

func (l *harborRegistryApiV2RepositoryLister) IsApplicableFor(ociRepo *OCIChartRepository) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRepo.url.String())
	orasRegistry, err := newRemoteOrasRegistry(ociRepo)
	if err != nil {
		return false, err
	} else {
		pong, err := l.ping(orasRegistry)
		if err != nil {
			pong = fmt.Sprintf("%v", err)
		}
		log.Infof("Harbor v2 Registry [%s PlainHTTP=%t] Ping: %s",
			ociRepo.url.String(), orasRegistry.PlainHTTP, pong)
		if err != nil {
			return false, err
		}

		projectName, err := l.projectName(ociRepo)
		if err != nil {
			return false, err
		}

		err = l.projectExists(projectName, orasRegistry)
		return err == nil, err
	}
}

func (l *harborRegistryApiV2RepositoryLister) ListRepositoryNames(ociRepo *OCIChartRepository) ([]string, error) {
	log.Infof("+ListRepositoryNames(%s)", ociRepo.url.String())
	orasRegistry, err := newRemoteOrasRegistry(ociRepo)
	if err != nil {
		return nil, err
	} else {
		projectName, err := l.projectName(ociRepo)
		if err != nil {
			return nil, err
		}
		repos := []string{}
		for page, pageSize, more := 1, 10, true; more; page++ {
			onePage, err := l.listProjectRepositories(projectName, orasRegistry, page, pageSize)
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
}

func (l *harborRegistryApiV2RepositoryLister) projectName(ociRepo *OCIChartRepository) (string, error) {
	path := strings.TrimPrefix(ociRepo.url.Path, "/")
	segments := strings.SplitN(path, "/", 1)
	if len(segments) > 0 {
		return segments[0], nil
	} else {
		return "", fmt.Errorf("unexpected URL format: [%s]", ociRepo.url.String())
	}
}

// ref https://demo.goharbor.io/#/ping/getPing
func (l *harborRegistryApiV2RepositoryLister) ping(r *orasregistryremotev2.Registry) (string, error) {
	url := l.buildRegistryBaseURL(r.PlainHTTP, r.Reference) + "/ping"
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		if pong, err := ioutil.ReadAll(resp.Body); err != nil {
			return "", err
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

// ref https://demo.goharbor.io/#/project/getProject
// There is a faster way
//  ref https://demo.goharbor.io/#/project/headProject
// but it seems to result in 401 Unauthorized. Bug in ORAS?
func (l *harborRegistryApiV2RepositoryLister) projectExists(projectName string, r *orasregistryremotev2.Registry) error {
	url := fmt.Sprintf("%s/projects/%s",
		l.buildRegistryBaseURL(r.PlainHTTP, r.Reference), projectName)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := r.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil

	case http.StatusNotFound:
		return errdef.ErrNotFound

	case http.StatusForbidden:
		return errors.New("forbidden")

	default:
		err = parseHarborErrorResponse(resp)
		return err
	}
}

// https://demo.goharbor.io/#/repository/listRepositories
func (l *harborRegistryApiV2RepositoryLister) listProjectRepositories(projectName string, r *orasregistryremotev2.Registry, page, pageSize int) ([]string, error) {
	log.Infof("+listProjectRepositories(%s)", projectName)
	url := fmt.Sprintf("%s/projects/%s/repositories?page=%d&page_size=%d",
		l.buildRegistryBaseURL(r.PlainHTTP, r.Reference), projectName, page, pageSize)
	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := r.Client.Do(req)
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

// buildScheme returns HTTP scheme used to access the remote registry.
func (l *harborRegistryApiV2RepositoryLister) buildScheme(plainHTTP bool) string {
	if plainHTTP {
		return "http"
	}
	return "https"
}

// buildRegistryBaseURL builds the URL for accessing the base API.
func (l *harborRegistryApiV2RepositoryLister) buildRegistryBaseURL(plainHTTP bool, ref orasregistryv2.Reference) string {
	return fmt.Sprintf("%s://%s/api/v2.0", l.buildScheme(plainHTTP), ref.Host())
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
