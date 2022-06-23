// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	log "k8s.io/klog/v2"

	"helm.sh/helm/v3/pkg/registry"
	orascontext "oras.land/oras-go/pkg/context"
	orasregistry "oras.land/oras-go/pkg/registry"
	registryremote "oras.land/oras-go/pkg/registry/remote"
	registryauth "oras.land/oras-go/pkg/registry/remote/auth"
)

func NewGitHubRepositoryLister() OCIRepositoryLister {
	return &gitHubRepositoryLister{}
}

type gitHubRepositoryLister struct {
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#api-version-check
func (l *gitHubRepositoryLister) IsApplicableFor(ociRegistry *OCIRegistry) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRegistry.url.String())

	// ref https://github.com/helm/helm/blob/657850e44b880cca43d0606ebf5a54eb75362c3f/pkg/registry/client.go#L55
	registryAuthorizer := &registryauth.Client{
		Header:     http.Header{"User-Agent": {common.UserAgentString()}},
		Cache:      registryauth.DefaultCache,
		Credential: ociRegistry.registryCredentialFn,
	}

	// given ref like this
	//   ghcr.io/stefanprodan/charts/podinfo
	// will return
	//  {
	//    "Registry": "ghcr.io",
	//    "Repository": "stefanprodan/charts/podinfo",
	//    "Reference": ""
	// }
	ref := strings.TrimPrefix(ociRegistry.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	log.Infof("ref: [%s]", ref)

	parsedRef, err := orasregistry.ParseReference(ref)
	if err != nil {
		return false, err
	}
	log.Infof("parsed reference: [%s]", common.PrettyPrint(parsedRef))

	ociRepo := registryremote.Repository{
		Reference: parsedRef,
		Client:    registryAuthorizer,
	}

	// build the base endpoint of the remote registry.
	// Format: <scheme>://<registry>/v2/
	url := fmt.Sprintf("%s://%s/v2/", "https", ociRepo.Reference.Host())
	req, err := http.NewRequestWithContext(orascontext.Background(), http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := ociRepo.Client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusOK {
		// based on the presence of this header Docker-Distribution-Api-Version:[registry/2.0]
		// conclude this is a case for GitHubRepositoryLister, e.g.
		// +HTTP GET request:
		// URL: https://ghcr.io/v2/
		// -HTTP GET response: code: 200 OK
		// headers:
		//   map[
		//       Content-Length:[0] Content-Type:[application/json]
		//       Date:[Sun, 19 Jun 2022 05:08:18 GMT]
		//       Docker-Distribution-Api-Version:[registry/2.0]
		//       X-Github-Request-Id:[C4E4:2F9A:3069FD:914D65:62AEAF42]
		//   ]

		val, ok := resp.Header["Docker-Distribution-Api-Version"]
		if ok && len(val) == 1 && val[0] == "registry/2.0" {
			log.Info("-isApplicableFor(): yes")
			return true, nil
		}
	} else {
		log.Infof("isApplicableFor(): HTTP GET (%s) returned status [%d]", url, resp.StatusCode)
	}
	return false, nil
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
func (l *gitHubRepositoryLister) ListRepositoryNames() ([]string, error) {
	log.Info("+ListRepositoryNames()")
	// TODO (gfichtenholt) fix me
	return []string{"podinfo"}, nil
}
