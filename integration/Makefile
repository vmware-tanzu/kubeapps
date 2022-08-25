# Copyright 2019-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

IMAGE_TAG ?= dev
IMG_MODIFIER ?=

# Retrieve the latest version pushed upstream at
# https://hub.docker.com/r/kubeapps/integration-tests/tags?page=1&ordering=last_updated

default: build

build:
	docker build -t kubeapps/integration-tests$(IMG_MODIFIER):$(IMAGE_TAG) .

push:
	docker push kubeapps/integration-tests$(IMG_MODIFIER):$(IMAGE_TAG)

kubeapps/%:
	DOCKER_BUILDKIT=1 docker build -t kubeapps/$*$(IMG_MODIFIER):$(IMAGE_TAG) --build-arg "VERSION=${VERSION}" .
