package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"oras.land/oras-go/pkg/content"
	orascontext "oras.land/oras-go/pkg/context"
	"oras.land/oras-go/pkg/oras"
)

// Code from Helm Registry Client. Copied here since it belongs to a internal package.
// TODO: Use helm as a library instead once the code is moved from "internal/experimental".
// More info at: https://github.com/helm/helm/issues/9275
//
// https://github.com/helm/helm/blob/v3.7.1/internal/experimental/registry/util.go
// ctx retrieves a fresh context.
// deactivate verbose logging coming from ORAS (unless debug is enabled)
func ctx(out io.Writer, debug bool) context.Context {
	if !debug {
		return orascontext.Background()
	}
	ctx := orascontext.WithLoggerFromWriter(context.Background(), out)
	orascontext.GetLogger(ctx).Logger.SetLevel(logrus.DebugLevel)
	return ctx
}

// Code from Helm Registry Client. Copied here since it belongs to a internal package.
// TODO: Use helm as a library instead once the code is moved from "internal/experimental".
// More info at: https://github.com/helm/helm/issues/9275
//
// https://github.com/helm/helm/blob/v3.7.1/internal/experimental/registry/constants.go#L19
const (
	// HelmChartConfigMediaType is the reserved media type for the Helm chart manifest config
	HelmChartConfigMediaType = "application/vnd.cncf.helm.config.v1+json"

	// HelmChartContentLayerMediaType is the reserved media type for Helm chart package content
	HelmChartContentLayerMediaType = "application/tar+gzip"

	// OCIScheme is the URL scheme for OCI-based requests
	OCIScheme = "oci"

	// CredentialsFileBasename is the filename for auth credentials file
	CredentialsFileBasename = "registry.json"

	// ConfigMediaType is the reserved media type for the Helm chart manifest config
	ConfigMediaType = "application/vnd.cncf.helm.config.v1+json"

	// ChartLayerMediaType is the reserved media type for Helm chart package content
	ChartLayerMediaType = "application/vnd.cncf.helm.chart.content.v1.tar+gzip"

	// ProvLayerMediaType is the reserved media type for Helm chart provenance files
	ProvLayerMediaType = "application/vnd.cncf.helm.chart.provenance.v1.prov"

	// LegacyChartLayerMediaType is the legacy reserved media type for Helm chart package content.
	LegacyChartLayerMediaType = "application/tar+gzip"
)

// ChartPuller interface to pull a chart from an OCI registry
type ChartPuller interface {
	PullOCIChart(ociFullName string) (*bytes.Buffer, string, error)
}

// OCIPuller implements ChartPuller
type OCIPuller struct {
	Resolver remotes.Resolver
}

// Code from Helm Registry Client. Copied here since it belongs to a internal package.
// TODO: Use helm as a library instead once the code is moved from "internal/experimental".
// More info at: https://github.com/helm/helm/issues/9275
//
// The code has been slightly adapted from:
// https://github.com/helm/helm/blob/v3.7.1/internal/experimental/registry/client.go#L209
func (p *OCIPuller) PullOCIChart(ociFullName string) (*bytes.Buffer, string, error) {
	ociFullName = strings.TrimPrefix(ociFullName, fmt.Sprintf("%s://", OCIScheme))
	store := content.NewMemoryStore()
	allowedMediaTypes := []string{
		ConfigMediaType,
	}
	minNumDescriptors := 1 // 1 for the config
	minNumDescriptors++

	allowedMediaTypes = append(allowedMediaTypes, ChartLayerMediaType, LegacyChartLayerMediaType)
	manifest, descriptors, err := oras.Pull(ctx(os.Stdout, log.GetLevel() == log.TraceLevel), p.Resolver, ociFullName, store,
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(allowedMediaTypes))
	if err != nil {
		return nil, "", err
	}
	numDescriptors := len(descriptors)
	if numDescriptors < minNumDescriptors {
		return nil, manifest.Digest.String(), errors.New(
			fmt.Sprintf("manifest does not contain minimum number of descriptors (%d), descriptors found: %d",
				minNumDescriptors, numDescriptors))
	}
	var chartDescriptor *ocispec.Descriptor
	for _, descriptor := range descriptors {
		d := descriptor
		switch d.MediaType {
		case ChartLayerMediaType:
			chartDescriptor = &d
		case LegacyChartLayerMediaType:
			chartDescriptor = &d
			log.Warnf("Warning: chart media type %s is deprecated\n", LegacyChartLayerMediaType)
		}
	}
	if chartDescriptor == nil {
		return nil, manifest.Digest.String(), errors.New(
			fmt.Sprintf("manifest does not contain a layer with mediatype %s or %s",
				ChartLayerMediaType, LegacyChartLayerMediaType))
	}
	_, chartData, ok := store.Get(*chartDescriptor)
	if !ok {
		return nil, manifest.Digest.String(), errors.Errorf("Unable to retrieve blob with digest %s", chartDescriptor.Digest)
	}

	return bytes.NewBuffer(chartData), manifest.Digest.String(), nil
}
