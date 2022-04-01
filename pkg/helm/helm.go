// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package helm

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	helmregistry "helm.sh/helm/v3/pkg/registry"
	log "k8s.io/klog/v2"
	"oras.land/oras-go/pkg/content"
	orascontext "oras.land/oras-go/pkg/context"
	"oras.land/oras-go/pkg/oras"
	"oras.land/oras-go/pkg/registry"
)

type (
	// ChartPuller interface to pull a chart from an OCI registry
	ChartPuller interface {
		PullOCIChart(ociFullName string) (*bytes.Buffer, string, error)
	}

	// OCIPuller implements ChartPuller
	OCIPuller struct {
		Resolver remotes.Resolver
	}
)

// Code from Helm Registry Client. Copied here since it belonged to a internal package.
// TODO(agamez): Some adaptations on the Kubeapps/Helm side are still required to be fully able to
// use the Helm code a library instead
// More information:
// // https://github.com/helm/helm/issues/10623
// // https://github.com/vmware-tanzu/kubeapps/pull/4154
//
// This function has been slightly adapted from:
// https://github.com/helm/helm/blob/v3.8.0/pkg/registry/client.go#L249
func (p *OCIPuller) PullOCIChart(ref string) (*bytes.Buffer, string, error) {
	parsedRef, err := parseReference(ref)
	if err != nil {
		return nil, "", err
	}

	memoryStore := content.NewMemory()
	allowedMediaTypes := []string{
		helmregistry.ConfigMediaType,
	}
	minNumDescriptors := 1 // 1 for the config
	minNumDescriptors++
	allowedMediaTypes = append(allowedMediaTypes, helmregistry.ChartLayerMediaType, helmregistry.LegacyChartLayerMediaType)

	var descriptors, layers []ocispec.Descriptor
	registryStore := content.Registry{Resolver: p.Resolver}

	context := orascontext.WithLoggerFromWriter(context.Background(), os.Stdout)
	manifest, err := oras.Copy(context, registryStore, parsedRef.String(), memoryStore, "",
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(allowedMediaTypes),
		oras.WithLayerDescriptors(func(l []ocispec.Descriptor) {
			layers = l
		}))
	if err != nil {
		return nil, "", err
	}

	descriptors = append(descriptors, manifest)
	descriptors = append(descriptors, layers...)

	numDescriptors := len(descriptors)
	if numDescriptors < minNumDescriptors {
		return nil, "", fmt.Errorf("manifest does not contain minimum number of descriptors (%d), descriptors found: %d",
			minNumDescriptors, numDescriptors)
	}
	var configDescriptor *ocispec.Descriptor
	var chartDescriptor *ocispec.Descriptor
	for _, descriptor := range descriptors {
		d := descriptor
		switch d.MediaType {
		case helmregistry.ConfigMediaType:
			configDescriptor = &d
		case helmregistry.ChartLayerMediaType:
			chartDescriptor = &d
		case helmregistry.LegacyChartLayerMediaType:
			chartDescriptor = &d
			log.Infof("Warning: chart media type %s is deprecated\n", helmregistry.LegacyChartLayerMediaType)
		}
	}
	if configDescriptor == nil {
		return nil, "", fmt.Errorf("could not load config with mediatype %s", helmregistry.ConfigMediaType)
	}
	if chartDescriptor == nil {
		return nil, "", fmt.Errorf("manifest does not contain a layer with mediatype %s",
			helmregistry.ChartLayerMediaType)
	}

	_, chartData, ok := memoryStore.Get(*chartDescriptor)
	if !ok {
		return nil, manifest.Digest.String(), fmt.Errorf("Unable to retrieve blob with digest %s", chartDescriptor.Digest)
	}

	return bytes.NewBuffer(chartData), manifest.Digest.String(), nil
}

// Code from Helm Registry Client. Copied here since it belonged to a internal package.
// TODO(agamez): Some adaptations on the Kubeapps/Helm side are still required to be fully able to
// use the Helm code a library instead
// More information:
// // https://github.com/helm/helm/issues/10623
// // https://github.com/vmware-tanzu/kubeapps/pull/4154
//
// https://github.com/helm/helm/blob/v3.8.0/pkg/registry/util.go#L112
func parseReference(raw string) (registry.Reference, error) {
	parts := strings.Split(raw, ":")
	if len(parts) > 1 && !strings.Contains(parts[len(parts)-1], "/") {
		tag := parts[len(parts)-1]

		if tag != "" {
			// Replace any plus (+) signs with known underscore (_) conversion
			newTag := strings.ReplaceAll(tag, "+", "_")
			raw = strings.ReplaceAll(raw, tag, newTag)
		}
	}
	return registry.ParseReference(raw)
}
