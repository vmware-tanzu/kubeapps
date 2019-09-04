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
CERTS_DIR="${ROOT_DIR}/script/test-certs"
HELM_CLIENT_TLS_FLAGS="--tls --tls-cert ${CERTS_DIR}/helm.cert.pem --tls-key ${CERTS_DIR}/helm.key.pem"

source $ROOT_DIR/script/libtest.sh

echo "IMAGE TAG TO BE TESTED: $DEV_TAG"
echo "IMAGE_REPO_SUFFIX: $IMG_MODIFIER"

# Print cluster version
kubectl version

# Install Tiller with TLS support
kubectl -n kube-system create sa tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init \
  --service-account tiller \
  --tiller-tls \
  --tiller-tls-cert ${CERTS_DIR}/tiller.cert.pem \
  --tiller-tls-key ${CERTS_DIR}/tiller.key.pem \
  --tiller-tls-verify \
  --tls-ca-cert ${CERTS_DIR}/ca.cert.pem

# The flag --wait is not available when using TLS flags:
# https://github.com/helm/helm/issues/4050
echo "Waiting for Tiller to be ready ... "
cnt=60 # 60 retries (about 60s)
until helm version ${HELM_CLIENT_TLS_FLAGS} --tiller-connection-timeout 1; do
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
    ${HELM_CLIENT_TLS_FLAGS} \
    `# Tiller-proxy TLS flags` \
    --set tillerProxy.tls.key="$(cat ${CERTS_DIR}/helm.key.pem)" \
    --set tillerProxy.tls.cert="$(cat ${CERTS_DIR}/helm.cert.pem)" \
    `# Image flags` \
    --set apprepository.image.tag=$DEV_TAG \
    --set apprepository.image.repository=kubeapps/apprepository-controller$IMG_MODIFIER \
    --set dashboard.image.tag=$DEV_TAG \
    --set dashboard.image.repository=kubeapps/dashboard$IMG_MODIFIER \
    --set tillerProxy.image.tag=$DEV_TAG \
    --set tillerProxy.image.repository=kubeapps/tiller-proxy$IMG_MODIFIER

# Ensure that we are testing the correct image
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-tiller-proxy $DEV_TAG

# Wait for Kubeapps Pods
deployments=(
  kubeapps-ci
  kubeapps-ci-internal-apprepository-controller
  kubeapps-ci-internal-chartsvc
  kubeapps-ci-internal-tiller-proxy
  kubeapps-ci-internal-dashboard
  kubeapps-ci-mongodb
)
for dep in ${deployments[@]}; do
  k8s_wait_for_deployment kubeapps ${dep}
  echo "Deployment ${dep} ready"
done

# Wait for Kubeapps Jobs
k8s_wait_for_job_completed kubeapps apprepositories.kubeapps.com/repo-name=stable
echo "Job apprepositories.kubeapps.com/repo-name=stable ready"

echo "All deployments ready. PODs:"
kubectl get pods -n kubeapps -o wide

# Run helm tests
set +e

helm test ${HELM_CLIENT_TLS_FLAGS} kubeapps-ci
code=$?

set -e

if [[ "$code" != 0 ]]; then
  echo "PODS status on failure"
  kubectl get pods -n kubeapps
  for pod in $(kubectl get po -l release=kubeapps-ci -oname -n kubeapps); do
    echo "LOGS for pod $pod ------------"
    kubectl logs -n kubeapps $pod
  done;
  echo 
  echo "LOGS for chartsvc tests --------"
  kubectl logs kubeapps-ci-chartsvc-test --namespace kubeapps
  echo "LOGS for tiller-proxy tests --------"
  kubectl logs kubeapps-ci-tiller-proxy-test --namespace kubeapps
  echo "LOGS for dashboard tests --------"
  kubectl logs kubeapps-ci-dashboard-test --namespace kubeapps
fi

exit $code
