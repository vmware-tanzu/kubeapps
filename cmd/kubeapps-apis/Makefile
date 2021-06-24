# Copyright 2021 VMware. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
	go get \
		github.com/spf13/cobra/cobra \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
		github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway

# Not sure if it's worth having targets for buf, but at least the targets can
# be tooling agnostic.
generate:
	buf generate

run: build-plugins
	go run main.go --plugin-dir devel/ --unsafe-use-demo-sa=true --unsafe-local-dev-kubeconfig=true

test:
	go test ./...

.PHONY: build-plugins cli-dependencies generate run test
