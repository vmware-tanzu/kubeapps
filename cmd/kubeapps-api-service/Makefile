# This Makefile is for local development and testing only.

# Ensure the required version of cli tooling is installed for grpc (and cobra)
# The versions are stored in ../go.mod as usual.
# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
cli-dependencies:
	go install \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/spf13/cobra/cobra \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go

# Build the helm and kapp package plugins with the output in the devel directory.
# TODO: currently this just rebuilds them every time.
build-plugins:
	go build -o devel/helm-packagerepositories-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packagerepositories/helm/v1/plugin/main.go
	go build -o devel/helm-packages-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packages/helm/v1/plugin/main.go
	go build -o devel/kapp-controller-packagerepositories-v1-plugin.so -buildmode=plugin kubeappsapis/plugins/packagerepositories/kapp_controller/v1/plugin/main.go

run: build-plugins
	go run main.go --plugin-dir devel/

test:
	go test ./...

generate:
	buf generate

.PHONY: run test build-plugins cli-dependencies generate
