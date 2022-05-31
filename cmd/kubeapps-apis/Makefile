# Copyright 2021-2022 the Kubeapps contributors.
# SPDX-License-Identifier: Apache-2.0

# Build the helm and kapp package plugins with the output in the devel directory.
build-plugins:
	go build -o devel/kapp-controller-packages-v1alpha1-plugin.so -buildmode=plugin plugins/kapp_controller/packages/v1alpha1/*.go
	go build -o devel/fluxv2-packages-v1alpha1-plugin.so -buildmode=plugin plugins/fluxv2/packages/v1alpha1/*.go
	go build -o devel/helm-packages-v1alpha1-plugin.so -buildmode=plugin plugins/helm/packages/v1alpha1/*.go


# Ensure the required version of cli tooling in tools/tools.go is installed.
# The versions are stored in ../go.mod as usual and this same approach is used in
# both octant and pinniped for build tooling.
# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
cli-dependencies:
	go install \
		github.com/spf13/cobra-cli \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway

# Not sure if it's worth having targets for buf, but at least the targets can
# be tooling agnostic.
generate:
	buf generate

run: build-plugins
	go run main.go --plugin-dir devel/ --unsafe-local-dev-kubeconfig=true

test:
	go test ./...

.PHONY: build-plugins cli-dependencies generate run test
