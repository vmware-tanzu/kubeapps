IMPORT_PATH:= github.com/kubeapps/kubeapps
GO = /usr/bin/env go
GOFMT = /usr/bin/env gofmt
VERSION ?= dev-$(shell date +%FT%H-%M-%S-%Z)

BINARY ?= kubeapps
GO_PACKAGES = ./...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)
GO_FLAGS = -ldflags="-s -w -X github.com/kubeapps/kubeapps/cmd/kubeapps.VERSION=${VERSION}"
EMBEDDED_STATIC = generated/statik/statik.go

default: kubeapps

kubeapps: $(EMBEDDED_STATIC)
	$(GO) build -i -o $(BINARY) $(GO_FLAGS) $(IMPORT_PATH)

test: $(EMBEDDED_STATIC)
	$(GO) test $(GO_PACKAGES)

$(EMBEDDED_STATIC): static/kubeapps-objs.yaml
	# force compilation on current OS
	GOOS= $(GO) build -o statik ./vendor/github.com/rakyll/statik/statik.go
	$(GO) generate

static/kubeapps-objs.yaml:
	KUBECFG_JPATH=./manifests/lib:./manifests/vendor/kubecfg/lib:./manifests/vendor/ksonnet-lib kubecfg show manifests/kubeapps.jsonnet > static/kubeapps-objs.yaml

kubeapps/%:
	docker build -t kubeapps/$*:$(VERSION) -f cmd/$*/Dockerfile .
kubeapps/dashboard:
	docker build -t kubeapps/dashboard:$(VERSION) -f dashboard/Dockerfile dashboard/

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_PACKAGES)

clean:
	$(RM) ./kubeapps ./chart-repo ./statik $(EMBEDDED_STATIC) static/kubeapps-objs.yaml

.PHONY: default test fmt vet clean build-prep chart-repo kubeapps
