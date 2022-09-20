// Copyright 2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	apprepov1alpha1 "github.com/vmware-tanzu/kubeapps/cmd/apprepository-controller/pkg/apis/apprepository/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	log "k8s.io/klog/v2"
)

func (s *Server) getPkgRepositoryResource(ctx context.Context, cluster, namespace string) (dynamic.ResourceInterface, error) {
	dynClient, err := s.clientGetter.Dynamic(ctx, cluster)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{
		Group:    apprepov1alpha1.SchemeGroupVersion.Group,
		Version:  apprepov1alpha1.SchemeGroupVersion.Version,
		Resource: AppRepositoryResource}
	ri := dynClient.Resource(gvr).Namespace(namespace)
	log.Infof("+helm getPkgRepositoryResource [%v]", ri)
	return ri, nil
}

// getPkgRepository returns the package repository for the given cluster, namespace and identifier
func (s *Server) getPkgRepository(ctx context.Context, cluster, namespace, identifier string) (*apprepov1alpha1.AppRepository, *k8scorev1.Secret, *k8scorev1.Secret, error) {
	client, err := s.getClient(ctx, cluster, namespace)
	if err != nil {
		return nil, nil, nil, err
	}

	key := types.NamespacedName{
		Name:      identifier,
		Namespace: namespace,
	}
	appRepo := &apprepov1alpha1.AppRepository{}
	if err = client.Get(ctx, key, appRepo); err != nil {
		return nil, nil, nil, statuserror.FromK8sError("get", AppRepositoryKind, key.String(), err)
	}

	// Auth and TLS
	typedClient, err := s.clientGetter.Typed(ctx, cluster)
	if err != nil {
		return nil, nil, nil, err
	}
	auth := appRepo.Spec.Auth
	var caCertSecret *k8scorev1.Secret
	if auth.CustomCA != nil {
		secretName := auth.CustomCA.SecretKeyRef.Name
		//client.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, caCertSecret)
		caCertSecret, err = typedClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read custom CA secret %q: %v", auth.CustomCA.SecretKeyRef.Name, err)
		}
	}

	var authSecret *k8scorev1.Secret
	if auth.Header != nil {
		secretName := auth.Header.SecretKeyRef.Name
		authSecret, err = typedClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to read auth secret %q: %v", secretName, err)
		}
	}

	return appRepo, caCertSecret, authSecret, nil
}

// updatePkgRepository updates a package repository for the given cluster, namespace and identifier
func (s *Server) updatePkgRepository(ctx context.Context, cluster, namespace string, newPkgRepository *apprepov1alpha1.AppRepository) error {

	client, err := s.getClient(ctx, cluster, namespace)
	if err != nil {
		return err
	}

	if err = client.Update(ctx, newPkgRepository); err != nil {
		return statuserror.FromK8sError("update", AppRepositoryKind, newPkgRepository.Name, err)
	}
	return nil
}
