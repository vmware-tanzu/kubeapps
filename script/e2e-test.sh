#!/usr/bin/env bash

# Copyright (c) 2018 Bitnami
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

ROOT_DIR=`cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd`
DEV_TAG=${1:?}
IMG_MODIFIER=${2:-""}
HELM_TLS_FLAGS="--tls --tls-ca-cert ca.cert.pem --tls-cert helm.cert.pem --tls-key helm.key.pem"

source $ROOT_DIR/script/libtest.sh

# Setup tiller with TLS support
openssl genrsa -out ./ca.key.pem 4096
cat <<EOF >> tls_config
[ req ]
distinguished_name="req_distinguished_name"
prompt="no"

[ req_distinguished_name ]
C="ES"
ST="Andalucia"
L="Sevilla"
O="Kubeapps"
CN="localhost"
EOF
openssl req -key ca.key.pem -new -x509 -days 7300 -sha256 -out ca.cert.pem -config tls_config

## server key
openssl genrsa -out ./tiller.key.pem 4096
## client key
openssl genrsa -out ./helm.key.pem 4096

openssl req -key tiller.key.pem -new -sha256 -out tiller.csr.pem -config tls_config
openssl req -key helm.key.pem -new -sha256 -out helm.csr.pem -config tls_config

openssl x509 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in tiller.csr.pem -out tiller.cert.pem
openssl x509 -req -CA ca.cert.pem -CAkey ca.key.pem -CAcreateserial -in helm.csr.pem -out helm.cert.pem

# Start helm
helm init \
  --tiller-tls \
  --tiller-tls-cert ./tiller.cert.pem \
  --tiller-tls-key ./tiller.key.pem \
  --tiller-tls-verify \
  --tls-ca-cert ca.cert.pem

# The flag --wait is not available when using TLS flags:
# https://github.com/helm/helm/issues/4050
echo "Waiting for Tiller to be ready ... "
cnt=60 # 60 retries (about 60s)
until helm version ${HELM_TLS_FLAGS} --tiller-connection-timeout 1; do
  ((cnt=cnt-1)) || return 1
  sleep 1
done

# Add admin permissions to default user in kube-system namespace
kubectl get clusterrolebinding kube-dns-admin >& /dev/null || \
    kubectl create clusterrolebinding kube-dns-admin --serviceaccount=kube-system:default --clusterrole=cluster-admin 

# Install Kubeapps
helm dep up $ROOT_DIR/chart/kubeapps/
helm install --name kubeapps-ci --namespace kubeapps $ROOT_DIR/chart/kubeapps \
    `# Tiller TLS flags` \
    ${HELM_TLS_FLAGS} \
    `# cli TLS flags` \
    --set tillerProxy.tls.ca="$(cat ca.cert.pem)" \
    --set tillerProxy.tls.key="$(cat helm.key.pem)" \
    --set tillerProxy.tls.cert="$(cat helm.cert.pem)" \
    `# Image flags` \
    --set apprepository.image.tag=$DEV_TAG \
    --set apprepository.image.repository=kubeapps/apprepository-controller$IMG_MODIFIER \
    --set apprepository.syncImage.tag=$DEV_TAG \
    --set apprepository.syncImage.repository=kubeapps/chart-repo$IMG_MODIFIER \
    --set chartsvc.image.tag=$DEV_TAG \
    --set chartsvc.image.repository=kubeapps/chartsvc$IMG_MODIFIER \
    --set dashboard.image.tag=$DEV_TAG \
    --set dashboard.image.repository=kubeapps/dashboard$IMG_MODIFIER \
    --set tillerProxy.image.tag=$DEV_TAG \
    --set tillerProxy.image.repository=kubeapps/tiller-proxy$IMG_MODIFIER

# Ensure that we are testing the correct image
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-chartsvc $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-tiller-proxy $DEV_TAG

# Wait for Kubeapps Pods
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-internal-apprepository-controller
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-internal-chartsvc
k8s_wait_for_pod_ready kubeapps app=kubeapps-ci-internal-tiller-proxy
k8s_wait_for_pod_ready kubeapps app=mongodb
k8s_wait_for_pod_completed kubeapps apprepositories.kubeapps.com/repo-name=stable

# Run helm tests
helm test ${HELM_TLS_FLAGS} --cleanup kubeapps-ci
