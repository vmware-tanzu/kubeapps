#!/usr/bin/env bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

command -v golangci-lint > /dev/null 2>&1 || { echo "golangci-lint must be installed -> https://github.com/golangci/golangci-lint"; exit 1; }

PROJECT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." >/dev/null && pwd)
TIMEOUT="10m"

golangci-lint run --timeout=${TIMEOUT} "${PROJECT_DIR}/cmd/..."
golangci-lint run --timeout=${TIMEOUT} "${PROJECT_DIR}/pkg/..."
