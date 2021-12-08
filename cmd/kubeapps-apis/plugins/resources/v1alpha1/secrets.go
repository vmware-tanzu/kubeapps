/*
Copyright © 2021 VMware
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

package main

import (
	"context"

	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/plugins/resources/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	log "k8s.io/klog/v2"
)

// CreateSecret creates the secret in the given context if the user has the
// required RBAC
func (s *Server) CreateSecret(ctx context.Context, r *v1alpha1.CreateSecretRequest) (*v1alpha1.CreateSecretResponse, error) {
	namespace := r.GetContext().GetNamespace()
	cluster := r.GetContext().GetCluster()
	log.Infof("+resources CreateSecret (cluster: %q, namespace=%q)", cluster, namespace)

	typedClient, _, err := s.clientGetter(ctx, cluster)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get the k8s client: '%v'", err)
	}

	_, err = typedClient.CoreV1().Secrets(namespace).Create(ctx, &core.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      r.GetName(),
		},
		Type:       secretTypeForEnum(r.GetType()),
		StringData: r.GetStringData(),
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, errorByStatus("get", "Namespace", namespace, err)
	}

	return &v1alpha1.CreateSecretResponse{}, nil
}

func secretTypeForEnum(secretType v1alpha1.SecretType) core.SecretType {
	switch secretType {
	case v1alpha1.SecretType_SECRET_TYPE_OPAQUE_UNSPECIFIED:
		return core.SecretTypeOpaque
	case v1alpha1.SecretType_SECRET_TYPE_SERVICE_ACCOUNT_TOKEN:
		return core.SecretTypeServiceAccountToken
	case v1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG:
		return core.SecretTypeDockercfg
	case v1alpha1.SecretType_SECRET_TYPE_DOCKER_CONFIG_JSON:
		return core.SecretTypeDockerConfigJson
	case v1alpha1.SecretType_SECRET_TYPE_BASIC_AUTH:
		return core.SecretTypeBasicAuth
	case v1alpha1.SecretType_SECRET_TYPE_SSH_AUTH:
		return core.SecretTypeSSHAuth
	case v1alpha1.SecretType_SECRET_TYPE_TLS:
		return core.SecretTypeTLS
	case v1alpha1.SecretType_SECRET_TYPE_BOOTSTRAP_TOKEN:
		return core.SecretTypeBootstrapToken
	}
	return core.SecretTypeOpaque
}
