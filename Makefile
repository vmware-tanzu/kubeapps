IMAGE_REPO ?= kubeapps/dashboard
IMAGE_TAG ?= latest

ifeq "$(VERSION)" ""
	override VERSION = dev
endif

install:
	yarn install

test:
	yarn test

test-ci:
	yarn run test-ci

compile:
	yarn run compile

compile-aot:
	yarn run compile-aot

docker-build: compile-aot
	docker build --pull --rm -t ${IMAGE_REPO}:${IMAGE_TAG} rootfs/

set-version:
	sed -i src/version.ts -e 's/dev/${VERSION}/'
