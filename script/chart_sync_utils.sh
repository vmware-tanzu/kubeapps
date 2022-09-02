#!/bin/bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)

# Path of the Kubeapps chart in the charts repo.
# For instance, given "https://github.com/bitnami/charts/tree/master/bitnami/kubeapps" it should be "bitnami/kubeapps"
CHART_REPO_PATH="bitnami/kubeapps"

# Path of the Kubeapps chart in the Kubeapps repo.
# For instance, given "https://github.com/vmware-tanzu/kubeapps/tree/main/chart/kubeapps", it should be "chart/kubeapps"
KUBEAPPS_CHART_DIR="${PROJECT_DIR}/chart/kubeapps"

# Paths of the templates files, note they are also used elsewhere
PR_INTERNAL_TEMPLATE_FILE="${PROJECT_DIR}/script/tpl/PR_internal_chart_template.md"
PR_EXTERNAL_TEMPLATE_FILE="${PROJECT_DIR}/script/tpl/PR_external_chart_template.md"
RELEASE_NOTES_TEMPLATE_FILE="${PROJECT_DIR}/script/tpl/release_notes.md"

# Returns the tag for the latest release
latestReleaseTag() {
    local TARGET_REPO=${1:?}

    git -C "${TARGET_REPO}/.git" fetch --tags
    git -C "${TARGET_REPO}/.git" describe --tags "$(git rev-list --tags --max-count=1)"
}

configUser() {
    local TARGET_REPO=${1:?}
    local USERNAME=${2:?}
    local EMAIL=${3:?}
    local GPG_KEY=${4:?}

    cd "${TARGET_REPO}"
    git config user.name "${USERNAME}"
    git config user.email "${EMAIL}"
    git config user.signingkey "${GPG_KEY}"
    git config --global commit.gpgSign true
    git config --global tag.gpgSign true
    git config pull.rebase false
    cd -
}

replaceImage_latestToProduction() {
    local SERVICE=${1:?}
    local FILE=${2:?}

    local repoName="bitnami-docker-kubeapps-${SERVICE}"
    local currentImageEscaped="kubeapps\/${SERVICE}"
    local targetImageEscaped="bitnami\/kubeapps-${SERVICE}"

    # Prevent a wrong image name "bitnami/kubeapps-kubeapps-apis"
    # with a manual rename to "bitnami/kubeapps-apis"
    if [ "${targetImageEscaped}" == "bitnami\/kubeapps-kubeapps-apis" ]; then
        targetImageEscaped="bitnami\/kubeapps-apis"
    fi

    if [ "${repoName}" == "bitnami-docker-kubeapps-kubeapps-apis" ]; then
        repoName="bitnami-docker-kubeapps-apis"
    fi

    echo "Replacing ${SERVICE}"...

    local curl_opts=()
    if [[ $GITHUB_TOKEN != "" ]]; then
        curl_opts=(-s -H "Authorization: token ${GITHUB_TOKEN}")
    fi

    # Get the latest tag from the bitnami repository
    local tag=$(curl "${curl_opts[@]}" "https://api.github.com/repos/bitnami/${repoName}/tags" | jq -r '.[0].name')

    if [[ $tag == "" ]]; then
        echo "ERROR: Unable to obtain latest tag for ${repoName}. Stopping..."
        exit 1
    fi

    # Replace image and tag from the values.yaml
    sed -i.bk -e '1h;2,$H;$!d;g' -re \
    's/repository: '${currentImageEscaped}'\n    tag: latest/repository: '${targetImageEscaped}'\n    tag: '${tag}'/g' \
    "${FILE}"
    rm "${FILE}.bk"
}

replaceImage_productionToLatest() {
    local SERVICE=${1:?}
    local FILE=${2:?}

    local repoName="bitnami-docker-kubeapps-${SERVICE}"
    local currentImageEscaped="bitnami\/kubeapps-${SERVICE}"
    local targetImageEscaped="kubeapps\/${SERVICE}"

    # Prevent a wrong image name "bitnami/kubeapps-kubeapps-apis"
    # with a manual rename to "bitnami/kubeapps-apis"
    if [ "${currentImageEscaped}" == "bitnami\/kubeapps-kubeapps-apis" ]; then
        currentImageEscaped="bitnami\/kubeapps-apis"
    fi

    echo "Replacing ${SERVICE}"...

    # Replace image and tag from the values.yaml
    sed -i.bk -e '1h;2,$H;$!d;g' -re \
    's/repository: '${currentImageEscaped}'\n    tag: \S*/repository: '${targetImageEscaped}'\n    tag: latest/g' \
    "${FILE}"
    rm "${FILE}.bk"
}

