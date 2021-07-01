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
	"k8s.io/client-go/kubernetes"

	log "k8s.io/klog/v2"

	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	packagingGroup = "packaging.carvel.dev"

	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
	packageVersion   = "v1alpha1"
	packageResource  = "PackageInstall"
	packagesResource = "packageinstalls"

	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	installPackageVersion = "v1alpha1"
	repositoryResource    = "PackageRepository"
	repositoriesResource  = "packagerepositories"

	globalPackagingNamespace = "kapp-controller-packaging-global"
)

// Compile-time statement to ensure this service implementation satisfies the core packaging API
var _ corev1.PackagesServiceServer = (*Server)(nil)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedKappControllerPackagesServiceServer
	// clientGetter is a field so that it can be switched in tests for
	// a fake client. NewServer() below sets this automatically with the
	// non-test implementation.
	clientGetter server.KubernetesClientGetter
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(clientGetter server.KubernetesClientGetter) *Server {
	return &Server{
		clientGetter: clientGetter,
	}
}

// GetClients ensures a client getter is available and uses it to return both a typed and dynamic k8s client.
func (s *Server) GetClients(ctx context.Context) (kubernetes.Interface, dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	typedClient, dynamicClient, err := s.clientGetter(ctx)
	if err != nil {
		return nil, nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return typedClient, dynamicClient, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+kapp_controller GetAvailablePackageSummaries %s", contextMsg)

	namespace := ""
	if request.Context != nil {
		if request.Context.Cluster != "" {
			return nil, status.Errorf(codes.Unimplemented, "Not supported yet: request.Context.Cluster: [%v]", request.Context.Cluster)
		}
		if request.Context.Namespace != "" {
			namespace = request.Context.Namespace
		}
	}

	_, client, err := s.GetClients(ctx)
	if err != nil {
		return nil, err
	}

	packageResource := schema.GroupVersionResource{Group: packagingGroup, Version: packageVersion, Resource: packagesResource}

	pkgs, err := client.Resource(packageResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to list kapp-controller packages: %v", err))
	}

	responsePackages := []*corev1.AvailablePackageSummary{}
	for _, pkgUnstructured := range pkgs.Items {
		pkg, err := AvailablePackageSummaryFromUnstructured(&pkgUnstructured)
		if err != nil {
			return nil, err
		}
		responsePackages = append(responsePackages, pkg)
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackagesSummaries: responsePackages,
	}, nil
}

func AvailablePackageSummaryFromUnstructured(ap *unstructured.Unstructured) (*corev1.AvailablePackageSummary, error) {
	pkg := &corev1.AvailablePackageSummary{}

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
	name, found, err := unstructured.NestedString(ap.Object, "spec", "packageRef", "refName")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field spec.packageRef.refName not found on kapp-controller package: %v:\n%v", err, ap.Object)
	}
	pkg.DisplayName = name

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
	version, found, err := unstructured.NestedString(ap.Object, "status", "version")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field status.version not found on kapp-controller package: %v:\n%v", err, ap.Object)
	}
	pkg.LatestPkgVersion = version
	return pkg, nil
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *v1alpha1.GetPackageRepositoriesRequest) (*v1alpha1.GetPackageRepositoriesResponse, error) {
	contextMsg := ""
	if request.Context != nil {
		contextMsg = fmt.Sprintf("(cluster=[%s], namespace=[%s])", request.Context.Cluster, request.Context.Namespace)
	}

	log.Infof("+kapp_controller GetPackageRepositories %s", contextMsg)

	namespace := globalPackagingNamespace
	if request.Context != nil {
		if request.Context.Cluster != "" {
			return nil, status.Errorf(codes.Unimplemented, "Not supported yet: request.Context.Cluster: [%v]", request.Context.Cluster)
		}
		if request.Context.Namespace != "" {
			namespace = request.Context.Namespace
		}
	}

	_, client, err := s.GetClients(ctx)
	if err != nil {
		return nil, err
	}

	repositoryResource := schema.GroupVersionResource{Group: packagingGroup, Version: installPackageVersion, Resource: repositoriesResource}

	// Currently checks globally. Update to handle namespaced requests (?)
	repos, err := client.Resource(repositoryResource).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to list kapp-controller repositories: %w", err)
	}

	responseRepos := []*v1alpha1.PackageRepository{}
	for _, repoUnstructured := range repos.Items {
		repo, err := packageRepositoryFromUnstructured(&repoUnstructured)
		if err != nil {
			return nil, err
		}
		responseRepos = append(responseRepos, repo)
	}
	return &v1alpha1.GetPackageRepositoriesResponse{
		Repositories: responseRepos,
	}, nil
}

func packageRepositoryFromUnstructured(pr *unstructured.Unstructured) (*v1alpha1.PackageRepository, error) {
	repo := &v1alpha1.PackageRepository{}

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	name, found, err := unstructured.NestedString(pr.Object, "metadata", "name")
	if err != nil || !found || name == "" {
		return nil, status.Errorf(codes.Internal, "required field metadata.name not found on PackageRepository: %v:\n%v", err, pr.Object)
	}
	repo.Name = name

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	namespace, found, err := unstructured.NestedString(pr.Object, "metadata", "namespace")
	if err != nil || !found || namespace == "" {
		return nil, status.Errorf(codes.Internal, "required field metadata.namespace not found on PackageRepository: %v:\n%v", err, pr.Object)
	}
	repo.Namespace = namespace

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
