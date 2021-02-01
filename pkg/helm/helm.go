package helm

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd/remotes"
	"github.com/deislabs/oras/pkg/content"
	orascontext "github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

// Code from Helm Registry Client. Copied here since it belongs to a internal package.
// TODO: Use helm as a library instead once the code is moved from "internal/experimental".
// More info at: https://github.com/helm/helm/issues/9275
//
// https://github.com/helm/helm/blob/6297c021cbda1483d8c08a8ec6f4a99e38be7302/internal/experimental/registry/util.go
// ctx retrieves a fresh context.
// disable verbose logging coming from ORAS (unless debug is enabled)
func ctx(out io.Writer, debug bool) context.Context {
	if !debug {
		return orascontext.Background()
	}
	ctx := orascontext.WithLoggerFromWriter(context.Background(), out)
	orascontext.GetLogger(ctx).Logger.SetLevel(logrus.DebugLevel)
	return ctx
}

const (
	// HelmChartConfigMediaType is the reserved media type for the Helm chart manifest config
	HelmChartConfigMediaType = "application/vnd.cncf.helm.config.v1+json"

	// HelmChartContentLayerMediaType is the reserved media type for Helm chart package content
	HelmChartContentLayerMediaType = "application/tar+gzip"
)

// KnownMediaTypes returns a list of layer mediaTypes that the Helm client knows about
func KnownMediaTypes() []string {
	return []string{
		HelmChartConfigMediaType,
		HelmChartContentLayerMediaType,
	}
}

// ChartPuller interface to pull a chart from an OCI registry
type ChartPuller interface {
	PullOCIChart(ociFullName string) (*bytes.Buffer, string, error)
}

// OCIPuller implements ChartPuller
type OCIPuller struct {
	Resolver remotes.Resolver
}

// PullOCIChart Code from: https://github.com/helm/helm/blob/fee2257e3493e9d06ca6caa4be7ef7660842cbdb/internal/experimental/registry/client.go
func (p *OCIPuller) PullOCIChart(ociFullName string) (*bytes.Buffer, string, error) {
	store := content.NewMemoryStore()

	desc, layerDescriptors, err := oras.Pull(ctx(os.Stdout, log.GetLevel() == log.DebugLevel), p.Resolver, ociFullName, store,
		oras.WithPullEmptyNameAllowed(),
		oras.WithAllowedMediaTypes(KnownMediaTypes()))
	if err != nil {
		return nil, "", err
	}

	numLayers := len(layerDescriptors)
	if numLayers < 1 {
		return nil, "", fmt.Errorf("manifest does not contain at least 1 layer (total: %d)", numLayers)
	}

	var contentLayer *ocispec.Descriptor
	for _, layer := range layerDescriptors {
		layer := layer
		switch layer.MediaType {
		case HelmChartContentLayerMediaType:
			contentLayer = &layer
		}
	}

	if contentLayer == nil {
		return nil, "", errors.New(
			fmt.Sprintf("manifest does not contain a layer with mediatype %s",
				HelmChartContentLayerMediaType))
	}

	_, b, ok := store.Get(*contentLayer)
	if !ok {
		return nil, "", errors.Errorf("Unable to retrieve blob with digest %s", contentLayer.Digest)
	}

	return bytes.NewBuffer(b), desc.Digest.String(), nil
}
