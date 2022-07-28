// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/registry"
	"helm.sh/helm/v3/pkg/repo"
	log "k8s.io/klog/v2"
)

// This flavor of OCI lister Works with respect to those OCI registry vendors that implement
// Docker Registry API V2 or OCI Distribution Specification. For example, GitHub (ghcr.io)
// References:
// - https://docs.docker.com/registry/spec/api/#base
// - https://github.com/opencontainers/distribution-spec/blob/main/spec.md#api
func newFakeDockerRegistryApiV2RepositoryLister() OCIRepositoryLister {
	return &fakeDockerRegistryApiV2RepositoryLister{}
}

type fakeDockerRegistryApiV2RepositoryLister struct {
}

// ref https://github.com/distribution/distribution/blob/main/docs/spec/api.md#api-version-check
// also https://github.com/oras-project/oras-go/blob/14422086e418/registry/remote/registry.go
func (l *fakeDockerRegistryApiV2RepositoryLister) IsApplicableFor(ociRegistry *OCIRegistry) (bool, error) {
	log.Infof("+IsApplicableFor(%s)", ociRegistry.url.String())
	return true, nil
}

func (l *fakeDockerRegistryApiV2RepositoryLister) ListRepositoryNames(ociRegistry *OCIRegistry) ([]string, error) {
	log.Infof("+ListRepositoryNames(%s)", ociRegistry.url.String())
	return []string{
		strings.TrimPrefix(ociRegistry.url.Path, "/") + "/podinfo",
	}, nil
}

type fakeRegistryClientType struct {
}

func (r *fakeRegistryClientType) Login(host string, opts ...registry.LoginOption) error {
	log.Infof("+Login")
	return nil
}

func (r *fakeRegistryClientType) Logout(host string, opts ...registry.LogoutOption) error {
	log.Infof("+Logout")
	return nil
}

func (r *fakeRegistryClientType) Tags(url string) ([]string, error) {
	log.Infof("+Tags(%s)", url)
	return []string{"6.1.5"}, nil
}

func (r *fakeRegistryClientType) DownloadChart(chartVersion *repo.ChartVersion) (*bytes.Buffer, error) {
	log.Infof("+DownloadChart(%s)", chartVersion.Version)
	// TODO: return bytes from .tgz file
	return nil, fmt.Errorf("TODO implement fakeRegistryClient.DownloadChart")
}

func newFakeRegistryClient(isLogin bool, tlsConfig *tls.Config, getterOpts []getter.Option, helmGetter getter.Getter) (RegistryClient, string, error) {
	return &fakeRegistryClientType{}, "", nil
}

func initOciFakeClientBuilder() {
	log.Infof("+initOciFakeClientFactoryAndLister()")

	builtInRepoListers = []OCIRepositoryLister{
		newFakeDockerRegistryApiV2RepositoryLister(),
	}

	registryClientBuilderFn = func(isLogin bool, tlsConfig *tls.Config, getterOpts []getter.Option, helmGetter getter.Getter) (RegistryClient, string, error) {
		return newFakeRegistryClient(isLogin, tlsConfig, getterOpts, helmGetter)
	}
}
