#!/usr/bin/env bash

# Copyright (c) 2018-2020 Bitnami
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

set -o errexit
set -o pipefail

# Constants
ROOT_DIR="$(cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd)"
DEV_TAG=${1:?missing dev tag}
IMG_MODIFIER=${2:-""}
CERTS_DIR="${ROOT_DIR}/script/test-certs"
HELM_CLIENT_TLS_FLAGS=("--tls" "--tls-cert" "${CERTS_DIR}/helm.cert.pem" "--tls-key" "${CERTS_DIR}/helm.key.pem")

# Load Generic Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/liblog.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libutil.sh"

# Auxiliar functions

########################
# Test Helm
# Globals:
#   HELM_*
# Arguments: None
# Returns: None
#########################
testHelm() {
  info "Running Helm tests..."
  if [[ "$HELM_VERSION" =~ "v2" ]]; then
    helm test "${HELM_CLIENT_TLS_FLAGS[@]}" kubeapps-ci --cleanup
  else
    helm test -n kubeapps kubeapps-ci
  fi
}

########################
# Init Tiller with TLS support on clusters with RBAC enabled
# Globals: None
# Arguments: None
# Returns: None
#########################
tiller-init-rbac() {
    info "Installing Tiller..."
    kubectl create serviceaccount -n kube-system tiller
    kubectl create clusterrolebinding tiller-cluster-rule \
      --clusterrole=cluster-admin \
      --serviceaccount=kube-system:tiller
    # The flag --wait is not available when using TLS flags
    # ref: https://github.com/helm/helm/issues/4050
    helm init \
      --service-account tiller \
      --tiller-tls \
      --tiller-tls-cert "${CERTS_DIR}/tiller.cert.pem" \
      --tiller-tls-key "${CERTS_DIR}/tiller.key.pem" \
      --tiller-tls-verify \
      --tls-ca-cert "${CERTS_DIR}/ca.cert.pem"
    info "Waiting for Tiller to be ready ... "
    # Retries 60 times with 1 second interval
    retry_while "helm version ${HELM_CLIENT_TLS_FLAGS[*]} --tiller-connection-timeout 1" "60" "1"
}

########################
# Check if the pod that populates de OperatorHub catalog is running
# Globals: None
# Arguments: None
# Returns: None
#########################
isOperatorHubCatalogRunning() {
  kubectl get pod -n olm -l olm.catalogSource=operatorhubio-catalog -o jsonpath='{.items[0].status.phase}' | grep Running
  # Wait also for the catalog to be populated
  kubectl get packagemanifests.packages.operators.coreos.com | grep prometheus
}

########################
# Install OLM
# Globals: None
# Arguments:
#   $1: Version of OLM
# Returns: None
#########################
installOLM() {
    local release=$1
    info "Installing OLM ${release} ..."
    url=https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${release}
    namespace=olm

    kubectl apply -f ${url}/crds.yaml
    kubectl apply -f ${url}/olm.yaml

    # wait for deployments to be ready
    kubectl rollout status -w deployment/olm-operator --namespace="${namespace}"
    kubectl rollout status -w deployment/catalog-operator --namespace="${namespace}"
}

########################
# Install chartmuseum
# Globals: None
# Arguments:
#   $1: Username
#   $2: Password
# Returns: None
#########################
installChartmuseum() {
    local user=$1
    local password=$2
    info "Installing ChartMuseum ..."
    helm repo add stable https://kubernetes-charts.storage.googleapis.com
    helm repo up
    if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
      helm install --name chartmuseum --namespace kubeapps stable/chartmuseum \
        "${HELM_CLIENT_TLS_FLAGS[@]}" \
        --set env.open.DISABLE_API=false \
        --set persistence.enabled=true \
        --set secret.AUTH_USER=$user \
        --set secret.AUTH_PASS=$password
    else
      helm install chartmuseum --namespace kubeapps stable/chartmuseum \
        --set env.open.DISABLE_API=false \
        --set persistence.enabled=true \
        --set secret.AUTH_USER=$user \
        --set secret.AUTH_PASS=$password
    fi
    kubectl rollout status -w deployment/chartmuseum-chartmuseum --namespace=kubeapps
}

########################
# Push a chart to chartmusem
# Globals: None
# Arguments:
#   $1: chart
#   $2: version
#   $3: chartmuseum username
#   $4: chartmuseum password
# Returns: None
#########################
pushChart() {
    local chart=$1
    local version=$2
    local user=$3
    local password=$4
    info "Adding ${chart}-${version} to ChartMuseum ..."
    curl -LO "https://charts.bitnami.com/bitnami/${chart}-${version}.tgz"

    local POD_NAME=$(kubectl get pods --namespace kubeapps -l "app=chartmuseum" -l "release=chartmuseum" -o jsonpath="{.items[0].metadata.name}")
    /bin/sh -c "kubectl port-forward $POD_NAME 8080:8080 --namespace kubeapps &"
    sleep 2
    curl -u "${user}:${password}" --data-binary "@${chart}-${version}.tgz" http://localhost:8080/api/charts
    pkill -f "kubectl port-forward $POD_NAME 8080:8080 --namespace kubeapps"
}

