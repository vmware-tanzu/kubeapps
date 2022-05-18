// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha2 contains API Schema definitions for the  v1alpha2 API group
//+kubebuilder:object:generate=true
//+groupName=kubeapps.com
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "kubeapps.com", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
