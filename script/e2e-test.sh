#!/usr/bin/env bash

# Copyright 2018-2024 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

startTime=$(date +%s)

# Constants
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
ALL_TESTS="all"
MAIN_TESTS="main"
MAIN_TESTS_SUBGROUP="main-group-"
EXISTENT_MAIN_TESTS_SUBGROUPS=3
MULTICLUSTER_TESTS="multicluster"
MULTICLUSTER_NOKUBEAPPS_TESTS="multicluster-nokubeapps"
CARVEL_TESTS="carvel"
FLUX_TESTS="flux"
OPERATOR_TESTS="operators"
SUPPORTED_TESTS_GROUPS=("${ALL_TESTS}" "${MAIN_TESTS}" "${MULTICLUSTER_TESTS}" "${CARVEL_TESTS}" "${FLUX_TESTS}" "${OPERATOR_TESTS}" "${MULTICLUSTER_NOKUBEAPPS_TESTS}")
INTEGRATION_HOST=kubeapps-ci.kubeapps
INTEGRATION_ENTRYPOINT="http://${INTEGRATION_HOST}"

# Params
USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV:-"false"}
OLM_VERSION=${OLM_VERSION:-"v0.18.2"}
IMG_DEV_TAG=${IMG_DEV_TAG:?missing dev tag}
IMG_MODIFIER=${IMG_MODIFIER:-""}
TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES:-"4"}
DEX_IP=${DEX_IP:-"172.18.0.2"}
ADDITIONAL_CLUSTER_IP=${ADDITIONAL_CLUSTER_IP:-"172.18.0.3"}
KAPP_CONTROLLER_VERSION=${KAPP_CONTROLLER_VERSION:-"v0.42.0"}
CHARTMUSEUM_VERSION=${CHARTMUSEUM_VERSION:-"3.9.1"}
FLUX_VERSION=${FLUX_VERSION:-"v2.2.3"}
GKE_VERSION=${GKE_VERSION:-}
IMG_PREFIX=${IMG_PREFIX:-"kubeapps/"}
TESTS_GROUP=${TESTS_GROUP:-"${ALL_TESTS}"}
DEBUG_MODE=${DEBUG_MODE:-false}
TEST_LATEST_RELEASE=${TEST_LATEST_RELEASE:-false}

# shellcheck disable=SC2076
if [[ ! " ${SUPPORTED_TESTS_GROUPS[*]} " =~ " ${TESTS_GROUP} " && ! "${TESTS_GROUP}" == "${MAIN_TESTS_SUBGROUP}"* ]]; then
  # shellcheck disable=SC2046
  echo $(IFS=','; echo "The provided TESTS_GROUP [${TESTS_GROUP}] is not supported. Supported groups are: ${SUPPORTED_TESTS_GROUPS[*]}. Subgroups of the main group are also supported: ${MAIN_TESTS_SUBGROUP}*")
  exit 1
fi

# TODO(andresmgot): While we work with beta releases, the Bitnami pipeline
# removes the pre-release part of the tag
if [[ -n "${TEST_LATEST_RELEASE}" && "${TEST_LATEST_RELEASE}" != "false" ]]; then
  IMG_DEV_TAG=${IMG_DEV_TAG/-beta.*/}
fi

# Load Generic Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/liblog.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/libutil.sh"

# Get the load balancer IP
if [[ -z "${GKE_VERSION-}" ]]; then
  LOAD_BALANCER_IP=$DEX_IP
else
  LOAD_BALANCER_IP=$(kubectl -n nginx-ingress get service nginx-ingress-ingress-nginx-controller -o jsonpath="{.status.loadBalancer.ingress[].ip}")
fi

# Functions for local Docker registry mgmt
. "${ROOT_DIR}/script/local-docker-registry.sh"

# Functions for handling Chart Museum
. "${ROOT_DIR}/script/chart-museum.sh"

