#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

startTime=$(date +%s)

# Constants
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"
ALL_TESTS="all"
MAIN_TESTS="main"
MULTICLUSTER_TESTS="multicluster"
CARVEL_TESTS="carvel"
OPERATOR_TESTS="operator"
SUPPORTED_TESTS_GROUPS=("${ALL_TESTS}" "${MAIN_TESTS}" "${MULTICLUSTER_TESTS}" "${CARVEL_TESTS}" "${OPERATOR_TESTS}")

# Params
USE_MULTICLUSTER_OIDC_ENV=${1:-false}
OLM_VERSION=${2:-"v0.18.2"}
IMG_DEV_TAG=${3:?missing dev tag}
IMG_MODIFIER=${4:-""}
TEST_TIMEOUT_MINUTES=${5:-"4"}
DEX_IP=${6:-"172.18.0.2"}
ADDITIONAL_CLUSTER_IP=${7:-"172.18.0.3"}
KAPP_CONTROLLER_VERSION=${8:-"v0.41.2"}
CHARTMUSEUM_VERSION=${9:-"3.9.0"}
# check latest flux releases at https://github.com/fluxcd/flux2/releases
FLUX_VERSION=${10:-"v0.35.0"}
IMG_PREFIX=${IMG_PREFIX:-"kubeapps/"}
TESTS_GROUP=${TESTS_GROUP:-"${ALL_TESTS}"}

# shellcheck disable=SC2076
if [[ ! " ${SUPPORTED_TESTS_GROUPS[*]} " =~ " ${TESTS_GROUP} " ]]; then
  # shellcheck disable=SC2046
  echo $(IFS=','; echo "The provided TEST_GROUP [${TESTS_GROUP}] is not supported. Supported groups are: ${SUPPORTED_TESTS_GROUPS[*]}")
  exit 1
fi

# TODO(andresmgot): While we work with beta releases, the Bitnami pipeline
# removes the pre-release part of the tag
if [[ -n "${TEST_LATEST_RELEASE:-}" ]]; then
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
if [[ -z "${GKE_BRANCH-}" ]]; then
  LOAD_BALANCER_IP=$DEX_IP
else
  LOAD_BALANCER_IP=$(kubectl -n nginx-ingress get service nginx-ingress-ingress-nginx-controller -o jsonpath="{.status.loadBalancer.ingress[].ip}")
fi

# Functions for local Docker registry mgmt
. "${ROOT_DIR}/script/local-docker-registry.sh"

# Functions for handling Chart Museum
. "${ROOT_DIR}/script/chart-museum.sh"

info "TESTS GROUP: ${TESTS_GROUP}"
info "Root dir: ${ROOT_DIR}"
info "Use multicluster+OIDC: ${USE_MULTICLUSTER_OIDC_ENV}"
info "OLM version: ${OLM_VERSION}"
info "ChartMuseum version: ${CHARTMUSEUM_VERSION}"
info "Image tag: ${IMG_DEV_TAG}"
info "Image modifier: ${IMG_MODIFIER}"
info "Image prefix: ${IMG_PREFIX}"
info "Dex IP: ${DEX_IP}"
info "Additional cluster IP : ${ADDITIONAL_CLUSTER_IP}"
info "Load balancer IP : ${LOAD_BALANCER_IP}"
info "Test timeout minutes: ${TEST_TIMEOUT_MINUTES}"
info "Kapp Controller version: ${KAPP_CONTROLLER_VERSION}"
info "Cluster Version: $(kubectl version -o json | jq -r '.serverVersion.gitVersion')"
info "Kubectl Version: $(kubectl version -o json | jq -r '.clientVersion.gitVersion')"
echo ""

# Auxiliar functions

#
# Install an authenticated Docker registry inside the cluster
#
setupLocalDockerRegistry() {
    info "Installing local Docker registry with authentication"
    installLocalRegistry $ROOT_DIR

    info "Pushing test container to local Docker registry"
    pushContainerToLocalRegistry
}

