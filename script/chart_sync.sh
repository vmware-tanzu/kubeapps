#!/bin/bash
# Copyright (c) 2018 Bitnami
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

set -e

CHARTS_REPO="bitnami/charts"
CHART_REPO_PATH="bitnami/kubeapps"
PROJECT_DIR=`cd "$( dirname "${BASH_SOURCE[0]}" )/.." >/dev/null && pwd`
KUBEAPPS_CHART_DIR="${PROJECT_DIR}/chart/kubeapps"

changedVersion() {
    local currentVersion=$(cat "${KUBEAPPS_CHART_DIR}/Chart.yaml" | grep "version:")
    local externalVersion=$(curl -s https://raw.githubusercontent.com/${CHARTS_REPO}/master/${CHART_REPO_PATH}/Chart.yaml | grep "version:")
    # NOTE: If curl returns an error this will return always true
    [[ "$currentVersion" != "$externalVersion" ]]
}

configUser() {
    local targetRepo=${1:?}
    local user=${2:?}
    local email=${3:?}
    cd $targetRepo
    git config user.name "$user"
    git config user.email "$email"
    cd -
}

updateRepo() {
    local targetRepo=${1:?}
    local targetTag=${2:?}
    if [ ! -f "${targetRepo}/${CHART_REPO_PATH}/Chart.yaml" ]; then
        echo "Wrong repo path. You should provide the root of the repository" > /dev/stderr
        return 1
    fi
    rm -rf "${targetRepo}/${CHART_REPO_PATH}"
    cp -R "${KUBEAPPS_CHART_DIR}" "${targetRepo}/${CHART_REPO_PATH}"
    # Update Chart.yaml with new version
    sed -i.bk 's/appVersion: DEVEL/appVersion: '"${targetTag}"'/g' "${targetRepo}/${CHART_REPO_PATH}/Chart.yaml"
    rm "${targetRepo}/${CHART_REPO_PATH}/Chart.yaml.bk"
    # DANGER: This replaces any tag marked as latest in the values.yaml
    sed -i.bk 's/tag: latest/tag: '"${targetTag}"'/g' "${targetRepo}/${CHART_REPO_PATH}/values.yaml"
    rm "${targetRepo}/${CHART_REPO_PATH}/values.yaml.bk"
}

commitAndPushChanges() {
    local targetRepo=${1:?}
    local targetBranch=${2:-"master"}
    local targetVersion=${3:?}
    if [ ! -f "${targetRepo}/${CHART_REPO_PATH}/Chart.yaml" ]; then
        echo "Wrong repo path. You should provide the root of the repository" > /dev/stderr
        return 1
    fi
    cd $targetRepo
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" > /dev/stderr
        cd -
        return 1
    fi
    git add --all .
    git commit -m "kubeapps: bump chart version to $targetVersion"
    # NOTE: This expects to have a loaded SSH key
    git push origin $targetBranch
    cd -
}

user=${1:?}
email=${2:?}
if changedVersion; then
    tempDir=$(mktemp -u)/charts
    mkdir -p $tempDir
    git clone https://github.com/${CHARTS_REPO} $tempDir
    configUser $tempDir $user $email
    git fetch --tags
    latestVersion=$(git describe --tags $(git rev-list --tags --max-count=1))
    updateRepo $tempDir $latestVersion
    commitAndPushChanges $tempDir master $latestVersion
else
    echo "Skipping Chart sync. The version has not changed"
fi
