#!/usr/bin/env bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

command -v license-eye > /dev/null 2>&1 || { echo "license-eye must be installed -> https://github.com/apache/skywalking-eyes"; exit 1; }

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null && pwd)

license-eye -c "${PROJECT_DIR}/.licenserc.yaml" header check
