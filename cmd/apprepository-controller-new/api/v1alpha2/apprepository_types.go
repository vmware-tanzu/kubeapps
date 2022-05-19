// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make" to regenerate code after modifying this file

// AppRepositorySpec is the spec for an AppRepository resource
type AppRepositorySpec struct {
	Type string            `json:"type"`
	URL  string            `json:"url"`
	Auth AppRepositoryAuth `json:"auth,omitempty"`
	// DockerRegistrySecrets is a list of dockerconfigjson secrets which exist
	// in the same namespace as the AppRepository and should be included
	// automatically for matching images.
	DockerRegistrySecrets []string `json:"dockerRegistrySecrets,omitempty"`
	// In case of an OCI type, the list of repositories is needed
	// as there is no API for the index
	OCIRepositories []string `json:"ociRepositories,omitempty"`
	// TLSInsecureSkipVerify skips TLS verification
	TLSInsecureSkipVerify bool `json:"tlsInsecureSkipVerify,omitempty"`
	// FilterRule allows to filter packages based on a JQuery
	FilterRule FilterRuleSpec `json:"filterRule,omitempty"`
	// (optional) description
	Description string `json:"description,omitempty"`
	// PassCredentials allows passing credentials with requests to other domains linked from the repository
	PassCredentials bool `json:"passCredentials,omitempty"`
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

// FilterRuleSpec defines a set of rules and aggreagation logic
type FilterRuleSpec struct {
	JQ        string            `json:"jq"`
	Variables map[string]string `json:"variables,omitempty"`
}

// AppRepositoryStatus records the observed state of the HelmRepository.
type AppRepositoryStatus struct {
	// ObservedGeneration is the last observed generation of the HelmRepository
	// object.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds the conditions for the HelmRepository.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// URL is the dynamic fetch link for the latest Artifact.
	// It is provided on a "best effort" basis, and using the precise
	// HelmRepositoryStatus.Artifact data is recommended.
	// +optional
	URL string `json:"url,omitempty"`

	// Artifact represents the last successful HelmRepository reconciliation.
	// +optional
	Artifact *Artifact `json:"artifact,omitempty"`

	meta.ReconcileRequestStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:subresource:status

const (
	// IndexationFailedReason signals that the AppRepository index fetch
	// failed.
	IndexationFailedReason string = "IndexationFailed"
)

// GetConditions returns the status conditions of the object.
func (in AppRepository) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *AppRepository) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the source must be
// reconciled again.
func (in AppRepository) GetRequeueAfter() time.Duration {
	// TODO (gfichtenholt) AppRepositorySpec should have a field for this
	// something like
	// Interval at which to check the URL for updates.
	// +required
	// Interval metav1.Duration `json:"interval"`
	return time.Duration(10 * time.Minute)
}

// GetArtifact returns the latest artifact from the source if present in the
// status sub-resource.
func (in *AppRepository) GetArtifact() *Artifact {
	return in.Status.Artifact
}

// +genclient
// +genclient:Namespaced
// +kubebuilder:storageversion
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=apprepo
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="URL",type=string,JSONPath=`.spec.url`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
// AppRepository is the Schema for the apprepositories API
type AppRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRepositorySpec   `json:"spec,omitempty"`
	Status AppRepositoryStatus `json:"status,omitempty"`
}

// AppRepositoryList contains a list of AppRepository
// +kubebuilder:object:root=true
type AppRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRepository{}, &AppRepositoryList{})
}
