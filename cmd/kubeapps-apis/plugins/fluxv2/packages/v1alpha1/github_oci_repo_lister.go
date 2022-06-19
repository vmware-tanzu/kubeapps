// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/fluxv2/packages/v1alpha1/common"
	httpclient "github.com/vmware-tanzu/kubeapps/pkg/http-client"
	log "k8s.io/klog/v2"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/registry"
	dockerauth "oras.land/oras-go/pkg/auth/docker"
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
	log.Infof("+IsApplicableFor()")

	ref := strings.TrimPrefix(ociRegistry.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	log.Infof("ref: [%s]", ref)

	// given ref like this
	//   ghcr.io/stefanprodan/charts/podinfo
	// will return
	//  {
	//    "Registry": "ghcr.io",
	//    "Repository": "stefanprodan/charts/podinfo",
	//    "Reference": ""
	// }
	parsedRef, err := orasregistry.ParseReference(ref)
	if err != nil {
		return false, err
	}
	log.Infof("parsed reference: [%s]", common.PrettyPrint(parsedRef))

	credentialsFile := helmpath.ConfigPath(registry.CredentialsFileBasename)

	authClient, err := dockerauth.NewClientWithDockerFallback(credentialsFile)
	if err != nil {
		return false, err
	}

	registryAuthorizer := &registryauth.Client{
		Header: http.Header{"User-Agent": {"Helm/3.9.0"}},
		Cache:  registryauth.DefaultCache,
		Credential: func(ctx context.Context, reg string) (registryauth.Credential, error) {
			dockerClient, ok := authClient.(*dockerauth.Client)
			if !ok {
				return registryauth.EmptyCredential, errors.New("unable to obtain docker client")
			}

			username, password, err := dockerClient.Credential(reg)
			if err != nil {
				return registryauth.EmptyCredential, errors.New("unable to retrieve credentials")
			}

			log.Infof("=======> IsApplicableFor: registryAuthorizer: [%s] [%s...]", username, password[0:3])

			// A blank returned username and password value is a bearer token
			if username == "" && password != "" {
				log.Infof("IsApplicableFor: registryAuthorizer: [%s] [%s]", username, password)
				return registryauth.Credential{
					RefreshToken: password,
				}, nil
			}
			return registryauth.Credential{
				Username: username,
				Password: password,
			}, nil
		},
	}

	ociRepo := registryremote.Repository{
		Reference: parsedRef,
		Client:    registryAuthorizer,
	}

	// ref https://github.com/oras-project/oras-go/blob/main/registry/remote/url.go
	// buildScheme returns HTTP scheme used to access the remote registry.
	buildScheme := func(plainHTTP bool) string {
		if plainHTTP {
			return "http"
		}
		return "https"
	}

	ctx := ctxFn(io.Discard, true)
	// buildRepositoryBaseURL builds the base endpoint of the remote registry.
	// Format: <scheme>://<registry>/v2/
	url := fmt.Sprintf("%s://%s/v2/", buildScheme(ociRepo.PlainHTTP), ociRepo.Reference.Host())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := httpclient.New().Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode == http.StatusOK {
		// based on the presence of this here Docker-Distribution-Api-Version:[registry/2.0]
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
			return true, nil
		}
	} else {
		log.Infof("isApplicableFor: HTTP GET (%s) returned status [%d]", url, resp.StatusCode)
	}

	return false, nil
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
func (l *gitHubRepositoryLister) ListRepositoryNames() ([]string, error) {
	log.Infof("+ListRepositoryNames()")
	// TODO (gfichtenholt) fix me
	return []string{"podinfo"}, nil
}

func ctxFn(out io.Writer, debug bool) context.Context {
	if !debug {
		return orascontext.Background()
	}
	ctx := orascontext.WithLoggerFromWriter(context.Background(), out)
	orascontext.GetLogger(ctx).Logger.SetLevel(logrus.DebugLevel)
	return ctx
}
