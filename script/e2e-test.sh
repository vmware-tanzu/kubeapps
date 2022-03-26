#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# Constants
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
USE_MULTICLUSTER_OIDC_ENV=${1:-false}
OLM_VERSION=${2:-"v0.18.2"}
DEV_TAG=${3:?missing dev tag}
IMG_MODIFIER=${4:-""}
DOCKER_USERNAME=${5:-""}
DOCKER_PASSWORD=${6:-""}
TEST_TIMEOUT_MINUTES=${7:-4}
DEX_IP=${8:-"172.18.0.2"}
ADDITIONAL_CLUSTER_IP=${9:-"172.18.0.3"}

# TODO(andresmgot): While we work with beta releases, the Bitnami pipeline
# removes the pre-release part of the tag
if [[ -n "${TEST_LATEST_RELEASE:-}" ]]; then
  DEV_TAG=${DEV_TAG/-beta.*/}
fi

# Load Generic Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/liblog.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/libutil.sh"

info "Root dir: ${ROOT_DIR}"
info "Use multicluster+OIDC: ${USE_MULTICLUSTER_OIDC_ENV}"
info "OLM version: ${OLM_VERSION}"
info "Image tag: ${DEV_TAG}"
info "Image repo suffix: ${IMG_MODIFIER}"
info "Dex IP: ${DEX_IP}"
info "Additional cluster IP : ${ADDITIONAL_CLUSTER_IP}"
info "Test timeout: ${TEST_TIMEOUT_MINUTES}"
info "Cluster Version: $(kubectl version -o json | jq -r '.serverVersion.gitVersion')"
info "Kubectl Version: $(kubectl version -o json | jq -r '.clientVersion.gitVersion')"
echo ""

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
  helm test -n kubeapps kubeapps-ci
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

  kubectl apply -f "${url}/crds.yaml"
  kubectl wait --for=condition=Established -f "${url}/crds.yaml"
  kubectl apply -f "${url}/olm.yaml"

  # wait for deployments to be ready
  kubectl rollout status -w deployment/olm-operator --namespace="${namespace}"
  kubectl rollout status -w deployment/catalog-operator --namespace="${namespace}"

  retries=30
  until [[ $retries == 0 ]]; do
    new_csv_phase=$(kubectl get csv -n "${namespace}" packageserver -o jsonpath='{.status.phase}' 2>/dev/null || echo "Waiting for CSV to appear")
    if [[ $new_csv_phase != "${csv_phase:-}" ]]; then
      csv_phase=$new_csv_phase
      echo "CSV \"packageserver\" phase: $csv_phase"
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
  helm install chartmuseum --namespace kubeapps https://github.com/chartmuseum/charts/releases/download/chartmuseum-2.14.2/chartmuseum-2.14.2.tgz \
    --set env.open.DISABLE_API=false \
    --set persistence.enabled=true \
    --set secret.AUTH_USER=$user \
    --set secret.AUTH_PASS=$password
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
  prefix="kubeapps-"
  description="foo ${chart} chart for CI"

  info "Adding ${chart}-${version} to ChartMuseum ..."
  curl -LO "https://charts.bitnami.com/bitnami/${chart}-${version}.tgz"

  # Mutate the chart name and description, then re-package the tarball
  # For instance, the apache's Chart.yaml file becomes modified to:
  #   name: kubeapps-apache
  #   description: foo apache chart for CI
  # consequently, the new packaged chart is "${prefix}${chart}-${version}.tgz"
  # This workaround should mitigate https://github.com/kubeapps/kubeapps/issues/3339
  mkdir ./${chart}-${version}
  tar zxf ${chart}-${version}.tgz -C ./${chart}-${version}
  sed -i "s/name: ${chart}/name: ${prefix}${chart}/" ./${chart}-${version}/${chart}/Chart.yaml
  sed -i "0,/^\([[:space:]]*description: *\).*/s//\1${description}/" ./${chart}-${version}/${chart}/Chart.yaml
  helm package ./${chart}-${version}/${chart} -d .

  local POD_NAME=$(kubectl get pods --namespace kubeapps -l "app=chartmuseum" -l "release=chartmuseum" -o jsonpath="{.items[0].metadata.name}")
  /bin/sh -c "kubectl port-forward $POD_NAME 8080:8080 --namespace kubeapps &"
  sleep 2
  curl -u "${user}:${password}" --data-binary "@${prefix}${chart}-${version}.tgz" http://localhost:8080/api/charts
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
  info "Installing Kubeapps from ${chartSource}..."
  kubectl -n kubeapps delete secret localhost-tls || true

  # See https://stackoverflow.com/a/36296000 for "${arr[@]+"${arr[@]}"}" notation.
  cmd=(helm upgrade --install kubeapps-ci --namespace kubeapps "${chartSource}"
    "${img_flags[@]}"
    "${@:2}"
    "${multiclusterFlags[@]+"${multiclusterFlags[@]}"}"
    --set frontend.replicaCount=1
    --set kubeops.replicaCount=1
    --set dashboard.replicaCount=1
    --set kubeappsapis.replicaCount=2
    --set kubeops.enabled=true
    --set postgresql.architecture=standalone
    --set postgresql.primary.persistence.enabled=false
    --set postgresql.auth.password=password
    --set redis.auth.password=password
    --set apprepository.initialRepos[0].name=bitnami
    --set apprepository.initialRepos[0].url=http://chartmuseum-chartmuseum.kubeapps:8080
    --set apprepository.initialRepos[0].basicAuth.user=admin
    --set apprepository.initialRepos[0].basicAuth.password=password
    --set apprepository.globalReposNamespaceSuffix=-repos-global
    --wait)

  echo "${cmd[@]}"
  "${cmd[@]}"
}

