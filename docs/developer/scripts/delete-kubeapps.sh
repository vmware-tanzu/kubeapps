#!/usr/bin/env bash

# Copyright 2020-2021 VMware. All Rights Reserved.
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

# Constants
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)"

# Load Libraries
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/libtest.sh"
# shellcheck disable=SC1090
. "${ROOT_DIR}/script/liblog.sh"

namespace="kubeapps"
while [[ "$#" -gt 0 ]]; do
    case "$1" in
    -n | --namespace)
        shift
        namespace="${1:?missing namespace}"
        ;;
    *)
        echo "Invalid command line flag $1" >&2
        return 1
        ;;
    esac
    shift
done

# Uninstall Kubeapps
info "Uninstalling Kubeapps in namespace '$namespace'..."
silence helm uninstall kubeapps -n "$namespace"
silence kubectl delete rolebinding example-kubeapps-repositories-write -n "$namespace"
info "Deleting '$namespace' namespace..."
silence kubectl delete ns "$namespace"

# Delete serviceAccount
info "Deleting 'example' serviceAccount and related RBAC objects..."
silence kubectl delete serviceaccount example --namespace default
silence kubectl delete rolebinding example-edit
