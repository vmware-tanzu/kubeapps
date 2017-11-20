IMPORT_PATH:= github.com/kubeapps/kubeapps
GOBIN = go
# Force builds to only use vendor/'ed dependencies
# i.e. ignore local $GOPATH/src installed sources
GOPATH_TMP = $(CURDIR)/.GOPATH
GO = /usr/bin/env GOPATH=$(GOPATH_TMP) $(GOBIN)
GO_FLAGS =
GOFMT = gofmt
VERSION = dev-$(shell date +%FT%T%z)

OS = linux
ARCH = amd64
BINARY = kubeapps
GO_PACKAGES = $(IMPORT_PATH)/cmd $(IMPORT_PATH)/pkg/...
GO_FILES := $(shell find $(shell $(GOBIN) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)
GO_FLAGS = -ldflags="-w -X github.com/kubeapps/kubeapps/cmd.VERSION=${VERSION}"
EMBEDDED_STATIC = generated/statik/statik.go

default: binary

binary: build-prep $(EMBEDDED_STATIC)
	$(GO) build -i -o $(BINARY) $(GO_FLAGS) $(IMPORT_PATH)

test: build-prep $(EMBEDDED_STATIC)
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

$(EMBEDDED_STATIC): build-prep $(wilcard static/*)
	$(GO) build -o statik ./vendor/github.com/rakyll/statik/statik.go
	$(GO) generate

build-prep:
	mkdir -p $(dir $(GOPATH_TMP)/src/$(IMPORT_PATH))
	ln -snf $(CURDIR) $(GOPATH_TMP)/src/$(IMPORT_PATH)

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)

clean:
	$(RM) ./kubeapps ./statik $(EMBEDDED_STATIC)
	$(RM) -r $(GOPATH_TMP)

.PHONY: default binary test fmt vet clean build-prep
