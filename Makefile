# Copyright 2017-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

IMPORT_PATH:= github.com/vmware-tanzu/kubeapps
GO = /usr/bin/env go
GOFMT = /usr/bin/env gofmt
IMAGE_TAG ?= dev-$(shell date +%FT%H-%M-%S-%Z)
VERSION ?= $$(git rev-parse HEAD)
TARGET_ARCHITECTURE ?= amd64
SUPPORTED_ARCHITECTURES := amd64 arm64 riscv64 ppc64le s390x 386 arm/v7 arm/v6

ifeq ($(filter $(TARGET_ARCHITECTURE),$(SUPPORTED_ARCHITECTURES)),)
COMMA:=,
EMPTY:=
WHITESPACE:=$(EMPTY) $(EMPTY)
$(error The provided TARGET_ARCHITECTURE '$(TARGET_ARCHITECTURE)' is not supported, provide one of '$(subst $(WHITESPACE),$(COMMA)$(WHITESPACE),$(SUPPORTED_ARCHITECTURES))')
endif

default: all

include ./script/makefiles/cluster-kind.mk
include ./script/makefiles/cluster-kind-for-pinniped.mk
include ./script/makefiles/deploy-dev.mk
include ./script/makefiles/deploy-dev-for-pinniped.mk
include ./script/makefiles/site.mk

IMG_MODIFIER ?=

GO_PACKAGES = ./...
# GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

all: kubeapps/dashboard kubeapps/apprepository-controller kubeapps/asset-syncer kubeapps/pinniped-proxy kubeapps/kubeapps-apis

# TODO(miguel) Create Makefiles per component
# TODO(mnelson) Or at least don't send the whole repo as the context for each project.
# Currently the go projects include the whole repository as the docker context
# only because the shared pkg/ directories?
kubeapps/%:
	DOCKER_BUILDKIT=1 docker build --platform "linux/$(TARGET_ARCHITECTURE)" -t kubeapps/$*$(IMG_MODIFIER):$(IMAGE_TAG) --build-arg "VERSION=${VERSION}" -f cmd/$*/Dockerfile .

kubeapps/dashboard:
	docker build --platform "linux/$(TARGET_ARCHITECTURE)" -t kubeapps/dashboard$(IMG_MODIFIER):$(IMAGE_TAG) -f dashboard/Dockerfile dashboard/

test:
	$(GO) test $(GO_PACKAGES)

test-db:
	# It's not supported to run tests that involve a database in parallel since they are currently
	# using the same PG schema. We need to run them sequentially
	cd cmd/asset-syncer; ENABLE_PG_INTEGRATION_TESTS=1 go test -count=1 ./...

test-all: test-apprepository-controller test-dashboard

test-dashboard:
	yarn --cwd dashboard/ install --frozen-lockfile
	yarn --cwd=dashboard run lint
	CI=true yarn --cwd dashboard/ run test

test-%:
	$(GO) test -v $(IMPORT_PATH)/cmd/$*

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_PACKAGES)

.PHONY: default all test-all test test-dashboard fmt vet