########################
# Install Kubeapps or upgrades it if it's already installed
# Arguments:
#   $1: chart source
# Returns: None
#########################
installOrUpgradeKubeapps() {
    local chartSource=$1
    # Install Kubeapps
    info "Installing Kubeapps..."
    if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
      helm upgrade --install kubeapps-ci --namespace kubeapps "${chartSource}" \
        "${HELM_CLIENT_TLS_FLAGS[@]}" \
        --set tillerProxy.tls.key="$(cat "${CERTS_DIR}/helm.key.pem")" \
        --set tillerProxy.tls.cert="$(cat "${CERTS_DIR}/helm.cert.pem")" \
        --set featureFlags.operators=true \
        ${invalidateCacheFlag} \
        "${img_flags[@]}" \
        "${db_flags[@]}"
    else
      helm upgrade --install kubeapps-ci --namespace kubeapps "${chartSource}" \
        ${invalidateCacheFlag} \
        "${img_flags[@]}" \
        "${db_flags[@]}" \
        --set featureFlags.operators=true \
        --set useHelm3=true
    fi
}

# Operators are not supported in GKE 1.14 and flaky in 1.15
if [[ -z "${GKE_BRANCH-}" ]]; then
  installOLM 0.15.1
fi

info "IMAGE TAG TO BE TESTED: $DEV_TAG"
info "IMAGE_REPO_SUFFIX: $IMG_MODIFIER"
info "Cluster Version: $(kubectl version -o json | jq -r '.serverVersion.gitVersion')"
info "Kubectl Version: $(kubectl version -o json | jq -r '.clientVersion.gitVersion')"

db_flags=("--set" "mongodb.enabled=true" "--set" "postgresql.enabled=false")
[[ "${KUBEAPPS_DB:-}" == "postgresql" ]] && db_flags=("--set" "mongodb.enabled=false" "--set" "postgresql.enabled=true")

# Use dev images or Bitnami if testing the latest release
image_prefix="kubeapps/"
[[ -n "${TEST_LATEST_RELEASE:-}" ]] && image_prefix="bitnami/kubeapps-"
images=(
  "apprepository-controller"
  "asset-syncer"
  "assetsvc"
  "dashboard"
  "tiller-proxy"
  "kubeops"
)
images=("${images[@]/#/${image_prefix}}")
images=("${images[@]/%/${IMG_MODIFIER}}")
img_flags=(
  "--set" "apprepository.image.tag=${DEV_TAG}"
  "--set" "apprepository.image.repository=${images[0]}"
  "--set" "apprepository.syncImage.tag=${DEV_TAG}"
  "--set" "apprepository.syncImage.repository=${images[1]}"
  "--set" "assetsvc.image.tag=${DEV_TAG}"
  "--set" "assetsvc.image.repository=${images[2]}"
  "--set" "dashboard.image.tag=${DEV_TAG}"
  "--set" "dashboard.image.repository=${images[3]}"
  "--set" "tillerProxy.image.tag=${DEV_TAG}"
  "--set" "tillerProxy.image.repository=${images[4]}"
  "--set" "kubeops.image.tag=${DEV_TAG}"
  "--set" "kubeops.image.repository=${images[5]}"
)

# TODO(andresmgot): Remove this condition with the parameter in the next version
invalidateCacheFlag=""
if [[ -z "${TEST_LATEST_RELEASE:-}" ]]; then
  invalidateCacheFlag="--set featureFlags.invalidateCache=true"
fi

if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
  # Init Tiller
  tiller-init-rbac
fi
helm repo add bitnami https://charts.bitnami.com/bitnami
helm dep up "${ROOT_DIR}/chart/kubeapps"
kubectl create ns kubeapps

if [[ -n "${TEST_UPGRADE}" ]]; then
  # To test the upgrade, first install the latest version published
  info "Installing latest Kubeapps chart available"
  installOrUpgradeKubeapps bitnami/kubeapps
fi

installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps"
installChartmuseum admin password
pushChart apache 7.3.15 admin password
pushChart apache 7.3.16 admin password

# Ensure that we are testing the correct image
info ""
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller "$DEV_TAG"
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard "$DEV_TAG"
if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
  k8s_ensure_image kubeapps kubeapps-ci-internal-tiller-proxy "$DEV_TAG"
else
  k8s_ensure_image kubeapps kubeapps-ci-internal-kubeops "$DEV_TAG"
fi

# Wait for Kubeapps Pods
info "Waiting for Kubeapps components to be ready..."
deployments=(
  "kubeapps-ci"
  "kubeapps-ci-internal-apprepository-controller"
  "kubeapps-ci-internal-assetsvc"
  "kubeapps-ci-internal-dashboard"
)
for dep in "${deployments[@]}"; do
  k8s_wait_for_deployment kubeapps "$dep"
  info "Deployment ${dep} ready"
done
if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
  k8s_wait_for_deployment kubeapps kubeapps-ci-internal-tiller-proxy
else
  k8s_wait_for_deployment kubeapps kubeapps-ci-internal-kubeops
