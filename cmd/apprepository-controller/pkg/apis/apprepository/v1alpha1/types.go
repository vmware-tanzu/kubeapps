/*
Copyright 2017 Bitnami.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppRepository is a specification for an AppRepository resource
type AppRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRepositorySpec   `json:"spec"`
	Status AppRepositoryStatus `json:"status"`
}

// AppRepositorySpec is the spec for an AppRepository resource
type AppRepositorySpec struct {
	Type               string                 `json:"type"`
	URL                string                 `json:"url"`
	Auth               AppRepositoryAuth      `json:"auth,omitempty"`
	ResyncRequests     uint                   `json:"resyncRequests"`
	SyncJobPodTemplate corev1.PodTemplateSpec `json:"syncJobPodTemplate"`
	// DockerRegistrySecrets is a list of dockerconfigjson secrets which exist
	// in the same namespace as the AppRepository and should be included
	// automatically for matching images.
	DockerRegistrySecrets []string `json:"dockerRegistrySecrets,omitempty"`
	// In case of an OCI type, the list of repositories is needed
	// as there is no API for the index
	// DEPRECATED: Use FilterRules instead
	OCIRepositories []string `json:"ociRepositories,omitempty"`
	// TLSInsecureSkipVerify skips TLS verification
	TLSInsecureSkipVerify bool            `json:"tlsInsecureSkipVerify,omitempty"`
	FilterRules           FilterRulesSpec `json:"filterRules,omitempty"`
}

// AppRepositoryAuth is the auth for an AppRepository resource
type AppRepositoryAuth struct {
	Header   *AppRepositoryAuthHeader `json:"header,omitempty"`
	CustomCA *AppRepositoryCustomCA   `json:"customCA,omitempty"`
}

// AppRepositoryAuthHeader secret-key reference
type AppRepositoryAuthHeader struct {
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// AppRepositoryCustomCA secret-key reference
type AppRepositoryCustomCA struct {
	// Selects a key of a secret in the pod's namespace
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// FilterRulesSpec defines a set of rules and aggreagation logic
type FilterRulesSpec struct {
	AnyOf []FilterRule `json:"anyOf,omitempty"`
	AllOf []FilterRule `json:"allOf,omitempty"`
}

// FilterRule defines a rule
type FilterRule struct {
	JQ        string            `json:"jq"`
	Variables map[string]string `json:"variables,omitempty"`
}

// AppRepositoryStatus is the status for an AppRepository resource
type AppRepositoryStatus struct {
	Status string `json:"status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppRepositoryList is a list of AppRepository resources
type AppRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []AppRepository `json:"items"`
}
