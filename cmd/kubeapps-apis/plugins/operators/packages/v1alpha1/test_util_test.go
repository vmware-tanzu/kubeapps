// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	apimanifests "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	KubeappsCluster = "default"
)

func newCtrlClient(manifests []apimanifests.PackageManifest) client.WithWatch {
	// Register required schema definitions
	scheme := runtime.NewScheme()
	err := apimanifests.AddToScheme(scheme)
	if err != nil {
		log.Fatalf("%s", err)
	}

	rm := apimeta.NewDefaultRESTMapper(
		[]schema.GroupVersion{
			{Group: apimanifests.Group, Version: apimanifests.Version},
		})
	rm.Add(schema.GroupVersionKind{
		Group:   apimanifests.Group,
		Version: apimanifests.Version,
		Kind:    apimanifests.PackageManifestKind},
		apimeta.RESTScopeNamespace)

	ctrlClientBuilder := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRESTMapper(rm)
	var initLists []client.ObjectList
	if len(manifests) > 0 {
		initLists = append(initLists, &apimanifests.PackageManifestList{Items: manifests})
	}
	if len(initLists) > 0 {
		ctrlClientBuilder = ctrlClientBuilder.WithLists(initLists...)
	}
	return ctrlClientBuilder.Build()
}

// misc global vars that get re-used in multiple tests
var (
	packageManifestCRD = &apiextv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: apimanifests.Group,
		},
		Status: apiextv1.CustomResourceDefinitionStatus{
			Conditions: []apiextv1.CustomResourceDefinitionCondition{
				{
					Type:   "Established",
					Status: apiextv1.ConditionStatus(metav1.ConditionTrue),
				},
			},
			StoredVersions: []string{apimanifests.Version},
		},
	}
)
