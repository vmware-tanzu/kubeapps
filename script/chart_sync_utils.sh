#!/bin/bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)
source "${PROJECT_DIR}/script/lib/liblog.sh"

# Path of the Kubeapps chart in the charts repo.
# For instance, given "https://github.com/bitnami/charts/tree/main/bitnami/kubeapps" it should be "bitnami/kubeapps"
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
    info "getting latest release from ${TARGET_REPO}"

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

########################################################################################################################
# Syncs the fork of the charts repo with its upstream, then updates the local copy with the information from the local
# chart stored in the kubeapps repo, and applying to it the new version extracted from the passed TARGET_TAG param.
# Globals:
#   KUBEAPPS_CHART_DIR: Path of the Kubeapps chart in the Kubeapps repo.
#   CHART_REPO_PATH: Path of the Kubeapps chart in the charts repo.
# Arguments:
#   $1 - CHARTS_REPO_FORK_LOCAL_PATH: Path to the clone of the bitnami/charts repo in the local machine.
#   $2 - TARGET_TAG: Tag from which to take the new version for the Helm chart.
#   $3 - CHARTS_FORK_SSH_KEY_FILENAME: Filename of the SSH key to connect with the remote charts repo fork.
#   $4 - CHARTS_REPO_UPSTREAM: Name of the upstream version of the bitnami/charts repo without the GitHub part (eg. bitnami/charts).
#   $5 - CHARTS_REPO_UPSTREAM_BRANCH: Name of the main branch in the upstream of the charts repo.
#   $6 - CHARTS_REPO_FORK_BRANCH: Name of the main branch in the origin remove of the fork of charts repo.
# Returns:
#   0 - Success
#   1 - Failure
########################################################################################################################
updateRepoWithLocalChanges() {
    local CHARTS_REPO_FORK_LOCAL_PATH=${1:?}
    local TARGET_TAG=${2:?}
    local CHARTS_FORK_SSH_KEY_FILENAME=${3:?}
    local CHARTS_REPO_UPSTREAM=${4:?}
    local CHARTS_REPO_UPSTREAM_BRANCH=${5:?}
    local CHARTS_REPO_FORK_BRANCH=${6:?}

    local targetTagWithoutV=${TARGET_TAG#v}
    local targetChartPath="${CHARTS_REPO_FORK_LOCAL_PATH}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"

    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    # Fetch latest upstream changes, and commit&push them to the forked charts repo
    git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" remote add upstream "https://github.com/${CHARTS_REPO_UPSTREAM}.git"
    git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" pull upstream "${CHARTS_REPO_UPSTREAM_BRANCH}"
    # https://superuser.com/questions/232373/how-to-tell-git-which-private-key-to-use
    GIT_SSH_COMMAND="ssh -i ~/.ssh/${CHARTS_FORK_SSH_KEY_FILENAME}" git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" push origin "${CHARTS_REPO_FORK_BRANCH}"
    rm -rf "${targetChartPath}"
    cp -R "${KUBEAPPS_CHART_DIR}" "${targetChartPath}"

    # Update Chart.yaml with new version
    sed -i.bk 's/appVersion: DEVEL/appVersion: '"${targetTagWithoutV}"'/g' "${chartYaml}"
    rm "${targetChartPath}/Chart.yaml.bk"
    info "New version ${targetTagWithoutV} applied to file ${chartYaml}"

    # Replace images for the latest available
    # TODO: use the IMAGES_TO_PUSH var already set in the CI config
    replaceImage_latestToProduction dashboard "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction apprepository-controller "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction asset-syncer "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction pinniped-proxy "${targetChartPath}/values.yaml"
    replaceImage_latestToProduction kubeapps-apis "${targetChartPath}/values.yaml"
}

########################################################################################################################
# Syncs the fork of the charts repo with its upstream, then updates the local Helm chart stored in the kubeapps repo
# with the chart information taken from the bitnami charts repo, and applying to it the new version extracted from the
# passed TARGET_TAG param.
# Globals:
#   KUBEAPPS_CHART_DIR: Path of the Kubeapps chart in the Kubeapps repo.
#   CHART_REPO_PATH: Path of the Kubeapps chart in the charts repo.
# Arguments:
#   $1 - CHARTS_REPO_FORK_LOCAL_PATH: Path to the clone of the bitnami/charts repo in the local machine.
#   $2 - TARGET_TAG: Tag from which to take the new version for the Helm chart.
#   $3 - CHARTS_FORK_SSH_KEY_FILENAME: Filename of the SSH key to connect with the remote charts repo fork.
#   $4 - CHARTS_REPO_UPSTREAM: Name of the upstream version of the bitnami/charts repo without the GitHub part (eg. bitnami/charts).
#   $5 - CHARTS_REPO_UPSTREAM_BRANCH: Name of the main branch in the upstream of the charts repo.
#   $6 - CHARTS_REPO_FORK_BRANCH: Name of the main branch in the origin remove of the fork of charts repo.
# Returns:
#   0 - Success
#   1 - Failure
########################################################################################################################
updateRepoWithRemoteChanges() {
    local CHARTS_REPO_FORK_LOCAL_PATH=${1:?}
    local TARGET_TAG=${2:?}
    local CHARTS_FORK_SSH_KEY_FILENAME=${3:?}
    local CHARTS_REPO_UPSTREAM=${4:?}
    local CHARTS_REPO_UPSTREAM_BRANCH=${5:?}
    local CHARTS_REPO_FORK_BRANCH=${6:?}

    local targetTagWithoutV=${TARGET_TAG#v}
    local targetChartPath="${CHARTS_REPO_FORK_LOCAL_PATH}/${CHART_REPO_PATH}"
    local remoteChartYaml="${targetChartPath}/Chart.yaml"
    local localChartYaml="${KUBEAPPS_CHART_DIR}/Chart.yaml"

    if [ ! -f "${remoteChartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    # Fetch latest upstream changes, and commit&push them to the forked charts repo
    git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" remote add upstream "https://github.com/${CHARTS_REPO_UPSTREAM}.git"
    git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" pull upstream "${CHARTS_REPO_UPSTREAM_BRANCH}"
    # https://superuser.com/questions/232373/how-to-tell-git-which-private-key-to-use
    GIT_SSH_COMMAND="ssh -i ~/.ssh/${CHARTS_FORK_SSH_KEY_FILENAME}" git -C "${CHARTS_REPO_FORK_LOCAL_PATH}" push origin "${CHARTS_REPO_FORK_BRANCH}"
    rm -rf "${KUBEAPPS_CHART_DIR}"
    cp -R "${targetChartPath}" "${KUBEAPPS_CHART_DIR}"

    # Update Chart.yaml with new version
    sed -i.bk "s/appVersion: "${targetTagWithoutV}"/appVersion: DEVEL/g" "${localChartYaml}"
    rm "${KUBEAPPS_CHART_DIR}/Chart.yaml.bk"
    info "New version ${targetTagWithoutV} applied to file ${localChartYaml}"

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

########################################################################################################################
# Files a PR to update the Helm chart in the bitnami/charts repository to a new version.
# Globals:
#   KUBEAPPS_CHART_DIR: Path of the Kubeapps chart in the Kubeapps repo.
#   CHART_REPO_PATH: Path of the Kubeapps chart in the charts repo.
# Arguments:
#   $1 - LOCAL_CHARTS_REPO_PATH: Path to the clone of the bitnami/charts repo in the local machine.
#   $2 - TARGET_BRANCH: Name of the branch to create for the PR.
#   $3 - CHART_VERSION: New version for the chart.
#   $4 - CHARTS_REPO_UPSTREAM: Name of the upstream version of the bitnami/charts repo without the GitHub part (eg. bitnami/charts).
#   $5 - CHARTS_REPO_UPSTREAM_BRANCH: Name of the main branch in the upstream of the charts repo.
#   $6 - CHARTS_FORK_SSH_KEY_FILENAME: Name of the file with the SSH private key to connect with the upstream of the charts fork.
#   $7 - DEV_MODE: Indicates if it should be run in development mode, in this case we add a disclaimer to the PR description
#         alerting that it's a development PR and shouldn't be taken into account, between other customizations (branch name, etc).
# Returns:
#   0 - Success
#   1 - Failure
########################################################################################################################
commitAndSendExternalPR() {
    local LOCAL_CHARTS_REPO_PATH=${1:?}
    local TARGET_BRANCH=${2:?}
    local CHART_VERSION=${3:?}
    local CHARTS_REPO_UPSTREAM=${4:?}
    local CHARTS_REPO_UPSTREAM_BRANCH=${5:?}
    local CHARTS_FORK_SSH_KEY_FILENAME=${6:?}
    local DEV_MODE=${7-false}

    local targetChartPath="${LOCAL_CHARTS_REPO_PATH}/${CHART_REPO_PATH}"
    local chartYaml="${targetChartPath}/Chart.yaml"

    if [ ! -f "${chartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi
    cd "${LOCAL_CHARTS_REPO_PATH}"
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" >/dev/stderr
        cd -
        return 1
    fi

    PR_TITLE="[bitnami/kubeapps] Bump chart version to ${CHART_VERSION}"

    if [[ "${DEV_MODE}" == "true" ]]; then
      timestamp=$(date +%s)
      TARGET_BRANCH="${TARGET_BRANCH}-DEV-${timestamp}"
      PR_TITLE="DEV - ${PR_TITLE} - ${timestamp}"
      tmpfile=$(mktemp)
      echo "# :warning: THIS IS A DEVELOPMENT PR, DO NOT MERGE!"|cat - "${PR_EXTERNAL_TEMPLATE_FILE}" > "$tmpfile" && mv "$tmpfile" "${PR_EXTERNAL_TEMPLATE_FILE}"
    fi

    sed -i.bk -e "s/<USER>/$(git config user.name)/g" "${PR_EXTERNAL_TEMPLATE_FILE}"
    sed -i.bk -e "s/<EMAIL>/$(git config user.email)/g" "${PR_EXTERNAL_TEMPLATE_FILE}"
    git checkout -b "${TARGET_BRANCH}"
    git add --all .
    git commit --signoff -m "kubeapps: bump chart version to ${CHART_VERSION}"
    # NOTE: This expects to have a loaded SSH key
    if [[ $(GIT_SSH_COMMAND="ssh -i ~/.ssh/${CHARTS_FORK_SSH_KEY_FILENAME}" git ls-remote origin "${TARGET_BRANCH}" | wc -l) -eq 0 ]]; then
        GIT_SSH_COMMAND="ssh -i ~/.ssh/${CHARTS_FORK_SSH_KEY_FILENAME}" git push -u origin "${TARGET_BRANCH}"
        if [[ "${DEV_MODE}" != "true" ]]; then
          gh pr create -d -B "${CHARTS_REPO_UPSTREAM_BRANCH}" -R "${CHARTS_REPO_UPSTREAM}" -F "${PR_EXTERNAL_TEMPLATE_FILE}" --title "${PR_TITLE}"
        else
          echo "Skipping external PR because we are running in DEV_MODE"
        fi
    else
        echo "The remote branch '${TARGET_BRANCH}' already exists, please check if there is already an open PR at the repository '${CHARTS_REPO_UPSTREAM}'"
        return 1
    fi
    cd -
}

########################################################################################################################
# Updates the local Helm chart to a new version and files a PR against the upstream Kubeapps repo.
# Globals:
#   KUBEAPPS_CHART_DIR: Path of the Kubeapps chart in the Kubeapps repo.
# Arguments:
#   $1 - LOCAL_REPO_PATH: Path to the clone of the Kubeapps repo in the local machine.
#   $2 - TARGET_BRANCH: Name of the branch to create for the PR.
#   $3 - CHART_VERSION: New version for the chart.
#   $4 - UPSTREAM_REPO: Name of the upstream version of the kubeapps repo without the GitHub part (eg. vmware-tanzu/kubeapps).
#   $5 - UPSTREAM_MAIN_BRANCH: Name of the main branch in the upstream repo.
#   $6 - DEV_MODE: Indicates if it should be run in development mode, in this case we add a disclaimer to the PR description
#         alerting that it's a development PR and shouldn't be taken into account, between other customizations (branch name, etc).
# Returns:
#   0 - Success
#   1 - Failure
########################################################################################################################
commitAndSendInternalPR() {
    local LOCAL_REPO_PATH=${1:?}
    local TARGET_BRANCH=${2:?}
    local CHART_VERSION=${3:?}
    local UPSTREAM_REPO=${4:?}
    local UPSTREAM_MAIN_BRANCH=${5:?}
    local DEV_MODE=${6:-false}

    local targetChartPath="${KUBEAPPS_CHART_DIR}/Chart.yaml"
    local localChartYaml="${KUBEAPPS_CHART_DIR}/Chart.yaml"

    if [ ! -f "${localChartYaml}" ]; then
        echo "Wrong repo path. You should provide the root of the repository" >/dev/stderr
        return 1
    fi

    cd "${LOCAL_REPO_PATH}"
    if [[ ! $(git diff-index HEAD) ]]; then
        echo "Not found any change to commit" >/dev/stderr
        cd -
        return 1
    fi

    PR_TITLE="Sync chart with bitnami/kubeapps chart (version ${CHART_VERSION})"

    if [[ "${DEV_MODE}" == "true" ]]; then
        timestamp=$(date +%s)
        TARGET_BRANCH="${TARGET_BRANCH}-DEV-${timestamp}"
        PR_TITLE="DEV - ${PR_TITLE} - ${timestamp}"
        tmpfile=$(mktemp)
        echo "# :warning: THIS IS A DEVELOPMENT PR, DO NOT MERGE!"|cat - "${PR_INTERNAL_TEMPLATE_FILE}" > "$tmpfile" && mv "$tmpfile" "${PR_INTERNAL_TEMPLATE_FILE}"
    fi

    git checkout -b "${TARGET_BRANCH}"
    git add --all .
    git commit --signoff -m "bump chart version to ${CHART_VERSION}"
    # NOTE: This expects to have a loaded SSH key
    if [[ $(git ls-remote origin "${TARGET_BRANCH}" | wc -l) -eq 0 ]]; then
        git push -u origin "${TARGET_BRANCH}"
        gh pr create -d -B "${UPSTREAM_MAIN_BRANCH}" -R "${UPSTREAM_REPO}" -F "${PR_INTERNAL_TEMPLATE_FILE}" --title "${PR_TITLE}"
    else
        echo "The remote branch '${TARGET_BRANCH}' already exists, please check if there is already an open PR at the repository '${UPSTREAM_REPO}'"
        return 1
    fi
    cd -
}
