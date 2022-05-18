// Copyright XXXX-YYYY the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppRepositorySpec defines the desired state of AppRepository
type AppRepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AppRepository. Edit apprepository_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AppRepositoryStatus defines the observed state of AppRepository
type AppRepositoryStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AppRepository is the Schema for the apprepositories API
type AppRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppRepositorySpec   `json:"spec,omitempty"`
	Status AppRepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AppRepositoryList contains a list of AppRepository
type AppRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppRepository{}, &AppRepositoryList{})
}