info "###############################################################################################"
info "DEBUG_MODE: ${DEBUG_MODE}"
info "TESTS_GROUP: ${TESTS_GROUP}"
info "GKE_VERSION: ${GKE_VERSION}"
info "ROOT_DIR: ${ROOT_DIR}"
info "USE_MULTICLUSTER_OIDC_ENV: ${USE_MULTICLUSTER_OIDC_ENV}"
info "OLM_VERSION: ${OLM_VERSION}"
info "CHARTMUSEUM_VERSION: ${CHARTMUSEUM_VERSION}"
info "IMG_DEV_TAG: ${IMG_DEV_TAG}"
info "IMG_MODIFIER: ${IMG_MODIFIER}"
info "IMG_PREFIX: ${IMG_PREFIX}"
info "DEX_IP: ${DEX_IP}"
info "ADDITIONAL_CLUSTER_IP: ${ADDITIONAL_CLUSTER_IP}"
info "LOAD_BALANCER_IP: ${LOAD_BALANCER_IP}"
info "TEST_TIMEOUT_MINUTES: ${TEST_TIMEOUT_MINUTES}"
info "KAPP_CONTROLLER_VERSION: ${KAPP_CONTROLLER_VERSION}"
info "K8S SERVER VERSION: $(kubectl version -o json | jq -r '.serverVersion.gitVersion')"
info "KUBECTL VERSION: $(kubectl version -o json | jq -r '.clientVersion.gitVersion')"
info "###############################################################################################"

# Auxiliary functions

#
# Install an authenticated Docker registry inside the cluster
#
setupLocalDockerRegistry() {
    info "Installing local Docker registry with authentication"
    installLocalRegistry "${ROOT_DIR}"

    info "Pushing test container to local Docker registry"
    pushContainerToLocalRegistry
}

#
# Push a chart that uses container image from the local registry
#
pushLocalChart() {
    info "Packaging local test chart"
    helm package "${ROOT_DIR}/integration/charts/simplechart"

    info "Pushing local test chart to ChartMuseum"
    pushChartToChartMuseum "simplechart" "0.1.0" "simplechart-0.1.0.tgz"
}

########################################################################################################################
# Check if the pod that populates de OperatorHub catalog is running
# Globals: None
# Arguments: None
# Returns: None
########################################################################################################################
isOperatorHubCatalogRunning() {
  kubectl get pod -n olm -l olm.catalogSource=operatorhubio-catalog -o jsonpath='{.items[0].status.phase}' | grep Running
  # Wait also for the catalog to be populated
  kubectl get packagemanifests.packages.operators.coreos.com | grep prometheus
}

########################################################################################################################
# Install OLM
# Globals: None
# Arguments:
#   $1: Version of OLM
# Returns: None
########################################################################################################################
installOLM() {
  local release=$1
  info "Installing OLM ${release} ..."
  url="https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${release}"
  namespace=olm

  kubectl create -f "${url}/crds.yaml"
  kubectl wait --for=condition=Established -f "${url}/crds.yaml"
  kubectl create -f "${url}/olm.yaml"

  # wait for deployments to be ready
  kubectl rollout status -w deployment/olm-operator --namespace="${namespace}"
  kubectl rollout status -w deployment/catalog-operator --namespace="${namespace}"

  retries=30
  until [[ $retries == 0 ]]; do
    new_csv_phase=$(kubectl get csv -n "${namespace}" packageserver -o jsonpath='{.status.phase}' 2>/dev/null || echo "Waiting for CSV to appear")
    if [[ $new_csv_phase != "${csv_phase:-}" ]]; then
      csv_phase=$new_csv_phase
      echo "CSV \"packageserver\" phase: ${csv_phase}"
    fi
    if [[ "$new_csv_phase" == "Succeeded" ]]; then
      break
    fi
    sleep 10
    retries=$((retries - 1))
  done

  if [ $retries == 0 ]; then
    echo "CSV \"packageserver\" failed to reach phase succeeded"
    exit 1
  fi

  kubectl rollout status -w deployment/packageserver --namespace="${namespace}"
}

########################################################################################################################
# Push a chart to chartmusem
# Globals: None
# Arguments:
#   $1: chart
#   $2: version
# Returns: None
########################################################################################################################
pushChart() {
  local chart=$1
  local version=$2
  prefix="kubeapps-"
  description="foo ${chart} chart for CI"

  info "Adding ${chart}-${version} to ChartMuseum ..."
  pullBitnamiChart "${chart}" "${version}"

  # Mutate the chart name and description, then re-package the tarball
  # For instance, the apache's Chart.yaml file becomes modified to:
  #   name: kubeapps-apache
  #   description: foo apache chart for CI
  # consequently, the new packaged chart is "${prefix}${chart}-${version}.tgz"
  # This workaround should mitigate https://github.com/vmware-tanzu/kubeapps/issues/3339
  mkdir "./${chart}-${version}"
  tar zxf "${chart}-${version}.tgz" -C "./${chart}-${version}"
  # this relies on GNU sed, which is not the default on MacOS
  # ref https://gist.github.com/andre3k1/e3a1a7133fded5de5a9ee99c87c6fa0d
  sed -i "s/name: ${chart}/name: ${prefix}${chart}/" "./${chart}-${version}/${chart}/Chart.yaml"
  sed -i "0,/^\([[:space:]]*description: *\).*/s//\1${description}/" "./${chart}-${version}/${chart}/Chart.yaml"
  helm package "./${chart}-${version}/${chart}" -d .

  pushChartToChartMuseum "${chart}" "${version}" "${prefix}${chart}-${version}.tgz"
}

