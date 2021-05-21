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

	// v1 "github.com/kubeapps/kubeapps/cmd/kubeapps-api-service/kubeappsapis/core/packagerepositories/v1"
	// *sigh*, seems different versions of the k8s client.go (at the time of writing, kapp-controller
	// is using client-go v0.19.2) means that we can't use the client here directly, as get errors like:
	/*
				gitub.com/vmware-tanzu/carvel-kapp-controller@v0.18.0/pkg/client/clientset/versioned/typed/kappctrl/v1alpha1/app.go:58:5: not enough arguments in call to c.client.Get().Namespace(c.ns).Resource("apps").Name(name).VersionedParams(&options, scheme.ParameterCodec).Do
		        have ()
		        want (context.Context)
	*/
	// So instead we use the dynamic (untyped) client.
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/rest"
)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer

	clientGetter func(context.Context) (dynamic.Interface, error)
}

// clientForRequestContext returns a k8s client for use during interactions with the cluster.
// This will be updated to use the user credential from the request context but for now
// simply returns th in-cluster config (which is linked to a service-account with demo RBAC).
func clientForRequestContext(ctx context.Context) (dynamic.Interface, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get client config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	return client, nil
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer() *Server {
	return &Server{
		clientGetter: clientForRequestContext,
	}
}

// GetAvailablePackages streams the available packages based on the request.
func (s *Server) GetAvailablePackages(ctx context.Context, request *corev1.GetAvailablePackagesRequest) (*corev1.GetAvailablePackagesResponse, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	client, err := s.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}

	packageResource := schema.GroupVersionResource{Group: "package.carvel.dev", Version: "v1alpha1", Resource: "packages"}

	pkgs, err := client.Resource(packageResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list kapp-controller packages: %w", err)
	}

	responsePackages := []*corev1.AvailablePackage{}
	for _, pkgUnstructured := range pkgs.Items {
		pkg := &corev1.AvailablePackage{}
		name, found, err := unstructured.NestedString(pkgUnstructured.Object, "spec", "publicName")
		if err != nil || !found {
			return nil, fmt.Errorf("required field publicName not found on kapp-controller package: %w:\n%v", err, pkgUnstructured.Object)
		}
		pkg.Name = name

		version, found, err := unstructured.NestedString(pkgUnstructured.Object, "spec", "version")
		if err != nil || !found {
			return nil, fmt.Errorf("required field version not found on kapp-controller package: %w:\n%v", err, pkgUnstructured.Object)
		}
		pkg.Version = version
		responsePackages = append(responsePackages, pkg)
	}
	return &corev1.GetAvailablePackagesResponse{
		Packages: responsePackages,
	}, nil
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *corev1.GetPackageRepositoriesRequest) (*corev1.GetPackageRepositoriesResponse, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create incluster config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
	}

	repositoryResource := schema.GroupVersionResource{Group: "install.package.carvel.dev", Version: "v1alpha1", Resource: "packagerepositories"}

	// Currently checks globally. Update to handle namespaced requests (?)
	repos, err := client.Resource(repositoryResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list kapp-controller repositories: %w", err)
	}

	responseRepos := []*corev1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo := &corev1.PackageRepository{}
		name, found, err := unstructured.NestedString(repoUnstructured.Object, "metadata", "name")
		if err != nil || !found {
			return nil, fmt.Errorf("required field metadata.name not found on PackageRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Name = name

		// TODO(absoludity): kapp-controller may soon introduce namespaced packagerepositories

		// TODO(absoludity): When able to add unit-tests using the fake dynamic client,
		// write failing test for handling fetch types other than imgpkgBundle and fix.
		url, found, err := unstructured.NestedString(repoUnstructured.Object, "spec", "fetch", "imgpkgBundle", "image")
		if err != nil || !found {
			return nil, fmt.Errorf("required field spec.url not found on PackageRepository: %w:\n%v", err, repoUnstructured.Object)
		}
		repo.Url = url

		responseRepos = append(responseRepos, repo)
	}
	return &corev1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}
