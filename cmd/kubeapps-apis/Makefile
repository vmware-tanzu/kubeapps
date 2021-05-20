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
	go build -o devel/helm-operator-packages-v1alpha1-plugin.so -buildmode=plugin plugins/helm_fluxv2/packages/v1alpha1/*.go


# Ensure the required version of cli tooling in tools/tools.go is installed.
# The versions are stored in ../go.mod as usual and this same approach is used in
# both octant and pinniped for build tooling.
# See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
cli-dependencies:
	go get \
		github.com/spf13/cobra/cobra \
		google.golang.org/grpc/cmd/protoc-gen-go-grpc \
		google.golang.org/protobuf/cmd/protoc-gen-go

# Not sure if it's worth having targets for buf, but at least the targets can
# be tooling agnostic.
# NOTE: I have an issue with buf in that if I try to update the protobuf dependencies
# (defined in the buf.yaml) with `buf beta mod update`, the `buf.lock` file updates
# accordingly, but then am unable to run `buf generate` or any other command that
# pulls the dependencies to the local cache. I've reported this and created a minimal
# demo of the issue (at least, for me) at: https://github.com/bufbuild/buf/issues/294#issuecomment-837839297
# I'll update once I've resolved that.
generate:
	buf generate

run: build-plugins
	go run main.go --plugin-dir devel/

test:
	go test ./...

.PHONY: build-plugins cli-dependencies generate run test
