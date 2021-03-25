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
devel/helm-packages-v1-plugin.so:
	go build -o $@ -buildmode=plugin plugins/helm/packages/v1/main.go

devel/kapp-controller-packages-v1-plugin.so:
	go build -o $@ -buildmode=plugin plugins/kapp-controller/packages/v1/main.go

run: devel/helm-packages-v1-plugin.so devel/kapp-controller-packages-v1-plugin.so
	go run main.go --plugin-dir devel/

test:
	go test ./...

generate:
	buf generate

.PHONY: run test cli-dependencies generate
