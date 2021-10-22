#!/bin/bash

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

TAG=${1:?}

PROJECT_DIR=$(cd $(dirname $0)/.. && pwd)

source $(dirname $0)/release_utils.sh

if [[ -z "$REPO_NAME" || -z "$REPO_DOMAIN" ]]; then
  echo "Github repository not specified" >/dev/stderr
  exit 1
fi

if [[ -z "$GITHUB_TOKEN" ]]; then
  echo "Unable to release: Github Token not specified" >/dev/stderr
  exit 1
fi

repo_check=$(curl -H "Authorization: token $GITHUB_TOKEN" -s https://api.github.com/repos/$REPO_DOMAIN/$REPO_NAME)
if [[ $repo_check == *"Not Found"* ]]; then
  echo "Not found a Github repository for $REPO_DOMAIN/$REPO_NAME, it is not possible to publish it" >/dev/stderr
  exit 1
else
  RELEASE_ID=$(release_tag $TAG $REPO_DOMAIN $REPO_NAME | jq '.id')
fi

echo "RELEASE ID: $RELEASE_ID"
