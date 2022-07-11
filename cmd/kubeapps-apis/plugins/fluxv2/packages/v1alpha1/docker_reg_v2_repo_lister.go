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
	orasregistryv2 "oras.land/oras-go/v2/registry"
	orasregistryremotev2 "oras.land/oras-go/v2/registry/remote"
	orasregistryauthv2 "oras.land/oras-go/v2/registry/remote/auth"
)

// This flavor of OCI lister Works with respect to those OCI registry vendors that implement
// Docker Registry API V2 or OCI Distribution Specification. For example, GitHub (ghcr.io)
// References:
// - https://docs.docker.com/registry/spec/api/#base
// - https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api
func NewDockerRegistryApiV2RepositoryLister() OCIRepositoryLister {
	return &dockerRegistryApiV2RepositoryLister{}
}

type dockerRegistryApiV2RepositoryLister struct {
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#api-version-check
// also https://github.com/oras-project/oras-go/blob/14422086e418/registry/remote/registry.go
func (l *dockerRegistryApiV2RepositoryLister) IsApplicableFor(ociRegistry *OCIRegistry) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRegistry.url.String())

	orasRegistry, err := newRemoteOrasRegistry(ociRegistry)
	if err != nil {
		return false, err
	} else {
		ping := "OK"
		err = orasRegistry.Ping(context.Background())
		if err != nil {
			ping = fmt.Sprintf("%v", err)
		}
		log.Infof("ORAS v2 Registry [%s PlainHTTP=%t] PING: %s",
			ociRegistry.url.String(), orasRegistry.PlainHTTP, ping)
		return err == nil, err
	}
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
func (l *dockerRegistryApiV2RepositoryLister) ListRepositoryNames(ociRegistry *OCIRegistry) ([]string, error) {
	log.Infof("+ListRepositoryNames()")

	orasRegistry, err := newRemoteOrasRegistry(ociRegistry)
	if err != nil {
		return nil, err
	} else {
		// this is the way to stop the loop in
		// https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/registry.go#L112
		done := errors.New("(done) backstop")

		fn := func(repos []string) error {
			log.Infof("orasRegistry.Repositories fn: %s", repos)
			return done
		}

		// see https://github.com/vmware-tanzu/kubeapps/pull/4932#issuecomment-1164004999
		// and https://github.com/oras-project/oras-go/issues/196
		// TODO (gfichtenholt) need to append
		// "?last=" + orasRegistry.Reference.Repository
		// to req.Query so we don't start at the beggining of the alphabet

		// impl refs:
		// 1. https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/registry.go#L105
		// 2. https://github.com/oras-project/oras-go/blob/14422086e41897a44cb706726e687d39dc728805/registry/remote/url.go#L43
		err = orasRegistry.Repositories(context.Background(), fn)
		log.Infof("ORAS Repositories returned: %v", err)
		if err != nil && err != done {
			return nil, err
		}
		//repositoryList := []string{}
		//return repositoryList, nil
	}

	// OLD
	return []string{"stefanprodan/charts/podinfo"}, nil
}

func newRemoteOrasRegistry(ociRegistry *OCIRegistry) (*orasregistryremotev2.Registry, error) {
	ref := strings.TrimPrefix(ociRegistry.url.String(), fmt.Sprintf("%s://", registry.OCIScheme))
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
		Credential: ociRegistry.registryCredentialFn,
	}
	return orasRegistry, nil
}