########################################################################################################################
# Install kapp-controller
# Globals: None
# Arguments:
#   $1: Version of kapp-controller
# Returns: None
########################################################################################################################
installKappController() {
  local release=$1
  info "Installing kapp-controller ${release} ..."
  url="https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/${release}/release.yml"
  namespace=kapp-controller

  kubectl apply -f "${url}"

  # wait for deployment to be ready
  kubectl rollout status -w deployment/kapp-controller --namespace="${namespace}"

  # Add test repository.
	kubectl apply -f https://raw.githubusercontent.com/vmware-tanzu/carvel-kapp-controller/develop/examples/packaging-with-repo/package-repository.yml

  # Add a carvel-reconciler service account to the kubeapps-user-namespace with
  # cluster-admin.
  kubectl create serviceaccount carvel-reconciler -n kubeapps-user-namespace
  kubectl create clusterrolebinding carvel-reconciler --clusterrole=cluster-admin --serviceaccount kubeapps-user-namespace:carvel-reconciler
}

########################################################################################################################
# Install flux
# Globals: None
# Arguments:
#   $1: Version of flux
# Returns: None
########################################################################################################################
installFlux() {
  local release=$1
  info "Installing flux ${release} ..."
  url="https://github.com/fluxcd/flux2/releases/download/${release}/install.yaml"
  namespace=flux-system

  # this is a workaround for flux e2e tests failing when run by GitHub Action Runners
  # due to not being able to deploy source-controller pod error:
  # Warning  FailedScheduling  19s (x7 over 6m)  default-scheduler  0/1 nodes are available: 1 Insufficient cpu.
  curl -o /tmp/flux_install.yaml -LO "${url}"
  cat /tmp/flux_install.yaml | sed -e 's/cpu: 100m/cpu: 75m/g' | kubectl apply -f -
  kubectl --namespace ${namespace} scale --replicas=0 deployment/image-automation-controller deployment/image-reflector-controller deployment/kustomize-controller

  # wait for deployments to be ready
  k8s_wait_for_deployment ${namespace} helm-controller
  k8s_wait_for_deployment ${namespace} source-controller

  # Add test repository.
  info "Install flux helm repository"
  #kubectl apply -f https://raw.githubusercontent.com/fluxcd/source-controller/main/config/samples/source_v1_helmrepository.yaml
  kubectl apply -f "${ROOT_DIR}/script/assets/flux-sample-helm-repository.yaml"

  # Add a flux-reconciler service account to the kubeapps-user-namespace with
  # cluster-admin.
  kubectl create serviceaccount flux-reconciler -n kubeapps-user-namespace
  kubectl create clusterrolebinding flux-reconciler --clusterrole=cluster-admin --serviceaccount kubeapps-user-namespace:flux-reconciler
}

########################################################################################################################
# Creates a Yaml file with additional values for the Helm chart
# Arguments: None
# Returns: Path to the newly created file with additional values
########################################################################################################################
generateAdditionalValuesFile() {
  # Could be done better with $(cat <<EOF > ${ROOT_DIR}/additional_chart_values.yaml
  # But it was breaking the formatting of the file
  local valuesFile="${ROOT_DIR}/additional_chart_values.yaml"
  echo "ingress:
  enabled: true
  hostname: localhost
  tls: true
  selfSigned: true
  ingressClassName: nginx
  annotations:
    nginx.ingress.kubernetes.io/proxy-buffer-size: \"8k\"
    nginx.ingress.kubernetes.io/proxy-buffers: \"4.0\"
    nginx.ingress.kubernetes.io/proxy-read-timeout: \"600.0\"" > "${valuesFile}"
  echo "${valuesFile}"
}

