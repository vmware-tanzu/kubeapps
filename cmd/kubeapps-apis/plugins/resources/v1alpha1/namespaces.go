// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	authorizationapi "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"
	"strings"

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
	log.InfoS("+resources CheckNamespaceExists", "cluster", cluster, "namespace", namespace)

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
	log.InfoS("+resources CreateNamespace", "cluster", cluster, "namespace", namespace, "labels", r.Labels)

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
			Name:   namespace,
			Labels: r.Labels,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Namespace", namespace, err)
	}

	return &v1alpha1.CreateNamespaceResponse{}, nil
}

// GetNamespaceNames returns the list of namespace names from either the cluster or the incoming trusted namespaces.
// In any case, only if the user has the required RBAC.
func (s *Server) GetNamespaceNames(ctx context.Context, r *v1alpha1.GetNamespaceNamesRequest) (*v1alpha1.GetNamespaceNamesResponse, error) {
	cluster := r.GetCluster()
	log.InfoS("+resources GetNamespaceNames ", "cluster", cluster)

	// Check if there are trusted namespaces in the request
	trustedNamespaces, err := getTrustedNamespacesFromHeader(ctx, s.pluginConfig.TrustedNamespaces.HeaderName, s.pluginConfig.TrustedNamespaces.HeaderPattern)
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Namespaces", "", err)
	}

	namespaceList, err := s.GetAccessibleNamespaces(ctx, cluster, trustedNamespaces)
	if err != nil {
		return nil, statuserror.FromK8sError("list", "Namespaces", "", err)
	}

	namespaces := make([]string, len(namespaceList))
	for i, ns := range namespaceList {
		namespaces[i] = ns.Name
	}

	return &v1alpha1.GetNamespaceNamesResponse{
		NamespaceNames: namespaces,
	}, nil
}

// CanI Checks if the operation can be performed according to incoming auth rbac
func (s *Server) CanI(ctx context.Context, r *v1alpha1.CanIRequest) (*v1alpha1.CanIResponse, error) {
	if r.GetContext() == nil {
		return nil, status.Errorf(codes.InvalidArgument, "context parameter is required")
	}
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	if cluster == "" {
		return nil, status.Errorf(codes.InvalidArgument, "cluster parameter is required")
	}
	log.InfoS("+resources CanI", "cluster", cluster, "namespace", namespace, "group", r.GetGroup(), "resource", r.GetResource(), "verb", r.GetVerb())

	var typedClient kubernetes.Interface
	var err error
	if s.kubeappsCluster != cluster && strings.ToLower(r.GetVerb()) == "list" && strings.ToLower(r.GetResource()) == "namespaces" {
		// Listing namespaces in additional clusters might involve using the provided service account token
		typedClient, _, err = s.clusterServiceAccountClientGetter(ctx, cluster)
	} else {
		typedClient, _, err = s.clientGetter(ctx, cluster)
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	reviewResult, err := typedClient.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Group:     r.Group,
				Resource:  r.Resource,
				Verb:      r.Verb,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return &v1alpha1.CanIResponse{
		Allowed: reviewResult.Status.Allowed,
	}, nil
}
