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

// This flavor of OCI lister Works with respect to those OCI registry vendors that implement
// Docker Registry API V2 or OCI Distribution Specification. For example, GitHub (ghcr.io)
// References:
// - https://docs.docker.com/registry/spec/api/#base
// - https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api
func NewDockerRegistryApiV2RepositoryLister() OCIChartRepositoryLister {
	return &dockerRegistryApiV2RepositoryLister{}
}

type dockerRegistryApiV2RepositoryLister struct {
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#api-version-check
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
		return err == nil, err
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
				if lastRepoMatch = strings.HasPrefix(r, startAt+"/"); lastRepoMatch {
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
		Header:     orasregistryauthv2.DefaultClient.Header.Clone(),
		Cache:      orasregistryauthv2.DefaultCache,
		Credential: ociRepo.registryCredentialFn,
	}
	return orasRegistry, nil
}