########################################################################################################################
# Install Kubeapps or upgrades it if it's already installed
# Arguments:
#   $1: chart source
# Returns: None
########################################################################################################################
installOrUpgradeKubeapps() {
  local chartSource=$1
  # Install Kubeapps
  info "Installing Kubeapps from ${chartSource}..."
  kubectl -n kubeapps delete secret localhost-tls || true

  # See https://stackoverflow.com/a/36296000 for "${arr[@]+"${arr[@]}"}" notation.
  cmd=(helm upgrade --install kubeapps-ci --namespace kubeapps "${chartSource}"
    "${img_flags[@]}"
    "${multiclusterFlags[@]+"${multiclusterFlags[@]}"}"
    "${@:2}"
    --set frontend.replicaCount=1
    --set dashboard.replicaCount=1
    --set kubeappsapis.replicaCount=2
    --set postgresql.architecture=standalone
    --set postgresql.primary.persistence.enabled=false
    --set postgresql.auth.password=password
    --set redis.auth.password=password
    --set apprepository.initialRepos[0].name=bitnami
    --set apprepository.initialRepos[0].url=http://chartmuseum.chart-museum.svc.cluster.local:8080
    --set apprepository.initialRepos[0].basicAuth.user=admin
    --set apprepository.initialRepos[0].basicAuth.password=password
    --set apprepository.globalReposNamespaceSuffix=-repos-global
    --set global.security.allowInsecureImages=true
    --wait)

  echo "${cmd[@]}"
  "${cmd[@]}"
}

########################################################################################################################
# Formats the provided time in seconds.
# Arguments:
#   $1: time in seconds
# Returns: Time formatted as Xm Ys
########################################################################################################################
formattedElapsedTime() {
  local time=$1

  mins=$((time/60))
  secs=$((time%60))
  echo "${mins}m ${secs}s"
}

########################################################################################################################
# Returns the elapsed time since the given starting point.
# Arguments:
#   $1: Starting point in seconds (eg. `date +%s`)
# Returns: The elapsed time formatted as Xm Ys
########################################################################################################################
elapsedTimeSince() {
  local start=${1?:Start time not provided}
  local end

  end=$(date +%s)
  formattedElapsedTime $((end-start))
}

[[ "${DEBUG_MODE}" == "true" ]] && set -x;

if [[ "${DEBUG_MODE}" == "true" && -z ${GKE_VERSION} ]]; then
  info "Docker images loaded in the cluster:"
  docker exec kubeapps-ci-control-plane crictl images
fi

# Use dev images or Bitnami if testing the latest release
kubeapps_apis_image="kubeapps-apis"
[[ -n "${TEST_LATEST_RELEASE}" && "${TEST_LATEST_RELEASE}" != "false" ]] && IMG_PREFIX="bitnami/kubeapps-" && kubeapps_apis_image="apis"
images=(
  "apprepository-controller"
  "asset-syncer"
  "dashboard"
  "pinniped-proxy"
  "${kubeapps_apis_image}"
  "oci-catalog"
)
images=("${images[@]/#/${IMG_PREFIX}}")
images=("${images[@]/%/${IMG_MODIFIER}}")
img_flags=(
  "--set" "apprepository.image.tag=${IMG_DEV_TAG}"
  "--set" "apprepository.image.repository=${images[0]}"
  "--set" "apprepository.syncImage.tag=${IMG_DEV_TAG}"
  "--set" "apprepository.syncImage.repository=${images[1]}"
  "--set" "dashboard.image.tag=${IMG_DEV_TAG}"
  "--set" "dashboard.image.repository=${images[2]}"
  "--set" "pinnipedProxy.image.tag=${IMG_DEV_TAG}"
  "--set" "pinnipedProxy.image.repository=${images[3]}"
  "--set" "kubeappsapis.image.tag=${IMG_DEV_TAG}"
  "--set" "kubeappsapis.image.repository=${images[4]}"
  "--set" "ociCatalog.image.tag=${IMG_DEV_TAG}"
  "--set" "ociCatalog.image.repository=${images[5]}"
)

additional_flags_file=$(generateAdditionalValuesFile)