updateRepoWithLocalChanges() {
    local TARGET_REPO=${1:?}
    local TARGET_TAG=${2:?}
    local CHARTS_REPO_ORIGINAL=${3:?}
    local BRANCH_CHARTS_REPO_ORIGINAL=${4:?}
    local BRANCH_CHARTS_REPO_FORKED=${5:?}

    local targetTagWithoutV=${TARGET_TAG#v}
    local targetChartPath="${TARGET_REPO}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"

    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    # Fetch latest upstream changes, and commit&push them to the forked charts repo
    git -C "${TARGET_REPO}" remote add upstream "https://github.com/${CHARTS_REPO_ORIGINAL}.git"
    git -C "${TARGET_REPO}" pull upstream "${BRANCH_CHARTS_REPO_ORIGINAL}"
    git -C "${TARGET_REPO}" push origin "${BRANCH_CHARTS_REPO_FORKED}"
    rm -rf "${targetChartPath}"
    cp -R "${KUBEAPPS_CHART_DIR}" "${targetChartPath}"
    # Update Chart.yaml with new version
    sed -i.bk 's/appVersion: DEVEL/appVersion: '"${targetTagWithoutV}"'/g' "${chartYaml}"
    rm "${targetChartPath}/Chart.yaml.bk"
    # Replace images for the latest available
    # TODO: use the IMAGES_TO_PUSH var already set in the CI config
    replaceImage_latestToProduction dashboard "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction apprepository-controller "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction asset-syncer "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction pinniped-proxy "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction kubeapps-apis "${targetChartPath}/values.yaml"
}

updateRepoWithRemoteChanges() {
    local TARGET_REPO=${1:?}
    local TARGET_TAG=${2:?}
    local FORKED_SSH_KEY_FILENAME=${3:?}
    local CHARTS_REPO_ORIGINAL=${4:?}
    local BRANCH_CHARTS_REPO_ORIGINAL=${5:?}
    local BRANCH_CHARTS_REPO_FORKED=${6:?}

    local targetTagWithoutV=${TARGET_TAG#v}
    local targetChartPath="${TARGET_REPO}/${CHART_REPO_PATH}"
    local remoteChartYaml="${targetChartPath}/Chart.yaml"
    local localChartYaml="${KUBEAPPS_CHART_DIR}/Chart.yaml"

    if [ ! -f "${remoteChartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    # Fetch latest upstream changes, and commit&push them to the forked charts repo
    git -C "${TARGET_REPO}" remote add upstream "https://github.com/${CHARTS_REPO_ORIGINAL}.git"
    git -C "${TARGET_REPO}" pull upstream "${BRANCH_CHARTS_REPO_ORIGINAL}"

    # https://superuser.com/questions/232373/how-to-tell-git-which-private-key-to-use
    GIT_SSH_COMMAND="ssh -i ~/.ssh/${FORKED_SSH_KEY_FILENAME}" git -C "${TARGET_REPO}" push origin "${BRANCH_CHARTS_REPO_FORKED}"

    rm -rf "${KUBEAPPS_CHART_DIR}"
    cp -R "${targetChartPath}" "${KUBEAPPS_CHART_DIR}"
    # Update Chart.yaml with new version
    sed -i.bk "s/appVersion: "${targetTagWithoutV}"/appVersion: DEVEL/g" "${localChartYaml}"
    rm "${KUBEAPPS_CHART_DIR}/Chart.yaml.bk"
    # Replace images for the latest available
    # TODO: use the IMAGES_TO_PUSH var already set in the CI config
    replaceImage_productionToLatest dashboard "${KUBEAPPS_CHART_DIR}/values.yaml"
    replaceImage_productionToLatest apprepository-controller "${KUBEAPPS_CHART_DIR}/values.yaml"
    replaceImage_productionToLatest asset-syncer "${KUBEAPPS_CHART_DIR}/values.yaml"
    replaceImage_productionToLatest pinniped-proxy "${KUBEAPPS_CHART_DIR}/values.yaml"
    replaceImage_productionToLatest kubeapps-apis "${KUBEAPPS_CHART_DIR}/values.yaml"
}

generateReadme() {
    local README_GENERATOR_REPO=${1:?}
    local CHART_PATH=${2:?}

    TMP_DIR=$(mktemp -u)/readme
    local chartReadmePath="${CHART_PATH}/README.md"
    local chartValuesPath="${CHART_PATH}/values.yaml"

    git clone "https://github.com/${README_GENERATOR_REPO}" "${TMP_DIR}" --depth 1 --no-single-branch

    cd "${TMP_DIR}"
    npm install --production
    node bin/index.js -r "${chartReadmePath}" -v "${chartValuesPath}"
}

commitAndSendExternalPR() {
    local TARGET_REPO=${1:?}
    local TARGET_BRANCH=${2:?}
    local CHART_VERSION=${3:?}
    local CHARTS_REPO_ORIGINAL=${4:?}
    local BRANCH_CHARTS_REPO_ORIGINAL=${5:?}

    local targetChartPath="${TARGET_REPO}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"

    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    cd "${TARGET_REPO}"
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" >/dev/stderr
        cd -
        return 1
    fi
    sed -i.bk -e "s/<USER>/$(git config user.name)/g" "${PR_EXTERNAL_TEMPLATE_FILE}"
    sed -i.bk -e "s/<EMAIL>/$(git config user.email)/g" "${PR_EXTERNAL_TEMPLATE_FILE}"
    git checkout -b "${TARGET_BRANCH}"
    git add --all .
    git commit --signoff -m "kubeapps: bump chart version to ${CHART_VERSION}"
    # NOTE: This expects to have a loaded SSH key
    if [[ $(git ls-remote origin "${TARGET_BRANCH}" | wc -l) -eq 0 ]]; then
        git push -u origin "${TARGET_BRANCH}"
        gh pr create -d -B "${BRANCH_CHARTS_REPO_ORIGINAL}" -R "${CHARTS_REPO_ORIGINAL}" -F "${PR_EXTERNAL_TEMPLATE_FILE}" --title "[bitnami/kubeapps] Bump chart version to ${CHART_VERSION}"
    else
        echo "The remote branch '${TARGET_BRANCH}' already exists, please check if there is already an open PR at the repository '${CHARTS_REPO_ORIGINAL}'"
    fi
    cd -
}

commitAndSendInternalPR() {
    local TARGET_REPO=${1:?}
    local TARGET_BRANCH=${2:?}
    local CHART_VERSION=${3:?}
    local KUBEAPPS_REPO=${4:?}
    local BRANCH_KUBEAPPS_REPO=${5:?}

    local targetChartPath="${KUBEAPPS_CHART_DIR}/Chart.yaml"
    local localChartYaml="${KUBEAPPS_CHART_DIR}/Chart.yaml"

    if [ ! -f "${localChartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    cd "${TARGET_REPO}"
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" >/dev/stderr
        cd -
        return 1
    fi
    git checkout -b "${TARGET_BRANCH}"
    git add --all .
    git commit --signoff -m "bump chart version to ${CHART_VERSION}"
    # NOTE: This expects to have a loaded SSH key
    if [[ $(git ls-remote origin "${TARGET_BRANCH}" | wc -l) -eq 0 ]]; then
        git push -u origin "${TARGET_BRANCH}"
        gh pr create -d -B "${BRANCH_KUBEAPPS_REPO}" -R "${KUBEAPPS_REPO}" -F "${PR_INTERNAL_TEMPLATE_FILE}" --title "Sync chart with bitnami/kubeapps chart (version ${CHART_VERSION})"
    else
        echo "The remote branch '${TARGET_BRANCH}' already exists, please check if there is already an open PR at the repository '${KUBEAPPS_REPO}'"
    fi
    cd -
}
