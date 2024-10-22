#!/usr/bin/env bash

# Copyright 2022-2023 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null && pwd)"

. "${ROOT_DIR}/script/lib/liblog.sh"

########################################################################################################################
# Install GitHub CLI
# Globals: None
# Arguments:
#   $1: Version of GitHub
# Returns: None
########################################################################################################################
function installGithubCLI() {
  GITHUB_VERSION=${1:?GitHub version not provided}

  info "Installing GitHub CLI $GITHUB_VERSION"
  pushd /tmp
    wget "https://github.com/cli/cli/releases/download/v${GITHUB_VERSION}/gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    tar zxf "gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    rm "gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    sudo mv "gh_${GITHUB_VERSION}_linux_amd64/bin/gh" /usr/local/bin/
  popd
  info "Done"
}

########################################################################################################################
# Install Semver
# Globals: None
# Arguments:
#   $1: Version of Semver
# Returns: None
########################################################################################################################
function installSemver() {
  SEMVER_VERSION=${1:?Semver version not provided}

  info "Installing Semver ${SEMVER_VERSION}"
  pushd /tmp
    wget "https://github.com/fsaintjacques/semver-tool/archive/refs/tags/${SEMVER_VERSION}.tar.gz"
    tar zxf "${SEMVER_VERSION}.tar.gz"
    rm "${SEMVER_VERSION}.tar.gz"
    cd "semver-tool-${SEMVER_VERSION}"
    sudo make install
  popd
  info "Done"
}

########################################################################################################################
# Install GPG key
# Globals: None
# Arguments:
#   $1: GPG public key
#   $2: GPG private key
#   $1: CI BOT GPG
#   $1: CI BOT EMAIL
# Returns: None
########################################################################################################################
function installGPGKey() {
  info "Installing the GPG KEY"
  # Creating the files from the GPG_KEY_PUBLIC and GPG_KEY_PRIVATE env vars
  echo -e "${GPG_KEY_PUBLIC}" > /tmp/public.key
  echo -e "${GPG_KEY_PRIVATE}" > /tmp/private.key

  # Importing the GPG keys
  gpg --import /tmp/public.key
  gpg --import --no-tty --batch --yes /tmp/private.key

  info "Trusting the CI BOT GPG KEY ${CI_BOT_GPG}"
  # Trusting the imported GPG private key
  (echo 5; echo y; echo save) |  gpg --command-fd 0 --no-tty --no-greeting -q --edit-key "${CI_BOT_GPG}" trust

  # Listing the key to verify the import process succeeded
  gpg --list-secret-keys ${CI_BOT_EMAIL}
  info "Done"
}

########################################################################################################################
# Install Kind
# Globals: None
# Arguments:
#   $1: Version of Kind
# Returns: None
########################################################################################################################
function installKind() {
  KIND_VERSION=${1:?Kind version not provided}

  info "Installing Kind ${KIND_VERSION}"
  pushd /tmp
    curl -LO https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-Linux-amd64
    chmod +x kind-Linux-amd64
    sudo mv kind-Linux-amd64 /usr/local/bin/kind
  popd
  info "Done"
}

########################################################################################################################
# Install Kubectl
# Globals: None
# Arguments:
#   $1: Version of kubectl
# Returns: None
########################################################################################################################
function installKubectl() {
  KUBECTL_VERSION=${1:?kubectl version not provided}

  info "Installing Kubectl ${KUBECTL_VERSION}"
  pushd /tmp
    curl -LO https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
    chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin/kubectl
  popd
  info "Done"
}

########################################################################################################################
# Install Mkcert
# Globals: None
# Arguments:
#   $1: Version of mkcert
# Returns: None
########################################################################################################################
function installMkcert() {
  MKCERT_VERSION=${1:?Mkcert version not provided}

  info "Installing Mkcert ${MKCERT_VERSION}"
  pushd /tmp
    curl -LO "https://github.com/FiloSottile/mkcert/releases/download/${MKCERT_VERSION}/mkcert-${MKCERT_VERSION}-linux-amd64"
    chmod +x "mkcert-${MKCERT_VERSION}-linux-amd64"
    sudo mv "mkcert-${MKCERT_VERSION}-linux-amd64" /usr/local/bin/mkcert
    mkcert -install
  popd
  info "Done"
}

########################################################################################################################
# Install Helm
# Globals: None
# Arguments:
#   $1: Version of Helm
#   $2: Helm binary name (OPTIONAL, allows to install different versions of Helm)
# Returns: None
########################################################################################################################
function installHelm() {
  HELM_VERSION=${1:?Helm version not provided}
  HELM_BINARY_NAME=${2:-helm}

  info "Installing Helm ${HELM_VERSION} as ${HELM_BINARY_NAME}"
  pushd /tmp
    wget "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz"
    tar zxf "helm-$HELM_VERSION-linux-amd64.tar.gz"
    sudo mv linux-amd64/helm "/usr/local/bin/${HELM_BINARY_NAME}"
  popd
  info "Done"
}

########################################################################################################################
# Install GCloud SDK
# Globals:
#   GITHUB_ENV: An environment variable available in GitHub Actions' runners that allow setting environment variables
#               that will be available for the next steps of the job.
#   HOME: The HOME env var for the current user.
#   PATH: The system path for the current user.
# Arguments: None
# Returns: None
########################################################################################################################
function installGCloudSDK() {
  info "Installing GCloud SDK"
  gcloud --version && info "Already installed" && return 0;

  GCLOUD_PATH="${HOME}/google-cloud-sdk"
  echo "PATH=${PATH}:${GCLOUD_PATH}/bin" >> "${GITHUB_ENV}"
  echo "CLOUDSDK_CORE_DISABLE_PROMPTS=1" >> "${GITHUB_ENV}"
  if [ ! -d "${GCLOUD_PATH}/bin" ]; then
    rm -rf "${GCLOUD_PATH}"
    curl https://sdk.cloud.google.com | bash;
  fi
  info "Done"
}

