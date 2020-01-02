// The Dashboard doesn't understand the Helm 3 release format.
// This file is a compatibility layer that translates Helm 3 releases to a Helm 2-similar format suitable for the Dashboard.

package handler

import (
	"strings"

	"github.com/golang/protobuf/ptypes/any"
	"gopkg.in/yaml.v2"
	h3chart "helm.sh/helm/v3/pkg/chart"
	h3 "helm.sh/helm/v3/pkg/release"
	h2chart "k8s.io/helm/pkg/proto/hapi/chart"
	h2 "k8s.io/helm/pkg/proto/hapi/release"
)

// generatedYamlHeader is prepended to YAML generated from the internal map[string]interface{} representation.
const generatedYamlHeader = "# Not original YAML! Generated from parsed representation."

type dashboardCompatibleRelease struct {
	Name      string                    `json:"name,omitempty"`
	Info      h2.Info                   `json:"info,omitempty"`
	Chart     dashboardCompatibleChart  `json:"chart,omitempty"`
	Config    dashboardCompatibleConfig `json:"config,omitempty"`
	Manifest  string                    `json:"manifest,omitempty"`
	Version   int                       `json:"version,omitempty"`
	Namespace string                    `json:"namespace,omitempty"`
}

type dashboardCompatibleChart struct {
	Files     []*any.Any                `json:"files,omitempty"`
	Metadata  h3chart.Metadata          `json:"metadata,omitempty"`
	Templates []*h2chart.Template       `json:"templates,omitempty"`
	Values    dashboardCompatibleValues `json:"values,omitempty"`
}

type dashboardCompatibleValues struct {
	Raw string `json:"raw,omitempty"`
}

type dashboardCompatibleConfig struct {
	Raw string `json:"raw,omitempty"`
}

func newDashboardCompatibleRelease(h3r h3.Release) dashboardCompatibleRelease {
	return dashboardCompatibleRelease{
		Name:      h3r.Name,
		Info:      h2.Info{Status: compatibleStatus(*h3r.Info)},
		Chart:     compatibleChart(*h3r.Chart),
		Config:    compatibleConfig(h3r),
		Manifest:  h3r.Manifest,
		Version:   h3r.Version,
		Namespace: h3r.Namespace,
	}
}

func compatibleChart(h3c h3chart.Chart) dashboardCompatibleChart {
	return dashboardCompatibleChart{
		Files:     compatibleFiles(h3c.Files),
		Metadata:  *h3c.Metadata,
		Templates: compatibleTemplates(h3c.Templates),
		Values:    compatibleValues(h3c),
	}
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

func compatibleValues(h3c h3chart.Chart) dashboardCompatibleValues {
	return dashboardCompatibleValues{
		Raw: valuesToYaml(h3c.Values),
	}
}

func compatibleConfig(h3r h3.Release) dashboardCompatibleConfig {
	return dashboardCompatibleConfig{
		Raw: valuesToYaml(h3r.Config),
	}
}

// valuesToYaml serializes to YAML and prepends an informative header.
// It assumes that the serialization succeeds.
func valuesToYaml(values map[string]interface{}) string {
	marshaled, _ := yaml.Marshal(values)
	return generatedYamlHeader + "\n" + string(marshaled)
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
