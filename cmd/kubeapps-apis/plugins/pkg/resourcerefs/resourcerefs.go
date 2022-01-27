// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resourcerefs

import (
	"context"
	"errors"
	"io"
	"strings"

	pkgsGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	clientgetter "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmstoragedriver "helm.sh/helm/v3/pkg/storage/driver"
	k8syamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

type yamlMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type yamlResource struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   yamlMetadata   `json:"metadata"`
	Items      []yamlResource `json:"items"`
}

// resourceRefsFromManifest returns the resource refs for a given yaml manifest.
func ResourceRefsFromManifest(m, pkgNamespace string) ([]*pkgsGRPCv1alpha1.ResourceRef, error) {
	decoder := k8syamlutil.NewYAMLToJSONDecoder(strings.NewReader(m))
	refs := []*pkgsGRPCv1alpha1.ResourceRef{}
	doc := yamlResource{}
	for {
		err := decoder.Decode(&doc)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to decode yaml manifest: %v", err)
		}
		if doc.Kind == "" {
			continue
		}
		if doc.Kind == "List" || doc.Kind == "RoleList" || doc.Kind == "ClusterRoleList" {
			for _, i := range doc.Items {
				namespace := i.Metadata.Namespace
				if namespace == "" {
					namespace = pkgNamespace
				}
				refs = append(refs, &pkgsGRPCv1alpha1.ResourceRef{
					ApiVersion: i.APIVersion,
					Kind:       i.Kind,
					Name:       i.Metadata.Name,
					Namespace:  namespace,
				})
			}
			continue
		}
		// Helm does not require that the rendered manifest specifies the
		// resource namespace so some charts do not do so (ldap).  We explicitly
		// set the namespace for the resource ref so that it can be used as part
		// of the key for the resource ref.
		// TODO(minelson): At the moment we do not distinguish between
		// cluster-scoped and namespace-scoped resources for the refs.  This
		// does not affect the resources plugin fetching them correctly, but
		// would be better if we only set the namespace in the reference if (a)
		// it was not set in the manifest, and (b) it is a namespace-scoped
		// resource.
		namespace := doc.Metadata.Namespace
		if namespace == "" {
			namespace = pkgNamespace
		}
		refs = append(refs, &pkgsGRPCv1alpha1.ResourceRef{
			ApiVersion: doc.APIVersion,
			Kind:       doc.Kind,
			Name:       doc.Metadata.Name,
			Namespace:  namespace,
		})
	}

	return refs, nil
}

func GetInstalledPackageResourceRefs(
	ctx context.Context,
	request *pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsRequest,
	actionConfigGetter clientgetter.HelmActionConfigGetterFunc) (*pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse, error) {
	pkgRef := request.GetInstalledPackageRef()
	identifier := pkgRef.GetIdentifier()
	namespace := pkgRef.GetContext().GetNamespace()

	actionConfig, err := actionConfigGetter(ctx, namespace)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to create Helm action config: %v", err)
	}

	// Grab the released manifest from the release.
	// TODO(minelson): We're currently getting the resource refs for a package
	// install by checking the helm manifest, as we do for the helm plugin. With
	// certain assumptions about the RBAC of the Kubeapps user, we may be able
	// to instead query for labelled resources. See the discussion following for
	// more details:
	// https://github.com/kubeapps/kubeapps/pull/3811#issuecomment-977689570
	getcmd := helmaction.NewGet(actionConfig)
	release, err := getcmd.Run(identifier)
	if err != nil {
		if err == helmstoragedriver.ErrReleaseNotFound {
			return nil, grpcstatus.Errorf(grpccodes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", identifier, namespace, err)
		}
		return nil, grpcstatus.Errorf(grpccodes.Internal, "Unable to run Helm get action: %v", err)
	}

	refs, err := ResourceRefsFromManifest(release.Manifest, namespace)
	if err != nil {
		return nil, err
	}

	return &pkgsGRPCv1alpha1.GetInstalledPackageResourceRefsResponse{
		Context:      pkgRef.GetContext(),
		ResourceRefs: refs,
	}, nil
}
