// Copyright 2021-2022 the Kubeapps contributors.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	resourcesGRPCv1alpha1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	statuserror "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugins/pkg/statuserror"
	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// CreateSecret creates the secret in the given context if the user has the
// required RBAC
func (s *Server) CreateSecret(ctx context.Context, r *resourcesGRPCv1alpha1.CreateSecretRequest) (*resourcesGRPCv1alpha1.CreateSecretResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.Infof("+resources CreateSecret (cluster: %q, namespace=%q)", cluster, namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "unable to get the k8s client: '%v'", err)
	}

	_, err = typedClient.CoreV1().Secrets(namespace).Create(ctx, &k8scorev1.Secret{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       k8scorev1.ResourceSecrets.String(),
			APIVersion: k8scorev1.SchemeGroupVersion.WithResource(k8scorev1.ResourceSecrets.String()).String(),
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Namespace: namespace,
			Name:      r.GetName(),
		},
		Type:       k8sTypeForProtoType(r.GetType()),
		StringData: r.GetStringData(),
	}, k8smetav1.CreateOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("get", "Namespace", namespace, err)
	}

	return &resourcesGRPCv1alpha1.CreateSecretResponse{}, nil
}

func k8sTypeForProtoType(secretType resourcesGRPCv1alpha1.SecretType) k8scorev1.SecretType {
	switch secretType {
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED:
		return k8scorev1.SecretTypeOpaque
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_SERVICE_ACCOUNT_TOKEN:
		return k8scorev1.SecretTypeServiceAccountToken
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG:
		return k8scorev1.SecretTypeDockercfg
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG_JSON:
		return k8scorev1.SecretTypeDockerConfigJson
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_BASIC_AUTH:
		return k8scorev1.SecretTypeBasicAuth
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_SSH_AUTH:
		return k8scorev1.SecretTypeSSHAuth
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_TLS:
		return k8scorev1.SecretTypeTLS
	case resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_BOOTSTRAP_TOKEN:
		return k8scorev1.SecretTypeBootstrapToken
	}
	return k8scorev1.SecretTypeOpaque
}

func protoTypeForK8sType(secretType k8scorev1.SecretType) resourcesGRPCv1alpha1.SecretType {
	switch secretType {
	case k8scorev1.SecretTypeOpaque:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED
	case k8scorev1.SecretTypeServiceAccountToken:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_SERVICE_ACCOUNT_TOKEN
	case k8scorev1.SecretTypeDockercfg:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG
	case k8scorev1.SecretTypeDockerConfigJson:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG_JSON
	case k8scorev1.SecretTypeBasicAuth:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_BASIC_AUTH
	case k8scorev1.SecretTypeSSHAuth:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_SSH_AUTH
	case k8scorev1.SecretTypeTLS:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_TLS
	case k8scorev1.SecretTypeBootstrapToken:
		return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_BOOTSTRAP_TOKEN
	}
	return resourcesGRPCv1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED
}

// GetSecretNames returns a map of secret names with their types for the given
// context if the user has the required RBAC.
func (s *Server) GetSecretNames(ctx context.Context, r *resourcesGRPCv1alpha1.GetSecretNamesRequest) (*resourcesGRPCv1alpha1.GetSecretNamesResponse, error) {
	cluster := r.GetContext().GetCluster()
	namespace := r.GetContext().GetNamespace()
	log.Infof("+resources GetSecretNames (cluster: %q, namespace: %q)", cluster, namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, grpcstatus.Errorf(grpccodes.Internal, "unable to get the k8s client: '%v'", err)
	}

	secretList, err := typedClient.CoreV1().Secrets(namespace).List(ctx, k8smetav1.ListOptions{})
	if err != nil {
		return nil, statuserror.FromK8sError("list", "Secrets", "", err)
	}

	secrets := map[string]resourcesGRPCv1alpha1.SecretType{}
	for _, s := range secretList.Items {
		secrets[s.Name] = protoTypeForK8sType(s.Type)
	}

	return &resourcesGRPCv1alpha1.GetSecretNamesResponse{
		SecretNames: secrets,
	}, nil
}
