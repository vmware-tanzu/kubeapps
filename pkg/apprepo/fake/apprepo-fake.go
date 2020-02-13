/*
Copyright (c) 2019 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	"fmt"
	"io"

	v1alpha1 "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Handler represents a fake Handler for testing purposes
type Handler struct {
	AppRepos    []*v1alpha1.AppRepository
	CreatedRepo *v1alpha1.AppRepository
	Namespaces  []corev1.Namespace
	Secrets     []*corev1.Secret
	Err         error
}

// CreateAppRepository fake
func (c *Handler) CreateAppRepository(appRepoBody io.ReadCloser, requestNamespace, token string) (*v1alpha1.AppRepository, error) {
	c.AppRepos = append(c.AppRepos, c.CreatedRepo)
	return c.CreatedRepo, c.Err
}

// DeleteAppRepository fake
func (c *Handler) DeleteAppRepository(name, namespace, token string) error {
	return c.Err
}

// GetAppRepository fake
func (c *Handler) GetAppRepository(name, namespace, token string) (*v1alpha1.AppRepository, error) {
	for _, r := range c.AppRepos {
		if r.Name == name && r.Namespace == namespace {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

// GetNamespaces fake
func (c *Handler) GetNamespaces(token string) ([]corev1.Namespace, error) {
	return c.Namespaces, c.Err
}

// GetSecret fake
func (c *Handler) GetSecret(name, namespace, token string) (*corev1.Secret, error) {
	for _, r := range c.Secrets {
		if r.Name == name && r.Namespace == namespace {
			return r, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
