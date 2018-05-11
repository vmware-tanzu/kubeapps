/*
Copyright (c) 2017 Bitnami

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

package gke

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// BuildCrbObject builds the clusterrolebinding for granting
// cluster-admin permission to the active user.
func BuildCrbObject(user string) ([]*unstructured.Unstructured, error) {
	crb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "ClusterRoleBinding",
			"apiVersion": "rbac.authorization.k8s.io/v1beta1",
			"metadata": map[string]interface{}{
				"name": "kubeapps-cluster-admin",
			},
			"roleRef": map[string]interface{}{
				"apiGroup": "rbac.authorization.k8s.io",
				"kind":     "ClusterRole",
				"name":     "cluster-admin",
			},
			"subjects": []map[string]interface{}{
				{
					"apiGroup": "rbac.authorization.k8s.io",
					"kind":     "User",
					"name":     user,
				},
			},
		},
	}
	crbs := []*unstructured.Unstructured{}
	crbs = append(crbs, crb)
	return crbs, nil
}

// GetActiveUser returns user logging to gcloud
func GetActiveUser() (string, error) {
	stdout, err := exec.Command("gcloud", "config", "get-value", "core/account").Output()
	if err != nil {
		return "", fmt.Errorf("error executing `gcloud config get-value core/account`, please make sure the Google Cloud SDK is in your system path: %v", err)
	}
	activeUser := strings.TrimSpace(string(stdout))
	if activeUser == "" {
		return "", fmt.Errorf("could not get active user from your Google Cloud SDK configuration, please login using `gcloud auth login`")
	}
	return activeUser, nil
}
