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

const (
	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
	packageGroup     = "package.carvel.dev"
	packageVersion   = "v1alpha1"
	packagesResource = "packages"

	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	installPackageGroup   = "install.package.carvel.dev"
	installPackageVersion = "v1alpha1"
	repositoriesResource  = "packagerepositories"
)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer

	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
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

// getClient ensures a client getter is available and uses it to return the client.
func (s *Server) GetClient(ctx context.Context) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	client, err := s.clientGetter(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return client, nil
}

// GetAvailablePackages streams the available packages based on the request.
func (s *Server) GetAvailablePackages(ctx context.Context, request *corev1.GetAvailablePackagesRequest) (*corev1.GetAvailablePackagesResponse, error) {

	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	packageResource := schema.GroupVersionResource{Group: packageGroup, Version: packageVersion, Resource: packagesResource}

	pkgs, err := client.Resource(packageResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to list kapp-controller packages: %v", err))
	}

	responsePackages := []*corev1.AvailablePackage{}
	for _, pkgUnstructured := range pkgs.Items {
		pkg, err := availablePackageFromUnstructured(&pkgUnstructured)
		if err != nil {
			return nil, err
		}
		responsePackages = append(responsePackages, pkg)
	}
	return &corev1.GetAvailablePackagesResponse{
		Packages: responsePackages,
	}, nil
}

func availablePackageFromUnstructured(ap *unstructured.Unstructured) (*corev1.AvailablePackage, error) {
	pkg := &corev1.AvailablePackage{}
	name, found, err := unstructured.NestedString(ap.Object, "spec", "publicName")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field publicName not found on kapp-controller package: %v:\n%v", err, ap.Object)
	}
	pkg.Name = name

	version, found, err := unstructured.NestedString(ap.Object, "spec", "version")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field version not found on kapp-controller package: %v:\n%v", err, ap.Object)
	}
	pkg.Version = version
	return pkg, nil
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *corev1.GetPackageRepositoriesRequest) (*corev1.GetPackageRepositoriesResponse, error) {
	client, err := s.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	repositoryResource := schema.GroupVersionResource{Group: installPackageGroup, Version: installPackageVersion, Resource: repositoriesResource}

	// Currently checks globally. Update to handle namespaced requests (?)
	repos, err := client.Resource(repositoryResource).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list kapp-controller repositories: %w", err)
	}

	responseRepos := []*corev1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo, err := packageRepositoryFromUnstructured(&repoUnstructured)
		if err != nil {
			return nil, err
		}
		responseRepos = append(responseRepos, repo)
	}
	return &corev1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

func packageRepositoryFromUnstructured(pr *unstructured.Unstructured) (*corev1.PackageRepository, error) {
	repo := &corev1.PackageRepository{}
	name, found, err := unstructured.NestedString(pr.Object, "metadata", "name")
	if err != nil || !found || name == "" {
		return nil, status.Errorf(codes.Internal, "required field metadata.name not found on PackageRepository: %v:\n%v", err, pr.Object)
	}
	repo.Name = name

	// TODO(absoludity): kapp-controller may soon introduce namespaced packagerepositories

	// See the PackageRepository CR at
	// https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	valid_url_paths := [][]string{
		{"spec", "fetch", "imgpkgBundle", "image"},
		{"spec", "fetch", "image", "url"},
		{"spec", "fetch", "http", "url"},
		{"spec", "fetch", "git", "url"},
	}

	found = false
	url := ""
	for _, path := range valid_url_paths {
		url, found, err = unstructured.NestedString(pr.Object, path...)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error fetching nested string %v from %v:\n%v", path, pr.Object, err)
		}
		if found {
			break
		}
	}

	if !found {
		return nil, status.Errorf(codes.Internal, "packagerepository without fetch of one of imgpkgBundle, image, http or git: %v", pr.Object)
	}
	repo.Url = url
	return repo, nil
}
