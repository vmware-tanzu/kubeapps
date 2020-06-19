IMPORT_PATH:= github.com/kubeapps/kubeapps
GO = /usr/bin/env go
GOFMT = /usr/bin/env gofmt
IMAGE_TAG ?= dev-$(shell date +%FT%H-%M-%S-%Z)
VERSION ?= $$(git rev-parse HEAD)

default: all

include ./script/cluster-kind.mk
include ./script/cluster-openshift.mk
include ./script/deploy-dev.mk

IMG_MODIFIER ?=

GO_PACKAGES = ./...
# GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

all: kubeapps/dashboard kubeapps/apprepository-controller kubeapps/tiller-proxy kubeapps/kubeops kubeapps/assetsvc kubeapps/asset-syncer

# TODO(miguel) Create Makefiles per component
kubeapps/%:
	DOCKER_BUILDKIT=1 docker build -t kubeapps/$*$(IMG_MODIFIER):$(IMAGE_TAG) --build-arg "VERSION=${VERSION}" -f cmd/$*/Dockerfile .

kubeapps/dashboard:
	docker build -t kubeapps/dashboard$(IMG_MODIFIER):$(IMAGE_TAG) -f dashboard/Dockerfile dashboard/

test:
	$(GO) test $(GO_PACKAGES)

start-db:
	docker run --rm --publish 5432:5432 -e ALLOW_EMPTY_PASSWORD=yes bitnami/postgresql:11.6.0-debian-9-r0 &
	until psql -h 127.0.0.1 -U postgres -c "select 1" > /dev/null 2>&1; do echo "Waiting for postgres server..."; sleep 1; done

test-db: start-db
	cd cmd/asset-syncer; ENABLE_PG_INTEGRATION_TESTS=1 go test -count=1 ./...
	cd cmd/assetsvc; ENABLE_PG_INTEGRATION_TESTS=1 go test -count=1 ./...

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