#
# Push a chart that uses container image from the local registry
#
pushLocalChart() {
    info "Packaging local test chart"
    helm package $ROOT_DIR/integration/charts/simplechart

    info "Pushing local test chart to ChartMuseum"
    pushChartToChartMuseum "simplechart" "0.1.0" "simplechart-0.1.0.tgz"
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
# Push a chart to chartmusem
# Globals: None
# Arguments:
#   $1: chart
#   $2: version
# Returns: None
#########################
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
  mkdir ./${chart}-${version}
  tar zxf ${chart}-${version}.tgz -C ./${chart}-${version}
  # this relies on GNU sed, which is not the default on MacOS
  # ref https://gist.github.com/andre3k1/e3a1a7133fded5de5a9ee99c87c6fa0d
  sed -i "s/name: ${chart}/name: ${prefix}${chart}/" ./${chart}-${version}/${chart}/Chart.yaml
  sed -i "0,/^\([[:space:]]*description: *\).*/s//\1${description}/" ./${chart}-${version}/${chart}/Chart.yaml
  helm package ./${chart}-${version}/${chart} -d .

  pushChartToChartMuseum "${chart}" "${version}" "${prefix}${chart}-${version}.tgz"
}

########################
# Install kapp-controller
# Globals: None
# Arguments:
#   $1: Version of kapp-controller
# Returns: None
#########################
installKappController() {
  local release=$1
  info "Installing kapp-controller ${release} ..."
  url=https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/${release}/release.yml
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

########################
# Install flux
# Globals: None
# Arguments:
#   $1: Version of flux
# Returns: None
#########################
installFlux() {
  local release=$1
  info "Installing flux ${release} ..."
  url=https://github.com/fluxcd/flux2/releases/download/${release}/install.yaml
  namespace=flux-system

  kubectl apply -f "${url}"

  # wait for deployment to be ready
  kubectl rollout status -w deployment/helm-controller --namespace="${namespace}"
  kubectl rollout status -w deployment/source-controller --namespace="${namespace}"

  # Add test repository.
	kubectl apply -f https://raw.githubusercontent.com/fluxcd/source-controller/main/config/samples/source_v1beta2_helmrepository.yaml

  # Add a flux-reconciler service account to the kubeapps-user-namespace with
  # cluster-admin.
  kubectl create serviceaccount flux-reconciler -n kubeapps-user-namespace
  kubectl create clusterrolebinding flux-reconciler --clusterrole=cluster-admin --serviceaccount kubeapps-user-namespace:flux-reconciler
}

########################
# Creates a Yaml file with additional values for the Helm chart
# Arguments: None
# Returns: Path to the newly created file with additional values
#########################
generateAdditionalValuesFile() {
  # Could be done better with $(cat <<EOF > ${ROOT_DIR}/additional_chart_values.yaml
  # But it was breaking the formatting of the file
  local valuesFile=${ROOT_DIR}/additional_chart_values.yaml;
  echo "ingress:
  enabled: true
  hostname: localhost
  tls: true
  selfSigned: true
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/proxy-buffer-size: \"8k\"
    nginx.ingress.kubernetes.io/proxy-buffers: \"4.0\"
    nginx.ingress.kubernetes.io/proxy-read-timeout: \"600.0\"" > ${valuesFile}
  echo ${valuesFile}
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
    --wait)

  echo "${cmd[@]}"
  "${cmd[@]}"
}

