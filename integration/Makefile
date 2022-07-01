# Copyright 2019-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

IMAGE_TAG ?= dev

# Retrieve the latest version pushed upstream at
# https://hub.docker.com/r/kubeapps/integration-tests/tags?page=1&ordering=last_updated

default: build

build:
	docker build -t kubeapps/integration-tests:v2.0.0 .

push:
	docker push kubeapps/integration-tests:v2.0.0
