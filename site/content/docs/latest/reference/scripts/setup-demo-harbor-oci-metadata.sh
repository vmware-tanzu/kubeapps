#!/usr/bin/env bash

# Copyright 2023 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

# Constants
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../../.." >/dev/null && pwd)"
RESET='\033[0m'
GREEN='\033[38;5;2m'
RED='\033[38;5;1m'
YELLOW='\033[38;5;3m'

VERSION="0.1.0"
IMAGE="demo.goharbor.io/kubeapps-test/simplechart:${VERSION}"

# shellcheck disable=SC1090
. "${ROOT_DIR}/script/lib/liblog.sh"

info "Setting up kubeapps project on demo.goharbor with metadata referers..."

# Check if project exists - if not, create it.
http_code=$(curl --head https://demo.goharbor.io/api/v2.0/projects\?project_name\=kubeapps-test -u "${DEMO_USERNAME}:${DEMO_PASSWORD}" -o /dev/null -sSLw "%{http_code}")

if [[ "${http_code}" -eq 404  ]] ; then
  info "'kubeapps-test' project does not exist yet, creating..."
  curl \
    'https://demo.goharbor.io/api/v2.0/projects' \
    -u "${DEMO_USERNAME}:${DEMO_PASSWORD}" \
    -H 'accept: application/json' \
    -H 'X-Resource-Name-In-Location: false' \
    -H 'Content-Type: application/json' \
    -d '{
    "project_name": "kubeapps-test",
    "public": true,
    "metadata": {
      "public": "true",
      "enable_content_trust": "string",
      "enable_content_trust_cosign": "string",
      "prevent_vul": "string",
      "severity": "string",
      "auto_scan": "string",
      "reuse_sys_cve_allowlist": "string",
      "retention_id": ""
    }
  }'
elif [[ "${http_code}" -eq 200 ]] ; then
  info "Project kubeapps-test already exists."
else
  error "Unexpected http code: ${http_code}. Exiting"
fi

info "Creating and pushing chart to kubeapps-test project..."
helm package ./integration/charts/simplechart
echo ${DEMO_PASSWORD} | oras login --username ${DEMO_USERNAME} --password-stdin demo.goharbor.io

oras push --artifact-type "application/vnd.cncf.helm.config.v1" --export-manifest "./chart-manifest.json" ${IMAGE} "simplechart-0.1.0.tgz:application/vnd.oci.image.manifest.v1"
# helm push ./simplechart-${VERSION}.tgz oci://demo.goharbor.io/kubeapps-test

info "Attaching super-important-meta with chart..."
echo '{"artifact": "'${IMAGE}'", "signature": "trust me"}' > signature.json
oras attach --export-manifest "./signature-manifest.json" ${IMAGE} --artifact-type "application/vnd.example.signature.v1" "signature.json:application/json"

echo '{"artifact": "'${IMAGE}'", "sbom": "lots of materials"}' > sbom.json
oras attach --export-manifest "./sbom-manifest.json" --artifact-type "application/vnd.example.sbom.v1" ${IMAGE} "sbom.json:application/vnd.oci.image.manifest.v1"


info "Listing referrers of chart..."
oras discover -o tree ${IMAGE}