<<<<<<< HEAD
createServiceAccountsAndRBAC() {
  ## Create admin user
  # using YAML to workaround https://github.com/vmware-tanzu/kubeapps/pull/4772
  kubectl create --namespace kubeapps serviceaccount kubeapps-operator
  kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: kubeapps-operator-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: "kubeapps-operator"
EOF
  kubectl patch serviceaccount -n kubeapps kubeapps-operator -p '{"secrets": [{"name": "kubeapps-operator-token"}]}'
  kubectl create clusterrolebinding kubeapps-operator-admin --clusterrole=cluster-admin --serviceaccount kubeapps:kubeapps-operator
  kubectl create clusterrolebinding kubeapps-repositories-write --clusterrole kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
  kubectl create rolebinding kubeapps-sa-operator-apprepositories-write -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-write --serviceaccount kubeapps:kubeapps-operator
  ## Create view user
  kubectl create serviceaccount kubeapps-view -n kubeapps
  kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: kubeapps-view-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: "kubeapps-view"
EOF
  kubectl patch serviceaccount -n kubeapps kubeapps-view -p '{"secrets": [{"name": "kubeapps-view-token"}]}'
  kubectl create role view-secrets --verb=get,list,watch --resource=secrets
  kubectl create rolebinding kubeapps-view-secret --role view-secrets --serviceaccount kubeapps:kubeapps-view
  kubectl create clusterrolebinding kubeapps-view --clusterrole=view --serviceaccount kubeapps:kubeapps-view
  kubectl create rolebinding kubeapps-view-user-apprepo-read -n kubeapps-user-namespace --clusterrole=kubeapps:kubeapps:apprepositories-read --serviceaccount kubeapps:kubeapps-view
  kubectl create rolebinding kubeapps-view-user -n kubeapps-user-namespace --clusterrole=edit --serviceaccount kubeapps:kubeapps-view
  ## Create edit user
  kubectl create serviceaccount kubeapps-edit -n kubeapps
  kubectl create -f - <<EOF
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: kubeapps-edit-token
  namespace: kubeapps
  annotations:
    kubernetes.io/service-account.name: "kubeapps-edit"
EOF
  kubectl patch serviceaccount -n kubeapps kubeapps-edit -p '{"secrets": [{"name": "kubeapps-edit-token"}]}'
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
}

=======
########################
# Formats the provided time in seconds.
# Arguments:
#   $1: time in seconds
# Returns: Time formatted as Xm Ys
#########################
formattedElapsedTime() {
  time=$1

  mins=$((time/60))
  secs=$((time%60))
  echo "${mins}m ${secs}s"
}

if [[ "${DEBUG_MODE:-false}" == "true" ]]; then
  info "Docker images loaded in the cluster:"
  docker exec kubeapps-ci-control-plane crictl images
fi

>>>>>>> main
# Use dev images or Bitnami if testing the latest release
kubeapps_apis_image="kubeapps-apis"
[[ -n "${TEST_LATEST_RELEASE:-}" ]] && IMG_PREFIX="bitnami/kubeapps-" && kubeapps_apis_image="apis"
images=(
  "apprepository-controller"
  "asset-syncer"
  "dashboard"
  "pinniped-proxy"
  "${kubeapps_apis_image}"
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
)

additional_flags_file=$(generateAdditionalValuesFile)

if [ "$USE_MULTICLUSTER_OIDC_ENV" = true ]; then
  multiclusterFlags=(
    "--values" "${additional_flags_file}"
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
    "--set" "clusters[1].serviceToken=$(kubectl --context=kind-kubeapps-ci-additional --kubeconfig=${HOME}/.kube/kind-config-kubeapps-ci-additional get secret kubeapps-namespace-discovery -o go-template='{{.data.token | base64decode}}')"
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

# Install ChartMuseum
installChartMuseum "${CHARTMUSEUM_VERSION}"
pushChart apache 8.6.2
pushChart apache 8.6.3

# Install Kubeapps
installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps"
info "Waiting for Kubeapps components to be ready (local chart)..."
k8s_wait_for_deployment kubeapps kubeapps-ci

# Setting up local Docker registry if not in GKE
if [[ -z "${GKE_BRANCH-}" ]]; then
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
kubectl create deployment e2e-runner --image ${IMG_PREFIX}integration-tests${IMG_MODIFIER}:${IMG_DEV_TAG}
k8s_wait_for_deployment default e2e-runner
pod=$(kubectl get po -l app=e2e-runner -o custom-columns=:metadata.name --no-headers)
## Copy config and latest tests
for f in *.js; do
  kubectl cp "./${f}" "${pod}:/app/"
done

kubectl cp ./tests "${pod}:/app/"
info "Copied tests to e2e-runner pod ${pod}"

<<<<<<< HEAD
createServiceAccountsAndRBAC

## Retrieve tokens
admin_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
view_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-view -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"
edit_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-edit -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}' && echo)"

