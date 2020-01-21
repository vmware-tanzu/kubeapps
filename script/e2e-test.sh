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

function testHelm {
  if [[ "$HELM_VERSION" =~ "v2" ]]; then
    helm test ${HELM_CLIENT_TLS_FLAGS} kubeapps-ci --cleanup
  else
    helm test -n kubeapps kubeapps-ci
  fi
}

# Print cluster version
kubectl version

dbFlags="--set mongodb.enabled=true --set postgresql.enabled=false"
if [[ "${KUBEAPPS_DB}" == "postgresql" ]]; then
  dbFlags="--set mongodb.enabled=false --set postgresql.enabled=true"
fi

# Use dev images or bitnami if testing the latest release
apprepositoryControllerImage="kubeapps/apprepository-controller"
assetSyncerImage="kubeapps/asset-syncer"
assetsvcImage="kubeapps/assetsvc"
dashboardImage="kubeapps/dashboard"
tillerProxyImage="kubeapps/tiller-proxy"
kubeopsImage="kubeapps/kubeops"
if [[ -n "$TEST_LATEST_RELEASE" ]]; then
  apprepositoryControllerImage="bitnami/kubeapps-apprepository-controller"
  assetSyncerImage="bitnami/kubeapps-asset-syncer"
  assetsvcImage="bitnami/kubeapps-assetsvc"
  dashboardImage="bitnami/kubeapps-dashboard"
  tillerProxyImage="bitnami/kubeapps-tiller-proxy"
  kubeopsImage="bitnami/kubeapps-kubeops"
fi
imgFlags=(
  --set apprepository.image.tag=${DEV_TAG}
  --set apprepository.image.repository=${apprepositoryControllerImage}${IMG_MODIFIER}
  --set apprepository.syncImage.tag=${DEV_TAG}
  --set apprepository.syncImage.repository=${assetSyncerImage}$IMG_MODIFIER
  --set assetsvc.image.tag=${DEV_TAG}
  --set assetsvc.image.repository=${assetsvcImage}${IMG_MODIFIER}
  --set dashboard.image.tag=${DEV_TAG}
  --set dashboard.image.repository=${dashboardImage}${IMG_MODIFIER}
  --set tillerProxy.image.tag=${DEV_TAG}
  --set tillerProxy.image.repository=${tillerProxyImage}${IMG_MODIFIER}
  --set kubeops.image.tag=${DEV_TAG}
  --set kubeops.image.repository=${kubeopsImage}${IMG_MODIFIER}
)

if [[ "$HELM_VERSION" =~ "v2" ]]; then
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

  # Install Kubeapps
  helm dep up $ROOT_DIR/chart/kubeapps/
  helm install --name kubeapps-ci --namespace kubeapps $ROOT_DIR/chart/kubeapps \
    `# Tiller TLS flags` \
    ${HELM_CLIENT_TLS_FLAGS} \
    `# Tiller-proxy TLS flags` \
    --set tillerProxy.tls.key="$(cat ${CERTS_DIR}/helm.key.pem)" \
    --set tillerProxy.tls.cert="$(cat ${CERTS_DIR}/helm.cert.pem)" \
    `# Image flags` \
    "${imgFlags[@]}" \
    `# Database choice flags` \
    ${dbFlags}
else
  # Install Kubeapps
  kubectl create ns kubeapps
  helm dep up $ROOT_DIR/chart/kubeapps/
  helm install kubeapps-ci --namespace kubeapps $ROOT_DIR/chart/kubeapps \
    `# Image flags` \
    "${imgFlags[@]}" \
    `# Database choice flags` \
    ${dbFlags} \
    `# Enable Helm 3 flag` \
    --set useHelm3=true
fi

# Ensure that we are testing the correct image
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller $DEV_TAG
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard $DEV_TAG
if [[ "$HELM_VERSION" =~ "v2" ]]; then
  k8s_ensure_image kubeapps kubeapps-ci-internal-tiller-proxy $DEV_TAG
else
  k8s_ensure_image kubeapps kubeapps-ci-internal-kubeops $DEV_TAG