fi

# Wait for Kubeapps Jobs
# Clean up existing jobs
kubectl delete jobs -n kubeapps --all
# Trigger update of the bitnami repository
kubectl patch apprepositories.kubeapps.com -n kubeapps bitnami -p='[{"op": "replace", "path": "/spec/resyncRequests", "value":1}]' --type=json
k8s_wait_for_job_completed kubeapps apprepositories.kubeapps.com/repo-name=bitnami
info "Job apprepositories.kubeapps.com/repo-name=bitnami ready"

info "All deployments ready. PODs:"
kubectl get pods -n kubeapps -o wide

# Wait for all the endpoints to be ready
kubectl get ep --namespace=kubeapps
svcs=(
  "kubeapps-ci"
  "kubeapps-ci-internal-assetsvc"
  "kubeapps-ci-internal-dashboard"
)
for svc in "${svcs[@]}"; do
  k8s_wait_for_endpoints kubeapps "$svc" 2
  info "Endpoints for ${svc} available"
done

# Disable helm tests unless we are testing the latest release until
# we have released the code with per-namespace tests (since the helm
# tests for assetsvc needs to test the namespaced repo).
if [[ -z "${TEST_LATEST_RELEASE:-}" ]]; then
  # Run helm tests
  # Retry once if tests fail to avoid temporary issue
  if ! retry_while testHelm "2" "1"; then
    warn "PODS status on failure"
    kubectl get pods -n kubeapps
    for pod in $(kubectl get po -l release=kubeapps-ci -oname -n kubeapps); do
      warn "LOGS for pod $pod ------------"
      kubectl logs -n kubeapps "$pod"
    done;
    echo
    warn "LOGS for assetsvc tests --------"
    kubectl logs kubeapps-ci-assetsvc-test --namespace kubeapps
    warn "LOGS for tiller-proxy tests --------"
    kubectl logs kubeapps-ci-tiller-proxy-test --namespace kubeapps
    warn "LOGS for dashboard tests --------"
    kubectl logs kubeapps-ci-dashboard-test --namespace kubeapps
    exit 1
  fi
  info "Helm tests succeded!!"
fi

# Operators are not supported in GKE 1.14 and flaky in 1.15
if [[ -z "${GKE_BRANCH-}" ]]; then
  ## Wait for the Operator catalog to be populated
  info "Waiting for the OperatorHub Catalog to be ready ..."
  retry_while isOperatorHubCatalogRunning 24
fi

# Browser tests
cd "${ROOT_DIR}/integration"
kubectl apply -f manifests/executor.yaml
k8s_wait_for_deployment default integration
pod=$(kubectl get po -l run=integration -o jsonpath="{.items[0].metadata.name}")
## Copy config and latest tests
for f in *.js; do
  kubectl cp "./${f}" "${pod}:/app/"
done
testsToIgnore=()
# Operators are not supported in GKE 1.14 and flaky in 1.15, skipping test
if [[ -n "${GKE_BRANCH-}" ]]; then
  testsToIgnore=("operator-deployment.js" "${testsToIgnore[@]}")
fi
## Support for Docker registry secrets are not supported for Helm2, skipping that test
if [[ "${HELM_VERSION:-}" =~ "v2" ]]; then
  testsToIgnore=("create-private-registry.js" "${testsToIgnore[@]}")
fi
ignoreFlag=""
if [[ "${#testsToIgnore[@]}" > "0" ]]; then
  # Join tests to ignore
  testsToIgnore=$(printf "|%s" "${testsToIgnore[@]}")
  testsToIgnore=${testsToIgnore:1}
  ignoreFlag="--testPathIgnorePatterns '$testsToIgnore'"
fi
kubectl cp ./use-cases "${pod}:/app/"
## Create admin user
kubectl create serviceaccount kubeapps-operator -n kubeapps
kubectl create clusterrolebinding kubeapps-operator-admin --clusterrole=admin --serviceaccount kubeapps:kubeapps-operator
kubectl create clusterrolebinding kubeapps-repositories-write --clusterrole kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
## Create view user
kubectl create serviceaccount kubeapps-view -n kubeapps
kubectl create clusterrolebinding kubeapps-view --clusterrole=view --serviceaccount kubeapps:kubeapps-view
## Create edit user
kubectl create serviceaccount kubeapps-edit -n kubeapps
kubectl create rolebinding kubeapps-edit -n kubeapps --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
kubectl create rolebinding kubeapps-edit -n default --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
## Give the cluster some time to avoid issues like
## https://circleci.com/gh/kubeapps/kubeapps/16102
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-operator -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-view -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-edit -o name" "5" "1"
## Retrieve tokens
admin_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
view_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-view -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
edit_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-edit -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
## Run tests
info "Running Integration tests..."
if ! kubectl exec -it "$pod" -- /bin/sh -c "INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} EDIT_TOKEN=${edit_token} yarn start ${ignoreFlag}"; then
  ## Integration tests failed, get report screenshot
  warn "PODS status on failure"
  kubectl cp "${pod}:/app/reports" ./reports
  exit 1
fi
info "Integration tests succeded!!"
