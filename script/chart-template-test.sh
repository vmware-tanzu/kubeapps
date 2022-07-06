#!/usr/bin/env bash

# Copyright 2018-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null && pwd)
CHART_DIR=$ROOT_DIR/chart/kubeapps/

# Remove the kubeVersion field as it would fail otherwise as there isn't any k8s cluster
# when running this command from the CI
sed -i.bk -e "s/kubeVersion.*//g" "${CHART_DIR}Chart.yaml"

helm dep up "${CHART_DIR}"

# test with the minium supported helm version
helm template "${CHART_DIR}" --debug

# test with the latest stable helm version
helm-stable template "${CHART_DIR}" --debug
