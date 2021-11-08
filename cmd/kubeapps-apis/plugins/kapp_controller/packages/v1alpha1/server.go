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
	"sort"

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

	log "k8s.io/klog/v2"

	"github.com/Masterminds/semver/v3"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/kapp_controller/packages/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type clientGetter func(context.Context, string) (dynamic.Interface, error)

const (

	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#package-cr
	packageGroup             = "data.packaging.carvel.dev"
	packageVersion           = "v1alpha1"
	packageResource          = "Package"
	packageResources         = "packages"
	packageMetadataResource  = "PackageMetadata"
	packageMetadataResources = "packagemetadatas"

	// See https://carvel.dev/kapp-controller/docs/latest/packaging/#packagerepository-cr
	repositoryGroup      = "packaging.carvel.dev"
	repositoryVersion    = "v1alpha1"
	repositoryResource   = "PackageRepository"
	repositoriesResource = "packagerepositories"

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
	clientGetter clientGetter
}

// NewServer returns a Server automatically configured with a function to obtain
// the k8s client config.
func NewServer(configGetter server.KubernetesConfigGetter) *Server {
	return &Server{
		clientGetter: func(ctx context.Context, cluster string) (dynamic.Interface, error) {
			if configGetter == nil {
				return nil, status.Errorf(codes.Internal, "configGetter arg required")
			}
			config, err := configGetter(ctx, cluster)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get config : %v", err))
			}
			dynamicClient, err := dynamic.NewForConfig(config)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get dynamic client : %v", err))
			}
			return dynamicClient, nil
		},
	}
}

// getDynamicClient returns a dynamic k8s client.
func (s *Server) getDynamicClient(ctx context.Context, cluster string) (dynamic.Interface, error) {
	if s.clientGetter == nil {
		return nil, status.Errorf(codes.Internal, "server not configured with configGetter")
	}
	dynamicClient, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, fmt.Sprintf("unable to get client : %v", err))
	}
	return dynamicClient, nil
}

// GetAvailablePackageSummaries returns the available packages based on the request.
func (s *Server) GetAvailablePackageSummaries(ctx context.Context, request *corev1.GetAvailablePackageSummariesRequest) (*corev1.GetAvailablePackageSummariesResponse, error) {
	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	log.Infof("+kapp_controller GetAvailablePackageSummaries (cluster=%q, namespace:%q)", cluster, namespace)

	client, err := s.getDynamicClient(ctx, cluster)
	if err != nil {
		return nil, err
	}

	pkgGVR := schema.GroupVersionResource{Group: packageGroup, Version: packageVersion, Resource: packageResources}
	pkgs, err := client.Resource(pkgGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to list kapp-controller packages: %v", err))
	}
	pkgVersions, err := pkgVersionsMap(pkgs.Items)
	if err != nil {
		return nil, err
	}

	metaGVR := schema.GroupVersionResource{Group: packageGroup, Version: packageVersion, Resource: packageMetadataResources}
	pkgMetadatas, err := client.Resource(metaGVR).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to list kapp-controller package metadatas: %v", err))
	}

	responsePackages := make([]*corev1.AvailablePackageSummary, len(pkgMetadatas.Items))
	for i, pkgMetadata := range pkgMetadatas.Items {
		pkg, err := AvailablePackageSummaryFromUnstructured(&pkgMetadata, pkgVersions, cluster)
		if err != nil {
			return nil, err
		}
		responsePackages[i] = pkg
	}
	return &corev1.GetAvailablePackageSummariesResponse{
		AvailablePackageSummaries: responsePackages,
	}, nil
}

type packageSemver struct {
	pkg     *unstructured.Unstructured
	version *semver.Version
}

// pkgVersionsMap recturns a map of packages keyed by the packagemetadataName.
//
// A Package CR in carvel is really a particular version of a package, so we need
// to sort them by the package metadata name, since this is what they share in common.
// The packages are then sorted by version.
func pkgVersionsMap(packages []unstructured.Unstructured) (map[string][]packageSemver, error) {
	m := map[string][]packageSemver{}
	for _, pkgUnstructured := range packages {
		refName, found, err := unstructured.NestedString(pkgUnstructured.Object, "spec", "refName")
		if err != nil || !found || refName == "" {
			return nil, status.Errorf(codes.Internal, "required field spec.refName not found on kapp-controller Package: %v\n%v", err, pkgUnstructured)
		}
		version, found, err := unstructured.NestedString(pkgUnstructured.Object, "spec", "version")
		if err != nil || !found || version == "" {
			return nil, status.Errorf(codes.Internal, "required field spec.version not found on kapp-controller Package: %v\n%v", err, pkgUnstructured)
		}
		semverVersion, err := semver.NewVersion(version)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "required field spec.version was not semver compatible on kapp-controller Package: %v\n%v", err, pkgUnstructured)
		}

		m[refName] = append(m[refName], packageSemver{&pkgUnstructured, semverVersion})
	}

	for _, pkgVersions := range m {
		sort.Slice(pkgVersions, func(i, j int) bool {
			return pkgVersions[i].version.GreaterThan(pkgVersions[j].version)
		})
	}

	return m, nil
}

