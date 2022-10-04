#!/usr/bin/env bash

set -euo pipefile
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"

. "${ROOT_DIR}/script/lib/liblog.sh"

########################
# Install GitHub CLI
# Globals: None
# Arguments:
#   $1: Version of GitHub
# Returns: None
#########################
function installGithubCLI() {
  GITHUB_VERSION=${1:?GitHub version not provided}

  info "Installing GitHub CLI $GITHUB_VERSION"
  pusd /tmp
    wget "https://github.com/cli/cli/releases/download/v${GITHUB_VERSION}/gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    tar zxf "gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    rm "gh_${GITHUB_VERSION}_linux_amd64.tar.gz"
    sudo mv "gh_${GITHUB_VERSION}_linux_amd64/bin/gh" /usr/local/bin/
  popd
  info "Done"
}

########################
# Install Semver
# Globals: None
# Arguments:
#   $1: Version of Semver
# Returns: None
#########################
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

########################
# Install GPG key
# Globals: None
# Arguments:
#   $1: GPG public key
#   $2: GPG private key
#   $1: CI BOT GPG
#   $1: CI BOT EMAIL
# Returns: None
#########################
function installGPGKey() {
  GPG_KEY_PUBLIC=${1:?GPG public key not provided}
  GPG_KEY_PRIVATE=${2:?GPG private key not provided}
  CI_BOT_GPG=${3:?CI BOT GPG key not provided}
  CI_BOT_EMAIL=${4:?CI BOT EMAIL not provided}

  info "Installing the GPG KEY"
  # Creating the files from the GPG_KEY_PUBLIC and GPG_KEY_PRIVATE env vars
  echo -e "${GPG_KEY_PUBLIC}" > /tmp/public.key
  echo -e "${GPG_KEY_PRIVATE}" > /tmp/private.key

  # Importing the GPG keys
  gpg --import /tmp/public.key
  gpg --import --no-tty --batch --yes /tmp/private.key

  # Trusting the imported GPG private key
  (echo 5; echo y; echo save) |  gpg --command-fd 0 --no-tty --no-greeting -q --edit-key "${CI_BOT_GPG}" trust

  # Listing the key to verify the import process succeeded
  gpg --list-secret-keys ${CI_BOT_EMAIL}
  info "Done"
}

########################
# Install Kind
# Globals: None
# Arguments:
#   $1: Version of Kind
# Returns: None
#########################
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

########################
# Install Kubectl
# Globals: None
# Arguments:
#   $1: Version of kubectl
# Returns: None
#########################
function installKubectl() {
  KUBECTL_VERSION=${1:?kubectl version not provided}

  info "Installing Kubectl ${KUBECTL_VERSION}"
  pushd /tmp
    curl -LO https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
    chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin/kubectl
  popd
  info "Done"
}

########################
# Install Mkcert
# Globals: None
# Arguments:
#   $1: Version of mkcert
# Returns: None
#########################
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

########################
# Install Helm
# Globals: None
# Arguments:
#   $1: Version of Helm
#   $2: Helm binary name (OPTIONAL, allows to install different versions of Helm)
# Returns: None
#########################
function installHelm() {
  HELM_VERSION=${1:?Helm version not provided}
  HELM_BINARY_NAME=${2:-helm}

  info "Installing Helm ${HELM_VERSION} as ${HELM_BINARY_NAME}"
  pusd /tmp
    wget "https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz"
    tar zxf "helm-$HELM_VERSION-linux-amd64.tar.gz"
    sudo mv linux-amd64/helm "/usr/local/bin/${HELM_BINARY_NAME}"
  popd
  info "Done"
}
