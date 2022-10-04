#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

echo "Instaling Nginx"
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install -n nginx-ingress --create-namespace nginx-ingress ingress-nginx/ingress-nginx
