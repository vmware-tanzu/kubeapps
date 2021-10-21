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

function commit_list {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  git fetch --tags
  local previous_tag=$(git describe --abbrev=0 --tags $(git rev-list --tags --skip=1 --max-count=1))
  local commit_list=$(git log $previous_tag..$tag --pretty=format:"- %s %H (%an)")
  echo "$commit_list"
}

function get_release_notes {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local commits=$(commit_list $tag $repo_domain $repo_name)
  local current_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)
  local release_notes=$(echo "$(cat $current_dir/tpl/release_notes.tpl)${commits}" | sed 's/<<TAG>>/'"${tag}"'/g')
  local escaped_release_notes=$(echo "$release_notes" | sed -n -e 'H;${x;s/\n/\\\n/g;s/"/\\\"/g;p;}')
  echo "${escaped_release_notes}"
}

function get_release_body {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local release_notes=$(get_release_notes $tag $repo_domain $repo_name)
  echo '{
    "tag_name": "'$tag'",
    "target_commitish": "master",
    "name": "'$tag'",
    "body": "'$release_notes'",
    "draft": true,
    "prerelease": false
  }'
}

function update_release_tag {
  local tag=${1:?}
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local release_id=$(curl -H "Authorization: token $GITHUB_TOKEN" -s https://api.github.com/repos/$repo_domain/$repo_name/releases | jq --raw-output '.[0].id')
  local body=$(get_release_body $tag $repo_domain $repo_name)
  local release=$(curl -H "Authorization: token $GITHUB_TOKEN" -s --request PATCH --data $body https://api.github.com/repos/$repo_domain/$repo_name/releases/$release_id)
  echo $release
}

function release_tag {
  local tag=$1
  local repo_domain=${2:?}
  local repo_name=${3:?}
  local body=$(get_release_body $tag $repo_domain $repo_name)
  local release=$(curl -H "Authorization: token $GITHUB_TOKEN" -s --request POST --data "$body" https://api.github.com/repos/$repo_domain/$repo_name/releases)
  echo $release
}

function upload_asset {
  local repo_domain=${1:?}
  local repo_name=${2:?}
  local release_id=${3:?}
  local asset=${4:?}
  local filename=$(basename $asset)
  if [[ "$filename" == *".zip" ]]; then
    local content_type="application/zip"
  elif [[ "$filename" == *".yaml" ]]; then
    local content_type="text/yaml"
  else
    local content_type="application/octet-stream"
  fi
  curl -H "Authorization: token $GITHUB_TOKEN" \
    -H "Content-Type: $content_type" \
    --data-binary @"$asset" \
    "https://uploads.github.com/repos/$repo_domain/$repo_name/releases/$release_id/assets?name=$filename"
}
