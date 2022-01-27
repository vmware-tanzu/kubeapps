// Copyright 2017-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	apprepo "github.com/kubeapps/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = k8sschema.GroupVersion{Group: apprepo.GroupName, Version: "v1alpha1"}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) k8sschema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) k8sschema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// SchemeBuilder is the SchemeBuilder for AppRepository
	SchemeBuilder = k8sruntime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme is the function to add to the scheme for AppRepository
	AddToScheme = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *k8sruntime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&AppRepository{},
		&AppRepositoryList{},
	)
	k8smetav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
