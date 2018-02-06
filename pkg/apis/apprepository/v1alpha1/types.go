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
	Type           string `json:"type"`
	URL            string `json:"url"`
	ResyncRequests uint   `json:"resyncRequests"`
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
