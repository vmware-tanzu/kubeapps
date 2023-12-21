#!/usr/bin/env bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"

TEST_LATEST_RELEASE=${TEST_LATEST_RELEASE:-false}

# If we want to test the latest version instead we override the image to be used
if [[ "${TEST_LATEST_RELEASE}" == "true" ]]; then
  source "${ROOT_DIR}/script/chart_sync_utils.sh"
  latest="$(latestReleaseTag)"
  IMG_DEV_TAG=${latest/v/}
  IMG_MODIFIER=""
fi

if
  IMG_PREFIX="${IMG_PREFIX}" \
  TESTS_GROUP="${TESTS_GROUP}" \
  USE_MULTICLUSTER_OIDC_ENV="${USE_MULTICLUSTER_OIDC_ENV}" \
  OLM_VERSION="${OLM_VERSION}" \
  IMG_DEV_TAG="${IMG_DEV_TAG}" \
  IMG_MODIFIER="${IMG_MODIFIER}" \
  TEST_TIMEOUT_MINUTES="${TEST_TIMEOUT_MINUTES}" \
  DEFAULT_DEX_IP="${DEFAULT_DEX_IP:-}" \
  ADDITIONAL_CLUSTER_IP="${ADDITIONAL_CLUSTER_IP:-}" \
  KAPP_CONTROLLER_VERSION="${KAPP_CONTROLLER_VERSION}" \
  FLUX_VERSION="${FLUX_VERSION}" \
  CHARTMUSEUM_VERSION="${CHARTMUSEUM_VERSION}" \
  "${ROOT_DIR}/script/e2e-test.sh";
then
  echo "TEST_RESULT=0" >> "${GITHUB_ENV}"
  exit 0
fi

echo "TEST_RESULT=1" >> "${GITHUB_ENV}"
exit 1