if [ "$USE_MULTICLUSTER_OIDC_ENV" = true ]; then
  basicAuthFlags=(
    "--values" "${additional_flags_file}"
    "--set" "authProxy.enabled=true"
    "--set" "authProxy.provider=oidc"
    "--set" "authProxy.clientID=default"
    "--set" "authProxy.clientSecret=ZXhhbXBsZS1hcHAtc2VjcmV0"
    "--set" "authProxy.cookieSecret=bm90LWdvb2Qtc2VjcmV0Cg=="
    "--set" "authProxy.extraFlags[0]=\"--oidc-issuer-url=https://${DEX_IP}:32000\""
    "--set" "authProxy.extraFlags[1]=\"--scope=openid email groups audience:server:client_id:second-cluster audience:server:client_id:third-cluster\""
    "--set" "authProxy.extraFlags[2]=\"--ssl-insecure-skip-verify=true\""
    "--set" "authProxy.extraFlags[3]=\"--redirect-url=${INTEGRATION_ENTRYPOINT}/oauth2/callback\""
    "--set" "authProxy.extraFlags[4]=\"--cookie-secure=false\""
    "--set" "authProxy.extraFlags[5]=\"--cookie-domain=${INTEGRATION_HOST}\""
    "--set" "authProxy.extraFlags[6]=\"--whitelist-domain=${INTEGRATION_HOST}\""
    "--set" "authProxy.extraFlags[7]=\"--set-authorization-header=true\""
  )
  multiclusterFlags=(
    "--set" "clusters[0].name=default"
    "--set" "clusters[1].name=second-cluster"
    "--set" "clusters[1].apiServiceURL=https://${ADDITIONAL_CLUSTER_IP}:6443"
    "--set" "clusters[1].insecure=true"
    "--set" "clusters[1].serviceToken=$(kubectl --context=kind-kubeapps-ci-additional --kubeconfig="${HOME}/.kube/kind-config-kubeapps-ci-additional" get secret kubeapps-namespace-discovery -o go-template='{{.data.token | base64decode}}')"
  )
  multiclusterFlags+=("${basicAuthFlags[@]+"${basicAuthFlags[@]}"}")
fi

helm repo add bitnami https://charts.bitnami.com/bitnami
helm dep up "${ROOT_DIR}/chart/kubeapps"
kubectl create ns kubeapps
GLOBAL_REPOS_NS=kubeapps-repos-global

if [[ -n "${TEST_UPGRADE:-}" ]]; then
  # To test the upgrade, first install the latest version published
  info "Installing latest Kubeapps chart available"
  installOrUpgradeKubeapps bitnami/kubeapps \
    "--set" "apprepository.initialRepos={}"

  info "Waiting for Kubeapps components to be ready (bitnami chart)..."
  k8s_wait_for_deployment kubeapps kubeapps-ci
fi

# Install ChartMuseum
installChartMuseum "${CHARTMUSEUM_VERSION}"
pushChart apache 8.6.2
pushChart apache 8.6.3

# Install Kubeapps
installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps"
info "Waiting for Kubeapps components to be ready (local chart)..."
k8s_wait_for_deployment kubeapps kubeapps-ci

# Setting up local Docker registry if not in GKE
if [[ -z "${GKE_VERSION-}" ]]; then
  setupLocalDockerRegistry
  pushLocalChart
fi

# Ensure that we are testing the correct image
info ""
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller "$IMG_DEV_TAG"
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard "$IMG_DEV_TAG"
k8s_ensure_image kubeapps kubeapps-ci-internal-kubeappsapis "$IMG_DEV_TAG"

# Wait for Kubeapps Pods
info "Waiting for Kubeapps components to be ready..."
deployments=(
  "kubeapps-ci"
  "kubeapps-ci-internal-apprepository-controller"
  "kubeapps-ci-internal-dashboard"
  "kubeapps-ci-internal-kubeappsapis"
)
for dep in "${deployments[@]}"; do
  k8s_wait_for_deployment kubeapps "$dep"
  info "Deployment ${dep} ready"
done

# Wait for Kubeapps Jobs
# Clean up existing jobs
kubectl delete jobs -n kubeapps --all
# Trigger update of the bitnami repository
kubectl patch apprepositories.kubeapps.com -n ${GLOBAL_REPOS_NS} bitnami -p='[{"op": "replace", "path": "/spec/resyncRequests", "value":1}]' --type=json
k8s_wait_for_job_completed kubeapps apprepositories.kubeapps.com/repo-name=bitnami
info "Job apprepositories.kubeapps.com/repo-name=bitnami ready"

info "All deployments ready. PODs:"
kubectl get pods -n kubeapps -o wide

