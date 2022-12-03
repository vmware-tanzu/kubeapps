// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"testing"

	apimanifests "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/clientgetter"
	authorizationv1 "k8s.io/api/authorization/v1"
	apiextfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/storage/names"
	typfake "k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func newServerWithPackageManifests(t *testing.T, manifests []apimanifests.PackageManifest) (*Server, error) {
	typedClient := typfake.NewSimpleClientset()

	// ref https://stackoverflow.com/questions/68794562/kubernetes-fake-client-doesnt-handle-generatename-in-objectmeta/68794563#68794563
	typedClient.PrependReactor(
		"create", "*",
		func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			ret = action.(k8stesting.CreateAction).GetObject()
			meta, ok := ret.(metav1.Object)
			if !ok {
				return
			}
			if meta.GetName() == "" && meta.GetGenerateName() != "" {
				meta.SetName(names.SimpleNameGenerator.GenerateName(meta.GetGenerateName()))
			}
			return
		})

	// Creating an authorized clientGetter
	typedClient.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{Allowed: true},
		}, nil
	})

	apiExtClient := apiextfake.NewSimpleClientset(packageManifestCRD)
	ctrlClient := newCtrlClient(manifests)
	clientGetter := clientgetter.NewBuilder().
		WithApiExt(apiExtClient).
		WithTyped(typedClient).
		WithControllerRuntime(ctrlClient).
		Build()
	return newServer(t, clientGetter)
}

// This func does not create a kubernetes dynamic client. It is meant to work in conjunction with
// a call to fake.NewSimpleDynamicClientWithCustomListKinds. The reason for argument repos
// (unlike charts or releases) is that repos are treated special because
// a new instance of a Server object is only returned once the cache has been synced with indexed repos
func newServer(t *testing.T, clientGetter clientgetter.ClientProviderInterface) (*Server, error) {

	stopCh := make(chan struct{})
	t.Cleanup(func() { close(stopCh) })

	s := &Server{
		clientGetter:    clientGetter,
		kubeappsCluster: KubeappsCluster,
		pluginConfig:    NewDefaultPluginConfig(),
	}
	return s, nil
}
