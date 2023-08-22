// Copyright 2021-2023 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"

	authorizationapi "k8s.io/api/authorization/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/bufbuild/connect-go"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"github.com/vmware-tanzu/kubeapps/cmd/kubeapps-apis/plugins/pkg/connecterror"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// CheckNamespaceExists returns whether a namespace exists on the cluster, or
// an error if the user does not have the required RBAC.
func (s *Server) CheckNamespaceExists(ctx context.Context, r *connect.Request[v1alpha1.CheckNamespaceExistsRequest]) (*connect.Response[v1alpha1.CheckNamespaceExistsResponse], error) {
	namespace := r.Msg.GetContext().GetNamespace()
	cluster := r.Msg.GetContext().GetCluster()
	log.InfoS("+resources CheckNamespaceExists", "cluster", cluster, "namespace", namespace)

	typedClient, err := s.clientGetter.Typed(r.Header(), cluster)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the k8s client: '%w'", err))
	}

	_, err = typedClient.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return connect.NewResponse(&v1alpha1.CheckNamespaceExistsResponse{
				Exists: false,
			}), nil
		}
		return nil, connecterror.FromK8sError("get", "Namespace", namespace, err)
	}

	return connect.NewResponse(&v1alpha1.CheckNamespaceExistsResponse{
		Exists: true,
	}), nil
}

// CreateNamespace create the namespace for the given context
// if the user has the required RBAC
func (s *Server) CreateNamespace(ctx context.Context, r *connect.Request[v1alpha1.CreateNamespaceRequest]) (*connect.Response[v1alpha1.CreateNamespaceResponse], error) {
	namespace := r.Msg.GetContext().GetNamespace()
	cluster := r.Msg.GetContext().GetCluster()
	log.InfoS("+resources CreateNamespace", "cluster", cluster, "namespace", namespace, "labels", r.Msg.Labels)

	typedClient, err := s.clientGetter.Typed(r.Header(), cluster)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the k8s client: '%w'", err))
	}

	_, err = typedClient.CoreV1().Namespaces().Create(ctx, &core.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: r.Msg.Labels,
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, connecterror.FromK8sError("get", "Namespace", namespace, err)
	}

	return connect.NewResponse(&v1alpha1.CreateNamespaceResponse{}), nil
}

// GetNamespaceNames returns the list of namespace names from either the cluster or the incoming trusted namespaces.
// In any case, only if the user has the required RBAC.
func (s *Server) GetNamespaceNames(ctx context.Context, r *connect.Request[v1alpha1.GetNamespaceNamesRequest]) (*connect.Response[v1alpha1.GetNamespaceNamesResponse], error) {
	cluster := r.Msg.GetCluster()
	log.InfoS("+resources GetNamespaceNames ", "cluster", cluster)

	// Check if there are trusted namespaces in the request
	trustedNamespaces, err := getTrustedNamespacesFromHeader(ctx, s.pluginConfig.TrustedNamespaces.HeaderName, s.pluginConfig.TrustedNamespaces.HeaderPattern)
	if err != nil {
		return nil, connecterror.FromK8sError("get", "Namespaces", "", err)
	}

	namespaceList, err := s.GetAccessibleNamespaces(ctx, r.Header(), cluster, trustedNamespaces)
	if err != nil {
		return nil, connecterror.FromK8sError("list", "Namespaces", "", err)
	}

	namespaces := make([]string, len(namespaceList))
	for i, ns := range namespaceList {
		namespaces[i] = ns.Name
	}

	return connect.NewResponse(&v1alpha1.GetNamespaceNamesResponse{
		NamespaceNames: namespaces,
	}), nil
}

// CanI Checks if the operation can be performed according to incoming auth rbac
func (s *Server) CanI(ctx context.Context, r *connect.Request[v1alpha1.CanIRequest]) (*connect.Response[v1alpha1.CanIResponse], error) {
	if r.Msg.GetContext() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Context parameter is required"))
	}
	namespace := r.Msg.GetContext().GetNamespace()
	cluster := r.Msg.GetContext().GetCluster()
	if cluster == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Cluster parameter is required"))
	}
	log.InfoS("+resources CanI", "cluster", cluster, "namespace", namespace, "group", r.Msg.GetGroup(), "resource", r.Msg.GetResource(), "verb", r.Msg.GetVerb())

	var typedClient kubernetes.Interface
	var err error
	if s.kubeappsCluster != cluster && strings.ToLower(r.Msg.GetVerb()) == "list" && strings.ToLower(r.Msg.GetResource()) == "namespaces" {
		// Listing namespaces in additional clusters might involve using the provided service account token
		typedClient, err = s.clusterServiceAccountClientGetter.Typed(r.Header(), cluster)
	} else {
		typedClient, err = s.clientGetter.Typed(r.Header(), cluster)
	}
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("Unable to get the k8s client: '%w'", err))
	}

	reviewResult, err := typedClient.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &authorizationapi.SelfSubjectAccessReview{
		Spec: authorizationapi.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationapi.ResourceAttributes{
				Group:     r.Msg.Group,
				Resource:  r.Msg.Resource,
				Verb:      r.Msg.Verb,
				Namespace: namespace,
			},
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&v1alpha1.CanIResponse{
		Allowed: reviewResult.Status.Allowed,
	}), nil
}
