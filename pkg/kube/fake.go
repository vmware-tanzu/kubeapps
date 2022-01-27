// Copyright 2019-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package kube

import (
	"io"

	apprepov1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	k8sauthorizationv1 "k8s.io/api/authorization/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// FakeHandler represents a fake Handler for testing purposes
type FakeHandler struct {
	AppRepos    []*apprepov1alpha1.AppRepository
	CreatedRepo *apprepov1alpha1.AppRepository
	UpdatedRepo *apprepov1alpha1.AppRepository
	Namespaces  []k8scorev1.Namespace
	Secrets     []*k8scorev1.Secret
	ValRes      *ValidationResponse
	Options     KubeOptions
	Err         error
	Can         bool
}

// AsUser fakes user auth
func (c *FakeHandler) AsUser(token, cluster string) (handler, error) {
	return c, nil
}

// AsSVC fakes using current svcaccount
func (c *FakeHandler) AsSVC(cluster string) (handler, error) {
	return c, nil
}

// NS fakes returning header namespace options
func (c *FakeHandler) GetOptions() KubeOptions {
	return c.Options
}

// ListAppRepositories fake
func (c *FakeHandler) ListAppRepositories(requestNamespace string) (*apprepov1alpha1.AppRepositoryList, error) {
	appRepos := &apprepov1alpha1.AppRepositoryList{}
	for _, repo := range c.AppRepos {
		appRepos.Items = append(appRepos.Items, *repo)
	}
	return appRepos, c.Err
}

// CreateAppRepository fake
func (c *FakeHandler) CreateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*apprepov1alpha1.AppRepository, error) {
	c.AppRepos = append(c.AppRepos, c.CreatedRepo)
	return c.CreatedRepo, c.Err
}

// RefreshAppRepository fake
func (c *FakeHandler) RefreshAppRepository(repoName string, requestNamespace string) (*apprepov1alpha1.AppRepository, error) {
	return c.UpdatedRepo, c.Err
}

// UpdateAppRepository fake
func (c *FakeHandler) UpdateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*apprepov1alpha1.AppRepository, error) {
	return c.UpdatedRepo, c.Err
}

// DeleteAppRepository fake
func (c *FakeHandler) DeleteAppRepository(name, namespace string) error {
	return c.Err
}

// GetAppRepository fake
func (c *FakeHandler) GetAppRepository(name, namespace string) (*apprepov1alpha1.AppRepository, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	for _, r := range c.AppRepos {
		if r.Name == name && r.Namespace == namespace {
			return r, nil
		}
	}
	return nil, k8serrors.NewNotFound(k8sschema.GroupResource{}, "foo")
}

// GetNamespaces fake
func (c *FakeHandler) GetNamespaces(precheckedNamespaces []k8scorev1.Namespace) ([]k8scorev1.Namespace, error) {
	if len(precheckedNamespaces) > 0 {
		return precheckedNamespaces, c.Err
	}
	return c.Namespaces, c.Err
}

// GetSecret fake
func (c *FakeHandler) GetSecret(name, namespace string) (*k8scorev1.Secret, error) {
	for _, r := range c.Secrets {
		if r.Name == name && r.Namespace == namespace {
			return r, nil
		}
	}
	return nil, k8serrors.NewNotFound(k8sschema.GroupResource{}, "foo")
}

// ValidateAppRepository fake
func (c *FakeHandler) ValidateAppRepository(appRepoBody io.ReadCloser, requestNamespace string) (*ValidationResponse, error) {
	return c.ValRes, c.Err
}

// GetOperatorLogo fake
func (c *FakeHandler) GetOperatorLogo(namespace, name string) ([]byte, error) {
	return []byte{}, nil
}

// CanI fake
func (c *FakeHandler) CanI(resourceAttributes *k8sauthorizationv1.ResourceAttributes) (bool, error) {
	return c.Can, c.Err
}
