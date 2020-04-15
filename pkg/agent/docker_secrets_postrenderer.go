package agent

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"
)

// DockerSecretsPostRenderer is a helm post-renderer (see https://helm.sh/docs/topics/advanced/#post-rendering)
// which appends image pull secrets to container images which match specified registry domains.
type DockerSecretsPostRenderer struct {
	// secrets maps a registry domain to a single secret to be used for that domain.
	secrets map[string]string
}

// NewDockerSecretsPostRenderer returns a post renderer configured with the specified secrets.
func NewDockerSecretsPostRenderer(secrets map[string]string) *DockerSecretsPostRenderer {
	return &DockerSecretsPostRenderer{secrets}
}

// Run returns the rendered yaml including any additions of the post-renderer.
func (r *DockerSecretsPostRenderer) Run(renderedManifests *bytes.Buffer) (modifiedManifests *bytes.Buffer, err error) {
	if r.secrets == nil {
		return renderedManifests, nil
	}

	decoder := yaml.NewDecoder(renderedManifests)
	var resourceList []interface{}
	for {
		var resource interface{}
		err := decoder.Decode(&resource)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		resourceList = append(resourceList, resource)
	}

	// TODO: Update relevant container images with image pull secrets

	modifiedManifests = bytes.NewBuffer([]byte{})
	encoder := yaml.NewEncoder(modifiedManifests)
	defer encoder.Close()
	encoder.SetIndent(2)

	for _, resource := range resourceList {
		err = encoder.Encode(resource)
		if err != nil {
			return nil, err
		}
	}

	return modifiedManifests, nil
}