########################################################################################################################
# Install GCloud DEB package source to be able to install GCloud SKD and its components through APT.
# Globals: None
# Arguments: None
# Returns: None
########################################################################################################################
function installGCloudPackageSource() {
  info "Installing GCloud package source"
  cat /etc/apt/sources.list.d/google-cloud-sdk.list. &> /dev/null && info "Already installed" && return 0;

  echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
  sudo apt-get update
  info "Done"
}

########################################################################################################################
# Install gke-gcloud-auth-plugin
# Globals: None
# Arguments: None
# Returns: None
########################################################################################################################
function installGKEAuthPlugin() {
  info "Installing gke-gcloud-auth-plugin"
  gke-gcloud-auth-plugin --version &> /dev/null && info "Already installed" && return 0

  sudo apt-get install google-cloud-sdk-gke-gcloud-auth-plugin
  info "Done"
}

########################################################################################################################
# Configure GCloud
# Globals:
#   GITHUB_ENV: An environment variable available in GitHub Actions' runners that allow setting environment variables
#               that will be available for the next steps of the job.
# Arguments:
#   $1: GKE_PROJECT
#   $2: GCLOUD_KEY
# Returns: None
########################################################################################################################
function configureGCloud() {
  GKE_PROJECT=${1:?GKE_PROJECT not provided}
  GCLOUD_KEY=${2:=GCLOUD_KEY not provided}

  info "Configuring GCloud"
  installGCloudPackageSource
  gcloud -q config set project "${GKE_PROJECT}"
  export GOOGLE_APPLICATION_CREDENTIALS=/tmp/client_secrets.json
  # Just exporting the env var won't make it available for the next steps in the GHA's job, so we need the line below
  echo "GOOGLE_APPLICATION_CREDENTIALS=${GOOGLE_APPLICATION_CREDENTIALS}" >> "${GITHUB_ENV}"
  echo "${GCLOUD_KEY}" > "${GOOGLE_APPLICATION_CREDENTIALS}"
  gcloud -q auth activate-service-account --key-file "${GOOGLE_APPLICATION_CREDENTIALS}";
  installGKEAuthPlugin
  echo "USE_GKE_GCLOUD_AUTH_PLUGIN=True" >> "${GITHUB_ENV}"

  info "Done"
}

########################################################################################################################
# Escape and export as an environment variable the name of the GKE cluster
# Globals:
#   GITHUB_ENV: An environment variable available in GitHub Actions' runners that allow setting environment variables
#               that will be available for the next steps of the job.
# Arguments:
#   $1: GKE_CLUSTER
#   $2: GKE_RELEASE_CHANNEL
#   $3: GITHUB_REF_NAME
#   $4: TEST_LATEST_RELEASE
# Returns: None
########################################################################################################################
function exportEscapedGKEClusterName() {
  GKE_CLUSTER=${1:?GKE_CLUSTER not provided}
  GKE_RELEASE_CHANNEL=${2:?GKE_RELEASE_CHANNEL not provided}
  GITHUB_REF_NAME=${3:?GITHUB_REF_NAME not provided}
  TEST_LATEST_RELEASE=${4:?TEST_LATEST_RELEASE not provided}
  # Max len for a GKE cluster is 40 characters at the time of this writing
  local MAX_LENGTH=40 LATEST_RELEASE=0 len offset

  info "Exporting scaped GKE cluster name"
  [[ "${TEST_LATEST_RELEASE}" == "true" ]] && LATEST_RELEASE=1
  ESCAPED_GKE_CLUSTER=$(echo "${GKE_CLUSTER}-${GITHUB_REF_NAME}-${LATEST_RELEASE}-${GKE_RELEASE_CHANNEL}-ci" | sed 's/[^a-z0-9-]//g')

  # In case the name exceeds the max length allowed, we take a substring of MAX_LENGTH chars from the beginning to avoid
  # the "-gha" suffix
  len=${#ESCAPED_GKE_CLUSTER}
  if ((len > MAX_LENGTH)); then
    offset=$((len - MAX_LENGTH))
    ESCAPED_GKE_CLUSTER="${ESCAPED_GKE_CLUSTER:offset:MAX_LENGTH}"
  fi

  export ESCAPED_GKE_CLUSTER
  # Just exporting the env var won't make it available for the next steps in the GHA's job, so we need the line below
  echo "ESCAPED_GKE_CLUSTER=${ESCAPED_GKE_CLUSTER}" >> "${GITHUB_ENV}"
  info "Done"
}

########################################################################################################################
# Delete a GKE cluster
# Globals: None
# Arguments:
#   $1: GKE_ZONE
#   $2: The escaped GKE cluster's name.
# Returns: None
########################################################################################################################
function deleteGKECluster() {
  GKE_ZONE=${1:?GKE_ZONE not provided}
  CLUSTER_NAME=${2:?CLUSTER_NAME not provided}

  info "Deleting GKE cluster: ${CLUSTER_NAME}"
  gcloud container clusters delete --quiet --async --zone "${GKE_ZONE}" "${CLUSTER_NAME}"
  info "Done"
}