# Wait for all the endpoints to be ready
kubectl get ep --namespace=kubeapps
svcs=(
  "kubeapps-ci"
  "kubeapps-ci-internal-dashboard"
  "kubeapps-ci-internal-kubeappsapis"
)
for svc in "${svcs[@]}"; do
  k8s_wait_for_endpoints kubeapps "$svc" 1
  info "Endpoints for ${svc} available"
done

# Browser tests
cd "${ROOT_DIR}/integration"
info "Using E2E runner image '${IMG_PREFIX}integration-tests${IMG_MODIFIER}:${IMG_DEV_TAG}'"
kubectl create deployment e2e-runner --image "${IMG_PREFIX}integration-tests${IMG_MODIFIER}:${IMG_DEV_TAG}"
k8s_wait_for_deployment default e2e-runner
pod=$(kubectl get po -l app=e2e-runner -o custom-columns=:metadata.name --no-headers)
## Copy config and latest tests
for f in *.js; do
  kubectl cp "./${f}" "default/${pod}:/app/"
done

kubectl cp ./tests "default/${pod}:/app/"
info "Copied tests to e2e-runner pod default/${pod}"

## Create admin user with manual token
kubectl create serviceaccount kubeapps-operator -n kubeapps
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kubeapps-operator-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: kubeapps-operator
type: kubernetes.io/service-account-token
EOF
kubectl create clusterrolebinding kubeapps-operator-admin --clusterrole=cluster-admin --serviceaccount kubeapps:kubeapps-operator
kubectl create clusterrolebinding kubeapps-repositories-write --clusterrole kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
kubectl create rolebinding kubeapps-sa-operator-apprepositories-write -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
## Create view user
kubectl create serviceaccount kubeapps-view -n kubeapps
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kubeapps-view-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: kubeapps-view
type: kubernetes.io/service-account-token
EOF
kubectl create role view-secrets --verb=get,list,watch --resource=secrets
kubectl create rolebinding kubeapps-view-secret --role view-secrets --serviceaccount kubeapps:kubeapps-view
kubectl create clusterrolebinding kubeapps-view --clusterrole=view --serviceaccount kubeapps:kubeapps-view
kubectl create rolebinding kubeapps-view-user-apprepo-read -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-read --serviceaccount kubeapps:kubeapps-view
kubectl create rolebinding kubeapps-view-user -n kubeapps-user-namespace --clusterrole=edit --serviceaccount kubeapps:kubeapps-view
## Create edit user
kubectl create serviceaccount kubeapps-edit -n kubeapps
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: kubeapps-edit-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: kubeapps-edit
type: kubernetes.io/service-account-token
EOF
# TODO(minelson): Many of these roles/bindings need to be cleaned up. Some are
# unnecessary (with chart changes), some should not be created (such as edit
# here having the edit cluster role in the kubeapps namespace - should just be
# default). See https://github.com/vmware-tanzu/kubeapps/issues/4435
kubectl create rolebinding kubeapps-edit -n kubeapps --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
kubectl create rolebinding kubeapps-edit -n default --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
kubectl create clusterrolebinding kubeapps-repositories-read --clusterrole kubeapps:kubeapps:apprepositories-read --serviceaccount kubeapps:kubeapps-edit
# TODO(minelson): Similar to the `global-repos-read` rolebinding that the chart
# adds to the `kubeapps-repos-global` namespace for all authenticated users, we
# should eventually consider adding a similar rolebinding for secrets in the
# `kubeapps-repos-global` namespace also (but not if the global repos namespace
# is configured to be the kubeapps namespace, of course.) For now, explicit
# creation because CI tests with a repo with creds in the global repos ns.
# See https://github.com/vmware-tanzu/kubeapps/issues/4435
kubectl create role view-secrets -n ${GLOBAL_REPOS_NS} --verb=get,list,watch --resource=secrets
kubectl create rolebinding global-repos-secrets-read -n ${GLOBAL_REPOS_NS} --role=view-secrets --serviceaccount kubeapps:kubeapps-edit

## Give the cluster some time to avoid timeout issues
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-operator -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-view -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-edit -o name" "5" "1"
## Retrieve tokens
admin_token="$(kubectl get -n kubeapps secret kubeapps-operator-token -o go-template='{{.data.token | base64decode}}')"
view_token="$(kubectl get -n kubeapps secret kubeapps-view-token -o go-template='{{.data.token | base64decode}}')"
edit_token="$(kubectl get -n kubeapps secret kubeapps-edit-token -o go-template='{{.data.token | base64decode}}')"

info "Bootstrap time: $(elapsedTimeSince "$startTime")"

