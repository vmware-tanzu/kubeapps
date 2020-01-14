// The Dashboard doesn't understand the Helm 3 release format.
// This file is a compatibility layer that translates Helm 3 releases to a Helm 2-similar format suitable for the Dashboard.
// Note that h3.Release and h2.Release are not isomorphic, so it is impossible to map between them in general.

package helm3to2

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	h3chart "helm.sh/helm/v3/pkg/chart"
	h3 "helm.sh/helm/v3/pkg/release"
	h2chart "k8s.io/helm/pkg/proto/hapi/chart"
	h2 "k8s.io/helm/pkg/proto/hapi/release"
)

var (
	// ErrUnableToConvertWithoutInfo indicates that the input release had nil Info or Chart.
	ErrUnableToConvertWithoutInfo = fmt.Errorf("unable to convert release without info")
	// ErrFailedToParseDeletionTime indicates that the deletion time of the h3 chart could not be parsed.
	ErrFailedToParseDeletionTime = fmt.Errorf("failed to parse deletion time")
)

// Convert returns a Helm2 compatible release based on the info from a Helm3 release
// TODO: This method is meant to be deleted once the support for Helm2 is dropped
func Convert(h3r h3.Release) (h2.Release, error) {
	if h3r.Info == nil || h3r.Chart == nil || h3r.Chart.Metadata == nil {
		return h2.Release{}, ErrUnableToConvertWithoutInfo
	}
	var deleted *timestamp.Timestamp
	if !h3r.Info.Deleted.IsZero() {
		var err error
		deleted, err = ptypes.TimestampProto(h3r.Info.Deleted.Time)
		if err != nil {
			return h2.Release{}, fmt.Errorf("%w: %v", ErrFailedToParseDeletionTime, err)
		}
	}
	return h2.Release{
		Name:      h3r.Name,
		Info:      &h2.Info{Status: compatibleStatus(*h3r.Info), Deleted: deleted},
		Chart:     compatibleChart(*h3r.Chart),
		Config:    compatibleConfig(h3r),
		Manifest:  h3r.Manifest,
		Version:   int32(h3r.Version),
		Namespace: h3r.Namespace,
	}, nil
}

func compatibleChart(h3c h3chart.Chart) *h2chart.Chart {
	return &h2chart.Chart{
		Files:     compatibleFiles(h3c.Files),
		Metadata:  ConvertMetadata(*h3c.Metadata),
		Templates: compatibleTemplates(h3c.Templates),
		Values:    compatibleValues(h3c),
	}
}

// ConvertMetadata turns the Metadata from a Helm3 release to a Helm2 compatibility format.
func ConvertMetadata(h3m h3chart.Metadata) *h2chart.Metadata {
	return &h2chart.Metadata{
		Annotations:   h3m.Annotations,
		ApiVersion:    h3m.APIVersion,
		AppVersion:    h3m.AppVersion,
		Condition:     h3m.Condition,
		Deprecated:    h3m.Deprecated,
		Description:   h3m.Description,
		Engine:        "",
		Home:          h3m.Home,
		Icon:          h3m.Icon,
		Keywords:      h3m.Keywords,
		KubeVersion:   h3m.KubeVersion,
		Maintainers:   compatibleMaintainers(h3m.Maintainers),
		Name:          h3m.Name,
		Sources:       h3m.Sources,
		Tags:          h3m.Tags,
		TillerVersion: "",
		Version:       h3m.Version,
	}
}

func compatibleMaintainers(h3ms []*h3chart.Maintainer) []*h2chart.Maintainer {
	h2ms := make([]*h2chart.Maintainer, len(h3ms))
	for i, m := range h3ms {
		h2ms[i] = &h2chart.Maintainer{
			Name:  m.Name,
			Email: m.Email,
			Url:   m.URL,
		}
	}
	return h2ms
}

func compatibleFiles(h3files []*h3chart.File) []*any.Any {
	anys := make([]*any.Any, len(h3files))
	for i, f := range h3files {
		anys[i] = &any.Any{
			TypeUrl: f.Name,
			Value:   f.Data,
		}
	}
	return anys
}

func compatibleTemplates(h3templates []*h3chart.File) []*h2chart.Template {
	templates := make([]*h2chart.Template, len(h3templates))
	for i, t := range h3templates {
		templates[i] = &h2chart.Template{
			Name: t.Name,
			Data: t.Data,
		}
	}
	return templates
}

func compatibleValues(h3c h3chart.Chart) *h2chart.Config {
	return &h2chart.Config{
		Raw: valuesToYaml(h3c.Values),
	}
}

func compatibleConfig(h3r h3.Release) *h2chart.Config {
	return &h2chart.Config{
		Raw: valuesToYaml(h3r.Config),
	}
}

// valuesToYaml serializes to YAML and assumes that the serialization succeeded, logging an error otherwise.
func valuesToYaml(values map[string]interface{}) string {
	marshaled, err := yaml.Marshal(values)
	if err != nil {
		log.Errorf("Failed to serialize values to YAML: %s", err.Error())
	}
	return string(marshaled)
}

func compatibleStatus(h3info h3.Info) *h2.Status {
	return &h2.Status{
		Code:  compatibleStatusCode(h3info.Status),
		Notes: h3info.Notes,
	}
}

func compatibleStatusCode(h3status h3.Status) h2.Status_Code {
	// "delet" is not a typo; "uninstalling" should become "deleting", not "deleteing".
	withDelete := strings.ReplaceAll(h3status.String(), "uninstall", "delet")
	withUnderscores := strings.ReplaceAll(withDelete, "-", "_")
	withUpperCase := strings.ToUpper(withUnderscores)
	// If the key is not found in the map, Status_UNKNOWN (0) is returned.
	return h2.Status_Code(h2.Status_Code_value[withUpperCase])
}
