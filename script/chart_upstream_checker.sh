#!/bin/bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT_DIR/script/lib/liblog.sh"
source "$ROOT_DIR/script/chart_sync_utils.sh"

USERNAME=${1:?Missing git username}
EMAIL=${2:?Missing git email}
GPG_KEY=${3:?Missing git gpg key}
FORKED_SSH_KEY_FILENAME=${4:?Missing forked ssh key filename}
CHARTS_REPO_ORIGINAL=${5:?Missing base chart repository}
BRANCH_CHARTS_REPO_ORIGINAL=${6:?Missing base chart repository}
CHARTS_REPO_FORKED=${7:?Missing forked chart repository}
BRANCH_CHARTS_REPO_FORKED=${8:?Missing forked chart repository}
KUBEAPPS_REPO=${9:?Missing kubeapps repository}
BRANCH_KUBEAPPS_REPO=${10:?Missing kubeapps repository branch}
README_GENERATOR_REPO=${11:?Missing readme generator repository}
DEV_MODE=${12:-false}

info "USERNAME: ${USERNAME}"
info "EMAIL: ${EMAIL}"
info "GPG_KEY: ${GPG_KEY}"
info "FORKED_SSH_KEY_FILENAME: ${FORKED_SSH_KEY_FILENAME}"
info "CHARTS_REPO_ORIGINAL: ${CHARTS_REPO_ORIGINAL}"
info "BRANCH_CHARTS_REPO_ORIGINAL: ${BRANCH_CHARTS_REPO_ORIGINAL}"
info "CHARTS_REPO_FORKED: ${CHARTS_REPO_FORKED}"
info "BRANCH_CHARTS_REPO_FORKED: ${BRANCH_CHARTS_REPO_FORKED}"
info "KUBEAPPS_REPO: ${KUBEAPPS_REPO}"
info "BRANCH_KUBEAPPS_REPO: ${BRANCH_KUBEAPPS_REPO}"
info "README_GENERATOR_REPO: ${README_GENERATOR_REPO}"
info "DEV_MODE: ${DEV_MODE}"

if [[ "${DEV_MODE}" == "true" ]]; then
  set -x
else
  echo "Not running in DEV_MODE"
  exit 1
fi

currentVersion=$(grep -oP '(?<=^version: ).*' <"${KUBEAPPS_CHART_DIR}/Chart.yaml")
externalVersion=$(curl -s "https://raw.githubusercontent.com/${CHARTS_REPO_ORIGINAL}/${BRANCH_CHARTS_REPO_ORIGINAL}/${CHART_REPO_PATH}/Chart.yaml" | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")

info "currentVersion: ${currentVersion}"
info "externalVersion: ${externalVersion}"

# If current version is less than the chart external version, then retrieve the changes and send an internal PR with them
if [[ ${semverCompare} -lt 0 ]]; then
    echo "Current chart version (${currentVersion}) is less than the chart external version (${externalVersion})"
    TMP_DIR=$(mktemp -u)/charts
    mkdir -p "${TMP_DIR}"

    git clone "https://github.com/${CHARTS_REPO_FORKED}" "${TMP_DIR}" --depth 1 --no-single-branch
    configUser "${TMP_DIR}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"
    configUser "${PROJECT_DIR}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"

    latestVersion=$(latestReleaseTag "${PROJECT_DIR}")
    prBranchName="sync-chart-changes-${externalVersion}"

    if [[ "${DEV_MODE}" == "true" ]]; then
      prBranchName="${prBranchName}-DEV"
    fi

    updateRepoWithRemoteChanges "${TMP_DIR}" "${latestVersion}" "${FORKED_SSH_KEY_FILENAME}" "${CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_FORKED}"
    generateReadme "${README_GENERATOR_REPO}" "${KUBEAPPS_CHART_DIR}"
    commitAndSendInternalPR "${PROJECT_DIR}" "${prBranchName}" "${externalVersion}" "${KUBEAPPS_REPO}" "${BRANCH_KUBEAPPS_REPO}" "${DEV_MODE}"
elif [[ ${semverCompare} -gt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version (${currentVersion}) is greater than the chart external version (${externalVersion})"
else
    echo "Skipping Chart sync. The chart version (${currentVersion}) has not changed"
fi