# Use dev images or Bitnami if testing the latest release
image_prefix="kubeapps/"
kubeapps_apis_image="kubeapps-apis"
[[ -n "${TEST_LATEST_RELEASE:-}" ]] && image_prefix="bitnami/kubeapps-" && kubeapps_apis_image="apis"
images=(
  "apprepository-controller"
  "asset-syncer"
  "assetsvc"
  "dashboard"
  "kubeops"
  "pinniped-proxy"
  "${kubeapps_apis_image}"
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
  "--set" "kubeops.image.tag=${DEV_TAG}"
  "--set" "kubeops.image.repository=${images[4]}"
  "--set" "pinnipedProxy.image.tag=${DEV_TAG}"
  "--set" "pinnipedProxy.image.repository=${images[5]}"
  "--set" "kubeappsapis.image.tag=${DEV_TAG}"
  "--set" "kubeappsapis.image.repository=${images[6]}"
)

if [ "$USE_MULTICLUSTER_OIDC_ENV" = true ]; then
  multiclusterFlags=(
    "--set" "ingress.enabled=true"
    "--set" "ingress.hostname=localhost"
    "--set" "ingress.tls=true"
    "--set" "ingress.selfSigned=true"
    "--set" "authProxy.enabled=true"
    "--set" "authProxy.provider=oidc"
    "--set" "authProxy.clientID=default"
    "--set" "authProxy.clientSecret=ZXhhbXBsZS1hcHAtc2VjcmV0"
    "--set" "authProxy.cookieSecret=bm90LWdvb2Qtc2VjcmV0Cg=="
    "--set" "authProxy.extraFlags[0]=\"--oidc-issuer-url=https://${DEX_IP}:32000\""
    "--set" "authProxy.extraFlags[1]=\"--scope=openid email groups audience:server:client_id:second-cluster audience:server:client_id:third-cluster\""
    "--set" "authProxy.extraFlags[2]=\"--ssl-insecure-skip-verify=true\""
    "--set" "authProxy.extraFlags[3]=\"--redirect-url=http://kubeapps-ci.kubeapps/oauth2/callback\""
    "--set" "authProxy.extraFlags[4]=\"--cookie-secure=false\""
    "--set" "authProxy.extraFlags[5]=\"--cookie-domain=kubeapps-ci.kubeapps\""
    "--set" "authProxy.extraFlags[6]=\"--whitelist-domain=kubeapps-ci.kubeapps\""
    "--set" "authProxy.extraFlags[7]=\"--set-authorization-header=true\""
    "--set" "clusters[0].name=default"
    "--set" "clusters[1].name=second-cluster"
    "--set" "clusters[1].apiServiceURL=https://${ADDITIONAL_CLUSTER_IP}:6443"
    "--set" "clusters[1].insecure=true"
    "--set" "clusters[1].serviceToken=ZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklsbHpiSEp5TlZwM1QwaG9WSE5PYkhVdE5GQkRablY2TW0wd05rUmtMVmxFWVV4MlZEazNaeTEyUmxFaWZRLmV5SnBjM01pT2lKcmRXSmxjbTVsZEdWekwzTmxjblpwWTJWaFkyTnZkVzUwSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXVZVzFsYzNCaFkyVWlPaUprWldaaGRXeDBJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbXQxWW1WaGNIQnpMVzVoYldWemNHRmpaUzFrYVhOamIzWmxjbmt0ZEc5clpXNHRjV295Ym1naUxDSnJkV0psY201bGRHVnpMbWx2TDNObGNuWnBZMlZoWTJOdmRXNTBMM05sY25acFkyVXRZV05qYjNWdWRDNXVZVzFsSWpvaWEzVmlaV0Z3Y0hNdGJtRnRaWE53WVdObExXUnBjMk52ZG1WeWVTSXNJbXQxWW1WeWJtVjBaWE11YVc4dmMyVnlkbWxqWldGalkyOTFiblF2YzJWeWRtbGpaUzFoWTJOdmRXNTBMblZwWkNJNkltVXhaakE1WmpSakxUTTRNemt0TkRJME15MWhZbUptTFRKaU5HWm1OREZrWW1RMllTSXNJbk4xWWlJNkluTjVjM1JsYlRwelpYSjJhV05sWVdOamIzVnVkRHBrWldaaGRXeDBPbXQxWW1WaGNIQnpMVzVoYldWemNHRmpaUzFrYVhOamIzWmxjbmtpZlEuTnh6V2dsUGlrVWpROVQ1NkpWM2xJN1VWTUVSR3J2bklPSHJENkh4dUVwR0luLWFUUzV5Q0pDa3Z0cTF6S3Z3b05sc2MyX0YxaTdFOUxWRGFwbC1UQlhleUN5Rl92S1B1TDF4dTdqZFBMZ1dKT1pQX3JMcXppaDV4ZlkxalFoOHNhdTRZclFJLUtqb3U1UkRRZ0tOQS1BaS1lRlFOZVh2bmlUNlBKYWVkc184V0t3dHRMMC1wdHpYRnBnOFl5dkx6N0U1UWdTR2tjNWpDVXlsS0RvZVRUaVRSOEc2RHFHYkFQQUYwREt0b3MybU9Geno4SlJYNHhoQmdvaUcxVTVmR1g4Z3hnTU1SV0VHRE9kaGMyeXRvcFdRUkRpYmhvaldNS3VDZlNua09zMDRGYTBkYmEwQ0NTbld2a29LZ3Z4QVR5aVVrWm9wV3VpZ1JJNFd5dDkzbXhR"
  )
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

installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps"
info "Waiting for Kubeapps components to be ready (local chart)..."
k8s_wait_for_deployment kubeapps kubeapps-ci
installChartmuseum admin password
pushChart apache 8.6.2 admin password
pushChart apache 8.6.3 admin password

# Ensure that we are testing the correct image
info ""
k8s_ensure_image kubeapps kubeapps-ci-internal-apprepository-controller "$DEV_TAG"
k8s_ensure_image kubeapps kubeapps-ci-internal-dashboard "$DEV_TAG"
k8s_ensure_image kubeapps kubeapps-ci-internal-kubeappsapis "$DEV_TAG"

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

# Deactivate helm tests unless we are testing the latest release until
# we have released the code with per-namespace tests (since the helm
# tests for assetsvc needs to test the namespaced repo).
if [[ -z "${TEST_LATEST_RELEASE:-}" ]]; then
  # Run helm tests
  # Retry once if tests fail to avoid temporary issue
  if ! retry_while testHelm "2" "1"; then
    warn "PODS status on failure"
    kubectl get pods -n kubeapps
    for pod in $(kubectl get po -l='app.kubernetes.io/managed-by=Helm,app.kubernetes.io/instance=kubeapps-ci' -oname -n kubeapps); do
      warn "LOGS for pod $pod ------------"
      if [[ "$pod" =~ .*internal.* ]]; then
        kubectl logs -n kubeapps "$pod"
      else
        kubectl logs -n kubeapps "$pod" nginx
        kubectl logs -n kubeapps "$pod" auth-proxy
      fi
    done
    echo
    warn "LOGS for dashboard tests --------"
    kubectl logs kubeapps-ci-dashboard-test --namespace kubeapps
    exit 1
  fi
  info "Helm tests succeeded!"
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

