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

# Returns the tag for the latest release
latestReleaseTag() {
  git describe --tags $(git rev-list --tags --max-count=1)
}

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
    local targetChartPath="${targetRepo}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"
    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" > /dev/stderr
        return 1
    fi
    rm -rf "${targetChartPath}"
    cp -R "${KUBEAPPS_CHART_DIR}" "${targetChartPath}"
    # Update Chart.yaml with new version
    sed -i.bk 's/appVersion: DEVEL/appVersion: '"${targetTag}"'/g' "${chartYaml}"
    rm "${targetChartPath}/Chart.yaml.bk"
    # DANGER: This replaces any tag marked as latest in the values.yaml
    sed -i.bk 's/tag: latest/tag: '"${targetTag}"'/g' "${targetChartPath}/values.yaml"
    rm "${targetChartPath}/values.yaml.bk"
}

commitAndPushChanges() {
    local targetRepo=${1:?}
    local targetBranch=${2:-"master"}
    local targetChartPath="${targetRepo}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"
    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" > /dev/stderr
        return 1
    fi
    cd $targetRepo
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" > /dev/stderr
        cd -
        return 1
    fi
    local chartVersion=$(grep -w version: ${chartYaml} | awk '{print $2}')
    git add --all .
    git commit -m "kubeapps: bump chart version to $chartVersion"
    # NOTE: This expects to have a loaded SSH key
    git push origin $targetBranch
    cd -
}
