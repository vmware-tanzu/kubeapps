/*
Copyright © 2021 VMware
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
package main

import (
	"context"
	"fmt"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	pkgResource             = "Package"
	pkgsResource            = "packages"
	pkgMetadataResource     = "PackageMetadata"
	pkgMetadatasResource    = "packagemetadatas"
	pkgRepositoryResource   = "PackageRepository"
	pkgRepositoriesResource = "packagerepositories"
	pkgInstallResource      = "PackageInstall"
	pkgInstallsResource     = "packageinstalls"
	appResource             = "App"
	appsResource            = "apps"
	appLabelKey             = "kapp.k14s.io/app"
)

// Dynamic ResourceInterface getters to encapsulate the logic of getting the proper group version API resources

// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
func (s *Server) getPkgResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	_, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    datapackagingv1alpha1.SchemeGroupVersion.Group,
		Version:  datapackagingv1alpha1.SchemeGroupVersion.Version,
		Resource: pkgsResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	return ri, nil
}

// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-metadata
func (s *Server) getPkgMetadataResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	_, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    datapackagingv1alpha1.SchemeGroupVersion.Group,
		Version:  datapackagingv1alpha1.SchemeGroupVersion.Version,
		Resource: pkgMetadatasResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	return ri, nil
}

// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-install
func (s *Server) getPkgInstallResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	_, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    packagingv1alpha1.SchemeGroupVersion.Group,
		Version:  packagingv1alpha1.SchemeGroupVersion.Version,
		Resource: pkgInstallsResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	return ri, nil
}

// See https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
func (s *Server) getPkgRepositoryResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	_, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    packagingv1alpha1.SchemeGroupVersion.Group,
		Version:  packagingv1alpha1.SchemeGroupVersion.Version,
		Resource: pkgRepositoriesResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	return ri, nil
}

// See https://carvel.dev/kapp-controller/docs/latest/app-spec/
func (s *Server) getAppResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	_, dynClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    kappctrlv1alpha1.SchemeGroupVersion.Group,
		Version:  kappctrlv1alpha1.SchemeGroupVersion.Version,
		Resource: appsResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	return ri, nil
}

//  Single resource getters

// getPkg returns the package for the given cluster, namespace and identifier
func (s *Server) getPkg(ctx context.Context, cluster, namespace, identifier string) (*datapackagingv1alpha1.Package, error) {
	var pkg datapackagingv1alpha1.Package
	resource, err := s.getPkgResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.Get(ctx, identifier, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

// getPkgMetadata returns the package metadata for the given cluster, namespace and identifier
func (s *Server) getPkgMetadata(ctx context.Context, cluster, namespace, identifier string) (*datapackagingv1alpha1.PackageMetadata, error) {
	var pkgMetadata datapackagingv1alpha1.PackageMetadata
	resource, err := s.getPkgMetadataResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.Get(ctx, identifier, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &pkgMetadata)
	if err != nil {
		return nil, err
	}
	return &pkgMetadata, nil
}

// getPkgInstall returns the package install for the given cluster, namespace and identifier
func (s *Server) getPkgInstall(ctx context.Context, cluster, namespace, identifier string) (*packagingv1alpha1.PackageInstall, error) {
	var pkgInstall packagingv1alpha1.PackageInstall
	resource, err := s.getPkgInstallResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.Get(ctx, identifier, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &pkgInstall)
	if err != nil {
		return nil, err
	}
	return &pkgInstall, nil
}

// getPkgRepository returns the package repository for the given cluster, namespace and identifier
func (s *Server) getPkgRepository(ctx context.Context, cluster, namespace, identifier string) (*packagingv1alpha1.PackageRepository, error) {
	var pkgRepository packagingv1alpha1.PackageRepository
	resource, err := s.getPkgRepositoryResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.Get(ctx, identifier, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &pkgRepository)
	if err != nil {
		return nil, err
	}
	return &pkgRepository, nil
}

// getApp returns the app for the given cluster, namespace and identifier
func (s *Server) getApp(ctx context.Context, cluster, namespace, identifier string) (*kappctrlv1alpha1.App, error) {
	var app kappctrlv1alpha1.App
	resource, err := s.getAppResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.Get(ctx, identifier, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

//  List of resources getters

// getPkgs returns the list of packages for the given cluster and namespace
func (s *Server) getPkgs(ctx context.Context, cluster, namespace string) ([]*datapackagingv1alpha1.Package, error) {
	return s.getPkgsWithFieldSelector(ctx, cluster, namespace, "")
}

// getPkgs returns the list of packages for the given cluster and namespace
func (s *Server) getPkgsWithFieldSelector(ctx context.Context, cluster, namespace, fieldSelector string) ([]*datapackagingv1alpha1.Package, error) {
	resource, err := s.getPkgResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{}
	if fieldSelector != "" {
		listOptions.FieldSelector = fieldSelector
	}
	// TODO(agamez): this function takes way too long (1-2 seconds!). Try to reduce it
	// More context at: https://github.com/kubeapps/kubeapps/pull/3784#discussion_r756259504
	unstructured, err := resource.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	var pkgs []*datapackagingv1alpha1.Package
	for _, unstructured := range unstructured.Items {
		pkg := &datapackagingv1alpha1.Package{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, pkg)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

// getPkgMetadatas returns the list of package metadatas for the given cluster and namespace
func (s *Server) getPkgMetadatas(ctx context.Context, cluster, namespace string) ([]*datapackagingv1alpha1.PackageMetadata, error) {
	resource, err := s.getPkgMetadataResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var pkgMetadatas []*datapackagingv1alpha1.PackageMetadata
	for _, unstructured := range unstructured.Items {
		pkgMetadata := &datapackagingv1alpha1.PackageMetadata{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, pkgMetadata)
		if err != nil {
			return nil, err
		}
		pkgMetadatas = append(pkgMetadatas, pkgMetadata)
	}
	return pkgMetadatas, nil
}

// getPkgInstalls returns the list of package installs for the given cluster and namespace
func (s *Server) getPkgInstalls(ctx context.Context, cluster, namespace string) ([]*packagingv1alpha1.PackageInstall, error) {
	resource, err := s.getPkgInstallResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var pkgInstalls []*packagingv1alpha1.PackageInstall
	for _, unstructured := range unstructured.Items {
		pkgInstall := &packagingv1alpha1.PackageInstall{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, pkgInstall)
		if err != nil {
			return nil, err
		}
		pkgInstalls = append(pkgInstalls, pkgInstall)
	}
	return pkgInstalls, nil
}

// getPkgRepositories returns the list of package repositories for the given cluster and namespace
func (s *Server) getPkgRepositories(ctx context.Context, cluster, namespace string) ([]*packagingv1alpha1.PackageRepository, error) {
	resource, err := s.getPkgRepositoryResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var pkgRepositories []*packagingv1alpha1.PackageRepository
	for _, unstructured := range unstructured.Items {
		pkgRepository := &packagingv1alpha1.PackageRepository{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, pkgRepository)
		if err != nil {
			return nil, err
		}
		pkgRepositories = append(pkgRepositories, pkgRepository)
	}
	return pkgRepositories, nil
}

// getApps returns the list of apps for the given cluster and namespace
func (s *Server) getApps(ctx context.Context, cluster, namespace, identifier string) ([]*kappctrlv1alpha1.App, error) {
	resource, err := s.getAppResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}
	unstructured, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var apps []*kappctrlv1alpha1.App
	for _, unstructured := range unstructured.Items {
		app := &kappctrlv1alpha1.App{}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructured.Object, app)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}
	return apps, nil
}

// Creation functions

// createPkgInstall creates a package install for the given cluster, namespace and identifier
func (s *Server) createPkgInstall(ctx context.Context, cluster, namespace string, newPkgInstall *packagingv1alpha1.PackageInstall) (*packagingv1alpha1.PackageInstall, error) {
	var pkgInstall packagingv1alpha1.PackageInstall
	resource, err := s.getPkgInstallResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}

	unstructuredPkgInstallContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newPkgInstall)
	if err != nil {
		return nil, err
	}
	unstructuredPkgInstall := unstructured.Unstructured{}
	unstructuredPkgInstall.SetUnstructuredContent(unstructuredPkgInstallContent)

	unstructuredNewPkgInstall, err := resource.Create(ctx, &unstructuredPkgInstall, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredNewPkgInstall.Object, &pkgInstall)
	if err != nil {
		return nil, err
	}
	return &pkgInstall, nil
}

// Deletion functions

// deletePkgInstall deletes a package install for the given cluster, namespace and identifier
func (s *Server) deletePkgInstall(ctx context.Context, cluster, namespace, identifier string) error {
	resource, err := s.getPkgInstallResource(ctx, cluster, namespace)
	if err != nil {
		return err
	}
	err = resource.Delete(ctx, identifier, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// Update functions

// createPkgInstall creates a package install for the given cluster, namespace and identifier
func (s *Server) updatePkgInstall(ctx context.Context, cluster, namespace string, newPkgInstall *packagingv1alpha1.PackageInstall) (*packagingv1alpha1.PackageInstall, error) {
	var pkgInstall packagingv1alpha1.PackageInstall
	resource, err := s.getPkgInstallResource(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}

	unstructuredPkgInstallContent, err := runtime.DefaultUnstructuredConverter.ToUnstructured(newPkgInstall)
	if err != nil {
		return nil, err
	}
	unstructuredPkgInstall := unstructured.Unstructured{}
	unstructuredPkgInstall.SetUnstructuredContent(unstructuredPkgInstallContent)

	unstructuredNewPkgInstall, err := resource.Update(ctx, &unstructuredPkgInstall, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}

	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredNewPkgInstall.Object, &pkgInstall)
	if err != nil {
		return nil, err
	}
	return &pkgInstall, nil
}

// inspectKappK8sResources returns the list of k8s resources matching the given listOptions
func (s *Server) inspectKappK8sResources(ctx context.Context, cluster, namespace, packageId string) ([]*corev1.ResourceRef, error) {
	// As per https://github.com/vmware-tanzu/carvel-kapp-controller/blob/v0.31.0/pkg/deploy/kapp.go#L151
	appName := fmt.Sprintf("%s-ctrl", packageId)

	refs := []*corev1.ResourceRef{}

	// Get the Kapp different clients
	appsClient, resourcesClient, failingAPIServicesPolicy, _, err := s.GetKappClients(ctx, cluster, namespace)
	if err != nil {
		return nil, err
	}

	// Fetch the Kapp App
	app, err := appsClient.Find(appName)
	if err != nil {
		return nil, err
	}

	// Fetch the GroupVersions used by the app
	usedGVs, err := app.UsedGVs()
	if err != nil {
		return nil, err
	}

	// Mark those GVs as required
	failingAPIServicesPolicy.MarkRequiredGVs(usedGVs)

	// Create a k8s label selector for the app
	labelSelector, err := app.LabelSelector()
	if err != nil {
		return nil, err
	}

	// List the k8s resources that match the label selector
	resources, err := resourcesClient.List(labelSelector, nil)
	if err != nil {
		return nil, err
	}

	// For each resource, generate and append the ResourceRef
	for _, resource := range resources {
		refs = append(refs, &corev1.ResourceRef{
			ApiVersion: resource.GroupVersion().String(),
			Kind:       resource.Kind(),
			Name:       resource.Name(),
			Namespace:  resource.Namespace(),
		})
	}
	return refs, nil
}
