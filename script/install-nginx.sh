#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

echo "Installing Nginx"
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install -n nginx-ingress --create-namespace nginx-ingress ingress-nginx/ingress-nginx

# Wait for load balancer to get the IP
LB_IP=""
while [ -z $LB_IP ]; do
  echo "Waiting for external IP"
  LB_IP=$(kubectl -n nginx-ingress get service nginx-ingress-ingress-nginx-controller --template="{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}")
  [ -z "$LB_IP" ] && sleep 10
done
echo 'Load balancer got IP assigned: '$LB_IP