########################################################################################################################
# Returns the test command to be executed to run the e2e tests in the e2e runner image.
# Arguments:
#   $1: Test group to be executed (eg. main)
#   $2: Timeout in minutes. Optional, default: 20
# Returns: the command to be executed (a yarn command at the time of this writing).
########################################################################################################################
getTestCommand() {
  local tests_group=${1:?Missing test group}
  local timeout=${2:-20}
  echo "
    CI_TIMEOUT_MINUTES=${timeout} \
    DOCKER_USERNAME=${DOCKER_USERNAME} \
    DOCKER_PASSWORD=${DOCKER_PASSWORD} \
    DOCKER_REGISTRY_URL=${DOCKER_REGISTRY_URL} \
    TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} \
    INTEGRATION_ENTRYPOINT=${INTEGRATION_ENTRYPOINT} \
    USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} \
    ADMIN_TOKEN=${admin_token} \
    VIEW_TOKEN=${view_token} \
    EDIT_TOKEN=${edit_token} \
    yarn test \"tests/${tests_group}/\"
    "
}

########################################################################################################################
# Run a subgroup of the Main tests group.
# Arguments:
#   $1: Subgroup to run (eg. main-group-1)
# Returns: None
########################################################################################################################
runMainTestsSubgroup() {
  local subgroup=${1:?Missing main tests subgroup to run}
  local test_command=$(getTestCommand "${subgroup}" "20")

  sectionStartTime=$(date +%s)
  info "Running Main Integration tests subgroup [${subgroup}] without k8s API access..."
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Main integration tests subgroup [${subgroup}] succeeded!!"
  info "Execution time: $(elapsedTimeSince "$sectionStartTime")"
}

######################################
######## Main tests SUBGROUPS ########
######################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${MAIN_TESTS_SUBGROUP}"* ]]; then
  if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" ]]; then
    # Run all subgroups
    for group in $(seq 1 "${EXISTENT_MAIN_TESTS_SUBGROUPS}"); do
      subgroup="${MAIN_TESTS_SUBGROUP}${group}"
      runMainTestsSubgroup "${subgroup}"
    done
  else
    # Run a specific subgroup
    runMainTestsSubgroup "${TESTS_GROUP}"
  fi
fi

###########################################
######## Multi-cluster tests group ########
###########################################
if [[ -z "${GKE_VERSION-}" && ("${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${MULTICLUSTER_TESTS}") ]]; then
  sectionStartTime=$(date +%s)
  info "Running multi-cluster integration tests..."
  test_command=$(getTestCommand "${MULTICLUSTER_TESTS}" "40")
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Multi-cluster integration tests succeeded!!"
  info "Multi-cluster tests execution time: $(elapsedTimeSince "$sectionStartTime")"
fi

####################################
######## Carvel tests group ########
####################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${CARVEL_TESTS}" ]]; then
  sectionStartTime=$(date +%s)

  ## Upgrade and run Carvel test
  installKappController "${KAPP_CONTROLLER_VERSION}"
  info "Updating Kubeapps with carvel support"
  installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
    "--set" "packaging.helm.enabled=false" \
    "--set" "packaging.carvel.enabled=true"

  info "Waiting for updated Kubeapps components to be ready..."
  k8s_wait_for_deployment kubeapps kubeapps-ci

  info "Running carvel integration test..."
  test_command=$(getTestCommand "${CARVEL_TESTS}" "20")
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Carvel integration tests succeeded!!"
  info "Carvel tests execution time: $(elapsedTimeSince "$sectionStartTime")"
fi

####################################
######## Flux tests group ########
####################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${FLUX_TESTS}" ]]; then
  sectionStartTime=$(date +%s)

  ## Upgrade and run Flux test
  installFlux "${FLUX_VERSION}"
  info "Updating Kubeapps with flux support"
  installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
    "--set" "packaging.flux.enabled=true" \
    "--set" "packaging.helm.enabled=false" \
    "--set" "packaging.carvel.enabled=false"

  info "Waiting for updated Kubeapps components to be ready..."
  k8s_wait_for_deployment kubeapps kubeapps-ci

  info "Running flux integration test..."
  test_command=$(getTestCommand "${FLUX_TESTS}" "20")
  info "${test_command}"

  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Flux integration tests succeeded!"
  info "Flux tests execution time: $(elapsedTimeSince "$sectionStartTime")"
fi

