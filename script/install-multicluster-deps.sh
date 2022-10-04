#!/usr/bin/env bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
IFS=$'\t\n'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
DEFAULT_DEX_IP=${DEFAULT_DEX_IP:-"172.18.0.2"}

sed -i -e "s/172.18.0.2/$DEFAULT_DEX_IP/g;s/localhost/kubeapps-ci.kubeapps/g" "${ROOT_DIR}/site/content/docs/latest/reference/manifests/kubeapps-local-dev-dex-values.yaml"
helm repo add dex https://charts.dexidp.io

# Install dex
kubectl create namespace dex
helm install dex dex/dex --version 0.5.0 --namespace dex --values "${ROOT_DIR}/site/content/docs/latest/reference/manifests/kubeapps-local-dev-dex-values.yaml"

# Install openldap
helm repo add stable https://charts.helm.sh/stable
kubectl create namespace ldap
helm install ldap stable/openldap --namespace ldap

# Create certs
kubectl -n dex create secret tls dex-web-server-tls --key "${ROOT_DIR}/devel/dex.key" --cert "${ROOT_DIR}/devel/dex.crt"
mkcert -key-file "${ROOT_DIR}/devel/localhost-key.pem" -cert-file "${ROOT_DIR}/devel/localhost-cert.pem" localhost kubeapps-ci.kubeapps "${DEFAULT_DEX_IP}"
