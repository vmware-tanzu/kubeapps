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
EMBEDDED_STATIC = generated/statik/statik.go

default: binary

binary: $(EMBEDDED_STATIC)
	$(GO) build -o $(BINARY) $(GO_FLAGS) .

test: $(EMBEDDED_STATIC)
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

$(EMBEDDED_STATIC): $(wilcard static/*)
	$(GO) build -o statik ./vendor/github.com/rakyll/statik/statik.go
	$(GO) generate

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)

clean:
	$(RM) kubeapps $(EMBEDDED_STATIC)

.PHONY: default binary test fmt vet clean
