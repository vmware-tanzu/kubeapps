// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
)

// This flavor of OCI lister Works with respect to those OCI registry vendors that implement
// Docker Registry API V2 or OCI Distribution Specification. For example, GitHub (ghcr.io)
// References:
// - https://docs.docker.com/registry/spec/api/#base
// - https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api
func newFakeDockerRegistryApiV2RepositoryLister(t *testing.T, r []fakeRepo) OCIRepositoryLister {
	return &fakeDockerRegistryApiV2RepositoryLister{
		t:            t,
		repositories: r,
	}
}

type fakeDockerRegistryApiV2RepositoryLister struct {
	t            *testing.T
	repositories []fakeRepo
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#api-version-check
// also https://github.com/oras-project/oras-go/blob/14422086e418/registry/remote/registry.go
func (fake *fakeDockerRegistryApiV2RepositoryLister) IsApplicableFor(ociRegistry *OCIRegistry) (bool, error) {
	fake.t.Logf("+IsApplicableFor(%s)", ociRegistry.url.String())
	return true, nil
}

// given an OCIRegistry instance, returns a list of repository names, e.g.
// given an OCIRegistry instance with url "oci://ghcr.io/stefanprodan/charts"
//    may return ["stefanprodan/charts/podinfo", "stefanprodan/charts/podinfo-2"]
// ref: https://github.com/distribution/distribution/blob/main/docs/spec/api.md#listing-repositories
func (fake *fakeDockerRegistryApiV2RepositoryLister) ListRepositoryNames(ociRegistry *OCIRegistry) ([]string, error) {
	fake.t.Logf("+ListRepositoryNames(%s)", ociRegistry.url.String())
	names := []string{}
	prefix := strings.TrimPrefix(ociRegistry.url.Path, "/")
	for _, r := range fake.repositories {
		names = append(names, prefix+"/"+r.name)
	}
	fake.t.Logf("-ListRepositoryNames(%s): returned %s", ociRegistry.url.String(), names)
	return names, nil
}

type fakeRegistryClientType struct {
	t            *testing.T
	repositories []fakeRepo
}

func (fake *fakeRegistryClientType) Login(host string, opts ...registry.LoginOption) error {
	fake.t.Logf("+Login")
	return nil
}

func (fake *fakeRegistryClientType) Logout(host string, opts ...registry.LogoutOption) error {
	fake.t.Logf("+Logout")
	return nil
}

func (fake *fakeRegistryClientType) Tags(ref string) ([]string, error) {
	fake.t.Logf("+Tags(%s)", ref)

	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 1 {
		return nil, fmt.Errorf("ref [%s] missing repository", ref)
	}
	parts = strings.Split(parts[1], "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("ref [%s] is not in expected format", ref)
	}
	refRepository := parts[2]
	refChart := refRepository // just for now
	for _, r := range fake.repositories {
		if refRepository == r.name {
			for _, c := range r.charts {
				if refChart == c.name {
					tags := []string{}
					for _, cv := range c.versions {
						tags = append(tags, cv.version)
					}
					return tags, nil
				}
			}
			return nil, fmt.Errorf("no charts found for ref [%s] in the registry", ref)
		}
	}
	return nil, fmt.Errorf("no repositories found for ref [%s] in the registry", ref)
}

func (fake *fakeRegistryClientType) DownloadChart(chartVersion *repo.ChartVersion) (*bytes.Buffer, error) {
	fake.t.Logf("+DownloadChart(%s)", chartVersion.Version)
	// TODO: return bytes from .tgz file
	return nil, fmt.Errorf("TODO implement fakeRegistryClient.DownloadChart")
}

func newFakeRegistryClient(t *testing.T, r []fakeRepo) (RegistryClient, string, error) {
	return &fakeRegistryClientType{
		t:            t,
		repositories: r,
	}, "", nil
}

type fakeChartVersion struct {
	version  string
	tgzBytes []byte
}

type fakeChart struct {
	name     string
	versions []fakeChartVersion
}

type fakeRepo struct {
	name   string
	charts []fakeChart
}

type fakeClientData struct {
	repositories []fakeRepo
}

func initOciFakeClientBuilder(t *testing.T, data fakeClientData) {
	t.Logf("+initOciFakeClientFactoryAndLister()")

	builtInRepoListers = []OCIRepositoryLister{
		newFakeDockerRegistryApiV2RepositoryLister(t, data.repositories),
	}

	registryClientBuilderFn = func(isLogin bool, tlsConfig *tls.Config, getterOpts []getter.Option, helmGetter getter.Getter) (RegistryClient, string, error) {
		return newFakeRegistryClient(t, data.repositories)
	}
}
