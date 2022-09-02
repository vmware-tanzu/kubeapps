// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	log "k8s.io/klog/v2"

	"helm.sh/helm/v3/pkg/registry"

	// ORAS => OCI Registry AS Storage
	// project home page: https://oras.land/
	// releases: https://github.com/oras-project/oras-go/releases

	orasregistryv2 "oras.land/oras-go/v2/registry"
	orasregistryremotev2 "oras.land/oras-go/v2/registry/remote"
	orasregistryauthv2 "oras.land/oras-go/v2/registry/remote/auth"
)

// This flavor of OCI repsitory lister works with respect to those OCI registry vendors
// that implement Docker Registry HTTP API V2 or OCI Distribution Specification.
// For example, GitHub (ghcr.io), harbor and GCP Artifact Repositories are
// known to suport this API, with some caveats:
// - harbor does support Docker Registry HTTP API V2 when robot accounts are used
// References:
// - https://github.com/distribution/distribution/blob/main/docs/spec/api.md

func NewDockerRegistryApiV2RepositoryLister() OCIChartRepositoryLister {
	return &dockerRegistryApiV2RepositoryLister{}
}

type dockerRegistryApiV2RepositoryLister struct {
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
// also https://github.com/oras-project/oras-go/blob/14422086e418/registry/remote/registry.go
func (l *dockerRegistryApiV2RepositoryLister) IsApplicableFor(ociRepo *OCIChartRepository) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRepo.url.String())

	orasRegistry, err := newRemoteOrasRegistry(ociRepo)
	if err != nil {
		return false, err
	} else {
		ping := "OK"
		err = orasRegistry.Ping(context.Background())
		if err != nil {
			ping = fmt.Sprintf("%v", err)
		}
		log.Infof("ORAS v2 Registry [%s PlainHTTP=%t] PING: %s",
			ociRepo.url.String(), orasRegistry.PlainHTTP, ping)
		if err != nil {
			return false, err
		}
		done := errors.New("(done) backstop")
		fn := func(repos []string) error { return done }
		// basic sanity check that we CAN call Repositories in general
		// and avoid things like this later:
		// GET "https://demo.goharbor.io/v2/_catalog?last=stefanprodan-podinfo-clone":
		// unexpected status code 401: unauthorized: unauthorized to list catalog:
		// unauthorized to list catalog
		err = orasRegistry.Repositories(context.Background(), "", fn)
		if err == done {
			// everything looks kosher
			return true, nil
		} else {
			log.Infof("This lister is not applicable due to: %v", err)
			return false, err
		}
	}
}

// given an OCIChartRepository instance, returns a list of repository names, e.g.
// given an OCIChartRepository instance with url "oci://ghcr.io/stefanprodan/charts"
//    may return ["stefanprodan/charts/podinfo", "stefanprodan/charts/podinfo-2"]
// ref: https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
func (l *dockerRegistryApiV2RepositoryLister) ListRepositoryNames(ociRepo *OCIChartRepository) ([]string, error) {
	log.Infof("+ListRepositoryNames(%s)", ociRepo.url.String())

	orasRegistry, err := newRemoteOrasRegistry(ociRepo)
	if err != nil {
		return nil, err
	} else {
		// this is where we will start, e.g. "stefanprodan/charts"
		startAt := strings.Trim(ociRepo.url.Path, "/")

		repositoryList := []string{}

		// this is the way to stop the loop in
		// https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/registry.go#L112
		done := errors.New("(done) backstop")

		fn := func(repos []string) error {
			log.Infof("orasRegistry.Repositories fn: %s", repos)
			lastRepoMatch := false
			for _, r := range repos {
				// Examples:
				// GitHub and Harbor: stefanprodan-podinfo-clone/podinfo
				// GCP Artifact Repository: vmware-kubeapps-ci/stefanprodan-podinfo-clone/podinfo
				lastRepoMatch =
					strings.HasPrefix(r, startAt+"/") ||
						strings.Contains(r, "/"+startAt+"/")
				if lastRepoMatch {
					repositoryList = append(repositoryList, r)
				}
			}
			if !lastRepoMatch {
				return done
			} else {
				return nil
			}
		}

		// impl refs:
		// 1. https://github.com/oras-project/oras-go/blob/4660638096b4b4b5c368ce98cd7040485b5ad776/registry/remote/registry.go#L105
		// 2. https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/url.go#L43
		err = orasRegistry.Repositories(context.Background(), startAt, fn)
		log.Infof("ORAS .Repositories() returned err: %v", err)
		if err != nil && err != done {
			return nil, err
		}
		log.Infof("-ListRepositoryNames(%s): returned %s", ociRepo.url.String(), repositoryList)
		return repositoryList, nil
	}
}

func newRemoteOrasRegistry(ociRepo *OCIChartRepository) (*orasregistryremotev2.Registry, error) {
	ref := strings.TrimPrefix(ociRepo.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
	parsedRef, err := orasregistryv2.ParseReference(ref)
	if err != nil {
		return nil, err
	}
	orasRegistry, err := orasregistryremotev2.NewRegistry(parsedRef.Registry)
	if err != nil {
		return nil, err
	}
	orasRegistry.Client = &orasregistryauthv2.Client{
		Header: orasregistryauthv2.DefaultClient.Header.Clone(),
		// not using the cache for now to avoid things like
		// https://github.com/vmware-tanzu/kubeapps/issues/5219#issuecomment-1233738309
		// also orasRegistry today is a short lived object, so caching tokens is a waste
		// per ORAS code: If nil, no cache is used
		Cache:      nil,
		Credential: ociRepo.registryCredentialFn,
	}
	return orasRegistry, nil
}
