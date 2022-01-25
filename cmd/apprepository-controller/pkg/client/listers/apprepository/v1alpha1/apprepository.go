// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// AppRepositoryLister helps list AppRepositories.
type AppRepositoryLister interface {
	// List lists all AppRepositories in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.AppRepository, err error)
	// AppRepositories returns an object that can list and get AppRepositories.
	AppRepositories(namespace string) AppRepositoryNamespaceLister
	AppRepositoryListerExpansion
}

// appRepositoryLister implements the AppRepositoryLister interface.
type appRepositoryLister struct {
	indexer cache.Indexer
}

// NewAppRepositoryLister returns a new AppRepositoryLister.
func NewAppRepositoryLister(indexer cache.Indexer) AppRepositoryLister {
	return &appRepositoryLister{indexer: indexer}
}

// List lists all AppRepositories in the indexer.
func (s *appRepositoryLister) List(selector labels.Selector) (ret []*v1alpha1.AppRepository, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AppRepository))
	})
	return ret, err
}

// AppRepositories returns an object that can list and get AppRepositories.
func (s *appRepositoryLister) AppRepositories(namespace string) AppRepositoryNamespaceLister {
	return appRepositoryNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// AppRepositoryNamespaceLister helps list and get AppRepositories.
type AppRepositoryNamespaceLister interface {
	// List lists all AppRepositories in the indexer for a given namespace.
	List(selector labels.Selector) (ret []*v1alpha1.AppRepository, err error)
	// Get retrieves the AppRepository from the indexer for a given namespace and name.
	Get(name string) (*v1alpha1.AppRepository, error)
	AppRepositoryNamespaceListerExpansion
}

// appRepositoryNamespaceLister implements the AppRepositoryNamespaceLister
// interface.
type appRepositoryNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all AppRepositories in the indexer for a given namespace.
func (s appRepositoryNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.AppRepository, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.AppRepository))
	})
	return ret, err
}

// Get retrieves the AppRepository from the indexer for a given namespace and name.
func (s appRepositoryNamespaceLister) Get(name string) (*v1alpha1.AppRepository, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("apprepository"), name)
	}
	return obj.(*v1alpha1.AppRepository), nil
}
