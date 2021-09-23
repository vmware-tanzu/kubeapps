IMAGE_TAG ?= dev

default: build

build:
	docker build -t kubeapps/integration-tests:$(IMAGE_TAG) .

push:
	docker push kubeapps/integration-tests:$(IMAGE_TAG)
