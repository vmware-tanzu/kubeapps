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
CHARTS_FORK_SSH_KEY_FILENAME=${4:?Missing forked ssh key filename}
CHARTS_REPO_UPSTREAM=${5:?Missing chart repository upstream (eg. bitnami/charts)}
CHARTS_REPO_UPSTREAM_BRANCH=${6:?Missing chart repository upstream\'s  main branch name (eg. main)}
CHARTS_REPO_FORK=${7:?Missing chart repository fork (eg. kubeapps-bot/charts)}
CHARTS_REPO_FORK_BRANCH=${8:?Missing chart repository fork\'s main branch name (eg. main)}
LOCAL_KUBEAPPS_REPO_PATH=${PROJECT_DIR:?PROJECT_DIR not defined}

info "LOCAL_KUBEAPPS_REPO_PATH: ${LOCAL_KUBEAPPS_REPO_PATH}"
info "USERNAME: ${USERNAME}"
info "EMAIL: ${EMAIL}"
info "GPG_KEY: ${GPG_KEY}"
info "CHARTS_FORK_SSH_KEY_FILENAME: ${CHARTS_FORK_SSH_KEY_FILENAME}"
info "CHARTS_REPO_UPSTREAM: ${CHARTS_REPO_UPSTREAM}"
info "CHARTS_REPO_UPSTREAM_BRANCH: ${CHARTS_REPO_UPSTREAM_BRANCH}"
info "CHARTS_REPO_FORK: ${CHARTS_REPO_FORK}"
info "CHARTS_REPO_FORK_BRANCH: ${CHARTS_REPO_FORK_BRANCH}"

currentVersion=$(grep -oP '(?<=^version: ).*' <"${KUBEAPPS_CHART_DIR}/Chart.yaml")
externalVersion=$(curl -s "https://raw.githubusercontent.com/${CHARTS_REPO_UPSTREAM}/${CHARTS_REPO_UPSTREAM_BRANCH}/${CHART_REPO_PATH}/Chart.yaml" | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")

info "currentVersion: ${currentVersion}"
info "externalVersion: ${externalVersion}"


# If current version is greater than the chart external version, then send a PR bumping up the version externally
if [[ ${semverCompare} -gt 0 ]]; then
    echo "Current chart version (${currentVersion}) is greater than the chart external version (${externalVersion})"
    CHARTS_FORK_LOCAL_PATH=$(mktemp -u)/charts
    mkdir -p "${CHARTS_FORK_LOCAL_PATH}"

    GIT_SSH_COMMAND="ssh -i ~/.ssh/${CHARTS_FORK_SSH_KEY_FILENAME}" git clone "git@github.com:${CHARTS_REPO_FORK}" "${CHARTS_FORK_LOCAL_PATH}" --depth 1 --no-single-branch
    configUser "${CHARTS_FORK_LOCAL_PATH}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"
    configUser "${LOCAL_KUBEAPPS_REPO_PATH}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"

    latestVersion=$(latestReleaseTag "${LOCAL_KUBEAPPS_REPO_PATH}")
    prBranchName="kubeapps-bump-${currentVersion}"

    updateRepoWithLocalChanges "${CHARTS_FORK_LOCAL_PATH}" "${latestVersion}" "${CHARTS_FORK_SSH_KEY_FILENAME}" "${CHARTS_REPO_UPSTREAM}" "${CHARTS_REPO_UPSTREAM_BRANCH}" "${CHARTS_REPO_FORK_BRANCH}"
    commitAndSendExternalPR "${CHARTS_FORK_LOCAL_PATH}" "${prBranchName}" "${currentVersion}" "${CHARTS_REPO_UPSTREAM}" "${CHARTS_REPO_UPSTREAM_BRANCH}" "${CHARTS_FORK_SSH_KEY_FILENAME}"
elif [[ ${semverCompare} -lt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version (${currentVersion}) is less than the chart external version (${externalVersion})"
else
    echo "Skipping Chart sync. The chart version (${currentVersion}) has not changed"
fi
