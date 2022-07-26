// Copyright 2020-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	"github.com/distribution/distribution/reference"
	"gopkg.in/yaml.v3" // The usual "sigs.k8s.io/yaml" is not used because we're dealing with unstructured yaml directly
	log "k8s.io/klog/v2"
)

const (
	IndexDockerIO = "index.docker.io"
	DockerIO      = "docker.io"
)

// DockerSecretsPostRenderer is a helm post-renderer (see https://helm.sh/docs/topics/advanced/#post-rendering)
// which appends image pull secrets to container images which match specified registry domains.
type DockerSecretsPostRenderer struct {
	// secrets maps a registry domain to a single secret to be used for that domain.
	secrets map[string]string
}

// NewDockerSecretsPostRenderer returns a post renderer configured with the specified secrets.
func NewDockerSecretsPostRenderer(secrets map[string]string) (*DockerSecretsPostRenderer, error) {
	r := &DockerSecretsPostRenderer{}
	r.secrets = map[string]string{}
	// Docker authentication credentials can be stored as either the registry domain
	// or explicitly with the protocol and potential path of the server.
	// We want to compare on the registry domain only when making the decision whether to
	// include the imagePullSecret, but note this does not change the server reference in
	// the secret itself.
	for registryServer, secretName := range secrets {
		// To use net/url to parse the domain, a protocol must be present.
		if !strings.HasPrefix(registryServer, "https://") && !strings.HasPrefix(registryServer, "http://") {
			registryServer = "https://" + registryServer
		}
		u, err := url.Parse(registryServer)
		if err != nil {
			return nil, err
		}
		r.secrets[u.Host] = secretName

		// A special case for docker hub, where authentication credentials for dockerhub must
		// be for the registry server index.docker.io, yet the reference is for just docker.io.
		if u.Host == IndexDockerIO {
			r.secrets[DockerIO] = secretName
		}
	}
	return r, nil
}

func (r *DockerSecretsPostRenderer) processResourceList(resourceList []interface{}) {
	for _, resourceItem := range resourceList {
		resource, ok := resourceItem.(map[string]interface{})
		if !ok {
			continue
		}
		kindValue, ok := resource["kind"]
		if !ok {
			log.Errorf("invalid resource: no kind. %+v", resource)
			continue
		}

		kind, ok := kindValue.(string)
		if !ok {
			log.Errorf("invalid resource: non-string resource kind. %+v", resource)
			continue
		}
		if items, ok := resource["items"]; ok {
			if itemsSlice, ok := items.([]interface{}); ok {
				r.processResourceList(itemsSlice)
			} else {
				log.Errorf("Items of list type did not contain a slice: %+v", resource)
			}
			continue
		}

		podSpec := getResourcePodSpec(kind, resource)
		if podSpec == nil {
			continue
		}
		r.updatePodSpecWithPullSecrets(podSpec)
	}
}