#######################################
######## Operators tests group ########
#######################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${OPERATOR_TESTS}" ]]; then
  sectionStartTime=$(date +%s)
  ## Upgrade and run operators test
  # Operators are not supported in GKE 1.14 and flaky in 1.15, skipping test
  if [[ -z "${GKE_VERSION-}" ]] && [[ -n "${TEST_OPERATORS-}" ]]; then
    installOLM "${OLM_VERSION}"

    # Update Kubeapps settings to enable operators and hence proxying
    # to k8s API server. Don't change the packaging setting to avoid
    # re-installing postgres.
    info "Installing latest Kubeapps chart available"
    installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
      "--set" "packaging.helm.enabled=false" \
      "--set" "packaging.carvel.enabled=true" \
      "--set" "featureFlags.operators=true"

    info "Waiting for Kubeapps components to be ready (bitnami chart)..."
    k8s_wait_for_deployment kubeapps kubeapps-ci

    ## Wait for the Operator catalog to be populated
    info "Waiting for the OperatorHub Catalog to be ready ..."
    retry_while isOperatorHubCatalogRunning 24

    info "Running operators integration test with k8s API access..."
    test_command=$(getTestCommand "${OPERATOR_TESTS}" "20")
    info "${test_command}"
    if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
      ## Integration tests failed, get report screenshot
      warn "PODS status on failure"
      kubectl cp "${pod}:/app/reports" ./reports
      exit 1
    fi
    info "Operators integration tests (with k8s API access) succeeded!!"
    info "Operators tests execution time: $(elapsedTimeSince "$sectionStartTime")"
  fi
fi

############################################################
######## Multi-cluster without Kubeapps tests group ########
############################################################
if [[ -z "${GKE_VERSION-}" && ("${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${MULTICLUSTER_NOKUBEAPPS_TESTS}") ]]; then
  sectionStartTime=$(date +%s)
  info "Running multi-cluster (without Kubeapps cluster) integration tests..."

  info "Updating Kubeapps to exclude Kubeapps cluster from the list of clusters"

  # Update Kubeapps
  kubeappsChartPath="${ROOT_DIR}/chart/kubeapps"
  info "Installing Kubeapps from ${kubeappsChartPath}..."
  kubectl -n kubeapps delete secret localhost-tls || true

  # See https://stackoverflow.com/a/36296000 for "${arr[@]+"${arr[@]}"}" notation.
  cmd=(helm upgrade --install kubeapps-ci --namespace kubeapps "${kubeappsChartPath}"
    "${img_flags[@]}"
    "${basicAuthFlags[@]+"${basicAuthFlags[@]}"}"
    --set clusters[0].name=second-cluster
    --set clusters[0].apiServiceURL=https://${ADDITIONAL_CLUSTER_IP}:6443
    --set clusters[0].insecure=true
    --set clusters[0].serviceToken=$(kubectl --context=kind-kubeapps-ci-additional --kubeconfig=${HOME}/.kube/kind-config-kubeapps-ci-additional get secret kubeapps-namespace-discovery -o go-template='{{.data.token | base64decode}}')
    --set frontend.replicaCount=1
    --set dashboard.replicaCount=1
    --set kubeappsapis.replicaCount=2
    --set postgresql.architecture=standalone
    --set postgresql.primary.persistence.enabled=false
    --set postgresql.auth.password=password
    --set redis.auth.password=password
    --set apprepository.initialRepos[0].name=bitnami
    --set apprepository.initialRepos[0].url=http://chartmuseum.chart-museum.svc.cluster.local:8080
    --set apprepository.initialRepos[0].basicAuth.user=admin
    --set apprepository.initialRepos[0].basicAuth.password=password
    --set apprepository.globalReposNamespaceSuffix=-repos-global
    --set global.postgresql.auth.postgresPassword=password
    --set global.security.allowInsecureImages=true
    --wait)

  echo "${cmd[@]}"
  "${cmd[@]}"

  info "Waiting for updated Kubeapps components to be ready..."
  k8s_wait_for_deployment kubeapps kubeapps-ci

  test_command=$(getTestCommand "${MULTICLUSTER_NOKUBEAPPS_TESTS}" "40")
  info "${test_command}"

  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Multi-cluster integration tests succeeded!!"
  info "Multi-cluster tests execution time:$(elapsedTimeSince "$sectionStartTime")"
fi

info "Integration tests succeeded!"
info "Total execution time: $(elapsedTimeSince "$startTime")"