func AvailablePackageSummaryFromUnstructured(pkgMeta *unstructured.Unstructured, pkgVersions map[string][]packageSemver, cluster string) (*corev1.AvailablePackageSummary, error) {
	summary := &corev1.AvailablePackageSummary{}

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#package-metadata
	name, found, err := unstructured.NestedString(pkgMeta.Object, "metadata", "name")
	if err != nil || !found || name == "" {
		return nil, status.Errorf(codes.Internal, "required field metadata.name not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	namespace, found, err := unstructured.NestedString(pkgMeta.Object, "metadata", "namespace")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field metadata.namespace not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	summary.Name = name
	displayName, found, err := unstructured.NestedString(pkgMeta.Object, "spec", "displayName")
	if err != nil || !found {
		return nil, status.Errorf(codes.Internal, "required field spec.displayName not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	summary.DisplayName = displayName

	// https://carvel.dev/kapp-controller/docs/latest/packaging/#package-metadata
	versions := pkgVersions[name]
	if len(versions) == 0 {
		return nil, status.Errorf(codes.Internal, "no packages found for kapp-controller package metadata: %v", pkgMeta.Object)

	}
	summary.LatestVersion = &corev1.PackageAppVersion{
		PkgVersion: versions[0].version.String(),
	}

	// Carvel uses base64-encoded SVG data for IconSVGBase64, whereas we need
	// a url, so convert to a data-url.
	iconSVGBase64, found, err := unstructured.NestedString(pkgMeta.Object, "spec", "iconSVGBase64")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "required field spec.iconSVGBase64 not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	if found && iconSVGBase64 != "" {
		summary.IconUrl = fmt.Sprintf("data:image/svg+xml;base64,%s", iconSVGBase64)
	}

	shortDescription, found, err := unstructured.NestedString(pkgMeta.Object, "spec", "shortDescription")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "required field spec.shortDescription not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	if found {
		summary.ShortDescription = shortDescription
	}

	categories, found, err := unstructured.NestedStringSlice(pkgMeta.Object, "spec", "categories")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "required field spec.categories not found on kapp-controller packagemetadata: %v\n%v", err, pkgMeta.Object)
	}
	if found {
		summary.Categories = categories
	}

	summary.AvailablePackageRef = &corev1.AvailablePackageReference{
		Context: &corev1.Context{
			Cluster:   cluster,
			Namespace: namespace,
		},
		Plugin:     &pluginDetail,
		Identifier: name,
	}

	return summary, nil
}

// GetAvailablePackageVersions returns the package versions managed by the 'helm' plugin
func (s *Server) GetAvailablePackageVersions(ctx context.Context, request *corev1.GetAvailablePackageVersionsRequest) (*corev1.GetAvailablePackageVersionsResponse, error) {
	namespace := request.GetAvailablePackageRef().GetContext().GetNamespace()
	cluster := request.GetAvailablePackageRef().GetContext().GetCluster()
	log.Infof("+kapp-controller GetAvailablePackageVersions (cluster=%q, namespace=%q)", cluster, namespace)

	identifier := request.GetAvailablePackageRef().GetIdentifier()
	if namespace == "" || identifier == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Required context or identifier not provided")
	}
	log.Errorf("request available pkg ref was: %+v", request.GetAvailablePackageRef())

	client, err := s.getDynamicClient(ctx, cluster)
	if err != nil {
		return nil, err
	}

	// Use the field selector to return only Package CRs that match on the spec.refName.
	pkgGVR := schema.GroupVersionResource{Group: packageGroup, Version: packageVersion, Resource: packageResources}
	pkgs, err := client.Resource(pkgGVR).Namespace(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.refName=%s", identifier),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("unable to list kapp-controller packages: %v", err))
	}

	pkgVersions, err := pkgVersionsMap(pkgs.Items)
	if err != nil {
		return nil, err
	}

	// TODO(minelson): support configurable version summary for kapp-controller pkgs.
	versions := make([]*corev1.PackageAppVersion, len(pkgVersions[identifier]))
	for i, v := range pkgVersions[identifier] {
		versions[i] = &corev1.PackageAppVersion{
			PkgVersion: v.version.String(),
		}
	}

	return &corev1.GetAvailablePackageVersionsResponse{
		PackageAppVersions: versions,
	}, nil
}

// GetPackageRepositories returns the package repositories based on the request.
func (s *Server) GetPackageRepositories(ctx context.Context, request *v1alpha1.GetPackageRepositoriesRequest) (*v1alpha1.GetPackageRepositoriesResponse, error) {
	namespace := request.GetContext().GetNamespace()
	cluster := request.GetContext().GetCluster()
	log.Infof("+kapp_controller GetPackageRepositories (cluster=%q, namespace=%q)", cluster, namespace)

	client, err := s.getDynamicClient(ctx, cluster)
	if err != nil {
		return nil, err
	}

	repositoryResource := schema.GroupVersionResource{Group: repositoryGroup, Version: repositoryVersion, Resource: repositoriesResource}

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
