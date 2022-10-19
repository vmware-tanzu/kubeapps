#!/bin/bash

# Copyright 2018-2022 the Kubeapps contributors.
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
DEV_MODE=${9:-false}

info "USERNAME: ${USERNAME}"
info "EMAIL: ${EMAIL}"
info "GPG_KEY: ${GPG_KEY}"
info "FORKED_SSH_KEY_FILENAME: ${FORKED_SSH_KEY_FILENAME}"
info "CHARTS_REPO_ORIGINAL: ${CHARTS_REPO_ORIGINAL}"
info "BRANCH_CHARTS_REPO_ORIGINAL: ${BRANCH_CHARTS_REPO_ORIGINAL}"
info "CHARTS_REPO_FORKED: ${CHARTS_REPO_FORKED}"
info "BRANCH_CHARTS_REPO_FORKED: ${BRANCH_CHARTS_REPO_FORKED}"
info "DEV_MODE: ${DEV_MODE}"

if [[ "${DEV_MODE}" == "true" ]]; then
  set -x
fi

currentVersion=$(grep -oP '(?<=^version: ).*' <"${KUBEAPPS_CHART_DIR}/Chart.yaml")
externalVersion=$(curl -s "https://raw.githubusercontent.com/${CHARTS_REPO_ORIGINAL}/${BRANCH_CHARTS_REPO_ORIGINAL}/${CHART_REPO_PATH}/Chart.yaml" | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")

info "currentVersion: ${currentVersion}"
info "externalVersion: ${externalVersion}"


# If current version is greater than the chart external version, then send a PR bumping up the version externally
if [[ ${semverCompare} -gt 0 ]]; then
    echo "Current chart version (${currentVersion}) is greater than the chart external version (${externalVersion})"
    CHARTS_FORK_LOCAL_PATH=$(mktemp -u)/charts
    mkdir -p "${CHARTS_FORK_LOCAL_PATH}"

#    git clone "https://github.com/${CHARTS_REPO_FORKED}" "${CHARTS_FORK_LOCAL_PATH}" --depth 1 --no-single-branch
    GIT_SSH_COMMAND="ssh -i ~/.ssh/${FORKED_SSH_KEY_FILENAME}" git clone "git@github.com:${CHARTS_REPO_FORKED}" "${CHARTS_FORK_LOCAL_PATH}" --depth 1 --no-single-branch
    configUser "${CHARTS_FORK_LOCAL_PATH}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"
    configUser "${PROJECT_DIR}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"

    latestVersion=$(latestReleaseTag "${PROJECT_DIR}")
    prBranchName="kubeapps-bump-${currentVersion}"

    updateRepoWithLocalChanges "${CHARTS_FORK_LOCAL_PATH}" "${latestVersion}" "${FORKED_SSH_KEY_FILENAME}" "${CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_FORKED}"
    commitAndSendExternalPR "${CHARTS_FORK_LOCAL_PATH}" "${prBranchName}" "${currentVersion}" "${CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_ORIGINAL}" "${DEV_MODE}"
elif [[ ${semverCompare} -lt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version (${currentVersion}) is less than the chart external version (${externalVersion})"
else
    echo "Skipping Chart sync. The chart version (${currentVersion}) has not changed"
fi
