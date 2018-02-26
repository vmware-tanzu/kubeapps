IMPORT_PATH:= github.com/kubeapps/kubeapps
GOBIN = go
# Force builds to only use vendor/'ed dependencies
# i.e. ignore local $GOPATH/src installed sources
GOPATH_TMP = $(CURDIR)/.GOPATH
GO = /usr/bin/env GOPATH=$(GOPATH_TMP) $(GOBIN)
GOFMT = gofmt
VERSION = dev-$(shell date +%FT%T%z)

BINARY = kubeapps
GO_PACKAGES = $(IMPORT_PATH)/cmd/kubeapps $(IMPORT_PATH)/cmd/chart-repo $(IMPORT_PATH)/pkg/...
GO_FILES := $(shell find $(shell $(GOBIN) list -f '{{.Dir}}' $(GO_PACKAGES)) -name \*.go)
GO_FLAGS = -ldflags='-w -X github.com/kubeapps/kubeapps/cmd/kubeapps.VERSION=${VERSION}'
GO_XFLAGS =
EMBEDDED_STATIC = generated/statik/statik.go

# Cross-compilation env
GO_BUILD_ENV_linux-amd64 = GOOS=linux GOARCH=amd64
GO_BUILD_ENV_darwin-amd64 = PATH=$(PATH):$(PWD)/osxcross/target/bin/ CC=x86_64-apple-darwin15-clang CXX=x86_64-apple-darwin15-clang++ GOOS=darwin GOARCH=amd64
GO_BUILD_ENV_windows-amd64.exe = CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64

default: kubeapps

kubeapps: build-prep $(EMBEDDED_STATIC)
	CGO_ENABLED=1 $(GO) build -i -o $(BINARY) $(GO_FLAGS) $(IMPORT_PATH)

kubeapps-%: build-prep $(EMBEDDED_STATIC)-travis
	CGO_ENABLED=1 $(GO_BUILD_ENV_$*) $(GO) build -i -o kubeapps-$* $(GO_FLAGS) $(GO_XFLAGS) $(IMPORT_PATH)

osxcross:
	git clone https://github.com/tpoechtrager/osxcross
	sudo $(PWD)/osxcross/tools/get_dependencies.sh
	wget -O $(PWD)/osxcross/tarballs/MacOSX10.11.sdk.tar.xz \
		https://storage.googleapis.com/osx-cross-compiler/MacOSX10.11.sdk.tar.xz
	UNATTENDED=1 OSX_VERSION_MIN=10.9 $(PWD)/osxcross/build.sh

test: build-prep $(EMBEDDED_STATIC)
	$(GO) test $(GO_FLAGS) $(GO_PACKAGES)

$(EMBEDDED_STATIC): build-prep static/kubeapps-objs.yaml
	$(GO) build -o statik ./vendor/github.com/rakyll/statik/statik.go
	$(GO) generate

$(EMBEDDED_STATIC)-travis: build-prep static/kubeapps-objs.yaml
	GOOS=linux $(GO) build -o statik ./vendor/github.com/rakyll/statik/statik.go
	GOOS=linux $(GO) generate

static/kubeapps-objs.yaml:
	KUBECFG_JPATH=./manifests/lib:./manifests/vendor/kubecfg/lib:./manifests/vendor/ksonnet-lib kubecfg show manifests/kubeapps.jsonnet > static/kubeapps-objs.yaml

build-prep:
	mkdir -p $(dir $(GOPATH_TMP)/src/$(IMPORT_PATH))
	ln -snf $(CURDIR) $(GOPATH_TMP)/src/$(IMPORT_PATH)

chart-repo:
	docker build -t kubeapps/chart-repo:$(VERSION) -f cmd/chart-repo/Dockerfile .

fmt:
	$(GOFMT) -s -w $(GO_FILES)

vet:
	$(GO) vet $(GO_FLAGS) $(GO_PACKAGES)

clean:
	$(RM) ./kubeapps ./chart-repo ./statik $(EMBEDDED_STATIC)
	$(RM) -r $(GOPATH_TMP)

.PHONY: default test fmt vet clean build-prep chart-repo osxcross
