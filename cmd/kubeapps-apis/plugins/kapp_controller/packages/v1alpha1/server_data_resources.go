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
	"sync"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	datapackagingv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
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

// appLabelIdentifier returns the app label identifier for the given Kapp app
//
// Apparently, App CRs not being created by a "kapp deploy" command don't have the proper annotations.
// So, in order to retrieve the annotation value,
// we have to get the ConfigMap <AppName>-ctrl and, then, fetch the value of the key "labelValue" in data.spec.
// See https://kubernetes.slack.com/archives/CH8KCCKA5/p1637842398026700
// https://github.com/vmware-tanzu/carvel-kapp-controller/issues/430
func (s *Server) appLabelIdentifier(ctx context.Context, cluster, namespace, installedPackageRefId string) (string, error) {
	typedClient, _, err := s.GetClients(ctx, cluster)
	if err != nil {
		return "", status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	// the ConfigMap name is, by convention, "<appname>-ctrl", but it will change in the near future
	cmName := fmt.Sprintf("%s-ctrl", installedPackageRefId)
	cm, err := typedClient.CoreV1().ConfigMaps(namespace).Get(ctx, cmName, metav1.GetOptions{})

	if err != nil || cm.Data["spec"] != "" {
		log.Warning(errorByStatus("get", "ConfigMap", cmName, err))
	}

	appLabelValue := extractValueFromJson(cm.Data["spec"], "labelValue")

	return appLabelValue, nil
}

// findMatchingK8sResources returns the list of k8s resources matching the given listOptions
// Code inspired by https://github.com/kubernetes/kubectl/blob/release-1.22/pkg/cmd/apiresources/apiresources.go#L142
func (s *Server) findMatchingK8sResources(ctx context.Context, cluster string, listOptions metav1.ListOptions) ([]*corev1.ResourceRef, error) {
	typedClient, dynamicClient, err := s.GetClients(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	refs := []*corev1.ResourceRef{}

	// fetch every possible kubernetes resource list that is available in the cluster
	// TODO(agamez): this call may be expensive and subject to be cached;
	// have a look at the CachedDiscoveryClient in k8s
	// https://github.com/kubernetes/client-go/blob/release-1.22/discovery/cached/disk/cached_discovery.go
	apiResourceLists, err := typedClient.Discovery().ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	// for each kubernetes resource list, filter out those not being having the verb "list"
	for _, apiResourceList := range apiResourceLists {
		if len(apiResourceList.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}

		// create hannel to collect the matching k8s resources thus avoid blocking the waitgroup
		ch := make(chan resourceAndGvk)

		wg := sync.WaitGroup{}

		// for each kubernetes resource, spwan a goroutine to check if it matches the listOptions
		for _, resource := range apiResourceList.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}
			// filter to resources that support the specified "List" verb
			if !sets.NewString(resource.Verbs...).HasAll([]string{"list"}...) {
				continue
			}
			wg.Add(1)
			// spawn a goroutine to fetch the matching k8s resources
			go func(resource metav1.APIResource, ch chan resourceAndGvk, wg *sync.WaitGroup) {
				// convert the gvk to a gvr
				gvk := gv.WithKind(resource.Kind)
				gvr, _, err := s.kindToResource(s.restMapper, gvk)
				if err != nil {
					ch <- resourceAndGvk{nil, nil, err}
				}

				// ignore the namespace as the resources may appear in different ns from which the PackageInstall is created
				resources, err := dynamicClient.Resource(gvr).List(ctx, listOptions)
				if err != nil {
					ch <- resourceAndGvk{nil, nil, err}
				}

				// for each found matching k8s resource, add it to the channel
				for _, resource := range resources.Items {
					ch <- resourceAndGvk{gvk: &gvk, resource: &resource}
				}
				wg.Done()
			}(resource, ch, &wg)
		}

		// wait every goroutine to finish and close the channel
		go func() {
			wg.Wait()
			close(ch)
		}()

		// once closed, iterate over the closed channel and push the objects to the response
		for resourceAndGvk := range ch {
			// skip failing resourceAndGvk
			if resourceAndGvk.err != nil {
				continue
			}
			refs = append(refs, &corev1.ResourceRef{
				ApiVersion: resourceAndGvk.gvk.GroupVersion().String(),
				Kind:       resourceAndGvk.gvk.Kind,
				Name:       resourceAndGvk.resource.GetName(),
				Namespace:  resourceAndGvk.resource.GetNamespace(),
			})
		}
	}
	return refs, nil
}
