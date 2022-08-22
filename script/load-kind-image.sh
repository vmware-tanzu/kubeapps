#!/bin/bash

# Copyright 2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

IMAGE=$1
MAIN_CLUSTER=$2
ADDITIONAL_CLUSTER=$3

echo "Loading image ${IMAGE} into clusters ${MAIN_CLUSTER} and ${ADDITIONAL_CLUSTER}"
docker pull ${IMAGE}
kind load docker-image ${IMAGE} --name ${MAIN_CLUSTER}
kind load docker-image ${IMAGE} --name ${ADDITIONAL_CLUSTER}
