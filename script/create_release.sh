#!/usr/bin/env bash

# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

source $(dirname "$0")/chart_sync_utils.sh

TAG=${1:?Missing tag}
KUBEAPPS_REPO=${2:?Missing kubeapps repo}
DEV_MODE=${DEV_MODE:-false}
GH_TOKEN=${GH_TOKEN:?Missing GitHub token}

if [[ -z "${TAG}" ]]; then
  echo "A git tag is required for creating a release"
  exit 1
fi

if [[ "${DEV_MODE}" != "false" ]]; then
  gh release create -R "${KUBEAPPS_REPO}" -d "${TAG}" -t "${TAG}" -F "${RELEASE_NOTES_TEMPLATE_FILE}"
else
  echo "Skipping release due to DEV_MODE!"
fi
