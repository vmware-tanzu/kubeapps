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
CHARTS_REPO_UPSTREAM=${5:?Missing base chart repository}
CHARTS_REPO_UPSTREAM_BRANCH=${6:?Missing base chart repository}
CHARTS_REPO_FORK=${7:?Missing forked chart repository}
CHARTS_REPO_FORK_BRANCH=${8:?Missing forked chart repository}
KUBEAPPS_REPO_UPSTREAM=${9:?Missing kubeapps repository}
KUBEAPPS_REPO_UPSTREAM_BRANCH=${10:?Missing kubeapps repository branch}
README_GENERATOR_REPO=${11:?Missing readme generator repository}
LOCAL_KUBEAPPS_REPO_PATH=${PROJECT_DIR:?PROJECT_DIR not defined}

info "LOCAL_KUBEAPPS_REPO_PATH: ${LOCAL_KUBEAPPS_REPO_PATH}"
info "USERNAME: ${USERNAME}"
info "EMAIL: ${EMAIL}"
info "GPG_KEY: ${GPG_KEY}"
info "FORKED_SSH_KEY_FILENAME: ${FORKED_SSH_KEY_FILENAME}"
info "CHARTS_REPO_UPSTREAM: ${CHARTS_REPO_UPSTREAM}"
info "CHARTS_REPO_UPSTREAM_BRANCH: ${CHARTS_REPO_UPSTREAM_BRANCH}"
info "CHARTS_REPO_FORK: ${CHARTS_REPO_FORK}"
info "CHARTS_REPO_FORK_BRANCH: ${CHARTS_REPO_FORK_BRANCH}"
info "KUBEAPPS_REPO_UPSTREAM: ${KUBEAPPS_REPO_UPSTREAM}"
info "KUBEAPPS_REPO_UPSTREAM_BRANCH: ${KUBEAPPS_REPO_UPSTREAM_BRANCH}"
info "README_GENERATOR_REPO: ${README_GENERATOR_REPO}"

currentVersion=$(grep -oP '(?<=^version: ).*' <"${KUBEAPPS_CHART_DIR}/Chart.yaml")
externalVersion=$(curl -s "https://raw.githubusercontent.com/${CHARTS_REPO_UPSTREAM}/${CHARTS_REPO_UPSTREAM_BRANCH}/${CHART_REPO_PATH}/Chart.yaml" | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")

info "currentVersion: ${currentVersion}"
info "externalVersion: ${externalVersion}"

# If current version is less than the chart external version, then retrieve the changes and send an internal PR with them
if [[ ${semverCompare} -lt 0 ]]; then
    echo "Current chart version (${currentVersion}) is less than the chart external version (${externalVersion})"
    LOCAL_CHARTS_REPO_FORK=$(mktemp -u)/charts
    mkdir -p "${LOCAL_CHARTS_REPO_FORK}"

    GIT_SSH_COMMAND="ssh -i ~/.ssh/${FORKED_SSH_KEY_FILENAME}" git clone "git@github.com:${CHARTS_REPO_FORK}" "${LOCAL_CHARTS_REPO_FORK}" --depth 1 --no-single-branch
    configUser "${LOCAL_CHARTS_REPO_FORK}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"
    configUser "${LOCAL_KUBEAPPS_REPO_PATH}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"

    latestVersion=$(latestReleaseTag "${LOCAL_KUBEAPPS_REPO_PATH}")
    prBranchName="sync-chart-changes-${externalVersion}"

    updateRepoWithRemoteChanges "${LOCAL_CHARTS_REPO_FORK}" "${latestVersion}" "${FORKED_SSH_KEY_FILENAME}" "${CHARTS_REPO_UPSTREAM}" "${CHARTS_REPO_UPSTREAM_BRANCH}" "${CHARTS_REPO_FORK_BRANCH}"
    generateReadme "${README_GENERATOR_REPO}" "${KUBEAPPS_CHART_DIR}"
    commitAndSendInternalPR "${LOCAL_KUBEAPPS_REPO_PATH}" "${prBranchName}" "${externalVersion}" "${KUBEAPPS_REPO_UPSTREAM}" "${KUBEAPPS_REPO_UPSTREAM_BRANCH}"
elif [[ ${semverCompare} -gt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version (${currentVersion}) is greater than the chart external version (${externalVersion})"
else
    echo "Skipping Chart sync. The chart version (${currentVersion}) has not changed"
fi
