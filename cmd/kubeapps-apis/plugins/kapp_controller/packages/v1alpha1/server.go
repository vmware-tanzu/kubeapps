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
	"k8s.io/client-go/rest"
)

// Server implements the kapp-controller packages v1alpha1 interface.
type Server struct {
	v1alpha1.UnimplementedPackagesServiceServer
}

// GetAvailablePackages streams the available packages based on the request.
func (s *Server) GetAvailablePackages(ctx context.Context, request *corev1.GetAvailablePackagesRequest) (*corev1.GetAvailablePackagesResponse, error) {
	// TODO: replace incluster config with the user config using token from request meta.
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to create incluster config: %w", err)
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create dynamic client: %w", err)
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
