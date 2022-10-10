#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"

source "${ROOT_DIR}/script/lib/liblog.sh"

# If we want to test the latest version instead we override the image to be used
if [[ -n "${TEST_LATEST_RELEASE:-}" ]]; then
  source "${ROOT_DIR}/script/chart_sync_utils.sh"
  latest="$(latestReleaseTag)"
  IMG_DEV_TAG=${latest/v/}
  IMG_MODIFIER=""
fi

test_command="${ROOT_DIR}/script/e2e-test.sh \
  ${USE_MULTICLUSTER_OIDC_ENV} \
  ${OLM_VERSION} \
  ${IMG_DEV_TAG} \
  ${IMG_MODIFIER} \
  ${TEST_TIMEOUT_MINUTES} \
  ${DEFAULT_DEX_IP} \
  ${ADDITIONAL_CLUSTER_IP} \
  ${KAPP_CONTROLLER_VERSION} \
  ${CHARTMUSEUM_VERSION}"

info "IMG_PREFIX: $IMG_PREFIX"
info "TESTS_GROUP: $TESTS_GROUP"
info "CMD: $test_command"

if IMG_PREFIX=${IMG_PREFIX} TESTS_GROUP=${TESTS_GROUP} ${test_command}; then
  echo "TEST_RESULT=0" >> "${GITHUB_ENV}"
  exit 0
fi

echo "TEST_RESULT=1" >> "${GITHUB_ENV}"
exit 1
