GO = go
GO_FLAGS =
GOFMT = gofmt
VERSION = dev-$(shell date +%FT%T%z)

OS = linux
ARCH = amd64
BINARY = kubeapps
GO_PACKAGES = ./cmd/...
GO_FILES := $(shell find $(shell $(GO) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)
GO_FLAGS = -ldflags="-w -X github.com/kubeapps/installer/cmd.VERSION=${VERSION}"

.PHONY: binary

default: binary

binary:
	$(GO) build -o $(BINARY) $(GO_FLAGS) .

test:
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)