info "Running main Integration tests without k8s API access..."
if ! kubectl exec -it "$pod" -- /bin/sh -c "CI_TIMEOUT_MINUTES=40 DOCKER_USERNAME=${DOCKER_USERNAME} DOCKER_PASSWORD=${DOCKER_PASSWORD} DOCKER_REGISTRY_URL=${DOCKER_REGISTRY_URL} TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} EDIT_TOKEN=${edit_token} yarn test ${testsArgs}"; then
  ## Integration tests failed, get report screenshot
  warn "PODS status on failure"
  kubectl cp "${pod}:/app/reports" ./reports
  exit 1
fi
info "Main integration tests succeeded!!"

## Upgrade and run Carvel test
installKappController "${KAPP_CONTROLLER_VERSION}"
info "Updating Kubeapps with carvel support"
installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
  "--set" "packaging.helm.enabled=false" \
  "--set" "packaging.carvel.enabled=true"
=======
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
admin_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-operator -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}')"
view_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-view -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}')"
edit_token="$(kubectl get -n kubeapps secret "$(kubectl get -n kubeapps serviceaccount kubeapps-edit -o jsonpath='{.secrets[].name}')" -o go-template='{{.data.token | base64decode}}')"

endTime=$(date +%s)
info "Bootstrap time: $(formattedElapsedTime endTime-startTime)"

##################################
######## Main tests group ########
##################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${MAIN_TESTS}" ]]; then
  sectionStartTime=$(date +%s)
  info "Running main Integration tests without k8s API access..."
  test_command="
    CI_TIMEOUT_MINUTES=40 \
    DOCKER_USERNAME=${DOCKER_USERNAME} \
    DOCKER_PASSWORD=${DOCKER_PASSWORD} \
    DOCKER_REGISTRY_URL=${DOCKER_REGISTRY_URL} \
    TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} \
    INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps \
    USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} \
    ADMIN_TOKEN=${admin_token} \
    VIEW_TOKEN=${view_token} \
    EDIT_TOKEN=${edit_token} \
    yarn test \"tests/main/\"
    "
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Main integration tests succeeded!!"

  sectionEndTime=$(date +%s)
  info "Main tests execution time: $(formattedElapsedTime sectionEndTime-sectionStartTime)"
fi
>>>>>>> main

###########################################
######## Multi-cluster tests group ########
###########################################
if [[ -z "${GKE_BRANCH-}" && ("${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${MULTICLUSTER_TESTS}") ]]; then
  sectionStartTime=$(date +%s)
  info "Running multi-cluster integration tests..."
  test_command="
    CI_TIMEOUT_MINUTES=40 \
    DOCKER_USERNAME=${DOCKER_USERNAME} \
    DOCKER_PASSWORD=${DOCKER_PASSWORD} \
    DOCKER_REGISTRY_URL=${DOCKER_REGISTRY_URL} \
    TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} \
    INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps \
    USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} \
    ADMIN_TOKEN=${admin_token} \
    VIEW_TOKEN=${view_token} \
    EDIT_TOKEN=${edit_token} \
    yarn test \"tests/multicluster/\"
    "
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Multi-cluster integration tests succeeded!!"

  sectionEndTime=$(date +%s)
  info "Multi-cluster tests execution time: $(formattedElapsedTime sectionEndTime-sectionStartTime)"
fi

