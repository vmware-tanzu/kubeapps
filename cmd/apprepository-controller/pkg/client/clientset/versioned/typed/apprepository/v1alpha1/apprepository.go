/*
Copyright 2018 Bitnami.

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

package v1alpha1

import (
	v1alpha1 "github.com/kubeapps/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	scheme "github.com/kubeapps/apprepository-controller/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AppRepositoriesGetter has a method to return a AppRepositoryInterface.
// A group's client should implement this interface.
type AppRepositoriesGetter interface {
	AppRepositories(namespace string) AppRepositoryInterface
}

// AppRepositoryInterface has methods to work with AppRepository resources.
type AppRepositoryInterface interface {
	Create(*v1alpha1.AppRepository) (*v1alpha1.AppRepository, error)
	Update(*v1alpha1.AppRepository) (*v1alpha1.AppRepository, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.AppRepository, error)
	List(opts v1.ListOptions) (*v1alpha1.AppRepositoryList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AppRepository, err error)
	AppRepositoryExpansion
}

// appRepositories implements AppRepositoryInterface
type appRepositories struct {
	client rest.Interface
	ns     string
}

// newAppRepositories returns a AppRepositories
func newAppRepositories(c *KubeappsV1alpha1Client, namespace string) *appRepositories {
	return &appRepositories{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the appRepository, and returns the corresponding appRepository object, and an error if there is any.
func (c *appRepositories) Get(name string, options v1.GetOptions) (result *v1alpha1.AppRepository, err error) {
	result = &v1alpha1.AppRepository{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("apprepositories").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AppRepositories that match those selectors.
func (c *appRepositories) List(opts v1.ListOptions) (result *v1alpha1.AppRepositoryList, err error) {
	result = &v1alpha1.AppRepositoryList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("apprepositories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested appRepositories.
func (c *appRepositories) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("apprepositories").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a appRepository and creates it.  Returns the server's representation of the appRepository, and an error, if there is any.
func (c *appRepositories) Create(appRepository *v1alpha1.AppRepository) (result *v1alpha1.AppRepository, err error) {
	result = &v1alpha1.AppRepository{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("apprepositories").
		Body(appRepository).
		Do().
		Into(result)
	return
}

// Update takes the representation of a appRepository and updates it. Returns the server's representation of the appRepository, and an error, if there is any.
func (c *appRepositories) Update(appRepository *v1alpha1.AppRepository) (result *v1alpha1.AppRepository, err error) {
	result = &v1alpha1.AppRepository{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("apprepositories").
		Name(appRepository.Name).
		Body(appRepository).
		Do().
		Into(result)
	return
}

// Delete takes name of the appRepository and deletes it. Returns an error if one occurs.
func (c *appRepositories) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("apprepositories").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *appRepositories) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("apprepositories").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched appRepository.
func (c *appRepositories) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.AppRepository, err error) {
	result = &v1alpha1.AppRepository{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("apprepositories").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