fi

# Wait for Kubeapps Pods
deployments=(
  kubeapps-ci
  kubeapps-ci-internal-apprepository-controller
  kubeapps-ci-internal-assetsvc
  kubeapps-ci-internal-dashboard
)
for dep in ${deployments[@]}; do
  k8s_wait_for_deployment kubeapps ${dep}
  echo "Deployment ${dep} ready"
done
if [[ "$HELM_VERSION" =~ "v2" ]]; then
  k8s_wait_for_deployment kubeapps kubeapps-ci-internal-tiller-proxy
else
  k8s_wait_for_deployment kubeapps kubeapps-ci-internal-kubeops
fi


# Wait for Kubeapps Jobs
k8s_wait_for_job_completed kubeapps apprepositories.kubeapps.com/repo-name=stable
echo "Job apprepositories.kubeapps.com/repo-name=stable ready"

echo "All deployments ready. PODs:"
kubectl get pods -n kubeapps -o wide

# Wait for all the endpoints to be ready
kubectl get ep --namespace=kubeapps
svcs=(
  kubeapps-ci
  kubeapps-ci-internal-assetsvc
  kubeapps-ci-internal-dashboard
)
for svc in ${svcs[@]}; do
  k8s_wait_for_endpoint kubeapps ${svc} 2
  echo "Endpoints for ${svc} available"
done

# Run helm tests
set +e

testHelm
code=$?

if [[ "$code" != 0 ]]; then
  echo "Helm test failed, retrying..."
  # Avoid temporary issues, retry
  testHelm
  code=$?
fi

set -e

if [[ "$code" != 0 ]]; then
  echo "PODS status on failure"
  kubectl get pods -n kubeapps
  for pod in $(kubectl get po -l release=kubeapps-ci -oname -n kubeapps); do
    echo "LOGS for pod $pod ------------"
    kubectl logs -n kubeapps $pod
  done;
  echo 
  echo "LOGS for assetsvc tests --------"
  kubectl logs kubeapps-ci-assetsvc-test --namespace kubeapps
  echo "LOGS for tiller-proxy tests --------"
  kubectl logs kubeapps-ci-tiller-proxy-test --namespace kubeapps
  echo "LOGS for dashboard tests --------"
  kubectl logs kubeapps-ci-dashboard-test --namespace kubeapps

  exit $code
fi

# Browser tests
cd $ROOT_DIR/integration
kubectl apply -f manifests/executor.yaml
k8s_wait_for_deployment default integration
pod=$(kubectl get po -l run=integration -o jsonpath="{.items[0].metadata.name}")
## Copy config and latest tests
for f in `ls *.js`; do kubectl cp ./${f} ${pod}:/app/; done
kubectl cp ./use-cases ${pod}:/app/
## Create admin user
kubectl create serviceaccount kubeapps-operator -n kubeapps
kubectl create clusterrolebinding kubeapps-operator-admin --clusterrole=admin --serviceaccount kubeapps:kubeapps-operator
kubectl create -n kubeapps rolebinding kubeapps-repositories-write --role=kubeapps-ci-repositories-write --serviceaccount kubeapps:kubeapps-operator
admin_token=`kubectl get -n kubeapps secret $(kubectl get -n kubeapps serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}') -o go-template='{{.data.token | base64decode}}' && echo`
## Create view user
kubectl create serviceaccount kubeapps-view -n kubeapps
kubectl create clusterrolebinding kubeapps-view --clusterrole=view --serviceaccount kubeapps:kubeapps-view
view_token=`kubectl get -n kubeapps secret $(kubectl get -n kubeapps serviceaccount kubeapps-view -o jsonpath='{.secrets[].name}') -o go-template='{{.data.token | base64decode}}' && echo`
## Run tests
set +e
kubectl exec -it ${pod} -- /bin/sh -c "INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} yarn start"
code=$?
set -e
if [[ "$code" != 0 ]]; then
  ### Browser tests failed, get report screenshot
  echo "PODS status on failure"
  kubectl cp ${pod}:/app/reports ./reports
fi

exit $code