// Run returns the rendered yaml including any additions of the post-renderer.
// An error is only returned if the manifests cannot be parsed or re-rendered.
func (r *DockerSecretsPostRenderer) Run(renderedManifests *bytes.Buffer) (modifiedManifests *bytes.Buffer, err error) {
	if len(r.secrets) == 0 {
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

	// TODO(mnelson): If re-rendering the entire manifest creates issues, we
	// could instead find the correct byte position and insert the image pull
	// secret into the byte stream at the relevant points, but this will be
	// more complex.
	r.processResourceList(resourceList)

	modifiedManifests = bytes.NewBuffer([]byte{})
	encoder := yaml.NewEncoder(modifiedManifests)
	defer encoder.Close()

	for _, resource := range resourceList {
		err = encoder.Encode(resource)
		if err != nil {
			return nil, err
		}
	}

	return modifiedManifests, nil
}

// updatePodSpecWithPullSecrets updates the podSpec inline with the relevant pull secrets.
// We do not parse the yaml into actual Kubernetes objects since we want to be
// independent of api versions. This requires special care and limitations, so
// we limit our assumptions of the untyped handling to the following:
// - The pod spec includes a 'containers' key with a slice value
// - Each container value is a map with an 'image' key and string value.
// An invalid resource doc is logged but left for the k8s API to respond to.
func (r *DockerSecretsPostRenderer) updatePodSpecWithPullSecrets(podSpec map[string]interface{}) {
	containersObject, ok := podSpec["containers"]
	if !ok {
		log.Errorf("podSpec contained no containers key: %+v", podSpec)
		return
	}
	containers, ok := containersObject.([]interface{})
	if !ok {
		log.Errorf("podSpec containers key is not a slice: %+v", podSpec)
		return
	}

	// If there are existing pull secrets, initialise our slice with that value
	// and additionally initialize a map keyed by secret name which we can
	// use to test existence more easily.
	var imagePullSecrets []map[string]interface{}
	existingNames := map[string]bool{}
	if existingPullSecrets, ok := podSpec["imagePullSecrets"]; ok {
		for _, s := range existingPullSecrets.([]interface{}) {
			pullSecret := s.(map[string]interface{})
			if name, ok := pullSecret["name"]; ok {
				if n, ok := name.(string); ok {
					existingNames[n] = true
					imagePullSecrets = append(imagePullSecrets, map[string]interface{}{"name": n})
				}
			}
		}
	}

	for _, c := range containers {
		container, ok := c.(map[string]interface{})
		if !ok {
			log.Errorf("pod spec container is not a map: %+v", c)
			continue
		}
		image, ok := container["image"].(string)
		if !ok {
			// NOTE: in https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#container-v1-core
			// the image is optional to allow higher level config management to default or override (such as
			// deployments or statefulsets), but both only define pod templates which in turn define containers?
			log.Errorf("pod spec container does not define an string image: %+v", container)
			continue
		}

		ref, err := reference.ParseNormalizedNamed(image)
		if err != nil {
			log.Errorf("unable to parse image reference: %q", image)
			continue
		}
		imageDomain := reference.Domain(ref)

		secretName, ok := r.secrets[imageDomain]
		if !ok {
			continue
		}
		// Only add the secret if it's not already included in the image pull secrets.
		if _, ok := existingNames[secretName]; !ok {
			log.Infof("appending imagePullSecret %q for fetching image %s", secretName, image)
			imagePullSecrets = append(imagePullSecrets, map[string]interface{}{"name": secretName})
			existingNames[secretName] = true
		}
	}

	if len(imagePullSecrets) > 0 {
		podSpec["imagePullSecrets"] = imagePullSecrets
	}
}

// getResourcePodSpec checks the kind of the resource and extracts the pod spec accordingly.
// We do not parse the yaml into actual Kubernetes objects since we want to be
// independent of api versions. This requires special care and limitations, so
// we limit our assumptions of the untyped handling to the following, with any
// invalid docs ignored and left for the API server to respond accordingly:
// - A resource doc is a map with a "kind" key with a string value
// - A pod resource doc has a "spec" key containing a map
func getResourcePodSpec(kind string, resource map[string]interface{}) map[string]interface{} {
	switch kind {
	case "Pod":
		return getMapForKeys([]string{"spec"}, resource)
	case "DaemonSet", "Deployment", "Job", "ReplicaSet", "ReplicationController", "StatefulSet":
		// These resources all include a spec.template.spec PodSpec.
		// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#podtemplatespec-v1-core
		return getMapForKeys([]string{"spec", "template", "spec"}, resource)
	case "PodTemplate":
		return getMapForKeys([]string{"template", "spec"}, resource)
	case "CronJob":
		// A CronJob spec contains a jobTemplate:
		// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/#cronjob-v1-batch
		return getMapForKeys([]string{"spec", "jobTemplate", "spec", "template", "spec"}, resource)
	}

	return nil
}

func getMapForKeys(keys []string, m map[string]interface{}) map[string]interface{} {
	current := m
	var ok bool
	for _, k := range keys {
		current, ok = current[k].(map[string]interface{})
		if !ok {
			log.Errorf("invalid resource: non-map %q, in %+v", k, m)
			return nil
		}
	}
	return current
}
