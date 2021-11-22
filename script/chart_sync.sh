#!/bin/bash

# Copyright 2018-2021 VMware. All Rights Reserved.
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
set -o nounset
set -o pipefail

source $(dirname $0)/chart_sync_utils.sh

USERNAME=${1:?Missing git username}
EMAIL=${2:?Missing git email}
GPG_KEY=${3:?Missing git gpg key}
CHARTS_REPO_ORIGINAL=${4:?Missing base chart repository}
BRANCH_CHARTS_REPO_ORIGINAL=${5:?Missing base chart repository branch}
CHARTS_REPO_FORKED=${6:?Missing forked chart repository}
BRANCH_CHARTS_REPO_FORKED=${7:?Missing forked chart repository branch}

currentVersion=$(grep -oP '(?<=^version: ).*' <"${KUBEAPPS_CHART_DIR}/Chart.yaml")
externalVersion=$(curl -s "https://raw.githubusercontent.com/${CHARTS_REPO_ORIGINAL}/${BRANCH_CHARTS_REPO_ORIGINAL}/${CHART_REPO_PATH}/Chart.yaml" | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")

# If current version is greater than the chart external version, then send a PR bumping up the version externally
if [[ ${semverCompare} -gt 0 ]]; then
    echo "Current chart version (${currentVersion}) is greater than the chart external version (${externalVersion})"
    TMP_DIR=$(mktemp -u)/charts
    mkdir -p "${TMP_DIR}"

    git clone "https://github.com/${CHARTS_REPO_FORKED}" "${TMP_DIR}" --depth 1 --no-single-branch
    configUser "${TMP_DIR}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"
    configUser "${PROJECT_DIR}" "${USERNAME}" "${EMAIL}" "${GPG_KEY}"

    latestVersion=$(latestReleaseTag "${PROJECT_DIR}")
    prBranchName="kubeapps-bump-${currentVersion}"

    updateRepoWithLocalChanges "${TMP_DIR}" "${latestVersion}" "${CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_FORKED}"
    commitAndSendExternalPR "${TMP_DIR}" "${prBranchName}" "${currentVersion}" "${CHARTS_REPO_ORIGINAL}" "${BRANCH_CHARTS_REPO_ORIGINAL}"
elif [[ ${semverCompare} -lt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version (${currentVersion}) is less than the chart external version (${externalVersion})"
else
    echo "Skipping Chart sync. The chart version (${currentVersion}) has not changed"
fi