# Set tests to be run
# Playwright does not allow to ignore tests on command line, only in config file
testsToRun=("tests/main/")
# Skip the multicluster scenario for GKE
if [[ -z "${GKE_BRANCH-}" ]]; then
  testsToRun+=("tests/multicluster/")
fi
testsArgs="$(printf "%s " "${testsToRun[@]}")"

kubectl cp ./tests "${pod}:/app/"
info "Copied tests to integration pod ${pod}"
## Create admin user
kubectl create serviceaccount kubeapps-operator -n kubeapps
kubectl create clusterrolebinding kubeapps-operator-admin --clusterrole=cluster-admin --serviceaccount kubeapps:kubeapps-operator
kubectl create clusterrolebinding kubeapps-repositories-write --clusterrole kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
kubectl create rolebinding kubeapps-sa-operator-apprepositories-write -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
## Create view user
kubectl create serviceaccount kubeapps-view -n kubeapps
kubectl create role view-secrets --verb=get,list,watch --resource=secrets
kubectl create rolebinding kubeapps-view-secret --role view-secrets --serviceaccount kubeapps:kubeapps-view
kubectl create clusterrolebinding kubeapps-view --clusterrole=view --serviceaccount kubeapps:kubeapps-view
kubectl create rolebinding kubeapps-view-user-apprepo-read -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-read --serviceaccount kubeapps:kubeapps-view
kubectl create rolebinding kubeapps-view-user -n kubeapps-user-namespace --clusterrole=edit --serviceaccount kubeapps:kubeapps-view
## Create edit user
kubectl create serviceaccount kubeapps-edit -n kubeapps
kubectl create rolebinding kubeapps-edit -n kubeapps --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
kubectl create rolebinding kubeapps-edit -n default --clusterrole=edit --serviceaccount kubeapps:kubeapps-edit
kubectl create rolebinding kubeapps-repositories-read -n kubeapps --clusterrole kubeapps:kubeapps:apprepositories-read --serviceaccount kubeapps:kubeapps-edit

## Give the cluster some time to avoid issues like
## https://circleci.com/gh/kubeapps/kubeapps/16102
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-operator -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-view -o name" "5" "1"
retry_while "kubectl get -n kubeapps serviceaccount kubeapps-edit -o name" "5" "1"
## Retrieve tokens
admin_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
view_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-view -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
edit_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-edit -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"

info "Running main Integration tests without k8s API access..."
if ! kubectl exec -it "$pod" -- /bin/sh -c "CI_TIMEOUT_MINUTES=40 DOCKER_USERNAME=${DOCKER_USERNAME} DOCKER_PASSWORD=${DOCKER_PASSWORD} TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} EDIT_TOKEN=${edit_token} yarn test ${testsArgs}"; then
  ## Integration tests failed, get report screenshot
  warn "PODS status on failure"
  kubectl cp "${pod}:/app/reports" ./reports
  exit 1
fi
info "Main integration tests succeeded!!"

## Upgrade and run operator test
# Operators are not supported in GKE 1.14 and flaky in 1.15, skipping test
if [[ -z "${GKE_BRANCH-}" ]] && [[ -n "${TEST_OPERATORS-}" ]]; then
  installOLM "${OLM_VERSION}"

  # Update Kubeapps settings to enable operators and hence proxying
  # to k8s API server.
  info "Installing latest Kubeapps chart available"
  installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
    "--set" "featureFlags.operators=true"

  info "Waiting for Kubeapps components to be ready (bitnami chart)..."
  k8s_wait_for_deployment kubeapps kubeapps-ci

  ## Wait for the Operator catalog to be populated
  info "Waiting for the OperatorHub Catalog to be ready ..."
  retry_while isOperatorHubCatalogRunning 24

  info "Running operator integration test with k8s API access..."
  if ! kubectl exec -it "$pod" -- /bin/sh -c "CI_TIMEOUT_MINUTES=20 TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} EDIT_TOKEN=${edit_token} yarn test \"tests/operators/\""; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Operator integration tests (with k8s API access) succeeded!!"
fi
info "Integration tests succeeded!"
