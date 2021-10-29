#!/usr/bin/env bash

# Copyright 2018-2021 VMware. All Rights Reserved.
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

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)
CHART_DIR=$ROOT_DIR/chart/kubeapps/

helm dep up "${CHART_DIR}"

# test with the minium supported helm version
helm template "${CHART_DIR}" --debug

# test with the latest stable helm version
helm-stable template "${CHART_DIR}" --debug
