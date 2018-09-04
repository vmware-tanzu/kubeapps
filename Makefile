IMPORT_PATH:= github.com/kubeapps/kubeapps
GO = /usr/bin/env go
GOFMT = /usr/bin/env gofmt
VERSION ?= dev-$(shell date +%FT%H-%M-%S-%Z)

GO_PACKAGES = ./...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)

default: all

all: kubeapps/dashboard kubeapps/chartsvc kubeapps/chart-repo kubeapps/apprepository-controller

# TODO(miguel) Create Makefiles per component
kubeapps/%:
	CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -o ./cmd/$*/$*-static ./cmd/$*
	docker build -t kubeapps/$*:$(VERSION) -f cmd/$*/Dockerfile cmd/$*

kubeapps/dashboard:
	docker build -t kubeapps/dashboard:$(VERSION) -f dashboard/Dockerfile dashboard/

test:
	$(GO) test $(GO_PACKAGES)

test-all: test-chartsvc test-chart-repo test-apprepository-controller test-dashboard

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
