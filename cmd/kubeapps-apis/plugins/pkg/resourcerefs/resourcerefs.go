// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package resourcerefs

import (
	"context"
	goerrs "errors"
	"io"
	"strings"

	corev1 "github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/storage/driver"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/yaml"
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
func ResourceRefsFromManifest(m, pkgNamespace string) ([]*corev1.ResourceRef, error) {
	decoder := yaml.NewYAMLToJSONDecoder(strings.NewReader(m))
	refs := []*corev1.ResourceRef{}
	doc := yamlResource{}
	for {
		err := decoder.Decode(&doc)
		if err != nil {
			if goerrs.Is(err, io.EOF) {
				break
			}
			return nil, status.Errorf(codes.Internal, "Unable to decode yaml manifest: %v", err)
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
				refs = append(refs, &corev1.ResourceRef{
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
		refs = append(refs, &corev1.ResourceRef{
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
	helmReleaseName types.NamespacedName,
	actionConfigGetter clientgetter.HelmActionConfigGetterFunc) ([]*corev1.ResourceRef, error) {
	namespace := helmReleaseName.Namespace

	actionConfig, err := actionConfigGetter(ctx, namespace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unable to create Helm action config: %v", err)
	}

	// Grab the released manifest from the release.
	// TODO(minelson): We're currently getting the resource refs for a package
	// install by checking the helm manifest, as we do for the helm plugin. With
	// certain assumptions about the RBAC of the Kubeapps user, we may be able
	// to instead query for labelled resources. See the discussion following for
	// more details:
	// https://github.com/vmware-tanzu/kubeapps/pull/3811#issuecomment-977689570
	getcmd := action.NewGet(actionConfig)
	release, err := getcmd.Run(helmReleaseName.Name)
	if err != nil {
		if err == driver.ErrReleaseNotFound {
			return nil, status.Errorf(codes.NotFound, "Unable to find Helm release %q in namespace %q: %+v", helmReleaseName, namespace, err)
		}
		return nil, status.Errorf(codes.Internal, "Unable to run Helm get action: %v", err)
	}

	refs, err := ResourceRefsFromManifest(release.Manifest, namespace)
	if err != nil {
		return nil, err
	} else {
		return refs, nil
	}
}
