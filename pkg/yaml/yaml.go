/*
Copyright (c) 2018 Bitnami

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package yaml

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/ksonnet/kubecfg/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// ParseObjects returns an Unstructured object list based on the content of a YAML manifest
func ParseObjects(manifest string) ([]*unstructured.Unstructured, error) {
	r := strings.NewReader(manifest)
	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	ret := []runtime.Object{}
	nullResult := []byte("null")

	for {
		// This reader will return a single K8s resource at the time based on the --- separator
		objManifest, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(objManifest) == 0 {
			continue
		}
		jsondata, err := yaml.ToJSON(objManifest)
		if err != nil {
			return nil, err
		}

		// It is also possible that the provided yaml file is empty from the point of view
		// of the toJSON parser. For example if the yaml only contain comments.
		// In which case the returned  will be "null"
		if bytes.Equal(jsondata, nullResult) {
			continue
		}

		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return utils.FlattenToV1(ret), nil
}
