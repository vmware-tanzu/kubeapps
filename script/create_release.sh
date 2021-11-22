#!/bin/bash

# Copyright 2021 VMware. All Rights Reserved.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

source $(dirname $0)/chart_sync_utils.sh

TAG=${1:?Missing tag}
KUBEAPPS_REPO=${2:?Missing kubeapps repo}

if [[ -z "${TAG}" ]]; then
  echo "A git tag is required for creating a release"
  exit 1
fi

gh release create -R "${KUBEAPPS_REPO}" -d "${TAG}" -t "${TAG}" -F "${RELEASE_NOTES_TEMPLATE_FILE}"
