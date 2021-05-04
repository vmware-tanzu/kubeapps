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

source $(dirname $0)/chart_sync_utils.sh

user=${1:?}
email=${2:?}
if changedVersion; then
    tempDir=$(mktemp -u)/charts
    mkdir -p $tempDir
    git clone https://github.com/${CHARTS_REPO} $tempDir  --depth 1 --no-single-branch 
    configUser $tempDir $user $email
    configUser $PROJECT_DIR $user $email
    latestVersion=$(latestReleaseTag $PROJECT_DIR)
    updateRepoWithRemoteChanges $tempDir $latestVersion
    commitAndSendInternalPR ${PROJECT_DIR} "sync-chart-changes-${latestVersion}"
else
    echo "Skipping Chart sync. The version has not changed"
fi
