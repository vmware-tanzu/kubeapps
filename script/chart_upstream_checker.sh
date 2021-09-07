#!/bin/bash
# Copyright (c) 2021 Bitnami
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

user=${1:?}
email=${2:?}
gpg=${3:?}
forkSSHKeyFilename=${4:?}

currentVersion=$(cat "${KUBEAPPS_CHART_DIR}/Chart.yaml" | grep -oP '(?<=^version: ).*')
externalVersion=$(curl -s https://raw.githubusercontent.com/${CHARTS_REPO_ORIGINAL}/master/${CHART_REPO_PATH}/Chart.yaml | grep -oP '(?<=^version: ).*')
semverCompare=$(semver compare "${currentVersion}" "${externalVersion}")
# If current version is less than the chart external version, then retrieve the changes and send an internal PR with them
if [[ ${semverCompare} -lt 0 ]]; then
    echo "Current chart version ("${currentVersion}") is less than the chart external version ("${externalVersion}")"
    tempDir=$(mktemp -u)/charts
    mkdir -p $tempDir
    git clone https://github.com/${CHARTS_REPO} $tempDir --depth 1 --no-single-branch
    configUser $tempDir $user $email $gpg
    configUser $PROJECT_DIR $user $email $gpg
    latestVersion=$(latestReleaseTag $PROJECT_DIR)
    updateRepoWithRemoteChanges $tempDir $latestVersion $forkSSHKeyFilename
    commitAndSendInternalPR ${PROJECT_DIR} "sync-chart-changes-${externalVersion}" ${externalVersion}
elif [[ ${semverCompare} -gt 0 ]]; then
    echo "Skipping Chart sync. WARNING Current chart version ("${currentVersion}") is greater than the chart external version ("${externalVersion}")"
else
    echo "Skipping Chart sync. The chart version ("${currentVersion}") has not changed"
fi