<<<<<<< HEAD
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
if ! kubectl exec -it "$pod" -- /bin/sh -c "CI_TIMEOUT_MINUTES=20 TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} ADMIN_TOKEN=${admin_token} VIEW_TOKEN=${view_token} EDIT_TOKEN=${edit_token} yarn test \"tests/flux/\""; then
  ## Integration tests failed, get report screenshot
  warn "PODS status on failure"
  kubectl cp "${pod}:/app/reports" ./reports
  exit 1
fi
info "Flux integration tests succeeded!"

## Upgrade and run operator test
# Operators are not supported in GKE 1.14 and flaky in 1.15, skipping test
if [[ -z "${GKE_BRANCH-}" ]] && [[ -n "${TEST_OPERATORS-}" ]]; then
  installOLM "${OLM_VERSION}"
=======
####################################
######## Carvel tests group ########
####################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${CARVEL_TESTS}" ]]; then
  sectionStartTime=$(date +%s)
>>>>>>> main

  ## Upgrade and run Carvel test
  installKappController "${KAPP_CONTROLLER_VERSION}"
  info "Updating Kubeapps with carvel support"
  installOrUpgradeKubeapps "${ROOT_DIR}/chart/kubeapps" \
    "--set" "packaging.helm.enabled=false" \
    "--set" "packaging.carvel.enabled=true"

  info "Waiting for updated Kubeapps components to be ready..."
  k8s_wait_for_deployment kubeapps kubeapps-ci

  info "Running carvel integration test..."
  test_command="
    CI_TIMEOUT_MINUTES=20 \
    TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} \
    INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps \
    USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} \
    ADMIN_TOKEN=${admin_token} \
    VIEW_TOKEN=${view_token} \
    EDIT_TOKEN=${edit_token} \
    yarn test \"tests/carvel/\"
    "
  info "${test_command}"
  if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
    ## Integration tests failed, get report screenshot
    warn "PODS status on failure"
    kubectl cp "${pod}:/app/reports" ./reports
    exit 1
  fi
  info "Carvel integration tests succeeded!!"

  sectionEndTime=$(date +%s)
  info "Carvel tests execution time: $(formattedElapsedTime sectionEndTime-sectionStartTime)"
fi

#######################################
######## Operators tests group ########
#######################################
if [[ "${TESTS_GROUP}" == "${ALL_TESTS}" || "${TESTS_GROUP}" == "${OPERATOR_TESTS}" ]]; then
  sectionStartTime=$(date +%s)
  ## Upgrade and run operator test
  # Operators are not supported in GKE 1.14 and flaky in 1.15, skipping test
  if [[ -z "${GKE_BRANCH-}" ]] && [[ -n "${TEST_OPERATORS-}" ]]; then
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

    info "Running operator integration test with k8s API access..."
    test_command="
      CI_TIMEOUT_MINUTES=20 \
      TEST_TIMEOUT_MINUTES=${TEST_TIMEOUT_MINUTES} \
      INTEGRATION_ENTRYPOINT=http://kubeapps-ci.kubeapps \
      USE_MULTICLUSTER_OIDC_ENV=${USE_MULTICLUSTER_OIDC_ENV} \
      ADMIN_TOKEN=${admin_token} \
      VIEW_TOKEN=${view_token} \
      EDIT_TOKEN=${edit_token} \
      yarn test \"tests/operators/\"
      "
    if ! kubectl exec -it "$pod" -- /bin/sh -c "${test_command}"; then
      ## Integration tests failed, get report screenshot
      warn "PODS status on failure"
      kubectl cp "${pod}:/app/reports" ./reports
      exit 1
    fi
    info "Operator integration tests (with k8s API access) succeeded!!"

    sectionEndTime=$(date +%s)
    info "Operator tests execution time: $(formattedElapsedTime sectionEndTime-sectionStartTime)"
  fi
fi

info "Integration tests succeeded!"

totalTime=$(date +%s)
info "Total execution time: $(formattedElapsedTime totalTime-startTime)"
