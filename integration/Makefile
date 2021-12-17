IMAGE_TAG ?= dev

# Retrieve the latest version pushed upstream at
# https://hub.docker.com/r/kubeapps/integration-tests/tags?page=1&ordering=last_updated

default: build

build:
	docker build -t kubeapps/integration-tests:$(IMAGE_TAG) .

push:
	docker push kubeapps/integration-tests:$(IMAGE_TAG)
