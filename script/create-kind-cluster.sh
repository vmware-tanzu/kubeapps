#!/usr/bin/env bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -uo pipefail
IFS=$'\t\n'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
K8S_KIND_VERSION=${K8S_KIND_VERSION:?"not provided"}
DEFAULT_DEX_IP=${DEFAULT_DEX_IP:-"172.18.0.2"}
CLUSTER_CONFIG_FILE=${CLUSTER_CONFIG_FILE:-"${ROOT_DIR}/site/content/docs/latest/reference/manifests/kubeapps-local-dev-apiserver-config.yaml"}
CLUSTER_NAME=${CLUSTER_NAME:-"kubeapps-ci"}
KUBECONFIG=${KUBECONFIG:-"${HOME}/.kube/kind-config-kubeapps-ci"}
CONTEXT=${CONTEXT:-"kind-kubeapps-ci"}

. "${ROOT_DIR}/script/lib/liblog.sh"

info "K8S_KIND_VERSION: ${K8S_KIND_VERSION}"
info "DEFAULT_DEX_IP: ${DEFAULT_DEX_IP}"
info "CLUSTER_CONFIG_FILE: ${CLUSTER_CONFIG_FILE}"
info "CLUSTER_NAME: ${CLUSTER_NAME}"
info "KUBECONFIG: ${KUBECONFIG}"
info "CONTEXT: ${CONTEXT}"

function createKindCluster() {
  kind create cluster --image "kindest/node:${K8S_KIND_VERSION}" --name "${CLUSTER_NAME}" --config="${CLUSTER_CONFIG_FILE}" --kubeconfig="${KUBECONFIG}" --retain --wait 120s &&
  kubectl --context "${CONTEXT}" --kubeconfig="${KUBECONFIG}" apply -f ${ROOT_DIR}/site/content/docs/latest/reference/manifests/kubeapps-local-dev-users-rbac.yaml &&
  kubectl --context "${CONTEXT}" --kubeconfig="${KUBECONFIG}" apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml &&
  sleep 5 &&
  kubectl wait --context "${CONTEXT}" --kubeconfig="${KUBECONFIG}" --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s &&
  kubectl create rolebinding kubeapps-view-secret-oidc --context "${CONTEXT}" --kubeconfig="${KUBECONFIG}" --role view-secrets --user oidc:kubeapps-user@example.com &&
  kubectl create clusterrolebinding kubeapps-view-oidc --context "${CONTEXT}" --kubeconfig="${KUBECONFIG}" --clusterrole=view --user oidc:kubeapps-user@example.com &&
  echo "Cluster created"
}

sed -i "s/172.18.0.2/$DEFAULT_DEX_IP/g" "${CLUSTER_CONFIG_FILE}"
{
  echo "Creating cluster..."
  createKindCluster
} || {
  echo "Cluster creation failed, retrying..."
  kind delete clusters kubeapps-ci || true
  createKindCluster
} || {
  echo "Error while creating the cluster after retry"
  exit 1
}
