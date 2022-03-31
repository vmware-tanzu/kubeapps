// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// CheckNamespaceExists returns whether a namespace exists on the cluster, or
// an error if the user does not have the required RBAC.
func (s *Server) CheckNamespaceExists(ctx context.Context, r *v1alpha1.CheckNamespaceExistsRequest) (*v1alpha1.CheckNamespaceExistsResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.Infof("+resources CheckNamespaceExists (cluster: %q, namespace=%q)", cluster, namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	_, err = typedClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return &v1alpha1.CheckNamespaceExistsResponse{
				Exists: false,
			}, nil
		}
		return nil, statuserror.FromK8sError("get", "Namespace", namespace, err)
	}

	return &v1alpha1.CheckNamespaceExistsResponse{
		Exists: true,
	}, nil
}

// CreateNamespace create the namespace for the given context
// if the user has the required RBAC
func (s *Server) CreateNamespace(ctx context.Context, r *v1alpha1.CreateNamespaceRequest) (*v1alpha1.CreateNamespaceResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.Infof("+resources CreateNamespace (cluster: %q, namespace=%q)", cluster, namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	_, err = typedClient.CoreV1().Namespaces().Create(ctx, &core.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Namespace", namespace, err)
	}

	return &v1alpha1.CreateNamespaceResponse{}, nil
}

// GetNamespaceNames returns the list of namespace names for a cluster if the
// user has the required RBAC.
//
// Note that we can't yet use this from the dashboard to replace the similar endpoint
// in kubeops until we update to ensure a configured service account can also be
// passed in (resources plugin config) and used if the user does not have RBAC.
func (s *Server) GetNamespaceNames(ctx context.Context, r *v1alpha1.GetNamespaceNamesRequest) (*v1alpha1.GetNamespaceNamesResponse, error) {
	cluster := r.GetCluster()
	log.Infof("+resources GetNamespaceNames (cluster: %q)", cluster)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	namespaceList, err := typedClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("list", "Namespaces", "", err)
	}

	namespaces := make([]string, len(namespaceList.Items))
	for i, ns := range namespaceList.Items {
		namespaces[i] = ns.Name
	}

	return &v1alpha1.GetNamespaceNamesResponse{
		NamespaceNames: namespaces,
	}, nil
}
